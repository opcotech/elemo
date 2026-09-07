import type { ReactNode } from "react";

export function PageHeader({
  title,
  description,
  eyebrow,
  leading,
  meta,
  actions,
}: {
  title: ReactNode;
  description?: ReactNode;
  eyebrow?: ReactNode;
  leading?: ReactNode;
  meta?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
      {leading}
      <div className="min-w-0 flex-1">
        {eyebrow && (
          <div className="mb-1 flex items-center gap-1 font-medium text-muted-foreground text-xs uppercase tracking-wide">
            {eyebrow}
          </div>
        )}
        <h1 className="text-balance font-semibold text-2xl tracking-tight sm:text-3xl">
          {title}
        </h1>
        {description && (
          <div className="mt-1 max-w-3xl text-pretty text-muted-foreground text-sm leading-6">
            {description}
          </div>
        )}
        {meta && <div className="mt-3">{meta}</div>}
      </div>
      {actions && (
        <div className="flex shrink-0 items-center gap-2">{actions}</div>
      )}
    </div>
  );
}
