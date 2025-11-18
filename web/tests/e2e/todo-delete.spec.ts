import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { openTodoListViaUI } from "./helpers/todo";
import { TodoCreateFormSection } from "./sections/todo-create-form-section";
import { TodoSection } from "./sections/todo-section";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { generateTitleOnly } from "./utils/todo";

test.describe("@todo.delete TODO Delete E2E Tests", () => {
  test("should delete TODO item", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateTitleOnly();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Delete todo").click();
    
    await waitForSuccessToast(page, "Todo deleted");
    expect(await todoItem.isVisible()).toBe(false);
  });

  test("should delete completed TODO item", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateTitleOnly();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Mark as complete").click();
    await waitForSuccessToast(page, "marked as completed");
    
    await todoItem.hover();
    await todoItem.getByTitle("Delete todo").click();
    
    await waitForSuccessToast(page, "Todo deleted");
    expect(await todoItem.isVisible()).toBe(false);
  });

  test("should show empty state after deleting all TODOs", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateTitleOnly();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Delete todo").click();
    await waitForSuccessToast(page, "Todo deleted");
    
    expect(await todoSection.getEmptyState().isVisible()).toBe(true);
  });
});
