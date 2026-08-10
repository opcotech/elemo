import type { QuickCreateType } from "@/components/quick-create/types";
import { QUICK_CREATE_EVENT } from "@/components/quick-create/types";

export function openQuickCreate(type: QuickCreateType = "todo") {
  if (typeof window !== "undefined") {
    window.dispatchEvent(
      new CustomEvent<QuickCreateType>(QUICK_CREATE_EVENT, { detail: type })
    );
  }
}
