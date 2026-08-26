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
  grantActionsToUser,
  grantMembershipToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomSlug, getRandomString } from "./utils/random";

import type { Client } from "@/lib/api/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type { User } from "@/lib/api/types";

test.describe("@settings.organization-projects-edit Organization Projects Edit E2E Tests", () => {
  let ownerUser: User;
  let writerUser: User;
  let readerUser: User;
  let organizationId: string;
  let organizationSlug: string;
  let namespaceId: string;
  let namespaceSlug: string;
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
      name: `Project Edit Org ${getRandomString(8)}`,
      email: `project-edit-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;
    organizationSlug = organization.slug;

    namespaceSlug = getRandomSlug("ns");
    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { organizationRef: organizationId },
      body: {
        slug: namespaceSlug,
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

  const seedProjectWithWriteAccess = async (
    testConfig: Parameters<typeof grantActionsToUser>[0]
  ) => {
    const project = await createProject(
      ownerApiClient,
      organizationId,
      namespaceId,
      {
        key: getRandomProjectKey(),
        name: getFullProjectName(),
        description: `Project description ${getRandomString(8)}`,
      }
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
    await projectEditPage.goto(organizationSlug, namespaceSlug, project.key);
    await projectEditPage.projectEditForm.waitForLoad();

    const updatedName = `Updated ${getRandomString(6)}`;
    const updatedDescription = `Updated description ${getRandomString(8)}`;

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.fillFields({
      Name: updatedName,
      Description: updatedDescription,
    });
    await expect(projectEditPage.projectEditForm.getField("Name")).toHaveValue(
      updatedName
    );
    await projectEditPage.projectEditForm.selectStatus("Pending");
    await projectEditPage.projectEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "Project updated");

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationSlug}/namespaces/${namespaceSlug}/projects/${project.key}$`
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
    await namespaceDetailsPage.goto(organizationSlug, namespaceSlug);
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
    await projectEditPage.goto(organizationSlug, namespaceSlug, project.key);
    await projectEditPage.projectEditForm.waitForLoad();

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.submit("Save Changes");
    await expect(fieldMessage("Name")).toHaveText(
      /too small|invalid input|required/i
    );

    await expect(
      projectEditPage.projectEditForm.getField("Key")
    ).toBeDisabled();
    await expect(projectEditPage.projectEditForm.getField("Key")).toHaveValue(
      project.key
    );
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
    await projectEditPage.goto(organizationSlug, namespaceSlug, project.key);
    await projectEditPage.projectEditForm.waitForLoad();

    await projectEditPage.projectEditForm.clearField("Name");
    await projectEditPage.projectEditForm.fillField(
      "Name",
      `Cancel ${getRandomString(6)}`
    );
    await projectEditPage.projectEditForm.cancel();

    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationSlug}/namespaces/${namespaceSlug}/projects/${project.key}$`
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
    const project = await createProject(
      ownerApiClient,
      organizationId,
      namespaceId,
      {
        key: getRandomProjectKey(),
        name: getFullProjectName(),
        description: `Project description ${getRandomString(8)}`,
      }
    );

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

    const projectEditPage = new SettingsOrganizationProjectEditPage(page);
    await projectEditPage.goto(organizationSlug, namespaceSlug, project.key);

    await expect(page).toHaveURL(/\/permission-denied(?:\?|$)/);
  });
});
