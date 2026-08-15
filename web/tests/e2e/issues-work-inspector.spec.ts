import { createIssue, updateIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForAnimations } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage, WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.work-inspector Issue Work Inspector E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Inspector",
    });
  });

  const seedAssignedIssue = async () => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Inspect ${getRandomString(8)}`,
    });
    return updateIssue(workspace.client, issue.id, {
      assignees: [workspace.owner.id],
    });
  };

  const openProjectWork = async (page: Parameters<typeof loginUser>[0]) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const workPage = new WorkPage(page);
    await workPage.gotoProjectWork(workspace.namespaceId, workspace.projectId);
    await workPage.waitForLoad(`${workspace.projectName} / Work`);
    await workPage.surface.selectLayout("List");
    return workPage;
  };

  const selectedPattern = (key: string) =>
    new RegExp(`selected=work(%3A|:)${key}`);

  test("should inspect an issue and set selected in the URL", async ({
    page,
  }) => {
    const issue = await seedAssignedIssue();
    const workPage = await openProjectWork(page);

    await workPage.surface.inspect(issue.key, issue.title);
    await workPage.inspector.waitForLoad(issue.key);
    await expect(page).toHaveURL(selectedPattern(issue.key));
    await expect(workPage.inspector.getInspector(issue.key)).toBeVisible();
  });

  test("should close the inspector and clear selection", async ({ page }) => {
    const issue = await seedAssignedIssue();
    const workPage = await openProjectWork(page);

    await workPage.surface.inspect(issue.key, issue.title);
    await workPage.inspector.waitForLoad(issue.key);
    await expect(page).toHaveURL(selectedPattern(issue.key));

    await workPage.inspector.close(issue.key);
    await expect(workPage.inspector.getInspector(issue.key)).toHaveCount(0);
    await expect(page).not.toHaveURL(/selected=/);
  });

  test("should open the full issue page from the inspector", async ({
    page,
  }) => {
    const issue = await seedAssignedIssue();
    const workPage = await openProjectWork(page);

    await workPage.surface.inspect(issue.key, issue.title);
    await workPage.inspector.waitForLoad(issue.key);
    await workPage.inspector.openFullPage(issue.key);

    await expect(page).toHaveURL(
      new RegExp(`/work/${workspace.namespaceId}/${issue.key}`)
    );
    const workItem = new WorkItemPage(page);
    await workItem.waitForLoad();
    await expect(workItem.getTitleButton()).toContainText(issue.title);
  });

  test("should use a sheet inspector overlay on tablet", async ({ page }) => {
    const issue = await seedAssignedIssue();
    await page.setViewportSize({ width: 800, height: 900 });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const workPage = new WorkPage(page);
    await workPage.gotoMyWork();
    await workPage.waitForLoad("My Work");
    await workPage.surface.selectLayout("List");

    await workPage.surface.inspect(issue.key, issue.title);
    await expect(page).toHaveURL(selectedPattern(issue.key));

    const overlay = workPage.inspector.getOverlay();
    await expect(overlay).toBeVisible();
    await waitForAnimations(overlay);
    await expect(workPage.inspector.getInspector(issue.key)).toBeVisible();

    await workPage.inspector.close(issue.key);
    await expect(overlay).not.toBeVisible();
    await expect(page).not.toHaveURL(/selected=/);
  });
});
