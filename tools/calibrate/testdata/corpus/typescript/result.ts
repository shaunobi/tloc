export type Result<T, E> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };

export const success = <T>(value: T): Result<T, never> => ({ ok: true, value });
export const failure = <E>(error: E): Result<never, E> => ({ ok: false, error });

export function map<T, U, E>(result: Result<T, E>, transform: (value: T) => U): Result<U, E> {
  return result.ok ? success(transform(result.value)) : result;
}

export class ValidationError extends Error {
  constructor(
    readonly field: string,
    message: string,
  ) {
    super(`${field}: ${message}`);
    this.name = "ValidationError";
  }
}

export function parsePort(value: unknown): Result<number, ValidationError> {
  const port = typeof value === "string" ? Number.parseInt(value, 10) : Number(value);
  if (!Number.isInteger(port) || port < 1 || port > 65_535) {
    return failure(new ValidationError("port", `expected 1..65535, received ${String(value)}`));
  }
  return success(port);
}
