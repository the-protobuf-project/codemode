// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"sort"

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
	exports := make([]string, len(files))
	for i, file := range files {
		exports[i] = relImport(f.path, file, cat.Config.Extension())
	}
	var codecs []string
	for _, m := range p.Messages {
		f.local(m.File, m.Codec())
		codecs = append(codecs, m.Codec())
	}
	if len(codecs) > 0 {
		f.runtime("registerTypes")
	}
	f.render("package_index.ts.tpl", struct{ Exports, Codecs []string }{exports, codecs})
	f.emit(gen)
}

// serviceRef is one namespace in codemode/index.ts.
type serviceRef struct {
	Namespace, Name, Connector string
}

// packageRef is one package re-exported by the root index.ts.
type packageRef struct {
	Alias, Path string
}

// emitCodemodeIndex writes codemode/index.ts, which builds every connector
// from service implementations, and the root index.ts.
func (cat *Catalog) emitCodemodeIndex(gen *protogen.Plugin) {
	f := cat.newFile("codemode/index.ts", "")
	f.notes = []string{"Wires generated connectors into the Code Mode runtime."}
	f.runtimeType("Connector", "ConnectorOptions")
	ext := cat.Config.Extension()
	var services []serviceRef
	for _, s := range cat.Services {
		f.localType(s.Dir+"/service.ts", s.Name)
		f.local(s.Dir+"/connector.ts", s.connectorFunc())
		services = append(services, serviceRef{s.Namespace, s.Name, s.connectorFunc()})
	}
	f.render("codemode_index.ts.tpl", struct {
		Catalog, Config string
		Services        []serviceRef
	}{"./catalog" + ext, "./config" + ext, services})
	f.emit(gen)

	root := cat.newFile("index.ts", "")
	var packages []packageRef
	for _, p := range cat.Packages {
		packages = append(packages, packageRef{lowerFirstWord(pascal(string(p.Name))), "./" + p.Dir + "/index" + ext})
	}
	root.render("root_index.ts.tpl", struct {
		Packages []packageRef
		Codemode string
	}{packages, "./codemode/index" + ext})
	root.emit(gen)
}
