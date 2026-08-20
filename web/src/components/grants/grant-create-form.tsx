import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";

import { ActionMultiSelect } from "@/components/roles/action-multi-select";
import { EntitySelect } from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationTeamsGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/api/query-options";
import {
  v1OrganizationMembersGet,
  v1OrganizationRolesGet,
  v1OrganizationTeamsGet,
  v1PermissionsCreate,
} from "@/lib/api/sdk";
import type {
  ResourceType as ApiResourceType,
  GrantCreate,
  GrantPrincipalType,
  Options,
  OrganizationMember,
  Role,
  Team,
  V1PermissionsCreateData,
} from "@/lib/api/types";
import { Action, ResourceType, can } from "@/lib/auth/permissions";
import { showSuccessToast } from "@/lib/toast";
import { getInitials } from "@/lib/utils";

const grantFormSchema = z
  .object({
    principalType: z.enum(["User", "Team", "Organization"]),
    principalId: z.string().min(1, "Principal is required"),
    roleId: z.string().optional(),
    actions: z.array(z.string()).optional(),
    scopeType: z.string().min(1),
    scopeId: z.string().min(1, "Scope is required"),
  })
  .refine(
    (value) =>
      Boolean(value.roleId) || (value.actions && value.actions.length > 0),
    {
      message: "Select a role or at least one action",
      path: ["roleId"],
    }
  );

type GrantFormValues = z.infer<typeof grantFormSchema>;

const principalTypeItems: Record<GrantPrincipalType, string> = {
  User: "User",
  Team: "Team",
  Organization: "Organization",
};

const scopeTypeItems: Partial<Record<ApiResourceType, string>> = {
  Organization: "Organization",
  Namespace: "Namespace",
  Project: "Project",
  Document: "Document",
  Folder: "Folder",
  Issue: "Issue",
  Role: "Role",
  Team: "Team",
};

interface GrantCreateFormProps {
  organizationId: string;
}

export function GrantCreateForm({ organizationId }: GrantCreateFormProps) {
  const { data: permissions } = useQuery(
    v1PermissionResourceGetOptions({
      path: {
        resourceId: `${ResourceType.Organization}:${organizationId}`,
      },
    })
  );
  const canManageGrants = can(permissions, Action.PermissionManage);

  const form = useForm<GrantFormValues>({
    resolver: zodResolver(grantFormSchema),
    defaultValues: {
      principalType: "User",
      principalId: "",
      roleId: "",
      actions: [],
      scopeType: "Organization",
      scopeId: organizationId,
    },
  });

  const principalType = useWatch({
    control: form.control,
    name: "principalType",
  });

  const membersOptions = v1OrganizationMembersGetOptions({
    path: { id: organizationId },
  });
  const { data: membersPage } = useQuery({
    ...collectedListQuery<OrganizationMember>(
      membersOptions,
      async (pageToken, signal) => {
        const { data } = await v1OrganizationMembersGet({
          path: { id: organizationId },
          query: cursorPageQuery(pageToken),
          signal,
          throwOnError: true,
        });
        return data;
      }
    ),
    enabled: canManageGrants && principalType === "User",
  });

  const teamsOptions = v1OrganizationTeamsGetOptions({
    path: { id: organizationId },
  });
  const { data: teamsPage } = useQuery({
    ...collectedListQuery<Team>(teamsOptions, async (pageToken, signal) => {
      const { data } = await v1OrganizationTeamsGet({
        path: { id: organizationId },
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
    enabled: canManageGrants && principalType === "Team",
  });

  const rolesOptions = v1OrganizationRolesGetOptions({
    path: { id: organizationId },
  });
  const { data: rolesPage } = useQuery({
    ...collectedListQuery<Role>(rolesOptions, async (pageToken, signal) => {
      const { data } = await v1OrganizationRolesGet({
        path: { id: organizationId },
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
    enabled: canManageGrants,
  });

  const mutation = useFormMutation<
    { id: string },
    Options<V1PermissionsCreateData>,
    GrantFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1PermissionsCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Grant created",
    errorMessagePrefix: "Failed to create grant",
    resetFormOnSuccess: true,
    transformValues: (values) => {
      const body: GrantCreate = {
        principal: {
          resourceType: values.principalType,
          id: values.principalId,
        },
        scope: {
          resourceType: values.scopeType as ApiResourceType,
          id: values.scopeId,
        },
      };
      if (values.roleId) {
        body.role_id = values.roleId;
      }
      if (values.actions && values.actions.length > 0) {
        body.actions = values.actions;
      }
      return { body };
    },
    onSuccess: () => {
      showSuccessToast(
        "Grant created",
        "The principal can now perform the selected actions on this scope"
      );
      form.reset({
        principalType: "User",
        principalId: "",
        roleId: "",
        actions: [],
        scopeType: "Organization",
        scopeId: organizationId,
      });
    },
  });

  if (!canManageGrants) {
    return null;
  }

  const memberOptions = (membersPage?.items ?? []).map((member) => ({
    value: member.id,
    title: `${member.first_name} ${member.last_name}`,
    description: member.email,
    avatarSrc: member.picture || null,
    avatarFallback: getInitials(member.first_name, member.last_name),
  }));

  const teamOptions = (teamsPage?.items ?? []).map((team) => ({
    value: team.id,
    title: team.name,
    description: team.description || undefined,
  }));

  const roleItems = Object.fromEntries(
    (rolesPage?.items ?? []).map((role) => [
      role.id,
      `${role.name} (${role.key})`,
    ])
  );

  return (
    <FormCard
      data-section="grant-create-form"
      title="Create grant"
      description="Bind a user, team, or organization principal to a role or actions on a scope. Scope defaults to this organization."
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Create Grant"
    >
      <FieldProvider {...form}>
        <FieldGroup>
          <ControlledField
            control={form.control}
            name="principalType"
            render={({ field }) => (
              <Field>
                <FieldLabel>Principal type</FieldLabel>
                <FieldControl>
                  <Select
                    value={field.value}
                    onValueChange={(value) => {
                      field.onChange(value);
                      form.setValue("principalId", "");
                    }}
                    items={principalTypeItems}
                    disabled={mutation.isPending}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select principal type" />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(principalTypeItems).map(
                        ([value, label]) => (
                          <SelectItem key={value} value={value}>
                            {label}
                          </SelectItem>
                        )
                      )}
                    </SelectContent>
                  </Select>
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          {principalType === "User" ? (
            <ControlledField
              control={form.control}
              name="principalId"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Principal</FieldLabel>
                  <FieldControl>
                    <EntitySelect
                      options={memberOptions}
                      value={field.value}
                      onValueChange={field.onChange}
                      placeholder="Choose a user"
                      disabled={mutation.isPending}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          ) : principalType === "Team" ? (
            <ControlledField
              control={form.control}
              name="principalId"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Principal</FieldLabel>
                  <FieldControl>
                    <EntitySelect
                      options={teamOptions}
                      value={field.value}
                      onValueChange={field.onChange}
                      placeholder="Choose a team"
                      disabled={mutation.isPending}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          ) : (
            <ControlledField
              control={form.control}
              name="principalId"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Organization ID</FieldLabel>
                  <FieldDescription>
                    Defaults to this organization. Use another organization ID
                    for cross-org grants.
                  </FieldDescription>
                  <FieldControl>
                    <Input
                      placeholder={organizationId}
                      {...field}
                      disabled={mutation.isPending}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          )}

          <ControlledField
            control={form.control}
            name="scopeType"
            render={({ field }) => (
              <Field>
                <FieldLabel>Scope type</FieldLabel>
                <FieldControl>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    items={scopeTypeItems}
                    disabled={mutation.isPending}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select scope type" />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(scopeTypeItems).map(([value, label]) => (
                        <SelectItem key={value} value={value}>
                          {label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ControlledField
            control={form.control}
            name="scopeId"
            render={({ field }) => (
              <Field>
                <FieldLabel>Scope ID</FieldLabel>
                <FieldDescription>
                  Resource the grant applies to. Defaults to this organization.
                </FieldDescription>
                <FieldControl>
                  <Input {...field} disabled={mutation.isPending} />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ControlledField
            control={form.control}
            name="roleId"
            render={({ field }) => (
              <Field>
                <FieldLabel>Role</FieldLabel>
                <FieldDescription>
                  Optional. The role's bundled actions are unioned with explicit
                  actions.
                </FieldDescription>
                <FieldControl>
                  <Select
                    value={field.value || "none"}
                    onValueChange={(value) =>
                      field.onChange(value === "none" ? "" : value)
                    }
                    items={{ none: "None", ...roleItems }}
                    disabled={mutation.isPending}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Optional role bundle" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">None</SelectItem>
                      {Object.entries(roleItems).map(([value, label]) => (
                        <SelectItem key={value} value={value}>
                          {label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ActionMultiSelect
            control={form.control}
            isPending={mutation.isPending}
            description="Optional explicit actions granted on the scope."
          />
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
