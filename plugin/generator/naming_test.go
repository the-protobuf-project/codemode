// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"
	"testing"
)

func TestMethodName(t *testing.T) {
	cases := []struct{ rpc, style, want string }{
		{"CreateBook", "lowerCamel", "createBook"},
		{"HTTPGet", "lowerCamel", "httpGet"},
		{"ListISBNs", "lowerCamel", "listISBNs"},
		{"CreateBook", "snake", "create_book"},
		{"Delete", "lowerCamel", "delete"},
	}
	for _, c := range cases {
		if got := methodName(c.rpc, c.style); got != c.want {
			t.Errorf("methodName(%q, %q) = %q, want %q", c.rpc, c.style, got, c.want)
		}
	}
}

func TestNamespaceName(t *testing.T) {
	cases := map[string]string{
		"LibraryService": "library",
		"Library":        "library",
		"DeleteService":  "delete_",
		"HTTPProxy":      "http_proxy",
	}
	for in, want := range cases {
		if got := namespaceName(in); got != want {
			t.Errorf("namespaceName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPropertyKey(t *testing.T) {
	for in, want := range map[string]string{"title": "title", "page2": "page2", "x-y": `"x-y"`, "2d": `"2d"`} {
		if got := propertyKey(in); got != want {
			t.Errorf("propertyKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJSDoc(t *testing.T) {
	if got := jsdoc("", ""); got != "" {
		t.Errorf("empty jsdoc = %q", got)
	}
	if got := jsdoc("  ", "One line."); got != "  /** One line. */\n" {
		t.Errorf("one-line jsdoc = %q", got)
	}
	got := jsdoc("", "Closes */ early.", "Required.")
	if strings.Contains(got, "*/ early") || !strings.Contains(got, " *\n * Required.") {
		t.Errorf("jsdoc did not escape or separate notes:\n%s", got)
	}
}

func TestTSStrings(t *testing.T) {
	if got := tsString("a\"b\n"); got != `"a\"b\n"` {
		t.Errorf("tsString = %s", got)
	}
	if got := tsTemplate("x `y` ${z}"); got != "`x \\`y\\` \\${z}`" {
		t.Errorf("tsTemplate = %s", got)
	}
}

func TestConfigDefaultsAndValidation(t *testing.T) {
	c, err := LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if c.Naming != "lowerCamel" || c.Discovery != "progressive" || c.Approval.Destructive != "require" || !c.InferFromAIP() {
		t.Errorf("unexpected defaults: %+v", c)
	}
	if !c.Selects("anything.v1.Service") {
		t.Error("an empty include list must select everything")
	}
	c.Include = []string{"library.v1.*"}
	c.Exclude = []string{"*.Internal*"}
	if !c.Selects("library.v1.LibraryService") || c.Selects("library.v1.InternalService") || c.Selects("other.v1.S") {
		t.Error("include/exclude patterns did not apply")
	}
	bad := &Config{Naming: "kebab"}
	if err := bad.defaults(); err == nil {
		t.Error("an unknown naming style must be rejected")
	}
	if _, err := Millis("soon"); err == nil {
		t.Error("a bad duration must be rejected")
	}
}
