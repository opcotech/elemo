import { useQuery } from "@tanstack/react-query";

import { v1TodosGetOptions } from "@/lib/api";

export function useTodos() {
  return useQuery({
    ...v1TodosGetOptions(),
  });
}
