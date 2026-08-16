import { expect, test } from "./fixtures";
import { loginUser } from "./utils/auth";

test.describe("@operational Authenticated operational smoke", () => {
  test.beforeEach(async ({ page, userPersona }) => {
    await loginUser(page, userPersona.credentials);
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  test("opens command and quick-create flows from the keyboard", async ({
    page,
  }) => {
    await page.keyboard.press("Control+K");
    const commandDialog = page.getByRole("dialog", {
      name: "Search or run a command",
    });
    await expect(commandDialog).toBeVisible();
    await commandDialog
      .getByPlaceholder("Search entities, navigation, or commands...")
      .fill("quick create");
    await commandDialog.getByText("Quick create", { exact: true }).click();

    await expect(
      page.getByRole("dialog", { name: "Quick create" })
    ).toBeVisible();
    await page.getByRole("button", { name: "Cancel" }).click();

    await page.keyboard.press("c");
    await expect(
      page.getByRole("dialog", { name: "Quick create" })
    ).toBeVisible();
  });
});
