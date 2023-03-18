package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opcotech/elemo/internal/model"
)

func TestWithUserEmailAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeUserEmail, Value: attribute.StringValue("test@example.com")}
	assert.Equal(t, attr, WithUserEmailAttribute("test@example.com"))
}

func TestWithQueryCollectionAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryLabels, Value: attribute.StringSliceValue([]string{"test"})}
	assert.Equal(t, attr, WithQueryLabelsAttribute("test"))
}

func TestWithQueryKeyAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryID, Value: attribute.StringValue("1234")}
	assert.Equal(t, attr, WithQueryIDAttribute("1234"))
}

func TestWithQueryLimitAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryLimit, Value: attribute.IntValue(10)}
	assert.Equal(t, attr, WithQueryLimitAttribute(10))
}

func TestWithQueryOffsetAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryOffset, Value: attribute.IntValue(10)}
	assert.Equal(t, attr, WithQueryOffsetAttribute(10))
}

func TestWithQueryDepthAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryDepth, Value: attribute.IntValue(2)}
	assert.Equal(t, attr, WithQueryDepthAttribute(2))
}

func TestWithQueryPatchAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryPatchLen, Value: attribute.IntValue(2)}
	assert.Equal(t, attr, WithQueryPatchLenAttribute(2))
}

func TestWithQueryResultLenAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeQueryResultLen, Value: attribute.IntValue(2)}
	assert.Equal(t, attr, WithQueryResultLenAttribute(2))
}

func TestWithSystemHealthStatusAttribute(t *testing.T) {
	t.Parallel()

	attr := attribute.KeyValue{Key: AttributeSystemHealthStatus, Value: attribute.StringValue("healthy")}
	assert.Equal(t, attr, WithSystemHealthStatusAttribute(model.HealthStatusHealthy))
}
