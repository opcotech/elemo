/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { PermissionKind } from './PermissionKind';

/**
 * A permission in the system.
 */
export type Permission = {
  /**
   * Unique identifier of the user.
   */
  id: string;
  kind: PermissionKind;
  subject: string;
  target: string;
  /**
   * Date when the user was created.
   */
  created_at: string;
  /**
   * Date when the user was updated.
   */
  updated_at: string | null;
};
