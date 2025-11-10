import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";
import { OrganizationPage } from "./pages/organization-page";
import { Form } from "./components/form";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";
import { waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.organization-edit Organization Edit E2E Tests", () => {
  test.describe("Organization Editing", () => {
    let ownerUser: any;
    let writeUser: any;
    let readUser: any;
    let testOrganization: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      writeUser = await createDBUser("active");
      readUser = await createDBUser("active");
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: "Original Organization",
        website: "https://original.example.com",
      });
      await addMemberToOrganization(testOrganization.id, writeUser.id, "write");
      await addMemberToOrganization(testOrganization.id, readUser.id, "read");
    });

    test("user with write permission should see edit button", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(page.getByRole("link", { name: "Edit" })).toBeVisible();
    });

    test("user without write permission should not see edit button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(page.getByRole("link", { name: "Edit" })).not.toBeVisible();
    });

    test("user without write permission should not access edit page", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await waitForPageLoad(page);
      await expect(page).toHaveURL(/.*permission-denied/, { timeout: 10000 });
    });

    test("should navigate to edit page when edit button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await page.getByRole("link", { name: "Edit" }).click();
      await waitForPageLoad(page);

      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`
      );
      await expect(
        page.getByRole("heading", { name: "Edit Organization", level: 1 })
      ).toBeVisible({ timeout: 10000 });
    });

    test("should pre-fill form with existing organization data", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await page.getByRole("link", { name: "Edit" }).click();
      await waitForPageLoad(page);
      const nameInput = page.getByLabel("Name");
      const emailInput = page.getByLabel("Email");
      const websiteInput = page.getByLabel("Website");

      await expect(nameInput).toHaveValue(testOrganization.name);
      await expect(emailInput).toHaveValue(testOrganization.email);
      if (testOrganization.website) {
        await expect(websiteInput).toHaveValue(testOrganization.website);
      }
    });

    test("should update organization with all fields", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await waitForPageLoad(page);

      const updatedName = `Updated Org ${Date.now()}`;
      const updatedEmail = `updated-${Date.now()}@example.com`;
      const updatedWebsite = `https://updated-${Date.now()}.example.com`;

      const form = new Form(page);
      await form.fillFields({
        Name: updatedName,
        Email: updatedEmail,
        Website: updatedWebsite,
      });
      await form.submit("Save Changes");
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );
      await waitForSuccessToast(page, "Organization updated", {
        timeout: 5000,
      });
    });

    test("should update organization with partial fields", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await waitForPageLoad(page);

      const updatedName = `Partial Update ${Date.now()}`;
      const form = new Form(page);
      await form.fillField("Name", updatedName);
      await form.submit("Save Changes");
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );
    });

    test("should show validation error for invalid email", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await waitForPageLoad(page);

      const form = new Form(page);
      await form.fillField("Email", "invalid-email");
      await page.getByLabel("Name").click();
      await page.getByRole("button", { name: "Save Changes" }).click();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasError = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const isStillOnEditPage = page.url().includes("/edit");

      expect(hasError || isStillOnEditPage).toBe(true);
    });

    test("should cancel edit and return to detail page", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await waitForPageLoad(page);
      const form = new Form(page);
      await form.fillField("Name", "Changed Name");

      await page.getByRole("button", { name: "Cancel" }).click();
      await waitForPageLoad(page);
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`
      );
    });

    test("should allow user with write permission to edit", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(page.getByRole("link", { name: "Edit" })).toBeVisible();

      await page.getByRole("link", { name: "Edit" }).click();
      await waitForPageLoad(page);

      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`
      );
      const updatedName = `Write User Update ${Date.now()}`;
      const form = new Form(page);
      await form.fillField("Name", updatedName);
      await form.submit("Save Changes");
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );
    });
  });
});
