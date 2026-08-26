import { describe, expect, it } from "vitest";

import {
  LIBRARY_ROOT_FOLDER_LABEL,
  LIBRARY_ROOT_FOLDER_VALUE,
  documentLibraryHref,
  documentLibraryKindFromType,
  documentLibraryListItems,
  documentLibrarySearchParams,
  documentLibrarySearchSchema,
  documentLibraryTypeLabel,
  filterDocumentLibraryListItems,
  folderMoveTargets,
  folderPathLabel,
  libraryBrowseCrumbs,
  libraryFolderPickerOptions,
  resolveDocumentLibrarySearch,
} from "./library";
import type { LibraryFolderOption } from "./library";

describe("document library search", () => {
  it("treats a missing search as the library root", () => {
    expect(resolveDocumentLibrarySearch({})).toEqual({ showAll: false });
    expect(documentLibrarySearchParams({})).toEqual({});
  });

  it("keeps a folder browse param", () => {
    expect(resolveDocumentLibrarySearch({ folder: "folder-1" })).toEqual({
      folderId: "folder-1",
      showAll: false,
    });
    expect(documentLibrarySearchParams({ folderId: "folder-1" })).toEqual({
      folder: "folder-1",
    });
  });

  it("ignores folder when All is selected", () => {
    expect(
      resolveDocumentLibrarySearch({ folder: "folder-1", all: true })
    ).toEqual({ showAll: true });
    expect(
      documentLibrarySearchParams({ folderId: "folder-1", showAll: true })
    ).toEqual({ all: true });
  });

  it("coerces all=true from the query string", () => {
    expect(documentLibrarySearchSchema.parse({ all: "true" })).toEqual({
      all: true,
    });
    expect(documentLibrarySearchSchema.parse({ all: "false" })).toEqual({});
  });
});

describe("document library helpers", () => {
  it("builds library document hrefs", () => {
    expect(
      documentLibraryHref({
        kind: "namespace",
        organizationSlug: "acme",
        namespaceSlug: "platform",
      })
    ).toBe("/organizations/acme/namespaces/platform/documents");
    expect(
      documentLibraryHref({
        kind: "organization",
        organizationSlug: "acme",
      })
    ).toBe("/organizations/acme/documents");
    expect(documentLibraryKindFromType("Namespace")).toBe("namespace");
    expect(documentLibraryKindFromType("Organization")).toBe("organization");
    expect(documentLibraryTypeLabel("organization")).toBe("Organization");
    expect(documentLibraryTypeLabel("namespace")).toBe("Namespace");
  });

  it("lists organization libraries before namespaces, each sorted by name", () => {
    expect(
      documentLibraryListItems(
        [
          {
            id: "org-b",
            slug: "zebra-org",
            name: "Zebra Org",
            document_count: 2,
          },
          { id: "org-a", slug: "alpha-org", name: "Alpha Org" },
        ],
        [
          {
            id: "ns-b",
            slug: "ops",
            organizationSlug: "acme",
            name: "Ops",
            document_count: 4,
          },
          {
            id: "ns-a",
            slug: "docs",
            organizationSlug: "acme",
            name: "Docs",
          },
        ]
      )
    ).toEqual([
      {
        id: "org-a",
        kind: "organization",
        name: "Alpha Org",
        typeLabel: "Organization",
        href: "/organizations/alpha-org/documents",
        documentCount: 0,
      },
      {
        id: "org-b",
        kind: "organization",
        name: "Zebra Org",
        typeLabel: "Organization",
        href: "/organizations/zebra-org/documents",
        documentCount: 2,
      },
      {
        id: "ns-a",
        kind: "namespace",
        name: "Docs",
        typeLabel: "Namespace",
        href: "/organizations/acme/namespaces/docs/documents",
        documentCount: 0,
      },
      {
        id: "ns-b",
        kind: "namespace",
        name: "Ops",
        typeLabel: "Namespace",
        href: "/organizations/acme/namespaces/ops/documents",
        documentCount: 4,
      },
    ]);
  });

  it("filters library entries by name or type", () => {
    const items = documentLibraryListItems(
      [{ id: "org-1", slug: "acme", name: "Acme" }],
      [
        {
          id: "ns-1",
          slug: "platform",
          organizationSlug: "acme",
          name: "Platform",
        },
      ]
    );

    expect(filterDocumentLibraryListItems(items, "platform")).toEqual([
      items[1],
    ]);
    expect(filterDocumentLibraryListItems(items, "organization")).toEqual([
      items[0],
    ]);
    expect(filterDocumentLibraryListItems(items, "  ")).toEqual(items);
  });

  it("joins folder path labels", () => {
    expect(folderPathLabel(["Guides", "Architecture"])).toBe(
      "Guides / Architecture"
    );
  });
});

describe("libraryBrowseCrumbs", () => {
  it("marks the library root as current when the path is empty", () => {
    expect(libraryBrowseCrumbs([])).toEqual([
      { name: "Library", current: true },
    ]);
  });

  it("builds a library-to-folder trail and marks the last folder current", () => {
    expect(
      libraryBrowseCrumbs([
        { id: "guides", name: "Guides" },
        { id: "architecture", name: "Architecture" },
      ])
    ).toEqual([
      { name: "Library", current: false },
      { folderId: "guides", name: "Guides", current: false },
      { folderId: "architecture", name: "Architecture", current: true },
    ]);
  });
});

function folderOption(
  id: string,
  name: string,
  parentId: string | null,
  pathLabel = name
): LibraryFolderOption {
  return { id, name, pathLabel, parentId };
}

describe("folderMoveTargets", () => {
  const guides = folderOption("guides", "Guides", null);
  const architecture = folderOption(
    "architecture",
    "Architecture",
    "guides",
    "Guides / Architecture"
  );
  const adr = folderOption(
    "adr",
    "ADRs",
    "architecture",
    "Guides / Architecture / ADRs"
  );
  const drafts = folderOption("drafts", "Drafts", null);
  const folders = [guides, architecture, adr, drafts];

  it("excludes the folder and its descendants", () => {
    expect(
      folderMoveTargets(folders, "guides").map((folder) => folder.id)
    ).toEqual(["drafts"]);
    expect(
      folderMoveTargets(folders, "architecture").map((folder) => folder.id)
    ).toEqual(["guides", "drafts"]);
  });

  it("keeps siblings and ancestors when moving a nested folder", () => {
    expect(
      folderMoveTargets(folders, "adr").map((folder) => folder.id)
    ).toEqual(["guides", "architecture", "drafts"]);
  });

  it("returns no folder targets when the library only has that folder", () => {
    expect(folderMoveTargets([guides], "guides")).toEqual([]);
  });
});

describe("libraryFolderPickerOptions", () => {
  it("puts Library root first and uses path labels for nested folders", () => {
    const options = libraryFolderPickerOptions([
      folderOption("guides", "Guides", null),
      folderOption(
        "architecture",
        "Architecture",
        "guides",
        "Guides / Architecture"
      ),
    ]);

    expect(options[0]).toEqual({
      value: LIBRARY_ROOT_FOLDER_VALUE,
      title: LIBRARY_ROOT_FOLDER_LABEL,
      searchText: "library root unfiled",
    });
    expect(options.slice(1).map((option) => option.title)).toEqual([
      "Guides",
      "Guides / Architecture",
    ]);
    expect(options[2]?.searchText).toContain("Architecture");
  });
});
