package repository

import (
	"context"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
)

// MakeUserSystemOwner elevates the user to system owner.
func MakeUserSystemOwner(userID model.ID, db *neo4j.Database) error {
	ctx := context.Background()

	cypher := `
	MATCH (u:` + userID.Label() + ` {id: $id})
	MATCH (r:` + model.ResourceTypeRole.String() + ` {id: $role_label, system: true})
	CREATE (u)-[:` + neo4j.EdgeKindMemberOf.String() + `]->(r)`

	params := map[string]any{
		"id":         userID.String(),
		"role_label": model.SystemRoleOwner.String(),
		"perm_kind":  model.PermissionKindAll.String(),
	}

	_, err := db.WriteSession(ctx).Run(ctx, cypher, params)
	if err != nil {
		return err
	}

	// This is an ugly hack to make sure the transaction is committed before
	// the next test starts. Without this, the next test will may be flaky.
	time.Sleep(1 * time.Second)
	return nil
}
