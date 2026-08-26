import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { Eye, EyeOff, Lock, Users } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { Spinner } from "@/components/ui/spinner";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1OrganizationMembersAcceptMutation } from "@/lib/api/mutation-options";
import { v1OrganizationMembersAccept } from "@/lib/api/sdk";
import type { Options, V1OrganizationMembersAcceptData } from "@/lib/api/types";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const passwordSchema = z
  .object({
    password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .max(64, "Password must be less than 64 characters"),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type PasswordFormData = z.infer<typeof passwordSchema>;

export function OrganizationInviteAcceptForm() {
  const navigate = useNavigate();
  const { organization, token } = useSearch({ from: "/organizations/join" });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [needsPassword, setNeedsPassword] = useState<boolean | null>(null);

  const form = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  // Mutation for accepting with password (using use-form-mutation)
  const acceptWithPasswordMutation = useFormMutation<
    unknown,
    Options<V1OrganizationMembersAcceptData>,
    PasswordFormData
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationMembersAccept({
        path: variables.path,
        body: variables.body,
        auth: () => undefined,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Invitation accepted",
    successDescription:
      "You have successfully joined the organization. You can now log in.",
    errorMessagePrefix: "Failed to accept invitation",
    navigateOnSuccess: (navigateTo) =>
      navigateTo({ to: "/login", search: { redirect: undefined } }),
    transformValues: (values) => {
      if (!organization || !token) {
        throw new Error(
          "Invalid invitation link - missing required parameters"
        );
      }
      return {
        path: {
          organizationRef: organization,
        },
        body: {
          token,
          password: values.password,
        },
      };
    },
    onError: (error) => {
      const errorMessage = error.message || "Failed to accept invitation";

      // Check if error indicates password is required for pending users
      if (
        errorMessage.toLowerCase().includes("password") &&
        errorMessage.toLowerCase().includes("required")
      ) {
        setNeedsPassword(true);
      }
    },
  });

  // Mutation for accepting without password (needs to stay separate)
  const acceptWithoutPasswordMutation = useMutation({
    ...v1OrganizationMembersAcceptMutation({
      auth: () => undefined,
    }),
    onSuccess: () => {
      showSuccessToast(
        "Invitation accepted",
        "You have successfully joined the organization. You can now log in."
      );
      navigate({ to: "/login", search: { redirect: undefined } });
    },
    onError: (error) => {
      const errorMessage = error.message || "Failed to accept invitation";

      // Check if error indicates password is required for pending users
      if (
        errorMessage.toLowerCase().includes("password") &&
        errorMessage.toLowerCase().includes("required")
      ) {
        setNeedsPassword(true);
        return;
      }
      showErrorToast("Failed to accept invitation", errorMessage);
    },
  });

  const handleAcceptWithoutPassword = () => {
    if (!organization || !token) {
      showErrorToast(
        "Invalid invitation link",
        "The invitation link is invalid or missing required parameters."
      );
      return;
    }

    acceptWithoutPasswordMutation.mutate({
      path: {
        organizationRef: organization,
      },
      body: {
        token,
      },
    });
  };

  if (!organization || !token) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1">
            <CardTitle className="text-center text-2xl font-bold">
              Invalid Invitation Link
            </CardTitle>
            <CardDescription className="text-center">
              The invitation link is invalid or missing required parameters.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Alert variant="destructive">
                <AlertDescription>
                  Please use the invitation link from your email.
                </AlertDescription>
              </Alert>
              <div className="flex justify-center">
                <Link
                  to="/login"
                  search={{ redirect: undefined }}
                  className="text-primary hover:text-primary/80 text-sm hover:underline"
                >
                  Go to login
                </Link>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // If we get an error that password is required, show password form
  // Otherwise, try accepting without password first
  const showPasswordForm = needsPassword === true;

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <div className="mb-2 flex items-center justify-center">
            <Users className="text-primary h-8 w-8" />
          </div>
          <CardTitle className="text-center text-2xl font-bold">
            {showPasswordForm ? "Set Your Password" : "Accept Invitation"}
          </CardTitle>
          <CardDescription className="text-center">
            {showPasswordForm
              ? "Create a password to complete your account setup"
              : "You've been invited to join an organization"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {showPasswordForm ? (
            <FieldProvider {...form}>
              <form
                onSubmit={acceptWithPasswordMutation.handleSubmit}
                className="space-y-4"
              >
                {acceptWithPasswordMutation.isError &&
                  acceptWithPasswordMutation.error?.message &&
                  !acceptWithPasswordMutation.error.message
                    .toLowerCase()
                    .includes("password") && (
                    <Alert variant="destructive">
                      <AlertDescription>
                        {acceptWithPasswordMutation.error.message}
                      </AlertDescription>
                    </Alert>
                  )}

                <FieldGroup>
                  <ControlledField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                      <Field>
                        <FieldLabel>Password</FieldLabel>
                        <div className="relative">
                          <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                          <FieldControl>
                            <Input
                              type={showPassword ? "text" : "password"}
                              placeholder="Enter your password"
                              className="pr-10 pl-10"
                              {...field}
                              disabled={acceptWithPasswordMutation.isPending}
                              autoComplete="new-password"
                            />
                          </FieldControl>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                            onClick={() => setShowPassword(!showPassword)}
                            disabled={acceptWithPasswordMutation.isPending}
                            aria-label={
                              showPassword ? "Hide password" : "Show password"
                            }
                          >
                            {showPassword ? (
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
                        <FieldLabel>Confirm Password</FieldLabel>
                        <div className="relative">
                          <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                          <FieldControl>
                            <Input
                              type={showConfirmPassword ? "text" : "password"}
                              placeholder="Confirm your password"
                              className="pr-10 pl-10"
                              {...field}
                              disabled={acceptWithPasswordMutation.isPending}
                              autoComplete="new-password"
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
                            disabled={acceptWithPasswordMutation.isPending}
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

                <div className="pt-4">
                  <Button
                    type="submit"
                    className="flex w-full items-center"
                    disabled={acceptWithPasswordMutation.isPending}
                    aria-label="Accept invitation"
                  >
                    {acceptWithPasswordMutation.isPending ? (
                      <>
                        <Spinner size="xs" className="mr-2" />
                        <span>Accepting invitation...</span>
                      </>
                    ) : (
                      "Accept Invitation"
                    )}
                  </Button>
                </div>
              </form>
            </FieldProvider>
          ) : (
            <div className="space-y-4">
              {acceptWithoutPasswordMutation.isError && !needsPassword && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {acceptWithoutPasswordMutation.error?.message}
                  </AlertDescription>
                </Alert>
              )}

              <div className="pt-4">
                <Button
                  type="button"
                  onClick={handleAcceptWithoutPassword}
                  className="flex w-full items-center"
                  disabled={acceptWithoutPasswordMutation.isPending}
                  aria-label="Accept invitation"
                >
                  {acceptWithoutPasswordMutation.isPending ? (
                    <>
                      <Spinner size="xs" className="mr-2" />
                      <span>Accepting invitation...</span>
                    </>
                  ) : (
                    "Accept Invitation"
                  )}
                </Button>
              </div>
            </div>
          )}

          <div className="flex justify-center pt-4">
            <Link
              to="/login"
              search={{ redirect: undefined }}
              className="text-muted-foreground hover:text-primary text-sm hover:underline"
            >
              Back to login
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
