import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Eye, EyeOff, Lock } from "lucide-react";
import { useState } from "react";
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
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/hooks/use-auth";
import { v1UserUpdateMutation } from "@/lib/api/mutation-options";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export const Route = createFileRoute("/_authenticated/settings/security")({
  staticData: {
    breadcrumb: "Password & authentication",
  },
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
  const { user } = useAuth();
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const form = useForm<PasswordChangeFormData>({
    resolver: zodResolver(passwordChangeSchema),
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
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

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Password & Authentication</h1>
        <p className="text-muted-foreground mt-2">
          Manage your password and authentication settings.
        </p>
      </div>

      <FieldProvider {...form}>
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

              <FieldGroup>
                <ControlledField
                  control={form.control}
                  name="currentPassword"
                  render={({ field }) => (
                    <Field>
                      <FieldLabel>Current Password</FieldLabel>
                      <div className="relative">
                        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                        <FieldControl>
                          <Input
                            type={showCurrentPassword ? "text" : "password"}
                            placeholder="Enter your current password"
                            className="pr-10 pl-10"
                            {...field}
                            disabled={updatePasswordMutation.isPending}
                          />
                        </FieldControl>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                          onClick={() =>
                            setShowCurrentPassword(!showCurrentPassword)
                          }
                          disabled={updatePasswordMutation.isPending}
                          aria-label={
                            showCurrentPassword
                              ? "Hide password"
                              : "Show password"
                          }
                        >
                          {showCurrentPassword ? (
                            <EyeOff className="text-muted-foreground h-4 w-4" />
                          ) : (
                            <Eye className="text-muted-foreground h-4 w-4" />
                          )}
                        </Button>
                      </div>
                      <FieldError />
                    </Field>
                  )}
                />

                <ControlledField
                  control={form.control}
                  name="newPassword"
                  render={({ field }) => (
                    <Field>
                      <FieldLabel>New Password</FieldLabel>
                      <div className="relative">
                        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                        <FieldControl>
                          <Input
                            type={showNewPassword ? "text" : "password"}
                            placeholder="Enter your new password"
                            className="pr-10 pl-10"
                            {...field}
                            disabled={updatePasswordMutation.isPending}
                          />
                        </FieldControl>
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
                      <FieldError />
                    </Field>
                  )}
                />

                <ControlledField
                  control={form.control}
                  name="confirmPassword"
                  render={({ field }) => (
                    <Field>
                      <FieldLabel>Confirm New Password</FieldLabel>
                      <div className="relative">
                        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                        <FieldControl>
                          <Input
                            type={showConfirmPassword ? "text" : "password"}
                            placeholder="Confirm your new password"
                            className="pr-10 pl-10"
                            {...field}
                            disabled={updatePasswordMutation.isPending}
                          />
                        </FieldControl>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                          onClick={() =>
                            setShowConfirmPassword(!showConfirmPassword)
                          }
                          disabled={updatePasswordMutation.isPending}
                          aria-label={
                            showConfirmPassword
                              ? "Hide password"
                              : "Show password"
                          }
                        >
                          {showConfirmPassword ? (
                            <EyeOff className="text-muted-foreground h-4 w-4" />
                          ) : (
                            <Eye className="text-muted-foreground h-4 w-4" />
                          )}
                        </Button>
                      </div>
                      <FieldError />
                    </Field>
                  )}
                />
              </FieldGroup>
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
      </FieldProvider>
    </div>
  );
}
