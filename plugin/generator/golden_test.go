// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// exampleGen is the committed output the generator must reproduce exactly.
const exampleGen = "../../examples/library/gen"

// generateLibrary runs the plugin over testdata/library.binpb, built with
// `buf build examples/library/protobuf -o plugin/generator/testdata/library.binpb`.
func generateLibrary(t *testing.T, param string) *pluginpb.CodeGeneratorResponse {
	t.Helper()
	data, err := os.ReadFile("testdata/library.binpb")
	if err != nil {
		t.Fatal(err)
	}
	set := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, set); err != nil {
		t.Fatal(err)
	}
	req := &pluginpb.CodeGeneratorRequest{Parameter: proto.String(param), ProtoFile: set.GetFile()}
	for _, f := range set.GetFile() {
		if strings.HasPrefix(f.GetName(), "library/") {
			req.FileToGenerate = append(req.FileToGenerate, f.GetName())
		}
		if f.GetOptions().GetGoPackage() == "" {
			if f.Options == nil {
				f.Options = &descriptorpb.FileOptions{}
			}
			f.Options.GoPackage = proto.String("codemode.invalid/" + filepath.Dir(f.GetName()))
		}
	}
	var opts Options
	gen, err := protogen.Options{ParamFunc: func(name, value string) error {
		switch name {
		case "config":
			opts.Config = value
		case "runtime_prefix":
			opts.RuntimePrefix = value
		}
		return nil
	}}.New(req)
	if err != nil {
		t.Fatal(err)
	}
	PluginVersion = "v0.1.0"
	if err := Generate(gen, opts); err != nil {
		t.Fatal(err)
	}
	res := gen.Response()
	if res.GetError() != "" {
		t.Fatal(res.GetError())
	}
	return res
}

func TestGoldenLibrary(t *testing.T) {
	res := generateLibrary(t, "config=../../examples/library/codemode.yaml,runtime_prefix=@the-protobuf-project/runtime-ts-")
	seen := map[string]bool{}
	for _, f := range res.GetFile() {
		seen[f.GetName()] = true
		want, err := os.ReadFile(filepath.Join(exampleGen, f.GetName()))
		if err != nil {
			t.Errorf("%s: generated but not committed (run `just generate`)", f.GetName())
			continue
		}
		if string(want) != f.GetContent() {
			t.Errorf("%s: differs from the committed example (run `just generate`)\n%s", f.GetName(), firstDiff(string(want), f.GetContent()))
		}
	}
	err := filepath.WalkDir(exampleGen, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(exampleGen, path)
		if !seen[filepath.ToSlash(rel)] {
			t.Errorf("%s: committed but no longer generated", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestImportExtensionOption(t *testing.T) {
	ext := ".js"
	cfg := &Config{ImportExtension: &ext}
	if got := relImport("library/v1/messages/book.ts", "library/v1/enums/book_state.ts", cfg.Extension()); got != "../enums/book_state.js" {
		t.Errorf("relImport = %q", got)
	}
	if got := relImport("index.ts", "library/v1/index.ts", ""); got != "./library/v1/index" {
		t.Errorf("relImport = %q", got)
	}
}

// firstDiff shows the first line where two texts differ.
func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return fmt.Sprintf("line %d:\n  committed: %q\n  generated: %q", i+1, wl, gl)
		}
	}
	return ""
}
