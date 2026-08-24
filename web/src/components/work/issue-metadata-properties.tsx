import type { ReactNode } from "react";

import { IssueReportedBy } from "./issue-reported-by";
import { formatDateTime } from "./utils";

import { InternalLink } from "@/components/ui/internal-link";
import type { PersonAvatarStackPerson } from "@/components/ui/person-avatar-stack";
import { PropertyList } from "@/components/ui/property-list";
import { internalPath } from "@/lib/internal-url";
import type { DataSource } from "@/lib/work/model";

export function IssueParentLink({
  parent,
  namespaceId,
}: {
  parent?: {
    key: string;
    title: string;
    namespaceId?: string;
  } | null;
  namespaceId?: string | null;
}) {
  if (!parent) {
    return <span className="px-2">None</span>;
  }

  const parentNamespaceId = parent.namespaceId ?? namespaceId;
  if (!parentNamespaceId) {
    return (
      <span className="px-2">
        {parent.key} {parent.title}
      </span>
    );
  }

  return (
    <InternalLink
      className="text-primary px-2 underline-offset-4 hover:underline"
      to={internalPath(`/work/${parentNamespaceId}/${parent.key}`)}
    >
      {parent.key} {parent.title}
    </InternalLink>
  );
}

export function IssueMetadataProperties({
  namespaceId,
  namespaceLabel,
  projectId,
  projectLabel,
  parent,
  reportedById,
  createdAt,
  updatedAt,
  compact = false,
  reporterPeople,
  reporterDataSource = "api",
}: {
  namespaceId?: string | null;
  namespaceLabel: string;
  projectId?: string | null;
  projectLabel: string;
  parent?: ReactNode;
  reportedById: string;
  createdAt?: string | null;
  updatedAt?: string | null;
  compact?: boolean;
  reporterPeople?: readonly PersonAvatarStackPerson[];
  reporterDataSource?: DataSource;
}) {
  return (
    <PropertyList
      compact={compact}
      items={[
        {
          label: "Namespace",
          value: namespaceId ? (
            <InternalLink
              className="text-primary px-2 underline-offset-4 hover:underline"
              to={internalPath(`/namespaces/${namespaceId}`)}
            >
              {namespaceLabel}
            </InternalLink>
          ) : (
            <span className="px-2">Unknown</span>
          ),
        },
        {
          label: "Project",
          value:
            namespaceId && projectId ? (
              <InternalLink
                className="text-primary px-2 underline-offset-4 hover:underline"
                to={internalPath(
                  `/namespaces/${namespaceId}/projects/${projectId}/work`
                )}
              >
                {projectLabel}
              </InternalLink>
            ) : (
              <span className="px-2">{projectLabel}</span>
            ),
        },
        {
          label: "Parent",
          value: parent ?? <span className="px-2">None</span>,
        },
        {
          label: "Reported by",
          value: (
            <div className="px-2">
              <IssueReportedBy
                userId={reportedById}
                namespaceId={namespaceId}
                people={reporterPeople}
                dataSource={reporterDataSource}
              />
            </div>
          ),
        },
        {
          label: "Created at",
          value: <span className="px-2">{formatDateTime(createdAt)}</span>,
        },
        {
          label: "Updated at",
          value: <span className="px-2">{formatDateTime(updatedAt)}</span>,
        },
      ]}
    />
  );
}
