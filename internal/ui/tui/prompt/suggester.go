package prompt

// suggester holds the suggestion list and the selection cursor.
type suggester struct {
	items  []string
	cursor int
}

// newSuggester builds an empty suggestion list.
func newSuggester() suggester { return suggester{} }

// set replaces the items and clamps the cursor.
func (s *suggester) set(items []string) {
	s.items = items
	if s.cursor >= len(items) {
		s.cursor = 0
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

// cycle moves the selection cursor by delta (wrapping).
func (s *suggester) cycle(delta int) {
	if len(s.items) == 0 {
		return
	}
	s.cursor = (s.cursor + delta + len(s.items)) % len(s.items)
}
