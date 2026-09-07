package main

import (
	"fmt"
	"strconv"
	"strings"
)

// simpleEscapes covers two-byte sequences: ESC followed by a single final
// byte, no parameters involved.
var simpleEscapes = map[byte]string{
	'7':  "DECSC: save cursor position and attributes",
	'8':  "DECRC: restore cursor position and attributes",
	'=':  "DECKPAM: set application keypad mode",
	'>':  "DECKPNM: set normal keypad mode",
	'c':  "RIS: reset terminal to initial state",
	'D':  "IND: move cursor down one line, scrolling if needed",
	'E':  "NEL: move cursor to start of next line",
	'H':  "HTS: set a horizontal tab stop at the cursor column",
	'M':  "RI: move cursor up one line, scrolling if needed",
	'Z':  "DECID: identify terminal (obsolete, superseded by CSI c)",
	'\\': "ST: string terminator",
}

// sgrNames covers the common Select Graphic Rendition parameters. It is
// intentionally not exhaustive yet - 256-color and truecolor sequences
// (38/48;5;n and 38/48;2;r;g;b) aren't decoded to their color name.
var sgrNames = map[int]string{
	0:  "reset all attributes",
	1:  "bold",
	2:  "faint",
	3:  "italic",
	4:  "underline",
	5:  "slow blink",
	6:  "rapid blink",
	7:  "reverse video",
	8:  "conceal",
	9:  "strikethrough",
	22: "normal intensity",
	23: "not italic",
	24: "not underlined",
	25: "not blinking",
	27: "not reversed",
	28: "reveal",
	29: "not strikethrough",

	30: "set foreground color to black",
	31: "set foreground color to red",
	32: "set foreground color to green",
	33: "set foreground color to yellow",
	34: "set foreground color to blue",
	35: "set foreground color to magenta",
	36: "set foreground color to cyan",
	37: "set foreground color to white",
	39: "reset foreground color to default",

	40: "set background color to black",
	41: "set background color to red",
	42: "set background color to green",
	43: "set background color to yellow",
	44: "set background color to blue",
	45: "set background color to magenta",
	46: "set background color to cyan",
	47: "set background color to white",
	49: "reset background color to default",

	90: "set foreground color to bright black",
	91: "set foreground color to bright red",
	92: "set foreground color to bright green",
	93: "set foreground color to bright yellow",
	94: "set foreground color to bright blue",
	95: "set foreground color to bright magenta",
	96: "set foreground color to bright cyan",
	97: "set foreground color to bright white",

	100: "set background color to bright black",
	101: "set background color to bright red",
	102: "set background color to bright green",
	103: "set background color to bright yellow",
	104: "set background color to bright blue",
	105: "set background color to bright magenta",
	106: "set background color to bright cyan",
	107: "set background color to bright white",
}

func withDefault(params, def string) string {
	if params == "" {
		return def
	}
	return params
}

func explainCSI(final byte, params string) string {
	switch final {
	case 'A':
		return fmt.Sprintf("CSI %s%c  cursor up %s line(s)", params, final, withDefault(params, "1"))
	case 'B':
		return fmt.Sprintf("CSI %s%c  cursor down %s line(s)", params, final, withDefault(params, "1"))
	case 'C':
		return fmt.Sprintf("CSI %s%c  cursor forward %s column(s)", params, final, withDefault(params, "1"))
	case 'D':
		return fmt.Sprintf("CSI %s%c  cursor back %s column(s)", params, final, withDefault(params, "1"))
	case 'E':
		return fmt.Sprintf("CSI %s%c  cursor to start of line, %s line(s) down", params, final, withDefault(params, "1"))
	case 'F':
		return fmt.Sprintf("CSI %s%c  cursor to start of line, %s line(s) up", params, final, withDefault(params, "1"))
	case 'G':
		return fmt.Sprintf("CSI %s%c  move cursor to column %s", params, final, withDefault(params, "1"))
	case 'H', 'f':
		return fmt.Sprintf("CSI %s%c  move cursor to row;column %s", params, final, withDefault(params, "1;1"))
	case 'J':
		return fmt.Sprintf("CSI %s%c  erase in display (mode %s)", params, final, withDefault(params, "0"))
	case 'K':
		return fmt.Sprintf("CSI %s%c  erase in line (mode %s)", params, final, withDefault(params, "0"))
	case 'S':
		return fmt.Sprintf("CSI %s%c  scroll up %s line(s)", params, final, withDefault(params, "1"))
	case 'T':
		return fmt.Sprintf("CSI %s%c  scroll down %s line(s)", params, final, withDefault(params, "1"))
	case 'm':
		return fmt.Sprintf("CSI %s%c  SGR: %s", params, final, explainSGR(params))
	case 'n':
		return fmt.Sprintf("CSI %s%c  device status report (type %s)", params, final, withDefault(params, "0"))
	case 's':
		return "CSI s  save cursor position"
	case 'u':
		return "CSI u  restore cursor position"
	case 'h':
		return fmt.Sprintf("CSI %s%c  set mode %s", params, final, params)
	case 'l':
		return fmt.Sprintf("CSI %s%c  reset mode %s", params, final, params)
	default:
		return fmt.Sprintf("CSI %s%c  unknown final byte 0x%02X (not in lookup table)", params, final, final)
	}
}

func explainSGR(params string) string {
	if params == "" {
		params = "0"
	}
	parts := strings.Split(params, ";")
	descs := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			p = "0"
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			descs = append(descs, fmt.Sprintf("invalid parameter %q", p))
			continue
		}
		if name, ok := sgrNames[n]; ok {
			descs = append(descs, name)
		} else {
			descs = append(descs, fmt.Sprintf("unknown SGR code %d", n))
		}
	}
	return strings.Join(descs, ", ")
}

func explainOSC(body string) string {
	code := body
	rest := ""
	if i := strings.IndexByte(body, ';'); i >= 0 {
		code = body[:i]
		rest = body[i+1:]
	}
	switch code {
	case "0":
		return fmt.Sprintf("set icon name and window title to %q", rest)
	case "1":
		return fmt.Sprintf("set icon name to %q", rest)
	case "2":
		return fmt.Sprintf("set window title to %q", rest)
	case "4":
		return fmt.Sprintf("set or query a color palette entry (%s)", rest)
	case "7":
		return fmt.Sprintf("report current working directory as %q", rest)
	case "8":
		return "hyperlink"
	case "52":
		return "clipboard operation"
	default:
		return fmt.Sprintf("unknown OSC command %q (not in lookup table)", code)
	}
}
