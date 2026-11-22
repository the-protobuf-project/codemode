# Runtime

The Code Mode runtime lives in [runtime-ts](https://github.com/the-protobuf-project/runtime-ts), the TypeScript runtime
shared by every plugin, as the `@the-protobuf-project/codemode` package. It runs model code against generated connectors.
It keeps no state in memory between calls: executions, their call logs and
snippets live in a `Store`, so a runtime can be recreated, or moved, at any time.

```ts
const runtime = createCodemode({
  catalog,                                   // gen/codemode/catalog.ts
  config,                                    // gen/codemode/config.ts
  connectors: connectors({ library: client }),
  executor: new QuickJSExecutor(),
  store: new SqliteStore(new DatabaseSync("codemode.db")),
  hooks: { onPassEnd, disposeExecution },
});
```

## Executions

`execute({ code })` returns one of:

```ts
{ status: "completed", executionId, result, logs? }
{ status: "paused", executionId, pending: [{ seq, path, args, message? }] }
{ status: "error", executionId, error, logs? }
```

| Method | What it does |
| --- | --- |
| `approve({ executionId })` | Approves the pending call and runs the code again from the top |
| `reject({ executionId, seq })` | Ends the execution; earlier calls stay applied |
| `rollback({ executionId })` | Runs reverts for applied calls, newest first; reports skips and failures |
| `pending(executionId?)` | Calls waiting for approval |
| `executions(limit)`, `getExecution(id)`, `deleteExecution(id)` | Audit history |
| `pruneExecutions(keep)` | Keeps the newest finished executions (runs after every execute) |
| `expirePaused({ maxAgeMs })` | Rejects stale paused runs and fails stale running ones |
| `search(query)`, `describe(target)` | The same discovery the sandbox has |
| `saveSnippet(name, { executionId })`, `snippets()`, `deleteSnippet(name)` | Reusable code |

## Replay

Each pass runs the code from the top. Every connector call and
`codemode.step()` gets a sequence number. A call already in the log returns its
recorded result (or error); read-only methods run again. A call that differs
from the log at its sequence number fails the execution, so code must call
methods one at a time, in a fixed order, and wrap time, randomness and other
changing values in `codemode.step(name, fn)`.

## Executors

| Executor | Import | Runs on | Isolation |
| --- | --- | --- | --- |
| `QuickJSExecutor` | `/codemode/quickjs` | Node, Bun, Deno, browsers | QuickJS in WebAssembly; memory cap; interrupt at the deadline |
| `DenoExecutor` | `/codemode/deno` | Hosts with `deno` installed | A process with every permission denied; killed at the deadline |
| `WorkerLoaderExecutor` | `/codemode/cloudflare` | Cloudflare Workers | A fresh isolate per pass; outbound blocked; calls over RPC |
| `IframeExecutor` | `/codemode/browser` | Browsers | `sandbox="allow-scripts"` iframe with a no-network CSP |

All of them run the same script and speak the same bridge, so behaviour does
not change with the executor.

## Stores

`MemoryStore` for tests; `SqliteStore` over `bun:sqlite`, `node:sqlite` (Node
22.5+, Deno) or `durableObjectSql(ctx.storage.sql)` on Cloudflare.

## Surfaces

- `runtime.tool()` — a framework-neutral `{ name, description, inputSchema, execute }`
- `createMcpServer(runtime, { expose })` from `/codemode/mcp` — `code`, or `search` and
  `execute`, plus `codemode_approve` and `codemode_reject`. When the client
  supports elicitation, approvals are asked there and the call resumes in place.
- `codemodeTool(runtime)` from `/codemode/ai` — a Vercel AI SDK tool
- `cloudflareTools(connector, catalog)` from `/codemode/cloudflare` — tools for
  `@cloudflare/codemode`'s own durable runtime
