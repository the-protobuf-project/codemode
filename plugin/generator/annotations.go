// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/the-protobuf-project/codemode/protobuf/codemodepb"
	"github.com/the-protobuf-project/mcp/protobuf/mcppb"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ext reads extension xt from opts, or returns nil when it is absent.
func ext[T any](opts proto.Message, xt protoreflect.ExtensionType) T {
	var zero T
	if opts == nil || !proto.HasExtension(opts, xt) {
		return zero
	}
	v, _ := proto.GetExtension(opts, xt).(T)
	return v
}

func serviceOptions(sd protoreflect.ServiceDescriptor) (*codemodepb.CodemodeService, *mcppb.McpServiceOptions) {
	return ext[*codemodepb.CodemodeService](sd.Options(), codemodepb.E_Service),
		ext[*mcppb.McpServiceOptions](sd.Options(), mcppb.E_Service)
}

func methodOptions(md protoreflect.MethodDescriptor) (*codemodepb.CodemodeMethod, *mcppb.McpToolOptions, *mcppb.McpElicitation) {
	return ext[*codemodepb.CodemodeMethod](md.Options(), codemodepb.E_Method),
		ext[*mcppb.McpToolOptions](md.Options(), mcppb.E_Tool),
		ext[*mcppb.McpElicitation](md.Options(), mcppb.E_Elicitation)
}

// fieldBehaviors returns a field's google.api.field_behavior values.
func fieldBehaviors(fd protoreflect.FieldDescriptor) []annotations.FieldBehavior {
	return ext[[]annotations.FieldBehavior](fd.Options(), annotations.E_FieldBehavior)
}

func hasBehavior(fd protoreflect.FieldDescriptor, b annotations.FieldBehavior) bool {
	for _, got := range fieldBehaviors(fd) {
		if got == b {
			return true
		}
	}
	return false
}

// isRequired reports REQUIRED field_behavior or (buf.validate.field).required.
func isRequired(fd protoreflect.FieldDescriptor) bool {
	if hasBehavior(fd, annotations.FieldBehavior_REQUIRED) {
		return true
	}
	rules := ext[*validate.FieldRules](fd.Options(), validate.E_Field)
	return rules.GetRequired()
}

// fieldDoc is a field's mcp.v1 description, falling back to its comment.
func fieldDoc(fd protoreflect.FieldDescriptor, comment string) string {
	if o := ext[*mcppb.McpFieldOptions](fd.Options(), mcppb.E_Field); o.GetDescription() != "" {
		return o.GetDescription()
	}
	return comment
}

// enumDoc is an enum's mcp.v1 description, falling back to its comment.
func enumDoc(ed protoreflect.EnumDescriptor, comment string) string {
	if o := ext[*mcppb.McpEnumOptions](ed.Options(), mcppb.E_Enum); o.GetDescription() != "" {
		return o.GetDescription()
	}
	return comment
}

// enumValueDoc is an enum value's mcp.v1 description, falling back to its comment.
func enumValueDoc(vd protoreflect.EnumValueDescriptor, comment string) string {
	if o := ext[*mcppb.McpEnumValueOptions](vd.Options(), mcppb.E_EnumValue); o.GetDescription() != "" {
		return o.GetDescription()
	}
	return comment
}

// resourceOf returns a message's google.api.resource, or nil.
func resourceOf(md protoreflect.MessageDescriptor) *annotations.ResourceDescriptor {
	return ext[*annotations.ResourceDescriptor](md.Options(), annotations.E_Resource)
}

// fileResources returns a file's google.api.resource_definition entries.
func fileResources(fd protoreflect.FileDescriptor) []*annotations.ResourceDescriptor {
	return ext[[]*annotations.ResourceDescriptor](fd.Options(), annotations.E_ResourceDefinition)
}

// resourceReference returns the resource type a string field refers to.
func resourceReference(fd protoreflect.FieldDescriptor) string {
	return ext[*annotations.ResourceReference](fd.Options(), annotations.E_ResourceReference).GetType()
}

// httpRule returns a method's google.api.http binding, or nil.
func httpRule(md protoreflect.MethodDescriptor) *HTTPRule {
	r := ext[*annotations.HttpRule](md.Options(), annotations.E_Http)
	if r == nil {
		return nil
	}
	out := &HTTPRule{Body: r.GetBody(), ResponseBody: r.GetResponseBody()}
	switch p := r.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		out.Method, out.Path = "GET", p.Get
	case *annotations.HttpRule_Post:
		out.Method, out.Path = "POST", p.Post
	case *annotations.HttpRule_Put:
		out.Method, out.Path = "PUT", p.Put
	case *annotations.HttpRule_Patch:
		out.Method, out.Path = "PATCH", p.Patch
	case *annotations.HttpRule_Delete:
		out.Method, out.Path = "DELETE", p.Delete
	case *annotations.HttpRule_Custom:
		out.Method, out.Path = strings.ToUpper(p.Custom.GetKind()), p.Custom.GetPath()
	default:
		return nil
	}
	return out
}

// methodIdempotency returns the method's idempotency_level option name, or ""
// when unset.
func methodIdempotency(md protoreflect.MethodDescriptor) string {
	opts, _ := md.Options().(*descriptorpb.MethodOptions)
	if l := opts.GetIdempotencyLevel(); l != descriptorpb.MethodOptions_IDEMPOTENCY_UNKNOWN {
		return l.String()
	}
	return ""
}
