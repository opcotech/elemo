import { createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.detail Document Detail E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Detail",
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
        title: name ?? `Detail ${getRandomString(8)}`,
        excerpt: `Detail excerpt ${getRandomString(8)}`,
        content: `# Overview\n\nBody ${getRandomString(8)}`,
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

  test("should show the document title, excerpt, and content", async ({
    page,
  }) => {
    const name = `Detail title ${getRandomString(8)}`;
    const { documentPage, document } = await openDocument(page, name);

    await expect(page).toHaveURL(new RegExp(`/documents/${document.id}`));
    await expect(documentPage.editor.getTitleInput()).toHaveValue(name);
    await expect(documentPage.editor.getExcerptInput()).toHaveValue(
      document.excerpt ?? ""
    );
    await expect(documentPage.editor.getContentEditor()).toContainText(
      "Overview"
    );
  });

  test("should copy the document link", async ({ page }) => {
    const { documentPage } = await openDocument(page);

    await page.evaluate(() => {
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: {
          writeText: () => undefined,
        },
      });
    });

    await documentPage.editor.clickMoreAction("Copy link");
    await waitForSuccessToast(page, "Copied");
  });

  test("should expose more actions including delete", async ({ page }) => {
    const { documentPage } = await openDocument(page);

    await documentPage.editor.openMoreActions();
    await expect(
      page.getByRole("menuitem", { name: "View relationships" })
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Copy link" })
    ).toBeVisible();
    await expect(page.getByRole("menuitem", { name: "Delete" })).toBeVisible();
  });
});
