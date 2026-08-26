import type { ReactNode } from "react";

import { IssueReportedBy } from "./issue-reported-by";
import { formatDateTime } from "./utils";

import { InternalLink } from "@/components/ui/internal-link";
import type { PersonAvatarStackPerson } from "@/components/ui/person-avatar-stack";
import { PropertyList } from "@/components/ui/property-list";
import { internalPath } from "@/lib/internal-url";
import { namespacePath, projectWorkPath, workItemPath } from "@/lib/paths";
import type { DataSource } from "@/lib/work/model";

export function IssueParentLink({
  parent,
  organizationSlug,
  namespaceSlug,
}: {
  parent?: {
    key: string;
    title: string;
    namespaceSlug?: string;
    organizationSlug?: string;
  } | null;
  organizationSlug?: string | null;
  namespaceSlug?: string | null;
}) {
  if (!parent) {
    return <span className="px-2">None</span>;
  }

  const parentOrganizationSlug = parent.organizationSlug ?? organizationSlug;
  const parentNamespaceSlug = parent.namespaceSlug ?? namespaceSlug;
  if (!parentOrganizationSlug || !parentNamespaceSlug) {
    return (
      <span className="px-2">
        {parent.key} {parent.title}
      </span>
    );
  }

  return (
    <InternalLink
      className="text-primary px-2 underline-offset-4 hover:underline"
      to={internalPath(
        workItemPath({
          organizationSlug: parentOrganizationSlug,
          namespaceSlug: parentNamespaceSlug,
          issueKey: parent.key,
        })
      )}
    >
      {parent.key} {parent.title}
    </InternalLink>
  );
}

export function IssueMetadataProperties({
  organizationSlug,
  namespaceSlug,
  namespaceLabel,
  projectKey,
  projectLabel,
  parent,
  reportedById,
  createdAt,
  updatedAt,
  compact = false,
  reporterPeople,
  reporterDataSource = "api",
  namespaceId,
}: {
  organizationSlug?: string | null;
  namespaceSlug?: string | null;
  namespaceLabel: string;
  projectKey?: string | null;
  projectLabel: string;
  parent?: ReactNode;
  reportedById: string;
  createdAt?: string | null;
  updatedAt?: string | null;
  compact?: boolean;
  reporterPeople?: readonly PersonAvatarStackPerson[];
  reporterDataSource?: DataSource;
  namespaceId?: string | null;
}) {
  return (
    <PropertyList
      compact={compact}
      items={[
        {
          label: "Namespace",
          value:
            organizationSlug && namespaceSlug ? (
              <InternalLink
                className="text-primary px-2 underline-offset-4 hover:underline"
                to={internalPath(
                  namespacePath({ organizationSlug, namespaceSlug })
                )}
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
            organizationSlug && namespaceSlug && projectKey ? (
              <InternalLink
                className="text-primary px-2 underline-offset-4 hover:underline"
                to={internalPath(
                  projectWorkPath({
                    organizationSlug,
                    namespaceSlug,
                    projectKey,
                  })
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
