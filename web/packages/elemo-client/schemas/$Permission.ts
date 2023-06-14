/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $Permission = {
    description: `A permission in the system.`,
    properties: {
        id: {
            type: 'string',
            description: `Unique identifier of the user.`,
            isRequired: true,
        },
        kind: {
            type: 'PermissionKind',
            isRequired: true,
        },
        subject: {
            type: 'string',
            isRequired: true,
        },
        target: {
            type: 'string',
            isRequired: true,
        },
        created_at: {
            type: 'string',
            description: `Date when the user was created.`,
            isRequired: true,
            format: 'date-time',
        },
        updated_at: {
            type: 'string',
            description: `Date when the user was updated.`,
            isRequired: true,
            isNullable: true,
            format: 'date-time',
        },
    },
} as const;
