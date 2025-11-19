import type { Page } from "@playwright/test";

export async function waitForTodoSheet(page: Page, timeout = 5000): Promise<void> {
  await page.getByRole("heading", { name: "Todo Items" }).waitFor({ timeout });
}

export async function openTodoListViaUI(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Show todo list" }).click();
  await waitForTodoSheet(page);
}

export async function openTodoListViaKeyboard(page: Page): Promise<void> {
  await page.keyboard.press("Shift+T");
  await page.keyboard.press("S");
  await waitForTodoSheet(page);
}

export async function openTodoListViaCommandPalette(page: Page): Promise<void> {
  const isMac = process.platform === "darwin";
  const modifier = isMac ? "Meta" : "Control";
  
  await page.keyboard.press(`${modifier}+K`);
  await page.getByPlaceholder(/Type a command/i).waitFor();
  await page.getByRole("option", { name: /Show Todos/i }).click();
  await waitForTodoSheet(page);
}

export async function openAddTodoFormViaKeyboard(page: Page): Promise<void> {
  // Shift + T + N
  await page.keyboard.press("Shift+T");
  await page.keyboard.press("N");
  await page.getByRole("heading", { name: "Add Todo" }).waitFor();
}

export async function openAddTodoFormViaCommandPalette(page: Page): Promise<void> {
  const isMac = process.platform === "darwin";
  const modifier = isMac ? "Meta" : "Control";
  
  await page.keyboard.press(`${modifier}+K`);
  await page.getByPlaceholder(/Type a command/i).waitFor();
  await page.getByRole("option", { name: /Add Todo/i }).click();
  await page.getByRole("heading", { name: "Add Todo" }).waitFor();
}