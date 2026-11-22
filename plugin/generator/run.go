// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Run is protogen.Options.Run for a plugin that emits no Go. protogen insists
// every file has a Go import path; TypeScript output has no use for one, so
// files without go_package get a placeholder before protogen sees them.
func Run(opts protogen.Options, f func(*protogen.Plugin) error) {
	if err := run(opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func run(opts protogen.Options, f func(*protogen.Plugin) error) error {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(in, req); err != nil {
		return err
	}
	for _, fd := range req.GetProtoFile() {
		if fd.GetOptions().GetGoPackage() != "" {
			continue
		}
		if fd.Options == nil {
			fd.Options = &descriptorpb.FileOptions{}
		}
		dir := fd.GetName()[:max(strings.LastIndex(fd.GetName(), "/"), 0)]
		fd.Options.GoPackage = proto.String("codemode.invalid/" + dir)
	}
	gen, err := opts.New(req)
	if err != nil {
		return err
	}
	if err := f(gen); err != nil {
		gen.Error(err)
	}
	out, err := proto.Marshal(gen.Response())
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
