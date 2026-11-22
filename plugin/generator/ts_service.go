// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// serviceView is the data of service.ts.tpl, client.ts.tpl and
// connector.ts.tpl.
type serviceView struct {
	Doc, Name, TypeName, Namespace, Instructions, Func string
	Methods                                            []methodView
}

// methodView is one RPC, with every name and type already resolved and
// imported for the file it is rendered into.
type methodView struct {
	Doc, Name, Client, RPC  string
	Input, Output           string
	InputCodec, OutputCodec string
	HTTP, Idempotency       string
	Approval                bool
	Replay, ApprovalMsg     string
	Revert                  *revertView
}

// viewParts says which parts of a method a file uses, so it imports only those.
type viewParts struct {
	types, codecs, reverts bool
}

// newServiceView resolves a service for one file, importing what it names.
func newServiceView(f *tsFile, s *Service, parts viewParts) serviceView {
	n := moduleNamer{f}
	v := serviceView{
		Doc: jsdoc("", s.Doc), Name: s.Name, TypeName: string(s.Proto.Desc.FullName()),
		Namespace: s.Namespace, Instructions: s.Instructions, Func: s.connectorFunc(),
	}
	for _, m := range s.Methods {
		mv := methodView{
			Doc: jsdoc("  ", m.Doc), Name: m.Name, Client: m.Client(), RPC: string(m.Proto.Desc.Name()),
			Idempotency: methodIdempotency(m.Proto.Desc),
			Approval:    m.Approval, Replay: m.Replay, ApprovalMsg: m.ApprovalMsg,
		}
		if parts.types {
			mv.Input, mv.Output = refType(m.Input, n), refType(m.Output, n)
		}
		if parts.codecs {
			mv.InputCodec, mv.OutputCodec = refCodec(f, m.Input), refCodec(f, m.Output)
		}
		if h := m.HTTP; h != nil {
			mv.HTTP = "{ method: " + tsString(h.Method) + ", path: " + tsString(h.Path)
			if h.Body != "" {
				mv.HTTP += ", body: " + tsString(h.Body)
			}
			if h.ResponseBody != "" {
				mv.HTTP += ", responseBody: " + tsString(h.ResponseBody)
			}
			mv.HTTP += " }"
		}
		if parts.reverts && m.Revert != nil {
			mv.Revert = newRevertView(n, m.Revert)
		}
		v.Methods = append(v.Methods, mv)
	}
	return v
}

// emitService writes services/<name>/: the service interface and descriptor,
// a client over any transport, the Code Mode connector and an index.
func (cat *Catalog) emitService(gen *protogen.Plugin, s *Service) {
	ext := cat.Config.Extension()

	f := cat.newFile(s.Dir+"/service", s.Source)
	f.runtimeType("CallOptions", "MethodDescriptor")
	f.render("service.ts.tpl", newServiceView(f, s, viewParts{types: true, codecs: true}))
	f.emit(gen)

	c := cat.newFile(s.Dir+"/client", s.Source)
	c.runtimeType("CallOptions", "Transport")
	c.local(s.Dir+"/service", s.Name+"Descriptor")
	c.localType(s.Dir+"/service", s.Name)
	c.render("client.ts.tpl", newServiceView(c, s, viewParts{types: true}))
	c.emit(gen)

	k := cat.newFile(s.Dir+"/connector", s.Source)
	k.runtime("connector", "unaryMethod")
	k.runtimeType("Connector", "ConnectorOptions")
	k.localType(s.Dir+"/service", s.Name)
	k.render("connector.ts.tpl", newServiceView(k, s, viewParts{codecs: true, reverts: true}))
	k.emit(gen)

	i := cat.newFile(s.Dir+"/index", s.Source)
	i.render("service_index.ts.tpl", struct{ Service, Client, Connector string }{
		"./service" + ext, "./client" + ext, "./connector" + ext,
	})
	i.emit(gen)
}

// refCodec imports and returns the codec of an RPC input or output.
func refCodec(f *tsFile, t TypeRef) string {
	if t.WKT != "" {
		f.runtime(tsWKT(t.WKT).Codec)
		return tsWKT(t.WKT).Codec
	}
	f.local(t.Message.Path, t.Message.Codec())
	return t.Message.Codec()
}

// Client is the method's name on the generated interface and client, which
// is always lowerCamel whatever the sandbox naming.
func (m *Method) Client() string {
	return lowerFirstWord(string(m.Proto.Desc.Name()))
}
