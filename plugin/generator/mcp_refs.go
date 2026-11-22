// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"sort"
	"strings"

	mcp "github.com/the-protobuf-project/mcp/plugin/generator"
)

// rewriteToolRefs replaces protoc-gen-mcp tool names in documentation, such
// as "hold_service-create_hold_v1", with the sandbox path the model calls,
// "holds.createHold". mcp.v1 descriptions name other tools that way, and in
// Code Mode those names do not exist.
func (cat *Catalog) rewriteToolRefs() {
	var pairs []string
	for _, s := range cat.Services {
		for _, m := range s.Methods {
			pairs = append(pairs, mcp.BuildToolName(string(m.Proto.Desc.FullName())), m.Path)
		}
	}
	if len(pairs) == 0 {
		return
	}
	// Longest names first, so no name is replaced inside a longer one.
	idx := make([]int, len(pairs)/2)
	for i := range idx {
		idx[i] = i * 2
	}
	sort.Slice(idx, func(a, b int) bool { return len(pairs[idx[a]]) > len(pairs[idx[b]]) })
	ordered := make([]string, 0, len(pairs))
	for _, i := range idx {
		ordered = append(ordered, pairs[i], pairs[i+1])
	}
	r := strings.NewReplacer(ordered...)
	for _, s := range cat.Services {
		s.Doc, s.Instructions = r.Replace(s.Doc), r.Replace(s.Instructions)
		for _, m := range s.Methods {
			m.Doc = r.Replace(m.Doc)
		}
	}
	for _, p := range cat.Packages {
		for _, m := range p.Messages {
			m.Doc = r.Replace(m.Doc)
			for _, f := range m.Fields {
				f.Doc = r.Replace(f.Doc)
			}
		}
	}
}
