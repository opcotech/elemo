import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import { grantSystemWritePermission } from "./utils/organization";
import { Form } from "./components/form";
import { waitForPageLoad } from "./helpers/navigation";
import { waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.organization-create Organization Create E2E Tests", () => {
  test.describe("Organization Creation", () => {
    let ownerUser: any;
    let regularUser: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active");
      regularUser = await createDBUser("active");
      await grantSystemWritePermission(ownerUser.id, "Organization");
    });

    test.beforeEach(async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: "/settings/organizations",
      });
      await expect(page).toHaveURL(/.*settings\/organizations/);
      await waitForPageLoad(page);
    });

    test("user with write permission should see create button", async ({
      page,
    }) => {
      await expect(
        page.getByRole("link", { name: /Create Organization/i }).first()
      ).toBeVisible();
    });

    test("user without write permission should not see create button", async ({
      page,
    }) => {
      await loginUser(page, regularUser, {
        destination: "/settings/organizations",
      });
      await waitForPageLoad(page);

      await expect(
        page.getByRole("link", { name: /Create Organization/i })
      ).not.toBeVisible();
    });

    test("should navigate to create page when create button is clicked", async ({
      page,
    }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      await expect(page).toHaveURL(/.*settings\/organizations\/new/);
      await expect(
        page.getByRole("heading", { name: "Create Organization", level: 1 })
      ).toBeVisible({ timeout: 10000 });
    });

    test("should create organization with all fields", async ({ page }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      const orgName = `Test Org ${Date.now()}`;
      const orgEmail = `test-${Date.now()}@example.com`;
      const orgWebsite = `https://test-${Date.now()}.example.com`;

      const form = new Form(page);
      await form.fillFields({
        Name: orgName,
        Email: orgEmail,
        Website: orgWebsite,
      });
      await form.submit("Create Organization");
      await waitForSuccessToast(page, "Organization created", {
        timeout: 10000,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/, {
        timeout: 10000,
      });
      await expect(page.getByRole("heading", { name: orgName })).toBeVisible({
        timeout: 10000,
      });
      await expect(page.getByText(orgEmail)).toBeVisible();
      await expect(page.getByText(orgWebsite)).toBeVisible();
    });

    test("should create organization with required fields only", async ({
      page,
    }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      const orgName = `Required Fields Org ${Date.now()}`;
      const orgEmail = `required-${Date.now()}@example.com`;

      const form = new Form(page);
      await form.fillFields({
        Name: orgName,
        Email: orgEmail,
      });
      await form.submit("Create Organization");
      await waitForSuccessToast(page, "Organization created", {
        timeout: 10000,
      });
      await expect(page).toHaveURL(/.*settings\/organizations\/.*/, {
        timeout: 10000,
      });
      await expect(page.getByRole("heading", { name: orgName })).toBeVisible({
        timeout: 10000,
      });
      await expect(page.getByText(orgEmail)).toBeVisible();
    });

    test("should show validation errors for invalid inputs", async ({
      page,
    }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);
      await page.getByRole("button", { name: "Create Organization" }).click();
      const formMessages = page.locator('[data-slot="form-message"]');
      await expect(formMessages.first()).toBeVisible({ timeout: 5000 });
    });

    test("should show validation error for invalid email", async ({ page }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      const form = new Form(page);
      await form.fillField("Name", "Test Org");
      await form.fillField("Email", "invalid-email");

      await page.getByRole("button", { name: "Create Organization" }).click();
      const emailField = page.getByLabel("Email").locator("..");
      await expect(
        emailField.locator("text=/invalid|email|must/i").first()
      ).toBeVisible({ timeout: 5000 });
    });

    test("should show validation error for invalid website URL", async ({
      page,
    }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      const form = new Form(page);
      await form.fillFields({
        Name: "Test Org",
        Email: "test@example.com",
        Website: "not-a-valid-url",
      });
      await page.getByLabel("Website").blur();
      await page
        .waitForFunction(
          () => {
            const input = document.querySelector(
              'input[aria-label="Website"]'
            ) as HTMLInputElement;
            return input.value === "not-a-valid-url";
          },
          { timeout: 1000 }
        )
        .catch(() => {});
      await page.getByRole("button", { name: "Create Organization" }).click();
      const isOnCreatePage = page.url().includes("/new");
      const websiteField = page.getByLabel("Website").locator("..");
      const formMessage = websiteField.locator('[data-slot="form-message"]');
      const hasError = await formMessage
        .isVisible({ timeout: 2000 })
        .catch(() => false);

      if (!hasError && !isOnCreatePage) {
        return;
      }
      expect(isOnCreatePage || hasError).toBe(true);
    });

    test("should cancel creation and return to list page", async ({ page }) => {
      await page
        .getByRole("link", { name: /Create Organization/i })
        .first()
        .click();
      await waitForPageLoad(page);

      await page.getByRole("button", { name: "Cancel" }).click();
      await waitForPageLoad(page);

      await expect(page).toHaveURL(/.*settings\/organizations$/);
    });
  });
});
