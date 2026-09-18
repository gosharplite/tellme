package history

// ToolUsageRow is one tool's line in the offline tool-usage report (round 026).
// Round 051 (R5.5 of #92; ADR 0020) folds this (formerly `ui.ToolUsageRow`) onto
// the existing `history.ToolUsageCounts` — one concept, one name: the row embeds
// the counts rather than duplicating OK/Error/Timeout.
type ToolUsageRow struct {
	Tool   string
	Counts ToolUsageCounts
}

// OK / Error / Timeout are convenience accessors onto the embedded counts, so a
// formatter reads the flat shape without reaching through Counts.
func (r ToolUsageRow) OK() int      { return r.Counts.OK }
func (r ToolUsageRow) Error() int   { return r.Counts.Error }
func (r ToolUsageRow) Timeout() int { return r.Counts.Timeout }

// Total is the tool's total invocation count.
func (r ToolUsageRow) Total() int { return r.Counts.Total() }
