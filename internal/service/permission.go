package service

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

func ctxUserPermitted(ctx context.Context, repo repository.PermissionRepository, target model.ID, permissions ...model.PermissionKind) bool {
	span := trace.SpanFromContext(ctx)

	var hasPerm bool
	var err error

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	span.AddEvent("check permission")
	hasPerm, err = repo.HasPermission(ctx, userID, target, append(permissions, model.PermissionKindAll)...)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		return false
	}
	span.AddEvent("permission checked")

	return hasPerm
}
