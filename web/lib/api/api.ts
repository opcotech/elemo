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

/**
 * User
 * A user in the system.
 */
export interface User {
  /**
   * Unique identifier of the user.
   * @example "9bsv0s46s6s002p9ltq0"
   */
  id: string;
  /**
   * The unique username of the user.
   * @minLength 3
   * @maxLength 50
   * @pattern ^[a-z0-9-_]{3,50}$
   * @example "test-user"
   */
  username: string;
  /**
   * First name of the user.
   * @minLength 1
   * @maxLength 50
   * @example "Test"
   */
  first_name: string | null;
  /**
   * Last name of the user.
   * @minLength 1
   * @maxLength 50
   * @example "User"
   */
  last_name: string | null;
  /**
   * Email address of the user.
   * @format email
   * @minLength 6
   * @maxLength 254
   * @example "user@example.com"
   */
  email: string;
  /**
   * Profile picture of the user.
   * @format uri
   * @maxLength 2000
   * @example "https://example.com/users/my-user.png"
   */
  picture: string | null;
  /**
   * Work title of the user.
   * @minLength 3
   * @maxLength 50
   * @example "Senior Software Engineer"
   */
  title: string | null;
  /**
   * Self description of the user.
   * @maxLength 500
   * @example "I'm working smart on software."
   */
  bio: string | null;
  /**
   * Working address of the user.
   * @maxLength 500
   * @example "Remote"
   */
  address: string | null;
  /**
   * Phone number of the user.
   * @minLength 7
   * @maxLength 16
   * @example "+15555551234"
   */
  phone: string | null;
  /**
   * Links to show on profile page.
   * @uniqueItems true
   */
  links: string[] | null;
  /**
   * Languages of the user.
   * @uniqueItems true
   */
  languages: Language[];
  /** Status of the user. */
  status: UserStatus;
  /**
   * Date when the user was created.
   * @format date-time
   */
  created_at: string;
  /**
   * Date when the user was updated.
   * @format date-time
   */
  updated_at: string | null;
}

/**
 * UserStatus
 * Status of the user.
 * @example "active"
 */
export enum UserStatus {
  Active = 'active',
  Pending = 'pending',
  Inactive = 'inactive',
  Deleted = 'deleted'
}

/**
 * Organization
 * An organization in the system.
 */
export interface Organization {
  /**
   * Unique identifier of the organization.
   * @example "9bsv0s46s6s002p9ltq0"
   */
  id: string;
  /**
   * Name of the organization.
   * @minLength 1
   * @maxLength 120
   * @example "ACME Inc."
   */
  name: string;
  /**
   * Email address of the organization.
   * @format email
   * @minLength 6
   * @maxLength 254
   * @example "info@example.com"
   */
  email: string;
  /**
   * Logo of the organization.
   * @format uri
   * @maxLength 2000
   * @example "https://example.com/static/logo.png"
   */
  logo: string | null;
  /**
   * Work title of the user.
   * @format uri
   * @maxLength 2000
   * @example "https://example.com"
   */
  website: string | null;
  /** Status of the organization. */
  status: OrganizationStatus;
  /**
   * IDs of the users in the organization.
   * @uniqueItems true
   */
  members: string[];
  /**
   * IDs of the teams in the organization.
   * @uniqueItems true
   */
  teams: string[];
  /**
   * IDs of the namespaces in the organization.
   * @uniqueItems true
   */
  namespaces: string[];
  /**
   * Date when the organization was created.
   * @format date-time
   */
  created_at: string;
  /**
   * Date when the organization was updated.
   * @format date-time
   */
  updated_at: string | null;
}

/**
 * OrganizationStatus
 * Status of the organization.
 * @example "active"
 */
export enum OrganizationStatus {
  Active = 'active',
  Deleted = 'deleted'
}

/**
 * Todo
 * A todo item belonging to a user.
 */
export interface Todo {
  /**
   * Unique identifier of the todo .
   * @example "9bsv0s46s6s002p9ltq0"
   */
  id: string;
  /**
   * Title of the todo item.
   * @minLength 3
   * @maxLength 250
   * @example "Do something great"
   */
  title: string;
  /**
   * Description of the todo item.
   * @minLength 10
   * @maxLength 500
   * @example "I'll make the world a better place today."
   */
  description: string;
  /** Priority of the todo item. */
  priority: TodoPriority;
  /**
   * Status of the todo item.
   * @default true
   */
  completed: boolean;
  /** ID of the user who owns the todo item. */
  owned_by: string;
  /** ID of the user who created the todo item. */
  created_by: string;
  /**
   * Completion due date of the todo item.
   * @format date-time
   */
  due_date: string | null;
  /**
   * Date when the todo item was created.
   * @format date-time
   */
  created_at: string;
  /**
   * Date when the todo item was updated.
   * @format date-time
   */
  updated_at: string | null;
}

/**
 * Priority of the todo item.
 * @minLength 6
 * @maxLength 9
 * @example "urgent"
 */
export enum TodoPriority {
  Normal = 'normal',
  Important = 'important',
  Urgent = 'urgent',
  Critical = 'critical'
}

/** SystemHealth */
export interface SystemHealth {
  /**
   * Health of the cache database.
   * @minLength 7
   * @maxLength 9
   */
  cache_database: SystemHealthCacheDatabase;
  /**
   * Health of the graph database.
   * @minLength 7
   * @maxLength 9
   */
  graph_database: SystemHealthGraphDatabase;
  /**
   * Health of the relational database.
   * @minLength 7
   * @maxLength 9
   */
  relational_database: SystemHealthRelationalDatabase;
  /**
   * Health of the license.
   * @minLength 7
   * @maxLength 9
   */
  license: SystemHealthLicense;
  /**
   * Health of the message queue.
   * @minLength 7
   * @maxLength 9
   */
  message_queue: SystemHealthMessageQueue;
}

/** SystemVersion */
export interface SystemVersion {
  /**
   * Version of the application.
   * @pattern ^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$
   */
  version: string;
  /**
   * Commit hash of the build.
   * @pattern ^[0-9a-f]{5,40}$
   */
  commit: string;
  /**
   * Build date and time of the application.
   * @format date-time
   */
  date: string;
  /**
   * Go version used to build the application.
   * @pattern ^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$
   */
  go_version: string;
}

/** SystemLicense */
export interface SystemLicense {
  /** Unique ID identifying the license. */
  id: string;
  /** Name of the organization the license belongs to. */
  organization: string;
  /**
   * Email address of the licensee.
   * @format email
   * @minLength 6
   * @maxLength 254
   * @example "info@example.com"
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
  /**
   * Features enabled by the license.
   * @uniqueItems true
   */
  features: SystemLicenseFeatures[];
  /**
   * Date and time when the license expires.
   * @format date-time
   */
  expires_at: string;
}

/**
 * HTTPError
 * HTTP error description.
 */
export interface HTTPError {
  /** Description of the error. */
  message: string;
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

/**
 * Health of the cache database.
 * @minLength 7
 * @maxLength 9
 */
export enum SystemHealthCacheDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/**
 * Health of the graph database.
 * @minLength 7
 * @maxLength 9
 */
export enum SystemHealthGraphDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/**
 * Health of the relational database.
 * @minLength 7
 * @maxLength 9
 */
export enum SystemHealthRelationalDatabase {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/**
 * Health of the license.
 * @minLength 7
 * @maxLength 9
 */
export enum SystemHealthLicense {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

/**
 * Health of the message queue.
 * @minLength 7
 * @maxLength 9
 */
export enum SystemHealthMessageQueue {
  Healthy = 'healthy',
  Unhealthy = 'unhealthy',
  Unknown = 'unknown'
}

export enum SystemLicenseFeatures {
  Components = 'components',
  CustomStatuses = 'custom_statuses',
  CustomFields = 'custom_fields',
  MultipleAssignees = 'multiple_assignees',
  Releases = 'releases'
}

export interface V1UsersGetParams {
  /**
   * Number of resources to skip.
   * @min 0
   * @default 0
   */
  offset?: number;
  /**
   * Number of resources to return.
   * @min 1
   * @max 1000
   * @default 100
   */
  limit?: number;
}

/** @uniqueItems true */
export type V1UsersGetData = User[];

export type V1UserGetData = User;

export interface V1UserDeleteParams {
  /** Irreversibly delete the user. */
  force?: boolean;
  /**
   * ID of the resource.
   * @example "9bsv0s46s6s002p9ltq0"
   */
  id: string;
}

export type V1UserDeleteData = any;

export type V1UserUpdateData = User;

export interface V1TodosGetParams {
  /**
   * Number of resources to skip.
   * @min 0
   * @default 0
   */
  offset?: number;
  /**
   * Number of resources to return.
   * @min 1
   * @max 1000
   * @default 100
   */
  limit?: number;
  /** Completion status of the items. */
  completed?: boolean;
}

/** @uniqueItems true */
export type V1TodosGetData = Todo[];

export type V1TodoGetData = Todo;

export type V1TodoDeleteData = any;

export type V1TodoUpdateData = Todo;

export interface V1OrganizationsGetParams {
  /**
   * Number of resources to skip.
   * @min 0
   * @default 0
   */
  offset?: number;
  /**
   * Number of resources to return.
   * @min 1
   * @max 1000
   * @default 100
   */
  limit?: number;
}

/** @uniqueItems true */
export type V1OrganizationsGetData = Organization[];

export type V1OrganizationGetData = Organization;

export interface V1OrganizationDeleteParams {
  /** Irreversibly delete the user. */
  force?: boolean;
  /**
   * ID of the resource.
   * @example "9bsv0s46s6s002p9ltq0"
   */
  id: string;
}

export type V1OrganizationDeleteData = any;

export type V1OrganizationUpdateData = Organization;

/** @uniqueItems true */
export type V1OrganizationMembersGetData = User[];

export interface V1OrganizationMembersAddPayload {
  /**
   * ID of the user to add.
   * @example "9bsv0s46s6s002p9ltq0"
   */
  user_id: string;
}

export type V1OrganizationMembersRemoveData = any;

export type V1SystemHealthData = SystemHealth;

export enum User4 {
  OK = 'OK'
}

export enum V1SystemHeartbeatData {
  OK = 'OK'
}

export type V1SystemLicenseData = SystemLicense;

export type V1SystemVersionData = SystemVersion;

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
     * @description Returns the paginated list of users
     *
     * @tags Users
     * @name V1UsersGet
     * @summary Get all users
     * @request GET:/v1/users
     * @secure
     * @response `200` `V1UsersGetData` OK
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1UsersGet: (query: V1UsersGetParams, params: RequestParams = {}) =>
      this.request<V1UsersGetData, HTTPError>({
        path: `/v1/users`,
        method: 'GET',
        query: query,
        secure: true,
        ...params
      }),

    /**
 * @description Create a new user.
 *
 * @name V1UsersCreate
 * @summary Create new user
 * @request POST:/v1/users
 * @secure
 * @response `201` `{
  \** ID of the newly created resource. *\
    id: string,

}`
 * @response `400` `HTTPError`
 * @response `401` `HTTPError`
 * @response `403` `HTTPError`
 * @response `500` `HTTPError`
 */
    v1UsersCreate: (
      data: {
        /**
         * The unique username of the user.
         * @minLength 3
         * @maxLength 50
         * @pattern ^[a-z0-9-_]{3,50}$
         * @example "test-user"
         */
        username: string;
        /**
         * First name of the user.
         * @minLength 1
         * @maxLength 50
         * @example "Test"
         */
        first_name?: string | null;
        /**
         * Last name of the user.
         * @minLength 1
         * @maxLength 50
         * @example "User"
         */
        last_name?: string | null;
        /**
         * Email address of the user.
         * @format email
         * @minLength 6
         * @maxLength 254
         * @example "user@example.com"
         */
        email: string;
        /**
         * Password of the user.
         * @format password
         * @minLength 8
         * @maxLength 64
         * @example "super-secret"
         */
        password: string;
        /**
         * Profile picture of the user.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com/users/my-user.png"
         */
        picture?: string | null;
        /**
         * Work title of the user.
         * @minLength 3
         * @maxLength 50
         * @example "Senior Software Engineer"
         */
        title?: string | null;
        /**
         * Self description of the user.
         * @maxLength 500
         * @example "I'm working smart on software."
         */
        bio?: string | null;
        /**
         * Working address of the user.
         * @maxLength 500
         * @example "Remote"
         */
        address?: string | null;
        /**
         * Phone number of the user.
         * @minLength 7
         * @maxLength 16
         * @example "+15555551234"
         */
        phone?: string | null;
        /**
         * Links to show on profile page.
         * @uniqueItems true
         */
        links?: string[] | null;
        /**
         * Languages of the user.
         * @uniqueItems true
         */
        languages?: Language[] | null;
      },
      params: RequestParams = {}
    ) =>
      this.request<
        {
          /** ID of the newly created resource. */
          id: string;
        },
        HTTPError
      >({
        path: `/v1/users`,
        method: 'POST',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Return the requested user by its ID.
     *
     * @tags Users
     * @name V1UserGet
     * @summary Get user
     * @request GET:/v1/users/{id}
     * @secure
     * @response `200` `V1UserGetData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1UserGet: (id: string, params: RequestParams = {}) =>
      this.request<V1UserGetData, HTTPError>({
        path: `/v1/users/${id}`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Delete a user by its ID. The user is not deleted irreversibly until the "force" parameter is set to true.
     *
     * @name V1UserDelete
     * @summary Delete the user with the given ID.
     * @request DELETE:/v1/users/{id}
     * @secure
     * @response `204` `V1UserDeleteData` No Content
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1UserDelete: ({ id, ...query }: V1UserDeleteParams, params: RequestParams = {}) =>
      this.request<V1UserDeleteData, HTTPError>({
        path: `/v1/users/${id}`,
        method: 'DELETE',
        query: query,
        secure: true,
        ...params
      }),

    /**
     * @description Update the given user.
     *
     * @name V1UserUpdate
     * @summary Update user
     * @request PATCH:/v1/users/{id}
     * @secure
     * @response `200` `V1UserUpdateData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1UserUpdate: (
      id: string,
      data: {
        /**
         * The unique username of the user.
         * @minLength 3
         * @maxLength 50
         * @pattern ^[a-z0-9-_]{3,50}$
         * @example "test-user"
         */
        username?: string;
        /**
         * First name of the user.
         * @minLength 1
         * @maxLength 50
         * @example "Test"
         */
        first_name?: string;
        /**
         * Last name of the user.
         * @minLength 1
         * @maxLength 50
         * @example "User"
         */
        last_name?: string;
        /**
         * Email address of the user.
         * @format email
         * @minLength 6
         * @maxLength 254
         * @example "user@example.com"
         */
        email?: string;
        /**
         * Password of the user.
         * @format password
         * @minLength 8
         * @maxLength 64
         * @example "super-secret"
         */
        password?: string;
        /**
         * Profile picture of the user.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com/users/my-user.png"
         */
        picture?: string;
        /**
         * Work title of the user.
         * @minLength 3
         * @maxLength 50
         * @example "Senior Software Engineer"
         */
        title?: string;
        /**
         * Self description of the user.
         * @maxLength 500
         * @example "I'm working smart on software."
         */
        bio?: string;
        /**
         * Working address of the user.
         * @maxLength 500
         * @example "Remote"
         */
        address?: string;
        /**
         * Phone number of the user.
         * @minLength 7
         * @maxLength 16
         * @example "+15555551234"
         */
        phone?: string;
        /**
         * Links to show on profile page.
         * @uniqueItems true
         */
        links?: string[];
        /**
         * Languages of the user.
         * @uniqueItems true
         */
        languages?: Language[];
        /** Status of the user. */
        status?: UserStatus;
      },
      params: RequestParams = {}
    ) =>
      this.request<V1UserUpdateData, HTTPError>({
        path: `/v1/users/${id}`,
        method: 'PATCH',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Returns all todo items belonging to the current user.
     *
     * @tags Todos
     * @name V1TodosGet
     * @summary Get todo item
     * @request GET:/v1/todos
     * @secure
     * @response `200` `V1TodosGetData` OK
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1TodosGet: (query: V1TodosGetParams, params: RequestParams = {}) =>
      this.request<V1TodosGetData, HTTPError>({
        path: `/v1/todos`,
        method: 'GET',
        query: query,
        secure: true,
        ...params
      }),

    /**
 * @description Create a new todo item.
 *
 * @tags Todos
 * @name V1TodosCreate
 * @summary Create todo item
 * @request POST:/v1/todos
 * @secure
 * @response `201` `{
  \** ID of the newly created resource. *\
    id: string,

}`
 * @response `400` `HTTPError`
 * @response `401` `HTTPError`
 * @response `403` `HTTPError`
 * @response `500` `HTTPError`
 */
    v1TodosCreate: (
      data: {
        /**
         * Title of the todo item.
         * @minLength 3
         * @maxLength 250
         * @example "Do something great"
         */
        title: string;
        /**
         * Description of the todo item.
         * @minLength 10
         * @maxLength 500
         * @example "I'll make the world a better place today."
         */
        description?: string;
        /** Priority of the todo item. */
        priority: TodoPriority;
        /** ID of the user who owns the todo item. */
        owned_by: string;
        /** Completion due date of the todo item. */
        due_date?: string;
      },
      params: RequestParams = {}
    ) =>
      this.request<
        {
          /** ID of the newly created resource. */
          id: string;
        },
        HTTPError
      >({
        path: `/v1/todos`,
        method: 'POST',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Return a todo item based on the todo id belonging to the current user.
     *
     * @tags Todos
     * @name V1TodoGet
     * @summary Get todo item
     * @request GET:/v1/todos/{id}
     * @secure
     * @response `200` `V1TodoGetData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1TodoGet: (id: string, params: RequestParams = {}) =>
      this.request<V1TodoGetData, HTTPError>({
        path: `/v1/todos/${id}`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Delete todo by its ID.
     *
     * @tags Todos
     * @name V1TodoDelete
     * @summary Delete todo item
     * @request DELETE:/v1/todos/{id}
     * @secure
     * @response `204` `V1TodoDeleteData` No Content
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1TodoDelete: (id: string, params: RequestParams = {}) =>
      this.request<V1TodoDeleteData, HTTPError>({
        path: `/v1/todos/${id}`,
        method: 'DELETE',
        secure: true,
        ...params
      }),

    /**
     * @description Update the given todo
     *
     * @tags Todos
     * @name V1TodoUpdate
     * @summary Update todo
     * @request PATCH:/v1/todos/{id}
     * @secure
     * @response `200` `V1TodoUpdateData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1TodoUpdate: (
      id: string,
      data: {
        /**
         * Title of the todo item.
         * @minLength 3
         * @maxLength 250
         * @example "Do something great"
         */
        title?: string;
        /**
         * Description of the todo item.
         * @minLength 10
         * @maxLength 500
         * @example "I'll make the world a better place today."
         */
        description?: string;
        /** Priority of the todo item. */
        priority?: TodoPriority;
        /** Completion status of the todo item. */
        completed?: boolean;
        /** ID of the user who owns the todo item. */
        owned_by?: string;
        /** Completion due date of the todo item. */
        due_date?: string;
      },
      params: RequestParams = {}
    ) =>
      this.request<V1TodoUpdateData, HTTPError>({
        path: `/v1/todos/${id}`,
        method: 'PATCH',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Returns the list of organizations in the system.
     *
     * @tags Organizations
     * @name V1OrganizationsGet
     * @summary Get organizations
     * @request GET:/v1/organizations
     * @secure
     * @response `200` `V1OrganizationsGetData` OK
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationsGet: (query: V1OrganizationsGetParams, params: RequestParams = {}) =>
      this.request<V1OrganizationsGetData, HTTPError>({
        path: `/v1/organizations`,
        method: 'GET',
        query: query,
        secure: true,
        ...params
      }),

    /**
 * @description Create a new organization.
 *
 * @tags Organizations
 * @name V1OrganizationsCreate
 * @summary Create organization
 * @request POST:/v1/organizations
 * @secure
 * @response `201` `{
  \** ID of the newly created resource. *\
    id: string,

}`
 * @response `400` `HTTPError`
 * @response `401` `HTTPError`
 * @response `403` `HTTPError`
 * @response `500` `HTTPError`
 */
    v1OrganizationsCreate: (
      data: {
        /**
         * Name of the organization.
         * @minLength 1
         * @maxLength 120
         * @example "ACME Inc."
         */
        name: string;
        /**
         * Email address of the organization.
         * @format email
         * @minLength 6
         * @maxLength 254
         * @example "info@example.com"
         */
        email: string;
        /**
         * Logo of the organization.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com/static/logo.png"
         */
        logo?: string;
        /**
         * Work title of the user.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com"
         */
        website?: string;
      },
      params: RequestParams = {}
    ) =>
      this.request<
        {
          /** ID of the newly created resource. */
          id: string;
        },
        HTTPError
      >({
        path: `/v1/organizations`,
        method: 'POST',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Returns the given organization by its ID.
     *
     * @tags Organizations
     * @name V1OrganizationGet
     * @summary Get organization
     * @request GET:/v1/organizations/{id}
     * @secure
     * @response `200` `V1OrganizationGetData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationGet: (id: string, params: RequestParams = {}) =>
      this.request<V1OrganizationGetData, HTTPError>({
        path: `/v1/organizations/${id}`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Delete the organization by its ID.
     *
     * @tags Organizations
     * @name V1OrganizationDelete
     * @summary Delete organization
     * @request DELETE:/v1/organizations/{id}
     * @secure
     * @response `204` `V1OrganizationDeleteData` No Content
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationDelete: ({ id, ...query }: V1OrganizationDeleteParams, params: RequestParams = {}) =>
      this.request<V1OrganizationDeleteData, HTTPError>({
        path: `/v1/organizations/${id}`,
        method: 'DELETE',
        query: query,
        secure: true,
        ...params
      }),

    /**
     * @description Update the organization by its ID.
     *
     * @tags Organizations
     * @name V1OrganizationUpdate
     * @summary Update organization
     * @request PATCH:/v1/organizations/{id}
     * @secure
     * @response `200` `V1OrganizationUpdateData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationUpdate: (
      id: string,
      data: {
        /**
         * Name of the organization.
         * @minLength 1
         * @maxLength 120
         * @example "ACME Inc."
         */
        name?: string;
        /**
         * Email address of the organization.
         * @format email
         * @minLength 6
         * @maxLength 254
         * @example "info@example.com"
         */
        email?: string;
        /**
         * Logo of the organization.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com/static/logo.png"
         */
        logo?: string;
        /**
         * Work title of the user.
         * @format uri
         * @maxLength 2000
         * @example "https://example.com"
         */
        website?: string;
        /** Status of the organization. */
        status?: OrganizationStatus;
      },
      params: RequestParams = {}
    ) =>
      this.request<V1OrganizationUpdateData, HTTPError>({
        path: `/v1/organizations/${id}`,
        method: 'PATCH',
        body: data,
        secure: true,
        ...params
      }),

    /**
     * @description Return the users that are members of the organization.
     *
     * @tags Organizations, Users
     * @name V1OrganizationMembersGet
     * @summary Get organization members
     * @request GET:/v1/organizations/{id}/members
     * @secure
     * @response `200` `V1OrganizationMembersGetData` OK
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationMembersGet: (id: string, params: RequestParams = {}) =>
      this.request<V1OrganizationMembersGetData, HTTPError>({
        path: `/v1/organizations/${id}/members`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
 * @description Add an existing user to an organization.
 *
 * @tags Organizations, Users
 * @name V1OrganizationMembersAdd
 * @summary Add organization member
 * @request POST:/v1/organizations/{id}/members
 * @secure
 * @response `201` `{
  \** ID of the newly created resource. *\
    id: string,

}`
 * @response `400` `HTTPError`
 * @response `401` `HTTPError`
 * @response `403` `HTTPError`
 * @response `404` `HTTPError`
 * @response `500` `HTTPError`
 */
    v1OrganizationMembersAdd: (id: string, data: V1OrganizationMembersAddPayload, params: RequestParams = {}) =>
      this.request<
        {
          /** ID of the newly created resource. */
          id: string;
        },
        HTTPError
      >({
        path: `/v1/organizations/${id}/members`,
        method: 'POST',
        body: data,
        secure: true,
        type: ContentType.Json,
        ...params
      }),

    /**
     * @description Removes a member from the organization
     *
     * @tags Organizations, Users
     * @name V1OrganizationMembersRemove
     * @summary Remove organization member
     * @request DELETE:/v1/organizations/{id}/members/{user_id}
     * @secure
     * @response `204` `V1OrganizationMembersRemoveData` No Content
     * @response `400` `HTTPError`
     * @response `401` `HTTPError`
     * @response `403` `HTTPError`
     * @response `404` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1OrganizationMembersRemove: (id: string, userId: string, params: RequestParams = {}) =>
      this.request<V1OrganizationMembersRemoveData, HTTPError>({
        path: `/v1/organizations/${id}/members/${userId}`,
        method: 'DELETE',
        secure: true,
        ...params
      }),

    /**
     * @description Returns the health of registered components.
     *
     * @tags System
     * @name V1SystemHealth
     * @summary Get system health
     * @request GET:/v1/system/health
     * @response `200` `V1SystemHealthData` OK
     * @response `500` `HTTPError`
     */
    v1SystemHealth: (params: RequestParams = {}) =>
      this.request<V1SystemHealthData, HTTPError>({
        path: `/v1/system/health`,
        method: 'GET',
        ...params
      }),

    /**
     * @description Returns 200 OK if the service is reachable.
     *
     * @tags System
     * @name V1SystemHeartbeat
     * @summary Get heartbeat
     * @request GET:/v1/system/heartbeat
     * @response `200` `V1SystemHeartbeatData` OK
     * @response `500` `HTTPError`
     */
    v1SystemHeartbeat: (params: RequestParams = {}) =>
      this.request<V1SystemHeartbeatData, HTTPError>({
        path: `/v1/system/heartbeat`,
        method: 'GET',
        ...params
      }),

    /**
     * @description Return the license information. The license information is only available to entitled users.
     *
     * @tags System
     * @name V1SystemLicense
     * @summary Get license info
     * @request GET:/v1/system/license
     * @secure
     * @response `200` `V1SystemLicenseData` OK
     * @response `403` `HTTPError`
     * @response `500` `HTTPError`
     */
    v1SystemLicense: (params: RequestParams = {}) =>
      this.request<V1SystemLicenseData, HTTPError>({
        path: `/v1/system/license`,
        method: 'GET',
        secure: true,
        ...params
      }),

    /**
     * @description Returns the version information of the system.
     *
     * @tags System
     * @name V1SystemVersion
     * @summary Get system version
     * @request GET:/v1/system/version
     * @response `200` `V1SystemVersionData` OK
     * @response `500` `HTTPError`
     */
    v1SystemVersion: (params: RequestParams = {}) =>
      this.request<V1SystemVersionData, HTTPError>({
        path: `/v1/system/version`,
        method: 'GET',
        ...params
      })
  };
}
