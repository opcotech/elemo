import { Badge } from "@/components/ui/badge";
import { pluralize } from "@/lib/utils";

export function CountBadge({
  count,
  singular,
  plural,
}: {
  count: number;
  singular: string;
  plural: string;
}) {
  return (
    <Badge variant="secondary">
      {count} {pluralize(count, singular, plural)}
    </Badge>
  );
}
