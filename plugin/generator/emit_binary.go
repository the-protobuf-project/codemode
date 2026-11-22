// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// writeElem returns the statement that writes one element v as field no.
func writeElem(f *tsFile, fd *Field, no int, v string) string {
	if fd.Message != nil || fd.WKT != nil {
		f.runtime("binary")
		return fmt.Sprintf("binary.message(w, %d, %s, %s)", no, codecOf(f, fd), v)
	}
	s := scalars[fd.Desc.Kind()]
	return fmt.Sprintf("w.tag(%d, WireType.%s).%s(%s)", no, s.WireType, s.Method, binaryIn(f, fd, v))
}

// binaryIn converts a protojson-shaped scalar to what BinaryWriter takes.
func binaryIn(f *tsFile, fd *Field, v string) string {
	switch {
	case fd.Enum != nil:
		return codecOf(f, fd) + ".toNumber(" + v + ")"
	case fd.Desc.Kind() == protoreflect.BytesKind:
		f.runtime("binary")
		return "binary.bytesIn(" + v + ")"
	}
	return v
}

// readElem returns the expression that reads one element from reader r.
func readElem(f *tsFile, fd *Field, r string, n typeNamer) string {
	switch {
	case fd.Message != nil || fd.WKT != nil:
		f.runtime("binary")
		return "binary.readMessage(" + r + ", " + codecOf(f, fd) + ")"
	case fd.Enum != nil:
		return codecOf(f, fd) + ".fromNumber(" + r + ".int32())"
	case fd.Desc.Kind() == protoreflect.BytesKind:
		f.runtime("binary")
		return "binary.bytesOut(" + r + ".bytes())"
	case fd.Resource != nil:
		return r + ".string() as " + n.resource(fd.Resource)
	}
	return r + "." + scalars[fd.Desc.Kind()].Method + "()"
}

// mapKeyIn converts a map key, always a string in protojson, for the writer.
func mapKeyIn(fd *Field) string {
	switch k := fd.Desc.Kind(); {
	case k == protoreflect.BoolKind:
		return `k === "true"`
	case k == protoreflect.StringKind || is64Bit(k):
		return "k"
	}
	return "Number(k)"
}

// emitEncode writes encode: the binary format, defaults left out.
func emitEncode(f *tsFile, m *Message) {
	f.runtime("BinaryWriter")
	if len(m.Fields) == 0 {
		f.P("  encode(_m: ", m.Name, ", w: BinaryWriter = new BinaryWriter()): BinaryWriter {")
		f.P("    return w;")
		f.P("  },")
		return
	}
	f.runtime("WireType")
	f.P("  encode(m: ", m.Name, ", w: BinaryWriter = new BinaryWriter()): BinaryWriter {")
	for _, fd := range m.Fields {
		no := int(fd.Desc.Number())
		a := access(fd)
		switch {
		case fd.MapKey != nil:
			f.P("    for (const [k, v] of Object.entries(", a, " ?? {})) {")
			f.P("      w.tag(", no, ", WireType.LengthDelimited).fork();")
			f.P("      ", writeElem(f, fd.MapKey, 1, mapKeyIn(fd.MapKey)), ";")
			f.P("      ", writeElem(f, fd.MapValue, 2, "v"), ";")
			f.P("      w.join();")
			f.P("    }")
		case fd.Desc.IsList() && fd.Desc.IsPacked() && isScalar(fd):
			s := scalars[fd.Desc.Kind()]
			f.P("    if (", a, "?.length) {")
			f.P("      w.tag(", no, ", WireType.LengthDelimited).fork();")
			f.P("      for (const v of ", a, ") w.", s.Method, "(", binaryIn(f, fd, "v"), ");")
			f.P("      w.join();")
			f.P("    }")
		case fd.Desc.IsList():
			f.P("    for (const v of ", a, " ?? []) ", writeElem(f, fd, no, "v"), ";")
		default:
			f.P("    if (", presentCheck(f, fd), ") ", writeElem(f, fd, no, a), ";")
		}
	}
	f.P("    return w;")
	f.P("  },")
}

// emitDecode writes decode: defaults first, then each field as it arrives.
func emitDecode(f *tsFile, m *Message, n moduleNamer) {
	f.runtime("BinaryReader")
	f.P("  decode(input: Uint8Array | BinaryReader, length?: number): ", m.Name, " {")
	f.P("    const r = input instanceof BinaryReader ? input : new BinaryReader(input);")
	f.P("    const end = r.end(length);")
	f.P("    const m: ", m.Name, " = {")
	for _, fd := range m.Fields {
		switch {
		case fd.MapKey != nil:
			f.P("      ", propertyKey(fd.JSONName), ": {},")
		case fd.Desc.IsList():
			f.P("      ", propertyKey(fd.JSONName), ": [],")
		case !fd.Explicit && isScalar(fd):
			f.P("      ", propertyKey(fd.JSONName), ": ", zero(f, fd, n), ",")
		case fd.Required && !isScalar(fd):
			f.P("      ", propertyKey(fd.JSONName), ": ", mapValueZero(f, fd, n), ",")
		}
	}
	f.P("    };")
	f.P("    while (r.pos < end) {")
	f.P("      const [no, wt] = r.tag();")
	f.P("      switch (no) {")
	for _, fd := range m.Fields {
		a := access(fd)
		f.P("        case ", fd.Desc.Number(), ":")
		switch {
		case fd.MapKey != nil:
			f.runtime("binary")
			read := "binary.mapEntry(r, (x) => " + readElem(f, fd.MapKey, "x", n) + ", (x) => " + readElem(f, fd.MapValue, "x", n) +
				", " + zero(f, fd.MapKey, n) + ", () => " + mapValueZero(f, fd.MapValue, n) + ")"
			f.P("          {")
			f.P("            const [k, v] = ", read, ";")
			f.P("            (", a, " ??= {})[String(k)] = v;")
			f.P("          }")
		case fd.Desc.IsList() && isScalar(fd) && scalars[fd.Desc.Kind()].Packable:
			f.runtime("binary")
			f.P("          binary.packed(r, wt, (x) => ", readElem(f, fd, "x", n), ", (", a, " ??= []));")
		case fd.Desc.IsList():
			f.P("          (", a, " ??= []).push(", readElem(f, fd, "r", n), ");")
		default:
			f.P("          ", a, " = ", readElem(f, fd, "r", n), ";")
			if fd.Oneof != nil {
				for _, other := range fd.Oneof.Fields {
					if other != fd {
						f.P("          ", access(other), " = undefined;")
					}
				}
			}
		}
		f.P("          break;")
	}
	f.P("        default:")
	f.P("          r.skip(wt, no);")
	f.P("      }")
	f.P("    }")
	f.P("    return m;")
	f.P("  },")
}

// mapValueZero is the value a map entry gets when its value is missing.
func mapValueZero(f *tsFile, fd *Field, n typeNamer) string {
	if fd.Message != nil || fd.WKT != nil {
		return codecOf(f, fd) + ".decode(new Uint8Array(0))"
	}
	return zero(f, fd, n)
}
