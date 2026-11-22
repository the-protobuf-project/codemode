// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Build turns the files protoc asked for into the Code Mode catalog.
func Build(gen *protogen.Plugin, cfg *Config) (*Catalog, error) {
	if err := cfg.checkOverrides(gen.Files); err != nil {
		return nil, err
	}
	a, err := buildAIP(gen)
	if err != nil {
		return nil, err
	}
	cat := &Catalog{
		aip:        a,
		Config:     cfg,
		Version:    protocVersion(gen),
		byMessage:  map[protoreflect.FullName]*Message{},
		byEnum:     map[protoreflect.FullName]*Enum{},
		byResource: map[string]*Resource{},
	}
	ix := newIndex(gen)
	col := &collector{ix: ix, messages: map[protoreflect.FullName]*protogen.Message{}, enums: map[protoreflect.FullName]*protogen.Enum{}}
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		col.addFile(f)
		for _, s := range f.Services {
			for _, m := range s.Methods {
				col.addMessage(m.Input)
				col.addMessage(m.Output)
			}
		}
	}
	for _, e := range col.sortedEnums() {
		cat.addEnum(e)
	}
	msgs := col.sortedMessages()
	for _, m := range msgs {
		cat.addMessage(m)
	}
	for _, m := range msgs {
		cat.buildFields(cat.byMessage[m.Desc.FullName()], m)
	}
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, s := range f.Services {
			if err := cat.addService(s, f); err != nil {
				return nil, err
			}
		}
	}
	for _, p := range cat.Packages {
		p.Doc = packageDoc(gen.Files, p.Name)
	}
	cat.assignSandboxNames()
	return cat, nil
}

func (cat *Catalog) addEnum(e *protogen.Enum) {
	p := cat.pkg(e.Desc.ParentFile().Package())
	name := typeName(e.Desc)
	en := &Enum{
		Desc:    e.Desc,
		Package: p,
		Name:    name,
		Path:    p.Dir + "/enums/" + fileStem(name),
		Source:  e.Desc.ParentFile().Path(),
		Doc:     enumDoc(e.Desc, comment(e.Comments)),
	}
	for _, v := range e.Values {
		en.Values = append(en.Values, EnumValue{
			Name:       string(v.Desc.Name()),
			Number:     int32(v.Desc.Number()),
			Doc:        enumValueDoc(v.Desc, comment(v.Comments)),
			Deprecated: isDeprecated(v.Desc),
		})
	}
	p.Enums = append(p.Enums, en)
	cat.byEnum[e.Desc.FullName()] = en
}

func (cat *Catalog) addMessage(m *protogen.Message) {
	p := cat.pkg(m.Desc.ParentFile().Package())
	name := typeName(m.Desc)
	msg := &Message{
		Desc:    m.Desc,
		Package: p,
		Name:    name,
		Path:    p.Dir + "/messages/" + fileStem(name),
		Source:  m.Desc.ParentFile().Path(),
		Doc:     comment(m.Comments),
	}
	p.Messages = append(p.Messages, msg)
	cat.byMessage[m.Desc.FullName()] = msg
}

// typeRef resolves an RPC input or output type.
func (cat *Catalog) typeRef(m *protogen.Message) (TypeRef, error) {
	if isWellKnown(m.Desc.FullName()) {
		return TypeRef{WKT: m.Desc.FullName()}, nil
	}
	if msg := cat.byMessage[m.Desc.FullName()]; msg != nil {
		return TypeRef{Message: msg}, nil
	}
	return TypeRef{}, fmt.Errorf("message %s is not collected", m.Desc.FullName())
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("v%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
	if v.GetSuffix() != "" {
		s += "-" + v.GetSuffix()
	}
	return s
}

// assignSandboxNames gives every type a name that is unique across packages
// in sandbox.d.ts, prefixing the package only where two packages collide.
func (cat *Catalog) assignSandboxNames() {
	count := map[string]int{}
	for _, p := range cat.Packages {
		for _, m := range p.Messages {
			count[m.Name]++
		}
		for _, e := range p.Enums {
			count[e.Name]++
		}
		for _, r := range p.Resources {
			count[r.Name]++
		}
	}
	unique := func(p *Package, name string) string {
		if count[name] > 1 {
			return pascal(string(p.Name)) + name
		}
		return name
	}
	for _, p := range cat.Packages {
		for _, m := range p.Messages {
			m.Sandbox = unique(p, m.Name)
		}
		for _, e := range p.Enums {
			e.Sandbox = unique(p, e.Name)
		}
		for _, r := range p.Resources {
			r.Sandbox = unique(p, r.Name)
		}
	}
}
