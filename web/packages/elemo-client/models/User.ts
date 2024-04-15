/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Language } from './Language';
import type { UserStatus } from './UserStatus';
/**
 * A user in the system.
 */
export type User = {
  /**
   * Unique identifier of the user.
   */
  id: string;
  /**
   * The unique username of the user.
   */
  username: string;
  /**
   * First name of the user.
   */
  first_name: string | null;
  /**
   * Last name of the user.
   */
  last_name: string | null;
  /**
   * Email address of the user.
   */
  email: string;
  /**
   * Profile picture of the user.
   */
  picture: string | null;
  /**
   * Work title of the user.
   */
  title: string | null;
  /**
   * Self description of the user.
   */
  bio: string | null;
  /**
   * Working address of the user.
   */
  address: string | null;
  /**
   * Phone number of the user.
   */
  phone: string | null;
  /**
   * Links to show on profile page.
   */
  links: Array<string> | null;
  /**
   * Languages of the user.
   */
  languages: Array<Language>;
  status: UserStatus;
  /**
   * Date when the user was created.
   */
  created_at: string;
  /**
   * Date when the user was updated.
   */
  updated_at: string | null;
};
