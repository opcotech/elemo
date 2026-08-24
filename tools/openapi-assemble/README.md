# OpenAPI assemble

Merges split OpenAPI YAML fragments under `api/openapi/src/` into the committed
bundle `api/openapi/openapi.yaml`.

## Usage

```bash
make generate.openapi
```

Or from the repository root, with absolute paths (required because
`go -C` runs the tool in `tools/openapi-assemble`):

```bash
go -C tools/openapi-assemble run . \
  -src "$(pwd)/api/openapi/src" \
  -out "$(pwd)/api/openapi/openapi.yaml"
```

Edit fragments in `api/openapi/src/`. Do not edit the assembled bundle.

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `-src` | `api/openapi/src` | Fragment directory |
| `-out` | `api/openapi/openapi.yaml` | Assembled spec |
| `-split-from` | empty | Extract fragments from a bundled spec into `-src` |
