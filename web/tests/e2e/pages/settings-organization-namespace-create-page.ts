import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";

import { settingsNamespaceNewPath } from "@/lib/paths";

export class SettingsOrganizationNamespaceCreatePage extends BaseComponent {
  public readonly namespaceForm: Form;

  constructor(page: Page) {
    super(page);
    this.namespaceForm = new Form(page);
  }

  async goto(organizationSlug: string): Promise<void> {
    await this.gotoForOrganization(organizationSlug);
  }

  async gotoForOrganization(organizationSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsNamespaceNewPath({ organizationSlug })
    );
  }

  async gotoGlobal(): Promise<void> {
    await navigateAndWait(this.page, "/settings/namespaces/new");
  }

  async selectOrganization(organizationName: string): Promise<void> {
    let trigger = this.page
      .getByRole("combobox", { name: /organization/i })
      .first();

    if ((await trigger.count()) === 0) {
      trigger = this.page
        .locator("label", { hasText: "Organization" })
        .locator("..")
        .locator("button")
        .first();
    }

    await trigger.click();
    await this.page
      .getByRole("option", { name: new RegExp(organizationName, "i") })
      .first()
      .click();
  }
}
