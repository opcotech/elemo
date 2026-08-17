import { createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.edit Document Edit E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Edit",
    });
  });

  const openDocument = async (page: Parameters<typeof loginUser>[0]) => {
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      {
        title: `Edit ${getRandomString(8)}`,
        excerpt: `Original excerpt ${getRandomString(8)}`,
        content: `# Seed\n\nOriginal body ${getRandomString(8)}`,
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

  test("should update the document title", async ({ page }) => {
    const { documentPage } = await openDocument(page);
    const updatedTitle = `Title ${getRandomString(8)}`;

    await documentPage.editor.editTitle(updatedTitle);
    await documentPage.editor.save();
    await waitForSuccessToast(page, "Title updated");
    await expect(documentPage.editor.getTitleInput()).toHaveValue(updatedTitle);
  });

  test("should update the document excerpt", async ({ page }) => {
    const { documentPage } = await openDocument(page);
    const updatedExcerpt = `Updated excerpt ${getRandomString(8)}`;

    await documentPage.editor.editExcerpt(updatedExcerpt);
    await documentPage.editor.save();
    await waitForSuccessToast(page, "Excerpt updated");
    await expect(documentPage.editor.getExcerptInput()).toHaveValue(
      updatedExcerpt
    );
  });

  test("should update the document content", async ({ page }) => {
    const { documentPage } = await openDocument(page);
    const addition = `Addition ${getRandomString(8)}`;

    await documentPage.editor.typeContent(addition);
    await documentPage.editor.save();
    await waitForSuccessToast(page, "Content updated");
    await expect(documentPage.editor.getContentEditor()).toContainText(
      addition
    );
  });

  test("should discard unsaved name changes", async ({ page }) => {
    const { documentPage, document } = await openDocument(page);

    await documentPage.editor.editTitle(`Discard ${getRandomString(8)}`);
    await documentPage.editor.discard();
    await expect(documentPage.editor.getTitleInput()).toHaveValue(
      document.title
    );
    await expect(documentPage.editor.getSaveButton()).toHaveCount(0);
  });

  test("should reject a title that is too short", async ({ page }) => {
    const { documentPage } = await openDocument(page);

    await documentPage.editor.editTitle("ab");
    await documentPage.editor.save();
    await expect(documentPage.editor.getTitleInput()).toHaveAttribute(
      "aria-invalid",
      "true"
    );
  });
});
