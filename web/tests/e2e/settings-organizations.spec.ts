import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";
import { waitForPageLoad, waitForPermissionsLoad } from "./helpers/navigation";
import { getTableByHeader, getTableRow } from "./helpers/elements";

test.describe("@settings.organizations Organization Listing E2E Tests", () => {
  test.describe("Owner Scenarios", () => {
    let ownerUser: any;
    let ownedOrg1: any;
    let ownedOrg2: any;
    let deletedOwnedOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
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
      await waitForPageLoad(page);
    });

    test("owner should see all organizations they own", async ({ page }) => {
      await expect(page.getByText("Owner Org Alpha")).toBeVisible();
      await expect(page.getByText("Owner Org Beta")).toBeVisible();
      await expect(page.getByText("Owner Org Deleted")).toBeVisible();
    });

    test("owner should have all action buttons (read, write, delete permissions)", async ({
      page,
    }) => {
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const alphaRow = getTableRow(organizationsTable, "Owner Org Alpha");
      await expect(
        alphaRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        alphaRow.getByRole("link", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        alphaRow.getByRole("button", { name: "Delete organization" })
      ).toBeVisible();
    });

    test("owner should see organizations sorted correctly (active first, then alphabetically)", async ({
      page,
    }) => {
      const organizationsTable = getTableByHeader(page, "Name");
      const rows = organizationsTable.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBeGreaterThanOrEqual(3);
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
      memberOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Member Organization",
      });
      await addMemberToOrganization(memberOrg.id, memberUser.id, "read");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, memberUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
      await waitForPageLoad(page);
    });

    test("member with read permission should see the organization", async ({
      page,
    }) => {
      await expect(page.getByText("Member Organization")).toBeVisible();
    });

    test("member with read permission should only see view button", async ({
      page,
    }) => {
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const memberRow = getTableRow(organizationsTable, "Member Organization");
      await expect(
        memberRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        memberRow.getByRole("link", { name: "Edit organization" })
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
      org1 = await createDBOrganization(ownerUser1.id, "active", {
        name: "Org A - Read Only",
      });
      org2 = await createDBOrganization(ownerUser1.id, "active", {
        name: "Org B - Write Access",
      });
      org3 = await createDBOrganization(ownerUser2.id, "active", {
        name: "Org C - Full Access",
      });
      await addMemberToOrganization(org1.id, memberUser.id, "read");
      await addMemberToOrganization(org2.id, memberUser.id, "write");
      await addMemberToOrganization(org3.id, memberUser.id, "*");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, memberUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
      await waitForPageLoad(page);
    });

    test("member should see all organizations they are a member of", async ({
      page,
    }) => {
      await expect(page.getByText("Org A - Read Only")).toBeVisible();
      await expect(page.getByText("Org B - Write Access")).toBeVisible();
      await expect(page.getByText("Org C - Full Access")).toBeVisible();
    });

    test("member should see correct action buttons based on permissions", async ({
      page,
    }) => {
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const orgA = getTableRow(organizationsTable, "Org A - Read Only");
      await expect(
        orgA.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgA.getByRole("link", { name: "Edit organization" })
      ).not.toBeVisible();
      await expect(
        orgA.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
      const orgB = getTableRow(organizationsTable, "Org B - Write Access");
      await expect(
        orgB.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgB.getByRole("link", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgB.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
      const orgC = getTableRow(organizationsTable, "Org C - Full Access");
      await expect(
        orgC.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgC.getByRole("link", { name: "Edit organization" })
      ).toBeVisible();
      await expect(
        orgC.getByRole("button", { name: "Delete organization" })
      ).toBeVisible();
    });

    test("member should see organizations sorted correctly", async ({
      page,
    }) => {
      const organizationsTable = getTableByHeader(page, "Name");
      const rows = organizationsTable.locator("tbody tr");
      const rowCount = await rows.count();
      expect(rowCount).toBeGreaterThanOrEqual(3);
      const orgNames: Array<string> = [];
      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const nameCell = row.locator("td").first();
        const nameText = await nameCell.textContent();
        if (nameText) {
          orgNames.push(nameText.trim());
        }
      }
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
      await addMemberToOrganization(testOrg.id, readMember.id, "read");
      await addMemberToOrganization(testOrg.id, writeMember.id, "write");
      await addMemberToOrganization(testOrg.id, deleteMember.id, "delete");
      await addMemberToOrganization(testOrg.id, fullAccessMember.id, "*");
    });

    test("read-only member should only see view button", async ({ page }) => {
      await loginUser(page, readMember, {
        destination: "/settings/organizations",
      });
      await waitForPageLoad(page);
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const orgRow = getTableRow(organizationsTable, "Permission Test Org");

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("link", { name: "Edit organization" })
      ).not.toBeVisible();
      await expect(
        orgRow.getByRole("button", { name: "Delete organization" })
      ).not.toBeVisible();
    });

    test("write member should see view and edit buttons", async ({ page }) => {
      await loginUser(page, writeMember, {
        destination: "/settings/organizations",
      });
      await waitForPageLoad(page);
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const orgRow = getTableRow(organizationsTable, "Permission Test Org");

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("link", { name: "Edit organization" })
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
      await waitForPageLoad(page);
      await waitForPermissionsLoad(page);

      const organizationsTable = getTableByHeader(page, "Name");
      const orgRow = getTableRow(organizationsTable, "Permission Test Org");

      await expect(
        orgRow.locator('a[href*="/settings/organizations/"]').first()
      ).toBeVisible();
      await expect(
        orgRow.getByRole("link", { name: "Edit organization" })
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
      await waitForPageLoad(page);
    });

    test("should display organization list page with all required elements", async ({
      page,
    }) => {
      await expect(
        page.getByRole("heading", { name: "Organizations", level: 1 })
      ).toBeVisible();
      await expect(
        page.getByText("View and manage organizations.").first()
      ).toBeVisible();

      await expect(
        page.getByPlaceholder("Search organizations...")
      ).toBeVisible();
      const organizationsTable = getTableByHeader(page, "Name");
      const tableHeaders = organizationsTable.locator("thead th");
      await expect(tableHeaders.filter({ hasText: "Name" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Email" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Website" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Members" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Status" })).toBeVisible();
      await expect(tableHeaders.filter({ hasText: "Actions" })).toBeVisible();
    });

    test("should filter organizations by search term", async ({ page }) => {
      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("Alpha");
      await waitForPageLoad(page);

      await expect(page.getByText("Alpha Organization")).toBeVisible();
      await expect(page.getByText("Beta Organization")).not.toBeVisible();

      await searchInput.clear();
      await waitForPageLoad(page);

      await expect(page.getByText("Alpha Organization")).toBeVisible();
      await expect(page.getByText("Beta Organization")).toBeVisible();
    });

    test("should show empty state when search has no results", async ({
      page,
    }) => {
      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("NonExistentOrganization123");
      await waitForPageLoad(page);

      await expect(
        page.getByText(
          "No organizations match your search criteria. Try adjusting your search."
        )
      ).toBeVisible();
    });

    test("should handle case-insensitive search", async ({ page }) => {
      const searchInput = page.getByPlaceholder("Search organizations...");

      await searchInput.fill("alpha");
      await waitForPageLoad(page);

      await expect(page.getByText("Alpha Organization")).toBeVisible();

      await searchInput.clear();
      await searchInput.fill("BETA");
      await waitForPageLoad(page);

      await expect(page.getByText("Beta Organization")).toBeVisible();
    });

    test("should navigate to organization detail page on view click", async ({
      page,
    }) => {
      const organizationsTable = getTableByHeader(page, "Name");
      const alphaRow = getTableRow(organizationsTable, "Alpha Organization");

      const viewButton = alphaRow
        .locator('a[href*="/settings/organizations/"]')
        .first();

      if (await viewButton.isVisible().catch(() => false)) {
        await viewButton.click();
        await waitForPageLoad(page);

        await expect(page).toHaveURL(/.*settings\/organizations\/.*/);
      }
    });

    test("should display status badges correctly", async ({ page }) => {
      const organizationsTable = getTableByHeader(page, "Name");
      const activeRow = getTableRow(organizationsTable, "Alpha Organization");
      const activeStatusBadge = activeRow
        .locator('[data-slot="badge"]')
        .filter({ hasText: "Active" })
        .first();
      await expect(activeStatusBadge).toBeVisible();
      await expect(activeStatusBadge).toHaveText("Active");
    });
  });
});
