/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Todo } from '../models/Todo';
import type { TodoPriority } from '../models/TodoPriority';

import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';

export class TodosService {

    /**
     * Get todo item
     * Returns all todo items belonging to the current user.
     * @param offset Number of resources to skip.
     * @param limit Number of resources to return.
     * @param completed Completion status of the items.
     * @returns Todo OK
     * @throws ApiError
     */
    public static v1TodosGet(
        offset?: number,
        limit: number = 100,
        completed?: boolean,
    ): CancelablePromise<Array<Todo>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/todos',
            query: {
                'offset': offset,
                'limit': limit,
                'completed': completed,
            },
            errors: {
                401: `Unauthorized request`,
                403: `Forbidden`,
                404: `The requested resource not found`,
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Create todo item
     * Create a new todo item.
     * @param requestBody
     * @returns any Example response
     * @throws ApiError
     */
    public static v1TodosCreate(
        requestBody?: {
            /**
             * Title of the todo item.
             */
            title: string;
            /**
             * Description of the todo item.
             */
            description?: string | null;
            priority: TodoPriority;
            /**
             * ID of the user who owns the todo item.
             */
            owned_by: string;
            /**
             * Completion due date of the todo item.
             */
            due_date?: string | null;
        },
    ): CancelablePromise<{
        /**
         * ID of the newly created resource.
         */
        id: string;
    }> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/todos',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                401: `Unauthorized request`,
                403: `Forbidden`,
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Get todo item
     * Return a todo item based on the todo id belonging to the current user.
     * @param id ID of the resource.
     * @returns Todo OK
     * @throws ApiError
     */
    public static v1TodoGet(
        id: string,
    ): CancelablePromise<Todo> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/todos/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Bad request`,
                401: `Unauthorized request`,
                403: `Forbidden`,
                404: `The requested resource not found`,
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Delete todo item
     * Delete todo by its ID.
     * @param id ID of the resource.
     * @returns void
     * @throws ApiError
     */
    public static v1TodoDelete(
        id: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/v1/todos/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Bad request`,
                401: `Unauthorized request`,
                403: `Forbidden`,
                404: `The requested resource not found`,
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Update todo
     * Update the given todo
     * @param id ID of the resource.
     * @param requestBody
     * @returns Todo OK
     * @throws ApiError
     */
    public static v1TodoUpdate(
        id: string,
        requestBody?: {
            /**
             * Title of the todo item.
             */
            title?: string;
            /**
             * Description of the todo item.
             */
            description?: string | null;
            priority?: TodoPriority;
            /**
             * Completion status of the todo item.
             */
            completed?: boolean;
            /**
             * ID of the user who owns the todo item.
             */
            owned_by?: string;
            /**
             * Completion due date of the todo item.
             */
            due_date?: string | null;
        },
    ): CancelablePromise<Todo> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/v1/todos/{id}',
            path: {
                'id': id,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                401: `Unauthorized request`,
                403: `Forbidden`,
                404: `The requested resource not found`,
                500: `Internal Server Error`,
            },
        });
    }

}
