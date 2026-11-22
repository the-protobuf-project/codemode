// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Command protoc-gen-codemode generates a Code Mode TypeScript codebase from
// annotated protobuf services.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/the-protobuf-project/codemode/plugin/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

// version is set at build time with -ldflags "-X main.version=v0.1.0".
var version = ""

func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("protoc-gen-codemode %s\n", resolveVersion())
		return
	}
	var flags flag.FlagSet
	config := flags.String("config", "", "Path to codemode.yaml, relative to where buf runs.")
	ext := flags.String("import_extension", ".ts", `Extension on relative imports: ".ts", ".js" or "none".`)
	runtimePrefix := flags.String("runtime_prefix", "", `Prefix of the runtime package names; code imports <prefix>protobuf, <prefix>network and <prefix>codemode.`)

	generator.PluginVersion = resolveVersion()

	generator.Run(protogen.Options{ParamFunc: flags.Set}, func(gen *protogen.Plugin) error {
		opts := generator.Options{Config: *config, RuntimePrefix: *runtimePrefix}
		set := false
		flags.Visit(func(f *flag.Flag) { set = set || f.Name == "import_extension" })
		if set {
			e := *ext
			if e == "none" {
				e = ""
			}
			opts.ImportExtension = &e
		}
		return generator.Generate(gen, opts)
	})
}
