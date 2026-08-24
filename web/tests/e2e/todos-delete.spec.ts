import { createTodo } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { TodoSheetSection } from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import type { Client } from "@/lib/api/client";
import type { User } from "@/lib/api/types";

test.describe("@todos.delete Todo delete E2E Tests", () => {
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

  test("deletes a todo and removes it from the list", async ({ page }) => {
    const title = `Delete ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    await expect(todoSheet.getItem(title)).toBeVisible();

    await todoSheet.clickDeleteItem(title);
    await waitForSuccessToast(page, "Todo deleted");
    await expect(todoSheet.getItem(title)).not.toBeVisible();
  });
});
