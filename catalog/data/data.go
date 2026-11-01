// Package data contains the reviewed Provider maintenance records.
package data

import "embed"

// Providers contains immutable inputs. ReadFile returns an independent slice.
// Runtime ai embeds only the generated artifact and does not import this package.
//
//go:embed providers/*.json
var Providers embed.FS
