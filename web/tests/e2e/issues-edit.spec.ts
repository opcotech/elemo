import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

import type { Issue } from "@/lib/api/types";

test.describe("@issues.edit Issue Edit E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Edit",
    });
  });

  const ownerName = () =>
    `${workspace.owner.first_name ?? ""} ${workspace.owner.last_name ?? ""}`.trim();

  const seedIssue = async (title?: string) =>
    createIssue(workspace.client, workspace.projectId, {
      title: title ?? `Edit ${getRandomString(8)}`,
    });

  const openIssue = async (
    page: Parameters<typeof loginUser>[0],
    issue: Issue
  ) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const workItem = new WorkItemPage(page);
    await workItem.goto(workspace.namespaceId, issue.key);
    await workItem.waitForLoad();
    return workItem;
  };

  test("should update the inline title", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);
    const updatedTitle = `Title ${getRandomString(8)}`;

    await workItem.editTitle(updatedTitle);
    await waitForSuccessToast(page, "Title updated");
    await expect(workItem.getTitleButton()).toContainText(updatedTitle);
  });

  test("should save a description", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);
    const description = `Description ${getRandomString(8)}`;

    await workItem.startDescriptionEdit();
    const editor = workItem
      .getDescriptionSection()
      .getByLabel("Issue description");
    await expect(editor).toBeVisible();
    await editor.click();
    await editor.pressSequentially(description, { delay: 10 });
    await workItem.saveDescription();
    await waitForSuccessToast(page, "Description updated");
    await expect(workItem.getDescriptionSection()).toContainText(description);
  });

  test("should update kind", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.selectKind("Story");
    await waitForSuccessToast(page, "Kind set to Story");
    await expect(workItem.details.getKindSelect()).toContainText("Story");
  });

  test("should update status", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.selectStatus("In progress");
    await waitForSuccessToast(page, "Status set to In progress");
    await expect(workItem.details.getStatusSelect()).toContainText(
      "In progress"
    );
  });

  test("should update resolution", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.selectResolution("Fixed");
    await waitForSuccessToast(page, "Resolution set to Fixed");
    await expect(workItem.details.getResolutionSelect()).toContainText("Fixed");
  });

  test("should update priority", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.selectPriority("High");
    await waitForSuccessToast(page, "Priority set to High");
    await expect(workItem.details.getPrioritySelect()).toContainText("High");
  });

  test("should set and clear the start date", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.getStartDatePicker().click();
    const calendar = page.locator("[data-slot='calendar']");
    await expect(calendar).toBeVisible();
    await calendar
      .locator(`[data-day="${new Date().toLocaleDateString()}"]`)
      .click();
    await waitForSuccessToast(page, "Start date updated");
    await expect(workItem.details.getClearStartDateButton()).toBeVisible();

    await workItem.details.getClearStartDateButton().click();
    await waitForSuccessToast(page, "Start date cleared");
    await expect(workItem.details.getClearStartDateButton()).toHaveCount(0);
  });

  test("should set and clear the due date", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);

    await workItem.details.getDueDatePicker().click();
    const calendar = page.locator("[data-slot='calendar']");
    await expect(calendar).toBeVisible();
    await calendar
      .locator(`[data-day="${new Date().toLocaleDateString()}"]`)
      .click();
    await waitForSuccessToast(page, "Due date updated");
    await expect(workItem.details.getClearDueDateButton()).toBeVisible();

    await workItem.details.getClearDueDateButton().click();
    await waitForSuccessToast(page, "Due date cleared");
    await expect(workItem.details.getClearDueDateButton()).toHaveCount(0);
  });

  test("should set an org member as an assignee", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);
    const name = ownerName();

    await workItem.details.openAssignees();
    await expect(
      page.locator('[data-slot="command-item"]').filter({ hasText: name })
    ).toBeVisible();
    await workItem.details.selectPersonOption(name);
    await page.keyboard.press("Escape");
    await waitForSuccessToast(page, "Assignees updated");
    await expect(workItem.details.getAssigneesSelect()).toContainText(name);
  });

  test("should set an org member as a reviewer", async ({ page }) => {
    const issue = await seedIssue();
    const workItem = await openIssue(page, issue);
    const name = ownerName();

    await workItem.details.openReviewers();
    await expect(
      page.locator('[data-slot="command-item"]').filter({ hasText: name })
    ).toBeVisible();
    await workItem.details.selectPersonOption(name);
    await page.keyboard.press("Escape");
    await waitForSuccessToast(page, "Reviewers updated");
    await expect(workItem.details.getReviewersSelect()).toContainText(name);
  });

  test("should set and clear the parent issue", async ({ page }) => {
    const parent = await seedIssue(`Parent ${getRandomString(8)}`);
    const child = await seedIssue(`Child ${getRandomString(8)}`);
    const workItem = await openIssue(page, child);

    await workItem.details.selectParent(parent.key);
    await waitForSuccessToast(page, `Parent set to ${parent.key}`);
    await expect(workItem.details.getParentSelect()).toContainText(parent.key);

    await workItem.details.clearParent();
    await waitForSuccessToast(page, "Parent cleared");
    await expect(workItem.details.getParentSelect()).toContainText("None");
  });
});
