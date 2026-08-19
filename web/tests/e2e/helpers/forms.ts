import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

/**
 * Fill a locator using trusted key events so Base UI + React controlled
 * inputs stay in sync (Playwright fill() often times out on Firefox/WebKit).
 */
export async function fillLocator(
  field: Locator,
  value: string
): Promise<void> {
  await expect(field).toBeVisible();
  await expect(field).toBeEditable();

  // Prefer a single fill() so URL-synced / remounting inputs keep one value.
  // Fall back to key events for Base UI + Firefox/WebKit, and when fill()
  // updates the DOM without committing React Hook Form state.
  await field.click();
  await field.fill(value);
  await field.blur();
  try {
    await expect(field).toHaveValue(value, { timeout: 1_000 });
    return;
  } catch {
    // continue with sequential typing
  }

  for (let attempt = 0; attempt < 2; attempt++) {
    await field.click();
    await field.press("ControlOrMeta+A");
    await field.press("Backspace");
    if (value.length > 0) {
      await field.pressSequentially(value, { delay: 10 });
    }
    try {
      await expect(field).toHaveValue(value, { timeout: 1_000 });
      return;
    } catch {
      if (attempt === 1) {
        throw new Error(
          "Failed to fill locator with a stable value after retry"
        );
      }
    }
  }
}

/**
 * Clear a form field by its label.
 * @param page - Playwright page object
 * @param label - Label text of the form field
 */
export async function clearFormField(page: Page, label: string): Promise<void> {
  const field = page.getByLabel(label, { exact: true });
  await fillLocator(field, "");
}

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
  await fillLocator(page.getByLabel(label, { exact: true }), value);
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
 * Returns the field error locator associated with a specific field label.
 * Useful for asserting validation errors scoped to a particular input.
 * @param page - Playwright page object
 * @param label - Label text of the form field
 */
export function getFormFieldMessage(page: Page, label: string) {
  return page
    .locator("[data-slot='field']")
    .filter({ has: page.getByLabel(label, { exact: true }) })
    .locator("[data-slot='field-error']");
}
