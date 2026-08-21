package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/metrics"
)

// Neo4jRecordReader decodes a single Neo4j record into T.
type Neo4jRecordReader[T any] func(record *neo4j.Record) (T, error)

// Neo4jExecuteReadPlan runs a root query and optional loaders inside one
// managed read transaction. All results are copied before the callback returns.
func Neo4jExecuteReadPlan(ctx context.Context, db *Neo4jDatabase, plan QueryPlan, run func(tx neo4j.ManagedTransaction) error) error {
	if err := plan.Validate(); err != nil {
		return err
	}

	session := db.ReadSession(ctx)
	defer func(ctx context.Context, sess neo4j.Session) {
		if err := sess.Close(ctx); err != nil {
			log.Error(ctx, err)
		}
	}(ctx, session)

	_, err := neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (any, error) {
		if err := run(tx); err != nil {
			return nil, err
		}
		return struct{}{}, nil
	})
	return err
}

// Neo4jRunQuery executes a compiled query and collects all records via reader.
func Neo4jRunQuery[T any](ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, reader Neo4jRecordReader[T]) (items []T, qs *QuerySummary, err error) {
	started := time.Now()
	defer func() {
		metrics.ObserveNeo4jQuery(query.Name, time.Since(started), len(items), err)
	}()

	result, err := tx.Run(ctx, query.Cypher, query.Params)
	if err != nil {
		return nil, nil, err
	}

	items = make([]T, 0)
	for result.Next(ctx) {
		item, err := reader(result.Record())
		if err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := result.Err(); err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return nil, nil, err
	}

	qs = &QuerySummary{
		Name:                 query.Name,
		Fingerprint:          query.Fingerprint(),
		Counters:             summary.Counters(),
		ResultAvailableAfter: summary.ResultAvailableAfter(),
		ResultConsumedAfter:  summary.ResultConsumedAfter(),
	}
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(
			attribute.String("db.neo4j.query.name", qs.Name),
			attribute.String("db.neo4j.query.fingerprint", qs.Fingerprint),
			attribute.Int64("db.neo4j.query.result_available_ms", qs.ResultAvailableAfter.Milliseconds()),
			attribute.Int64("db.neo4j.query.result_consumed_ms", qs.ResultConsumedAfter.Milliseconds()),
			attribute.Int64("db.neo4j.query.elapsed_ms", time.Since(started).Milliseconds()),
		)
	}

	return items, qs, nil
}

// Neo4jRunQuerySingle executes a compiled query expecting exactly one record.
func Neo4jRunQuerySingle[T any](ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, reader Neo4jRecordReader[T]) (T, *QuerySummary, error) {
	var zero T
	items, summary, err := Neo4jRunQuery(ctx, tx, query, reader)
	if err != nil {
		return zero, summary, err
	}
	if len(items) == 0 {
		return zero, summary, ErrNotFound
	}
	if len(items) > 1 {
		return zero, summary, ErrMalformedResult
	}
	return items[0], summary, nil
}
