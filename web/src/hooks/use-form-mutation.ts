import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryKey } from "@tanstack/react-query";
import { useNavigate, useRouter } from "@tanstack/react-router";
import type { FieldValues, UseFormReturn } from "react-hook-form";

import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

type MutationCallback<TValue> = (value: TValue) => void | Promise<void>;

type NavigateOnSuccess<TData> = (
  navigate: ReturnType<typeof useNavigate>,
  data: TData
) => void | Promise<void>;

interface UseFormMutationOptions<
  TData,
  TVariables,
  TFormValues extends FieldValues,
  TError extends Error = Error,
> {
  mutationFn: (variables: TVariables) => Promise<TData>;
  form: UseFormReturn<TFormValues>;
  onSuccess?: MutationCallback<TData>;
  onError?: MutationCallback<TError>;
  successMessage?: string;
  successDescription?: string;
  errorMessagePrefix?: string;
  queryKeysToInvalidate?: QueryKey[];
  navigateOnSuccess?: NavigateOnSuccess<TData>;
  resetFormOnSuccess?: boolean;
  transformValues?: (values: TFormValues) => TVariables;
}

/**
 * Generic hook for form mutations with standardized handling:
 * - Form submission with react-hook-form
 * - Loading/error states
 * - Query invalidation
 * - Toast notifications
 * - Optional navigation on success
 * - Optional form reset on success
 */
export function useFormMutation<
  TData,
  TVariables,
  TFormValues extends FieldValues,
  TError extends Error = Error,
>({
  mutationFn,
  form,
  onSuccess,
  onError,
  successMessage,
  successDescription,
  errorMessagePrefix = "Failed to save",
  queryKeysToInvalidate = [],
  navigateOnSuccess,
  resetFormOnSuccess = false,
  transformValues,
}: UseFormMutationOptions<TData, TVariables, TFormValues, TError>) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const router = useRouter();

  const mutation = useMutation<TData, TError, TVariables>({
    mutationFn,
    onSuccess: (data) =>
      runMutationSuccessWorkflow({
        invalidateQueries: queryKeysToInvalidate.map(
          (queryKey) => () => queryClient.invalidateQueries({ queryKey })
        ),
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          async () => {
            if (successMessage) {
              showSuccessToast(
                successMessage,
                successDescription || `${successMessage} successfully`
              );
            }

            if (resetFormOnSuccess) {
              form.reset();
            }

            await onSuccess?.(data);
          },
        ],
        navigate: navigateOnSuccess
          ? () => navigateOnSuccess(navigate, data)
          : undefined,
      }),
    onError: async (error: TError) => {
      const errorMessage = error.message || "Unknown error occurred";
      showErrorToast(errorMessagePrefix, errorMessage);
      await onError?.(error);
    },
  });

  const handleSubmit = form.handleSubmit((values) => {
    const variables = transformValues
      ? transformValues(values)
      : (values as unknown as TVariables);
    mutation.mutate(variables);
  });

  return {
    ...mutation,
    handleSubmit,
    isPending: mutation.isPending,
    isError: mutation.isError,
    error: mutation.error,
  };
}
