const HTTP_PROTOCOL = /^https?:\/\//i;
const MAX_ISSUE_LINK_LABEL_LENGTH = 120;

export function parseIssueLinkUrl(
  value: string
): { ok: true; url: string } | { ok: false; error: string } {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return { ok: false, error: "Enter a URL" };
  }

  const candidate = HTTP_PROTOCOL.test(trimmed)
    ? trimmed
    : `https://${trimmed}`;

  let parsed: URL;
  try {
    parsed = new URL(candidate);
  } catch {
    return { ok: false, error: "Enter a valid URL" };
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return { ok: false, error: "Enter a valid URL" };
  }
  if (!parsed.hostname) {
    return { ok: false, error: "Enter a valid URL" };
  }

  return { ok: true, url: parsed.href };
}

export function parseIssueLink(
  url: string,
  label: string
): { ok: true; url: string; label: string } | { ok: false; error: string } {
  const parsed = parseIssueLinkUrl(url);
  if (!parsed.ok) {
    return parsed;
  }

  const trimmedLabel = label.trim();
  if (trimmedLabel.length === 0) {
    return { ok: false, error: "Enter a label" };
  }
  if (trimmedLabel.length > MAX_ISSUE_LINK_LABEL_LENGTH) {
    return { ok: false, error: "Label is too long" };
  }

  return { ok: true, url: parsed.url, label: trimmedLabel };
}

export function issueLinkHostname(url: string): string | null {
  try {
    const hostname = new URL(url).hostname;
    return hostname.length > 0 ? hostname : null;
  } catch {
    return null;
  }
}

export function issueLinkFaviconSrc(hostname: string): string {
  return `https://www.google.com/s2/favicons?domain=${encodeURIComponent(hostname)}&sz=32`;
}
