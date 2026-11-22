// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"strings"
	"testing"
)

func TestResolveLanguages(t *testing.T) {
	for _, lang := range []string{"", "typescript", "ts", "all", "typescript, ts"} {
		got, err := resolveLanguages(lang)
		if err != nil || len(got) == 0 || got[0].Name() != "typescript" {
			t.Errorf("resolveLanguages(%q) = %v, %v; want typescript", lang, got, err)
		}
	}
	_, err := resolveLanguages("go")
	if err == nil || !strings.Contains(err.Error(), `unsupported language "go" (supported: typescript, all)`) {
		t.Errorf("resolveLanguages(go) error = %v", err)
	}
}
