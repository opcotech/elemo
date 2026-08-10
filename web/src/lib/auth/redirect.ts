import { z } from "zod";

const FALLBACK_REDIRECT = "/";

export function sanitizeRedirectTarget(
  candidate: string | undefined,
  origin?: string
): string {
  if (!candidate || candidate.includes("\\") || /[\r\n\0]/.test(candidate)) {
    return FALLBACK_REDIRECT;
  }

  const baseOrigin = origin || "https://elemo.invalid";
  try {
    const target = new URL(candidate, baseOrigin);
    if (
      target.origin !== new URL(baseOrigin).origin ||
      !target.pathname.startsWith("/") ||
      target.pathname === "/login"
    ) {
      return FALLBACK_REDIRECT;
    }
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return FALLBACK_REDIRECT;
  }
}

export const safeRedirectSearchSchema = z.object({
  redirect: z
    .string()
    .max(2048)
    .refine(
      (value) =>
        value.startsWith("/") && sanitizeRedirectTarget(value) === value,
      "Invalid redirect target"
    )
    .optional(),
});
