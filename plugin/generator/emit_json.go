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

// emitFromJSON writes fromJson: validate protojson, fill defaults.
func emitFromJSON(f *tsFile, m *Message, n moduleNamer) {
	f.runtime("protojson")
	f.P("  fromJson(json: unknown, path: string = ", tsString(string(m.Desc.FullName())), "): ", m.Name, " {")
	if len(m.Fields) == 0 {
		f.P("    protojson.object(json, path);")
		f.P("    return {};")
		f.P("  },")
		return
	}
	f.P("    const o = protojson.object(json, path);")
	f.P("    const m: ", m.Name, " = {")
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
		expr := reader + "(protojson.get(o, " + tsString(fd.JSONName) + ", " + tsString(fd.ProtoName) + "), path + " + tsString("."+fd.JSONName) + ")"
		switch {
		case fd.MapKey != nil:
			expr += " ?? {}"
		case fd.Desc.IsList():
			expr += " ?? []"
		case !fd.Explicit && isScalar(fd):
			expr += " ?? " + zero(f, fd, n)
		case fd.Required && !isScalar(fd):
			expr += " ?? protojson.missing(path + " + tsString("."+fd.JSONName) + ")"
		}
		if fd.Resource != nil && fd.MapKey == nil {
			cast := tsType(fd, n)
			if fd.Explicit {
				cast += " | undefined"
			}
			expr = "(" + expr + ") as " + cast
		}
		f.P("      ", propertyKey(fd.JSONName), ": ", expr, ",")
	}
	f.P("    };")
	for _, o := range m.Oneofs {
		names := make([]string, len(o.Fields))
		for i, of := range o.Fields {
			names[i] = tsString(of.JSONName)
		}
		f.P("    protojson.oneof(m, path, ", tsString(o.Name), ", [", strings.Join(names, ", "), "]);")
	}
	f.P("    return m;")
	f.P("  },")
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

// emitToJSON writes toJson: protojson with default values left out, or, with
// emitDefaults, written out, which is what sandboxed code receives.
func emitToJSON(f *tsFile, m *Message) {
	f.runtimeType("JsonObject", "ToJsonOptions")
	if len(m.Fields) == 0 {
		f.P("  toJson(): JsonObject {")
		f.P("    return {};")
		f.P("  },")
		return
	}
	f.P("  toJson(m: ", m.Name, ", options?: ToJsonOptions): JsonObject {")
	if usesDefaults(m) {
		f.P("    const all = options?.emitDefaults === true;")
	}
	f.P("    const o: JsonObject = {};")
	for _, fd := range m.Fields {
		a := access(fd)
		key := "o[" + tsString(fd.JSONName) + "]"
		switch {
		case fd.MapKey != nil:
			f.P("    if (all || protojson.nonEmpty(", a, ")) ", key, " = protojson.mapOut(", a, " ?? {}, (v) => ", jsonOut(f, fd.MapValue, "v"), ");")
		case fd.Desc.IsList():
			out := a + " ?? []"
			if conv := jsonOut(f, fd, "v"); conv != "v" {
				out = "(" + out + ").map((v) => " + conv + ")"
			}
			f.P("    if (all || ", a, "?.length) ", key, " = ", out, ";")
		case !fd.Explicit && isScalar(fd):
			cond := presentCheck(f, fd)
			if strings.Contains(cond, " ") {
				cond = "(" + cond + ")"
			}
			value := a + " ?? " + zero(f, fd, moduleNamer{f})
			if conv := jsonOut(f, fd, "v"); conv != "v" {
				value = jsonOut(f, fd, value)
			}
			f.P("    if (all || ", cond, ") ", key, " = ", value, ";")
		default:
			f.P("    if (", presentCheck(f, fd), ") ", key, " = ", jsonOut(f, fd, a), ";")
		}
	}
	f.P("    return o;")
	f.P("  },")
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
