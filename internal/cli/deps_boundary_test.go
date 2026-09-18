package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestDependenciesSeamIsNarrow pins review-deferral F-6 (round 051 / ADR 0020):
// every function that RECEIVES a deps.Dependencies parameter must read at most
// TWO of its fields — the narrow-seam rule — except the two composition-bearing
// functions (`run`, `runTurn`) that legitimately hold the bag (the falsifiable
// acceptance recorded by the PR #104 review: *"every function receiving
// Dependencies reads ≤2 of its fields."*).
//
// It is a source-level audit over the package's PRODUCTION files: for each
// function with a `deps.Dependencies` parameter, count the DISTINCT field
// selectors applied to that parameter.
func TestDependenciesSeamIsNarrow(t *testing.T) {
	// The composition bearers that legitimately hold the bag: `runTurn` (the turn
	// composition) and `renderTurn` (which forwards the bag to runTurn). `run`
	// takes `Options`, NOT a `deps.Dependencies`, so it is not exempt (round-051
	// review R-51-2).
	bagHolders := map[string]bool{"renderTurn": true, "runTurn": true}
	exemptSeen := map[string]bool{}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	type violation struct {
		fn    string
		field []string
	}
	var violations []violation
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if bagHolders[fd.Name.Name] && depsParamName(fd) != "" {
				exemptSeen[fd.Name.Name] = true
			}
			if fs, bad := wideBagReads(fd, bagHolders); bad {
				violations = append(violations, violation{fd.Name.Name, fs})
			}
			return true
		})
	}
	for name := range bagHolders {
		if !exemptSeen[name] {
			t.Fatalf("F-6 liveness: exempt composition bearer %q is no longer a function receiving deps.Dependencies — the exemption is dead (round-051 review R-51-2; the round-047 assertSanctionedInUse analogue)", name)
		}
	}
	if len(violations) > 0 {
		var b strings.Builder
		for _, v := range violations {
			b.WriteString(v.fn + ": " + strings.Join(v.field, ", ") + "\n")
		}
		t.Fatalf("F-6 violation — function(s) receiving deps.Dependencies read >2 fields:\n%s", b.String())
	}
}

// wideBagReads reports the distinct deps.Dependencies field selectors read by a
// function body when it exceeds the narrow-seam budget of two (false otherwise).
func wideBagReads(fd *ast.FuncDecl, bagHolders map[string]bool) ([]string, bool) {
	if fd.Recv != nil || fd.Body == nil || bagHolders[fd.Name.Name] {
		return nil, false
	}
	param := depsParamName(fd)
	if param == "" {
		return nil, false
	}
	fields := map[string]bool{}
	ast.Inspect(fd.Body, func(m ast.Node) bool {
		se, ok := m.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := se.X.(*ast.Ident); ok && id.Name == param {
			fields[se.Sel.Name] = true
		}
		return true
	})
	if len(fields) <= 2 {
		return nil, false
	}
	var fs []string
	for f := range fields {
		fs = append(fs, f)
	}
	sort.Strings(fs)
	return fs, true
}

// depsParamName returns the first parameter named with a deps.Dependencies type.
func depsParamName(fd *ast.FuncDecl) string {
	if fd.Type.Params == nil {
		return ""
	}
	for _, p := range fd.Type.Params.List {
		if isDepsDependencies(p.Type) && len(p.Names) > 0 {
			return p.Names[0].Name
		}
	}
	return ""
}

func isDepsDependencies(e ast.Expr) bool {
	se, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := se.X.(*ast.Ident)
	return ok && id.Name == "deps" && se.Sel.Name == "Dependencies"
}
