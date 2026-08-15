import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import type { PersonAvatarStackPerson } from "@/components/ui/person-avatar-stack";
import type { DataSource } from "@/lib/mock-data";
import { resolveReportedByPerson } from "@/lib/work/resolve-work-people";
import { useOrganizationMembersForNamespace } from "@/lib/work/use-organization-members-for-namespace";

export function IssueReportedBy({
  userId,
  namespaceId,
  people = [],
  dataSource = "api",
}: {
  userId: string;
  namespaceId?: string | null;
  people?: readonly PersonAvatarStackPerson[];
  dataSource?: DataSource;
}) {
  const isApi = dataSource === "api";
  const { members } = useOrganizationMembersForNamespace(
    namespaceId ?? undefined,
    { enabled: isApi && Boolean(namespaceId) }
  );

  if (!userId) {
    return (
      <PersonAvatarStack people={[]} size="sm" showNames emptyLabel="Unknown" />
    );
  }

  const person = resolveReportedByPerson(userId, {
    members: isApi ? members : undefined,
    people,
    useMockFallback: !isApi,
  });

  return <PersonAvatarStack people={[person]} size="sm" showNames />;
}
