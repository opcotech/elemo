/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $User = {
    description: `A user in the system.`,
    properties: {
        id: {
            type: 'string',
            description: `Unique identifier of the user.`,
            isRequired: true,
        },
        username: {
            type: 'string',
            description: `The unique username of the user.`,
            isRequired: true,
            maxLength: 50,
            minLength: 3,
            pattern: '^[a-z0-9-_]{3,50}$',
        },
        first_name: {
            type: 'string',
            description: `First name of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 50,
            minLength: 1,
        },
        last_name: {
            type: 'string',
            description: `Last name of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 50,
            minLength: 1,
        },
        email: {
            type: 'string',
            description: `Email address of the user.`,
            isRequired: true,
            format: 'email',
            maxLength: 254,
            minLength: 6,
        },
        picture: {
            type: 'string',
            description: `Profile picture of the user.`,
            isRequired: true,
            isNullable: true,
            format: 'uri',
            maxLength: 2000,
        },
        title: {
            type: 'string',
            description: `Work title of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 50,
            minLength: 3,
        },
        bio: {
            type: 'string',
            description: `Self description of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 500,
        },
        address: {
            type: 'string',
            description: `Working address of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 500,
            minLength: 3,
        },
        phone: {
            type: 'string',
            description: `Phone number of the user.`,
            isRequired: true,
            isNullable: true,
            maxLength: 16,
            minLength: 7,
        },
        links: {
            type: 'array',
            contains: {
                type: 'string',
                format: 'uri',
                maxLength: 2000,
            },
            isRequired: true,
            isNullable: true,
        },
        languages: {
            type: 'array',
            contains: {
                type: 'Language',
            },
            isRequired: true,
        },
        status: {
            type: 'UserStatus',
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
