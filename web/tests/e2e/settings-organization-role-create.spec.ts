import { createGrant, createOrganization, createRole } from "./api";
import { expect, test } from "./fixtures";
import { getFormFieldMessage, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationRoleCreatePage,
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

test.describe("@settings.organization-role-create Organization Role Creation E2E Tests", () => {
  let testUser: User;
  let readOnlyUser: User;
  let organizationId: string;
  let organizationSlug: string;

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
      name: `Test Org Role Create ${uniqueId}`,
      email: `test-role-create-${uniqueId}@example.com`,
    });
    organizationId = organization.id;
    organizationSlug = organization.slug;

    await grantMembershipToUser(
      testConfig,
      readOnlyUser.email,
      "Organization",
      organizationId
    );
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

  test("should display list of organization roles", async ({ page }) => {
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();

    await expect(orgDetailsPage.roles.getSectionContainer()).toBeVisible();
  });

  test("should allow creating a new role", async ({ page }) => {
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();
    await orgDetailsPage.roles.clickCreateRoleButton();

    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Test Role ${getRandomString(8)}`;
    const roleDescription = `Test role description ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
      Description: roleDescription,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should grant organization.update when a role bundle is assigned at org scope", async ({
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
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.organizationInfo.waitForLoad();
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();

    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Editors ${getRandomString(8)}`,
      description: "Can update the organization.",
      actions: ["organization.update"],
    });
    await createGrant(apiClient, {
      principal: { resourceType: "User", id: scenarioUser.id },
      scope: { resourceType: "Organization", id: organizationId },
      role_id: role.id,
    });

    await loginUser(page, {
      email: scenarioUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.organizationInfo.waitForLoad();
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeTruthy();
  });

  test("should show validation errors for invalid inputs", async ({ page }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();
    const nameError = getFormFieldMessage(page, "Name");

    await roleCreatePage.roleCreateForm.fillFields({
      Name: "AB",
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await expect(nameError).toHaveText(/invalid input/i);
  });

  test("should save role and show success message", async ({ page }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Success Test Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should persist role after page reload", async ({ page }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Persist Test Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    await page.reload();

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should not show create role button without role.manage", async ({
    page,
  }) => {
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();

    expect(await orgDetailsPage.roles.hasCreateRoleButton()).toBeFalsy();
  });

  test("should not show edit role button without role.manage", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Test Role Edit Perm ${getRandomString(8)}`,
      actions: ["project.read"],
    });

    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();
    await expect(
      roleRow.getByRole("button", { name: /edit role/i })
    ).not.toBeVisible();
  });

  test("should not show delete role button without role.manage", async ({
    page,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const role = await createRole(apiClient, organizationId, {
      name: `Test Role Delete Perm ${getRandomString(8)}`,
      actions: ["project.read"],
    });

    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(role.name);
    await expect(roleRow).toBeVisible();
    await expect(
      roleRow.getByRole("button", { name: /delete role/i })
    ).not.toBeVisible();
  });

  test("should create role with only name (description optional)", async ({
    page,
  }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();

    const roleName = `Name Only Role ${getRandomString(8)}`;
    await roleCreatePage.roleCreateForm.fillField("Name", roleName);
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
    await expect(orgDetailsPage.roles.getRowByRoleName(roleName)).toBeVisible();
  });

  test("should allow canceling role creation and return to organization details", async ({
    page,
  }) => {
    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();

    await roleCreatePage.roleCreateForm.fillField(
      "Name",
      `Cancel Test ${getRandomString(8)}`
    );
    await roleCreatePage.roleCreateForm.cancel();

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.roles.waitForLoad();
  });

  test("should show role in list with correct details", async ({ page }) => {
    const roleName = `Details Test Role ${getRandomString(8)}`;
    const roleDescription = `This is a detailed description ${getRandomString(8)}`;

    const roleCreatePage = new SettingsOrganizationRoleCreatePage(page);
    await roleCreatePage.goto(organizationSlug);
    await roleCreatePage.roleCreateForm.waitForLoad();

    await roleCreatePage.roleCreateForm.fillFields({
      Name: roleName,
      Description: roleDescription,
    });
    await roleCreatePage.roleCreateForm.submit("Create Role");
    await waitForSuccessToast(page, "created");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    await orgDetailsPage.roles.waitForLoad();

    const roleRow = orgDetailsPage.roles.getRowByRoleName(roleName);
    await expect(roleRow).toBeVisible();
    await expect(roleRow.getByText(roleName)).toBeVisible();
    await expect(roleRow.getByText(roleDescription)).toBeVisible();
  });
});
