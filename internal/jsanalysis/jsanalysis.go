package jsanalysis

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"codecompass/internal/critical"
	"codecompass/internal/spellchecker"
	"codecompass/internal/types"
)

// JSAnalysisResult contains all JavaScript analysis results
type JSAnalysisResult struct {
	MissingImports    []MissingImport
	UnusedImports     []UnusedImport
	ComplexityIssues  []ComplexityIssue
	DeadCode          []DeadCode
	SecurityIssues    []SecurityIssue
	CircularDeps      []CircularDependency
	SpellingIssues    []SpellingIssue
	CriticalIssues    []critical.CriticalIssue
}

// Issue types for JavaScript analysis
type MissingImport struct {
	FilePath    string
	Line        int
	Variable    string
	Suggestion  string
	Description string
	Fix         string
	Context     string
}

type UnusedImport struct {
	FilePath    string
	Line        int
	ImportName  string
	ImportPath  string
	Description string
	Fix         string
}

type ComplexityIssue struct {
	FilePath     string
	Line         int
	FunctionName string
	Complexity   int
	Type         string // "cyclomatic", "nesting", "length"
	Description  string
	Suggestions  []string
	Threshold    int
}

type DeadCode struct {
	FilePath    string
	Line        int
	Type        string // "function", "variable", "export"
	Name        string
	Description string
	Suggestions []string
}

type SecurityIssue struct {
	FilePath     string
	Line         int
	Type         string // "console.log", "eval", "innerHTML", etc.
	Description  string
	Severity     string // "low", "medium", "high"
	Impact       string
	Remediation  string
	References   []string
}

type CircularDependency struct {
	Files       []string
	Cycle       []string
	Description string
	Impact      string
}

type SpellingIssue struct {
	FilePath    string
	Line        int
	Column      int
	Word        string
	Context     string
	Type        string   // "comment", "string", "identifier"
	Suggestions []string
	Severity    string
}

// Regular expressions for analysis
var (
	importRegex       = regexp.MustCompile(`^(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|import\s+['"]([^'"]+)['"]|const\s+.*?\s*=\s*require\(['"]([^'"]+)['"]\))`)
	exportRegex       = regexp.MustCompile(`^(?:export\s+(?:default\s+)?(?:function|class|const|let|var)\s+(\w+)|export\s*\{([^}]+)\})`)
	functionRegex     = regexp.MustCompile(`(?:function\s+(\w+)|(\w+)\s*[:=]\s*(?:function|\([^)]*\)\s*=>)|class\s+(\w+))`)
	variableRegex     = regexp.MustCompile(`(?:^|\s)(?:const|let|var)\s+(\w+)`)
	consoleLogRegex   = regexp.MustCompile(`console\.(?:log|warn|error|debug|info)`)
	evalRegex         = regexp.MustCompile(`\beval\s*\(`)
	innerHTMLRegex    = regexp.MustCompile(`\.innerHTML\s*=`)
	undefinedVarRegex = regexp.MustCompile(`\b([A-Z][a-zA-Z0-9]*)\b`) // Common pattern for missing imports
)

// RunJSAnalysis performs comprehensive JavaScript static analysis
func RunJSAnalysis(trackedFiles map[string]bool, enableSpellCheck bool) (*JSAnalysisResult, error) {
	result := &JSAnalysisResult{
		MissingImports:   []MissingImport{},
		UnusedImports:    []UnusedImport{},
		ComplexityIssues: []ComplexityIssue{},
		DeadCode:         []DeadCode{},
		SecurityIssues:   []SecurityIssue{},
		CircularDeps:     []CircularDependency{},
		SpellingIssues:   []SpellingIssue{},
		CriticalIssues:   []critical.CriticalIssue{},
	}

	// Initialize spell checker only if enabled
	var spellChecker *spellchecker.SpellChecker
	if enableSpellCheck {
		var err error
		spellChecker, err = spellchecker.NewSpellChecker()
		if err != nil {
			fmt.Printf("Warning: Failed to initialize spell checker: %v\n", err)
			spellChecker = nil
		} else {
			// Try to load external dictionary
			spellChecker.LoadExternalDictionary("")
		}
	}

	// Process each JavaScript/TypeScript file
	for filePath := range trackedFiles {
		if !isJavaScriptFile(filePath) {
			continue
		}

		if err := analyzeFile(filePath, result, spellChecker); err != nil {
			// Continue with other files if one fails
			fmt.Printf("Warning: Failed to analyze %s: %v\n", filePath, err)
		}

		// Run critical analysis for server-breaking issues
		if content, err := os.ReadFile(filePath); err == nil {
			criticalAnalyzer := critical.NewCriticalAnalyzer()
			criticalIssues := criticalAnalyzer.AnalyzeCriticalIssues(filePath, string(content), []types.Issue{})
			result.CriticalIssues = append(result.CriticalIssues, criticalIssues...)
		}
	}

	return result, nil
}

// isJavaScriptFile checks if the file is a JavaScript/TypeScript file
func isJavaScriptFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	if !(ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" || ext == ".mjs" || ext == ".cjs") {
		return false
	}
	
	// Skip minified files (common patterns)
	baseName := filepath.Base(filePath)
	if strings.Contains(baseName, ".min.") || 
	   strings.Contains(baseName, ".bundle.") ||
	   strings.Contains(filePath, "node_modules") ||
	   strings.Contains(filePath, "vendor") ||
	   strings.Contains(filePath, "dist/") ||
	   strings.Contains(filePath, "build/") {
		return false
	}
	
	// Skip very large files (likely minified or generated)
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() > 1024*1024 { // Skip files larger than 1MB
			return false
		}
	}
	
	return true
}

// analyzeFile performs analysis on a single file
func analyzeFile(filePath string, result *JSAnalysisResult, spellChecker *spellchecker.SpellChecker) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max token size
	lineNum := 0
	
	var imports []ImportInfo
	var exports []string
	var functions []string
	var variables []string
	var usedIdentifiers []string
	
	functionStack := 0
	maxNesting := 0
	currentFunctionStart := 0
	currentFunctionName := ""

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "*") {
			continue
		}
		
		// Skip very long lines (likely minified code)
		if len(line) > 500 {
			continue
		}

		// Analyze imports
		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			importPath := ""
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					importPath = matches[i]
					break
				}
			}
			imports = append(imports, ImportInfo{
				Line: lineNum,
				Path: importPath,
				Raw:  line,
			})
		}

		// Analyze exports
		if matches := exportRegex.FindStringSubmatch(line); matches != nil {
			if matches[1] != "" {
				exports = append(exports, matches[1])
			} else if matches[2] != "" {
				// Handle export { name1, name2 }
				exportNames := strings.Split(matches[2], ",")
				for _, name := range exportNames {
					exports = append(exports, strings.TrimSpace(name))
				}
			}
		}

		// Analyze functions
		if matches := functionRegex.FindStringSubmatch(line); matches != nil {
			funcName := ""
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					funcName = matches[i]
					break
				}
			}
			if funcName != "" {
				functions = append(functions, funcName)
				currentFunctionStart = lineNum
				currentFunctionName = funcName
			}
		}

		// Analyze variables
		if matches := variableRegex.FindStringSubmatch(line); matches != nil {
			variables = append(variables, matches[1])
		}

		// Track nesting level for complexity
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		functionStack += openBraces - closeBraces
		if functionStack > maxNesting {
			maxNesting = functionStack
		}

		// Check for function end and complexity
		if strings.Contains(line, "}") && currentFunctionName != "" && functionStack <= 0 {
			funcLength := lineNum - currentFunctionStart
			if funcLength > 50 { // Configurable threshold
				result.ComplexityIssues = append(result.ComplexityIssues, ComplexityIssue{
					FilePath:     filePath,
					Line:         currentFunctionStart,
					FunctionName: currentFunctionName,
					Complexity:   funcLength,
					Type:         "length",
					Description:  fmt.Sprintf("Function '%s' is too long (%d lines)", currentFunctionName, funcLength),
					Suggestions:  []string{"Break into smaller functions", "Extract reusable logic", "Consider using early returns"},
					Threshold:    50,
				})
			}
			if maxNesting > 4 { // Configurable threshold
				result.ComplexityIssues = append(result.ComplexityIssues, ComplexityIssue{
					FilePath:     filePath,
					Line:         currentFunctionStart,
					FunctionName: currentFunctionName,
					Complexity:   maxNesting,
					Type:         "nesting",
					Description:  fmt.Sprintf("Function '%s' has excessive nesting depth (%d levels)", currentFunctionName, maxNesting),
					Suggestions:  []string{"Use early returns", "Extract nested logic into functions", "Consider guard clauses", "Flatten conditional structures"},
					Threshold:    4,
				})
			}
			currentFunctionName = ""
			maxNesting = 0
		}

		// Security analysis
		if consoleLogRegex.MatchString(line) {
			result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
				FilePath:    filePath,
				Line:        lineNum,
				Type:        "console.log",
				Description: "Console statement found - should be removed in production",
				Severity:    "low",
				Impact:      "Performance impact and information disclosure in production",
				Remediation: "Remove console statements or use a logging library with proper levels",
				References:  []string{"https://eslint.org/docs/rules/no-console"},
			})
		}

		if evalRegex.MatchString(line) {
			result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
				FilePath:    filePath,
				Line:        lineNum,
				Type:        "eval",
				Description: "Use of eval() is dangerous and should be avoided",
				Severity:    "high",
				Impact:      "Code injection vulnerability, performance issues",
				Remediation: "Use JSON.parse() for JSON, or rewrite logic to avoid eval()",
				References:  []string{"https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/eval#never_use_eval!"},
			})
		}

		if innerHTMLRegex.MatchString(line) {
			result.SecurityIssues = append(result.SecurityIssues, SecurityIssue{
				FilePath:    filePath,
				Line:        lineNum,
				Type:        "innerHTML",
				Description: "innerHTML assignment can lead to XSS vulnerabilities",
				Severity:    "medium",
				Impact:      "Cross-site scripting (XSS) vulnerability if content is user-controlled",
				Remediation: "Use textContent, createElement, or a safe DOM library",
				References:  []string{"https://owasp.org/www-community/attacks/xss/", "https://developer.mozilla.org/en-US/docs/Web/API/Element/innerHTML#security_considerations"},
			})
		}

		// Collect used identifiers for missing import detection
		words := strings.Fields(line)
		for _, word := range words {
			// Remove common JavaScript operators and keywords
			cleanWord := strings.Trim(word, "()[]{}.,;:")
			if isIdentifier(cleanWord) && !isJavaScriptKeyword(cleanWord) {
				usedIdentifiers = append(usedIdentifiers, cleanWord)
			}
		}
	}

	// Check for unused imports
	for _, imp := range imports {
		if !isImportUsed(imp, usedIdentifiers) {
			importName := extractImportName(imp.Raw)
			result.UnusedImports = append(result.UnusedImports, UnusedImport{
				FilePath:    filePath,
				Line:        imp.Line,
				ImportName:  importName,
				ImportPath:  imp.Path,
				Description: fmt.Sprintf("Import '%s' from '%s' is unused", importName, imp.Path),
				Fix:         fmt.Sprintf("Remove the unused import: %s", imp.Raw),
			})
		}
	}

	// Check for potential missing imports
	checkMissingImports(filePath, usedIdentifiers, imports, result)

	// Spell checking
	if spellChecker != nil {
		spellIssues, err := spellChecker.CheckFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Spell check failed for %s: %v\n", filePath, err)
		} else {
			// Convert spell checker issues to our format
			for _, issue := range spellIssues {
				result.SpellingIssues = append(result.SpellingIssues, SpellingIssue{
					FilePath:    filePath,
					Line:        issue.Line,
					Column:      issue.Column,
					Word:        issue.Word,
					Context:     issue.Context,
					Type:        issue.Type,
					Suggestions: issue.Suggestions,
					Severity:    issue.Severity,
				})
			}
		}
	}

	return scanner.Err()
}

type ImportInfo struct {
	Line int
	Path string
	Raw  string
}

// Helper functions
func isIdentifier(word string) bool {
	if len(word) == 0 {
		return false
	}
	// Check if it starts with uppercase (common pattern for imported classes/components)
	return regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`).MatchString(word)
}

func isJavaScriptKeyword(word string) bool {
	keywords := map[string]bool{
		"function": true, "var": true, "let": true, "const": true,
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "default": true, "break": true, "continue": true,
		"return": true, "try": true, "catch": true, "finally": true, "throw": true,
		"new": true, "this": true, "super": true, "class": true, "extends": true,
		"import": true, "export": true, "from": true, "as": true,
		"async": true, "await": true, "yield": true, "typeof": true, "instanceof": true,
		"true": true, "false": true, "null": true, "undefined": true,
	}
	return keywords[word]
}

func isImportUsed(imp ImportInfo, usedIdentifiers []string) bool {
	importName := extractImportName(imp.Raw)
	for _, identifier := range usedIdentifiers {
		if identifier == importName {
			return true
		}
	}
	return false
}

func extractImportName(importLine string) string {
	// Simple extraction - could be improved
	if strings.Contains(importLine, "import") && strings.Contains(importLine, "from") {
		parts := strings.Split(importLine, "from")
		if len(parts) > 0 {
			importPart := strings.TrimSpace(parts[0])
			importPart = strings.TrimPrefix(importPart, "import")
			importPart = strings.TrimSpace(importPart)
			importPart = strings.Trim(importPart, "{}")
			return strings.TrimSpace(importPart)
		}
	}
	return ""
}

func checkMissingImports(filePath string, usedIdentifiers []string, imports []ImportInfo, result *JSAnalysisResult) {
	// Create map of imported identifiers
	importedIds := make(map[string]bool)
	for _, imp := range imports {
		importedIds[extractImportName(imp.Raw)] = true
	}

	// Check for common missing imports
	commonLibraries := map[string]string{
		"React":     "react",
		"useState":  "react",
		"useEffect": "react",
		"useContext": "react",
		"useReducer": "react",
		"Component": "react",
		"PropTypes": "prop-types",
		"axios":     "axios",
		"lodash":    "lodash",
		"moment":    "moment",
		"dayjs":     "dayjs",
		"_":         "lodash",
		"jQuery":    "jquery",
		"$":         "jquery",
		"Express":   "express",
		"Mongoose":  "mongoose",
	}

	identifierCount := make(map[string]int)
	for _, id := range usedIdentifiers {
		identifierCount[id]++
	}

	for id, count := range identifierCount {
		if count > 1 && !importedIds[id] { // Used multiple times but not imported
			if suggestion, exists := commonLibraries[id]; exists {
				result.MissingImports = append(result.MissingImports, MissingImport{
					FilePath:    filePath,
					Line:        1, // Top of file
					Variable:    id,
					Suggestion:  suggestion,
					Description: fmt.Sprintf("'%s' is used %d times but not imported", id, count),
					Fix:         fmt.Sprintf("Add import: import %s from '%s';", id, suggestion),
					Context:     fmt.Sprintf("Used %d times in this file", count),
				})
			}
		}
	}
}

// ConvertToIssues converts JSAnalysisResult to standard Issue format for integration
func (result *JSAnalysisResult) ConvertToIssues() []types.Issue {
	var issues []types.Issue

	// Convert missing imports
	for _, mi := range result.MissingImports {
		issues = append(issues, types.Issue{
			FilePath: mi.FilePath,
			Line:     mi.Line,
			RuleID:   "missing-import",
			Message:  fmt.Sprintf("Missing import: %s (suggested: %s)", mi.Variable, mi.Suggestion),
			Severity: 1, // Warning
		})
	}

	// Convert unused imports
	for _, ui := range result.UnusedImports {
		issues = append(issues, types.Issue{
			FilePath: ui.FilePath,
			Line:     ui.Line,
			RuleID:   "unused-import",
			Message:  fmt.Sprintf("Unused import: %s from %s", ui.ImportName, ui.ImportPath),
			Severity: 1, // Warning
		})
	}

	// Convert complexity issues
	for _, ci := range result.ComplexityIssues {
		severity := 1
		if ci.Complexity > 100 || ci.Type == "nesting" && ci.Complexity > 6 {
			severity = 2 // Error
		}
		issues = append(issues, types.Issue{
			FilePath: ci.FilePath,
			Line:     ci.Line,
			RuleID:   fmt.Sprintf("complexity-%s", ci.Type),
			Message:  fmt.Sprintf("High %s complexity in %s: %d", ci.Type, ci.FunctionName, ci.Complexity),
			Severity: severity,
		})
	}

	// Convert security issues
	for _, si := range result.SecurityIssues {
		severity := 1
		switch si.Severity {
		case "high":
			severity = 2
		case "medium":
			severity = 2
		case "low":
			severity = 1
		}
		issues = append(issues, types.Issue{
			FilePath: si.FilePath,
			Line:     si.Line,
			RuleID:   fmt.Sprintf("security-%s", si.Type),
			Message:  si.Description,
			Severity: severity,
		})
	}

	// Convert spelling issues
	for _, si := range result.SpellingIssues {
		severity := 1 // Generally low severity for spelling
		if si.Severity == "medium" {
			severity = 1
		}
		
		message := fmt.Sprintf("Spelling: '%s' in %s", si.Word, si.Context)
		if len(si.Suggestions) > 0 {
			message += fmt.Sprintf(" (suggestions: %s)", strings.Join(si.Suggestions, ", "))
		}
		
		issues = append(issues, types.Issue{
			FilePath: si.FilePath,
			Line:     si.Line,
			RuleID:   fmt.Sprintf("spelling-%s", si.Type),
			Message:  message,
			Severity: severity,
		})
	}

	// Convert critical issues (highest priority)
	for _, ci := range result.CriticalIssues {
		severity := 2 // Error by default
		if ci.Severity == "MEDIUM" {
			severity = 1 // Warning
		}
		
		message := fmt.Sprintf("[%s] %s - %s", ci.Severity, ci.Message, ci.Impact)
		
		issues = append(issues, types.Issue{
			FilePath: ci.FilePath,
			Line:     ci.Line,
			RuleID:   fmt.Sprintf("critical-%s", strings.ToLower(ci.Type)),
			Message:  message,
			Severity: severity,
		})
	}

	return issues
}

// AnalyzeSingleFile performs detailed analysis on a specific file
func AnalyzeSingleFile(filePath string, detailed bool, enableSpellCheck bool) (*JSAnalysisResult, error) {
	if !isJavaScriptFile(filePath) {
		return nil, fmt.Errorf("file %s is not a JavaScript/TypeScript file", filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file %s does not exist: %v", filePath, err)
	}

	result := &JSAnalysisResult{
		MissingImports:   []MissingImport{},
		UnusedImports:    []UnusedImport{},
		ComplexityIssues: []ComplexityIssue{},
		DeadCode:         []DeadCode{},
		SecurityIssues:   []SecurityIssue{},
		CircularDeps:     []CircularDependency{},
		SpellingIssues:   []SpellingIssue{},
		CriticalIssues:   []critical.CriticalIssue{},
	}

	// Initialize spell checker only if enabled
	var spellChecker *spellchecker.SpellChecker
	if enableSpellCheck {
		var err error
		spellChecker, err = spellchecker.NewSpellChecker()
		if err != nil {
			fmt.Printf("Warning: Failed to initialize spell checker: %v\n", err)
			spellChecker = nil
		} else {
			spellChecker.LoadExternalDictionary("")
		}
	}

	// Analyze the file
	if err := analyzeFile(filePath, result, spellChecker); err != nil {
		return nil, fmt.Errorf("failed to analyze file %s: %v", filePath, err)
	}

	// Run critical analysis for server-breaking issues
	if content, err := os.ReadFile(filePath); err == nil {
		criticalAnalyzer := critical.NewCriticalAnalyzer()
		criticalIssues := criticalAnalyzer.AnalyzeCriticalIssues(filePath, string(content), []types.Issue{})
		result.CriticalIssues = append(result.CriticalIssues, criticalIssues...)
	}

	return result, nil
}

// PrintDetailedResults prints detailed analysis results for a file
func (result *JSAnalysisResult) PrintDetailedResults(filePath string) {
	fmt.Printf("\n🔍 Detailed Analysis for: %s\n", filePath)
	fmt.Println("=" + strings.Repeat("=", len(filePath)+25))

	// Missing Imports
	if len(result.MissingImports) > 0 {
		fmt.Printf("\n📦 Missing Imports (%d):\n", len(result.MissingImports))
		for i, mi := range result.MissingImports {
			fmt.Printf("  %d. Line %d: %s\n", i+1, mi.Line, mi.Description)
			fmt.Printf("     💡 Fix: %s\n", mi.Fix)
			if mi.Context != "" {
				fmt.Printf("     📝 Context: %s\n", mi.Context)
			}
			fmt.Println()
		}
	}

	// Unused Imports
	if len(result.UnusedImports) > 0 {
		fmt.Printf("\n🗑️  Unused Imports (%d):\n", len(result.UnusedImports))
		for i, ui := range result.UnusedImports {
			fmt.Printf("  %d. Line %d: %s\n", i+1, ui.Line, ui.Description)
			fmt.Printf("     💡 Fix: %s\n", ui.Fix)
			fmt.Println()
		}
	}

	// Complexity Issues
	if len(result.ComplexityIssues) > 0 {
		fmt.Printf("\n🧮 Complexity Issues (%d):\n", len(result.ComplexityIssues))
		for i, ci := range result.ComplexityIssues {
			fmt.Printf("  %d. Line %d: %s\n", i+1, ci.Line, ci.Description)
			fmt.Printf("     📊 %s: %d (threshold: %d)\n", strings.Title(ci.Type), ci.Complexity, ci.Threshold)
			if len(ci.Suggestions) > 0 {
				fmt.Printf("     💡 Suggestions:\n")
				for _, suggestion := range ci.Suggestions {
					fmt.Printf("        • %s\n", suggestion)
				}
			}
			fmt.Println()
		}
	}

	// Security Issues
	if len(result.SecurityIssues) > 0 {
		fmt.Printf("\n🔒 Security Issues (%d):\n", len(result.SecurityIssues))
		for i, si := range result.SecurityIssues {
			fmt.Printf("  %d. Line %d: %s [%s]\n", i+1, si.Line, si.Description, strings.ToUpper(si.Severity))
			if si.Impact != "" {
				fmt.Printf("     ⚠️  Impact: %s\n", si.Impact)
			}
			if si.Remediation != "" {
				fmt.Printf("     💡 Remediation: %s\n", si.Remediation)
			}
			if len(si.References) > 0 {
				fmt.Printf("     🔗 References:\n")
				for _, ref := range si.References {
					fmt.Printf("        • %s\n", ref)
				}
			}
			fmt.Println()
		}
	}

	// Spelling Issues
	if len(result.SpellingIssues) > 0 {
		fmt.Printf("\n📝 Spelling Issues (%d):\n", len(result.SpellingIssues))
		for i, si := range result.SpellingIssues {
			fmt.Printf("  %d. Line %d: '%s' in %s\n", i+1, si.Line, si.Word, si.Context)
			if len(si.Suggestions) > 0 {
				fmt.Printf("     💡 Suggestions: %s\n", strings.Join(si.Suggestions, ", "))
			}
			fmt.Println()
		}
	}

	// Critical Issues (show first as highest priority)
	if len(result.CriticalIssues) > 0 {
		fmt.Printf("\n🚨 CRITICAL SERVER-BREAKING ISSUES (%d):\n", len(result.CriticalIssues))
		for i, ci := range result.CriticalIssues {
			severityIcon := "🔴"
			if ci.Severity == "HIGH" {
				severityIcon = "🟠"
			} else if ci.Severity == "MEDIUM" {
				severityIcon = "🟡"
			}
			
			fmt.Printf("  %d. Line %d: %s %s [%s]\n", i+1, ci.Line, severityIcon, ci.Message, ci.Severity)
			fmt.Printf("     💥 Impact: %s\n", ci.Impact)
			fmt.Printf("     🔧 Fix: %s\n", ci.FixSuggestion)
			fmt.Println()
		}
	}

	// Summary
	total := len(result.MissingImports) + len(result.UnusedImports) + 
			len(result.ComplexityIssues) + len(result.SecurityIssues) + 
			len(result.SpellingIssues) + len(result.CriticalIssues)
	
	fmt.Printf("\n📊 Summary: %d total issues found\n", total)
	
	if len(result.CriticalIssues) > 0 {
		fmt.Printf("🚨 %d CRITICAL issues that can break the server\n", len(result.CriticalIssues))
		fmt.Println("⚠️  RECOMMENDATION: Fix critical issues before deploying!")
	} else if total == 0 {
		fmt.Println("🎉 No issues found! Great job!")
	}
}