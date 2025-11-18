import { expect, test } from "./fixtures";
import {openTodoListViaCommandPalette, openTodoListViaKeyboard, openTodoListViaUI, } from "./helpers/todo";
import { TodoSection } from "./sections/todo-section";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";

test.describe("@todo.open-close TODO List Open/Close E2E Tests", () => {
  test("should open TODO list via UI button click", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    
    expect(await todoSection.isOpen()).toBe(true);
  });

  test("should close TODO list via Escape key", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.close();
    
    expect(await todoSection.isOpen()).toBe(false);
  });

  test("should open TODO list via keyboard shortcut (Shift+T+S)", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    
    await openTodoListViaKeyboard(page);
    await todoSection.waitForLoad();
    
    expect(await todoSection.isOpen()).toBe(true);
  });

  test("should open TODO list via command palette", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    
    await openTodoListViaCommandPalette(page);
    await todoSection.waitForLoad();
    
    expect(await todoSection.isOpen()).toBe(true);
  });
});