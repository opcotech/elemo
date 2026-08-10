import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { DetailField } from "@/components/shared/detail-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Namespace, Permission } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { formatDate } from "@/lib/format-date";
import { pluralize } from "@/lib/utils";

interface NamespaceDetailInfoProps {
  namespace: Namespace;
  permissions: Permission[];
  organizationId: string;
  organizationName: string;
}

export function NamespaceDetailInfo({
  namespace,
  permissions,
  organizationId,
  organizationName,
}: NamespaceDetailInfoProps) {
  const hasWritePermission = can(permissions, "write");

  const projectCount = namespace.projects?.length || 0;
  const documentCount = namespace.documents?.length || 0;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle>Namespace Information</CardTitle>
            <CardDescription>
              Details about the namespace and its resources.
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              render={
                <Link
                  to="/namespaces/$namespaceId"
                  params={{ namespaceId: namespace.id }}
                />
              }
            >
              Open namespace
            </Button>
            {hasWritePermission && (
              <Button
                variant="outline"
                size="sm"
                render={
                  <Link
                    to="/settings/organizations/$organizationId/namespaces/$namespaceId/edit"
                    params={{
                      organizationId,
                      namespaceId: namespace.id,
                    }}
                  />
                }
              >
                <Edit className="size-4" />
                Edit
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <DetailField label="Name" value={namespace.name} />

          <DetailField label="Organization">
            <Link
              to="/settings/organizations/$organizationId"
              params={{ organizationId }}
              className="text-primary hover:underline"
            >
              {organizationName}
            </Link>
          </DetailField>

          <DetailField label="Description" value={namespace.description} />

          <DetailField label="Projects">
            <Badge variant="secondary">
              {projectCount} {pluralize(projectCount, "project", "projects")}
            </Badge>
          </DetailField>

          <DetailField label="Documents">
            <Badge variant="secondary">
              {documentCount}{" "}
              {pluralize(documentCount, "document", "documents")}
            </Badge>
          </DetailField>

          <DetailField
            label="Created At"
            value={formatDate(namespace.created_at)}
          />
        </div>
      </CardContent>
    </Card>
  );
}
