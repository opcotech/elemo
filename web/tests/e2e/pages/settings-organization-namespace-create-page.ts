import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";

export class SettingsOrganizationNamespaceCreatePage extends BaseComponent {
  public readonly namespaceForm: Form;

  constructor(page: Page) {
    super(page);
    this.namespaceForm = new Form(page);
  }

  async goto(organizationId: string): Promise<void> {
    await this.gotoForOrganization(organizationId);
  }

  async gotoForOrganization(organizationId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/new`
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
