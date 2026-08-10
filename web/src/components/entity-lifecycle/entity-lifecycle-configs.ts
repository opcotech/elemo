import type { EntityLifecycleConfig } from "./entity-lifecycle";

import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1OrganizationGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import type {
  Namespace,
  Options,
  Organization,
  PartialProject,
  Permission,
  Project,
  Role,
  V1NamespaceDeleteData,
  V1OrganizationDeleteData,
  V1OrganizationRoleDeleteData,
  V1ProjectDeleteData,
} from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { zOrganizationStatus } from "@/lib/client/zod.gen";

interface OrganizationLifecycleEntity extends Pick<
  Organization,
  "id" | "name" | "status"
> {}

interface NamespaceLifecycleContext {
  organizationId: string;
}

interface ProjectLifecycleContext extends NamespaceLifecycleContext {
  namespaceId: string;
}

interface RoleLifecycleContext {
  organizationId: string;
}

type OrganizationDeleteVariables = Options<V1OrganizationDeleteData>;
type NamespaceDeleteVariables = Options<V1NamespaceDeleteData>;
type ProjectDeleteVariables = Options<V1ProjectDeleteData>;
type RoleDeleteVariables = Options<V1OrganizationRoleDeleteData>;

const hasDeletePermission = (_entity: unknown, permissions: Permission[]) =>
  can(permissions, "delete");

export const organizationLifecycleConfig: EntityLifecycleConfig<
  OrganizationLifecycleEntity,
  undefined,
  OrganizationDeleteVariables
> & {
  dangerZone: NonNullable<
    EntityLifecycleConfig<
      OrganizationLifecycleEntity,
      undefined,
      OrganizationDeleteVariables
    >["dangerZone"]
  >;
} = {
  entityName: "Organization",
  deleteDialog: {
    title: (organization) =>
      `Are you sure you want to delete ${organization.name}?`,
    description:
      "This will mark the organization as deleted. This action cannot be undone.",
    consequences: [
      "The organization will be marked as deleted",
      "All organization members will lose access",
      "Organization data will be hidden from listings",
      "You will be redirected to the organizations list",
    ],
  },
  dangerZone: {
    dataSection: "organization-danger-zone",
    description: "Irreversible actions for this organization",
    summary:
      "Deleting an organization will mark it as deleted and hide it from listings. This action cannot be undone.",
    consequences: [
      "All organization members will lose access",
      "Organization data will be hidden from search and listings",
      "The organization will be marked as deleted",
      "This action is permanent and cannot be reversed",
    ],
    buttonLabel: "Delete Organization",
  },
  canDelete: (organization, permissions) =>
    organization.status === zOrganizationStatus.enum.active &&
    can(permissions, "delete"),
  deleteVariables: (organization) => ({
    path: { id: organization.id },
    query: { force: false },
  }),
  queryKeys: (organization) => [
    v1OrganizationGetOptions({
      path: { id: organization.id },
    }).queryKey,
    v1OrganizationsGetOptions().queryKey,
    accessibleNamespacesQueryKey,
  ],
  navigateAfterDelete: (navigate) =>
    navigate({ to: "/settings/organizations" }),
};

export const namespaceLifecycleConfig: EntityLifecycleConfig<
  Namespace,
  NamespaceLifecycleContext,
  NamespaceDeleteVariables
> & {
  dangerZone: NonNullable<
    EntityLifecycleConfig<
      Namespace,
      NamespaceLifecycleContext,
      NamespaceDeleteVariables
    >["dangerZone"]
  >;
} = {
  entityName: "Namespace",
  deleteDialog: {
    title: (namespace) => `Are you sure you want to delete ${namespace.name}?`,
    description:
      "This will permanently delete the namespace. This action cannot be undone.",
    consequences: [
      "The namespace will be permanently deleted",
      "Projects and documents in this namespace will remain but will no longer be associated with the namespace",
    ],
  },
  dangerZone: {
    dataSection: "namespace-danger-zone",
    description: "Irreversible actions for this namespace",
    summary:
      "Deleting a namespace permanently removes it from this organization. This action cannot be undone.",
    consequences: [
      "The namespace will be permanently deleted",
      "Projects and documents will remain but will no longer be associated with the namespace",
      "You will be redirected to the organization details page",
    ],
    buttonLabel: "Delete Namespace",
  },
  canDelete: hasDeletePermission,
  deleteVariables: (namespace) => ({
    path: { id: namespace.id },
  }),
  queryKeys: (namespace, { organizationId }) => [
    v1OrganizationsNamespacesGetOptions({
      path: { id: organizationId },
    }).queryKey,
    v1NamespaceGetOptions({
      path: { id: namespace.id },
    }).queryKey,
    accessibleNamespacesQueryKey,
  ],
  navigateAfterDelete: (navigate, { organizationId }) =>
    navigate({
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    }),
};

type ProjectLifecycleEntity = Pick<Project | PartialProject, "id" | "name">;

export const projectLifecycleConfig: EntityLifecycleConfig<
  ProjectLifecycleEntity,
  ProjectLifecycleContext,
  ProjectDeleteVariables
> & {
  dangerZone: NonNullable<
    EntityLifecycleConfig<
      ProjectLifecycleEntity,
      ProjectLifecycleContext,
      ProjectDeleteVariables
    >["dangerZone"]
  >;
} = {
  entityName: "Project",
  deleteDialog: {
    title: (project) => `Are you sure you want to delete ${project.name}?`,
    description:
      "This will permanently delete the project. This action cannot be undone.",
    consequences: [
      "The project will be permanently deleted",
      "Documents and issues in this project will remain but will no longer be associated with the project",
    ],
  },
  dangerZone: {
    dataSection: "project-danger-zone",
    description: "Irreversible actions for this project",
    summary:
      "Deleting a project permanently removes it from this namespace. This action cannot be undone.",
    consequences: [
      "The project will be permanently deleted",
      "Documents and issues will remain but will no longer be associated with the project",
      "You will be redirected to the namespace details page",
    ],
    buttonLabel: "Delete Project",
  },
  canDelete: hasDeletePermission,
  deleteVariables: (project) => ({
    path: { id: project.id },
  }),
  queryKeys: (project, { namespaceId }) => [
    v1ProjectGetOptions({
      path: { id: project.id },
    }).queryKey,
    v1NamespaceGetOptions({
      path: { id: namespaceId },
    }).queryKey,
    v1NamespacesProjectsGetOptions({
      path: { id: namespaceId },
    }).queryKey,
    accessibleNamespacesQueryKey,
  ],
  navigateAfterDelete: (navigate, { organizationId, namespaceId }) =>
    navigate({
      to: "/settings/organizations/$organizationId/namespaces/$namespaceId",
      params: { organizationId, namespaceId },
    }),
};

export const roleLifecycleConfig: EntityLifecycleConfig<
  Role,
  RoleLifecycleContext,
  RoleDeleteVariables
> = {
  entityName: "Role",
  deleteDialog: {
    title: (role) => `Are you sure you want to delete ${role.name}?`,
    description:
      "This will permanently delete the role. This action cannot be undone.",
    consequences: [
      "The role will be permanently deleted",
      "All members assigned to this role will lose their role assignment",
      "Role permissions will be removed",
    ],
  },
  canDelete: hasDeletePermission,
  deleteVariables: (role, { organizationId }) => ({
    path: {
      id: organizationId,
      role_id: role.id,
    },
  }),
  queryKeys: (role, { organizationId }) => [
    v1OrganizationRolesGetOptions({
      path: { id: organizationId },
    }).queryKey,
    v1OrganizationRoleGetOptions({
      path: {
        id: organizationId,
        role_id: role.id,
      },
    }).queryKey,
  ],
};
