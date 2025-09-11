package spellchecker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// SpellChecker represents a spell checking instance
type SpellChecker struct {
	dictionary    map[string]bool
	techTerms     map[string]bool
	jsKeywords    map[string]bool
	suggestions   map[string][]string
	wordRegex     *regexp.Regexp
	camelCaseRegex *regexp.Regexp
}

// SpellIssue represents a spelling issue found in code
type SpellIssue struct {
	Word        string
	Line        int
	Column      int
	Context     string
	Type        string   // "comment", "string", "identifier"
	Suggestions []string
	Severity    string   // "low", "medium", "high"
}

// NewSpellChecker creates a new spell checker instance
func NewSpellChecker() (*SpellChecker, error) {
	sc := &SpellChecker{
		dictionary:     make(map[string]bool),
		techTerms:      make(map[string]bool),
		jsKeywords:     make(map[string]bool),
		suggestions:    make(map[string][]string),
		wordRegex:      regexp.MustCompile(`\b[A-Za-z][a-zA-Z']*\b`),
		camelCaseRegex: regexp.MustCompile(`[A-Z][a-z]+|[a-z]+`),
	}

	// Load built-in dictionaries
	if err := sc.loadBuiltinDictionaries(); err != nil {
		return nil, err
	}

	return sc, nil
}

// loadBuiltinDictionaries loads built-in word lists
func (sc *SpellChecker) loadBuiltinDictionaries() error {
	// Common English words (subset of most frequently used words)
	commonWords := []string{
		// Articles, pronouns, prepositions
		"the", "be", "to", "of", "and", "a", "in", "that", "have", "it", "for", "not", "on", "with", "as", "you", "do", "at", "this", "but", "his", "by", "from", "they", "we", "say", "her", "she", "or", "an", "will", "my", "one", "all", "would", "there", "their",

		// Common verbs
		"get", "make", "go", "know", "take", "see", "come", "could", "time", "work", "first", "way", "about", "many", "then", "may", "use", "find", "give", "over", "think", "also", "after", "back", "other", "call", "want", "look", "right", "now", "its", "before", "here", "through", "when", "where", "much", "should", "well", "people",

		// Common nouns
		"year", "work", "part", "number", "way", "day", "man", "world", "life", "hand", "system", "program", "question", "government", "company", "problem", "service", "thing", "woman", "place", "case", "point", "group", "fact", "money", "story", "example", "state", "business", "night", "area", "water", "book", "eye", "head", "information", "school", "room", "home", "office",

		// Common adjectives
		"good", "new", "different", "following", "same", "important", "small", "large", "next", "early", "young", "public", "bad", "few", "own", "general", "high", "long", "great", "little", "old", "right", "big", "national", "major", "best", "economic", "strong", "possible", "whole", "free", "military", "true", "federal", "international", "full", "special", "easy", "clear", "recent",

		// Programming common words
		"data", "function", "method", "class", "object", "property", "value", "parameter", "argument", "variable", "constant", "array", "list", "string", "number", "boolean", "null", "undefined", "return", "result", "response", "request", "error", "exception", "message", "event", "handler", "callback", "promise", "async", "await", "component", "module", "library", "framework", "application", "service", "client", "server", "database", "file", "path", "url", "api", "json", "xml", "html", "css", "javascript", "typescript", "node", "react", "vue", "angular", "express", "mongoose", "axios", "lodash", "moment", "jquery", "bootstrap",
	}

	// Load common words
	for _, word := range commonWords {
		sc.dictionary[strings.ToLower(word)] = true
	}

	// Technical terms and abbreviations commonly used in programming
	techTerms := []string{
		// General tech terms
		"api", "url", "uri", "http", "https", "ftp", "ssl", "tls", "tcp", "udp", "ip", "dns", "cdn", "aws", "gcp", "azure", "docker", "kubernetes", "git", "github", "gitlab", "bitbucket", "npm", "yarn", "webpack", "babel", "eslint", "prettier", "jest", "mocha", "chai", "cypress", "selenium", "jenkins", "travis", "circleci",

		// JavaScript/Web specific
		"dom", "bom", "ajax", "xhr", "fetch", "cors", "jwt", "oauth", "ssr", "spa", "pwa", "seo", "ui", "ux", "css", "sass", "scss", "less", "html", "jsx", "tsx", "json", "xml", "yaml", "svg", "png", "jpg", "jpeg", "gif", "webp", "ico",

		// Database terms
		"sql", "nosql", "mongodb", "mysql", "postgresql", "sqlite", "redis", "elasticsearch", "crud", "orm", "odm", "schema", "migration", "seed", "transaction", "rollback", "commit",

		// Development terms
		"ide", "vscode", "atom", "sublime", "vim", "emacs", "cli", "gui", "sdk", "jdk", "ndk", "api", "rest", "graphql", "websocket", "microservice", "monolith", "devops", "cicd", "qa", "uat", "prod", "dev", "staging", "localhost", "debug", "profiling", "optimization", "refactor", "tech", "stack", "backend", "frontend", "fullstack",

		// File extensions and formats
		"js", "ts", "jsx", "tsx", "css", "scss", "sass", "less", "html", "htm", "xml", "json", "yaml", "yml", "md", "txt", "csv", "pdf", "doc", "docx", "xls", "xlsx", "zip", "tar", "gz", "env", "config", "conf", "ini", "log",

		// Common abbreviations
		"etc", "i.e", "e.g", "vs", "aka", "faq", "tbd", "todo", "fixme", "hack", "temp", "tmp", "min", "max", "avg", "std", "dev", "prod", "qa", "uat", "src", "lib", "dist", "build", "public", "assets", "static", "components", "utils", "helpers", "services", "controllers", "models", "views", "routes", "middleware", "plugins", "extensions", "addons",
	}

	for _, term := range techTerms {
		sc.techTerms[strings.ToLower(term)] = true
		sc.dictionary[strings.ToLower(term)] = true
	}

	// JavaScript keywords and built-ins
	jsTerms := []string{
		// Keywords
		"abstract", "arguments", "await", "boolean", "break", "byte", "case", "catch", "char", "class", "const", "continue", "debugger", "default", "delete", "do", "double", "else", "enum", "eval", "export", "extends", "false", "final", "finally", "float", "for", "function", "goto", "if", "implements", "import", "in", "instanceof", "int", "interface", "let", "long", "native", "new", "null", "package", "private", "protected", "public", "return", "short", "static", "super", "switch", "synchronized", "this", "throw", "throws", "transient", "true", "try", "typeof", "var", "void", "volatile", "while", "with", "yield",

		// Built-in objects
		"array", "boolean", "date", "error", "function", "json", "math", "number", "object", "regexp", "string", "symbol", "promise", "proxy", "reflect", "map", "set", "weakmap", "weakset", "arraybuffer", "dataview", "int8array", "uint8array", "int16array", "uint16array", "int32array", "uint32array", "float32array", "float64array",

		// Global functions and properties
		"isnan", "isfinite", "parseint", "parsefloat", "encodeuricomponent", "decodeuricomponent", "encodeuri", "decodeuri", "escape", "unescape", "console", "window", "document", "navigator", "location", "history", "screen", "settimeout", "setinterval", "cleartimeout", "clearinterval", "requestanimationframe", "cancelanimationframe",

		// Common method names
		"push", "pop", "shift", "unshift", "slice", "splice", "concat", "join", "split", "reverse", "sort", "filter", "map", "reduce", "foreach", "find", "findindex", "includes", "indexof", "lastindexof", "startswith", "endswith", "replace", "match", "search", "substring", "substr", "trim", "tolowercase", "touppercase", "charat", "charcodeat", "fromcharcode",
	}

	for _, term := range jsTerms {
		sc.jsKeywords[strings.ToLower(term)] = true
		sc.dictionary[strings.ToLower(term)] = true
	}

	// Load common misspellings and their corrections
	sc.loadCommonMisspellings()

	return nil
}

// loadCommonMisspellings loads common programming-related misspellings
func (sc *SpellChecker) loadCommonMisspellings() {
	corrections := map[string][]string{
		// Common programming misspellings
		"lenght":       {"length"},
		"widht":        {"width"},
		"hieght":       {"height"},
		"calback":      {"callback"},
		"fucntion":     {"function"},
		"funciton":     {"function"},
		"funtion":      {"function"},
		"retrun":       {"return"},
		"reutrn":       {"return"},
		"varialbe":     {"variable"},
		"vairable":     {"variable"},
		"variabel":     {"variable"},
		"compnent":     {"component"},
		"componenet":   {"component"},
		"compoent":     {"component"},
		"handelr":      {"handler"},
		"handlr":       {"handler"},
		"handel":       {"handle"},
		"seperate":     {"separate"},
		"seperator":    {"separator"},
		"recieve":      {"receive"},
		"recieved":     {"received"},
		"acheive":      {"achieve"},
		"acheived":     {"achieved"},
		"occured":      {"occurred"},
		"occurence":    {"occurrence"},
		"definately":   {"definitely"},
		"seperated":    {"separated"},
		"calender":     {"calendar"},
		"succesful":    {"successful"},
		"successfull":  {"successful"},
		"accomodate":   {"accommodate"},
		"acomodate":    {"accommodate"},
		"begining":     {"beginning"},
		"comming":      {"coming"},
		"geting":       {"getting"},
		"runing":       {"running"},
		"stoping":      {"stopping"},
		"writting":     {"writing"},
		"payed":        {"paid"},

		// Tech-specific misspellings
		"authetication": {"authentication"},
		"authenication": {"authentication"},
		"authroization": {"authorization"},
		"syncronous":    {"synchronous"},
		"asyncronous":   {"asynchronous"},
		"peformance":    {"performance"},
		"preformance":   {"performance"},
		"responsability": {"responsibility"},
		"responsable":   {"responsible"},
		"initalize":     {"initialize"},
		"intialize":     {"initialize"},
		"paramater":     {"parameter"},
		"paramter":      {"parameter"},
		"algoritm":      {"algorithm"},
		"algorithem":    {"algorithm"},
		"refference":    {"reference"},
		"referance":     {"reference"},
		"dependancy":    {"dependency"},
		"dependecy":     {"dependency"},
		"libary":        {"library"},
		"libaray":       {"library"},
		"framwork":      {"framework"},
		"framewrok":     {"framework"},
		"databse":       {"database"},
		"databas":       {"database"},
	}

	for wrong, correct := range corrections {
		sc.suggestions[wrong] = correct
	}
}

// CheckFile checks spelling in a JavaScript/TypeScript file
func (sc *SpellChecker) CheckFile(filePath string) ([]SpellIssue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []SpellIssue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regular expressions for different contexts
	commentRegex := regexp.MustCompile(`//(.*)$|/\*([^*]|\*[^/])*\*/`)
	stringRegex := regexp.MustCompile(`["']([^"'\\]|\\.)*["']`)
	identifierRegex := regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check comments
		commentMatches := commentRegex.FindAllStringSubmatch(line, -1)
		for _, match := range commentMatches {
			if len(match) > 1 {
				commentText := match[1]
				if commentText == "" && len(match) > 2 {
					commentText = match[2]
				}
				issues = append(issues, sc.checkTextSpelling(commentText, lineNum, "comment", line)...)
			}
		}

		// Check string literals
		stringMatches := stringRegex.FindAllString(line, -1)
		for _, match := range stringMatches {
			// Remove quotes and check content
			content := strings.Trim(match, `"'`)
			issues = append(issues, sc.checkTextSpelling(content, lineNum, "string", line)...)
		}

		// Check identifiers (variable names, function names, etc.)
		identifierMatches := identifierRegex.FindAllString(line, -1)
		for _, identifier := range identifierMatches {
			// Skip if it's a known keyword or library
			if sc.jsKeywords[strings.ToLower(identifier)] {
				continue
			}
			
			// Check camelCase words
			issues = append(issues, sc.checkCamelCaseSpelling(identifier, lineNum, "identifier", line)...)
		}
	}

	return issues, scanner.Err()
}

// checkTextSpelling checks spelling in plain text (comments, strings)
func (sc *SpellChecker) checkTextSpelling(text string, lineNum int, contextType string, fullLine string) []SpellIssue {
	var issues []SpellIssue
	
	words := sc.wordRegex.FindAllString(text, -1)
	for _, word := range words {
		if !sc.isWordCorrect(word) {
			suggestions := sc.getSuggestions(word)
			severity := "medium"
			if len(suggestions) == 0 {
				severity = "low"
			}
			
			issues = append(issues, SpellIssue{
				Word:        word,
				Line:        lineNum,
				Column:      strings.Index(fullLine, word) + 1,
				Context:     contextType,
				Type:        contextType,
				Suggestions: suggestions,
				Severity:    severity,
			})
		}
	}
	
	return issues
}

// checkCamelCaseSpelling checks spelling in camelCase identifiers
func (sc *SpellChecker) checkCamelCaseSpelling(identifier string, lineNum int, contextType string, fullLine string) []SpellIssue {
	var issues []SpellIssue
	
	// Split camelCase into individual words
	words := sc.camelCaseRegex.FindAllString(identifier, -1)
	
	for _, word := range words {
		if len(word) > 2 && !sc.isWordCorrect(word) { // Skip very short words
			suggestions := sc.getSuggestions(word)
			severity := "low" // Lower severity for identifiers
			
			issues = append(issues, SpellIssue{
				Word:        word,
				Line:        lineNum,
				Column:      strings.Index(fullLine, identifier) + 1,
				Context:     fmt.Sprintf("identifier '%s'", identifier),
				Type:        contextType,
				Suggestions: suggestions,
				Severity:    severity,
			})
		}
	}
	
	return issues
}

// isWordCorrect checks if a word is spelled correctly
func (sc *SpellChecker) isWordCorrect(word string) bool {
	word = strings.ToLower(word)
	
	// Check dictionary
	if sc.dictionary[word] {
		return true
	}
	
	// Check if it's a proper noun (starts with uppercase)
	if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
		return true
	}
	
	// Check if it's all uppercase (might be an acronym)
	if strings.ToUpper(word) == word && len(word) > 1 {
		return true
	}
	
	// Check if it contains numbers (might be a version or identifier)
	for _, r := range word {
		if unicode.IsDigit(r) {
			return true
		}
	}
	
	return false
}

// getSuggestions gets spelling suggestions for a word
func (sc *SpellChecker) getSuggestions(word string) []string {
	word = strings.ToLower(word)
	
	// Check if we have specific suggestions
	if suggestions, exists := sc.suggestions[word]; exists {
		return suggestions
	}
	
	// Generate simple suggestions using edit distance
	suggestions := sc.generateSuggestions(word)
	return suggestions
}

// generateSuggestions generates simple spelling suggestions
func (sc *SpellChecker) generateSuggestions(word string) []string {
	var suggestions []string
	word = strings.ToLower(word)
	
	// Fast suggestions - only check similar length words first
	targetLen := len(word)
	
	// Check tech terms first (higher priority)
	for techWord := range sc.techTerms {
		if len(techWord) >= targetLen-2 && len(techWord) <= targetLen+2 {
			if sc.editDistance(word, techWord) <= 1 && len(techWord) > 2 {
				suggestions = append(suggestions, techWord)
				if len(suggestions) >= 3 {
					return suggestions
				}
			}
		}
	}
	
	// Then check dictionary words with length filter
	checked := 0
	maxCheck := 1000 // Limit how many words we check for performance
	for dictWord := range sc.dictionary {
		if checked >= maxCheck {
			break
		}
		
		if len(dictWord) >= targetLen-2 && len(dictWord) <= targetLen+2 {
			if sc.editDistance(word, dictWord) <= 2 && len(dictWord) > 2 {
				suggestions = append(suggestions, dictWord)
				if len(suggestions) >= 3 {
					break
				}
			}
		}
		checked++
	}
	
	return suggestions
}

// editDistance calculates the Levenshtein distance between two strings
func (sc *SpellChecker) editDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// LoadExternalDictionary loads an external dictionary file
func (sc *SpellChecker) LoadExternalDictionary(dictPath string) error {
	// Try to load from common dictionary paths
	possiblePaths := []string{
		dictPath,
		filepath.Join(os.Getenv("HOME"), ".codecompass", "dictionary.txt"),
		"/usr/share/dict/words",
		"/usr/dict/words",
		"/usr/share/hunspell/en_US.dic",
	}
	
	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		
		if err := sc.loadDictionaryFile(path); err == nil {
			fmt.Printf("📚 Loaded external dictionary from: %s\n", path)
			return nil
		}
	}
	
	// If no external dictionary found, continue with built-in
	fmt.Printf("📝 Using built-in dictionary (%d words)\n", len(sc.dictionary))
	return nil
}

// loadDictionaryFile loads words from a dictionary file
func (sc *SpellChecker) loadDictionaryFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	count := 0
	maxWords := 50000 // Limit dictionary size for performance
	
	for scanner.Scan() && count < maxWords {
		word := strings.TrimSpace(strings.ToLower(scanner.Text()))
		// Skip very short or very long words, proper nouns, and words with special characters
		if len(word) >= 3 && len(word) <= 15 && word != "" {
			// Skip words that start with uppercase (proper nouns in some dictionaries)
			firstChar := scanner.Text()[0]
			if firstChar >= 'a' && firstChar <= 'z' {
				// Only include common words (basic filter)
				if !strings.ContainsAny(word, "'-.,;:!?()[]{}\"") {
					sc.dictionary[word] = true
					count++
				}
			}
		}
	}
	
	if count > 0 {
		fmt.Printf("📖 Loaded %d words from dictionary\n", count)
	}
	
	return scanner.Err()
}