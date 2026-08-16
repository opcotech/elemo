import { expect, test } from "./fixtures";
import {
  fillLocator,
  pressPaletteShortcut,
  runPaletteCommand,
  waitForSuccessToast,
} from "./helpers";
import { HomePage } from "./pages";
import {
  QuickCreateSection,
  TodoFormSection,
  TodoSheetSection,
} from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";
import { getRandomString } from "./utils/random";

test.describe("@todos.create Todo create E2E Tests", () => {
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

  test("creates a todo from the Add Todo button in the sheet", async ({
    page,
  }) => {
    const title = `Sheet create ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    await todoForm.fillAndSubmitAdd({ title });
    await waitForSuccessToast(page, "Todo added successfully");
    await expect(todoForm.getAddDialog()).not.toBeVisible();

    await expect(todoSheet.getGroup("Later")).toBeVisible();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("creates a todo from the command palette Add Todo action", async ({
    page,
  }) => {
    const title = `Palette create ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await runPaletteCommand(page, "Add Todo");
    await todoForm.fillAndSubmitAdd({ title }, { allowHidden: true });
    await waitForSuccessToast(page, "Todo added successfully");
    await todoSheet.waitForOpen();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("creates a todo with Shift+T+N while the palette is open", async ({
    page,
  }) => {
    const title = `Shortcut create ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await pressPaletteShortcut(page, "n");
    await todoForm.fillAndSubmitAdd({ title }, { allowHidden: true });
    await waitForSuccessToast(page, "Todo added successfully");
    await todoSheet.waitForOpen();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("creates with title only", async ({ page }) => {
    const title = `Title only ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    await todoForm.fillAndSubmitAdd({ title });
    await waitForSuccessToast(page, "Todo added successfully");

    const item = todoSheet.getItem(title);
    await expect(item.getByText(title, { exact: true })).toBeVisible();
    await expect(item.getByText("Normal", { exact: true })).toBeVisible();
  });

  test("creates with title and description", async ({ page }) => {
    const title = `Title desc ${getRandomString(8)}`;
    const description = `Description ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    await todoForm.fillAndSubmitAdd({ title, description });
    await waitForSuccessToast(page, "Todo added successfully");

    const item = todoSheet.getItem(title);
    await expect(item.getByText(title, { exact: true })).toBeVisible();
    await expect(item.getByText(description)).toBeVisible();
  });

  test("creates with title and priority", async ({ page }) => {
    const title = `Title priority ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    await todoForm.fillAndSubmitAdd({ title, priority: "Urgent" });
    await waitForSuccessToast(page, "Todo added successfully");

    const item = todoSheet.getItem(title);
    await expect(item.getByText(title, { exact: true })).toBeVisible();
    await expect(item.getByText("Urgent", { exact: true })).toBeVisible();
  });

  test("creates with title, description, and priority", async ({ page }) => {
    const title = `Full fields ${getRandomString(8)}`;
    const description = `Full description ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    await todoForm.fillAndSubmitAdd({
      title,
      description,
      priority: "Critical",
    });
    await waitForSuccessToast(page, "Todo added successfully");

    const item = todoSheet.getItem(title);
    await expect(item.getByText(title, { exact: true })).toBeVisible();
    await expect(item.getByText(description)).toBeVisible();
    await expect(item.getByText("Critical", { exact: true })).toBeVisible();
  });

  test("keeps the add dialog open when Create more is checked", async ({
    page,
  }) => {
    const firstTitle = `Create more first ${getRandomString(8)}`;
    const secondTitle = `Create more second ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await todoSheet.clickAddTodo();
    const dialog = await todoForm.waitForAddDialog();
    await todoForm.setCreateMore(true);
    await todoForm.fillTodoFields(dialog, { title: firstTitle });
    await todoForm.submitAdd();
    await waitForSuccessToast(page, "Todo added successfully");
    await expect(dialog).toBeVisible();
    await expect(dialog.getByLabel("Title")).toHaveValue("");

    await fillLocator(dialog.getByLabel("Title"), secondTitle);
    await todoForm.submitAdd();
    await expect(dialog.getByLabel("Title")).toHaveValue("");
    await expect(dialog).toBeVisible();

    await dialog.getByRole("button", { name: "Cancel" }).click();
    await expect(dialog).not.toBeVisible();
    await expect(todoSheet.getItem(firstTitle)).toBeVisible();
    await expect(todoSheet.getItem(secondTitle)).toBeVisible();
  });

  test("creates a personal todo from Quick create", async ({ page }) => {
    const title = `Quick create ${getRandomString(8)}`;
    const homePage = new HomePage(page);
    const todoSheet = new TodoSheetSection(page);

    await homePage.clickCreate();
    await homePage.quickCreate.waitForLoad();
    await expect(homePage.quickCreate.getEntityTypeSelect()).toContainText(
      "Personal todo"
    );
    await homePage.quickCreate.fillField("Title", title);
    await homePage.quickCreate.submitTodo();
    await waitForSuccessToast(page, "Todo added successfully");

    await todoSheet.openFromHeader();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("creates a personal todo with the C shortcut", async ({ page }) => {
    const title = `Shortcut C ${getRandomString(8)}`;
    const quickCreate = new QuickCreateSection(page);
    const todoSheet = new TodoSheetSection(page);

    await page.keyboard.press("c");
    await quickCreate.waitForLoad();
    await expect(quickCreate.getEntityTypeSelect()).toContainText(
      "Personal todo"
    );
    await quickCreate.fillField("Title", title);
    await quickCreate.submitTodo();
    await waitForSuccessToast(page, "Todo added successfully");

    await todoSheet.openFromHeader();
    await expect(todoSheet.getItem(title)).toBeVisible();
  });

  test("creates a todo from the sheet empty-state Add Todo button", async ({
    page,
  }) => {
    const title = `Empty state ${getRandomString(8)}`;
    const todoSheet = new TodoSheetSection(page);
    const todoForm = new TodoFormSection(page);

    await todoSheet.openFromHeader();
    await expect(
      todoSheet.getSheet().getByText("No todos found")
    ).toBeVisible();
    await todoSheet.clickEmptyStateAddTodo();
    await todoForm.fillAndSubmitAdd({ title });
    await waitForSuccessToast(page, "Todo added successfully");
    await expect(todoSheet.getItem(title)).toBeVisible();
  });
});
