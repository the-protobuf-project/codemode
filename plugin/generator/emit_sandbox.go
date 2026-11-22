// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"sort"
	"strings"
)

// codemodeGlobal declares the runtime API every sandbox has.
const codemodeGlobal = `/** A search hit: a method or a saved snippet. */
interface CodemodeSearchResult {
  path: string;
  kind: "method" | "snippet";
  description?: string;
  requiresApproval?: boolean;
  score: number;
}

/** The error every failed call rejects with. Branch on code. */
interface CodemodeError extends Error {
  code: "CANCELLED" | "UNKNOWN" | "INVALID_ARGUMENT" | "DEADLINE_EXCEEDED" | "NOT_FOUND" | "ALREADY_EXISTS" | "PERMISSION_DENIED" | "RESOURCE_EXHAUSTED" | "FAILED_PRECONDITION" | "ABORTED" | "OUT_OF_RANGE" | "UNIMPLEMENTED" | "INTERNAL" | "UNAVAILABLE" | "DATA_LOSS" | "UNAUTHENTICATED";
  reason?: string;
  details?: unknown[];
}

declare const codemode: {
  /** Finds methods and snippets by keyword. */
  search(query: string): Promise<{ results: CodemodeSearchResult[]; total: number; truncated: boolean }>;
  /** Returns TypeScript for a namespace, "namespace.method", or snippet. */
  describe(target: string): Promise<{ path: string; kind: "namespace" | "method" | "snippet"; description?: string; requiresApproval?: boolean; types: string }>;
  /** Records a non-deterministic value (time, randomness) so replays reuse it. */
  step<T>(name: string, fn: () => T | Promise<T>): Promise<T>;
  /** Runs a saved snippet. */
  run(name: string, input?: unknown): Promise<unknown>;
};
`

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

// methodDecl is one method's signature inside its namespace declaration.
func methodDecl(m *Method) string {
	n := sandboxNamer{}
	var notes []string
	if m.Approval {
		notes = append(notes, "Requires approval: the execution pauses until a person approves this call.")
	}
	if m.Revert != nil {
		notes = append(notes, "Undone on rollback by "+m.Service.Namespace+"."+m.Revert.Method.Name+".")
	}
	return jsdoc("  ", m.Doc, notes...) +
		"  " + m.Name + "(input: " + refType(m.Input, n) + "): Promise<" + refType(m.Output, n) + ">;\n"
}

// namespaceDecl declares one service's sandbox global.
func namespaceDecl(s *Service) string {
	var b strings.Builder
	b.WriteString(jsdoc("", s.Instructions))
	b.WriteString("declare const " + s.Namespace + ": {\n")
	for i, m := range s.Methods {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(methodDecl(m))
	}
	b.WriteString("};\n")
	return b.String()
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

// sandboxDTS is the whole of sandbox.d.ts: what a model sees with inline
// discovery, and what editors use to check sandbox code.
func (cat *Catalog) sandboxDTS() string {
	var b strings.Builder
	b.WriteString(sandboxPrelude + "\n" + codemodeGlobal)
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
	for _, name := range names {
		b.WriteString("\n" + types[name])
	}
	for _, s := range cat.Services {
		b.WriteString("\n" + namespaceDecl(s))
	}
	return b.String()
}
