import { describe, expect, it } from "vitest";

import { removableInputGroupDangerClassName } from "./removable-input-group";

describe("RemovableInputGroup", () => {
  it("applies sidebar-style destructive hover to the whole group", () => {
    expect(removableInputGroupDangerClassName).toContain(
      "has-[[data-slot=input-group-remove]:hover]:bg-destructive/10"
    );
    expect(removableInputGroupDangerClassName).toContain(
      "has-[[data-slot=input-group-remove]:hover]:hover:bg-destructive/10"
    );
    expect(removableInputGroupDangerClassName).toContain(
      "has-[[data-slot=input-group-remove]:hover]:text-destructive"
    );
    expect(removableInputGroupDangerClassName).toContain(
      "has-[[data-slot=input-group-remove]:hover]:ring-destructive/20"
    );
    expect(removableInputGroupDangerClassName).toContain(
      "has-[[data-slot=input-group-remove]:hover]:**:data-[slot=dropdown-menu-trigger]:text-destructive"
    );
  });
});
