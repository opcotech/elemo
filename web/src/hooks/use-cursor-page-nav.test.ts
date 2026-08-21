import { parseHTML } from "linkedom";
import { act, createElement, useEffect } from "react";
import { createRoot } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useCursorPageNav } from "./use-cursor-page-nav";

type Nav = ReturnType<typeof useCursorPageNav>;

function Probe({
  resetKey,
  onNav,
}: {
  resetKey?: string;
  onNav: (nav: Nav) => void;
}) {
  const nav = useCursorPageNav({ resetKey });
  useEffect(() => {
    onNav(nav);
  }, [nav, onNav]);
  return null;
}

describe("useCursorPageNav", () => {
  let root: ReturnType<typeof createRoot>;
  let latest: Nav | undefined;

  beforeEach(() => {
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
  });

  function render(resetKey?: string) {
    act(() => {
      root.render(
        createElement(Probe, {
          resetKey,
          onNav: (nav) => {
            latest = nav;
          },
        })
      );
    });
  }

  it("starts on the first page", () => {
    render();
    expect(latest?.pageToken).toBeUndefined();
    expect(latest?.canGoPrevious).toBe(false);
  });

  it("advances and rewinds the previous-token stack", () => {
    render();

    act(() => {
      latest?.goNext("cursor-2");
    });
    expect(latest?.pageToken).toBe("cursor-2");
    expect(latest?.canGoPrevious).toBe(true);

    act(() => {
      latest?.goNext("cursor-3");
    });
    expect(latest?.pageToken).toBe("cursor-3");

    act(() => {
      latest?.goPrevious();
    });
    expect(latest?.pageToken).toBe("cursor-2");

    act(() => {
      latest?.goPrevious();
    });
    expect(latest?.pageToken).toBeUndefined();
    expect(latest?.canGoPrevious).toBe(false);
  });

  it("resets to the first page when resetKey changes", () => {
    render("q:alpha");
    act(() => {
      latest?.goNext("cursor-2");
    });
    expect(latest?.pageToken).toBe("cursor-2");

    render("q:beta");
    expect(latest?.pageToken).toBeUndefined();
    expect(latest?.canGoPrevious).toBe(false);
  });
});
