import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@issues.links Issue Links E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Links",
    });
  });

  const openIssue = async (page: Parameters<typeof loginUser>[0]) => {
    const issue = await createIssue(workspace.client, workspace.projectId, {
      title: `Links ${getRandomString(8)}`,
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
    await workItem.links.waitForLoad();
    return workItem;
  };

  test("should add an external link", async ({ page }) => {
    const workItem = await openIssue(page);
    const label = `Docs ${getRandomString(6)}`;

    await workItem.links.addLink({
      url: "https://example.com/issue-link",
      label,
    });
    await waitForSuccessToast(page, "Link added");
    await expect(workItem.links.getLink(label)).toBeVisible();
  });

  test("should edit an external link", async ({ page }) => {
    const workItem = await openIssue(page);
    const label = `Original ${getRandomString(6)}`;
    const updatedLabel = `Updated ${getRandomString(6)}`;

    await workItem.links.addLink({
      url: "https://example.com/original",
      label,
    });
    await waitForSuccessToast(page, "Link added");

    await workItem.links.editLink(label, {
      url: "https://example.com/updated",
      label: updatedLabel,
    });
    await waitForSuccessToast(page, "Link updated");
    await expect(workItem.links.getLink(updatedLabel)).toBeVisible();
    await expect(workItem.links.getLink(label)).toHaveCount(0);
  });

  test("should remove an external link", async ({ page }) => {
    const workItem = await openIssue(page);
    const label = `Remove ${getRandomString(6)}`;

    await workItem.links.addLink({
      url: "https://example.com/remove",
      label,
    });
    await waitForSuccessToast(page, "Link added");
    await expect(workItem.links.getLink(label)).toBeVisible();

    await workItem.links.removeLink(label);
    await waitForSuccessToast(page, "Link removed");
    await expect(workItem.links.getLink(label)).toHaveCount(0);
    await expect(workItem.links.getSectionContainer()).toContainText(
      "No links"
    );
  });
});
