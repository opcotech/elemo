import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Eye, EyeOff, Lock } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/hooks/use-auth";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { v1UserUpdateMutation } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export const Route = createFileRoute("/settings/security")({
  beforeLoad: requireAuthBeforeLoad,
  component: SecuritySettings,
});

const passwordChangeSchema = z
  .object({
    currentPassword: z.string().min(8, "Current password is required"),
    newPassword: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .max(64, "Password must be less than 64 characters"),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type PasswordChangeFormData = z.infer<typeof passwordChangeSchema>;

function SecuritySettings() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { user } = useAuth();
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const form = useForm<PasswordChangeFormData>({
    resolver: zodResolver(passwordChangeSchema),
  });

  const updatePasswordMutation = useMutation({
    ...v1UserUpdateMutation(),
    onSuccess: () => {
      showSuccessToast(
        "Password updated successfully",
        "Your password has been changed successfully"
      );
    },
    onError: (error) => {
      showErrorToast("Failed to update password", error.message);
    },
  });

  const onSubmit = (values: PasswordChangeFormData) => {
    updatePasswordMutation.mutate(
      {
        path: { id: user!.id },
        body: {
          password: values.currentPassword,
          new_password: values.newPassword,
        },
      },
      {
        onSuccess: () => {
          form.reset();
        },
      }
    );
  };

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Password & Authentication",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Password & Authentication</h1>
        <p className="mt-2 text-gray-600">
          Manage your password and authentication settings.
        </p>
      </div>

      <form onSubmit={form.handleSubmit(onSubmit)}>
        <Card>
          <CardHeader>
            <CardTitle>Change Password</CardTitle>
          </CardHeader>
          <CardContent>
            {updatePasswordMutation.isError && (
              <Alert variant="destructive" className="mb-6">
                <AlertDescription>
                  {updatePasswordMutation.error.message}
                </AlertDescription>
              </Alert>
            )}

            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="currentPassword">Current Password</Label>
                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="currentPassword"
                    type={showCurrentPassword ? "text" : "password"}
                    placeholder="Enter your current password"
                    className="pr-10 pl-10"
                    {...form.register("currentPassword")}
                    disabled={updatePasswordMutation.isPending}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                    disabled={updatePasswordMutation.isPending}
                    aria-label={
                      showCurrentPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showCurrentPassword ? (
                      <EyeOff className="text-muted-foreground h-4 w-4" />
                    ) : (
                      <Eye className="text-muted-foreground h-4 w-4" />
                    )}
                  </Button>
                </div>
                {form.formState.errors.currentPassword && (
                  <p className="text-sm text-red-600">
                    {form.formState.errors.currentPassword.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="newPassword">New Password</Label>
                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="newPassword"
                    type={showNewPassword ? "text" : "password"}
                    placeholder="Enter your new password"
                    className="pr-10 pl-10"
                    {...form.register("newPassword")}
                    disabled={updatePasswordMutation.isPending}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowNewPassword(!showNewPassword)}
                    disabled={updatePasswordMutation.isPending}
                    aria-label={
                      showNewPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showNewPassword ? (
                      <EyeOff className="text-muted-foreground h-4 w-4" />
                    ) : (
                      <Eye className="text-muted-foreground h-4 w-4" />
                    )}
                  </Button>
                </div>
                {form.formState.errors.newPassword && (
                  <p className="text-sm text-red-600">
                    {form.formState.errors.newPassword.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm New Password</Label>
                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="confirmPassword"
                    type={showConfirmPassword ? "text" : "password"}
                    placeholder="Confirm your new password"
                    className="pr-10 pl-10"
                    {...form.register("confirmPassword")}
                    disabled={updatePasswordMutation.isPending}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    disabled={updatePasswordMutation.isPending}
                    aria-label={
                      showConfirmPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showConfirmPassword ? (
                      <EyeOff className="text-muted-foreground h-4 w-4" />
                    ) : (
                      <Eye className="text-muted-foreground h-4 w-4" />
                    )}
                  </Button>
                </div>
                {form.formState.errors.confirmPassword && (
                  <p className="text-sm text-red-600">
                    {form.formState.errors.confirmPassword.message}
                  </p>
                )}
              </div>
            </div>
          </CardContent>

          <CardFooter className="flex justify-end">
            <Button type="submit" disabled={updatePasswordMutation.isPending}>
              {updatePasswordMutation.isPending ? (
                <>
                  <Spinner size="xs" className="mr-2" />
                  <span>Updating Password...</span>
                </>
              ) : (
                "Update Password"
              )}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
}
