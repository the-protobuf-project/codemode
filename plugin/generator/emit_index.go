// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// emitPackageIndex writes <package>/index.ts: every type and service in the
// package, and registration of its codecs for google.protobuf.Any.
func (cat *Catalog) emitPackageIndex(gen *protogen.Plugin, p *Package) {
	f := cat.newFile(p.Dir+"/index.ts", "")
	f.notes = []string{"package: " + string(p.Name)}
	var files []string
	for _, m := range p.Messages {
		files = append(files, m.File)
	}
	for _, e := range p.Enums {
		files = append(files, e.File)
	}
	for _, r := range p.Resources {
		files = append(files, r.File)
	}
	for _, s := range p.Services {
		files = append(files, s.Dir+"/index.ts")
	}
	sort.Strings(files)
	for _, file := range files {
		f.P("export * from ", tsString(relImport(f.path, file, cat.Config.Extension())), ";")
	}
	if len(p.Messages) > 0 {
		codecs := make([]string, len(p.Messages))
		for i, m := range p.Messages {
			f.local(m.File, m.Codec())
			codecs[i] = m.Codec()
		}
		f.runtime("registerTypes")
		f.P()
		f.P("registerTypes([", strings.Join(codecs, ", "), "]);")
	}
	f.emit(gen)
}

// emitCodemodeIndex writes codemode/index.ts, which builds every connector
// from service implementations, and the root index.ts.
func (cat *Catalog) emitCodemodeIndex(gen *protogen.Plugin) {
	f := cat.newFile("codemode/index.ts", "")
	f.notes = []string{"Wires generated connectors into the Code Mode runtime."}
	f.runtimeType("Connector", "ConnectorOptions")
	ext := cat.Config.Extension()
	f.P("export { catalog } from ", tsString("./catalog"+ext), ";")
	f.P("export { config } from ", tsString("./config"+ext), ";")
	f.P()
	f.P("/** The implementation behind each sandbox namespace: a generated client or your own object. */")
	f.P("export interface Services {")
	for _, s := range cat.Services {
		f.localType(s.Dir+"/service.ts", s.Name)
		f.P("  ", s.Namespace, ": ", s.Name, ";")
	}
	f.P("}")
	f.P()
	f.P("/** Builds the connector for every namespace. */")
	f.P("export function connectors(services: Services, options: { [K in keyof Services]?: ConnectorOptions } = {}): Connector[] {")
	f.P("  return [")
	for _, s := range cat.Services {
		f.local(s.Dir+"/connector.ts", s.connectorFunc())
		f.P("    ", s.connectorFunc(), "(services.", s.Namespace, ", options.", s.Namespace, "),")
	}
	f.P("  ];")
	f.P("}")
	f.emit(gen)

	root := cat.newFile("index.ts", "")
	for _, p := range cat.Packages {
		alias := lowerFirstWord(pascal(string(p.Name)))
		root.P("export * as ", alias, " from ", tsString("./"+p.Dir+"/index"+ext), ";")
	}
	root.P("export * from ", tsString("./codemode/index"+ext), ";")
	root.emit(gen)
}
