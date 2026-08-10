import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";

import { ApiError } from "@/lib/api/errors";

export const cacheProfiles = {
  reference: {
    staleTime: 15 * 60 * 1000,
    gcTime: 60 * 60 * 1000,
  },
  entity: {
    staleTime: 5 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
  },
  volatile: {
    staleTime: 15 * 1000,
    gcTime: 5 * 60 * 1000,
  },
} as const;

function retryTransientFailure(failureCount: number, error: unknown) {
  if (failureCount >= 2) {
    return false;
  }
  if (!(error instanceof ApiError)) {
    return failureCount < 1;
  }
  return (
    error.status === 408 ||
    error.status === 429 ||
    (error.status >= 500 && error.status <= 599)
  );
}

function logUnexpectedError(error: unknown) {
  if (!(error instanceof ApiError) || error.status >= 500) {
    console.error("Unexpected data operation failure", error);
  }
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: logUnexpectedError,
    }),
    mutationCache: new MutationCache({
      onError: logUnexpectedError,
    }),
    defaultOptions: {
      queries: {
        retry: retryTransientFailure,
        ...cacheProfiles.entity,
        refetchOnWindowFocus: false,
        refetchOnReconnect: true,
      },
      mutations: {
        retry: false,
      },
    },
  });
}
