import { Link } from "@tanstack/react-router";
import { Edit, ListTree, Puzzle } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CountBadge } from "@/components/ui/count-badge";
import { DetailCard } from "@/components/ui/detail-card";
import { DetailField } from "@/components/ui/detail-field";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Project } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import { formatDate } from "@/lib/format-date";

interface ProjectDetailInfoProps {
  project: Project;
  organizationSlug: string;
  namespaceSlug: string;
  namespaceName: string;
}

export function ProjectDetailInfo({
  project,
  organizationSlug,
  namespaceSlug,
  namespaceName,
}: ProjectDetailInfoProps) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );

  const hasWritePermission = can(permissions, Action.ProjectUpdate);

  const documentCount = project.document_count ?? 0;
  const issueCount = project.issue_count ?? 0;

  return (
    <DetailCard
      data-section="project-info"
      title="Project Information"
      description="Details about the project and its resources."
      actions={
        <>
          <Button
            variant="outline"
            size="sm"
            render={
              <Link
                to="/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
                params={{
                  organizationSlug,
                  namespaceSlug,
                  projectKey: project.key,
                }}
              />
            }
          >
            Open project
          </Button>
          <Button
            variant="outline"
            size="sm"
            render={
              <Link
                to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/custom-fields"
                params={{
                  organizationSlug,
                  namespaceSlug,
                  projectKey: project.key,
                }}
              />
            }
          >
            <ListTree className="size-4" />
            Custom fields
          </Button>
          <Button
            variant="outline"
            size="sm"
            render={
              <Link
                to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/plugins"
                params={{
                  organizationSlug,
                  namespaceSlug,
                  projectKey: project.key,
                }}
              />
            }
          >
            <Puzzle className="size-4" />
            Plugins
          </Button>
          {hasWritePermission && (
            <Button
              variant="outline"
              size="sm"
              render={
                <Link
                  to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/edit"
                  params={{
                    organizationSlug,
                    namespaceSlug,
                    projectKey: project.key,
                  }}
                />
              }
            >
              <Edit className="size-4" />
              Edit
            </Button>
          )}
        </>
      }
    >
      <DetailField label="Key" value={project.key} />

      <DetailField label="Name" value={project.name} />

      <DetailField label="Description" value={project.description} />

      <DetailField label="Status">
        <Badge variant={project.status === "active" ? "success" : "secondary"}>
          {project.status}
        </Badge>
      </DetailField>

      <DetailField label="Namespace">
        <Link
          to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug"
          params={{ organizationSlug, namespaceSlug }}
          className="text-primary hover:underline"
        >
          {namespaceName}
        </Link>
      </DetailField>

      <DetailField label="Documents">
        <CountBadge
          count={documentCount}
          singular="document"
          plural="documents"
        />
      </DetailField>

      <DetailField label="Issues">
        <CountBadge count={issueCount} singular="issue" plural="issues" />
      </DetailField>

      <DetailField label="Created At" value={formatDate(project.created_at)} />

      <DetailField label="Updated At" value={formatDate(project.updated_at)} />
    </DetailCard>
  );
}
