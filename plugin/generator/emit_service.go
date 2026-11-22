// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// emitService writes services/<name>/: the service interface and descriptor,
// a client over any transport, the Code Mode connector and an index.
func (cat *Catalog) emitService(gen *protogen.Plugin, s *Service) {
	cat.emitServiceInterface(gen, s)
	cat.emitClient(gen, s)
	cat.emitConnector(gen, s)
	f := cat.newFile(s.Dir+"/index.ts", s.Source)
	f.P("export * from ", tsString("./service"+cat.Config.Extension()), ";")
	f.P("export * from ", tsString("./client"+cat.Config.Extension()), ";")
	f.P("export * from ", tsString("./connector"+cat.Config.Extension()), ";")
	f.emit(gen)
}

func (cat *Catalog) emitServiceInterface(gen *protogen.Plugin, s *Service) {
	f := cat.newFile(s.Dir+"/service.ts", s.Source)
	n := moduleNamer{f}
	f.runtimeType("CallOptions", "MethodDescriptor")
	f.body.WriteString(jsdoc("", s.Doc))
	f.P("export interface ", s.Name, " {")
	for i, m := range s.Methods {
		if i > 0 {
			f.P()
		}
		f.body.WriteString(jsdoc("  ", m.Doc))
		f.P("  ", m.Client(), "(request: ", refType(m.Input, n), ", options?: CallOptions): Promise<", refType(m.Output, n), ">;")
	}
	f.P("}")
	f.P()
	f.P("/** Describes each {@link ", s.Name, "} RPC for transports. */")
	f.P("export const ", s.Name, "Descriptor = {")
	f.P("  typeName: ", tsString(string(s.Proto.Desc.FullName())), ",")
	f.P("  methods: {")
	for _, m := range s.Methods {
		f.P("    ", m.Client(), ": {")
		f.P("      service: ", tsString(string(s.Proto.Desc.FullName())), ",")
		f.P("      name: ", tsString(string(m.Proto.Desc.Name())), ",")
		f.P("      input: ", refCodec(f, m.Input), ",")
		f.P("      output: ", refCodec(f, m.Output), ",")
		if h := m.HTTP; h != nil {
			http := "{ method: " + tsString(h.Method) + ", path: " + tsString(h.Path)
			if h.Body != "" {
				http += ", body: " + tsString(h.Body)
			}
			if h.ResponseBody != "" {
				http += ", responseBody: " + tsString(h.ResponseBody)
			}
			f.P("      http: ", http, " },")
		}
		if level := methodIdempotency(m.Proto.Desc); level != "" {
			f.P("      idempotency: ", tsString(level), ",")
		}
		f.P("    } satisfies MethodDescriptor<", refType(m.Input, n), ", ", refType(m.Output, n), ">,")
	}
	f.P("  },")
	f.P("} as const;")
	f.emit(gen)
}

func (cat *Catalog) emitClient(gen *protogen.Plugin, s *Service) {
	f := cat.newFile(s.Dir+"/client.ts", s.Source)
	n := moduleNamer{f}
	f.runtimeType("CallOptions", "Transport")
	f.local(s.Dir+"/service.ts", s.Name+"Descriptor")
	f.localType(s.Dir+"/service.ts", s.Name)
	f.P("/** Calls {@link ", s.Name, "} over a transport: HTTP, Connect, or a service binding. */")
	f.P("export class ", s.Name, "Client implements ", s.Name, " {")
	f.P("  readonly transport: Transport;")
	f.P()
	f.P("  constructor(transport: Transport) {")
	f.P("    this.transport = transport;")
	f.P("  }")
	for _, m := range s.Methods {
		f.P()
		f.body.WriteString(jsdoc("  ", m.Doc))
		f.P("  ", m.Client(), "(request: ", refType(m.Input, n), ", options?: CallOptions): Promise<", refType(m.Output, n), "> {")
		f.P("    return this.transport.unary(", s.Name, "Descriptor.methods.", m.Client(), ", request, options);")
		f.P("  }")
	}
	f.P("}")
	f.emit(gen)
}

// refCodec imports and returns the codec of an RPC input or output.
func refCodec(f *tsFile, t TypeRef) string {
	if t.WKT != nil {
		f.runtime(t.WKT.Codec)
		return t.WKT.Codec
	}
	f.local(t.Message.File, t.Message.Codec())
	return t.Message.Codec()
}

// Client is the method's name on the generated interface and client, which
// is always lowerCamel whatever the sandbox naming.
func (m *Method) Client() string {
	return lowerFirstWord(string(m.Proto.Desc.Name()))
}
