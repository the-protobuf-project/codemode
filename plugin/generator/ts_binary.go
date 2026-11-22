// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// writeElem returns the statement that writes one element v as field no.
func writeElem(f *tsFile, fd *Field, no int, v string) string {
	if fd.Message != nil || fd.WKT != "" {
		f.runtime("binary")
		return fmt.Sprintf("binary.message(w, %d, %s, %s)", no, codecOf(f, fd), v)
	}
	f.runtime("WireType")
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
	case fd.Message != nil || fd.WKT != "":
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

// encodeLines is the body of encode: the binary format, defaults left out.
// Blocks span several lines, indented relative to the method body.
func encodeLines(f *tsFile, m *Message) []string {
	var out []string
	for _, fd := range m.Fields {
		no := int(fd.Desc.Number())
		a := access(fd)
		switch {
		case fd.MapKey != nil:
			f.runtime("WireType")
			out = append(out,
				"for (const [k, v] of Object.entries("+a+" ?? {})) {",
				fmt.Sprintf("  w.tag(%d, WireType.LengthDelimited).fork();", no),
				"  "+writeElem(f, fd.MapKey, 1, mapKeyIn(fd.MapKey))+";",
				"  "+writeElem(f, fd.MapValue, 2, "v")+";",
				"  w.join();",
				"}")
		case fd.Desc.IsList() && fd.Desc.IsPacked() && isScalar(fd):
			f.runtime("WireType")
			out = append(out,
				"if ("+a+"?.length) {",
				fmt.Sprintf("  w.tag(%d, WireType.LengthDelimited).fork();", no),
				"  for (const v of "+a+") w."+scalars[fd.Desc.Kind()].Method+"("+binaryIn(f, fd, "v")+");",
				"  w.join();",
				"}")
		case fd.Desc.IsList():
			out = append(out, "for (const v of "+a+" ?? []) "+writeElem(f, fd, no, "v")+";")
		default:
			out = append(out, "if ("+presentCheck(f, fd)+") "+writeElem(f, fd, no, a)+";")
		}
	}
	return out
}

// decodeInit is the object decode starts from: defaults for fields without
// presence, and empty messages for required ones.
func decodeInit(f *tsFile, m *Message, n moduleNamer) []keyExpr {
	var out []keyExpr
	for _, fd := range m.Fields {
		key := propertyKey(fd.JSONName)
		switch {
		case fd.MapKey != nil:
			out = append(out, keyExpr{key, "{}"})
		case fd.Desc.IsList():
			out = append(out, keyExpr{key, "[]"})
		case !fd.Explicit && isScalar(fd):
			out = append(out, keyExpr{key, zero(f, fd, n)})
		case fd.Required && !isScalar(fd):
			out = append(out, keyExpr{key, mapValueZero(f, fd, n)})
		}
	}
	return out
}

// decodeCase is one field number's branch of decode's switch.
type decodeCase struct {
	No    int
	Lines []string
}

// decodeCases reads each field as it arrives; a oneof member clears the others.
func decodeCases(f *tsFile, m *Message, n moduleNamer) []decodeCase {
	var out []decodeCase
	for _, fd := range m.Fields {
		a := access(fd)
		c := decodeCase{No: int(fd.Desc.Number())}
		switch {
		case fd.MapKey != nil:
			f.runtime("binary")
			read := "binary.mapEntry(r, (x) => " + readElem(f, fd.MapKey, "x", n) + ", (x) => " + readElem(f, fd.MapValue, "x", n) +
				", " + zero(f, fd.MapKey, n) + ", () => " + mapValueZero(f, fd.MapValue, n) + ")"
			c.Lines = []string{"{", "  const [k, v] = " + read + ";", "  (" + a + " ??= {})[String(k)] = v;", "}"}
		case fd.Desc.IsList() && isScalar(fd) && scalars[fd.Desc.Kind()].Packable:
			f.runtime("binary")
			c.Lines = []string{"binary.packed(r, wt, (x) => " + readElem(f, fd, "x", n) + ", (" + a + " ??= []));"}
		case fd.Desc.IsList():
			c.Lines = []string{"(" + a + " ??= []).push(" + readElem(f, fd, "r", n) + ");"}
		default:
			c.Lines = []string{a + " = " + readElem(f, fd, "r", n) + ";"}
			if fd.Oneof != nil {
				for _, other := range fd.Oneof.Fields {
					if other != fd {
						c.Lines = append(c.Lines, access(other)+" = undefined;")
					}
				}
			}
		}
		out = append(out, c)
	}
	return out
}

// mapValueZero is the value a map entry gets when its value is missing.
func mapValueZero(f *tsFile, fd *Field, n typeNamer) string {
	if fd.Message != nil || fd.WKT != "" {
		return codecOf(f, fd) + ".decode(new Uint8Array(0))"
	}
	return zero(f, fd, n)
}
