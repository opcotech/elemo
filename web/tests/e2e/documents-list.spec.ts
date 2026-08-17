import {
  createNamespaceDocument,
  createOrganizationDocument,
  createProject,
  createProjectDocument,
  getRandomProjectKey,
} from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage, DocumentsListPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.list Document List E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents List",
    });
  });

  const loginOwner = async (page: Parameters<typeof loginUser>[0]) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
  };

  test("should show an empty state on a project with no documents", async ({
    page,
  }) => {
    const project = await createProject(
      workspace.client,
      workspace.namespaceId,
      {
        key: getRandomProjectKey(),
        name: `Empty docs ${getRandomString(8)}`,
      }
    );

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoProject(workspace.namespaceId, project.id);
    await listPage.waitForLoad();
    await expect(page.getByText("No related documents")).toBeVisible();
  });

  test("should search project documents by title", async ({ page }) => {
    const token = getRandomString(8);
    const matchTitle = `Needle ${token}`;
    const otherTitle = `Other ${getRandomString(8)}`;
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: matchTitle,
      excerpt: "Searchable summary for the matching document",
    });
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: otherTitle,
      excerpt: "Different summary that should be filtered out",
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoProject(workspace.namespaceId, workspace.projectId);
    await listPage.waitForLoad();
    await listPage.list.search(token);

    await expect(listPage.list.getDocumentLink(matchTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(otherTitle)).toHaveCount(0);
  });

  test("should sort project documents by title", async ({ page }) => {
    const firstTitle = `Alpha sort ${getRandomString(8)}`;
    const lastTitle = `Zulu sort ${getRandomString(8)}`;
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: lastTitle,
    });
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: firstTitle,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoProject(workspace.namespaceId, workspace.projectId);
    await listPage.waitForLoad();
    await expect(listPage.list.getCreatorButton()).toBeVisible();
    await listPage.list.selectSort("Title");

    const titles = await listPage.list
      .getSectionContainer()
      .locator("h2")
      .allTextContents();
    expect(titles.indexOf(firstTitle)).toBeGreaterThanOrEqual(0);
    expect(titles.indexOf(lastTitle)).toBeGreaterThan(
      titles.indexOf(firstTitle)
    );
  });

  test("should open a document from the project list", async ({ page }) => {
    const title = `Open list ${getRandomString(8)}`;
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      { title }
    );

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoProject(workspace.namespaceId, workspace.projectId);
    await listPage.waitForLoad();
    await listPage.list.clickDocument(title);

    const documentPage = new DocumentPage(page);
    await documentPage.waitForLoad();
    await expect(page).toHaveURL(new RegExp(`/documents/${document.id}`));
    await expect(documentPage.editor.getTitleInput()).toHaveValue(title);
  });

  test("should list an organization document", async ({ page }) => {
    const title = `Org list ${getRandomString(8)}`;
    await createOrganizationDocument(
      workspace.client,
      workspace.organizationId,
      { title }
    );

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoOrganization(workspace.organizationId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();
  });

  test("should list a namespace document", async ({ page }) => {
    const title = `Namespace list ${getRandomString(8)}`;
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();
  });
});
