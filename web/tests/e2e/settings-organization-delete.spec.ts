import { createOrganization } from "./api";
import { Dialog } from "./components";
import { expect, test } from "./fixtures";
import { waitForPermissionsLoad, waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationsPage,
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

test.describe("@settings.organization-delete Organization Delete E2E Tests", () => {
  let ownerUser: User;

  test.beforeAll(async ({ testConfig }) => {
    ownerUser = await createUser(testConfig);

    // Grant system owner membership so users can create organizations
    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);
  });

  test("should delete organization", async ({ page, createApiClient }) => {
    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );

    const orgToDelete = await createOrganization(apiClient, {
      name: `Org To Delete ${getRandomString(8)}`,
      email: `delete-test-${getRandomString(8)}@example.com`,
    });

    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(orgToDelete.id);
    await orgDetailsPage.dangerZone.waitForLoad();

    // Click delete button
    await orgDetailsPage.dangerZone.clickDeleteButton();

    // Wait for and confirm delete dialog
    const dialog = new Dialog(page);
    await dialog.waitFor("Are you sure you want to delete");
    await dialog.confirm("Delete");

    // Wait for success toast
    await waitForSuccessToast(page, "deleted");

    // Verify organization appears as deleted
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.organizations.waitForLoad();
    expect(
      orgsPage.organizations
        .getRowByOrganizationName(orgToDelete.name)
        .getByText("Deleted")
    ).toBeTruthy();
  });

  test("should delete organization with delete permission", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );

    const orgToDelete = await createOrganization(apiClient, {
      name: `Org To Delete ${getRandomString(8)}`,
      email: `delete-permission-test-${getRandomString(8)}@example.com`,
    });

    const deletePermissionUser = await createUser(testConfig);

    // Grant delete permission to deletePermissionUser for this org
    await grantMembershipToUser(
      testConfig,
      deletePermissionUser.email,
      "Organization",
      orgToDelete.id
    );
    await grantPermissionToUser(
      testConfig,
      deletePermissionUser.email,
      "Organization",
      orgToDelete.id,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      deletePermissionUser.email,
      "Organization",
      orgToDelete.id,
      "delete"
    );

    await loginUser(page, {
      email: deletePermissionUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(orgToDelete.id);
    await orgDetailsPage.dangerZone.waitForLoad();

    // Click delete button
    await orgDetailsPage.dangerZone.clickDeleteButton();

    // Wait for and confirm delete dialog
    const dialog = new Dialog(page);
    await dialog.waitFor("Are you sure you want to delete");
    await dialog.confirm("Delete");

    // Wait for success toast
    await waitForSuccessToast(page, "deleted");

    // Verify organization appears as deleted
    const orgsPage = new SettingsOrganizationsPage(page);
    await orgsPage.organizations.waitForLoad();
    expect(
      orgsPage.organizations
        .getRowByOrganizationName(orgToDelete.name)
        .getByText("Deleted")
    ).toBeTruthy();
  });

  test("should not show delete button without delete permission", async ({
    page,
    createApiClient,
    testConfig,
  }) => {
    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );

    const organization = await createOrganization(apiClient, {
      name: `Test Org ${getRandomString(8)}`,
      email: `test-${getRandomString(8)}@example.com`,
    });

    const noDeletePermissionUser = await createUser(testConfig);

    // Grant membership but only read permission to noDeletePermissionUser
    // This organization is used in the "should not show delete button" test
    await grantMembershipToUser(
      testConfig,
      noDeletePermissionUser.email,
      "Organization",
      organization.id
    );
    await grantPermissionToUser(
      testConfig,
      noDeletePermissionUser.email,
      "Organization",
      organization.id,
      "read"
    );

    await loginUser(page, {
      email: noDeletePermissionUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organization.id);
    await waitForPermissionsLoad(page, organization.id);

    // Verify danger zone is not visible
    expect(await orgDetailsPage.dangerZone.isVisible()).toBeFalsy();
    expect(await orgDetailsPage.dangerZone.hasDeleteButton()).toBeFalsy();
  });
});
