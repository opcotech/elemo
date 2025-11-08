import { navigateAndWait, waitForPageLoad } from "../helpers/navigation";
import { getSession } from "./db";
import { generateXid } from "./xid";
import { getRandomString } from "./random";
import type { Page } from "@playwright/test";
import type { User, UserStatus } from "@/lib/api";

export const USER_DEFAULT_PASSWORD = "AppleTree123";
export const USER_DEFAULT_PASSWORD_HASH =
  "$2a$10$CYRw/WtES8e8d4di2uIddevV9MO2.tI0G8R8QZEnF5dyx8S4Wqt6e";
export async function createDBUser(
  status: UserStatus = "active",
  overrides?: Partial<User>
): Promise<User> {
  const user: User = {
    id: await generateXid(),
    username: getRandomString(8).toLowerCase(),
    first_name: getRandomString(),
    last_name: getRandomString(),
    email: `${getRandomString()}-test@example.com`.toLowerCase(),
    picture: "https://picsum.photos/id/64/100/100",
    title: "Senior Test User",
    bio: "I am a senior test user",
    address: "123 Main St, Anytown, USA",
    phone: "555-555-5555",
    links: ["https://example.com"],
    languages: ["en"],
    status: status,
    created_at: new Date().toISOString(),
    updated_at: null,
    ...overrides,
  };

  try {
    const resp = await getSession().executeWrite((tx: any) => {
      const query = `
      CREATE (u:User $user)
      WITH u SET u.password = $password
      RETURN u
    `;

      return tx.run(query, { user, password: USER_DEFAULT_PASSWORD_HASH });
    });

    return resp.records[0].get("u").properties;
  } finally {
    await getSession().close();
  }
}

/**
 * Helper function to perform login flow in e2e tests.
 *
 * @param page - Playwright page object
 * @param user - User object with email property
 * @param options - Optional configuration
 * @param options.destination - URL to navigate to after successful login (e.g., "/settings/organizations")
 * @param options.throwOnFailure - Whether to throw an error if login fails (default: true)
 * @returns Promise<boolean> - Returns true if login was successful, false otherwise
 */
export async function loginUser(
  page: Page,
  user: { email: string },
  options?: {
    destination?: string;
    throwOnFailure?: boolean;
  }
): Promise<boolean> {
  const { destination, throwOnFailure = true } = options || {};
  await navigateAndWait(page, "/login");
  await page.getByLabel("Email").fill(user.email);
  await page
    .getByRole("textbox", { name: "Password" })
    .fill(USER_DEFAULT_PASSWORD);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((url) => !url.pathname.includes("/login"), {
    timeout: 10000,
  });
  await waitForPageLoad(page);
  await page
    .waitForFunction(
      () => {
        const buttons = document.querySelectorAll("button");
        for (const button of buttons) {
          if (button.textContent.includes("Signing in...")) {
            return false;
          }
        }
        return true;
      },
      { timeout: 10000 }
    )
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
