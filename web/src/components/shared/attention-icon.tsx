import {
  AlertTriangleIcon,
  CheckCircle2Icon,
  CircleDotIcon,
} from "lucide-react";

export function AttentionIcon({ severity }: { severity: string }) {
  if (severity === "critical") {
    return <AlertTriangleIcon className="size-4 text-destructive" />;
  }
  if (severity === "warning") {
    return <CircleDotIcon className="size-4 text-warning-on-subtle" />;
  }
  return <CheckCircle2Icon className="size-4 text-info" />;
}
