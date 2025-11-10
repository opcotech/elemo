import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import { createDBOrganization } from "./utils/organization";
import { generateXid } from "./utils/xid";
import { OrganizationPage } from "./pages/organization-page";
import { waitForPageLoad } from "./helpers/navigation";

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
    const orgPage = new OrganizationPage(page, testOrganization.id);
    await orgPage.goto();
    await expect(page).toHaveURL(
      new RegExp(`.*settings/organizations/${testOrganization.id}`)
    );
    await expect(
      page.getByRole("heading", { name: testOrganization.name })
    ).toBeVisible({ timeout: 10000 });
    await expect(
      page.getByText("Organization Information", { exact: true })
    ).toBeVisible({ timeout: 5000 });
    await expect(
      page.getByText("Details about the organization and its status.")
    ).toBeVisible();
    await expect(
      page.locator("label").filter({ hasText: "Name" })
    ).toBeVisible();
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
    const statusField = page
      .locator("label")
      .filter({ hasText: "Status" })
      .locator("..")
      .locator("div.mt-1");
    await expect(statusField.getByText("Active")).toBeVisible();

    await expect(
      page.locator("label").filter({ hasText: "Created At" })
    ).toBeVisible();
    const createdDate = page.locator(
      "text=/.*(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec).*\\d{4}.*/"
    );
    await expect(createdDate.first()).toBeVisible();
  });

  test("should display error message when navigating to non-existent organization", async ({
    page,
  }) => {
    const fakeOrgId = await generateXid();
    await page.goto(`/settings/organizations/${fakeOrgId}`);
    await waitForPageLoad(page);
    await expect(page).toHaveURL(
      new RegExp(`.*settings/organizations/${fakeOrgId}`)
    );
    await expect(
      page.getByRole("heading", { name: "Organization Details" })
    ).toBeVisible();
    await expect(
      page.getByText(
        "Organization not found. Please check the URL and try again."
      )
    ).toBeVisible();
    const errorAlert = page.locator('[role="alert"]');
    await expect(errorAlert).toBeVisible();
  });
});
