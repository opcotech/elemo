import { createTodo } from "./api";
import { expect, test } from "./fixtures";
import { HomePage } from "./pages";
import { TodoSheetSection } from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

test.describe("@todos.home Todo home E2E Tests", () => {
  test("shows a personal todo in the Home preview", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    const testUser = await createUser(testConfig);
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const title = `Home preview ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const homePage = new HomePage(page);
    await homePage.waitForLoad();
    await homePage.personalTodos.waitForLoad();
    await expect(homePage.personalTodos.getTodo(title)).toBeVisible();
  });

  test("opens the todo sheet from View all", async ({
    page,
    testConfig,
    createApiClient,
  }) => {
    const testUser = await createUser(testConfig);
    const apiClient = await createApiClient(
      testUser.email,
      USER_DEFAULT_PASSWORD
    );
    const title = `Home view all ${getRandomString(8)}`;
    await createTodo(apiClient, { title, owned_by: testUser.id });

    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const homePage = new HomePage(page);
    const todoSheet = new TodoSheetSection(page);
    await homePage.waitForLoad();
    await homePage.personalTodos.waitForLoad();
    await expect(homePage.personalTodos.getTodo(title)).toBeVisible();

    await homePage.personalTodos.getViewAllButton().click();
    await todoSheet.waitForOpen();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("opens Quick create from the empty Add todo action", async ({
    page,
    testConfig,
  }) => {
    const testUser = await createUser(testConfig);
    await loginUser(page, {
      email: testUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const homePage = new HomePage(page);
    await homePage.waitForLoad();
    await homePage.personalTodos.waitForLoad();
    await expect(page.getByText("No open todos")).toBeVisible();

    await homePage.personalTodos.getAddTodoButton().click();
    await homePage.quickCreate.waitForLoad();
  });
});
