import { Folder } from "lucide-react";
import { useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
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
import type { NamespaceProject } from "@/lib/api";

function NamespaceProjectsListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Key</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-16" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-5 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

interface NamespaceProjectsListProps {
  projects: NamespaceProject[];
  isLoading: boolean;
  error: unknown;
}

export function NamespaceProjectsList({
  projects,
  isLoading,
  error,
}: NamespaceProjectsListProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredProjects = useMemo(() => {
    if (!searchTerm.trim()) return projects;
    const term = searchTerm.toLowerCase();
    return projects.filter(
      (project) =>
        project.name.toLowerCase().includes(term) ||
        project.key.toLowerCase().includes(term) ||
        (project.description &&
          project.description.toLowerCase().includes(term))
    );
  }, [projects, searchTerm]);

  const emptyState =
    projects.length === 0
      ? {
          icon: <Folder />,
          title: "No projects found",
          description:
            "Projects will appear here when they are added to this namespace.",
        }
      : filteredProjects.length === 0 && searchTerm.trim()
        ? {
            icon: <Folder />,
            title: "No projects found",
            description:
              "No projects match your search criteria. Try adjusting your search.",
          }
        : undefined;

  const shouldShowSearch = projects.length > 0 || searchTerm.trim() !== "";

  return (
    <ListContainer
      title="Projects"
      description="Projects in this namespace."
      isLoading={isLoading}
      error={error}
      emptyState={emptyState}
      searchInput={
        shouldShowSearch ? (
          <SearchInput
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search projects..."
            disabled={isLoading}
          />
        ) : undefined
      }
    >
      {isLoading ? (
        <NamespaceProjectsListSkeleton />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Key</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredProjects.map((project) => (
              <TableRow key={project.id}>
                <TableCell className="font-mono text-sm">
                  {project.key}
                </TableCell>
                <TableCell className="font-medium">{project.name}</TableCell>
                <TableCell>
                  <span className="text-muted-foreground text-sm">
                    {project.description || "—"}
                  </span>
                </TableCell>
                <TableCell>
                  <Badge
                    variant={
                      project.status === "active" ? "success" : "secondary"
                    }
                  >
                    {project.status}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </ListContainer>
  );
}
