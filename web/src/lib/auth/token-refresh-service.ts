import { authClient } from "./auth-client";
import {
  clearSession,
  getRefreshToken,
  getTokenTimeRemaining,
  storeTokens,
} from "./session";

export class TokenRefreshService {
  private static instance: TokenRefreshService | null = null;
  private refreshTimeoutId: number | null = null;
  private isRefreshing = false;
  private refreshPromise: Promise<void> | null = null;

  private constructor() {}

  static getInstance(): TokenRefreshService {
    if (!TokenRefreshService.instance) {
      TokenRefreshService.instance = new TokenRefreshService();
    }
    return TokenRefreshService.instance;
  }

  /**
   * Start automatic token refresh scheduling
   */
  async startAutoRefresh(): Promise<void> {
    await this.scheduleNextRefresh();
  }

  /**
   * Stop automatic token refresh
   */
  stopAutoRefresh(): void {
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
      this.refreshTimeoutId = null;
    }
  }

  /**
   * Schedule the next token refresh based on token expiry
   */
  private async scheduleNextRefresh(): Promise<void> {
    // Clear any existing timeout
    this.stopAutoRefresh();

    try {
      const timeRemaining = await getTokenTimeRemaining();

      if (timeRemaining <= 0) {
        // Token already expired, try to refresh immediately
        await this.refreshToken();
        return;
      }

      // Schedule refresh 5 minutes before expiry, but at least 1 minute from now
      const refreshIn = Math.max(timeRemaining - 300000, 60000); // 5 minutes buffer, minimum 1 minute

      this.refreshTimeoutId = window.setTimeout(async () => {
        await this.refreshToken();
      }, refreshIn);

      console.debug(
        `Token refresh scheduled in ${Math.round(refreshIn / 1000 / 60)} minutes`
      );
    } catch (error) {
      console.error("Failed to schedule token refresh:", error);
      // Retry in 5 minutes if scheduling fails
      this.refreshTimeoutId = window.setTimeout(async () => {
        await this.scheduleNextRefresh();
      }, 300000);
    }
  }

  /**
   * Refresh the access token using the refresh token
   */
  async refreshToken(): Promise<void> {
    // Prevent multiple simultaneous refresh attempts
    if (this.isRefreshing) {
      return this.refreshPromise || Promise.resolve();
    }

    this.isRefreshing = true;
    this.refreshPromise = this.performRefresh();

    try {
      await this.refreshPromise;
    } finally {
      this.isRefreshing = false;
      this.refreshPromise = null;
    }
  }

  /**
   * Perform the actual token refresh
   */
  private async performRefresh(): Promise<void> {
    try {
      const refreshToken = getRefreshToken();

      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      console.debug("Refreshing access token...");
      const newTokens = await authClient.refreshToken(refreshToken);
      await storeTokens(newTokens);

      console.debug("Access token refreshed successfully");

      // Schedule the next refresh
      await this.scheduleNextRefresh();

      // Dispatch custom event to notify components
      window.dispatchEvent(
        new CustomEvent("tokenRefreshed", {
          detail: { tokens: newTokens },
        })
      );
    } catch (error) {
      console.error("Token refresh failed:", error);

      // If refresh fails, clear session and dispatch logout event
      clearSession();
      window.dispatchEvent(
        new CustomEvent("tokenRefreshFailed", {
          detail: { error },
        })
      );

      throw error;
    }
  }

  /**
   * Force an immediate token refresh
   */
  async forceRefresh(): Promise<void> {
    this.stopAutoRefresh();
    await this.refreshToken();
  }

  /**
   * Check if a refresh is currently in progress
   */
  isRefreshInProgress(): boolean {
    return this.isRefreshing;
  }

  /**
   * Get time until next scheduled refresh (in milliseconds)
   */
  getTimeUntilNextRefresh(): number {
    if (!this.refreshTimeoutId) return 0;

    // This is an approximation since we can't get exact timeout remaining
    return 0;
  }
}

// Export singleton instance
export const tokenRefreshService = TokenRefreshService.getInstance();

// Utility functions for easier usage
export const startTokenRefresh = () => tokenRefreshService.startAutoRefresh();
export const stopTokenRefresh = () => tokenRefreshService.stopAutoRefresh();
export const refreshTokenNow = () => tokenRefreshService.refreshToken();
export const forceTokenRefresh = () => tokenRefreshService.forceRefresh();
