package spellcheck

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"codecompass/internal/config"
	"codecompass/internal/git"
	"codecompass/internal/types"
	"codecompass/internal/utils"

	"github.com/sajari/fuzzy"
)

type SpellChecker struct {
	model      *fuzzy.Model
	customDict map[string]bool
	progDict   map[string]bool
}

func NewSpellChecker(cfg *config.Config) (*SpellChecker, error) {
	// Initialize fuzzy model
	model := fuzzy.NewModel()
	model.SetThreshold(1) // Set edit distance threshold

	// Load common English words
	commonWords := loadCommonWords()
	model.Train(commonWords)

	return &SpellChecker{
		model:      model,
		customDict: createCustomDict(cfg.CustomWords),
		progDict:   createProgrammingDict(),
	}, nil
}

func (sc *SpellChecker) IsCorrect(word string) bool {
	return isCorrectlySpelled(word)
}

func (sc *SpellChecker) GetSuggestions(word string) []string {
	var suggestions []string

	// Get suggestions from fuzzy model
	fuzzySuggestions := sc.model.Suggestions(word, false)
	suggestions = append(suggestions, fuzzySuggestions...)

	// Add custom suggestions for common programming typos
	commonTypos := map[string][]string{
		"recieve":    {"receive"},
		"lenght":     {"length"},
		"widht":      {"width"},
		"heigth":     {"height"},
		"parmeter":   {"parameter"},
		"seperate":   {"separate"},
		"definately": {"definitely"},
		"occured":    {"occurred"},
		"compnent":   {"component"},
		"functon":    {"function"},
		"retrun":     {"return"},
		"calback":    {"callback"},
		"promis":     {"promise"},
		"responce":   {"response"},
		"requets":    {"request"},
		"intialize":  {"initialize"},
		"seralize":   {"serialize"},
		"destory":    {"destroy"},
		"conection":  {"connection"},
		"sucessful":  {"successful"},
	}

	lowerWord := strings.ToLower(word)
	if typoSuggestions, exists := commonTypos[lowerWord]; exists {
		// Add custom suggestions first (more relevant for programming)
		result := make([]string, 0, len(typoSuggestions)+len(suggestions))
		result = append(result, typoSuggestions...)
		result = append(result, suggestions...)
		suggestions = result
	}

	// Remove duplicates and limit suggestions
	seen := make(map[string]bool)
	uniqueSuggestions := make([]string, 0)
	for _, suggestion := range suggestions {
		if !seen[suggestion] && len(uniqueSuggestions) < 5 {
			seen[suggestion] = true
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
		}
	}

	return uniqueSuggestions
}

func loadCommonWords() []string {
	// Common English words - in production, load from external file
	commonWords := []string{
		"the", "and", "for", "are", "but", "not", "you", "all", "can", "had",
		"her", "was", "one", "our", "out", "day", "get", "has", "him", "his",
		"how", "its", "may", "new", "now", "old", "see", "two", "way", "who",
		"been", "each", "from", "have", "here", "just", "like", "long", "make",
		"many", "over", "such", "take", "than", "them", "well", "when", "with",
		"word", "work", "call", "came", "come", "find", "first", "good", "great",
		"group", "hand", "help", "kind", "know", "last", "left", "life", "line",
		"list", "look", "made", "most", "move", "must", "name", "need", "next",
		"only", "open", "part", "play", "right", "said", "same", "seem", "show",
		"small", "sound", "still", "time", "turn", "want", "water", "went", "what",
		"where", "will", "write", "would", "year", "your", "about", "after", "again",
		"against", "always", "another", "around", "because", "before", "being",
		"between", "could", "create", "delete", "different", "during", "every",
		"example", "expected", "function", "general", "getting", "having", "however",
		"important", "including", "inside", "instead", "might", "nothing", "number",
		"object", "other", "people", "person", "place", "please", "point", "possible",
		"problem", "public", "really", "return", "should", "simple", "since",
		"something", "state", "string", "system", "their", "there", "these", "thing",
		"think", "those", "three", "through", "today", "together", "under", "until",
		"value", "while", "without", "world", "years", "yourself",
		// Add more common words as needed
	}

	return commonWords
}

func createCustomDict(customWords []string) map[string]bool {
	dict := make(map[string]bool)
	for _, word := range customWords {
		dict[strings.ToLower(word)] = true
	}
	return dict
}

func createProgrammingDict() map[string]bool {
	return map[string]bool{
		// Programming languages and frameworks
		"javascript": true, "typescript": true, "nodejs": true, "reactjs": true,
		"angular": true, "vuejs": true, "svelte": true, "nextjs": true,
		"webpack": true, "babel": true, "eslint": true, "prettier": true,
		"jest": true, "cypress": true, "storybook": true, "vite": true,

		// Common programming terms
		"async": true, "await": true, "callback": true, "promise": true,
		"json": true, "api": true, "rest": true, "graphql": true,
		"auth": true, "oauth": true, "jwt": true, "cors": true,
		"http": true, "https": true, "url": true, "uri": true,
		"css": true, "html": true, "dom": true, "svg": true,
		"regex": true, "uuid": true, "csrf": true, "xss": true,

		// Database and backend
		"sql": true, "nosql": true, "mongodb": true, "postgresql": true,
		"redis": true, "elasticsearch": true, "microservice": true,
		"orm": true, "crud": true,

		// DevOps and tools
		"docker": true, "kubernetes": true, "nginx": true, "apache": true,
		"aws": true, "gcp": true, "azure": true, "terraform": true,
		"github": true, "gitlab": true, "bitbucket": true, "jenkins": true,

		// Common abbreviations
		"btn": true, "img": true, "src": true, "dest": true, "tmp": true,
		"config": true, "env": true, "dev": true, "prod": true, "repo": true,
		"lib": true, "util": true, "ctrl": true, "param": true, "arg": true,
		"idx": true, "len": true, "str": true, "num": true, "obj": true,

		// File extensions
		"js": true, "ts": true, "jsx": true, "tsx": true, "vue": true,
		"scss": true, "sass": true, "xml": true,
		"yaml": true, "toml": true, "md": true, "txt": true,
	}
}

// Common programming terms that shouldn't be flagged as misspelled
var programmingDictionary = map[string]bool{
	// Common programming terms
	"async": true, "await": true, "bool": true, "const": true, "enum": true,
	"func": true, "impl": true, "init": true, "len": true, "mut": true,
	"npm": true, "ptr": true, "ref": true, "struct": true, "typeof": true,
	"var": true, "void": true, "webpack": true, "json": true, "api": true,
	"http": true, "https": true, "url": true, "uri": true, "uuid": true,
	"auth": true, "oauth": true, "jwt": true, "cors": true, "csrf": true,
	"sql": true, "nosql": true, "db": true, "orm": true, "crud": true,
	"ui": true, "ux": true, "css": true, "html": true, "dom": true,
	"xml": true, "yaml": true, "toml": true, "csv": true, "pdf": true,
	"img": true, "svg": true, "png": true, "jpg": true, "jpeg": true,
	"utf": true, "ascii": true, "unicode": true, "regex": true, "regexp": true,
	"app": true, "config": true, "env": true, "dev": true, "prod": true,
	"test": true, "spec": true, "mock": true, "stub": true, "lint": true,
	"eslint": true, "prettier": true, "babel": true, "typescript": true,
	"javascript": true, "nodejs": true, "react": true, "vue": true,
	"angular": true, "jquery": true, "lodash": true, "axios": true,
	"github": true, "gitlab": true, "bitbucket": true, "docker": true,
	"kubernetes": true, "aws": true, "gcp": true, "azure": true,
	// Common abbreviations
	"btn": true, "src": true, "dest": true, "temp": true,
	"tmp": true, "min": true, "max": true, "avg": true, "std": true,
	"admin": true, "user": true, "usr": true, "pwd": true,
	"repo": true, "pkg": true, "lib": true, "util": true, "utils": true,
	"ctrl": true, "cmd": true, "arg": true, "args": true, "param": true,
	"params": true, "opt": true, "opts": true, "req": true, "res": true,
	"err": true, "ctx": true, "cfg": true, "idx": true, "num": true,
}

// Basic English dictionary (in real implementation, you'd load from a file)
var basicDictionary = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true,
	"not": true, "you": true, "all": true, "can": true, "had": true,
	"her": true, "was": true, "one": true, "our": true, "out": true,
	"day": true, "get": true, "has": true, "him": true, "his": true,
	"how": true, "its": true, "may": true, "new": true, "now": true,
	"old": true, "see": true, "two": true, "way": true, "who": true,
	"boy": true, "did": true, "man": true, "end": true, "why": true,
	"let": true, "put": true, "say": true, "she": true, "too": true,
	"use": true, "been": true, "each": true, "from": true, "have": true,
	"here": true, "just": true, "like": true, "long": true, "make": true,
	"many": true, "over": true, "such": true, "take": true, "than": true,
	"them": true, "well": true, "when": true, "with": true, "word": true,
	"work": true, "call": true, "came": true, "come": true, "find": true,
	"first": true, "good": true, "great": true, "group": true, "hand": true,
	"help": true, "kind": true, "know": true, "last": true, "left": true,
	"life": true, "line": true, "list": true, "look": true, "made": true,
	"most": true, "move": true, "must": true, "name": true, "need": true,
	"next": true, "only": true, "open": true, "part": true, "play": true,
	"right": true, "said": true, "same": true, "seem": true, "show": true,
	"small": true, "sound": true, "still": true, "time": true, "turn": true,
	"want": true, "water": true, "went": true, "what": true, "where": true,
	"will": true, "write": true, "would": true, "year": true, "your": true, "hello": true, "test": true, "code": true,
	// Add more common words as needed
	"about": true, "after": true, "again": true, "against": true, "always": true,
	"another": true, "around": true, "because": true, "before": true, "being": true,
	"between": true, "could": true, "create": true, "delete": true, "different": true,
	"during": true, "every": true, "example": true, "expected": true, "function": true,
	"general": true, "getting": true, "having": true, "however": true, "important": true,
	"including": true, "inside": true, "instead": true, "might": true, "nothing": true,
	"number": true, "object": true, "other": true, "people": true, "person": true,
	"place": true, "please": true, "point": true, "possible": true, "problem": true,
	"public": true, "really": true, "return": true, "should": true, "simple": true,
	"since": true, "something": true, "state": true, "string": true, "system": true,
	"their": true, "there": true, "these": true, "thing": true, "think": true,
	"those": true, "three": true, "through": true, "today": true, "together": true,
	"under": true, "until": true, "value": true, "while": true, "without": true,
	"world": true, "years": true, "yourself": true,
}

func AnalyzeSpelling(trackedFiles map[string]bool, cfg *config.Config) ([]types.SpellCheckEntry, map[string]*types.SpellCheckAuthorStats, error) {
	spellChecker, err := NewSpellChecker(cfg)
	if err != nil {
		return nil, nil, err
	}

	var entries []types.SpellCheckEntry
	authorStats := make(map[string]*types.SpellCheckAuthorStats)

	for filePath := range trackedFiles {
		if !isSpellCheckFile(filePath, cfg) {
			continue
		}

		entry, fileAuthorStats, err := analyzeFileSpelling(filePath, spellChecker, cfg)
		if err != nil {
			continue
		}

		if entry.TotalWords > 0 {
			entries = append(entries, entry)

			// Merge file author stats
			for email, stats := range fileAuthorStats {
				if authorStats[email] == nil {
					authorStats[email] = &types.SpellCheckAuthorStats{
						Name:           stats.Name,
						Email:          email,
						TotalErrors:    0,
						Files:          make(map[string]int),
						CommonMistakes: make(map[string]int),
					}
				}

				authorStats[email].TotalErrors += stats.TotalErrors
				authorStats[email].Files[filePath] += stats.TotalErrors
				for mistake, count := range stats.CommonMistakes {
					authorStats[email].CommonMistakes[mistake] += count
				}
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ErrorRate > entries[j].ErrorRate
	})

	return entries, authorStats, nil
}

func analyzeFileSpelling(filePath string, spellChecker *SpellChecker, cfg *config.Config) (types.SpellCheckEntry, map[string]*types.SpellCheckAuthorStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return types.SpellCheckEntry{}, nil, err
	}
	defer file.Close()

	entry := types.SpellCheckEntry{
		Path:            filePath,
		TopMisspellings: make(map[string]int),
		Issues:          []types.SpellIssue{},
	}

	authorStats := make(map[string]*types.SpellCheckAuthorStats)

	// Get git blame for this file
	blameMap, err := getBlameForSpellCheck(filePath, cfg)
	if err != nil {
		blameMap = make(map[int]types.BlameInfo)
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Focus mainly on comments and documentation
	commentRegex := regexp.MustCompile(`//\s*(.+)|/\*([^*]|\*[^/])*\*/|#\s*(.+)|<!--([^>]*)-->`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip lines that are mostly code
		if isCodeLine(line) {
			continue
		}

		// Get blame info for this line
		var blameInfo *types.BlameInfo
		if info, exists := blameMap[lineNum]; exists {
			blameInfo = &info
		}

		// Check comments (primary focus)
		if matches := commentRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				for i := 1; i < len(match); i++ {
					if match[i] != "" && len(strings.TrimSpace(match[i])) > 3 {
						analyzeText(match[i], lineNum, "comment", &entry, authorStats, blameInfo, spellChecker)
					}
				}
			}
		}
	}

	// Calculate error rate
	if entry.TotalWords > 0 {
		entry.ErrorRate = float64(entry.MisspelledWords) / float64(entry.TotalWords) * 100
	}

	return entry, authorStats, scanner.Err()
}

func getBlameForSpellCheck(filePath string, cfg *config.Config) (map[int]types.BlameInfo, error) {
	var mu sync.Mutex
	var warningLogs []string
	semaphore := utils.NewSemaphore(cfg.GetConcurrency())

	return git.BlameFile(filePath, &warningLogs, &mu, semaphore)
}

func analyzeText(text string, lineNum int, contextType string, entry *types.SpellCheckEntry, authorStats map[string]*types.SpellCheckAuthorStats, blameInfo *types.BlameInfo, spellChecker *SpellChecker) {
	words := extractWords(text)

	for _, word := range words {
		// Skip very short words or likely code
		if len(word) < 4 || isLikelyCode(word) {
			continue
		}

		entry.TotalWords++

		if !spellChecker.IsCorrect(word) {
			entry.MisspelledWords++
			entry.TopMisspellings[strings.ToLower(word)]++

			issue := types.SpellIssue{
				Word:        word,
				Line:        lineNum,
				Context:     text,
				Type:        contextType,
				Suggestions: spellChecker.GetSuggestions(word),
				Author:      getAuthorName(blameInfo),
				AuthorEmail: getAuthorEmail(blameInfo),
			}
			entry.Issues = append(entry.Issues, issue)

			// Track author statistics
			if blameInfo != nil {
				updateAuthorStats(authorStats, blameInfo, word)
			}
		}
	}
}

func analyzeIdentifier(identifier string, lineNum int, entry *types.SpellCheckEntry, authorStats map[string]*types.SpellCheckAuthorStats, blameInfo *types.BlameInfo, customDict map[string]bool) {
	words := splitIdentifier(identifier)

	for _, word := range words {
		if len(word) < 3 {
			continue
		}

		entry.TotalWords++

		if !isCorrectlySpelledWithCustom(word, customDict) && !isCommonAbbreviation(word) {
			entry.MisspelledWords++
			entry.TopMisspellings[strings.ToLower(word)]++

			issue := types.SpellIssue{
				Word:        word,
				Line:        lineNum,
				Context:     identifier,
				Type:        "identifier",
				Suggestions: getSuggestions(word),
				Author:      getAuthorName(blameInfo),
				AuthorEmail: getAuthorEmail(blameInfo),
			}
			entry.Issues = append(entry.Issues, issue)

			// Track author statistics
			if blameInfo != nil {
				updateAuthorStats(authorStats, blameInfo, word)
			}
		}
	}
}

func isCorrectlySpelledWithCustom(word string, customDict map[string]bool) bool {
	lowerWord := strings.ToLower(word)

	// Check custom dictionary first
	if customDict[lowerWord] {
		return true
	}

	// Then check built-in dictionaries
	return isCorrectlySpelled(word)
}

func updateAuthorStats(authorStats map[string]*types.SpellCheckAuthorStats, blameInfo *types.BlameInfo, word string) {
	email := blameInfo.Email
	if authorStats[email] == nil {
		authorStats[email] = &types.SpellCheckAuthorStats{
			Name:           blameInfo.Name,
			Email:          email,
			TotalErrors:    0,
			Files:          make(map[string]int),
			CommonMistakes: make(map[string]int),
		}
	}

	authorStats[email].TotalErrors++
	authorStats[email].CommonMistakes[strings.ToLower(word)]++
}

func getAuthorName(blameInfo *types.BlameInfo) string {
	if blameInfo != nil {
		return blameInfo.Name
	}
	return "unknown"
}

func getAuthorEmail(blameInfo *types.BlameInfo) string {
	if blameInfo != nil {
		return blameInfo.Email
	}
	return "unknown"
}

func isSpellCheckFile(filePath string, cfg *config.Config) bool {
	if !cfg.SpellCheckEnabled {
		return false
	}

	// Check against ignored paths
	for _, ignorePath := range cfg.SpellCheckIgnorePaths {
		if strings.Contains(filePath, ignorePath) {
			return false
		}
	}

	// Check file extension
	for _, ext := range cfg.SpellCheckExtensions {
		if strings.HasSuffix(strings.ToLower(filePath), strings.ToLower(ext)) {
			return true
		}
	}

	return false
}

func extractWords(text string) []string {
	// Simple word extraction
	wordRegex := regexp.MustCompile(`[a-zA-Z]+`)
	return wordRegex.FindAllString(text, -1)
}

func splitIdentifier(identifier string) []string {
	var words []string
	var currentWord strings.Builder

	for i, char := range identifier {
		if i > 0 && unicode.IsUpper(char) && !unicode.IsUpper(rune(identifier[i-1])) {
			// CamelCase boundary
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}

		if char == '_' || char == '-' {
			// snake_case or kebab-case boundary
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			continue
		}

		currentWord.WriteRune(char)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

func isCorrectlySpelled(word string) bool {
	lowerWord := strings.ToLower(word)

	// Check programming dictionary first
	if programmingDictionary[lowerWord] {
		return true
	}

	// Check basic English dictionary
	if basicDictionary[lowerWord] {
		return true
	}

	// Additional checks for common patterns
	if isCommonPattern(lowerWord) {
		return true
	}

	return false
}

func isCommonPattern(word string) bool {
	// URLs, emails, etc.
	if strings.Contains(word, "http") || strings.Contains(word, "www") || strings.Contains(word, "@") {
		return true
	}

	// File extensions
	extensions := []string{"js", "ts", "jsx", "tsx", "css", "html", "json", "xml", "md", "txt", "log"}
	for _, ext := range extensions {
		if word == ext {
			return true
		}
	}

	// Common abbreviations in programming
	if len(word) <= 4 && strings.ToUpper(word) == word {
		return true // Likely an acronym
	}

	return false
}

func isCommonAbbreviation(word string) bool {
	commonAbbreviations := map[string]bool{
		"btn": true, "img": true, "src": true, "dest": true,
		"tmp": true, "temp": true, "min": true, "max": true,
		"idx": true, "num": true, "str": true, "obj": true,
		"arr": true, "len": true, "cnt": true, "pos": true,
	}

	return commonAbbreviations[strings.ToLower(word)]
}

func isKeyword(word string) bool {
	keywords := map[string]bool{
		"abstract": true, "arguments": true, "await": true, "boolean": true,
		"break": true, "byte": true, "case": true, "catch": true, "char": true,
		"class": true, "const": true, "continue": true, "debugger": true,
		"default": true, "delete": true, "do": true, "double": true, "else": true,
		"enum": true, "eval": true, "export": true, "extends": true, "false": true,
		"final": true, "finally": true, "float": true, "for": true, "function": true,
		"goto": true, "if": true, "implements": true, "import": true, "in": true,
		"instanceof": true, "int": true, "interface": true, "let": true,
		"long": true, "native": true, "new": true, "null": true, "package": true,
		"private": true, "protected": true, "public": true, "return": true,
		"short": true, "static": true, "super": true, "switch": true,
		"synchronized": true, "this": true, "throw": true, "throws": true,
		"transient": true, "true": true, "try": true, "typeof": true, "var": true,
		"void": true, "volatile": true, "while": true, "with": true, "yield": true,
	}

	return keywords[strings.ToLower(word)]
}

func isTextFile(filePath string) bool {
	textExtensions := []string{
		".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".cpp", ".c", ".go",
		".rs", ".php", ".rb", ".swift", ".kt", ".scala", ".clj", ".hs",
		".css", ".scss", ".sass", ".less", ".html", ".htm", ".xml",
		".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
		".md", ".txt", ".rst", ".adoc", ".tex", ".rtf",
		".sh", ".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd",
		".sql", ".graphql", ".gql", ".proto", ".thrift",
	}

	for _, ext := range textExtensions {
		if strings.HasSuffix(strings.ToLower(filePath), ext) {
			return true
		}
	}

	return false
}

func getSuggestions(word string) []string {
	// Simple suggestion algorithm - in a real implementation,
	// you might use a more sophisticated algorithm like Levenshtein distance
	var suggestions []string
	lowerWord := strings.ToLower(word)

	// Check for common typos
	commonTypos := map[string][]string{
		"teh":        {"the"},
		"adn":        {"and"},
		"taht":       {"that"},
		"hte":        {"the"},
		"recieve":    {"receive"},
		"occured":    {"occurred"},
		"seperate":   {"separate"},
		"definately": {"definitely"},
		"neccessary": {"necessary"},
		"begining":   {"beginning"},
		"parmeter":   {"parameter"},
		"lenght":     {"length"},
		"widht":      {"width"},
		"heigth":     {"height"},
	}

	if typoSuggestions, exists := commonTypos[lowerWord]; exists {
		suggestions = append(suggestions, typoSuggestions...)
	}

	// Add some basic suggestions based on dictionary
	for dictWord := range basicDictionary {
		if len(dictWord) == len(word) && levenshteinDistance(lowerWord, dictWord) == 1 {
			suggestions = append(suggestions, dictWord)
			if len(suggestions) >= 3 { // Limit suggestions
				break
			}
		}
	}

	return suggestions
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i // Fix: assign to first element, not entire slice
	}

	for j := 1; j <= len(b); j++ {
		matrix[0][j] = j // Fix: assign to matrix[0][j], not matrix[j]
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = minOfThree(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func minOfThree(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// Helper functions for better filtering
func isCodeLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return true
	}

	// Skip lines with lots of symbols
	symbolCount := 0
	for _, char := range trimmed {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) {
			symbolCount++
		}
	}

	return float64(symbolCount)/float64(len(trimmed)) > 0.5
}

func isHumanReadableString(content string) bool {
	// Skip very short strings
	if len(content) < 5 {
		return false
	}

	// Skip strings that look like code/config
	codePatterns := []string{
		"http://", "https://", "www.", ".com", ".org", ".net",
		"<", ">", "{", "}", "[", "]", "(", ")",
		"=", "+", "-", "*", "/", "%", "&", "|",
	}

	for _, pattern := range codePatterns {
		if strings.Contains(content, pattern) {
			return false
		}
	}

	// Check if it has reasonable letter-to-other ratio
	letterCount := 0
	for _, char := range content {
		if unicode.IsLetter(char) {
			letterCount++
		}
	}

	// At least 70% letters for human-readable text
	return float64(letterCount)/float64(len(content)) > 0.7
}

func isLikelyCode(word string) bool {
	// Skip camelCase variables
	hasLower := false
	hasUpper := false

	for _, char := range word {
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsUpper(char) {
			hasUpper = true
		}
	}

	// Skip typical camelCase
	if hasLower && hasUpper && len(word) < 20 {
		return true
	}

	// Skip words with numbers
	for _, char := range word {
		if unicode.IsDigit(char) {
			return true
		}
	}

	// Skip all-caps words
	if strings.ToUpper(word) == word && len(word) > 1 {
		return true
	}

	return false
}
