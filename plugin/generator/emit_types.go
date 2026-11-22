// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// enumDecl renders an enum as a union of its value names.
func enumDecl(e *Enum, name, export string) string {
	var b strings.Builder
	b.WriteString(jsdoc("", e.Doc))
	b.WriteString(export + "type " + name + " =\n")
	for i, v := range e.Values {
		b.WriteString(jsdoc("  ", v.Doc))
		end := ""
		if i == len(e.Values)-1 {
			end = ";"
		}
		b.WriteString("  | " + tsString(v.Name) + end + "\n")
	}
	return b.String()
}

// emitEnum writes enums/<name>.ts: the union type and its codec.
func (cat *Catalog) emitEnum(gen *protogen.Plugin, e *Enum) {
	f := cat.newFile(e.File, e.Source)
	f.runtime("enumCodec")
	f.runtimeType("EnumCodec")
	f.body.WriteString(enumDecl(e, e.Name, "export "))
	f.P()
	f.P("/** Converts {@link ", e.Name, "} between value names and numbers. */")
	f.P("export const ", e.Codec(), ": EnumCodec<", e.Name, "> = enumCodec(", tsString(string(e.Desc.FullName())), ", [")
	for _, v := range e.Values {
		f.P("  [", tsString(v.Name), ", ", v.Number, "],")
	}
	f.P("]);")
	f.emit(gen)
}

// resourceDecl renders a resource name as a template literal type.
func resourceDecl(r *Resource, name, export string) string {
	return jsdoc("", fmt.Sprintf("Resource name of `%s`: `%s`.", r.Type, r.Pattern)) +
		export + "type " + name + " = `" + r.Template + "`;\n"
}

// emitResource writes resources/<name>.ts: the name type, and a value of the
// same name that formats and parses it.
func (cat *Catalog) emitResource(gen *protogen.Plugin, r *Resource) {
	f := cat.newFile(r.File, "")
	f.notes = []string{"resource: " + r.Type}
	f.runtime("resourceName")
	f.runtimeType("ResourceName")
	f.body.WriteString(resourceDecl(r, r.Name, "export "))
	f.P()
	vars := make([]string, len(r.Vars))
	for i, v := range r.Vars {
		vars[i] = tsString(v)
	}
	varType := "never"
	if len(vars) > 0 {
		varType = strings.Join(vars, " | ")
	}
	f.P("/** Formats, parses and checks {@link ", r.Name, "} values. */")
	f.P("export const ", r.Name, ": ResourceName<", varType, ", ", r.Name, "> = resourceName(", tsString(r.Type), ", ", tsString(r.Pattern), ");")
	f.emit(gen)
}
