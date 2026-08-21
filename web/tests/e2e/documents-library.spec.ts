import {
  createAuthenticatedClient,
  createNamespaceDocument,
  createNamespaceFolder,
  createProjectDocument,
  updateDocument,
} from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage, DocumentsListPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { Document } from "@/lib/api/types";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

function createdTime(document: Pick<Document, "created_at">): number {
  if (!document.created_at) {
    return 0;
  }
  const time = new Date(document.created_at).getTime();
  return Number.isNaN(time) ? 0 : time;
}

function titlesByCreated(
  documents: Pick<Document, "title" | "created_at">[],
  direction: "asc" | "desc"
): string[] {
  return [...documents]
    .sort((left, right) => {
      const delta = createdTime(left) - createdTime(right);
      const createdDelta = direction === "asc" ? delta : -delta;
      return createdDelta !== 0
        ? createdDelta
        : left.title.localeCompare(right.title);
    })
    .map((document) => document.title);
}

test.describe("@documents.library Document Library E2E Tests", () => {
  let workspace: OwnerWorkspace;

  // Isolated workspace per test so parallel workers never share folders,
  // documents, or creator filters on the same library list.
  test.beforeEach(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Library",
    });
  });

  const loginOwner = async (page: Parameters<typeof loginUser>[0]) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
  };

  test("should open the documents hub, list libraries, and browse one", async ({
    page,
  }) => {
    const folderName = `Hub ${getRandomString(8)}`;
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: folderName,
    });

    await loginOwner(page);
    await page.getByRole("link", { name: "Documents", exact: true }).click();

    const listPage = new DocumentsListPage(page);
    await expect(page).toHaveURL(/\/documents\/?$/);
    await expect(
      page.getByRole("heading", { name: "Documents", exact: true })
    ).toBeVisible();
    await expect(
      listPage.getHubLibraryLink(workspace.organizationName)
    ).toBeVisible();
    await expect(
      listPage.getHubLibraryLink(workspace.namespaceName)
    ).toBeVisible();
    await expect(
      listPage.getHubLibraryLink(workspace.organizationName)
    ).toContainText("Organization");
    await expect(
      listPage.getHubLibraryLink(workspace.namespaceName)
    ).toContainText("Namespace");

    await listPage.getHubLibraryLink(workspace.namespaceName).click();
    await expect(page).toHaveURL(
      new RegExp(`/namespaces/${workspace.namespaceId}/documents`)
    );
    await listPage.waitForLoad();
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();
    await expect(listPage.list.getBrowseMenuButton()).toBeVisible();
    await expect(listPage.list.getNewFolderButton()).toBeVisible();
  });

  test("should browse folders, move a document, and return it to the library root", async ({
    page,
  }) => {
    const folderName = `Guides ${getRandomString(8)}`;
    const title = `Library move ${getRandomString(8)}`;
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();

    await listPage.createFolder(folderName);
    await waitForSuccessToast(page, "Folder created");
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();

    await listPage.moveDocument(title, folderName);
    await waitForSuccessToast(page, "Document moved");
    await expect(listPage.list.getDocumentLink(title)).toHaveCount(0);

    await listPage.list.clickFolder(folderName);
    await expect(page).toHaveURL(/[?&]folder=/);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();

    const documentPage = new DocumentPage(page);
    await listPage.list.clickDocument(title);
    await documentPage.waitForLoad();
    await expect(documentPage.getLocation()).toContainText(
      workspace.namespaceName
    );
    await expect(documentPage.getLocation()).toContainText(folderName);

    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await listPage.list.clickFolder(folderName);
    await listPage.waitForLoad();
    await listPage.moveDocument(title, "Library root");
    await waitForSuccessToast(page, "Document moved");

    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();
  });

  test("should list folder documents in All and keep project-created docs on both lists", async ({
    page,
  }) => {
    const folderName = `Architecture ${getRandomString(8)}`;
    const filedTitle = `Filed ${getRandomString(8)}`;
    const projectTitle = `Project related ${getRandomString(8)}`;
    const folder = await createNamespaceFolder(
      workspace.client,
      workspace.namespaceId,
      { name: folderName }
    );
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title: filedTitle,
    });
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: projectTitle,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();
    await expect(listPage.list.getDocumentLink(projectTitle)).toBeVisible();

    await listPage.moveDocument(filedTitle, folderName);
    await waitForSuccessToast(page, "Document moved");
    await expect(listPage.list.getDocumentLink(filedTitle)).toHaveCount(0);
    await expect(listPage.list.getDocumentLink(projectTitle)).toBeVisible();

    await listPage.list.clickAll();
    await expect(page).toHaveURL(/[?&]all=true/);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(filedTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(projectTitle)).toBeVisible();
    await expect(listPage.list.getFolderLink(folderName)).toHaveCount(0);

    await listPage.gotoProject(workspace.namespaceId, workspace.projectId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(projectTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(filedTitle)).toHaveCount(0);

    await listPage.gotoNamespace(workspace.namespaceId, { folder: folder.id });
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(filedTitle)).toBeVisible();
  });

  test("should rename a folder from the library browse list", async ({
    page,
  }) => {
    const folderName = `Rename me ${getRandomString(8)}`;
    const nextName = `Renamed ${getRandomString(8)}`;
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: folderName,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();

    await listPage.renameFolder(folderName, nextName);
    await waitForSuccessToast(page, "Folder renamed");
    await expect(listPage.list.getFolderLink(folderName)).toHaveCount(0);
    await expect(listPage.list.getFolderLink(nextName)).toBeVisible();
  });

  test("should rename a document from the library browse list", async ({
    page,
  }) => {
    const title = `Rename doc ${getRandomString(8)}`;
    const nextTitle = `Renamed doc ${getRandomString(8)}`;
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();

    await listPage.renameDocument(title, nextTitle);
    await waitForSuccessToast(page, "Document renamed");
    await expect(listPage.list.getDocumentLink(title)).toHaveCount(0);
    await expect(listPage.list.getDocumentLink(nextTitle)).toBeVisible();
  });

  test("should delete a folder and reparent children one level up", async ({
    page,
  }) => {
    const parentName = `Parent ${getRandomString(8)}`;
    const childName = `Child ${getRandomString(8)}`;
    const title = `Reparented ${getRandomString(8)}`;
    const parent = await createNamespaceFolder(
      workspace.client,
      workspace.namespaceId,
      { name: parentName }
    );
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: childName,
      parent_id: parent.id,
    });
    const document = await createNamespaceDocument(
      workspace.client,
      workspace.namespaceId,
      { title }
    );
    await updateDocument(workspace.client, document.id, {
      folder_id: parent.id,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getFolderLink(parentName)).toBeVisible();
    await expect(listPage.list.getFolderLink(childName)).toHaveCount(0);
    await expect(listPage.list.getDocumentLink(title)).toHaveCount(0);

    await listPage.deleteFolder(parentName);
    await waitForSuccessToast(page, "Folder deleted");
    await expect(listPage.list.getFolderLink(parentName)).toHaveCount(0);
    await expect(listPage.list.getFolderLink(childName)).toBeVisible();
    await expect(listPage.list.getDocumentLink(title)).toBeVisible();
  });

  test("should move a folder within the library and hide invalid targets", async ({
    page,
  }) => {
    const sourceName = `Source ${getRandomString(8)}`;
    const nestedName = `Nested ${getRandomString(8)}`;
    const destinationName = `Destination ${getRandomString(8)}`;
    const source = await createNamespaceFolder(
      workspace.client,
      workspace.namespaceId,
      { name: sourceName }
    );
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: nestedName,
      parent_id: source.id,
    });
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: destinationName,
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();

    await listPage.openFolderMovePicker(sourceName);
    await expect(
      page.getByRole("option", { name: destinationName, exact: true })
    ).toBeVisible();
    await expect(
      page.getByRole("option", { name: sourceName, exact: true })
    ).toHaveCount(0);
    await expect(
      page.getByRole("option", {
        name: `${sourceName} / ${nestedName}`,
        exact: true,
      })
    ).toHaveCount(0);
    await page
      .getByRole("option", { name: destinationName, exact: true })
      .click();
    await page
      .getByRole("dialog", { name: "Move folder" })
      .getByRole("button", { name: "Move", exact: true })
      .click();
    await waitForSuccessToast(page, "Folder moved");
    await expect(listPage.list.getFolderLink(sourceName)).toHaveCount(0);

    await listPage.list.clickFolder(destinationName);
    await listPage.waitForLoad();
    await expect(listPage.list.getFolderLink(sourceName)).toBeVisible();
  });

  test("should search library documents and folders by name", async ({
    page,
  }) => {
    const token = getRandomString(8);
    const matchTitle = `Needle ${token}`;
    const otherTitle = `Other ${getRandomString(8)}`;
    const folderName = `Search folder ${getRandomString(8)}`;
    await createNamespaceFolder(workspace.client, workspace.namespaceId, {
      name: folderName,
    });
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title: matchTitle,
      excerpt: "Searchable summary for the matching document",
    });
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title: otherTitle,
      excerpt: "Different summary that should be filtered out",
    });

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(matchTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(otherTitle)).toBeVisible();
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();

    await listPage.list.search(token);
    await expect(listPage.list.getDocumentLink(matchTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(otherTitle)).toHaveCount(0);
    await expect(listPage.list.getFolderLink(folderName)).toHaveCount(0);

    await listPage.list.search("Searchable summary");
    await expect(listPage.list.getDocumentLink(matchTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(otherTitle)).toHaveCount(0);
    await expect(listPage.list.getFolderLink(folderName)).toHaveCount(0);

    await listPage.list.search(folderName);
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();
    await expect(listPage.list.getDocumentLink(matchTitle)).toHaveCount(0);
    await expect(listPage.list.getDocumentLink(otherTitle)).toHaveCount(0);

    await listPage.list.search(`missing-${getRandomString(8)}`);
    await expect(page.getByText("No document matches")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Clear search" })
    ).toBeVisible();

    await page.getByRole("button", { name: "Clear search" }).click();
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(matchTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(otherTitle)).toBeVisible();
    await expect(listPage.list.getFolderLink(folderName)).toBeVisible();
  });

  test("should sort library documents by title, created, and oldest", async ({
    page,
  }) => {
    const firstTitle = `Alpha lib ${getRandomString(8)}`;
    const lastTitle = `Zulu lib ${getRandomString(8)}`;
    const lastDocument = await createNamespaceDocument(
      workspace.client,
      workspace.namespaceId,
      {
        title: lastTitle,
      }
    );
    const firstDocument = await createNamespaceDocument(
      workspace.client,
      workspace.namespaceId,
      {
        title: firstTitle,
      }
    );

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(firstTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(lastTitle)).toBeVisible();

    await listPage.list.selectSort("Title");
    await expect(listPage.list.getDocumentTitles()).toHaveText([
      firstTitle,
      lastTitle,
    ]);

    await listPage.list.selectSort("Oldest");
    await expect(listPage.list.getDocumentTitles()).toHaveText(
      titlesByCreated([lastDocument, firstDocument], "asc")
    );

    await listPage.list.selectSort("Created");
    await expect(listPage.list.getDocumentTitles()).toHaveText(
      titlesByCreated([lastDocument, firstDocument], "desc")
    );
  });

  test("should filter library documents by creator", async ({
    page,
    testConfig,
  }) => {
    const ownerTitle = `Owner doc ${getRandomString(8)}`;
    const writerTitle = `Writer doc ${getRandomString(8)}`;
    const writer = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      writer.email,
      "Organization",
      workspace.organizationId
    );
    await grantActionsToUser(
      testConfig,
      writer.email,
      "Organization",
      workspace.organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      writer.email,
      "Namespace",
      workspace.namespaceId,
      ["namespace.read"]
    );
    await grantActionsToUser(
      testConfig,
      writer.email,
      "Namespace",
      workspace.namespaceId,
      ["namespace.update", "project.create", "document.create", "folder.create"]
    );
    const writerClient = await createAuthenticatedClient(
      writer.email,
      USER_DEFAULT_PASSWORD
    );
    await createNamespaceDocument(workspace.client, workspace.namespaceId, {
      title: ownerTitle,
    });
    await createNamespaceDocument(writerClient, workspace.namespaceId, {
      title: writerTitle,
    });

    const ownerName = `${workspace.owner.first_name} ${workspace.owner.last_name}`;
    const writerName = `${writer.first_name} ${writer.last_name}`;

    await loginOwner(page);
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(ownerTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(writerTitle)).toBeVisible();

    await listPage.list.selectCreator(ownerName);
    await expect(listPage.list.getDocumentLink(ownerTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(writerTitle)).toHaveCount(0);

    await listPage.list.selectCreator(writerName);
    await expect(listPage.list.getDocumentLink(writerTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(ownerTitle)).toHaveCount(0);

    await listPage.list.selectCreator("Anyone");
    await expect(listPage.list.getDocumentLink(ownerTitle)).toBeVisible();
    await expect(listPage.list.getDocumentLink(writerTitle)).toBeVisible();
  });

  test("should change a document library and clear its folder", async ({
    page,
  }) => {
    const folderName = `Relocate ${getRandomString(8)}`;
    const title = `Change library ${getRandomString(8)}`;
    const folder = await createNamespaceFolder(
      workspace.client,
      workspace.namespaceId,
      { name: folderName }
    );
    const document = await createNamespaceDocument(
      workspace.client,
      workspace.namespaceId,
      { title }
    );
    await updateDocument(workspace.client, document.id, {
      folder_id: folder.id,
    });

    await loginOwner(page);
    const documentPage = new DocumentPage(page);
    await documentPage.goto(document.id);
    await documentPage.waitForLoad();
    await expect(documentPage.getLocation()).toContainText(
      workspace.namespaceName
    );
    await expect(documentPage.getLocation()).toContainText(folderName);

    await documentPage.changeLibrary(
      `Organization · ${workspace.organizationName}`
    );
    await waitForSuccessToast(page, "Library updated");
    await expect(documentPage.getLocation()).toContainText(
      workspace.organizationName
    );
    await expect(documentPage.getLocation()).not.toContainText(folderName);

    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toHaveCount(0);

    await expect(async () => {
      await listPage.gotoOrganization(workspace.organizationId);
      await expect(page).toHaveURL(
        new RegExp(`/organizations/${workspace.organizationId}/documents`)
      );
      await listPage.waitForLoad();
      await expect(listPage.list.getDocumentLink(title)).toBeVisible({
        timeout: 8_000,
      });
    }).toPass({ timeout: 25_000 });
  });

  test("should change a document library to another namespace", async ({
    page,
  }) => {
    const title = `Namespace hop ${getRandomString(8)}`;
    const document = await createNamespaceDocument(
      workspace.client,
      workspace.namespaceId,
      { title }
    );
    const nextNamespaceName = `Library hop ${getRandomString(8)}`;
    const nextNamespace = await v1OrganizationsNamespacesCreate({
      client: workspace.client,
      path: { id: workspace.organizationId },
      body: {
        name: nextNamespaceName,
        description: "Destination library for a document move",
      },
      throwOnError: true,
    });
    const nextNamespaceId = nextNamespace.data.id ?? "";

    await loginOwner(page);
    const documentPage = new DocumentPage(page);
    await documentPage.goto(document.id);
    await documentPage.waitForLoad();
    await expect(documentPage.getLocation()).toContainText(
      workspace.namespaceName
    );

    await documentPage.changeLibrary(`Namespace · ${nextNamespaceName}`);
    await waitForSuccessToast(page, "Library updated");
    await expect(documentPage.getLocation()).toContainText(nextNamespaceName);
    await expect(documentPage.getLocation()).not.toContainText(
      workspace.namespaceName
    );

    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(workspace.namespaceId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(title)).toHaveCount(0);

    await expect(async () => {
      await listPage.gotoNamespace(nextNamespaceId);
      await expect(page).toHaveURL(
        new RegExp(`/namespaces/${nextNamespaceId}/documents`)
      );
      await listPage.waitForLoad();
      await expect(listPage.list.getDocumentLink(title)).toBeVisible({
        timeout: 8_000,
      });
    }).toPass({ timeout: 25_000 });
  });
});
