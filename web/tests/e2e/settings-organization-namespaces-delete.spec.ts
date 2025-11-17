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

test.describe("@settings.organization-namespaces-delete Organization Namespaces Delete E2E Tests", () => {
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

  test("should show delete namespace button only for members with namespace delete permission", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "delete"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    expect(
      await orgDetailsPage.namespaces.hasDeleteNamespaceButton(namespace.name)
    ).toBeTruthy();

    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();
    expect(
      await orgDetailsPage.namespaces.hasDeleteNamespaceButton(namespace.name)
    ).toBeFalsy();
  });

  test("should allow members with namespace delete permission to delete namespace", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "delete"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    await orgDetailsPage.namespaces.deleteNamespace(namespace.name);
    await orgDetailsPage.namespaces.waitForLoad();
    expect(
      await orgDetailsPage.namespaces.hasNamespace(namespace.name)
    ).toBeFalsy();
  });

  test("should display confirmation dialog when deleting a namespace", async ({
    page,
    testConfig,
  }) => {
    const namespace = await createNamespaceViaApi();

    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Namespace",
      namespace.id,
      "delete"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.namespaces.waitForLoad();

    await orgDetailsPage.namespaces.openDeleteNamespaceDialog(namespace.name);

    const dialog = page.getByRole("alertdialog", {
      name: new RegExp(
        `Are you sure you want to delete ${namespace.name}`,
        "i"
      ),
    });
    await expect(dialog).toBeVisible();
    await expect(
      dialog.getByText(/This will permanently delete the namespace/i)
    ).toBeVisible();
  });
});
