import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.work-surface Issue Work Surface E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Surface",
    });
  });

  const seedIssue = async (title?: string) =>
    createIssue(workspace.client, workspace.projectId, {
      title: title ?? `Surface ${getRandomString(8)}`,
    });

  const loginOwner = async (page: Parameters<typeof loginUser>[0]) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
  };

  const openProjectWork = async (page: Parameters<typeof loginUser>[0]) => {
    await loginOwner(page);
    const workPage = new WorkPage(page);
    await workPage.gotoProjectWork(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await workPage.waitForLoad(`${workspace.projectName} / Work`);
    return workPage;
  };

  test("should switch project Work layouts in the URL", async ({ page }) => {
    const issue = await seedIssue();
    const workPage = await openProjectWork(page);

    await workPage.surface.selectLayout("List");
    await expect(page).toHaveURL(/layout=list/);
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();
    await expect(workPage.surface.getLayoutButton("List")).toHaveAttribute(
      "aria-pressed",
      "true"
    );

    await workPage.surface.selectLayout("Table");
    await expect(page).toHaveURL(/layout=table/);
    await expect(workPage.surface.getLayoutButton("Table")).toHaveAttribute(
      "aria-pressed",
      "true"
    );
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();

    await workPage.surface.selectLayout("Board");
    await expect(page).toHaveURL(/layout=board/);
    await expect(workPage.surface.getLayoutButton("Board")).toHaveAttribute(
      "aria-pressed",
      "true"
    );
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();

    await workPage.surface.selectLayout("Timeline");
    await expect(page).toHaveURL(/layout=timeline/);
    await expect(workPage.surface.getLayoutButton("Timeline")).toHaveAttribute(
      "aria-pressed",
      "true"
    );
  });

  test("should show seeded issues on namespace Work and switch layouts", async ({
    page,
  }) => {
    const issue = await seedIssue(`Namespace ${getRandomString(8)}`);
    await loginOwner(page);

    const workPage = new WorkPage(page);
    await workPage.gotoNamespaceWork(
      workspace.organizationSlug,
      workspace.namespaceSlug
    );
    await workPage.waitForLoad(`${workspace.namespaceName} / Work`);
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();

    await workPage.surface.selectLayout("List");
    await expect(page).toHaveURL(/layout=list/);
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();

    await workPage.surface.selectLayout("Table");
    await expect(page).toHaveURL(/layout=table/);
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();
  });

  test("should filter work by title", async ({ page }) => {
    const matching = await seedIssue(`FilterTitle${getRandomString(8)}`);
    const other = await seedIssue(`OtherTitle${getRandomString(8)}`);
    const workPage = await openProjectWork(page);
    await workPage.surface.selectLayout("List");

    await workPage.surface.fillFilter(matching.title);
    await page.keyboard.press("Escape");
    await expect(page).toHaveURL(
      new RegExp(`filter=${encodeURIComponent(matching.title)}`)
    );
    await expect(workPage.surface.getWorkKeyLink(matching.key)).toBeVisible();
    await expect(workPage.surface.getWorkKeyLink(other.key)).toHaveCount(0);

    await workPage.surface.clearFilter();
    await expect(page).not.toHaveURL(/filter=/);
    await expect(workPage.surface.getWorkKeyLink(other.key)).toBeVisible();
  });

  test("should filter work by key", async ({ page }) => {
    const matching = await seedIssue();
    const other = await seedIssue();
    const workPage = await openProjectWork(page);
    await workPage.surface.selectLayout("List");

    await workPage.surface.fillFilter(matching.key);
    await page.keyboard.press("Escape");
    await expect(page).toHaveURL(
      new RegExp(`filter=${encodeURIComponent(matching.key)}`)
    );
    await expect(workPage.surface.getWorkKeyLink(matching.key)).toBeVisible();
    await expect(workPage.surface.getWorkKeyLink(other.key)).toHaveCount(0);
  });

  test("should update group, sort, and display chips in the URL", async ({
    page,
  }) => {
    await seedIssue();
    const workPage = await openProjectWork(page);

    await workPage.surface.selectGroup("Priority");
    await expect(page).toHaveURL(/group=priority/);
    await expect(workPage.surface.getGroupButton()).toContainText("Priority");

    await workPage.surface.selectSort("Due date");
    await expect(page).toHaveURL(/sort=dueDate(%3A|:)asc/);
    await expect(workPage.surface.getSortButton()).toContainText("Due date");

    await workPage.surface.selectDisplay("Compact");
    await expect(page).toHaveURL(/display=compact/);
    await expect(workPage.surface.getDisplayButton()).toContainText("Compact");
  });
});
