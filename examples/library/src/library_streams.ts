// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// The library's streaming RPCs, built on its unary ones so the same checks apply.

import type { libraryV1 } from "../gen/index.ts";

type Library = Pick<libraryV1.LibraryService, "listBooks" | "getBook" | "createBook">;

/** Each book on a shelf, a page at a time underneath. */
export async function* streamBooks(lib: Library, req: libraryV1.StreamBooksRequest): AsyncGenerator<libraryV1.Book> {
  let pageToken = "";
  do {
    const page = await lib.listBooks({ parent: req.parent, pageSize: 2, pageToken });
    yield* page.books ?? [];
    pageToken = page.nextPageToken ?? "";
  } while (pageToken);
}

/** Adds every book as it arrives, and says how many. */
export async function importBooks(
  lib: Library,
  requests: AsyncIterable<libraryV1.ImportBooksRequest>,
): Promise<libraryV1.ImportBooksResponse> {
  let imported = 0;
  for await (const req of requests) {
    await lib.createBook({ parent: req.parent, book: req.book });
    imported++;
  }
  return { imported };
}

/** Looks up each book as its name arrives. */
export async function* lookupBooks(
  lib: Library,
  requests: AsyncIterable<libraryV1.LookupBooksRequest>,
): AsyncGenerator<libraryV1.Book> {
  for await (const req of requests) yield await lib.getBook({ name: req.name });
}
