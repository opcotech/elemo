import { createOrganization, createRole } from "./api";
import { Dialog } from "./components";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { SettingsOrganizationDetailsPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";

test.describe("@settings.organization-role-delete Organization Role Delete E2E Tests", () => {
  let testUser: User;
  let readOnlyUser: User;
  let organizationId: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    testUser = await createUser(testConfig);
    readOnlyUser = await createUser(testConfig);

    await grantOrganizationCreateToUser(testConfig, testUser.email);

    // Create organization via API
    const uniqueId = getRandomString(8);
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Role Delete ${uniqueId}`,
      email: `test-role-delete-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    // Grant read-only user read permission on the organization
    await grantActionsToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should display current role details before deletion", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Display Test Role ${getRandomString(8)}`,
      description: `Test description ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify role is visible
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();
    await expect(roleRow.getByText(role.name)).toBeVisible();
  });

  test("should allow deleting role", async ({ page, createApiClient }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Delete Test Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    // Click delete button
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    // Wait for and confirm delete dialog
    const dialog = new Dialog(page);
    await dialog.waitFor(`Are you sure you want to delete ${role.name}?`);
    await dialog.confirm("Delete");

    // Wait for success toast
    await waitForSuccessToast(page, "deleted");

    // Verify role is removed from the list
    await orgDetailsPage.roles.waitForLoad();
    await expect(
      orgDetailsPage.roles.getRowByRoleName(role.name)
    ).not.toBeVisible();
  });

  test("should show confirmation dialog with role details", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Confirmation Test Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Click delete button
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    // Verify dialog shows role name
    const dialog = new Dialog(page);
    await dialog.waitFor(`Are you sure you want to delete ${role.name}?`);

    // Verify dialog shows consequences
    const dialogContent = dialog.getContent();
    await expect(
      dialogContent.getByText("The role will be permanently deleted")
    ).toBeVisible();
    await expect(
      dialogContent.getByText(
        "Grants that reference this role will no longer include its bundled actions"
      )
    ).toBeVisible();
    await expect(
      dialogContent.getByText("This action cannot be undone", { exact: true })
    ).toBeVisible();

    // Cancel the dialog
    await dialog.cancel();

    // Verify role still exists
    await orgDetailsPage.roles.waitForLoad();
    await expect(
      orgDetailsPage.roles.getRowByRoleName(role.name)
    ).toBeVisible();
  });

  test("should show success message on deletion", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Success Message Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Delete role
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.confirm("Delete");

    // Verify success toast
    await waitForSuccessToast(page, "deleted");
    await expect(
      page.getByText("The role has been deleted successfully")
    ).toBeVisible();
  });

  test("should allow canceling deletion", async ({ page, createApiClient }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Cancel Delete Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Click delete button
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    // Wait for dialog and cancel
    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.cancel();

    // Verify role still exists
    await orgDetailsPage.roles.waitForLoad();
    await expect(roleRow).toBeVisible();
  });

  test("should not show delete button without role delete permission", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `No Delete Permission Role ${getRandomString(8)}`,
    });

    // Login as read-only user
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    // Verify delete button is not visible
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await expect(deleteButton).not.toBeVisible();
  });

  test("should show delete button for users with delete permission", async ({
    page,
    createApiClient,
    testConfig,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Delete Permission Role ${getRandomString(8)}`,
    });

    // Create user with delete permission
    const deletePermissionUser = await createUser(testConfig);
    await grantActionsToUser(
      testConfig,
      deletePermissionUser.email,
      "Organization",
      organizationId,
      ["organization.read", "role.manage"]
    );

    // Login as user with delete permission
    await loginUser(page, {
      email: deletePermissionUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify delete button is visible
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await expect(deleteButton).toBeVisible();
  });

  test("should remove role from list after deletion", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );

    // Create multiple roles
    const role1Name = `Remove Test Role 1 ${getRandomString(8)}`;
    const role2Name = `Remove Test Role 2 ${getRandomString(8)}`;
    await createRole(apiClient, organizationId, { name: role1Name });
    await createRole(apiClient, organizationId, { name: role2Name });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Verify both roles exist
    const roleRow1 = orgDetailsPage.roles.getRowByRoleName(role1Name);
    const roleRow2 = orgDetailsPage.roles.getRowByRoleName(role2Name);
    await expect(roleRow1).toBeVisible();
    await expect(roleRow2).toBeVisible();

    // Delete first role
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role1Name);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.confirm("Delete");
    await waitForSuccessToast(page, "deleted");

    // Verify first role is removed but second still exists
    await orgDetailsPage.roles.waitForLoad();
    await expect(roleRow1).not.toBeVisible();
    await expect(roleRow2).toBeVisible();
  });

  test("should persist deletion after page reload", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Persist Delete Role ${getRandomString(8)}`,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Delete role
    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.confirm("Delete");
    await waitForSuccessToast(page, "deleted");

    // Reload page
    await page.reload();
    await orgDetailsPage.roles.waitForLoad();

    // Verify role is still gone
    await expect(roleRow).not.toBeVisible();
  });

  test("should keep seeded role templates after deleting a custom role", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const testOrg = await createOrganization(apiClient, {
      name: `Templates Org ${uniqueId}`,
      email: `templates-${uniqueId}@example.com`,
    });

    const roleName = `Custom Role ${getRandomString(8)}`;
    await createRole(apiClient, testOrg.id, { name: roleName });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(testOrg.id);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(roleName);
    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.confirm("Delete");
    await waitForSuccessToast(page, "deleted");

    await orgDetailsPage.roles.waitForLoad();
    expect(await orgDetailsPage.roles.hasRole(roleName)).toBeFalsy();
    expect(await orgDetailsPage.roles.hasEmptyState()).toBeFalsy();
    await expect(
      orgDetailsPage.roles.getRowByRoleName("Organization admin")
    ).toBeVisible();
  });

  test("should delete a custom role bundle", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Custom Bundle ${getRandomString(8)}`,
      actions: ["project.read"],
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();

    const deleteButton = roleRow.getByRole("button", {
      name: /delete role/i,
    });
    await deleteButton.click();

    const dialog = new Dialog(page);
    await dialog.waitFor();
    await dialog.confirm("Delete");
    await waitForSuccessToast(page, "deleted");

    await orgDetailsPage.roles.waitForLoad();
    await expect(roleRow).not.toBeVisible();
  });
});
