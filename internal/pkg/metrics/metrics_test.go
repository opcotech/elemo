package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func histogramSampleCount(t *testing.T, name, query, result string) uint64 {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["query"] == query && labels["result"] == result {
				return metric.GetHistogram().GetSampleCount()
			}
		}
	}
	return 0
}

func TestObserveNeo4jQuery_recordsCompiledQueryName(t *testing.T) {
	query := "issue.list_for_project"
	before := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", query, ResultOK)

	ObserveNeo4jQuery(query, 25*time.Millisecond, 10, nil)

	after := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", query, ResultOK)
	require.GreaterOrEqual(t, after, before+1)
}

func TestObserveNeo4jQuery_recordsErrors(t *testing.T) {
	query := "issue.list_for_namespace"
	before := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", query, ResultError)

	ObserveNeo4jQuery(query, time.Millisecond, 0, errors.New("boom"))

	after := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", query, ResultError)
	require.GreaterOrEqual(t, after, before+1)
}

func TestObserveNeo4jQuery_emptyNameUsesUncompiled(t *testing.T) {
	before := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", Neo4jUncompiledQuery, ResultOK)

	ObserveNeo4jQuery("", time.Millisecond, 0, nil)

	after := histogramSampleCount(t, "elemo_neo4j_query_duration_seconds", Neo4jUncompiledQuery, ResultOK)
	require.GreaterOrEqual(t, after, before+1)
}

func TestObserveIssueList_incrementsCacheCounter(t *testing.T) {
	before := testutil.ToFloat64(IssueListCacheRequests.WithLabelValues(IssueListScopeProject, ResultHit))

	ObserveIssueList(IssueListScopeProject, ResultHit, time.Millisecond)

	after := testutil.ToFloat64(IssueListCacheRequests.WithLabelValues(IssueListScopeProject, ResultHit))
	require.GreaterOrEqual(t, after, before+1)
}

func TestObserveCacheGet_incrementsCounterByResult(t *testing.T) {
	for _, result := range []string{ResultHit, ResultMiss, ResultError} {
		t.Run(result, func(t *testing.T) {
			before := testutil.ToFloat64(CacheRequests.WithLabelValues(result))

			ObserveCacheGet(result)

			after := testutil.ToFloat64(CacheRequests.WithLabelValues(result))
			require.GreaterOrEqual(t, after, before+1)
		})
	}
}

func TestSetRedisStatsFunc_reportsPoolStats(t *testing.T) {
	SetRedisStatsFunc(func() RedisPoolStats {
		return RedisPoolStats{Total: 10, Idle: 4, Timeouts: 2}
	})

	stats := loadRedisStats()
	require.Equal(t, 10.0, stats.Total)
	require.Equal(t, 4.0, stats.Idle)
	require.Equal(t, 2.0, stats.Timeouts)
}
