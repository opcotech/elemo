/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $SystemVersion = {
  properties: {
    version: {
      type: 'string',
      description: `Version of the application.`,
      isRequired: true,
      pattern:
        '^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$'
    },
    commit: {
      type: 'string',
      description: `Commit hash of the build.`,
      isRequired: true,
      pattern: '^[0-9a-f]{5,40}$'
    },
    date: {
      type: 'string',
      description: `Build date and time of the application.`,
      isRequired: true,
      format: 'date-time'
    },
    go_version: {
      type: 'string',
      description: `Go version used to build the application.`,
      isRequired: true,
      pattern:
        '^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$'
    }
  }
} as const;
