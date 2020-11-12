package unitconst

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var (
	flagTypes string
)

func init() {
	Analyzer.Flags.StringVar(&flagTypes, "type", "time.Duration", "type of constant(comma separated)")
}

const doc = "unitconst finds using untyped constant as specified type"

// Analyzer finds using untyped constant as specified type.
var Analyzer = &analysis.Analyzer{
	Name: "unitconst",
	Doc:  doc,
	Run:  new(analyzer).run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

type hashedType struct {
	Org   *types.Named
	Name  string
	Under string
}

type analyzer struct {
	pass    *analysis.Pass
	inspect *inspector.Inspector
	types   []*hashedType
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	if err := a.init(pass); err != nil {
		return nil, err
	}

	if len(a.types) == 0 {
		return nil, nil
	}

	exprs := a.constExprs()
	fset := token.NewFileSet()
	td := &tmplData{
		Exprs: exprs,
		Types: a.types,
	}

	f, err := a.parse(fset, td)
	if err != nil {
		return nil, err
	}

	_, info, err := a.typeCheck(fset, f)
	if err != nil {
		return nil, err
	}

	done := map[token.Pos]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 {
			return true
		}

		ident := spec.Names[0]
		if ident.Name[0] != 'v' {
			return true
		}

		t := info.TypeOf(ident)
		if a.isTarget(t) {
			return true
		}

		p, err := strconv.Atoi(ident.Name[1:])
		if err != nil {
			return true
		}
		pos := token.Pos(p)

		if !done[pos] {
			pass.Reportf(pos, "must not use a untyped constant without a unit")
		}
		done[pos] = true

		return true
	})

	return nil, nil
}

func (a *analyzer) init(pass *analysis.Pass) error {
	a.pass = pass
	a.inspect = pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ts, err := a.getTypes(a.pass.Pkg, stringSet(strings.Split(flagTypes, ",")))
	if err != nil {
		return err
	}
	a.types = make([]*hashedType, len(ts))
	hasher := typeutil.MakeHasher()
	for i := range ts {
		h := hasher.Hash(ts[i])
		a.types[i] = &hashedType{
			Org:   ts[i],
			Name:  fmt.Sprintf("t%d", h),
			Under: analysisutil.Under(ts[i]).String(),
		}
	}
	return nil
}

func (a *analyzer) hash(t types.Type) *hashedType {
	for i := range a.types {
		if types.Identical(a.types[i].Org, t) {
			return a.types[i]
		}
	}
	return nil
}

func (a *analyzer) getTypes(pkg *types.Package, names []string) ([]*types.Named, error) {
	pkgToTypes := map[string][]string{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		ss := strings.Split(n, ".")
		if len(ss) != 2 {
			return nil, fmt.Errorf("invalid type: %s", n)
		}
		pkgToTypes[ss[0]] = append(pkgToTypes[ss[0]], ss[1])
	}

	var ts []*types.Named
	for _, p := range append(pkg.Imports(), pkg) {
		tnames, ok := pkgToTypes[p.Path()]
		if !ok {
			continue
		}

		for _, tn := range tnames {
			if obj := p.Scope().Lookup(tn); obj != nil {
				if named, ok := obj.Type().(*types.Named); ok {
					ts = append(ts, named)
				}
			}
		}
	}
	return ts, nil
}

func (a *analyzer) constExprs() map[token.Pos]string {
	exprs := map[token.Pos]string{}

	a.inspect.Nodes(nil, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}

		// Skip named const decl
		decl, ok := n.(*ast.GenDecl)
		if ok && decl.Tok == token.CONST {
			return false
		}

		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}

		switch expr := expr.(type) {
		case *ast.Ident, *ast.SelectorExpr:
			return false
		case *ast.CallExpr:
			if tv := a.pass.TypesInfo.Types[expr]; tv.Value != nil {
				return false
			}

			for _, arg := range expr.Args {
				tv := a.pass.TypesInfo.Types[arg]
				if tv.Value != nil && a.hash(tv.Type) != nil {
					pos := arg.Pos() // pos must be got before expand
					expandedExpr := a.expandNamedConstAll(arg)
					exprs[pos] = exprToString(expandedExpr)
				}
			}
		default:
			tv := a.pass.TypesInfo.Types[expr]
			if tv.Value != nil && a.hash(tv.Type) != nil {
				expandedExpr := a.expandNamedConstAll(expr)
				exprs[expr.Pos()] = exprToString(expandedExpr)
			}
		}
		return false
	})

	return exprs
}

func (a *analyzer) isTarget(t types.Type) bool {
	if t == nil {
		return false
	}

	splited := strings.Split(t.String(), ".")
	tn := splited[0]
	if len(splited) > 1 {
		tn = splited[len(splited) - 1]
	}
	for i := range a.types {
		if a.types[i].Name == tn {
			return true
		}
	}
	return false
}

func (a *analyzer) expandNamedConstAll(expr ast.Expr) ast.Expr {
	r, ok := astutil.Apply(expr, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.Ident:
			obj := a.pass.TypesInfo.ObjectOf(n)
			tv := a.pass.TypesInfo.Types[n]
			if tv.Value != nil {
				v := a.expandNamedConst(tv.Value)
				switch t := obj.Type().(type) {
				case *types.Named:
					h := a.hash(t)
					if h == nil {
						return false
					}
					fun, err := parser.ParseExpr(h.Name)
					if err != nil {
						return false
					}
					cast := &ast.CallExpr{
						Fun:  fun,
						Args: []ast.Expr{v},
					}
					c.Replace(cast)
				default:
					c.Replace(v)
				}
			}
			return false
		case *ast.SelectorExpr:
			obj := a.pass.TypesInfo.ObjectOf(n.Sel)
			switch obj := obj.(type) {
			case *types.Const:
				v := a.expandNamedConst(obj.Val())
				h := a.hash(obj.Type())
				if h == nil {
					return false
				}
				fun, err := parser.ParseExpr(h.Name)
				if err != nil {
					return false
				}
				cast := &ast.CallExpr{
					Fun:  fun,
					Args: []ast.Expr{v},
				}
				c.Replace(cast)
			case *types.TypeName:
				h := a.hash(obj.Type())
				if h == nil {
					return false
				}
				t, err := parser.ParseExpr(h.Name)
				if err != nil {
					return false
				}
				c.Replace(t)
			}
		}
		return true
	}, nil).(ast.Expr)

	if ok {
		return r
	}

	return nil
}

func (a *analyzer) expandNamedConst(cnst constant.Value) ast.Expr {
	switch cnst.Kind() {
	case constant.Bool:
		return &ast.Ident{
			Name: cnst.String(),
		}
	case constant.String:
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: cnst.ExactString(),
		}
	case constant.Int:
		return &ast.BasicLit{
			Kind:  token.INT,
			Value: cnst.ExactString(),
		}
	case constant.Float:
		return &ast.BasicLit{
			Kind:  token.FLOAT,
			Value: cnst.ExactString(),
		}
	case constant.Complex:
		real := constant.Real(cnst)
		imag := constant.Imag(cnst)
		return &ast.BinaryExpr{
			X:  a.expandNamedConst(real),
			Op: token.ADD,
			Y:  a.expandNamedConst(imag),
		}
	}
	return nil
}

func (a *analyzer) parse(fset *token.FileSet, d *tmplData) (*ast.File, error) {
	var src bytes.Buffer
	if err := srcTmpl.Execute(&src, d); err != nil {
		return nil, err
	}

	f, err := parser.ParseFile(fset, "a.go", &src, 0)
	if err != nil {
		return nil, err
	}
	// for debug
	// println(src.String())

	return f, nil
}

func (a *analyzer) typeCheck(fset *token.FileSet, f *ast.File) (*types.Package, *types.Info, error) {

	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
	}
	config := &types.Config{
		Importer: importer.Default(),
	}

	pkg, err := config.Check("a", fset, []*ast.File{f}, info)
	if err != nil {
		return nil, nil, err
	}

	return pkg, info, nil
}
