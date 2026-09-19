package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// readImageToolName is the wire-valid canonical identifier (round 062; ADR 0032).
const readImageToolName = "read_image"

// imageMaxBytes is the inline image ceiling (32 MiB) — the OpenAI-compatible
// wire's inline limit accepted by DeepSeek (ADR 0032 D6). An image past it is a
// LOUD recoverable tool result, never truncated or partially sent.
const imageMaxBytes = 32 << 20

// The loud refusals (recoverable tool results fed back to the model). The
// `ERROR:` prefix matches the sibling readers' nil-error recoverable class.
const (
	imageTooLargeResult = "ERROR: the image is too large to send inline (it exceeds the 32 MiB limit)."
	notAPictureResult   = "ERROR: the content is not a supported picture (expected JPEG, PNG, GIF, or WebP)."
)

// readImage is the `read_image` agent tool (round 062): it reads a local image
// and attaches it to the model-visible request. It is offered to the model ONLY
// when the selected provider declares `VISION: true` (the assemblage gate in
// cmd/tellme); the tool itself simply reads and attaches.
type readImage struct{}

// NewReadImageTool builds the `read_image` agent tool.
func NewReadImageTool() domaintools.Tool { return readImage{} }

// Name is the wire-valid canonical identifier.
func (readImage) Name() string { return readImageToolName }

// Description is the model-facing summary.
func (readImage) Description() string {
	return "Read a local image file (JPEG, PNG, GIF, or WebP) and show it to the model. The selected provider must be able to see images."
}

// Contract declares the tool's protective per-call timeout — the local-reader
// class (a local file read; round-062 research D1).
func (readImage) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: readerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments: the mandatory
// `filepath` + the mandatory `reason`, plus the two resource params the shared
// builder declares, so `required ⊆ properties` holds (round 031).
func (readImage) Parameters() json.RawMessage {
	return resourceSchema(`"filepath":{"type":"string","description":"The path to the image file to read."}`, `"filepath","reason"`, readerDefaultTimeout)
}

// Execute reads the named file, resolves its kind from the file's CONTENT
// (magic bytes — never the extension), checks the 32 MiB inline ceiling, and
// attaches the image to the current tool call's collector (llm.AttachMedia) so
// the loop folds it back onto a `user` message (ADR 0032 D5/D6/D7). A file that
// is too large, or not a supported picture, returns a LOUD recoverable result
// (nil error) naming the problem; the image never reaches the request.
func (readImage) Execute(ctx context.Context, arguments string, _ domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	var args struct {
		FilePath string `json:"filepath"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("read_image: invalid arguments: %w", err)
	}
	if args.FilePath == "" {
		return "", fmt.Errorf("read_image: filepath argument is required")
	}
	f, err := os.Open(args.FilePath)
	if err != nil {
		return "", fmt.Errorf("read_image: failed to read file: %w", err)
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("read_image: failed to read file: %w", err)
	}
	if info.IsDir() {
		return "ERROR: the path is a directory; read_image expects an image file.", nil
	}
	if info.Size() > imageMaxBytes {
		return imageTooLargeResult, nil
	}
	// Read the whole file (bounded one byte past the ceiling so a racing growth
	// cannot exceed it silently).
	data, err := io.ReadAll(io.LimitReader(f, imageMaxBytes+1))
	if err != nil {
		return "", fmt.Errorf("read_image: failed to read file: %w", err)
	}
	if len(data) > imageMaxBytes {
		return imageTooLargeResult, nil
	}
	mime, ok := imageMIME(data)
	if !ok {
		return notAPictureResult, nil
	}
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	llm.AttachMedia(ctx, llm.MediaPart{MIMEType: mime, Data: data})
	return fmt.Sprintf("Successfully read image from %s (%s, %d bytes)", args.FilePath, mime, len(data)), nil
}

// imageMIME resolves the media type from the file's CONTENT (magic bytes) for
// the four formats DeepSeek documents — JPEG, PNG, GIF, WebP — or reports false
// for anything else. The name and any declared MIME are deliberately ignored
// (ADR 0032 D5).
func imageMIME(data []byte) (string, bool) {
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", true
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png", true
	case bytes.HasPrefix(data, []byte("GIF87a")), bytes.HasPrefix(data, []byte("GIF89a")):
		return "image/gif", true
	case len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return "image/webp", true
	}
	return "", false
}
