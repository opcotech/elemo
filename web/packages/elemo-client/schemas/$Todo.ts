/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $Todo = {
    description: `A todo item belonging to a user.`,
    properties: {
        id: {
            type: 'string',
            description: `Unique identifier of the todo.`,
            isRequired: true,
        },
        title: {
            type: 'string',
            description: `Title of the todo item.`,
            isRequired: true,
            maxLength: 250,
            minLength: 3,
        },
        description: {
            type: 'string',
            description: `Description of the todo item.`,
            isRequired: true,
            maxLength: 500,
            minLength: 10,
        },
        priority: {
            type: 'TodoPriority',
            isRequired: true,
        },
        completed: {
            type: 'boolean',
            description: `Status of the todo item.`,
            isRequired: true,
        },
        owned_by: {
            type: 'string',
            description: `ID of the user who owns the todo item.`,
            isRequired: true,
        },
        created_by: {
            type: 'string',
            description: `ID of the user who created the todo item.`,
            isRequired: true,
        },
        due_date: {
            type: 'string',
            description: `Completion due date of the todo item.`,
            isRequired: true,
            isNullable: true,
            format: 'date-time',
        },
        created_at: {
            type: 'string',
            description: `Date when the todo item was created.`,
            isRequired: true,
            format: 'date-time',
        },
        updated_at: {
            type: 'string',
            description: `Date when the todo item was updated.`,
            isRequired: true,
            isNullable: true,
            format: 'date-time',
        },
    },
} as const;
