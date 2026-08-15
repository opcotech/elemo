import type { IssueResolution } from "@/lib/api/types";

export const issueResolutions: readonly IssueResolution[] = [
  "none",
  "fixed",
  "duplicate",
  "won't fix",
  "invalid",
  "incomplete",
  "cannot reproduce",
];

export const issueResolutionLabels: Record<IssueResolution, string> = {
  none: "None",
  fixed: "Fixed",
  duplicate: "Duplicate",
  "won't fix": "Won't fix",
  invalid: "Invalid",
  incomplete: "Incomplete",
  "cannot reproduce": "Cannot reproduce",
};
