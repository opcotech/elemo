import { expect, test } from "./fixtures";
import { waitForErrorToast, waitForSuccessToast } from "./helpers";
import { SettingsSecurityPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";

test.describe("@settings.security Password Change E2E Tests", () => {
  test.beforeEach(async ({ page, userPersona }) => {
    await loginUser(page, userPersona.credentials);
  });

  test("should show validation errors for invalid form inputs", async ({
    page,
  }) => {
    const securityPage = new SettingsSecurityPage(page);
    await securityPage.goto();
    await securityPage.security.waitForLoad();

    // Empty form
    await securityPage.security.submitPasswordChange();
    await expect(page.getByText("Current password is required")).toBeVisible();

    // Password too short
    await securityPage.security.fillPasswordFields({
      currentPassword: USER_DEFAULT_PASSWORD,
      newPassword: "short",
      confirmPassword: "short",
    });
    await securityPage.security.submitPasswordChange();
    await expect(
      page.getByText("Password must be at least 8 characters")
    ).toBeVisible();

    // Password too long
    const longPassword = "a".repeat(65);
    await securityPage.security.fillPasswordFields({
      currentPassword: USER_DEFAULT_PASSWORD,
      newPassword: longPassword,
      confirmPassword: longPassword,
    });
    await securityPage.security.submitPasswordChange();
    await expect(
      page.getByText("Password must be less than 64 characters")
    ).toBeVisible();

    // Passwords don't match
    await securityPage.security.fillPasswordFields({
      currentPassword: USER_DEFAULT_PASSWORD,
      newPassword: "NewPassword123!",
      confirmPassword: "DifferentPassword123!",
    });
    await securityPage.security.submitPasswordChange();
    await expect(page.getByText("Passwords don't match")).toBeVisible();
  });

  test("should show error when current password is incorrect", async ({
    page,
  }) => {
    const securityPage = new SettingsSecurityPage(page);
    await securityPage.goto();
    await securityPage.security.waitForLoad();

    await securityPage.security.fillPasswordFields({
      currentPassword: "WrongPassword123!",
      newPassword: "NewSecurePassword123!",
      confirmPassword: "NewSecurePassword123!",
    });
    await securityPage.security.submitPasswordChange();
    await waitForErrorToast(page);
  });

  test("should successfully update password", async ({ page }) => {
    const securityPage = new SettingsSecurityPage(page);
    await securityPage.goto();
    await securityPage.security.waitForLoad();

    await securityPage.security.fillPasswordFields({
      currentPassword: USER_DEFAULT_PASSWORD,
      newPassword: "NewSecurePassword123!",
      confirmPassword: "NewSecurePassword123!",
    });
    await securityPage.security.submitPasswordChange();
    await waitForSuccessToast(page, "Password updated successfully");
  });
});
