import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryKey, UseMutationOptions } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { showErrorToast, showSuccessToast } from "@/lib/toast";

/**
 * Generic hook for delete mutations with standardized handling:
 * - Query invalidation
 * - Toast notifications
 * - Optional navigation on success
 * - Error handling
 */
export function useDeleteMutation<
  TData = unknown,
  TVariables = unknown,
  TError = Error,
  TContext = unknown,
>({
  mutationOptions,
  successMessage,
  successDescription,
  errorMessagePrefix = "Failed to delete",
  queryKeysToInvalidate = [],
  onSuccess,
  onError,
  navigateOnSuccess,
}: {
  mutationOptions: UseMutationOptions<TData, TError, TVariables, TContext>;
  successMessage: string;
  successDescription?: string;
  errorMessagePrefix?: string;
  queryKeysToInvalidate?: QueryKey[];
  onSuccess?: (data: TData) => void;
  onError?: (error: TError) => void;
  navigateOnSuccess?: string | { to: string; params?: Record<string, string> };
}) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation<TData, TError, TVariables, TContext>({
    ...mutationOptions,
    onSuccess: (data, variables, context, mutation) => {
      // Call original onSuccess if provided
      mutationOptions.onSuccess?.(data, variables, context, mutation);

      // Invalidate queries
      queryKeysToInvalidate.forEach((queryKey) => {
        queryClient.invalidateQueries({ queryKey });
      });

      // Show success toast
      showSuccessToast(
        successMessage,
        successDescription || `${successMessage} successfully`
      );

      // Call custom success handler
      onSuccess?.(data);

      // Navigate if specified
      if (navigateOnSuccess) {
        if (typeof navigateOnSuccess === "string") {
          navigate({ to: navigateOnSuccess });
        } else {
          navigate({
            to: navigateOnSuccess.to as any,
            params: navigateOnSuccess.params as any,
          });
        }
      }
    },
    onError: (error, variables, context, mutation) => {
      // Call original onError if provided
      mutationOptions.onError?.(error, variables, context, mutation);

      const errorMessage =
        error instanceof Error ? error.message : "Unknown error occurred";
      showErrorToast(errorMessagePrefix, errorMessage);
      onError?.(error);
    },
  });
}
