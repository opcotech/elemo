package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlphaKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AAA", alphaKey(0))
	assert.Equal(t, "AAB", alphaKey(1))
	assert.Equal(t, "AAZ", alphaKey(25))
	assert.Equal(t, "ABA", alphaKey(26))

	keys := make(map[string]struct{}, 250)
	for i := range 250 {
		key := alphaKey(i)
		require.Regexp(t, `^[A-Z]{3}$`, key)
		_, exists := keys[key]
		require.False(t, exists, "duplicate key %s", key)
		keys[key] = struct{}{}
	}
	assert.Len(t, keys, 250)
}

func TestFullScenarioCounts(t *testing.T) {
	t.Parallel()

	spec := fullScenario()
	assert.Equal(t, 300, spec.main.userCount)
	assert.Len(t, spec.main.teams, 12)
	assert.Len(t, spec.partners, 5)
	assert.Equal(t, 19, spec.liveProjectCount())
	assert.Equal(t, 72, spec.migratedProjectCount())
	assert.GreaterOrEqual(t, spec.liveProjectMinIssues(), 100)
	assert.GreaterOrEqual(t, spec.documentCount, 200)
	assert.LessOrEqual(t, spec.documentCount, 350)

	requiredTeams := []string{
		"Engineering", "Design", "Management", "Operations", "Human Resources",
		"Sales", "Client Relations",
	}
	got := make(map[string]struct{}, len(spec.main.teams))
	for _, team := range spec.main.teams {
		got[team.name] = struct{}{}
	}
	for _, name := range requiredTeams {
		_, ok := got[name]
		assert.True(t, ok, "missing team %s", name)
	}

	for _, ns := range spec.main.namespaces {
		if ns.name == migratedNS {
			assert.GreaterOrEqual(t, ns.migratedIssueMin, 75)
			assert.LessOrEqual(t, ns.migratedIssueMax, 300)
			continue
		}
		assert.GreaterOrEqual(t, len(ns.projects), 2)
		assert.LessOrEqual(t, len(ns.projects), 5)
		for _, project := range ns.projects {
			assert.GreaterOrEqual(t, project.issueCount, 100)
			assert.True(t, project.live)
		}
	}

	assert.NotEmpty(t, spec.collaborations)
}

func TestSmokeScenarioCounts(t *testing.T) {
	t.Parallel()

	spec := smokeScenario()
	assert.Equal(t, 8, spec.main.userCount)
	assert.Len(t, spec.partners, 1)
	assert.Equal(t, 1, spec.migratedProjectCount())
	assert.Less(t, spec.documentCount, 200)
	assert.Equal(t, 20, spec.documentCount)
}
