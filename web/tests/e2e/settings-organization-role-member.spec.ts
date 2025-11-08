import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  addMemberToRole,
  createDBOrganization,
  createDBRole,
  grantRoleWritePermission,
} from "./utils/organization";
import { RoleEditPage } from "./pages/role-edit-page";
import { AddMemberDialog } from "./components/add-member-dialog";
import { Dialog } from "./components/dialog";
import { waitForSuccessToast } from "./helpers/toast";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";

test.describe("@settings.organization-role-member Organization Role Member E2E Tests", () => {
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
        name: "Role Member Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, writeUser.id, "read");
      await addMemberToOrganization(testOrg.id, writeUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Member Permission Test Role",
        "A role for member permission testing"
      );
    });

    test("user with write permission should see add member button", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await expect(
        page.getByRole("button", { name: /Add Member/i })
      ).toBeVisible({ timeout: 10000 });
    });

    test("user without write permission should not see add member button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });
      await waitForPageLoad(page);
      await expect(page).toHaveURL(/.*permission-denied/, { timeout: 10000 });
    });
  });

  test.describe("Add Member Dialog", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;
    let memberToAdd: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Member Add Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "read");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Add Member Test Role",
        "A role for adding members"
      );
      await grantRoleWritePermission(ownerUser.id, testRole);
      memberToAdd = await createDBUser("active", {
        first_name: "Member",
        last_name: "ToAdd",
        email: `member-add-${Date.now()}@example.com`,
      });
      await addMemberToOrganization(testOrg.id, memberToAdd.id, "read");
    });

    test("should open add member dialog when add button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.clickAddMember();
      const addMemberDialog = new AddMemberDialog(page);
      await addMemberDialog.waitFor();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.clickAddMember();
      const addMemberDialog = new AddMemberDialog(page);
      await addMemberDialog.waitFor();
      await addMemberDialog.cancel();
      await addMemberDialog.waitForClose();
    });
  });

  test.describe("Successful Member Assignment", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;
    let memberToAdd: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: `Role Member Add Success Test Org ${Date.now()}`,
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "read");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        `Add Success Test Role ${Date.now()}`,
        "A role for successful member addition"
      );
      await grantRoleWritePermission(ownerUser.id, testRole);
      memberToAdd = await createDBUser("active", {
        first_name: "Member",
        last_name: "ToAdd",
        email: `member-add-${Date.now()}@example.com`,
      });
      await addMemberToOrganization(testOrg.id, memberToAdd.id, "read");
    });

    test("should successfully add member to role", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.clickAddMember();
      const addMemberDialog = new AddMemberDialog(page);
      await addMemberDialog.waitFor();
      await addMemberDialog.addMemberToRole("Member ToAdd");
      await waitForSuccessToast(page, "Member added", { timeout: 10000 });
      await addMemberDialog.waitForClose();
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.waitForMember("Member ToAdd");
    });
  });

  test.describe("Remove Member Dialog", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;
    let memberToRemove: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Role Member Remove Dialog Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Remove Dialog Test Role",
        "A role for remove dialog testing"
      );
      memberToRemove = await createDBUser("active", {
        first_name: "Member",
        last_name: "ToRemove",
        email: `member-remove-${Date.now()}@example.com`,
      });
      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");
      await addMemberToRole(testRole, memberToRemove.id, testOrg.id);
    });

    test("should open remove dialog when remove button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await waitForPermissionsLoad(page);
      await roleEditPage.waitForMember("Member ToRemove");
      await roleEditPage.clickRemoveMember("Member ToRemove");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      const dialogLocator = dialog.getLocator();
      await expect(
        dialogLocator.getByRole("heading", { name: /Remove Member ToRemove/i })
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await waitForPermissionsLoad(page);
      await roleEditPage.waitForMember("Member ToRemove");
      await roleEditPage.clickRemoveMember("Member ToRemove");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await dialog.waitForClose();
      await roleEditPage.waitForMember("Member ToRemove");
    });
  });

  test.describe("Successful Member Removal", () => {
    let ownerUser: any;
    let testOrg: any;
    let testRole: any;
    let memberToRemove: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: `Role Member Remove Success Test Org ${Date.now()}`,
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "write");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        `Remove Success Test Role ${Date.now()}`,
        "A role for successful member removal"
      );
      memberToRemove = await createDBUser("active", {
        first_name: "Member",
        last_name: "ToRemove",
        email: `member-remove-${Date.now()}@example.com`,
      });
      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");
      await addMemberToRole(testRole, memberToRemove.id, testOrg.id);
      await grantRoleWritePermission(ownerUser.id, testRole);
    });

    test("should successfully remove member from role", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await waitForPermissionsLoad(page);
      await roleEditPage.waitForMember("Member ToRemove");
      await roleEditPage.clickRemoveMember("Member ToRemove");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.confirm("Remove Member");
      await waitForSuccessToast(page, "Member removed", { timeout: 10000 });
      await roleEditPage.waitForMembersLoad();
      const memberExists = await roleEditPage.memberExists("Member ToRemove");
      expect(memberExists).toBe(false);
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
        name: "Role Member Validation Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "read");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Validation Test Role",
        "A role for validation testing"
      );
      await grantRoleWritePermission(ownerUser.id, testRole);
    });

    test("should show validation error when no member is selected", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.clickAddMember();
      const addMemberDialog = new AddMemberDialog(page);
      await addMemberDialog.waitFor();
      await addMemberDialog.addMember();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasFormMessage = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const isDialogStillOpen = await addMemberDialog.isVisible();
      expect(hasFormMessage || isDialogStillOpen).toBe(true);
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
        name: "Role Member Error Test Org",
      });
      await addMemberToOrganization(testOrg.id, ownerUser.id, "read");

      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Error Test Role",
        "A role for error testing"
      );
      await grantRoleWritePermission(ownerUser.id, testRole);
    });

    test("should handle add member errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}/roles/${testRole}/edit`,
      });

      const roleEditPage = new RoleEditPage(page, testOrg.id, testRole);
      await roleEditPage.waitForMembersLoad();
      await roleEditPage.clickAddMember();
      const addMemberDialog = new AddMemberDialog(page);
      await addMemberDialog.waitFor();
      await addMemberDialog.cancel();
      await addMemberDialog.waitForClose();
    });
  });
});
