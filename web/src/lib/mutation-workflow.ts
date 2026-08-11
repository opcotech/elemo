import type { QueryClient, QueryKey } from "@tanstack/react-query";

type MutationWorkflowStep = () => void | Promise<void>;

interface MutationSuccessWorkflow {
  invalidateQueries?: MutationWorkflowStep[];
  invalidateRouter?: MutationWorkflowStep;
  callbacks?: MutationWorkflowStep[];
  navigate?: MutationWorkflowStep;
}

/**
 * Runs post-mutation work in cache, router, callback, then navigation phases.
 * Each phase completes before the next starts.
 */
export async function runMutationSuccessWorkflow({
  invalidateQueries = [],
  invalidateRouter,
  callbacks = [],
  navigate,
}: MutationSuccessWorkflow) {
  await Promise.all(invalidateQueries.map((invalidate) => invalidate()));
  await invalidateRouter?.();

  for (const callback of callbacks) {
    await callback();
  }

  await navigate?.();
}

interface OptimisticQueryContext<TData> {
  previous: TData | undefined;
}

export function rollbackOptimisticQueryData<TData>(
  queryClient: QueryClient,
  queryKey: QueryKey,
  context: OptimisticQueryContext<TData> | undefined
) {
  if (!context) {
    return;
  }

  if (context.previous === undefined) {
    queryClient.removeQueries({ queryKey, exact: true });
    return;
  }

  queryClient.setQueryData(queryKey, context.previous);
}
