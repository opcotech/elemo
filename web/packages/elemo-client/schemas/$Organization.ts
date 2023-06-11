/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $Organization = {
    description: `An organization in the system.`,
    properties: {
        id: {
            type: 'string',
            description: `Unique identifier of the organization.`,
            isRequired: true,
        },
        name: {
            type: 'string',
            description: `Name of the organization.`,
            isRequired: true,
            maxLength: 120,
            minLength: 1,
        },
        email: {
            type: 'string',
            description: `Email address of the organization.`,
            isRequired: true,
            format: 'email',
            maxLength: 254,
            minLength: 6,
        },
        logo: {
            type: 'string',
            description: `Logo of the organization.`,
            isRequired: true,
            isNullable: true,
            format: 'uri',
            maxLength: 2000,
        },
        website: {
            type: 'string',
            description: `Work title of the user.`,
            isRequired: true,
            isNullable: true,
            format: 'uri',
            maxLength: 2000,
        },
        status: {
            type: 'OrganizationStatus',
            isRequired: true,
        },
        members: {
            type: 'array',
            contains: {
                type: 'string',
            },
            isRequired: true,
        },
        teams: {
            type: 'array',
            contains: {
                type: 'string',
            },
            isRequired: true,
        },
        namespaces: {
            type: 'array',
            contains: {
                type: 'string',
            },
            isRequired: true,
        },
        created_at: {
            type: 'string',
            description: `Date when the organization was created.`,
            isRequired: true,
            format: 'date-time',
        },
        updated_at: {
            type: 'string',
            description: `Date when the organization was updated.`,
            isRequired: true,
            isNullable: true,
            format: 'date-time',
        },
    },
} as const;
