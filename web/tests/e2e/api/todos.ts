import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/client/client";
import { v1TodoGet, v1TodosCreate } from "@/lib/client/sdk.gen";
import type { Todo, TodoCreate } from "@/lib/client/types.gen";

/**
 * Create a todo via API, then fetch the full todo by ID.
 *
 * @param client - Authenticated API client
 * @param todoData - Todo data (title and owned_by are required)
 * @returns Created todo with ID
 */
export async function createTodo(
  client: Client,
  todoData: Partial<TodoCreate> & { title: string; owned_by: string }
): Promise<Todo> {
  const todoCreateData: TodoCreate = {
    title: todoData.title,
    owned_by: todoData.owned_by,
    description: todoData.description,
    priority: todoData.priority ?? "normal",
    due_date: todoData.due_date,
  };

  const response = await withErrorHandling(
    async () => {
      return await v1TodosCreate({
        client,
        body: todoCreateData,
        throwOnError: true,
      });
    },
    {
      endpoint: "/v1/todos",
      method: "POST",
    }
  );

  const todoId = response.data.id || "";

  const todoResponse = await withErrorHandling(
    async () => {
      return await v1TodoGet({
        client,
        path: { id: todoId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/todos/${todoId}`,
      method: "GET",
    }
  );

  return todoResponse.data;
}
