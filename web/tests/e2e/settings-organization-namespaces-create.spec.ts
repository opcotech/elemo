import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationNamespaceCreatePage,
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
import type { Client } from "@/lib/client/client";

test.describe("@settings.organization-namespaces-create Organization Namespaces Create E2E Tests", () => {
  let ownerUser: User;
  let writerUser: User;
  let readerUser: User;
  let organizationId: string;
  let organizationName: string;
  let ownerApiClient: Client;

  const getFullNamespaceName = () => `Namespace ${getRandomString(8)}`;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    writerUser = await createUser(testConfig);
    readerUser = await createUser(testConfig);

    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);

    ownerApiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(ownerApiClient, {
      name: `Namespaces Org ${getRandomString(8)}`,
      email: `namespaces-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;
    organizationName = organization.name;

    await grantMembershipToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      "write"
    );

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId,
      "read"
    );
  });

  test("should allow creating a namespace via global form when user has organization write permission", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceCreatePage = new SettingsOrganizationNamespaceCreatePage(
      page
    );
    await namespaceCreatePage.gotoGlobal();
    await namespaceCreatePage.selectOrganization(organizationName);

    const namespaceName = getFullNamespaceName();
    const namespaceDescription = `Namespace description ${getRandomString(8)}`;
    await namespaceCreatePage.namespaceForm.fillFields({
      Name: namespaceName,
      Description: namespaceDescription,
    });
    await namespaceCreatePage.namespaceForm.submit("Create Namespace");
    await waitForSuccessToast(page, "Namespace created");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    await expect(
      orgDetailsPage.namespaces.getRowByNamespaceName(namespaceName)
    ).toBeVisible();
  });

  test("should navigate to create namespace page when clicking create button", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    await orgDetailsPage.namespaces.clickCreateNamespaceButton();

    // Should navigate to create namespace page
    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationId}/namespaces/new`)
    );

    // Verify form fields are present
    await expect(page.getByRole("textbox", { name: "Name" })).toBeVisible();
  });

  test("should allow creating namespace with only name field", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceCreatePage = new SettingsOrganizationNamespaceCreatePage(
      page
    );
    await namespaceCreatePage.gotoGlobal();
    await namespaceCreatePage.selectOrganization(organizationName);

    const namespaceName = getFullNamespaceName();
    await namespaceCreatePage.namespaceForm.fillFields({
      Name: namespaceName,
    });
    await namespaceCreatePage.namespaceForm.submit("Create Namespace");
    await waitForSuccessToast(page, "Namespace created");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    await expect(
      orgDetailsPage.namespaces.getRowByNamespaceName(namespaceName)
    ).toBeVisible();
  });
});
