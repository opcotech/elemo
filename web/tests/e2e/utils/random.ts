/**
 * Utility functions for generating random strings for test data.
 * Used to ensure unique test data across parallel test runs.
 */

/**
 * Generate a random string of specified length.
 * Uses alphanumeric characters (a-z, A-Z, 0-9).
 *
 * @param length - Length of the random string (default: 10)
 * @returns Random string
 */
export function getRandomString(length: number = 10): string {
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}
