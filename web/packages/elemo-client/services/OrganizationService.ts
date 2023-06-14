/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Organization } from '../models/Organization';
import type { OrganizationStatus } from '../models/OrganizationStatus';
import type { User } from '../models/User';

import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';

export class OrganizationService {
  /**
   * Get organizations
   * Returns the list of organizations in the system.
   * @param offset Number of resources to skip.
   * @param limit Number of resources to return.
   * @returns Organization OK
   * @throws ApiError
   */
  public static v1OrganizationsGet(offset?: number, limit: number = 100): CancelablePromise<Array<Organization>> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/organizations',
      query: {
        offset: offset,
        limit: limit
      },
      errors: {
        401: `Unauthorized request`,
        403: `Forbidden`,
        500: `Internal Server Error`
      }
    });
  }

  /**
   * Create organization
   * Create a new organization.
   * @param requestBody
   * @returns any Example response
   * @throws ApiError
   */
  public static v1OrganizationsCreate(requestBody?: {
    /**
     * Name of the organization.
     */
    name: string;
    /**
     * Email address of the organization.
     */
    email: string;
    /**
     * Logo of the organization.
     */
    logo?: string;
    /**
     * Work title of the user.
     */
    website?: string;
  }): CancelablePromise<{
    /**
     * ID of the newly created resource.
     */
    id: string;
  }> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/v1/organizations',
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
   * Get organization
   * Returns the given organization by its ID.
   * @param id ID of the resource.
   * @returns Organization OK
   * @throws ApiError
   */
  public static v1OrganizationGet(id: string): CancelablePromise<Organization> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/organizations/{id}',
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
   * Delete organization
   * Delete the organization by its ID.
   * @param id ID of the resource.
   * @param force Irreversibly delete the user.
   * @returns void
   * @throws ApiError
   */
  public static v1OrganizationDelete(id: string, force?: boolean): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/v1/organizations/{id}',
      path: {
        id: id
      },
      query: {
        force: force
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
   * Update organization
   * Update the organization by its ID.
   * @param id ID of the resource.
   * @param requestBody
   * @returns Organization OK
   * @throws ApiError
   */
  public static v1OrganizationUpdate(
    id: string,
    requestBody?: {
      /**
       * Name of the organization.
       */
      name?: string;
      /**
       * Email address of the organization.
       */
      email?: string;
      /**
       * Logo of the organization.
       */
      logo?: string;
      /**
       * Work title of the user.
       */
      website?: string;
      status?: OrganizationStatus;
    }
  ): CancelablePromise<Organization> {
    return __request(OpenAPI, {
      method: 'PATCH',
      url: '/v1/organizations/{id}',
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
   * Get organization members
   * Return the users that are members of the organization.
   * @param id ID of the resource.
   * @returns User OK
   * @throws ApiError
   */
  public static v1OrganizationMembersGet(id: string): CancelablePromise<Array<User>> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/organizations/{id}/members',
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
   * Add organization member
   * Add an existing user to an organization.
   * @param id ID of the resource.
   * @param requestBody
   * @returns any Example response
   * @throws ApiError
   */
  public static v1OrganizationMembersAdd(
    id: string,
    requestBody?: {
      /**
       * ID of the user to add.
       */
      user_id: string;
    }
  ): CancelablePromise<{
    /**
     * ID of the newly created resource.
     */
    id: string;
  }> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/v1/organizations/{id}/members',
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
   * Remove organization member
   * Removes a member from the organization
   * @param id ID of the resource.
   * @param userId ID of the user.
   * @returns void
   * @throws ApiError
   */
  public static v1OrganizationMembersRemove(id: string, userId: string): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/v1/organizations/{id}/members/{user_id}',
      path: {
        id: id,
        user_id: userId
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
}
