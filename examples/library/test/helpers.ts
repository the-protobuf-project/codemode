// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

import { CodemodeRuntime, MemoryStore, type Store } from "@the-protobuf-project/codemode";
import { QuickJSExecutor } from "@the-protobuf-project/codemode/quickjs";
import { connectHandler, connectTransport } from "@the-protobuf-project/network";
import { catalog, config, connectors, libraryV1 } from "../gen/index.ts";
import { InMemoryLibrary } from "../src/library.ts";

/**
 * A runtime over a fresh seeded library, reached through the generated client over Connect,
 * or called in-process with direct: connectHandler serves unary calls only, so streaming
 * RPCs are tested in-process (and against the Go and Rust servers in runtime-ts).
 */
export function setup(options: { store?: Store; discovery?: "inline" | "progressive"; direct?: boolean } = {}) {
  const library = InMemoryLibrary.seeded();
  const handler = connectHandler([{ descriptor: libraryV1.LibraryServiceDescriptor, impl: library }]);
  const client = new libraryV1.LibraryServiceClient(
    connectTransport({ baseUrl: "http://library.test", fetch: (url, init) => handler(new Request(url, init)) }),
  );
  const runtime = new CodemodeRuntime({
    catalog,
    config: {
      ...config,
      discovery: options.discovery ?? config.discovery,
      executor: { ...config.executor, timeoutMs: 2000 },
    },
    connectors: connectors({ library: options.direct ? library : client }),
    executor: new QuickJSExecutor(),
    store: options.store ?? new MemoryStore(),
  });
  return { library, client, runtime };
}
