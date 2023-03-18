package tracing

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	AttributeUserEmail      = "user.email"
	AttributeQueryLabels    = "query.labels"
	AttributeQueryID        = "query.id"
	AttributeQueryLimit     = "query.limit"
	AttributeQueryOffset    = "query.offset"
	AttributeQueryDepth     = "query.depth"
	AttributeQueryPatchLen  = "query.patch_length"
	AttributeQueryResultLen = "query.result_length"
)

// WithUserEmailAttribute adds the user email to the span.
func WithUserEmailAttribute(email string) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeUserEmail,
		Value: attribute.StringValue(email),
	}
}

// WithQueryLabelsAttribute adds the node labels to the span.
func WithQueryLabelsAttribute(labels ...string) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryLabels,
		Value: attribute.StringSliceValue(labels),
	}
}

// WithQueryIDAttribute adds the inner of the updated entity to the span.
func WithQueryIDAttribute(id string) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryID,
		Value: attribute.StringValue(id),
	}
}

// WithQueryLimitAttribute adds the query limit to the span.
func WithQueryLimitAttribute(limit int) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryLimit,
		Value: attribute.IntValue(limit),
	}
}

// WithQueryOffsetAttribute adds the query offset to the span.
func WithQueryOffsetAttribute(offset int) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryOffset,
		Value: attribute.IntValue(offset),
	}
}

// WithQueryDepthAttribute adds the query depth to the span.
func WithQueryDepthAttribute(depth int) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryDepth,
		Value: attribute.IntValue(depth),
	}
}

// WithQueryPatchLenAttribute adds the query patch to the span.
func WithQueryPatchLenAttribute(length int) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryPatchLen,
		Value: attribute.IntValue(length),
	}
}

// WithQueryResultLenAttribute adds the query result length to the span.
func WithQueryResultLenAttribute(length int) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   AttributeQueryResultLen,
		Value: attribute.IntValue(length),
	}
}
