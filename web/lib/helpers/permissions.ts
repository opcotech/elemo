import { Permission, PermissionKind, ResourceType } from '@/lib/api';

export type Actions = {
  canView: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canCreate?: boolean;
};

function hasPermission(permission: Permission, kind: PermissionKind) {
  return permission.kind === kind || permission.kind === PermissionKind._;
}

function hasSomePermission(permissions: Permission[], kind: PermissionKind) {
  return permissions.some((p) => p.kind === kind || p.kind === PermissionKind._);
}

export function getActions(data: Permission | Permission[]): Actions {
  let getter = data instanceof Array ? hasSomePermission : hasPermission;

  const canCreate = getter(data as never, PermissionKind.CREATE);
  const canDelete = getter(data as never, PermissionKind.DELETE);
  const canEdit = getter(data as never, PermissionKind.WRITE);
  const canView = getter(data as never, PermissionKind.READ) || canCreate || canEdit || canDelete;

  return {
    canView,
    canEdit,
    canDelete,
    canCreate
  };
}

export function formatResourceId(resourceType: ResourceType, resourceId: string): string {
  return `${resourceType}:${resourceId}`;
}
