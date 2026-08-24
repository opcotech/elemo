package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

const (
	testCtxKey CtxKey = "test-ctx-key"
)

func TestCtxUserID(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		expectedResult string
		expectNonEmpty bool
	}{
		{
			name:           "should return empty string when context has no user ID",
			ctx:            context.Background(),
			expectedResult: "",
		},
		{
			name:           "should return machine user string when context has machine user",
			ctx:            context.WithValue(context.Background(), CtxKeyUserID, CtxMachineUser),
			expectedResult: "machine",
		},
		{
			name: "should return user ID string when context has user ID",
			ctx: context.WithValue(
				context.Background(),
				CtxKeyUserID,
				model.MustNewID(model.ResourceTypeUser),
			),
			expectNonEmpty: true,
		},
		{
			name:           "should return empty string when context has invalid user ID type",
			ctx:            context.WithValue(context.Background(), CtxKeyUserID, "invalid-type"),
			expectedResult: "",
		},
		{
			name: "should return empty string when context has different key",
			ctx: context.WithValue(
				context.Background(),
				testCtxKey,
				model.MustNewID(model.ResourceTypeUser),
			),
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expectNonEmpty {
				result := CtxUserID(tt.ctx)
				assert.NotEmpty(t, result)
			} else {
				result := CtxUserID(tt.ctx)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCtxUserIDWithMachineUser(t *testing.T) {
	t.Run("should handle machine user correctly", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), CtxKeyUserID, CtxMachineUser)
		result := CtxUserID(ctx)
		assert.Equal(t, "machine", result)
	})

	t.Run("should return empty string for different machine user kind", func(t *testing.T) {
		t.Parallel()

		differentMachineUser := CtxMachineUserKind("different-machine")
		ctx := context.WithValue(context.Background(), CtxKeyUserID, differentMachineUser)
		result := CtxUserID(ctx)
		assert.Equal(t, "", result)
	})
}

func TestCtxUserIDWithUserID(t *testing.T) {
	t.Run("should extract user ID string", func(t *testing.T) {
		t.Parallel()

		userID := model.MustNewID(model.ResourceTypeUser)
		ctx := context.WithValue(context.Background(), CtxKeyUserID, userID)
		result := CtxUserID(ctx)

		require.NotEmpty(t, result)
		assert.Equal(t, userID.String(), result)
	})

	t.Run("should handle multiple users with different IDs", func(t *testing.T) {
		t.Parallel()

		userID1 := model.MustNewID(model.ResourceTypeUser)
		userID2 := model.MustNewID(model.ResourceTypeUser)

		ctx1 := context.WithValue(context.Background(), CtxKeyUserID, userID1)
		ctx2 := context.WithValue(context.Background(), CtxKeyUserID, userID2)

		result1 := CtxUserID(ctx1)
		result2 := CtxUserID(ctx2)

		require.NotEmpty(t, result1)
		require.NotEmpty(t, result2)
		assert.NotEqual(t, result1, result2, "Different users should have different IDs")
	})
}

func TestCtxUserIDEdgeCases(t *testing.T) {
	t.Run("should handle empty context", func(t *testing.T) {
		t.Parallel()

		result := CtxUserID(context.Background())
		assert.Equal(t, "", result)
	})

	t.Run("should handle context with nil value", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), CtxKeyUserID, nil)
		result := CtxUserID(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("should handle context with wrong type assertion", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), CtxKeyUserID, 123)
		result := CtxUserID(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("should handle context with string that is not machine user", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), CtxKeyUserID, "not-machine")
		result := CtxUserID(ctx)
		assert.Equal(t, "", result)
	})
}

func TestCtxKeyConstants(t *testing.T) {
	t.Run("should have correct user ID key", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, CtxKey("userID"), CtxKeyUserID)
	})

	t.Run("should have correct logger key", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, CtxKey("logger"), CtxKeyLogger)
	})

	t.Run("should have correct machine user kind", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, CtxMachineUserKind("machine"), CtxMachineUser)
	})
}

func TestCtxKeyType(t *testing.T) {
	t.Run("should be able to use CtxKey as context key", func(t *testing.T) {
		t.Parallel()

		key := CtxKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")

		value, ok := ctx.Value(key).(string)
		require.True(t, ok)
		assert.Equal(t, "test-value", value)
	})
}

func TestCtxMachineUserKindType(t *testing.T) {
	t.Run("should be able to use CtxMachineUserKind as context value", func(t *testing.T) {
		t.Parallel()

		machineUser := CtxMachineUserKind("test-machine")
		ctx := context.WithValue(context.Background(), CtxKeyUserID, machineUser)

		value, ok := ctx.Value(CtxKeyUserID).(CtxMachineUserKind)
		require.True(t, ok)
		assert.Equal(t, machineUser, value)
	})
}

func TestCtxUserIDValue(t *testing.T) {
	t.Parallel()

	t.Run("returns user id", func(t *testing.T) {
		t.Parallel()
		userID := model.MustNewID(model.ResourceTypeUser)
		ctx := context.WithValue(context.Background(), CtxKeyUserID, userID)
		got, ok := CtxUserIDValue(ctx)
		require.True(t, ok)
		assert.Equal(t, userID, got)
	})

	t.Run("missing user", func(t *testing.T) {
		t.Parallel()
		_, ok := CtxUserIDValue(context.Background())
		assert.False(t, ok)
	})
}

func TestCtxUserIDNestedContext(t *testing.T) {
	t.Run("should work with nested contexts", func(t *testing.T) {
		t.Parallel()

		userID := model.MustNewID(model.ResourceTypeUser)
		ctx1 := context.WithValue(context.Background(), CtxKeyUserID, userID)
		ctx2 := context.WithValue(ctx1, testCtxKey, "other-value")

		result := CtxUserID(ctx2)
		assert.Equal(t, userID.String(), result)
	})

	t.Run("should work with context cancellation", func(t *testing.T) {
		t.Parallel()

		userID := model.MustNewID(model.ResourceTypeUser)
		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, CtxKeyUserID, userID)

		result := CtxUserID(ctx)
		assert.Equal(t, userID.String(), result)

		cancel()

		resultAfterCancel := CtxUserID(ctx)
		assert.Equal(t, userID.String(), resultAfterCancel)
	})
}
