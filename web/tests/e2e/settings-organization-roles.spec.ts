import { createOrganization, createRole } from "./api";
import { expect, test } from "./fixtures";
import { SettingsOrganizationDetailsPage } from "./pages";
import { loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@settings.organization-roles Organization Roles List E2E Tests", () => {
  test("should show empty state when no roles exist", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Roles Empty ${getRandomString(8)}`,
      email: `test-roles-empty-${getRandomString(8)}@example.com`,
    });

    await loginUser(page, ownerPersona.credentials);

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.id);
    await orgDetailsPage.roles.waitForLoad();

    // Verify empty state is visible
    expect(await orgDetailsPage.roles.hasEmptyState()).toBeTruthy();
    await expect(page.getByText("No roles found")).toBeVisible();
    await expect(
      page.getByText(
        "Roles help organize permissions and member access. Create a role to get started."
      )
    ).toBeVisible();

    // Verify create button is visible
    expect(await orgDetailsPage.roles.hasCreateRoleButton()).toBeTruthy();
  });

  test("should create roles and display in list", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Roles List ${getRandomString(8)}`,
      email: `test-roles-list-${getRandomString(8)}@example.com`,
    });

    // Create multiple roles via API
    const roles = [
      {
        name: `Role 1 ${getRandomString(8)}`,
        description: `Description for role 1 ${getRandomString(8)}`,
      },
      {
        name: `Role 2 ${getRandomString(8)}`,
        description: `Description for role 2 ${getRandomString(8)}`,
      },
      {
        name: `Role 3 ${getRandomString(8)}`,
      },
    ];

    const createdRoles = [];
    for (const roleData of roles) {
      const role = await createRole(apiClient, organization.id, roleData);
      createdRoles.push(role);
    }

    await loginUser(page, ownerPersona.credentials);

    // Navigate to organization details page
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.id);
    await orgDetailsPage.roles.waitForLoad();

    // Verify all roles appear in the table
    for (const role of createdRoles) {
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name)
      ).toBeVisible();
      // Verify role name is displayed
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name).getByText(role.name)
      ).toBeVisible();
    }

    // Verify role count matches
    const roleCount = await orgDetailsPage.roles.getRoleCount();
    expect(roleCount).toBeGreaterThanOrEqual(createdRoles.length);

    // Verify role details are displayed
    for (const role of createdRoles) {
      const row = orgDetailsPage.roles.getRowByRoleName(role.name);

      // Verify name
      await expect(row.getByText(role.name)).toBeVisible();

      // Verify description if present
      if (role.description) {
        await expect(row.getByText(role.description)).toBeVisible();
      }

      // Verify member count badge (should show "1 member" for new roles)
      await expect(row.getByText("1 member")).toBeVisible();
    }
  });

  test("should search roles", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Roles Search ${getRandomString(8)}`,
      email: `test-roles-search-${getRandomString(8)}@example.com`,
    });

    // Create multiple roles with different names
    const roles = [
      {
        name: `Admin Role ${getRandomString(8)}`,
        description: `Admin description ${getRandomString(8)}`,
      },
      {
        name: `Developer Role ${getRandomString(8)}`,
        description: `Developer description ${getRandomString(8)}`,
      },
      {
        name: `Manager Role ${getRandomString(8)}`,
      },
    ];

    const createdRoles = [];
    for (const roleData of roles) {
      const role = await createRole(apiClient, organization.id, roleData);
      createdRoles.push(role);
    }

    await loginUser(page, ownerPersona.credentials);

    // Navigate to organization details page
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.id);
    await orgDetailsPage.roles.waitForLoad();

    // Search for a specific role name
    await orgDetailsPage.roles.search(createdRoles[0].name);

    // Verify only matching role is visible
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[0].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[1].name)
    ).toBeFalsy();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[2].name)
    ).toBeFalsy();

    // Clear search and verify all roles are visible again
    await orgDetailsPage.roles.search("");
    for (const role of createdRoles) {
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name)
      ).toBeVisible();
    }

    // Test search by partial name
    await orgDetailsPage.roles.search("Admin");
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[0].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[1].name)
    ).toBeFalsy();

    // Test search by description
    await orgDetailsPage.roles.search("Developer");
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[1].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[0].name)
    ).toBeFalsy();
  });
});
