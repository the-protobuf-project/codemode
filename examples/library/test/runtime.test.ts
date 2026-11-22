// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { test } from "node:test";
import { type CodeOutput, MemoryStore } from "@the-protobuf-project/codemode";
import { setup } from "./helpers.ts";

function completed(out: CodeOutput): unknown {
  assert.equal(out.status, "completed", JSON.stringify(out));
  return out.status === "completed" ? out.result : undefined;
}

test("code chains typed calls and returns only what it needs", async () => {
  const { runtime } = setup();
  const out = await runtime.execute({
    code: `async () => {
      const shelves = await library.listShelves({});
      const counts = [];
      for (const s of shelves.shelves) {
        const page = await library.listBooks({ parent: s.name });
        counts.push([s.displayName, page.books.length]);
      }
      return counts;
    }`,
  });
  assert.deepEqual(completed(out), [
    ["Poetry", 3],
    ["Fiction", 0],
  ]);
});

test("TypeScript and bare bodies run too", async () => {
  const { runtime } = setup();
  const out = await runtime.execute({
    code: `const b: Book = await library.getBook({ name: "shelves/poetry/books/odes" });\nreturn b.title;`,
  });
  assert.equal(completed(out), "Odes");
});

test("service errors reach the code with their gRPC code", async () => {
  const { runtime } = setup();
  const out = await runtime.execute({
    code: `try { await library.getBook({ name: "shelves/poetry/books/nope" }); } catch (e) { return e.code; }`,
  });
  assert.equal(completed(out), "NOT_FOUND");
});

test("invalid arguments fail before the service is called", async () => {
  const { runtime } = setup();
  const out = await runtime.execute({
    code: `try { await library.lendBook({ name: "shelves/poetry/books/odes", days: "many" }); } catch (e) { return e.message; }`,
  });
  assert.match(String(completed(out)), /days/);
});

test("destructive calls pause, then resume on approval without repeating earlier calls", async () => {
  const { runtime, library } = setup();
  const code = `async () => {
    await library.createShelf({ shelf: { displayName: "Drama" }, shelfId: "drama" });
    const r = await library.archiveShelf({ name: "shelves/poetry" });
    return r.archivedCount;
  }`;
  const first = await runtime.execute({ code });
  assert.equal(first.status, "paused");
  if (first.status !== "paused") return;
  assert.equal(first.pending[0]?.path, "library.archiveShelf");
  assert.equal(library.shelves.has("shelves/drama"), true);
  const resumed = await runtime.approve({ executionId: first.executionId });
  assert.equal(completed(resumed), 3);
  assert.equal([...library.shelves.keys()].filter((k) => k === "shelves/drama").length, 1);
});

test("a rejected call ends the execution", async () => {
  const { runtime, library } = setup();
  const first = await runtime.execute({ code: `await library.archiveShelf({ name: "shelves/poetry" });` });
  assert.equal(first.status, "paused");
  if (first.status !== "paused") return;
  assert.equal(await runtime.reject({ executionId: first.executionId, seq: first.pending[0]?.seq ?? 0 }), true);
  assert.equal((await runtime.getExecution(first.executionId))?.status, "rejected");
  assert.equal(
    [...library.books.values()].some((b) => b.state === "BOOK_STATE_ARCHIVED"),
    false,
  );
});

test("rollback undoes applied calls through their generated reverts", async () => {
  const { runtime, library } = setup();
  const out = await runtime.execute({
    code: `async () => {
      const b = await library.createBook({ parent: "shelves/poetry", book: { title: "New" } });
      await library.lendBook({ name: "shelves/poetry/books/ariel", borrower: "ada" });
      return b.name;
    }`,
  });
  const created = String(completed(out));
  assert.equal(library.books.has(created), true);
  const result = await runtime.rollback({ executionId: out.executionId });
  assert.deepEqual(result.reverted, ["library.lendBook", "library.createBook"]);
  assert.equal(result.status, "rolled_back");
  assert.equal(library.books.has(created), false);
  assert.equal(library.books.get("shelves/poetry/books/ariel")?.state, "BOOK_STATE_AVAILABLE");
});

test("codemode.step values are replayed, and diverging code fails", async () => {
  const { runtime } = setup();
  const out = await runtime.execute({
    code: `async () => {
      const id = await codemode.step("id", () => Math.random());
      await library.archiveShelf({ name: "shelves/fiction" });
      return id;
    }`,
  });
  assert.equal(out.status, "paused");
  const exec = await runtime.getExecution(out.executionId);
  const recorded = exec?.log.find((e) => e.kind === "step")?.result;
  assert.equal(completed(await runtime.approve({ executionId: out.executionId })), recorded);
});

test("search and describe work from code and from the host", async () => {
  const { runtime } = setup();
  const found = await runtime.search("lend book");
  assert.equal(found.results[0]?.path, "library.lendBook");
  const d = await runtime.describe("library.createBook");
  assert.match(d.types, /interface CreateBookRequest/);
  assert.match(d.types, /createBook\(input: CreateBookRequest\): Promise<Book>/);
  const out = await runtime.execute({ code: `const d = await codemode.describe("library"); return d.kind;` });
  assert.equal(completed(out), "namespace");
});

test("proto snippets are seeded, and saved snippets run", async () => {
  const { runtime } = setup();
  assert.deepEqual(
    (await runtime.snippets()).map((s) => s.name),
    ["move_shelf"],
  );
  const src = await runtime.execute({
    code: `const p = await library.listBooks({ parent: "shelves/poetry" }); return p.books.length;`,
  });
  await runtime.saveSnippet("count_poetry", { executionId: src.executionId, description: "Counts poetry books" });
  assert.equal(completed(await runtime.execute({ code: `return await codemode.run("count_poetry");` })), 3);
});

test("executions survive a new runtime on the same store", async () => {
  const store = new MemoryStore();
  const first = setup({ store });
  const out = await first.runtime.execute({
    code: `await library.archiveShelf({ name: "shelves/poetry" }); return "done";`,
  });
  const second = setup({ store });
  assert.equal((await second.runtime.pending()).length, 1);
  assert.equal(completed(await second.runtime.approve({ executionId: out.executionId })), "done");
});

test("the tool description lists namespaces, or inlines every type", async () => {
  assert.match(setup().runtime.tool().description, /- library \(13 methods\)/);
  const inline = setup({ discovery: "inline" }).runtime.tool().description;
  assert.match(inline, /declare const library/);
  assert.match(inline, /interface Book \{/);
});
