package steps

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 062 (ADR 0032) E2E steps: the image journey's Given/Then sentences.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the workspace holds an image file "([^"]*)" that is a (\w+) picture$`, givenWorkspaceImage)
		ctx.Given(`^the workspace holds an image file "([^"]*)" larger than the image size limit$`, givenWorkspaceOversizeImage)
		ctx.Given(`^the workspace holds a file "([^"]*)" whose content is not a supported picture$`, givenWorkspaceNotAPicture)
		ctx.Given(`^a configured provider "([^"]*)" that can take images and whose endpoint asks tellme to read "([^"]*)" and then answers with "([^"]*)"$`, givenVisionProviderReadsImage)
		ctx.Given(`^a configured provider "([^"]*)" that can take images whose endpoint reports the offered tools and then answers with "([^"]*)"$`, givenVisionProviderReportsTools)
		ctx.Given(`^a configured provider "([^"]*)" that cannot take images whose endpoint reports the offered tools and then answers with "([^"]*)"$`, givenBlindProviderReportsTools)

		ctx.Then(`^the request carried the image file "([^"]*)"$`, thenRequestCarriedImage)
		ctx.Then(`^the image was attached as a "([^"]*)" picture$`, thenImageAttachedAs)
		ctx.Then(`^the request carried no image$`, thenRequestCarriedNoImage)
		ctx.Then(`^the request offered the read-image tool$`, thenOfferedReadImage)
		ctx.Then(`^the request offered no read-image tool$`, thenOfferedNoReadImage)
		ctx.Then(`^the tool result reported the image is too large$`, thenToolResultImageTooLarge)
		ctx.Then(`^the tool result reported the content is not a supported picture$`, thenToolResultNotAPicture)
	})
}

// imageBytes returns a minimal valid picture of the named kind (a real magic-byte
// header so the content sniffer resolves it).
func imageBytes(kind string) ([]byte, bool) {
	switch strings.ToUpper(kind) {
	case "PNG":
		return []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R'}, true
	case "JPEG":
		return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}, true
	case "GIF":
		return []byte("GIF89a\x01\x00\x01\x00"), true
	case "WEBP":
		return append([]byte("RIFF\x00\x00\x00\x00WEBP"), 0x00), true
	}
	return nil, false
}

// givenWorkspaceImage writes a small valid picture of the named kind into the
// run's working directory.
func givenWorkspaceImage(ctx context.Context, name, kind string) error {
	sc := scenarioFrom(ctx)
	data, ok := imageBytes(kind)
	if !ok {
		return fmt.Errorf("unsupported picture kind %q", kind)
	}
	return sc.writeWorkFile(name, string(data))
}

// givenWorkspaceOversizeImage writes a sparse picture larger than the 32 MiB
// inline ceiling into the run's working directory.
func givenWorkspaceOversizeImage(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	p := filepath.Join(sc.workDir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}); err != nil {
		return err
	}
	const over = 32<<20 + 1
	if _, err := f.Seek(int64(over-1), 0); err != nil {
		return err
	}
	_, err = f.Write([]byte{0})
	return err
}

// givenWorkspaceNotAPicture writes a file whose bytes are not a supported
// picture, whatever its name suggests.
func givenWorkspaceNotAPicture(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	return sc.writeWorkFile(name, "this is plainly not a picture")
}

// writeVisionConfig writes the default config selecting `provider` (an
// OpenAI-compatible family entry pointing at the fake) with VISION set per the
// vision argument.
func writeVisionConfig(sc *scenarioContext, provider, url string, vision bool) error {
	cfg := fmt.Sprintf("MODE: butler\n"+
		"PERSON: \"e2e persona\"\n"+
		"SELECTED_PROVIDER: %s\n"+
		"PROVIDERS:\n"+
		"  %s:\n"+
		"    TYPE: deepseek\n"+
		"    MODEL: deepseek-v4-flash\n"+
		"    URL: %s\n"+
		"    API_KEY: test-key\n"+
		"    VISION: %t\n",
		provider, provider, url, vision)
	return sc.writeFile("configs/butler.yaml", []byte(cfg))
}

// givenVisionProviderReadsImage arranges a vision-enabled provider whose fake
// scripts one `read_image` call (for {path}), then answers {answer}.
func givenVisionProviderReadsImage(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	args, err := json.Marshal(map[string]any{
		"filepath": path,
		"reason":   "look at the picture",
	})
	if err != nil {
		return err
	}
	f.Script(
		fakeprovider.Reply{ToolName: "read_image", Arguments: string(args)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_image"
	sc.registerFake(provider, f)
	return writeVisionConfig(sc, provider, f.URL(), true)
}

// givenVisionProviderReportsTools arranges a vision-enabled provider whose fake
// reports the offered tool set then answers {answer}.
func givenVisionProviderReportsTools(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Answer(unescapeText(answer))
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return writeVisionConfig(sc, provider, f.URL(), true)
}

// givenBlindProviderReportsTools arranges a provider WITHOUT vision (no VISION
// key) whose fake reports the offered tool set then answers {answer}.
func givenBlindProviderReportsTools(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Answer(unescapeText(answer))
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return writeVisionConfig(sc, provider, f.URL(), false)
}

// lastBody returns the most recent recorded request body across the scenario's
// fakes (the single fake every round-062 feature uses).
func lastBody(sc *scenarioContext) string {
	f := sc.onlyFake()
	if f == nil {
		return ""
	}
	return f.BodyAt(-1)
}

// imageURIs extracts every `image_url.url` string from a recorded request body.
func imageURIs(body string) []string {
	var req struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return nil
	}
	var uris []string
	for _, m := range req.Messages {
		var parts []struct {
			Type     string `json:"type"`
			ImageURL struct {
				URL string `json:"url"`
			} `json:"image_url"`
		}
		if err := json.Unmarshal(m.Content, &parts); err != nil {
			continue
		}
		for _, p := range parts {
			if p.Type == "image_url" && p.ImageURL.URL != "" {
				uris = append(uris, p.ImageURL.URL)
			}
		}
	}
	return uris
}

// requestContentText returns the concatenated string content of every recorded
// message (used to find a folded-back tool result).
func requestContentText(body string) string {
	var req struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, m := range req.Messages {
		var s string
		if err := json.Unmarshal(m.Content, &s); err == nil {
			sb.WriteString(s)
			sb.WriteString("\n")
			continue
		}
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(m.Content, &parts); err == nil {
			for _, p := range parts {
				if p.Type == "text" {
					sb.WriteString(p.Text)
					sb.WriteString("\n")
				}
			}
		}
	}
	return sb.String()
}

// thenRequestCarriedImage asserts the recorded request carries the named file's
// image bytes exactly (decoded from the inline data URI).
func thenRequestCarriedImage(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	want, err := os.ReadFile(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return err
	}
	body := lastBody(sc)
	for _, uri := range imageURIs(body) {
		const marker = ";base64,"
		i := strings.Index(uri, marker)
		if i < 0 {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(uri[i+len(marker):])
		if err != nil {
			continue
		}
		if string(data) == string(want) {
			return nil
		}
	}
	return fmt.Errorf("the request did not carry the image file %q (image URIs: %v)", name, imageURIs(body))
}

// thenImageAttachedAs asserts the request's image block declares the MIME type.
func thenImageAttachedAs(ctx context.Context, mime string) error {
	sc := scenarioFrom(ctx)
	for _, uri := range imageURIs(lastBody(sc)) {
		if strings.HasPrefix(uri, "data:"+mime+";base64,") {
			return nil
		}
	}
	return fmt.Errorf("no image block declares the media type %q", mime)
}

// thenRequestCarriedNoImage asserts NO image block reached the request.
func thenRequestCarriedNoImage(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if uris := imageURIs(lastBody(sc)); len(uris) > 0 {
		return fmt.Errorf("the request carried an image: %v", uris)
	}
	return nil
}

// thenOfferedReadImage asserts the offered tool set contains `read_image`.
func thenOfferedReadImage(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider")
	}
	for _, n := range f.ToolNamesAt(-1) {
		if n == "read_image" {
			return nil
		}
	}
	return fmt.Errorf("read_image not offered; offered = %v", f.ToolNamesAt(-1))
}

// thenOfferedNoReadImage asserts the offered tool set contains no `read_image`.
func thenOfferedNoReadImage(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider")
	}
	for _, n := range f.ToolNamesAt(-1) {
		if n == "read_image" {
			return fmt.Errorf("read_image was offered to a provider that cannot take images")
		}
	}
	return nil
}

// thenToolResultImageTooLarge asserts the folded-back tool result reports the
// image is too large.
func thenToolResultImageTooLarge(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(requestContentText(lastBody(sc)), "too large") {
		return nil
	}
	return fmt.Errorf("the folded-back tool result did not report the image is too large")
}

// thenToolResultNotAPicture asserts the folded-back tool result reports the
// content is not a supported picture.
func thenToolResultNotAPicture(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(requestContentText(lastBody(sc)), "not a supported picture") {
		return nil
	}
	return fmt.Errorf("the folded-back tool result did not report a not-a-picture refusal")
}
