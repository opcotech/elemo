import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import {
  clickUntilVisible,
  navigateAndWait,
  waitForElementVisible,
} from "../helpers";
import { SectionContainerMixin } from "../mixins";
import { QuickCreateSection } from "../sections/quick-create-section";

/**
 * Continue working list on Home.
 */
class HomeContinueWorkingSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='home-continue-working']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getViewAllLink(): Locator {
    return this.getSectionContainer().getByRole("link", { name: "View all" });
  }

  getWorkItem(key: string): Locator {
    return this.getSectionContainer().getByRole("link", {
      name: key,
      exact: true,
    });
  }
}

/**
 * Personal todos preview on Home.
 */
class HomePersonalTodosSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='home-personal-todos']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getViewAllButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "View all",
      exact: true,
    });
  }

  getAddTodoButton(): Locator {
    return this.getSectionContainer().getByRole("button", { name: "Add todo" });
  }

  getTodo(title: string): Locator {
    return this.getSectionContainer()
      .getByRole("button")
      .filter({ hasText: title });
  }

  getGroup(label: string): Locator {
    return this.getSectionContainer().getByRole("list", {
      name: `${label} todos`,
    });
  }
}

/**
 * Page Object Model for Home.
 */
export class HomePage extends BaseComponent {
  public readonly continueWorking: HomeContinueWorkingSection;
  public readonly personalTodos: HomePersonalTodosSection;
  public readonly quickCreate: QuickCreateSection;

  constructor(page: Page) {
    super(page);
    this.continueWorking = new HomeContinueWorkingSection(page);
    this.personalTodos = new HomePersonalTodosSection(page);
    this.quickCreate = new QuickCreateSection(page);
  }

  async goto(): Promise<void> {
    await navigateAndWait(this.page, "/");
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Home" }),
      options
    );
  }

  getCreateButton(): Locator {
    return this.page.getByRole("button", { name: "Create", exact: true });
  }

  async clickCreate(): Promise<void> {
    await clickUntilVisible(
      this.getCreateButton(),
      this.page.getByRole("dialog", { name: "Quick create" })
    );
  }
}
