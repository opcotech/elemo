export class ApiError extends Error {
  readonly status: number;
  readonly details: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
  }
}

export function toApiError(error: unknown, response?: Response): ApiError {
  const status = response?.status ?? 500;
  const details =
    error && typeof error === "object"
      ? error
      : error
        ? { message: error }
        : {};
  const candidate =
    details && typeof details === "object" && "message" in details
      ? details.message
      : undefined;
  const message =
    typeof candidate === "string" && candidate
      ? candidate
      : response?.statusText || "API request failed";

  return new ApiError(status, message, details);
}

export function isApiError(error: unknown, status?: number): error is ApiError {
  return (
    error instanceof ApiError &&
    (status === undefined || error.status === status)
  );
}

function getErrorStatus(error: unknown): number | undefined {
  if (error instanceof ApiError) {
    return error.status;
  }
  if (!error || typeof error !== "object") {
    return undefined;
  }
  if ("status" in error && typeof error.status === "number") {
    return error.status;
  }
  if ("cause" in error) {
    return getErrorStatus(error.cause);
  }
  return undefined;
}

export const isPermissionDenied = (error: unknown): boolean =>
  getErrorStatus(error) === 403;

export const isNotFound = (error: unknown): boolean =>
  getErrorStatus(error) === 404;

export function isNotFoundOrForbidden(error: unknown): boolean {
  return isNotFound(error) || isPermissionDenied(error);
}
