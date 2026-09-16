package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domainskills "github.com/gosharplite/tellme/internal/domain/skills"
)

// T017 [UNIT] — pins the `list_skills` tool contract (Decision 6/7; techstack
// `Skills listing tool (list_skills)` row).
func TestListSkills(t *testing.T) {
	t.Parallel()

	t.Run("lists name, description, and location, path-sorted", func(t *testing.T) {
		t.Parallel()
		tool := NewSkillsTool(func() ([]domainskills.Skill, error) {
			return []domainskills.Skill{
				{Name: "beta", Description: "Beta skill", Location: "/skills/beta/SKILL.md"},
				{Name: "alpha", Description: "Alpha skill", Location: "/skills/alpha/SKILL.md"},
			}, nil
		})
		out, err := tool.Execute(context.Background(), `{"reason":"list the skills"}`, 1<<20)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		for _, want := range []string{"alpha", "Alpha skill", "/skills/alpha/SKILL.md", "beta", "/skills/beta/SKILL.md"} {
			if !strings.Contains(out, want) {
				t.Errorf("listing missing %q; got %q", want, out)
			}
		}
		if strings.Index(out, "alpha") > strings.Index(out, "beta") {
			t.Errorf("listing is not path-sorted; got %q", out)
		}
		if strings.Contains(out, timeoutMarker) {
			t.Errorf("a non-expired deadline returned a timeout result: %q", out)
		}
	})

	t.Run("reports an empty catalog as a result", func(t *testing.T) {
		t.Parallel()
		tool := NewSkillsTool(nil) // unbound default → empty catalog
		out, err := tool.Execute(context.Background(), `{"reason":"list"}`, 1<<20)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if !strings.Contains(strings.ToLower(out), "no skills") {
			t.Errorf("empty catalog not reported as a result; got %q", out)
		}
	})

	t.Run("declares a well-formed schema and the reader-class timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewSkillsTool(nil)
		if tool.Name() != "list_skills" {
			t.Errorf("Name() = %q; want list_skills", tool.Name())
		}
		if tool.Contract().DefaultTimeout != readerDefaultTimeout {
			t.Errorf("DefaultTimeout = %v; want %v", tool.Contract().DefaultTimeout, readerDefaultTimeout)
		}
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if err := json.Unmarshal(tool.Parameters(), &schema); err != nil {
			t.Fatalf("schema is not a JSON object: %v", err)
		}
		if len(schema.Required) == 0 {
			t.Errorf("schema declares no required fields; want at least `reason`")
		}
		for _, r := range schema.Required {
			if _, ok := schema.Properties[r]; !ok {
				t.Errorf("required %q is not declared under properties (required ⊆ properties)", r)
			}
		}
	})
}
