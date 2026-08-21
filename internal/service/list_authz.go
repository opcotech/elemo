package service

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

func listGrantCoversRoot(scopeIDs, ancestry []model.ID) bool {
	if len(scopeIDs) == 0 || len(ancestry) == 0 {
		return false
	}
	granted := make(map[string]struct{}, len(scopeIDs))
	for _, id := range scopeIDs {
		granted[id.String()] = struct{}{}
	}
	for _, id := range ancestry {
		if _, ok := granted[id.String()]; ok {
			return true
		}
	}
	return false
}

// resolvedListScopeIDs returns grant scopes that still need a per-row EXISTS
// filter. allowed is false when the actor has no grant for action. A grant on
// the list root or any ancestor already authorizes every descendant, so
// scopeIDs is nil and the query can skip that predicate.
func resolvedListScopeIDs(ctx context.Context, perm PermissionService, root model.ID, action model.Action) (scopeIDs []model.ID, allowed bool, err error) {
	scopeIDs, err = perm.CtxUserListGrantScopes(ctx, action)
	if err != nil {
		return nil, false, err
	}
	if len(scopeIDs) == 0 {
		return nil, false, nil
	}
	ancestry, err := perm.ListScopeAncestry(ctx, root)
	if err != nil {
		return nil, false, err
	}
	if listGrantCoversRoot(scopeIDs, ancestry) {
		return nil, true, nil
	}
	return scopeIDs, true, nil
}
