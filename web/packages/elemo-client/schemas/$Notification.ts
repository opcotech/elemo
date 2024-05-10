/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export const $Notification = {
  description: `An in-app notification sent to the user.`,
  properties: {
    id: {
      type: 'string',
      description: `Unique identifier of the in-app notification.`,
      isRequired: true
    },
    title: {
      type: 'string',
      description: `Title of the in-app notification.`,
      isRequired: true,
      maxLength: 120,
      minLength: 3
    },
    description: {
      type: 'string',
      description: `Description of the in-app notification.`,
      isRequired: true,
      maxLength: 500,
      minLength: 5
    },
    recipient: {
      type: 'string',
      description: `ID of the user who got notified.`,
      isRequired: true
    },
    read: {
      type: 'boolean',
      description: `Whether the notification was read by the user.`,
      isRequired: true
    },
    created_at: {
      type: 'string',
      description: `Date when the todo item was created.`,
      isRequired: true,
      format: 'date-time'
    },
    updated_at: {
      type: 'string',
      description: `Date when the in-app notification was updated.`,
      isRequired: true,
      isNullable: true,
      format: 'date-time'
    }
  }
} as const;
