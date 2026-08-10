import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

import { fillFormField, navigateAndWait, waitForPageLoad } from "../helpers";

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
  // Clear prior session so tests can switch users without hitting
  // redirectIfAuthenticated on /login. Wait for any in-flight redirect
  // (e.g. post password-reset) to settle before navigating again.
  await page.context().clearCookies();
  await page.waitForLoadState("domcontentloaded").catch(() => undefined);
  await navigateAndWait(page, "/login");
  const signIn = page.getByRole("button", { name: "Sign in" });
  await expect(signIn).toBeVisible();
  await expect(signIn).toBeEnabled();
  await fillFormField(page, "Email", credentials.email);
  await fillFormField(page, "Password", credentials.password);
  await signIn.click();

  await page.waitForURL((url) => !url.pathname.includes("/login"), {
    timeout: 15_000,
  });
  await waitForPageLoad(page);

  // Wait for the authenticated shell instead of a single visibility snapshot —
  // WebKit and slow home loaders often leave /login before the sidebar paints.
  const shell = page.getByRole("link", { name: "Elemo", exact: true });
  try {
    await shell.waitFor({ state: "visible", timeout: 15_000 });
  } catch {
    if (throwOnFailure) {
      throw new Error("Login failed - authenticated shell not found");
    }
    return false;
  }

  if (destination) {
    await navigateAndWait(page, destination);
  }

  return true;
}

/**
 * Open the sidebar user menu and log out.
 * Matches NavUser by email (always shown) or optional display name.
 */
export async function logoutUser(
  page: Page,
  user: { email: string; first_name?: string; last_name?: string }
): Promise<void> {
  const userMenuTrigger = page
    .locator("button")
    .filter({ hasText: user.email })
    .first();

  await userMenuTrigger.click();
  await page.getByRole("menuitem", { name: "Log out" }).click();
  await page.waitForURL((url) => url.pathname.includes("/login"));
}
