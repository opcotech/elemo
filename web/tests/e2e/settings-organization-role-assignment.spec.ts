import { createOrganization, createRole } from "./api";
import { Dialog } from "./components";
import { expect, test } from "./fixtures";
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

test.describe("@settings.organization-role-assignment Organization Role Assignment E2E Tests", () => {
  let testUser: User;
  let member1: User;
  let member2: User;
  let organizationId: string;
  let roleId: string;
  let roleName: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    testUser = await createUser(testConfig);
    member1 = await createUser(testConfig);
    member2 = await createUser(testConfig);

    // Grant system owner membership so user can create organizations
    await grantSystemOwnerMembershipToUser(testConfig, testUser.email);

    // Create organization via API
    const uniqueId = getRandomString(8);
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Role Assignment ${uniqueId}`,
      email: `test-role-assignment-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    // Add members to organization using DB utility
    await grantMembershipToUser(
      testConfig,
      member1.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      member1.email,
      "Organization",
      organizationId,
      "read"
    );
    await grantMembershipToUser(
      testConfig,
      member2.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      member2.email,
      "Organization",
      organizationId,
      "read"
    );

    // Create a role
    roleName = `Test Role ${getRandomString(8)}`;
    const role = await createRole(apiClient, organizationId, {
      name: roleName,
    });
    roleId = role.id;
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should display current role members", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.members.waitForLoad();

    // The role creator should be listed in the members table
    const fullName = `${testUser.first_name} ${testUser.last_name}`;
    await expect(
      roleEditPage.members.getRowByMemberName(fullName)
    ).toBeVisible();
  });

  test("should allow assigning members to roles", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.members.waitForLoad();

    // Add member using the section helper
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Verify member appears in the list
    await expect(
      roleEditPage.members.getRowByMemberName(member1FullName)
    ).toBeVisible();
  });

  test("should show success message on assignment", async ({ page }) => {
    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, roleId);
    await roleEditPage.members.waitForLoad();

    // Add member2 using the section helper
    const member2FullName = `${member2.first_name} ${member2.last_name}`;
    await roleEditPage.members.addMember(member2FullName);

    // Verify success toast (already waited in addMember)
    await expect(
      page.getByText("Member added to role successfully")
    ).toBeVisible();
  });

  test("should show success message on revocation", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role and add a member for this test
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Revoke Test Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add member1 first
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Now remove the member using the section helper
    await roleEditPage.members.removeMember(member1FullName);

    // Verify success toast (already waited in removeMember)
    await expect(
      page.getByText("Member removed from role successfully")
    ).toBeVisible();
  });

  test("should allow removing members from roles", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role and add a member for this test
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Remove Test Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add member2 first
    const member2FullName = `${member2.first_name} ${member2.last_name}`;
    await roleEditPage.members.addMember(member2FullName);

    // Verify member is in the list
    await expect(
      roleEditPage.members.getRowByMemberName(member2FullName)
    ).toBeVisible();

    // Remove the member using the section helper
    await roleEditPage.members.removeMember(member2FullName);

    // Verify member is removed from the list
    await expect(
      roleEditPage.members.getRowByMemberName(member2FullName)
    ).not.toBeVisible();
  });

  test("should show confirmation dialog when removing member", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role and add a member
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Confirm Remove Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add member1
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Click remove button
    const removeButton =
      roleEditPage.members.getRemoveMemberButton(member1FullName);
    await removeButton.click();

    // Verify dialog shows member info and consequences
    const dialog = new Dialog(page);
    await dialog.waitFor(`Remove ${member1FullName}`);

    await expect(
      page.getByText(
        "The member will lose all permissions assigned to this role"
      )
    ).toBeVisible();
    await expect(
      page.getByText(
        "The member will lose access to all resources assigned to this role"
      )
    ).toBeVisible();

    // Cancel the dialog
    await dialog.cancel();

    // Verify member still exists
    await expect(
      roleEditPage.members.getRowByMemberName(member1FullName)
    ).toBeVisible();
  });

  test("should persist member assignments after page reload", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role and add a member
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Persist Assignment Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add member1
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Reload page
    await page.reload();
    await roleEditPage.members.waitForLoad();

    // Verify member still appears
    await expect(
      roleEditPage.members.getRowByMemberName(member1FullName)
    ).toBeVisible();
  });

  test("should show empty state when no members assigned", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh organization and role with no members
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const emptyOrg = await createOrganization(apiClient, {
      name: `Empty Members Org ${uniqueId}`,
      email: `empty-members-${uniqueId}@example.com`,
    });

    const emptyRole = await createRole(apiClient, emptyOrg.id, {
      name: `Empty Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(emptyOrg.id, emptyRole.id);
    await roleEditPage.members.waitForLoad();

    // The creator should be listed in the members table
    const creatorFullName = `${testUser.first_name} ${testUser.last_name}`;
    await expect(
      roleEditPage.members.getRowByMemberName(creatorFullName)
    ).toBeVisible();
  });

  test("should not show already assigned members in add dialog", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Already Assigned Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add member1 first
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Try to add another member - member1 should not appear in the list
    await roleEditPage.members.clickAddMemberButton();

    const dialog = new Dialog(page);
    await dialog.waitFor("Add Member to Role");

    const selectTrigger = page.getByRole("combobox");
    await selectTrigger.click();

    // Member2 should still be available
    const member2FullName = `${member2.first_name} ${member2.last_name}`;
    const member2Option = page.getByRole("option", { name: member2FullName });
    await expect(member2Option).toBeVisible();

    // Close the dialog by pressing Escape
    await page.keyboard.press("Escape");
  });

  test("should update member count in roles list after assignment", async ({
    page,
    createApiClient,
  }) => {
    // Create a fresh role
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const testRole = await createRole(apiClient, organizationId, {
      name: `Member Count Role ${getRandomString(8)}`,
    });

    const roleEditPage = new SettingsOrganizationRoleEditPage(page);
    await roleEditPage.goto(organizationId, testRole.id);
    await roleEditPage.members.waitForLoad();

    // Add a member
    const member1FullName = `${member1.first_name} ${member1.last_name}`;
    await roleEditPage.members.addMember(member1FullName);

    // Navigate back to organization details to see updated count
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.roles.waitForLoad();

    // Find the role row and verify member count increased (should be 2 members now)
    const roleRow = orgDetailsPage.roles.getRowByRoleName(testRole.name);
    await expect(roleRow.getByText("2 members")).toBeVisible();
  });
});
