import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { waitForErrorToast, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationEditPage,
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

test.describe("@settings.organization-edit Organization Edit E2E Tests", () => {
  let ownerUser: User;
  let memberUser: User;
  let readOnlyMemberUser: User;
  let organizationId: string;
  let organizationName: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    memberUser = await createUser(testConfig);
    readOnlyMemberUser = await createUser(testConfig);

    // Grant system owner membership so user can create organizations
    // Using DB helper since API doesn't allow creating system-level permissions
    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);

    // Create organization via API first with unique name
    const uniqueId = getRandomString(8);
    organizationName = `Test Org ${uniqueId}`;
    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(apiClient, {
      name: organizationName,
      email: `test-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    // Grant membership to member user with write permission
    await grantMembershipToUser(
      testConfig,
      memberUser.email,
      "Organization",
      organization.id
    );
    await grantPermissionToUser(
      testConfig,
      memberUser.email,
      "Organization",
      organization.id,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      memberUser.email,
      "Organization",
      organization.id,
      "write"
    );

    // Grant membership to read-only member user with only read permission
    await grantMembershipToUser(
      testConfig,
      readOnlyMemberUser.email,
      "Organization",
      organization.id
    );
    await grantPermissionToUser(
      testConfig,
      readOnlyMemberUser.email,
      "Organization",
      organization.id,
      "read"
    );
  });

  test("should edit organization", async ({ page }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();

    // Click edit button
    await orgDetailsPage.organizationInfo.clickEditOrganizationButton();

    // Wait for edit page to load
    const orgEditPage = new SettingsOrganizationEditPage(page);
    await orgEditPage.organizationEditForm.waitForLoad();

    // Fill form
    const updatedEmail = `updated-${getRandomString()}@example.com`;
    await orgEditPage.organizationEditForm.fillField("Email", updatedEmail);
    await orgEditPage.organizationEditForm.submit("Save Changes");

    // Wait for success toast
    await waitForSuccessToast(page, "updated");
  });

  test("should show validation errors for invalid form inputs", async ({
    page,
  }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgEditPage = new SettingsOrganizationEditPage(page);
    await orgEditPage.goto(organizationId);
    await orgEditPage.organizationEditForm.waitForLoad();

    // Try submitting with empty name
    await orgEditPage.organizationEditForm.clearField("Name");
    await orgEditPage.organizationEditForm.fillField(
      "Email",
      "test@example.com"
    );
    await orgEditPage.organizationEditForm.submit("Save Changes");
    await expect(
      page.getByText(/too small: expected string to have >=1 characters/i)
    ).toBeVisible();

    // Try submitting with empty email
    await orgEditPage.organizationEditForm.fillField("Name", "Test Org");
    await orgEditPage.organizationEditForm.clearField("Email");
    await orgEditPage.organizationEditForm.submit("Save Changes");
    await expect(page.getByText(/invalid email address/i)).toBeVisible();

    // Fill name but invalid email
    await orgEditPage.organizationEditForm.fillField("Name", "Test Org");
    await orgEditPage.organizationEditForm.fillField("Email", "invalid-email");
    await orgEditPage.organizationEditForm.submit("Save Changes");
    await expect(page.getByText(/invalid email address/i)).toBeVisible();
  });

  test("should show error when updating to a duplicate organization", async ({
    page,
    createApiClient,
  }) => {
    // Create another organization via API
    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const duplicateOrg = await createOrganization(apiClient, {
      name: `Duplicate Org ${getRandomString()}`,
      email: `duplicate-${getRandomString()}@example.com`,
    });

    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgEditPage = new SettingsOrganizationEditPage(page);
    await orgEditPage.goto(organizationId);
    await orgEditPage.organizationEditForm.waitForLoad();

    // Try updating with duplicate name
    await orgEditPage.organizationEditForm.fillField("Name", duplicateOrg.name);
    await orgEditPage.organizationEditForm.submit("Save Changes");
    await waitForErrorToast(page);

    // Try updating with duplicate email
    await orgEditPage.organizationEditForm.fillField("Name", organizationName);
    await orgEditPage.organizationEditForm.fillField(
      "Email",
      duplicateOrg.email
    );
    await orgEditPage.organizationEditForm.submit("Save Changes");
    await waitForErrorToast(page);
  });

  test("should see the edit button with write permission for non-owner", async ({
    page,
  }) => {
    await loginUser(page, {
      email: memberUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();

    // Verify the edit button is visible
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeTruthy();
  });

  test("should not see the edit button without write permission", async ({
    page,
  }) => {
    await loginUser(page, {
      email: readOnlyMemberUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.organizationInfo.waitForLoad();

    // Verify the edit button is visible
    expect(
      await orgDetailsPage.organizationInfo.hasEditOrganizationButton()
    ).toBeFalsy();
  });
});
