import type { AuthConfig, AuthTokens, LoginCredentials } from "./types";

/** Pure OAuth client — no Vite/config dependency (safe for Playwright/Node). */
export class AuthClient {
  private baseUrl: string;
  private config: AuthConfig;

  constructor(baseUrl: string, authConfig: AuthConfig) {
    this.baseUrl = baseUrl;
    this.config = authConfig;
  }

  /**
   * Authenticate user with email and password using OAuth password grant
   */
  async login(credentials: LoginCredentials): Promise<AuthTokens> {
    const url = `${this.baseUrl}${this.config.tokenUrl}`;
    const body = new URLSearchParams({
      grant_type: "password",
      username: credentials.email,
      password: credentials.password,
      client_id: this.config.clientId,
      client_secret: this.config.clientSecret,
      scope: this.config.scopes.join(" "),
    });

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.error_description || errorData.message || "Login failed"
      );
    }

    const tokens: AuthTokens = await response.json();
    return tokens;
  }

  /**
   * Refresh access token using refresh token
   */
  async refreshToken(refreshToken: string): Promise<AuthTokens> {
    const response = await fetch(`${this.baseUrl}${this.config.tokenUrl}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        refresh_token: refreshToken,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.error_description ||
          errorData.message ||
          "Token refresh failed"
      );
    }

    const tokens: AuthTokens = await response.json();
    return tokens;
  }

  /**
   * Validate current access token
   */
  async validateToken(accessToken: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/v1/users`, {
        method: "HEAD",
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}
