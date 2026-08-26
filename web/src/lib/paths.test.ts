import { describe, expect, it } from "vitest";

import {
  documentPath,
  namespacePath,
  organizationPath,
  projectPath,
  settingsNamespacePath,
  settingsOrganizationPath,
  settingsProjectPath,
  workItemPath,
} from "./paths";

describe("canonical browser paths", () => {
  it("builds hierarchical org, namespace, project, work, and document hrefs", () => {
    expect(organizationPath({ organizationSlug: "acme" })).toBe(
      "/organizations/acme"
    );
    expect(
      namespacePath({ organizationSlug: "acme", namespaceSlug: "platform" })
    ).toBe("/organizations/acme/namespaces/platform");
    expect(
      projectPath({
        organizationSlug: "acme",
        namespaceSlug: "platform",
        projectKey: "PLAT",
      })
    ).toBe("/organizations/acme/namespaces/platform/projects/PLAT");
    expect(
      workItemPath({
        organizationSlug: "acme",
        namespaceSlug: "platform",
        issueKey: "PLAT-1",
      })
    ).toBe("/work/acme/platform/PLAT-1");
    expect(documentPath("9bsv0s46s6s002p9ltq0")).toBe(
      "/documents/9bsv0s46s6s002p9ltq0"
    );
  });

  it("builds settings hierarchy with the same slug and key params", () => {
    expect(settingsOrganizationPath({ organizationSlug: "acme" })).toBe(
      "/settings/organizations/acme"
    );
    expect(
      settingsNamespacePath({
        organizationSlug: "acme",
        namespaceSlug: "platform",
      })
    ).toBe("/settings/organizations/acme/namespaces/platform");
    expect(
      settingsProjectPath({
        organizationSlug: "acme",
        namespaceSlug: "platform",
        projectKey: "PLAT",
      })
    ).toBe("/settings/organizations/acme/namespaces/platform/projects/PLAT");
  });
});
