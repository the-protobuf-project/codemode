// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Catalog is the Code Mode model: every package with generated types, the
// services exposed in the sandbox, and the snippets shipped with them.
type Catalog struct {
	Packages   []*Package
	Services   []*Service
	Snippets   []*Snippet
	Config     *Config
	Version    string // protoc version, for file headers
	byMessage  map[protoreflect.FullName]*Message
	byEnum     map[protoreflect.FullName]*Enum
	byResource map[string]*Resource
	aip        *aip
}

// Package groups the generated files of one proto package.
type Package struct {
	Name      protoreflect.FullName // "library.v1"
	Dir       string                // "library/v1"
	Doc       string                // comments above `package`
	Messages  []*Message
	Enums     []*Enum
	Resources []*Resource
	Services  []*Service
}

// Message is one generated interface and its codec.
type Message struct {
	Desc     protoreflect.MessageDescriptor
	Package  *Package
	Name     string // TypeScript name: "Book", nested "BookEdition"
	Sandbox  string // unique name in sandbox.d.ts
	Path     string // output path without extension: "library/v1/messages/book"
	Source   string // the .proto it came from
	Doc      string
	Fields   []*Field
	Oneofs   []*Oneof
	Resource *Resource // set when the message is a google.api.resource
}

// Codec is the name of the message's codec constant.
func (m *Message) Codec() string { return m.Name + "Codec" }

// Field is one message field, resolved to its TypeScript shape.
type Field struct {
	Desc       protoreflect.FieldDescriptor
	JSONName   string // lowerCamel property name
	ProtoName  string
	Doc        string
	Required   bool // REQUIRED field_behavior or buf.validate required
	OutputOnly bool
	Explicit   bool // has presence: proto3 optional, proto2 optional, message
	Oneof      *Oneof
	Message    *Message // message-typed fields, except well-known types
	Enum       *Enum
	WKT        protoreflect.FullName // well-known type, e.g. google.protobuf.Timestamp
	Resource   *Resource             // string fields that hold a resource name
	MapKey     *Field                // map fields: the entry's key and value
	MapValue   *Field
}

// Oneof is a set of fields of which at most one is set.
type Oneof struct {
	Name   string
	Doc    string
	Fields []*Field
}

// Enum is one generated string-union type and its codec.
type Enum struct {
	Desc    protoreflect.EnumDescriptor
	Package *Package
	Name    string
	Sandbox string
	Path    string // output path without extension
	Source  string
	Doc     string
	Values  []EnumValue
}

// Codec is the name of the enum's codec constant.
func (e *Enum) Codec() string { return e.Name + "Codec" }

// EnumValue is one enum value.
type EnumValue struct {
	Name       string
	Number     int32
	Doc        string
	Deprecated bool
}

// Resource is a google.api.resource pattern, generated as a typed name.
type Resource struct {
	Type     string // "library.example.com/Book"
	Pattern  string // "shelves/{shelf}/books/{book}"
	Package  *Package
	Name     string // "BookName"
	Sandbox  string
	Path     string // output path without extension
	Vars     []string
	Template string // "shelves/${string}/books/${string}"
}

// Service is one gRPC service exposed in the sandbox.
type Service struct {
	Proto        *protogen.Service
	Package      *Package
	Name         string // "LibraryService"
	Dir          string // "library/v1/services/library_service"
	Source       string
	Namespace    string // sandbox global, e.g. "library"
	Instructions string
	Doc          string
	Methods      []*Method
}

// Method is one unary RPC exposed as a sandbox method.
type Method struct {
	Proto       *protogen.Method
	Service     *Service
	Name        string // "createBook"
	Path        string // "library.createBook"
	Doc         string
	Input       TypeRef
	Output      TypeRef
	Approval    bool
	Replay      string // "log" or "reexecute"
	ReadOnly    bool
	Idempotent  bool
	Revert      *Revert
	HTTP        *HTTPRule
	Examples    []string
	ApprovalMsg string // elicitation message shown when approval is asked
}

// TypeRef is a message type used as an RPC input or output.
type TypeRef struct {
	Message *Message
	WKT     protoreflect.FullName // well-known type, e.g. google.protobuf.Empty
}

// Revert is a compensating call, resolved against the service's methods.
type Revert struct {
	Method *Method
	Fields []RevertField
}

// RevertField sets one field of the revert request from args or result.
type RevertField struct {
	Field  *Field
	Source string // "args" or "result"
	Path   []string
}

// HTTPRule is a google.api.http binding.
type HTTPRule struct {
	Method       string
	Path         string
	Body         string
	ResponseBody string
}

// Snippet is reusable code shipped in codemode.v1.service.snippets.
type Snippet struct {
	Name        string
	Description string
	Code        string
	Input       *Message
}
