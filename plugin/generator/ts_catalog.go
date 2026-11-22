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

// catalogNamespace is one namespace in catalog.ts.
type catalogNamespace struct {
	Name, Service, Instructions, Decl, Methods string
}

// catalogMethod is one method in catalog.ts.
type catalogMethod struct {
	Path, Namespace, Name, RPC, Description, Replay, RevertedBy string
	Approval, ReadOnly                                          bool
	Decl, Types, Examples                                       string
}

// catalogSnippet is one snippet in catalog.ts.
type catalogSnippet struct {
	Name, Description, Code, InputType string
}

func quoteAll(vals []string, quote func(string) string) string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = quote(v)
	}
	return strings.Join(out, ", ")
}

// emitCatalog writes codemode/catalog.ts, the data search() and describe()
// read, and codemode/sandbox.d.ts, the same declarations as one file.
func (cat *Catalog) emitCatalog(gen *protogen.Plugin) {
	f := cat.newFile("codemode/catalog", "")
	f.notes = []string{"The sandbox API: every namespace, method and type the model can use."}
	f.runtimeType("Catalog")
	var v struct {
		Prelude    string
		Types      []keyExpr
		Namespaces []catalogNamespace
		Methods    []catalogMethod
		Snippets   []catalogSnippet
	}
	v.Prelude = renderString("sandbox-prelude", nil)
	types := cat.sandboxTypes()
	names := make([]string, 0, len(types))
	for k := range types {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		v.Types = append(v.Types, keyExpr{propertyKey(k), types[k]})
	}
	for _, s := range cat.Services {
		paths := make([]string, len(s.Methods))
		for i, m := range s.Methods {
			paths[i] = m.Path
		}
		v.Namespaces = append(v.Namespaces, catalogNamespace{
			s.Namespace, string(s.Proto.Desc.FullName()), s.Instructions, namespaceDecl(s), quoteAll(paths, tsString),
		})
		for _, m := range s.Methods {
			cm := catalogMethod{
				Path: m.Path, Namespace: s.Namespace, Name: m.Name, RPC: string(m.Proto.Desc.FullName()),
				Description: m.Doc, Replay: m.Replay, Approval: m.Approval, ReadOnly: m.ReadOnly,
				Decl: methodDecl(m), Types: quoteAll(typeClosure(m), tsString), Examples: quoteAll(m.Examples, tsTemplate),
			}
			if m.Revert != nil {
				cm.RevertedBy = s.Namespace + "." + m.Revert.Method.Name
			}
			v.Methods = append(v.Methods, cm)
		}
	}
	for _, sn := range cat.Snippets {
		cs := catalogSnippet{Name: sn.Name, Description: sn.Description, Code: sn.Code}
		if sn.Input != nil {
			cs.InputType = sn.Input.Sandbox
		}
		v.Snippets = append(v.Snippets, cs)
	}
	f.render("catalog.ts.tpl", v)
	f.emit(gen)

	banner := header.Render("//", header.Info{PluginVersion: PluginVersion, ProtocVersion: cat.Version,
		Notes: []string{"Everything code in the Code Mode sandbox can use."}})
	var namespaces []sandboxNamespace
	for _, s := range cat.Services {
		namespaces = append(namespaces, newSandboxNamespace(s))
	}
	emitRaw(gen, "codemode/sandbox.d.ts", renderString("sandbox.d.ts.tpl", struct {
		Banner     string
		Types      []string
		Namespaces []sandboxNamespace
	}{banner, cat.sandboxTypesUsed(), namespaces}))
}

// emitConfig writes codemode/config.ts from codemode.yaml's runtime settings.
func (cat *Catalog) emitConfig(gen *protogen.Plugin) {
	c := cat.Config
	timeout, _ := Millis(c.Executor.Timeout)
	ttl, _ := Millis(c.Runtime.PausedTTL)
	f := cat.newFile("codemode/config", "")
	f.notes = []string{"Runtime settings from codemode.yaml."}
	f.runtimeType("CodemodeConfig")
	f.render("config.ts.tpl", struct {
		Discovery, Outbound, Allow, Expose string
		Executor, Runtime                  []keyExpr
	}{
		Discovery: c.Discovery, Outbound: c.Executor.Outbound, Expose: c.MCP.Expose,
		Allow:    quoteAll(c.Executor.Allow, tsString),
		Executor: positive("timeoutMs", timeout, "memoryMb", int64(c.Executor.MemoryMB)),
		Runtime: positive("maxExecutions", int64(c.Runtime.MaxExecutions), "pausedTtlMs", ttl,
			"maxValueBytes", int64(c.Runtime.MaxValueBytes), "resultBudgetTokens", int64(c.Runtime.ResultBudgetTokens)),
	})
	f.emit(gen)
}

// positive keeps the key/value pairs whose value is set (above zero).
func positive(pairs ...any) []keyExpr {
	var out []keyExpr
	for i := 0; i+1 < len(pairs); i += 2 {
		if v, _ := pairs[i+1].(int64); v > 0 {
			out = append(out, keyExpr{pairs[i].(string), fmt.Sprint(v)})
		}
	}
	return out
}
