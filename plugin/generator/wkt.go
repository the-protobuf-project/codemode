// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "google.golang.org/protobuf/reflect/protoreflect"

// wellKnownTypes are the google.protobuf messages every language's runtime
// ships natively, so they are referenced by name and never generated.
var wellKnownTypes = map[protoreflect.FullName]bool{
	"google.protobuf.Timestamp": true, "google.protobuf.Duration": true, "google.protobuf.FieldMask": true,
	"google.protobuf.Struct": true, "google.protobuf.Value": true, "google.protobuf.ListValue": true,
	"google.protobuf.Empty": true, "google.protobuf.Any": true,
	"google.protobuf.DoubleValue": true, "google.protobuf.FloatValue": true, "google.protobuf.Int64Value": true,
	"google.protobuf.UInt64Value": true, "google.protobuf.Int32Value": true, "google.protobuf.UInt32Value": true,
	"google.protobuf.BoolValue": true, "google.protobuf.StringValue": true, "google.protobuf.BytesValue": true,
}

// isWellKnown reports whether a message is a runtime-provided well-known type.
func isWellKnown(name protoreflect.FullName) bool {
	return wellKnownTypes[name]
}
