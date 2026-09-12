// Package ui renders the provider's answer as formatted Markdown for tellme's
// default output path, using the reference's renderer (glamour). It is the
// project's first presentation dependency (round-006 research Decision 1).
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

// defaultStyle is the fallback glamour style when GLAMOUR_STYLE is unset
// (round-006 research Decision 2).
const defaultStyle = "dark"

// Renderer renders Markdown answers to ANSI. Construct one with NewRenderer.
type Renderer struct {
	style    string
	warnOnce sync.Once
}

// NewRenderer builds a renderer whose style resolves from GLAMOUR_STYLE
// (falling back to defaultStyle), matching the reference (round-006 research
// Decision 2).
func NewRenderer() *Renderer {
	style := os.Getenv("GLAMOUR_STYLE")
	if style == "" {
		style = defaultStyle
	}
	return &Renderer{style: style}
}

// Render renders markdown to ANSI, word-wrapped at width columns (width <= 0
// uses glamour's built-in default). The LaTeX→Unicode sanitizer runs first
// (round-006 research Decision 5). On any renderer-construction or render
// failure it returns the sanitized text UNCHANGED with degraded=true, so the
// caller can fall back to raw output without failing the turn (round-006
// research Decision 6; reference ADR-007 parity).
func (r *Renderer) Render(markdown string, width int) (out string, degraded bool) {
	sanitized := sanitizeForTerminal(markdown)
	opts := []glamour.TermRendererOption{
		glamour.WithStandardStyle(r.style),
		glamour.WithEmoji(),
	}
	if width > 0 {
		opts = append(opts, glamour.WithWordWrap(width))
	}
	tr, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return sanitized, true
	}
	rendered, err := tr.Render(sanitized)
	if err != nil {
		return sanitized, true
	}
	return rendered, false
}

// WarnDegraded emits the markdown-degradation warning at most once per renderer
// (round-006 research Decision 6). It is a NON-class warning: it deliberately
// does NOT use the reserved `tellme: ` prefix, which is kept exclusively for the
// frozen class phrases — so the closed phrase vocabulary is untouched and the
// "exactly one `tellme: ` line" contract holds.
func (r *Renderer) WarnDegraded(w io.Writer) {
	r.warnOnce.Do(func() {
		_, _ = fmt.Fprintln(w, "[WARN] markdown rendering degraded, falling back to raw text")
	})
}

// sanitizeForTerminal converts common LaTeX/Math notation LLMs emit into
// terminal-friendly Unicode before rendering (round-006 research Decision 5),
// mirroring the reference's sanitizeForTerminal.
func sanitizeForTerminal(text string) string {
	replacements := []struct{ old, new string }{
		{`$\leftrightarrow$`, "↔"},
		{`$\rightarrow$`, "→"},
		{`$\leftarrow$`, "←"},
		{`$\Rightarrow$`, "⇒"},
		{`$\Leftarrow$`, "⇐"},
		{`$\dots$`, "..."},
		{`$\cdot$`, "·"},
		{`$\times$`, "×"},
		{`$\checkmark$`, "✓"},
	}
	for _, rep := range replacements {
		text = strings.ReplaceAll(text, rep.old, rep.new)
	}
	return text
}
