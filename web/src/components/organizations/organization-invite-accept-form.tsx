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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { v1OrganizationMembersAcceptMutation } from "@/lib/api";
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

  const acceptInvitationMutation = useMutation({
    ...v1OrganizationMembersAcceptMutation({
      auth: () => undefined, // Public endpoint, no authentication required
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
        // This is expected for pending users - show password form without error toast
        setNeedsPassword(true);
        return; // Don't show error toast for this expected case
      }
      // Show error toast for actual errors
      showErrorToast("Failed to accept invitation", errorMessage);
    },
  });

  const onSubmit = (values: PasswordFormData) => {
    if (!organization || !token) {
      showErrorToast(
        "Invalid invitation link",
        "The invitation link is invalid or missing required parameters."
      );
      return;
    }

    acceptInvitationMutation.mutate({
      path: {
        id: organization,
      },
      body: {
        token,
        password: values.password,
      },
    });
  };

  const handleAcceptWithoutPassword = () => {
    if (!organization || !token) {
      showErrorToast(
        "Invalid invitation link",
        "The invitation link is invalid or missing required parameters."
      );
      return;
    }

    acceptInvitationMutation.mutate({
      path: {
        id: organization,
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
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              {acceptInvitationMutation.isError &&
                acceptInvitationMutation.error?.message &&
                !acceptInvitationMutation.error.message
                  .toLowerCase()
                  .includes("password") && (
                  <Alert variant="destructive">
                    <AlertDescription>
                      {acceptInvitationMutation.error.message}
                    </AlertDescription>
                  </Alert>
                )}

              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder="Enter your password"
                    className="pr-10 pl-10"
                    {...form.register("password")}
                    disabled={acceptInvitationMutation.isPending}
                    autoComplete="new-password"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowPassword(!showPassword)}
                    disabled={acceptInvitationMutation.isPending}
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
                {form.formState.errors.password && (
                  <p className="text-sm text-red-600">
                    {form.formState.errors.password.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm Password</Label>
                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="confirmPassword"
                    type={showConfirmPassword ? "text" : "password"}
                    placeholder="Confirm your password"
                    className="pr-10 pl-10"
                    {...form.register("confirmPassword")}
                    disabled={acceptInvitationMutation.isPending}
                    autoComplete="new-password"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    disabled={acceptInvitationMutation.isPending}
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

              <div className="pt-4">
                <Button
                  type="submit"
                  className="flex w-full items-center"
                  disabled={acceptInvitationMutation.isPending}
                  aria-label="Accept invitation"
                >
                  {acceptInvitationMutation.isPending ? (
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
          ) : (
            <div className="space-y-4">
              {acceptInvitationMutation.isError && !needsPassword && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {acceptInvitationMutation.error?.message}
                  </AlertDescription>
                </Alert>
              )}

              <div className="pt-4">
                <Button
                  type="button"
                  onClick={handleAcceptWithoutPassword}
                  className="flex w-full items-center"
                  disabled={acceptInvitationMutation.isPending}
                  aria-label="Accept invitation"
                >
                  {acceptInvitationMutation.isPending ? (
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
