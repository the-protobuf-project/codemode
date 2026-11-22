// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// typeNamer chooses names for referenced types. Generated modules import
// them; sandbox.d.ts uses each type's globally unique sandbox name.
type typeNamer interface {
	message(m *Message) string
	enum(e *Enum) string
	resource(r *Resource) string
	wellKnown(name protoreflect.FullName) string
}

// moduleNamer names types inside a generated file, importing as it goes.
type moduleNamer struct{ f *tsFile }

func (n moduleNamer) message(m *Message) string {
	n.f.localType(m.Path, m.Name)
	return m.Name
}

func (n moduleNamer) enum(e *Enum) string {
	n.f.localType(e.Path, e.Name)
	return e.Name
}

func (n moduleNamer) resource(r *Resource) string {
	n.f.localType(r.Path, r.Name)
	return r.Name
}

func (n moduleNamer) wellKnown(name protoreflect.FullName) string {
	w := tsWKT(name)
	n.f.runtimeType(w.Imports...)
	return w.Type
}

// sandboxNamer names types in sandbox.d.ts, where everything is global.
type sandboxNamer struct{}

func (sandboxNamer) message(m *Message) string                   { return m.Sandbox }
func (sandboxNamer) enum(e *Enum) string                         { return e.Sandbox }
func (sandboxNamer) resource(r *Resource) string                 { return r.Sandbox }
func (sandboxNamer) wellKnown(name protoreflect.FullName) string { return tsWKT(name).Type }

// tsType is a field's TypeScript type.
func tsType(f *Field, n typeNamer) string {
	if f.MapKey != nil {
		return "Record<string, " + tsType(f.MapValue, n) + ">"
	}
	t := elemType(f, n)
	if f.Desc.IsList() {
		if strings.Contains(t, " ") {
			t = "(" + t + ")"
		}
		return t + "[]"
	}
	return t
}

// elemType is the type of one element of a field.
func elemType(f *Field, n typeNamer) string {
	switch {
	case f.Resource != nil:
		return n.resource(f.Resource)
	case f.Enum != nil:
		return n.enum(f.Enum)
	case f.Message != nil:
		return n.message(f.Message)
	case f.WKT != "":
		return n.wellKnown(f.WKT)
	}
	return scalars[f.Desc.Kind()].TS
}

// refType is the TypeScript type of an RPC input or output.
func refType(t TypeRef, n typeNamer) string {
	if t.WKT != "" {
		return n.wellKnown(t.WKT)
	}
	return n.message(t.Message)
}

// isScalar reports a field whose elements are strings, numbers or booleans.
func isScalar(f *Field) bool {
	return f.Message == nil && f.WKT == "" && f.Desc.Kind() != protoreflect.GroupKind
}

// zero is the default value of a scalar or enum element.
func zero(file *tsFile, f *Field, n typeNamer) string {
	switch {
	case f.Enum != nil:
		if file != nil {
			file.local(f.Enum.Path, f.Enum.Codec())
		}
		return f.Enum.Codec() + ".defaultValue"
	case f.Resource != nil:
		return `"" as ` + n.resource(f.Resource)
	}
	return scalars[f.Desc.Kind()].Zero
}

// fieldDoc renders the JSDoc lines for a field: its comment, then notes
// derived from annotations.
func fieldNotes(f *Field) []string {
	var notes []string
	if f.Desc.Kind() == protoreflect.MessageKind && f.WKT != "" && tsWKT(f.WKT).Doc != "" {
		notes = append(notes, tsWKT(f.WKT).Doc)
	}
	if is64Bit(f.Desc.Kind()) {
		notes = append(notes, "64-bit integer as a decimal string.")
	}
	if f.Desc.Kind() == protoreflect.BytesKind {
		notes = append(notes, "Bytes as base64.")
	}
	if f.Resource != nil {
		notes = append(notes, "Resource name: `"+f.Resource.Pattern+"`.")
	}
	if f.Required {
		notes = append(notes, "Required.")
	}
	if f.OutputOnly {
		notes = append(notes, "Output only; ignored in requests.")
	}
	if f.Oneof != nil {
		names := make([]string, len(f.Oneof.Fields))
		for i, o := range f.Oneof.Fields {
			names[i] = o.JSONName
		}
		notes = append(notes, "Oneof `"+f.Oneof.Name+"`: set at most one of "+strings.Join(names, ", ")+".")
	}
	return notes
}
