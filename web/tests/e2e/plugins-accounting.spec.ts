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

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../.."
);
const timeTrackingZip = path.join(
  repoRoot,
  "build/plugins/com.elemo.timetracking.zip"
);
const accountingZip = path.join(
  repoRoot,
  "build/plugins/com.elemo.accounting.zip"
);

test.describe("@plugins.accounting Accounting plugin", () => {
  test.skip(
    !existsSync(timeTrackingZip) || !existsSync(accountingZip),
    "plugin zips missing; run make plugins"
  );

  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Plugin Accounting",
    });
  });

  test("installs, binds time source, assigns a budget, and counts logged time", async ({
    page,
  }) => {
    test.setTimeout(90_000);
    const suffix = getRandomString(8);
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Accounting ${suffix}`,
      kind: "task",
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    await page.goto("/settings/plugins");
    await expect(
      page
        .getByRole("button", { name: "Upgrade" })
        .first()
        .or(page.getByText("No plugins installed"))
    ).toBeVisible({ timeout: 30_000 });
    const timeTrackingPlugin = page
      .locator("li")
      .filter({ hasText: "com.elemo.timetracking" });
    if ((await timeTrackingPlugin.count()) > 0) {
      await timeTrackingPlugin.getByRole("button", { name: "Upgrade" }).click();
      await page
        .locator('input[type="file"]')
        .last()
        .setInputFiles(timeTrackingZip);
      await waitForSuccessToast(page, "Plugin upgraded", { timeout: 30_000 });
    } else {
      await page.getByRole("button", { name: "Install package" }).click();
      await page
        .locator('input[type="file"]')
        .first()
        .setInputFiles(timeTrackingZip);
      await waitForSuccessToast(page, "Plugin installed", { timeout: 30_000 });
    }
    const accountingPlugin = page
      .locator("li")
      .filter({ hasText: "com.elemo.accounting" });
    if ((await accountingPlugin.count()) > 0) {
      await accountingPlugin.getByRole("button", { name: "Upgrade" }).click();
      await page
        .locator('input[type="file"]')
        .last()
        .setInputFiles(accountingZip);
      await waitForSuccessToast(page, "Plugin upgraded", { timeout: 30_000 });
    } else {
      await page.getByRole("button", { name: "Install package" }).click();
      await page
        .locator('input[type="file"]')
        .first()
        .setInputFiles(accountingZip);
      await waitForSuccessToast(page, "Plugin installed", { timeout: 30_000 });
    }
    await expect(page.getByText("com.elemo.timetracking")).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.getByText("com.elemo.accounting")).toBeVisible({
      timeout: 30_000,
    });

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/plugins`
    );
    await page
      .locator("li")
      .filter({ hasText: "com.elemo.timetracking" })
      .getByRole("button", { name: "Enable" })
      .click();
    await waitForSuccessToast(page, "Plugin enabled");
    await page
      .locator("li")
      .filter({ hasText: "com.elemo.accounting" })
      .getByRole("button", { name: "Enable" })
      .click();
    await waitForSuccessToast(page, "Plugin enabled");

    const binding = page.getByTestId("plugin-binding-time_source");
    await expect(binding).toBeVisible({ timeout: 15_000 });
    await binding.getByRole("combobox").first().click();
    await page.getByRole("option", { name: "Time Tracking" }).click();
    await binding.getByRole("combobox").nth(1).click();
    await page.getByRole("option", { name: "TimeEntry" }).click();
    await binding.getByTestId("plugin-binding-save").click();
    await waitForSuccessToast(page, "Plugin config saved");

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/accounts`
    );
    await expect(page.getByTestId("accounting-accounts")).toBeVisible({
      timeout: 15_000,
    });
    await page.getByRole("button", { name: "New account" }).first().click();
    await page.locator("#account-code").fill(`ACC-${suffix.slice(0, 4)}`);
    await page.locator("#account-name").fill("Consulting");
    await page
      .locator("#account-description")
      .fill("Client consulting and delivery work");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Create" })
      .click();
    await waitForSuccessToast(page, "Account created");

    const accountRow = page.getByTestId("accounting-account-row");
    await expect(accountRow).toContainText("Consulting");
    await expect(accountRow).toContainText(
      "Client consulting and delivery work"
    );
    await accountRow.getByRole("button", { name: "Account actions" }).click();
    await page.getByRole("menuitem", { name: "Edit" }).click();
    await page.locator("#account-name").fill("Consulting desk");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save" })
      .click();
    await waitForSuccessToast(page, "Account updated");
    await expect(accountRow).toContainText("Consulting desk");

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/budgets`
    );
    await expect(page.getByTestId("accounting-budgets")).toBeVisible({
      timeout: 15_000,
    });
    await page.getByRole("button", { name: "New budget" }).first().click();
    await page.locator("#budget-name").fill("Q1");
    await page.locator("#budget-hours").fill("40");
    await page.locator("#budget-threshold").fill("80");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Create" })
      .click();
    await waitForSuccessToast(page, "Budget created");
    const budgetRow = page.getByTestId("accounting-budget-row");
    await expect(budgetRow).toContainText("Q1");
    await expect(budgetRow).toContainText("80%");

    await budgetRow.getByRole("button", { name: "Budget actions" }).click();
    await page.getByRole("menuitem", { name: "Edit" }).click();
    await page.locator("#budget-name").fill("Q1 retainer");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save" })
      .click();
    await waitForSuccessToast(page, "Budget updated");
    await expect(budgetRow).toContainText("Q1 retainer");

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/namespaces/${workspace.namespaceSlug}/projects/${workspace.projectKey}`
    );
    const projectBudget = page.getByTestId("accounting-project-settings");
    await expect(projectBudget).toBeVisible({ timeout: 15_000 });
    await projectBudget
      .getByRole("button", { name: "Assigned budget" })
      .click();
    await expect(page.getByRole("option", { name: /Q1 retainer/ })).toBeVisible(
      {
        timeout: 15_000,
      }
    );
    await page.getByRole("option", { name: /Q1 retainer/ }).click();
    await waitForSuccessToast(page, "Budget assigned");

    const workItem = new WorkItemPage(page);
    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      issue.key
    );
    await workItem.waitForLoad();

    const issueBudget = page.getByTestId("accounting-issue-sidebar");
    await expect(issueBudget).toBeVisible({ timeout: 15_000 });
    await expect(
      issueBudget.getByRole("button", { name: "Assigned budget" })
    ).toContainText("Use project budget");
    await expect(
      issueBudget.getByRole("button", { name: "Clear" })
    ).toHaveCount(0);
    await expect(page.getByRole("tab", { name: "Billing" })).toHaveCount(0);

    const sidebar = page.getByTestId("timetracking-sidebar");
    await expect(sidebar).toBeVisible({ timeout: 15_000 });
    await sidebar.getByRole("button", { name: "Log time" }).click();
    await page.getByTestId("timetracking-hours").fill("1");
    await page.getByTestId("timetracking-minutes").fill("0");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save" })
      .click();
    await waitForSuccessToast(page, "Time logged");

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/namespaces/${workspace.namespaceSlug}/projects/${workspace.projectKey}`
    );
    await expect(page.getByTestId("accounting-budget-remaining")).toContainText(
      "Used 1h of 40h",
      { timeout: 15_000 }
    );

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/report`
    );
    const report = page.getByTestId("accounting-report");
    await expect(report).toBeVisible({ timeout: 15_000 });
    await expect(report.getByTestId("accounting-report-summary")).toContainText(
      /Allocated\s*40h\s*Used\s*1h\s*Remaining\s*39h\s*Utilization\s*3%/
    );
    await expect(report.getByTestId("accounting-report-filters")).toBeVisible();
    await expect(
      report.getByTestId("accounting-usage-trend-chart")
    ).toHaveCount(0);
    await expect(
      page.getByText("This plugin page failed to load.")
    ).toHaveCount(0);
    await expect(
      report.getByTestId("accounting-threshold-warning")
    ).toHaveCount(0);
    await report.getByRole("button", { name: "Filter by account" }).click();
    await page.getByRole("option", { name: /Consulting desk/ }).click();
    await expect(report).toBeVisible();
    await expect(
      page.getByText("This plugin page failed to load.")
    ).toHaveCount(0);
    await report.getByRole("button", { name: "Filter by account" }).click();
    await page.getByRole("option", { name: "All accounts" }).click();
    await expect(report).toBeVisible();
    await expect(
      page.getByText("This plugin page failed to load.")
    ).toHaveCount(0);
    const reportRow = report.getByTestId("accounting-report-account-row");
    await expect(reportRow).toContainText("Consulting desk");
    await reportRow
      .getByRole("button", { name: /Expand .*Consulting desk usage/ })
      .click();
    const timelogs = report.getByTestId("accounting-report-timelogs");
    await expect(timelogs).toBeVisible();
    await expect(timelogs).toContainText(issue.title);
    await expect(timelogs).toContainText("1h");
    await expect(
      timelogs.getByRole("link", { name: new RegExp(issue.title) })
    ).toHaveAttribute(
      "href",
      `/work/${workspace.organizationSlug}/${workspace.namespaceSlug}/${issue.key}`
    );

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/budgets`
    );
    await expect(page.getByTestId("accounting-budget-row")).toBeVisible({
      timeout: 15_000,
    });
    await page
      .getByTestId("accounting-budget-row")
      .getByRole("button", { name: "Budget actions" })
      .click();
    await page.getByRole("menuitem", { name: "Edit" }).click();
    await page.locator("#budget-threshold").fill("1");
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save" })
      .click();
    await waitForSuccessToast(page, "Budget updated");

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/report`
    );
    await expect(report).toBeVisible({ timeout: 15_000 });
    await expect(
      page.getByText("This plugin page failed to load.")
    ).toHaveCount(0);
    await expect(
      report.getByTestId("accounting-threshold-warning").first()
    ).toBeVisible();

    await page.goto(
      `/settings/organizations/${workspace.organizationSlug}/namespaces/${workspace.namespaceSlug}/projects/${workspace.projectKey}`
    );
    const assigned = page.getByTestId("accounting-project-settings");
    await expect(assigned).toBeVisible({ timeout: 15_000 });
    await assigned.getByRole("button", { name: "Assigned budget" }).click();
    await page.getByRole("option", { name: "No budget" }).click();
    await waitForSuccessToast(page, "Budget cleared");
    await expect(
      assigned.getByTestId("accounting-budget-remaining")
    ).toHaveCount(0);

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/accounts`
    );
    await expect(page.getByTestId("accounting-account-row")).toBeVisible({
      timeout: 15_000,
    });
    await page
      .getByTestId("accounting-account-row")
      .getByRole("button", { name: "Account actions" })
      .click();
    await page.getByRole("menuitem", { name: "Delete" }).click();
    await expect(page.getByRole("alertdialog")).toContainText(
      "Account has budgets"
    );
    await page.getByRole("button", { name: "Cancel" }).click();

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/budgets`
    );
    await expect(page.getByTestId("accounting-budget-row")).toBeVisible({
      timeout: 15_000,
    });
    await page
      .getByTestId("accounting-budget-row")
      .getByRole("button", { name: "Budget actions" })
      .click();
    await page.getByRole("menuitem", { name: "Delete" }).click();
    await page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Delete" })
      .click();
    await waitForSuccessToast(page, "Budget deleted");

    await page.goto(
      `/organizations/${workspace.organizationSlug}/plugins/com.elemo.accounting/accounts`
    );
    await page
      .getByTestId("accounting-account-row")
      .getByRole("button", { name: "Account actions" })
      .click();
    await page.getByRole("menuitem", { name: "Delete" }).click();
    await page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Delete" })
      .click();
    await waitForSuccessToast(page, "Account deleted");
    await expect(page.getByTestId("accounting-account-row")).toHaveCount(0);
  });
});
