import {
  createGrant,
  createOrganization,
  createRole,
  deleteGrant,
  updateRole,
} from "./api";
import { expect, test } from "./fixtures";
import { getFormFieldMessage, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationRoleEditPage,
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

test.describe("@settings.organization-role-edit Organization Role Edit E2E Tests", () => {
  let testUser: User;
  let readOnlyUser: User;
  let organizationId: string;
  let roleId: string;
  const initialRoleName = `Test Role ${getRandomString(8)}`;
  const initialRoleDescription = `Test role description ${getRandomString(8)}`;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    testUser = await createUser(testConfig);
    readOnlyUser = await createUser(testConfig);

    await grantOrganizationCreateToUser(testConfig, testUser.email);

    const uniqueId = getRandomString(8);
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Role Edit ${uniqueId}`,
      email: `test-role-edit-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    const role = await createRole(apiClient, organizationId, {
      name: initialRoleName,
      description: initialRoleDescription,
      actions: ["project.read"],
    });
    roleId = role.id;

    await grantActionsToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );
    await grantMembershipToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId
    );
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should display current role details", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const nameField = roleEditPage.roleEditForm.getField("Name");
    const descriptionField = roleEditPage.roleEditForm.getField("Description");

    await expect(nameField).toHaveValue(initialRoleName);
    await expect(descriptionField).toHaveValue(initialRoleDescription);
  });

  test("should allow editing role name", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const newName = `Updated Role ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}`)
    );

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    expect(await orgDetailsPage.roles.hasRole(newName)).toBeTruthy();
  });

  test("should allow editing role description", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const currentNameField = roleEditPage.roleEditForm.getField("Name");
    const currentRoleName = await currentNameField.inputValue();

    const newDescription = `Updated description ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.fillField("Description", newDescription);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}`)
    );

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    const roleRow = orgDetailsPage.roles.getRowByRoleName(currentRoleName);
    await expect(roleRow.getByText(newDescription)).toBeVisible();
  });

  test("should allow editing both name and description", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const newName = `Both Updated ${getRandomString(8)}`;
    const newDescription = `Both updated desc ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.fillField("Description", newDescription);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    const roleRow = orgDetailsPage.roles.getRowByRoleName(newName);
    await expect(roleRow).toBeVisible();
    await expect(roleRow.getByText(newDescription)).toBeVisible();
  });

  test("should show validation errors for invalid inputs", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();
    const nameError = getFormFieldMessage(page, "Name");

    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", "AB");
    await roleEditPage.roleEditForm.submit("Save Changes");
    await expect(nameError).toHaveText(/invalid input/i);
  });

  test("should persist role changes after page reload", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const newName = `Reload Role ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await page.reload();
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(newName)).toBeVisible();
  });

  test("should preserve unchanged fields when updating", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const descriptionField = roleEditPage.roleEditForm.getField("Description");
    const originalDescription = await descriptionField.inputValue();

    const newName = `Only Name ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const nameField = roleEditPage.roleEditForm.getField("Name");
    await expect(nameField).toHaveValue(newName);
    await expect(descriptionField).toHaveValue(originalDescription);
  });

  test("should grant and revoke organization.update via a role bundle", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    const scenarioUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId
    );
    await grantActionsToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );

    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    let orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();

    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Org Update Bundle ${getRandomString(8)}`,
      actions: ["organization.update"],
    });
    const grant = await createGrant(apiClient, {
      principal: { resourceType: "User", id: scenarioUser.id },
      scope: { resourceType: "Organization", id: organizationId },
      role_id: role.id,
    });

    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeTruthy();

    await updateRole(apiClient, organizationId, role.id, { actions: [] });
    await deleteGrant(apiClient, grant.id);

    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();
  });
});
