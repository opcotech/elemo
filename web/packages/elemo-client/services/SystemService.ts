/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { SystemHealth } from '../models/SystemHealth';
import type { SystemLicense } from '../models/SystemLicense';
import type { SystemVersion } from '../models/SystemVersion';

import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';

export class SystemService {

    /**
     * Get system health
     * Returns the health of registered components.
     * @returns SystemHealth OK
     * @throws ApiError
     */
    public static v1SystemHealth(): CancelablePromise<SystemHealth> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/system/health',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Get heartbeat
     * Returns 200 OK if the service is reachable.
     * @returns string OK
     * @throws ApiError
     */
    public static v1SystemHeartbeat(): CancelablePromise<'OK'> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/system/heartbeat',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Get license info
     * Return the license information. The license information is only available to entitled users.
     * @returns SystemLicense OK
     * @throws ApiError
     */
    public static v1SystemLicense(): CancelablePromise<SystemLicense> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/system/license',
            errors: {
                403: `Forbidden`,
                500: `Internal Server Error`,
            },
        });
    }

    /**
     * Get system version
     * Returns the version information of the system.
     * @returns SystemVersion OK
     * @throws ApiError
     */
    public static v1SystemVersion(): CancelablePromise<SystemVersion> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/system/version',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }

}
