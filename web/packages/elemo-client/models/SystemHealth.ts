/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type SystemHealth = {
    /**
     * Health of the cache database.
     */
    cache_database: SystemHealth.cache_database;
    /**
     * Health of the graph database.
     */
    graph_database: SystemHealth.graph_database;
    /**
     * Health of the relational database.
     */
    relational_database: SystemHealth.relational_database;
    /**
     * Health of the license.
     */
    license: SystemHealth.license;
    /**
     * Health of the message queue.
     */
    message_queue: SystemHealth.message_queue;
};
export namespace SystemHealth {
    /**
     * Health of the cache database.
     */
    export enum cache_database {
        HEALTHY = 'healthy',
        UNHEALTHY = 'unhealthy',
        UNKNOWN = 'unknown',
    }
    /**
     * Health of the graph database.
     */
    export enum graph_database {
        HEALTHY = 'healthy',
        UNHEALTHY = 'unhealthy',
        UNKNOWN = 'unknown',
    }
    /**
     * Health of the relational database.
     */
    export enum relational_database {
        HEALTHY = 'healthy',
        UNHEALTHY = 'unhealthy',
        UNKNOWN = 'unknown',
    }
    /**
     * Health of the license.
     */
    export enum license {
        HEALTHY = 'healthy',
        UNHEALTHY = 'unhealthy',
        UNKNOWN = 'unknown',
    }
    /**
     * Health of the message queue.
     */
    export enum message_queue {
        HEALTHY = 'healthy',
        UNHEALTHY = 'unhealthy',
        UNKNOWN = 'unknown',
    }
}

