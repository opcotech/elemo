import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryKey } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import type { UseFormReturn } from "react-hook-form";

import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface UseFormMutationOptions<TData, TVariables, TFormValues> {
  mutationFn: (variables: TVariables) => Promise<TData>;
  form: UseFormReturn<TFormValues>;
  onSuccess?: (data: TData) => void;
  onError?: (error: Error) => void;
  successMessage: string;
  successDescription?: string;
  errorMessagePrefix?: string;
  queryKeysToInvalidate?: QueryKey[];
  navigateOnSuccess?: string | { to: string; params?: Record<string, string> };
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
export function useFormMutation<TData, TVariables, TFormValues>({
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
}: UseFormMutationOptions<TData, TVariables, TFormValues>) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const mutation = useMutation({
    mutationFn,
    onSuccess: (data) => {
      // Invalidate queries
      queryKeysToInvalidate.forEach((queryKey) => {
        queryClient.invalidateQueries({ queryKey });
      });

      // Show success toast
      showSuccessToast(
        successMessage,
        successDescription || `${successMessage} successfully`
      );

      // Reset form if requested
      if (resetFormOnSuccess) {
        form.reset();
      }

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
    onError: (error: Error) => {
      const errorMessage = error.message || "Unknown error occurred";
      showErrorToast(errorMessagePrefix, errorMessage);
      onError?.(error);
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
