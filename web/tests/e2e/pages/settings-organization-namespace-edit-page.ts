import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";

export class SettingsOrganizationNamespaceEditPage extends BaseComponent {
  public readonly namespaceForm: Form;

  constructor(page: Page) {
    super(page);
    this.namespaceForm = new Form(page);
  }

  async goto(organizationId: string, namespaceId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/${namespaceId}/edit`
    );
  }
}
