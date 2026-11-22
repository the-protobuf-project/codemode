// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigOverrides(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "codemode.yaml")
	yaml := `version: v1
services:
  library.v1.LibraryService:
    namespace: books
    instructions: Ask before lending.
methods:
  library.v1.LibraryService.GetBook:
    approval: required
  library.v1.LibraryService.ListShelves:
    exclude: true
  library.v1.LibraryService.LendBook:
    revert:
      method: ReturnBook
      fields: {name: args.name}
`
	if err := os.WriteFile(cfg, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	res := generateLibrary(t, "lang=typescript,config="+cfg)
	var sandbox string
	for _, f := range res.GetFile() {
		if f.GetName() == "codemode/sandbox.d.ts" {
			sandbox = f.GetContent()
		}
	}
	for _, want := range []string{
		"declare const books: {",
		"Ask before lending.",
		"Undone on rollback by books.returnBook.",
	} {
		if !strings.Contains(sandbox, want) {
			t.Errorf("sandbox.d.ts lacks %q", want)
		}
	}
	if strings.Contains(sandbox, "listShelves(") {
		t.Error("listShelves is excluded but still in sandbox.d.ts")
	}
	get := sandbox[strings.Index(sandbox, "Gets a book."):]
	if !strings.Contains(get[:strings.Index(get, "getBook(")], "Requires approval") {
		t.Error("getBook should require approval")
	}
}

func TestConfigOverridesRejectTypos(t *testing.T) {
	for name, c := range map[string]*Config{
		"unknown method":  {Methods: map[string]MethodOverride{"library.v1.LibraryService.GetBok": {}}},
		"unknown service": {Services: map[string]ServiceOverride{"library.v1.Library": {}}},
		"bad approval":    {Methods: map[string]MethodOverride{"a.B.C": {Approval: "maybe"}}},
	} {
		if err := c.checkOverrides(nil); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
