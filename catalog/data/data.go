// Package data contains the reviewed Provider maintenance records.
package data

import "embed"

// Providers contains immutable inputs. ReadFile returns an independent slice.
// Runtime ai embeds only the generated artifact and does not import this package.
//
//go:embed providers/*.json
var Providers embed.FS

// Baseline retains an adopted review bundle for future maintenance exports.
// The README is a sentinel for the initial historical migration only.
//
//go:embed all:baseline
var Baseline embed.FS
