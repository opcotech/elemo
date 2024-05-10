/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Notification } from '../models/Notification';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class NotificationService {
  /**
   * Get all in-app notification of the requesting user.
   * Returns the paginated list of in-app notifications
   * @param offset Number of resources to skip.
   * @param limit Number of resources to return.
   * @returns Notification OK
   * @throws ApiError
   */
  public static v1NotificationsGet(offset?: number, limit: number = 100): CancelablePromise<Array<Notification>> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/notifications',
      query: {
        offset: offset,
        limit: limit
      },
      errors: {
        400: `Bad request`,
        401: `Unauthorized request`,
        403: `Forbidden`,
        500: `Internal Server Error`
      }
    });
  }
  /**
   * Get an in-app notification
   * Return the requested notification by its ID.
   * @param id ID of the resource.
   * @returns Notification OK
   * @throws ApiError
   */
  public static v1NotificationGet(id: string): CancelablePromise<Notification> {
    return __request(OpenAPI, {
      method: 'GET',
      url: '/v1/notifications/{id}',
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
   * Update an in-app notification
   * Update the given user.
   * @param id ID of the resource.
   * @param requestBody
   * @returns Notification OK
   * @throws ApiError
   */
  public static v1NotificationUpdate(
    id: string,
    requestBody?: {
      /**
       * Whether the notification was read by the user.
       */
      read: boolean;
    }
  ): CancelablePromise<Notification> {
    return __request(OpenAPI, {
      method: 'PATCH',
      url: '/v1/notifications/{id}',
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
   * Delete the notification with the given ID.
   * Delete a notification by its ID.
   * @param id ID of the resource.
   * @returns void
   * @throws ApiError
   */
  public static v1NotificationDelete(id: string): CancelablePromise<void> {
    return __request(OpenAPI, {
      method: 'DELETE',
      url: '/v1/notifications/{id}',
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
}
