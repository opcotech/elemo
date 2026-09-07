# OpenAPI assemble

Merges split OpenAPI YAML fragments under `api/openapi/src/` into the committed
bundle `api/openapi/openapi.yaml`.

## Usage

```bash
mise run generate-openapi
```

Or from the repository root (`go -C` changes cwd into the nested module;
relative `-src`/`-out` paths are still resolved from the repository root):

```bash
go -C tools/openapi-assemble run .
go -C tools/openapi-assemble run . \
  -src api/openapi/src \
  -out api/openapi/openapi.yaml
```

Edit fragments in `api/openapi/src/`. Do not edit the assembled bundle.

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `-src` | `api/openapi/src` | Fragment directory |
| `-out` | `api/openapi/openapi.yaml` | Assembled spec |
| `-split-from` | empty | Extract fragments from a bundled spec into `-src` |
