import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationsPage,
} from "./pages";
import { loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@settings.organizations Organization Listing E2E Tests", () => {
  test("should display organizations table", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const uniqueId = getRandomString(8);
    const organizations = [
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
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    for (const organization of organizations) {
      await createOrganization(apiClient, {
        name: organization.name,
        email: organization.email,
      });
    }

    await loginUser(page, ownerPersona.credentials);

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

  test("should search organizations", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const uniqueId = getRandomString(8);
    const organizations = [
      {
        name: `Search Org 1 ${uniqueId}`,
        email: `search1-${uniqueId}@example.com`,
      },
      {
        name: `Search Org 2 ${uniqueId}`,
        email: `search2-${uniqueId}@example.com`,
      },
      {
        name: `Search Org 3 ${uniqueId}`,
        email: `search3-${uniqueId}@example.com`,
      },
    ];

    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    for (const organization of organizations) {
      await createOrganization(apiClient, {
        name: organization.name,
        email: organization.email,
      });
    }

    await loginUser(page, ownerPersona.credentials);

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

  test("should navigate to organization details", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const uniqueId = getRandomString(8);
    const organizationName = `Navigate Org ${uniqueId}`;

    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    await createOrganization(apiClient, {
      name: organizationName,
      email: `navigate-${uniqueId}@example.com`,
    });

    await loginUser(page, ownerPersona.credentials);

    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Click on organization
    await orgsPage.organizations.clickOrganization(organizationName);

    // Verify navigation to organization details page
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.organizationInfo.waitForLoad();
  });
});
