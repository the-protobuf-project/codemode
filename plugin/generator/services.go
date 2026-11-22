// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/the-protobuf-project/codemode/protobuf/codemodepb"
	"github.com/the-protobuf-project/mcp/protobuf/mcppb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// addService adds one service and its unary methods, unless config or
// annotations exclude it.
func (cat *Catalog) addService(s *protogen.Service, f *protogen.File) error {
	full := string(s.Desc.FullName())
	opts, mcpOpts := serviceOptions(s.Desc)
	if !cat.Config.Selects(full) || opts.GetExclude() {
		return nil
	}
	p := cat.pkg(f.Desc.Package())
	name := string(s.Desc.Name())
	svc := &Service{
		Proto:     s,
		Package:   p,
		Name:      name,
		Dir:       p.Dir + "/services/" + fileStem(name),
		Source:    f.Desc.Path(),
		Namespace: opts.GetNamespace(),
		Doc:       comment(s.Comments),
	}
	if svc.Namespace == "" {
		svc.Namespace = namespaceName(name)
	}
	if !token.IsIdentifier(svc.Namespace) || reserved[svc.Namespace] {
		return fmt.Errorf("%s: namespace %q is not a usable JavaScript identifier", full, svc.Namespace)
	}
	for _, other := range cat.Services {
		if other.Namespace == svc.Namespace {
			return fmt.Errorf("%s: namespace %q is already used by %s", full, svc.Namespace, other.Proto.Desc.FullName())
		}
	}
	svc.Instructions = firstNonEmpty(opts.GetInstructions(), mcpOpts.GetApp().GetDescription(), svc.Doc)
	for _, m := range s.Methods {
		if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
			continue
		}
		meth, err := cat.method(svc, m)
		if err != nil {
			return err
		}
		if meth != nil {
			svc.Methods = append(svc.Methods, meth)
		}
	}
	for _, m := range svc.Methods {
		mo, _, _ := methodOptions(m.Proto.Desc)
		if err := cat.resolveRevert(svc, m, mo.GetRevert()); err != nil {
			return fmt.Errorf("%s: %w", m.Proto.Desc.FullName(), err)
		}
	}
	for _, sn := range opts.GetSnippets() {
		snip := &Snippet{Name: sn.GetName(), Description: sn.GetDescription(), Code: sn.GetCode()}
		if t := sn.GetInputType(); t != "" {
			if snip.Input = cat.byMessage[protoreflect.FullName(t)]; snip.Input == nil {
				return fmt.Errorf("%s: snippet %q input_type %s is not in the generated files", full, snip.Name, t)
			}
		}
		cat.Snippets = append(cat.Snippets, snip)
	}
	p.Services = append(p.Services, svc)
	cat.Services = append(cat.Services, svc)
	return nil
}

// method builds one sandbox method, applying codemode.v1 over mcp.v1 over
// what the AIP method kind implies.
func (cat *Catalog) method(svc *Service, m *protogen.Method) (*Method, error) {
	opts, tool, elicit := methodOptions(m.Desc)
	if opts.GetExclude() {
		return nil, nil
	}
	in, err := cat.typeRef(m.Input)
	if err != nil {
		return nil, err
	}
	out, err := cat.typeRef(m.Output)
	if err != nil {
		return nil, err
	}
	rpc := string(m.Desc.Name())
	name := methodName(rpc, cat.Config.Naming)
	meth := &Method{
		Proto: m, Service: svc, Name: name, Path: svc.Namespace + "." + name,
		Doc:   firstNonEmpty(tool.GetDescription(), comment(m.Comments)),
		Input: in, Output: out, HTTP: httpRule(m.Desc), Examples: opts.GetExamples(),
		ApprovalMsg: elicit.GetMessage(),
	}
	level := methodIdempotency(m.Desc)
	readOnlyKind := hasAnyPrefix(rpc, "Get", "List", "Search", "BatchGet", "Lookup") ||
		level == "NO_SIDE_EFFECTS" || (meth.HTTP != nil && meth.HTTP.Method == "GET")
	if tool == nil {
		tool = &mcppb.McpToolOptions{}
	}
	meth.ReadOnly = boolOr(tool.ReadOnly, readOnlyKind)
	destructive := boolOr(tool.Destructive, cat.Config.InferFromAIP() && hasAnyPrefix(rpc, "Delete", "Purge", "Remove"))
	meth.Idempotent = boolOr(tool.Idempotent, meth.ReadOnly || level == "IDEMPOTENT")

	switch opts.GetApproval() {
	case codemodepb.Approval_APPROVAL_REQUIRED:
		meth.Approval = true
	case codemodepb.Approval_APPROVAL_NONE:
		meth.Approval = false
	default:
		meth.Approval = (destructive && !meth.ReadOnly && cat.Config.Approval.Destructive == "require") ||
			(elicit != nil && cat.Config.Approval.Elicitation == "require")
	}
	switch opts.GetReplay() {
	case codemodepb.Replay_REPLAY_LOG:
		meth.Replay = "log"
	case codemodepb.Replay_REPLAY_REEXECUTE:
		if meth.Approval {
			return nil, fmt.Errorf("%s: REPLAY_REEXECUTE cannot be combined with approval", m.Desc.FullName())
		}
		meth.Replay = "reexecute"
	default:
		meth.Replay = "log"
		if meth.ReadOnly && !meth.Approval {
			meth.Replay = "reexecute"
		}
	}
	return meth, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

func boolOr(p *bool, fallback bool) bool {
	if p != nil {
		return *p
	}
	return fallback
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
