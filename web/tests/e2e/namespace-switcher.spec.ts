import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantOrganizationCreateToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@namespace.switcher Namespace Switcher E2E Tests", () => {
  let ownerUser: User;
  let ownerApiClient: Client;
  let organizationId: string;
  let organizationName: string;
  let namespaceAId: string;
  let namespaceAName: string;
  let namespaceBId: string;
  let namespaceBName: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    await grantOrganizationCreateToUser(testConfig, ownerUser.email);

    ownerApiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );

    const uniqueId = getRandomString(8);
    organizationName = `Switcher Org ${uniqueId}`;
    const organization = await createOrganization(ownerApiClient, {
      name: organizationName,
      email: `switcher-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    namespaceAName = `Namespace A ${uniqueId}`;
    const namespaceAResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: namespaceAName,
        description: `First namespace ${uniqueId}`,
      },
      throwOnError: true,
    });
    namespaceAId = namespaceAResponse.data.id ?? "";

    namespaceBName = `Namespace B ${uniqueId}`;
    const namespaceBResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: namespaceBName,
        description: `Second namespace ${uniqueId}`,
      },
      throwOnError: true,
    });
    namespaceBId = namespaceBResponse.data.id ?? "";
  });

  test("should switch between namespaces and update shell context", async ({
    page,
  }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    const switcher = page.getByRole("button", {
      name: /Switch namespace, current:/,
    });
    await expect(switcher).toBeVisible();
    await switcher.click();

    await page.getByRole("menuitem", { name: namespaceAName }).click();
    await expect(page).toHaveURL(new RegExp(`/namespaces/${namespaceAId}`));
    await expect(
      page.getByRole("button", {
        name: `Switch namespace, current: ${namespaceAName}`,
      })
    ).toBeVisible();

    await page
      .getByRole("button", {
        name: `Switch namespace, current: ${namespaceAName}`,
      })
      .click();
    await page.getByRole("menuitem", { name: namespaceBName }).click();
    await expect(page).toHaveURL(new RegExp(`/namespaces/${namespaceBId}`));
    await expect(
      page.getByRole("button", {
        name: `Switch namespace, current: ${namespaceBName}`,
      })
    ).toBeVisible();

    // Active namespace shows a check; reopen menu to verify both remain listed
    await page
      .getByRole("button", {
        name: `Switch namespace, current: ${namespaceBName}`,
      })
      .click();
    await expect(
      page.getByRole("menuitem", { name: namespaceAName })
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: namespaceBName })
    ).toBeVisible();
    await expect(
      page.locator('[data-slot="dropdown-menu-label"]').filter({
        hasText: organizationName,
      })
    ).toBeVisible();
  });

  test("should show empty state when user has no namespaces", async ({
    page,
    testConfig,
  }) => {
    const lonelyUser = await createUser(testConfig);
    await loginUser(page, {
      email: lonelyUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    const switcher = page.getByRole("button", {
      name: /Switch namespace, current: Choose namespace/,
    });
    await expect(switcher).toBeVisible();
    await switcher.click();
    await expect(page.getByText("No namespaces available.")).toBeVisible();
  });
});
