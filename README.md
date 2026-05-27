# protoc-gen-errors-md

A protoc plugin that renders Kratos `ErrorReason` enums into a Markdown reference table, suitable for publishing as a stable contract for frontend and client teams.

The plugin scans every enum that carries the `(errors.default_code)` option (defined by [`github.com/go-kratos/kratos/v2`](https://github.com/go-kratos/kratos)) and emits one Markdown file per proto that declares such an enum, written next to the source proto. Files without a qualifying enum produce no output.

## Install

```bash
go install github.com/vektrel/protoc-gen-errors-md/cmd/protoc-gen-errors-md@latest
```

## Usage with buf

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-errors-md
    out: .
    opt:
      - paths=source_relative
```

Then run:

```bash
buf generate
```

For each input proto that declares a qualifying enum, the plugin writes `<proto>.md` next to it (for example `user_error.proto` produces `user_error.md`). `paths=source_relative` keeps the Markdown beside its source proto; omit it to use the import-path layout.

## Input format

The plugin only processes enums annotated with Kratos error options. Example input:

```proto
syntax = "proto3";

package api.helloworld.v1;

import "errors/errors.proto";

option go_package = "example.com/api/helloworld/v1;v1";

enum ErrorReason {
  option (errors.default_code) = 500;

  // User does not exist.
  USER_NOT_FOUND = 0 [(errors.code) = 404];
  // Request payload is missing required fields.
  CONTENT_MISSING = 1 [(errors.code) = 400];
}
```

Produces:

```markdown
# api.helloworld.v1 error codes

> Default HTTP code: `500` (any reason not listed below uses this code).

| Reason | HTTP | Description |
|---|---|---|
| `USER_NOT_FOUND` | 404 | User does not exist. |
| `CONTENT_MISSING` | 400 | Request payload is missing required fields. |
```

## Formatting rules

- The H1 heading uses the full proto `package` declaration.
- The default HTTP code is taken from `(errors.default_code)`. Values without an explicit `(errors.code)` fall back to it.
- Leading comments become the description column. Newlines and runs of whitespace are collapsed to a single space.
- `|` and `\` in descriptions are escaped (`\|`, `\\`) so they render correctly inside the Markdown table.
- Enum values whose number is `0` **and** whose name ends with `_UNSPECIFIED` are skipped, matching the proto3 zero-value convention.
- Rows preserve the source order from the `.proto` file.
- If a file contains multiple qualifying enums, each gets an `## <EnumName>` section under a single H1.

## Development

```bash
go test ./...      # unit and golden tests
go vet ./...
go build ./...
```

Golden files live under `internal/render/testdata/`. To regenerate them after an intentional format change:

```bash
UPDATE_GOLDEN=1 go test ./internal/render/...
```

`protoc` and the `errors/errors.proto` from the Kratos module cache are both required for the golden tests; they are skipped automatically if either is missing.

## License

MIT
