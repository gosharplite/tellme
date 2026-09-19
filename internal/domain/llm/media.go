package llm

import "context"

// MediaPart is one piece of non-text media attached to a conversation message
// (round 062; ADR 0032). Data is the raw file content and MIMEType is resolved
// from the content (magic bytes), never the file name nor a declared MIME.
type MediaPart struct {
	MIMEType string
	Data     []byte
}

// mediaKey is the private context key for a per-call media collector.
type mediaKey struct{}

// WithMediaCollector returns a context that collects media a tool attaches
// during one call. The agent loop installs a FRESH collector per tool call (so
// the media belongs to that call), and a tool attaches with AttachMedia.
func WithMediaCollector(ctx context.Context, sink *[]MediaPart) context.Context {
	return context.WithValue(ctx, mediaKey{}, sink)
}

// AttachMedia appends media to the current call's collector, if one is
// installed. A context without a collector (e.g. a unit test calling a tool
// directly, or a non-agent caller) is a silent no-op, so a tool can always call
// it without a nil check.
func AttachMedia(ctx context.Context, parts ...MediaPart) {
	if sink, ok := ctx.Value(mediaKey{}).(*[]MediaPart); ok && sink != nil {
		*sink = append(*sink, parts...)
	}
}
