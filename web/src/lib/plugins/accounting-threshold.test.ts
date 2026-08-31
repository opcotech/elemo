import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import {
  progressIndicatorClass,
  thresholdReached,
} from "../../../../plugins/accounting/frontend/src/nodes";

const reportSource = readFileSync(
  resolve(
    dirname(fileURLToPath(import.meta.url)),
    "../../../../plugins/accounting/frontend/src/report.tsx"
  ),
  "utf8"
);

describe("accounting threshold progress", () => {
  it("stays primary below the threshold", () => {
    expect(thresholdReached(79, 80)).toBe(false);
    expect(progressIndicatorClass(false)).toBeUndefined();
  });

  it("warns at and above the threshold", () => {
    expect(thresholdReached(80, 80)).toBe(true);
    expect(thresholdReached(81, 80)).toBe(true);
    expect(progressIndicatorClass(true)).toBe("bg-warning");
  });

  it("uses compact avatars on the person filter", () => {
    expect(reportSource).toContain("avatarSrc");
    expect(reportSource).toContain("avatarFallback");
  });
});
