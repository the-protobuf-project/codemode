// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/the-protobuf-project/codemode/protobuf/codemodepb"
)

// resolveRevert checks a codemode.v1.method.revert against the service and
// resolves its field paths to JSON names, so the generated revert compiles.
func (cat *Catalog) resolveRevert(svc *Service, m *Method, r *codemodepb.Revert) error {
	if r == nil {
		return nil
	}
	if m.Streaming != "" {
		return fmt.Errorf("%s is a streaming method; a revert undoes one unary call", m.Proto.Desc.FullName())
	}
	var target *Method
	for _, other := range svc.Methods {
		if string(other.Proto.Desc.Name()) == r.GetMethod() {
			target = other
		}
	}
	if target == nil {
		return fmt.Errorf("revert method %q is not a unary method of %s", r.GetMethod(), svc.Name)
	}
	if target.Streaming != "" {
		return fmt.Errorf("revert method %s is a streaming method", r.GetMethod())
	}
	if target.Approval {
		return fmt.Errorf("revert method %s requires approval; rollback cannot pause", r.GetMethod())
	}
	if target.Input.Message == nil {
		return fmt.Errorf("revert method %s must take a generated message", r.GetMethod())
	}
	keys := make([]string, 0, len(r.GetFields()))
	for k := range r.GetFields() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rev := &Revert{Method: target}
	for _, key := range keys {
		field := fieldByProtoName(target.Input.Message, key)
		if field == nil {
			return fmt.Errorf("revert field %q is not in %s", key, target.Input.Message.Desc.FullName())
		}
		src := r.GetFields()[key]
		parts := strings.Split(src, ".")
		var from *Message
		switch parts[0] {
		case "args":
			from = m.Input.Message
		case "result":
			from = m.Output.Message
		default:
			return fmt.Errorf("revert source %q must start with args. or result.", src)
		}
		path, err := jsonPath(from, parts[1:], src)
		if err != nil {
			return err
		}
		rev.Fields = append(rev.Fields, RevertField{Field: field, Source: parts[0], Path: path})
	}
	m.Revert = rev
	return nil
}

// jsonPath maps proto field names along a path to JSON names.
func jsonPath(msg *Message, parts []string, src string) ([]string, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("revert source %q names no field", src)
	}
	var out []string
	for i, part := range parts {
		if msg == nil {
			return nil, fmt.Errorf("revert source %q goes through a non-message field", src)
		}
		f := fieldByProtoName(msg, part)
		if f == nil {
			return nil, fmt.Errorf("revert source %q: %s has no field %q", src, msg.Desc.FullName(), part)
		}
		out = append(out, f.JSONName)
		if i < len(parts)-1 {
			msg = f.Message
		}
	}
	return out, nil
}

func fieldByProtoName(msg *Message, name string) *Field {
	for _, f := range msg.Fields {
		if f.ProtoName == name {
			return f
		}
	}
	return nil
}
