package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type Config struct {
	DisableLowercaseCheck bool `json:"disable_lowercase_check"`
	DisableEnglishCheck bool `json:"disable_english_check"`
	DisableSpecialCharCheck bool `json:"disable_special_char_check"`
	DisableSensitiveCheck bool `json:"disable_sensitive_check"`
	CustomSensitiveKeywords []string `json:"custom_sensitive_keywords"`
}

var Analyzer = &analysis.Analyzer{
	Name:     "logchecklint",
	Doc:      "checks log messages for common issues: uppercase start, non-English text, special characters, sensitive data",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var logFunctions = map[string]map[string]bool{
	"log/slog": {
		"Info":  true,
		"Error": true,
		"Warn":  true,
		"Debug": true,
		"Log":   true,
	},
	"slog": {
		"Info":  true,
		"Error": true,
		"Warn":  true,
		"Debug": true,
		"Log":   true,
	},
}

var zapMethods = map[string]bool{
	"Info":   true,
	"Error":  true,
	"Warn":   true,
	"Debug":  true,
	"Fatal":  true,
	"Panic":  true,
	"DPanic": true,
	"Infof":   true,
	"Errorf":  true,
	"Warnf":   true,
	"Debugf":  true,
	"Fatalf":  true,
	"Panicf":  true,
	"DPanicf": true,
	"Infow":   true,
	"Errorw":  true,
	"Warnw":   true,
	"Debugw":  true,
	"Fatalw":  true,
	"Panicw":  true,
	"DPanicw": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		msg, pos := extractLogMessage(call)
		if msg == "" {
			return
		}

		checkAndReport(pass, msg, pos)
	})

	return nil, nil
}

func RunWithConfig(cfg Config) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}

		insp.Preorder(nodeFilter, func(n ast.Node) {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return
			}

			msg, pos := extractLogMessage(call)
			if msg == "" {
				return
			}

			checkAndReportWithConfig(pass, msg, pos, cfg)
		})

		return nil, nil
	}
}

func extractLogMessage(call *ast.CallExpr) (string, token.Pos) {
	if len(call.Args) == 0 {
		return "", token.NoPos
	}

	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		funcName := fn.Sel.Name

		if ident, ok := fn.X.(*ast.Ident); ok {
			if ident.Name == "slog" {
				if _, exists := logFunctions["slog"][funcName]; exists {
					return extractStringArg(call, funcName)
				}
			}
			if zapMethods[funcName] {
				return extractStringArg(call, funcName)
			}
		}

		if _, ok := fn.X.(*ast.CallExpr); ok {
			if zapMethods[funcName] {
				return extractStringArg(call, funcName)
			}
		}

		if ident, ok := fn.X.(*ast.Ident); ok {
			if ident.Name == "log" {
				switch funcName {
				case "Info", "Error", "Warn", "Debug", "Fatal", "Panic",
					"Infof", "Errorf", "Warnf", "Debugf", "Fatalf", "Panicf":
					return extractStringArg(call, funcName)
				}
			}
		}
	}

	return "", token.NoPos
}

func extractStringArg(call *ast.CallExpr, funcName string) (string, token.Pos) {
	if funcName == "Log" && len(call.Args) >= 3 {
		if lit, ok := call.Args[2].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return strings.Trim(lit.Value, "`\"" ), lit.Pos()
		}
		return "", token.NoPos
	}

	if len(call.Args) >= 1 {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return strings.Trim(lit.Value, "`\""), lit.Pos()
		}
	}

	return "", token.NoPos
}

func checkAndReport(pass *analysis.Pass, msg string, pos token.Pos) {
	if CheckLowercaseStart(msg) {
		lower := strings.ToLower(string(msg[0])) + msg[1:]
		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Message:  "log message should start with a lowercase letter",
			Category: "logchecklint",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "convert first letter to lowercase",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     pos,
							End:     pos + token.Pos(len(msg)+2),
							NewText: []byte(`"` + lower + `"`),
						},
					},
				},
			},
		})
	}

	if CheckEnglishOnly(msg) {
		pass.Reportf(pos, "log message should be in English only")
	}

	if CheckSpecialCharsOrEmoji(msg) {
		cleaned := cleanSpecialChars(msg)
		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Message:  "log message should not contain special characters or emoji",
			Category: "logchecklint",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "remove special characters and emoji",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     pos,
							End:     pos + token.Pos(len(msg)+2),
							NewText: []byte(`"` + cleaned + `"`),
						},
					},
				},
			},
		})
	}

	if CheckSensitiveData(msg) {
		pass.Reportf(pos, "log message may contain sensitive data")
	}
}

func checkAndReportWithConfig(pass *analysis.Pass, msg string, pos token.Pos, cfg Config) {
	if !cfg.DisableLowercaseCheck && CheckLowercaseStart(msg) {
		pass.Reportf(pos, "log message should start with a lowercase letter")
	}

	if !cfg.DisableEnglishCheck && CheckEnglishOnly(msg) {
		pass.Reportf(pos, "log message should be in English only")
	}

	if !cfg.DisableSpecialCharCheck && CheckSpecialCharsOrEmoji(msg) {
		pass.Reportf(pos, "log message should not contain special characters or emoji")
	}

	if !cfg.DisableSensitiveCheck {
		if len(cfg.CustomSensitiveKeywords) > 0 {
			if CheckSensitiveDataWithCustomKeywords(msg, cfg.CustomSensitiveKeywords) {
				pass.Reportf(pos, "log message may contain sensitive data")
			}
		} else {
			if CheckSensitiveData(msg) {
				pass.Reportf(pos, "log message may contain sensitive data")
			}
		}
	}
}

func cleanSpecialChars(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isEmoji(r) {
			continue
		}
		if strings.ContainsRune(specialChars, r) {
			continue
		}
		if !unicode.IsPrint(r) && r != ' ' {
			continue
		}
		b.WriteRune(r)
	}
	result := b.String()
	// Clean up repeated dots
	for strings.Contains(result, "...") {
		result = strings.ReplaceAll(result, "...", ".")
	}
	return strings.TrimSpace(result)
}
