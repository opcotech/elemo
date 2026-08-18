import { useQuery } from "@tanstack/react-query";
import { FolderIcon, Link2Icon } from "lucide-react";
import { Fragment } from "react";
import type { ReactNode } from "react";

import { EntityIcon } from "@/components/shared/entity-link";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { InternalLink } from "@/components/ui/internal-link";
import type { Document, DocumentRelation } from "@/lib/api/types";
import {
  documentLibraryPageHref,
  documentLibrarySearchParams,
  folderPathQuery,
} from "@/lib/documents/library";
import type { InternalPath } from "@/lib/internal-url";
import { internalPath } from "@/lib/internal-url";

function documentRelationHref(
  library: Document["library"],
  relation: DocumentRelation
): InternalPath | null {
  if (library.type !== "Namespace") {
    return null;
  }
  if (relation.type === "Project") {
    return internalPath(`/namespaces/${library.id}/projects/${relation.id}`);
  }
  return internalPath(`/work/${library.id}/${relation.name}`);
}

function MetaRow({
  label,
  icon,
  children,
}: {
  label: string;
  icon: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="col-span-2 grid grid-cols-subgrid items-center gap-x-4 py-2.5">
      <dt className="text-muted-foreground flex min-h-5 items-center gap-2 text-xs font-medium whitespace-nowrap">
        {icon}
        {label}
      </dt>
      <dd className="flex min-h-5 min-w-0 items-center">{children}</dd>
    </div>
  );
}

export function DocumentLocation({ document }: { document: Document }) {
  const folderId = document.folder?.id;
  const { data: folderPath } = useQuery({
    ...folderPathQuery(folderId ?? ""),
    enabled: Boolean(folderId),
  });
  const documentsHref = documentLibraryPageHref(document.library);
  const crumbs = folderPath ?? [];

  return (
    <div
      className="border-border/80 bg-card overflow-hidden rounded-lg border"
      data-section="document-location"
    >
      <dl className="divide-border/60 grid grid-cols-[7.5rem_minmax(0,1fr)] divide-y px-4">
        <MetaRow
          label="Location"
          icon={<FolderIcon className="size-3.5" aria-hidden />}
        >
          <Breadcrumb>
            <BreadcrumbList className="gap-1 text-xs leading-5">
              <BreadcrumbItem>
                <BreadcrumbLink
                  render={<InternalLink to={documentsHref} search={{}} />}
                >
                  {document.library.name}
                </BreadcrumbLink>
              </BreadcrumbItem>
              {crumbs.map((folder) => (
                <Fragment key={folder.id}>
                  <BreadcrumbSeparator className="flex items-center [&>svg]:size-3" />
                  <BreadcrumbItem>
                    <BreadcrumbLink
                      render={
                        <InternalLink
                          to={documentsHref}
                          search={documentLibrarySearchParams({
                            folderId: folder.id,
                          })}
                        />
                      }
                    >
                      {folder.name}
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                </Fragment>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
        </MetaRow>
        {document.relations.length > 0 ? (
          <MetaRow
            label="Related"
            icon={<Link2Icon className="size-3.5" aria-hidden />}
          >
            <div className="flex min-w-0 flex-wrap items-center gap-1.5">
              {document.relations.map((relation) => {
                const href = documentRelationHref(document.library, relation);
                return (
                  <Badge
                    key={`${relation.type}-${relation.id}`}
                    variant="secondary"
                    className="max-w-44 gap-1 pl-1.5"
                    title={relation.type}
                    render={href ? <InternalLink to={href} /> : undefined}
                  >
                    <EntityIcon
                      type={
                        relation.type === "Project" ? "project" : "work-item"
                      }
                      className="size-3"
                    />
                    <span className="truncate">{relation.name}</span>
                  </Badge>
                );
              })}
            </div>
          </MetaRow>
        ) : null}
      </dl>
    </div>
  );
}
