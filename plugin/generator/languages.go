// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// LangAll requests every language the plugin emits.
const LangAll = "all"

// Language renders the catalog as one host language's codebase: typed
// messages, clients, connectors and the runtime wiring. The model it reads is
// language-neutral; each language keeps its emitters in <lang>_*.go files and
// its layouts in templates/<lang>/.
//
// Whatever the host language, sandboxed code is JavaScript, so every language
// describes the sandbox API with the same TypeScript declarations.
type Language interface {
	// Name is the value of the lang plugin option, e.g. "typescript".
	Name() string
	// Aliases are other accepted spellings, e.g. "ts".
	Aliases() []string
	Generate(gen *protogen.Plugin, cat *Catalog) error
}

// languages lists every language the plugin emits. A new host language,
// say Go for a Go Code Mode runtime, adds one entry here.
var languages = []Language{typescript{}}

// languageNames lists the supported languages, for errors and the factory.
func languageNames() []string {
	names := make([]string, len(languages))
	for i, l := range languages {
		names[i] = l.Name()
	}
	return names
}

// resolveLanguages turns the lang option, a comma-separated list or "all",
// into languages. Empty means the default, TypeScript.
func resolveLanguages(lang string) ([]Language, error) {
	if lang == "" {
		lang = "typescript"
	}
	if lang == LangAll {
		return languages, nil
	}
	var out []Language
	for _, want := range strings.Split(lang, ",") {
		want = strings.TrimSpace(want)
		l := findLanguage(want)
		if l == nil {
			return nil, fmt.Errorf("unsupported language %q (supported: %s, %s)", want, strings.Join(languageNames(), ", "), LangAll)
		}
		out = append(out, l)
	}
	return out, nil
}

func findLanguage(name string) Language {
	for _, l := range languages {
		if l.Name() == name {
			return l
		}
		for _, a := range l.Aliases() {
			if a == name {
				return l
			}
		}
	}
	return nil
}
