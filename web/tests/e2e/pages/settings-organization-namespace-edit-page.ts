import type { Page } from "@playwright/test";
import { settingsNamespaceEditPath } from "@/lib/paths";
import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";

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
