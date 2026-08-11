import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryKey, UseMutationOptions } from "@tanstack/react-query";
import { useNavigate, useRouter } from "@tanstack/react-router";

import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

type MutationCallback<TValue> = (value: TValue) => void | Promise<void>;

type NavigateOnSuccess = (
  navigate: ReturnType<typeof useNavigate>
) => void | Promise<void>;

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
  onSuccess?: MutationCallback<TData>;
  onError?: MutationCallback<TError>;
  navigateOnSuccess?: NavigateOnSuccess;
}) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const router = useRouter();

  return useMutation<TData, TError, TVariables, TContext>({
    ...mutationOptions,
    onSuccess: (data, variables, context, mutation) =>
      runMutationSuccessWorkflow({
        invalidateQueries: queryKeysToInvalidate.map(
          (queryKey) => () => queryClient.invalidateQueries({ queryKey })
        ),
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          async () => {
            await mutationOptions.onSuccess?.(
              data,
              variables,
              context,
              mutation
            );
          },
          () => {
            showSuccessToast(
              successMessage,
              successDescription || `${successMessage} successfully`
            );
          },
          async () => {
            await onSuccess?.(data);
          },
        ],
        navigate: navigateOnSuccess
          ? () => navigateOnSuccess(navigate)
          : undefined,
      }),
    onError: async (error, variables, context, mutation) => {
      await mutationOptions.onError?.(error, variables, context, mutation);
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error occurred";
      showErrorToast(errorMessagePrefix, errorMessage);
      await onError?.(error);
    },
  });
}
