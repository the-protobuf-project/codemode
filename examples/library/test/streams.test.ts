// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { test } from "node:test";
import type { CodeOutput } from "@the-protobuf-project/codemode";
import { setup } from "./helpers.ts";

function completed(out: CodeOutput): unknown {
  assert.equal(out.status, "completed", JSON.stringify(out));
  return out.status === "completed" ? out.result : undefined;
}

test("a server stream reads every book, or stops at the limit and says so", async () => {
  const { runtime } = setup({ direct: true });
  const all = await runtime.execute({
    code: `async () => {
      const { messages, complete } = await library.streamBooks({ parent: "shelves/poetry" });
      return [messages.map((b) => b.title), complete];
    }`,
  });
  assert.deepEqual(completed(all), [["Leaves of Grass", "Odes", "Ariel"], true]);
  const two = await runtime.execute({
    code: `async () => {
      const r = await library.streamBooks({ parent: "shelves/poetry" }, { limit: 2 });
      return [r.messages.length, r.complete];
    }`,
  });
  assert.deepEqual(completed(two), [2, false]);
});

test("a client stream waits for approval, adds every book, and replays without adding them twice", async () => {
  const { runtime, library } = setup({ direct: true });
  const before = library.books.size;
  const paused = await runtime.execute({
    code: `async () => {
      const shelf = await library.getShelf({ name: "shelves/fiction" });
      const { imported } = await library.importBooks([
        { parent: shelf.name, book: { title: "Dune" } },
        { parent: shelf.name, book: { title: "Emma" } },
      ]);
      return imported;
    }`,
  });
  assert.equal(paused.status, "paused", JSON.stringify(paused));
  assert.equal(library.books.size, before, "nothing is added before approval");
  if (paused.status !== "paused") return;
  const done = await runtime.approve({ executionId: paused.executionId, seq: paused.pending[0]?.seq ?? 0 });
  assert.equal(completed(done), 2);
  assert.equal(library.books.size, before + 2);
});

test("a bidi stream answers each request, and a failure reaches the script with its code", async () => {
  const { runtime } = setup({ direct: true });
  const out = await runtime.execute({
    code: `async () => {
      const found = await library.lookupBooks([{ name: "shelves/poetry/books/odes" }, { name: "shelves/poetry/books/ariel" }]);
      let code = "";
      try { await library.lookupBooks([{ name: "shelves/poetry/books/missing" }]); } catch (e) { code = e.code; }
      return [found.messages.map((b) => b.title), found.complete, code];
    }`,
  });
  assert.deepEqual(completed(out), [["Odes", "Ariel"], true, "NOT_FOUND"]);
});

test("the model is told how streams behave, in the tool and in describe()", async () => {
  const { runtime } = setup({ direct: true });
  assert.match(runtime.tool().description, /Streaming methods are one call each/);
  const { types } = await runtime.describe("library.lookupBooks");
  assert.match(
    types,
    /lookupBooks\(inputs: LookupBooksRequest\[\], options\?: StreamOptions\): Promise<StreamResult<Book>>/,
  );
});
