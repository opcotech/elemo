import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";

test.describe("@settings.organizations Organization Listing E2E Tests", () => {
  test.describe("Owner Scenarios", () => {
    let ownerUser: any;
    let ownedOrg1: any;
    let ownedOrg2: any;
    let deletedOwnedOrg: any;

    test.beforeAll(async () => {
      // Create an owner user
      ownerUser = await createDBUser("active");

      // Create organizations owned by this user
      ownedOrg1 = await createDBOrganization(ownerUser.id, "active", {
        name: "Owner Org Alpha",
      });
      ownedOrg2 = await createDBOrganization(ownerUser.id, "active", {
        name: "Owner Org Beta",
      });
      deletedOwnedOrg = await createDBOrganization(ownerUser.id, "deleted", {
        name: "Owner Org Deleted",
      });
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
    });

    test("owner should see all organizations they own", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Owner Org Alpha")).toBeVisible();
      await expect(page.getByText("Owner Org Beta")).toBeVisible();
      await expect(page.getByText("Owner Org Deleted")).toBeVisible();
    });

    test("owner should have all action buttons (read, write, delete permissions)", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const alphaRow = page
        .locator("tbody tr")
        .filter({ hasText: "Owner Org Alpha" })
        .first();

      // Owner has "*" permission, so should see all buttons
      await expect(
        alphaRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      // Edit button is disabled but should be visible (getByRole handles sr-only spans)
      await expect(
        alphaRow.getByRole("button", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        alphaRow.getByRole("button", { name: "Delete organization" })
      ).toBeVisible();
    });

    test("owner should see organizations sorted correctly (active first, then alphabetically)", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const rows = page.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBeGreaterThanOrEqual(3);

      // Check that active organizations appear before deleted ones
      let foundDeleted = false;
      let lastActiveIndex = -1;

      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const statusCell = row.locator("td").nth(3);
        const statusText = await statusCell.textContent();

        if (statusText?.includes("Deleted")) {
          foundDeleted = true;
          if (lastActiveIndex === -1) {
            lastActiveIndex = i - 1;
          }
        } else if (statusText?.includes("Active")) {
          if (foundDeleted) {
            throw new Error(
              "Active organization found after deleted organization"
            );
          }
          lastActiveIndex = i;
        }
      }

      // Within active organizations, check alphabetical order
      const activeRows: Array<string> = [];
      for (let i = 0; i <= lastActiveIndex; i++) {
        const row = rows.nth(i);
        const nameCell = row.locator("td").first();
        const nameText = await nameCell.textContent();
        if (nameText) {
          activeRows.push(nameText.trim());
        }
      }

      const sortedActiveRows = [...activeRows].sort();
      expect(activeRows).toEqual(sortedActiveRows);
    });
  });

  test.describe("Member Scenarios - Single Organization", () => {
    let ownerUser: any;
    let memberUser: any;
    let memberOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      memberUser = await createDBUser("active");

      // Create organization owned by ownerUser
      memberOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Member Organization",
      });

      // Add memberUser as a member with read permission
      await addMemberToOrganization(memberOrg.id, memberUser.id, "read");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, memberUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
    });

    test("member with read permission should see the organization", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Member Organization")).toBeVisible();
    });

    test("member with read permission should only see view button", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const memberRow = page
        .locator("tbody tr")
        .filter({ hasText: "Member Organization" })
        .first();

      // Should have view button
      await expect(
        memberRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();

      // Should NOT have edit or delete buttons
      await expect(
        memberRow.getByRole("button", { name: "Edit organization" })
      ).not.toBeVisible();
      await expect(
        memberRow.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
    });
  });

  test.describe("Member Scenarios - Multiple Organizations", () => {
    let ownerUser1: any;
    let ownerUser2: any;
    let memberUser: any;
    let org1: any;
    let org2: any;
    let org3: any;

    test.beforeAll(async () => {
      ownerUser1 = await createDBUser("active");
      ownerUser2 = await createDBUser("active");
      memberUser = await createDBUser("active");

      // Create multiple organizations
      org1 = await createDBOrganization(ownerUser1.id, "active", {
        name: "Org A - Read Only",
      });
      org2 = await createDBOrganization(ownerUser1.id, "active", {
        name: "Org B - Write Access",
      });
      org3 = await createDBOrganization(ownerUser2.id, "active", {
        name: "Org C - Full Access",
      });

      // Add memberUser to all organizations with different permissions
      await addMemberToOrganization(org1.id, memberUser.id, "read");
      await addMemberToOrganization(org2.id, memberUser.id, "write");
      await addMemberToOrganization(org3.id, memberUser.id, "*");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, memberUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
    });

    test("member should see all organizations they are a member of", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Org A - Read Only")).toBeVisible();
      await expect(page.getByText("Org B - Write Access")).toBeVisible();
      await expect(page.getByText("Org C - Full Access")).toBeVisible();
    });

    test("member should see correct action buttons based on permissions", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      // Check Org A - Read Only (read permission)
      const orgA = page
        .locator("tbody tr")
        .filter({ hasText: "Org A - Read Only" })
        .first();

      await expect(
        orgA.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgA.getByRole("button", { name: "Edit organization" })
      ).not.toBeVisible();
      await expect(
        orgA.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();

      // Check Org B - Write Access (write permission)
      const orgB = page
        .locator("tbody tr")
        .filter({ hasText: "Org B - Write Access" })
        .first();

      await expect(
        orgB.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgB.getByRole("button", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgB.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();

      // Check Org C - Full Access (* permission)
      const orgC = page
        .locator("tbody tr")
        .filter({ hasText: "Org C - Full Access" })
        .first();

      await expect(
        orgC.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgC.getByRole("button", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgC.getByRole("button", { name: "Delete organization" })
      ).toBeVisible();
    });

    test("member should see organizations sorted correctly", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const rows = page.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBeGreaterThanOrEqual(3);

      // Get organization names in order
      const orgNames: Array<string> = [];
      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const nameCell = row.locator("td").first();
        const nameText = await nameCell.textContent();
        if (nameText) {
          orgNames.push(nameText.trim());
        }
      }

      // Organizations should be sorted alphabetically (all are active)
      const sortedNames = [...orgNames].sort();
      expect(orgNames).toEqual(sortedNames);
    });
  });

  test.describe("Permission-Based UI Visibility", () => {
    let ownerUser: any;
    let readMember: any;
    let writeMember: any;
    let deleteMember: any;
    let fullAccessMember: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      readMember = await createDBUser("active");
      writeMember = await createDBUser("active");
      deleteMember = await createDBUser("active");
      fullAccessMember = await createDBUser("active");

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Permission Test Org",
      });

      // Add members with different permissions
      await addMemberToOrganization(testOrg.id, readMember.id, "read");
      await addMemberToOrganization(testOrg.id, writeMember.id, "write");
      await addMemberToOrganization(testOrg.id, deleteMember.id, "delete");
      await addMemberToOrganization(testOrg.id, fullAccessMember.id, "*");
    });

    test("read-only member should only see view button", async ({ page }) => {
      await loginUser(page, readMember, {
        destination: "/settings/organizations",
      });

      const orgRow = page
        .locator("tbody tr")
        .filter({ hasText: "Permission Test Org" })
        .first();

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Edit organization" })
      ).not.toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
    });

    test("write member should see view and edit buttons", async ({ page }) => {
      await loginUser(page, writeMember, {
        destination: "/settings/organizations",
      });

      const orgRow = page
        .locator("tbody tr")
        .filter({ hasText: "Permission Test Org" })
        .first();

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
    });

    test("full access member should see all action buttons", async ({
      page,
    }) => {
      await loginUser(page, fullAccessMember, {
        destination: "/settings/organizations",
      });

      const orgRow = page
        .locator("tbody tr")
        .filter({ hasText: "Permission Test Org" })
        .first();

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Delete organization" })
      ).toBeVisible();
    });
  });

  test.describe("Common Functionality", () => {
    let testUser: any;
    let testOrg1: any;
    let testOrg2: any;

    test.beforeAll(async () => {
      testUser = await createDBUser("active");

      testOrg1 = await createDBOrganization(testUser.id, "active", {
        name: "Alpha Organization",
      });
      testOrg2 = await createDBOrganization(testUser.id, "active", {
        name: "Beta Organization",
      });
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, testUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
    });

    test("should display organization list page with all required elements", async ({
      page,
    }) => {
      await expect(
        page.getByRole("heading", { name: "Organizations" })
      ).toBeVisible();
      await expect(
        page.getByText("View and manage organizations.").first()
      ).toBeVisible();

      await expect(
        page.getByPlaceholder("Search organizations...")
      ).toBeVisible();

      // Check table headers using text content instead of exact role match
      await page.waitForLoadState("networkidle");
      const tableHeaders = page.locator("thead th");
      await expect(tableHeaders.filter({ hasText: "Name" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Email" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Website" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Members" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Status" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Actions" })).toBeVisible();
    });

    test("should filter organizations by search term", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("Alpha");
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Alpha Organization")).toBeVisible();
      await expect(page.getByText("Beta Organization")).not.toBeVisible();

      await searchInput.clear();
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Alpha Organization")).toBeVisible();
      await expect(page.getByText("Beta Organization")).toBeVisible();
    });

    test("should show empty state when search has no results", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("NonExistentOrganization123");
      await page.waitForLoadState("networkidle");

      await expect(
        page.getByText("No organizations found matching your search.")
      ).toBeVisible();
    });

    test("should handle case-insensitive search", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("alpha");
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Alpha Organization")).toBeVisible();

      await searchInput.clear();
      await searchInput.fill("BETA");
      await page.waitForLoadState("networkidle");

      await expect(page.getByText("Beta Organization")).toBeVisible();
    });

    test("should navigate to organization detail page on view click", async ({
      page,
    }) => {
      await page.waitForLoadState("networkidle");

      const alphaRow = page
        .locator("tbody tr")
        .filter({ hasText: "Alpha Organization" })
        .first();

      const viewButton = alphaRow
        .locator('a[href*="/settings/organizations/"]')
        .first();

      if (await viewButton.isVisible()) {
        await viewButton.click();
        await page.waitForLoadState("networkidle");

        await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
      }
    });

    test("should display status badges correctly", async ({ page }) => {
      await page.waitForLoadState("networkidle");

      const activeRow = page
        .locator("tbody tr")
        .filter({ hasText: "Alpha Organization" })
        .first();
      const activeStatusBadge = activeRow
        .locator('[data-slot="badge"]')
        .filter({ hasText: "Active" })
        .first();
      await expect(activeStatusBadge).toBeVisible();
      await expect(activeStatusBadge).toHaveText("Active");
    });
  });
});
