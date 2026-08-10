import type { Page } from "@playwright/test";

import { createOrganization, createProject, getRandomProjectKey } from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser, grantSystemOwnerMembershipToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";

const WORK_LABEL = "LMO-101 Preserve work context across projections";
const WORK_TITLE = "Preserve work context across projections";
const DOCUMENT_LABEL = "Work projection model";
const NAVIGATION_STORAGE_KEY = "elemo_navigation_context";

async function expectRecentTypePersisted(page: Page, type: string) {
  await expect
    .poll(async () =>
      page.evaluate(
        ({ key, entityType }) => {
          const raw = window.localStorage.getItem(key);
          if (!raw) {
            return false;
          }
          try {
            const parsed = JSON.parse(raw) as {
              recentEntities?: { type?: string }[];
            };
            return (
              parsed.recentEntities?.some(
                (entity) => entity.type === entityType
              ) ?? false
            );
          } catch {
            return false;
          }
        },
        { key: NAVIGATION_STORAGE_KEY, entityType: type }
      )
    )
    .toBe(true);
}

test.describe("@operational Recent sidebar entities", () => {
  let ownerUser: User;
  let ownerApiClient: Client;
  let namespaceId: string;
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
      name: `Recent Entities Org ${getRandomString(8)}`,
      email: `recent-entities-${getRandomString(8)}@example.com`,
    });

    const namespaceResponse = await v1OrganizationsNamespacesCreate({
      client: ownerApiClient,
      path: { id: organization.id },
      body: {
        name: `Recent Entities Namespace ${getRandomString(8)}`,
        description: `Namespace for recent entities ${getRandomString(8)}`,
      },
      throwOnError: true,
    });
    namespaceId = namespaceResponse.data.id ?? "";

    projectName = `Recent Entities Project ${getRandomString(8)}`;
    const project = await createProject(ownerApiClient, namespaceId, {
      key: getRandomProjectKey(),
      name: projectName,
      description: `Project for recent entities ${getRandomString(8)}`,
    });
    projectId = project.id;
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: ownerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await page.evaluate((key) => {
      window.localStorage.removeItem(key);
    }, NAVIGATION_STORAGE_KEY);
    await page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  test("remembers work, documents, and projects in sidebar order and supports remove", async ({
    page,
  }) => {
    const sidebar = page.locator(
      '[data-slot="sidebar"][data-variant="sidebar"]'
    );

    await page.goto("/work/lmo-101", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: WORK_TITLE })).toBeVisible();
    await expect(
      sidebar.getByText("Recent Work Items", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: WORK_LABEL, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "work");

    await page.goto("/documents/document-projection-model", {
      waitUntil: "domcontentloaded",
    });
    await expect(
      page.getByRole("heading", { name: DOCUMENT_LABEL })
    ).toBeVisible();
    await expect(
      sidebar.getByText("Recent Documents", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: DOCUMENT_LABEL, exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: WORK_LABEL, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "document");

    await page.goto(`/namespaces/${namespaceId}/projects/${projectId}`, {
      waitUntil: "domcontentloaded",
    });
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();
    await expect(
      sidebar.getByText("Recent Projects", { exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: projectName, exact: true })
    ).toBeVisible();
    await expectRecentTypePersisted(page, "project");

    await page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    const workSection = sidebar.getByText("Recent Work Items", { exact: true });
    const documentsSection = sidebar.getByText("Recent Documents", {
      exact: true,
    });
    const projectsSection = sidebar.getByText("Recent Projects", {
      exact: true,
    });

    await expect(workSection).toBeVisible();
    await expect(documentsSection).toBeVisible();
    await expect(projectsSection).toBeVisible();

    const workBox = await workSection.boundingBox();
    const documentsBox = await documentsSection.boundingBox();
    const projectsBox = await projectsSection.boundingBox();
    expect(workBox).toBeTruthy();
    expect(documentsBox).toBeTruthy();
    expect(projectsBox).toBeTruthy();
    expect(workBox!.y).toBeLessThan(documentsBox!.y);
    expect(documentsBox!.y).toBeLessThan(projectsBox!.y);

    const workLink = sidebar.getByRole("link", {
      name: WORK_LABEL,
      exact: true,
    });
    await workLink.hover();
    await sidebar
      .getByRole("button", { name: `Remove ${WORK_LABEL} from recents` })
      .click();
    await expect(workLink).toHaveCount(0);
    await expect(workSection).toHaveCount(0);

    await page.reload();
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: WORK_LABEL, exact: true })
    ).toHaveCount(0);
    await expect(
      sidebar.getByText("Recent Work Items", { exact: true })
    ).toHaveCount(0);
    await expect(
      sidebar.getByRole("link", { name: DOCUMENT_LABEL, exact: true })
    ).toBeVisible();
    await expect(
      sidebar.getByRole("link", { name: projectName, exact: true })
    ).toBeVisible();

    await sidebar
      .getByRole("link", { name: DOCUMENT_LABEL, exact: true })
      .click();
    await expect(page).toHaveURL("/documents/document-projection-model");
    await expect(
      page.getByRole("heading", { name: DOCUMENT_LABEL })
    ).toBeVisible();
  });
});
