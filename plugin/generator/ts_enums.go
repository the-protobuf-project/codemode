// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// enumDeclView is the "enum" template's data: a string union.
type enumDeclView struct {
	Doc, Export, Name string
	Values            []enumValueView
}

type enumValueView struct {
	Doc, Name string
	Number    int32
	Last      bool
}

func newEnumDecl(e *Enum, name, export string) enumDeclView {
	v := enumDeclView{Doc: jsdoc("", e.Doc), Export: export, Name: name}
	for i, val := range e.Values {
		v.Values = append(v.Values, enumValueView{
			Doc: jsdoc("  ", val.Doc), Name: val.Name, Number: val.Number, Last: i == len(e.Values)-1,
		})
	}
	return v
}

// enumDecl renders an enum as a union of its value names.
func enumDecl(e *Enum, name, export string) string {
	return renderString("enum", newEnumDecl(e, name, export))
}

// emitEnum writes enums/<name>.ts: the union type and its codec.
func (cat *Catalog) emitEnum(gen *protogen.Plugin, e *Enum) {
	f := cat.newFile(e.Path, e.Source)
	f.runtime("enumCodec")
	f.runtimeType("EnumCodec")
	f.render("enum.ts.tpl", struct {
		Decl            enumDeclView
		Codec, TypeName string
	}{newEnumDecl(e, e.Name, "export "), e.Codec(), string(e.Desc.FullName())})
	f.emit(gen)
}

// resourceDeclView is the "resource" template's data: a template literal type.
type resourceDeclView struct {
	Doc, Export, Name, Template string
}

func newResourceDecl(r *Resource, name, export string) resourceDeclView {
	doc := jsdoc("", fmt.Sprintf("Resource name of `%s`: `%s`.", r.Type, r.Pattern))
	return resourceDeclView{Doc: doc, Export: export, Name: name, Template: r.Template}
}

// resourceDecl renders a resource name as a template literal type.
func resourceDecl(r *Resource, name, export string) string {
	return renderString("resource", newResourceDecl(r, name, export))
}

// emitResource writes resources/<name>.ts: the name type, and a value of the
// same name that formats and parses it.
func (cat *Catalog) emitResource(gen *protogen.Plugin, r *Resource) {
	f := cat.newFile(r.Path, "")
	f.notes = []string{"resource: " + r.Type}
	f.runtime("resourceName")
	f.runtimeType("ResourceName")
	vars := make([]string, len(r.Vars))
	for i, v := range r.Vars {
		vars[i] = tsString(v)
	}
	varType := "never"
	if len(vars) > 0 {
		varType = strings.Join(vars, " | ")
	}
	f.render("resource.ts.tpl", struct {
		Decl                resourceDeclView
		Vars, Type, Pattern string
	}{newResourceDecl(r, r.Name, "export "), varType, r.Type, r.Pattern})
	f.emit(gen)
}
