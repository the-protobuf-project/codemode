// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"path"
	"strings"

	"github.com/the-protobuf-project/protokit/naming"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// reserved lists JavaScript words that cannot name a sandbox global.
var reserved = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true, "continue": true,
	"debugger": true, "default": true, "delete": true, "do": true, "else": true, "enum": true,
	"export": true, "extends": true, "false": true, "finally": true, "for": true, "function": true,
	"if": true, "import": true, "in": true, "instanceof": true, "new": true, "null": true,
	"return": true, "super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true, "with": true,
	"yield": true, "let": true, "static": true, "implements": true, "interface": true,
	"package": true, "private": true, "protected": true, "public": true, "await": true,
	"codemode": true, "console": true, "globalThis": true, "undefined": true,
}

// packageDir maps a proto package to its output directory: "library.v1" ->
// "library/v1".
func packageDir(pkg protoreflect.FullName) string {
	return strings.ReplaceAll(string(pkg), ".", "/")
}

// typeName flattens a message or enum's nesting into one PascalCase name:
// "library.v1.Book.Edition" -> "BookEdition".
func typeName(desc protoreflect.Descriptor) string {
	pkg := string(desc.ParentFile().Package())
	local := strings.TrimPrefix(string(desc.FullName()), pkg+".")
	return strings.ReplaceAll(local, ".", "")
}

// fileStem is a type's snake_case file name: "BookEdition" -> "book_edition".
func fileStem(name string) string {
	return naming.SnakeCase(name)
}

// methodName is an RPC's sandbox method name in the configured style.
func methodName(rpc string, style string) string {
	if style == "snake" {
		return naming.SnakeCase(rpc)
	}
	return lowerFirstWord(rpc)
}

// lowerFirstWord lowercases a leading acronym or word: "CreateBook" ->
// "createBook", "HTTPGet" -> "httpGet".
func lowerFirstWord(s string) string {
	r := []rune(s)
	n := 0
	for n < len(r) && r[n] >= 'A' && r[n] <= 'Z' {
		n++
	}
	if n > 1 && n < len(r) {
		n--
	}
	return strings.ToLower(string(r[:n])) + string(r[n:])
}

// namespaceName is a service's default sandbox global: "LibraryService" ->
// "library".
func namespaceName(service string) string {
	n := strings.TrimSuffix(naming.SnakeCase(service), "_service")
	return safeIdent(n)
}

// safeIdent makes s a usable global by suffixing reserved words.
func safeIdent(s string) string {
	s = naming.SanitizeIdent(s)
	if reserved[s] {
		return s + "_"
	}
	return s
}

// pascal converts a snake_case or dotted name to PascalCase: "library.v1" ->
// "LibraryV1".
func pascal(s string) string {
	var b strings.Builder
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == '.' || r == '_' }) {
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	return b.String()
}

// relImport returns the import specifier for target as seen from the file
// from, both relative to the output root, with ext replacing ".ts".
func relImport(from, target, ext string) string {
	rel, err := relPath(path.Dir(from), target)
	if err != nil {
		rel = target
	}
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return strings.TrimSuffix(rel, ".ts") + ext
}

func relPath(fromDir, target string) (string, error) {
	from := strings.Split(path.Clean(fromDir), "/")
	to := strings.Split(path.Clean(target), "/")
	if fromDir == "." {
		from = nil
	}
	i := 0
	for i < len(from) && i < len(to)-1 && from[i] == to[i] {
		i++
	}
	up := strings.Repeat("../", len(from)-i)
	return up + strings.Join(to[i:], "/"), nil
}

// propertyKey quotes a property name that is not a plain identifier.
func propertyKey(name string) string {
	for i, r := range name {
		ok := r == '_' || r == '$' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if !ok {
			return `"` + name + `"`
		}
	}
	return name
}
