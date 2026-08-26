import { z } from "zod";

const SLUG_MIN_LENGTH = 3;
const SLUG_MAX_LENGTH = 50;
const PROJECT_KEY_MIN = 2;
const PROJECT_KEY_MAX = 6;

/** Canonical slug: lowercase ASCII kebab-case. */
const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

/** xid string encoding: 20 characters from Crockford base32 without checksum letters. */
const XID_PATTERN = /^[0-9a-v]{20}$/;

const ISSUE_KEY_PATTERN = /^[A-Z]{2,6}-\d+$/;

export const RESERVED_ORGANIZATION_SLUGS = new Set(["join", "new"]);
export const RESERVED_NAMESPACE_SLUGS = new Set(["new"]);
export const RESERVED_PROJECT_KEY = "NEW";

export type SlugKind = "organization" | "namespace";

export function isXidShaped(value: string): boolean {
  return XID_PATTERN.test(value);
}

export function isCanonicalSlug(value: string): boolean {
  if (value !== value.trim()) {
    return false;
  }
  if (value.length < SLUG_MIN_LENGTH || value.length > SLUG_MAX_LENGTH) {
    return false;
  }
  if (!SLUG_PATTERN.test(value)) {
    return false;
  }
  return !isXidShaped(value);
}

export function isCanonicalOrganizationSlug(value: string): boolean {
  return isCanonicalSlug(value) && !RESERVED_ORGANIZATION_SLUGS.has(value);
}

export function isCanonicalNamespaceSlug(value: string): boolean {
  return isCanonicalSlug(value) && !RESERVED_NAMESPACE_SLUGS.has(value);
}

export function isCanonicalProjectKey(value: string): boolean {
  if (value !== value.toUpperCase()) {
    return false;
  }
  if (value.length < PROJECT_KEY_MIN || value.length > PROJECT_KEY_MAX) {
    return false;
  }
  if (!/^[A-Z]+$/.test(value)) {
    return false;
  }
  return value !== RESERVED_PROJECT_KEY;
}

export function isCanonicalIssueKey(value: string): boolean {
  return ISSUE_KEY_PATTERN.test(value);
}

export function organizationSlugMessage(value: string): string | undefined {
  return slugFieldMessage(value, "organization");
}

export function namespaceSlugMessage(value: string): string | undefined {
  return slugFieldMessage(value, "namespace");
}

export const organizationSlugFormSchema = z
  .string()
  .superRefine((value, ctx) => {
    const message = organizationSlugMessage(value);
    if (message) {
      ctx.addIssue({ code: "custom", message });
    }
  });

export const namespaceSlugFormSchema = z.string().superRefine((value, ctx) => {
  const message = namespaceSlugMessage(value);
  if (message) {
    ctx.addIssue({ code: "custom", message });
  }
});

function slugFieldMessage(value: string, kind: SlugKind): string | undefined {
  if (!value) {
    return "Slug is required";
  }
  if (value !== value.trim()) {
    return "Slug cannot include leading or trailing spaces";
  }
  if (value.length < SLUG_MIN_LENGTH || value.length > SLUG_MAX_LENGTH) {
    return `Slug must be ${SLUG_MIN_LENGTH}–${SLUG_MAX_LENGTH} characters`;
  }
  if (!SLUG_PATTERN.test(value)) {
    return "Use lowercase letters, numbers, and single hyphens";
  }
  if (isXidShaped(value)) {
    return "Slug cannot look like an identifier";
  }
  if (kind === "organization" && RESERVED_ORGANIZATION_SLUGS.has(value)) {
    return "This slug is reserved";
  }
  if (kind === "namespace" && RESERVED_NAMESPACE_SLUGS.has(value)) {
    return "This slug is reserved";
  }
  return undefined;
}

/**
 * Derive a canonical slug from a display name. Already-canonical input is kept.
 * Invalid input may yield an empty string.
 */
export function suggestSlug(name: string): string {
  let decoded = name;
  try {
    decoded = decodeURIComponent(name);
  } catch {
    decoded = name;
  }

  let suggested = "";
  let lastHyphen = true;
  for (const character of decoded.toLowerCase()) {
    const code = character.codePointAt(0) ?? 0;
    const isAlphaNum =
      (code >= 97 && code <= 122) || (code >= 48 && code <= 57);
    const isSeparator =
      character === "-" ||
      character === "_" ||
      character === " " ||
      isUnicodeSpace(character);

    if (isAlphaNum) {
      suggested += character;
      lastHyphen = false;
      continue;
    }
    if (isSeparator && !lastHyphen && suggested.length > 0) {
      suggested += "-";
      lastHyphen = true;
    }
  }

  suggested = suggested.replace(/^-+|-+$/g, "");
  return isCanonicalSlug(suggested) ? suggested : "";
}

function isUnicodeSpace(character: string): boolean {
  return /^\s$/u.test(character);
}
