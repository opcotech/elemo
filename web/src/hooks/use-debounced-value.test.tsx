import { parseHTML } from "linkedom";
import { act, createElement, useEffect } from "react";
import { createRoot } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useDebouncedValue } from "./use-debounced-value";

function Probe({
  value,
  delayMs,
  onValue,
}: {
  value: string;
  delayMs: number;
  onValue: (value: string) => void;
}) {
  const debounced = useDebouncedValue(value, delayMs);
  useEffect(() => {
    onValue(debounced);
  }, [debounced, onValue]);
  return null;
}

describe("useDebouncedValue", () => {
  let root: ReturnType<typeof createRoot>;
  let latest: string | undefined;

  beforeEach(() => {
    vi.useFakeTimers();
    const { window } = parseHTML(
      "<!doctype html><html><body><div id='root'></div></body></html>"
    );
    vi.stubGlobal("window", window);
    vi.stubGlobal("document", window.document);
    const container = window.document.getElementById("root");
    if (!container) {
      throw new Error("Expected a root container");
    }
    root = createRoot(container);
    latest = undefined;
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  function render(value: string, delayMs = 300) {
    act(() => {
      root.render(
        createElement(Probe, {
          value,
          delayMs,
          onValue: (next) => {
            latest = next;
          },
        })
      );
    });
  }

  it("keeps the previous value until the delay elapses", () => {
    render("");
    expect(latest).toBe("");

    render("projection");
    expect(latest).toBe("");

    act(() => {
      vi.advanceTimersByTime(299);
    });
    expect(latest).toBe("");

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(latest).toBe("projection");
  });

  it("emits only the last value after rapid changes", () => {
    render("");
    render("p");
    render("pr");
    render("projection");
    expect(latest).toBe("");

    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(latest).toBe("projection");
  });
});
