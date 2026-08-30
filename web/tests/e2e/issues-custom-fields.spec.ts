import { createCustomField, createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.custom-fields Issue Custom Fields E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issue Custom Fields",
    });
  });

  test("should edit a typed custom field on an issue", async ({ page }) => {
    const suffix = getRandomString(8).toLowerCase();
    const field = await createCustomField(workspace.client, {
      key: `release_note_${suffix}`,
      name: "Release note",
      kind: "text",
      scope_id: workspace.projectId,
      scope_type: "Project",
      target_type: "Issue",
      schema: { kind: "text" },
    });
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Custom field issue ${suffix}`,
      kind: "task",
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workItem = new WorkItemPage(page);
    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      issue.key
    );
    await workItem.waitForLoad();
    await workItem.customFields.waitForLoad();
    await expect(workItem.customFields.getSectionContainer()).toContainText(
      field.name
    );

    await workItem.customFields.fillTextField(field.name, "ready to ship");
    await waitForSuccessToast(page, "Custom field updated");
    await expect(workItem.customFields.getFieldInput(field.name)).toHaveValue(
      "ready to ship"
    );
  });
});
