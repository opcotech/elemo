# License generator

This tool generates a license file that is read by Elemo. To generate a
license, first get the private key from the secret store.

## Usage

```bash
Usage of license-generator:
  -email string
        License email
  -features string
        Comma-separated list of features
  -license string
        Output license file (default "license.key")
  -organization string
        License organization
  -private-key string
        The private key to use
  -quota string
        Comma-separated key-value pairs of quotas
  -validity-period int
        License validity period in days (default 365)
```

## Example

```bash
go run tools/license-generator/main.go \
    -email" services@opcotech.com" \
    -organization "Open Code Technologies FZC" \
    -private-key "configs/keys/license-signer.key" \
    -quota-seats 100
```
