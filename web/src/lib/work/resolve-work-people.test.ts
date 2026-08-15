import { describe, expect, it } from "vitest";

import {
  partialUsersFromIds,
  resolveReportedByPerson,
} from "./resolve-work-people";

describe("partialUsersFromIds", () => {
  it("prefers existing users, then catalog names, then the raw id", () => {
    expect(
      partialUsersFromIds(
        ["user-1", "user-2", "user-3"],
        [
          {
            id: "user-1",
            first_name: "Existing",
            last_name: "Name",
            picture: null,
          },
        ],
        [
          {
            id: "user-1",
            first_name: "Ada",
            last_name: "Lovelace",
            picture: null,
          },
          {
            id: "user-2",
            first_name: "Grace",
            last_name: "Hopper",
            picture: null,
          },
        ]
      )
    ).toEqual([
      {
        id: "user-1",
        first_name: "Existing",
        last_name: "Name",
        picture: null,
      },
      {
        id: "user-2",
        first_name: "Grace",
        last_name: "Hopper",
        picture: null,
      },
      {
        id: "user-3",
        first_name: "user-3",
        last_name: "",
        picture: null,
      },
    ]);
  });
});

describe("resolveReportedByPerson", () => {
  it("prefers organization members over issue people", () => {
    expect(
      resolveReportedByPerson("user-1", {
        members: [
          {
            id: "user-1",
            first_name: "Ada",
            last_name: "Lovelace",
            picture: "https://example.test/ada.png",
          },
        ],
        people: [{ id: "user-1", name: "Assignee Name", picture: null }],
      })
    ).toEqual({
      id: "user-1",
      name: "Ada Lovelace",
      picture: "https://example.test/ada.png",
    });
  });

  it("uses assignees or reviewers when the reporter is among them", () => {
    expect(
      resolveReportedByPerson("user-2", {
        members: [],
        people: [
          { id: "user-2", name: "Grace Hopper", picture: null },
          { id: "user-3", name: "Katherine Johnson", picture: null },
        ],
      })
    ).toEqual({
      id: "user-2",
      name: "Grace Hopper",
      picture: null,
    });
  });

  it("falls back to the raw id without a mock lookup", () => {
    expect(
      resolveReportedByPerson("user-unknown", {
        members: [],
        people: [{ id: "user-2", name: "Grace Hopper", picture: null }],
      })
    ).toEqual({
      id: "user-unknown",
      name: "user-unknown",
      picture: null,
    });
  });

  it("uses mock people only when requested", () => {
    expect(
      resolveReportedByPerson("person-ada", { useMockFallback: true })
    ).toEqual({
      id: "person-ada",
      name: "Ada Lovelace",
      picture: null,
    });
    expect(resolveReportedByPerson("person-ada")).toEqual({
      id: "person-ada",
      name: "person-ada",
      picture: null,
    });
  });
});
