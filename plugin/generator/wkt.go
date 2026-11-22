// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "google.golang.org/protobuf/reflect/protoreflect"

// WellKnown is a google.protobuf type the runtime ships a codec for. Its
// TypeScript shape is its protojson form, so the model, the sandbox and the
// client all see the same value.
type WellKnown struct {
	Type    string // TypeScript type, e.g. "string"
	Codec   string // runtime export, e.g. "TimestampCodec"
	Imports []string
	Doc     string
}

var wellKnown = map[protoreflect.FullName]*WellKnown{
	"google.protobuf.Timestamp":   {Type: "string", Codec: "TimestampCodec", Doc: "RFC 3339 timestamp, e.g. \"2026-11-22T10:00:00Z\"."},
	"google.protobuf.Duration":    {Type: "string", Codec: "DurationCodec", Doc: "Duration in seconds with an \"s\" suffix, e.g. \"1.5s\"."},
	"google.protobuf.FieldMask":   {Type: "string", Codec: "FieldMaskCodec", Doc: "Comma-separated lowerCamel field paths, e.g. \"title,author\"."},
	"google.protobuf.Struct":      {Type: "JsonObject", Codec: "StructCodec", Imports: []string{"JsonObject"}},
	"google.protobuf.Value":       {Type: "JsonValue", Codec: "ValueCodec", Imports: []string{"JsonValue"}},
	"google.protobuf.ListValue":   {Type: "JsonValue[]", Codec: "ListValueCodec", Imports: []string{"JsonValue"}},
	"google.protobuf.Empty":       {Type: "Empty", Codec: "EmptyCodec", Imports: []string{"Empty"}},
	"google.protobuf.Any":         {Type: "AnyMessage", Codec: "AnyCodec", Imports: []string{"AnyMessage"}},
	"google.protobuf.DoubleValue": {Type: "number | null", Codec: "DoubleValueCodec"},
	"google.protobuf.FloatValue":  {Type: "number | null", Codec: "FloatValueCodec"},
	"google.protobuf.Int64Value":  {Type: "string | null", Codec: "Int64ValueCodec", Doc: "64-bit integer as a decimal string."},
	"google.protobuf.UInt64Value": {Type: "string | null", Codec: "UInt64ValueCodec", Doc: "64-bit integer as a decimal string."},
	"google.protobuf.Int32Value":  {Type: "number | null", Codec: "Int32ValueCodec"},
	"google.protobuf.UInt32Value": {Type: "number | null", Codec: "UInt32ValueCodec"},
	"google.protobuf.BoolValue":   {Type: "boolean | null", Codec: "BoolValueCodec"},
	"google.protobuf.StringValue": {Type: "string | null", Codec: "StringValueCodec"},
	"google.protobuf.BytesValue":  {Type: "string | null", Codec: "BytesValueCodec", Doc: "Bytes as base64."},
}

// sandboxPrelude declares the runtime types well-known types refer to, once,
// at the top of sandbox.d.ts.
const sandboxPrelude = `/** Any JSON value. */
type JsonValue = null | boolean | number | string | JsonValue[] | JsonObject;
/** A JSON object. */
interface JsonObject { [key: string]: JsonValue }
/** google.protobuf.Empty. */
type Empty = Record<string, never>;
/** google.protobuf.Any: "@type" plus the packed message's fields. */
interface AnyMessage { "@type": string; [field: string]: unknown }
`

// lookupWellKnown returns the runtime codec for a google.protobuf message.
func lookupWellKnown(name protoreflect.FullName) *WellKnown {
	return wellKnown[name]
}
