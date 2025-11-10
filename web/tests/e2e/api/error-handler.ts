import type { Response } from "@hey-api/client-fetch";

export interface APIErrorContext {
  endpoint: string;
  method: string;
  statusCode?: number;
  errorMessage?: string;
  testName?: string;
}

/**
 * Enhanced error handler for API calls.
 * Provides detailed error messages with context for debugging.
 */
export function handleAPIError(
  error: unknown,
  context: APIErrorContext
): never {
  let message = `API call failed: ${context.method} ${context.endpoint}`;

  if (context.testName) {
    message += ` (in test: ${context.testName})`;
  }

  console.error(message, error);

  if (error instanceof Error) {
    // Check if it's a fetch error
    if ("response" in error) {
      const response = (error as { response?: Response }).response;
      if (response) {
        message += `\n  Status: ${response.status} ${response.statusText}`;
        context.statusCode = response.status;

        // Try to extract error message from response
        try {
          const errorData = (error as { data?: unknown }).data;
          if (errorData && typeof errorData === "object") {
            const errorObj = errorData as Record<string, unknown>;
            if (errorObj.message) {
              message += `\n  Message: ${String(errorObj.message)}`;
              context.errorMessage = String(errorObj.message);
            } else if (errorObj.error_description) {
              message += `\n  Error: ${String(errorObj.error_description)}`;
              context.errorMessage = String(errorObj.error_description);
            }
          }
        } catch {
          // Ignore JSON parsing errors
        }
      }
    }

    message += `\n  Original error: ${error.message}`;
  } else {
    message += `\n  Unknown error: ${String(error)}`;
  }

  // Add suggestions based on status code
  if (context.statusCode === 401) {
    message += "\n  Suggestion: Check authentication token and credentials";
  } else if (context.statusCode === 403) {
    message += "\n  Suggestion: Verify user has required permissions";
  } else if (context.statusCode === 404) {
    message += "\n  Suggestion: Check if resource exists";
  } else if (context.statusCode === 500) {
    message += "\n  Suggestion: Check backend logs for server errors";
  }

  const enhancedError = new Error(message);
  (enhancedError as unknown as { context: APIErrorContext }).context = context;
  throw enhancedError;
}

/**
 * Wrap an API call with error handling.
 */
export async function withErrorHandling<T>(
  apiCall: () => Promise<T>,
  context: Omit<APIErrorContext, "statusCode" | "errorMessage">
): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    handleAPIError(error, context);
  }
}
