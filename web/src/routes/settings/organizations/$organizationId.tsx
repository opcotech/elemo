import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { isNotFound, v1OrganizationGetOptions } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";
import { formatDate } from "@/lib/utils";

export const Route = createFileRoute("/settings/organizations/$organizationId")(
  {
    beforeLoad: requireAuthBeforeLoad,
    component: OrganizationDetailPage,
  }
);

function OrganizationDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="mb-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="mt-2 h-5 w-96" />
      </div>

      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="mt-2 h-4 w-80" />
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i}>
                <Skeleton className="h-4 w-20" />
                <Skeleton className="mt-1 h-5 w-32" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function OrganizationDetailHeader({
  title,
  description = "View organization information.",
}: {
  title: string;
  description?: string;
}) {
  return (
    <div className="mb-6">
      <h1 className="text-2xl font-bold">{title}</h1>
      <p className="mt-2 text-gray-600">{description}</p>
    </div>
  );
}

function OrganizationNotFound() {
  return (
    <div className="space-y-6">
      <OrganizationDetailHeader title="Organization Details" />

      <Alert variant="destructive">
        <AlertDescription>
          Organization not found. Please check the URL and try again.
        </AlertDescription>
      </Alert>
    </div>
  );
}

function OrganizationDetailError() {
  return (
    <div className="space-y-6">
      <OrganizationDetailHeader title="Organization Details" />

      <Alert variant="destructive">
        <AlertDescription>
          Failed to load organization details. Please try again later.
        </AlertDescription>
      </Alert>
    </div>
  );
}

function OrganizationDetailField({
  label,
  value,
  children,
}: {
  label: string;
  value?: string | null;
  children?: React.ReactNode;
}) {
  return (
    <div>
      <label className="text-muted-foreground text-sm font-medium">
        {label}
      </label>
      {children ? (
        <div className="mt-1 text-sm">{children}</div>
      ) : (
        <p className="mt-1 text-sm">
          {value || <span className="text-muted-foreground">—</span>}
        </p>
      )}
    </div>
  );
}

function OrganizationDetailPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId } = Route.useParams();
  const {
    data: organization,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  useEffect(() => {
    if (!organization) return;

    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Organizations",
        href: "/settings/organizations",
        isNavigatable: true,
      },
      {
        label: organization.name,
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization]);

  if (isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !organization) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  return (
    <div className="space-y-6">
      <OrganizationDetailHeader title={organization.name} />

      <Card>
        <CardHeader>
          <CardTitle>Organization Information</CardTitle>
          <CardDescription>
            Details about the organization and its status.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <OrganizationDetailField label="Name" value={organization.name} />

            <OrganizationDetailField label="Email" value={organization.email} />

            <OrganizationDetailField label="Website">
              {organization.website ? (
                <a
                  href={organization.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {organization.website}
                </a>
              ) : (
                <span className="text-muted-foreground">—</span>
              )}
            </OrganizationDetailField>

            <OrganizationDetailField label="Status">
              {organization.status === "active" ? (
                <Badge variant="success">Active</Badge>
              ) : (
                <Badge variant="destructive">Deleted</Badge>
              )}
            </OrganizationDetailField>

            <OrganizationDetailField
              label="Created At"
              value={formatDate(organization.created_at)}
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
