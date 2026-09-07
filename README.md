# escq

A command-line tool that answers one question: what does this terminal
escape sequence do?

Terminal output is full of bytes you can't read directly - `ESC [ 3 1 m`
turns text red, `ESC ] 0 ; ... BEL` sets a window title, and so on. When
you're staring at a log file, a captured pty session, or a program that's
clearly emitting *something* broken, it's tedious to work out by hand
which sequence you're looking at, or exactly where it breaks. `escq`
reads a stream of bytes, walks through it, and for every escape sequence
it finds prints where it is and what it means. Plain text is skipped over.

## Usage

```
escq [file]
```

With no argument it reads from stdin.

```
$ printf '\033[31mHello\033[0m\n' | escq
1:1  CSI 31m  SGR: set foreground color to red
1:11  CSI 0m  SGR: reset all attributes
```

Each line is `<line>:<column>  <sequence>  <explanation>`, where the
line and column point at the ESC byte that started the sequence.

## Error messages

Most tools that touch escape sequences either render them (so you can't
see the bytes) or dump raw hex (so you have to decode the grammar
yourself). When `escq` hits a sequence that's actually malformed - not
just unrecognized, but structurally broken - it stops and explains
exactly where and why, with the offending bytes shown and a caret
pointing at the problem:

```
$ printf 'bad: \033[3' | escq
escq: incomplete CSI sequence: reached end of input before a final byte (0x40-0x7E)
  at line 1, column 6

    \e[3
        ^
```

Control bytes in the context line are spelled out (`\e` for ESC, `\xHH`
for other control bytes) so the caret lines up with something you can
actually read, rather than sitting under invisible bytes.

## What it recognizes today

- CSI sequences (`ESC [ ... final`): cursor movement, erase in
  display/line, SGR (colors and text attributes), save/restore cursor,
  device status reports, mode set/reset.
- OSC sequences (`ESC ] ... BEL` or `ESC ] ... ESC \`): window/icon
  title, working directory reporting, hyperlinks, clipboard, by number.
- The common two-byte ESC sequences: save/restore cursor, keypad mode,
  reset, index/reverse-index, next line, tab stop.

Sequences outside these tables are reported as "unknown" rather than as
errors - not being in the lookup table isn't the same as being invalid.
Errors are reserved for input that breaks the actual escape sequence
grammar: an unterminated CSI or OSC, or a byte that can't legally appear
where it does.

## Building

```
go build .
```

No third-party dependencies - standard library only.

## License

MIT, see [LICENSE](LICENSE).
