//go:build quality

// The compile-blind argument-swap gate. Everything lives in this build-tagged
// test file so it runs only under `-tags=quality` (the Makefile
// `verify-params` target) and never in the default `go test ./...` run. See
// docs/decisions/0068-parameter-shape-gate.md and specs/truth/techstack.md.
//
// The scan is the WITNESS (issue #210): per site it emits file:line, name, the
// arms that fired, and the structural type string it keyed on, and the report
// prints every baseline figure with its label and universe. The figures live in
// the script, never in prose — that is the durable lesson of the five drafts
// behind issue #210, where every wrong number came from memory.
package params

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

// updateAllowlist regenerates the committed allow-list from the gate's own
// output, refusing to GROW it (ADR 0020 fold R-51-1). Passed through the test
// binary: `go test ... -args -update-allowlist`.
var updateAllowlist = flag.Bool("update-allowlist", false,
	"regenerate tools/params/allowlist.txt from the gate's own output (refuses to grow)")

// Rule thresholds. They are constants, not flags: widening the rule is an ADR
// decision (docs/decisions/0068-parameter-shape-gate.md §D), never a knob.
const (
	armAMinArity = 4 // arm (a) only applies at n >= 4
	armARun      = 2 // ...and needs an adjacent identical-type run of >= 2
	armBRun      = 3 // arm (b): arity-independent run of >= 3
	armCRun      = 2 // arm (c): adjacent run of >= 2 func-typed parameters
	armDMinElts  = 2 // arm (d): >= 2 unkeyed elements in a named struct literal
)

// param is one flattened parameter: `a, b, c float64` is three params.
type param struct {
	key      string // structural type key (slices/chans/funcs expanded)
	isFunc   bool   // the parameter's type is a func type
	exempt   bool   // context.Context / any / interface{} (or `error` in results)
	variadic bool
}

// decl is one scanned named signature: a FuncDecl or a top-level interface
// method.
type decl struct {
	file string // module-relative, forward slashes
	line int
	name string
	// qual is the receiver type (a method) or the declaring interface (an
	// interface method); it qualifies the allow-list identity so two same-named
	// methods in one file cannot share an entry (review item 3).
	qual    string
	isIface bool
	isTest  bool
	n       int
	longest int    // longest adjacent identical-type run (exempt runs skipped)
	runKey  string // the type string the longest run keyed on
	funcRun int    // longest adjacent func-typed run (arm c)
	identFn int    // longest adjacent identical-signature func run
	resRun  int    // longest adjacent identical-type result run (recorded only)
	resKey  string
}

// site is one red: a declaration with arms fired, or an unkeyed struct literal.
type site struct {
	file  string
	line  int
	name  string
	arms  string
	shape string
	key   string
	// elts is an arm-(d) literal's element count. It is aggregated with its
	// same-identity siblings when rows collapse (see redSites).
	elts int
}

// tree is one scan of the whole tree.
type tree struct {
	decls   []decl
	unkeyed []site
	aliases int
}

// arms returns the arm letters that fire for a declaration, in `abcd` order.
func (d decl) arms() string {
	var b strings.Builder
	if d.n >= armAMinArity && d.longest >= armARun {
		b.WriteByte('a')
	}
	if d.longest >= armBRun {
		b.WriteByte('b')
	}
	if d.funcRun >= armCRun {
		b.WriteByte('c')
	}
	return b.String()
}

// shape renders the evidence the arms keyed on.
func (d decl) shape() string {
	parts := []string{fmt.Sprintf("n=%d", d.n)}
	switch {
	case d.longest >= armARun:
		parts = append(parts, fmt.Sprintf("run=%d %s", d.longest, d.runKey))
	case d.funcRun >= armCRun:
		parts = append(parts, fmt.Sprintf("func-run=%d", d.funcRun))
	default:
		parts = append(parts, "run=1")
	}
	return strings.Join(parts, " ")
}

// asSite converts a declaration into its red-set line.
func (d decl) asSite() site {
	marker := ""
	if d.isIface {
		marker = " [iface]"
	}
	return site{
		file:  d.file,
		line:  d.line,
		name:  d.display(),
		arms:  d.arms(),
		shape: d.shape(),
		key:   d.key() + marker,
	}
}

// key is the allow-list identity of a declaration: `file:name`, or
// `file:Recv.Method` / `file:Iface.Method` when the declaration is qualified. Line
// numbers are deliberately excluded so a cosmetic edit elsewhere in the file does
// not stale the entry; a rename or a move must re-justify the entry, which is the
// point. The qualifier is REQUIRED, not cosmetic: `internal/infrastructure/tools/
// writer.go` declares `Execute` on two receivers, and
// `internal/infrastructure/mcp/client.go` declares `Close` on two — an unqualified
// key would let one entry silently allow the other.
func (d decl) key() string {
	if d.qual == "" {
		return d.file + ":" + d.name
	}
	return d.file + ":" + d.qual + "." + d.name
}

// display is the name the report prints: qualified when it can be.
func (d decl) display() string {
	if d.qual == "" {
		return d.name
	}
	return d.qual + "." + d.name
}

// exemptKey reports whether a PARAMETER's type key is exempt from the swap-run
// rule. `error` is deliberately absent: it is exempt only in a RESULT list (the
// idiomatic trailing error), which newDecl marks there. Two adjacent `error`
// parameters swap just as silently as two adjacent ints.
func exemptKey(k string) bool {
	switch k {
	case "context.Context", "any", "interface{}":
		return true
	}
	return false
}

// typeKey renders a STRUCTURAL type key: slices, channels, maps and func
// signatures are expanded, never flattened to "func". A flattened key is what
// made three of four earlier drafts measure the wrong set (issue #210 §Provenance).
func typeKey(e ast.Expr) string {
	if k, ok := simpleKey(e); ok {
		return k
	}
	return compoundKey(e)
}

// simpleKey renders the one-token type shapes; ok=false defers to compoundKey.
func simpleKey(e ast.Expr) (string, bool) {
	switch t := e.(type) {
	case nil:
		return "?", true
	case *ast.Ident:
		return t.Name, true
	case *ast.SelectorExpr:
		return typeKey(t.X) + "." + t.Sel.Name, true
	case *ast.StarExpr:
		return "*" + typeKey(t.X), true
	case *ast.ParenExpr:
		return typeKey(t.X), true
	case *ast.Ellipsis:
		return "..." + typeKey(t.Elt), true
	}
	return "", false
}

// compoundKey renders the composite type shapes.
func compoundKey(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.ArrayType:
		return arrayKey(t)
	case *ast.MapType:
		return "map[" + typeKey(t.Key) + "]" + typeKey(t.Value)
	case *ast.ChanType:
		return chanKey(t)
	case *ast.FuncType:
		return "func(" + fieldsKey(t.Params) + ")(" + fieldsKey(t.Results) + ")"
	case *ast.InterfaceType:
		return interfaceKey(t)
	case *ast.StructType:
		return structKey(t)
	case *ast.IndexExpr:
		return typeKey(t.X) + "[" + typeKey(t.Index) + "]"
	case *ast.IndexListExpr:
		return typeKey(t.X) + "[" + indexListKey(t.Indices) + "]"
	default:
		return fmt.Sprintf("%T", e)
	}
}

// arrayKey renders a slice or a fixed-length array.
func arrayKey(t *ast.ArrayType) string {
	if t.Len == nil {
		return "[]" + typeKey(t.Elt)
	}
	return "[N]" + typeKey(t.Elt)
}

// interfaceKey renders an interface type (an empty one is `interface{}`).
func interfaceKey(t *ast.InterfaceType) string {
	if t.Methods == nil || len(t.Methods.List) == 0 {
		return "interface{}"
	}
	return "interface{…}"
}

// structKey renders a struct type (an empty one is `struct{}`).
func structKey(t *ast.StructType) string {
	if t.Fields == nil || len(t.Fields.List) == 0 {
		return "struct{}"
	}
	return "struct{…}"
}

// indexListKey renders a generic type's index list.
func indexListKey(indices []ast.Expr) string {
	parts := make([]string, 0, len(indices))
	for _, ix := range indices {
		parts = append(parts, typeKey(ix))
	}
	return strings.Join(parts, ", ")
}

// chanKey renders a channel type, direction included.
func chanKey(t *ast.ChanType) string {
	switch t.Dir {
	case ast.SEND:
		return "chan<- " + typeKey(t.Value)
	case ast.RECV:
		return "<-chan " + typeKey(t.Value)
	default:
		return "chan " + typeKey(t.Value)
	}
}

// fieldsKey renders a field list with runs flattened (`a, b int` -> `int, int`).
func fieldsKey(fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	var out []string
	for _, f := range fl.List {
		count := len(f.Names)
		if count == 0 {
			count = 1
		}
		k := typeKey(f.Type)
		for range count {
			out = append(out, k)
		}
	}
	return strings.Join(out, ", ")
}

// paramsOf flattens a parameter list, marking the trailing variadic.
func paramsOf(fl *ast.FieldList) []param {
	if fl == nil {
		return nil
	}
	var out []param
	for _, f := range fl.List {
		k := typeKey(f.Type)
		count := len(f.Names)
		if count == 0 {
			count = 1
		}
		_, isEllipsis := f.Type.(*ast.Ellipsis)
		for i := range count {
			out = append(out, param{
				key:      k,
				isFunc:   isFuncType(f.Type),
				exempt:   exemptKey(k),
				variadic: isEllipsis && i == count-1,
			})
		}
	}
	return out
}

// isFuncType reports whether a parameter's type is a func type.
func isFuncType(e ast.Expr) bool {
	if p, ok := e.(*ast.ParenExpr); ok {
		e = p.X
	}
	_, ok := e.(*ast.FuncType)
	return ok
}

// runsOf returns the longest adjacent identical-type run and the key it keyed
// on. Exempt parameters break a run (they are not "the same type" as a
// neighbour), and a trailing variadic is excluded (a lone `...T` tail is the
// idiomatic collector, not a swappable pair).
func runsOf(ps []param) (longest int, key string) {
	n := len(ps)
	if n > 0 && ps[n-1].variadic {
		n--
	}
	for i := 0; i < n; {
		if ps[i].exempt {
			i++
			continue
		}
		j := i
		for j+1 < n && !ps[j+1].exempt && ps[j+1].key == ps[i].key {
			j++
		}
		if l := j - i + 1; l >= armARun && l > longest {
			longest, key = l, ps[i].key
		}
		i = j + 1
	}
	return longest, key
}

// funcRunsOf returns the longest adjacent func-typed run and the longest run
// whose members are all the SAME func signature (recorded, expected 0).
func funcRunsOf(ps []param) (longest, identical int) {
	n := len(ps)
	for i := 0; i < n; {
		if !ps[i].isFunc {
			i++
			continue
		}
		j := i
		for j+1 < n && ps[j+1].isFunc {
			j++
		}
		l := j - i + 1
		if l >= armCRun {
			longest = max(longest, l)
			same := true
			for k := i + 1; k <= j; k++ {
				if ps[k].key != ps[i].key {
					same = false
					break
				}
			}
			if same {
				identical = max(identical, l)
			}
		}
		i = j + 1
	}
	return longest, identical
}

// declSpec is one named signature's scanned inputs. It is a named struct because
// this gate enforces exactly that idiom on the tree it scans (review item 5): the
// positional form is n = 8 with three adjacent identical-type runs — the gate
// would red on itself.
type declSpec struct {
	fset    *token.FileSet
	params  *ast.FieldList
	results *ast.FieldList
	file    string
	name    string
	qual    string
	pos     token.Pos
	isTest  bool
	isIface bool
}

// newDecl assembles one scanned declaration.
func newDecl(spec declSpec) decl {
	ps := paramsOf(spec.params)
	d := decl{
		file:    spec.file,
		line:    spec.fset.Position(spec.pos).Line,
		name:    spec.name,
		qual:    spec.qual,
		isIface: spec.isIface,
		isTest:  spec.isTest,
		n:       len(ps),
	}
	d.longest, d.runKey = runsOf(ps)
	d.funcRun, d.identFn = funcRunsOf(ps)
	res := paramsOf(spec.results)
	for i := range res {
		if res[i].key == "error" {
			res[i].exempt = true
		}
	}
	d.resRun, d.resKey = runsOf(res)
	return d
}

// declsOf extracts every named signature in one file: FuncDecls plus top-level
// interface methods (an inline interface type is NOT a TypeSpec of its own and
// is out of scope — issue #210 G4).
func declsOf(fset *token.FileSet, f *ast.File, file string, isTest bool) []decl {
	var out []decl
	for _, d := range f.Decls {
		switch n := d.(type) {
		case *ast.FuncDecl:
			out = append(out, newDecl(declSpec{
				fset:    fset,
				params:  n.Type.Params,
				results: n.Type.Results,
				file:    file,
				name:    n.Name.Name,
				qual:    receiverName(n.Recv),
				pos:     n.Pos(),
				isTest:  isTest,
			}))
		case *ast.GenDecl:
			out = append(out, interfaceMethods(fset, n, file, isTest)...)
		}
	}
	return out
}

// interfaceMethods returns the methods of every top-level interface TypeSpec.
func interfaceMethods(fset *token.FileSet, gd *ast.GenDecl, file string, isTest bool) []decl {
	if gd.Tok != token.TYPE {
		return nil
	}
	var out []decl
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok || it.Methods == nil {
			continue
		}
		for _, m := range it.Methods.List {
			ft, ok := m.Type.(*ast.FuncType)
			if !ok || len(m.Names) == 0 {
				continue
			}
			out = append(out, newDecl(declSpec{
				fset:    fset,
				params:  ft.Params,
				results: ft.Results,
				file:    file,
				name:    m.Names[0].Name,
				qual:    ts.Name.Name,
				pos:     m.Pos(),
				isTest:  isTest,
				isIface: true,
			}))
		}
	}
	return out
}

// receiverName returns a method's receiver type name, or "" for a plain function.
func receiverName(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	return structNameOf(fl.List[0].Type)
}

// declaredContainers collects the SAME-PACKAGE named array/slice/map types in one
// package directory (`type table []row`), so an elided literal under such a name
// can be attributed. Whole-file walk, like declaredStructs, so a function-local
// `type table []row` counts. A named type from another package is never here — it
// is opaque by design (resolving it needs that package's AST).
func declaredContainers(parsed map[string]*ast.File) map[string]containerDecl {
	out := map[string]containerDecl{}
	for _, f := range parsed {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			switch ts.Type.(type) {
			case *ast.ArrayType, *ast.MapType:
				out[ts.Name.Name] = containerDecl{expr: ts.Type, params: typeParamNames(ts)}
			}
			return true
		})
	}
	return out
}

// typeParamNames returns a generic type declaration's type-parameter names, in
// order (`type Table[T any] []T` → ["T"]).
func typeParamNames(ts *ast.TypeSpec) []string {
	if ts.TypeParams == nil {
		return nil
	}
	var out []string
	for _, f := range ts.TypeParams.List {
		for _, n := range f.Names {
			out = append(out, n.Name)
		}
	}
	return out
}

// declaredStructs collects the named struct types declared in one package
// directory (production and test files alike — an arm-(d) literal may name a
// struct declared in either). It walks the WHOLE file, so a FUNCTION-LOCAL
// `type note struct{…}` counts too — a test table declared beside its literal is
// the common shape (review item 2).
func declaredStructs(parsed map[string]*ast.File) map[string]bool {
	out := map[string]bool{}
	for _, f := range parsed {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, isStruct := ts.Type.(*ast.StructType); isStruct {
				out[ts.Name.Name] = true
			}
			return true
		})
	}
	return out
}

// structNameOf returns the declared (possibly generic) struct name a type
// expression names, or "" when it names no identifiable type — an anonymous
// struct, a slice, a map, a func. Pointers and parens are transparent, so a
// method receiver `*client` and a `*T{…}` literal both yield their base name.
func structNameOf(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ParenExpr:
		return structNameOf(t.X)
	case *ast.StarExpr:
		return structNameOf(t.X)
	case *ast.IndexExpr:
		return structNameOf(t.X)
	case *ast.IndexListExpr:
		return structNameOf(t.X)
	}
	return ""
}

// containerDecl is one same-package named container type: its underlying array /
// slice / map type and its type parameters (nil for a plain `type table []row`).
type containerDecl struct {
	expr   ast.Expr
	params []string
}

// bound is a resolved container: the underlying type expression plus any type
// substitution its instantiation implies.
type bound struct {
	expr  ast.Expr
	subst map[string]ast.Expr
}

// resolveContainer follows a type expression to a CONTAINER (array, slice or map)
// through the SAME-PACKAGE named container types — `type table []row` makes
// `table{{…}}` a table of `row`, and `elemType(Ident "table")` would otherwise be
// nil (re-review item 2). A GENERIC instantiation (`Table[Pair]`, `type Table[T
// any] []T`) is bound by substituting each type parameter with its argument, which
// is what makes `Table[Pair]{{…}}` attributable too.
//
// A named type from ANOTHER package (`pkg.Table`) is opaque by DESIGN, exactly like
// `time.Time{}`: resolving it would need that package's AST. That is the one
// recorded limit (RF-068-3b).
//
// The hop count is bounded so a pathological chain cannot spin; Go's own type
// rules make a container cycle impossible, so the bound is belt-and-braces.
func resolveContainer(e ast.Expr, named map[string]containerDecl) bound {
	var subst map[string]ast.Expr
	for range 8 {
		if p, ok := e.(*ast.ParenExpr); ok {
			e = p.X
			continue
		}
		decl, args, ok := containerLookup(e, named)
		if !ok {
			return bound{expr: e, subst: subst}
		}
		if len(decl.params) != len(args) {
			return bound{expr: e, subst: subst} // an uninstantiated generic: opaque
		}
		if len(args) > 0 {
			subst = make(map[string]ast.Expr, len(args))
			for i, p := range decl.params {
				subst[p] = args[i]
			}
		}
		e = decl.expr
	}
	return bound{expr: e, subst: subst}
}

// containerLookup resolves one named-container hop: a bare `table`, or an
// instantiation `Table[A]` / `Table[A, B]`. ok=false for anything else (a
// selector from another package, an unbound type parameter, a non-container).
func containerLookup(e ast.Expr, named map[string]containerDecl) (containerDecl, []ast.Expr, bool) {
	switch t := e.(type) {
	case *ast.Ident:
		decl, ok := named[t.Name]
		return decl, nil, ok
	case *ast.IndexExpr:
		id, ok := t.X.(*ast.Ident)
		if !ok {
			return containerDecl{}, nil, false
		}
		decl, ok := named[id.Name]
		return decl, []ast.Expr{t.Index}, ok
	case *ast.IndexListExpr:
		id, ok := t.X.(*ast.Ident)
		if !ok {
			return containerDecl{}, nil, false
		}
		decl, ok := named[id.Name]
		return decl, t.Indices, ok
	}
	return containerDecl{}, nil, false
}

// substitute resolves a bare type parameter through an instantiation's binding.
func substitute(e ast.Expr, subst map[string]ast.Expr) ast.Expr {
	if id, ok := e.(*ast.Ident); ok && subst != nil {
		if r, ok := subst[id.Name]; ok {
			return r
		}
	}
	return e
}

// elemType returns the TYPE EXPRESSION a container's elements have: `[]T` and
// `[N]T` yield `T`, `map[K]V` yields `V`. It returns the EXPRESSION rather than a
// name so NESTED containers resolve one level per elided literal —
// `[][]T{{{1, 2, 3, 4}}}` reaches T only at the third level.
func elemType(e ast.Expr, named map[string]containerDecl) ast.Expr {
	b := resolveContainer(e, named)
	switch t := b.expr.(type) {
	case *ast.ArrayType:
		return substitute(t.Elt, b.subst)
	case *ast.MapType:
		return substitute(t.Value, b.subst)
	}
	return nil
}

// keyType returns the TYPE EXPRESSION a map's keys have, or nil when e is not a
// map container.
func keyType(e ast.Expr, named map[string]containerDecl) ast.Expr {
	b := resolveContainer(e, named)
	if t, ok := b.expr.(*ast.MapType); ok {
		return substitute(t.Key, b.subst)
	}
	return nil
}

// allUnkeyed reports whether every element of a literal is positional.
func allUnkeyed(elts []ast.Expr) bool {
	if len(elts) == 0 {
		return false
	}
	for _, e := range elts {
		if _, keyed := e.(*ast.KeyValueExpr); keyed {
			return false
		}
	}
	return true
}

// effectiveLitTypes resolves, for every composite literal in the file, the type
// expression it denotes: its own `Type`, or — when the type is ELIDED — the type
// it inherits from its position inside an enclosing literal.
//
// It is a WORKLIST from the explicitly-typed literals outward, and it assigns a
// map's KEY position the map's key type and its value position the element type.
// One attribution per literal: the key of `map[T]T{{…}: {…}}` is ONE site and its
// value is another — a stack walk that gave every child the value type counted
// the key twice once key and value were the same struct (re-review item 3).
//
// A literal with neither an own type nor a container position (an anonymous
// `struct{…}{…}`) maps to nil — deliberately: Go permits type elision ONLY in an
// array/slice/map literal's element or key position, so those are the only
// positions this can reach.
func effectiveLitTypes(f *ast.File, named map[string]containerDecl) map[*ast.CompositeLit]ast.Expr {
	out := map[*ast.CompositeLit]ast.Expr{}
	var queue []*ast.CompositeLit
	ast.Inspect(f, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok && cl.Type != nil {
			out[cl] = cl.Type
			queue = append(queue, cl)
		}
		return true
	})
	assign := func(child ast.Expr, t ast.Expr) {
		cl, ok := child.(*ast.CompositeLit)
		if !ok || cl.Type != nil || t == nil {
			return
		}
		out[cl] = t
		queue = append(queue, cl)
	}
	for len(queue) > 0 {
		cl := queue[0]
		queue = queue[1:]
		eff := out[cl]
		et, kt := elemType(eff, named), keyType(eff, named)
		for _, e := range cl.Elts {
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				assign(kv.Key, kt)
				assign(kv.Value, et)
				continue
			}
			assign(e, et)
		}
	}
	return out
}

// armDSites returns every arm-(d) site in one file: same-package named-struct
// literals with >= 2 unkeyed elements. It covers every position Go allows a type
// to be elided in —
//
//	[]T{{1, 2}}                 // list element, elided (the test-table shape)
//	map[K]T{"k": {1, 2}}        // map VALUE, elided
//	[][]T{{{1, 2}}}             // nested container, elided twice
//	map[T]int{{1, 2}: 1}        // struct-typed map KEY, elided
//	T{1, 2} / &T{1, 2}          // named directly (incl. a pointer element)
//	pair[int]{1, 2}             // generic instantiation
//	table{{1, 2}}               // SAME-PACKAGE named container (type table []T)
//
// …plus a literal that names no type and inherits none is skipped: it is an
// anonymous struct literal, not a same-package named one. Field-KEYED literals are
// correctly not flagged.
func armDSites(fset *token.FileSet, f *ast.File, file string, structs map[string]bool, named map[string]containerDecl) []site {
	var out []site
	for cl, t := range effectiveLitTypes(f, named) {
		name := structNameOf(t)
		if name == "" || !structs[name] || !allUnkeyed(cl.Elts) || len(cl.Elts) < armDMinElts {
			continue
		}
		line := fset.Position(cl.Pos()).Line
		out = append(out, site{
			file: file,
			line: line,
			name: name,
			arms: "d",
			elts: len(cl.Elts),
			key:  fmt.Sprintf("%s:%d:%s", file, line, name),
		})
	}
	return out
}

// countAliases counts type aliases (`type A = B`) — 0 expected; a non-zero count
// would make the syntactic type identity a false-negative source (issue #210 G3).
func countAliases(f *ast.File) int {
	count := 0
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if ok && ts.Assign != token.NoPos {
				count++
			}
		}
	}
	return count
}

// goFilesUnder returns the .go files under dir, grouped by directory.
func goFilesUnder(t *testing.T, dir string) map[string][]string {
	t.Helper()
	dirs := map[string][]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parent := filepath.Dir(path)
		dirs[parent] = append(dirs[parent], path)
		return nil
	})
	if err != nil {
		t.Fatalf("cannot walk %s: %v", dir, err)
	}
	return dirs
}

// scanTree scans the gate's governed subtree: internal/**.
func scanTree(t *testing.T, root string) tree {
	return scanSubtree(t, scanTarget{root: root, sub: "internal"})
}

// scanTarget names one subtree to scan. It is a NAMED struct and not a
// `root, sub string` pair because this gate must satisfy its own arms, and the
// self-application pin scans the very file these functions live in.
type scanTarget struct {
	root string // module root
	sub  string // module-relative subtree, e.g. "internal"
	// pinProduction treats `_test.go` files as PRODUCTION. The self-application pin
	// needs it: every line of this package lives in params_test.go, and redSites()
	// skips test declarations — without it the pin could only ever observe arm (d)
	// on the gate's own source, so it would not have redded on the OLD newDecl
	// either (re-review residual 1). It is never set by the gate itself.
	pinProduction bool
}

// scanSubtree parses one subtree and returns the declarations, the arm-(d) sites,
// and the alias count. Build-tag-gated files are parsed as written (the scan is
// text-level; the union over tags is not simulated).
func scanSubtree(t *testing.T, target scanTarget) tree {
	t.Helper()
	fset := token.NewFileSet()
	var out tree
	dirs := goFilesUnder(t, filepath.Join(target.root, target.sub))
	for _, dir := range slices.Sorted(maps.Keys(dirs)) {
		files := dirs[dir]
		slices.Sort(files)
		parsed := make(map[string]*ast.File, len(files))
		for _, p := range files {
			f, err := parser.ParseFile(fset, p, nil, 0)
			if err != nil {
				t.Fatalf("cannot parse %s: %v", p, err)
			}
			parsed[p] = f
		}
		structs := declaredStructs(parsed)
		named := declaredContainers(parsed)
		for _, p := range files {
			f := parsed[p]
			file := relPath(target.root, p)
			isTest := strings.HasSuffix(p, "_test.go") && !target.pinProduction
			out.decls = append(out.decls, declsOf(fset, f, file, isTest)...)
			out.unkeyed = append(out.unkeyed, armDSites(fset, f, file, structs, named)...)
			if !isTest { // figure Q's universe is production code
				out.aliases += countAliases(f)
			}
		}
	}
	return out
}

// relPath makes a path module-relative with forward slashes.
func relPath(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

// redSites returns the sorted red set: production declarations whose arms fire,
// plus every arm-(d) call-site literal (production AND test files — the
// same-package literals `go vet`'s composites analyser cannot see).
func (tr tree) redSites() []site {
	var out []site
	for _, d := range tr.decls {
		if d.isTest || d.arms() == "" {
			continue
		}
		out = append(out, d.asSite())
	}
	// Literals sharing one identity (file:line:Type) collapse to ONE row: that is
	// the allow-list identity, so a duplicate row would be noise and the report
	// must stay a diff against it. The row still SAYS how many literals it stands
	// for, so `[]T{{…}, {…}}` cannot read as a single literal (re-review nit).
	order := make([]string, 0, len(tr.unkeyed))
	agg := make(map[string]*litAgg, len(tr.unkeyed))
	for _, u := range tr.unkeyed {
		a, ok := agg[u.key]
		if !ok {
			a = &litAgg{site: u, minElts: u.elts, maxElts: u.elts}
			agg[u.key] = a
			order = append(order, u.key)
		}
		a.count++
		a.minElts = min(a.minElts, u.elts)
		a.maxElts = max(a.maxElts, u.elts)
	}
	for _, k := range order {
		a := agg[k]
		a.shape = fmt.Sprintf("%d unkeyed elements", a.minElts)
		if a.count > 1 {
			a.shape = fmt.Sprintf("%d literals, %d-%d unkeyed elements", a.count, a.minElts, a.maxElts)
		}
		out = append(out, a.site)
	}
	// name is the final tiebreak: two different struct types on one line would
	// otherwise print (and land in the allow-list) in map-iteration order.
	sort.Slice(out, func(i, j int) bool {
		if out[i].file != out[j].file {
			return out[i].file < out[j].file
		}
		if out[i].line != out[j].line {
			return out[i].line < out[j].line
		}
		return out[i].name < out[j].name
	})
	return out
}

// litAgg accumulates the arm-(d) literals that share one allow-list identity.
type litAgg struct {
	site
	count   int
	minElts int
	maxElts int
}

// count returns how many production declarations match a predicate, and how many
// interface methods do.
func (tr tree) count(isIface bool, pred func(decl) bool) int {
	n := 0
	for _, d := range tr.decls {
		if d.isTest || d.isIface != isIface || !pred(d) {
			continue
		}
		n++
	}
	return n
}

// figure is one labelled baseline figure: its label, its universe, and how to
// count it.
type figure struct {
	label    string
	universe string
	count    func(tr tree) int
}

// figures is the labelled baseline. Every number the plan quotes comes from
// here — never from prose (issue #210 §Provenance).
func figures() []figure {
	anyDecl := func(decl) bool { return true }
	return []figure{
		{label: "A. FuncDecl total (incl. zero-param)", universe: "internal/**, _test.go excluded, FuncDecl only", count: func(tr tree) int {
			return tr.count(false, anyDecl)
		}},
		{label: "B. ...of which zero-parameter", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n == 0 })
		}},
		{label: "C. ...with >= 1 parameter", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n > 0 })
		}},
		{label: "C'. ...with >= 1 parameter — interface methods", universe: "internal/** top-level interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.n > 0 })
		}},
		{label: "D. n >= 5", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n >= 5 })
		}},
		{label: "E. n >= 6", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n >= 6 })
		}},
		{label: "E'. n >= 6 — interface methods", universe: "internal/** interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.n >= 6 })
		}},
		{label: "F. n >= 4 AND run >= 2", universe: "internal/** FuncDecl, _test.go excluded, ctx/any runs exempt", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n >= 4 && d.longest >= armARun })
		}},
		{label: "F'. n >= 4 AND run >= 2 — interface methods", universe: "internal/** interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.n >= 4 && d.longest >= armARun })
		}},
		{label: "G. n >= 5 AND run >= 2", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n >= 5 && d.longest >= armARun })
		}},
		{label: "G'. n >= 5 AND run >= 2 — interface methods", universe: "internal/** interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.n >= 5 && d.longest >= armARun })
		}},
		{label: "H. longest run >= 3", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.longest >= armBRun })
		}},
		{label: "I. any run >= 2", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.longest >= armARun })
		}},
		{label: "J. RED — arms (a)/(b)", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return strings.ContainsAny(d.arms(), "ab") })
		}},
		{label: "J'. RED — arms (a)/(b) — interface methods", universe: "internal/** interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return strings.ContainsAny(d.arms(), "ab") })
		}},
		{label: "K. RED — arms (a)/(b)/(c)", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.arms() != "" })
		}},
		{label: "K'. RED — arms (a)/(b)/(c) — interface methods", universe: "internal/** interface methods, _test.go excluded", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.arms() != "" })
		}},
		{label: "L. RED — arm (d), same-package unkeyed literals", universe: "internal/** incl. _test.go, one row per file:line:Type", count: func(tr tree) int {
			// Counted as SITES (deduped by the allow-list identity), so L agrees
			// with the red set row for row.
			keys := map[string]bool{}
			for _, u := range tr.unkeyed {
				keys[u.key] = true
			}
			return len(keys)
		}},
		{label: "M. remainder — _test.go arms (a)/(b)/(c)", universe: "internal/**/*_test.go (out of scope for the gate)", count: func(tr tree) int {
			n := 0
			for _, d := range tr.decls {
				if d.isTest && d.arms() != "" {
					n++
				}
			}
			return n
		}},
		{label: "N. remainder — interface methods below the arm-(a) arity", universe: "internal/**: n <= 3 AND run >= 2", count: func(tr tree) int {
			return tr.count(true, func(d decl) bool { return d.n <= 3 && d.longest >= armARun })
		}},
		{label: "O. remainder — FuncDecl n <= 3 AND run == 2", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.n <= 3 && d.longest == armARun })
		}},
		{label: "P. remainder — FuncDecl with an identical-type RESULT run (G2)", universe: "internal/** FuncDecl, _test.go excluded, `error` exempt", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.resRun >= armARun })
		}},
		{label: "Q. type aliases (G3: 0 expected)", universe: "internal/**, _test.go excluded", count: func(tr tree) int {
			return tr.aliases
		}},
		{label: "R. identical-signature func runs (arm (c) is NOT a swap rule)", universe: "internal/** FuncDecl, _test.go excluded", count: func(tr tree) int {
			return tr.count(false, func(d decl) bool { return d.identFn >= armCRun })
		}},
	}
}

// renderFigures prints every figure with its label and universe — the standing
// requirement of issue #210 (a number without its universe is how four drafts
// drifted).
func renderFigures(tr tree) string {
	var b strings.Builder
	b.WriteString("tellme parameter-shape baseline — every figure is labelled with its universe\n")
	for _, f := range figures() {
		fmt.Fprintf(&b, "  %-62s %4d   [%s]\n", f.label, f.count(tr), f.universe)
	}
	return b.String()
}

// renderRedSet prints the red set in the gate's own `file:line name arms shape`
// form — the gate's first red run IS the specification.
func renderRedSet(sites []site) string {
	var b strings.Builder
	for _, s := range sites {
		fmt.Fprintf(&b, "%-58s %-28s arms=%s  %s\n", fmt.Sprintf("%s:%d", s.file, s.line), s.name, s.arms, s.shape)
	}
	return b.String()
}

// readAllowlist parses the committed allow-list. Every entry MUST carry a
// recorded reason; a malformed or reasonless entry is a hard failure (never
// "no allow-list configured").
func readAllowlist(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("allow-list unreadable (%s): %v — an absent/unreadable allow-list MUST fail, never be treated as 'nothing allowed'", path, err)
	}
	out := map[string]string{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, reason, _ := strings.Cut(line, "#")
		key = strings.TrimSpace(key)
		reason = strings.TrimSpace(reason)
		if key == "" {
			t.Fatalf("malformed allow-list line %q in %s (want `<file>:<name>  # <reason>`)", line, path)
		}
		if reason == "" {
			t.Fatalf("allow-list entry %q in %s carries no recorded reason — every exemption must be justified", key, path)
		}
		out[key] = reason
	}
	return out
}

// writeAllowlist regenerates the allow-list from the gate's own output. It
// refuses to GROW (ADR 0020 fold R-51-1): a new violation must be FIXED, not
// baselined.
func writeAllowlist(t *testing.T, path string, sites []site, prev map[string]string) {
	t.Helper()
	var added []string
	for _, s := range sites {
		if _, ok := prev[s.key]; !ok {
			added = append(added, fmt.Sprintf("%s (%s arms=%s)", s.key, s.name, s.arms))
		}
	}
	if len(added) > 0 {
		t.Fatalf("refusing to GROW the allow-list (%d -> %d): new red(s) %v — a new swap hazard must be REFACTORED, not allow-listed (ADR 0020 fold R-51-1)", len(prev), len(sites), added)
	}
	var b strings.Builder
	b.WriteString("# tools/params/allowlist.txt — parameter-shape gate allow-list (ADR 0068).\n")
	b.WriteString("#\n")
	b.WriteString("# Generated from the gate's own output — NEVER hand-edit. Regenerate with:\n")
	b.WriteString("#   make verify-params-update\n")
	b.WriteString("#\n")
	b.WriteString("# A line `<file>:<name>  # <reason>` exempts ONE known, reviewed red. The list is a\n")
	b.WriteString("# fail-on-stale ratchet: an entry that no longer matches a violation FAILS the gate,\n")
	b.WriteString("# and -update-allowlist refuses to GROW it (ADR 0020 fold R-51-1).\n")
	b.WriteString("#\n")
	if len(sites) == 0 {
		b.WriteString("# Empty at round 098 (option A: every site fixed).\n")
	}
	for _, s := range sites {
		fmt.Fprintf(&b, "%s  # %s arms=%s — %s\n", s.key, s.name, s.arms, prev[s.key])
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("cannot write allow-list %s: %v", path, err)
	}
}

// diffAllowlist returns the fresh (red but not allow-listed) and stale
// (allow-listed but no longer red) entries.
func diffAllowlist(sites []site, allow map[string]string) (fresh, stale []string) {
	red := make(map[string]bool, len(sites))
	for _, s := range sites {
		red[s.key] = true
		if _, ok := allow[s.key]; !ok {
			fresh = append(fresh, fmt.Sprintf("%s:%d %s arms=%s  %s", s.file, s.line, s.name, s.arms, s.shape))
		}
	}
	for k := range allow {
		if !red[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(fresh)
	sort.Strings(stale)
	return fresh, stale
}

// moduleRoot walks up from the test process's CWD to the directory holding
// go.mod (a Go test's CWD is its package directory, inside the module).
func moduleRoot(t *testing.T) string {
	t.Helper()
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		starts = append(starts, filepath.Dir(file))
	}
	for _, start := range starts {
		dir := start
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	t.Fatalf("go.mod not found above the test CWD or the gate's source path")
	return ""
}

// assertEnumeration asserts the scan is anchored to the whole tree, not to one
// subdirectory (a vacuous green otherwise).
func assertEnumeration(t *testing.T, tr tree) {
	t.Helper()
	const minDecls = 400
	if len(tr.decls) < minDecls {
		t.Fatalf("scan produced only %d declaration(s); expected the whole tree (>= %d) — is the scan anchored to internal/**?", len(tr.decls), minDecls)
	}
	seen := map[string]bool{}
	for _, d := range tr.decls {
		seen[path.Dir(d.file)] = true
	}
	for _, want := range []string{"internal/cli", "internal/ui", "internal/domain/render", "internal/infrastructure/tools"} {
		if !seen[want] {
			t.Fatalf("scan is missing known package %q — the scan is not the whole tree", want)
		}
	}
}

// parseSynthetic parses one in-memory source file. The gate's rule logic is
// pinned on SYNTHETIC input so a claim like "arm (a) fires at n = 4" is carried by
// the gate's own code, not by the shape of today's tree (ADR 0041 D2: a check must
// demonstrate a catch).
func parseSynthetic(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "synthetic.go", src, 0)
	if err != nil {
		t.Fatalf("synthetic source does not parse: %v\n%s", err, src)
	}
	return fset, f
}

// selfTestArms pins arms (a)/(b)/(c), the interface scope, and the three
// exemptions on synthetic signatures.
func selfTestArms(t *testing.T) {
	t.Helper()
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"arity-4 + two adjacent strings → (a)",
			"package p\nfunc f(a, b string, c int, d bool) {}\n", "a"},
		{"run of 3 at arity 3 → (b), arity-independent",
			"package p\nfunc f(a, b, c int) {}\n", "b"},
		{"run of 4 → (a) and (b)",
			"package p\nfunc f(a, b, c, d float64) {}\n", "ab"},
		{"no adjacent pair → clean",
			"package p\nfunc f(a string, b int, c string, d int) {}\n", ""},
		{"non-scalar run (io.ReadCloser) still reds",
			"package p\nfunc f(a int, b, c io.ReadCloser, d int) {}\n", "a"},
		{"two adjacent closures → (c)",
			"package p\nfunc f(a func(), b func(int)) {}\n", "c"},
		{"identical-signature closures are a run too",
			"package p\nfunc f(a func(), b func()) {}\n", "c"},
		{"a run of context.Context is exempt",
			"package p\nfunc f(a, b context.Context, c int, d bool) {}\n", ""},
		{"a run of any is exempt",
			"package p\nfunc f(a, b any, c int, d bool) {}\n", ""},
		{"a lone trailing variadic does not extend a run",
			"package p\nfunc f(a string, rest ...string) {}\n", ""},
		{"an interface method is scanned",
			"package p\ntype I interface { M(a, b, c int, d string) }\n", "ab"},
	}
	for _, tc := range cases {
		fset, f := parseSynthetic(t, tc.src)
		ds := declsOf(fset, f, "synthetic.go", false)
		if len(ds) != 1 {
			t.Fatalf("%s: expected exactly 1 declaration, got %d", tc.name, len(ds))
		}
		if got := ds[0].arms(); got != tc.want {
			t.Errorf("%s: arms = %q, want %q (%s)", tc.name, got, tc.want, ds[0].shape())
		}
	}
}

// selfTestUnkeyed pins arm (d) on EVERY valid-Go position a type may be elided in —
// an explicit type, a list/map element (the test-table shape), a NESTED container,
// a struct-typed MAP KEY, a SAME-PACKAGE NAMED container (`type table []pair`) and
// its generic form (`type gen[T any] []T`; `gen[pair]{{…}}`) — plus the negatives
// that keep the arm honest: a field-KEYED literal, a cross-package named STRUCT
// (`time.Time{}`) and a cross-package named CONTAINER (`other.Table{{…}}`).
func selfTestUnkeyed(t *testing.T) {
	t.Helper()
	src := `package p

type pair struct{ a, b int }

var explicit = pair{1, 2}
var keyed = pair{a: 1, b: 2}
var single = pair{a: 1}
var foreign = time.Time{}
var listOfElided = []pair{{1, 2}, {3, 4}}
var mapOfElided = map[string]pair{"k": {1, 2}}
var ptrOfElided = []*pair{{1, 2}}
var arrayOfElided = [2]pair{{1, 2}, {3, 4}}
var nested = [][]pair{{{1, 2}}}
var mapKey = map[pair]int{{1, 2}: 1}
var keyedList = []pair{{a: 1, b: 2}}

type table []pair
type pairMap map[string]pair

var namedSlice = table{{1, 2}}
var namedMap = pairMap{"k": {1, 2}}
var mapBoth = map[pair]pair{{1, 2}: {3, 4}}
var foreignNamed = other.Table{{1, 2}}

type gen[T any] []T
type genMap[K comparable, V any] map[K]V

var namedGeneric = gen[pair]{{1, 2}}
var namedGenericMap = genMap[string, pair]{"k": {1, 2}}
`
	fset, f := parseSynthetic(t, src)
	structs := declaredStructs(map[string]*ast.File{"synthetic.go": f})
	// Through redSites, so the identity collapse (and the "N literals" shape) is
	// pinned too, not only the attribution.
	rows := tree{unkeyed: armDSites(fset, f, "synthetic.go", structs, declaredContainers(map[string]*ast.File{"synthetic.go": f}))}.redSites()
	var got []string
	for _, s := range rows {
		got = append(got, fmt.Sprintf("%d:%s:%d", s.line, s.name, s.elts))
	}
	// One row per line (redSites collapses shared identities). 20/21 are the NAMED
	// container types, 28/29 their GENERIC forms; 22 is the map whose key AND value
	// are the same struct — one row, but it must SAY two literals (asserted below);
	// 23 (`other.Table`), 8 (`time.Time`) and 15 (field-keyed) must stay silent.
	want := []string{"5:pair:2", "9:pair:2", "10:pair:2", "11:pair:2", "12:pair:2", "13:pair:2",
		"14:pair:2", "20:pair:2", "21:pair:2", "22:pair:2", "28:pair:2", "29:pair:2"}
	if !slices.Equal(got, want) {
		t.Fatalf("arm (d) self-test:\n got  %v\n want %v\n(every elided shape must be attributed; the keyed literals and the\ncross-package selectors must not)", got, want)
	}
	// The map whose key and value are the same struct is TWO literals, so its row
	// must stand for two — the double-count bug made it say three (re-review item 3).
	var both string
	for _, r := range rows {
		if r.line == 22 {
			both = r.shape
		}
	}
	if want := "2 literals, 2-2 unkeyed elements"; both != want {
		t.Fatalf("arm (d) map key+value self-test: shape = %q, want %q", both, want)
	}
}

// selfTestUnkeyedLocalStruct pins the second half of residual 2: a struct type
// declared INSIDE a function must be collected, because that is where a test table
// is usually declared.
func selfTestUnkeyedLocalStruct(t *testing.T) {
	t.Helper()
	src := `package p

func f() {
	type note struct{ attempt, total int }
	_ = []note{{2, 3}, {3, 3}}
}
`
	fset, f := parseSynthetic(t, src)
	structs := declaredStructs(map[string]*ast.File{"synthetic.go": f})
	rows := tree{unkeyed: armDSites(fset, f, "synthetic.go", structs, declaredContainers(map[string]*ast.File{"synthetic.go": f}))}.redSites()
	if len(rows) != 1 || rows[0].name != "note" || rows[0].arms != "d" {
		t.Fatalf("arm (d) local-struct self-test: got %+v, want exactly the function-local `note` table", rows)
	}
	// Two literals on one line share an identity, so the row must SAY it stands for
	// two (re-review nit).
	if want := "2 literals, 2-2 unkeyed elements"; rows[0].shape != want {
		t.Fatalf("arm (d) collapse self-test: shape = %q, want %q", rows[0].shape, want)
	}
}

// selfTestAllowlist pins the fail-on-stale diff on synthetic sites.
func selfTestAllowlist(t *testing.T) {
	t.Helper()
	sites := []site{{file: "a.go", line: 1, name: "F", arms: "a", key: "a.go:F"}}
	if fresh, stale := diffAllowlist(sites, map[string]string{"a.go:F": "known"}); len(fresh) != 0 || len(stale) != 0 {
		t.Fatalf("allow-list self-test: an allow-listed site reported fresh=%v stale=%v", fresh, stale)
	}
	if fresh, _ := diffAllowlist(sites, map[string]string{}); len(fresh) != 1 {
		t.Fatalf("allow-list self-test: an unlisted site must be fresh, got %v", fresh)
	}
	if _, stale := diffAllowlist(nil, map[string]string{"a.go:F": "known"}); len(stale) != 1 || stale[0] != "a.go:F" {
		t.Fatalf("allow-list self-test: an entry with no violation must be stale, got %v", stale)
	}
}

// selfTestSelfApplication proves the gate satisfies the arms it enforces: its OWN
// package is scanned and must report no site. It is a PIN, not a widened walk —
// the governed scope stays `internal/**` (issue #210 §scope).
func selfTestSelfApplication(t *testing.T) {
	t.Helper()
	root := moduleRoot(t)
	pin := scanTarget{root: root, sub: "tools/params", pinProduction: true}
	if sites := scanSubtree(t, pin).redSites(); len(sites) != 0 {
		t.Fatalf("the gate reds on its own source (it does not satisfy its own arms):\n%s", renderRedSet(sites))
	}
}

// TestVerifyParameterShapes is the gate: red on every non-allow-listed
// swap-hazard site, and red on a stale allow-list entry.
func TestVerifyParameterShapes(t *testing.T) {
	root := moduleRoot(t)
	tr := scanTree(t, root)
	assertEnumeration(t, tr)
	t.Logf("\n%s", renderFigures(tr))

	// The rule logic is pinned on synthetic input BEFORE the tree is judged, so a
	// deleted or renamed self-test reds instead of silently passing (the
	// tools/arch name-set defence).
	want := []string{"arms", "unkeyed", "unkeyed-local-struct", "allow-list", "self-application"}
	var ran []string
	run := func(name string, fn func(*testing.T)) {
		t.Run(name, fn)
		ran = append(ran, name)
	}
	run("arms", selfTestArms)
	run("unkeyed", selfTestUnkeyed)
	run("unkeyed-local-struct", selfTestUnkeyedLocalStruct)
	run("allow-list", selfTestAllowlist)
	run("self-application", selfTestSelfApplication)
	if !slices.Equal(ran, want) {
		t.Fatalf("internal error: expected self-tests %v to run, got %v", want, ran)
	}

	sites := tr.redSites()
	path := filepath.Join(root, "tools", "params", "allowlist.txt")
	prev := readAllowlist(t, path)
	if *updateAllowlist {
		writeAllowlist(t, path, sites, prev)
		t.Logf("allow-list regenerated from the gate: %d entr(ies)", len(sites))
		return
	}

	fresh, stale := diffAllowlist(sites, prev)
	if len(fresh) > 0 {
		t.Errorf("parameter-shape gate: %d unallow-listed swap-hazard site(s):\n  %s", len(fresh), strings.Join(fresh, "\n  "))
	}
	if len(stale) > 0 {
		t.Errorf("parameter-shape gate: %d stale allow-list entr(ies) in %s — remove them:\n  %s", len(stale), path, strings.Join(stale, "\n  "))
	}
	if len(sites) == 0 && len(prev) == 0 {
		t.Logf("parameter-shape gate: %d site(s), allow-list empty — GREEN", len(sites))
	}
}

// TestParameterShapeReport prints the full labelled report (figures + the red
// set with arms). It never fails; it is the script that OWNS the numbers the
// plan quotes. Run it with:
//
//	go test -count=1 -v -tags=quality -run TestParameterShapeReport ./tools/params
func TestParameterShapeReport(t *testing.T) {
	root := moduleRoot(t)
	tr := scanTree(t, root)
	sites := tr.redSites()

	var b strings.Builder
	b.WriteString(renderFigures(tr))
	b.WriteString("\n")
	fmt.Fprintf(&b, "RED SET — %d site(s)\n", len(sites))
	b.WriteString(renderRedSet(sites))
	b.WriteString("\n")
	byArms := map[string]int{}
	byPkg := map[string]int{}
	for _, s := range sites {
		byArms[s.arms]++
		parts := strings.Split(s.file, "/")
		byPkg[strings.Join(parts[:2], "/")]++
	}
	fmt.Fprintf(&b, "arms: %v\n", sortedCounts(byArms))
	fmt.Fprintf(&b, "packages: %v\n", sortedCounts(byPkg))
	t.Logf("\n%s", b.String())
}

// sortedCounts renders a count map in a stable order.
func sortedCounts(m map[string]int) string {
	keys := slices.Sorted(maps.Keys(m))
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(parts, " ")
}
