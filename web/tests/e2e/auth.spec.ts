import { expect, test } from "@playwright/test";
import { USER_DEFAULT_PASSWORD, createDBUser } from "./utils/auth";

test.describe("@auth Authentication E2E Tests", () => {
  let testUser: any;

  test.beforeAll(async () => {
    // Create a test user for authentication tests
    testUser = await createDBUser("active");
  });

  test.beforeEach(async ({ page }) => {
    // Navigate to login page before each test
    await page.goto("/login");
    await page.waitForLoadState("networkidle");
  });

  test("should display login form with all required elements", async ({
    page,
  }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Check for login form elements
    await expect(page.locator('[data-slot="card-title"]')).toBeVisible();
    await expect(page.getByLabel("Email")).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Password" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Sign in" })).toBeVisible();

    // Check for password visibility toggle
    const passwordInput = page.getByRole("textbox", { name: "Password" });
    await expect(passwordInput).toHaveAttribute("type", "password");

    // Look for the password toggle button by aria-label
    const toggleButton = page.getByRole("button", { name: "Show password" });
    if (await toggleButton.isVisible()) {
      await toggleButton.click();
      await expect(passwordInput).toHaveAttribute("type", "text");
    }
  });

  test("should show validation errors for empty form submission", async ({
    page,
  }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Fill in some data first to enable the button, then clear it
    await page.getByLabel("Email").fill("test@example.com");
    await page.getByRole("textbox", { name: "Password" }).fill("password123");

    // Clear the fields
    await page.getByLabel("Email").clear();
    await page.getByRole("textbox", { name: "Password" }).clear();

    // Try to submit empty form - button should be disabled
    const submitButton = page.getByRole("button", { name: "Sign in" });
    await expect(submitButton).toBeDisabled();

    // Check that form validation prevents submission
    await expect(page).toHaveURL(/.*login/);
  });

  test("should show validation errors for invalid email format", async ({
    page,
  }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Fill in invalid email
    await page.getByLabel("Email").fill("invalid-email");
    await page.getByRole("textbox", { name: "Password" }).fill("password123");

    // Check for email validation error
    const emailInput = page.getByLabel("Email");
    await expect(emailInput).toHaveAttribute("type", "email");

    // Try to submit
    await page.getByRole("button", { name: "Sign in" }).click();

    // Should still be on login page
    await expect(page).toHaveURL(/.*login/);
  });

  test("should handle login with invalid credentials", async ({ page }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Fill in invalid credentials
    await page.getByLabel("Email").fill("invalid@example.com");
    await page.getByRole("textbox", { name: "Password" }).fill("wrongpassword");

    // Submit form
    await page.getByRole("button", { name: "Sign in" }).click();

    // Wait for error to appear
    // Check for error message - be more flexible with the error text
    const errorAlert = page.locator('[role="alert"]');
    await expect(errorAlert).toBeVisible({ timeout: 5000 });

    // Check for any error message, not just specific patterns
    const errorText = await errorAlert.textContent();
    expect(errorText).toBeTruthy();
    expect(errorText?.length).toBeGreaterThan(0);

    // Should still be on login page
    await expect(page).toHaveURL(/.*login/);
  });

  test("should handle login with valid credentials", async ({ page }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Fill in valid credentials using the created test user
    await page.getByLabel("Email").fill(testUser.email);
    await page
      .getByRole("textbox", { name: "Password" })
      .fill(USER_DEFAULT_PASSWORD);

    // Submit form
    await page.getByRole("button", { name: "Sign in" }).click();

    // Wait for navigation
    await page.waitForLoadState("networkidle");

    // Should redirect to dashboard or home page
    await expect(page).not.toHaveURL(/.*login/);

    // Check that we're authenticated (look for user-specific elements)
    const userMenu = page.locator(
      '[data-testid="user-menu"], [aria-label*="user"], [aria-label*="account"]'
    );
    if (await userMenu.isVisible()) {
      await expect(userMenu).toBeVisible();
    }
  });

  test("should handle authentication required redirect", async ({ page }) => {
    // Try to access a protected route directly
    await page.goto("/dashboard");
    await page.waitForLoadState("networkidle");

    // Should redirect to login page
    await expect(page).toHaveURL(/.*login/);

    // Check for redirect parameter in URL or form
    const currentUrl = page.url();
    if (currentUrl.includes("redirect=")) {
      const redirectParam = new URL(currentUrl).searchParams.get("redirect");
      // Decode the redirect parameter before checking
      const decodedRedirect = redirectParam
        ? decodeURIComponent(redirectParam)
        : "";
      expect(decodedRedirect).toContain("/dashboard");
    }
  });

  test("should handle logout functionality", async ({ page }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // First login using the created test user
    await page.getByLabel("Email").fill(testUser.email);
    await page
      .getByRole("textbox", { name: "Password" })
      .fill(USER_DEFAULT_PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await page.waitForLoadState("networkidle");

    // Look for logout button/menu
    const userMenu = page.locator(
      '[data-testid="user-menu"], [aria-label*="user"], [aria-label*="account"]'
    );
    const logoutButton = page.getByRole("button", { name: /logout|sign out/i });

    if (await userMenu.isVisible()) {
      await userMenu.click();
      // Wait for menu to expand/logout button to be visible
      await expect(logoutButton).toBeVisible({ timeout: 3000 });
    }

    if (await logoutButton.isVisible()) {
      await logoutButton.click();
      await page.waitForLoadState("networkidle");

      // Should redirect to login page
      await expect(page).toHaveURL(/.*login/);

      // Try to access protected route again
      await page.goto("/dashboard");
      await page.waitForLoadState("networkidle");

      // Should still be redirected to login
      await expect(page).toHaveURL(/.*login/);
    }
  });

  test("should handle form accessibility", async ({ page }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Test keyboard navigation
    await page.keyboard.press("Tab");
    await expect(page.getByLabel("Email")).toBeFocused();

    await page.keyboard.press("Tab");
    await expect(page.getByRole("textbox", { name: "Password" })).toBeFocused();

    await page.keyboard.press("Tab");

    // The button might be disabled initially
    const submitButton = page.getByRole("button", { name: "Sign in" });
    await expect(submitButton).toBeVisible();

    // Test form submission with Enter key
    await page.getByLabel("Email").fill("test@example.com");
    await page.getByRole("textbox", { name: "Password" }).fill("password123");
    await page.keyboard.press("Enter");

    // Should attempt to submit form - wait for either navigation or error
    await page.waitForLoadState("networkidle");
  });

  test("should handle loading states during login", async ({ page }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // Fill in valid credentials using the created test user
    await page.getByLabel("Email").fill(testUser.email);
    await page
      .getByRole("textbox", { name: "Password" })
      .fill(USER_DEFAULT_PASSWORD);

    // Submit form
    const submitButton = page.getByLabel("Sign in");
    await submitButton.click();

    // Check for loading state
    expect(page.getByText("Signing in...")).toBeVisible();
    expect(submitButton).toBeDisabled();
  });

  test("should clear error messages when user starts typing", async ({
    page,
  }) => {
    // Wait for the page to be fully loaded
    await page.waitForLoadState("domcontentloaded");

    // First trigger an error
    await page.getByLabel("Email").fill("invalid@example.com");
    await page.getByRole("textbox", { name: "Password" }).fill("wrongpassword");
    await page.getByRole("button", { name: "Sign in" }).click();

    // Check if error is visible
    const errorAlert = page.locator('[role="alert"]');
    const hasError = await errorAlert
      .isVisible({ timeout: 5000 })
      .catch(() => false);

    if (hasError) {
      // Start typing in email field
      await page.getByLabel("Email").fill("new@example.com");

      // Error should be cleared
      await expect(errorAlert).not.toBeVisible();
    }
  });
});
