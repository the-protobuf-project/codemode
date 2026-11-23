// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { test } from "node:test";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";
import { ElicitRequestSchema } from "@modelcontextprotocol/sdk/types.js";
import { createMcpServer } from "@the-protobuf-project/codemode/mcp";
import { setup } from "./helpers.ts";

async function connect(options: { expose?: "code" | "search_execute"; approve?: boolean }) {
  const { runtime, library } = setup();
  const server = createMcpServer(runtime, { name: "library", expose: options.expose ?? "code" });
  const elicit = options.approve !== undefined;
  const client = new Client(
    { name: "test", version: "1.0.0" },
    { capabilities: elicit ? { elicitation: { form: {} } } : {} },
  );
  const asked: string[] = [];
  if (elicit) {
    client.setRequestHandler(ElicitRequestSchema, async (req) => {
      asked.push(req.params.message);
      return { action: "accept", content: { approve: options.approve === true } };
    });
  }
  const [a, b] = InMemoryTransport.createLinkedPair();
  await Promise.all([server.connect(a), client.connect(b)]);
  return { client, library, asked };
}

function structured(result: Awaited<ReturnType<Client["callTool"]>>): Record<string, unknown> {
  return result.structuredContent as Record<string, unknown>;
}

test("the code tool runs code and returns structured output", async () => {
  const { client } = await connect({});
  const { tools } = await client.listTools();
  assert.deepEqual(
    tools.map((t) => t.name),
    ["code", "codemode_status", "codemode_approve", "codemode_reject"],
  );
  const out = await client.callTool({
    name: "code",
    arguments: { code: "const s = await library.getShelf({ name: 'shelves/poetry' }); return s.bookCount;" },
  });
  assert.deepEqual(structured(out).result, 3);
});

test("approval is asked through elicitation when the client supports it", async () => {
  const { client, library, asked } = await connect({ approve: true });
  const out = await client.callTool({
    name: "code",
    arguments: { code: "return (await library.archiveShelf({ name: 'shelves/poetry' })).archivedCount;" },
  });
  assert.equal(structured(out).status, "completed");
  assert.equal(structured(out).result, 3);
  assert.match(asked[0] ?? "", /library\.archiveShelf/);
  assert.equal(
    [...library.books.values()].every((b) => b.state === "BOOK_STATE_ARCHIVED"),
    true,
  );
});

test("a declined elicitation rejects the call", async () => {
  const { client, library } = await connect({ approve: false });
  const out = await client.callTool({
    name: "code",
    arguments: { code: "await library.archiveShelf({ name: 'shelves/poetry' });" },
  });
  assert.equal(out.isError, true);
  assert.equal(
    [...library.books.values()].some((b) => b.state === "BOOK_STATE_ARCHIVED"),
    false,
  );
});

test("without elicitation the client resumes with codemode_approve", async () => {
  const { client } = await connect({});
  const paused = structured(
    await client.callTool({
      name: "code",
      arguments: { code: "await library.archiveShelf({ name: 'shelves/fiction' }); return 'ok';" },
    }),
  );
  assert.equal(paused.status, "paused");
  const done = await client.callTool({ name: "codemode_approve", arguments: { executionId: paused.executionId } });
  assert.equal(structured(done).result, "ok");
});

test("search_execute exposes search and execute", async () => {
  const { client } = await connect({ expose: "search_execute" });
  const { tools } = await client.listTools();
  assert.deepEqual(tools.map((t) => t.name).slice(0, 2), ["search", "execute"]);
  const found = await client.callTool({ name: "search", arguments: { query: "return lent book" } });
  assert.match(JSON.stringify(found.content), /library\.returnBook/);
});
