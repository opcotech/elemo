import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForElementVisible } from "../helpers";
import { QuickCreateSection } from "../sections/quick-create-section";
import { WorkInspectorSection } from "../sections/work-inspector-section";
import { WorkSurfaceSection } from "../sections/work-surface-section";

/**
 * Page Object Model for Work surfaces: My Work, namespace Work, and project Work.
 */
export class WorkPage extends BaseComponent {
  public readonly surface: WorkSurfaceSection;
  public readonly inspector: WorkInspectorSection;
  public readonly quickCreate: QuickCreateSection;

  constructor(page: Page) {
    super(page);
    this.surface = new WorkSurfaceSection(page);
    this.inspector = new WorkInspectorSection(page);
    this.quickCreate = new QuickCreateSection(page);
  }

  async gotoMyWork(): Promise<void> {
    await navigateAndWait(this.page, "/my-work");
  }

  async gotoNamespaceWork(namespaceId: string): Promise<void> {
    await navigateAndWait(this.page, `/namespaces/${namespaceId}/work`);
  }

  async gotoProjectWork(namespaceId: string, projectId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/namespaces/${namespaceId}/projects/${projectId}/work`
    );
  }

  async waitForLoad(heading?: string): Promise<void> {
    const title = heading
      ? this.page.getByRole("heading", { name: heading })
      : this.page
          .getByRole("heading", { level: 1 })
          .filter({ hasText: /Work/ });
    await waitForElementVisible(title);
    await this.surface.waitForLoad();
  }
}
