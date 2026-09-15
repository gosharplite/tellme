package tools

// binary.go (round 021): the stdlib binary-probe helper.
//
// It mirrors the reference's null-byte probe (persistence.IsBinary): a NUL byte
// within the leading bytes marks content as binary rather than text. Wired into
// read_files (T038).

// isBinary reports whether content looks like a binary blob rather than text — a
// NUL byte within the leading bytes (matching the reference probe).
func isBinary(content []byte) bool {
	const probe = 8000
	n := len(content)
	if n > probe {
		n = probe
	}
	for i := 0; i < n; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
