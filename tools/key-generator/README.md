# Key generator

This tool generates a private and public key pair that is used to sign
licenses.

## Usage

```bash
Usage of generate-key:
  -private string
        Output private key file (default "private.key")
  -public string
        Output public key file (default "public.key")
```

## Example

```bash
go run tools/generate-key/main.go
```
