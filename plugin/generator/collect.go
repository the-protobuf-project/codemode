// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"sort"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// index finds the protogen wrapper, with its comments, for any descriptor in
// the request, including imported files.
type index struct {
	messages map[protoreflect.FullName]*protogen.Message
	enums    map[protoreflect.FullName]*protogen.Enum
	files    []*protogen.File
}

func newIndex(gen *protogen.Plugin) *index {
	ix := &index{
		messages: map[protoreflect.FullName]*protogen.Message{},
		enums:    map[protoreflect.FullName]*protogen.Enum{},
		files:    gen.Files,
	}
	var walk func(ms []*protogen.Message)
	walk = func(ms []*protogen.Message) {
		for _, m := range ms {
			ix.messages[m.Desc.FullName()] = m
			for _, e := range m.Enums {
				ix.enums[e.Desc.FullName()] = e
			}
			walk(m.Messages)
		}
	}
	for _, f := range gen.Files {
		for _, e := range f.Enums {
			ix.enums[e.Desc.FullName()] = e
		}
		walk(f.Messages)
	}
	return ix
}

// collector gathers the messages and enums the generated code needs: every
// type declared in the files being generated, and every type they reach.
type collector struct {
	ix       *index
	messages map[protoreflect.FullName]*protogen.Message
	enums    map[protoreflect.FullName]*protogen.Enum
}

func (c *collector) addMessage(m *protogen.Message) {
	name := m.Desc.FullName()
	if m.Desc.IsMapEntry() || lookupWellKnown(name) != nil || c.messages[name] != nil {
		if m.Desc.IsMapEntry() {
			c.addFields(m)
		}
		return
	}
	c.messages[name] = m
	c.addFields(m)
}

func (c *collector) addFields(m *protogen.Message) {
	for _, f := range m.Fields {
		if f.Message != nil {
			c.addMessage(f.Message)
		}
		if f.Enum != nil {
			c.enums[f.Enum.Desc.FullName()] = f.Enum
		}
	}
}

func (c *collector) addFile(f *protogen.File) {
	var walk func(ms []*protogen.Message)
	walk = func(ms []*protogen.Message) {
		for _, m := range ms {
			c.addMessage(m)
			for _, e := range m.Enums {
				c.enums[e.Desc.FullName()] = e
			}
			walk(m.Messages)
		}
	}
	for _, e := range f.Enums {
		c.enums[e.Desc.FullName()] = e
	}
	walk(f.Messages)
}

// sortedMessages returns the collected messages in full-name order, so output
// is deterministic.
func (c *collector) sortedMessages() []*protogen.Message {
	out := make([]*protogen.Message, 0, len(c.messages))
	for _, m := range c.messages {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Desc.FullName() < out[j].Desc.FullName() })
	return out
}

func (c *collector) sortedEnums() []*protogen.Enum {
	out := make([]*protogen.Enum, 0, len(c.enums))
	for _, e := range c.enums {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Desc.FullName() < out[j].Desc.FullName() })
	return out
}

// pkg returns the catalog package for name, creating it on first use.
func (cat *Catalog) pkg(name protoreflect.FullName) *Package {
	for _, p := range cat.Packages {
		if p.Name == name {
			return p
		}
	}
	p := &Package{Name: name, Dir: packageDir(name)}
	cat.Packages = append(cat.Packages, p)
	sort.Slice(cat.Packages, func(i, j int) bool { return cat.Packages[i].Name < cat.Packages[j].Name })
	return p
}
