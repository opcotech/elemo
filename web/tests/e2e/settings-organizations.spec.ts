import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationsPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantSystemOwnerMembershipToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api";

test.describe("@settings.organizations Organization Listing E2E Tests", () => {
  let testUser: User;
  let organizations: { name: string; email: string }[];

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    testUser = await createUser(testConfig);

    // Grant system owner membership so user can create organizations
    // Using DB helper since API doesn't allow creating system-level permissions
    await grantSystemOwnerMembershipToUser(testConfig, testUser.email);

    // Create unique organizations to avoid conflicts
    const uniqueId = getRandomString(8);
    organizations = [
      {
        name: `Test Org 1 ${uniqueId}`,
        email: `test1-${uniqueId}@example.com`,
      },
      {
        name: `Test Org 2 ${uniqueId}`,
        email: `test2-${uniqueId}@example.com`,
      },
      {
        name: `Test Org 3 ${uniqueId}`,
        email: `test3-${uniqueId}@example.com`,
      },
    ];

    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    for (const organization of organizations) {
      await createOrganization(apiClient, {
        name: organization.name,
        email: organization.email,
      });
    }
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should display organizations table", async ({ page }) => {
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Verify table is visible
    const table = orgsPage.organizations.getTable();
    await expect(table).toBeVisible();

    // Verify all organizations are displayed
    for (const org of organizations) {
      await expect(
        orgsPage.organizations.getRowByOrganizationName(org.name)
      ).toBeVisible();
    }

    // Verify organization count matches
    const count = await orgsPage.organizations.getOrganizationCount();
    expect(count).toBeGreaterThanOrEqual(organizations.length);
  });

  test("should search organizations", async ({ page }) => {
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Search for first organization
    await orgsPage.organizations.search(organizations[0].name);

    // Verify only matching organization is visible
    await expect(
      orgsPage.organizations.getRowByOrganizationName(organizations[0].name)
    ).toBeVisible();
    await expect(
      orgsPage.organizations.hasOrganization(organizations[1].name)
    ).resolves.toBe(false);
    await expect(
      orgsPage.organizations.hasOrganization(organizations[2].name)
    ).resolves.toBe(false);

    // Clear search and verify all organizations are visible again
    await orgsPage.organizations.search("");
    for (const org of organizations) {
      await expect(
        orgsPage.organizations.getRowByOrganizationName(org.name)
      ).toBeVisible();
    }
  });

  test("should navigate to organization details", async ({ page }) => {
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Click on first organization
    await orgsPage.organizations.clickOrganization(organizations[0].name);

    // Verify navigation to organization details page
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.organizationInfo.waitForLoad();
  });
});
