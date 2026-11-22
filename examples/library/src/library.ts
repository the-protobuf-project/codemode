// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import { StatusError } from "@the-protobuf-project/runtime-ts-errors";
import type { Empty } from "@the-protobuf-project/runtime-ts-protobuf";
import { libraryV1 } from "../gen/index.ts";

type Book = libraryV1.Book;
type Shelf = libraryV1.Shelf;

const { BookName, ShelfName } = libraryV1;

/**
 * An in-memory LibraryService. It implements the generated interface
 * directly, so the connector can call it without a network hop; tests and
 * the demo serve it over Connect too.
 */
export class InMemoryLibrary implements libraryV1.LibraryService {
  readonly shelves = new Map<string, Shelf>();
  readonly books = new Map<string, Book>();
  private nextId = 1;

  /** Adds the sample shelves and books the demo and tests start from. */
  static seeded(): InMemoryLibrary {
    const lib = new InMemoryLibrary();
    const poetry = lib.put({ name: ShelfName.format({ shelf: "poetry" }), displayName: "Poetry", genre: "GENRE_POETRY" });
    lib.put({ name: ShelfName.format({ shelf: "fiction" }), displayName: "Fiction", genre: "GENRE_FICTION" });
    for (const [id, title, author] of [
      ["leaves", "Leaves of Grass", "Walt Whitman"],
      ["odes", "Odes", "John Keats"],
      ["ariel", "Ariel", "Sylvia Plath"],
    ] as const) {
      lib.books.set(BookName.format({ shelf: "poetry", book: id }), {
        name: BookName.format({ shelf: "poetry", book: id }), title, authors: [author],
        state: "BOOK_STATE_AVAILABLE", pageCount: "120", createTime: "2026-01-01T00:00:00Z",
      });
    }
    poetry.bookCount = 3;
    return lib;
  }

  private put(shelf: Shelf): Shelf {
    const s: Shelf = { bookCount: 0, createTime: new Date().toISOString(), ...shelf };
    this.shelves.set(s.name as string, s);
    return s;
  }

  private shelf(name: string): Shelf {
    return this.shelves.get(name) ?? fail("NOT_FOUND", `${name} does not exist`);
  }

  private book(name: string): Book {
    return this.books.get(name) ?? fail("NOT_FOUND", `${name} does not exist`);
  }

  private count(shelf: string): void {
    this.shelf(shelf).bookCount = [...this.books.keys()].filter((n) => n.startsWith(`${shelf}/`)).length;
  }

  async listShelves(req: libraryV1.ListShelvesRequest): Promise<libraryV1.ListShelvesResponse> {
    return page([...this.shelves.values()], req.pageSize, req.pageToken, (shelves, next) => ({ shelves, nextPageToken: next }));
  }

  async getShelf(req: libraryV1.GetShelfRequest): Promise<Shelf> {
    return this.shelf(req.name);
  }

  async createShelf(req: libraryV1.CreateShelfRequest): Promise<Shelf> {
    const id = req.shelfId || `shelf${this.nextId++}`;
    const name = ShelfName.format({ shelf: id });
    if (this.shelves.has(name)) fail("ALREADY_EXISTS", `${name} already exists`);
    return this.put({ ...req.shelf, name, bookCount: 0 });
  }

  async deleteShelf(req: libraryV1.DeleteShelfRequest): Promise<Empty> {
    this.shelf(req.name);
    const held = [...this.books.keys()].filter((n) => n.startsWith(`${req.name}/`));
    if (held.length > 0 && !req.force) fail("FAILED_PRECONDITION", `${req.name} still holds ${held.length} books`);
    for (const n of held) this.books.delete(n);
    this.shelves.delete(req.name);
    return {};
  }

  async listBooks(req: libraryV1.ListBooksRequest): Promise<libraryV1.ListBooksResponse> {
    this.shelf(req.parent);
    const all = [...this.books.values()].filter(
      (b) => b.name?.startsWith(`${req.parent}/`) && (!req.state || req.state === "BOOK_STATE_UNSPECIFIED" || b.state === req.state),
    );
    return page(all, req.pageSize, req.pageToken, (books, next) => ({ books, nextPageToken: next }));
  }

  async getBook(req: libraryV1.GetBookRequest): Promise<Book> {
    return this.book(req.name);
  }

  async createBook(req: libraryV1.CreateBookRequest): Promise<Book> {
    this.shelf(req.parent);
    const name = BookName.format({ shelf: ShelfName.parse(req.parent).shelf, book: `book${this.nextId++}` });
    const book: Book = { ...req.book, name, state: "BOOK_STATE_AVAILABLE", createTime: new Date().toISOString() };
    this.books.set(name, book);
    this.count(req.parent);
    return book;
  }

  async updateBook(req: libraryV1.UpdateBookRequest): Promise<Book> {
    const current = this.book(req.book.name ?? "");
    const paths = req.updateMask ? req.updateMask.split(",") : Object.keys(req.book);
    const next: Record<string, unknown> = { ...current };
    const changes: Record<string, unknown> = { ...req.book };
    for (const p of paths) {
      if (p !== "name" && p !== "state" && p !== "createTime") next[p] = changes[p];
    }
    const updated = libraryV1.BookCodec.fromJson(next);
    this.books.set(current.name as string, updated);
    return updated;
  }

  async deleteBook(req: libraryV1.DeleteBookRequest): Promise<Empty> {
    this.book(req.name);
    this.books.delete(req.name);
    this.count(`shelves/${BookName.parse(req.name).shelf}`);
    return {};
  }

  async lendBook(req: libraryV1.LendBookRequest): Promise<Book> {
    const book = this.book(req.name);
    if (book.state !== "BOOK_STATE_AVAILABLE") fail("FAILED_PRECONDITION", `${req.name} is not available`);
    const due = new Date(Date.now() + (req.days || 14) * 86_400_000).toISOString().replace(/\.\d+Z$/, "Z");
    Object.assign(book, { state: "BOOK_STATE_LENT", borrower: req.borrower, dueTime: due });
    return book;
  }

  async returnBook(req: libraryV1.ReturnBookRequest): Promise<Book> {
    const book = this.book(req.name);
    if (book.state !== "BOOK_STATE_LENT") fail("FAILED_PRECONDITION", `${req.name} is not lent out`);
    Object.assign(book, { state: "BOOK_STATE_AVAILABLE", borrower: "" });
    delete book.dueTime;
    return book;
  }

  async moveBook(req: libraryV1.MoveBookRequest): Promise<Book> {
    const book = this.book(req.name);
    this.shelf(req.destination);
    const from = `shelves/${BookName.parse(req.name).shelf}`;
    const name = BookName.format({ shelf: ShelfName.parse(req.destination).shelf, book: BookName.parse(req.name).book });
    this.books.delete(req.name);
    const moved: Book = { ...book, name };
    this.books.set(name, moved);
    this.count(from);
    this.count(req.destination);
    return moved;
  }

  async archiveShelf(req: libraryV1.ArchiveShelfRequest): Promise<libraryV1.ArchiveShelfResponse> {
    this.shelf(req.name);
    let archivedCount = 0;
    for (const b of this.books.values()) {
      if (b.name?.startsWith(`${req.name}/`) && b.state !== "BOOK_STATE_ARCHIVED") {
        b.state = "BOOK_STATE_ARCHIVED";
        archivedCount++;
      }
    }
    return { archivedCount };
  }
}

function fail(code: "NOT_FOUND" | "ALREADY_EXISTS" | "FAILED_PRECONDITION", message: string): never {
  throw new StatusError(code, message);
}

function page<T, R>(all: T[], size: number | undefined, token: string | undefined, wrap: (items: T[], next: string) => R): R {
  const start = token ? Number(token) : 0;
  const end = start + (size && size > 0 ? size : 50);
  return wrap(all.slice(start, end), end < all.length ? String(end) : "");
}
