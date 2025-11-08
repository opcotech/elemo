import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
  createDBRole,
} from "./utils/organization";
import { OrganizationPage } from "./pages/organization-page";
import { Dialog } from "./components/dialog";
import { waitForPermissionsLoad } from "./helpers/navigation";
import { waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.organization-role-delete Organization Role Delete E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let deleteUser: any;
    let readUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });
      deleteUser = await createDBUser("active", {
        first_name: "Delete",
        last_name: "User",
      });
      readUser = await createDBUser("active", {
        first_name: "Read",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Delete Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
      await addMemberToOrganization(testOrg.id, deleteUser.id, "read");
      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Delete Test Role",
        "A test role for delete permission testing"
      );
    });

    test("user with delete permission should see delete button", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await orgPage.waitForRole("Delete Test Role");
      await waitForPermissionsLoad(page);
      const hasDeleteButton = await orgPage.hasDeleteButton("Delete Test Role");
      expect(hasDeleteButton).toBe(true);
    });

    test("user with read permission should not see delete button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await orgPage.waitForRole("Delete Test Role");
      await waitForPermissionsLoad(page);
      const hasDeleteButton = await orgPage.hasDeleteButton("Delete Test Role");
      expect(hasDeleteButton).toBe(false);
    });
  });

  test.describe("Delete Dialog", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Delete Dialog Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Dialog Test Role",
        "A role for dialog testing"
      );
    });

    test("should open delete dialog when delete button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await orgPage.waitForRole("Dialog Test Role");
      await waitForPermissionsLoad(page);
      await orgPage.clickDeleteRole("Dialog Test Role");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      const dialogLocator = dialog.getLocator();
      await expect(
        dialogLocator.getByText(
          "Are you sure you want to delete Dialog Test Role?"
        )
      ).toBeVisible();
      await expect(
        dialogLocator.getByText(
          "This will permanently delete the role. This action cannot be undone."
        )
      ).toBeVisible();
      await expect(
        dialogLocator.getByText("The role will be permanently deleted")
      ).toBeVisible();
      await expect(
        dialogLocator.getByText(
          "All members assigned to this role will lose their role assignment"
        )
      ).toBeVisible();
      await expect(
        dialogLocator.getByText("Role permissions will be removed")
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await orgPage.waitForRole("Dialog Test Role");
      await waitForPermissionsLoad(page);
      await orgPage.clickDeleteRole("Dialog Test Role");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
    });
  });

  test.describe("Successful Deletion", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: `Role Delete Success Test Org ${Date.now()}`,
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");
    });

    test("should successfully delete role and update list", async ({
      page,
    }) => {
      const roleToDelete = await createDBRole(
        testOrg.id,
        ownerUser.id,
        `Delete Success Role ${Date.now()}`,
        "A role to be deleted"
      );

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await waitForPermissionsLoad(page);
      const rolesTable = orgPage.getRolesTable();
      const roleRows = rolesTable.locator("tbody tr");
      const rowCount = await roleRows.count();

      let roleRow: any;
      let roleNameToDelete = "";

      for (let i = 0; i < rowCount; i++) {
        const row = roleRows.nth(i);
        const deleteButton = row.getByRole("button", { name: "Delete role" });
        const hasDeleteButton = await deleteButton.count();
        if (hasDeleteButton > 0) {
          roleRow = row;
          const nameCell = row.locator("td").first();
          roleNameToDelete = (await nameCell.textContent()) || "";
          break;
        }
      }

      expect(roleRow).toBeDefined();
      expect(roleNameToDelete).toBeTruthy();
      await roleRow.getByRole("button", { name: "Delete role" }).click();
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.confirm("Delete");
      await waitForSuccessToast(page, "Role deleted", { timeout: 10000 });
      await expect(rolesTable.getByText(roleNameToDelete)).not.toBeVisible({
        timeout: 5000,
      });
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
        name: "Role Delete Error Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Error Test Role",
        "A role for error testing"
      );
    });

    test("should handle deletion errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const orgPage = new OrganizationPage(page, testOrg.id);
      await orgPage.waitForRolesLoad();
      await orgPage.waitForRole("Error Test Role");
      await waitForPermissionsLoad(page);
      await orgPage.clickDeleteRole("Error Test Role");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
    });
  });
});
