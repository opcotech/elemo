import { describe, expect, it } from "vitest";

import {
  isPluginJavaScriptSource,
  isSafeAssetRel,
  pluginFrontendAssetPath,
} from "./asset-path";

describe("pluginFrontendAssetPath", () => {
  it("builds a relative Go API path for a valid frontend entry", () => {
    expect(
      pluginFrontendAssetPath(
        "com.elemo.timetracking",
        "1.0.0",
        "frontend/index.js"
      )
    ).toBe("/v1/plugins/com.elemo.timetracking/assets/1.0.0/frontend/index.js");
  });

  it("rejects traversal and empty entrypoints", () => {
    expect(
      pluginFrontendAssetPath("com.elemo.timetracking", "1.0.0", "../secret")
    ).toBeUndefined();
    expect(
      pluginFrontendAssetPath(
        "com.elemo.timetracking",
        "1.0.0",
        "foo/../../secret"
      )
    ).toBeUndefined();
    expect(pluginFrontendAssetPath("com.elemo.timetracking", "1.0.0", "")).toBe(
      undefined
    );
    expect(isSafeAssetRel("frontend/index.js")).toBe(true);
    expect(isSafeAssetRel("../x")).toBe(false);
  });
});

describe("isPluginJavaScriptSource", () => {
  it("accepts javascript content types and ESM source", () => {
    expect(isPluginJavaScriptSource("not js", "text/javascript")).toBe(true);
    expect(isPluginJavaScriptSource('import x from "y";\n', "text/plain")).toBe(
      true
    );
    expect(isPluginJavaScriptSource("export default {};\n", "")).toBe(true);
  });

  it("rejects HTML shells", () => {
    expect(
      isPluginJavaScriptSource("<!DOCTYPE html><html></html>", "text/html")
    ).toBe(false);
  });
});
