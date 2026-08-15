import {
  AlertTriangleIcon,
  CheckCircle2Icon,
  CircleDotIcon,
} from "lucide-react";

export function AttentionIcon({ severity }: { severity: string }) {
  if (severity === "critical") {
    return <AlertTriangleIcon className="text-destructive size-4" />;
  }
  if (severity === "warning") {
    return <CircleDotIcon className="text-warning-on-subtle size-4" />;
  }
  return <CheckCircle2Icon className="text-info size-4" />;
}
