// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/the-protobuf-project/protokit/header"
	"google.golang.org/protobuf/compiler/protogen"
)

// PluginVersion is stamped into every generated header; main sets it.
var PluginVersion = "dev"

// tsFile accumulates one generated TypeScript file: its imports, kept apart
// from the body so they can be sorted and merged, and the body itself.
type tsFile struct {
	cat     *Catalog
	path    string
	source  string
	notes   []string
	imports map[string]map[string]bool // specifier -> name -> type-only
	body    strings.Builder
}

func (cat *Catalog) newFile(path, source string) *tsFile {
	return &tsFile{cat: cat, path: path, source: source, imports: map[string]map[string]bool{}}
}

func (f *tsFile) add(spec, name string, typeOnly bool) {
	set := f.imports[spec]
	if set == nil {
		set = map[string]bool{}
		f.imports[spec] = set
	}
	if prev, seen := set[name]; seen {
		set[name] = prev && typeOnly
		return
	}
	set[name] = typeOnly
}

// runtimeModules maps runtime exports that do not live in the protobuf
// package to the runtime-ts package that has them. Everything else is in
// <prefix>protobuf.
var runtimeModules = map[string]string{
	"CallOptions": "network", "MethodDescriptor": "network", "Transport": "network",
	"connector": "codemode", "unaryMethod": "codemode", "Connector": "codemode",
	"ConnectorOptions": "codemode", "Catalog": "codemode", "CodemodeConfig": "codemode",
}

// runtimeSpec is the import specifier of the runtime package exporting name.
func (f *tsFile) runtimeSpec(name string) string {
	mod, ok := runtimeModules[name]
	if !ok {
		mod = "protobuf"
	}
	return f.cat.Config.RuntimePrefix + mod
}

// runtime imports values from the runtime package.
func (f *tsFile) runtime(names ...string) {
	for _, n := range names {
		f.add(f.runtimeSpec(n), n, false)
	}
}

// runtimeType imports types from the runtime package.
func (f *tsFile) runtimeType(names ...string) {
	for _, n := range names {
		f.add(f.runtimeSpec(n), n, true)
	}
}

// local imports values from another generated file; a no-op for this file.
func (f *tsFile) local(target string, names ...string) {
	f.importLocal(target, false, names)
}

// localType imports types from another generated file.
func (f *tsFile) localType(target string, names ...string) {
	f.importLocal(target, true, names)
}

func (f *tsFile) importLocal(target string, typeOnly bool, names []string) {
	if target == f.path {
		return
	}
	spec := relImport(f.path, target, f.cat.Config.Extension())
	for _, n := range names {
		f.add(spec, n, typeOnly)
	}
}

// content renders the header, the sorted imports and the body.
func (f *tsFile) content() string {
	var b strings.Builder
	b.WriteString(header.Render("//", header.Info{
		PluginVersion: PluginVersion,
		ProtocVersion: f.cat.Version,
		Source:        f.source,
		Notes:         f.notes,
	}))
	b.WriteString("\n\n")
	specs := make([]string, 0, len(f.imports))
	for s := range f.imports {
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool {
		li, lj := strings.HasPrefix(specs[i], "."), strings.HasPrefix(specs[j], ".")
		if li != lj {
			return !li
		}
		return specs[i] < specs[j]
	})
	for _, s := range specs {
		b.WriteString(importLine(s, f.imports[s]))
	}
	if len(specs) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(strings.TrimRight(f.body.String(), "\n"))
	b.WriteByte('\n')
	return b.String()
}

func importLine(spec string, names map[string]bool) string {
	list := make([]string, 0, len(names))
	allTypes := true
	for n, typeOnly := range names {
		list = append(list, n)
		allTypes = allTypes && typeOnly
	}
	sort.Strings(list)
	keyword := "import"
	if allTypes {
		keyword = "import type"
	} else {
		for i, n := range list {
			if names[n] {
				list[i] = "type " + n
			}
		}
	}
	line := fmt.Sprintf("%s { %s } from %q;\n", keyword, strings.Join(list, ", "), spec)
	if len(line) <= maxLine {
		return line
	}
	return fmt.Sprintf("%s {\n  %s,\n} from %q;\n", keyword, strings.Join(list, ",\n  "), spec)
}

// maxLine is the width past which an import is split one name per line.
const maxLine = 100

// emit hands the finished file to protoc.
func (f *tsFile) emit(gen *protogen.Plugin) {
	g := gen.NewGeneratedFile(f.path, "")
	g.P(f.content())
}

// raw is a file with no imports or header logic: JSON and declaration text.
func emitRaw(gen *protogen.Plugin, path, content string) {
	g := gen.NewGeneratedFile(path, "")
	g.P(strings.TrimRight(content, "\n"))
}
