/** Minimal stub so Storybook can import app modules that use TanStack Start. */
type AnyFn = (...args: never[]) => unknown;

function createBuilder() {
  const builder: Record<string, AnyFn> = {};
  builder.validator = () => builder;
  builder.inputValidator = () => builder;
  builder.middleware = () => builder;
  builder.handler = () => async () => undefined;
  builder.server = (fn?: AnyFn) => fn ?? builder;
  builder.client = (fn?: AnyFn) => fn ?? builder;
  return builder;
}

export function createServerFn(_options?: unknown) {
  return createBuilder();
}

export function createMiddleware(_options?: unknown) {
  return createBuilder();
}

export function getCookie(_name: string) {
  return undefined;
}

export function setCookie(_name: string, _value: unknown, _options?: unknown) {
  return undefined;
}

export function deleteCookie(_name: string, _options?: unknown) {
  return undefined;
}

export default {};
