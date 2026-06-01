import type { FieldValues, Path, UseFormSetError } from "react-hook-form";
import { ApiError } from "@/lib/api";

/**
 * Map backend validation field errors (422) onto the corresponding RHF fields.
 * Returns true if field-level errors were applied.
 */
export function applyApiErrors<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
): boolean {
  if (error instanceof ApiError && error.details && error.details.length > 0) {
    for (const detail of error.details) {
      setError(detail.field as Path<T>, {
        type: "server",
        message: detail.message,
      });
    }
    return true;
  }
  return false;
}

/** Extract a human-readable message from any thrown value. */
export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) return error.message;
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan tak terduga";
}
