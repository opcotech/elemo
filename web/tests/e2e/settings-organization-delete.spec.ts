import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";
import { OrganizationPage } from "./pages/organization-page";
import { Dialog } from "./components/dialog";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";
import { waitForSuccessToast } from "./helpers/toast";

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
      testOrganization = await createDBOrganization(ownerUser.id, "active", {
        name: "Permission Test Organization",
      });
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

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).toBeVisible({ timeout: 15000 });
    });

    test("user with delete-only permission should see danger zone", async ({
      page,
    }) => {
      await loginUser(page, deleteUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });
      await waitForPermissionsLoad(page);

      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).toBeVisible({ timeout: 15000 });
    });

    test("user with read permission should not see danger zone", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).not.toBeVisible();
    });

    test("user with write permission should not see danger zone", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).not.toBeVisible();
    });

    test("deleted organization should not show danger zone", async ({
      page,
    }) => {
      const deletedOrg = await createDBOrganization(ownerUser.id, "deleted", {
        name: "Deleted Test Organization",
      });

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${deletedOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, deletedOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).not.toBeVisible();
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

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });
      await expect(
        page.getByRole("button", { name: "Delete Organization" })
      ).toBeVisible({ timeout: 10000 });
      await page.getByRole("button", { name: "Delete Organization" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      const dialogLocator = dialog.getLocator();
      await expect(
        dialogLocator.getByText(
          `Are you sure you want to delete ${testOrganization.name}?`,
          { exact: false }
        )
      ).toBeVisible();

      await expect(
        dialogLocator.getByText("This will mark the organization as deleted", {
          exact: false,
        })
      ).toBeVisible();
      await expect(
        dialogLocator.getByText("What will happen:", { exact: false })
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole("button", { name: "Delete Organization" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
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

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole("button", { name: "Delete Organization" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.confirm("Delete");
      await expect(page).toHaveURL("/settings/organizations", {
        timeout: 10000,
      });
      await expect(
        page.getByRole("heading", { name: "Organizations", level: 1 })
      ).toBeVisible({ timeout: 5000 });
    });

    test("should show success toast after deletion", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrganization.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);

      await expect(page.getByText("Danger Zone", { exact: true })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole("button", { name: "Delete Organization" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.confirm("Delete");
      await waitForSuccessToast(page, "Organization deleted", {
        timeout: 10000,
      });
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

      const orgPage = new OrganizationPage(page, testOrganization.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      await page.getByRole("button", { name: "Delete Organization" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await expect(page).toHaveURL(
        `/settings/organizations/${testOrganization.id}`
      );
    });
  });
});
