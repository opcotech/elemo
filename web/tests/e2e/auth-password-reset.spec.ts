import { expect, test } from "./fixtures";
import { waitForErrorToast, waitForSuccessToast } from "./helpers";
import { ForgotPasswordPage, LoginPage, ResetPasswordPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser, logoutUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getPasswordResetTokenFromEmail, waitForEmail } from "./utils/mailpit";
import { getRandomString } from "./utils/random";

test.describe("@auth.password-reset Password Reset E2E Tests", () => {
  test("should navigate to forgot password from login", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login.waitForLoad();

    await page.getByRole("link", { name: "Forgot password?" }).click();
    await expect(page).toHaveURL(/\/forgot-password/);
    await expect(
      page.getByText("Forgot your password?", { exact: true })
    ).toBeVisible();
  });

  test("should request reset, set a new password via email link, and login", async ({
    page,
    testConfig,
  }) => {
    const testUser = await createUser(testConfig);
    const newPassword = `Reset${getRandomString(8)}1a`;

    const forgotPage = new ForgotPasswordPage(page);
    await forgotPage.goto();
    await forgotPage.form.waitForLoad();
    await forgotPage.form.requestReset(testUser.email);

    const email = await waitForEmail(testUser.email, 15000);
    expect(email).not.toBeNull();

    const token = await getPasswordResetTokenFromEmail(testUser.email);
    expect(token).toBeTruthy();

    const resetPage = new ResetPasswordPage(page);
    await resetPage.goto(token!);
    await resetPage.form.waitForLoad();
    await resetPage.form.resetPassword(newPassword);
    await waitForSuccessToast(page, "Password reset successfully");
    await expect(page).toHaveURL(/\/login/);

    await loginUser(page, {
      email: testUser.email,
      password: newPassword,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();

    await logoutUser(page, testUser);
    await expect(page).toHaveURL(/\/login/);

    const loginPage = new LoginPage(page);
    await loginPage.login.waitForLoad();
    await loginPage.login.login(testUser.email, USER_DEFAULT_PASSWORD);
    await waitForErrorToast(page, undefined);
    await expect(page).toHaveURL(/\/login/);
  });

  test("should show validation errors for invalid reset form inputs", async ({
    page,
    testConfig,
  }) => {
    const testUser = await createUser(testConfig);

    const forgotPage = new ForgotPasswordPage(page);
    await forgotPage.goto();
    await forgotPage.form.waitForLoad();
    await forgotPage.form.requestReset(testUser.email);

    const email = await waitForEmail(testUser.email, 15000);
    expect(email).not.toBeNull();
    const token = await getPasswordResetTokenFromEmail(testUser.email);
    expect(token).toBeTruthy();

    const resetPage = new ResetPasswordPage(page);
    await resetPage.goto(token!);
    await resetPage.form.waitForLoad();

    await resetPage.form.fillPasswordFields({
      password: "short",
      confirmPassword: "short",
    });
    await resetPage.form.submit();
    await expect(
      page.getByText(/Password must be at least 8 characters/i)
    ).toBeVisible();
    await expect(page).toHaveURL(/\/reset-password/);

    await resetPage.form.fillPasswordFields({
      password: "ValidPassword1",
      confirmPassword: "DifferentPassword1",
    });
    await resetPage.form.submit();
    await expect(page.getByText(/Passwords don't match/i)).toBeVisible();
    await expect(page).toHaveURL(/\/reset-password/);
  });
});
