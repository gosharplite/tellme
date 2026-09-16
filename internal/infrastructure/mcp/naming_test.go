package mcp

import (
	"strings"
	"testing"
)

// T026 [UNIT] — deterministic naming: namespacing + the derived 64-byte budget
// with an 8-hex SHA-256 suffix (FR-004 / TD3).
func TestNamespacedName_ShortIsPlainNamespacing(t *testing.T) {
	if got := NamespacedName("shop", "lookup_price"); got != "mcp_shop_lookup_price" {
		t.Fatalf("got %q, want mcp_shop_lookup_price", got)
	}
}

func TestNamespacedName_LongIsBounded(t *testing.T) {
	server := strings.Repeat("s", 24)
	tool := strings.Repeat("t", 200)
	got := NamespacedName(server, tool)
	if len(got) > 64 {
		t.Fatalf("len(%q)=%d exceeds the 64-byte wire maximum", got, len(got))
	}
	if !strings.HasPrefix(got, "mcp_"+server+"_") {
		t.Fatalf("the namespaced prefix was lost: %q", got)
	}
	// the trailing 8 chars are the hex hash suffix (the byte before it is '_')
	hash := got[len(got)-8:]
	for _, r := range hash {
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Fatalf("the suffix %q is not an 8-hex SHA-256 prefix", hash)
		}
	}
}

func TestNamespacedName_TwoLongToolsDistinct(t *testing.T) {
	server := strings.Repeat("s", 24)
	prefix := strings.Repeat("t", 60) // beyond the retained prefix budget
	a := NamespacedName(server, prefix+"aaaa")
	b := NamespacedName(server, prefix+"bbbb")
	if a == b {
		t.Fatalf("two distinct long tools collided: %q", a)
	}
	if len(a) > 64 || len(b) > 64 {
		t.Fatalf("names exceed 64 bytes: %q %q", a, b)
	}
}

func TestNamespacedName_CrossServerDistinct(t *testing.T) {
	if NamespacedName("a", "tool") == NamespacedName("b", "tool") {
		t.Fatal("the same tool on two servers must not collide")
	}
}
