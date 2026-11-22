// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "google.golang.org/protobuf/reflect/protoreflect"

// scalar describes how one protobuf scalar kind maps to TypeScript, protojson
// and the binary format.
type scalar struct {
	TS       string // TypeScript type
	JSON     string // protojson reader
	Method   string // BinaryWriter and BinaryReader method
	WireType string // WireType member
	Zero     string // default value literal
	Float    bool   // written through protojson.floatOut
	Packable bool
}

var scalars = map[protoreflect.Kind]scalar{
	protoreflect.DoubleKind:   {TS: "number", JSON: "float", Method: "double", WireType: "Fixed64", Zero: "0", Float: true, Packable: true},
	protoreflect.FloatKind:    {TS: "number", JSON: "float", Method: "float", WireType: "Fixed32", Zero: "0", Float: true, Packable: true},
	protoreflect.Int32Kind:    {TS: "number", JSON: "int32", Method: "int32", WireType: "Varint", Zero: "0", Packable: true},
	protoreflect.Sint32Kind:   {TS: "number", JSON: "int32", Method: "sint32", WireType: "Varint", Zero: "0", Packable: true},
	protoreflect.Sfixed32Kind: {TS: "number", JSON: "int32", Method: "sfixed32", WireType: "Fixed32", Zero: "0", Packable: true},
	protoreflect.Uint32Kind:   {TS: "number", JSON: "uint32", Method: "uint32", WireType: "Varint", Zero: "0", Packable: true},
	protoreflect.Fixed32Kind:  {TS: "number", JSON: "uint32", Method: "fixed32", WireType: "Fixed32", Zero: "0", Packable: true},
	protoreflect.Int64Kind:    {TS: "string", JSON: "int64", Method: "int64", WireType: "Varint", Zero: `"0"`, Packable: true},
	protoreflect.Sint64Kind:   {TS: "string", JSON: "int64", Method: "sint64", WireType: "Varint", Zero: `"0"`, Packable: true},
	protoreflect.Sfixed64Kind: {TS: "string", JSON: "int64", Method: "sfixed64", WireType: "Fixed64", Zero: `"0"`, Packable: true},
	protoreflect.Uint64Kind:   {TS: "string", JSON: "uint64", Method: "uint64", WireType: "Varint", Zero: `"0"`, Packable: true},
	protoreflect.Fixed64Kind:  {TS: "string", JSON: "uint64", Method: "fixed64", WireType: "Fixed64", Zero: `"0"`, Packable: true},
	protoreflect.BoolKind:     {TS: "boolean", JSON: "bool", Method: "bool", WireType: "Varint", Zero: "false", Packable: true},
	protoreflect.StringKind:   {TS: "string", JSON: "string", Method: "string", WireType: "LengthDelimited", Zero: `""`},
	protoreflect.BytesKind:    {TS: "string", JSON: "bytes", Method: "bytes", WireType: "LengthDelimited", Zero: `""`},
	protoreflect.EnumKind:     {Method: "int32", WireType: "Varint", Packable: true},
}

// is64Bit reports whether a kind is a 64-bit integer, carried as a string.
func is64Bit(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return true
	}
	return false
}
