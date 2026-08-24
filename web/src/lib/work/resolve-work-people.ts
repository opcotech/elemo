import type { PersonAvatarStackPerson } from "@/components/ui/person-avatar-stack";
import type { OrganizationMember, PartialUser } from "@/lib/api/types";
import { getPerson } from "@/lib/mock-data";
import type { WorkItem, WorkPerson } from "@/lib/work/model";

export function personDisplayName(
  person: Pick<PartialUser, "id" | "first_name" | "last_name">
): string {
  return `${person.first_name} ${person.last_name}`.trim() || person.id;
}

export function partialUsersFromIds(
  ids: readonly string[],
  existing: readonly PartialUser[] = [],
  catalog: readonly Pick<
    PartialUser,
    "id" | "first_name" | "last_name" | "picture"
  >[] = []
): PartialUser[] {
  const byId = new Map<string, PartialUser>();
  for (const user of catalog) {
    byId.set(user.id, {
      id: user.id,
      first_name: user.first_name,
      last_name: user.last_name,
      picture: user.picture ?? null,
    });
  }
  for (const user of existing) {
    byId.set(user.id, user);
  }
  return ids.map(
    (id) =>
      byId.get(id) ?? {
        id,
        first_name: id,
        last_name: "",
        picture: null,
      }
  );
}

export function partialUserToPerson(
  user: PartialUser
): PersonAvatarStackPerson {
  return {
    id: user.id,
    name: personDisplayName(user),
    picture: user.picture,
  };
}

export function resolveWorkPeople(
  ids: readonly string[],
  members?: readonly OrganizationMember[] | null
): PersonAvatarStackPerson[] {
  const memberById = new Map(
    (members ?? []).map((member) => [member.id, member] as const)
  );

  return ids.map((id) => {
    const member = memberById.get(id);
    if (member) {
      return {
        id,
        name: personDisplayName(member),
        picture: member.picture,
      };
    }

    const mock = getPerson(id);
    if (mock) {
      return {
        id,
        name: mock.displayName,
        picture: mock.avatarUrl,
      };
    }

    return {
      id,
      name: id,
      picture: null,
    };
  });
}

function uniquePeople(
  people: readonly PersonAvatarStackPerson[]
): PersonAvatarStackPerson[] {
  const seen = new Set<string>();
  const unique: PersonAvatarStackPerson[] = [];

  for (const person of people) {
    if (seen.has(person.id)) {
      continue;
    }
    seen.add(person.id);
    unique.push(person);
  }

  return unique;
}

/** Assignees first, then reviewers; de-dupe by id. */
export function resolveWorkAssignmentPeople(
  assigneeIds: readonly string[],
  reviewerIds: readonly string[],
  members?: readonly OrganizationMember[] | null
): PersonAvatarStackPerson[] {
  return uniquePeople([
    ...resolveWorkPeople(assigneeIds, members),
    ...resolveWorkPeople(reviewerIds, members),
  ]);
}

export function workItemPeople(
  people: readonly WorkPerson[] | undefined,
  ids: readonly string[]
): PersonAvatarStackPerson[] {
  if (people && people.length > 0) {
    return people.map((person) => ({
      id: person.id,
      name: person.name,
      picture: person.picture,
    }));
  }
  return resolveWorkPeople(ids);
}

export function workItemAssignmentPeople(
  item: WorkItem
): PersonAvatarStackPerson[] {
  const fromShapes = uniquePeople([
    ...workItemPeople(item.assignees, []),
    ...workItemPeople(item.reviewers, []),
  ]);
  if (fromShapes.length > 0) {
    return fromShapes;
  }
  return resolveWorkAssignmentPeople(item.assigneeIds, item.reviewerIds);
}

/** Selected issue people first, then catalog extras. */
export function mergeWorkPeople(
  selected: readonly PartialUser[],
  catalog?: readonly OrganizationMember[] | null
): PersonAvatarStackPerson[] {
  return uniquePeople([
    ...selected.map(partialUserToPerson),
    ...(catalog ?? []).map(partialUserToPerson),
  ]);
}

/** Resolve a reporter from members, then issue people, then a raw id. */
export function resolveReportedByPerson(
  userId: string,
  options: {
    members?: readonly PartialUser[] | null;
    people?: readonly PersonAvatarStackPerson[] | null;
    useMockFallback?: boolean;
  } = {}
): PersonAvatarStackPerson {
  const member = (options.members ?? []).find((person) => person.id === userId);
  if (member) {
    return partialUserToPerson(member);
  }

  const fromPeople = (options.people ?? []).find(
    (person) => person.id === userId
  );
  if (fromPeople) {
    return fromPeople;
  }

  if (options.useMockFallback) {
    const mock = getPerson(userId);
    if (mock) {
      return {
        id: mock.id,
        name: mock.displayName,
        picture: mock.avatarUrl,
      };
    }
  }

  return {
    id: userId,
    name: userId,
    picture: null,
  };
}
