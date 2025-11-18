import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { openTodoListViaUI } from "./helpers/todo";
import { TodoCreateFormSection } from "./sections/todo-create-form-section";
import { TodoSection } from "./sections/todo-section";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { generateTitleOnly } from "./utils/todo";

test.describe("@todo.complete TODO Complete/Uncomplete E2E Tests", () => {
  test("should mark TODO as complete", async ({ page, testConfig }) => {
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
    expect(await todoItem.getByText("Completed").isVisible()).toBe(true);
  });

  test("should mark completed TODO as incomplete", async ({ page, testConfig }) => {
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
    await todoItem.getByTitle("Mark as incomplete").click();
    
    await waitForSuccessToast(page, "marked as incomplete");
    expect(await todoItem.getByText("Completed").isVisible()).toBe(false);
  });

  test("should apply completed styling to completed TODO", async ({ page, testConfig }) => {
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
    
    // Check for strikethrough on title
    const title = todoItem.locator("h4").first();
    await expect(title).toHaveClass(/line-through/);
    
    // Check for reduced opacity
    await expect(todoItem).toHaveClass(/opacity-75/);
  });
});