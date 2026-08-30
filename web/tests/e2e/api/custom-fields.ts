import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import { v1CustomFieldsCreate } from "@/lib/api/sdk";
import type { CustomFieldCreate, CustomFieldDefinition } from "@/lib/api/types";

export async function createCustomField(
  client: Client,
  body: CustomFieldCreate
): Promise<CustomFieldDefinition> {
  const response = await withErrorHandling(
    async () => {
      return await v1CustomFieldsCreate({
        client,
        body,
        throwOnError: true,
      });
    },
    {
      endpoint: "/v1/custom-fields",
      method: "POST",
    }
  );

  return response.data;
}
