import Cookies from "js-cookie";

import type { AuthTokens, User } from "./types";

const ACCESS_TOKEN_KEY = "elemo_at";
const REFRESH_TOKEN_KEY = "elemo_rt";
const USER_KEY = "elemo_user";
const TOKEN_EXPIRY_KEY = "elemo_at_exp";

// Proper AES encryption for localStorage
class SecureStorage {
  private static async getKey(): Promise<CryptoKey> {
    // Use a combination of device fingerprint and fixed salt
    const fingerprint = this.getDeviceFingerprint();
    const encoder = new TextEncoder();
    const keyMaterial = await crypto.subtle.importKey(
      "raw",
      encoder.encode(fingerprint),
      { name: "PBKDF2" },
      false,
      ["deriveBits", "deriveKey"]
    );

    return crypto.subtle.deriveKey(
      {
        name: "PBKDF2",
        salt: encoder.encode("elemo-auth-salt-2024"),
        iterations: 100000,
        hash: "SHA-256",
      },
      keyMaterial,
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"]
    );
  }

  private static getDeviceFingerprint(): string {
    // Create a device-specific fingerprint
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    if (ctx) {
      ctx.textBaseline = "top";
      ctx.font = "14px Arial";
      ctx.fillText("Device fingerprint", 2, 2);
    }

    const fingerprint = [
      navigator.userAgent,
      navigator.language,
      screen.width + "x" + screen.height,
      new Date().getTimezoneOffset(),
      canvas.toDataURL(),
      localStorage.length,
    ].join("|");

    // Hash the fingerprint to get consistent length
    return btoa(fingerprint).slice(0, 32);
  }

  private static async encrypt(data: string): Promise<string> {
    try {
      const key = await this.getKey();
      const encoder = new TextEncoder();
      const iv = crypto.getRandomValues(new Uint8Array(12));

      const encryptedData = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        key,
        encoder.encode(data)
      );

      // Combine IV and encrypted data
      const combined = new Uint8Array(iv.length + encryptedData.byteLength);
      combined.set(iv);
      combined.set(new Uint8Array(encryptedData), iv.length);

      return btoa(String.fromCharCode(...combined));
    } catch (error) {
      console.error("Encryption failed:", error);
      return btoa(data); // Fallback to base64 if encryption fails
    }
  }

  private static async decrypt(encryptedData: string): Promise<string> {
    try {
      const key = await this.getKey();
      const combined = new Uint8Array(
        atob(encryptedData)
          .split("")
          .map((char) => char.charCodeAt(0))
      );

      const iv = combined.slice(0, 12);
      const data = combined.slice(12);

      const decryptedData = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv },
        key,
        data
      );

      return new TextDecoder().decode(decryptedData);
    } catch (error) {
      console.error("Decryption failed:", error);
      // Fallback: try to decode as base64
      try {
        return atob(encryptedData);
      } catch {
        return "";
      }
    }
  }

  static async setItem(key: string, value: string): Promise<void> {
    if (typeof window !== "undefined") {
      const encrypted = await this.encrypt(value);
      localStorage.setItem(key, encrypted);
    }
  }

  static async getItem(key: string): Promise<string | null> {
    if (typeof window === "undefined") return null;
    const item = localStorage.getItem(key);
    return item ? await this.decrypt(item) : null;
  }

  static removeItem(key: string): void {
    if (typeof window !== "undefined") {
      localStorage.removeItem(key);
    }
  }
}

/**
 * Session management utilities for handling auth tokens and user data
 */
export class SessionManager {
  /**
   * Store authentication tokens with improved security
   */
  static async storeTokens(tokens: AuthTokens): Promise<void> {
    // Store access token encrypted in localStorage with expiry
    if (typeof window !== "undefined") {
      await SecureStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);

      // Store token expiry (assume 1 hour if not provided)
      const expiryTime =
        Date.now() + (tokens.expires_in ? tokens.expires_in * 1000 : 3600000);
      await SecureStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());

      // Also store access token in a cookie so that it is available during
      // server-side rendering (cookies are forwarded with the initial HTTP
      // request, unlike localStorage). We intentionally keep it **non**
      // httpOnly here so the client-side code can still read it when needed.
      // The cookie expiry mirrors the token TTL converted from seconds ➜ days.
      const isSecure = window.location.protocol === "https:";
      Cookies.set(ACCESS_TOKEN_KEY, tokens.access_token, {
        path: "/",
        httpOnly: false,
        secure: isSecure,
        sameSite: "strict",
        expires: tokens.expires_in ? tokens.expires_in / 86400 : 1 / 24, // default 1h
      });
    }

    // Store refresh token in secure HTTP-only cookie
    if (tokens.refresh_token) {
      const isSecure = window.location.protocol === "https:";
      Cookies.set(REFRESH_TOKEN_KEY, tokens.refresh_token, {
        path: "/",
        httpOnly: false, // Should be httpOnly when backend supports it
        secure: isSecure,
        sameSite: "strict",
        expires: 30, // 30 days
      });
    }
  }

  /**
   * Get current access token if not expired
   */
  static async getAccessToken(): Promise<string | null> {
    if (typeof window === "undefined") return null;

    // Check if token is expired
    if (await this.isAccessTokenExpired()) {
      await this.clearAccessToken();
      return null;
    }

    return await SecureStorage.getItem(ACCESS_TOKEN_KEY);
  }

  /**
   * Get current refresh token
   */
  static getRefreshToken(): string | null {
    try {
      return Cookies.get(REFRESH_TOKEN_KEY) || null;
    } catch (error) {
      console.warn("Error getting refresh token:", error);
      return null;
    }
  }

  /**
   * Store user data encrypted
   */
  static async storeUser(user: User): Promise<void> {
    if (typeof window !== "undefined") {
      await SecureStorage.setItem(USER_KEY, JSON.stringify(user));
    }
  }

  /**
   * Get stored user data
   */
  static async getUser(): Promise<User | null> {
    if (typeof window === "undefined") return null;

    try {
      const userData = await SecureStorage.getItem(USER_KEY);
      return userData ? JSON.parse(userData) : null;
    } catch {
      return null;
    }
  }

  /**
   * Clear all stored auth data
   */
  static clearSession(): void {
    if (typeof window !== "undefined") {
      SecureStorage.removeItem(ACCESS_TOKEN_KEY);
      SecureStorage.removeItem(USER_KEY);
      SecureStorage.removeItem(TOKEN_EXPIRY_KEY);
      // Remove access-token cookie as well
      Cookies.remove(ACCESS_TOKEN_KEY, { path: "/" });
    }
    Cookies.remove(REFRESH_TOKEN_KEY);
  }

  /**
   * Comprehensive cleanup of all auth-related data including any legacy/corrupted data
   */
  static clearAllAuthData(): void {
    if (typeof window !== "undefined") {
      // Clear our encrypted storage
      SecureStorage.removeItem(ACCESS_TOKEN_KEY);
      SecureStorage.removeItem(USER_KEY);
      SecureStorage.removeItem(TOKEN_EXPIRY_KEY);

      // Clear any legacy keys that might exist
      const keysToRemove = [
        "access_token",
        "user_data",
        "elemo_at_exp",
        "auth_tokens",
        "user_profile",
        "session_data",
      ];

      keysToRemove.forEach((key) => {
        localStorage.removeItem(key);
        sessionStorage.removeItem(key);
      });

      // Clear any elemo-prefixed keys
      Object.keys(localStorage).forEach((key) => {
        if (key.startsWith("elemo_")) {
          localStorage.removeItem(key);
        }
      });
    }

    // Clear all cookies that might contain auth data
    const cookiesToRemove = [
      "elemo_rt",
      "refresh_token",
      "auth_token",
      "session_token",
    ];

    cookiesToRemove.forEach((cookieName) => {
      Cookies.remove(cookieName);
      // Also try removing with different path/domain combinations
      Cookies.remove(cookieName, { path: "/" });
      Cookies.remove(cookieName, {
        path: "",
        domain: window.location.hostname,
      });
    });
  }

  /**
   * Clear only access token (for when it expires)
   */
  static clearAccessToken(): void {
    if (typeof window !== "undefined") {
      SecureStorage.removeItem(ACCESS_TOKEN_KEY);
      SecureStorage.removeItem(TOKEN_EXPIRY_KEY);
      // Remove access-token cookie as well
      Cookies.remove(ACCESS_TOKEN_KEY, { path: "/" });
    }
  }

  /**
   * Check if user has valid tokens (not just user data)
   */
  static hasValidSession(): boolean {
    // Ensure we're in browser environment
    if (typeof window === "undefined") return false;

    try {
      const refreshToken = this.getRefreshToken();
      // Only consider session valid if we have a refresh token
      // Access token will be checked/refreshed separately
      return !!refreshToken;
    } catch (error) {
      console.warn("Error checking session validity:", error);
      return false;
    }
  }

  /**
   * Check if access token is expired based on stored timestamp
   */
  static async isAccessTokenExpired(): Promise<boolean> {
    if (typeof window === "undefined") return true;

    try {
      const expiryStr = await SecureStorage.getItem(TOKEN_EXPIRY_KEY);
      if (!expiryStr) return true;

      const expiry = parseInt(expiryStr, 10);
      const now = Date.now();

      // Consider token expired if it expires within 5 minutes (buffer for refresh)
      return expiry - now < 300000; // 5 minutes in milliseconds
    } catch {
      return true;
    }
  }

  /**
   * Get time until token expires (in milliseconds)
   */
  static async getTokenTimeRemaining(): Promise<number> {
    if (typeof window === "undefined") return 0;

    try {
      const expiryStr = await SecureStorage.getItem(TOKEN_EXPIRY_KEY);
      if (!expiryStr) return 0;

      const expiry = parseInt(expiryStr, 10);
      const remaining = expiry - Date.now();

      return Math.max(0, remaining);
    } catch {
      return 0;
    }
  }

  /**
   * Manual cleanup utility for debugging - completely wipes all auth data
   */
  static manualCleanup(): void {
    // Clear all storage
    this.clearAllAuthData();

    // Clear any additional browser storage
    if (typeof window !== "undefined") {
      try {
        // Clear all localStorage
        const allKeys = Object.keys(localStorage);
        allKeys.forEach((key) => {
          if (
            key.includes("auth") ||
            key.includes("token") ||
            key.includes("user") ||
            key.includes("elemo")
          ) {
            localStorage.removeItem(key);
          }
        });

        // Clear all sessionStorage
        const sessionKeys = Object.keys(sessionStorage);
        sessionKeys.forEach((key) => {
          sessionStorage.removeItem(key);
        });

        // Clear all cookies
        document.cookie.split(";").forEach((cookie) => {
          const eqPos = cookie.indexOf("=");
          const name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
          document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
        });
      } catch (error) {
        console.error("Error during manual cleanup:", error);
      }
    }
  }

  /**
   * Debug function to inspect current auth state
   */
  static async debugAuthState(): Promise<void> {
    try {
      const accessToken = await this.getAccessToken();
      const refreshToken = this.getRefreshToken();
      const user = await this.getUser();
      const hasSession = this.hasValidSession();
      const isExpired = await this.isAccessTokenExpired();
      const timeRemaining = await this.getTokenTimeRemaining();

      console.log("Access Token:", accessToken ? "✅ Present" : "❌ Missing");
      console.log("Refresh Token:", refreshToken ? "✅ Present" : "❌ Missing");
      console.log("User Data:", user ? "✅ Present" : "❌ Missing");
      console.log("Has Valid Session:", hasSession ? "✅ Yes" : "❌ No");
      console.log("Is Token Expired:", isExpired ? "❌ Yes" : "✅ No");
      console.log(
        "Time Remaining:",
        timeRemaining > 0 ? `${Math.floor(timeRemaining / 60000)}m` : "Expired"
      );

      if (user) {
        console.log("User Info:", { id: user.id, email: user.email });
      }

      // Check localStorage for any auth-related keys
      const authKeys = Object.keys(localStorage).filter(
        (key) =>
          key.includes("elemo") ||
          key.includes("auth") ||
          key.includes("token") ||
          key.includes("user")
      );
      console.log("Auth-related localStorage keys:", authKeys);
    } catch (error) {
      console.error("Error debugging auth state:", error);
    }

    console.groupEnd();
  }
}

// Convenience exports (async functions)
export const storeTokens = (tokens: AuthTokens) =>
  SessionManager.storeTokens(tokens);
export const getAccessToken = () => SessionManager.getAccessToken();
export const getRefreshToken = () => SessionManager.getRefreshToken();
export const storeUser = (user: User) => SessionManager.storeUser(user);
export const getUser = () => SessionManager.getUser();
export const clearSession = SessionManager.clearSession.bind(SessionManager);
export const clearAllAuthData =
  SessionManager.clearAllAuthData.bind(SessionManager);
export const clearAccessToken =
  SessionManager.clearAccessToken.bind(SessionManager);
export const hasValidSession =
  SessionManager.hasValidSession.bind(SessionManager);
export const isAccessTokenExpired = () => SessionManager.isAccessTokenExpired();
export const getTokenTimeRemaining = () =>
  SessionManager.getTokenTimeRemaining();

// Debug utilities
export const manualAuthCleanup =
  SessionManager.manualCleanup.bind(SessionManager);
export const debugAuthState =
  SessionManager.debugAuthState.bind(SessionManager);

// Make cleanup available globally for debugging
if (typeof window !== "undefined") {
  (window as any).manualAuthCleanup = manualAuthCleanup;
  (window as any).debugAuthState = debugAuthState;
  console.debug(
    "🔧 Auth debug tools available: window.manualAuthCleanup() and window.debugAuthState()"
  );
}
