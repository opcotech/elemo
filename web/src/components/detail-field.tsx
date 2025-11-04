/**
 * Reusable component for displaying a label-value pair in detail views.
 * Can display either a simple string value or custom children content.
 */
export function DetailField({
  label,
  value,
  children,
}: {
  label: string;
  value?: string | null;
  children?: React.ReactNode;
}) {
  return (
    <div>
      <label className="text-muted-foreground text-sm font-medium">
        {label}
      </label>
      {children ? (
        <div className="mt-1 text-sm">{children}</div>
      ) : (
        <p className="mt-1 text-sm">
          {value || <span className="text-muted-foreground">—</span>}
        </p>
      )}
    </div>
  );
}
