import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

const pluginZip = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../build/plugins/com.elemo.timetracking.zip"
);

test.describe("@plugins.timetracking Time Tracking plugin", () => {
  test.skip(
    !existsSync(pluginZip),
    "build/plugins/com.elemo.timetracking.zip is missing; run make plugins.timetracking"
  );

  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Plugin Time Tracking",
    });
  });

  test("installs, logs time with description, edits, moves, and deletes", async ({
    page,
  }) => {
    const suffix = getRandomString(8);
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Time tracking ${suffix}`,
      kind: "task",
    });
    const other = await createIssue(workspace.client, workspace.projectId, {
      title: `Time tracking other ${suffix}`,
      kind: "task",
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    await page.goto("/settings/plugins");
    await page.getByRole("button", { name: "Install package" }).click();
    await page.locator('input[type="file"]').first().setInputFiles(pluginZip);
    await waitForSuccessToast(page, "Plugin installed", { timeout: 30_000 });
    await expect(page.getByText("com.elemo.timetracking")).toBeVisible({
      timeout: 30_000,
    });

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/plugins`
    );
    await page.getByRole("button", { name: "Enable" }).click();
    await waitForSuccessToast(page, "Plugin enabled");
    await expect(page.getByText("Enabled")).toBeVisible();

    const workItem = new WorkItemPage(page);
    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      issue.key
    );
    await workItem.waitForLoad();

    const sidebar = page.getByTestId("timetracking-sidebar");
    await expect(sidebar).toBeVisible({ timeout: 15_000 });
    await expect(sidebar.getByRole("button", { name: "Report" })).toHaveCount(
      0
    );
    await expect(sidebar.getByRole("link", { name: "Report" })).toHaveCount(0);
    await sidebar
      .getByTestId("timetracking-description")
      .fill("pair programming");
    await sidebar.getByRole("button", { name: "Start" }).click();
    await waitForSuccessToast(page, "Timer started");
    await page.waitForTimeout(1100);
    await sidebar.getByRole("button", { name: "Stop" }).click();
    await waitForSuccessToast(page, "Time logged");

    const loggedTab = page.getByRole("tab", { name: "Logged time" });
    await expect(loggedTab).toBeVisible({ timeout: 15_000 });
    await loggedTab.click();
    const logged = page.getByTestId("timetracking-logged-time");
    await expect(logged).toBeVisible({ timeout: 15_000 });
    const entry = logged.getByTestId("timetracking-entry");
    await expect(entry).toBeVisible({ timeout: 15_000 });
    await expect(page.getByTestId("timetracking-entry-note")).toContainText(
      "pair programming"
    );

    await entry.getByRole("button", { name: "Time entry actions" }).click();
    await page.getByRole("menuitem", { name: "Edit" }).click();
    await page.getByTestId("timetracking-edit-note").fill("updated note");
    await page.getByRole("button", { name: "Save" }).click();
    await waitForSuccessToast(page, "Time entry updated");
    await expect(page.getByTestId("timetracking-entry-note")).toContainText(
      "updated note"
    );

    await logged.getByTestId("timetracking-report-link").click();
    const filteredReport = page.getByTestId("timetracking-report");
    await expect(filteredReport).toBeVisible({ timeout: 15_000 });
    expect(new URL(page.url()).searchParams.get("workItem")).toBe(issue.id);
    await expect(
      filteredReport.getByTestId("timetracking-report-group")
    ).toHaveCount(1);
    await expect(filteredReport.getByText(issue.key)).toBeVisible();
    await expect(
      filteredReport.getByTestId("timetracking-entry-note")
    ).toContainText("updated note");

    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      issue.key
    );
    await workItem.waitForLoad();
    await expect(page.getByRole("tab", { name: "Logged time" })).toBeVisible({
      timeout: 15_000,
    });
    await page.getByRole("tab", { name: "Logged time" }).click();
    const entryToMove = page
      .getByTestId("timetracking-logged-time")
      .getByTestId("timetracking-entry");
    await expect(entryToMove).toBeVisible({ timeout: 15_000 });

    await entryToMove
      .getByRole("button", { name: "Time entry actions" })
      .click();
    await page.getByRole("menuitem", { name: "Move" }).click();
    await page.getByRole("button", { name: "Select a work item" }).click();
    await page.getByText(other.key, { exact: false }).click();
    await page.getByRole("button", { name: "Move" }).click();
    await waitForSuccessToast(page, "Time entry moved");
    await expect(logged.getByTestId("timetracking-entry")).toHaveCount(0, {
      timeout: 15_000,
    });

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.timetracking/report`
    );
    const report = page.getByTestId("timetracking-report");
    await expect(report).toBeVisible({ timeout: 15_000 });
    await expect(report.getByTestId("timetracking-report-group")).toBeVisible({
      timeout: 15_000,
    });
    await expect(report.getByText(other.key)).toBeVisible();
    await report.getByRole("searchbox").fill("updated note");
    await expect(report.getByTestId("timetracking-report-group")).toHaveCount(
      1
    );
    await report.getByRole("searchbox").fill("zzzz-no-match");
    await expect(report.getByTestId("timetracking-report-group")).toHaveCount(
      0
    );

    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      other.key
    );
    await workItem.waitForLoad();
    await expect(page.getByTestId("timetracking-sidebar")).toBeVisible({
      timeout: 15_000,
    });
    await page.getByRole("tab", { name: "Logged time" }).click();
    const movedEntry = page
      .getByTestId("timetracking-logged-time")
      .getByTestId("timetracking-entry");
    await expect(movedEntry).toBeVisible({ timeout: 15_000 });
    await movedEntry
      .getByRole("button", { name: "Time entry actions" })
      .click();
    await page.getByRole("menuitem", { name: "Delete" }).click();
    await page.getByRole("button", { name: "Delete" }).click();
    await waitForSuccessToast(page, "Time entry deleted");
    await expect(
      page
        .getByTestId("timetracking-logged-time")
        .getByTestId("timetracking-entry")
    ).toHaveCount(0, { timeout: 15_000 });

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/plugins`
    );
    await page.getByRole("button", { name: "Disable" }).click();
    await waitForSuccessToast(page, "Plugin disabled");
    await expect(page.getByText("Disabled")).toBeVisible();

    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      issue.key
    );
    await workItem.waitForLoad();
    await expect(page.getByTestId("timetracking-sidebar")).toHaveCount(0);
  });
});
