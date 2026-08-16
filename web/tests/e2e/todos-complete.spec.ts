import { createTodo } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { TodoSheetSection } from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";
import type { Client } from "@/lib/client/client";

test.describe("@todos.complete Todo complete E2E Tests", () => {
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

  test("marks a todo complete and incomplete", async ({ page }) => {
    const title = `Complete ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    const item = todoSheet.getItem(title);

    await todoSheet.markComplete(title);
    await waitForSuccessToast(page, "Todo updated");
    await expect(todoSheet.getIncompleteCheckbox(title)).toBeChecked();

    await todoSheet.markIncomplete(title);
    await waitForSuccessToast(page, "Todo updated");
    await expect(todoSheet.getCompleteCheckbox(title)).not.toBeChecked();
    await item.hover();
    await expect(item.getByRole("button", { name: "Edit todo" })).toBeVisible();
  });

  test("hides Edit when the todo is completed", async ({ page }) => {
    const title = `Hide edit ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    const item = todoSheet.getItem(title);

    await todoSheet.markComplete(title);
    await waitForSuccessToast(page, "Todo updated");
    await item.hover();
    await expect(item.getByRole("button", { name: "Edit todo" })).toHaveCount(
      0
    );
  });
});
