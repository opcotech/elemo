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
	assertNoBannedProjectKeys(t, spec)
}

func TestSmokeScenarioCounts(t *testing.T) {
	t.Parallel()

	spec := smokeScenario()
	assert.Equal(t, 8, spec.main.userCount)
	assert.Equal(t, adminEmail, spec.main.adminEmail)
	assert.Len(t, spec.partners, 5)
	assert.Equal(t, 1, spec.migratedProjectCount())
	assert.Less(t, spec.documentCount, 200)
	assert.Equal(t, 20, spec.documentCount)
	assertNoBannedProjectKeys(t, spec)

	wantAdmins := map[string]string{
		"Kite Analytics":        "maya.chen@kite.example",
		"Harbor Logistics":      "luis.navarro@harbor.example",
		"Nimbus Cloud":          "priya.shah@nimbus.example",
		"Brightline Design":     "aisha.okoro@brightline.example",
		"Fieldstone Consulting": "jordan.lee@fieldstone.example",
	}
	gotAdmins := make(map[string]string, len(spec.partners))
	for _, partner := range spec.partners {
		gotAdmins[partner.name] = partner.adminEmail
	}
	assert.Equal(t, wantAdmins, gotAdmins)

	keysByOrg := projectKeysByOrg(spec)
	for _, collab := range spec.collaborations {
		if collab.toProjectKey == "" {
			continue
		}
		_, ok := keysByOrg[collab.toOrg][collab.toProjectKey]
		assert.True(t, ok, "collaboration %s -> %s missing project %s", collab.fromOrg, collab.toOrg, collab.toProjectKey)
	}
}

func assertNoBannedProjectKeys(t *testing.T, spec scenarioSpec) {
	t.Helper()

	banned := map[string]struct{}{"ANAL": {}}
	for orgName, keys := range projectKeysByOrg(spec) {
		for key := range keys {
			_, hit := banned[key]
			assert.False(t, hit, "org %s has banned project key %s", orgName, key)
		}
	}
	for _, collab := range spec.collaborations {
		_, hit := banned[collab.toProjectKey]
		assert.False(t, hit, "collaboration targets banned project key %s", collab.toProjectKey)
	}
}

func projectKeysByOrg(spec scenarioSpec) map[string]map[string]struct{} {
	orgs := append([]orgSpec{spec.main}, spec.partners...)
	out := make(map[string]map[string]struct{}, len(orgs))
	for _, org := range orgs {
		keys := make(map[string]struct{})
		for _, ns := range org.namespaces {
			for _, project := range ns.projects {
				keys[project.key] = struct{}{}
			}
		}
		out[org.name] = keys
	}
	return out
}

func TestScenarioSlugUniqueness(t *testing.T) {
	t.Parallel()

	for _, spec := range []scenarioSpec{fullScenario(), smokeScenario()} {
		orgSlugs := make(map[string]string)
		platformOrgs := 0
		for _, org := range append([]orgSpec{spec.main}, spec.partners...) {
			require.NotEmpty(t, org.slug, "org %s missing slug", org.name)
			_, exists := orgSlugs[org.slug]
			require.False(t, exists, "duplicate organization slug %s", org.slug)
			orgSlugs[org.slug] = org.name

			nsSlugs := make(map[string]string)
			for _, ns := range org.namespaces {
				require.NotEmpty(t, ns.slug, "namespace %s in %s missing slug", ns.name, org.name)
				_, exists := nsSlugs[ns.slug]
				require.False(t, exists, "duplicate namespace slug %s in %s", ns.slug, org.name)
				nsSlugs[ns.slug] = ns.name
				if ns.slug == "platform" {
					platformOrgs++
				}
			}
		}
		require.Equal(t, "elemo", spec.main.slug)
		assert.GreaterOrEqual(t, platformOrgs, 2, "platform namespace slug should be reused across organizations")
	}
}
