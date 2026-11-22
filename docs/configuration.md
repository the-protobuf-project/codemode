# Configuration

## Plugin parameters

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-codemode
    out: gen
    strategy: all                       # one catalog across every service
    opt:
      - config=codemode.yaml            # relative to where buf runs
      - import_extension=.ts            # .ts (Node, Bun, Deno), .js (compiled), none (bundlers)
      - runtime_prefix=@the-protobuf-project/   # code imports <prefix>protobuf, network, codemode
```

## codemode.yaml

Every key is optional. Generation settings shape the output; runtime settings
are written to `codemode/config.ts`, which you pass to `createCodemode`.
[`schema/codemode.schema.json`](../schema/codemode.schema.json) validates it in
editors: add `# yaml-language-server: $schema=<path>` at the top.

```yaml
version: v1
include: ["library.v1.*"]   # service full-name globs; default: all
exclude: ["*.Internal*"]
naming: lowerCamel          # lowerCamel: library.createBook | snake: library.create_book
approval:
  destructive: require      # require | allow
  elicitation: require      # require | allow
  infer_from_aip: true      # unannotated Delete*, Purge*, Remove* count as destructive
discovery: progressive      # progressive: namespaces in the description | inline: every type
executor:
  timeout: 30s              # per sandbox pass
  memory_mb: 64
  outbound: none            # none | allowlist
  allow: []                 # with allowlist: "api.example.com", "*.example.com" (HTTPS only)
runtime:
  max_executions: 50        # finished executions kept
  paused_ttl: 24h           # expirePaused() default
  max_value_bytes: 1000000  # largest argument, result or step value recorded
  result_budget_tokens: 6000
mcp:
  expose: code              # code | search_execute
```

## Choosing discovery

`progressive` keeps the tool description short: the namespaces, their
instructions, and the `codemode` API. The model calls
`codemode.describe("library.createBook")` for a method's types. Use it for large
APIs.

`inline` puts every type and method in the description, so the model needs no
lookups. It suits APIs of up to about 30 methods.
