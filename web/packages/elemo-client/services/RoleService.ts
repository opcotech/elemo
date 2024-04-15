/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Role } from '../models/Role';
import type { User } from '../models/User';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RoleService {
    /**
     * Get organization roles
     * Return the roles that are assigned to the organization.
     * @param id ID of the resource.
     * @param offset Number of resources to skip.
     * @param limit Number of resources to return.
     * @returns Role OK
     * @throws ApiError
     */
    public static v1OrganizationRolesGet(
        id: string,
        offset?: number,
        limit: number = 100,
    ): CancelablePromise<Array<Role>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations/${id}/roles',
            path: {
                'id': id,
            },
            query: {
                'offset': offset,
                'limit': limit,
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
     * Create a new role in the organization
     * Create a new role and assign it to the organization.
     * @param id ID of the resource.
     * @param requestBody
     * @returns any Example response
     * @throws ApiError
     */
    public static v1OrganizationRolesCreate(
        id: string,
        requestBody?: {
            /**
             * Name of the role.
             */
            name: string;
            /**
             * Description of the role.
             */
            description?: string;
        },
    ): CancelablePromise<{
        /**
         * ID of the newly created resource.
         */
        id: string;
    }> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/organizations/${id}/roles',
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
    /**
     * Get organization role
     * Returns the given organization by its ID.
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @returns Role OK
     * @throws ApiError
     */
    public static v1OrganizationRoleGet(
        id: string,
        roleId: string,
    ): CancelablePromise<Role> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations/${id}/roles/{role_id}',
            path: {
                'id': id,
                'role_id': roleId,
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
     * Update organization role
     * Update the organization role by its ID.
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @param requestBody
     * @returns Role OK
     * @throws ApiError
     */
    public static v1OrganizationRoleUpdate(
        id: string,
        roleId: string,
        requestBody?: {
            /**
             * Name of the role.
             */
            name?: string;
            /**
             * Description of the role.
             */
            description?: string;
        },
    ): CancelablePromise<Role> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/v1/organizations/${id}/roles/{role_id}',
            path: {
                'id': id,
                'role_id': roleId,
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
    /**
     * Delete organization role
     * Deletes a role that is assigned to the organization.
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @returns void
     * @throws ApiError
     */
    public static v1OrganizationRoleDelete(
        id: string,
        roleId: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/v1/organizations/${id}/roles/{role_id}',
            path: {
                'id': id,
                'role_id': roleId,
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
     * Get organization role members
     * Return the users that are members of the organization's role.
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @returns User OK
     * @throws ApiError
     */
    public static v1OrganizationRoleMembersGet(
        id: string,
        roleId: string,
    ): CancelablePromise<Array<User>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations/{id}/roles/{role_id}/members',
            path: {
                'id': id,
                'role_id': roleId,
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
     * Add organization role member
     * Add an existing user to an organization's role.
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @param requestBody
     * @returns any Example response
     * @throws ApiError
     */
    public static v1OrganizationRoleMembersAdd(
        id: string,
        roleId: string,
        requestBody?: {
            /**
             * ID of the user to add.
             */
            user_id: string;
        },
    ): CancelablePromise<{
        /**
         * ID of the newly created resource.
         */
        id: string;
    }> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/organizations/{id}/roles/{role_id}/members',
            path: {
                'id': id,
                'role_id': roleId,
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
    /**
     * Remove organization role member
     * Removes a member from the organization's role
     * @param id ID of the resource.
     * @param roleId ID of the role.
     * @param userId ID of the user.
     * @returns void
     * @throws ApiError
     */
    public static v1OrganizationRoleMemberRemove(
        id: string,
        roleId: string,
        userId: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/v1/organizations/{id}/roles/{role_id}/members/{user_id}',
            path: {
                'id': id,
                'role_id': roleId,
                'user_id': userId,
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
}
