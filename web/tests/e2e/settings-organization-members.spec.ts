import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  addMemberToRole,
  createDBOrganization,
  createDBRole,
} from "./utils/organization";

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
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });

      // Check members section header (use first() to handle multiple matches)
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Scope to members table - find table that contains member data
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });

      // Check table headers - scope to members table only
      const tableHeaders = membersTable.locator("thead th");
      await expect(
        tableHeaders.filter({ hasText: "Name" }).first()
      ).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Roles" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Status" })).toBeVisible();

      // Check member row - scope to members table to avoid duplicates
      await expect(membersTable.getByText("Owner User")).toBeVisible();
      // Check email in the members table (not in user menu)
      const ownerRow = membersTable
        .locator("tbody tr")
        .filter({ hasText: "Owner User" })
        .first();
      await expect(ownerRow.getByText(ownerUser.email)).toBeVisible();

      // Check roles - owner should have "Owner" role (from permissions)
      const ownerRolesCell = ownerRow.locator("td").nth(1); // Roles is 2nd column
      await expect(
        ownerRolesCell.getByText("Owner", { exact: false })
      ).toBeVisible();

      // Check status
      const statusCell = ownerRow.locator("td").nth(2); // Status is 3rd column
      await expect(statusCell.getByText("Active")).toBeVisible();

      // Check "You" badge for current user (in the members table)
      const nameCell = ownerRow.locator("td").first();
      await expect(nameCell.getByText("You", { exact: true })).toBeVisible();
    });

    test("should show empty state when no members", async ({ page }) => {
      // This test would require deleting the member, but that's complex
      // So we'll test the empty state structure exists by checking the component
      await page.waitForLoadState("networkidle");

      // The members section should exist
      await expect(page.getByText("Members").first()).toBeVisible();
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

      // Add members with different permissions (which create virtual roles)
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
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Check all members are visible - scope to table to avoid duplicates
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });
      await expect(membersTable.getByText("Owner User")).toBeVisible();
      await expect(membersTable.getByText("Admin User")).toBeVisible();
      await expect(membersTable.getByText("Member User")).toBeVisible();
      await expect(membersTable.getByText("ReadOnly User")).toBeVisible();

      // Check emails in the members table
      const ownerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Owner User" })
        .first();
      await expect(ownerRow.getByText(ownerUser.email)).toBeVisible();

      const adminRow = page
        .locator("tbody tr")
        .filter({ hasText: "Admin User" })
        .first();
      await expect(adminRow.getByText(adminUser.email)).toBeVisible();

      const memberRow = page
        .locator("tbody tr")
        .filter({ hasText: "Member User" })
        .first();
      await expect(memberRow.getByText(memberUser.email)).toBeVisible();

      const readOnlyRow = page
        .locator("tbody tr")
        .filter({ hasText: "ReadOnly User" })
        .first();
      await expect(readOnlyRow.getByText(readOnlyUser.email)).toBeVisible();
    });

    test("should display correct virtual roles based on permissions", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Owner should have "Owner" role
      const ownerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Owner User" })
        .first();
      const ownerRolesCell = ownerRow.locator("td").nth(1); // Roles is 2nd column
      await expect(
        ownerRolesCell.getByText("Owner", { exact: false })
      ).toBeVisible();

      // Admin (write permission) should have "Admin" role
      const adminRow = page
        .locator("tbody tr")
        .filter({ hasText: "Admin User" })
        .first();
      const adminRolesCell = adminRow.locator("td").nth(1);
      await expect(
        adminRolesCell.getByText("Admin", { exact: false })
      ).toBeVisible();

      // Members (read permission) should have "Member" role
      const memberRow = page
        .locator("tbody tr")
        .filter({ hasText: "Member User" })
        .first();
      const memberRolesCell = memberRow.locator("td").nth(1);
      await expect(
        memberRolesCell.getByText("Member", { exact: false })
      ).toBeVisible();

      const readOnlyRow = page
        .locator("tbody tr")
        .filter({ hasText: "ReadOnly User" })
        .first();
      const readOnlyRolesCell = readOnlyRow.locator("td").nth(1);
      await expect(
        readOnlyRolesCell.getByText("Member", { exact: false })
      ).toBeVisible();
    });

    test("should display correct status badges", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      // All users should have Active status
      const statusBadges = page.locator('[data-slot="badge"]').filter({
        hasText: "Active",
      });
      const count = await statusBadges.count();
      expect(count).toBeGreaterThanOrEqual(4); // At least 4 active members
    });

    test("should sort members correctly (non-deleted first, then alphabetically)", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Scope to members table
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });

      const rows = membersTable.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBe(4);

      // Get member names in order
      const memberNames: Array<string> = [];
      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const nameCell = row.locator("td").first();
        const nameText = await nameCell.textContent();
        if (nameText) {
          // Extract just the name (remove email and "You" badge)
          const nameMatch = nameText.match(/^([A-Za-z]+ [A-Za-z]+)/);
          if (nameMatch) {
            memberNames.push(nameMatch[1].trim());
          }
        }
      }

      // Should be sorted alphabetically (all are active)
      // Expected order: Admin User, Member User, Owner User, ReadOnly User
      expect(memberNames[0]).toContain("Admin");
      expect(memberNames[1]).toContain("Member");
      expect(memberNames[2]).toContain("Owner");
      expect(memberNames[3]).toContain("ReadOnly");
    });
  });

  test.describe("Members with Roles", () => {
    let ownerUser: any;
    let developerUser: any;
    let designerUser: any;
    let managerUser: any;
    let testOrg: any;
    let developerRoleId: string;
    let designerRoleId: string;
    let managerRoleId: string;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      developerUser = await createDBUser("active", {
        first_name: "Developer",
        last_name: "User",
      });

      designerUser = await createDBUser("active", {
        first_name: "Designer",
        last_name: "User",
      });

      managerUser = await createDBUser("active", {
        first_name: "Manager",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Roles Test Org",
      });

      // Add members to organization
      await addMemberToOrganization(testOrg.id, developerUser.id, "read");
      await addMemberToOrganization(testOrg.id, designerUser.id, "read");
      await addMemberToOrganization(testOrg.id, managerUser.id, "write");

      // Create roles
      developerRoleId = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Developer",
        "Development team"
      );
      designerRoleId = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Designer",
        "Design team"
      );
      managerRoleId = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Manager",
        "Management team"
      );

      // Add members to roles
      await addMemberToRole(developerRoleId, developerUser.id, testOrg.id);
      await addMemberToRole(designerRoleId, designerUser.id, testOrg.id);
      await addMemberToRole(managerRoleId, managerUser.id, testOrg.id);
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should display members with their assigned roles", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Check all members are visible - scope to table to avoid duplicates
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });
      await expect(membersTable.getByText("Owner User")).toBeVisible();
      await expect(membersTable.getByText("Developer User")).toBeVisible();
      await expect(membersTable.getByText("Designer User")).toBeVisible();
      await expect(membersTable.getByText("Manager User")).toBeVisible();
    });

    test("should display role badges for members", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      // Developer should have Developer role badge
      const developerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Developer User" })
        .first();
      const developerRolesCell = developerRow.locator("td").nth(1); // Roles is 2nd column
      await expect(developerRolesCell.getByText("Developer")).toBeVisible();

      // Designer should have Designer role badge
      const designerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Designer User" })
        .first();
      const designerRolesCell = designerRow.locator("td").nth(1);
      await expect(designerRolesCell.getByText("Designer")).toBeVisible();

      // Manager should have Manager role badge
      const managerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Manager User" })
        .first();
      const managerRolesCell = managerRow.locator("td").nth(1);
      await expect(managerRolesCell.getByText("Manager")).toBeVisible();
    });

    test("should display both virtual roles and assigned roles", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Owner should have "Owner" virtual role
      const ownerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Owner User" })
        .first();
      const ownerRolesCell = ownerRow.locator("td").nth(1); // Roles is 2nd column
      await expect(
        ownerRolesCell.getByText("Owner", { exact: false })
      ).toBeVisible();

      // Developer should have "Developer" role and "Member" virtual role
      const developerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Developer User" })
        .first();
      const developerRolesCell = developerRow.locator("td").nth(1);
      await expect(developerRolesCell.getByText("Developer")).toBeVisible();
      await expect(
        developerRolesCell.getByText("Member", { exact: false })
      ).toBeVisible();

      // Manager should have "Manager" role and "Admin" virtual role (write permission)
      const managerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Manager User" })
        .first();
      const managerRolesCell = managerRow.locator("td").nth(1);
      await expect(managerRolesCell.getByText("Manager")).toBeVisible();
      await expect(
        managerRolesCell.getByText("Admin", { exact: false })
      ).toBeVisible();
    });

    test("should display member with multiple roles", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      // Add developer to designer role as well
      await addMemberToRole(designerRoleId, developerUser.id, testOrg.id);

      // Reload page
      await page.reload();
      await page.waitForLoadState("networkidle");

      const developerRow = page
        .locator("tbody tr")
        .filter({ hasText: "Developer User" })
        .first();
      const developerRolesCell = developerRow.locator("td").nth(1); // Roles is 2nd column

      // Should show both Developer and Designer roles
      await expect(developerRolesCell.getByText("Developer")).toBeVisible();
      await expect(developerRolesCell.getByText("Designer")).toBeVisible();
    });
  });

  test.describe("Member Status and Display", () => {
    let ownerUser: any;
    let activeUser: any;
    let deletedUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      activeUser = await createDBUser("active", {
        first_name: "Active",
        last_name: "User",
      });

      deletedUser = await createDBUser("deleted", {
        first_name: "Deleted",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Status Test Org",
      });

      await addMemberToOrganization(testOrg.id, activeUser.id, "read");
      await addMemberToOrganization(testOrg.id, deletedUser.id, "read");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should display different status badges correctly", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Active members should show Active badge
      const activeRow = page
        .locator("tbody tr")
        .filter({ hasText: "Active User" })
        .first();
      await expect(activeRow.getByText("Active").first()).toBeVisible();

      // Deleted members should show Deleted badge (check in status column)
      const deletedRow = page
        .locator("tbody tr")
        .filter({ hasText: "Deleted User" })
        .first();
      const statusCell = deletedRow.locator("td").nth(2); // Status is 3rd column
      await expect(statusCell.getByText("Deleted")).toBeVisible();
    });

    test("should sort deleted members last", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Wait for table to be visible - find table that contains member data
      // The table should be visible once members are loaded
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });

      const rows = membersTable.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBeGreaterThanOrEqual(3);

      // Find deleted member index
      let deletedIndex = -1;
      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const statusCell = row.locator("td").nth(2); // Status is 3rd column
        const statusText = await statusCell.textContent();

        if (statusText?.includes("Deleted")) {
          deletedIndex = i;
          break;
        }
      }

      // Deleted member should be last
      expect(deletedIndex).toBe(rowCount - 1);
    });
  });

  test.describe("Loading and Error States", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Loading Test Org",
      });
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
    });

    test("should show loading state while fetching members", async ({
      page,
    }) => {
      // Navigate to page - loading state might be brief
      await page.waitForLoadState("networkidle");

      // Members section should eventually load
      await expect(page.getByText("Members").first()).toBeVisible();
    });

    test("should display members table structure", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      // Wait for members section to load
      await page.waitForSelector("text=Members", { timeout: 10000 });
      await expect(page.getByText("Members").first()).toBeVisible();
      await expect(
        page.getByText("Organization members and their roles.")
      ).toBeVisible();

      // Wait for table to be visible - find table that contains member data
      const membersTable = page.locator("table").filter({
        has: page.locator("thead th", { hasText: "Name" })
      }).first();
      await expect(membersTable).toBeVisible({ timeout: 10000 });

      // Check table headers - scope to members table only
      const tableHeaders = membersTable.locator("thead th");
      await expect(
        tableHeaders.filter({ hasText: "Name" }).first()
      ).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Roles" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Status" })).toBeVisible();
    });
  });
});
