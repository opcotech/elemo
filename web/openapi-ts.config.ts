import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: `${process.env.ROOT_DIR}/api/openapi/openapi.yaml`,
  output: {
    path: `${process.env.PACKAGE_DIR}`,
    clean: true,
    postProcess: ["prettier"],
  },
  plugins: [
    {
      name: "@hey-api/client-fetch",
    },
    {
      name: "@hey-api/typescript",
    },
    {
      name: "@hey-api/sdk",
      client: "@hey-api/client-fetch",
      operations: "flat",
    },
    {
      name: "@tanstack/react-query",
      infiniteQueryKeys: false,
      infiniteQueryOptions: false,
      mutationKeys: false,
      mutationOptions: true,
      queryKeys: true,
      queryOptions: true,
    },
    {
      name: "zod",
      requests: true,
    },
  ],
});