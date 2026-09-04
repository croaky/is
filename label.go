package is

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// label returns the source text of argument arg in the assertion call
// that the frame skip levels up made, so a failure can name what it
// compared. It returns "" when the source can't be read, and for a
// literal argument, where "[]int{1, 2} = []int{1, 2}" is noise.
func label(skip int, name string, arg int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	// -trimpath, which the Checkfile's test entry passes, leaves a
	// module-relative path that won't open. go test runs each binary
	// in its own package directory, so the base name resolves.
	if !filepath.IsAbs(file) {
		file = filepath.Base(file)
	}
	f, fset := parseFile(file)
	if f == nil {
		return ""
	}
	call := callAt(f, fset, line, name)
	if call == nil || arg >= len(call.Args) {
		return ""
	}
	expr := call.Args[arg]
	if isLiteral(expr) {
		return ""
	}
	var b strings.Builder
	if err := printer.Fprint(&b, fset, expr); err != nil {
		return ""
	}
	// A multi-line argument keeps its newlines and indentation from the
	// source, which reads badly on one failure line.
	return strings.Join(strings.Fields(b.String()), " ")
}

// callAt returns the innermost call to name covering line. Requiring
// the callee's name rejects a same-named file from another package,
// which is what a trimmed path can resolve to. Matching a node whose
// range covers the line, rather than the line's own text, keeps the
// label for a call spread over several lines.
func callAt(f *ast.File, fset *token.FileSet, line int, name string) *ast.CallExpr {
	var best *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		start := fset.Position(call.Pos()).Line
		end := fset.Position(call.End()).Line
		if line < start || line > end {
			return true
		}
		if callName(call.Fun) != name {
			return true
		}
		if best == nil || call.Pos() > best.Pos() {
			best = call
		}
		return true
	})
	return best
}

func callName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	case *ast.IndexExpr:
		return callName(f.X)
	default:
		return ""
	}
}

func isLiteral(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit, *ast.CompositeLit, *ast.FuncLit:
		return true
	case *ast.Ident:
		return e.Name == "true" || e.Name == "false" || e.Name == "nil"
	default:
		return false
	}
}

var files sync.Map // file name -> *parsed

type parsed struct {
	once sync.Once
	file *ast.File
	fset *token.FileSet
}

func parseFile(name string) (*ast.File, *token.FileSet) {
	v, _ := files.LoadOrStore(name, &parsed{})
	p := v.(*parsed)
	p.once.Do(func() {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			return
		}
		p.file, p.fset = f, fset
	})
	return p.file, p.fset
}
