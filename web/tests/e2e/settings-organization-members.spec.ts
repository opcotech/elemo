import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  addMemberToRole,
  createDBOrganization,
  createDBRole,
} from "./utils/organization";
import { MembersPage } from "./pages/members-page";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";

test.describe("@settings.organization-members Organization Members Listing E2E Tests", () => {
  test.describe("Single Member", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Single Member Org",
      });
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should display single member correctly", async ({ page }) => {
      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();
      const membersTable = membersPage.getMembersTable();
      const tableHeaders = membersTable.locator("thead th");
      await expect(
        tableHeaders.filter({ hasText: "Name" }).first()
      ).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Roles" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Status" })).toBeVisible();
      await membersPage.waitForMember("Owner User");
      const ownerRow = membersPage.getMemberRow("Owner User");
      await expect(ownerRow.getByText(ownerUser.email)).toBeVisible();
      const ownerRolesCell = ownerRow.locator("td").nth(1);
      await expect(
        ownerRolesCell.getByText("Owner", { exact: false })
      ).toBeVisible();
      const statusCell = ownerRow.locator("td").nth(2);
      await expect(statusCell.getByText("Active")).toBeVisible();
      const nameCell = ownerRow.locator("td").first();
      await expect(nameCell.getByText("You", { exact: true })).toBeVisible();
    });
  });

  test.describe("Multiple Members", () => {
    let ownerUser: any;
    let adminUser: any;
    let memberUser: any;
    let readOnlyUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      adminUser = await createDBUser("active", {
        first_name: "Admin",
        last_name: "User",
      });

      memberUser = await createDBUser("active", {
        first_name: "Member",
        last_name: "User",
      });

      readOnlyUser = await createDBUser("active", {
        first_name: "ReadOnly",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Multiple Members Org",
      });
      await addMemberToOrganization(testOrg.id, adminUser.id, "write");
      await addMemberToOrganization(testOrg.id, memberUser.id, "read");
      await addMemberToOrganization(testOrg.id, readOnlyUser.id, "read");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should display all members correctly", async ({ page }) => {
      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();

      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();
      const membersTable = membersPage.getMembersTable();
      await expect(membersTable.getByText("Owner User")).toBeVisible();
      await expect(membersTable.getByText("Admin User")).toBeVisible();
      await expect(membersTable.getByText("Member User")).toBeVisible();
      await expect(membersTable.getByText("ReadOnly User")).toBeVisible();
      await expect(
        membersPage.getMemberRow("Owner User").getByText(ownerUser.email)
      ).toBeVisible();
      await expect(
        membersPage.getMemberRow("Admin User").getByText(adminUser.email)
      ).toBeVisible();
      await expect(
        membersPage.getMemberRow("Member User").getByText(memberUser.email)
      ).toBeVisible();
      await expect(
        membersPage.getMemberRow("ReadOnly User").getByText(readOnlyUser.email)
      ).toBeVisible();
    });

    test("should display correct virtual roles based on permissions", async ({
      page,
    }) => {
      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const ownerRow = membersPage.getMemberRow("Owner User");
      const ownerRolesCell = ownerRow.locator("td").nth(1);
      await expect(
        ownerRolesCell.getByText("Owner", { exact: false })
      ).toBeVisible();
      const adminRow = membersPage.getMemberRow("Admin User");
      const adminRolesCell = adminRow.locator("td").nth(1);
      await expect(
        adminRolesCell.getByText("Admin", { exact: false })
      ).toBeVisible();
      const memberRow = membersPage.getMemberRow("Member User");
      const memberRolesCell = memberRow.locator("td").nth(1);
      await expect(
        memberRolesCell.getByText("Member", { exact: false })
      ).toBeVisible();
    });

    test("should display correct status for all members", async ({ page }) => {
      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const ownerRow = membersPage.getMemberRow("Owner User");
      const ownerStatusCell = ownerRow.locator("td").nth(2);
      await expect(ownerStatusCell.getByText("Active")).toBeVisible();

      const adminRow = membersPage.getMemberRow("Admin User");
      const adminStatusCell = adminRow.locator("td").nth(2);
      await expect(adminStatusCell.getByText("Active")).toBeVisible();
    });
  });

  test.describe("Members with Roles", () => {
    let ownerUser: any;
    let memberUser: any;
    let testOrg: any;
    let testRole: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      memberUser = await createDBUser("active", {
        first_name: "Member",
        last_name: "WithRole",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Members With Roles Org",
      });

      await addMemberToOrganization(testOrg.id, memberUser.id, "read");
      testRole = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Developer",
        "Development team role"
      );
      await addMemberToRole(testRole, memberUser.id, testOrg.id);
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should display assigned roles for members", async ({ page }) => {
      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const memberRow = membersPage.getMemberRow("Member WithRole");
      const rolesCell = memberRow.locator("td").nth(1);
      await expect(rolesCell.getByText("Developer")).toBeVisible();
    });
  });
});
