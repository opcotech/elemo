import { Client, type ApiConfig } from './api';
import { getErrorMessage } from './helpers';

let token: string | null = null;

export function setToken(newToken: string | null): void {
  token = newToken;
}

export function createClient(config: ApiConfig): Client<{ token: string }> {
  const client = new Client<{ token: string }>(config);

  if (token) {
    client.setSecurityData({ token });
  }

  return client;
}

const baseClient = createClient({
  baseUrl: process.env.NEXT_PUBLIC_ELEMO_BASE_URL
});

const client = new Proxy(baseClient, {
  get(_, key: keyof typeof baseClient) {
    const client = createClient({
      baseUrl: process.env.NEXT_PUBLIC_ELEMO_BASE_URL,
      baseApiParams: {
        headers: {
          Authorization: token ? `Bearer ${token}` : ''
        }
      }
    });

    return client[key];
  }
});

export default client;
export * from './api';
export { getErrorMessage };
