import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { bindPluginRuntimeImports } from "./runtime";

const TIME_TRACKING_HEADER = `import { defineElemoPlugin } from "@elemo/plugin-sdk";
import { Button, Card, CardContent, CardHeader, CardTitle, EmptyState } from "@elemo/plugin-ui";
import { useEffect, useState } from "react";
import { jsx, jsxs } from "react/jsx-runtime";
var PLUGIN_ID = "com.elemo.timetracking";
export { src_default as default };
`;

describe("bindPluginRuntimeImports", () => {
  it("rewrites host package imports to the in-page runtime", () => {
    const bound = bindPluginRuntimeImports(TIME_TRACKING_HEADER);
    expect(bound).toContain(
      "const __rt = globalThis.__ELEMO_PLUGIN_RUNTIME__;"
    );
    expect(bound).toContain("const { defineElemoPlugin } = __rt.PluginSDK;");
    expect(bound).toContain(
      "const { Button, Card, CardContent, CardHeader, CardTitle, EmptyState } = __rt.PluginUI;"
    );
    expect(bound).toContain("const { useEffect, useState } = __rt.React;");
    expect(bound).toContain("const { jsx, jsxs } = __rt.jsxRuntime;");
    expect(bound).toContain('var PLUGIN_ID = "com.elemo.timetracking";');
    expect(bound).toContain("export { src_default as default };");
    expect(bound).not.toMatch(/^import /m);
  });

  it("rejects imports the host does not provide", () => {
    expect(() => bindPluginRuntimeImports('import fs from "fs";\n')).toThrow(
      /is not provided by the host/
    );
  });

  it("rewrites lucide-react imports to the in-page runtime", () => {
    const bound = bindPluginRuntimeImports(
      'import { ArrowUpRight, Clock, Pencil } from "lucide-react";\n'
    );
    expect(bound).toContain(
      "const { ArrowUpRight, Clock, Pencil } = __rt.Lucide;"
    );
  });

  it("rewrites the built Time Tracking frontend module", () => {
    const source = readFileSync(
      resolve(
        dirname(fileURLToPath(import.meta.url)),
        "../../../../plugins/timetracking/frontend/index.js"
      ),
      "utf8"
    );
    const bound = bindPluginRuntimeImports(source);
    expect(bound).not.toMatch(/^import /m);
    expect(bound).toContain("const { defineElemoPlugin } = __rt.PluginSDK;");
    expect(bound).toContain("export { src_default as default };");
  });

  it("rewrites accounting lucide warning icons to the host runtime", () => {
    const bound = bindPluginRuntimeImports(
      'import { CircleAlert, TriangleAlert, Wallet } from "lucide-react";\n'
    );
    expect(bound).toContain(
      "const { CircleAlert, TriangleAlert, Wallet } = __rt.Lucide;"
    );
  });
});
