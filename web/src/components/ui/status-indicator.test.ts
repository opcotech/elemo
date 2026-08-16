import { describe, expect, it } from "vitest";

import { resolveStatusTone } from "./status-indicator";

describe("resolveStatusTone", () => {
  it("maps open issues to the info tone", () => {
    expect(resolveStatusTone("open")).toBe("info");
    expect(resolveStatusTone("Open")).toBe("info");
    expect(resolveStatusTone("backlog")).toBe("info");
  });

  it("keeps existing status tones", () => {
    expect(resolveStatusTone("blocked")).toBe("danger");
    expect(resolveStatusTone("done")).toBe("success");
    expect(resolveStatusTone("in progress")).toBe("primary");
    expect(resolveStatusTone("pending")).toBe("warning");
  });
});
