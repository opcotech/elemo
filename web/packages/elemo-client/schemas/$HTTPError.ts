/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $HTTPError = {
    description: `HTTP error description.`,
    properties: {
        message: {
            type: 'string',
            description: `Description of the error.`,
            isRequired: true,
        },
    },
} as const;
