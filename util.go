package unitconst

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

// remove duplicated elements
func stringSet(ss []string) []string {
	m := map[string]bool{}
	var set []string
	for _, s := range ss {
		if !m[s] {
			set = append(set, s)
		}
		m[s] = true
	}
	return set
}

func exprToString(expr ast.Expr) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}
