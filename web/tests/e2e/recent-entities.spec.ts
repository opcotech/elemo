import type { Page } from "@playwright/test";

import { createIssue, createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers";
import { DocumentPage, WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

const NAVIGATION_STORAGE_KEY = "elemo_navigation_context";

async function expectRecentTypePersisted(page: Page, type: string) {
  await expect
    .poll(async () =>
      page.evaluate(
        ({ key, entityType }) => {
          const raw = window.localStorage.getItem(key);
          if (!raw) {
            return false;
          }
          try {
            const parsed = JSON.parse(raw) as {
              recentEntities?: { type?: string }[];
            };
            return (
              parsed.recentEntities?.some(
                (entity) => entity.type === entityType
              ) ?? false
            );
          } catch {
            return false;
          }
        },
        { key: NAVIGATION_STORAGE_KEY, entityType: type }
      )
    )
    .toBe(true);
}

test.describe("@operational Recent sidebar entities", () => {
  let workspace: OwnerWorkspace;
  let workKey: string;
  let workTitle: string;
  let workLabel: string;
  let documentId: string;
  let documentLabel: string;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Recent Entities",
    });
    workTitle = `Preserve work context ${getRandomString(8)}`;
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: workTitle,
    });
    if (!issue.key) {
      throw new Error("Created issue is missing a key");
    }
    workKey = issue.key;
    workLabel = `${workKey} ${workTitle}`;

    documentLabel = `Work projection model ${getRandomString(8)}`;
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      { title: documentLabel }
    );
    documentId = document.id;
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await page.evaluate((key) => {
      window.localStorage.removeItem(key);
    }, NAVIGATION_STORAGE_KEY);
    await page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  test("remembers work, documents, and projects in sidebar order and supports remove", async ({
    page,
  }) => {
    const sidebar = page.locator(
      '[data-slot="sidebar"][data-variant="sidebar"]'
    );
    const workItemPage = new WorkItemPage(page);
    const documentPage = new DocumentPage(page);

    await workItemPage.goto(workspace.namespaceId, workKey);
    await workItemPage.waitForLoad();
    await expect(workItemPage.getTitleButton()).toHaveText(workTitle);
    await expect(
      sidebar.getByText("Recent Work Items", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: workLabel, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "work");

    await documentPage.goto(documentId);
    await documentPage.waitForLoad();
    await expect(documentPage.editor.getTitleInput()).toHaveValue(
      documentLabel
    );
    await expect(
      sidebar.getByText("Recent Documents", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: documentLabel, exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: workLabel, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "document");

    await page.goto(
      `/namespaces/${workspace.namespaceId}/projects/${workspace.projectId}`,
      {
        waitUntil: "domcontentloaded",
      }
    );
    await expect(
      page.getByRole("heading", { name: workspace.projectName })
    ).toBeVisible();
    await expect(
      sidebar.getByText("Recent Projects", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: workspace.projectName, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "project");

    await page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    const workSection = sidebar.getByText("Recent Work Items", { exact: true });
    const documentsSection = sidebar.getByText("Recent Documents", {
      exact: true,
    });
    const projectsSection = sidebar.getByText("Recent Projects", {
      exact: true,
    });

    await expect(workSection).toBeVisible();
    await expect(documentsSection).toBeVisible();
    await expect(projectsSection).toBeVisible();

    const workBox = await workSection.boundingBox();
    const documentsBox = await documentsSection.boundingBox();
    const projectsBox = await projectsSection.boundingBox();
    expect(workBox).toBeTruthy();
    expect(documentsBox).toBeTruthy();
    expect(projectsBox).toBeTruthy();
    expect(workBox!.y).toBeLessThan(documentsBox!.y);
    expect(documentsBox!.y).toBeLessThan(projectsBox!.y);

    const workLink = sidebar.getByRole("link", {
      name: workLabel,
      exact: true,
    });
    await workLink.hover();
    await sidebar
      .getByRole("button", { name: `Remove ${workLabel} from recents` })
      .click();
    await expect(workLink).toHaveCount(0);
    await expect(workSection).toHaveCount(0);

    await page.reload();
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: workLabel, exact: true })
    ).toHaveCount(0);
    await expect(
      sidebar.getByText("Recent Work Items", { exact: true })
    ).toHaveCount(0);
    await expect(
      sidebar.getByRole("link", { name: documentLabel, exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", {
        name: workspace.projectName,
        exact: true,
      })
    ).toBeVisible();

    await sidebar
      .getByRole("link", { name: documentLabel, exact: true })
      .click();
    await expect(page).toHaveURL(`/documents/${documentId}`);
    await documentPage.waitForLoad();
    await expect(documentPage.editor.getTitleInput()).toHaveValue(
      documentLabel
    );
  });
});
