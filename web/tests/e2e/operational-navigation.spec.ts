import { expect, test } from "./fixtures";
import { loginUser } from "./utils/auth";

test.describe("@operational Authenticated operational smoke", () => {
  test.beforeEach(async ({ page, userPersona }) => {
    await loginUser(page, userPersona.credentials);
    await expect(page.getByRole("heading", { name: "Home" })).toBeVisible();
  });

  // FIXME: re-enable once work items can be created via API (LMO-101 is fixture-only).
  test.skip("navigates operational routes and preserves projection context", async ({
    page,
  }) => {
    await page
      .getByRole("link", { name: "My Work", exact: true })
      .first()
      .click();
    await expect(page.getByRole("heading", { name: "My Work" })).toBeVisible();

    await page.getByRole("button", { name: "Table", exact: true }).click();
    await expect(page).toHaveURL(/\/my-work\?.*layout=table/);

    await page
      .getByRole("button", { name: /^Inspect LMO-101:/ })
      .first()
      .click();
    await expect(page).toHaveURL(/selected=work%3Aele-101/);

    const inspector = page.getByRole("dialog");
    await expect(inspector).toBeVisible();
    await expect(
      inspector.getByRole("complementary", { name: "LMO-101 details" })
    ).toBeVisible();

    // Sheet is modal — close before changing layout, then re-open selection.
    await inspector.getByRole("button", { name: "Close inspector" }).click();
    await expect(inspector).not.toBeVisible();
    await expect(page).not.toHaveURL(/selected=/);

    await page.getByRole("button", { name: "List", exact: true }).click();
    await expect(page).toHaveURL(/layout=list/);

    await page
      .getByRole("button", { name: /^Inspect LMO-101:/ })
      .first()
      .click();
    await expect(page).toHaveURL(/layout=list/);
    await expect(page).toHaveURL(/selected=work%3Aele-101/);
    await expect(inspector).toBeVisible();
    await expect(
      inspector.getByRole("complementary", { name: "LMO-101 details" })
    ).toBeVisible();

    // URL-owned projection: layout + selection survive a direct navigation.
    await page.goto("/my-work?layout=table&selected=work%3Aele-101");
    await expect(page.getByRole("heading", { name: "My Work" })).toBeVisible();
    await expect(page).toHaveURL(/layout=table/);
    await expect(page).toHaveURL(/selected=work%3Aele-101/);
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(
      page.getByRole("complementary", { name: "LMO-101 details" })
    ).toBeVisible();

    await page.getByRole("button", { name: "Close inspector" }).click();
    await expect(page.getByRole("dialog")).not.toBeVisible();
    await expect(page).not.toHaveURL(/selected=/);

    await page.goto("/search");
    await expect(page.getByRole("heading", { name: "Search" })).toBeVisible();
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

  // FIXME: re-enable once work items can be created via API (LMO-101 is fixture-only).
  test.skip("uses a sheet inspector overlay on tablet", async ({ page }) => {
    await page.setViewportSize({ width: 800, height: 900 });
    await page.goto("/my-work?layout=list");
    await expect(page.getByRole("heading", { name: "My Work" })).toBeVisible();

    await page
      .getByRole("button", { name: /^Inspect LMO-101:/ })
      .first()
      .click();
    await expect(page).toHaveURL(/selected=work%3Aele-101/);

    const overlay = page.getByRole("dialog");
    await expect(overlay).toBeVisible();
    await expect(
      overlay.getByRole("complementary", { name: "LMO-101 details" })
    ).toBeVisible();
    await overlay.getByRole("button", { name: "Close inspector" }).click();
    await expect(overlay).not.toBeVisible();
    await expect(page).not.toHaveURL(/selected=/);
  });
});
