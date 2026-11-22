// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"
	"text/template"

	"github.com/the-protobuf-project/codemode/plugin/generator/templates"
)

// funcs are the helpers templates may call. They only format values; every
// name, type and expression is computed before rendering.
var funcs = template.FuncMap{
	"str":  tsString,
	"tpl":  tsTemplate,
	"join": strings.Join,
}

// tpls holds every template, parsed once. protokit's templates.MustParse is
// not used because it takes no FuncMap, and text/template rejects a reference
// to an unknown function at parse time.
var tpls = template.Must(template.New("typescript").Funcs(funcs).ParseFS(templates.FS, "typescript/*.tpl"))

// renderString executes the named template and returns its output. A failure
// is a bug in a template or view, so it panics rather than emitting a partial
// file.
func renderString(name string, data any) string {
	var b strings.Builder
	if err := tpls.ExecuteTemplate(&b, name, data); err != nil {
		panic("generator: render " + name + ": " + err.Error())
	}
	return b.String()
}

// render appends the named template's output to the file body.
func (f *tsFile) render(name string, data any) {
	f.body.WriteString(renderString(name, data))
}
