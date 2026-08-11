import { BriefcaseBusinessIcon } from "lucide-react";

export function ContextLine({
  namespace,
  project,
}: {
  namespace?: string;
  project?: string;
}) {
  if (!namespace && !project) return null;
  return (
    <div className="text-muted-foreground flex items-center gap-2 text-xs">
      <BriefcaseBusinessIcon className="size-3.5" />
      <span>{[namespace, project].filter(Boolean).join(" / ")}</span>
    </div>
  );
}
