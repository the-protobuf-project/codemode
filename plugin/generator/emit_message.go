// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// emitMessage writes messages/<name>.ts: the interface and its codec.
func (cat *Catalog) emitMessage(gen *protogen.Plugin, m *Message) {
	f := cat.newFile(m.File, m.Source)
	n := moduleNamer{f}
	f.body.WriteString(interfaceDecl(m, n, "export "))
	f.P()
	emitCodec(f, m, n)
	f.emit(gen)
}

// interfaceDecl renders a message as an interface. The sandbox uses the
// same text with sandbox names and no export.
func interfaceDecl(m *Message, n typeNamer, export string) string {
	var b strings.Builder
	name := n.message(m)
	b.WriteString(jsdoc("", m.Doc, messageNotes(m)...))
	if len(m.Fields) == 0 {
		b.WriteString(export + "type " + name + " = Record<string, never>;\n")
		return b.String()
	}
	b.WriteString(export + "interface " + name + " {\n")
	for i, f := range m.Fields {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(jsdoc("  ", f.Doc, fieldNotes(f)...))
		opt := "?"
		if f.Required {
			opt = ""
		}
		b.WriteString("  " + propertyKey(f.JSONName) + opt + ": " + tsType(f, n) + ";\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func messageNotes(m *Message) []string {
	if m.Resource == nil {
		return nil
	}
	return []string{"Resource `" + m.Resource.Type + "`, named `" + m.Resource.Pattern + "`."}
}

// emitCodec writes the <Name>Codec constant.
func emitCodec(f *tsFile, m *Message, n moduleNamer) {
	f.runtimeType("MessageCodec")
	name := m.Name
	f.P("/** Converts {@link ", name, "} between its interface, protojson and the protobuf binary format. */")
	f.P("export const ", m.Codec(), ": MessageCodec<", name, "> = {")
	f.P("  typeName: ", tsString(string(m.Desc.FullName())), ",")
	f.P()
	emitFromJSON(f, m, n)
	f.P()
	emitToJSON(f, m)
	f.P()
	emitEncode(f, m)
	f.P()
	emitDecode(f, m, n)
	f.P("};")
}

// codecOf returns the codec expression for a message, enum or well-known
// element, importing it.
func codecOf(f *tsFile, fd *Field) string {
	switch {
	case fd.Message != nil:
		f.local(fd.Message.File, fd.Message.Codec())
		return fd.Message.Codec()
	case fd.Enum != nil:
		f.local(fd.Enum.File, fd.Enum.Codec())
		return fd.Enum.Codec()
	case fd.WKT != nil:
		f.runtime(fd.WKT.Codec)
		return fd.WKT.Codec
	}
	return ""
}

// access is the property access for a field on m.
func access(fd *Field) string {
	if propertyKey(fd.JSONName) != fd.JSONName {
		return "m[" + propertyKey(fd.JSONName) + "]"
	}
	return "m." + fd.JSONName
}

// presentCheck is the condition under which encoders write a singular field.
func presentCheck(f *tsFile, fd *Field) string {
	a := access(fd)
	switch {
	case fd.WKT != nil && fd.WKT.Codec == "ValueCodec":
		return a + " !== undefined"
	case fd.Message != nil || fd.WKT != nil:
		return a + " != null"
	case fd.Explicit:
		return a + " !== undefined"
	case fd.Enum != nil:
		return a + " !== undefined && " + a + " !== " + codecOf(f, fd) + ".defaultValue"
	}
	switch scalars[fd.Desc.Kind()].TS {
	case "number":
		return a + " !== undefined && " + a + " !== 0"
	case "string":
		if is64Bit(fd.Desc.Kind()) {
			return a + ` !== undefined && ` + a + ` !== "0"`
		}
	}
	return a
}
