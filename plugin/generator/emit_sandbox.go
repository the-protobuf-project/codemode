// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "sort"

// sandboxTypes returns the declaration of every generated type, keyed by its
// sandbox name.
func (cat *Catalog) sandboxTypes() map[string]string {
	n := sandboxNamer{}
	out := map[string]string{}
	for _, p := range cat.Packages {
		for _, m := range p.Messages {
			out[m.Sandbox] = interfaceDecl(m, n, "")
		}
		for _, e := range p.Enums {
			out[e.Sandbox] = enumDecl(e, e.Sandbox, "")
		}
		for _, r := range p.Resources {
			out[r.Sandbox] = resourceDecl(r, r.Sandbox, "")
		}
	}
	return out
}

// sandboxMethod is the "method" template's data: one signature in a namespace.
type sandboxMethod struct {
	Doc, Name, Input, Output string
}

func newSandboxMethod(m *Method) sandboxMethod {
	n := sandboxNamer{}
	var notes []string
	if m.Approval {
		notes = append(notes, "Requires approval: the execution pauses until a person approves this call.")
	}
	if m.Revert != nil {
		notes = append(notes, "Undone on rollback by "+m.Service.Namespace+"."+m.Revert.Method.Name+".")
	}
	return sandboxMethod{jsdoc("  ", m.Doc, notes...), m.Name, refType(m.Input, n), refType(m.Output, n)}
}

// methodDecl is one method's signature inside its namespace declaration.
func methodDecl(m *Method) string {
	return renderString("method", newSandboxMethod(m))
}

// sandboxNamespace is the "namespace" template's data: one sandbox global.
type sandboxNamespace struct {
	Doc, Name string
	Methods   []sandboxMethod
}

func newSandboxNamespace(s *Service) sandboxNamespace {
	v := sandboxNamespace{Doc: jsdoc("", s.Instructions), Name: s.Namespace}
	for _, m := range s.Methods {
		v.Methods = append(v.Methods, newSandboxMethod(m))
	}
	return v
}

// namespaceDecl declares one service's sandbox global.
func namespaceDecl(s *Service) string {
	return renderString("namespace", newSandboxNamespace(s))
}

// typeClosure returns the sandbox names of every type a method's input and
// output reach, sorted.
func typeClosure(m *Method) []string {
	seen := map[string]bool{}
	var visit func(msg *Message)
	visitField := func(f *Field) {
		for _, x := range []*Field{f, f.MapValue} {
			if x == nil {
				continue
			}
			if x.Enum != nil {
				seen[x.Enum.Sandbox] = true
			}
			if x.Resource != nil {
				seen[x.Resource.Sandbox] = true
			}
			if x.Message != nil {
				visit(x.Message)
			}
		}
	}
	visit = func(msg *Message) {
		if seen[msg.Sandbox] {
			return
		}
		seen[msg.Sandbox] = true
		for _, f := range msg.Fields {
			visitField(f)
		}
	}
	for _, t := range []TypeRef{m.Input, m.Output} {
		if t.Message != nil {
			visit(t.Message)
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// sandboxTypesUsed returns the declarations of the types some method
// reaches, in name order: the types sandbox.d.ts declares.
func (cat *Catalog) sandboxTypesUsed() []string {
	types := cat.sandboxTypes()
	used := map[string]bool{}
	for _, s := range cat.Services {
		for _, m := range s.Methods {
			for _, t := range typeClosure(m) {
				used[t] = true
			}
		}
	}
	names := make([]string, 0, len(used))
	for k := range used {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = types[name]
	}
	return out
}
