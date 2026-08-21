# Workload Prefill

Destructive demo seeder that wipes Neo4j, Meilisearch, Redis, and Postgres
tokens, then fills a mature-company world through Elemo's internal services.

This does **not** replace `assets/queries/demo.cypher`. Use `scripts/setup.sh`
for the small ACME workspace, and this tool when you need a large tenant for
product demos.

## Prerequisites

- Backend stack running (`make start.backend`) with LocalStack, Neo4j,
  Postgres, Redis, and Meilisearch
- Config file, usually `configs/development/config.local.gen.yml`
- Run from the repository root so relative license, query, and S3 paths resolve

The generated development license already has quotas of 99999, which is enough
for the full profile.

## Usage

```bash
go run ./tools/workload-prefill \
  -config configs/development/config.local.gen.yml \
  -yes
```

Or:

```bash
make demo.prefill
```

The `-yes` flag is required. The run **deletes all graph data**, rebuilds
bootstrap constraints, truncates `user_tokens` and `notifications`, flushes
Redis, and clears the search index.

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `-config` | `$ELEMO_CONFIG` | Elemo YAML config |
| `-yes` | false | Confirm wipe |
| `-profile` | `full` | `full` or `smoke` |
| `-concurrency` | 8 | Parallel project issue seeding |
| `-seed` | 42 | RNG seed |
| `-password` | `AppleTree123` | Password for every user |
| `-queries-dir` | `assets/queries` | `bootstrap.cypher` / `bootstrap.sql` |
| `-skip-reindex` | false | Skip Meilisearch rebuild |

`smoke` is a tiny subset (about 10 users, a few projects, tens of issues) for
iterating on the tool itself.

## Full profile

Main organization **Meridian Systems** (`meridian.example`):

- 300 users and 12 teams
- Namespaces: Product, Platform, Operations, Customer (19 live projects total,
  2–5 per namespace, at least 100 issues each)
- **Migrated** namespace: 72 archived projects, 100–300 issues each
- 280 documents split across org, namespace, and project libraries

Five partner organizations (Kite Analytics, Harbor Logistics, Nimbus Cloud,
Brightline Design, Fieldstone Consulting) collaborate through ReBAC grants
rather than dual org membership, plus a few dual-org members.

Expected volume is about 100 projects and 15k–20k issues. A full run takes
several minutes on a local Docker stack.

## Logins

Every account uses the same password (`AppleTree123` unless `-password` is set).

| Email | Role |
| --- | --- |
| `demo@meridian.example` | Installation `organization.create` + Meridian org-admin |
| `maya.chen@kite.example` | Kite Analytics org-admin |
| `luis.navarro@harbor.example` | Harbor Logistics org-admin |
| `priya.shah@nimbus.example` | Nimbus Cloud org-admin |
| `aisha.okoro@brightline.example` | Brightline Design org-admin |
| `jordan.lee@fieldstone.example` | Fieldstone Consulting org-admin |
