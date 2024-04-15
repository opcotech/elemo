/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * A role in the system.
 */
export type Role = {
    /**
     * Unique identifier of the role.
     */
    id: string;
    /**
     * Name of the role.
     */
    name: string;
    /**
     * Description of the role.
     */
    description?: string | null;
    /**
     * IDs of the users assigned to the role.
     */
    members: Array<string>;
    /**
     * IDs of the permissions assigned to the role.
     */
    permissions: Array<string>;
    /**
     * Date when the organization was created.
     */
    created_at: string;
    /**
     * Date when the organization was updated.
     */
    updated_at: string | null;
};

