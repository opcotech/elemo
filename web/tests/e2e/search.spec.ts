import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { searchPaletteResult, seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { HomePage, SearchPage, WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@search Search E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Search",
    });
  });

  test("opens an indexed issue from the command palette", async ({ page }) => {
    const title = `Palette search ${getRandomString(8)}`;
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title,
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const home = new HomePage(page);
    await home.waitForLoad();

    const result = await searchPaletteResult(page, title);
    await result.click();

    const workItem = new WorkItemPage(page);
    await workItem.waitForLoad();
    await expect(page).toHaveURL(
      new RegExp(`/work/${workspace.namespaceId}/${issue.key}`)
    );
    await expect(workItem.getTitleButton()).toContainText(title);
  });

  test("filters the search page by type and namespace", async ({ page }) => {
    const title = `Page search ${getRandomString(8)}`;
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title,
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const searchPage = new SearchPage(page);
    await expect(async () => {
      await searchPage.goto({
        q: title,
        type: "Issue",
        namespace_id: workspace.namespaceId,
      });
      await expect(searchPage.getTypeFilter()).toContainText("Issues");
      await expect(searchPage.getResultLink(title)).toBeVisible({
        timeout: 5_000,
      });
    }).toPass({ timeout: 30_000 });
    await searchPage.getResultLink(title).click();

    const workItem = new WorkItemPage(page);
    await workItem.waitForLoad();
    await expect(page).toHaveURL(
      new RegExp(`/work/${workspace.namespaceId}/${issue.key}`)
    );
  });
});
