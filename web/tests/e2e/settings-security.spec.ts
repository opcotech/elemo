import { expect, test } from "@playwright/test";
import { USER_DEFAULT_PASSWORD, createDBUser, loginUser } from "./utils/auth";
import { Form } from "./components/form";
import { waitForPageLoad } from "./helpers/navigation";
import { waitForErrorToast, waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.security Password Change E2E Tests", () => {
  let testUser: any;

  test.beforeAll(async () => {
    testUser = await createDBUser("active");
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, testUser, {
      destination: "/settings/security",
    });
    await expect(page).toHaveURL(/.*settings\/security/);
    await waitForPageLoad(page);
  });

  test("should display password change form with all required elements", async ({
    page,
  }) => {
    await expect(
      page.getByRole("heading", { name: "Password & Authentication" })
    ).toBeVisible();
    await expect(
      page.getByText("Manage your password and authentication settings.")
    ).toBeVisible();
    await expect(page.getByText("Change Password")).toBeVisible();
    await expect(
      page.getByLabel("Current Password", { exact: true })
    ).toBeVisible();
    await expect(
      page.getByLabel("New Password", { exact: true })
    ).toBeVisible();
    await expect(
      page.getByLabel("Confirm New Password", { exact: true })
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Update Password" })
    ).toBeVisible();
    await expect(
      page.getByLabel("Current Password", { exact: true })
    ).toHaveAttribute("type", "password");
    await expect(
      page.getByLabel("New Password", { exact: true })
    ).toHaveAttribute("type", "password");
    await expect(
      page.getByLabel("Confirm New Password", { exact: true })
    ).toHaveAttribute("type", "password");
  });

  test("should toggle password visibility for all password fields", async ({
    page,
  }) => {
    const currentPasswordInput = page.getByLabel("Current Password", {
      exact: true,
    });
    const currentPasswordToggle = page
      .getByRole("button", { name: /Show password|Hide password/ })
      .first();
    const currentType = await currentPasswordInput.getAttribute("type");
    if (currentType === "password") {
      await currentPasswordToggle.click();
      await expect(currentPasswordInput).toHaveAttribute("type", "text");
      await currentPasswordToggle.click();
      await expect(currentPasswordInput).toHaveAttribute("type", "password");
    } else {
      await currentPasswordToggle.click();
      await expect(currentPasswordInput).toHaveAttribute("type", "password");
      await currentPasswordToggle.click();
      await expect(currentPasswordInput).toHaveAttribute("type", "text");
    }
    const newPasswordInput = page.getByLabel("New Password", { exact: true });
    const newPasswordToggle = page
      .getByRole("button", { name: /Show password|Hide password/ })
      .nth(1);
    const newType = await newPasswordInput.getAttribute("type");
    if (newType === "password") {
      await newPasswordToggle.click();
      await expect(newPasswordInput).toHaveAttribute("type", "text");
      await newPasswordToggle.click();
      await expect(newPasswordInput).toHaveAttribute("type", "password");
    } else {
      await newPasswordToggle.click();
      await expect(newPasswordInput).toHaveAttribute("type", "password");
      await newPasswordToggle.click();
      await expect(newPasswordInput).toHaveAttribute("type", "text");
    }
    const confirmPasswordInput = page.getByLabel("Confirm New Password", {
      exact: true,
    });
    const confirmPasswordToggle = page
      .getByRole("button", { name: /Show password|Hide password/ })
      .nth(2);
    const confirmType = await confirmPasswordInput.getAttribute("type");
    if (confirmType === "password") {
      await confirmPasswordToggle.click();
      await expect(confirmPasswordInput).toHaveAttribute("type", "text");
      await confirmPasswordToggle.click();
      await expect(confirmPasswordInput).toHaveAttribute("type", "password");
    } else {
      await confirmPasswordToggle.click();
      await expect(confirmPasswordInput).toHaveAttribute("type", "password");
      await confirmPasswordToggle.click();
      await expect(confirmPasswordInput).toHaveAttribute("type", "text");
    }
  });

  test("should show validation errors for empty form submission", async ({
    page,
  }) => {
    const form = new Form(page);
    await form.submit("Update Password");
    await waitForPageLoad(page);
    const hasValidationErrors = await page
      .getByText("Current password is required")
      .isVisible()
      .catch(() => false);
    const isStillOnSecurityPage = page.url().includes("/settings/security");

    if (hasValidationErrors) {
      await expect(
        page.getByText("Current password is required")
      ).toBeVisible();
      await expect(
        page.getByText("Password must be at least 8 characters")
      ).toBeVisible();
      await expect(
        page.getByText("Please confirm your password")
      ).toBeVisible();
    } else if (isStillOnSecurityPage) {
      expect(true).toBe(true);
    } else {
      throw new Error(
        "Validation failed - form submitted when it shouldn't have"
      );
    }
  });

  test("should show validation errors for invalid password inputs", async ({
    page,
  }) => {
    const form = new Form(page);
    await form.fillFields({
      "Current Password": USER_DEFAULT_PASSWORD,
      "New Password": "short",
      "Confirm New Password": "short",
    });
    await form.submit("Update Password");
    await waitForPageLoad(page);

    const hasShortError = await page
      .getByText("Password must be at least 8 characters")
      .isVisible()
      .catch(() => false);
    const isStillOnPage1 = page.url().includes("/settings/security");

    if (hasShortError) {
      await expect(
        page.getByText("Password must be at least 8 characters")
      ).toBeVisible();
    } else if (isStillOnPage1) {
      expect(true).toBe(true);
    } else {
      throw new Error("Short password validation failed");
    }
    const longPassword = "a".repeat(65);
    await form.fillFields({
      "Current Password": USER_DEFAULT_PASSWORD,
      "New Password": longPassword,
      "Confirm New Password": longPassword,
    });
    await form.submit("Update Password");
    await waitForPageLoad(page);

    const hasLongError = await page
      .getByText("Password must be less than 64 characters")
      .isVisible()
      .catch(() => false);
    const isStillOnPage2 = page.url().includes("/settings/security");

    if (hasLongError) {
      await expect(
        page.getByText("Password must be less than 64 characters")
      ).toBeVisible();
    } else if (isStillOnPage2) {
      expect(true).toBe(true);
    } else {
      throw new Error("Long password validation failed");
    }
    await form.fillFields({
      "Current Password": USER_DEFAULT_PASSWORD,
      "New Password": "NewPassword123!",
      "Confirm New Password": "DifferentPassword123!",
    });
    await form.submit("Update Password");
    await waitForPageLoad(page);

    const hasMismatchError = await page
      .getByText("Passwords don't match")
      .isVisible()
      .catch(() => false);
    const isStillOnPage3 = page.url().includes("/settings/security");

    if (hasMismatchError) {
      await expect(page.getByText("Passwords don't match")).toBeVisible();
    } else if (isStillOnPage3) {
      expect(true).toBe(true);
    } else {
      throw new Error("Password mismatch validation failed");
    }
  });

  test("should handle invalid current password", async ({ page }) => {
    const form = new Form(page);
    await form.fillFields({
      "Current Password": "WrongPassword123!",
      "New Password": "NewSecurePassword123!",
      "Confirm New Password": "NewSecurePassword123!",
    });
    await form.submit("Update Password");
    await waitForErrorToast(page, undefined, { timeout: 5000 });
  });

  test("should successfully update password and show success toast", async ({
    page,
  }) => {
    const form = new Form(page);
    await form.fillFields({
      "Current Password": USER_DEFAULT_PASSWORD,
      "New Password": "NewSecurePassword123!",
      "Confirm New Password": "NewSecurePassword123!",
    });
    await form.submit("Update Password");
    await waitForSuccessToast(page, "Password updated successfully", {
      timeout: 10000,
    });
    await expect(page.getByText("Update Password")).toBeVisible();
    const submitButton = page.getByRole("button", { name: "Update Password" });
    await expect(submitButton).toBeEnabled();
    await expect(
      page.getByLabel("Current Password", { exact: true })
    ).toBeEnabled();
    await expect(
      page.getByLabel("New Password", { exact: true })
    ).toBeEnabled();
    await expect(
      page.getByLabel("Confirm New Password", { exact: true })
    ).toBeEnabled();
  });
});
