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
	Lang            string  // comma-separated languages, or "all"; default "typescript"
	Config          string  // path to codemode.yaml
	ImportExtension *string // TypeScript: extension on relative imports
	RuntimePrefix   string  // TypeScript: prefix of the runtime package names
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
		cfg.TypeScript.ImportExtension = s.Options.ImportExtension
	}
	if s.Options.RuntimePrefix != "" {
		cfg.TypeScript.RuntimePrefix = s.Options.RuntimePrefix
	}
	return Build(ctx.Plugin, cfg)
}

// Target renders the catalog in one of the supported languages.
type Target struct{}

// Name identifies the target in a factory registry.
func (Target) Name() string { return "codemode" }

// Languages lists the languages the target emits.
func (Target) Languages() []string { return languageNames() }

// Generate writes one language's codebase.
func (Target) Generate(ctx factory.Ctx, cat *Catalog, lang string) error {
	l := findLanguage(lang)
	if l == nil {
		return fmt.Errorf("unsupported language %q", lang)
	}
	return l.Generate(ctx.Plugin, cat)
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
	langs, err := resolveLanguages(opts.Lang)
	if err != nil {
		return err
	}
	cat, err := reg.Sources["proto"].Build(ctx)
	if err != nil {
		return err
	}
	for _, l := range langs {
		if err := reg.Targets["codemode"].Generate(ctx, cat, l.Name()); err != nil {
			return fmt.Errorf("%s: %w", l.Name(), err)
		}
	}
	return nil
}
