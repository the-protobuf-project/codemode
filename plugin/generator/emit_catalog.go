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

// emitCatalog writes codemode/catalog.ts, the data search() and describe()
// read, and codemode/sandbox.d.ts, the same declarations as one file.
func (cat *Catalog) emitCatalog(gen *protogen.Plugin) {
	f := cat.newFile("codemode/catalog.ts", "")
	f.notes = []string{"The sandbox API: every namespace, method and type the model can use."}
	f.runtimeType("Catalog")
	f.P("export const catalog: Catalog = {")
	f.P("  prelude: ", tsTemplate(sandboxPrelude+"\n"+codemodeGlobal), ",")
	f.P("  types: {")
	types := cat.sandboxTypes()
	names := make([]string, 0, len(types))
	for k := range types {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		f.P("    ", propertyKey(k), ": ", tsTemplate(types[k]), ",")
	}
	f.P("  },")
	f.P("  namespaces: [")
	for _, s := range cat.Services {
		methods := make([]string, len(s.Methods))
		for i, m := range s.Methods {
			methods[i] = tsString(m.Path)
		}
		f.P("    {")
		f.P("      name: ", tsString(s.Namespace), ",")
		f.P("      service: ", tsString(string(s.Proto.Desc.FullName())), ",")
		f.P("      instructions: ", tsString(s.Instructions), ",")
		f.P("      declaration: ", tsTemplate(namespaceDecl(s)), ",")
		f.P("      methods: [", strings.Join(methods, ", "), "],")
		f.P("    },")
	}
	f.P("  ],")
	f.P("  methods: [")
	for _, s := range cat.Services {
		for _, m := range s.Methods {
			emitCatalogMethod(f, m)
		}
	}
	f.P("  ],")
	f.P("  snippets: [")
	for _, sn := range cat.Snippets {
		f.P("    {")
		f.P("      name: ", tsString(sn.Name), ",")
		f.P("      description: ", tsString(sn.Description), ",")
		f.P("      code: ", tsTemplate(sn.Code), ",")
		if sn.Input != nil {
			f.P("      inputType: ", tsString(sn.Input.Sandbox), ",")
		}
		f.P("    },")
	}
	f.P("  ],")
	f.P("};")
	f.emit(gen)

	banner := header.Render("//", header.Info{PluginVersion: PluginVersion, ProtocVersion: cat.Version,
		Notes: []string{"Everything code in the Code Mode sandbox can use."}})
	emitRaw(gen, "codemode/sandbox.d.ts", banner+"\n\n"+cat.sandboxDTS())
}

func emitCatalogMethod(f *tsFile, m *Method) {
	types := typeClosure(m)
	quoted := make([]string, len(types))
	for i, t := range types {
		quoted[i] = tsString(t)
	}
	examples := make([]string, len(m.Examples))
	for i, e := range m.Examples {
		examples[i] = tsTemplate(e)
	}
	f.P("    {")
	f.P("      path: ", tsString(m.Path), ",")
	f.P("      namespace: ", tsString(m.Service.Namespace), ",")
	f.P("      name: ", tsString(m.Name), ",")
	f.P("      rpc: ", tsString(string(m.Proto.Desc.FullName())), ",")
	f.P("      description: ", tsString(m.Doc), ",")
	f.P("      requiresApproval: ", m.Approval, ",")
	f.P("      replay: ", tsString(m.Replay), ",")
	f.P("      readOnly: ", m.ReadOnly, ",")
	if m.Revert != nil {
		f.P("      revertedBy: ", tsString(m.Service.Namespace+"."+m.Revert.Method.Name), ",")
	}
	f.P("      declaration: ", tsTemplate(methodDecl(m)), ",")
	f.P("      types: [", strings.Join(quoted, ", "), "],")
	f.P("      examples: [", strings.Join(examples, ", "), "],")
	f.P("    },")
}

// emitConfig writes codemode/config.ts from codemode.yaml's runtime settings.
func (cat *Catalog) emitConfig(gen *protogen.Plugin) {
	c := cat.Config
	timeout, _ := Millis(c.Executor.Timeout)
	ttl, _ := Millis(c.Runtime.PausedTTL)
	f := cat.newFile("codemode/config.ts", "")
	f.notes = []string{"Runtime settings from codemode.yaml."}
	f.runtimeType("CodemodeConfig")
	f.P("export const config: CodemodeConfig = {")
	f.P("  discovery: ", tsString(c.Discovery), ",")
	f.P("  executor: {")
	optNum(f, "timeoutMs", timeout)
	optNum(f, "memoryMb", int64(c.Executor.MemoryMB))
	f.P("    outbound: ", tsString(c.Executor.Outbound), ",")
	allow := make([]string, len(c.Executor.Allow))
	for i, a := range c.Executor.Allow {
		allow[i] = tsString(a)
	}
	f.P("    allow: [", strings.Join(allow, ", "), "],")
	f.P("  },")
	f.P("  runtime: {")
	optNum(f, "maxExecutions", int64(c.Runtime.MaxExecutions))
	optNum(f, "pausedTtlMs", ttl)
	optNum(f, "maxValueBytes", int64(c.Runtime.MaxValueBytes))
	optNum(f, "resultBudgetTokens", int64(c.Runtime.ResultBudgetTokens))
	f.P("  },")
	f.P("  mcp: { expose: ", tsString(c.MCP.Expose), " },")
	f.P("};")
	f.emit(gen)
}

func optNum(f *tsFile, key string, v int64) {
	if v > 0 {
		f.P(fmt.Sprintf("    %s: %d,", key, v))
	}
}
