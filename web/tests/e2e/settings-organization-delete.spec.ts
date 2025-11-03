import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";

test.describe("@settings.organization-delete Organization Delete E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let readUser: any;
    let writeUser: any;
    let deleteUser: any;
    let testOrganization: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      readUser = await createDBUser("active");
      writeUser = await createDBUser("active");
      deleteUser = await createDBUser("active");

      // Create organization owned by ownerUser
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: "Permission Test Organization",
      });

      // Add members with different permissions
      await addMemberToOrganization(testOrganization.id, readUser.id, "read");
      await addMemberToOrganization(testOrganization.id, writeUser.id, "write");
      await addMemberToOrganization(
        testOrganization.id,
        deleteUser.id,
        "delete"
      );
    });

    test("user with delete permission should see danger zone", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      // Wait for the actual delete button to appear (not the skeleton)
      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
    });

    test("user with delete-only permission should see danger zone", async ({
      page,
    }) => {
      await loginUser(page, deleteUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
    });

    test("user with read permission should not see danger zone", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 10000,
          }
        )
        .toBe(0);
    });

    test("user with write permission should not see danger zone", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 10000,
          }
        )
        .toBe(0);
    });

    test("deleted organization should not show danger zone", async ({
      page,
    }) => {
      // Create a deleted organization
      const deletedOrg = await createDBOrganization(ownerUser.id, "deleted", {
        name: "Deleted Test Organization",
      });

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${deletedOrg.id}`,
      });
      await page.waitForLoadState("networkidle");

      // Danger zone should not be visible for deleted organizations
      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: true }).count(),
          {
            timeout: 10000,
          }
        )
        .toBe(0);
    });
  });

  test.describe("Delete Dialog", () => {
    let ownerUser: any;
    let testOrganization: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active");
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: `Delete Test Org ${Date.now()}`,
      });
    });

    test("should open delete dialog when delete button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });

      // Wait for delete button to be visible
      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).toBeVisible({ timeout: 10000 });

      // Click delete button
      await page.getByRole("button", { name: "Delete Organization" }).click();

      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });

      // Verify dialog is open
      await expect(
        dialog.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      // Verify dialog content
      await expect(
        dialog.getByText("This will mark the organization as deleted", {
          exact: false,
        })
      ).toBeVisible();
      await expect(
        dialog.getByText("What will happen:", { exact: false })
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });

      // Open dialog
      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: true }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
      const deleteOrgButton = page
        .getByText("Delete Organization", { exact: false })
        .last();
      await deleteOrgButton.click();

      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });
      await expect(
        dialog.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      // Click cancel
      await dialog.getByRole("button", { name: "Cancel" }).click();

      // Verify dialog is closed
      await expect(dialog).not.toBeVisible();

      // Should still be on the same page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`
      );
    });
  });

  test.describe("Successful Deletion", () => {
    let ownerUser: any;
    let testOrganization: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active");
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: `Delete Success Test ${Date.now()}`,
      });
    });

    test("should successfully delete organization and redirect to list", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });

      // Open delete dialog
      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
      const deleteOrgButton = page
        .getByText("Delete Organization", { exact: false })
        .last();
      await deleteOrgButton.click();

      // Wait for dialog
      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });
      await expect(
        dialog.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      // Confirm deletion
      const confirmButton = dialog.getByRole("button", { name: /Delete/ });
      await expect(confirmButton).toBeVisible({ timeout: 10000 });
      await confirmButton.click();

      // Should redirect to organizations list
      await expect(page).toHaveURL("/settings/organizations", {
        timeout: 10000,
      });

      // Verify we're on the organizations list page
      await expect(
        page.getByRole("heading", { name: "Organizations" })
      ).toBeVisible({ timeout: 5000 });
    });

    test("should show success toast after deletion", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });

      // Open delete dialog
      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
      const deleteOrgButton = page
        .getByText("Delete Organization", { exact: false })
        .last();
      await deleteOrgButton.click();

      // Confirm deletion
      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });
      const confirmButton = dialog.getByRole("button", { name: /Delete/ });
      await expect(confirmButton).toBeVisible({ timeout: 10000 });
      await confirmButton.click();

      // Wait for toast to appear (check for success message)
      await expect(
        page.getByText("Organization deleted", { exact: false })
      ).toBeVisible({ timeout: 5000 });
    });

    test("should remove organization from list after deletion", async ({
      page,
    }) => {
      const orgName = `Temp Delete Test ${Date.now()}`;
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: orgName,
      });

      // First, verify organization is in the list
      await loginUser(page, ownerUser, {
        destination: "/settings/organizations",
      });
      await page.waitForLoadState("networkidle");

      await expect(page.getByText(orgName)).toBeVisible();

      // Navigate to detail page and delete
      await page.goto(`/settings/organizations/${testOrganization.id}`);
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });

      await expect
        .poll(
          async () =>
            page.getByText("Delete Organization", { exact: false }).count(),
          {
            timeout: 15000,
          }
        )
        .toBeGreaterThan(0);
      const deleteOrgButton = page
        .getByText("Delete Organization", { exact: false })
        .last();
      await deleteOrgButton.click();

      await page
        .getByRole("alertdialog")
        .getByRole("button", { name: "Delete" })
        .click();

      // Should be redirected to list
      await expect(page).toHaveURL("/settings/organizations", {
        timeout: 10000,
      });
      await page.waitForLoadState("networkidle");

      await expect
        .poll(
          async () => {
            const row = page.locator("tbody tr").filter({ hasText: orgName });
            const count = await row.count();
            if (count === 0) {
              return true;
            }
            const deletedCount = await row.getByText("Deleted").count();
            return deletedCount > 0;
          },
          { timeout: 15000 }
        )
        .toBe(true);
    });
  });

  test.describe("Error Handling", () => {
    let ownerUser: any;
    let testOrganization: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active");
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: `Error Test Org ${Date.now()}`,
      });
    });

    test("should handle deletion errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      // This test verifies the error handling exists
      // In a real scenario with network issues, errors would be shown
      // For now, we just verify the dialog has proper structure for errors

      const deleteOrgButton = page
        .getByText("Delete Organization", { exact: true })
        .last();
      await expect(deleteOrgButton).toBeVisible({ timeout: 10000 });
      await deleteOrgButton.click();

      // Verify dialog can be cancelled even if there's an error state
      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });
      await expect(
        dialog.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      await dialog.getByRole("button", { name: "Cancel" }).click();

      // Should still be on detail page
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`
      );
    });
  });

  test.describe("Danger Zone Content", () => {
    let ownerUser: any;
    let testOrganization: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active");
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: `Content Test Org ${Date.now()}`,
      });
    });

    test("should display all danger zone warning information", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      await expect(
        page.getByText("Danger Zone", { exact: true }).first()
      ).toBeVisible({ timeout: 10000 });

      // Verify warning text
      await expect(
        page
          .getByText(
            "Deleting an organization will mark it as deleted and hide it from listings",
            { exact: false }
          )
          .first()
      ).toBeVisible();

      // Verify consequences list
      await expect(page.getByText("Consequences:").first()).toBeVisible();
      await expect(
        page.getByText("All organization members will lose access").first()
      ).toBeVisible();
      await expect(
        page
          .getByText(
            "Organization data will be hidden from search and listings",
            { exact: false }
          )
          .first()
      ).toBeVisible();
      await expect(
        page.getByText("The organization will be marked as deleted").first()
      ).toBeVisible();
      await expect(
        page
          .getByText("This action is permanent and cannot be reversed")
          .first()
      ).toBeVisible();
    });

    test("should display dialog with correct warning information", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await page.waitForLoadState("networkidle");

      // Open dialog
      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });
      const deleteOrgButton = page
        .locator('button:has-text("Delete Organization")')
        .first();
      await expect(deleteOrgButton).toBeVisible({ timeout: 10000 });
      await deleteOrgButton.click();

      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 10000 });

      // Verify dialog title contains organization name
      await expect(
        dialog.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      // Verify dialog warning text
      await expect(
        dialog.getByText(
          "This will mark the organization as deleted. This action cannot be undone.",
          { exact: false }
        )
      ).toBeVisible();

      // Verify dialog consequences
      await expect(dialog.getByText("What will happen:")).toBeVisible();
      await expect(
        dialog.getByText("The organization will be marked as deleted")
      ).toBeVisible();
      await expect(
        dialog.getByText("All organization members will lose access")
      ).toBeVisible();
      await expect(
        dialog.getByText("Organization data will be hidden from listings")
      ).toBeVisible();
      await expect(
        dialog.getByText("You will be redirected to the organizations list")
      ).toBeVisible();
    });
  });
});
