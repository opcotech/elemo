/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type SystemLicense = {
    /**
     * Unique ID identifying the license.
     */
    id: string;
    /**
     * Name of the organization the license belongs to.
     */
    organization: string;
    /**
     * Email address of the licensee.
     */
    email: string;
    /**
     * Quotas available for the license.
     */
    quotas: {
        /**
         * Number of documents can exist in the system.
         */
        documents: number;
        /**
         * Number of namespaces can exist in the system.
         */
        namespaces: number;
        /**
         * Number of organizations active can exist in the system.
         */
        organizations: number;
        /**
         * Number of projects can exist in the system.
         */
        projects: number;
        /**
         * Number of roles can exist in the system.
         */
        roles: number;
        /**
         * Number of active or pending users can exist in the system.
         */
        users: number;
    };
    /**
     * Features enabled by the license.
     */
    features: Array<'components' | 'custom_statuses' | 'custom_fields' | 'multiple_assignees' | 'releases'>;
    /**
     * Date and time when the license expires.
     */
    expires_at: string;
};

