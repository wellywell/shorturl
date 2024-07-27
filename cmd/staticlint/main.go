package main

import (
	"go/ast"

	"github.com/Antonboom/testifylint/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var NoOsExitAnalizer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for os.Exit() in main() function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if f, ok := node.(*ast.FuncDecl); ok {
				if f.Name.Name == "main" {
					ast.Inspect(f, func(node ast.Node) bool {
						if c, ok := node.(*ast.CallExpr); ok {
							if s, ok := c.Fun.(*ast.SelectorExpr); ok {
								if s.Sel.Name == "Exit" {
									if pkg, ok := s.X.(*ast.Ident); ok {
										if pkg.Name == "os" {
											pass.Reportf(s.Sel.NamePos, "os.Exit in main")
										}
									}
								}
							}
						}
						return true
					})

				}
			}
			return true
		})
	}
	return nil, nil
}

func main() {

	mychecks :=
		[]*analysis.Analyzer{
			printf.Analyzer,
			shadow.Analyzer,
			appends.Analyzer,
			assign.Analyzer,
			atomic.Analyzer,
			atomicalign.Analyzer,
			bools.Analyzer,
			composite.Analyzer,
			copylock.Analyzer,
			deepequalerrors.Analyzer,
			defers.Analyzer,
			directive.Analyzer,
			errorsas.Analyzer,
			fieldalignment.Analyzer,
			httpmux.Analyzer,
			httpresponse.Analyzer,
			ifaceassert.Analyzer,
			loopclosure.Analyzer,
			lostcancel.Analyzer,
			nilfunc.Analyzer,
			reflectvaluecompare.Analyzer,
			shift.Analyzer,
			stdmethods.Analyzer,
			stdversion.Analyzer,
			stringintconv.Analyzer,
			structtag.Analyzer,
			testinggoroutine.Analyzer,
			tests.Analyzer,
			timeformat.Analyzer,
			unmarshal.Analyzer,
			unreachable.Analyzer,
			unusedresult.Analyzer,
			unusedwrite.Analyzer,
			usesgenerics.Analyzer,
			errcheck.Analyzer,
			analyzer.New(),
			NoOsExitAnalizer,
		}
	for _, v := range staticcheck.Analyzers {
		// static check
		{
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		// анализ стиля
		{
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
