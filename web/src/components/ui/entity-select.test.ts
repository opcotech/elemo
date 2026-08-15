import { describe, expect, it } from "vitest";

import { optionCommandValue } from "./entity-select";

describe("optionCommandValue", () => {
  it("includes the option value so duplicate titles stay distinct", () => {
    expect(
      optionCommandValue({
        value: "user-1",
        title: "Ada Lovelace",
        description: "Engineer",
      })
    ).toBe("Ada Lovelace Engineer user-1");
    expect(
      optionCommandValue({
        value: "user-2",
        title: "Ada Lovelace",
        description: "Engineer",
      })
    ).toBe("Ada Lovelace Engineer user-2");
  });
});
