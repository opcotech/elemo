import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { DetailField } from "@/components/detail-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Project } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { formatDate, pluralize } from "@/lib/utils";

interface ProjectDetailInfoProps {
  project: Project;
  organizationId: string;
  namespaceId: string;
  namespaceName: string;
}

export function ProjectDetailInfo({
  project,
  organizationId,
  namespaceId,
  namespaceName,
}: ProjectDetailInfoProps) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );

  const hasWritePermission = can(permissions, "write");

  const documentCount = project.documents?.length || 0;
  const issueCount = project.issues?.length || 0;

  return (
    <Card data-section="project-info">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle>Project Information</CardTitle>
            <CardDescription>
              Details about the project and its resources.
            </CardDescription>
          </div>
          {hasWritePermission && (
            <Button variant="outline" size="sm" asChild>
              <Link
                to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/edit"
                params={{
                  organizationId,
                  namespaceId,
                  projectId: project.id,
                }}
              >
                <Edit className="size-4" />
                Edit
              </Link>
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <DetailField label="Key" value={project.key} />

          <DetailField label="Name" value={project.name} />

          <DetailField label="Description" value={project.description} />

          <DetailField label="Status">
            <Badge
              variant={project.status === "active" ? "success" : "secondary"}
            >
              {project.status}
            </Badge>
          </DetailField>

          <DetailField label="Namespace">
            <Link
              to="/settings/organizations/$organizationId/namespaces/$namespaceId"
              params={{ organizationId, namespaceId }}
              className="text-primary hover:underline"
            >
              {namespaceName}
            </Link>
          </DetailField>

          <DetailField label="Documents">
            <Badge variant="secondary">
              {documentCount}{" "}
              {pluralize(documentCount, "document", "documents")}
            </Badge>
          </DetailField>

          <DetailField label="Issues">
            <Badge variant="secondary">
              {issueCount} {pluralize(issueCount, "issue", "issues")}
            </Badge>
          </DetailField>

          <DetailField
            label="Created At"
            value={formatDate(project.created_at)}
          />

          <DetailField
            label="Updated At"
            value={formatDate(project.updated_at)}
          />
        </div>
      </CardContent>
    </Card>
  );
}
