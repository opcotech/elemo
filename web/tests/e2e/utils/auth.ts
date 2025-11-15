import type { Page } from "@playwright/test";

import { navigateAndWait, waitForPageLoad } from "../helpers";

import type { LoginCredentials } from "@/lib/auth/types";

export const USER_DEFAULT_PASSWORD = "AppleTree123";
export const USER_DEFAULT_PASSWORD_HASH =
  "$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C";

/**
 * Helper function to perform login flow in e2e tests.
 *
 * @param page - Playwright page object
 * @param credentials - Login credentials
 * @param options - Optional configuration
 * @param options.destination - URL to navigate to after successful login (e.g., "/settings/organizations")
 * @param options.throwOnFailure - Whether to throw an error if login fails (default: true)
 * @returns Promise<boolean> - Returns true if login was successful, false otherwise
 */
export async function loginUser(
  page: Page,
  credentials: LoginCredentials,
  options?: {
    destination?: string;
    throwOnFailure?: boolean;
  }
): Promise<boolean> {
  const { destination, throwOnFailure = true } = options || {};
  await navigateAndWait(page, "/login");
  await page.getByLabel("Email").fill(credentials.email);
  await page
    .getByRole("textbox", { name: "Password" })
    .fill(credentials.password);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((url) => !url.pathname.includes("/login"));
  await waitForPageLoad(page);
  await page
    .waitForFunction(() => {
      const buttons = document.querySelectorAll("button");
      for (const button of buttons) {
        if (
          button.textContent &&
          button.textContent.includes("Signing in...")
        ) {
          return false;
        }
      }
      return true;
    })
    .catch(() => {});
  const isOnDashboard = await page
    .getByText("Welcome back!")
    .isVisible()
    .catch(() => false);

  if (!isOnDashboard) {
    if (throwOnFailure) {
      throw new Error("Login failed - dashboard not found");
    }
    return false;
  }
  if (destination) {
    await navigateAndWait(page, destination);
  }

  return true;
}
