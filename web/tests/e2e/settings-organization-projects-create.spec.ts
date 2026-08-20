import { createOrganization, getRandomProjectKey } from "./api";
import { expect, test } from "./fixtures";
import { getFormFieldMessage, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationNamespaceDetailsPage,
  SettingsOrganizationProjectCreatePage,
  SettingsOrganizationProjectDetailsPage,
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

test.describe("@settings.organization-projects-create Organization Projects Create E2E Tests", () => {
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
      name: `Projects Org ${getRandomString(8)}`,
      email: `projects-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;

    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: `Projects Namespace ${getRandomString(8)}`,
        description: `Namespace for project create tests ${getRandomString(8)}`,
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
    await grantActionsToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespaceId,
      ["namespace.update", "project.create", "document.create", "folder.create"]
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

  test("should navigate to create project page when clicking create button", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();
    await namespaceDetailsPage.projects.clickCreateProjectButton();

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/new`
      )
    );
    await expect(page.getByRole("textbox", { name: "Key" })).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Name" })).toBeVisible();
  });

  test("should create a project and navigate to project detail", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectCreatePage = new SettingsOrganizationProjectCreatePage(page);
    await projectCreatePage.goto(organizationId, namespaceId);
    await projectCreatePage.projectForm.waitForLoad();

    const projectKey = getRandomProjectKey();
    const projectName = getFullProjectName();
    const projectDescription = `Project description ${getRandomString(8)}`;

    await projectCreatePage.projectForm.fillFields({
      Key: projectKey,
      Name: projectName,
      Description: projectDescription,
    });
    await projectCreatePage.projectForm.submit("Create Project");
    await waitForSuccessToast(page, "Project created");

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/(?!new$)[^/]+$`
      )
    );

    // Creator receives project-maintainer actions on the new project.
    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.waitForLoad();
    await projectDetailsPage.projectInfo.waitForLoad();

    expect(await projectDetailsPage.getTitleText()).toContain(projectName);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Key")
    ).toContainText(projectKey);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Name")
    ).toContainText(projectName);
    expect(await projectDetailsPage.projectInfo.hasEditProjectButton()).toBe(
      true
    );
    expect(await projectDetailsPage.dangerZone.hasDeleteButton()).toBe(true);
  });

  test("should allow creating a project with only key and name", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectCreatePage = new SettingsOrganizationProjectCreatePage(page);
    await projectCreatePage.goto(organizationId, namespaceId);
    await projectCreatePage.projectForm.waitForLoad();

    const projectKey = getRandomProjectKey();
    const projectName = getFullProjectName();

    await projectCreatePage.projectForm.fillFields({
      Key: projectKey,
      Name: projectName,
    });
    await projectCreatePage.projectForm.submit("Create Project");
    await waitForSuccessToast(page, "Project created");

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/(?!new$)[^/]+$`
      )
    );

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.waitForLoad();
    expect(await projectDetailsPage.getTitleText()).toContain(projectName);
  });

  test("should show validation errors for invalid project inputs", async ({
    page,
  }) => {
    const fieldMessage = (label: string) => getFormFieldMessage(page, label);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectCreatePage = new SettingsOrganizationProjectCreatePage(page);
    await projectCreatePage.goto(organizationId, namespaceId);
    await projectCreatePage.projectForm.waitForLoad();

    await projectCreatePage.projectForm.submit("Create Project");
    await expect(fieldMessage("Key")).toHaveText(
      /too small|invalid input|required/i
    );
    await expect(fieldMessage("Name")).toHaveText(
      /too small|invalid input|required/i
    );

    await projectCreatePage.projectForm.fillFields({
      Key: "AB1",
      Name: "Valid Project Name",
    });
    await projectCreatePage.projectForm.submit("Create Project");
    await expect(fieldMessage("Key")).toHaveText(/ascii letters/i);
  });

  test("should cancel create and return to namespace details", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectCreatePage = new SettingsOrganizationProjectCreatePage(page);
    await projectCreatePage.goto(organizationId, namespaceId);
    await projectCreatePage.projectForm.waitForLoad();
    await projectCreatePage.projectForm.cancel();

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}$`
      )
    );
  });

  test("should not allow creating a project without namespace write permission", async ({
    page,
  }) => {
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectCreatePage = new SettingsOrganizationProjectCreatePage(page);
    await projectCreatePage.goto(organizationId, namespaceId);

    await expect(page).toHaveURL(/\/permission-denied(?:\?|$)/);
  });
});
