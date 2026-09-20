package tools

// MediaPart is one piece of non-text media a tool's result may carry (round 070;
// ADR 0040 — relocated here from internal/domain/llm, which now references this
// type). MIMEType is resolved from the content (magic bytes), never the file name
// nor a declared MIME; Data is the raw file content. It is the INPUT to a
// conversation message's media (internal/domain/llm serializes it on the selected
// provider's wire).
//
// It is owned by the tool domain because it is a tool's OUTPUT artifact: a
// media-producing tool returns it in-band (see MediaTool), so the tool no longer
// reaches into the conversation model to name its own result type.
type MediaPart struct {
	MIMEType string
	Data     []byte
}
