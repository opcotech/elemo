# Key generator

Vendor-only tool that writes an Ed25519 key pair used to sign AirGap licenses.
The private key must never enter the repository, CI, or a distributed artifact.

The public key is hex-encoded (64 characters). The matching private key is
hex-encoded seed+public (128 characters). Both files are written with mode `0600`.

## Usage

```bash
Usage of key-generator:
  -private string
        Output private key file (hex-encoded Ed25519 seed+public) (default "private.key")
  -public string
        Output public key file (hex-encoded Ed25519 public key) (default "public.key")
```

## Example

```bash
go run ./tools/key-generator \
    -private "/secure/path/elemo-airgap-2026.key" \
    -public "/secure/path/elemo-airgap-2026.pub"
```

Embed the public key as `internal/entitlement/license/vendorkeys/<key-id>.pub`. The file
name without `.pub` is the key ID that `tools/license-generator` must pass
as `-key-id`.
