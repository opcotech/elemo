import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
    input: `${process.env.ROOT_DIR}/api/openapi/openapi.yaml`,
    output: `${process.env.PACKAGE_DIR}`,
    plugins: [
      {
        name: '@tanstack/react-query',
      },
      {
        name: 'zod',
        requests: true,
      },
    ],
  });
  