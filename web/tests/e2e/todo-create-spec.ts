import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { openAddTodoFormViaCommandPalette, openAddTodoFormViaKeyboard, openTodoListViaUI } from "./helpers/todo";
import { TodoCreateFormSection } from "./sections/todo-create-form-section";
import { TodoSection } from "./sections/todo-section";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { generateComplete, generateTitleOnly, generateWithDescription, generateWithPriority } from "./utils/todo";

test.describe("@todo.create TODO Create E2E Tests", () => {
  test("should create TODO with title only via UI button", async ({ page, testConfig }) => {
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
    expect(await todoSection.getTodoByTitle(todoData.title).isVisible()).toBe(true);
  });

  test("should create TODO with title and description", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateWithDescription();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    
    await waitForSuccessToast(page, "Todo added successfully");
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    expect(await todoItem.isVisible()).toBe(true);
    expect(await todoItem.getByText(todoData.description!).isVisible()).toBe(true);
  });

  test("should create TODO with title and priority", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateWithPriority("urgent");
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    
    await waitForSuccessToast(page, "Todo added successfully");
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    expect(await todoItem.isVisible()).toBe(true);
    expect(await todoItem.getByText("urgent").isVisible()).toBe(true);
  });

  test("should create TODO with all fields", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateComplete();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    
    await waitForSuccessToast(page, "Todo added successfully");
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    expect(await todoItem.isVisible()).toBe(true);
  });

  test("should create TODO via keyboard shortcut (Shift+T+N)", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateTitleOnly();
    
    await openAddTodoFormViaKeyboard(page);
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    
    await waitForSuccessToast(page, "Todo added successfully");
  });

  test("should create TODO via command palette", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const createForm = new TodoCreateFormSection(page);
    const todoData = generateTitleOnly();
    
    await openAddTodoFormViaCommandPalette(page);
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    
    await waitForSuccessToast(page, "Todo added successfully");
  });

  test("should create multiple TODOs with 'Create More' checkbox", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const todo1 = generateTitleOnly();
    const todo2 = generateTitleOnly();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo({ ...todo1, createMore: true });
    
    await waitForSuccessToast(page, "Todo added successfully");
    
    // Form should still be open
    await createForm.waitForLoad();
    await createForm.createTodo(todo2);
    
    await waitForSuccessToast(page, "Todo added successfully");
    expect(await todoSection.getTodoCount()).toBeGreaterThanOrEqual(2);
  });
});