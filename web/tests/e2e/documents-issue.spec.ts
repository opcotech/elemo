import {
  createIssue,
  createIssueDocument,
  createNamespaceDocument,
} from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage, WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.issue Document Issue E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Issue",
    });
  });

  const openIssue = async (
    page: Parameters<typeof loginUser>[0],
    title?: string
  ) => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: title ?? `Docs issue ${getRandomString(8)}`,
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workItem = new WorkItemPage(page);
    await workItem.goto(workspace.namespaceId, issue.key);
    await workItem.waitForLoad();
    await workItem.documents.waitForLoad();
    return { workItem, issue };
  };

  test("should show an empty linked documents state", async ({ page }) => {
    const { workItem } = await openIssue(page);
    await expect(workItem.documents.getEmptyState()).toBeVisible();
  });

  test("should create a linked document from the issue page", async ({
    page,
  }) => {
    const { workItem } = await openIssue(page);
    const name = `Issue doc ${getRandomString(8)}`;

    await workItem.documents.openCreateDialog();
    await workItem.documents.fillTitle(name);
    await workItem.documents.submitCreate();
    await waitForSuccessToast(page, "Document created successfully");

    const documentPage = new DocumentPage(page);
    await documentPage.waitForLoad();
    await expect(page).toHaveURL(/\/documents\//);
    await expect(documentPage.editor.getTitleInput()).toHaveValue(name);
  });

  test("should list an existing linked document", async ({ page }) => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Linked ${getRandomString(8)}`,
    });
    const name = `Linked doc ${getRandomString(8)}`;
    await createIssueDocument(workspace.client, issue.id, { title: name });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workItem = new WorkItemPage(page);
    await workItem.goto(workspace.namespaceId, issue.key);
    await workItem.waitForLoad();
    await workItem.documents.waitForLoad();
    await expect(workItem.documents.getDocumentLink(name)).toBeVisible();
  });

  test("should link an existing namespace document to an issue", async ({
    page,
  }) => {
    const name = `Available doc ${getRandomString(8)}`;
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title: name,
    });
    const { workItem } = await openIssue(page);

    await workItem.documents.linkDocument(name);
    await waitForSuccessToast(page, "Document linked");
    await expect(workItem.documents.getDocumentLink(name)).toBeVisible();
  });
});
