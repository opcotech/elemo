import { CircleDot } from "lucide-react";
import { useMemo, useState } from "react";

import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

function ProjectIssuesListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>ID</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-48" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

interface ProjectIssuesListProps {
  issues: string[];
  isLoading: boolean;
  error: unknown;
}

export function ProjectIssuesList({
  issues,
  isLoading,
  error,
}: ProjectIssuesListProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredIssues = useMemo(() => {
    if (!searchTerm.trim()) return issues;
    const term = searchTerm.toLowerCase();
    return issues.filter((issueId) => issueId.toLowerCase().includes(term));
  }, [issues, searchTerm]);

  const emptyState =
    issues.length === 0
      ? {
          icon: <CircleDot />,
          title: "No issues found",
          description:
            "Issues will appear here when they are added to this project.",
        }
      : filteredIssues.length === 0 && searchTerm.trim()
        ? {
            icon: <CircleDot />,
            title: "No issues found",
            description:
              "No issues match your search criteria. Try adjusting your search.",
          }
        : undefined;

  const shouldShowSearch = issues.length > 0 || searchTerm.trim() !== "";

  return (
    <ListContainer
      data-section="project-issues"
      title="Issues"
      description="Issues in this project."
      isLoading={isLoading}
      error={error}
      emptyState={emptyState}
      searchInput={
        shouldShowSearch ? (
          <SearchInput
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search issues..."
            disabled={isLoading}
          />
        ) : undefined
      }
    >
      {isLoading ? (
        <ProjectIssuesListSkeleton />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredIssues.map((issueId) => (
              <TableRow key={issueId}>
                <TableCell className="font-mono text-sm">{issueId}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </ListContainer>
  );
}
