import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import { openTodoListViaUI } from "./helpers/todo";
import { TodoCreateFormSection } from "./sections/todo-create-form-section";
import { TodoEditFormSection } from "./sections/todo-edit-form-section";
import { TodoSection } from "./sections/todo-section";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { generateComplete } from "./utils/todo";

test.describe("@todo.update TODO Update E2E Tests", () => {
  test("should update TODO title", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const editForm = new TodoEditFormSection(page);
    const todoData = generateComplete();
    
    // Create a TODO first
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    // Update the title
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Edit todo").click();
    await editForm.waitForLoad();
    
    const newTitle = `Updated ${todoData.title}`;
    await editForm.updateTodo({ title: newTitle });
    
    await waitForSuccessToast(page, "Todo updated successfully");
    expect(await todoSection.getTodoByTitle(newTitle).isVisible()).toBe(true);
  });

  test("should update TODO description", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const editForm = new TodoEditFormSection(page);
    const todoData = generateComplete();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Edit todo").click();
    await editForm.waitForLoad();
    
    const newDescription = "Updated description";
    await editForm.updateTodo({ description: newDescription });
    
    await waitForSuccessToast(page, "Todo updated successfully");
    expect(await todoItem.getByText(newDescription).isVisible()).toBe(true);
  });

  test("should update TODO priority", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const editForm = new TodoEditFormSection(page);
    const todoData = generateComplete();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Edit todo").click();
    await editForm.waitForLoad();
    
    await editForm.updateTodo({ priority: "critical" });
    
    await waitForSuccessToast(page, "Todo updated successfully");
    expect(await todoItem.getByText("critical").isVisible()).toBe(true);
  });

  test("should update multiple fields simultaneously", async ({ page, testConfig }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, { email: testUser.email, password: USER_DEFAULT_PASSWORD });
    
    const todoSection = new TodoSection(page);
    const createForm = new TodoCreateFormSection(page);
    const editForm = new TodoEditFormSection(page);
    const todoData = generateComplete();
    
    await openTodoListViaUI(page);
    await todoSection.waitForLoad();
    await todoSection.clickAddTodo();
    await createForm.waitForLoad();
    await createForm.createTodo(todoData);
    await waitForSuccessToast(page, "Todo added successfully");
    
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Edit todo").click();
    await editForm.waitForLoad();
    
    const newTitle = `Multi-update ${todoData.title}`;
    const newDescription = "Multi-field update description";
    await editForm.updateTodo({ 
      title: newTitle, 
      description: newDescription,
      priority: "urgent"
    });
    
    await waitForSuccessToast(page, "Todo updated successfully");
    const updatedItem = todoSection.getTodoByTitle(newTitle);
    expect(await updatedItem.isVisible()).toBe(true);
    expect(await updatedItem.getByText(newDescription).isVisible()).toBe(true);
    expect(await updatedItem.getByText("urgent").isVisible()).toBe(true);
  });

  test("should not allow editing completed TODOs", async ({ page, testConfig }) => {
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
    
    // Complete the TODO
    const todoItem = todoSection.getTodoByTitle(todoData.title);
    await todoItem.hover();
    await todoItem.getByTitle("Mark as complete").click();
    await waitForSuccessToast(page, "Todo updated");
    
    // Try to edit - button should be disabled
    await todoItem.hover();
    const editButton = todoItem.getByTitle("Edit todo");
    expect(await editButton.isDisabled()).toBe(true);
  });
});