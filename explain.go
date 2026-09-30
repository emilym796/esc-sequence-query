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

// decPrivateModeNames covers the DEC private modes set/reset with CSI ?h
// and CSI ?l - the ones actually seen in the wild (alternate screen,
// mouse reporting, bracketed paste) rather than the full DEC catalogue.
var decPrivateModeNames = map[int]string{
	1:    "DECCKM: cursor keys send application sequences",
	2:    "DECANM: VT52 mode",
	3:    "DECCOLM: 132 column mode",
	4:    "DECSCLM: smooth scroll",
	5:    "DECSCNM: reverse video",
	6:    "DECOM: origin mode",
	7:    "DECAWM: auto-wrap mode",
	8:    "DECARM: auto-repeat keys",
	9:    "X10 mouse reporting",
	12:   "cursor blinking",
	25:   "DECTCEM: cursor visible",
	47:   "use alternate screen buffer",
	1000: "VT200 mouse reporting (button press/release)",
	1002: "button-event mouse reporting (with drag)",
	1003: "any-event mouse reporting",
	1004: "focus reporting",
	1005: "UTF-8 mouse mode",
	1006: "SGR mouse mode",
	1015: "urxvt mouse mode",
	1047: "use alternate screen buffer, clear on exit",
	1048: "save/restore cursor position",
	1049: "save cursor and use alternate screen buffer, clear on exit",
	2004: "bracketed paste mode",
}

// ansiModeNames covers the (non-DEC) standard modes settable with plain
// CSI h / CSI l, without a leading '?'. There aren't many of these in
// practical use, so the table is small on purpose.
var ansiModeNames = map[int]string{
	2:  "KAM: keyboard action mode",
	4:  "IRM: insert/replace mode",
	12: "SRM: send/receive (local echo) mode",
	20: "LNM: automatic newline mode",
}

// explainMode describes the parameters of a CSI h/l sequence. A leading
// '?' marks a DEC private mode rather than a standard ANSI one, and the
// two have separate, non-overlapping numbering, so they need separate
// tables.
func explainMode(params string) string {
	private := strings.HasPrefix(params, "?")
	body := params
	if private {
		body = params[1:]
	}
	if body == "" {
		return "no parameter given"
	}
	table := ansiModeNames
	kind := "mode"
	if private {
		table = decPrivateModeNames
		kind = "DEC private mode"
	}
	parts := strings.Split(body, ";")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			names = append(names, fmt.Sprintf("invalid parameter %q", p))
			continue
		}
		if name, ok := table[n]; ok {
			names = append(names, name)
		} else {
			names = append(names, fmt.Sprintf("unknown %s %d", kind, n))
		}
	}
	return strings.Join(names, ", ")
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
		return fmt.Sprintf("CSI %s%c  set mode: %s", params, final, explainMode(params))
	case 'l':
		return fmt.Sprintf("CSI %s%c  reset mode: %s", params, final, explainMode(params))
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

// explainOSCPalette describes the argument of OSC 4, which is a list of
// index;spec pairs. A spec of "?" asks the terminal to report the entry
// instead of changing it, so the two cases read differently.
func explainOSCPalette(args string) string {
	if args == "" {
		return "palette entry: no parameters given"
	}
	parts := strings.Split(args, ";")
	if len(parts)%2 != 0 {
		return fmt.Sprintf("palette entry: index %q has no color spec", parts[len(parts)-1])
	}
	descs := make([]string, 0, len(parts)/2)
	for i := 0; i < len(parts); i += 2 {
		idx, err := strconv.Atoi(parts[i])
		if err != nil || idx < 0 || idx > 255 {
			descs = append(descs, fmt.Sprintf("invalid palette index %q", parts[i]))
			continue
		}
		if parts[i+1] == "?" {
			descs = append(descs, fmt.Sprintf("query palette entry %d", idx))
		} else {
			descs = append(descs, fmt.Sprintf("set palette entry %d to %s", idx, parts[i+1]))
		}
	}
	return strings.Join(descs, ", ")
}

// explainOSCHyperlink describes the argument of OSC 8, "params;URI". The
// URI is everything after the first semicolon, since URIs may contain
// semicolons themselves. An empty URI ends the current link.
func explainOSCHyperlink(args string) string {
	i := strings.IndexByte(args, ';')
	if i < 0 {
		return "hyperlink: missing ';' between parameters and URI"
	}
	params, uri := args[:i], args[i+1:]
	if uri == "" {
		return "end hyperlink"
	}
	desc := fmt.Sprintf("start hyperlink to %q", uri)
	if params == "" {
		return desc
	}
	var opts []string
	for _, kv := range strings.Split(params, ":") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			opts = append(opts, fmt.Sprintf("malformed parameter %q", kv))
		} else if k == "id" {
			opts = append(opts, fmt.Sprintf("id %q", v))
		} else {
			opts = append(opts, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return desc + " (" + strings.Join(opts, ", ") + ")"
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
		return explainOSCPalette(rest)
	case "7":
		return fmt.Sprintf("report current working directory as %q", rest)
	case "8":
		return explainOSCHyperlink(rest)
	case "52":
		return "clipboard operation"
	default:
		return fmt.Sprintf("unknown OSC command %q (not in lookup table)", code)
	}
}
