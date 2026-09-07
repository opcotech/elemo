import { Link } from "@tanstack/react-router";
import { Home, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
            <Search className="h-8 w-8 text-muted-foreground" />
          </div>
          <CardTitle className="font-bold text-2xl">Page Not Found</CardTitle>
          <CardDescription>
            Sorry, we couldn't find the page you're looking for. It might have
            been moved, deleted, or you entered the wrong URL.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2 text-muted-foreground text-sm">
            <p>Here are some things you can try:</p>
            <ul className="list-disc space-y-1 pl-4">
              <li>Check the URL for typos</li>
              <li>Go back to the previous page</li>
              <li>Return to Home</li>
            </ul>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Button render={<Link to="/" />} className="flex-1">
              <Home className="size-4" />
              Go to Home
            </Button>
            <Button variant="outline" onClick={() => window.history.back()}>
              Go Back
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
