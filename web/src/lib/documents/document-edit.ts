import { z } from "zod";

const titleSchema = z.string().min(3).max(120);
const excerptSchema = z.string().min(10).max(500);

export const DOCUMENT_EXCERPT_MIN_LENGTH = 10;
export const DOCUMENT_EXCERPT_AUTO_LENGTH = 100;

/** Build a document excerpt from body text when the field is left empty. */
export function excerptFromContent(
  content: string,
  maxLength = DOCUMENT_EXCERPT_AUTO_LENGTH
): string | null {
  const collapsed = content.replace(/\s+/g, " ").trim();
  if (!collapsed) {
    return null;
  }

  const excerpt =
    collapsed.length <= maxLength
      ? collapsed
      : collapsed.slice(0, maxLength).trimEnd();
  if (excerpt.length < DOCUMENT_EXCERPT_MIN_LENGTH) {
    return null;
  }

  return excerpt;
}

/** Validate a document title. */
export function parseDocumentTitle(
  value: string
): { ok: true; title: string } | { ok: false; error: string } {
  const trimmed = value.trim();
  const result = titleSchema.safeParse(trimmed);
  if (!result.success) {
    return {
      ok: false,
      error: result.error.issues[0]?.message ?? "Invalid title",
    };
  }
  return { ok: true, title: result.data };
}

/** Validate a document excerpt. Empty clears the excerpt. */
export function parseDocumentExcerpt(
  value: string
): { ok: true; excerpt: string | null } | { ok: false; error: string } {
  const trimmed = value.trim();
  if (!trimmed) {
    return { ok: true, excerpt: null };
  }

  const result = excerptSchema.safeParse(trimmed);
  if (!result.success) {
    return {
      ok: false,
      error: result.error.issues[0]?.message ?? "Invalid excerpt",
    };
  }
  return { ok: true, excerpt: result.data };
}

/**
 * Resolve the excerpt to persist. An empty field uses the first 100
 * characters of the document body; a filled field is validated as-is.
 */
export function resolveDocumentExcerpt(
  excerpt: string,
  content: string
): { ok: true; excerpt: string | null } | { ok: false; error: string } {
  if (!excerpt.trim()) {
    return { ok: true, excerpt: excerptFromContent(content) };
  }
  return parseDocumentExcerpt(excerpt);
}

/**
 * Normalize document body for PATCH. Empty markdown is allowed.
 */
export function parseDocumentContent(
  markdown: string
): { ok: true; content: string } | { ok: false; error: string } {
  return { ok: true, content: markdown.trim() };
}
