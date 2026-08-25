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

function numericStatus(value: unknown): number | undefined {
  return typeof value === "number" && Number.isInteger(value)
    ? value
    : undefined;
}

function getErrorStatus(error: unknown): number | undefined {
  if (error instanceof ApiError) {
    return error.status;
  }
  if (!error || typeof error !== "object") {
    return undefined;
  }
  const record = error as Record<string, unknown>;
  const direct = numericStatus(record.status);
  if (direct !== undefined) {
    return direct;
  }
  if (record.response && typeof record.response === "object") {
    const nested = numericStatus(
      (record.response as Record<string, unknown>).status
    );
    if (nested !== undefined) {
      return nested;
    }
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

export const isConflict = (error: unknown): boolean =>
  getErrorStatus(error) === 409;

export function isNotFoundOrForbidden(error: unknown): boolean {
  return isNotFound(error) || isPermissionDenied(error);
}

export function throwIfApiFailed<T>(result: {
  data?: T;
  error?: unknown;
  response?: Response;
}): T {
  const status =
    result.response && typeof result.response.status === "number"
      ? result.response.status
      : undefined;
  if (result.error || (status !== undefined && status >= 400)) {
    if (result.error instanceof Error) {
      throw result.error;
    }
    throw toApiError(result.error, result.response);
  }
  if (result.data === undefined || result.data === null) {
    throw toApiError(result.error, result.response);
  }
  return result.data;
}
