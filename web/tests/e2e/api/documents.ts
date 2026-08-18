import { withErrorHandling } from "./error-handler";

import type { Document, DocumentCreate, DocumentPatch } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import {
  v1DocumentGet,
  v1DocumentUpdate,
  v1IssuesDocumentsCreate,
  v1NamespacesDocumentsCreate,
  v1OrganizationsDocumentsCreate,
  v1ProjectsDocumentsCreate,
} from "@/lib/client/sdk.gen";

type DocumentCreateFields = Partial<DocumentCreate> & { title: string };

function documentCreateBody(
  documentData: DocumentCreateFields
): DocumentCreate {
  return {
    title: documentData.title,
    content: documentData.content ?? `# ${documentData.title.trim()}`,
    ...(documentData.excerpt ? { excerpt: documentData.excerpt } : {}),
  };
}

async function getDocument(
  client: Client,
  documentId: string
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1DocumentGet({
        client,
        path: { id: documentId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/documents/${documentId}`,
      method: "GET",
    }
  );

  return response.data;
}

/**
 * Create a project document via API, then fetch the full document by ID.
 */
export async function createProjectDocument(
  client: Client,
  projectId: string,
  documentData: DocumentCreateFields
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1ProjectsDocumentsCreate({
        client,
        path: { id: projectId },
        body: documentCreateBody(documentData),
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/projects/${projectId}/documents`,
      method: "POST",
    }
  );

  return getDocument(client, response.data.id || "");
}

/**
 * Create an organization document via API, then fetch the full document by ID.
 */
export async function createOrganizationDocument(
  client: Client,
  organizationId: string,
  documentData: DocumentCreateFields
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationsDocumentsCreate({
        client,
        path: { id: organizationId },
        body: documentCreateBody(documentData),
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/documents`,
      method: "POST",
    }
  );

  return getDocument(client, response.data.id || "");
}

/**
 * Create a namespace document via API, then fetch the full document by ID.
 */
export async function createNamespaceDocument(
  client: Client,
  namespaceId: string,
  documentData: DocumentCreateFields
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1NamespacesDocumentsCreate({
        client,
        path: { id: namespaceId },
        body: documentCreateBody(documentData),
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/namespaces/${namespaceId}/documents`,
      method: "POST",
    }
  );

  return getDocument(client, response.data.id || "");
}

/**
 * Create an issue-linked document via API, then fetch the full document by ID.
 */
export async function createIssueDocument(
  client: Client,
  issueId: string,
  documentData: DocumentCreateFields
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1IssuesDocumentsCreate({
        client,
        path: { id: issueId },
        body: documentCreateBody(documentData),
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/issues/${issueId}/documents`,
      method: "POST",
    }
  );

  return getDocument(client, response.data.id || "");
}

/**
 * Update a document via API.
 */
export async function updateDocument(
  client: Client,
  documentId: string,
  patch: DocumentPatch
): Promise<Document> {
  const response = await withErrorHandling(
    async () => {
      return await v1DocumentUpdate({
        client,
        path: { id: documentId },
        body: patch,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/documents/${documentId}`,
      method: "PATCH",
    }
  );

  return response.data;
}
