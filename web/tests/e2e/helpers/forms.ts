import { expect } from "@playwright/test";
import { waitForPageLoad } from "./navigation";
import type { Page } from "@playwright/test";

/**
 * Fill a form field by its label.
 * @param page - Playwright page object
 * @param label - Label text of the form field
 * @param value - Value to fill in
 */
export async function fillFormField(
  page: Page,
  label: string,
  value: string
): Promise<void> {
  const field = page.getByLabel(label, { exact: true });
  await expect(field)
    .toBeVisible({ timeout: 10000 })
    .catch(() => {});
  await field.fill(value);
}

/**
 * Submit a form by clicking the submit button.
 * @param page - Playwright page object
 * @param buttonText - Text of the submit button
 */
export async function submitForm(
  page: Page,
  buttonText: string
): Promise<void> {
  const submitButton = page.getByRole("button", { name: buttonText });
  await submitButton.click();
}

/**
 * Wait for form submission to complete.
 * This waits for navigation or network activity to settle after form submission.
 * @param page - Playwright page object
 */
export async function waitForFormSubmission(page: Page): Promise<void> {
  await page
    .waitForLoadState("networkidle", { timeout: 10000 })
    .catch(() => {});
  await waitForPageLoad(page);
}
