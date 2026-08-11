import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/dashboard")({
  beforeLoad: () => {
    throw redirect({
      to: "/",
      replace: true,
    });
  },
  component: () => null,
});
