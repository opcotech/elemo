import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import {
  getFormFieldMessage,
  waitForErrorToast,
  waitForSuccessToast,
} from "./helpers";
import {
  SettingsOrganizationCreatePage,
  SettingsOrganizationDetailsPage,
  SettingsOrganizationsPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantSystemOwnerMembershipToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api";

test.describe("@settings.organization-create Organization Creation E2E Tests", () => {
  let testUser: User;
  let readOnlyUser: User;

  test.beforeAll(async ({ testConfig }) => {
    testUser = await createUser(testConfig);
    readOnlyUser = await createUser(testConfig);

    // Grant system owner membership so user can create organizations
    // Using DB helper since API doesn't allow creating system-level permissions
    await grantSystemOwnerMembershipToUser(testConfig, testUser.email);
  });

  test("should create organization and display in list", async ({ page }) => {
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Click create button
    await orgsPage.organizations.clickCreateOrganizationButton();

    const orgCreatePage = new SettingsOrganizationCreatePage(page);
    await orgCreatePage.organizationCreateForm.waitForLoad();

    // Fill organization form
    const orgName = `Test Org ${Date.now()}`;
    const orgEmail = `test-${Date.now()}@example.com`;
    await orgCreatePage.organizationCreateForm.fillFields({
      Name: orgName,
      Email: orgEmail,
    });
    await orgCreatePage.organizationCreateForm.submit("Create");

    // Then check for success toast
    await waitForSuccessToast(page, "created");

    // Verify organization details page shows the created organization
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.organizationInfo.waitForLoad();

    // Navigate back to list and verify organization appears
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();
    expect(await orgsPage.organizations.hasOrganization(orgName)).toBeTruthy();
  });

  test("should show validation errors for invalid form inputs", async ({
    page,
  }) => {
    const fieldMessage = (label: string) => getFormFieldMessage(page, label);

    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgCreatePage = new SettingsOrganizationCreatePage(page);
    await orgCreatePage.goto();
    await orgCreatePage.organizationCreateForm.waitForLoad();

    // Try submitting empty form
    await orgCreatePage.organizationCreateForm.submit("Create");
    await expect(fieldMessage("Name")).toHaveText(/invalid input/i);
    await expect(fieldMessage("Email")).toHaveText(/invalid input/i);

    // Fill name but invalid email
    await orgCreatePage.organizationCreateForm.fillField("Name", "Test Org");
    await orgCreatePage.organizationCreateForm.fillField(
      "Email",
      "invalid-email"
    );
    await orgCreatePage.organizationCreateForm.submit("Create");
    await expect(fieldMessage("Name")).toHaveCount(0);
    await expect(fieldMessage("Email")).toHaveText(/invalid input/i);

    // Fill valid email but empty name
    await orgCreatePage.organizationCreateForm.fillField("Name", "");
    await orgCreatePage.organizationCreateForm.fillField(
      "Email",
      "test@example.com"
    );
    await orgCreatePage.organizationCreateForm.submit("Create");
    await expect(fieldMessage("Name")).toHaveText(/invalid input/i);
    await expect(fieldMessage("Email")).toHaveCount(0);
  });

  test("should show error when creating duplicate organization", async ({
    page,
    createApiClient,
  }) => {
    const orgData = {
      name: `Existing Org ${getRandomString()}`,
      email: `duplicate-${getRandomString()}@example.com`,
    };

    // Create organization via API first
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    await createOrganization(apiClient, orgData);

    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Wait for create page to load
    const orgCreatePage = new SettingsOrganizationCreatePage(page);
    await orgCreatePage.goto();
    await orgCreatePage.organizationCreateForm.waitForLoad();

    // Fill form with existing organization name
    await orgCreatePage.organizationCreateForm.fillFields({
      Name: orgData.name,
      Email: `${getRandomString()}@example.com`,
    });
    await orgCreatePage.organizationCreateForm.submit("Create");
    await waitForErrorToast(page);

    // Fill form with existing organization email
    await orgCreatePage.organizationCreateForm.fillFields({
      Name: `${getRandomString()}`,
      Email: orgData.email,
    });
    await orgCreatePage.organizationCreateForm.submit("Create");
    await waitForErrorToast(page);
  });

  test("should not see the create organization button without create permission", async ({
    page,
  }) => {
    await loginUser(page, {
      email: readOnlyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    // Wait for organizations page to load
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();

    // Verify the create button is not visible
    expect(
      await orgsPage.organizations.hasCreateOrganizationButton()
    ).toBeFalsy();
  });
});
