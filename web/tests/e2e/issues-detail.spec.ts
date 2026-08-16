import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.detail Issue Detail E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Detail",
    });
  });

  const openIssue = async (
    page: Parameters<typeof loginUser>[0],
    title?: string
  ) => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: title ?? `Detail ${getRandomString(8)}`,
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workItem = new WorkItemPage(page);
    await workItem.goto(workspace.namespaceId, issue.key);
    await workItem.waitForLoad();
    return { workItem, issue };
  };

  test("should show the issue key and title on the work item page", async ({
    page,
  }) => {
    const title = `Detail title ${getRandomString(8)}`;
    const { workItem, issue } = await openIssue(page, title);

    await expect(page).toHaveURL(
      new RegExp(`/work/${workspace.namespaceId}/${issue.key}`)
    );
    await expect(
      page.getByText(issue.key, { exact: true }).first()
    ).toBeVisible();
    await expect(workItem.getTitleButton()).toContainText(title);
  });

  test("should copy the issue link", async ({ page }) => {
    const { workItem } = await openIssue(page);

    // WebKit/Firefox often reject the Clipboard API in Playwright.
    await page.evaluate(() => {
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: {
          writeText: () => undefined,
        },
      });
    });

    await workItem.copyLink();
    await waitForSuccessToast(page, "Copied");
  });

  test("should expose more actions including delete", async ({ page }) => {
    const { workItem } = await openIssue(page);

    await workItem.openMoreActions();
    await expect(
      page.getByRole("menuitem", { name: "View relationships" })
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Copy link" })
    ).toBeVisible();
    await expect(page.getByRole("menuitem", { name: "Delete" })).toBeVisible();
  });

  test("should keep comments disabled at most once", async ({ page }) => {
    await openIssue(page);

    const commentField = page.getByPlaceholder(
      "Comments are not available for this work item yet."
    );
    await expect(commentField).toHaveCount(1);
    await expect(commentField).toBeDisabled();
    await expect(page.getByRole("button", { name: "Send" })).toBeDisabled();
  });
});
