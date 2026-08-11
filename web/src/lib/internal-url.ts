export type InternalPath = `/${string}`;

function hasControlOrWhitespace(value: string) {
  return [...value].some((character) => {
    const codePoint = character.codePointAt(0) ?? 0;
    return codePoint <= 0x20 || codePoint === 0x7f;
  });
}

export function isSafeInternalPath(value: unknown): value is InternalPath {
  if (
    typeof value !== "string" ||
    !value.startsWith("/") ||
    value.startsWith("//") ||
    value.includes("\\") ||
    hasControlOrWhitespace(value)
  ) {
    return false;
  }

  try {
    const base = new URL("https://elemo.invalid");
    const parsed = new URL(value, base);
    return parsed.origin === base.origin && parsed.pathname.startsWith("/");
  } catch {
    return false;
  }
}

export function internalPath(value: string): InternalPath {
  if (!isSafeInternalPath(value)) {
    throw new Error("Expected a safe internal path");
  }
  return value;
}
