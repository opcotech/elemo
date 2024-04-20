/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $SystemLicense = {
    properties: {
        id: {
            type: 'string',
            description: `Unique ID identifying the license.`,
            isRequired: true,
        },
        organization: {
            type: 'string',
            description: `Name of the organization the license belongs to.`,
            isRequired: true,
        },
        email: {
            type: 'string',
            description: `Email address of the licensee.`,
            isRequired: true,
            format: 'email',
            maxLength: 254,
            minLength: 6,
        },
        quotas: {
            description: `Quotas available for the license.`,
            properties: {
                documents: {
                    type: 'number',
                    description: `Number of documents can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
                namespaces: {
                    type: 'number',
                    description: `Number of namespaces can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
                organizations: {
                    type: 'number',
                    description: `Number of organizations active can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
                projects: {
                    type: 'number',
                    description: `Number of projects can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
                roles: {
                    type: 'number',
                    description: `Number of roles can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
                users: {
                    type: 'number',
                    description: `Number of active or pending users can exist in the system.`,
                    isRequired: true,
                    minimum: 1,
                },
            },
            isRequired: true,
        },
        features: {
            type: 'array',
            contains: {
                type: 'Enum',
            },
            isRequired: true,
        },
        expires_at: {
            type: 'string',
            description: `Date and time when the license expires.`,
            isRequired: true,
            format: 'date-time',
        },
    },
} as const;
