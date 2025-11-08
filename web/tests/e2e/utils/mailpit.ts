/**
 * Helper functions for interacting with Mailpit API to fetch emails and extract tokens.
 * Mailpit runs locally on port 8025 for testing purposes.
 */

const MAILPIT_API_URL =
  process.env.MAILPIT_API_URL || "http://localhost:8025/api/v1";

interface MailpitMessage {
  ID: string;
  Created: string;
  From: {
    Name: string;
    Address: string;
  };
  To: Array<{
    Name: string;
    Address: string;
  }>;
  Subject: string;
  Size: number;
}

interface MailpitMessageDetail extends MailpitMessage {
  HTML: string;
  Text: string;
  Date?: string;
}

/**
 * Fetches the latest email message for a given recipient email address.
 */
export async function getLatestEmailForRecipient(
  recipientEmail: string
): Promise<MailpitMessageDetail | null> {
  try {
    const response = await fetch(`${MAILPIT_API_URL}/messages?limit=100`);
    if (!response.ok) {
      throw new Error(`Mailpit API error: ${response.statusText}`);
    }

    const data = await response.json();
    const messages: Array<MailpitMessage> = data.messages || [];
    const message = messages
      .filter((msg) =>
        msg.To.some(
          (to) => to.Address.toLowerCase() === recipientEmail.toLowerCase()
        )
      )
      .sort(
        (a, b) => new Date(b.Created).getTime() - new Date(a.Created).getTime()
      )[0];
    const detailResponse = await fetch(
      `${MAILPIT_API_URL}/message/${message.ID}`
    );
    if (!detailResponse.ok) {
      throw new Error(`Mailpit API error: ${detailResponse.statusText}`);
    }

    const messageDetail: MailpitMessageDetail = await detailResponse.json();
    return messageDetail;
  } catch (error) {
    console.error("Error fetching email from Mailpit:", error);
    return null;
  }
}

/**
 * Extracts invitation token from an organization invitation email.
 * The token is typically in the invitation URL in the format:
 * /organizations/join?organization={orgId}&token={token}
 * Tokens can be base64-encoded and may contain =, /, + characters
 */
export async function getInvitationTokenFromEmail(
  recipientEmail: string
): Promise<string | null> {
  const email = await getLatestEmailForRecipient(recipientEmail);
  if (!email) {
    return null;
  }
  const htmlContent = email.HTML || email.Text || "";
  const decodedContent = htmlContent
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'");
  const tokenMatch = decodedContent.match(/token=([A-Za-z0-9+/=_-]+)/i);
  if (tokenMatch && tokenMatch[1]) {
    try {
      return decodeURIComponent(tokenMatch[1]);
    } catch {
      return tokenMatch[1];
    }
  }
  const urlMatch = decodedContent.match(
    /\/organizations\/join[^"'\s]*token=([A-Za-z0-9+/=_-]+)/i
  );
  if (urlMatch && urlMatch[1]) {
    try {
      return decodeURIComponent(urlMatch[1]);
    } catch {
      return urlMatch[1];
    }
  }

  return null;
}

/**
 * Waits for an email to arrive for a recipient, with timeout.
 */
export async function waitForEmail(
  recipientEmail: string,
  timeoutMs: number = 10000,
  checkIntervalMs: number = 500
): Promise<MailpitMessageDetail | null> {
  const startTime = Date.now();

  while (Date.now() - startTime < timeoutMs) {
    const email = await getLatestEmailForRecipient(recipientEmail);
    if (email) {
      const emailTimeStr = email.Created || (email as any).Date;
      if (emailTimeStr) {
        const emailTime = new Date(emailTimeStr).getTime();
        const now = Date.now();
        if (now - emailTime < 60000) {
          return email;
        }
      } else {
        return email;
      }
    }
    await new Promise((resolve) => setTimeout(resolve, checkIntervalMs));
  }

  return null;
}
