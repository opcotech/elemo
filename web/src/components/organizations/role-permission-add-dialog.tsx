import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { Spinner } from "@/components/ui/spinner";
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import {
  v1OrganizationRolePermissionAddMutation,
  v1OrganizationRolePermissionsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

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

interface RolePermissionAddDialogProps {
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function RolePermissionAddDialog({
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RolePermissionAddDialogProps) {
  const queryClient = useQueryClient();

  const form = useForm<PermissionFormValues>({
    resolver: zodResolver(permissionFormSchema),
    defaultValues: {
      resourceType:
        ResourceType.Organization as PermissionFormValues["resourceType"],
      resourceId: "",
      kind: "read",
    },
  });

  const mutation = useMutation(v1OrganizationRolePermissionAddMutation());

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      form.reset();
    }
    onOpenChange(newOpen);
  };

  const onSubmit = (values: PermissionFormValues) => {
    // Validate that resource type is organization-scoped
    if (!ORGANIZATION_SCOPED_RESOURCE_TYPES.includes(values.resourceType)) {
      showErrorToast(
        "Invalid resource type",
        "Only organization-scoped resources can be assigned to roles"
      );
      return;
    }

    const target = withResourceType(values.resourceType, values.resourceId);

    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: roleId,
        },
        body: {
          target,
          kind: values.kind,
        },
      },
      {
        onSuccess: () => {
          showSuccessToast("Permission added", "Permission added successfully");
          form.reset();
          onOpenChange(false);

          // Invalidate queries to refresh the permissions list
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolePermissionsGetOptions({
              path: {
                id: organizationId,
                role_id: roleId,
              },
            }).queryKey,
          });

          onSuccess?.();
        },
        onError: (error) => {
          showErrorToast("Failed to add permission", error.message);
        },
      }
    );
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            <DialogHeader>
              <DialogTitle>Add Permission</DialogTitle>
            </DialogHeader>

            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Failed to add permission</AlertTitle>
                <AlertDescription>{mutation.error.message}</AlertDescription>
              </Alert>
            )}

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

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={mutation.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? (
                  <>
                    <Spinner size="xs" className="mr-0.5 text-white" />
                    <span>Adding...</span>
                  </>
                ) : (
                  "Add Permission"
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
