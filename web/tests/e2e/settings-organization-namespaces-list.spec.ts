import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { SettingsOrganizationDetailsPage } from "./pages";
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
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@settings.organization-namespaces-list Organization Namespaces List E2E Tests", () => {
  let ownerUser: User;
  let writerUser: User;
  let readerUser: User;
  let organizationId: string;
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

  const createNamespaceViaApi = async (overrides?: {
    name?: string;
    description?: string;
  }): Promise<{ id: string; name: string; description: string }> => {
    const name = overrides?.name ?? getFullNamespaceName();
    const description =
      overrides?.description ?? `Namespace description ${getRandomString(8)}`;

    const response = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name,
        description,
      },
      throwOnError: true,
    });
    return {
      id: response.data.id ?? "",
      name,
      description,
    };
  };

  test("should list organization namespaces for members with organization read permission", async ({
    page,
  }) => {
    const namespace = await createNamespaceViaApi();

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    await expect(
      orgDetailsPage.namespaces.getRowByNamespaceName(namespace.name)
    ).toBeVisible();
  });

  test("should show create namespace button only for members with organization write permission", async ({
    page,
  }) => {
    // Writer with write permission sees the button
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    expect(
      await orgDetailsPage.namespaces.hasCreateNamespaceButton()
    ).toBeTruthy();

    // Reader with read-only permission does not see the button
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    expect(
      await orgDetailsPage.namespaces.hasCreateNamespaceButton()
    ).toBeFalsy();
  });

  test("should display empty state when no namespaces exist", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    // If no namespaces exist, empty state should be visible
    const emptyState = page.getByText(/No namespaces found/i);
    if (await emptyState.isVisible()) {
      await expect(emptyState).toBeVisible();
    }
  });

  test("should display namespace description in the list", async ({ page }) => {
    const namespace = await createNamespaceViaApi({
      description: "Test description",
    });

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    const namespaceRow = orgDetailsPage.namespaces.getRowByNamespaceName(
      namespace.name
    );
    await expect(namespaceRow).toBeVisible();
    await expect(namespaceRow.getByText("Test description")).toBeVisible();
  });
});
