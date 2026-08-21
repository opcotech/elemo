# k6 tests

## Work-item latency benchmark

`test.k6.work-item-latency` runs an authenticated benchmark over project,
namespace, and user issue-list endpoints. It records cold and warm page-1 and
collect-all timings, payload bytes, and cache speedup tags per scope/page.

Required environment variables:

- `AUTH_CLIENT_ID`
- `AUTH_CLIENT_SECRET`

Optional overrides:

- `K6_USER_EMAIL` (default `demo@meridian.example`)
- `K6_USER_PASSWORD` (default `AppleTree123`)
- `K6_PROJECT_ID`, `K6_NAMESPACE_ID`, `K6_USER_ID` (auto-discovered when empty)
- `K6_ISSUE_Q`, `K6_ISSUE_STATUS`, `K6_ISSUE_PRIORITY`, `K6_ISSUE_ORDER`
- `K6_ISSUE_PAGE_SIZE` (default `100`)
- `K6_ISSUE_MAX_PAGES` (default `10`)
- `K6_SCOPES` (comma-separated subset of `project,namespace,user`)
- `BASE_URL` (default `http://127.0.0.1:35478`)

Run:

```bash
make start.backend start.monitoring
make demo.prefill
make test.k6.work-item-latency
```

Use `tests/config/work-item-latency.json` as the default benchmark scenario
configuration. `tests/main.js` can load another config by setting `K6_CONFIG`.
