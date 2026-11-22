// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"github.com/the-protobuf-project/protokit/service"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// aip is protokit's service IR for the request: the google.api.* vocabulary
// (http, resource, resource_reference, field_behavior) already parsed and
// checked, indexed for the catalog build.
type aip struct {
	ir      *service.IR
	methods map[protoreflect.FullName]*service.Method
	files   map[string]protoreflect.FullName // file-level resource type -> package
}

func buildAIP(gen *protogen.Plugin) (*aip, error) {
	ir, err := service.Build(gen, service.Options{})
	if err != nil {
		return nil, err
	}
	a := &aip{ir: ir, methods: map[protoreflect.FullName]*service.Method{}, files: map[string]protoreflect.FullName{}}
	for _, s := range ir.Services {
		for _, m := range s.Methods {
			a.methods[protoreflect.FullName(m.FullName)] = m
		}
	}
	// protokit reads resources off messages only; add file-level
	// google.api.resource_definition entries it does not see.
	for _, f := range gen.Files {
		for _, rd := range fileResources(f.Desc) {
			a.define(rd, f.Desc.Package())
		}
	}
	return a, nil
}

func (a *aip) define(rd *annotations.ResourceDescriptor, pkg protoreflect.FullName) {
	if _, seen := a.ir.Resources[rd.GetType()]; seen || rd.GetType() == "" {
		return
	}
	a.files[rd.GetType()] = pkg
	r := &service.Resource{Type: rd.GetType(), NameField: rd.GetNameField(), Singular: rd.GetSingular(), Plural: rd.GetPlural()}
	for _, raw := range rd.GetPattern() {
		if p, err := service.ParsePattern(raw); err == nil {
			r.Patterns = append(r.Patterns, p)
		}
	}
	if len(r.Patterns) > 0 {
		a.ir.Resources[r.Type] = r
	}
}

// field returns protokit's view of a field. protokit only materializes
// messages an RPC reaches, so a message used only by a snippet gets the two
// annotations Code Mode needs read here instead.
func (a *aip) field(fd protoreflect.FieldDescriptor) *service.Field {
	if m := a.ir.Messages[string(fd.ContainingMessage().FullName())]; m != nil {
		return m.Field(string(fd.Name()))
	}
	f := &service.Field{Name: string(fd.Name())}
	for _, b := range ext[[]annotations.FieldBehavior](fd.Options(), annotations.E_FieldBehavior) {
		switch b {
		case annotations.FieldBehavior_REQUIRED:
			f.Behavior = append(f.Behavior, service.BehaviorRequired)
		case annotations.FieldBehavior_OUTPUT_ONLY:
			f.Behavior = append(f.Behavior, service.BehaviorOutputOnly)
		}
	}
	f.RefType = ext[*annotations.ResourceReference](fd.Options(), annotations.E_ResourceReference).GetType()
	return f
}

// messageResource returns the google.api.resource a message declares, or nil.
func (a *aip) messageResource(md protoreflect.MessageDescriptor) *service.Resource {
	if m := a.ir.Messages[string(md.FullName())]; m != nil {
		return m.Resource
	}
	return nil
}

// httpRule returns a method's primary google.api.http binding, or nil.
func (a *aip) httpRule(md protoreflect.MethodDescriptor) *HTTPRule {
	m := a.methods[md.FullName()]
	if m == nil || len(m.Bindings) == 0 {
		return nil
	}
	b := m.Bindings[0]
	out := &HTTPRule{Method: b.HTTPMethod, Path: b.Template.Raw}
	switch {
	case b.Body == nil:
	case b.Body.Wildcard:
		out.Body = "*"
	case b.Body.Field != nil:
		out.Body = strings.Join(b.Body.Field.Proto, ".")
	}
	if b.ResponseBody != nil {
		out.ResponseBody = strings.Join(b.ResponseBody.Proto, ".")
	}
	return out
}

// patternTemplate renders a resource pattern as a TypeScript template literal type,
// each variable a ${string}.
func patternTemplate(p *service.Pattern) string {
	parts := make([]string, len(p.Segments))
	for i, s := range p.Segments {
		parts[i] = s.Literal
		if s.Var != "" {
			parts[i] = "${string}"
		}
	}
	return strings.Join(parts, "/")
}
