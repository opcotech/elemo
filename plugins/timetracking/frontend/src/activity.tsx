import type { ElemoPluginAPI } from "@elemo/plugin-sdk";
import { EmptyState } from "@elemo/plugin-ui";
import { Timer } from "lucide-react";

import { TimeEntryList, useIssueEntries } from "./entries";

const PLUGIN_ID = "com.elemo.timetracking";

export function LoggedTime(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const issueId = String(props.issueId ?? "");
  const organizationSlug = String(props.organizationSlug ?? "");
  const projectId =
    typeof props.projectId === "string" ? props.projectId : undefined;
  const { entries, error, reload } = useIssueEntries(elemo, issueId);
  const reportHref =
    organizationSlug && issueId
      ? `/organizations/${organizationSlug}/plugins/${PLUGIN_ID}/report?workItem=${encodeURIComponent(issueId)}`
      : undefined;

  if (error) {
    return (
      <div data-testid="timetracking-logged-time" className="space-y-3">
        <ReportLink href={reportHref} />
        <EmptyState compact title="Logged time" description={error} />
      </div>
    );
  }

  if (entries.length === 0) {
    return (
      <div data-testid="timetracking-logged-time" className="space-y-3">
        <ReportLink href={reportHref} />
        <EmptyState
          compact
          icon={<Timer />}
          title="No time logged"
          description="Start a timer or log time from the sidebar to add entries here."
        />
      </div>
    );
  }

  return (
    <div data-testid="timetracking-logged-time" className="space-y-3">
      <ReportLink href={reportHref} />
      <TimeEntryList
        elemo={elemo}
        entries={entries}
        projectId={projectId}
        compact
        onChanged={reload}
      />
    </div>
  );
}

function ReportLink({ href }: { href?: string }) {
  if (!href) {
    return null;
  }
  return (
    <p className="text-muted-foreground text-right text-sm">
      View the full{" "}
      <a
        href={href}
        data-testid="timetracking-report-link"
        className="text-foreground underline underline-offset-3"
      >
        time report
      </a>
    </p>
  );
}
