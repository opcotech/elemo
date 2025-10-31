import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import { createDBOrganization } from "./utils/organization";
import { generateXid } from "./utils/xid";

test.describe("@settings.organization-details Organization Details E2E Tests", () => {
  let ownerUser: any;
  let testOrganization: any;

  test.beforeAll(async () => {
    ownerUser = await createDBUser("active");
    testOrganization = await createDBOrganization(ownerUser.id, "active", {
      name: "Test Organization Details",
      website: "https://test-org.example.com",
    });
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, ownerUser);
  });

  test("should navigate to an existing organization and display details", async ({
    page,
  }) => {
    // Navigate to organization details page
    await page.goto(`/settings/organizations/${testOrganization.id}`);
    await page.waitForLoadState("networkidle");

    // Wait for the page to load and the query to complete
    // Wait for either the organization name (success) or error message
    await page.waitForFunction(
      () => {
        const heading = document.querySelector("h1");
        const alert = document.querySelector('[role="alert"]');
        return (
          (heading &&
            heading.textContent.includes("Test Organization Details")) ||
          alert !== null
        );
      },
      { timeout: 10000 }
    );

    // Verify URL
    await expect(page).toHaveURL(
      new RegExp(`.*settings/organizations/${testOrganization.id}`)
    );

    // Verify header shows organization name
    await expect(
      page.getByRole("heading", { name: testOrganization.name })
    ).toBeVisible({ timeout: 5000 });

    // Verify description is visible
    await expect(
      page.getByText("View organization information.")
    ).toBeVisible();

    // Verify card title - wait for it to appear
    // CardTitle is a div, not a heading, so we use getByText with exact match
    await expect(
      page.getByText("Organization Information", { exact: true })
    ).toBeVisible({ timeout: 5000 });

    // Verify card description
    await expect(
      page.getByText("Details about the organization and its status.")
    ).toBeVisible();

    // Verify organization fields are displayed
    await expect(
      page.locator("label").filter({ hasText: "Name" })
    ).toBeVisible();
    // Organization name appears in multiple places, check the one in the Name field
    const nameField = page.locator("text=Name").locator("..").locator("p");
    await expect(
      nameField.filter({ hasText: testOrganization.name })
    ).toBeVisible();

    await expect(
      page.locator("label").filter({ hasText: "Email" })
    ).toBeVisible();
    const emailField = page.locator("text=Email").locator("..").locator("p");
    await expect(
      emailField.filter({ hasText: testOrganization.email })
    ).toBeVisible();

    await expect(page.getByText("Website")).toBeVisible();
    const websiteLink = page
      .locator('a[href="https://test-org.example.com"]')
      .first();
    await expect(websiteLink).toBeVisible();
    await expect(websiteLink).toHaveAttribute("target", "_blank");
    await expect(websiteLink).toHaveAttribute("rel", "noopener noreferrer");

    await expect(
      page.locator("label").filter({ hasText: "Status" })
    ).toBeVisible();
    // "Active" appears in both organization details and members table, check in the details section
    const statusField = page
      .locator("label")
      .filter({ hasText: "Status" })
      .locator("..")
      .locator("div.mt-1");
    await expect(statusField.getByText("Active")).toBeVisible();

    await expect(
      page.locator("label").filter({ hasText: "Created At" })
    ).toBeVisible();
    // Created date should be formatted and visible
    const createdDate = page.locator(
      "text=/.*(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec).*\\d{4}.*/"
    );
    await expect(createdDate.first()).toBeVisible();
  });

  test("should display error message when navigating to non-existent organization", async ({
    page,
  }) => {
    // Generate a fake organization ID
    const fakeOrgId = await generateXid();

    // Navigate to non-existent organization details page
    await page.goto(`/settings/organizations/${fakeOrgId}`);
    await page.waitForLoadState("networkidle");

    // Verify URL
    await expect(page).toHaveURL(
      new RegExp(`.*settings/organizations/${fakeOrgId}`)
    );

    // Verify error header
    await expect(
      page.getByRole("heading", { name: "Organization Details" })
    ).toBeVisible();

    // Verify error message
    await expect(
      page.getByText(
        "Organization not found. Please check the URL and try again."
      )
    ).toBeVisible();

    // Verify error alert is visible
    const errorAlert = page.locator('[role="alert"]');
    await expect(errorAlert).toBeVisible();
  });
});
