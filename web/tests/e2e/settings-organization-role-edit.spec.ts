import { createOrganization, createRole } from "./api";
import { expect, test } from "./fixtures";
import {
  getFormFieldMessage,
  waitForPageLoad,
  waitForPermissionsLoad,
  waitForSuccessToast,
} from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationRoleEditPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantPermissionToUser,
  grantSystemOwnerMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api";

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

    // Grant system owner membership so user can create organizations
    await grantSystemOwnerMembershipToUser(testConfig, testUser.email);

    // Create organization via API
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

    // Create a role to edit
    const role = await createRole(apiClient, organizationId, {
      name: initialRoleName,
      description: initialRoleDescription,
    });
    roleId = role.id;

    // Grant read-only user read permission on the organization
    await grantPermissionToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId,
      "read"
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

    // Verify form fields are populated with current values
    const nameField = roleEditPage.roleEditForm.getField("Name");
    const descriptionField = roleEditPage.roleEditForm.getField("Description");

    await expect(nameField).toHaveValue(initialRoleName);
    await expect(descriptionField).toHaveValue(initialRoleDescription);
  });

  test("should allow editing role name", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Edit role name
    const newName = `Updated Role ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");

    // Verify success toast
    await waitForSuccessToast(page, "updated");

    // Verify navigated back to organization details
    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}`)
    );

    // Verify role name updated in the list
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

    // Edit role description
    const newDescription = `Updated description ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.fillField("Description", newDescription);
    await roleEditPage.roleEditForm.submit("Save Changes");

    // Verify success toast
    await waitForSuccessToast(page, "updated");

    // Verify navigated back to organization details
    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}`)
    );

    // Verify description is updated in the roles list
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    const roleRow = orgDetailsPage.roles.getRowByRoleName(currentRoleName);
    await expect(roleRow.getByText(newDescription)).toBeVisible();
  });

  test("should allow editing both name and description", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Edit both fields
    const newName = `Both Updated ${getRandomString(8)}`;
    const newDescription = `Both updated desc ${getRandomString(8)}`;

    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.fillField("Description", newDescription);
    await roleEditPage.roleEditForm.submit("Save Changes");

    // Verify success toast
    await waitForSuccessToast(page, "updated");

    // Verify both changes persisted
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

    // Enter invalid input (empty name which is required)
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", "");
    await roleEditPage.roleEditForm.submit("Save Changes");

    // Verify validation error is shown for the name field
    await expect(nameError).toHaveText(/invalid input/i);
  });

  test("should save changes and show success message", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Make a change
    const newName = `Success Test ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");

    // Verify success toast is shown
    await waitForSuccessToast(page, "updated");
  });

  test("should persist changes after page reload", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Edit role
    const newName = `Persist Test ${getRandomString(8)}`;
    const newDescription = `Persist desc ${getRandomString(8)}`;

    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.fillField("Description", newDescription);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    // Navigate back to edit page
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Verify changes persisted
    const nameField = roleEditPage.roleEditForm.getField("Name");
    const descriptionField = roleEditPage.roleEditForm.getField("Description");

    await expect(nameField).toHaveValue(newName);
    await expect(descriptionField).toHaveValue(newDescription);

    // Reload the page
    await page.reload();
    await roleEditPage.roleEditForm.waitForLoad();

    // Verify changes still persist after reload
    await expect(nameField).toHaveValue(newName);
    await expect(descriptionField).toHaveValue(newDescription);
  });

  test("should allow clearing description field", async ({ page }) => {
    // First, create a role with description
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Clear the description
    await roleEditPage.roleEditForm.clearField("Description");
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    // Verify description was cleared
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    const descriptionField = roleEditPage.roleEditForm.getField("Description");
    await expect(descriptionField).toHaveValue("");
  });

  test("should allow canceling edit and return to organization details", async ({
    page,
  }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Make a change
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField(
      "Name",
      `Cancel Test ${getRandomString(8)}`
    );

    // Click cancel
    await roleEditPage.roleEditForm.cancel();

    // Verify navigated back to organization details
    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}`)
    );
  });

  test("should not allow editing without role write permission", async ({
    page,
  }) => {
    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Try to navigate to edit page - should be redirected or show error
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);

    // Wait for page and permissions to load
    await waitForPageLoad(page);

    // Should not be on the edit page anymore (redirected due to lack of permission)
    const currentUrl = page.url();
    expect(currentUrl).not.toContain("/edit");
  });

  test("should show edit button in roles list for users with write permission", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role for this test to avoid conflicts with previous edits
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const freshRole = await createRole(apiClient, organizationId, {
      name: `Fresh Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Get the role row for the fresh role
    const roleRow = orgDetailsPage.roles.getRowByRoleName(freshRole.name);
    await expect(roleRow).toBeVisible();

    // Verify edit button is visible
    const editButton = roleRow.getByRole("link", { name: /edit role/i });
    await expect(editButton).toBeVisible();
  });

  test("should navigate to edit page when clicking edit button", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role for this test to avoid conflicts with previous edits
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const freshRoleName = `Navigate Role ${getRandomString(8)}`;
    const freshRole = await createRole(apiClient, organizationId, {
      name: freshRoleName,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Click edit button
    const roleRow = orgDetailsPage.roles.getRowByRoleName(freshRoleName);
    const editButton = roleRow.getByRole("link", { name: /edit role/i });
    await editButton.click();

    // Verify navigated to edit page
    await expect(page).toHaveURL(
      new RegExp(
        `/settings/organizations/${organizationId}/roles/${freshRole.id}/edit`
      )
    );

    // Verify edit form is loaded
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.roleEditForm.waitForLoad();
  });

  test("should preserve unchanged fields when updating", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Get original description
    const descriptionField = roleEditPage.roleEditForm.getField("Description");
    const originalDescription = await descriptionField.inputValue();

    // Only change the name
    const newName = `Only Name ${getRandomString(8)}`;
    await roleEditPage.roleEditForm.clearField("Name");
    await roleEditPage.roleEditForm.fillField("Name", newName);
    await roleEditPage.roleEditForm.submit("Save Changes");
    await waitForSuccessToast(page, "updated");

    // Go back to edit page
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.roleEditForm.waitForLoad();

    // Verify name changed but description stayed the same
    const nameField = roleEditPage.roleEditForm.getField("Name");
    await expect(nameField).toHaveValue(newName);
    await expect(descriptionField).toHaveValue(originalDescription);
  });

  test("should grant and revoke organization write access for role members via permissions", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    test.setTimeout(60_000);

    const scenarioUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId,
      "read"
    );
    const memberFullName = `${scenarioUser.first_name} ${scenarioUser.last_name}`;
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const permissionRole = await createRole(apiClient, organizationId, {
      name: `Org Permission Role ${getRandomString(8)}`,
    });

    // Baseline: member cannot edit organization
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    let orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    await waitForPermissionsLoad(page, organizationId);
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();

    // Owner grants organization write permission via role permissions
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, permissionRole.id);
    await roleEditPage.roleEditForm.waitForLoad();
    await roleEditPage.permissions.waitForLoad();
    await roleEditPage.permissions.addPermission({
      resourceType: "Organization",
      resourceId: organizationId,
      permissionKind: "write",
    });
    await roleEditPage.members.waitForLoad();
    await roleEditPage.members.addMember(memberFullName);
    await expect(
      roleEditPage.members.getRowByMemberName(memberFullName)
    ).toBeVisible();

    // Member now has access to edit organization
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    await waitForPermissionsLoad(page, organizationId);
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeTruthy();

    // Remove permission
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await roleEditPage.goto(organizationId, permissionRole.id);
    await roleEditPage.roleEditForm.waitForLoad();
    await roleEditPage.permissions.waitForLoad();
    await roleEditPage.permissions.removePermission({
      resourceType: "Organization",
      resourceId: organizationId,
      permissionKind: "write",
    });

    // Member loses edit access again
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();
    await waitForPermissionsLoad(page, organizationId);
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();
  });

  test("should grant and revoke role write access for members via role permissions", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    test.setTimeout(60_000);

    const scenarioUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      scenarioUser.email,
      "Organization",
      organizationId,
      "read"
    );
    const memberFullName = `${scenarioUser.first_name} ${scenarioUser.last_name}`;
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const roleWithPermissions = await createRole(apiClient, organizationId, {
      name: `Role Write Permission ${getRandomString(8)}`,
    });

    // Baseline: member cannot manage the role
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    let orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();
    let roleRow = orgDetailsPage.roles.getRowByRoleName(
      roleWithPermissions.name
    );
    await expect(roleRow).toBeVisible();
    await expect(
      roleRow.getByRole("button", { name: /add member/i })
    ).not.toBeVisible();
    await expect(
      roleRow.getByRole("link", { name: /edit role/i })
    ).not.toBeVisible();

    // Owner grants role write permission and assigns member
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleWithPermissions.id);
    await roleEditPage.roleEditForm.waitForLoad();
    await roleEditPage.permissions.waitForLoad();
    await roleEditPage.permissions.addPermission({
      resourceType: "Role",
      resourceId: roleWithPermissions.id,
      permissionKind: "write",
    });
    await roleEditPage.members.waitForLoad();
    await roleEditPage.members.addMember(memberFullName);
    await expect(
      roleEditPage.members.getRowByMemberName(memberFullName)
    ).toBeVisible();

    // Member should now see role actions
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();
    roleRow = orgDetailsPage.roles.getRowByRoleName(roleWithPermissions.name);
    await expect(
      roleRow.getByRole("button", { name: /add member/i })
    ).toBeVisible();
    await expect(
      roleRow.getByRole("link", { name: /edit role/i })
    ).toBeVisible();

    // Remove permission
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await roleEditPage.goto(organizationId, roleWithPermissions.id);
    await roleEditPage.roleEditForm.waitForLoad();
    await roleEditPage.permissions.waitForLoad();
    await roleEditPage.permissions.removePermission({
      resourceType: "Role",
      resourceId: roleWithPermissions.id,
      permissionKind: "write",
    });

    // Member loses role management actions again
    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();
    roleRow = orgDetailsPage.roles.getRowByRoleName(roleWithPermissions.name);
    await expect(
      roleRow.getByRole("button", { name: /add member/i })
    ).not.toBeVisible();
    await expect(
      roleRow.getByRole("link", { name: /edit role/i })
    ).not.toBeVisible();
  });
});
