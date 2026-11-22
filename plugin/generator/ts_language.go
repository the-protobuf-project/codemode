// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "google.golang.org/protobuf/compiler/protogen"

// typescript emits the TypeScript codebase: interfaces with protojson and
// binary codecs, clients, Code Mode connectors, the catalog and sandbox.d.ts,
// against the runtime-ts packages. Its emitters are the ts_*.go files and its
// layouts templates/typescript/*.tpl.
type typescript struct{}

func (typescript) Name() string { return "typescript" }

func (typescript) Aliases() []string { return []string{"ts"} }

func (typescript) Generate(gen *protogen.Plugin, cat *Catalog) error {
	for _, p := range cat.Packages {
		for _, m := range p.Messages {
			cat.emitMessage(gen, m)
		}
		for _, e := range p.Enums {
			cat.emitEnum(gen, e)
		}
		for _, r := range p.Resources {
			cat.emitResource(gen, r)
		}
		for _, s := range p.Services {
			cat.emitService(gen, s)
		}
		cat.emitPackageIndex(gen, p)
	}
	cat.emitCatalog(gen)
	cat.emitConfig(gen)
	cat.emitCodemodeIndex(gen)
	return nil
}
