import { mockPeople } from "./fixtures";
import type { Person } from "./types";

export interface AuthenticatedDemoIdentity {
  readonly email?: string | null;
  readonly username?: string | null;
}

/**
 * Connects a real authenticated identity to fixture-owned people. Email is the
 * strongest match, followed by username/handle. Until the people API owns this
 * domain, unmatched users intentionally receive the first documented demo
 * person so personal projections remain deterministic.
 */
export function resolveDemoPerson(
  identity: AuthenticatedDemoIdentity | null | undefined,
  people: readonly Person[] = mockPeople
): Person {
  const fallback = people[0];
  if (!fallback) {
    throw new Error("The demo person resolver requires at least one fixture");
  }

  const email = identity?.email?.trim().toLowerCase();
  const username = identity?.username?.trim().replace(/^@/, "").toLowerCase();

  return (
    people.find((person) => email && person.email.toLowerCase() === email) ??
    people.find(
      (person) => username && person.handle.toLowerCase() === username
    ) ??
    fallback
  );
}
