import { createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage, DocumentsListPage, HomePage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.delete Document Delete E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Delete",
    });
  });

  const openDocument = async (
    page: Parameters<typeof loginUser>[0],
    name?: string
  ) => {
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      {
        title: name ?? `Delete ${getRandomString(8)}`,
      }
    );

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const documentPage = new DocumentPage(page);
    await documentPage.goto(document.id);
    await documentPage.waitForLoad();
    return { documentPage, document };
  };

  test("should delete a document from more actions and leave the page", async ({
    page,
  }) => {
    const name = `Delete gone ${getRandomString(8)}`;
    const { documentPage, document } = await openDocument(page, name);

    await documentPage.openDeleteDialog();
    await expect(
      page.getByRole("heading", {
        name: `Are you sure you want to delete ${name}?`,
      })
    ).toBeVisible();
    await documentPage.confirmDelete();
    await waitForSuccessToast(page, "Document deleted");

    const homePage = new HomePage(page);
    await homePage.waitForLoad();
    await expect(page).not.toHaveURL(new RegExp(`/documents/${document.id}`));

    const listPage = new DocumentsListPage(page);
    await listPage.gotoProject(workspace.namespaceId, workspace.projectId);
    await listPage.waitForLoad();
    await expect(listPage.list.getDocumentLink(name)).toHaveCount(0);
  });

  test("should cancel delete and keep the document", async ({ page }) => {
    const name = `Delete keep ${getRandomString(8)}`;
    const { documentPage, document } = await openDocument(page, name);

    await documentPage.openDeleteDialog();
    await documentPage.cancelDelete();
    await expect(page).toHaveURL(new RegExp(`/documents/${document.id}`));
    await expect(documentPage.editor.getTitleInput()).toHaveValue(name);
  });
});
