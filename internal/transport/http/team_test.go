package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newServiceTeam() *service.Team {
	return &service.Team{
		ID:          model.MustNewID(model.ResourceTypeTeam),
		Name:        "Platform",
		Description: "Platform team",
		MemberCount: convert.ToPointer(int64(3)),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

func TestCreateTeamJSONRequestBodyToCreateTeamOpts(t *testing.T) {
	t.Parallel()

	t.Run("with description", func(t *testing.T) {
		t.Parallel()
		description := "Platform team"
		opts := createTeamJSONRequestBodyToCreateTeamOpts(&api.V1OrganizationTeamsCreateJSONRequestBody{
			Name:        "Platform",
			Description: &description,
		})
		assert.Equal(t, "Platform", opts.Name)
		assert.Equal(t, description, opts.Description)
	})

	t.Run("name only", func(t *testing.T) {
		t.Parallel()
		opts := createTeamJSONRequestBodyToCreateTeamOpts(&api.V1OrganizationTeamsCreateJSONRequestBody{
			Name: "Platform",
		})
		assert.Equal(t, "Platform", opts.Name)
		assert.Empty(t, opts.Description)
	})
}

func TestUpdateTeamJSONRequestBodyToUpdateTeamOpts(t *testing.T) {
	t.Parallel()

	name := "updated-team"
	opts := updateTeamJSONRequestBodyToUpdateTeamOpts(&api.V1OrganizationTeamUpdateJSONRequestBody{
		Name:        &name,
		Description: optional.Some("updated description"),
	})
	assert.Equal(t, optional.Some("updated-team"), opts.Name)
	assert.Equal(t, optional.Some("updated description"), opts.Description)
}

func TestTeamToDTO(t *testing.T) {
	t.Parallel()

	team := newServiceTeam()
	dto := teamToDTO(team)
	assert.Equal(t, team.ID.String(), dto.Id)
	assert.Equal(t, team.Name, dto.Name)
	require.NotNil(t, dto.Description)
	assert.Equal(t, team.Description, *dto.Description)
	require.NotNil(t, dto.MemberCount)
	assert.Equal(t, int64(3), *dto.MemberCount)
}
