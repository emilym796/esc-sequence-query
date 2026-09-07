package main

import (
	"fmt"
	"strings"
)

// ParseError describes a point at which the input violated the escape
// sequence grammar. Raw holds the bytes of the sequence read so far,
// starting with the ESC byte. CaretIndex points into Raw at the byte
// that caused the problem, or equals len(Raw) if the problem is that
// input ran out before the sequence was complete.
type ParseError struct {
	Pos        Position
	Raw        []byte
	CaretIndex int
	Message    string
}

func (e *ParseError) Error() string {
	vis, caretCol := visualize(e.Raw, e.CaretIndex)
	return fmt.Sprintf(
		"escq: %s\n  at line %d, column %d\n\n    %s\n    %s^",
		e.Message, e.Pos.Line, e.Pos.Col, vis, strings.Repeat(" ", caretCol),
	)
}

// visualize renders raw as a printable string, escaping ESC and other
// control bytes so multi-byte escapes stay legible, and returns the
// column within that rendered string at which caretIndex falls so the
// caller can print a caret lined up underneath the right byte even
// though escaped bytes take up more than one printed column.
func visualize(raw []byte, caretIndex int) (string, int) {
	var b strings.Builder
	caretCol := 0
	for i, c := range raw {
		if i == caretIndex {
			caretCol = b.Len()
		}
		b.WriteString(displayByte(c))
	}
	if caretIndex >= len(raw) {
		caretCol = b.Len()
	}
	return b.String(), caretCol
}
