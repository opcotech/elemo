import type { getTestConfig } from "./test-config";

/**
 * Verify that the backend API is accessible.
 * Throws an error if the API is not reachable.
 */
export async function verifyBackendAPI(
  config: ReturnType<typeof getTestConfig>
) {
  try {
    const response = await fetch(`${config.apiBaseUrl}/v1/system/health`, {
      method: "GET",
      signal: AbortSignal.timeout(5000),
    });

    if (!response.ok) {
      throw new Error(
        `Backend API health check failed with status ${response.status}`
      );
    }

    console.log("Backend API is accessible");
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Unknown error checking backend API";
    throw new Error(
      `Failed to verify backend API accessibility: ${message}. ` +
        `Make sure the backend is running at ${config.apiBaseUrl}`
    );
  }
}
