// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

// Package templates holds the layout of every generated file. Templates are
// presentation only: the generator computes names, types and expressions and
// passes them in as view structs.
package templates

import "embed"

// FS holds the templates, addressed by file name.
//
//go:embed typescript/*.tpl
var FS embed.FS
