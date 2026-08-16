import { expect, test } from "./fixtures";
import { pressPaletteShortcut, runPaletteCommand } from "./helpers";
import { TodoSheetSection } from "./sections";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import { createUser } from "./utils/db";

test.describe("@todos.open Todo open E2E Tests", () => {
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

  test("opens and closes the todo list from the header", async ({ page }) => {
    const todoSheet = new TodoSheetSection(page);
    const sheet = await todoSheet.openFromHeader();
    await expect(
      sheet.getByText(/No todos found|open ·|completed/)
    ).toBeVisible();

    await todoSheet.closeWithButton();
  });

  test("opens the todo list from the command palette", async ({ page }) => {
    const todoSheet = new TodoSheetSection(page);
    await runPaletteCommand(page, "Show Todos");
    await todoSheet.waitForOpen();
  });

  test("opens the todo list with Shift+T+S while the palette is open", async ({
    page,
  }) => {
    const todoSheet = new TodoSheetSection(page);
    await pressPaletteShortcut(page, "s");
    await todoSheet.waitForOpen();
  });

  test("closes the todo list with Escape", async ({ page }) => {
    const todoSheet = new TodoSheetSection(page);
    await todoSheet.openFromHeader();
    await todoSheet.closeWithEscape();
  });
});
