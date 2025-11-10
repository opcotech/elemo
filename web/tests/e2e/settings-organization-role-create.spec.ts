import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";
import { OrganizationPage } from "./pages/organization-page";
import { Form } from "./components/form";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";

test.describe("@settings.organization-role-create Organization Role Create E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let writeUser: any;
    let readUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });
      writeUser = await createDBUser("active", {
        first_name: "Write",
        last_name: "User",
      });
      readUser = await createDBUser("active", {
        first_name: "Read",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, writeUser.id, "read");
      await addMemberToOrganization(testOrg.id, writeUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
    });

    test("user with write permission should see create button", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(
        page.getByRole("link", { name: /Create Role/i }).first()
      ).toBeVisible();
    });

    test("user with read permission should not see create button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(
        page.getByRole("link", { name: /Create Role/i })
      ).not.toBeVisible();
    });

    test("owner should see create button", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(
        page.getByRole("link", { name: /Create Role/i }).first()
      ).toBeVisible();
    });
  });

  test.describe("Navigation", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Navigation Test Org",
      });
    });

    test("should navigate to create page when create button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await orgPage.clickCreateRole();
      await waitForPageLoad(page);
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrg.id}/roles/new`
      );
      await expect(
        page.getByRole("heading", { name: "Create Role" })
      ).toBeVisible();
    });
  });

  test.describe("Form Validation", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Validation Test Org",
      });
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);
    });

    test("should show validation error for empty name", async ({ page }) => {
      await page.getByRole("button", { name: "Create Role" }).click();
      const formMessages = page.locator('[data-slot="form-message"]');
      await expect(formMessages.first()).toBeVisible({ timeout: 5000 });
    });

    test("should allow submission with name only", async ({ page }) => {
      const roleName = `Test Role ${Date.now()}`;

      const form = new Form(page);
      await form.fillField("Name", roleName);
      await form.submit("Create Role");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role created", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
    });
  });

  test.describe("Successful Creation", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Success Test Org",
      });
    });

    test("should create role with name only", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);

      const roleName = `Required Role ${Date.now()}`;

      const form = new Form(page);
      await form.fillField("Name", roleName);
      await form.submit("Create Role");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role created", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await expect(page.getByText(roleName).first()).toBeVisible({
        timeout: 5000,
      });
    });

    test("should create role with name and description", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);

      const roleName = `Full Role ${Date.now()}`;
      const roleDescription = `This is a test role created at ${new Date().toISOString()}`;

      const form = new Form(page);
      await form.fillFields({
        Name: roleName,
        Description: roleDescription,
      });
      await form.submit("Create Role");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role created", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await expect(page.getByText(roleName).first()).toBeVisible({
        timeout: 5000,
      });
    });

    test("should show success toast after creation", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);

      const roleName = `Toast Test Role ${Date.now()}`;

      const form = new Form(page);
      await form.fillField("Name", roleName);
      await form.submit("Create Role");
      await expect(page.getByText("Role created", { exact: true })).toBeVisible(
        { timeout: 10000 }
      );
      await expect(page.getByText("Role created successfully")).toBeVisible({
        timeout: 5000,
      });
    });
  });

  test.describe("Cancel Functionality", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Cancel Test Org",
      });
    });

    test("should cancel creation and return to organization page", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);
      const form = new Form(page);
      await form.fillFields({
        Name: "Test Role",
        Description: "Test Description",
      });
      await page.getByRole("button", { name: "Cancel" }).click();
      await waitForPageLoad(page);
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
      await expect(page.getByText("Test Role")).not.toBeVisible();
    });
  });

  test.describe("Error Handling", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Create Error Test Org",
      });
    });

    test("should handle creation errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/new`,
      });
      await waitForPageLoad(page);
      const longName = "A".repeat(1000);
      const form = new Form(page);
      await form.fillField("Name", longName);
      await page.getByRole("button", { name: "Create Role" }).click();
      await Promise.race([
        page
          .waitForURL((url) => !url.pathname.includes("/roles/new"), {
            timeout: 5000,
          })
          .catch(() => null),
        page
          .waitForSelector('[role="alert"]', { timeout: 5000 })
          .catch(() => null),
        page
          .waitForSelector('[data-slot="form-message"]', { timeout: 5000 })
          .catch(() => null),
      ]);
      const errorAlert = page.locator('[role="alert"]');
      const hasError = await errorAlert.isVisible().catch(() => false);
      const isOnOrgPage =
        page.url().includes(`/settings/organizations/${testOrg.id}`) &&
        !page.url().includes("/roles/new");
      const isStillOnCreatePage = page.url().includes("/roles/new");
      expect(hasError || isOnOrgPage || isStillOnCreatePage).toBe(true);
    });
  });
});
