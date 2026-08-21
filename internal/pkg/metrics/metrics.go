// Package metrics declares Elemo Prometheus instrumentation.
//
// Label values are bounded: query names, cache scopes, and coarse operation
// kinds. Never put Cypher, SQL, URLs, or entity IDs on labels.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ResultOK    = "ok"
	ResultError = "error"
	ResultHit   = "hit"
	ResultMiss  = "miss"

	Neo4jUncompiledQuery = "uncompiled"

	IssueListScopeProject   = "project"
	IssueListScopeNamespace = "namespace"
	IssueListScopeUser      = "user"
)

var queryDurationBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30,
}

var queryRowBuckets = []float64{
	1, 5, 10, 25, 50, 100, 250, 500, 1000,
}

// Dashboard: histogram_quantile(0.99, sum by (le, query) (rate(elemo_neo4j_query_duration_seconds_bucket[5m])))
// Dashboard: topk(10, histogram_quantile(0.99, sum by (le, query) (rate(elemo_neo4j_query_duration_seconds_bucket[5m]))))
var Neo4jQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "elemo",
		Subsystem: "neo4j",
		Name:      "query_duration_seconds",
		Help:      "Neo4j query duration in seconds.",
		Buckets:   queryDurationBuckets,
	},
	[]string{"query", "result"},
)

// Dashboard: histogram_quantile(0.95, sum by (le, query) (rate(elemo_neo4j_query_rows_bucket[5m])))
var Neo4jQueryRows = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "elemo",
		Subsystem: "neo4j",
		Name:      "query_rows",
		Help:      "Rows returned by a Neo4j query.",
		Buckets:   queryRowBuckets,
	},
	[]string{"query"},
)

// Dashboard: histogram_quantile(0.99, sum by (le, operation) (rate(elemo_pg_query_duration_seconds_bucket[5m])))
var PGQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "elemo",
		Subsystem: "pg",
		Name:      "query_duration_seconds",
		Help:      "Postgres query duration in seconds.",
		Buckets:   queryDurationBuckets,
	},
	[]string{"operation", "result"},
)

// Dashboard: histogram_quantile(0.99, sum by (le, scope) (rate(elemo_issue_list_duration_seconds_bucket[5m])))
var IssueListDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "elemo",
		Subsystem: "issue_list",
		Name:      "duration_seconds",
		Help:      "Issue list request duration including cache lookup.",
		Buckets:   queryDurationBuckets,
	},
	[]string{"scope", "result"},
)

// Dashboard: sum by (scope, result) (rate(elemo_issue_list_cache_requests_total[5m]))
// Dashboard: sum(rate(elemo_issue_list_cache_requests_total{result="hit"}[5m])) / sum(rate(elemo_issue_list_cache_requests_total[5m]))
var IssueListCacheRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "elemo",
		Subsystem: "issue_list",
		Name:      "cache_requests_total",
		Help:      "Issue list cache lookups by scope and result.",
	},
	[]string{"scope", "result"},
)

func resultLabel(err error) string {
	if err != nil {
		return ResultError
	}
	return ResultOK
}

// ObserveNeo4jQuery records Neo4j query latency and row count.
func ObserveNeo4jQuery(query string, duration time.Duration, rows int, err error) {
	if query == "" {
		query = Neo4jUncompiledQuery
	}
	Neo4jQueryDuration.WithLabelValues(query, resultLabel(err)).Observe(duration.Seconds())
	if err == nil {
		Neo4jQueryRows.WithLabelValues(query).Observe(float64(rows))
	}
}

// ObservePGQuery records Postgres query latency by operation kind.
func ObservePGQuery(operation string, duration time.Duration, err error) {
	PGQueryDuration.WithLabelValues(operation, resultLabel(err)).Observe(duration.Seconds())
}

// ObserveIssueList records issue-list latency and cache hit/miss/error.
func ObserveIssueList(scope, result string, duration time.Duration) {
	IssueListDuration.WithLabelValues(scope, result).Observe(duration.Seconds())
	IssueListCacheRequests.WithLabelValues(scope, result).Inc()
}
