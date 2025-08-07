import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Mail } from "lucide-react";
import { useState } from "react";

import { Spinner } from "../ui/spinner";

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
import { v1UserRequestPasswordResetOptions } from "@/lib/api";

export function PasswordResetRequestForm() {
  const [email, setEmail] = useState("");

  const {
    isLoading,
    isPending,
    error,
    refetch: requestPasswordReset,
  } = useQuery({
    enabled: false,
    ...v1UserRequestPasswordResetOptions({
      query: {
        email,
      },
    }),
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      return;
    }

    await requestPasswordReset();
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
  };

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-center text-2xl font-bold">
            Forgot your password?
          </CardTitle>
          <CardDescription className="text-center">
            Enter your email to reset your password
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error.message}</AlertDescription>
              </Alert>
            )}

            {!isPending && !error && (
              <Alert variant="success">
                <AlertDescription>
                  If an account with this email exists, you will receive an
                  email with a link to reset your password.
                </AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <div className="relative">
                <Mail className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter your email"
                  value={email}
                  onChange={handleEmailChange}
                  className="pl-10"
                  required
                  disabled={isLoading}
                  autoComplete="email"
                />
              </div>
            </div>

            <div className="pt-4">
              <Button
                type="submit"
                className="flex w-full items-center"
                disabled={isLoading || !email}
                aria-label="Reset password"
              >
                {isLoading ? (
                  <>
                    <Spinner size="xs" className="mr-0.5 text-white" />
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
        </CardContent>
      </Card>
    </div>
  );
}
