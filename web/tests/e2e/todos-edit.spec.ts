import { createTodo } from "./api";
import { expect, test } from "./fixtures";
import { fillLocator, waitForSuccessToast } from "./helpers";
import { TodoFormSection, TodoSheetSection } from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { Client } from "@/lib/api/client";
import type { User } from "@/lib/api/types";

function localNoonIso(daysFromToday: number): string {
  const date = new Date();
  date.setHours(12, 0, 0, 0);
  date.setDate(date.getDate() + daysFromToday);
  return date.toISOString();
}

function localNoonDate(daysFromToday: number): Date {
  const date = new Date();
  date.setHours(12, 0, 0, 0);
  date.setDate(date.getDate() + daysFromToday);
  return date;
}

test.describe("@todos.edit Todo edit E2E Tests", () => {
  // Fresh user per test so earlier creates do not push completed items
  // below the sheet ScrollArea viewport.
  let testUser: User;
  let apiClient: Client;

  test.beforeEach(async ({ page, testConfig, createApiClient }) => {
    testUser = await createUser(testConfig);
    apiClient = await createApiClient(testUser.email, USER_DEFAULT_PASSWORD);
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  test("updates the title", async ({ page }) => {
    const title = `Update title ${getRandomString(8)}`;
    const updatedTitle = `Updated ${title}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();
    await expect(todoSheet.getItem(title)).toBeVisible();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await fillLocator(editDialog.getByLabel("Title"), updatedTitle);
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    await expect(
      todoSheet.getSheet().getByText(updatedTitle, { exact: true })
    ).toBeVisible();
    await expect(
      todoSheet.getSheet().getByText(title, { exact: true })
    ).not.toBeVisible();
  });

  test("updates the description", async ({ page }) => {
    const title = `Update desc ${getRandomString(8)}`;
    const description = `Original description ${getRandomString(8)}`;
    const updatedDescription = `Updated description ${getRandomString(8)}`;
    await createTodo(apiClient, {
      title,
      description,
      owned_by: testUser.id,
    });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await todoForm.fillDescription(editDialog, updatedDescription);
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    const item = todoSheet.getItem(title);
    await expect(item.getByText(updatedDescription)).toBeVisible();
    await expect(item.getByText(description)).not.toBeVisible();
  });

  test("updates the priority", async ({ page }) => {
    const title = `Update priority ${getRandomString(8)}`;
    await createTodo(apiClient, {
      title,
      priority: "normal",
      owned_by: testUser.id,
    });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await editDialog.getByRole("combobox").click();
    await page.getByRole("option", { name: "Important" }).click();
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    await expect(
      todoSheet.getItem(title).getByText("Important", { exact: true })
    ).toBeVisible();
  });

  test("updates title, description, and priority together", async ({
    page,
  }) => {
    const title = `Update combo ${getRandomString(8)}`;
    const updatedTitle = `Combo ${getRandomString(8)}`;
    const updatedDescription = `Combo description ${getRandomString(8)}`;
    await createTodo(apiClient, {
      title,
      description: "Initial description text",
      priority: "normal",
      owned_by: testUser.id,
    });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await todoForm.fillTodoFields(editDialog, {
      title: updatedTitle,
      description: updatedDescription,
      priority: "Urgent",
    });
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    const item = todoSheet.getItem(updatedTitle);
    await expect(item.getByText(updatedTitle, { exact: true })).toBeVisible();
    await expect(item.getByText(updatedDescription)).toBeVisible();
    await expect(item.getByText("Urgent", { exact: true })).toBeVisible();
  });

  test("sets a due date and shows the todo in Today", async ({ page }) => {
    const title = `Due today ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();
    await expect(todoSheet.getGroup("Later")).toBeVisible();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await todoForm.selectDueDate(editDialog, localNoonDate(0));
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    await expect(todoSheet.getGroup("Today")).toBeVisible();
    await expect(
      todoSheet
        .getGroup("Today")
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();
  });

  test("changes the due date and moves the item to Tomorrow", async ({
    page,
  }) => {
    const title = `Due tomorrow ${getRandomString(8)}`;
    await createTodo(apiClient, {
      title,
      owned_by: testUser.id,
      due_date: localNoonIso(0),
    });

    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);
    await todoSheet.openFromHeader();
    await expect(
      todoSheet
        .getGroup("Today")
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();

    await todoSheet.clickEditItem(title);
    const editDialog = await todoForm.waitForEditDialog();
    await todoForm.selectDueDate(editDialog, localNoonDate(1));
    await todoForm.submitEdit();
    await waitForSuccessToast(page, "Todo updated successfully");

    await expect(todoSheet.getGroup("Tomorrow")).toBeVisible();
    await expect(
      todoSheet
        .getGroup("Tomorrow")
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();
  });

  test("shows a todo without a due date in Later", async ({ page }) => {
    const title = `No due date ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    await expect(todoSheet.getGroup("Later")).toBeVisible();
    await expect(
      todoSheet
        .getGroup("Later")
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();
  });

  test("shows a todo due after this week in Later", async ({ page }) => {
    const title = `Due later ${getRandomString(8)}`;
    await createTodo(apiClient, {
      title,
      owned_by: testUser.id,
      due_date: localNoonIso(14),
    });

    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    await expect(todoSheet.getGroup("Later")).toBeVisible();
    await expect(
      todoSheet
        .getGroup("Later")
        .getByRole("listitem")
        .filter({ hasText: title })
    ).toBeVisible();
  });
});
