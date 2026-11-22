// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "strings"

// jsdoc renders text and extra paragraphs as a JSDoc block at indent. It
// returns "" when there is nothing to say.
func jsdoc(indent, text string, notes ...string) string {
	var paras []string
	if t := strings.TrimSpace(text); t != "" {
		paras = append(paras, t)
	}
	if len(notes) > 0 {
		paras = append(paras, strings.Join(notes, "\n"))
	}
	if len(paras) == 0 {
		return ""
	}
	body := strings.ReplaceAll(strings.Join(paras, "\n\n"), "*/", "*\\/")
	lines := strings.Split(body, "\n")
	if len(lines) == 1 {
		return indent + "/** " + lines[0] + " */\n"
	}
	var b strings.Builder
	b.WriteString(indent + "/**\n")
	for _, l := range lines {
		l = strings.TrimRight(l, " ")
		if l == "" {
			b.WriteString(indent + " *\n")
		} else {
			b.WriteString(indent + " * " + l + "\n")
		}
	}
	b.WriteString(indent + " */\n")
	return b.String()
}

// tsString quotes s as a TypeScript string literal.
func tsString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`, " ", ` `, " ", ` `)
	return `"` + r.Replace(s) + `"`
}

// tsTemplate quotes s as a template literal, for long multi-line text.
func tsTemplate(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "`", "\\`", "${", "\\${")
	return "`" + r.Replace(s) + "`"
}
