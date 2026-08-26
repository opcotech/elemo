import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";

import { settingsNamespaceEditPath } from "@/lib/paths";

export class SettingsOrganizationNamespaceEditPage extends BaseComponent {
  public readonly namespaceForm: Form;

  constructor(page: Page) {
    super(page);
    this.namespaceForm = new Form(page);
  }

  async goto(organizationSlug: string, namespaceSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsNamespaceEditPath({ organizationSlug, namespaceSlug })
    );
  }
}
