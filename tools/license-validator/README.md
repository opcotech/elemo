# License validator

Validates an License file against the embedded vendor trust set, or
against a supplied public key. It prints the verified payload and the
evaluated lifecycle state (`valid`, `grace`, `expired`, `not_yet_valid`,
`invalid`).

Use this tool offline. It does not contact a license server.

## Usage

```bash
Usage of license-validator:
  -installation-id string
        Optional logical installation UUID used for lifecycle evaluation
  -key-id string
        Key identifier for -public-key (default "elemo-airgap-2026")
  -license string
        License file to validate
  -public-key string
        Optional hex-encoded public key file. Defaults to the embedded vendor trust.
```

## Example

```bash
go run ./tools/license-validator \
    -license "/secure/path/acme-airgap.json" \
    -installation-id "11111111-1111-1111-1111-111111111111"
```

Omit `-installation-id` to evaluate against the installation UUID encoded
in the payload. Pass `-public-key` only when checking a key that is not
yet embedded in the production AirGap artifact.
