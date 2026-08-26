import { createNamespaceDocument, createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import {
  DocumentPage,
  DocumentsListPage,
  HomePage,
  ProjectPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";

test.describe("@documents.create Document Create E2E Tests", () => {
  let workspace: OwnerWorkspace;
  let readerUser: User;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Create",
    });
    readerUser = await createUser(testConfig);

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      workspace.organizationId
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Organization",
      workspace.organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      workspace.namespaceId,
      ["namespace.read"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      workspace.projectId,
      ["project.read", "issue.read", "document.read"]
    );
  });

  const openProjectDocuments = async (
    page: Parameters<typeof loginUser>[0],
    email = workspace.owner.email
  ) => {
    await loginUser(page, {
      email,
      password: USER_DEFAULT_PASSWORD,
    });
    const documents = new DocumentsListPage(page);
    await documents.gotoProject(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await documents.waitForLoad();
    return documents;
  };

  test("should create a document from the project documents page", async ({
    page,
  }) => {
    const documents = await openProjectDocuments(page);
    const name = `Create ${getRandomString(8)}`;

    await documents.list.clickCreate();
    await documents.quickCreate.waitForLoad();
    await documents.quickCreate.fillDocumentTitle(name);
    await documents.quickCreate.submitDocument();
    await waitForSuccessToast(page, "Document created successfully");

    const documentPage = new DocumentPage(page);
    await documentPage.waitForLoad();
    await expect(page).toHaveURL(/\/documents\//);
    await expect(documentPage.editor.getTitleInput()).toHaveValue(name);
  });

  test("should show Create unavailable without parent context on Home", async ({
    page,
  }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const homePage = new HomePage(page);
    await homePage.goto();
    await homePage.waitForLoad();
    await homePage.clickCreate();
    await homePage.quickCreate.waitForLoad();
    await homePage.quickCreate.selectEntityType("Document");

    await expect(
      homePage.quickCreate.getCreateUnavailableButton()
    ).toBeVisible();
    await expect(
      homePage.quickCreate.getCreateUnavailableButton()
    ).toBeDisabled();
  });

  test("should not show create on the documents list for a reader", async ({
    page,
  }) => {
    await createProjectDocument(workspace.client, workspace.projectId, {
      title: `Reader list ${getRandomString(8)}`,
    });
    const documents = await openProjectDocuments(page, readerUser.email);
    await expect(documents.list.getCreateButton()).toHaveCount(0);
  });

  test("should create a document from the project overview", async ({
    page,
  }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const projectPage = new ProjectPage(page);
    await projectPage.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await projectPage.waitForLoad();
    const name = `Overview ${getRandomString(8)}`;

    await projectPage.documents.openCreateDialog();
    await projectPage.documents.fillTitle(name);
    await projectPage.documents.submitCreate();
    await waitForSuccessToast(page, "Document created successfully");

    const documentPage = new DocumentPage(page);
    await documentPage.waitForLoad();
    await expect(page).toHaveURL(/\/documents\//);
    await expect(documentPage.editor.getTitleInput()).toHaveValue(name);
  });

  test("should link an existing namespace document from the project overview", async ({
    page,
  }) => {
    const name = `Overview link ${getRandomString(8)}`;
    await createNamespaceDocument(
      workspace.client,
      workspace.organizationId,
      workspace.namespaceId,
      {
        title: name,
      }
    );
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const projectPage = new ProjectPage(page);
    await projectPage.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await projectPage.waitForLoad();

    await projectPage.documents.linkDocument(name);
    await waitForSuccessToast(page, "Document linked");
    await expect(projectPage.documents.getDocumentLink(name)).toBeVisible();
  });
});
