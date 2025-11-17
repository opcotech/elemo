import { zodResolver } from "@hookform/resolvers/zod";
import { Trash2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
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
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import { formatResourceId } from "@/lib/utils";

const permissionFormSchema = z.object({
  resourceType: z.enum([
    ResourceType.Organization,
    ResourceType.Namespace,
    ResourceType.Document,
    ResourceType.Project,
    ResourceType.Role,
  ]),
  resourceId: z.string().min(1, "Resource ID is required"),
  kind: z.enum(["read", "write", "create", "delete", "*"]),
});

type PermissionFormValues = z.infer<typeof permissionFormSchema>;

const ORGANIZATION_SCOPED_RESOURCE_TYPES = [
  ResourceType.Organization,
  ResourceType.Namespace,
  ResourceType.Document,
  ResourceType.Project,
  ResourceType.Role,
] as const;

const PERMISSION_KINDS: ("read" | "write" | "create" | "delete" | "*")[] = [
  "read",
  "write",
  "create",
  "delete",
  "*",
];

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
      resourceType:
        ResourceType.Organization as PermissionFormValues["resourceType"],
      resourceId: "",
      kind: "read",
    },
  });

  const onSubmit = (values: PermissionFormValues) => {
    const target = withResourceType(values.resourceType, values.resourceId);

    const permission: PendingPermission = {
      resourceType: values.resourceType,
      resourceId: values.resourceId,
      kind: values.kind,
      target,
    };

    onAddPermission(permission);
    form.reset({
      resourceType:
        ResourceType.Organization as PermissionFormValues["resourceType"],
      resourceId: "",
      kind: "read",
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
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-4 rounded-md border p-4"
          >
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <FormField
                control={form.control}
                name="resourceType"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Resource Type</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select resource type" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {ORGANIZATION_SCOPED_RESOURCE_TYPES.map((type) => (
                          <SelectItem key={type} value={type}>
                            {type}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="resourceId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Resource ID</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter resource ID" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="kind"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Permission Kind</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select permission kind" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PERMISSION_KINDS.map((kind) => (
                          <SelectItem key={kind} value={kind}>
                            {kind}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="flex justify-end">
              <Button type="submit" size="sm">
                Add Permission
              </Button>
            </div>
          </form>
        </Form>

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
