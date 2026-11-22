// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Codecs and the wire format, used by generated messages.
export { binary, fromBinary, fromJsonString, toBinary, toJsonString, type ValueReader } from "./codec/binary.ts";
export { decodeBase64, encodeBase64 } from "./codec/base64.ts";
export { enumCodec } from "./codec/enum.ts";
export { protojson, type JsonReader } from "./codec/protojson.ts";
export { resourceName, type ResourceName } from "./codec/resource.ts";
export type { EnumCodec, JsonObject, JsonValue, MessageCodec, MessageOf, ToJsonOptions } from "./codec/types.ts";
export { BinaryReader } from "./wire/reader.ts";
export { BinaryWriter, WireType, type Int64Input } from "./wire/writer.ts";
export { lookupType, registerTypes } from "./registry.ts";

// Well-known types.
export { AnyCodec, EmptyCodec, type AnyMessage, type Empty } from "./wkt/any.ts";
export { FieldMaskCodec } from "./wkt/field_mask.ts";
export { ListValueCodec, StructCodec, ValueCodec } from "./wkt/struct.ts";
export { DurationCodec, TimestampCodec } from "./wkt/time.ts";
export {
  BoolValueCodec, BytesValueCodec, DoubleValueCodec, FloatValueCodec, Int32ValueCodec,
  Int64ValueCodec, StringValueCodec, UInt32ValueCodec, UInt64ValueCodec,
} from "./wkt/wrappers.ts";

// Errors.
export {
  CodemodeError, codeFromConnect, codeFromHttpStatus, codeFromNumber, isCode, toCodemodeError,
  type Code, type CodemodeErrorOptions,
} from "./errors.ts";

// Transports, used by generated clients.
export {
  callHeaders, callSignal,
  type CallOptions, type Fetch, type FetchTransportOptions, type HttpRule, type MethodDescriptor, type Transport,
} from "./transport/types.ts";
export { httpTransport } from "./transport/http.ts";
export { connectTransport, type ConnectTransportOptions } from "./transport/connect.ts";
export { connectHandler, type ServiceDescriptor, type ServiceImpl } from "./transport/connect_handler.ts";

// Connectors, used by generated connectors.
export { connector, unaryMethod, type ConnectorSpec, type UnaryMethodSpec } from "./connector/connector.ts";
export type { CallContext, Connector, ConnectorMethod, ConnectorOptions, Replay } from "./connector/types.ts";

// Catalog and discovery.
export type { Catalog, CatalogMethod, CatalogNamespace, CatalogSnippet, CodemodeConfig } from "./catalog/types.ts";
export { describe, search, type Description, type SearchResult, type SearchResults } from "./catalog/search.ts";

// The durable runtime.
export { createCodemode, type CodemodeOptions } from "./runtime/create.ts";
export { CodemodeRuntime } from "./runtime/runtime.ts";
export type {
  CodeInput, CodeOutput, Execution, ExecutionStatus, Hooks, LogEntry, PendingAction, RollbackResult, SerializedError, Snippet, TerminalStatus,
} from "./runtime/types.ts";
export type { CodemodeTool } from "./runtime/tool.ts";
export { truncateResult, truncateText, type TruncateOptions } from "./runtime/truncate.ts";

// Executors and stores.
export type { Executor, ExecuteRequest, ExecuteResult, HostCall } from "./executor/types.ts";
export { normalizeCode } from "./executor/normalize.ts";
export { MemoryStore } from "./store/memory.ts";
export type { Store } from "./store/types.ts";
export { SqliteStore, durableObjectSql, type DurableObjectSql, type SqliteDatabase } from "./store/sqlite.ts";
export type { SaveSnippetOptions } from "./runtime/snippets.ts";
export { DEFAULT_MAX_EXECUTIONS, DEFAULT_PAUSED_TTL_MS, MAX_DURABLE_VALUE_BYTES } from "./runtime/values.ts";
