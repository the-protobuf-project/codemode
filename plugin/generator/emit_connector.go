// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// connectorFunc is the generated connector factory: "libraryServiceConnector".
func (s *Service) connectorFunc() string {
	return lowerFirstWord(s.Name) + "Connector"
}

func (cat *Catalog) emitConnector(gen *protogen.Plugin, s *Service) {
	f := cat.newFile(s.Dir+"/connector.ts", s.Source)
	n := moduleNamer{f}
	f.runtime("connector", "unaryMethod")
	f.runtimeType("Connector", "ConnectorOptions")
	f.localType(s.Dir+"/service.ts", s.Name)
	f.P("/**")
	f.P(" * Exposes {@link ", s.Name, "} in the Code Mode sandbox as `", s.Namespace, "`.")
	f.P(" * Pass the generated client, or any object implementing the interface.")
	f.P(" */")
	f.P("export function ", s.connectorFunc(), "(service: ", s.Name, ", options: ConnectorOptions = {}): Connector {")
	f.P("  return connector(")
	f.P("    {")
	f.P("      namespace: ", tsString(s.Namespace), ",")
	if s.Instructions != "" {
		f.P("      instructions: ", tsString(s.Instructions), ",")
	}
	f.P("      methods: [")
	for _, m := range s.Methods {
		f.P("        unaryMethod({")
		f.P("          name: ", tsString(m.Name), ",")
		f.P("          input: ", refCodec(f, m.Input), ",")
		f.P("          output: ", refCodec(f, m.Output), ",")
		f.P("          requiresApproval: ", m.Approval, ",")
		f.P("          replay: ", tsString(m.Replay), ",")
		if m.ApprovalMsg != "" {
			f.P("          approvalMessage: ", tsString(m.ApprovalMsg), ",")
		}
		f.P("          call: (request, context) => service.", m.Client(), "(request, { signal: context.signal }),")
		if r := m.Revert; r != nil {
			emitRevert(f, n, m, r)
		}
		f.P("        }),")
	}
	f.P("      ],")
	f.P("    },")
	f.P("    options,")
	f.P("  );")
	f.P("}")
	f.emit(gen)
}

// emitRevert writes the compensating call: the revert RPC, with its request
// built from the original request and response.
func emitRevert(f *tsFile, n moduleNamer, m *Method, r *Revert) {
	var fields []string
	for _, rf := range r.Fields {
		root := "request"
		if rf.Source == "result" {
			root = "response"
		}
		fields = append(fields, propertyKey(rf.Field.JSONName)+": "+root+"."+strings.Join(rf.Path, "?."))
	}
	reqType := refType(r.Method.Input, n)
	req := "{ " + strings.Join(fields, ", ") + " } as " + reqType
	if len(fields) == 0 {
		req = "{} as " + reqType
	}
	used := func(src string) string {
		for _, rf := range r.Fields {
			if rf.Source == src {
				if src == "args" {
					return "request"
				}
				return "response"
			}
		}
		if src == "args" {
			return "_request"
		}
		return "_response"
	}
	f.P("          revert: async (", used("args"), ", ", used("result"), ", context) => {")
	f.P("            await service.", r.Method.Client(), "(", req, ", { signal: context.signal });")
	f.P("          },")
}
