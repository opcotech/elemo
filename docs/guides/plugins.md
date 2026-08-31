# Plugin developer guide

Elemo plugins install, enable, and disable at runtime. The host process does
not restart. Packages are zip archives with a versioned `plugin.yaml`, an
optional WASI WASM backend, and an optional frontend ESM module.

## Package layout

```text
plugin.yaml
backend/plugin.wasm      # optional; WASI reactor (see Backend)
frontend/index.js        # optional; ESM with React/SDK as externals
```

Maximum zip size is 32MB. Paths cannot contain `..` or absolute segments.

### Manifest

```yaml
schemaVersion: 1
id: com.example.plugin
name: Example
version: 1.0.0
requires:
  pluginApi: "^1"
backend:
  entry: backend/plugin.wasm
frontend:
  entry: frontend/index.js
capabilities:
  - issues.read
  - graph.read
  - graph.write
slots:
  - issue.sidebar
graph:
  foreign:
    - name: LoggedTime
      parent: Issue
      properties:
        - name: seconds
          type: integer
          required: true
  nodes:
    - kind: TimeEntry
      scope:
        parent: Issue
      properties:
        - name: seconds
          type: integer
          required: true
  relations:
    - kind: LOGGED_ON
      from: TimeEntry
      to: Issue
config:
  - name: time_source
    type: graph_binding
    foreign: LoggedTime
    required: false
```

Unknown `schemaVersion` values and Plugin API versions other than `^1` / `1`
are rejected.

## Backend (WASM)

Compile a **WASI reactor** with the public SDK [`sdk/plugin`](../../sdk/plugin).
Do **not** import `github.com/opcotech/elemo/internal`.

```bash
CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o backend/plugin.wasm .
```

`-buildmode=c-shared` exports `_initialize`. A default WASI command (`_start`
only) cannot run `//go:wasmexport` functions; the host rejects it.

The guest ABI exports `elemo_call`, `elemo_start`, `elemo_stop`, and imports
`elemo.host_call`. Host methods are versioned JSON (`issues.get`,
`graph.nodes.create`, `plugin.storage.set`, `plugin.config.get`, …). Every host
call except `plugin.config.get` requires the capability declared at install
**and** the calling user’s ReBAC grant. `plugin.config.get` is available to
any active plugin so it can read its own nearest-ancestor activation config.

The production runtime is **wazero** (WASI). Timeouts come from
`plugin.execution_timeout`. WASM traps are isolated; they never panic the
host.

## Plugin graph

Plugin domain entities are Neo4j nodes with label `Extension`, discriminated
by `plugin_id` and `kind`. The host writes `IN_SCOPE_OF` from
`graph.nodes[].scope.parent`. Domain relations are namespaced
`EXT__{plugin_id}__{KIND}` and never collide with core `EdgeKind` values.

Plugins cannot submit Cypher or create `IN_SCOPE_OF` / `GRANTED`. Relating a
core resource additionally requires that resource’s update action (for
example `issue.update`). Relating to a `User` is allowed only when the
target is the calling user (so plugins can attach `LOGGED_BY`-style edges
without a `user.update` grant, and cannot spoof another person). Relating
*to* a kind owned by the calling plugin requires `extension.read` on the
target, not `extension.update`.

`graph.foreign` declares a kind alias (name, parent, required properties).
At activation, a `graph_binding` config field stores `{ plugin_id, kind }`
for that alias. The host checks the target plugin is installed and active,
the kind exists, `scope.parent` matches, and property names/types are
compatible. Domain edges stay typed as the **declaring** plugin’s
`EXT__{plugin}__{KIND}`. Uninstall of the declaring plugin deletes those
edges; uninstall of the bound plugin `DETACH DELETE`s its nodes and the
edges go with them.

Foreign reads use `graph.nodes.get` / `list` with `ownerPluginId`. They
are allowed only when the caller has a resolved binding for that
owner+kind, `graph.read`, both plugins active, and `extension.read`.

`CreateRelation` enforces cardinality (`many-to-one` / `one-to-one` reject
an existing outgoing edge; `one-to-many` / `one-to-one` reject an existing
incoming). `graph.relations.list` accepts `direction` (`outgoing`,
`incoming`, `both`; default `outgoing`). `graph.nodes.list` accepts
`equals` for property equality filters.

Activation config lives on `plugin_activations.config`, not
`plugin_storage`. Declared field types are `string`, `integer`, `boolean`,
and `graph_binding`. HTTP `GET`/`PATCH /v1/plugins/{pluginId}/config` with
`scope_id` / `scope_type` requires `plugin.manage`. WASM
`plugin.config.get` walks the nearest enabled ancestor, same as
`requireActive`.

JSONB `plugin_storage` remains for private non-graph data (drafts, timers).

Uninstall deletes all `Extension` nodes for that `plugin_id` and all
`EXT__{plugin}__*` relationships. Disable leaves graph rows in place but
returns 404 until the plugin is enabled again.

Plugins may subscribe to `issue.*`, `project.*`, and `extension.created` /
`updated` / `deleted`. Extension events for another plugin are delivered
only when the subscriber has a resolved `graph.foreign` binding for that
`plugin_id`+`kind` (or the event is their own).

Plugins may subscribe to `issue.*`, `project.*`, and `extension.created` /
`updated` / `deleted`. Extension events for another plugin are delivered
only when the subscriber has a resolved `graph.foreign` binding for that
`plugin_id`+`kind` (or the event is their own).

## Frontend

Frontend modules load as client-only ESM via `import()`. React,
`react/jsx-runtime`, `lucide-react`, `@elemo/plugin-sdk`, and
`@elemo/plugin-ui` are provided by an import map pointing at host shims.
Mark those packages as externals when bundling. `@elemo/plugin-ui` re-exports
the same shadcn primitives the app uses (page headers, progress, tabs,
accordion, dialogs, date picker, person avatars, searchable selects, and
more) plus `showSuccessToast` / `showErrorToast` so plugin mutations use
the host toaster. `lucide-react` is a curated host icon set, not the full
package.

```ts
import { defineElemoPlugin } from "@elemo/plugin-sdk";

export default defineElemoPlugin({
  id: "com.example.plugin",
  activate(elemo) {
    elemo.slots.register("issue.sidebar", Sidebar);
    elemo.slots.register("issue.activity", LoggedTime, {
      title: "Logged time",
    });
    elemo.routes.register("report", ReportPage);
    return () => {};
  },
});
```

Plugin pages live under
`/organizations/{org}/plugins/{pluginId}/{splat}`. JSON APIs go through the
BFF. Asset URLs are same-origin (`/plugin-assets/...`) so `import()` does
not send a bearer token to the Go API.

Slots in v1: `issue.sidebar`, `issue.actions`, `issue.activity`,
`organization.settings`, `project.settings`, `project.sidebar`.
`issue.activity` contributions render as extra tabs inside the work item
Activity accordion. Pass `{ title }` when registering so the tab label is
not the plugin id. The host binding picker on the activation page is used
for `graph_binding` fields; plugins do not render that form themselves.

`graph.nodes.move` reparents an extension (`IN_SCOPE_OF`) and retargets
declared many-to-one / one-to-one domain edges whose `to` kind matches the
new parent. List and get responses include `parent_id` / `parent_type`.
When a node kind declares `user_id`, create stamps it from the caller and
ignores a guest-supplied value.

## Authorization

- `plugin.install` — Installation only (upload, upgrade, uninstall)
- `plugin.manage` — enable/disable on org / namespace / project
- `extension.create|read|update|delete` — ReBAC on plugin graph nodes.
  Organization members receive `extension.read` so they can see org-scoped
  plugin nodes (for example Account/Budget).
- Frontend discovery of *active* plugins requires read on the scope, not
  `plugin.manage`

Installing a plugin grants no implicit power. Capabilities are a closed set.

## Trust model

Frontend plugins run in the Elemo origin with full page privileges. v1
allows only administrator-installed packages. Iframe sandbox is not built.

## Reference plugins

Build the Time Tracking zip:

```bash
make plugins.timetracking
```

The archive is written to `build/plugins/com.elemo.timetracking.zip`.
Install it from **Settings → Plugins**, then enable it on an organization.
Open a work item: the sidebar timer (per user) writes a `TimeEntry` with
`seconds`, optional `note` (description), and host-stamped `user_id`. Stop
or **Log time** creates the node `IN_SCOPE_OF` the issue with `LOGGED_ON`
to the work item and `LOGGED_BY` to the calling user.
The Activity accordion adds a **Logged time** tab with a compact, headless
table of entries (edit / move / delete) and a right-aligned “View the full
time report” link that opens
`/organizations/{slug}/plugins/com.elemo.timetracking/report?workItem={issueId}`
pre-filtered to that work item. The standalone report uses the same page
header pattern as the rest of Elemo: an eyebrow, title, summary metrics,
and a wrapping filter card (search, work item, dates, person). Groups of
entries sit in bordered sections under that. The sidebar timer card does
not link to the report; per-user totals sit in a headerless table under
the card.

Build the Accounting zip:

```bash
make plugins.accounting
```

The archive is written to `build/plugins/com.elemo.accounting.zip`. Enable
it on an organization, then bind **Time source** on the activation page to
`com.elemo.timetracking` / `TimeEntry` (or another plugin kind that
matches the `LoggedTime` foreign shape). Chart of accounts and hour
budgets live under
`/organizations/{slug}/plugins/com.elemo.accounting/...`. Accounts and
hour budgets are created, edited, and deleted from those pages. Deleting
an account is blocked while it still has budgets. Project and work item
sidebars assign a budget with a searchable selector; clearing uses the
selector’s empty option (no separate Clear button). Logged time with a
bound source is counted against the assigned envelope via
`COUNTED_AGAINST`. Envelopes and actuals are integer seconds, shown as
`Xh Ym`.

Elemo does not import `plugins/timetracking` or `plugins/accounting`.

## Known limitations (v1)

- No marketplace, OCI, or signatures
- No iframe sandbox or SSR of plugin UI
- No plugin SQL/Cypher migrations; graph upgrades are additive-only
- No plugin-to-plugin install dependencies
- No per-kind ReBAC actions (`account.approve`)
- No Meilisearch indexing of `Extension` nodes
- No WebSocket push; UI refreshes via React Query after admin mutations
  and `refetchOnWindowFocus`
- Frontend plugins share the Elemo origin
