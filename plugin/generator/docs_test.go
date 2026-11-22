// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "testing"

func TestCommentText(t *testing.T) {
	cases := map[string]string{
		" One line.\n":                        "One line.",
		" First.\n\n Second,\n   indented.\n": "First.\n\nSecond,\n  indented.",
		" See [Book][library.v1.Book].\n":     "See Book.",
		" See [the book][].\n":                "See the book.",
		"\n Leading blank line.\n\n":          "Leading blank line.",
		" Keeps [brackets] and a[0] alone.\n": "Keeps [brackets] and a[0] alone.",
	}
	for in, want := range cases {
		if got := commentText(in); got != want {
			t.Errorf("commentText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinDocs(t *testing.T) {
	if got := joinDocs("Gets a book.", "", "Gets a book.", "Includes the borrower."); got != "Gets a book.\n\nIncludes the borrower." {
		t.Errorf("joinDocs = %q", got)
	}
}
