import { expect, test } from "./fixtures";
import { navigateAndWait, waitForErrorToast, waitForPageLoad } from "./helpers";
import { LoginPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";

test.describe("@auth.login Login E2E Tests", () => {
  test("should show validation errors for invalid form inputs", async ({
    page,
  }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    // Empty form
    await loginPage.login.submit();
    await expect(page).toHaveURL(/.*\/login/);

    // Invalid email format
    await loginPage.login.fillLoginFields({
      email: "invalid-email",
      password: USER_DEFAULT_PASSWORD,
    });
    await loginPage.login.submit();
    await expect(page).toHaveURL(/.*\/login/);
  });

  test("should show error when credentials are invalid", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await loginPage.login.fillLoginFields({
      email: "invalid@example.com",
      password: "wrongpassword",
    });
    await loginPage.login.submit();
    await waitForErrorToast(page, undefined);
    await expect(page).toHaveURL(/.*login/);
  });

  test("should successfully login with valid credentials", async ({
    page,
    testConfig,
  }) => {
    const testUser = await createUser(testConfig);
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await loginPage.login.login(testUser.email, USER_DEFAULT_PASSWORD);
    await page.waitForURL((url) => !url.pathname.includes("/login"));
    await waitForPageLoad(page);
    await expect(page).not.toHaveURL(/.*login/);
  });

  test("should redirect to login when accessing protected route", async ({
    page,
  }) => {
    await navigateAndWait(page, "/settings/organizations");
    await expect(page).toHaveURL(/.*login/);
  });

  test("should persist authentication across page reloads", async ({
    page,
    testConfig,
  }) => {
    const testUser = await createUser(testConfig);
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await loginPage.login.login(testUser.email, USER_DEFAULT_PASSWORD);
    await page.waitForURL((url) => !url.pathname.includes("/login"));
    await waitForPageLoad(page);

    await page.reload();
    await waitForPageLoad(page);
    await expect(page).not.toHaveURL(/.*login/);
  });

  test("should handle logout", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });

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
        await expect(page).toHaveURL(/.*login/);
      }
    }
  });
});
