/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $Role = {
  description: `A role in the system.`,
  properties: {
    id: {
      type: 'string',
      description: `Unique identifier of the role.`,
      isRequired: true
    },
    name: {
      type: 'string',
      description: `Name of the role.`,
      isRequired: true,
      maxLength: 120,
      minLength: 3
    },
    description: {
      type: 'string',
      description: `Description of the role.`,
      isNullable: true,
      maxLength: 500,
      minLength: 5
    },
    members: {
      type: 'array',
      contains: {
        type: 'string'
      },
      isRequired: true
    },
    permissions: {
      type: 'array',
      contains: {
        type: 'string'
      },
      isRequired: true
    },
    created_at: {
      type: 'string',
      description: `Date when the organization was created.`,
      isRequired: true,
      format: 'date-time'
    },
    updated_at: {
      type: 'string',
      description: `Date when the organization was updated.`,
      isRequired: true,
      isNullable: true,
      format: 'date-time'
    }
  }
} as const;
