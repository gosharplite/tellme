package steps

import "strings"

// unescapeText decodes the escapes \n, \t, \\, and \xHH (hex byte) in a Gherkin
// quoted parameter, so a multi-line or control-byte-bearing value (e.g. a
// newline-terminated or ANSI-bearing answer) can be written on a single Gherkin
// line without a docstring. Any other backslash sequence is left verbatim.
//
// round-005 correction pass (grill Q5/Q6): the single-line `"([^"]*)"` matchers
// could not express the newline- or ANSI-bearing answer class, which left the
// output-contract assertions vacuous on the only tested inputs. Decoding here
// makes that class representable and the contract falsifiable.
func unescapeText(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case 'n':
			b.WriteByte('\n')
			i++
		case 't':
			b.WriteByte('\t')
			i++
		case '\\':
			b.WriteByte('\\')
			i++
		case 'x':
			if i+3 < len(s) {
				hi, ok1 := hexVal(s[i+2])
				lo, ok2 := hexVal(s[i+3])
				if ok1 && ok2 {
					b.WriteByte(hi<<4 | lo)
					i += 3
					continue
				}
			}
			b.WriteByte(s[i])
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// hexVal returns the numeric value of a hexadecimal digit (0-9, a-f, A-F).
func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
