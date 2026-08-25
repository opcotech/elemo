import { createIssue } from "./api";
import { expect, test } from "./fixtures";
import {
  seedOwnerWorkspace,
  waitForErrorToast,
  waitForSuccessToast,
} from "./helpers";
import type { OwnerWorkspace } from "./helpers/workspace";
import { HomePage, WorkItemPage, WorkPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";

test.describe("@issues.create Issue Create E2E Tests", () => {
  let workspace: OwnerWorkspace;
  let writerUser: User;
  let readerUser: User;

  test.beforeAll(async ({ testConfig }) => {
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Issues Create",
    });
    writerUser = await createUser(testConfig);
    readerUser = await createUser(testConfig);

    await grantMembershipToUser(
      testConfig,
      writerUser.email,
      "Organization",
      workspace.organizationId
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Organization",
      workspace.organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      workspace.namespaceId,
      ["namespace.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      workspace.projectId,
      ["project.read", "issue.read", "document.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      workspace.projectId,
      [
        "project.update",
        "project.members.manage",
        "issue.create",
        "issue.update",
        "issue.assign",
        "document.create",
        "document.update",
        "folder.create",
      ]
    );

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      workspace.organizationId
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Organization",
      workspace.organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      workspace.namespaceId,
      ["namespace.read"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      workspace.projectId,
      ["project.read", "issue.read", "document.read"]
    );
  });

  const openProjectWork = async (page: Parameters<typeof loginUser>[0]) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const workPage = new WorkPage(page);
    await workPage.gotoProjectWork(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await workPage.waitForLoad(`${workspace.projectName} / Work`);
    return workPage;
  };

  const inspectButtonForTitle = (
    page: Parameters<typeof loginUser>[0],
    title: string
  ) =>
    page.getByRole("button", {
      name: new RegExp(`^Inspect .+: ${title}`),
    });

  const openCreatedIssue = async (workPage: WorkPage, title: string) => {
    const page = workPage.getPage();
    await workPage.surface.selectLayout("List");
    await expect(page).toHaveURL(/layout=list/);
    await workPage.surface
      .getSectionContainer()
      .getByRole("listitem")
      .filter({ hasText: title })
      .getByRole("link")
      .click();

    const workItem = new WorkItemPage(page);
    await workItem.waitForLoad();
    return workItem;
  };

  test("should create an issue from project Work with a title", async ({
    page,
  }) => {
    const workPage = await openProjectWork(page);
    const title = `Create title ${getRandomString(8)}`;

    await workPage.surface.clickCreate();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.submitWork();
    await waitForSuccessToast(page, "Issue created successfully");

    await expect(inspectButtonForTitle(page, title).first()).toBeVisible();
  });

  test("should create an issue with a selected kind", async ({ page }) => {
    const workPage = await openProjectWork(page);
    const title = `Create kind ${getRandomString(8)}`;

    await workPage.surface.clickCreate();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.fillWorkKind("Bug");
    await workPage.quickCreate.submitWork();
    await waitForSuccessToast(page, "Issue created successfully");

    const workItem = await openCreatedIssue(workPage, title);
    await expect(workItem.details.getKindSelect()).toContainText("Bug");
  });

  test("should create an issue with a description from more properties", async ({
    page,
  }) => {
    const workPage = await openProjectWork(page);
    const title = `Create description ${getRandomString(8)}`;
    const description = `Created description ${getRandomString(8)}`;

    await workPage.surface.clickCreate();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.fillWorkDescription(description);
    await workPage.quickCreate.submitWork();
    await waitForSuccessToast(page, "Issue created successfully");

    const workItem = await openCreatedIssue(workPage, title);
    await expect(workItem.getDescriptionSection()).toContainText(description);
  });

  test("should show a created issue on the list and board", async ({
    page,
  }) => {
    const workPage = await openProjectWork(page);
    const title = `Create layouts ${getRandomString(8)}`;

    await workPage.surface.clickCreate();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.submitWork();
    await waitForSuccessToast(page, "Issue created successfully");

    await expect(inspectButtonForTitle(page, title).first()).toBeVisible();

    await workPage.surface.selectLayout("List");
    await expect(page).toHaveURL(/layout=list/);
    await expect(
      workPage.surface
        .getSectionContainer()
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();

    await workPage.surface.selectLayout("Board");
    await expect(page).toHaveURL(/layout=board/);
    await expect(inspectButtonForTitle(page, title).first()).toBeVisible();
  });

  test("should create an issue from board Add work to a column", async ({
    page,
  }) => {
    await createIssue(workspace.client, workspace.projectId, {
      title: `Board seed ${getRandomString(8)}`,
    });

    const workPage = await openProjectWork(page);
    await workPage.surface.selectLayout("Board");
    await expect(page).toHaveURL(/layout=board/);

    const title = `Board add ${getRandomString(8)}`;
    await workPage.surface.getAddWorkToColumnButton("backlog").click();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.submitWork();
    await waitForSuccessToast(page, "Issue created successfully");

    await expect(inspectButtonForTitle(page, title).first()).toBeVisible();
  });

  test("should show Create unavailable without project context on Home", async ({
    page,
  }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const homePage = new HomePage(page);
    await homePage.goto();
    await homePage.waitForLoad();
    await homePage.clickCreate();
    await homePage.quickCreate.waitForLoad();
    await homePage.quickCreate.selectEntityType("Work item");

    await expect(
      homePage.quickCreate.getCreateUnavailableButton()
    ).toBeVisible();
    await expect(
      homePage.quickCreate.getCreateUnavailableButton()
    ).toBeDisabled();
  });

  test("should not allow a reader with only read permission to create", async ({
    page,
  }) => {
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const workPage = new WorkPage(page);
    await workPage.gotoProjectWork(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      workspace.projectKey
    );
    await workPage.waitForLoad(`${workspace.projectName} / Work`);

    const title = `Reader create ${getRandomString(8)}`;
    await workPage.surface.clickCreate();
    await workPage.quickCreate.waitForLoad();
    await workPage.quickCreate.fillWorkTitle(title);
    await workPage.quickCreate.submitWork();
    await waitForErrorToast(page, "Failed to create issue");
    await expect(inspectButtonForTitle(page, title)).toHaveCount(0);
  });
});
