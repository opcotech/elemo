/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { OrganizationStatus } from './OrganizationStatus';

/**
 * An organization in the system.
 */
export type Organization = {
    /**
     * Unique identifier of the organization.
     */
    id: string;
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
    logo: string | null;
    /**
     * Work title of the user.
     */
    website: string | null;
    status: OrganizationStatus;
    /**
     * IDs of the users in the organization.
     */
    members: Array<string>;
    /**
     * IDs of the teams in the organization.
     */
    teams: Array<string>;
    /**
     * IDs of the namespaces in the organization.
     */
    namespaces: Array<string>;
    /**
     * Date when the organization was created.
     */
    created_at: string;
    /**
     * Date when the organization was updated.
     */
    updated_at: string | null;
};

