package repository

import "github.com/opcotech/elemo/internal/model"

func PermissionGetQuery(id model.ID) (CompiledQuery, error) {
	if err := id.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name:   "permission.get",
		Cypher: "MATCH (s)-[p:" + EdgeKindHasPermission.String() + " {id: $id}]->(t) RETURN s, p, t",
		Params: map[string]any{"id": id.String()},
	}, nil
}

func PermissionGetBySubjectAndTargetQuery(subjectID, targetID model.ID) (CompiledQuery, error) {
	if err := subjectID.Validate(); err != nil {
		return CompiledQuery{}, err
	}
	if err := targetID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "permission.get_by_subject_and_target",
		Cypher: `
			MATCH (s:` + subjectID.Label() + ` {id: $subject_id}), (t:` + targetID.Label() + ` {id: $target_id})
			MATCH path=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(t)
			WITH s, t, last(relationships(path)) AS p
			WHERE type(p) = "` + EdgeKindHasPermission.String() + `"
			RETURN s, p, t`,
		Params: map[string]any{
			"subject_id": subjectID.String(),
			"target_id":  targetID.String(),
		},
	}, nil
}

func PermissionGetBySubjectQuery(subjectID model.ID) (CompiledQuery, error) {
	if err := subjectID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "permission.get_by_subject",
		Cypher: `
			MATCH (s:` + subjectID.Label() + ` {id: $subject_id})-[p:` + EdgeKindHasPermission.String() + `]->(t)
			RETURN s, p, t`,
		Params: map[string]any{"subject_id": subjectID.String()},
	}, nil
}

func PermissionGetByTargetQuery(targetID model.ID) (CompiledQuery, error) {
	if err := targetID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "permission.get_by_target",
		Cypher: `
			MATCH (s)-[p:` + EdgeKindHasPermission.String() + `]->(t:` + targetID.Label() + ` {id: $target_id})
			RETURN s, p, t`,
		Params: map[string]any{"target_id": targetID.String()},
	}, nil
}
