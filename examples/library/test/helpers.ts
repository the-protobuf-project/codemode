// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import { CodemodeRuntime, MemoryStore, type Store } from "@the-protobuf-project/runtime-ts-codemode";
import { connectHandler, connectTransport } from "@the-protobuf-project/runtime-ts-network";
import { QuickJSExecutor } from "@the-protobuf-project/runtime-ts-codemode/quickjs";
import { catalog, config, connectors, libraryV1 } from "../gen/index.ts";
import { InMemoryLibrary } from "../src/library.ts";

/** A runtime over a fresh seeded library, reached through the generated client over Connect. */
export function setup(options: { store?: Store; discovery?: "inline" | "progressive" } = {}) {
  const library = InMemoryLibrary.seeded();
  const handler = connectHandler([{ descriptor: libraryV1.LibraryServiceDescriptor, impl: library }]);
  const client = new libraryV1.LibraryServiceClient(
    connectTransport({ baseUrl: "http://library.test", fetch: (url, init) => handler(new Request(url, init)) }),
  );
  const runtime = new CodemodeRuntime({
    catalog,
    config: { ...config, discovery: options.discovery ?? config.discovery, executor: { ...config.executor, timeoutMs: 2000 } },
    connectors: connectors({ library: client }),
    executor: new QuickJSExecutor(),
    store: options.store ?? new MemoryStore(),
  });
  return { library, client, runtime };
}
