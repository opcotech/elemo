import { createOrganization, createRole } from "./api";
import { expect, test } from "./fixtures";
import { SettingsOrganizationDetailsPage } from "./pages";
import { loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

const SEEDED_ROLE_NAMES = [
  "Organization admin",
  "Organization member",
  "Namespace admin",
  "Project maintainer",
  "Project viewer",
];

test.describe("@settings.organization-roles Organization Roles List E2E Tests", () => {
  test("should display seeded role templates on a new organization", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    const organization = await createOrganization(apiClient, {
      name: `Test Org Roles Seeded ${getRandomString(8)}`,
      email: `test-roles-seeded-${getRandomString(8)}@example.com`,
    });

    await loginUser(page, ownerPersona.credentials);

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.slug);
    await orgDetailsPage.roles.waitForLoad();

    expect(await orgDetailsPage.roles.hasEmptyState()).toBeFalsy();
    for (const name of SEEDED_ROLE_NAMES) {
      await expect(orgDetailsPage.roles.getRowByRoleName(name)).toBeVisible();
    }
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

    const roles = [
      {
        name: `Role 1 ${getRandomString(8)}`,
        description: `Description for role 1 ${getRandomString(8)}`,
        actions: ["project.read"],
      },
      {
        name: `Role 2 ${getRandomString(8)}`,
        description: `Description for role 2 ${getRandomString(8)}`,
        actions: ["issue.read"],
      },
      {
        name: `Role 3 ${getRandomString(8)}`,
        actions: ["document.read"],
      },
    ];

    const createdRoles = [];
    for (const roleData of roles) {
      const role = await createRole(apiClient, organization.id, roleData);
      createdRoles.push(role);
    }

    await loginUser(page, ownerPersona.credentials);

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.slug);
    await orgDetailsPage.roles.waitForLoad();

    for (const role of createdRoles) {
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name)
      ).toBeVisible();
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name).getByText(role.name)
      ).toBeVisible();
    }

    const roleCount = await orgDetailsPage.roles.getRoleCount();
    expect(roleCount).toBeGreaterThanOrEqual(createdRoles.length);

    for (const role of createdRoles) {
      const row = orgDetailsPage.roles.getRowByRoleName(role.name);
      await expect(row.getByText(role.name)).toBeVisible();
      if (role.description) {
        await expect(row.getByText(role.description)).toBeVisible();
      }
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

    const unique = getRandomString(8);
    const roles = [
      {
        name: `Zebra ${unique}`,
        description: `Alpha description ${unique}`,
        actions: ["project.read"],
      },
      {
        name: `Yak ${unique}`,
        description: `Beta description ${unique}`,
        actions: ["issue.read"],
      },
      {
        name: `Xerus ${unique}`,
        actions: ["document.read"],
      },
    ];

    const createdRoles = [];
    for (const roleData of roles) {
      const role = await createRole(apiClient, organization.id, roleData);
      createdRoles.push(role);
    }

    await loginUser(page, ownerPersona.credentials);

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.slug);
    await orgDetailsPage.roles.waitForLoad();

    await orgDetailsPage.roles.search(createdRoles[0].name);
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[0].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[1].name)
    ).toBeFalsy();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[2].name)
    ).toBeFalsy();

    await orgDetailsPage.roles.search("");
    for (const role of createdRoles) {
      await expect(
        orgDetailsPage.roles.getRowByRoleName(role.name)
      ).toBeVisible();
    }

    await orgDetailsPage.roles.search("Zebra");
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[0].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[1].name)
    ).toBeFalsy();

    await orgDetailsPage.roles.search("Beta");
    await expect(
      orgDetailsPage.roles.getRowByRoleName(createdRoles[1].name)
    ).toBeVisible();
    expect(
      await orgDetailsPage.roles.hasRole(createdRoles[0].name)
    ).toBeFalsy();
  });
});
