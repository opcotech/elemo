# License generator

Vendor-only tool that signs an AirGap human-seat license. The private key
is required at issuance time and must not be stored in this repository.

The output is a versioned JSON envelope (`elemo.airgap.license.v1`) signed
with Ed25519. The payload binds a customer, a logical installation UUID,
and an active-human seat count. Ordinary product features and non-seat
quotas are not part of the license.

Write the file to a path the AirGap binary can read, then mount it as
`airgap.license_file` and restart the AirGap server.

## Usage

```bash
Usage of license-generator:
  -customer string
        Customer name
  -installation-id string
        Logical installation UUID
  -key-id string
        Signing key identifier (default "elemo-airgap-2026")
  -license string
        Output license file (default "license.json")
  -not-before-offset duration
        Offset from now for not_before (may be negative)
  -private-key string
        Hex-encoded Ed25519 private key file
  -seats uint
        Active human seat count
  -validity-period int
        License validity period in days from not-before (default 365)
```

## Example

```bash
go run ./tools/license-generator \
    -private-key "/secure/path/elemo-airgap-2026.key" \
    -key-id "elemo-airgap-2026" \
    -customer "ACME Corp" \
    -installation-id "11111111-1111-1111-1111-111111111111" \
    -seats 25 \
    -validity-period 365 \
    -license "/secure/path/acme-airgap.json"
```

The installation UUID is logged by an AirGap server at startup as
`airgap installation identity`. It is also returned to administrators on
`GET /v1/system/entitlements`.
