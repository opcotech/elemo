/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Permission } from '../models/Permission';
import type { PermissionKind } from '../models/PermissionKind';
import type { ResourceType } from '../models/ResourceType';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class PermissionService {
  /**
   * Create permission
   * Create a new permission for a subject to the given target.
   * @param requestBody
   * @returns any Example response
   * @throws ApiError
   */
  public static v1PermissionsCreate(requestBody?: {
    kind: PermissionKind;
    subject: {
      resourceType: ResourceType;
      id: string;
    };
    target: {
      resourceType: ResourceType;
      id: string;
    };
  }): CancelablePromise<{
    /**
     * ID of the newly created resource.
     */
    id: string;
  }> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/v1/permissions',
      body: requestBody,
      mediaType: 'application/json',
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        403: `Forbidden`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Update permission
   * Update a permission.
   * @param id ID of the resource.
   * @param requestBody
   * @returns Permission OK
   * @throws ApiError
   */
  public static v1PermissionUpdate(
    id: string,
    requestBody?: {
      kind: PermissionKind;
    }
  ): CancelablePromise<Permission> {
    return __request(OpenAPI, {
      method: 'PATCH',
      url: '/v1/permissions/{id}',
      path: {
        id: id
      },
      body: requestBody,
      mediaType: 'application/json',
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        403: `Forbidden`,
        404: `The requested resource not found`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Delete permission
   * Delete a permission by its ID.
   * @param id ID of the resource.
   * @returns void
   * @throws ApiError
   */
  public static v1PermissionDelete(id: string): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/v1/permissions/{id}',
      path: {
        id: id
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        403: `Forbidden`,
        404: `The requested resource not found`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Get permission
   * Get a permission by its ID.
   * @param id ID of the resource.
   * @returns Permission OK
   * @throws ApiError
   */
  public static v1PermissionGet(id: string): CancelablePromise<Permission> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/permissions/{id}',
      path: {
        id: id
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        404: `The requested resource not found`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Get permissions for a resource
   * Get all permissions the caller have for a given resource.
   * @param resourceId ID of the resource combined with its resource type.
   * @returns Permission OK
   * @throws ApiError
   */
  public static v1PermissionResourceGet(resourceId: string): CancelablePromise<Array<Permission>> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/permissions/resources/{resourceId}',
      path: {
        resourceId: resourceId
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        403: `Forbidden`,
        404: `The requested resource not found`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Check relations to resource
   * Check if the caller has any relations to a given resource.
   * @param resourceId ID of the resource combined with its resource type.
   * @returns boolean OK
   * @throws ApiError
   */
  public static v1PermissionHasRelations(resourceId: string): CancelablePromise<boolean> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/permissions/has-relations/{resourceId}',
      path: {
        resourceId: resourceId
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        404: `The requested resource not found`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Check system role assignment
   * Check if the user is member of one or more system roles.
   * @param roles ID of a role.
   * @returns boolean OK
   * @throws ApiError
   */
  public static v1PermissionHasSystemRole(roles: Array<'Owner' | 'Admin' | 'Support'>): CancelablePromise<boolean> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/permissions/has-system-role',
      query: {
        roles: roles
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        500: `Internal Server Error`
      }
    });
  }
}
