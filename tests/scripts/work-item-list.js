import http from 'k6/http';
import {check, fail, group} from 'k6';
import {Counter, Trend} from 'k6/metrics';

const OAUTH_SCOPES = [
  'user',
  'organization',
  'namespace',
  'project',
  'issue',
  'document',
  'label',
  'todo',
  'notification',
].join(' ');

const pageDuration = new Trend('issue_list_page_duration', true);
const pageOneDuration = new Trend('issue_list_page1_duration', true);
const collectDuration = new Trend('issue_list_collect_duration', true);
const collectPages = new Trend('issue_list_collect_pages', true);
const payloadBytes = new Counter('issue_list_payload_bytes');
const cacheSpeedupRatio = new Trend('issue_list_cache_speedup_ratio', true);

function toArray(value) {
  if (!value) {
    return [];
  }
  return String(value)
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function encodeQuery(query) {
  const pairs = [];
  for (const [key, raw] of Object.entries(query)) {
    if (raw == null || raw === '') {
      continue;
    }
    if (Array.isArray(raw)) {
      for (const value of raw) {
        pairs.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
      }
      continue;
    }
    pairs.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(raw))}`);
  }
  return pairs.join('&');
}

function parseBody(response, label) {
  try {
    return response.json();
  } catch (error) {
    fail(`${label}: unable to parse JSON response (${String(error)})`);
    return null;
  }
}

function authHeaders(accessToken) {
  return {
    Authorization: `Bearer ${accessToken}`,
    Accept: 'application/json',
  };
}

function oauthBaseUrl(baseUrl) {
  return baseUrl.replace(/\/api\/?$/, '');
}

function requiredEnv(name, fallback) {
  const value = __ENV[name] || fallback;
  if (!value) {
    fail(`Missing required environment variable: ${name}`);
  }
  return value;
}

function maybeEnv(name) {
  const value = __ENV[name];
  return value ? String(value).trim() : '';
}

function issueQueryFromEnv() {
  return {
    page_size: String(Number(__ENV.K6_ISSUE_PAGE_SIZE || 100)),
    q: maybeEnv('K6_ISSUE_Q'),
    status: toArray(__ENV.K6_ISSUE_STATUS),
    priority: toArray(__ENV.K6_ISSUE_PRIORITY),
    order: maybeEnv('K6_ISSUE_ORDER') || 'rank:asc',
  };
}

function createIssueListPath(scope, id) {
  if (scope === 'project') {
    return `/v1/projects/${id}/issues`;
  }
  if (scope === 'namespace') {
    return `/v1/namespaces/${id}/issues`;
  }
  return `/v1/users/${id}/issues`;
}

function getJson(baseUrl, path, accessToken, tags, query) {
  const queryString = encodeQuery(query ?? {});
  const url = queryString ? `${baseUrl}${path}?${queryString}` : `${baseUrl}${path}`;
  const response = http.get(url, {
    headers: authHeaders(accessToken),
    tags,
  });
  check(response, {
    [`${tags.scope}: status 200`]: (result) => result.status === 200,
  });
  return response;
}

function discoverNamespace(baseUrl, accessToken) {
  const response = getJson(
    baseUrl,
    '/v1/namespaces',
    accessToken,
    {scope: 'bootstrap', step: 'discover_namespace'},
    {page_size: '100'}
  );
  const body = parseBody(response, 'discover namespace');
  const namespace = body?.items?.[0];
  if (!namespace?.id) {
    fail('No namespace was returned while discovering namespace scope.');
  }
  return namespace.id;
}

function discoverProject(baseUrl, accessToken, namespaceId) {
  const response = getJson(
    baseUrl,
    `/v1/namespaces/${namespaceId}/projects`,
    accessToken,
    {scope: 'bootstrap', step: 'discover_project'},
    {page_size: '100'}
  );
  const body = parseBody(response, 'discover project');
  const project = body?.items?.[0];
  if (!project?.id) {
    fail(`No project found for namespace ${namespaceId}.`);
  }
  return project.id;
}

function discoverUser(baseUrl, accessToken, email) {
  const maxPages = Number(__ENV.K6_USER_DISCOVERY_MAX_PAGES || 20);
  let pageToken;
  for (let page = 1; page <= maxPages; page += 1) {
    const response = getJson(
      baseUrl,
      '/v1/users',
      accessToken,
      {scope: 'bootstrap', step: 'discover_user', page: String(page)},
      {
        page_size: '200',
        page_token: pageToken,
      }
    );
    const body = parseBody(response, `discover user page ${page}`);
    const users = body?.items ?? [];
    const matched = users.find((user) => user.email === email);
    if (matched?.id) {
      return matched.id;
    }
    const pageInfo = body?.page_info || {};
    if (!pageInfo.has_more || !pageInfo.next_page_token) {
      break;
    }
    pageToken = pageInfo.next_page_token;
  }
  fail(
    `Unable to locate benchmark user ${email} in /v1/users; set K6_USER_ID explicitly to skip discovery.`
  );
  return '';
}

function collectIssuePages(baseUrl, accessToken, scope, id, cacheState, query) {
  const maxPages = Number(__ENV.K6_ISSUE_MAX_PAGES || 10);
  let page = 0;
  let pageToken;
  const startedAt = Date.now();

  while (page < maxPages) {
    page += 1;
    const tags = {
      scope,
      page: String(page),
      cache_state: cacheState,
      mode: page === 1 ? 'page1' : 'collect',
    };
    const response = getJson(
      baseUrl,
      createIssueListPath(scope, id),
      accessToken,
      tags,
      {
        ...query,
        page_token: pageToken,
      }
    );
    pageDuration.add(response.timings.duration, tags);
    if (page === 1) {
      pageOneDuration.add(response.timings.duration, tags);
    }
    payloadBytes.add((response.body || '').length, tags);

    const body = parseBody(response, `${scope} page ${page}`);
    const pageInfo = body?.page_info || {};
    if (!pageInfo.has_more || !pageInfo.next_page_token) {
      break;
    }
    pageToken = pageInfo.next_page_token;
  }

  const totalDuration = Date.now() - startedAt;
  collectDuration.add(totalDuration, {scope, cache_state: cacheState});
  collectPages.add(page, {scope, cache_state: cacheState});

  return {
    pages: page,
    durationMs: totalDuration,
  };
}

function resolveScopeIds(baseUrl, accessToken, userEmail) {
  const namespaceId = maybeEnv('K6_NAMESPACE_ID') || discoverNamespace(baseUrl, accessToken);
  const projectId =
    maybeEnv('K6_PROJECT_ID') || discoverProject(baseUrl, accessToken, namespaceId);
  const userId = maybeEnv('K6_USER_ID') || discoverUser(baseUrl, accessToken, userEmail);

  return {namespaceId, projectId, userId};
}

function selectedScopes(ids) {
  const available = {
    project: {name: 'project', id: ids.projectId},
    namespace: {name: 'namespace', id: ids.namespaceId},
    user: {name: 'user', id: ids.userId},
  };
  const requested = toArray(__ENV.K6_SCOPES).map((scope) => scope.toLowerCase());
  if (!requested.length) {
    return [available.project, available.namespace, available.user];
  }

  const scopes = requested
    .map((scope) => available[scope])
    .filter(Boolean);
  if (!scopes.length) {
    fail(
      `K6_SCOPES did not contain a valid scope (expected any of: project,namespace,user; got: ${requested.join(',')}).`
    );
  }
  return scopes;
}

export function setupIssueListBenchmark() {
  const baseUrl = globalThis.BASE_URL || 'http://127.0.0.1:35478';
  const userEmail = requiredEnv('K6_USER_EMAIL', 'demo@meridian.example');
  const userPassword = requiredEnv('K6_USER_PASSWORD', 'AppleTree123');
  const clientId = requiredEnv('AUTH_CLIENT_ID');
  const clientSecret = requiredEnv('AUTH_CLIENT_SECRET');

  const tokenResponse = http.post(
    `${oauthBaseUrl(baseUrl)}/oauth/token`,
    encodeQuery({
      grant_type: 'password',
      username: userEmail,
      password: userPassword,
      client_id: clientId,
      client_secret: clientSecret,
      scope: OAUTH_SCOPES,
    }),
    {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      tags: {
        scope: 'bootstrap',
        step: 'oauth_token',
      },
    }
  );

  check(tokenResponse, {'oauth token status 200': (result) => result.status === 200});
  const tokenBody = parseBody(tokenResponse, 'oauth token');
  const accessToken = tokenBody?.access_token;
  if (!accessToken) {
    fail('OAuth token response did not include access_token.');
  }

  const ids = resolveScopeIds(baseUrl, accessToken, userEmail);
  return {
    baseUrl,
    accessToken,
    ids,
    query: issueQueryFromEnv(),
  };
}

export function runIssueListBenchmark(data) {
  const scopes = selectedScopes(data.ids);

  for (const scope of scopes) {
    group(`issue-list:${scope.name}`, function () {
      const cold = collectIssuePages(
        data.baseUrl,
        data.accessToken,
        scope.name,
        scope.id,
        'cold',
        data.query
      );
      const warm = collectIssuePages(
        data.baseUrl,
        data.accessToken,
        scope.name,
        scope.id,
        'warm',
        data.query
      );
      if (cold.durationMs > 0) {
        cacheSpeedupRatio.add(warm.durationMs / cold.durationMs, {scope: scope.name});
      }
    });
  }
}
