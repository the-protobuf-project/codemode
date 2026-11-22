// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "strings"

// jsonReader is the protojson reader for one element of a field.
func jsonReader(f *tsFile, fd *Field) string {
	switch {
	case fd.Enum != nil:
		return "protojson.enumOf(" + codecOf(f, fd) + ")"
	case fd.Message != nil || fd.WKT != nil:
		return "protojson.message(" + codecOf(f, fd) + ")"
	}
	return "protojson." + scalars[fd.Desc.Kind()].JSON
}

// fromJSONFields is each property of the object fromJson builds: read the
// protojson value, then fill the default or require it.
func fromJSONFields(f *tsFile, m *Message, n moduleNamer) []keyExpr {
	var out []keyExpr
	for _, fd := range m.Fields {
		var reader string
		switch {
		case fd.MapKey != nil:
			reader = "protojson.map(" + jsonReader(f, fd.MapValue) + ")"
		case fd.Desc.IsList():
			reader = "protojson.list(" + jsonReader(f, fd) + ")"
		default:
			reader = jsonReader(f, fd)
		}
		path := "path + " + tsString("."+fd.JSONName)
		expr := reader + "(protojson.get(o, " + tsString(fd.JSONName) + ", " + tsString(fd.ProtoName) + "), " + path + ")"
		switch {
		case fd.MapKey != nil:
			expr += " ?? {}"
		case fd.Desc.IsList():
			expr += " ?? []"
		case !fd.Explicit && isScalar(fd):
			expr += " ?? " + zero(f, fd, n)
		case fd.Required && !isScalar(fd):
			expr += " ?? protojson.missing(" + path + ")"
		}
		if fd.Resource != nil && fd.MapKey == nil {
			cast := tsType(fd, n)
			if fd.Explicit {
				cast += " | undefined"
			}
			expr = "(" + expr + ") as " + cast
		}
		out = append(out, keyExpr{propertyKey(fd.JSONName), expr})
	}
	return out
}

// oneofChecks lists each oneof and its member names, for protojson.oneof.
func oneofChecks(m *Message) []struct{ Name, Fields string } {
	var out []struct{ Name, Fields string }
	for _, o := range m.Oneofs {
		names := make([]string, len(o.Fields))
		for i, of := range o.Fields {
			names[i] = tsString(of.JSONName)
		}
		out = append(out, struct{ Name, Fields string }{o.Name, strings.Join(names, ", ")})
	}
	return out
}

// jsonOut converts one element value v to protojson.
func jsonOut(f *tsFile, fd *Field, v string) string {
	switch {
	case fd.Message != nil || fd.WKT != nil:
		return codecOf(f, fd) + ".toJson(" + v + ", options)"
	case scalars[fd.Desc.Kind()].Float:
		return "protojson.floatOut(" + v + ")"
	}
	return v
}

// toJSONLines is one statement per field: protojson with default values left
// out or, with emitDefaults, written out, which is what sandboxed code gets.
func toJSONLines(f *tsFile, m *Message) []string {
	var out []string
	for _, fd := range m.Fields {
		a := access(fd)
		key := "o[" + tsString(fd.JSONName) + "]"
		switch {
		case fd.MapKey != nil:
			out = append(out, "if (all || protojson.nonEmpty("+a+")) "+key+" = protojson.mapOut("+a+" ?? {}, (v) => "+jsonOut(f, fd.MapValue, "v")+");")
		case fd.Desc.IsList():
			value := a + " ?? []"
			if conv := jsonOut(f, fd, "v"); conv != "v" {
				value = "(" + value + ").map((v) => " + conv + ")"
			}
			out = append(out, "if (all || "+a+"?.length) "+key+" = "+value+";")
		case !fd.Explicit && isScalar(fd):
			cond := presentCheck(f, fd)
			if strings.Contains(cond, " ") {
				cond = "(" + cond + ")"
			}
			value := a + " ?? " + zero(f, fd, moduleNamer{f})
			if conv := jsonOut(f, fd, "v"); conv != "v" {
				value = jsonOut(f, fd, value)
			}
			out = append(out, "if (all || "+cond+") "+key+" = "+value+";")
		default:
			out = append(out, "if ("+presentCheck(f, fd)+") "+key+" = "+jsonOut(f, fd, a)+";")
		}
	}
	return out
}

// usesDefaults reports whether toJson writes a field that emitDefaults affects:
// a map, a list, or a scalar without presence.
func usesDefaults(m *Message) bool {
	for _, fd := range m.Fields {
		if fd.MapKey != nil || fd.Desc.IsList() || (!fd.Explicit && isScalar(fd)) {
			return true
		}
	}
	return false
}
