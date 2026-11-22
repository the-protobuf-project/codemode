// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Serves the library over MCP on stdio, as one `code` tool. Point an MCP
// client at `node examples/library/src/mcp.ts`.
// Executions are kept in SQLite, so paused ones survive a restart.

import { DatabaseSync } from "node:sqlite";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { createCodemode, SqliteStore, type SqliteDatabase } from "@the-protobuf-project/runtime/codemode";
import { createMcpServer } from "@the-protobuf-project/runtime/codemode/mcp";
import { QuickJSExecutor } from "@the-protobuf-project/runtime/codemode/quickjs";
import { catalog, config, connectors } from "../gen/index.ts";
import { InMemoryLibrary } from "./library.ts";

const db = new DatabaseSync(process.env["CODEMODE_DB"] ?? ":memory:") as unknown as SqliteDatabase;
const runtime = createCodemode({
  catalog,
  config,
  connectors: connectors({ library: InMemoryLibrary.seeded() }),
  executor: new QuickJSExecutor(),
  store: new SqliteStore(db),
});

const server = createMcpServer(runtime, { name: "library", expose: config.mcp?.expose ?? "code" });
await server.connect(new StdioServerTransport());
