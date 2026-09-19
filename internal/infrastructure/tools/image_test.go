package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// pngHeader is a minimal PNG magic-byte prefix; jpegHeader/gifHeader/webpHeader
// are the other three supported signatures (round-062; ADR 0032 D5).
var (
	pngHeader  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	gifHeader  = []byte("GIF89a....")
	webpHeader = append(append([]byte("RIFF"), 0x00, 0x00, 0x00, 0x00), []byte("WEBP")...)
)

// TestImageMIME pins the content sniff: the four supported kinds resolve from
// their magic bytes, and anything else is not a picture (the name is ignored).
func TestImageMIME(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
		ok   bool
	}{
		{"png", pngHeader, "image/png", true},
		{"jpeg", jpegHeader, "image/jpeg", true},
		{"gif", gifHeader, "image/gif", true},
		{"webp", webpHeader, "image/webp", true},
		{"text", []byte("hello, not a picture"), "", false},
		{"empty", nil, "", false},
		{"riff but not webp", []byte("RIFFxxxxAVI "), "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := imageMIME(c.data)
			if ok != c.ok || got != c.want {
				t.Errorf("imageMIME = (%q,%v), want (%q,%v)", got, ok, c.want, c.ok)
			}
		})
	}
}

// TestImageCeilingForFamily pins the single-owned family-aware ceiling (round
// 063; ADR 0033 D4): the Gemini/Vertex family is stricter than the
// OpenAI-compatible family, and any other/empty label defaults to the latter.
func TestImageCeilingForFamily(t *testing.T) {
	if got := ImageCeilingForFamily("gemini"); got != geminiImageCeiling {
		t.Errorf("gemini ceiling = %d, want %d", got, geminiImageCeiling)
	}
	if got := ImageCeilingForFamily("openai"); got != openAIImageCeiling {
		t.Errorf("openai ceiling = %d, want %d", got, openAIImageCeiling)
	}
	if got := ImageCeilingForFamily(""); got != openAIImageCeiling {
		t.Errorf("empty-family ceiling = %d, want the OpenAI-compatible default", got)
	}
	if geminiImageCeiling >= openAIImageCeiling {
		t.Errorf("the Gemini ceiling must be stricter than the OpenAI-compatible one (%d vs %d)", geminiImageCeiling, openAIImageCeiling)
	}
}

// TestReadImageAttachesMedia pins the happy path: the image is attached to the
// call's collector with the content-sniffed MIME and its exact bytes.
func TestReadImageAttachesMedia(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mystery.bin") // name says .bin; content says PNG
	if err := os.WriteFile(p, pngHeader, 0o644); err != nil {
		t.Fatal(err)
	}
	var media []llm.MediaPart
	ctx := llm.WithMediaCollector(context.Background(), &media)
	args, _ := json.Marshal(map[string]string{"filepath": p, "reason": "look"})
	res, err := NewReadImageTool(openAIImageCeiling).Execute(ctx, string(args), domaintools.ByteBudget(1<<20))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(res, "image/png") {
		t.Errorf("result = %q, want it to name the sniffed type", res)
	}
	if len(media) != 1 || media[0].MIMEType != "image/png" {
		t.Fatalf("media = %+v, want one image/png", media)
	}
	if string(media[0].Data) != string(pngHeader) {
		t.Errorf("attached bytes differ from the file's")
	}
}

// TestReadImageRefusesNotAPicture pins the loud refusal for a non-picture: a
// recoverable result (nil error), and NOTHING is attached.
func TestReadImageRefusesNotAPicture(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "notes.png") // name says .png; content is text
	if err := os.WriteFile(p, []byte("just text"), 0o644); err != nil {
		t.Fatal(err)
	}
	var media []llm.MediaPart
	ctx := llm.WithMediaCollector(context.Background(), &media)
	args, _ := json.Marshal(map[string]string{"filepath": p, "reason": "look"})
	res, err := NewReadImageTool(openAIImageCeiling).Execute(ctx, string(args), domaintools.ByteBudget(1<<20))
	if err != nil {
		t.Fatalf("Execute should return a recoverable result, got error %v", err)
	}
	if !strings.HasPrefix(res, "ERROR:") || !strings.Contains(res, "not a supported picture") {
		t.Errorf("result = %q, want the not-a-picture refusal", res)
	}
	if len(media) != 0 {
		t.Errorf("media attached for a non-picture: %+v", media)
	}
}

// TestReadImageRefusesOversize pins the ceiling boundary against the tool's
// RESOLVED ceiling (round 063): exactly at the limit is accepted, one byte over
// is the loud refusal naming that limit, with nothing attached.
func TestReadImageRefusesOversize(t *testing.T) {
	dir := t.TempDir()
	limit := geminiImageCeiling // the stricter family ceiling, to prove it is honoured
	atLimit := filepath.Join(dir, "at.png")
	if err := writeSparse(atLimit, pngHeader, limit); err != nil {
		t.Fatal(err)
	}
	over := filepath.Join(dir, "over.png")
	if err := writeSparse(over, pngHeader, limit+1); err != nil {
		t.Fatal(err)
	}

	var media []llm.MediaPart
	ctx := llm.WithMediaCollector(context.Background(), &media)

	atArgs, _ := json.Marshal(map[string]string{"filepath": atLimit, "reason": "look"})
	if _, err := NewReadImageTool(limit).Execute(ctx, string(atArgs), domaintools.ByteBudget(1)); err != nil {
		t.Fatalf("at-limit Execute: %v", err)
	}
	if len(media) != 1 {
		t.Errorf("at-limit media = %d, want 1 (the ceiling is inclusive)", len(media))
	}

	media = nil
	overArgs, _ := json.Marshal(map[string]string{"filepath": over, "reason": "look"})
	res, err := NewReadImageTool(limit).Execute(ctx, string(overArgs), domaintools.ByteBudget(1))
	if err != nil {
		t.Fatalf("oversize Execute should be a recoverable result, got %v", err)
	}
	if !strings.Contains(res, "too large") {
		t.Errorf("result = %q, want the too-large refusal", res)
	}
	if !strings.Contains(res, "14 MiB") {
		t.Errorf("result = %q, want it to name the resolved (Gemini) limit", res)
	}
	if len(media) != 0 {
		t.Errorf("media attached for an oversize image")
	}
}

// writeSparse writes `total` bytes: `header` then zero padding, using a sparse
// file so a 32 MiB fixture is cheap.
func writeSparse(path string, header []byte, total int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(header); err != nil {
		return err
	}
	if total > len(header) {
		if _, err := f.Seek(int64(total-1), 0); err != nil {
			return err
		}
		if _, err := f.Write([]byte{0}); err != nil {
			return err
		}
	}
	return nil
}
