import * as PluginSDK from "@elemo/plugin-sdk";
import * as PluginUI from "@elemo/plugin-ui";
import * as React from "react";
import * as jsxRuntime from "react/jsx-runtime";
import * as ReactDOM from "react-dom";

import * as Lucide from "./lucide-runtime";

import { showErrorToast, showSuccessToast } from "@/lib/toast";

declare global {
  interface Window {
    __ELEMO_PLUGIN_RUNTIME__?: {
      React: typeof React;
      ReactDOM: typeof ReactDOM;
      jsxRuntime: typeof jsxRuntime;
      PluginSDK: typeof PluginSDK;
      PluginUI: typeof PluginUI;
      Lucide: typeof Lucide;
    };
  }
}

export const PLUGIN_IMPORT_MAP = {
  imports: {
    react: "/plugin-runtime/react.js",
    "react-dom": "/plugin-runtime/react-dom.js",
    "react/jsx-runtime": "/plugin-runtime/jsx-runtime.js",
    "react/jsx-dev-runtime": "/plugin-runtime/jsx-runtime.js",
    "@elemo/plugin-sdk": "/plugin-runtime/plugin-sdk.js",
    "@elemo/plugin-ui": "/plugin-runtime/plugin-ui.js",
    "lucide-react": "/plugin-runtime/lucide-react.js",
  },
};

const HOST_PACKAGES: Record<string, string> = {
  react: "__rt.React",
  "react-dom": "__rt.ReactDOM",
  "react/jsx-runtime": "__rt.jsxRuntime",
  "react/jsx-dev-runtime": "__rt.jsxRuntime",
  "@elemo/plugin-sdk": "__rt.PluginSDK",
  "@elemo/plugin-ui": "__rt.PluginUI",
  "lucide-react": "__rt.Lucide",
};

const IMPORT_RE =
  /^import\s+([\s\S]*?)\s+from\s+["']([^"']+)["'];?[ \t]*\r?\n?/gm;

function bindingForImportClause(clause: string, runtime: string): string {
  const trimmed = clause.trim();
  const defaultAndNamed = trimmed.match(
    /^([A-Za-z_$][\w$]*)\s*,\s*(\{[\s\S]*\})$/
  );
  if (defaultAndNamed) {
    return `const ${defaultAndNamed[1]} = ${runtime}.default ?? ${runtime};\nconst ${defaultAndNamed[2]} = ${runtime};`;
  }
  if (trimmed.startsWith("* as ")) {
    return `const ${trimmed.slice(5).trim()} = ${runtime};`;
  }
  if (trimmed.startsWith("{")) {
    return `const ${trimmed} = ${runtime};`;
  }
  return `const ${trimmed} = ${runtime}.default ?? ${runtime};`;
}

/**
 * Rewrite host-provided bare imports to the in-page runtime. Vite/TanStack
 * register module scripts before an import map can take effect, so plugin
 * ESM cannot resolve `react` / `@elemo/plugin-sdk` natively.
 */
export function bindPluginRuntimeImports(source: string): string {
  const bindings: string[] = [];
  const stripped = source.replace(
    IMPORT_RE,
    (_full, clause: string, spec: string) => {
      const runtime = HOST_PACKAGES[spec];
      if (!runtime) {
        throw new Error(
          `Plugin import of "${spec}" is not provided by the host`
        );
      }
      bindings.push(bindingForImportClause(clause, runtime));
      return "";
    }
  );
  return [
    "const __rt = globalThis.__ELEMO_PLUGIN_RUNTIME__;",
    'if (!__rt) throw new Error("Elemo plugin runtime is not initialized");',
    ...bindings,
    stripped,
  ].join("\n");
}

export function ensurePluginRuntime() {
  if (typeof window === "undefined") {
    return;
  }
  window.__ELEMO_PLUGIN_RUNTIME__ = {
    React,
    ReactDOM,
    jsxRuntime,
    PluginSDK,
    PluginUI: Object.assign({}, PluginUI, {
      showErrorToast,
      showSuccessToast,
    }) as typeof PluginUI,
    Lucide,
  };
}
