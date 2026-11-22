// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// resourceDef is a google.api.resource pattern and where it was declared.
type resourceDef struct {
	pattern string
	pkg     protoreflect.FullName
	typ     string
}

var patternVar = regexp.MustCompile(`\{([a-z0-9_]+)(=[^}]*)?\}`)

// indexResources records every resource pattern in the request, so a
// resource_reference in one file can resolve a type declared in another.
func (cat *Catalog) indexResources(ix *index) {
	defs := map[string]resourceDef{}
	add := func(r *annotations.ResourceDescriptor, pkg protoreflect.FullName) {
		if r.GetType() != "" && len(r.GetPattern()) > 0 {
			if _, seen := defs[r.GetType()]; !seen {
				defs[r.GetType()] = resourceDef{pattern: r.GetPattern()[0], pkg: pkg, typ: r.GetType()}
			}
		}
	}
	for _, f := range ix.files {
		for _, r := range fileResources(f.Desc) {
			add(r, f.Desc.Package())
		}
	}
	for _, m := range ix.messages {
		if r := resourceOf(m.Desc); r != nil {
			add(r, m.Desc.ParentFile().Package())
		}
	}
	cat.resourceDefs = defs
}

// resource returns the generated name type for a resource type, creating it
// the first time a message or field uses it.
func (cat *Catalog) resource(typ string) *Resource {
	if r := cat.byResource[typ]; r != nil {
		return r
	}
	def, ok := cat.resourceDefs[typ]
	if !ok {
		return nil
	}
	p := cat.pkg(def.pkg)
	name := typ[strings.LastIndex(typ, "/")+1:] + "Name"
	for _, m := range p.Messages {
		if m.Name == name {
			name = strings.TrimSuffix(name, "Name") + "ResourceName"
		}
	}
	r := &Resource{
		Type:     typ,
		Pattern:  def.pattern,
		Package:  p,
		Name:     name,
		Path:     p.Dir + "/resources/" + fileStem(name),
		Template: patternVar.ReplaceAllString(def.pattern, "$${string}"),
	}
	for _, m := range patternVar.FindAllStringSubmatch(def.pattern, -1) {
		r.Vars = append(r.Vars, m[1])
	}
	p.Resources = append(p.Resources, r)
	cat.byResource[typ] = r
	return r
}

// buildFields resolves every field of msg to its TypeScript shape.
func (cat *Catalog) buildFields(msg *Message, m *protogen.Message) {
	nameField := ""
	if r := resourceOf(m.Desc); r != nil {
		msg.Resource = cat.resource(r.GetType())
		nameField = r.GetNameField()
		if nameField == "" {
			nameField = "name"
		}
	}
	oneofs := map[protoreflect.FullName]*Oneof{}
	for _, f := range m.Fields {
		fd := f.Desc
		field := cat.field(fd)
		field.Doc = fieldDoc(fd, comment(f.Comments))
		field.Required = isRequired(fd)
		field.OutputOnly = hasBehavior(fd, annotations.FieldBehavior_OUTPUT_ONLY)
		if string(fd.Name()) == nameField && fd.Kind() == protoreflect.StringKind {
			field.Resource = msg.Resource
		}
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			o := oneofs[od.FullName()]
			if o == nil {
				o = &Oneof{Name: string(od.Name())}
				oneofs[od.FullName()] = o
				msg.Oneofs = append(msg.Oneofs, o)
			}
			o.Fields = append(o.Fields, field)
			field.Oneof = o
		}
		msg.Fields = append(msg.Fields, field)
	}
}

// field resolves the type of one field; map entries resolve key and value.
func (cat *Catalog) field(fd protoreflect.FieldDescriptor) *Field {
	f := &Field{
		Desc:      fd,
		JSONName:  fd.JSONName(),
		ProtoName: string(fd.Name()),
		Explicit:  fd.HasPresence() && !fd.IsList() && !fd.IsMap(),
	}
	if fd.IsMap() {
		f.MapKey = cat.field(fd.MapKey())
		f.MapValue = cat.field(fd.MapValue())
		return f
	}
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if isWellKnown(fd.Message().FullName()) {
			f.WKT = fd.Message().FullName()
		} else {
			f.Message = cat.byMessage[fd.Message().FullName()]
		}
	case protoreflect.EnumKind:
		f.Enum = cat.byEnum[fd.Enum().FullName()]
	case protoreflect.StringKind:
		if t := resourceReference(fd); t != "" && t != "*" {
			f.Resource = cat.resource(t)
		}
	}
	return f
}
