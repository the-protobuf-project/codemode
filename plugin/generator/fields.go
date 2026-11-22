// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"github.com/the-protobuf-project/protokit/service"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// resource returns the generated name type for a resource type, creating it
// the first time a message or field uses it.
func (cat *Catalog) resource(typ string) *Resource {
	if r := cat.byResource[typ]; r != nil {
		return r
	}
	def := cat.aip.ir.Resources[typ]
	if def == nil || len(def.Patterns) == 0 {
		return nil
	}
	p := cat.pkg(cat.resourcePackage(def))
	name := typ[strings.LastIndex(typ, "/")+1:] + "Name"
	for _, m := range p.Messages {
		if m.Name == name {
			name = strings.TrimSuffix(name, "Name") + "ResourceName"
		}
	}
	pattern := def.Patterns[0]
	r := &Resource{
		Type:     typ,
		Pattern:  pattern.Raw,
		Package:  p,
		Name:     name,
		Path:     p.Dir + "/resources/" + fileStem(name),
		Template: patternTemplate(pattern),
		Vars:     pattern.Vars(),
	}
	p.Resources = append(p.Resources, r)
	cat.byResource[typ] = r
	return r
}

// resourcePackage is where a resource's name type is generated: beside the
// message that declares it, or in the file that defines it.
func (cat *Catalog) resourcePackage(def *service.Resource) protoreflect.FullName {
	if m := cat.byMessage[protoreflect.FullName(def.Message)]; m != nil {
		return m.Desc.ParentFile().Package()
	}
	return cat.aip.files[def.Type]
}

// buildFields resolves every field of msg to its TypeScript shape.
func (cat *Catalog) buildFields(msg *Message, m *protogen.Message) {
	nameField := ""
	if r := cat.aip.messageResource(m.Desc); r != nil {
		msg.Resource = cat.resource(r.Type)
		nameField = r.NameField
	}
	oneofs := map[protoreflect.FullName]*Oneof{}
	for _, f := range m.Fields {
		fd := f.Desc
		field := cat.field(fd)
		field.Doc = fieldDoc(fd, comment(f.Comments))
		ir := cat.aip.field(fd)
		field.Required = isRequired(fd, ir)
		field.OutputOnly = ir != nil && ir.Has(service.BehaviorOutputOnly)
		if string(fd.Name()) == nameField && fd.Kind() == protoreflect.StringKind {
			field.Resource = msg.Resource
		}
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			o := oneofs[od.FullName()]
			if o == nil {
				o = &Oneof{Name: string(od.Name()), Doc: oneofDoc(m, od)}
				oneofs[od.FullName()] = o
				msg.Oneofs = append(msg.Oneofs, o)
			}
			o.Fields = append(o.Fields, field)
			field.Oneof = o
		}
		msg.Fields = append(msg.Fields, field)
	}
}

// oneofDoc is the comment on a oneof declaration.
func oneofDoc(m *protogen.Message, od protoreflect.OneofDescriptor) string {
	for _, o := range m.Oneofs {
		if o.Desc == od {
			return comment(o.Comments)
		}
	}
	return ""
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
		if ir := cat.aip.field(fd); ir != nil && ir.RefType != "" && ir.RefType != "*" && !ir.RefChild {
			f.Resource = cat.resource(ir.RefType)
		}
	}
	return f
}
