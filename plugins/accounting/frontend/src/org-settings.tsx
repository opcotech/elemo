import type { ElemoPluginAPI } from "@elemo/plugin-sdk";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@elemo/plugin-ui";

import { pluginBasePath } from "./layout";

export function OrganizationStatus(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationSlug = String(props.organizationSlug ?? "");
  const base = pluginBasePath(organizationSlug, elemo.pluginId);

  return (
    <Card data-testid="accounting-org-settings">
      <CardHeader>
        <CardTitle>Accounting</CardTitle>
        <CardDescription>
          Manage the chart of accounts and hour envelopes. Bind a time source on
          this organization&apos;s plugin page so logged time counts against
          budgets.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2">
        <Button
          size="sm"
          nativeButton={false}
          render={<a href={`${base}/accounts`} />}
        >
          Accounts
        </Button>
        <Button
          size="sm"
          variant="outline"
          nativeButton={false}
          render={<a href={`${base}/budgets`} />}
        >
          Budgets
        </Button>
        <Button
          size="sm"
          variant="outline"
          nativeButton={false}
          render={<a href={`${base}/report`} />}
        >
          Report
        </Button>
      </CardContent>
    </Card>
  );
}
