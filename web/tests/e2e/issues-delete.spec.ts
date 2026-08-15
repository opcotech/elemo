import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage, WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.delete Issue Delete E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Delete",
    });
  });

  test("should delete an issue from more actions and return to project work", async ({
    page,
  }) => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Delete ${getRandomString(8)}`,
    });

    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workItem = new WorkItemPage(page);
    await workItem.goto(workspace.namespaceId, issue.key);
    await workItem.waitForLoad();

    await workItem.openDeleteDialog();
    await expect(
      page.getByRole("heading", { name: `Delete ${issue.key}?` })
    ).toBeVisible();
    await workItem.confirmDelete();
    await waitForSuccessToast(page, "Issue deleted");

    await expect(page).toHaveURL(
      new RegExp(
        `/namespaces/${workspace.namespaceId}/projects/${workspace.projectId}/work`
      )
    );

    const workPage = new WorkPage(page);
    await workPage.waitForLoad(`${workspace.projectName} / Work`);
    await expect(workPage.surface.getWorkKeyLink(issue.key)).toHaveCount(0);
    await expect(
      page.getByRole("button", { name: new RegExp(`^Inspect ${issue.key}:`) })
    ).toHaveCount(0);
  });
});
