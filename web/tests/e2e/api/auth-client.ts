import type { AuthTokens, LoginCredentials } from "@/lib/auth/types";

interface TestAuthConfig {
  clientId: string;
  clientSecret: string;
  tokenUrl: string;
  scopes: string[];
}

export class TestAuthClient {
  constructor(
    private readonly baseUrl: string,
    private readonly config: TestAuthConfig
  ) {}

  async login(credentials: LoginCredentials): Promise<AuthTokens> {
    const response = await fetch(`${this.baseUrl}${this.config.tokenUrl}`, {
      method: "POST",
      headers: {
        "content-type": "application/x-www-form-urlencoded",
      },
      body: new URLSearchParams({
        grant_type: "password",
        username: credentials.email,
        password: credentials.password,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret,
        scope: this.config.scopes.join(" "),
      }),
    });

    if (!response.ok) {
      const body = await response.text().catch(() => "");
      throw new Error(
        `Test login failed with status ${response.status}` +
          (body ? `: ${body}` : "")
      );
    }
    return response.json() as Promise<AuthTokens>;
  }

  async validateToken(accessToken: string): Promise<boolean> {
    const response = await fetch(`${this.baseUrl}/api/v1/users`, {
      method: "HEAD",
      headers: {
        authorization: `Bearer ${accessToken}`,
      },
    });
    return response.ok;
  }
}
