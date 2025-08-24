package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer - анализатор, запрещающий использование os.Exit в функции main пакета main
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc: `Запрещает использование os.Exit в функции main пакета main.

Этот анализатор проверяет, что в функции main пакета main не используется прямой вызов os.Exit.
Вместо этого рекомендуется использовать panic() или возвращать ошибки для корректного завершения программы.`,
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isOSExitCall(callExpr) {
				if isInMainFunction(pass, callExpr) {
					pass.Reportf(callExpr.Pos(), "прямой вызов os.Exit в функции main запрещен")
				}
			}

			return true
		})
	}

	return nil, nil
}

// isOSExitCall проверяет, является ли вызов os.Exit
func isOSExitCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name == "Exit" {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" {
			return true
		}
	}

	return false
}

// isInMainFunction проверяет, находится ли вызов в функции main
func isInMainFunction(pass *analysis.Pass, node ast.Node) bool {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		if pass.Fset.File(node.Pos()) == pass.Fset.File(file.Pos()) {
			var inMain bool
			ast.Inspect(file, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "main" {
					if node.Pos() >= funcDecl.Pos() && node.End() <= funcDecl.End() {
						inMain = true
						return false
					}
				}
				return true
			})
			if inMain {
				return true
			}
		}
	}

	return false
}
