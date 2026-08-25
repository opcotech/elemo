import { createOrganization, createProjectDocument } from "./api";
import { expect, test } from "./fixtures";
import { seedOwnerWorkspace } from "./helpers";
import type { OwnerWorkspace } from "./helpers";
import { DocumentPage, DocumentsListPage, WorkItemPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { getRandomSlug, getRandomString } from "./utils/random";

import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import { namespacePath } from "@/lib/paths";

test.describe("@routes.slugs Hierarchical URL identity E2E Tests", () => {
  let workspace: OwnerWorkspace;
  let secondOrganizationSlug: string;
  let secondNamespaceName: string;
  let sharedNamespaceSlug: string;
  let otherNamespaceSlug: string;

  test.beforeAll(async ({ testConfig }) => {
    sharedNamespaceSlug = getRandomSlug("platform");
    workspace = await seedOwnerWorkspace(testConfig, {
      namePrefix: "Slug Routes",
      namespaceSlug: sharedNamespaceSlug,
    });

    const secondOrganization = await createOrganization(workspace.client, {
      name: `Slug Routes Second ${getRandomString(8)}`,
      email: `slug-second-${getRandomString(8)}@example.com`,
    });
    secondOrganizationSlug = secondOrganization.slug;
    secondNamespaceName = `Second Platform ${getRandomString(8)}`;
    otherNamespaceSlug = getRandomSlug("ops");

    await v1OrganizationsNamespacesCreate({
      client: workspace.client,
      path: { organizationRef: secondOrganization.id },
      body: {
        name: secondNamespaceName,
        slug: sharedNamespaceSlug,
        description: "Same slug in a second organization",
      },
      throwOnError: true,
    });
    await v1OrganizationsNamespacesCreate({
      client: workspace.client,
      path: { organizationRef: secondOrganization.id },
      body: {
        name: `Other NS ${getRandomString(8)}`,
        slug: otherNamespaceSlug,
      },
      throwOnError: true,
    });
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, {
      email: workspace.owner.email,
      password: USER_DEFAULT_PASSWORD,
    });
  });

  test("should 404 xid-shaped organization and namespace segments", async ({
    page,
  }) => {
    await page.goto(`/organizations/${workspace.organizationId}`);
    await expect(page.getByText("Organization not found")).toBeVisible();

    await page.goto(
      `/organizations/${workspace.organizationSlug}/namespaces/${workspace.namespaceId}`
    );
    await expect(page.getByText("not found", { exact: false })).toBeVisible();
  });

  test("should 404 reserved slug lookups", async ({ page }) => {
    await page.goto("/organizations/new");
    await expect(page.getByText("Organization not found")).toBeVisible();

    await page.goto(
      `/organizations/${workspace.organizationSlug}/namespaces/new`
    );
    await expect(page.getByText("not found", { exact: false })).toBeVisible();
  });

  test("should open namespace documents without a scope query param", async ({
    page,
  }) => {
    const listPage = new DocumentsListPage(page);
    await listPage.gotoNamespace(
      workspace.organizationSlug,
      workspace.namespaceSlug
    );
    await expect(page).toHaveURL(
      new RegExp(
        `/organizations/${workspace.organizationSlug}/namespaces/${workspace.namespaceSlug}/documents$`
      )
    );
    await listPage.waitForLoad();
  });

  test("should keep work items on the hierarchical path", async ({ page }) => {
    const workItem = new WorkItemPage(page);
    await workItem.goto(
      workspace.organizationSlug,
      workspace.namespaceSlug,
      "NO-SUCH-1"
    );
    await expect(page.getByText("not found", { exact: false })).toBeVisible();
  });

  test("should 404 noncanonical organization slugs", async ({ page }) => {
    await page.goto("/organizations/AB");
    await expect(page.getByText("not found", { exact: false })).toBeVisible();
  });

  test("should 404 a namespace slug that belongs to another organization", async ({
    page,
  }) => {
    await page.goto(
      namespacePath({
        organizationSlug: workspace.organizationSlug,
        namespaceSlug: otherNamespaceSlug,
      })
    );
    await expect(page.getByText("not found", { exact: false })).toBeVisible();
  });

  test("should resolve the same namespace slug under two organizations", async ({
    page,
  }) => {
    await page.goto(
      namespacePath({
        organizationSlug: workspace.organizationSlug,
        namespaceSlug: sharedNamespaceSlug,
      })
    );
    await expect(
      page.getByRole("heading", { name: workspace.namespaceName })
    ).toBeVisible();

    await page.goto(
      namespacePath({
        organizationSlug: secondOrganizationSlug,
        namespaceSlug: sharedNamespaceSlug,
      })
    );
    await expect(
      page.getByRole("heading", { name: secondNamespaceName })
    ).toBeVisible();
  });

  test("should open document details without scope query params", async ({
    page,
  }) => {
    const document = await createProjectDocument(
      workspace.client,
      workspace.projectId,
      { title: `Slug Doc ${getRandomString(8)}` }
    );

    const documentPage = new DocumentPage(page);
    await documentPage.goto(document.id);
    await documentPage.waitForLoad();

    await expect(page).toHaveURL(new RegExp(`/documents/${document.id}$`));
    expect(new URL(page.url()).search).toBe("");
  });
});
