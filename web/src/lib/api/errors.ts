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

function getErrorCode(error: unknown): string | undefined {
  if (!error || typeof error !== "object") {
    return undefined;
  }
  const record = error as Record<string, unknown>;
  if (typeof record.code === "string") {
    return record.code;
  }
  if (record.details && typeof record.details === "object") {
    const details = record.details as Record<string, unknown>;
    if (typeof details.code === "string") {
      return details.code;
    }
  }
  if ("cause" in error) {
    return getErrorCode(error.cause);
  }
  return undefined;
}

export const isSeatLimitReached = (error: unknown): boolean =>
  isConflict(error) && getErrorCode(error) === "seat_limit_reached";

export const isActivationDenied = (error: unknown): boolean =>
  isConflict(error) && getErrorCode(error) === "activation_denied";

export const isEntitlementReadOnly = (error: unknown): boolean =>
  isConflict(error) && getErrorCode(error) === "entitlement_read_only";

export function entitlementActivationErrorMessage(
  error: unknown
): string | null {
  if (isSeatLimitReached(error)) {
    return "No active human seats remain. Ask an installation administrator to free a seat or install a license with a higher seat limit.";
  }
  if (isActivationDenied(error)) {
    return "Human user activation is disabled by this installation's AirGap entitlement. Ask an installation administrator to install a valid license.";
  }
  return null;
}

export function entitlementReadOnlyErrorMessage(error: unknown): string | null {
  if (isEntitlementReadOnly(error)) {
    return "This installation is read-only until a valid License is installed and the server is restarted.";
  }
  return entitlementActivationErrorMessage(error);
}

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
  if (status !== undefined && status >= 200 && status < 300) {
    return result.data as T;
  }
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
