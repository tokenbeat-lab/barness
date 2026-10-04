package ai

import "strings"

// argumentText is a tool call's argument JSON received so far (pi's
// partialJson). Appending grows one buffer instead of copying the whole text
// on every delta, so a call costs time linear in its argument size where
// `raw += delta` costs its square (issue 32). Bytes once written never
// change, so every text it returned stays valid. Share it by pointer: two
// copies appending to one buffer would overwrite each other's texts.
type argumentText struct{ b strings.Builder }

func newArgumentText(initial string) *argumentText {
	t := &argumentText{}
	t.b.WriteString(initial)
	return t
}

// append adds delta and returns the whole text so far.
func (t *argumentText) append(delta string) string {
	t.b.WriteString(delta)
	return t.b.String()
}

// String is the text so far.
func (t *argumentText) String() string { return t.b.String() }

// Len is the text's size in bytes.
func (t *argumentText) Len() int { return t.b.Len() }
