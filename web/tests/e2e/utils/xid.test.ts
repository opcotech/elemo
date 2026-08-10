import { describe, expect, it } from "vitest";

import { generateXid } from "./xid";

describe("generateXid", () => {
  it("returns a 20-character base32hex xid", () => {
    const id = generateXid();
    expect(id).toMatch(/^[0-9a-v]{20}$/);
  });

  it("returns unique values across rapid calls", () => {
    const ids = new Set(Array.from({ length: 100 }, () => generateXid()));
    expect(ids.size).toBe(100);
  });
});
