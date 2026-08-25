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
import { loginUser } from "./utils/auth";
import { getRandomString } from "./utils/random";

test.describe("@settings.organization-create Organization Creation E2E Tests", () => {
  test("should create organization and display in list", async ({
    page,
    ownerPersona,
  }) => {
    await loginUser(page, ownerPersona.credentials);

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
    await orgCreatePage.organizationCreateForm.submit("Create Organization");

    // Then check for success toast
    await waitForSuccessToast(page, "created");

    // Verify organization details page shows the created organization
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.organizationInfo.waitForLoad();

    // Navigate back to list and verify organization appears
    await orgsPage.goto();
    await orgsPage.organizations.waitForLoad();
    await expect(
      orgsPage.organizations.getRowByOrganizationName(orgName)
    ).toBeVisible();
  });

  test("should show validation errors for invalid form inputs", async ({
    page,
    ownerPersona,
  }) => {
    const fieldMessage = (label: string) => getFormFieldMessage(page, label);

    await loginUser(page, ownerPersona.credentials);

    const orgCreatePage = new SettingsOrganizationCreatePage(page);
    await orgCreatePage.goto();
    await orgCreatePage.organizationCreateForm.waitForLoad();

    // Try submitting empty form
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await expect(fieldMessage("Name")).toHaveText(/invalid input/i);
    await expect(fieldMessage("Slug")).toHaveText(/required/i);
    await expect(fieldMessage("Email")).toHaveText(/invalid input/i);

    // Fill name but invalid email
    await orgCreatePage.organizationCreateForm.fillField("Name", "Test Org");
    await orgCreatePage.organizationCreateForm.fillField(
      "Email",
      "invalid-email"
    );
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await expect(fieldMessage("Name")).toHaveCount(0);
    await expect(fieldMessage("Email")).toHaveText(/invalid input/i);

    // Fill valid email but empty name
    await orgCreatePage.organizationCreateForm.fillField("Name", "");
    await orgCreatePage.organizationCreateForm.fillField(
      "Email",
      "test@example.com"
    );
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await expect(fieldMessage("Name")).toHaveText(/invalid input/i);
    await expect(fieldMessage("Email")).toHaveCount(0);

    await orgCreatePage.organizationCreateForm.fillFields({
      Name: "Valid Organization",
      Slug: "AB",
      Email: "test@example.com",
    });
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await expect(fieldMessage("Slug")).toHaveText(/3–50 characters/i);

    await orgCreatePage.organizationCreateForm.fillField("Slug", "new");
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await expect(fieldMessage("Slug")).toHaveText(/reserved/i);
  });

  test("should show error when creating duplicate organization", async ({
    page,
    ownerPersona,
    createApiClient,
  }) => {
    const orgData = {
      name: `Existing Org ${getRandomString()}`,
      email: `duplicate-${getRandomString()}@example.com`,
    };

    // Create organization via API first
    const apiClient = await createApiClient(
      ownerPersona.credentials.email,
      ownerPersona.credentials.password
    );
    await createOrganization(apiClient, orgData);

    await loginUser(page, ownerPersona.credentials);

    // Wait for create page to load
    const orgCreatePage = new SettingsOrganizationCreatePage(page);
    await orgCreatePage.goto();
    await orgCreatePage.organizationCreateForm.waitForLoad();

    // Fill form with existing organization name
    await orgCreatePage.organizationCreateForm.fillFields({
      Name: orgData.name,
      Email: `${getRandomString()}@example.com`,
    });
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await waitForErrorToast(page);

    // Fill form with existing organization email
    await orgCreatePage.organizationCreateForm.fillFields({
      Name: `${getRandomString()}`,
      Email: orgData.email,
    });
    await orgCreatePage.organizationCreateForm.submit("Create Organization");
    await waitForErrorToast(page);
  });

  test("should not see the create organization button without create permission", async ({
    page,
    userPersona,
  }) => {
    await loginUser(page, userPersona.credentials);

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
