import { describe, expect, it } from "vitest";

import {
  ApiError,
  isConflict,
  isNotFound,
  isPermissionDenied,
  throwIfApiFailed,
  toApiError,
} from "@/lib/api/errors";

describe("api error helpers", () => {
  it("detects ApiError instances by status", () => {
    expect(isNotFound(new ApiError(404, "missing"))).toBe(true);
    expect(isPermissionDenied(new ApiError(403, "denied"))).toBe(true);
    expect(isConflict(new ApiError(409, "taken"))).toBe(true);
    expect(isNotFound(new ApiError(500, "boom"))).toBe(false);
  });

  it("detects plain objects that carry a numeric status", () => {
    expect(isNotFound({ status: 404, message: "missing" })).toBe(true);
    expect(isPermissionDenied({ status: 403 })).toBe(true);
    expect(isConflict({ response: { status: 409 } })).toBe(true);
  });

  it("walks nested cause chains for status", () => {
    expect(
      isNotFound({
        message: "loader failed",
        cause: new ApiError(404, "missing"),
      })
    ).toBe(true);
    expect(
      isPermissionDenied({
        message: "loader failed",
        cause: { status: 403 },
      })
    ).toBe(true);
  });

  it("builds ApiError values from responses", () => {
    const error = toApiError({ message: "gone" }, {
      status: 404,
      statusText: "Not Found",
    } as Response);
    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(404);
    expect(error.message).toBe("gone");
  });

  it("throws ApiError when the transport returns a failed response", () => {
    expect(() =>
      throwIfApiFailed({
        error: { message: "taken" },
        response: { status: 409, statusText: "Conflict" } as Response,
      })
    ).toThrow(ApiError);
  });

  it("accepts empty 204 responses as success", () => {
    expect(
      throwIfApiFailed({
        response: { status: 204, statusText: "No Content" } as Response,
      })
    ).toBeUndefined();
  });

  it("does not treat a 204 parse error as failure", () => {
    expect(
      throwIfApiFailed({
        error: new Error("Failed to parse JSON"),
        response: { status: 204, statusText: "No Content" } as Response,
      })
    ).toBeUndefined();
  });
});
