// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"sort"

	"github.com/the-protobuf-project/codemode/protobuf/codemodepb"
	"github.com/the-protobuf-project/mcp/protobuf/mcppb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ServiceOverride sets codemode.v1.service options from codemode.yaml, for
// protos the author cannot annotate. Set fields win over the annotation.
type ServiceOverride struct {
	Namespace    string `yaml:"namespace"`
	Instructions string `yaml:"instructions"`
	Exclude      bool   `yaml:"exclude"`
}

// MethodOverride sets codemode.v1.method options from codemode.yaml.
type MethodOverride struct {
	Exclude  bool            `yaml:"exclude"`
	Approval string          `yaml:"approval"` // required | none
	Replay   string          `yaml:"replay"`   // log | reexecute
	Revert   *RevertOverride `yaml:"revert"`
	Examples []string        `yaml:"examples"`
}

// RevertOverride is codemode.v1.Revert: the method that undoes this one, and
// its request fields from "args.<path>" or "result.<path>".
type RevertOverride struct {
	Method string            `yaml:"method"`
	Fields map[string]string `yaml:"fields"`
}

var (
	approvals = map[string]codemodepb.Approval{"required": codemodepb.Approval_APPROVAL_REQUIRED, "none": codemodepb.Approval_APPROVAL_NONE}
	replays   = map[string]codemodepb.Replay{"log": codemodepb.Replay_REPLAY_LOG, "reexecute": codemodepb.Replay_REPLAY_REEXECUTE}
)

// serviceOptions is the service's codemode.v1 annotation with codemode.yaml
// applied, and its mcp.v1 options.
func (cat *Catalog) serviceOptions(sd protoreflect.ServiceDescriptor) (*codemodepb.CodemodeService, *mcppb.McpServiceOptions) {
	opts, mcpOpts := serviceOptions(sd)
	o, ok := cat.Config.Services[string(sd.FullName())]
	if !ok {
		return opts, mcpOpts
	}
	out := &codemodepb.CodemodeService{}
	if opts != nil {
		out = proto.Clone(opts).(*codemodepb.CodemodeService)
	}
	if o.Namespace != "" {
		out.Namespace = proto.String(o.Namespace)
	}
	if o.Instructions != "" {
		out.Instructions = proto.String(o.Instructions)
	}
	out.Exclude = out.GetExclude() || o.Exclude
	return out, mcpOpts
}

// methodOptions is the method's codemode.v1 annotation with codemode.yaml
// applied, and its mcp.v1 options.
func (cat *Catalog) methodOptions(md protoreflect.MethodDescriptor) (*codemodepb.CodemodeMethod, *mcppb.McpToolOptions, *mcppb.McpElicitation) {
	opts, tool, elicit := methodOptions(md)
	o, ok := cat.Config.Methods[string(md.FullName())]
	if !ok {
		return opts, tool, elicit
	}
	out := &codemodepb.CodemodeMethod{}
	if opts != nil {
		out = proto.Clone(opts).(*codemodepb.CodemodeMethod)
	}
	out.Exclude = out.GetExclude() || o.Exclude
	if o.Approval != "" {
		out.Approval = approvals[o.Approval]
	}
	if o.Replay != "" {
		out.Replay = replays[o.Replay]
	}
	if o.Revert != nil {
		out.Revert = &codemodepb.Revert{Method: o.Revert.Method, Fields: o.Revert.Fields}
	}
	out.Examples = append(out.Examples, o.Examples...)
	return out, tool, elicit
}

// checkOverrides rejects codemode.yaml entries that name no service or method
// in the request, or use unknown values, so a typo cannot silently drop an
// approval.
func (c *Config) checkOverrides(files []*protogen.File) error {
	known := map[string]bool{}
	for _, f := range files {
		for _, s := range f.Services {
			known[string(s.Desc.FullName())] = true
			for _, m := range s.Methods {
				known[string(m.Desc.FullName())] = true
			}
		}
	}
	var errs []string
	for name := range c.Services {
		if !known[name] {
			errs = append(errs, fmt.Sprintf("services: %s is not a service in the request", name))
		}
	}
	for name, o := range c.Methods {
		if !known[name] {
			errs = append(errs, fmt.Sprintf("methods: %s is not a method in the request", name))
		}
		if _, ok := approvals[o.Approval]; o.Approval != "" && !ok {
			errs = append(errs, fmt.Sprintf("methods: %s: approval %q is not required or none", name, o.Approval))
		}
		if _, ok := replays[o.Replay]; o.Replay != "" && !ok {
			errs = append(errs, fmt.Sprintf("methods: %s: replay %q is not log or reexecute", name, o.Replay))
		}
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("codemode.yaml: %s", errs[0])
	}
	return nil
}
