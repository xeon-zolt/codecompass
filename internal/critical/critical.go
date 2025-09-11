package critical

import (
	"fmt"
	"regexp"
	"strings"

	"codecompass/internal/types"
)

// CriticalIssue represents an issue that can break the server
type CriticalIssue struct {
	FilePath     string
	Line         int
	Type         string
	Severity     string // "CRITICAL", "HIGH", "MEDIUM"
	Message      string
	Impact       string
	FixSuggestion string
	Category     string // "RUNTIME_ERROR", "IMPORT_ERROR", "TYPE_ERROR", "SYNTAX_ERROR"
}

// CriticalAnalyzer detects server-breaking issues
type CriticalAnalyzer struct {
	serverFrameworks []string
	criticalPatterns []CriticalPattern
}

type CriticalPattern struct {
	Pattern     *regexp.Regexp
	Type        string
	Severity    string
	Message     string
	Impact      string
	Category    string
	FixTemplate string
}

// NewCriticalAnalyzer creates a new critical issue analyzer
func NewCriticalAnalyzer() *CriticalAnalyzer {
	ca := &CriticalAnalyzer{
		serverFrameworks: []string{
			"express", "koa", "fastify", "hapi", "nest", "meteor",
			"next", "nuxt", "sveltekit", "remix", "gatsby",
			"apollo", "graphql", "mongoose", "prisma", "sequelize",
		},
	}
	
	ca.initializeCriticalPatterns()
	return ca
}

func (ca *CriticalAnalyzer) initializeCriticalPatterns() {
	ca.criticalPatterns = []CriticalPattern{
		// Only detect actual undefined variables/functions, not just keywords
		// This pattern is now disabled to reduce false positives
		// The missing imports detection is handled by analyzeMissingImports() function
		
		// Undefined middleware functions
		{
			Pattern:     regexp.MustCompile(`app\.(use|get|post|put|delete|patch)\([^)]*[A-Z][a-zA-Z]*[^)]*\)`),
			Type:        "UNDEFINED_MIDDLEWARE",
			Severity:    "CRITICAL", 
			Message:     "Middleware function used but not defined",
			Impact:      "Server routes will fail - TypeError: middleware is not a function",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Define middleware function or import it",
		},
		
		// Missing async/await on database operations
		{
			Pattern:     regexp.MustCompile(`\b(findOne|find|save|create|update|delete|aggregate)\([^)]*\)\s*[^.]`),
			Type:        "MISSING_ASYNC_DB",
			Severity:    "HIGH",
			Message:     "Database operation missing await",
			Impact:      "Database queries will return promises instead of data",
			Category:    "RUNTIME_ERROR", 
			FixTemplate: "Add 'await' before database operation or use .then()",
		},
		
		// Unhandled promise rejections
		{
			Pattern:     regexp.MustCompile(`\.catch\s*\(\s*\)\s*[;}]`),
			Type:        "EMPTY_CATCH",
			Severity:    "HIGH",
			Message:     "Empty catch block will hide errors",
			Impact:      "Unhandled promise rejections may crash Node.js process",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Add error handling in catch block",
		},
		
		// Missing error handling in async functions
		{
			Pattern:     regexp.MustCompile(`async\s+function\s+\w+\s*\([^)]*\)`),
			Type:        "MISSING_ERROR_HANDLING",
			Severity:    "HIGH",
			Message:     "Async function missing error handling",
			Impact:      "Unhandled exceptions may crash the server",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Wrap in try-catch block",
		},
		
		// Using res.send() multiple times
		{
			Pattern:     regexp.MustCompile(`res\.(send|json|status|end).*res\.(send|json|status|end)`),
			Type:        "MULTIPLE_RESPONSE",
			Severity:    "CRITICAL",
			Message:     "Multiple response methods called",
			Impact:      "Error: Cannot set headers after they are sent",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Use only one response method per request",
		},
		
		// Missing connection string or config - only for database connections
		{
			Pattern:     regexp.MustCompile(`(mongoose|db|database|sequelize|prisma)\.connect\(['"]\s*['"]|connect\(\s*\).*(?:mongoose|mongodb|postgres|mysql|database)`),
			Type:        "MISSING_CONNECTION_STRING",
			Severity:    "CRITICAL",
			Message:     "Database connection missing connection string",
			Impact:      "Database connection will fail on startup",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Provide valid connection string",
		},
		
		// Port conflicts or missing port
		{
			Pattern:     regexp.MustCompile(`listen\(\s*(?:process\.env\.PORT|PORT|3000)\s*\)`),
			Type:        "HARDCODED_PORT",
			Severity:    "MEDIUM",
			Message:     "Hardcoded port may cause conflicts in production",
			Impact:      "EADDRINUSE error if port is already in use",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Use process.env.PORT || fallback",
		},
		
		// Synchronous file operations in server code
		{
			Pattern:     regexp.MustCompile(`fs\.(readFileSync|writeFileSync|unlinkSync|mkdirSync)`),
			Type:        "SYNC_FILE_OPERATION",
			Severity:    "HIGH",
			Message:     "Synchronous file operation will block event loop",
			Impact:      "Server becomes unresponsive during file operations",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Use async versions: readFile, writeFile, etc.",
		},
		
		// Missing CORS configuration
		{
			Pattern:     regexp.MustCompile(`app\.(get|post|put|delete).*['"]/api`),
			Type:        "MISSING_CORS",
			Severity:    "MEDIUM",
			Message:     "API routes may need CORS configuration",
			Impact:      "Cross-origin requests will be blocked by browsers",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Add CORS middleware or headers",
		},
		
		// Memory leaks - event listeners not removed
		{
			Pattern:     regexp.MustCompile(`\.(addEventListener|on)\s*\(`),
			Type:        "POTENTIAL_MEMORY_LEAK",
			Severity:    "MEDIUM",
			Message:     "Event listener added but never removed",
			Impact:      "Memory leaks can cause server crashes over time",
			Category:    "RUNTIME_ERROR",
			FixTemplate: "Remove event listeners in cleanup functions",
		},
	}
}

// AnalyzeCriticalIssues detects server-breaking issues in code
func (ca *CriticalAnalyzer) AnalyzeCriticalIssues(filePath string, content string, existingIssues []types.Issue) []CriticalIssue {
	var critical []CriticalIssue
	
	lines := strings.Split(content, "\n")
	isServerFile := ca.isServerFile(filePath, content)
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		for _, pattern := range ca.criticalPatterns {
			if pattern.Pattern.MatchString(line) {
				// Higher severity for server files
				severity := pattern.Severity
				if isServerFile && severity == "MEDIUM" {
					severity = "HIGH"
				}
				
				critical = append(critical, CriticalIssue{
					FilePath:      filePath,
					Line:          lineNum + 1,
					Type:          pattern.Type,
					Severity:      severity,
					Message:       pattern.Message,
					Impact:        pattern.Impact,
					FixSuggestion: pattern.FixTemplate,
					Category:      pattern.Category,
				})
			}
		}
	}
	
	// Analyze import issues based on usage
	critical = append(critical, ca.analyzeMissingImports(filePath, content)...)
	
	return critical
}

func (ca *CriticalAnalyzer) isServerFile(filePath string, content string) bool {
	// Check if file is likely a server file
	serverIndicators := []string{
		"express", "app.listen", "server.listen", "require('http')",
		"import.*express", "app.use", "app.get", "app.post",
		"middleware", "router", "controller", "api/", "routes/",
		"mongoose", "sequelize", "prisma", "database", "connection",
	}
	
	lowerContent := strings.ToLower(content)
	lowerPath := strings.ToLower(filePath)
	
	// Check file path
	if strings.Contains(lowerPath, "server") || 
	   strings.Contains(lowerPath, "api") ||
	   strings.Contains(lowerPath, "routes") ||
	   strings.Contains(lowerPath, "controller") ||
	   strings.Contains(lowerPath, "middleware") {
		return true
	}
	
	// Check content for server indicators
	indicatorCount := 0
	for _, indicator := range serverIndicators {
		if strings.Contains(lowerContent, strings.ToLower(indicator)) {
			indicatorCount++
		}
	}
	
	return indicatorCount >= 2
}

func (ca *CriticalAnalyzer) analyzeMissingImports(filePath string, content string) []CriticalIssue {
	var critical []CriticalIssue
	
	// Get all imports
	importedModules := ca.getImportedModules(content)
	
	// Check for usage of common server modules without imports
	// Using more precise patterns to reduce false positives
	usagePatterns := map[string][]string{
		"express": {"express()", "express.Router()"},
		"mongoose": {"mongoose.connect", "mongoose.model", "mongoose.Schema"},
		"http": {"http.createServer", "http.get", "http.request"},
		"path": {"path.join", "path.resolve"},
		"fs": {"fs.readFile", "fs.writeFile", "fs.existsSync"},
		"crypto": {"crypto.createHash", "crypto.randomBytes"},
		"util": {"util.promisify", "util.inspect"},
		"querystring": {"querystring.parse", "querystring.stringify"},
		"url": {"url.parse", "new URL("},
	}
	
	for module, patterns := range usagePatterns {
		if !importedModules[module] {
			for _, pattern := range patterns {
				if strings.Contains(content, pattern) {
					severity := "HIGH"
					if module == "express" || module == "mongoose" || module == "http" {
						severity = "CRITICAL"
					}
					
					critical = append(critical, CriticalIssue{
						FilePath:      filePath,
						Line:          1, // Top of file
						Type:          "MISSING_CRITICAL_IMPORT",
						Severity:      severity,
						Message:       fmt.Sprintf("Missing import for '%s' module", module),
						Impact:        fmt.Sprintf("ReferenceError: %s is not defined - server will crash", module),
						FixSuggestion: fmt.Sprintf("Add: const %s = require('%s'); or import %s from '%s';", module, module, module, module),
						Category:      "IMPORT_ERROR",
					})
					break // Only report once per module
				}
			}
		}
	}
	
	return critical
}

func (ca *CriticalAnalyzer) getImportedModules(content string) map[string]bool {
	imported := make(map[string]bool)
	
	// Match require statements
	requireRegex := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
	matches := requireRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			module := strings.Split(match[1], "/")[0] // Get base module name
			imported[module] = true
		}
	}
	
	// Match ES6 import statements with more flexibility
	// Handles: import express from 'express', import { Router } from 'express', etc.
	importRegex := regexp.MustCompile(`import\s+(?:[^{}\s]+|\{[^}]*\}|\*\s+as\s+\w+)\s+from\s+['"]([^'"]+)['"]`)
	matches = importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			module := strings.Split(match[1], "/")[0] // Get base module name
			imported[module] = true
		}
	}
	
	return imported
}

// ConvertToIssues converts critical issues to standard Issue format
func ConvertCriticalToIssues(issues []CriticalIssue) []types.Issue {
	var converted []types.Issue
	
	for _, issue := range issues {
		severity := 2 // Default to error
		if issue.Severity == "MEDIUM" {
			severity = 1 // Warning
		}
		
		message := fmt.Sprintf("[%s] %s - %s", issue.Severity, issue.Message, issue.Impact)
		
		converted = append(converted, types.Issue{
			FilePath: issue.FilePath,
			Line:     issue.Line,
			RuleID:   fmt.Sprintf("critical-%s", strings.ToLower(issue.Type)),
			Message:  message,
			Severity: severity,
		})
	}
	
	return converted
}

// PrintCriticalAnalysis displays critical issues with detailed information
func PrintCriticalAnalysis(issues []CriticalIssue) {
	if len(issues) == 0 {
		fmt.Println("✅ No critical server-breaking issues detected!")
		return
	}
	
	// Group by severity
	critical := []CriticalIssue{}
	high := []CriticalIssue{}
	medium := []CriticalIssue{}
	
	for _, issue := range issues {
		switch issue.Severity {
		case "CRITICAL":
			critical = append(critical, issue)
		case "HIGH":
			high = append(high, issue)
		case "MEDIUM":
			medium = append(medium, issue)
		}
	}
	
	fmt.Println("🚨 CRITICAL SERVER ISSUES ANALYSIS")
	fmt.Println("=" + strings.Repeat("=", 45))
	fmt.Println()
	
	if len(critical) > 0 {
		fmt.Printf("🔴 CRITICAL ISSUES (%d) - WILL BREAK SERVER:\n", len(critical))
		for i, issue := range critical {
			fmt.Printf("  %d. %s:%d\n", i+1, issue.FilePath, issue.Line)
			fmt.Printf("     💥 %s\n", issue.Message)
			fmt.Printf("     ⚡ Impact: %s\n", issue.Impact)
			fmt.Printf("     🔧 Fix: %s\n", issue.FixSuggestion)
			fmt.Println()
		}
	}
	
	if len(high) > 0 {
		fmt.Printf("🟠 HIGH PRIORITY ISSUES (%d) - MAY BREAK SERVER:\n", len(high))
		for i, issue := range high {
			if i < 5 { // Show top 5
				fmt.Printf("  %d. %s:%d - %s\n", i+1, issue.FilePath, issue.Line, issue.Message)
				fmt.Printf("     🔧 Fix: %s\n", issue.FixSuggestion)
				fmt.Println()
			}
		}
		if len(high) > 5 {
			fmt.Printf("     ... and %d more high priority issues\n\n", len(high)-5)
		}
	}
	
	if len(medium) > 0 {
		fmt.Printf("🟡 MEDIUM ISSUES (%d) - SHOULD BE FIXED:\n", len(medium))
		for i, issue := range medium {
			if i < 3 { // Show top 3
				fmt.Printf("  %d. %s:%d - %s\n", i+1, issue.FilePath, issue.Line, issue.Message)
			}
		}
		if len(medium) > 3 {
			fmt.Printf("     ... and %d more medium priority issues\n", len(medium)-3)
		}
		fmt.Println()
	}
	
	// Summary and recommendations
	fmt.Println("📊 SUMMARY:")
	fmt.Printf("   🔴 Critical: %d (fix immediately)\n", len(critical))
	fmt.Printf("   🟠 High: %d (fix before deployment)\n", len(high))
	fmt.Printf("   🟡 Medium: %d (fix when possible)\n", len(medium))
	fmt.Println()
	
	if len(critical) > 0 {
		fmt.Println("⚠️  RECOMMENDATION: Do not deploy until critical issues are resolved!")
	} else if len(high) > 0 {
		fmt.Println("💡 RECOMMENDATION: Address high priority issues before production deployment")
	} else {
		fmt.Println("✅ No blocking issues found - safe to deploy")
	}
}