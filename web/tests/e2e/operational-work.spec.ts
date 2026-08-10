import { createOrganization, createProject, getRandomProjectKey } from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantSystemOwnerMembershipToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

test.describe("@operational Namespace and project Work surfaces", () => {
  let ownerUser: User;
  let ownerApiClient: Client;
  let organizationId: string;
  let namespaceId: string;
  let namespaceName: string;
  let projectId: string;
  let projectName: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);

    ownerApiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const organization = await createOrganization(ownerApiClient, {
      name: `Operational Work Org ${getRandomString(8)}`,
      email: `operational-work-${getRandomString(8)}@example.com`,
    });
    organizationId = organization.id;

    namespaceName = `Operational Namespace ${getRandomString(8)}`;
    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organizationId },
      body: {
        name: namespaceName,
        description: `Namespace for operational work ${getRandomString(8)}`,
      },
      throwOnError: true,
    });
    namespaceId = namespaceResponse.data.id ?? "";

    projectName = `Operational Project ${getRandomString(8)}`;
    const project = await createProject(ownerApiClient, namespaceId, {
      key: getRandomProjectKey(),
      name: projectName,
      description: `Project for operational work ${getRandomString(8)}`,
    });
    projectId = project.id;
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("opens namespace Work, owns projection URL state, and keeps empty fixtures scoped", async ({
    page,
  }) => {
    await page.goto(`/namespaces/${namespaceId}/work`);
    await expect(
      page.getByRole("heading", { name: `${namespaceName} / Work` })
    ).toBeVisible();
    await expect(page.getByText("No work yet")).toBeVisible();
    await expect(page.getByText("Illustrative work projection")).toBeVisible();

    await page.getByRole("button", { name: "Table", exact: true }).click();
    await expect(page).toHaveURL(/layout=table/);
    await expect(
      page.getByRole("button", { name: "Table", exact: true })
    ).toHaveAttribute("aria-pressed", "true");

    await page.getByRole("button", { name: "List", exact: true }).click();
    await expect(page).toHaveURL(/layout=list/);
    await expect(page.getByText("No work yet")).toBeVisible();
  });

  test("opens project Work from contextual navigation", async ({ page }) => {
    await page.goto(`/namespaces/${namespaceId}/projects/${projectId}`);
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page.getByRole("link", { name: "Work", exact: true }).click();
    await expect(page).toHaveURL(
      new RegExp(`/namespaces/${namespaceId}/projects/${projectId}/work`)
    );
    await expect(
      page.getByRole("heading", { name: `${projectName} / Work` })
    ).toBeVisible();
    await expect(page.getByText("No work yet")).toBeVisible();

    await page.getByRole("button", { name: "Board", exact: true }).click();
    await expect(page).toHaveURL(/layout=board/);
  });

  // FIXME: re-enable once work items can be created via API (LMO-101 is fixture-only).
  test.skip("uses a sheet inspector overlay on tablet for My Work", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 800, height: 900 });
    await page.goto("/my-work?layout=list");
    await expect(page.getByRole("heading", { name: "My Work" })).toBeVisible();
    await expect(page.getByText("2 items")).toBeVisible();

    await page
      .getByRole("button", { name: /^Inspect LMO-101:/ })
      .first()
      .click();
    await expect(page).toHaveURL(/selected=work%3Aele-101/);

    const overlay = page.getByRole("dialog");
    await expect(overlay).toBeVisible();
    await expect(
      overlay.getByRole("complementary", { name: "LMO-101 details" })
    ).toBeVisible();
    await overlay.getByRole("button", { name: "Close inspector" }).click();
    await expect(overlay).not.toBeVisible();
    await expect(page).not.toHaveURL(/selected=/);
  });
});
