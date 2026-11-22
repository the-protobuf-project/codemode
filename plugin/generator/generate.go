// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Package generator turns annotated protobuf services into a Code Mode
// TypeScript codebase: typed messages with protojson and binary codecs,
// clients, connectors, and the catalog the sandbox runtime reads.
package generator

import (
	"fmt"

	"github.com/the-protobuf-project/protokit/factory"
	"github.com/the-protobuf-project/protokit/header"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Options are the plugin parameters from buf.gen.yaml.
type Options struct {
	Config          string // path to codemode.yaml
	ImportExtension *string
	RuntimeScope    string // npm scope of the runtime packages
}

// ProtoSource builds the catalog from the CodeGeneratorRequest.
type ProtoSource struct{ Options Options }

// Name identifies the source in a factory registry.
func (ProtoSource) Name() string { return "proto" }

// Build loads codemode.yaml and builds the catalog.
func (s ProtoSource) Build(ctx factory.Ctx) (*Catalog, error) {
	if ctx.Plugin == nil {
		return nil, fmt.Errorf("proto source requires plugin (protoc) mode")
	}
	cfg, err := LoadConfig(s.Options.Config)
	if err != nil {
		return nil, err
	}
	if s.Options.ImportExtension != nil {
		cfg.ImportExtension = s.Options.ImportExtension
	}
	if s.Options.RuntimeScope != "" {
		cfg.RuntimeScope = s.Options.RuntimeScope
	}
	return Build(ctx.Plugin, cfg)
}

// Target renders the catalog as TypeScript.
type Target struct{}

// Name identifies the target in a factory registry.
func (Target) Name() string { return "codemode" }

// Languages lists the languages the target emits.
func (Target) Languages() []string { return []string{"typescript"} }

// Generate writes every file of the generated codebase.
func (Target) Generate(ctx factory.Ctx, cat *Catalog, _ string) error {
	gen := ctx.Plugin
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

// Generate runs the plugin over one CodeGeneratorRequest.
func Generate(gen *protogen.Plugin, opts Options) error {
	header.SetTool("protoc-gen-codemode")
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL |
		pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	reg := factory.NewRegistry[*Catalog]()
	reg.AddSource(ProtoSource{Options: opts})
	reg.AddTarget(Target{})
	ctx := factory.Ctx{Plugin: gen}
	cat, err := reg.Sources["proto"].Build(ctx)
	if err != nil {
		return err
	}
	return reg.Targets["codemode"].Generate(ctx, cat, "typescript")
}
