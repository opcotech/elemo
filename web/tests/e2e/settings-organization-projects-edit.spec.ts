import { createOrganization, createProject, getRandomProjectKey } from "./api";
import { expect, test } from "./fixtures";
import { getFormFieldMessage, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationNamespaceDetailsPage,
  SettingsOrganizationProjectDetailsPage,
  SettingsOrganizationProjectEditPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantPermissionToUser,
  grantSystemOwnerMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@settings.organization-projects-edit Organization Projects Edit E2E Tests", () => {
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

    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);

    ownerApiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(ownerApiClient, {
      name: `Project Edit Org ${getRandomString(8)}`,
      email: `project-edit-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;

    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: `Project Edit Namespace ${getRandomString(8)}`,
        description: `Namespace for project edit tests ${getRandomString(8)}`,
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
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespaceId,
      "read"
    );

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      namespaceId,
      "read"
    );
  });

  const seedProjectWithWriteAccess = async (
    testConfig: Parameters<typeof grantPermissionToUser>[0]
  ) => {
    const project = await createProject(ownerApiClient, namespaceId, {
      key: getRandomProjectKey(),
      name: getFullProjectName(),
      description: `Project description ${getRandomString(8)}`,
    });

    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Project",
      project.id,
      "write"
    );

    return project;
  };

  test("should allow members with project write permission to update project", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProjectWithWriteAccess(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.goto(organizationId, namespaceId, project.id);
    await projectEditPage.projectEditForm.waitForLoad();

    const updatedName = `Updated ${getRandomString(6)}`;
    const updatedDescription = `Updated description ${getRandomString(8)}`;

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.fillFields({
      Name: updatedName,
      Description: updatedDescription,
    });
    await projectEditPage.projectEditForm.selectStatus("Pending");
    await projectEditPage.projectEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "Project updated");

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${project.id}$`
      )
    );

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.waitForLoad();
    await projectDetailsPage.projectInfo.waitForLoad();

    await expect(
      page.getByRole("main").getByRole("heading", { level: 1 }).first()
    ).toContainText(updatedName);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Name")
    ).toContainText(updatedName);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Description")
    ).toContainText(updatedDescription);
    await expect(
      projectDetailsPage.projectInfo.getFieldValue("Status")
    ).toContainText("pending");

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespaceId);
    await namespaceDetailsPage.projects.waitForLoad();

    const projectRow =
      namespaceDetailsPage.projects.getRowByProjectName(updatedName);
    await expect(projectRow).toBeVisible();
    await expect(projectRow).toContainText("pending");
  });

  test("should show validation errors for invalid project edit inputs", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProjectWithWriteAccess(testConfig);
    const fieldMessage = (label: string) => getFormFieldMessage(page, label);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.goto(organizationId, namespaceId, project.id);
    await projectEditPage.projectEditForm.waitForLoad();

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.submit("Save Changes");
    await expect(fieldMessage("Name")).toHaveText(
      /too small|invalid input|required/i
    );

    await projectEditPage.projectEditForm.fillField("Name", "Valid Name");
    await projectEditPage.projectEditForm.clearField("Key");
    await projectEditPage.projectEditForm.fillField("Key", "AB1");
    await projectEditPage.projectEditForm.submit("Save Changes");
    await expect(fieldMessage("Key")).toHaveText(/ascii letters/i);
  });

  test("should cancel edit and return to project detail", async ({
    page,
    testConfig,
  }) => {
    const project = await seedProjectWithWriteAccess(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.goto(organizationId, namespaceId, project.id);
    await projectEditPage.projectEditForm.waitForLoad();

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.fillField(
      "Name",
      `Cancel ${getRandomString(6)}`
    );
    await projectEditPage.projectEditForm.cancel();

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${project.id}$`
      )
    );

    const projectDetailsPage = new SettingsOrganizationProjectDetailsPage(page);
    await projectDetailsPage.waitForLoad();
    expect(await projectDetailsPage.getTitleText()).toContain(project.name);
  });

  test("should not allow editing without project write permission", async ({
    page,
    testConfig,
  }) => {
    const project = await createProject(ownerApiClient, namespaceId, {
      key: getRandomProjectKey(),
      name: getFullProjectName(),
      description: `Project description ${getRandomString(8)}`,
    });

    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Project",
      project.id,
      "read"
    );

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.goto(organizationId, namespaceId, project.id);

    await expect(page).toHaveURL(/\/permission-denied(?:\?|$)/);
  });
});
