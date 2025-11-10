import dotenv from "dotenv";

import { USER_DEFAULT_PASSWORD } from "./auth";

// Load test environment variables
dotenv.config({ path: ".env.test.local", debug: false, quiet: true });

export interface TestConfig {
  apiBaseUrl: string;
  systemOwnerEmail: string;
  systemOwnerPassword: string;
  authClientId: string;
  authClientSecret: string;
  neo4jUrl: string;
  neo4jUser: string;
  neo4jPassword: string;
}

/**
 * Get test configuration from environment variables with sensible defaults.
 * Validates required values and provides clear error messages.
 */
export function getTestConfig(): TestConfig {
  const apiBaseUrl =
    process.env.E2E_API_BASE_URL ||
    process.env.VITE_API_BASE_URL ||
    "http://localhost:8080/api";

  const systemOwnerEmail =
    process.env.E2E_SYSTEM_OWNER_EMAIL || "e2e-test-owner@elemo.app";

  const systemOwnerPassword =
    process.env.E2E_SYSTEM_OWNER_PASSWORD || USER_DEFAULT_PASSWORD;

  const authClientId =
    process.env.E2E_AUTH_CLIENT_ID || process.env.VITE_AUTH_CLIENT_ID || "";

  const authClientSecret =
    process.env.E2E_AUTH_CLIENT_SECRET ||
    process.env.VITE_AUTH_CLIENT_SECRET ||
    "";

  const neo4jUrl = process.env.NEO4J_URL || "neo4j://localhost:7687";
  const neo4jUser = process.env.NEO4J_USER || "neo4j";
  const neo4jPassword = process.env.NEO4J_PASSWORD || "neo4jsecret";

  // Validate required values
  if (!authClientId) {
    throw new Error(
      "E2E_AUTH_CLIENT_ID or VITE_AUTH_CLIENT_ID environment variable is required"
    );
  }

  if (!authClientSecret) {
    throw new Error(
      "E2E_AUTH_CLIENT_SECRET or VITE_AUTH_CLIENT_SECRET environment variable is required"
    );
  }

  return {
    apiBaseUrl,
    systemOwnerEmail,
    systemOwnerPassword,
    authClientId,
    authClientSecret,
    neo4jUrl,
    neo4jUser,
    neo4jPassword,
  };
}
