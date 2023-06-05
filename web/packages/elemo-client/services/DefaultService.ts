/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Language } from '../models/Language';
import type { User } from '../models/User';
import type { UserStatus } from '../models/UserStatus';

import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';

export class DefaultService {
  /**
   * Create new user
   * Create a new user.
   * @param requestBody
   * @returns any Example response
   * @throws ApiError
   */
  public static v1UsersCreate(requestBody?: {
    /**
     * The unique username of the user.
     */
    username: string;
    /**
     * First name of the user.
     */
    first_name?: string | null;
    /**
     * Last name of the user.
     */
    last_name?: string | null;
    /**
     * Email address of the user.
     */
    email: string;
    /**
     * Password of the user.
     */
    password: string;
    /**
     * Profile picture of the user.
     */
    picture?: string | null;
    /**
     * Work title of the user.
     */
    title?: string | null;
    /**
     * Self description of the user.
     */
    bio?: string | null;
    /**
     * Working address of the user.
     */
    address?: string | null;
    /**
     * Phone number of the user.
     */
    phone?: string | null;
    /**
     * Links to show on profile page.
     */
    links?: Array<string> | null;
    /**
     * Languages of the user.
     */
    languages?: Array<Language> | null;
  }): CancelablePromise<{
    /**
     * ID of the newly created resource.
     */
    id: string;
  }> {
    return __request(OpenAPI, {
      method: 'POST',
      url: '/v1/users',
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
   * Delete the user with the given ID.
   * Delete a user by its ID. The user is not deleted irreversibly until the "force" parameter is set to true.
   * @param id ID of the resource.
   * @param force Irreversibly delete the user.
   * @returns void
   * @throws ApiError
   */
  public static v1UserDelete(id: string, force?: boolean): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/v1/users/{id}',
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
   * Update user
   * Update the given user.
   * @param id ID of the resource.
   * @param requestBody
   * @returns User OK
   * @throws ApiError
   */
  public static v1UserUpdate(
    id: string,
    requestBody?: {
      /**
       * The unique username of the user.
       */
      username?: string;
      /**
       * First name of the user.
       */
      first_name?: string;
      /**
       * Last name of the user.
       */
      last_name?: string;
      /**
       * Email address of the user.
       */
      email?: string;
      /**
       * Password of the user.
       */
      password?: string;
      /**
       * Profile picture of the user.
       */
      picture?: string;
      /**
       * Work title of the user.
       */
      title?: string;
      /**
       * Self description of the user.
       */
      bio?: string;
      /**
       * Working address of the user.
       */
      address?: string;
      /**
       * Phone number of the user.
       */
      phone?: string;
      /**
       * Links to show on profile page.
       */
      links?: Array<string>;
      /**
       * Languages of the user.
       */
      languages?: Array<Language>;
      status?: UserStatus;
    }
  ): CancelablePromise<User> {
    return __request(OpenAPI, {
      method: 'PATCH',
      url: '/v1/users/{id}',
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
}
