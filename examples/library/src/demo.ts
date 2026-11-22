// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Walks through Code Mode against the in-memory library: a read, an approval
// pause, resume, and rollback. Run with `bun run demo`.

import { type CodeOutput, createCodemode } from "@the-protobuf-project/codemode";
import { QuickJSExecutor } from "@the-protobuf-project/codemode/quickjs";
import { catalog, config, connectors } from "../gen/index.ts";
import { InMemoryLibrary } from "./library.ts";

const library = InMemoryLibrary.seeded();
const runtime = createCodemode({
  catalog,
  config,
  connectors: connectors({ library }),
  executor: new QuickJSExecutor(),
});

function show(title: string, out: unknown): void {
  console.log(`\n== ${title}\n${JSON.stringify(out, null, 2)}`);
}

console.log(runtime.tool().description.split("\n\nThe codemode API")[0]);

show(
  "Read: titles on the poetry shelf",
  await runtime.execute({
    code: `async () => {
    const page = await library.listBooks({ parent: "shelves/poetry" });
    return page.books.map((b) => b.title);
  }`,
  }),
);

const lend: CodeOutput = await runtime.execute({
  code: `async () => {
    const book = await library.createBook({ parent: "shelves/poetry", book: { title: "The Waste Land", authors: ["T. S. Eliot"] } });
    await library.lendBook({ name: book.name, borrower: "ada", days: 7 });
    const archived = await library.archiveShelf({ name: "shelves/fiction" });
    return { book: book.name, archived: archived.archivedCount };
  }`,
});
show("Write: pauses before the destructive archiveShelf", lend);

if (lend.status === "paused") {
  show(
    "Approved: resumes, replaying the create and lend from the log",
    await runtime.approve({ executionId: lend.executionId }),
  );
  show("Rollback: lend and create are undone, newest first", await runtime.rollback({ executionId: lend.executionId }));
}

console.log(
  `\n== Discovery: describe a method\n${(await runtime.describe("library.lendBook")).types.split("\n").slice(-6).join("\n")}`,
);
