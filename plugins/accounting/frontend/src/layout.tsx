import type { ElemoPluginAPI } from "@elemo/plugin-sdk";
import { Button, PageHeader, cn } from "@elemo/plugin-ui";
import type { ReactNode } from "react";

export type AccountingSection = "accounts" | "budgets" | "report";

export function pluginBasePath(
  organizationSlug: string,
  pluginId: string,
): string {
  return `/organizations/${organizationSlug}/plugins/${pluginId}`;
}

export function AccountingPage({
  elemo,
  organizationSlug,
  title,
  description,
  current,
  actions,
  children,
  testId,
}: {
  elemo: ElemoPluginAPI;
  organizationSlug: string;
  title: string;
  description: string;
  current: AccountingSection;
  actions?: ReactNode;
  children: ReactNode;
  testId: string;
}) {
  const base = pluginBasePath(organizationSlug, elemo.pluginId);
  const links: Array<{ id: AccountingSection; href: string; label: string }> = [
    { id: "accounts", href: `${base}/accounts`, label: "Accounts" },
    { id: "budgets", href: `${base}/budgets`, label: "Budgets" },
    { id: "report", href: `${base}/report`, label: "Report" },
  ];

  return (
    <div className="space-y-6" data-testid={testId}>
      <PageHeader
        eyebrow="Accounting"
        title={title}
        description={description}
        actions={actions}
      />
      <nav
        aria-label="Accounting"
        className="flex flex-wrap gap-1"
        data-testid="accounting-nav"
      >
        {links.map((link) => (
          <Button
            key={link.id}
            size="sm"
            variant={link.id === current ? "secondary" : "ghost"}
            nativeButton={false}
            render={
              <a
                href={link.href}
                aria-current={link.id === current ? "page" : undefined}
              />
            }
            className={cn(link.id === current && "pointer-events-none")}
          >
            {link.label}
          </Button>
        ))}
      </nav>
      {children}
    </div>
  );
}
