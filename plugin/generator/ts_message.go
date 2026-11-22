// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// interfaceView is the "interface" template's data.
type interfaceView struct {
	Doc, Export, Name string
	Fields            []propView
}

// propView is one interface property.
type propView struct {
	Doc, Key, Opt, Type string
}

// keyExpr is one property in an object literal.
type keyExpr struct {
	Key, Expr string
}

// messageView is message.ts.tpl's data: the interface and every statement of
// its codec, resolved for the file.
type messageView struct {
	Decl                  interfaceView
	Name, Codec, TypeName string
	Empty, NeedsAll       bool
	FromJSON, Init        []keyExpr
	Oneofs                []struct{ Name, Fields string }
	ToJSON, Encode        []string
	Cases                 []decodeCase
}

func newInterface(m *Message, n typeNamer, export string) interfaceView {
	v := interfaceView{Doc: jsdoc("", m.Doc, messageNotes(m)...), Export: export, Name: n.message(m)}
	for _, f := range m.Fields {
		opt := "?"
		if f.Required {
			opt = ""
		}
		v.Fields = append(v.Fields, propView{jsdoc("  ", f.Doc, fieldNotes(f)...), propertyKey(f.JSONName), opt, tsType(f, n)})
	}
	return v
}

// interfaceDecl renders a message as an interface. The sandbox uses the
// same template with sandbox names and no export.
func interfaceDecl(m *Message, n typeNamer, export string) string {
	return renderString("interface", newInterface(m, n, export))
}

func messageNotes(m *Message) []string {
	var notes []string
	if m.Resource != nil {
		notes = append(notes, "Resource `"+m.Resource.Type+"`, named `"+m.Resource.Pattern+"`.")
	}
	return append(notes, deprecated(isDeprecated(m.Desc))...)
}

// emitMessage writes messages/<name>.ts: the interface and its codec.
func (cat *Catalog) emitMessage(gen *protogen.Plugin, m *Message) {
	f := cat.newFile(m.Path, m.Source)
	n := moduleNamer{f}
	f.runtime("protojson", "BinaryWriter", "BinaryReader")
	f.runtimeType("MessageCodec", "JsonObject")
	v := messageView{
		Decl: newInterface(m, n, "export "), Name: m.Name, Codec: m.Codec(),
		TypeName: string(m.Desc.FullName()), Empty: len(m.Fields) == 0, NeedsAll: usesDefaults(m),
	}
	if !v.Empty {
		f.runtimeType("ToJsonOptions")
		v.FromJSON, v.Oneofs = fromJSONFields(f, m, n), oneofChecks(m)
		v.ToJSON = toJSONLines(f, m)
		v.Encode = encodeLines(f, m)
	}
	v.Init, v.Cases = decodeInit(f, m, n), decodeCases(f, m, n)
	f.render("message.ts.tpl", v)
	f.emit(gen)
}

// codecOf returns the codec expression for a message, enum or well-known
// element, importing it.
func codecOf(f *tsFile, fd *Field) string {
	switch {
	case fd.Message != nil:
		f.local(fd.Message.Path, fd.Message.Codec())
		return fd.Message.Codec()
	case fd.Enum != nil:
		f.local(fd.Enum.Path, fd.Enum.Codec())
		return fd.Enum.Codec()
	case fd.WKT != "":
		f.runtime(tsWKT(fd.WKT).Codec)
		return tsWKT(fd.WKT).Codec
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
	case fd.WKT != "" && tsWKT(fd.WKT).Codec == "ValueCodec":
		return a + " !== undefined"
	case fd.Message != nil || fd.WKT != "":
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
