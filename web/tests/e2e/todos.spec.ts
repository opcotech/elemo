import type { Locator, Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { waitForAnimations, waitForSuccessToast } from "./helpers";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

async function openCommandPalette(page: Page) {
  await page.keyboard.press("Control+K");
  const commandDialog = page.getByRole("dialog", {
    name: "Search or run a command",
  });
  await expect(commandDialog).toBeVisible();
  await waitForAnimations(commandDialog);
  return commandDialog;
}

async function runPaletteCommand(page: Page, commandTitle: string) {
  const commandDialog = await openCommandPalette(page);
  await commandDialog
    .getByPlaceholder("Search entities, navigation, or commands...")
    .fill(commandTitle);
  await commandDialog
    .locator('[data-slot="command-item"]')
    .filter({ hasText: commandTitle })
    .first()
    .click();
  await expect(commandDialog).not.toBeVisible();
}

async function pressPaletteShortcut(page: Page, secondKey: "s" | "n") {
  // Sequence shortcuts are handled only while the command palette is open.
  await openCommandPalette(page);
  await page.keyboard.down("Shift");
  await page.keyboard.press("t");
  await page.keyboard.press(secondKey);
  await page.keyboard.up("Shift");
}

function todoSheet(page: Page) {
  return page.getByRole("dialog", { name: "Todo Items" });
}

function todoItem(sheet: Locator, title: string) {
  return sheet.getByRole("listitem").filter({ hasText: title });
}

function todoGroup(sheet: Locator, label: string) {
  return sheet.getByRole("list", { name: `${label} todos` });
}

async function openTodoSheetFromHeader(page: Page) {
  await page.getByRole("button", { name: "Show todo list" }).click();
  const sheet = todoSheet(page);
  await expect(sheet).toBeVisible();
  await waitForAnimations(sheet);
  return sheet;
}

async function closeTodoSheet(page: Page) {
  const sheet = todoSheet(page);
  await page.keyboard.press("Escape");
  await expect(sheet).not.toBeVisible();
}

async function openAddTodoFromSheet(sheet: Locator) {
  await sheet.getByRole("button", { name: "Add Todo" }).first().click();
}

async function waitForPageAnimations(page: Page) {
  await page.evaluate(async () => {
    await Promise.all(
      document
        .getAnimations()
        .map((animation) => animation.finished.catch(() => undefined))
    );
  });
}

async function waitForAddTodoDialog(
  page: Page,
  options?: { allowHidden?: boolean }
) {
  if (options?.allowHidden) {
    // Sheet + Add Todo opened together: wait for enter animations, then accept
    // the dialog even if the sheet modal left it aria-hidden/inert.
    await waitForPageAnimations(page);
  }
  const addDialog = page.getByRole("dialog", {
    name: "Add Todo",
    includeHidden: options?.allowHidden,
  });
  await expect(addDialog).toBeVisible();
  await waitForAnimations(addDialog);
  return addDialog;
}

async function fillAndSubmitTodo(
  page: Page,
  fields: {
    title: string;
    description?: string;
    priority?: "Normal" | "Important" | "Urgent" | "Critical";
  },
  options?: { allowHidden?: boolean }
) {
  const allowHidden = options?.allowHidden ?? false;
  const addDialog = await waitForAddTodoDialog(page, { allowHidden });
  await addDialog.getByLabel("Title").fill(fields.title);
  if (fields.description !== undefined) {
    await addDialog.getByLabel("Description").fill(fields.description);
  }
  if (fields.priority) {
    await addDialog
      .getByRole("combobox")
      .click(allowHidden ? { force: true } : undefined);
    await page.getByRole("option", { name: fields.priority }).click();
  }
  await addDialog
    .getByRole("button", {
      name: "Add todo",
      includeHidden: allowHidden || undefined,
    })
    .click(allowHidden ? { force: true } : undefined);
  await waitForSuccessToast(page, "Todo added successfully");
  await expect(addDialog).not.toBeVisible();
}

async function createTodoViaSheet(
  page: Page,
  fields: {
    title: string;
    description?: string;
    priority?: "Normal" | "Important" | "Urgent" | "Critical";
  }
) {
  const sheet = await openTodoSheetFromHeader(page);
  await openAddTodoFromSheet(sheet);
  await fillAndSubmitTodo(page, fields);
  return sheet;
}

test.describe("@todos Todo E2E Tests", () => {
  // Fresh user per test so earlier creates do not push completed items
  // below the sheet ScrollArea viewport.
  test.beforeEach(async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  test.describe("Opening and closing the todo list", () => {
    test("opens and closes the todo list from the header", async ({ page }) => {
      const sheet = await openTodoSheetFromHeader(page);
      await expect(
        sheet.getByText(/No todos found|open ·|completed/)
      ).toBeVisible();

      await sheet.getByRole("button", { name: "Close" }).click();
      await expect(sheet).not.toBeVisible();
    });

    test("opens the todo list from the command palette", async ({ page }) => {
      await runPaletteCommand(page, "Show Todos");
      const sheet = todoSheet(page);
      await expect(sheet).toBeVisible();
      await waitForAnimations(sheet);
    });

    test("opens the todo list with Shift+T+S while the palette is open", async ({
      page,
    }) => {
      // Shortcut sequences are handled only while the command palette is open.
      await pressPaletteShortcut(page, "s");
      const sheet = todoSheet(page);
      await expect(sheet).toBeVisible();
      await waitForAnimations(sheet);
    });

    test("closes the todo list with Escape", async ({ page }) => {
      await openTodoSheetFromHeader(page);
      await closeTodoSheet(page);
    });
  });

  test.describe("Creating todo items", () => {
    test("creates a todo from the Add Todo button in the sheet", async ({
      page,
    }) => {
      const title = `Sheet create ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title });
      await expect(todoGroup(sheet, "Later")).toBeVisible();
      await expect(todoItem(sheet, title)).toBeVisible();
    });

    test("creates a todo from the command palette Add Todo action", async ({
      page,
    }) => {
      const title = `Palette create ${getRandomString(8)}`;
      await runPaletteCommand(page, "Add Todo");
      await fillAndSubmitTodo(page, { title }, { allowHidden: true });
      await expect(todoSheet(page)).toBeVisible();
      await expect(todoItem(todoSheet(page), title)).toBeVisible();
    });

    test("creates a todo with Shift+T+N while the palette is open", async ({
      page,
    }) => {
      const title = `Shortcut create ${getRandomString(8)}`;
      await pressPaletteShortcut(page, "n");
      await fillAndSubmitTodo(page, { title }, { allowHidden: true });
      await expect(todoSheet(page)).toBeVisible();
      await expect(todoItem(todoSheet(page), title)).toBeVisible();
    });
  });

  test.describe("Creating items with various fields", () => {
    test("creates with title only", async ({ page }) => {
      const title = `Title only ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title });
      const item = todoItem(sheet, title);
      await expect(item.getByText(title, { exact: true })).toBeVisible();
      await expect(item.getByText("normal", { exact: true })).toBeVisible();
    });

    test("creates with title and description", async ({ page }) => {
      const title = `Title desc ${getRandomString(8)}`;
      const description = `Description ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title, description });
      const item = todoItem(sheet, title);
      await expect(item.getByText(title, { exact: true })).toBeVisible();
      await expect(item.getByText(description)).toBeVisible();
    });

    test("creates with title and priority", async ({ page }) => {
      const title = `Title priority ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, {
        title,
        priority: "Urgent",
      });
      const item = todoItem(sheet, title);
      await expect(item.getByText(title, { exact: true })).toBeVisible();
      await expect(item.getByText("urgent", { exact: true })).toBeVisible();
    });

    test("creates with title, description, and priority", async ({ page }) => {
      const title = `Full fields ${getRandomString(8)}`;
      const description = `Full description ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, {
        title,
        description,
        priority: "Critical",
      });
      const item = todoItem(sheet, title);
      await expect(item.getByText(title, { exact: true })).toBeVisible();
      await expect(item.getByText(description)).toBeVisible();
      await expect(item.getByText("critical", { exact: true })).toBeVisible();
    });
  });

  test.describe("Updating todo items", () => {
    test("updates the title", async ({ page }) => {
      const title = `Update title ${getRandomString(8)}`;
      const updatedTitle = `Updated ${title}`;
      const sheet = await createTodoViaSheet(page, { title });

      await todoItem(sheet, title)
        .getByRole("button", { name: "Edit todo" })
        .click();
      const editDialog = page.getByRole("dialog", { name: "Edit Todo" });
      await expect(editDialog).toBeVisible();
      await waitForAnimations(editDialog);
      await editDialog.getByLabel("Title").fill(updatedTitle);
      await editDialog.getByRole("button", { name: "Update todo" }).click();
      await waitForSuccessToast(page, "Todo updated successfully");

      await expect(
        sheet.getByText(updatedTitle, { exact: true })
      ).toBeVisible();
      await expect(sheet.getByText(title, { exact: true })).not.toBeVisible();
    });

    test("updates the description", async ({ page }) => {
      const title = `Update desc ${getRandomString(8)}`;
      const description = `Original description ${getRandomString(8)}`;
      const updatedDescription = `Updated description ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title, description });

      await todoItem(sheet, title)
        .getByRole("button", { name: "Edit todo" })
        .click();
      const editDialog = page.getByRole("dialog", { name: "Edit Todo" });
      await editDialog.getByLabel("Description").fill(updatedDescription);
      await editDialog.getByRole("button", { name: "Update todo" }).click();
      await waitForSuccessToast(page, "Todo updated successfully");

      const item = todoItem(sheet, title);
      await expect(item.getByText(updatedDescription)).toBeVisible();
      await expect(item.getByText(description)).not.toBeVisible();
    });

    test("updates the priority", async ({ page }) => {
      const title = `Update priority ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, {
        title,
        priority: "Normal",
      });

      await todoItem(sheet, title)
        .getByRole("button", { name: "Edit todo" })
        .click();
      const editDialog = page.getByRole("dialog", { name: "Edit Todo" });
      await editDialog.getByRole("combobox").click();
      await page.getByRole("option", { name: "Important" }).click();
      await editDialog.getByRole("button", { name: "Update todo" }).click();
      await waitForSuccessToast(page, "Todo updated successfully");

      await expect(
        todoItem(sheet, title).getByText("important", { exact: true })
      ).toBeVisible();
    });

    test("updates title, description, and priority together", async ({
      page,
    }) => {
      const title = `Update combo ${getRandomString(8)}`;
      const updatedTitle = `Combo ${getRandomString(8)}`;
      const updatedDescription = `Combo description ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, {
        title,
        description: "Initial description text",
        priority: "Normal",
      });

      await todoItem(sheet, title)
        .getByRole("button", { name: "Edit todo" })
        .click();
      const editDialog = page.getByRole("dialog", { name: "Edit Todo" });
      await editDialog.getByLabel("Title").fill(updatedTitle);
      await editDialog.getByLabel("Description").fill(updatedDescription);
      await editDialog.getByRole("combobox").click();
      await page.getByRole("option", { name: "Urgent" }).click();
      await editDialog.getByRole("button", { name: "Update todo" }).click();
      await waitForSuccessToast(page, "Todo updated successfully");

      const item = todoItem(sheet, updatedTitle);
      await expect(item.getByText(updatedTitle, { exact: true })).toBeVisible();
      await expect(item.getByText(updatedDescription)).toBeVisible();
      await expect(item.getByText("urgent", { exact: true })).toBeVisible();
    });
  });

  test.describe("Completing and uncompleting items", () => {
    test("marks a todo complete and incomplete", async ({ page }) => {
      const title = `Complete ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title });
      const item = todoItem(sheet, title);

      await item
        .getByRole("checkbox", { name: `Mark "${title}" as complete` })
        .click();
      await waitForSuccessToast(page, "Todo updated");
      await expect(
        item.getByRole("checkbox", { name: `Mark "${title}" as incomplete` })
      ).toBeChecked();

      await item
        .getByRole("checkbox", { name: `Mark "${title}" as incomplete` })
        .click();
      await waitForSuccessToast(page, "Todo updated");
      await expect(
        item.getByRole("checkbox", { name: `Mark "${title}" as complete` })
      ).not.toBeChecked();
      await expect(
        item.getByRole("button", { name: "Edit todo" })
      ).toBeVisible();
    });
  });

  test.describe("Deleting items", () => {
    test("deletes a todo and removes it from the list", async ({ page }) => {
      const title = `Delete ${getRandomString(8)}`;
      const sheet = await createTodoViaSheet(page, { title });
      await expect(todoItem(sheet, title)).toBeVisible();

      await todoItem(sheet, title)
        .getByRole("button", { name: "Delete todo" })
        .click();
      await waitForSuccessToast(page, "Todo deleted");
      await expect(todoItem(sheet, title)).not.toBeVisible();
    });
  });
});
