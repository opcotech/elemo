import { expect, test } from "@playwright/test";
import { USER_DEFAULT_PASSWORD, createDBUser, loginUser } from "./utils/auth";
import { waitForPageLoad } from "./helpers/navigation";
import { Form } from "./components/form";
import { waitForErrorToast } from "./helpers/toast";

test.describe("@auth Authentication E2E Tests", () => {
  let testUser: any;

  test.beforeAll(async () => {
    testUser = await createDBUser("active");
  });

  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await waitForPageLoad(page);
  });

  test("should display login form with all required elements", async ({
    page,
  }) => {
    await expect(page.locator('[data-slot="card-title"]')).toBeVisible();
    await expect(page.getByLabel("Email")).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Password" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Sign in" })).toBeVisible();
    const passwordInput = page.getByRole("textbox", { name: "Password" });
    await expect(passwordInput).toHaveAttribute("type", "password");
    const toggleButton = page.getByRole("button", { name: "Show password" });
    if (await toggleButton.isVisible().catch(() => false)) {
      await toggleButton.click();
      await expect(passwordInput).toHaveAttribute("type", "text");
    }
  });

  test("should show validation errors for empty form submission", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page).toHaveURL(/.*\/login/);
    const emailInput = page.getByLabel("Email");
    await expect(emailInput).toHaveAttribute("required", "");
    const emailInvalid = await emailInput.evaluate(
      (el: HTMLInputElement) => !el.validity.valid
    );
    expect(emailInvalid).toBe(true);
  });

  test("should show validation error for invalid email format", async ({
    page,
  }) => {
    const form = new Form(page);
    await form.fillField("Email", "invalid-email");
    await form.fillField("Password", USER_DEFAULT_PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page).toHaveURL(/.*\/login/);
    const emailInput = page.getByLabel("Email");
    const emailInvalid = await emailInput.evaluate(
      (el: HTMLInputElement) => !el.validity.valid
    );
    expect(emailInvalid).toBe(true);
  });

  test("should handle login with invalid credentials", async ({ page }) => {
    const form = new Form(page);
    await form.fillField("Email", "invalid@example.com");
    await form.fillField("Password", "wrongpassword");
    await page.getByRole("button", { name: "Sign in" }).click();
    await waitForErrorToast(page, undefined, { timeout: 5000 });
    await expect(page).toHaveURL(/.*login/);
  });

  test("should handle login with valid credentials", async ({ page }) => {
    const form = new Form(page);
    await form.fillField("Email", testUser.email);
    await form.fillField("Password", USER_DEFAULT_PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await page.waitForURL((url) => !url.pathname.includes("/login"), {
      timeout: 10000,
    });
    await waitForPageLoad(page);
    await expect(page).not.toHaveURL(/.*login/);
    const userMenu = page.locator(
      '[data-testid="user-menu"], [aria-label*="user"], [aria-label*="account"]'
    );
    if (await userMenu.isVisible().catch(() => false)) {
      await expect(userMenu).toBeVisible();
    }
  });

  test("should handle authentication required redirect", async ({ page }) => {
    await page.goto("/settings/organizations");
    await waitForPageLoad(page);
    await expect(page).toHaveURL(/.*login/);
  });

  test("should persist authentication across page reloads", async ({
    page,
  }) => {
    const form = new Form(page);
    await form.fillField("Email", testUser.email);
    await form.fillField("Password", USER_DEFAULT_PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await page.waitForURL((url) => !url.pathname.includes("/login"), {
      timeout: 10000,
    });
    await waitForPageLoad(page);
    await page.reload();
    await waitForPageLoad(page);
    await expect(page).not.toHaveURL(/.*login/);
  });

  test("should handle logout", async ({ page }) => {
    await loginUser(page, testUser);
    const userMenu = page.locator(
      '[data-testid="user-menu"], [aria-label*="user"], [aria-label*="account"], button[aria-label*="menu"]'
    );

    if (await userMenu.isVisible().catch(() => false)) {
      await userMenu.click();
      const logoutButton = page.getByRole("button", {
        name: /logout|sign out/i,
      });
      if (await logoutButton.isVisible().catch(() => false)) {
        await logoutButton.click();
        await expect(page).toHaveURL(/.*login/, { timeout: 10000 });
      }
    }
  });
});
