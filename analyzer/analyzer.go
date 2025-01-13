package analyzer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `checks that the error style is the same in all app

This analyzer check for correctness of error operation and that error or message fields are set.
`

var ErrStyleAnalyzer = &analysis.Analyzer{
	Name:     "errstyle",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	opName  *string
	errType *string

	errorType = types.Universe.Lookup("error").Type()
)

const (
	opField  = "Op"
	errField = "Err"
	msgField = "Message"
)

func init() {
	opName = ErrStyleAnalyzer.Flags.String("op_name", "op", "name of op const")
	errType = ErrStyleAnalyzer.Flags.String("errType", "app.Error", "error type")
}

type fn struct {
	decl       *ast.FuncDecl
	errReturns []ast.Expr
	constDecls map[string]constDecl
	errVars    []*ast.CompositeLit
}

type constDecl struct {
	ident   *ast.Ident
	valNode ast.Node
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	fnList := make([]fn, 0)
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		function := n.(*ast.FuncDecl)

		// TODO: not need to check functions that returns nothing
		if function.Type.Results == nil {
			return
		}

		for i, v := range function.Type.Results.List {
			t := pass.TypesInfo.TypeOf(v.Type)
			if t == errorType {
				// not last in returns
				if i != len(function.Type.Results.List)-1 {
					pass.Report(analysis.Diagnostic{
						Pos:     v.Pos(),
						Message: "error must be last",
					})
					continue
				}

				fnList = append(fnList, fn{
					decl:       function,
					errReturns: make([]ast.Expr, 0),
					constDecls: make(map[string]constDecl, 0),
					errVars:    make([]*ast.CompositeLit, 0),
				})
			}
		}
	})

	populateFuncs(pass, fnList)
	checkConst(pass, fnList)
	checkVars(pass, fnList)
	return nil, nil
}

func populateFuncs(pass *analysis.Pass, fnList []fn) {
	for i, f := range fnList {
		ast.Inspect(f.decl.Body, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.GenDecl:
				if v.Tok != token.CONST {
					break
				}
				for _, cDecl := range v.Specs {
					if vSpec, ok := cDecl.(*ast.ValueSpec); ok {
						for ii, vv := range vSpec.Names {
							fnList[i].constDecls[vv.Name] = constDecl{ident: vv, valNode: vSpec.Values[ii]}
						}
					}
				}
			case *ast.CompositeLit:
				if v.Type == nil {
					break
				}

				if pass.TypesInfo.TypeOf(v.Type).String() == *errType {
					fnList[i].errVars = append(fnList[i].errVars, v)
				}
			}

			return true
		})
	}
}

func checkVars(pass *analysis.Pass, fnList []fn) {
	for _, fn := range fnList {
		_, opConstExists := fn.constDecls[*opName]
		for _, v := range fn.errVars {
			kv := map[string]ast.Expr{}
			for _, e := range v.Elts {
				if f, ok := e.(*ast.KeyValueExpr); ok {
					kv[f.Key.(*ast.Ident).Name] = f.Value
				}
			}

			opValue, opExists := kv[opField]
			switch opExists {
			case true:
				constName, isConst := isStringConst(pass, opValue)
				if !isConst {
					report := analysis.Diagnostic{
						Pos:     opValue.Pos(),
						Message: "const value must be used",
					}
					if opConstExists {
						report.SuggestedFixes = []analysis.SuggestedFix{
							{
								TextEdits: []analysis.TextEdit{
									{
										Pos:     opValue.Pos(),
										End:     opValue.End(),
										NewText: []byte(*opName),
									},
								},
							},
						}
					}
					pass.Report(report)
				}

				if isConst && opConstExists && constName != *opName {
					pass.Report(analysis.Diagnostic{
						Pos:     opValue.Pos(),
						Message: fmt.Sprintf("must use `%s` const", *opName),
						SuggestedFixes: []analysis.SuggestedFix{
							{
								TextEdits: []analysis.TextEdit{
									{
										Pos:     opValue.Pos(),
										End:     opValue.End(),
										NewText: []byte(*opName),
									},
								},
							},
						},
					})
				}
			case false:
				pass.Report(analysis.Diagnostic{
					Pos:     v.Pos(),
					Message: "error without operation",
				})
			}
			_, errExists := kv[errField]
			_, msgExists := kv[msgField]
			if !errExists && !msgExists {
				pass.Report(analysis.Diagnostic{Pos: v.Pos(), Message: "error don't have error or message"})
			}
		}
	}
}

func checkConst(pass *analysis.Pass, fnList []fn) {
	for _, fn := range fnList {
		validOp := nameOf(fn.decl)
		for k, v := range fn.constDecls {
			_, ok := isStringConst(pass, v.ident)
			if k == *opName && !ok {
				pass.Report(analysis.Diagnostic{Pos: v.ident.Pos(), Message: "operation constant must be string"})
			}

			constVal := stringConst(pass, v.ident)
			if constVal == validOp && k != *opName {
				pass.Report(analysis.Diagnostic{
					Pos:     v.ident.Pos(),
					Message: fmt.Sprintf("operation constant must be named `%s` not `%s`", *opName, k),
				})
			}

			if k == *opName && constVal != validOp {
				pass.Report(analysis.Diagnostic{
					Pos:     v.valNode.Pos(),
					Message: fmt.Sprintf("operation must be `%s` not `%s`", validOp, constVal),
					SuggestedFixes: []analysis.SuggestedFix{
						{
							TextEdits: []analysis.TextEdit{
								{
									Pos:     v.valNode.Pos() + 1,
									End:     v.valNode.End() - 1,
									NewText: []byte(validOp),
								},
							},
						},
					},
				})
			}
		}
	}
}

func nameOf(f *ast.FuncDecl) string {
	if r := f.Recv; r != nil && len(r.List) == 1 {
		t := r.List[0].Type
		if p, _ := t.(*ast.StarExpr); p != nil {
			t = p.X
		}
		if p, _ := t.(*ast.Ident); p != nil {
			return p.Name + "." + f.Name.Name
		}
	}
	return f.Name.Name
}

func isStringConst(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	c, ok := obj.(*types.Const)
	if !ok {
		return "", false
	}
	basic, ok := c.Type().(*types.Basic)
	if !ok {
		return "", false
	}
	if basic.Kind() != types.UntypedString && basic.Kind() != types.String {
		return "", false
	}
	return ident.Name, true
}

func stringConst(pass *analysis.Pass, expr ast.Expr) string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	val := pass.TypesInfo.ObjectOf(expr.(*ast.Ident)).(*types.Const).Val()
	return constant.StringVal(val)
}
