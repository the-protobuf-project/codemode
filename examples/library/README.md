# Library example

`LibraryService` manages shelves and books and lends books out. It exercises
every Code Mode feature: typed resources, enums, oneofs, maps, well-known
types, approval (`ArchiveShelf`, `MoveBook`), rollback (`CreateShelf`,
`CreateBook`, `LendBook`), a proto snippet (`move_shelf`) and examples.

| Path | What it is |
| --- | --- |
| `protobuf/library/v1/` | The annotated protos |
| `codemode.yaml` | How the example exposes the service |
| `gen/` | Generated code, committed so CI can check it stays in sync |
| `src/library.ts` | An in-memory `LibraryService` |
| `src/demo.ts` | Reads, an approval pause, resume and rollback |
| `src/mcp.ts` | The library as an MCP server on stdio |
| `test/` | Codec, runtime and MCP tests |

## Connect it to Claude

```sh
claude mcp add library -- node --conditions=source "$PWD/src/mcp.ts"   # from this directory
```

See the [main README](../../README.md#try-it-with-claude) for Claude Desktop and
other MCP clients.

## Run it

```sh
npm run demo            # from this directory
npm test                # Node; test:bun and test:deno for the other runtimes
just generate           # from the repository root, after changing a proto
```
