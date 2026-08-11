import { pluralize } from "@/lib/utils";

export function NamespaceEntitySubtitle({
  description,
  organizationName,
  projectCount,
  documentCount,
}: {
  description?: string | null;
  organizationName?: string;
  projectCount: number;
  documentCount: number;
}) {
  const meta = [
    organizationName,
    `${projectCount} ${pluralize(projectCount, "project", "projects")}`,
    `${documentCount} ${pluralize(documentCount, "document", "documents")}`,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <span className="flex min-w-0 flex-col gap-0.5">
      {description ? <span className="truncate">{description}</span> : null}
      <span className="truncate">{meta}</span>
    </span>
  );
}
