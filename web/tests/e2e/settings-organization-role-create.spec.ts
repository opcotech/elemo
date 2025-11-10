import { createOrganization, createRole } from "./api";
import { expect, test } from "./fixtures";
import { waitForPermissionsLoad, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationRoleCreatePage,
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

test.describe("@settings.organization-role-create Organization Role Creation E2E Tests", () => {
  let testUser: User;
  let readOnlyUser: User;
  let organizationId: string;

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
      name: `Test Org Role Create ${uniqueId}`,
      email: `test-role-create-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    // Grant read-only user read permission on the organization
    await grantMembershipToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId,
      "read"
    );
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should display list of organization roles", async ({ page }) => {
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify roles section is visible (by checking the section container)
    await expect(orgDetailsPage.roles.getSectionContainer()).toBeVisible();
  });

  test("should allow creating a new role", async ({ page }) => {
    // Navigate to role create page via the create button
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Click create role button
    await orgDetailsPage.roles.clickCreateRoleButton();

    // Fill role form
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Test Role ${getRandomString(8)}`;
    const roleDescription = `Test role description ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
      Description: roleDescription,
    });

    // Submit form
    await roleCreatePage.roleCreateForm.submit("Create Role");

    // Wait for success toast
    await waitForSuccessToast(page, "created");

    // Verify role appears in the list
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should grant organization write access to role members when creating role with permissions", async ({
    page,
    testConfig,
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
    const scenarioFullName = `${scenarioUser.first_name} ${scenarioUser.last_name}`;

    // Baseline: read-only member cannot edit organization
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

    // Owner logs in to create role with permissions
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();
    await orgDetailsPage.roles.clickCreateRoleButton();

    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Permission Role ${getRandomString(8)}`;
    const roleDescription = `Permission role description ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
      Description: roleDescription,
    });

    // Add organization write permission via the Permissions card
    await roleCreatePage.rolePermissionDraft.waitForLoad();
    await roleCreatePage.rolePermissionDraft.addPermission({
      resourceType: "Organization",
      resourceId: organizationId,
      permissionKind: "write",
    });

    // Submit creation with pending permission
    const permissionSubmitLabel = "Create Role with 1 Permission(s)";
    await roleCreatePage.roleCreateForm.submit(permissionSubmitLabel);
    await waitForSuccessToast(page, "created");

    // Navigate back to organization details and verify role exists
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();
    const newRoleRow = orgDetailsPage.roles.getRowByRoleName(roleName);
    await expect(newRoleRow).toBeVisible();

    // Assign the read-only member to the newly created role
    const editButton = newRoleRow.getByRole("link", { name: /edit role/i });
    await editButton.click();
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.roleEditForm.waitForLoad();
    await roleEditPage.members.waitForLoad();
    await roleEditPage.members.addMember(scenarioFullName);
    await expect(
      roleEditPage.members.getRowByMemberName(scenarioFullName)
    ).toBeVisible();

    // After assignment and permission, member should now see edit button
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
  });

  test("should show validation errors for invalid inputs", async ({ page }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    // Fill in a name that's too short (less than 3 characters)
    await roleCreatePage.roleCreateForm.fillFields({
      Name: "AB", // Only 2 characters
    });

    // Try submitting the form with invalid data
    await roleCreatePage.roleCreateForm.submit("Create Role");

    // Verify validation error is shown for the name field
    await expect(
      page.getByText(/too small: expected string to have >=3 characters/i)
    ).toBeVisible();
  });

  test("should save role and show success message", async ({ page }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    // Fill and submit form
    const roleName = `Success Test Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");

    // Verify success toast is shown and we're redirected to org details
    await waitForSuccessToast(page, "created");

    // Verify we're on the roles list page and the role exists
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();

    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should persist role after page reload", async ({ page }) => {
    // Create role
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Persist Test Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    // Reload page
    await page.reload();

    // Verify role still exists
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should not show create role button without organization write permission", async ({
    page,
  }) => {
    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Navigate to organization details
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify create button is not visible
    expect(await orgDetailsPage.roles.hasCreateRoleButton()).toBeFalsy();
  });

  test("should not show add member button without role write permission", async ({
    page,
    createApiClient,
  }) => {
    // Create a role as testUser (who has write permission)
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Test Role Member Perm ${getRandomString(8)}`,
    });

    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Navigate to organization details
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify role is visible but add member button is not
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    // Verify add member button (UserPlus icon) is not visible
    const addMemberButton = roleRow.getByRole("button", {
      name: /add member/i,
    });
    await expect(addMemberButton).not.toBeVisible();
  });

  test("should not show edit role button without role write permission", async ({
    page,
    createApiClient,
  }) => {
    // Create a role as testUser (who has write permission)
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Test Role Edit Perm ${getRandomString(8)}`,
    });

    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Navigate to organization details
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify role is visible but edit button is not
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    // Verify edit button (Edit icon) is not visible
    const editButton = roleRow.getByRole("button", { name: /edit role/i });
    await expect(editButton).not.toBeVisible();
  });

  test("should not show delete role button without role delete permission", async ({
    page,
    createApiClient,
  }) => {
    // Create a role as testUser (who has write permission)
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Test Role Delete Perm ${getRandomString(8)}`,
    });

    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Navigate to organization details
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify role is visible but delete button is not
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    // Verify delete button (Trash2 icon) is not visible
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await expect(deleteButton).not.toBeVisible();
  });

  test("should create role with only name (description optional)", async ({
    page,
  }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    // Fill only name field
    const roleName = `Name Only Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillField("Name", roleName);

    // Submit form
    await roleCreatePage.roleCreateForm.submit("Create Role");

    // Verify success
    await waitForSuccessToast(page, "created");

    // Verify role appears in list
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should allow canceling role creation and return to organization details", async ({
    page,
  }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    // Fill some data
    await roleCreatePage.roleCreateForm.fillField(
      "Name",
      `Cancel Test ${getRandomString(8)}`
    );

    // Click cancel button
    await roleCreatePage.roleCreateForm.cancel();

    // Verify navigated back to organization details
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
  });

  test("should show role in list with correct details", async ({ page }) => {
    const roleName = `Details Test Role ${getRandomString(8)}`;
    const roleDescription = `This is a detailed description ${getRandomString(8)}`;

    // Create role via UI
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationId);
    await roleCreatePage.roleCreateForm.waitForLoad();

    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
      Description: roleDescription,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    // Navigate to roles list
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify role details in the table
    const roleRow = orgDetailsPage.roles.getRowByRoleName(roleName);
    await expect(roleRow).toBeVisible();
    await expect(roleRow.getByText(roleName)).toBeVisible();
    await expect(roleRow.getByText(roleDescription)).toBeVisible();
    await expect(roleRow.getByText("1 member")).toBeVisible();
  });
});
