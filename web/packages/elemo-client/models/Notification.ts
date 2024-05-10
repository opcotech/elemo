/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * An in-app notification sent to the user.
 */
export type Notification = {
  /**
   * Unique identifier of the in-app notification.
   */
  id: string;
  /**
   * Title of the in-app notification.
   */
  title: string;
  /**
   * Description of the in-app notification.
   */
  description: string;
  /**
   * ID of the user who got notified.
   */
  recipient: string;
  /**
   * Whether the notification was read by the user.
   */
  read: boolean;
  /**
   * Date when the todo item was created.
   */
  created_at: string;
  /**
   * Date when the in-app notification was updated.
   */
  updated_at: string | null;
};
