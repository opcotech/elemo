import type { PluginGraphNode } from "@elemo/plugin-sdk";

export function asNodes(value: unknown): PluginGraphNode[] {
  return Array.isArray(value) ? (value as PluginGraphNode[]) : [];
}

export function propString(node: PluginGraphNode, key: string): string {
  const value = node.properties?.[key];
  return value == null ? "" : String(value);
}

export function propNumber(node: PluginGraphNode, key: string): number {
  const value = node.properties?.[key];
  const n = Number(value ?? 0);
  return Number.isFinite(n) ? n : 0;
}

export function accountLabel(node: PluginGraphNode): string {
  const code = propString(node, "code");
  const name = propString(node, "name");
  return [code, name].filter(Boolean).join(" ").trim() || node.id;
}

export function budgetLabel(node: PluginGraphNode): string {
  return propString(node, "name") || "Budget";
}

export function budgetThreshold(node: PluginGraphNode): number {
  const value = propNumber(node, "threshold");
  return value > 0 ? Math.min(100, Math.round(value)) : 80;
}

export function utilizationPercent(used: number, seconds: number): number {
  if (seconds <= 0) {
    return used > 0 ? 100 : 0;
  }
  return Math.round((used / seconds) * 100);
}

export function thresholdReached(percent: number, threshold: number): boolean {
  return percent >= threshold;
}

export function progressIndicatorClass(reached: boolean): string | undefined {
  return reached ? "bg-warning" : undefined;
}
