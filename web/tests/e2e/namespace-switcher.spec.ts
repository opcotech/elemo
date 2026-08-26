import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { fillLocator } from "./helpers";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantOrganizationCreateToUser } from "./utils/db";
import { getRandomSlug, getRandomString } from "./utils/random";

import type { Client } from "@/lib/api/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type { User } from "@/lib/api/types";

test.describe("@namespace.switcher Namespace Switcher E2E Tests", () => {
  let ownerUser: User;
  let ownerApiClient: Client;
  let organizationId: string;
  let organizationSlug: string;
  let organizationName: string;
  let namespaceASlug: string;
  let namespaceAName: string;
  let namespaceBSlug: string;
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
    organizationSlug = organization.slug;

    namespaceAName = `Alpha NS ${uniqueId}`;
    namespaceASlug = getRandomSlug("ns");
    await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { organizationRef: organizationId },
      body: {
        slug: namespaceASlug,
        name: namespaceAName,
        description: `First namespace ${uniqueId}`,
      },
      throwOnError: true,
    });

    namespaceBName = `Zulu NS ${uniqueId}`;
    namespaceBSlug = getRandomSlug("ns");
    await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { organizationRef: organizationId },
      body: {
        slug: namespaceBSlug,
        name: namespaceBName,
        description: `Second namespace ${uniqueId}`,
      },
      throwOnError: true,
    });
  });

  test("should switch between namespaces and update shell context", async ({
    page,
  }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    const switcher = page.getByRole("combobox", {
      name: /Switch namespace, current:/,
    });
    await expect(switcher).toBeVisible();
    await switcher.click();

    await page.getByRole("option", { name: namespaceAName }).click();
    await expect(page).toHaveURL(
      new RegExp(
        `/organizations/${organizationSlug}/namespaces/${namespaceASlug}`
      )
    );
    await expect(
      page.getByRole("combobox", {
        name: `Switch namespace, current: ${namespaceAName}`,
      })
    ).toBeVisible();

    await page
      .getByRole("combobox", {
        name: `Switch namespace, current: ${namespaceAName}`,
      })
      .click();
    await page.getByRole("option", { name: namespaceBName }).click();
    await expect(page).toHaveURL(
      new RegExp(
        `/organizations/${organizationSlug}/namespaces/${namespaceBSlug}`
      )
    );
    await expect(
      page.getByRole("combobox", {
        name: `Switch namespace, current: ${namespaceBName}`,
      })
    ).toBeVisible();

    // Active namespace shows a check; reopen menu to verify both remain listed
    await page
      .getByRole("combobox", {
        name: `Switch namespace, current: ${namespaceBName}`,
      })
      .click();
    await expect(
      page.getByRole("option", { name: namespaceAName })
    ).toBeVisible();
    await expect(
      page.getByRole("option", { name: namespaceBName })
    ).toBeVisible();
    await expect(
      page.locator("[cmdk-group-heading]").filter({
        hasText: organizationName,
      })
    ).toBeVisible();
  });

  test("should filter namespaces by search query", async ({ page }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    await page
      .getByRole("combobox", {
        name: /Switch namespace, current:/,
      })
      .click();

    const search = page.getByPlaceholder("Search namespaces...");
    await expect(search).toBeVisible();
    await fillLocator(search, "Zulu");
    await expect(
      page.getByRole("option", { name: namespaceBName })
    ).toBeVisible();
    await expect(
      page.getByRole("option", { name: namespaceAName })
    ).toBeHidden();

    await page.getByRole("option", { name: namespaceBName }).click();
    await expect(page).toHaveURL(
      new RegExp(
        `/organizations/${organizationSlug}/namespaces/${namespaceBSlug}`
      )
    );
    await expect(
      page.getByRole("combobox", {
        name: `Switch namespace, current: ${namespaceBName}`,
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

    const switcher = page.getByRole("combobox", {
      name: /Switch namespace, current: Choose namespace/,
    });
    await expect(switcher).toBeVisible();
    await switcher.click();
    await expect(page.getByText("No namespaces available.")).toBeVisible();
  });
});
