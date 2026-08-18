import { createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { DocumentPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@documents.editor Document Editor E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Documents Editor",
    });
  });

  const openDocument = async (
    page: Parameters<typeof loginUser>[0],
    content: string
  ) => {
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      {
        title: `Editor ${getRandomString(8)}`,
        content,
      }
    );

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const documentPage = new DocumentPage(page);
    await documentPage.goto(document.id);
    await documentPage.waitForLoad();
    return documentPage;
  };

  test("should open the table of contents and jump to a heading", async ({
    page,
  }) => {
    const documentPage = await openDocument(
      page,
      `# Alpha heading\n\nIntro paragraph.\n\n## Beta heading\n\nDetails here.`
    );

    await documentPage.editor.openToc();
    await expect(
      documentPage.editor.getTocHeading("Alpha heading")
    ).toBeVisible();
    await expect(
      documentPage.editor.getTocHeading("Beta heading")
    ).toBeVisible();

    await documentPage.editor.getTocHeading("Beta heading").click();
    await expect(
      documentPage.editor.getTocHeading("Beta heading")
    ).toHaveAttribute("aria-current", "true");
  });

  test("should add a heading from the toolbar and show it in the TOC", async ({
    page,
  }) => {
    const documentPage = await openDocument(
      page,
      `# Title heading\n\nParagraph to convert.`
    );

    const editor = documentPage.editor.getContentEditor();
    await editor.getByText("Paragraph to convert.").click();
    await documentPage.editor.getToolbarButton("Heading 2").click();
    await documentPage.editor.openToc();
    await expect(
      documentPage.editor.getToc().getByRole("button", { name: /convert/i })
    ).toBeVisible();
  });
});
