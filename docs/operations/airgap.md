# AirGap operations

This document describes vendor issuance and operator recovery for the
commercial AirGap artifact. Normal self-hosted Elemo does not read a
license and has no human-seat ceiling.

Cloud storage and retention quotas are a separate billing concern. They
are not part of this entitlement.

## Trust

AirGap licenses are vendor-signed Ed25519 JSON envelopes. The production
binary embeds public keys from `internal/entitlement/license/vendorkeys`.
Private keys are held by the vendor and are never committed or distributed.

## Issuance

1. Start an AirGap server against the customer database. The process logs
   `airgap installation identity` with the logical installation UUID. An
   administrator can also read it from `GET /v1/system/entitlements`.
2. Sign a license with `tools/license-generator`, binding that UUID and
   the purchased active-human seat count.
3. Mount the file as `airgap.license_file` and perform a rolling restart.

Self-hosted binaries ignore `airgap.license_file`. Compiling with
`-tags airgap` is the only way to produce the commercial artifact.

```bash
go build -tags airgap -o bin/elemo-airgap ./cmd/elemo
```

## Replacement

Replace the mounted file and restart AirGap nodes. The previously verified
entitlement stays in memory until restart. A malformed replacement puts
restarted nodes into restricted activation and read-only mode. Existing
users and data remain readable.

A replacement with a lower seat count never deactivates users. New
activations resume after active human usage falls below the new limit.

There is no upload API and no hidden network path.

## Seats and expiration

Only active human users consume a seat. Pending invitations, inactive
users, deleted users, and OAuth or API clients do not. Invitation
acceptance and any non-active-to-active transition consume a seat.

The license is valid through `expires_at`, then a 30-day grace period
allows seat-limited activations and ordinary writes. After grace, and
when the license is missing, invalid, or not yet valid, new activations
are denied and the instance is fail-closed read-only: user-initiated
creates, updates, and deletes are rejected. Login, password reset, user
deactivation, license recovery, export, backup, and background search or
custom-field reconciliation remain available. Restore a valid license
file and restart to leave read-only mode.

## Backup, restore, and reissue

The logical installation UUID lives on the singleton Neo4j
`Installation` node. A database backup preserves it. A restore of that
database keeps the existing license valid.

Losing the graph identity (new empty database, discarded volume) produces
a new UUID. The previous license will not bind. Ask the vendor to reissue
against the new identity.

Kubernetes rescheduling and node replacement do not change the UUID.

## Official updates

Access to newer official AirGap bundles ends at license expiration. That
gate is enforced by vendor distribution, not by disabling an installed
runtime. Installed versions remain readable. After grace they become
read-only until a valid license is restored. Grace does not extend access
to newer official bundles.

## Operator warnings

Administrators see `GET /v1/system/entitlements` and a settings warning
when the License is in grace, expired, invalid, missing, or
not-yet-valid. After grace, and when the license is missing, invalid, or
not yet valid, domain writes return HTTP 409 `entitlement_read_only`.
System health does not include license state. Normal self-hosted
deployments show no licensing UI.

Set `airgap.billing_email` to the operator mailbox that should receive
license expiration reminders. A verified license triggers one reminder per
day when it is within seven days of expiration, in the 30-day grace period,
or expired. Missing, invalid, and not-yet-valid licenses remain visible
through the administrator API, settings warning, and startup logs.
