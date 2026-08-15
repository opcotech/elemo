import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers";
import { WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";

test.describe("@operational Namespace and project Work surfaces", () => {
  let workspace: OwnerWorkspace;
  let workPage: WorkPage;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Operational Work",
    });
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    workPage = new WorkPage(page);
  });

  test("opens namespace Work, owns projection URL state, and keeps empty fixtures scoped", async ({
    page,
  }) => {
    await workPage.gotoNamespaceWork(workspace.namespaceId);
    await workPage.waitForLoad(`${workspace.namespaceName} / Work`);
    await expect(workPage.surface.getEmptyState()).toBeVisible();
    await expect(
      page.getByText("Live issues with illustrative extras")
    ).toBeVisible();

    await workPage.surface.selectLayout("Table");
    await expect(page).toHaveURL(/layout=table/);
    await expect(workPage.surface.getLayoutButton("Table")).toHaveAttribute(
      "aria-pressed",
      "true"
    );

    await workPage.surface.selectLayout("List");
    await expect(page).toHaveURL(/layout=list/);
    await expect(workPage.surface.getEmptyState()).toBeVisible();
  });

  test("opens project Work from contextual navigation", async ({ page }) => {
    await page.goto(
      `/namespaces/${workspace.namespaceId}/projects/${workspace.projectId}`
    );
    await expect(
      page.getByRole("heading", { name: workspace.projectName })
    ).toBeVisible();

    await page.getByRole("link", { name: "Work", exact: true }).click();
    await expect(page).toHaveURL(
      new RegExp(
        `/namespaces/${workspace.namespaceId}/projects/${workspace.projectId}/work`
      )
    );
    await workPage.waitForLoad(`${workspace.projectName} / Work`);
    await expect(workPage.surface.getEmptyState()).toBeVisible();

    await workPage.surface.selectLayout("Board");
    await expect(page).toHaveURL(/layout=board/);
  });
});
