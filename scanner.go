package main

import (
	"fmt"
	"io"
)

// Position is a 1-indexed line and column in the input, counted in bytes.
type Position struct {
	Line int
	Col  int
}

type scanner struct {
	data []byte
	pos  int
	line int
	col  int
}

func newScanner(data []byte) *scanner {
	return &scanner{data: data, line: 1, col: 1}
}

func (s *scanner) position() Position {
	return Position{Line: s.line, Col: s.col}
}

func (s *scanner) advance() (byte, bool) {
	if s.pos >= len(s.data) {
		return 0, false
	}
	b := s.data[s.pos]
	s.pos++
	if b == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	return b, true
}

// Scan walks data looking for escape sequences and writes a one-line
// explanation of each one it recognizes to w. Ordinary text is skipped
// over silently. It stops and returns an error on the first sequence
// that violates the escape sequence grammar itself (an unterminated
// CSI, an OSC with no terminator, and so on) - not merely on a
// sequence we don't happen to have a name for.
func Scan(data []byte, w io.Writer) error {
	s := newScanner(data)
	for {
		start := s.position()
		b, ok := s.advance()
		if !ok {
			return nil
		}
		if b != 0x1B {
			continue
		}
		desc, err := s.readEscape(start)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%d:%d  %s\n", start.Line, start.Col, desc)
	}
}

func (s *scanner) readEscape(start Position) (string, error) {
	raw := []byte{0x1B}
	next, ok := s.advance()
	if !ok {
		return "", &ParseError{
			Pos:        start,
			Raw:        raw,
			CaretIndex: len(raw),
			Message:    "incomplete escape sequence: ESC at end of input with no following byte",
		}
	}
	raw = append(raw, next)

	switch next {
	case '[':
		return s.readCSI(start, raw)
	case ']':
		return s.readOSC(start, raw)
	default:
		if desc, ok := simpleEscapes[next]; ok {
			return fmt.Sprintf("ESC %s  %s", displayByte(next), desc), nil
		}
		return fmt.Sprintf("ESC %s  unknown escape sequence (not in lookup table)", displayByte(next)), nil
	}
}

// readCSI consumes a Control Sequence Introducer: ESC '[' followed by any
// number of parameter bytes (0x30-0x3F) and intermediate bytes (0x20-0x2F),
// terminated by exactly one final byte (0x40-0x7E). Anything else in that
// position is a grammar violation, not just an unrecognized sequence.
func (s *scanner) readCSI(start Position, raw []byte) (string, error) {
	paramsStart := len(raw)
	for {
		b, ok := s.advance()
		if !ok {
			return "", &ParseError{
				Pos:        start,
				Raw:        raw,
				CaretIndex: len(raw),
				Message:    "incomplete CSI sequence: reached end of input before a final byte (0x40-0x7E)",
			}
		}
		raw = append(raw, b)
		switch {
		case b >= 0x30 && b <= 0x3F, b >= 0x20 && b <= 0x2F:
			continue
		case b >= 0x40 && b <= 0x7E:
			params := string(raw[paramsStart : len(raw)-1])
			return explainCSI(b, params), nil
		default:
			return "", &ParseError{
				Pos:        start,
				Raw:        raw,
				CaretIndex: len(raw) - 1,
				Message: fmt.Sprintf(
					"invalid byte 0x%02X in CSI sequence: expected a parameter byte (0x30-0x3F), intermediate byte (0x20-0x2F), or final byte (0x40-0x7E)",
					b,
				),
			}
		}
	}
}

// readOSC consumes an Operating System Command: ESC ']' followed by a body,
// terminated by either BEL (0x07) or the two-byte String Terminator ESC '\'.
func (s *scanner) readOSC(start Position, raw []byte) (string, error) {
	for {
		b, ok := s.advance()
		if !ok {
			return "", &ParseError{
				Pos:        start,
				Raw:        raw,
				CaretIndex: len(raw),
				Message:    "unterminated OSC sequence: expected BEL (0x07) or ST (ESC \\) terminator",
			}
		}
		raw = append(raw, b)
		switch b {
		case 0x07:
			body := string(raw[2 : len(raw)-1])
			return fmt.Sprintf("OSC %s BEL  %s", body, explainOSC(body)), nil
		case 0x1B:
			nb, ok := s.advance()
			if !ok {
				return "", &ParseError{
					Pos:        start,
					Raw:        raw,
					CaretIndex: len(raw),
					Message:    "unterminated OSC sequence: expected '\\' (0x5C) to complete the ST terminator",
				}
			}
			raw = append(raw, nb)
			if nb == '\\' {
				body := string(raw[2 : len(raw)-2])
				return fmt.Sprintf("OSC %s ST  %s", body, explainOSC(body)), nil
			}
			return "", &ParseError{
				Pos:        start,
				Raw:        raw,
				CaretIndex: len(raw) - 1,
				Message:    fmt.Sprintf("invalid OSC terminator: ESC followed by 0x%02X, expected '\\' (0x5C)", nb),
			}
		}
	}
}

// displayByte renders a single byte the way it should show up in an
// explanation or an error's context line: control bytes are spelled out
// so they can't be confused with literal text.
func displayByte(b byte) string {
	switch {
	case b == 0x1B:
		return "\\e"
	case b < 0x20 || b == 0x7F:
		return fmt.Sprintf("\\x%02X", b)
	default:
		return string(b)
	}
}
