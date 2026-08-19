package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// MakeUserSystemOwner grants the user organization.create on the Installation
// node so they can create organizations in tests.
func MakeUserSystemOwner(userID model.ID, db *repository.Neo4jDatabase) error {
	ctx := context.Background()
	installationID := model.InstallationID()

	cypher := `
	MERGE (i:` + installationID.Label() + ` {id: $installation_id})
	ON CREATE SET i.system = true, i.created_at = datetime()
	SET i.system = true
	WITH i
	MATCH (u:` + userID.Label() + ` {id: $user_id})
	SET u:` + model.LabelPrincipal + `
	MERGE (u)-[g:` + repository.EdgeKindGranted.String() + `]->(i)
	ON CREATE SET g.id = $grant_id, g.actions = $actions, g.created_at = datetime(), g.role_id = ""
	ON MATCH SET g.actions = $actions, g.updated_at = datetime()`

	params := map[string]any{
		"installation_id": installationID.String(),
		"user_id":         userID.String(),
		"grant_id":        model.MustNewID(model.ResourceTypePermission).String(),
		"actions":         []string{model.ActionOrganizationCreate.String()},
	}

	return repository.Neo4jExecuteWriteAndConsume(ctx, db, cypher, params)
}
