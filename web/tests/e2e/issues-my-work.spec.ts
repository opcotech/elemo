import { createIssue, updateIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { HomePage, WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.my-work Issue My Work E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues My Work",
    });
  });

  const seedAssignedIssue = async (status?: "open" | "in progress") => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Assigned ${getRandomString(8)}`,
      ...(status ? { status } : {}),
    });
    return updateIssue(workspace.client, issue.id, {
      assignees: [workspace.owner.id],
    });
  };

  test("should show an assigned issue on My Work", async ({ page }) => {
    const issue = await seedAssignedIssue();

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workPage = new WorkPage(page);
    await workPage.gotoMyWork();
    await workPage.waitForLoad("My Work");
    await workPage.surface.selectLayout("List");
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toBeVisible();
    await expect(
      page.getByRole("button", {
        name: new RegExp(`^Inspect ${issue.key}: ${issue.title}`),
      })
    ).toBeVisible();
  });

  test("should show in-progress assigned work on Home Continue working", async ({
    page,
  }) => {
    const issue = await seedAssignedIssue("in progress");

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const homePage = new HomePage(page);
    await homePage.goto();
    await homePage.waitForLoad();
    await homePage.continueWorking.waitForLoad();
    await expect(homePage.continueWorking.getWorkItem(issue.key)).toBeVisible();
  });
});
