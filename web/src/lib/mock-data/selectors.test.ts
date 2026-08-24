import { describe, expect, it } from "vitest";

import {
  API_BACKED_DOMAINS,
  MOCK_ONLY_DOMAINS,
  getDocumentBody,
  mockActivity,
  mockAttentionSignals,
  mockDocumentBodies,
  mockGlobalSearchEntries,
  mockPeople,
  mockRelations,
  mockSavedViews,
  mockTimeline,
  mockWorkItems,
  searchGlobalFixtures,
  selectActivity,
  selectAttentionSignals,
  selectDocumentBodies,
  selectPeople,
  selectRelations,
  selectSavedViews,
  selectTimeline,
  selectWorkItems,
} from ".";

describe("mock-data boundary", () => {
  it("keeps API-backed and fixture-only domains explicit and disjoint", () => {
    expect(API_BACKED_DOMAINS).toEqual(
      expect.arrayContaining([
        "organizations",
        "namespaces",
        "projects",
        "issues",
        "todos",
        "notifications",
        "memberships",
      ])
    );
    expect(MOCK_ONLY_DOMAINS).toEqual(
      expect.arrayContaining([
        "savedViews",
        "documentBodies",
        "relations",
        "activity",
        "timeline",
        "attentionSignals",
        "people",
        "globalSearch",
      ])
    );
    expect(
      API_BACKED_DOMAINS.filter((domain) =>
        (MOCK_ONLY_DOMAINS as readonly string[]).includes(domain)
      )
    ).toEqual([]);
  });

  it("marks every fixture record as mock-sourced", () => {
    const fixtureCollections = [
      mockWorkItems,
      mockSavedViews,
      mockDocumentBodies,
      mockRelations,
      mockActivity,
      mockTimeline,
      mockAttentionSignals,
      mockPeople,
      mockGlobalSearchEntries,
    ];

    expect(fixtureCollections.flat()).not.toHaveLength(0);
    expect(
      fixtureCollections.flat().every((record) => record.dataSource === "mock")
    ).toBe(true);
  });
});

describe("selectWorkItems", () => {
  it("applies global, namespace, project, and person scopes", () => {
    expect(selectWorkItems()).toHaveLength(mockWorkItems.length);
    expect(
      selectWorkItems({
        scope: { type: "namespace", namespaceId: "namespace-product" },
      }).map((item) => item.key)
    ).toEqual(["LMO-101", "LMO-201", "LMO-102", "LMO-103"]);
    expect(
      selectWorkItems({
        scope: {
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        },
      }).map((item) => item.key)
    ).toEqual(["LMO-101", "LMO-102", "LMO-103"]);
    expect(
      selectWorkItems({
        scope: { type: "person", personId: "person-katherine" },
      }).map((item) => item.key)
    ).toEqual(["OPS-301", "OPS-302"]);
  });

  it("combines text, status, priority, assignee, and label filters", () => {
    const result = selectWorkItems({
      filters: {
        text: "projection",
        statuses: ["in progress"],
        priorities: ["highest"],
        assigneeIds: ["person-ada"],
        labelIds: ["frontend", "navigation"],
      },
    });

    expect(result.map((item) => item.key)).toEqual(["LMO-101"]);
  });

  it.each([
    ["overdue", ["OPS-302", "LMO-101"]],
    ["today", ["LMO-102"]],
    ["upcoming", ["OPS-301", "LMO-201"]],
    ["none", ["LMO-103"]],
  ] as const)(
    "filters %s due dates against a supplied clock",
    (dueDate, keys) => {
      expect(
        selectWorkItems({
          filters: { dueDate },
          now: "2026-08-10T12:00:00.000Z",
          sort: [{ field: "dueDate", direction: "asc" }],
        }).map((item) => item.key)
      ).toEqual(keys);
    }
  );

  it("sorts by multiple fields without mutating fixture order", () => {
    const originalOrder = mockWorkItems.map((item) => item.id);
    const result = selectWorkItems({
      scope: { type: "namespace", namespaceId: "namespace-product" },
      sort: [
        { field: "priority", direction: "desc" },
        { field: "title", direction: "asc" },
      ],
    });

    expect(result.map((item) => item.key)).toEqual([
      "LMO-101",
      "LMO-102",
      "LMO-201",
      "LMO-103",
    ]);
    expect(mockWorkItems.map((item) => item.id)).toEqual(originalOrder);
  });

  it("always keeps missing due dates after dated work", () => {
    expect(
      selectWorkItems({
        scope: {
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        },
        sort: [{ field: "dueDate", direction: "desc" }],
      }).map((item) => item.key)
    ).toEqual(["LMO-102", "LMO-101", "LMO-103"]);
  });
});

describe("domain selectors", () => {
  it("makes global and namespace views available in a project scope", () => {
    expect(
      selectSavedViews({
        scope: {
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        },
      }).map((view) => view.id)
    ).toEqual([
      "view-my-urgent-work",
      "view-product-planning",
      "view-web-delivery",
    ]);

    expect(
      selectSavedViews({
        scope: {
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        },
        includeGlobal: false,
      }).map((view) => view.id)
    ).toEqual(["view-product-planning", "view-web-delivery"]);
  });

  it("selects document bodies by identity and scope", () => {
    expect(getDocumentBody("document-projection-model")?.title).toBe(
      "Work projection model"
    );
    expect(
      selectDocumentBodies({
        type: "namespace",
        namespaceId: "namespace-operations",
      }).map((document) => document.documentId)
    ).toEqual(["document-incident-review"]);
  });

  it("traverses relations by direction and kind", () => {
    expect(
      selectRelations({
        entity: { id: "lmo-101", type: "work-item" },
        direction: "incoming",
      }).map((relation) => relation.id)
    ).toEqual(["relation-projection-document"]);
    expect(
      selectRelations({
        entity: { id: "lmo-101", type: "work-item" },
        direction: "outgoing",
        kinds: ["depends-on"],
      }).map((relation) => relation.id)
    ).toEqual(["relation-work-dependency"]);
  });

  it("orders scoped activity newest-first and honors limits", () => {
    expect(
      selectActivity({
        entity: { id: "lmo-101", type: "work-item" },
        limit: 2,
      }).map((entry) => entry.id)
    ).toEqual(["activity-lmo-101-status", "activity-lmo-101-assigned"]);
  });

  it("filters timeline entries by scope and overlap range", () => {
    expect(
      selectTimeline({
        scope: {
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        },
        from: "2026-08-10T00:00:00.000Z",
        to: "2026-08-13T00:00:00.000Z",
      }).map((entry) => entry.id)
    ).toEqual(["timeline-lmo-101", "timeline-lmo-102"]);
  });

  it("hides acknowledged attention by default and prioritizes severity", () => {
    expect(
      selectAttentionSignals({
        scope: { type: "person", personId: "person-ada" },
      }).map((signal) => signal.id)
    ).toEqual(["attention-lmo-101-overdue", "attention-lmo-201-blocked"]);
    expect(
      selectAttentionSignals({
        personId: "person-grace",
        includeAcknowledged: true,
      }).map((signal) => signal.id)
    ).toEqual(["attention-lmo-102-mention"]);
  });

  it("filters presentation-only people by text and team", () => {
    expect(
      selectPeople({ text: "reliability", teams: ["Infrastructure"] }).map(
        (person) => person.id
      )
    ).toEqual(["person-katherine"]);
  });
});

describe("searchGlobalFixtures", () => {
  it("ranks title matches before keyword and description matches", () => {
    expect(
      searchGlobalFixtures({ text: "projection", limit: 3 }).map(
        (entry) => entry.id
      )
    ).toEqual(["lmo-101", "document-projection-model"]);
  });

  it("combines namespace scope and result-kind filters", () => {
    const results = searchGlobalFixtures({
      text: "notification",
      scope: {
        type: "namespace",
        namespaceId: "namespace-operations",
      },
      kinds: ["work-item", "document"],
    });

    expect(results.map((entry) => entry.id)).toEqual([
      "document-incident-review",
      "ops-301",
    ]);
    expect(results.every((entry) => entry.dataSource === "mock")).toBe(true);
  });

  it("returns no results for an empty query or zero limit", () => {
    expect(searchGlobalFixtures({ text: " " })).toEqual([]);
    expect(searchGlobalFixtures({ text: "work", limit: 0 })).toEqual([]);
  });
});
