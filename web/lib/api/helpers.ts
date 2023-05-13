export function getErrorMessage(e: unknown): string {
  if (e instanceof Error) {
    return e.message;
  }

  if (typeof e === 'string') {
    return e;
  }

  return 'Unexpected error occurred, please try again.';
}
