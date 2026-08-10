import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { Eye, EyeOff, Lock } from "lucide-react";
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
import { v1UserResetPasswordMutation } from "@/lib/api/mutation-options";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const passwordResetSchema = z
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

type PasswordResetFormData = z.infer<typeof passwordResetSchema>;

export function PasswordResetForm() {
  const navigate = useNavigate();
  const { token } = useSearch({ from: "/reset-password" });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const form = useForm<PasswordResetFormData>({
    resolver: zodResolver(passwordResetSchema),
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  const resetPasswordMutation = useMutation({
    ...v1UserResetPasswordMutation(),
    onSuccess: () => {
      showSuccessToast(
        "Password reset successfully",
        "Your password has been reset successfully. You can now log in with your new password."
      );
      navigate({ to: "/login", search: { redirect: undefined } });
    },
    onError: (error) => {
      showErrorToast("Failed to reset password", error.message);
    },
  });

  const onSubmit = (values: PasswordResetFormData) => {
    if (!token) {
      showErrorToast(
        "Invalid reset link",
        "The password reset link is invalid or missing."
      );
      return;
    }

    resetPasswordMutation.mutate({
      body: {
        token,
        password: values.password,
      },
    });
  };

  if (!token) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1">
            <CardTitle className="text-center text-2xl font-bold">
              Invalid Reset Link
            </CardTitle>
            <CardDescription className="text-center">
              The password reset link is invalid or missing.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Alert variant="destructive">
                <AlertDescription>
                  Please request a new password reset link.
                </AlertDescription>
              </Alert>
              <div className="flex justify-center">
                <Link
                  to="/forgot-password"
                  search={{ redirect: undefined }}
                  className="text-primary hover:text-primary/80 text-sm hover:underline"
                >
                  Request new reset link
                </Link>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-center text-2xl font-bold">
            Reset your password
          </CardTitle>
          <CardDescription className="text-center">
            Enter your new password below
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FieldProvider {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              {resetPasswordMutation.isError && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {resetPasswordMutation.error.message}
                  </AlertDescription>
                </Alert>
              )}

              <FieldGroup>
                <ControlledField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <Field>
                      <FieldLabel>New Password</FieldLabel>
                      <div className="relative">
                        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                        <FieldControl>
                          <Input
                            type={showPassword ? "text" : "password"}
                            placeholder="Enter your new password"
                            className="pr-10 pl-10"
                            {...field}
                            disabled={resetPasswordMutation.isPending}
                            autoComplete="new-password"
                          />
                        </FieldControl>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                          onClick={() => setShowPassword(!showPassword)}
                          disabled={resetPasswordMutation.isPending}
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
                      <FieldLabel>Confirm New Password</FieldLabel>
                      <div className="relative">
                        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                        <FieldControl>
                          <Input
                            type={showConfirmPassword ? "text" : "password"}
                            placeholder="Confirm your new password"
                            className="pr-10 pl-10"
                            {...field}
                            disabled={resetPasswordMutation.isPending}
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
                          disabled={resetPasswordMutation.isPending}
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
                  disabled={resetPasswordMutation.isPending}
                  aria-label="Reset password"
                >
                  {resetPasswordMutation.isPending ? (
                    <>
                      <Spinner size="xs" className="mr-2" />
                      <span>Resetting password...</span>
                    </>
                  ) : (
                    "Reset password"
                  )}
                </Button>
              </div>

              <div className="flex justify-center">
                <Link
                  to="/login"
                  search={{ redirect: undefined }}
                  className="text-muted-foreground hover:text-primary text-sm hover:underline"
                >
                  Back to login
                </Link>
              </div>
            </form>
          </FieldProvider>
        </CardContent>
      </Card>
    </div>
  );
}
