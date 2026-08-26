import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace, waitForSuccessToast } from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

import type { Issue } from "@/lib/api/types";

test.describe("@issues.relations Issue Relations E2E Tests", () => {
  let workspace: OwnerWorkspace;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Relations",
    });
  });

  const seedPair = async () => {
    const source = await createIssue(workspace.client, workspace.projectId, {
      title: `Source ${getRandomString(8)}`,
    });
    const related = await createIssue(workspace.client, workspace.projectId, {
      title: `Related ${getRandomString(8)}`,
    });
    return { source, related };
  };

  const openIssue = async (
    page: Parameters<typeof loginUser>[0],
    issue: Issue
  ) => {
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
    await workItem.relations.waitForLoad();
    return workItem;
  };

  test("should add a related to relation", async ({ page }) => {
    const { source, related } = await seedPair();
    const workItem = await openIssue(page, source);

    await workItem.relations.addRelation({
      kind: "Related to",
      issue: related.key,
    });
    await waitForSuccessToast(page, "Relation added");
    await expect(workItem.relations.getRelation(related.key)).toBeVisible();
    await expect(
      workItem.relations.getRelationKindButton(related.key)
    ).toContainText("Related to");
  });

  test("should add a blocks relation", async ({ page }) => {
    const { source, related } = await seedPair();
    const workItem = await openIssue(page, source);

    await workItem.relations.addRelation({
      kind: "Blocks",
      issue: related.key,
    });
    await waitForSuccessToast(page, "Relation added");
    await expect(workItem.relations.getRelation(related.key)).toBeVisible();
    await expect(
      workItem.relations.getRelationKindButton(related.key)
    ).toContainText("Blocks");
  });

  test("should change a relation kind", async ({ page }) => {
    const { source, related } = await seedPair();
    const workItem = await openIssue(page, source);

    await workItem.relations.addRelation({
      kind: "Related to",
      issue: related.key,
    });
    await waitForSuccessToast(page, "Relation added");

    await workItem.relations.changeKind(related.key, "Blocks");
    await waitForSuccessToast(page, "Relation updated");
    await expect(
      workItem.relations.getRelationKindButton(related.key)
    ).toContainText("Blocks");
  });

  test("should remove a relation", async ({ page }) => {
    const { source, related } = await seedPair();
    const workItem = await openIssue(page, source);

    await workItem.relations.addRelation({
      kind: "Related to",
      issue: related.key,
    });
    await waitForSuccessToast(page, "Relation added");

    await workItem.relations.removeRelation(related.key);
    await waitForSuccessToast(page, "Relation removed");
    await expect(workItem.relations.getRelation(related.key)).toHaveCount(0);
  });

  test("should navigate to the related issue", async ({ page }) => {
    const { source, related } = await seedPair();
    const workItem = await openIssue(page, source);

    await workItem.relations.addRelation({
      kind: "Related to",
      issue: related.key,
    });
    await waitForSuccessToast(page, "Relation added");

    await workItem.relations.getRelation(related.key).getByRole("link").click();
    await expect(page).toHaveURL(
      new RegExp(
        `/work/${workspace.organizationSlug}/${workspace.namespaceSlug}/${related.key}`
      )
    );
    const relatedPage = new WorkItemPage(page);
    await relatedPage.waitForLoad();
    await expect(
      page.getByText(related.key, { exact: true }).first()
    ).toBeVisible();
  });
});
