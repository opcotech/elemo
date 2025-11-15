import { Link } from "@tanstack/react-router";

interface ConditionalLinkProps {
  to: string;
  params?: Record<string, string>;
  condition: boolean;
  children: React.ReactNode;
}

export function ConditionalLink({
  to,
  params,
  condition,
  children,
}: ConditionalLinkProps) {
  return condition ? (
    <Link to={to} params={params} className="text-primary hover:underline">
      {children}
    </Link>
  ) : (
    <span className="text-muted-foreground">{children}</span>
  );
}
