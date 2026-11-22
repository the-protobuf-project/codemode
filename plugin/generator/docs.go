// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// comment is the documentation protoc kept for a declaration: its leading
// comment, then its trailing one, as plain text without comment markers.
func comment(c protogen.CommentSet) string {
	return joinDocs(commentText(string(c.Leading)), commentText(string(c.Trailing)))
}

// crossRef matches AIP-192 references: [Book][library.v1.Book] or [Book][].
var crossRef = regexp.MustCompile(`\[([^\[\]]+)\]\[[A-Za-z0-9_.]*\]`)

// commentText dedents a comment's lines, drops surrounding blank lines and
// turns AIP-192 references into their display text.
func commentText(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	indent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if n := len(l) - len(strings.TrimLeft(l, " ")); indent < 0 || n < indent {
			indent = n
		}
	}
	for i, l := range lines {
		if len(l) >= indent && indent > 0 {
			l = l[indent:]
		}
		lines[i] = strings.TrimRight(l, " \t")
	}
	text := strings.TrimSpace(strings.Join(lines, "\n"))
	return crossRef.ReplaceAllString(text, "$1")
}

// joinDocs joins non-empty, distinct paragraphs: a proto comment and an
// annotation's description both reach the generated docs.
func joinDocs(docs ...string) string {
	var out []string
	for _, d := range docs {
		d = strings.TrimSpace(d)
		if d == "" || contains(out, d) {
			continue
		}
		out = append(out, d)
	}
	return strings.Join(out, "\n\n")
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// isDeprecated reports the deprecated option on any declaration.
func isDeprecated(d protoreflect.Descriptor) bool {
	o, ok := d.Options().(interface{ GetDeprecated() bool })
	return ok && o.GetDeprecated()
}

// packageDoc joins the comments above `package` in each generated file of a
// package, in file order.
func packageDoc(files []*protogen.File, pkg protoreflect.FullName) string {
	var docs []string
	for _, f := range files {
		if !f.Generate || f.Desc.Package() != pkg {
			continue
		}
		loc := f.Desc.SourceLocations().ByPath(protoreflect.SourcePath{2}) // FileDescriptorProto.package
		docs = append(docs, commentText(loc.LeadingComments))
	}
	return joinDocs(docs...)
}
