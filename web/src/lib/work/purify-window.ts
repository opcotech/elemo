import type { WindowLike } from "dompurify";
import { DOMParser, Node, NodeFilter, parseHTML } from "linkedom";

const instanceFields = new WeakMap<object, Record<string, unknown>>();

/**
 * DOMPurify reads parentNode/ownerDocument/nodeType via prototype getters.
 * linkedom stores those as instance fields, so lookupGetter returns a stub
 * that always yields null and sanitization becomes a no-op.
 */
function defineInstanceFieldAccessor(proto: object, name: string): void {
  Object.defineProperty(proto, name, {
    configurable: true,
    enumerable: true,
    get() {
      return instanceFields.get(this)?.[name] ?? null;
    },
    set(value: unknown) {
      let bag = instanceFields.get(this);
      if (!bag) {
        bag = {};
        instanceFields.set(this, bag);
      }
      bag[name] = value;
    },
  });
}

type LinkedomDocument = {
  implementation?: unknown;
  createNodeIterator: (root: object, whatToShow?: number) => unknown;
  createTreeWalker: (root: object, whatToShow?: number) => unknown;
};

let linkedomPatched = false;

function patchLinkedomForDomPurify(): void {
  if (linkedomPatched) {
    return;
  }
  linkedomPatched = true;

  defineInstanceFieldAccessor(Node.prototype, "parentNode");
  defineInstanceFieldAccessor(Node.prototype, "ownerDocument");
  defineInstanceFieldAccessor(Node.prototype, "nodeType");

  const originalParseFromString = DOMParser.prototype.parseFromString;
  DOMParser.prototype.parseFromString = function (
    this: DOMParser,
    markup: string,
    mimeType: "text/html" | "image/svg+xml" | "text/xml",
    globals?: unknown
  ) {
    let html = markup;
    if (
      mimeType === "text/html" &&
      typeof html === "string" &&
      !/<html[\s>]/i.test(html)
    ) {
      html = `<!doctype html><html><head></head><body>${html}</body></html>`;
    }
    return originalParseFromString.call(this, html, mimeType, globals);
  } as typeof originalParseFromString;
}

function patchDocument(document: LinkedomDocument): void {
  const createHTMLDocument = () => {
    const parsed = parseHTML(
      "<!doctype html><html><head></head><body></body></html>",
      { NodeFilter }
    );
    patchDocument(parsed.document as unknown as LinkedomDocument);
    return parsed.document;
  };

  Object.defineProperty(document, "implementation", {
    configurable: true,
    value: {
      createHTMLDocument,
      createDocument: () => createHTMLDocument(),
    },
  });

  document.createNodeIterator = (root, whatToShow = -1) =>
    document.createTreeWalker(root, whatToShow);
}

/**
 * DOMPurify needs a Window-like root. Use the real browser window when
 * present; otherwise build a lightweight one with linkedom (SSR / Node tests).
 */
export function getPurifyWindow(): WindowLike {
  if (typeof globalThis.window !== "undefined" && globalThis.document) {
    return globalThis.window;
  }

  patchLinkedomForDomPurify();
  const { window, document } = parseHTML(
    "<!doctype html><html><body></body></html>",
    { NodeFilter }
  );
  patchDocument(document as unknown as LinkedomDocument);
  return window as unknown as WindowLike;
}
