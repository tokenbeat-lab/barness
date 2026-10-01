package ai

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

// SystemMessage changes the system instructions or the tool set at its place
// in the transcript (pi-ai's system message). The first one is the base
// prompt; each later one appends instructions, replaces or removes named
// sections and adds or removes tools from that point on. Request.SystemPrompt
// and Request.Tools are a shorthand for a leading SystemMessage.
//
// How later messages reach the wire depends on the target: a model that
// accepts system messages mid-conversation gets each one in place; any other
// gets one leading instruction replaying them all, in input order. Either way
// the request declares the tool set current after the last change.
type SystemMessage struct {
	// Content is instruction text: the base prompt on the leading message,
	// additional instructions on a later one.
	Content string `json:"content"`
	// Sections are named prompt parts rendered verbatim after Content. A
	// later message replaces a section by name or removes it. Order follows
	// the slice, except that names which are array indices ("0", "17") come
	// first in ascending order, as JavaScript object keys do in pi-ai; keep
	// names non-numeric to avoid surprises.
	Sections []SystemSection `json:"sections,omitempty"`
	// ToolsAdded are complete declarations that become available here. A
	// name already declared is redefined in its original position.
	ToolsAdded []Tool `json:"toolsAdded,omitempty"`
	// ToolsRemoved names tools that stop being available here. Removals of a
	// message apply before its additions.
	ToolsRemoved []string `json:"toolsRemoved,omitempty"`
}

func (SystemMessage) isMessage() {}

// SystemSection is one named prompt section of a SystemMessage.
type SystemSection struct {
	Name string `json:"name"`
	Text string `json:"text,omitempty"`
	// Removed drops the section (pi-ai's null value); Text must be empty.
	Removed bool `json:"removed,omitempty"`
}

func (m SystemMessage) clone() SystemMessage {
	m.Sections = slices.Clone(m.Sections)
	m.ToolsRemoved = slices.Clone(m.ToolsRemoved)
	if m.ToolsAdded != nil {
		tools := make([]Tool, len(m.ToolsAdded))
		for i, t := range m.ToolsAdded {
			tools[i] = t.clone()
		}
		m.ToolsAdded = tools
	}
	return m
}

// promptText renders the message as a complete prompt: its content, then
// each section's text (pi getSystemMessageText).
func (m SystemMessage) promptText() string {
	parts := []string{m.Content}
	for _, s := range jsOrderSections(m.Sections) {
		if !s.Removed {
			parts = append(parts, s.Text)
		}
	}
	return joinNonEmpty(parts, "\n\n")
}

// updateText renders a later message for a target that takes system messages
// mid-conversation, framing each section change by name (pi
// renderSystemMessageUpdate).
func (m SystemMessage) updateText() string {
	var parts []string
	if m.Content != "" {
		parts = append(parts, m.Content)
	}
	for _, s := range jsOrderSections(m.Sections) {
		if s.Removed {
			parts = append(parts, `Removed system prompt section "`+s.Name+`".`)
		} else {
			parts = append(parts, `Updated system prompt section "`+s.Name+`":`+"\n\n"+s.Text)
		}
	}
	return strings.Join(parts, "\n\n")
}

// normalizeTranscript folds the request's SystemPrompt and Tools into a
// leading SystemMessage ahead of its messages (pi normalizeContext). The
// result shares the request's messages; nothing in it is modified later.
func normalizeTranscript(req Request) []Message {
	if req.SystemPrompt == "" && len(req.Tools) == 0 {
		return req.Messages
	}
	head := SystemMessage{Content: req.SystemPrompt, ToolsAdded: req.Tools}
	return append([]Message{head}, req.Messages...)
}

// collapseSystemMessages replays every system message into one leading
// message holding the current prompt, sections and tools, and drops the rest
// (pi collapseSystemMessages). Without system messages it changes nothing.
func collapseSystemMessages(msgs []Message) []Message {
	var contents []string
	var sections orderedSections
	found := false
	rest := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		sm, ok := m.(SystemMessage)
		if !ok {
			rest = append(rest, m)
			continue
		}
		found = true
		if sm.Content != "" {
			contents = append(contents, sm.Content)
		}
		for _, s := range jsOrderSections(sm.Sections) {
			if s.Removed {
				sections.remove(s.Name)
			} else {
				sections.set(s.Name, s.Text)
			}
		}
	}
	if !found {
		return msgs
	}
	head := SystemMessage{Content: strings.Join(contents, "\n\n"), Sections: jsOrderSections(sections.list()), ToolsAdded: currentTools(msgs)}
	return append([]Message{head}, rest...)
}

// currentTools is the tool set after applying every system message's changes
// in order (pi getCurrentTools): a redefinition keeps the tool's position, a
// removal followed by an addition moves it to the end.
func currentTools(msgs []Message) []Tool {
	var names []string
	byName := map[string]Tool{}
	for _, m := range msgs {
		sm, ok := m.(SystemMessage)
		if !ok {
			continue
		}
		for _, name := range sm.ToolsRemoved {
			if _, ok := byName[name]; ok {
				delete(byName, name)
				names = slices.DeleteFunc(names, func(n string) bool { return n == name })
			}
		}
		for _, t := range sm.ToolsAdded {
			if _, ok := byName[t.Name]; !ok {
				names = append(names, t.Name)
			}
			byName[t.Name] = t
		}
	}
	tools := make([]Tool, len(names))
	for i, n := range names {
		tools[i] = byName[n]
	}
	return tools
}

// orderedSections is pi's insertion-ordered section Map.
type orderedSections struct {
	names []string
	text  map[string]string
}

func (s *orderedSections) set(name, text string) {
	if s.text == nil {
		s.text = map[string]string{}
	}
	if _, ok := s.text[name]; !ok {
		s.names = append(s.names, name)
	}
	s.text[name] = text
}

func (s *orderedSections) remove(name string) {
	if _, ok := s.text[name]; ok {
		delete(s.text, name)
		s.names = slices.DeleteFunc(s.names, func(n string) bool { return n == name })
	}
}

func (s *orderedSections) list() []SystemSection {
	out := make([]SystemSection, len(s.names))
	for i, n := range s.names {
		out[i] = SystemSection{Name: n, Text: s.text[n]}
	}
	return out
}

// jsOrderSections orders sections as JavaScript orders the keys of the object
// pi-ai holds them in: array-index names ascending, then the rest in their
// own order. Section names are unique per message (Request.validate).
func jsOrderSections(sections []SystemSection) []SystemSection {
	out := slices.Clone(sections)
	slices.SortStableFunc(out, func(a, b SystemSection) int {
		ai, aIndex := jsArrayIndex(a.Name)
		bi, bIndex := jsArrayIndex(b.Name)
		switch {
		case aIndex && bIndex:
			return cmp.Compare(ai, bi)
		case aIndex:
			return -1
		case bIndex:
			return 1
		}
		return 0
	})
	return out
}

// jsArrayIndex reports whether name is a canonical array index, the property
// keys JavaScript orders numerically before all others.
func jsArrayIndex(name string) (int64, bool) {
	if name == "" || (len(name) > 1 && name[0] == '0') {
		return 0, false
	}
	for _, c := range name {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(name, 10, 64)
	return n, err == nil && n < 1<<32-1
}

func joinNonEmpty(parts []string, sep string) string {
	return strings.Join(slices.DeleteFunc(parts, func(p string) bool { return p == "" }), sep)
}
