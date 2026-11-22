# Annotations

You need no new annotations to start: a proto annotated for
[`protoc-gen-mcp`](https://github.com/the-protobuf-project/mcp), or not
annotated at all, generates. `codemode.v1` adds the four things MCP has no word
for: rollback, replay, snippets and the sandbox namespace.

Import it from `protobuf/codemode/v1/annotations.proto`
(`buf.build/the-protobuf-project/codemode` once published).

## Precedence

For every method the generator decides approval and replay in this order:

1. `codemode.v1.method` — `approval`, `replay`
2. `mcp.v1.tool` hints — `read_only`, `destructive`, `idempotent` — and `mcp.v1.elicitation`
3. The AIP method kind — `Get`, `List`, `Search`, `BatchGet` are read-only; with
   `approval.infer_from_aip`, `Delete`, `Purge` and `Remove` are destructive
4. `codemode.yaml` — `approval.destructive` and `approval.elicitation` say
   whether destructive and elicitation methods pause

Read-only methods replay by running again; everything else replays from the log.
A method that needs approval always replays from the log.

## Reused from mcp.v1 and google.api

| Annotation | Code Mode meaning |
| --- | --- |
| `mcp.v1.tool.description`, RPC comment | Method JSDoc and search text |
| `mcp.v1.tool.destructive` | Pauses for approval |
| `mcp.v1.tool.read_only`, `idempotent` | Replay policy |
| `mcp.v1.elicitation.message` | Pauses for approval; the message is shown to the approver |
| `mcp.v1.service.app.description` | Namespace instructions |
| `mcp.v1.field`, `enum`, `enum_value` | JSDoc on generated types |
| `google.api.resource`, `resource_reference` | Typed resource names: `` `shelves/${string}` `` |
| `google.api.field_behavior` `REQUIRED` | Non-optional field; required messages are checked |
| `google.api.field_behavior` `OUTPUT_ONLY` | Documented as ignored in requests |
| `buf.validate.field.required` | Same as `REQUIRED` |
| `google.api.http` | The generated client's HTTP transcoding |

## codemode.v1.service

```proto
option (codemode.v1.service) = {
  namespace: "library"                 // sandbox global; default: snake_case name without _service
  instructions: "Shelves hold books."  // shown next to the namespace
  exclude: false                       // hide the whole service
  snippets: {
    name: "move_shelf"
    description: "Moves every book from one shelf to another."
    input_type: "library.v1.MoveShelfInput"
    code: "async (input) => { ... }"
  }
};
```

Snippets are seeded into the runtime's store on first use and run with
`await codemode.run("move_shelf", { from, to })`. A snippet saved at run time
with the same name wins over the proto's.

## codemode.v1.method

```proto
option (codemode.v1.method) = {
  approval: APPROVAL_REQUIRED          // or APPROVAL_NONE; default: inferred
  replay: REPLAY_LOG                   // or REPLAY_REEXECUTE; default: inferred
  examples: "const b = await library.getBook({ name });"
  revert: {
    method: "DeleteBook"               // a unary RPC in the same service
    fields: { key: "name" value: "result.name" }   // request field <- args.* or result.*
  }
};
```

The generator checks reverts at build time: the method must exist, must not
need approval, and every field path must resolve. `REPLAY_REEXECUTE` together
with approval is rejected, as in Cloudflare's runtime.
