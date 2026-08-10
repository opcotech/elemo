import { expect, test } from "./fixtures";
import { navigateAndWait, waitForErrorToast, waitForPageLoad } from "./helpers";
import { LoginPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser, logoutUser } from "./utils/auth";

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
    userPersona,
  }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await loginPage.login.login(
      userPersona.credentials.email,
      userPersona.credentials.password
    );
    await page.waitForURL((url) => !url.pathname.includes("/login"));
    await waitForPageLoad(page);
    await expect(page).not.toHaveURL(/.*login/);
    // Home route loader fetches namespaces + todos; allow pending spinner to clear.
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible({
      timeout: 15_000,
    });
  });

  test("should redirect to login when accessing protected route", async ({
    page,
  }) => {
    await navigateAndWait(page, "/settings/organizations");
    await expect(page).toHaveURL(/.*login/);
  });

  test("should persist authentication across page reloads", async ({
    page,
    userPersona,
  }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await loginPage.login.login(
      userPersona.credentials.email,
      userPersona.credentials.password
    );
    await page.waitForURL((url) => !url.pathname.includes("/login"));
    // SPA navigations do not re-fire document load events, so wait for
    // dashboard UI before reloading (avoids Firefox NS_BINDING_ABORTED).
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible({
      timeout: 15_000,
    });

    // Prefer goto(current URL) over reload — more reliable on Firefox when
    // in-flight module loads would otherwise abort the navigation.
    await page.goto(page.url(), { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible({
      timeout: 15_000,
    });
    await expect(page).not.toHaveURL(/.*login/);
  });

  test("should handle logout", async ({ page, userPersona }) => {
    await loginUser(page, userPersona.credentials);
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible({
      timeout: 15_000,
    });

    await logoutUser(page, userPersona.user);
    await expect(page).toHaveURL(/.*login/);
  });
});
