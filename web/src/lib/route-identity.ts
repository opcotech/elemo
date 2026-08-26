import { notFound } from "@tanstack/react-router";

import {
  isCanonicalIssueKey,
  isCanonicalNamespaceSlug,
  isCanonicalOrganizationSlug,
  isCanonicalProjectKey,
} from "@/lib/slug";

export function requireOrganizationSlug(value: string): string {
  if (!isCanonicalOrganizationSlug(value)) {
    throw notFound();
  }
  return value;
}

export function requireNamespaceSlug(value: string): string {
  if (!isCanonicalNamespaceSlug(value)) {
    throw notFound();
  }
  return value;
}

export function requireProjectKey(value: string): string {
  if (!isCanonicalProjectKey(value)) {
    throw notFound();
  }
  return value;
}

export function requireIssueKey(value: string): string {
  if (!isCanonicalIssueKey(value)) {
    throw notFound();
  }
  return value;
}

function entityId(value: unknown): string | undefined {
  if (!value || typeof value !== "object") {
    return undefined;
  }
  const id = (value as { id?: unknown }).id;
  return typeof id === "string" && id.length > 0 ? id : undefined;
}

function stringField(
  value: Record<string, unknown>,
  key: string
): string | undefined {
  const field = value[key];
  return typeof field === "string" && field.length > 0 ? field : undefined;
}

export interface ResolvedRouteIdentity {
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}

/** Pull resolved xids out of hierarchical loader data. Never copy URL params. */
export function identityFromLoaderData(data: unknown): ResolvedRouteIdentity {
  if (!data || typeof data !== "object") {
    return {};
  }
  const record = data as Record<string, unknown>;
  return {
    organizationId:
      entityId(record.organization) ?? stringField(record, "organizationId"),
    namespaceId:
      entityId(record.namespace) ?? stringField(record, "namespaceId"),
    projectId: entityId(record.project) ?? stringField(record, "projectId"),
  };
}

export function identityFromMatches(
  matches: readonly { loaderData?: unknown }[]
): ResolvedRouteIdentity {
  const identity: ResolvedRouteIdentity = {};
  for (const match of matches) {
    const next = identityFromLoaderData(match.loaderData);
    identity.organizationId = next.organizationId ?? identity.organizationId;
    identity.namespaceId = next.namespaceId ?? identity.namespaceId;
    identity.projectId = next.projectId ?? identity.projectId;
  }
  return identity;
}
