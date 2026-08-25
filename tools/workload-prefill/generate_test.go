package main

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/service"
)

func TestPickIssueAssignee(t *testing.T) {
	t.Parallel()

	admin := &service.User{Username: "demo"}
	other := &service.User{Username: "other"}
	assignees := []*service.User{admin, other}

	t.Run("skips archived work", func(t *testing.T) {
		t.Parallel()
		rng := rand.New(rand.NewPCG(1, 1)) //nolint:gosec // deterministic test fixture
		assert.Nil(t, pickIssueAssignee(rng, assignees, false, true, 0))
	})

	t.Run("assigns every fifth live issue to the demo admin", func(t *testing.T) {
		t.Parallel()
		rng := rand.New(rand.NewPCG(1, 1)) //nolint:gosec // deterministic test fixture
		got := pickIssueAssignee(rng, assignees, true, true, 0)
		require.NotNil(t, got)
		assert.Equal(t, admin, got)
		got = pickIssueAssignee(rng, assignees, true, true, 5)
		require.NotNil(t, got)
		assert.Equal(t, admin, got)
	})

	t.Run("does not force partner admins onto a cadence", func(t *testing.T) {
		t.Parallel()
		unassigned := 0
		for seed := range 20 {
			rng := rand.New(rand.NewPCG(uint64(seed), 1)) //nolint:gosec // deterministic test fixture
			if pickIssueAssignee(rng, assignees, true, false, 0) == nil {
				unassigned++
			}
		}
		assert.Greater(t, unassigned, 0)
	})
}
