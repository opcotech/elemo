export interface User {
  id: string;
  username: string;
  first_name: string | null;
  last_name: string | null;
  email: string;
  picture: string | null;
  title: string | null;
  bio: string | null;
  address: string | null;
  phone: string | null;
  links: string[] | null;
  languages: string[];
  status: string;
  created_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  token_type: string;
  expires_in: number;
  scope?: string;
}

export interface AuthState {
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthConfig {
  clientId: string;
  clientSecret: string;
  tokenUrl: string;
  scopes: string[];
}

export interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
  clearError: () => void;
}
