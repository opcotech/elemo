/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $SystemHealth = {
    properties: {
        cache_database: {
            type: 'Enum',
            isRequired: true,
        },
        graph_database: {
            type: 'Enum',
            isRequired: true,
        },
        relational_database: {
            type: 'Enum',
            isRequired: true,
        },
        license: {
            type: 'Enum',
            isRequired: true,
        },
        message_queue: {
            type: 'Enum',
            isRequired: true,
        },
    },
} as const;
