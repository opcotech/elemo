import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { Button } from "@/components/ui/button";
import { CountBadge } from "@/components/ui/count-badge";
import { DetailCard } from "@/components/ui/detail-card";
import { DetailField } from "@/components/ui/detail-field";
import type { EffectiveActions, Namespace } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import { formatDate } from "@/lib/format-date";

interface NamespaceDetailInfoProps {
  namespace: Namespace;
  permissions: EffectiveActions;
  organizationId: string;
  organizationName: string;
}

export function NamespaceDetailInfo({
  namespace,
  permissions,
  organizationId,
  organizationName,
}: NamespaceDetailInfoProps) {
  const hasWritePermission = can(permissions, Action.NamespaceUpdate);

  const projectCount = namespace.project_count ?? 0;
  const documentCount = namespace.document_count ?? 0;

  return (
    <DetailCard
      title="Namespace Information"
      description="Details about the namespace and its resources."
      actions={
        <>
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
        </>
      }
    >
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
        <CountBadge count={projectCount} singular="project" plural="projects" />
      </DetailField>

      <DetailField label="Documents">
        <CountBadge
          count={documentCount}
          singular="document"
          plural="documents"
        />
      </DetailField>

      <DetailField
        label="Created At"
        value={formatDate(namespace.created_at)}
      />
    </DetailCard>
  );
}
