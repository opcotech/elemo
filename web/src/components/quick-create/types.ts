export const QUICK_CREATE_EVENT = "elemo:quick-create";

export type QuickCreateType = "todo" | "work" | "document";

export function isQuickCreateType(value: unknown): value is QuickCreateType {
  return value === "todo" || value === "work" || value === "document";
}

export function isTypingTarget(target: EventTarget | null) {
  return (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    (target instanceof HTMLElement && target.isContentEditable)
  );
}

export interface QuickCreateKindProps {
  onCancel: () => void;
  onComplete: () => void;
}
