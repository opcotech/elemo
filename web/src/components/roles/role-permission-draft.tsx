import { zodResolver } from "@hookform/resolvers/zod";
import { Trash2 } from "lucide-react";
import { useForm } from "react-hook-form";

import {
  organizationScopedResourceType,
  permissionFormSchema,
} from "./permission-form-schema";
import type { PermissionFormValues } from "./permission-form-schema";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ResourceType } from "@/hooks/use-permissions";
import { withResourceType } from "@/hooks/use-permissions";
import { zPermissionKind } from "@/lib/client/zod.gen";
import { formatResourceId } from "@/lib/utils";

export interface PendingPermission {
  resourceType: string;
  resourceId: string;
  kind: string;
  target: string;
}

interface RolePermissionDraftProps {
  permissions: PendingPermission[];
  onAddPermission: (permission: PendingPermission) => void;
  onRemovePermission: (index: number) => void;
}

export function RolePermissionDraft({
  permissions,
  onAddPermission,
  onRemovePermission,
}: RolePermissionDraftProps) {
  const form = useForm<PermissionFormValues>({
    resolver: zodResolver(permissionFormSchema),
    defaultValues: {
      resourceType: organizationScopedResourceType.enum.Organization,
      resourceId: "",
      kind: zPermissionKind.enum.read,
    },
  });

  const onSubmit = (values: PermissionFormValues) => {
    const target = withResourceType(
      values.resourceType as ResourceType,
      values.resourceId
    );

    const permission: PendingPermission = {
      resourceType: values.resourceType,
      resourceId: values.resourceId,
      kind: values.kind,
      target,
    };

    onAddPermission(permission);
    form.reset({
      resourceType: organizationScopedResourceType.enum.Organization,
      resourceId: "",
      kind: zPermissionKind.enum.read,
    });
  };

  return (
    <Card data-section="role-permission-draft">
      <CardHeader>
        <CardTitle>Permissions</CardTitle>
        <CardDescription>
          Add permissions to assign to this role. Only organization-scoped
          resources can be assigned.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <FieldProvider {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-4 rounded-md border p-4"
          >
            <FieldGroup className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <ControlledField
                control={form.control}
                name="resourceType"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Resource Type</FieldLabel>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      items={Object.fromEntries(
                        organizationScopedResourceType.options.map((type) => [
                          type,
                          type,
                        ])
                      )}
                    >
                      <FieldControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select resource type" />
                        </SelectTrigger>
                      </FieldControl>
                      <SelectContent>
                        {organizationScopedResourceType.options.map((type) => (
                          <SelectItem key={type} value={type}>
                            {type}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FieldError />
                  </Field>
                )}
              />

              <ControlledField
                control={form.control}
                name="resourceId"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Resource ID</FieldLabel>
                    <FieldControl>
                      <Input placeholder="Enter resource ID" {...field} />
                    </FieldControl>
                    <FieldError />
                  </Field>
                )}
              />

              <ControlledField
                control={form.control}
                name="kind"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Permission Kind</FieldLabel>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      items={Object.fromEntries(
                        zPermissionKind.options.map((kind) => [kind, kind])
                      )}
                    >
                      <FieldControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select permission kind" />
                        </SelectTrigger>
                      </FieldControl>
                      <SelectContent>
                        {zPermissionKind.options.map((kind) => (
                          <SelectItem key={kind} value={kind}>
                            {kind}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FieldError />
                  </Field>
                )}
              />
            </FieldGroup>

            <div className="flex justify-end">
              <Button type="submit" size="sm">
                Add Permission
              </Button>
            </div>
          </form>
        </FieldProvider>

        {permissions.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource Type</TableHead>
                <TableHead>Resource ID</TableHead>
                <TableHead>Permission Kind</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {permissions.map((permission, index) => (
                <TableRow
                  key={`${permission.target}-${permission.kind}-${index}`}
                >
                  <TableCell className="font-medium">
                    {permission.resourceType}
                  </TableCell>
                  <TableCell>
                    {formatResourceId(permission.resourceId)}
                  </TableCell>
                  <TableCell>{permission.kind}</TableCell>
                  <TableCell className="text-right">
                    <Button
                      type="button"
                      variant="destructive-ghost"
                      size="sm"
                      onClick={() => onRemovePermission(index)}
                    >
                      <Trash2 className="h-4 w-4" />
                      <span className="sr-only">Remove permission</span>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
