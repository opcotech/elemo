/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

/** A user in the system. */
export interface User {
  /** Unique ID identifying the given user. */
  readonly id?: string;
  /**
   * Username of the user.
   * @minLength 3
   * @maxLength 50
   * @pattern ^[a-z0-9-_]{3,50}$
   */
  username?: string;
  /**
   * First name of the user.
   * @minLength 1
   * @maxLength 50
   */
  first_name?: string;
  /**
   * Last name of the user.
   * @minLength 1
   * @maxLength 50
   */
  last_name?: string;
  /**
   * Email address of the user.
   * @format email
   * @minLength 6
   * @maxLength 254
   */
  email?: string;
  /**
   * Work title of the user.
   * @minLength 3
   * @maxLength 50
   */
  title?: string;
  /**
   * Password of the user.
   * @format password
   * @minLength 8
   * @maxLength 64
   */
  password?: string;
  /**
   * Profile picture URL of the user.
   * @format uri
   * @maxLength 2000
   */
  picture?: string;
  /**
   * Self
   * @maxLength 500
   */
  bio?: string;
  /**
   * Working address of the user (if not remote).
   * @maxLength 500
   */
  address?: string;
  /**
   * Phone number of the user
   * @minLength 7
   * @maxLength 16
   */
  phone?: string;
  /** Links to show on profile page. */
  links?: string[];
  /** Languages of the user. */
  languages?: Language[];
  /** Status of the user. */
  status?: UserStatus;
  /**
   * Date when the user was created.
   * @format date-time
   */
  readonly created_at?: string;
  /**
   * Last date of user modification. If the user has never been modified, the field is not present.
   * @format date-time
   */
  readonly updated_at?: string;
}

/** Status of the user. */
export enum UserStatus {
  Active = 'active',
  Pending = 'pending',
  Inactive = 'inactive',
  Deleted = 'deleted'
}

/**
 * Two-letter ISO language code.
 * @minLength 2
 * @maxLength 2
 */
export enum Language {
  Aa = 'aa',
  Ab = 'ab',
  Ae = 'ae',
  Af = 'af',
  Ak = 'ak',
  Am = 'am',
  An = 'an',
  Ar = 'ar',
  As = 'as',
  Av = 'av',
  Ay = 'ay',
  Az = 'az',
  Ba = 'ba',
  Be = 'be',
  Bg = 'bg',
  Bh = 'bh',
  Bi = 'bi',
  Bm = 'bm',
  Bn = 'bn',
  Bo = 'bo',
  Br = 'br',
  Bs = 'bs',
  Ca = 'ca',
  Ce = 'ce',
  Ch = 'ch',
  Co = 'co',
  Cr = 'cr',
  Cs = 'cs',
  Cu = 'cu',
  Cv = 'cv',
  Cy = 'cy',
  Da = 'da',
  De = 'de',
  Dv = 'dv',
  Dz = 'dz',
  Ee = 'ee',
  El = 'el',
  En = 'en',
  Eo = 'eo',
  Es = 'es',
  Et = 'et',
  Eu = 'eu',
  Fa = 'fa',
  Ff = 'ff',
  Fi = 'fi',
  Fj = 'fj',
  Fo = 'fo',
  Fr = 'fr',
  Fy = 'fy',
  Ga = 'ga',
  Gd = 'gd',
  Gl = 'gl',
  Gn = 'gn',
  Gu = 'gu',
  Gv = 'gv',
  Ha = 'ha',
  He = 'he',
  Hi = 'hi',
  Ho = 'ho',
  Hr = 'hr',
  Ht = 'ht',
  Hu = 'hu',
  Hy = 'hy',
  Hz = 'hz',
  Ia = 'ia',
  Id = 'id',
  Ie = 'ie',
  Ig = 'ig',
  Ii = 'ii',
  Ik = 'ik',
  Io = 'io',
  Is = 'is',
  It = 'it',
  Iu = 'iu',
  Ja = 'ja',
  Jv = 'jv',
  Ka = 'ka',
  Kg = 'kg',
  Ki = 'ki',
  Kj = 'kj',
  Kk = 'kk',
  Kl = 'kl',
  Km = 'km',
  Kn = 'kn',
  Ko = 'ko',
  Kr = 'kr',
  Ks = 'ks',
  Ku = 'ku',
  Kv = 'kv',
  Kw = 'kw',
  Ky = 'ky',
  La = 'la',
  Lb = 'lb',
  Lg = 'lg',
  Li = 'li',
  Ln = 'ln',
  Lo = 'lo',
  Lt = 'lt',
  Lu = 'lu',
  Lv = 'lv',
  Mg = 'mg',
  Mh = 'mh',
  Mi = 'mi',
  Mk = 'mk',
  Ml = 'ml',
  Mn = 'mn',
  Mr = 'mr',
  Ms = 'ms',
  Mt = 'mt',
  My = 'my',
  Na = 'na',
  Nb = 'nb',
  Nd = 'nd',
  Ne = 'ne',
  Ng = 'ng',
  Nl = 'nl',
  Nn = 'nn',
  No = 'no',
  Nr = 'nr',
  Nv = 'nv',
  Ny = 'ny',
  Oc = 'oc',
  Oj = 'oj',
  Om = 'om',
  Or = 'or',
  Os = 'os',
  Pa = 'pa',
  Pi = 'pi',
  Pl = 'pl',
  Ps = 'ps',
  Pt = 'pt',
  Qu = 'qu',
  Rm = 'rm',
  Rn = 'rn',
  Ro = 'ro',
  Ru = 'ru',
  Rw = 'rw',
  Sa = 'sa',
  Sc = 'sc',
  Sd = 'sd',
  Se = 'se',
  Sg = 'sg',
  Si = 'si',
  Sk = 'sk',
  Sl = 'sl',
  Sm = 'sm',
  Sn = 'sn',
  So = 'so',
  Sq = 'sq',
  Sr = 'sr',
  Ss = 'ss',
  St = 'st',
  Su = 'su',
  Sv = 'sv',
  Sw = 'sw',
  Ta = 'ta',
  Te = 'te',
  Tg = 'tg',
  Th = 'th',
  Ti = 'ti',
  Tk = 'tk',
  Tl = 'tl',
  Tn = 'tn',
  To = 'to',
  Tr = 'tr',
  Ts = 'ts',
  Tt = 'tt',
  Tw = 'tw',
  Ty = 'ty',
  Ug = 'ug',
  Uk = 'uk',
  Ur = 'ur',
  Uz = 'uz',
  Ve = 've',
  Vi = 'vi',
  Vo = 'vo',
  Wa = 'wa',
  Wo = 'wo',
  Xh = 'xh',
  Yi = 'yi',
  Yo = 'yo',
  Za = 'za',
  Zh = 'zh',
  Zu = 'zu'
}

/** Todo item belonging to a user. */
export interface Todo {
  /** Unique identifier of the todo item. */
  readonly id?: string;
  /**
   * Title of the todo item.
   * @max 250
   * @minLength 3
   */
  title?: string;
  /**
   * Description of the todo item.
   * @min 10
   * @maxLength 500
   */
  description?: string;
  /** Priority of a todo item. */
  priority?: TodoPriority;
  /**
   * Status of the todo item.
   * @default false
   */
  completed?: boolean;
  /** ID of the user who owns the todo item. */
  owned_by?: string;
  /** ID of the user who created the todo item. */
  readonly created_by?: string;
  /**
   * Completion due date of the todo item.
   * @format date-time
   */
  due_date?: string;
  /**
   * Date when the todo item was created.
   * @format date-time
   */
  readonly created_at?: string;
  /**
   * Last date of todo item modification.
   * @format date-time
   */
  readonly updated_at?: string;
}

/** Priority of a todo item. */
export enum TodoPriority {
  Normal = 'normal',
  Important = 'important',
  Urgent = 'urgent',
  Critical = 'critical'
}

/** Registered license. */
export interface SystemLicense {
  /** Unique ID identifying the license. */
  id: string;
  /** Organization name. */
  organization: string;
  /**
   * Username of the user.
   * @format email
   * @minLength 6
   * @maxLength 254
   */
  email: string;
  /** Quotas available for the license. */
  quotas: {
    /**
     * Number of documents can exist in the system.
     * @min 1
     */
    documents: number;
    /**
     * Number of namespaces can exist in the system.
     * @min 1
     */
    namespaces: number;
    /**
     * Number of organizations active can exist in the system.
     * @min 1
     */
    organizations: number;
    /**
     * Number of projects can exist in the system.
     * @min 1
     */
    projects: number;
    /**
     * Number of roles can exist in the system.
     * @min 1
     */
    roles: number;
    /**
     * Number of active or pending users can exist in the system.
     * @min 1
     */
    users: number;
  };
  /** Features enabled for the license. */
  features: SystemLicenseFeatures[];
  /**
   * Date when the license expires.
   * @format date-time
   */
  expires_at: string;
}

/** Health of the system components. */
export interface SystemHealth {
  /** Health of the cache database. */
  cache_database: SystemHealthCacheDatabase;
  /** Health of the graph database. */
  graph_database: SystemHealthGraphDatabase;
  /** Health of the relational database. */
  relational_database: SystemHealthRelationalDatabase;
  /** Health of the license based on its validity. */
  license: SystemHealthLicense;
  /** Health of the message queue. */
  message_queue: SystemHealthMessageQueue;
}

/** Heartbeat response of the system. */
export enum SystemHeartbeat {
  OK = 'OK'
}

/** Version information of the system. */
export interface SystemVersionInfo {
  /** Version of the application. */
  version: string;
  /** Commit hash of the build. */
  commit: string;
  /** Build date of the application. */
  date: string;
  /** Go version used to build the application. */
  go_version: string;
}

export interface HTTPError {
  message: string;
}

export enum SystemLicenseFeatures {
  Components = 'components',
  CustomStatuses = 'custom_statuses',
  CustomFields = 'custom_fields',
  MultipleAssignees = 'multiple_assignees',
  Releases = 'releases'
}

/** Health of the cache database. */
export enum SystemHealthCacheDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/** Health of the graph database. */
export enum SystemHealthGraphDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/** Health of the relational database. */
export enum SystemHealthRelationalDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/** Health of the license based on its validity. */
export enum SystemHealthLicense {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/** Health of the message queue. */
export enum SystemHealthMessageQueue {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

export interface GetUsersParams {
  /**
   * Maximum number of results to return.
   * @min 1
   * @max 1000
   * @default 100
   */
  limit?: number;
  /**
   * Results to skip when paginating through a result set.
   * @min 0
   * @default 0
   */
  offset?: number;
}

export type GetUsersData = User[];

export interface CreateUserData {
  user_id: string;
}

export type GetUserData = User;

export type UpdateUserData = User;

export interface DeleteUserParams {
  /**
   * Force the operation. If set to true, the operation will be performed regardless its consequences. In case of some resources, this flag will call the delete operation on the resource instead of updating its status.
   * @default false
   */
  force?: boolean;
  /** Unique ID of the user. The ID may be set to `me` to get the current user. */
  userId: string;
}

export interface GetTodosParams {
  /**
   * Maximum number of results to return.
   * @min 1
   * @max 1000
   * @default 100
   */
  limit?: number;
  /**
   * Results to skip when paginating through a result set.
   * @min 0
   * @default 0
   */
  offset?: number;
  /** Filter by completed status. */
  completed?: boolean;
}

export type GetTodosData = Todo[];

export interface CreateTodoData {
  todo_id: string;
}

export type GetTodoData = Todo;

export type UpdateTodoData = Todo;

export type GetSystemHealthData = SystemHealth;

export type GetSystemHeartbeatData = SystemHeartbeat;

export type GetSystemVersionData = SystemVersionInfo;

export type GetSystemLicenseData = SystemLicense;

export type QueryParamsType = Record<string | number, any>;
export type ResponseFormat = keyof Omit<Body, 'body' | 'bodyUsed'>;

export interface FullRequestParams extends Omit<RequestInit, 'body'> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseFormat;
  /** request body */
  body?: unknown;
  /** base url */
  baseUrl?: string;
  /** request cancellation token */
  cancelToken?: CancelToken;
}

export type RequestParams = Omit<FullRequestParams, 'body' | 'method' | 'query' | 'path'>;

export interface ApiConfig<SecurityDataType = unknown> {
  baseUrl?: string;
  baseApiParams?: Omit<RequestParams, 'baseUrl' | 'cancelToken' | 'signal'>;
  securityWorker?: (securityData: SecurityDataType | null) => Promise<RequestParams | void> | RequestParams | void;
  customFetch?: typeof fetch;
}

export interface HttpResponse<D extends unknown, E extends unknown = unknown> extends Response {
  data: D;
  error: E;
}

type CancelToken = Symbol | string | number;

export enum ContentType {
  Json = 'application/json',
  FormData = 'multipart/form-data',
  UrlEncoded = 'application/x-www-form-urlencoded',
  Text = 'text/plain'
}

export class HttpClient<SecurityDataType = unknown> {
  public baseUrl: string = 'https://{site}.elemo.app/api';
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>['securityWorker'];
  private abortControllers = new Map<CancelToken, AbortController>();
  private customFetch = (...fetchParams: Parameters<typeof fetch>) => fetch(...fetchParams);

  private baseApiParams: RequestParams = {
    credentials: 'same-origin',
    headers: {},
    redirect: 'follow',
    referrerPolicy: 'no-referrer'
  };

  constructor(apiConfig: ApiConfig<SecurityDataType> = {}) {
    Object.assign(this, apiConfig);
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected encodeQueryParam(key: string, value: any) {
    const encodedKey = encodeURIComponent(key);
    return `${encodedKey}=${encodeURIComponent(typeof value === 'number' ? value : `${value}`)}`;
  }

  protected addQueryParam(query: QueryParamsType, key: string) {
    return this.encodeQueryParam(key, query[key]);
  }

  protected addArrayQueryParam(query: QueryParamsType, key: string) {
    const value = query[key];
    return value.map((v: any) => this.encodeQueryParam(key, v)).join('&');
  }

  protected toQueryString(rawQuery?: QueryParamsType): string {
    const query = rawQuery || {};
    const keys = Object.keys(query).filter((key) => 'undefined' !== typeof query[key]);
    return keys
      .map((key) => (Array.isArray(query[key]) ? this.addArrayQueryParam(query, key) : this.addQueryParam(query, key)))
      .join('&');
  }

  protected addQueryParams(rawQuery?: QueryParamsType): string {
    const queryString = this.toQueryString(rawQuery);
    return queryString ? `?${queryString}` : '';
  }

  private contentFormatters: Record<ContentType, (input: any) => any> = {
    [ContentType.Json]: (input: any) =>
      input !== null && (typeof input === 'object' || typeof input === 'string') ? JSON.stringify(input) : input,
    [ContentType.Text]: (input: any) => (input !== null && typeof input !== 'string' ? JSON.stringify(input) : input),
    [ContentType.FormData]: (input: any) =>
      Object.keys(input || {}).reduce((formData, key) => {
        const property = input[key];
        formData.append(
          key,
          property instanceof Blob
            ? property
            : typeof property === 'object' && property !== null
            ? JSON.stringify(property)
            : `${property}`
        );
        return formData;
      }, new FormData()),
    [ContentType.UrlEncoded]: (input: any) => this.toQueryString(input)
  };

  protected mergeRequestParams(params1: RequestParams, params2?: RequestParams): RequestParams {
    return {
      ...this.baseApiParams,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...(this.baseApiParams.headers || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {})
      }
    };
  }

  protected createAbortSignal = (cancelToken: CancelToken): AbortSignal | undefined => {
    if (this.abortControllers.has(cancelToken)) {
      const abortController = this.abortControllers.get(cancelToken);
      if (abortController) {
        return abortController.signal;
      }
      return void 0;
    }

    const abortController = new AbortController();
    this.abortControllers.set(cancelToken, abortController);
    return abortController.signal;
  };

  public abortRequest = (cancelToken: CancelToken) => {
    const abortController = this.abortControllers.get(cancelToken);

    if (abortController) {
      abortController.abort();
      this.abortControllers.delete(cancelToken);
    }
  };

  public request = async <T = any, E = any>({
    body,
    secure,
    path,
    type,
    query,
    format,
    baseUrl,
    cancelToken,
    ...params
  }: FullRequestParams): Promise<HttpResponse<T, E>> => {
    const secureParams =
      ((typeof secure === 'boolean' ? secure : this.baseApiParams.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const queryString = query && this.toQueryString(query);
    const payloadFormatter = this.contentFormatters[type || ContentType.Json];
    const responseFormat = format || requestParams.format;

    return this.customFetch(`${baseUrl || this.baseUrl || ''}${path}${queryString ? `?${queryString}` : ''}`, {
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type && type !== ContentType.FormData ? { 'Content-Type': type } : {})
      },
      signal: cancelToken ? this.createAbortSignal(cancelToken) : requestParams.signal,
      body: typeof body === 'undefined' || body === null ? null : payloadFormatter(body)
    }).then(async (response) => {
      const r = response as HttpResponse<T, E>;
      r.data = null as unknown as T;
      r.error = null as unknown as E;

      const data = !responseFormat
        ? r
        : await response[responseFormat]()
            .then((data) => {
              if (r.ok) {
                r.data = data;
              } else {
                r.error = data;
              }
              return r;
            })
            .catch((e) => {
              r.error = e;
              return r;
            });

      if (cancelToken) {
        this.abortControllers.delete(cancelToken);
      }

      if (!response.ok) throw data;
      return data;
    });
  };
}

/**
 * @title Elemo API
 * @version 0.1.0
 * @license Apache 2.0 (https://elemo.app/license)
 * @termsOfService https://elemo.app/terms-of-service
 * @baseUrl https://{site}.elemo.app/api
 * @contact Elemo Support <help@elemo.app> (https://elemo.app/contact)
 *
 * # Introduction
 *
 * The Elemo API allows you to manage users, organizations and other resources within Elemo in a programmatic way. The API is capable of doing all operations that can be executed from the user interface.
 *
 * You may use any tool that handles HTTP requests to interact with the API. However, the requests should be made using the HTTPS protocol so that traffic is encrypted.
 *
 * You need to obtain an [Access Token](https://en.wikipedia.org/wiki/Access_token) to call most of the API endpoints. The tokens are bound to users, therefore you must have a user in the system as well. Read more about obtaining an access token below.
 *
 * ## Requests
 *
 * The endpoints have support for the HTTP methods below. Please note that not all endpoints are supporting every HTTP method.
 *
 * | Method | Usage                                                                                                                        |
 * |----------|------------------------------------------------------------------------------------------------------------------------------|
 * | `GET`    | Used to retrieve information about one or many resources.                                                                    |
 * | `POST`   | Creates a new resource. The request must include all required attributes.                                                    |
 * | `PUT`    | Updates an existing resource. The request must include all required attributes.                                              |
 * | `PATCH`  | Partially updates an existing resource. The request attributes are not required. Most of the resources are supporting PATCH. |
 * | `DELETE` | Delete a resource from the system. Usually, this is an irreversible action.                                                  |
 *
 * ## Authentication
 *
 * Authentication is implemented based on the [OAuth 2.0](https://oauth.net/2/) specification supporting the password and authorization code flows. As a rule of thumb, whenever you need to interact with the API use authorization code flow and fallback to password flow if other flows cannot be implemented for any reason.
 *
 * After the token obtained, use the token as part of the `Authorization` HTTP header in the format of: `Authorization: Bearer {access_token}`
 *
 * ## Pagination
 *
 * All endpoints that are returning a list of resources supports and requires pagination. The default page size is `100` items, controlled by the `limit` query parameter. To utilize pagination, you may define `offset` in addition which skips the defined results. Therefore, if `offset` is set to `100`, the endpoint will return the next page's results. Sample request:
 *
 * ```shell
 * $ curl -H "Authorization: Bearer {access_token}" https://{site}.elemo.app/api/v1/users?limit=100&offset=100
 * ```
 *
 * Although the pagination is very flexible, it defines some constraints:
 *
 * * The maximum `limit` cannot be greater than `1000`
 * * The minimum `limit` cannot be less than `1`
 * * The minimum `offset` cannot be less than `0`
 *
 * ## Versioning
 *
 * ### APIs
 *
 * The endpoints are versioned and the version number is part of the path. When the required input or returned output of an endpoint is changed, the current version is being deprecated and a new version of the endpoint is created. Deprecated endpoints are removed when a new major application version is released.
 *
 * ### This specification
 *
 * In contrast with the APIs, this specification follows semantic versioning.
 */
export class Client<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
  v1 = {
    /**
     * @description Return all users within the given limit starting at the optionally defined offset.
     *
     * @tags Users
     * @name GetUsers
     * @summary Get all users.
     * @request GET:/v1/users
     * @secure
     * @response `200` `GetUsersData` List of users.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `default` `HTTPError`
     */
    getUsers: (query: GetUsersParams, params: RequestParams = {}) =>
      this.request<GetUsersData, HTTPError>({
        path: `/v1/users`,
        method: 'GET',
        query: query,
        secure: true,
        ...params
      }),

    /**
     * @description Creates a new user in the system, then returns its ID.
     *
     * @tags Users
     * @name CreateUser
     * @summary Create a new user.
     * @request POST:/v1/users
     * @secure
     * @response `201` `CreateUserData` Returns the ID of the newly created user.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `default` `HTTPError`
     */
    createUser: (data: User, params: RequestParams = {}) =>
      this.request<CreateUserData, HTTPError>({
        path: `/v1/users`,
        method: 'POST',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Returns the user specified in the request path.
     *
     * @tags Users
     * @name GetUser
     * @summary Get user by ID.
     * @request GET:/v1/users/{user_id}
     * @secure
     * @response `200` `GetUserData` The user object if exists.
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    getUser: (userId: string, params: RequestParams = {}) =>
      this.request<GetUserData, HTTPError>({
        path: `/v1/users/${userId}`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Updates an existing user by its ID.
     *
     * @tags Users
     * @name UpdateUser
     * @summary Update a user by its ID.
     * @request PATCH:/v1/users/{user_id}
     * @secure
     * @response `200` `UpdateUserData` The updated user object.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    updateUser: (
      userId: string,
      data: {
        /**
         * Username of the user.
         * @minLength 3
         * @maxLength 50
         * @pattern ^[a-z0-9-_]{3,50}$
         */
        username?: string;
        /**
         * First name of the user.
         * @minLength 1
         * @maxLength 50
         */
        first_name?: string;
        /**
         * Last name of the user.
         * @minLength 1
         * @maxLength 50
         */
        last_name?: string;
        /**
         * Email address of the user.
         * @format email
         * @minLength 6
         * @maxLength 254
         */
        email?: string;
        /**
         * Work title of the user.
         * @minLength 3
         * @maxLength 50
         */
        title?: string;
        /**
         * Password of the user.
         * @format password
         * @minLength 8
         * @maxLength 64
         */
        password?: string;
        /**
         * Profile picture URL of the user.
         * @format uri
         * @maxLength 2000
         */
        picture?: string;
        /**
         * Self
         * @maxLength 500
         */
        bio?: string;
        /**
         * Working address of the user (if not remote).
         * @maxLength 500
         */
        address?: string;
        /**
         * Phone number of the user
         * @minLength 7
         * @maxLength 16
         */
        phone?: string;
        /** Links to show on profile page. */
        links?: string[];
        /** Languages of the user. */
        languages?: Language[];
        /** Status of the user. */
        status?: UserStatus;
      },
      params: RequestParams = {}
    ) =>
      this.request<UpdateUserData, HTTPError>({
        path: `/v1/users/${userId}`,
        method: 'PATCH',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Deletes the user specified in the request path.
     *
     * @tags Users
     * @name DeleteUser
     * @summary Delete user by its ID.
     * @request DELETE:/v1/users/{user_id}
     * @secure
     * @response `204` `string | null`
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    deleteUser: ({ userId, ...query }: DeleteUserParams, params: RequestParams = {}) =>
      this.request<string | null, HTTPError>({
        path: `/v1/users/${userId}`,
        method: 'DELETE',
        query: query,
        secure: true,
        ...params
      }),

    /**
     * @description Return all todo items within the given limit starting at the optionally defined offset for the current user.
     *
     * @tags Todos
     * @name GetTodos
     * @summary Get user's todo items.
     * @request GET:/v1/todos
     * @secure
     * @response `200` `GetTodosData` The list of todo items belonging to the current user.
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    getTodos: (query: GetTodosParams, params: RequestParams = {}) =>
      this.request<GetTodosData, HTTPError>({
        path: `/v1/todos`,
        method: 'GET',
        query: query,
        secure: true,
        ...params
      }),

    /**
     * @description Creates a new todo item.
     *
     * @tags Todos
     * @name CreateTodo
     * @summary Create a new todo item.
     * @request POST:/v1/todos
     * @secure
     * @response `201` `CreateTodoData` Returns the newly created todo item's ID.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `default` `HTTPError`
     */
    createTodo: (data: Todo, params: RequestParams = {}) =>
      this.request<CreateTodoData, HTTPError>({
        path: `/v1/todos`,
        method: 'POST',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Return a todo item by its ID.
     *
     * @tags Todos
     * @name GetTodo
     * @summary Get todo item by ID.
     * @request GET:/v1/todos/{todo_id}
     * @secure
     * @response `200` `GetTodoData` The list of todo items belonging to the current user.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    getTodo: (todoId: string, params: RequestParams = {}) =>
      this.request<GetTodoData, HTTPError>({
        path: `/v1/todos/${todoId}`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Updates an existing todo item by its ID.
     *
     * @tags Todos
     * @name UpdateTodo
     * @summary Update todo item by ID.
     * @request PATCH:/v1/todos/{todo_id}
     * @secure
     * @response `200` `UpdateTodoData` The updated todo item.
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    updateTodo: (todoId: string, data: Todo, params: RequestParams = {}) =>
      this.request<UpdateTodoData, HTTPError>({
        path: `/v1/todos/${todoId}`,
        method: 'PATCH',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Delete a todo item by its ID.
     *
     * @tags Todos
     * @name DeleteTodo
     * @summary Delete todo item by ID.
     * @request DELETE:/v1/todos/{todo_id}
     * @secure
     * @response `204` `string | null`
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    deleteTodo: (todoId: string, params: RequestParams = {}) =>
      this.request<string | null, HTTPError>({
        path: `/v1/todos/${todoId}`,
        method: 'DELETE',
        secure: true,
        ...params
      }),

    /**
     * @description Returns the system health, listing all components and their status.
     *
     * @tags System
     * @name GetSystemHealth
     * @summary Get system status.
     * @request GET:/v1/system/health
     * @response `200` `GetSystemHealthData` The server can reply to health requests.
     * @response `default` `HTTPError`
     */
    getSystemHealth: (params: RequestParams = {}) =>
      this.request<GetSystemHealthData, HTTPError>({
        path: `/v1/system/health`,
        method: 'GET',
        ...params
      }),

    /**
     * @description Returns HTTP OK response if the server is up and running, however it doesn't mean that the server is healthy.
     *
     * @tags System
     * @name GetSystemHeartbeat
     * @summary Get heartbeat.
     * @request GET:/v1/system/heartbeat
     * @response `200` `GetSystemHeartbeatData` Response indicating that the server is running.
     * @response `default` `HTTPError`
     */
    getSystemHeartbeat: (params: RequestParams = {}) =>
      this.request<GetSystemHeartbeatData, HTTPError>({
        path: `/v1/system/heartbeat`,
        method: 'GET',
        ...params
      }),

    /**
     * @description Returns the version information and build details.
     *
     * @tags System
     * @name GetSystemVersion
     * @summary Get application version.
     * @request GET:/v1/system/version
     * @response `200` `GetSystemVersionData` Version information and build details of the system.
     * @response `default` `HTTPError`
     */
    getSystemVersion: (params: RequestParams = {}) =>
      this.request<GetSystemVersionData, HTTPError>({
        path: `/v1/system/version`,
        method: 'GET',
        ...params
      }),

    /**
     * @description Returns the registered license information.
     *
     * @tags System
     * @name GetSystemLicense
     * @summary Get the registered license.
     * @request GET:/v1/system/license
     * @secure
     * @response `200` `GetSystemLicenseData` The registered license.
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `default` `HTTPError`
     */
    getSystemLicense: (params: RequestParams = {}) =>
      this.request<GetSystemLicenseData, HTTPError>({
        path: `/v1/system/license`,
        method: 'GET',
        secure: true,
        ...params
      })
  };
}
