import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForElementVisible } from "../helpers";

export class SettingsOrganizationNamespaceDetailsPage extends BaseComponent {
  constructor(page: Page) {
    super(page);
  }

  async goto(organizationId: string, namespaceId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/${namespaceId}`
    );
  }

  async waitForLoad(): Promise<void> {
    const heading = this.page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .first();
    await waitForElementVisible(heading);
  }

  async getTitleText(): Promise<string> {
    const heading = this.page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .first();
    await waitForElementVisible(heading);
    return (await heading.textContent()) ?? "";
  }
}
