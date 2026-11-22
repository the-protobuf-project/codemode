// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "google.golang.org/protobuf/reflect/protoreflect"

// tsWellKnown is how TypeScript represents a google.protobuf type: its
// protojson form, with a codec the runtime ships. The model, the sandbox and
// the client all see the same value.
type tsWellKnown struct {
	Type    string // TypeScript type, e.g. "string"
	Codec   string // runtime export, e.g. "TimestampCodec"
	Imports []string
	Doc     string
}

var tsWellKnownTypes = map[protoreflect.FullName]*tsWellKnown{
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

// tsWKT returns the TypeScript mapping of a well-known type.
func tsWKT(name protoreflect.FullName) *tsWellKnown {
	return tsWellKnownTypes[name]
}
