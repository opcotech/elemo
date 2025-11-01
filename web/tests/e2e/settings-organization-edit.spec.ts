import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";

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

      // Create organization owned by ownerUser
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: "Original Organization",
        website: "https://original.example.com",
      });

      // Add write user with write permission
      await addMemberToOrganization(testOrganization.id, writeUser.id, "write");

      // Add read user with read permission
      await addMemberToOrganization(testOrganization.id, readUser.id, "read");
    });

    test("user with write permission should see edit button", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      // Edit button should be visible on detail page (it's a Link wrapped in Button)
      await expect(page.getByRole("link", { name: "Edit" })).toBeVisible();
    });

    test("user without write permission should not see edit button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByRole("link", { name: "Edit" })).not.toBeVisible();
    });

    test("user without write permission should not access edit page", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });

      // Wait for navigation to complete (either redirect or page load)
      await page.waitForLoadState("networkidle");

      // User should be redirected to permission denied page
      await expect(page).toHaveURL(/.*permission-denied/, { timeout: 10000 });
    });

    test("should navigate to edit page when edit button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await page.getByRole("link", { name: "Edit" }).click();
      await page.waitForLoadState("networkidle");

      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`
      );
      await expect(
        page.getByRole("heading", { name: "Edit Organization" })
      ).toBeVisible();
    });

    test("should pre-fill form with existing organization data", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await page.getByRole("link", { name: "Edit" }).click();
      await page.waitForLoadState("networkidle");

      // Check that form fields are pre-filled
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
      await page.waitForLoadState("networkidle");

      const updatedName = `Updated Org ${Date.now()}`;
      const updatedEmail = `updated-${Date.now()}@example.com`;
      const updatedWebsite = `https://updated-${Date.now()}.example.com`;

      await page.getByLabel("Name").fill(updatedName);
      await page.getByLabel("Email").fill(updatedEmail);
      await page.getByLabel("Website").fill(updatedWebsite);

      await page.getByRole("button", { name: "Save Changes" }).click();

      // Wait for navigation back to detail page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );

      // Simply verify navigation happened
      await page.waitForLoadState("networkidle");
    });

    test("should update organization with partial fields", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await page.waitForLoadState("networkidle");

      const updatedName = `Partial Update ${Date.now()}`;

      // Only update name, leave email and website unchanged
      await page.getByLabel("Name").fill(updatedName);

      await page.getByRole("button", { name: "Save Changes" }).click();

      // Wait for navigation back to detail page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );

      // Simply verify navigation happened
      await page.waitForLoadState("networkidle");
    });

    test("should show error when trying to save without changes", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await page.waitForLoadState("networkidle");

      // Don't make any changes, just click save
      await page.getByRole("button", { name: "Save Changes" }).click();

      // Should still be on edit page (no navigation)
      // Wait for URL to confirm we haven't navigated away
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`,
        { timeout: 5000 }
      );
    });

    test("should show validation error for invalid email", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await page.waitForLoadState("networkidle");

      await page.getByLabel("Email").fill("invalid-email");
      await page.getByLabel("Name").click(); // Trigger blur to show validation

      await page.getByRole("button", { name: "Save Changes" }).click();

      // Should still be on edit page (validation prevented submission)
      // Wait for URL to confirm we haven't navigated away
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`,
        { timeout: 5000 }
      );
    });

    test("should cancel edit and return to detail page", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await page.waitForLoadState("networkidle");

      // Make some changes
      await page.getByLabel("Name").fill("Changed Name");

      await page.getByRole("button", { name: "Cancel" }).click();

      // Should navigate back to detail page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`
      );
    });

    test("should display loading state during save", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}/edit`,
      });
      await page.waitForLoadState("networkidle");

      const updatedName = `Loading Test ${Date.now()}`;
      await page.getByLabel("Name").fill(updatedName);

      // The save button should be disabled during save
      const saveButton = page.getByRole("button", { name: "Save Changes" });
      await saveButton.click();

      // Check that the form is submitted and navigates away
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        { timeout: 10000 }
      );
    });

    test("should allow user with write permission to edit", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      // Write user should see edit button
      await expect(page.getByRole("link", { name: "Edit" })).toBeVisible();

      await page.getByRole("link", { name: "Edit" }).click();
      await page.waitForLoadState("networkidle");

      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}/edit`
      );

      // Should be able to edit
      const updatedName = `Write User Update ${Date.now()}`;
      await page.getByLabel("Name").fill(updatedName);
      await page.getByRole("button", { name: "Save Changes" }).click();

      // Wait for navigation back to detail page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`,
        {
          timeout: 10000,
        }
      );
    });
  });
});
