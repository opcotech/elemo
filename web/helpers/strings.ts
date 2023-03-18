/**
 * Concatenate arbitrary number of strings with a whitespace.
 * @param arr Array of strings to concatenate.
 * @returns Concatenated string.
 */
export function concat(...arr: (string | undefined)[]): string {
  return arr.filter(Boolean).join(' ');
}

/**
 * Format the error message of a form field.
 *
 * @param field The form field name.
 * @param message The error message.
 * @returns Formatted error message.
 */
export function formatErrorMessage(field: string, message?: string): string {
  if (!message) {
    return `Unknown error.`;
  }

  return message.replace('String', `The ${field}`) + '.';
}

/**
 * Capitalize the first letter of a string.
 *
 * @param str String to capitalize.
 * @returns Capitalized string.
 */
export function toCapitalCase(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

/**
 * Return the initials of a string.
 *
 * @param str String to get the initials from.
 * @returns Initials of the string.
 */
export function getInitials(str: string | undefined | null | (string | undefined | null)[]): string {
  if (!str) {
    return '';
  }

  if (Array.isArray(str)) {
    return str.map((s) => s?.charAt(0).toUpperCase()).join('');
  }

  return str
    .split(' ')
    .map((word) => word.charAt(0).toUpperCase())
    .join('');
}

/**
 * Slugify a string.
 * @param str String to slugify.
 * @returns Slugified string.
 */
export function slugify(str: string): string {
  return str
    .toLowerCase()
    .replace(/ /g, '-')
    .replace(/[^\w-]+/g, '');
}
