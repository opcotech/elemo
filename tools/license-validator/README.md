# License validator

This tool validates a license file that is read by Elemo. To validate a
license, first get the public key from the secret store.

## Usage

```bash
Usage of license-validator:
  -license string
        License to validate
  -public-key string
        The public key to use
```

## Example

```bash
go run tools/license-validator/main.go \
    -public-key "configs/keys/signing/public.key" \
    -license "configs/licenses/license.key"
```
