import { createOrganization, createProject, getRandomProjectKey } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationNamespaceDetailsPage,
  SettingsOrganizationProjectDetailsPage,
  SettingsOrganizationProjectEditPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantMembershipToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@settings.organization-projects-detail Organization Projects Detail E2E Tests", () => {
  let ownerUser: User;
  let writerUser: User;
  let readerUser: User;
  let organizationId: string;
  let namespaceId: string;
  let ownerApiClient: Client;

  const getFullProjectName = () => `Project ${getRandomString(8)}`;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    writerUser = await createUser(testConfig);
    readerUser = await createUser(testConfig);

    await grantOrganizationCreateToUser(testConfig, ownerUser.email);

    ownerApiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(ownerApiClient, {
      name: `Project Detail Org ${getRandomString(8)}`,
      email: `project-detail-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;

    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: `Project Detail Namespace ${getRandomString(8)}`,
        description: `Namespace for project detail tests ${getRandomString(8)}`,
      },
      throwOnError: true,
    });
    namespaceId = namespaceResponse.data.id ?? "";

    await grantMembershipToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespaceId,
      ["namespace.read"]
    );

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      namespaceId,
      ["namespace.read"]
    );
  });

  const seedProject = async (overrides?: {
    key?: string;
    name?: string;
    description?: string;
  }) => {
    const key = overrides?.key ?? getRandomProjectKey();
    const name = overrides?.name ?? getFullProjectName();
    const description =
      overrides?.description ?? `Project description ${getRandomString(8)}`;

    return createProject(ownerApiClient, namespaceId, {
      key,
      name,
      description,
    });
  };

  test("should allow members with project read permission to view project details", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProject();

    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.goto(organizationId, namespaceId, project.id);
    await projectDetailsPage.waitForLoad();
    await projectDetailsPage.projectInfo.waitForLoad();

    expect(await projectDetailsPage.getTitleText()).toContain(project.name);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Key")
    ).toContainText(project.key);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Name")
    ).toContainText(project.name);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Status")
    ).toContainText(/active|pending/i);
  });

  test("should show edit button only for members with project write permission", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProject();

    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
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

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.goto(organizationId, namespaceId, project.id);
    await projectDetailsPage.projectInfo.waitForLoad();

    expect(await projectDetailsPage.projectInfo.hasEditProjectButton()).toBe(
      true
    );
    await projectDetailsPage.projectInfo.clickEditProjectButton();

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.projectEditForm.waitForLoad();
    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${project.id}/edit`
      )
    );

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await projectDetailsPage.goto(organizationId, namespaceId, project.id);
    await projectDetailsPage.projectInfo.waitForLoad();

    expect(await projectDetailsPage.projectInfo.hasEditProjectButton()).toBe(
      false
    );
  });

  test("should navigate from namespace project list row to project detail when user has project read", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProject();

    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();

    const projectRow = namespaceDetailsPage.projects.getRowByProjectName(
      project.name
    );
    await expect(projectRow).toBeVisible();
    await expect(
      projectRow.getByRole("link", { name: project.name })
    ).toBeVisible();

    await namespaceDetailsPage.projects.clickProjectLink(project.name);

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${project.id}`
      )
    );

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.waitForLoad();
    expect(await projectDetailsPage.getTitleText()).toContain(project.name);
  });

  test("should not list projects without project read permission", async ({
    page,
  }) => {
    const project = await seedProject();

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();

    await expect(
      namespaceDetailsPage.projects.getRowByProjectName(project.name)
    ).not.toBeVisible();
  });

  test("should show access denied without project read permission", async ({
    page,
  }) => {
    const project = await seedProject();

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.goto(organizationId, namespaceId, project.id);
    await projectDetailsPage.waitForLoad();

    await expect(
      page.getByRole("heading", { name: "Access Denied" })
    ).toBeVisible();
    await expect(
      page.getByText("You do not have permission to view this project.")
    ).toBeVisible();
  });

  test("should show list action icons based on project write and delete permissions", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProject();

    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
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
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      ["project.delete"]
    );
    await grantActionsToUser(
      testConfig,
      readerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();

    const writerProjectRow = namespaceDetailsPage.projects.getRowByProjectName(
      project.name
    );
    await expect(writerProjectRow).toBeVisible();
    // Action controls use sr-only labels; wait with auto-retrying assertions
    // because per-project permission fetches can resolve after table load.
    await expect(
      writerProjectRow.getByRole("link", { name: /edit project/i })
    ).toBeVisible();
    await expect(
      writerProjectRow.getByRole("button", { name: /delete project/i })
    ).toBeVisible();

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();

    const readerProjectRow = namespaceDetailsPage.projects.getRowByProjectName(
      project.name
    );
    await expect(readerProjectRow).toBeVisible();
    await expect(
      readerProjectRow.getByRole("link", { name: /edit project/i })
    ).not.toBeVisible();
    await expect(
      readerProjectRow.getByRole("button", { name: /delete project/i })
    ).not.toBeVisible();
  });

  test("should allow deleting a project from the danger zone", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProject();

    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      ["project.read", "issue.read", "document.read"]
    );
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      ["project.delete"]
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.goto(organizationId, namespaceId, project.id);
    await projectDetailsPage.projectInfo.waitForLoad();
    await projectDetailsPage.dangerZone.waitForLoad();

    await projectDetailsPage.dangerZone.clickDeleteButton();
    await page.getByRole("button", { name: "Delete", exact: true }).click();
    await waitForSuccessToast(page, "Project deleted");

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}$`
      )
    );

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.projects.waitForLoad();
    expect(await namespaceDetailsPage.projects.hasProject(project.name)).toBe(
      false
    );
  });
});
