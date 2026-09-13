package prompt

// suggester holds the suggestion list and the selection cursor. Behaviour —
// seeding from the source, debounced refresh, and accept — lands with the
// Feature phase (round-015 T032). This is the landing skeleton.
type suggester struct {
	items  []string
	cursor int
}

// newSuggester builds an empty suggestion skeleton.
func newSuggester() suggester { return suggester{} }

// top is the currently selected suggestion (empty when the list is empty or the
// cursor is out of range). Skeleton accessor used by the model.
func (s *suggester) top() string {
	if s.cursor < 0 || s.cursor >= len(s.items) {
		return ""
	}
	return s.items[s.cursor]
}
