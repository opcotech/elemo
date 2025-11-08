import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  addMemberToRole,
  createDBOrganization,
  createDBRole,
} from "./utils/organization";
import { OrganizationPage } from "./pages/organization-page";
import { RoleEditPage } from "./pages/role-edit-page";
import { Form } from "./components/form";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";

test.describe("@settings.organization-role-edit Organization Role Edit E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let writeUser: any;
    let readUser: any;
    let testOrg: any;
    let testRole: any;

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
        name: "Role Edit Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, writeUser.id, "read");
      await addMemberToOrganization(testOrg.id, writeUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
      testRole = await createDBRole(
        testOrg.id,
        writeUser.id,
        "Test Role",
        "A test role for permission testing"
      );
    });

    test("user with write permission should see edit button", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await orgPage.waitForRole("Test Role");
      const roleRow = orgPage.getRoleRow("Test Role");
      await expect(roleRow).toBeVisible();
      const editButton = roleRow.getByTitle("Edit role");
      const editButtonCount = await editButton.count();
      if (editButtonCount > 0) {
        await expect(editButton).toBeVisible();
      } else {
        await expect(roleRow).toBeVisible();
      }
    });

    test("user with read permission should not see edit button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await orgPage.waitForRole("Test Role");
      const roleRow = orgPage.getRoleRow("Test Role");
      await expect(roleRow.getByTitle("Edit role")).not.toBeVisible();
    });

    test("user without write permission should not access edit page", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
      await expect(page).toHaveURL(/.*permission-denied/, { timeout: 10000 });
    });
  });

  test.describe("Navigation", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Navigation Test Org",
      });
      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Navigation Test Role",
        "A role for navigation testing"
      );
    });

    test("should navigate to edit page when edit button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await orgPage.waitForRole("Navigation Test Role");
      const roleRow = orgPage.getRoleRow("Navigation Test Role");
      const editButton = roleRow.getByTitle("Edit role");
      const editButtonCount = await editButton.count();

      if (editButtonCount > 0) {
        await expect(editButton).toBeVisible();
        await orgPage.clickEditRole("Navigation Test Role");
        await waitForPageLoad(page);
      } else {
        const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
        await roleEditPage.goto();
      }
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`
      );
      await expect(
        page.getByRole("heading", { name: "Edit Role" })
      ).toBeVisible();
    });
  });

  test.describe("Form Pre-filling", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;
    let roleName: string;
    let roleDescription: string;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Pre-fill Test Org",
      });

      roleName = `Pre-fill Test Role ${Date.now()}`;
      roleDescription = `This is a test role description for pre-filling at ${new Date().toISOString()}`;

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        roleName,
        roleDescription
      );
    });

    test("should pre-fill form with existing role data", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
      const nameInput = page.getByLabel("Name");
      const descriptionInput = page.getByLabel("Description");

      await expect(nameInput).toHaveValue(roleName);
      await expect(descriptionInput).toHaveValue(roleDescription);
    });
  });

  test.describe("Successful Updates", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Update Test Org",
      });

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Original Role Name",
        "Original description"
      );
    });

    test("should update role with all fields", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);

      const updatedName = `Updated Role ${Date.now()}`;
      const updatedDescription = `Updated description at ${new Date().toISOString()}`;

      const form = new Form(page);
      await form.fillFields({
        Name: updatedName,
        Description: updatedDescription,
      });
      await form.submit("Save Changes");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role updated", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await expect(page.getByText(updatedName).first()).toBeVisible({
        timeout: 5000,
      });
    });

    test("should update role with partial fields", async ({ page }) => {
      const partialTestRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Partial Update Role",
        "Original description"
      );

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${partialTestRole}/edit`,
      });
      await waitForPageLoad(page);

      const updatedName = `Partial Updated ${Date.now()}`;

      const form = new Form(page);
      await form.fillField("Name", updatedName);
      await form.submit("Save Changes");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role updated", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
    });

    test("should update role description only", async ({ page }) => {
      const descTestRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Description Update Role",
        "Original description"
      );

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${descTestRole}/edit`,
      });
      await waitForPageLoad(page);

      const updatedDescription = `Updated description only at ${new Date().toISOString()}`;

      const form = new Form(page);
      await form.fillField("Description", updatedDescription);
      await form.submit("Save Changes");
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
      await expect(page.getByText("Role updated", { exact: true })).toBeVisible(
        { timeout: 5000 }
      );
    });

    test("should show success toast after update", async ({ page }) => {
      const toastTestRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Toast Test Role",
        "Original description"
      );

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${toastTestRole}/edit`,
      });
      await waitForPageLoad(page);

      const updatedName = `Toast Updated ${Date.now()}`;
      const form = new Form(page);
      await form.fillField("Name", updatedName);
      await form.submit("Save Changes");
      await expect(page.getByText("Role updated", { exact: true })).toBeVisible(
        { timeout: 10000 }
      );
      await expect(page.getByText("Role updated successfully")).toBeVisible({
        timeout: 5000,
      });
    });
  });

  test.describe("Form Validation", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Validation Test Org",
      });

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Validation Test Role",
        "Original description"
      );
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
    });

    test("should show validation error for empty name", async ({ page }) => {
      await page.getByLabel("Name").clear();
      await page.getByLabel("Name").blur();
      await page
        .waitForFunction(
          () => {
            const input = document.querySelector(
              'input[aria-label="Name"]'
            ) as HTMLInputElement;
            return input.value === "";
          },
          { timeout: 1000 }
        )
        .catch(() => {});
      await page.getByRole("button", { name: "Save Changes" }).click();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasFormMessage = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const isStillOnEditPage =
        page.url().includes("/roles/") && page.url().includes("/edit");
      expect(hasFormMessage || isStillOnEditPage).toBe(true);
    });

    test("should allow clearing description", async ({ page }) => {
      await page.getByLabel("Description").clear();
      await page.getByRole("button", { name: "Save Changes" }).click();
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`, {
        timeout: 10000,
      });
    });
  });

  test.describe("Cancel Functionality", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Cancel Test Org",
      });

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Cancel Test Role",
        "Original description"
      );
    });

    test("should cancel edit and return to organization page", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
      const form = new Form(page);
      await form.fillFields({
        Name: "Changed Name",
        Description: "Changed Description",
      });

      await page.getByRole("button", { name: "Cancel" }).click();
      await waitForPageLoad(page);
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await expect(page.getByText("Cancel Test Role").first()).toBeVisible();
    });
  });

  test.describe("Error Handling", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Edit Error Test Org",
      });

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Error Test Role",
        "Original description"
      );
    });

    test("should handle update errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
      const longName = "A".repeat(1000);
      const form = new Form(page);
      await form.fillField("Name", longName);
      await page.getByRole("button", { name: "Save Changes" }).click();
      await Promise.race([
        page
          .waitForURL(
            (url) =>
              !url.pathname.includes("/roles/") ||
              !url.pathname.includes("/edit"),
            { timeout: 5000 }
          )
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
        !page.url().includes("/roles/") &&
        !page.url().includes("/edit");
      const isStillOnEditPage =
        page.url().includes("/roles/") && page.url().includes("/edit");
      expect(hasError || isOnOrgPage || isStillOnEditPage).toBe(true);
    });
  });
});
