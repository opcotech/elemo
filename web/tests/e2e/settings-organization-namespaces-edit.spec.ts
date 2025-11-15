import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import {
  SettingsOrganizationDetailsPage,
  SettingsOrganizationNamespaceDetailsPage,
  SettingsOrganizationNamespaceEditPage,
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
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@settings.organization-namespaces-edit Organization Namespaces Edit E2E Tests", () => {
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

  test("should allow members with namespace read permission to view namespace details", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      namespace.id,
      "read"
    );

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceDetailsPage = new SettingsOrganizationNamespaceDetailsPage(
      page
    );
    await namespaceDetailsPage.goto(organizationId, namespace.id);
    await namespaceDetailsPage.waitForLoad();

    expect(await namespaceDetailsPage.getTitleText()).toContain(namespace.name);
  });

  test("should allow members with namespace write permission to update namespace", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "write"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const namespaceEditPage = new SettingsOrganizationNamespaceEditPage(page);
    await namespaceEditPage.goto(organizationId, namespace.id);
    const updatedName = `Updated ${getRandomString(6)}`;
    await namespaceEditPage.namespaceForm.clearField("Name");
    await namespaceEditPage.namespaceForm.fillFields({
      Name: updatedName,
      Description: `Updated description ${getRandomString(6)}`,
    });
    await namespaceEditPage.namespaceForm.submit("Save Changes");
    await waitForSuccessToast(page, "Namespace updated");

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    await expect(
      orgDetailsPage.namespaces.getRowByNamespaceName(updatedName)
    ).toBeVisible();
  });

  test("should show edit button only for members with namespace write permission", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Namespace",
      namespace.id,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "write"
    );

    // Writer with write permission should see edit button
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
    // Edit button is a link with sr-only text, so we check for the link
    const editLink = namespaceRow
      .getByRole("link")
      .filter({ hasText: /edit namespace/i });
    await expect(editLink).toBeVisible();

    // Reader with read-only permission should not see edit button
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    const readerNamespaceRow = orgDetailsPage.namespaces.getRowByNamespaceName(
      namespace.name
    );
    await expect(readerNamespaceRow).toBeVisible();
    await expect(
      readerNamespaceRow
        .getByRole("link")
        .filter({ hasText: /edit namespace/i })
    ).not.toBeVisible();
  });
});
