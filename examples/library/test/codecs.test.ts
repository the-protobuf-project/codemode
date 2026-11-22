// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { test } from "node:test";
import { fromBinary, toBinary } from "@the-protobuf-project/codemode";
import { libraryV1 } from "../gen/index.ts";

const { BookCodec, CreateBookRequestCodec, ListBooksRequestCodec, BookName, ShelfName, GenreCodec } = libraryV1;

const book: libraryV1.Book = {
  name: "shelves/poetry/books/leaves",
  title: "Leaves of Grass",
  authors: ["Walt Whitman"],
  state: "BOOK_STATE_LENT",
  isbn: "9780140421996",
  pageCount: "9007199254740993",
  rating: 4.5,
  editions: [{ number: 1, year: 1855, publisher: "Self" }],
  labels: { language: "en" },
  dueTime: "2026-12-01T00:00:00Z",
  readingTime: "36000s",
};

test("Book round-trips through the binary format", () => {
  const decoded = fromBinary(BookCodec, toBinary(BookCodec, book));
  assert.deepEqual(BookCodec.toJson(decoded), BookCodec.toJson(book));
  assert.equal(decoded.pageCount, "9007199254740993", "64-bit integers keep full precision");
});

test("Book round-trips through protojson, dropping defaults", () => {
  const json = BookCodec.toJson({ ...book, borrower: "", authors: [] });
  assert.equal("borrower" in (json as object), false);
  assert.equal("authors" in (json as object), false);
  assert.deepEqual(BookCodec.toJson(BookCodec.fromJson(json)), json);
});

test("fromJson accepts proto field names and fills defaults", () => {
  const b = BookCodec.fromJson({ title: "T", page_count: 3, reading_time: "1.5s" });
  assert.equal(b.pageCount, "3");
  assert.equal(b.readingTime, "1.500s");
  assert.equal(b.state, "BOOK_STATE_UNSPECIFIED");
  assert.deepEqual(b.authors, []);
});

test("fromJson rejects bad input with a path", () => {
  assert.throws(() => BookCodec.fromJson({ title: 5 }), /library\.v1\.Book\.title/);
  assert.throws(() => BookCodec.fromJson({ title: "T", state: "BOOK_STATE_STOLEN" }), { code: "INVALID_ARGUMENT" });
  assert.throws(() => BookCodec.fromJson({ title: "T", isbn: "1", doi: "2" }), /oneof identifier/);
  assert.throws(() => BookCodec.fromJson({ title: "T", editions: [{ year: "soon" }] }), /editions\[0\]\.year/);
});

test("a REQUIRED message field must be present", () => {
  assert.throws(() => CreateBookRequestCodec.fromJson({ parent: "shelves/a" }), /book: required/);
});

test("the oneof keeps only the last member on decode", () => {
  const bytes = new Uint8Array([...toBinary(BookCodec, { title: "T", isbn: "1" }), ...toBinary(BookCodec, { title: "T", doi: "2" })]);
  const decoded = fromBinary(BookCodec, bytes);
  assert.equal(decoded.doi, "2");
  assert.equal(decoded.isbn, undefined);
});

test("enums accept numbers and encode as numbers", () => {
  const req = ListBooksRequestCodec.fromJson({ parent: "shelves/a", state: 2 });
  assert.equal(req.state, "BOOK_STATE_LENT");
  assert.equal(GenreCodec.toNumber("GENRE_POETRY"), 3);
});

test("resource names format, parse and match", () => {
  assert.equal(BookName.format({ shelf: "a", book: "b" }), "shelves/a/books/b");
  assert.deepEqual(BookName.parse("shelves/a/books/b"), { shelf: "a", book: "b" });
  assert.equal(ShelfName.matches("shelves/a/books/b"), false);
  assert.throws(() => BookName.format({ shelf: "a/b", book: "c" }), { code: "INVALID_ARGUMENT" });
});
