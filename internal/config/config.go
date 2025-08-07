package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	IgnoredFiles          []string
	IgnoredAuthors        []string
	IgnoredRules          []string
	IgnoredPaths          []string
	MaxFileSize           int
	MinCoverageThreshold  float64
	MaxConcurrentBlame    int
	CacheResults          bool
	EnableGitHooks        bool
	CustomSettings        map[string]string
	CustomWords           []string
	SpellCheckEnabled     bool
	SpellCheckExtensions  []string
	SpellCheckIgnorePaths []string
	RuffEnabled           bool
	RuffRules             []string
	RuffIgnorePaths       []string
}

func NewConfig() *Config {
	return &Config{
		IgnoredFiles:          []string{},
		IgnoredAuthors:        []string{},
		IgnoredRules:          []string{},
		IgnoredPaths:          []string{},
		MaxFileSize:           5000,
		MinCoverageThreshold:  80.0,
		MaxConcurrentBlame:    4,
		CacheResults:          true,
		EnableGitHooks:        false,
		CustomSettings:        make(map[string]string),
		CustomWords:           []string{},
		SpellCheckEnabled:     true,
		SpellCheckExtensions:  []string{".js", ".ts", ".jsx", ".tsx", ".md", ".txt"},
		SpellCheckIgnorePaths: []string{"node_modules", "dist", "build"},
		RuffEnabled:           true,
		RuffRules:             []string{},
		RuffIgnorePaths:       []string{"node_modules", "dist", "build"},
	}
}

func LoadConfig() (*Config, error) {
	config := NewConfig()

	// Look for config files in order of preference
	configFiles := []string{
		".codecompass.rc",
		".codecompass.config",
		"codecompass.config",
		".codecompass",
	}

	var configFile string
	for _, file := range configFiles {
		if _, err := os.Stat(file); err == nil {
			configFile = file
			break
		}
	}

	if configFile == "" {
		return config, nil // No config file found, return default config
	}

	fmt.Printf("🧭 Using config file: %s\n", configFile)
	return parseConfigFile(configFile, config)
}

func LoadConfigFromFile(filename string) (*Config, error) {
	config := NewConfig()
	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("config file not found: %s", filename)
	}
	return parseConfigFile(filename, config)
}

func parseConfigFile(filename string, config *Config) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				value = strings.Trim(value, `"'`)

				if err := config.parseKeyValue(key, value); err != nil {
					fmt.Printf("Warning: Line %d: %v\n", lineNum, err)
				}
			}
		}
	}

	return config, scanner.Err()
}

func (c *Config) parseKeyValue(key, value string) error {
	switch key {
	case "ignore-files":
		c.IgnoredFiles = append(c.IgnoredFiles, parseList(value)...)
	case "ignore-authors":
		c.IgnoredAuthors = append(c.IgnoredAuthors, parseList(value)...)
	case "ignore-rules":
		c.IgnoredRules = append(c.IgnoredRules, parseList(value)...)
	case "ignore-paths":
		c.IgnoredPaths = append(c.IgnoredPaths, parseList(value)...)
	case "max-file-size":
		if size, err := strconv.Atoi(value); err == nil {
			c.MaxFileSize = size
		} else {
			return fmt.Errorf("invalid max-file-size value: %s", value)
		}
	case "min-coverage-threshold":
		if threshold, err := strconv.ParseFloat(value, 64); err == nil {
			c.MinCoverageThreshold = threshold
		} else {
			return fmt.Errorf("invalid min-coverage-threshold value: %s", value)
		}
	case "max-concurrent-blame":
		if concurrent, err := strconv.Atoi(value); err == nil {
			c.MaxConcurrentBlame = concurrent
		} else {
			return fmt.Errorf("invalid max-concurrent-blame value: %s", value)
		}
	case "cache-results":
		c.CacheResults = strings.ToLower(value) == "true"
	case "enable-git-hooks":
		c.EnableGitHooks = strings.ToLower(value) == "true"
	case "custom-words":
		c.CustomWords = append(c.CustomWords, parseList(value)...)
	case "spellcheck-enabled":
		c.SpellCheckEnabled = strings.ToLower(value) == "true"
	case "spellcheck-extensions":
		c.SpellCheckExtensions = parseList(value)
	case "spellcheck-ignore-paths":
		c.SpellCheckIgnorePaths = parseList(value)
	case "ruff-enabled":
		c.RuffEnabled = strings.ToLower(value) == "true"
	case "ruff-rules":
		c.RuffRules = append(c.RuffRules, parseList(value)...)
	case "ruff-ignore-paths":
		c.RuffIgnorePaths = append(c.RuffIgnorePaths, parseList(value)...)
	default:
		c.CustomSettings[key] = value
	}
	return nil
}

func parseList(value string) []string {
	// Split by comma and clean up
	items := strings.Split(value, ",")
	var result []string

	for _, item := range items {
		cleaned := strings.TrimSpace(item)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}

	return result
}

func (c *Config) ShouldIgnoreFile(filePath string) bool {
	// Check if file is too large
	if c.MaxFileSize > 0 {
		if info, err := os.Stat(filePath); err == nil {
			if info.Size() > int64(c.MaxFileSize*1024) { // MaxFileSize is in KB
				return true
			}
		}
	}

	// Check against ignored file patterns
	for _, pattern := range c.IgnoredFiles {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		if strings.Contains(filePath, pattern) {
			return true
		}
		// Support glob patterns
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
	}

	// Check against ignored paths
	for _, pattern := range c.IgnoredPaths {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	return false
}

func (c *Config) ShouldIgnoreAuthor(email string, name string) bool {
	for _, ignored := range c.IgnoredAuthors {
		// Exact match
		if strings.EqualFold(ignored, email) || strings.EqualFold(ignored, name) {
			return true
		}
		// Partial match
		if strings.Contains(strings.ToLower(email), strings.ToLower(ignored)) {
			return true
		}
		if strings.Contains(strings.ToLower(name), strings.ToLower(ignored)) {
			return true
		}
	}
	return false
}

func (c *Config) ShouldIgnoreRule(ruleID string) bool {
	for _, ignored := range c.IgnoredRules {
		if ignored == ruleID {
			return true
		}
	}
	return false
}

func (c *Config) GetConcurrency() int {
	if c.MaxConcurrentBlame > 0 {
		return c.MaxConcurrentBlame
	}
	return 2 // Default
}

func GenerateConfigFile(filename string) error {
	if filename == "" {
		filename = ".codecompass.rc"
	}

	content := `# CodeCompass Configuration File 🧭
# Navigate your code quality with precision
# Lines starting with # are comments

# Ignore specific files (supports wildcards and globs)
ignore-files = "*.test.js,*.spec.js,*.d.ts,dist/*,build/*,node_modules/*"

# Ignore specific file paths
ignore-paths = "node_modules,coverage,.git,vendor,tmp,.next,out"

# Ignore specific authors (email or name patterns)
ignore-authors = "bot@company.com,dependabot,renovate,github-actions"

# Additional ESLint rules to ignore beyond command line
ignore-rules = "prefer-const,no-console"

# Maximum file size to analyze (in KB, 0 = no limit)
max-file-size = 5000

# Minimum coverage threshold for warnings (percentage)
min-coverage-threshold = 80

# Maximum concurrent git blame operations
max-concurrent-blame = 4

# Cache git blame results for better performance
cache-results = true

# Enable git hooks integration (experimental)
enable-git-hooks = false

# Custom project-specific settings
project-name = "My Project"
team-name = "Development Team"

# CodeCompass specific settings
show-compass-art = true
default-direction = "all"

# Spell check configuration
spellcheck-enabled = true
custom-words = "api,url,auth,oauth,async,await,json,xml,css,html,dom,ui,ux"
spellcheck-extensions = ".js,.ts,.jsx,.tsx,.md,.txt,.py,.java"
spellcheck-ignore-paths = "node_modules,dist,build,coverage"

# Ruff (Python Linter) configuration
ruff-enabled = true
ruff-rules = "E501,F401"
ruff-ignore-paths = "venv,.venv,migrations"

`

	return os.WriteFile(filename, []byte(content), 0644)
}

func (c *Config) PrintSummary() {
	fmt.Printf("🧭 Configuration Summary:\n")
	fmt.Printf("  • Ignored files: %d patterns\n", len(c.IgnoredFiles))
	fmt.Printf("  • Ignored authors: %d patterns\n", len(c.IgnoredAuthors))
	fmt.Printf("  • Ignored rules: %d rules\n", len(c.IgnoredRules))
	fmt.Printf("  • Max concurrent blame: %d\n", c.MaxConcurrentBlame)
	fmt.Printf("  • Cache results: %t\n", c.CacheResults)

	if c.MaxFileSize > 0 {
		fmt.Printf("  • Max file size: %d KB\n", c.MaxFileSize)
	}
}
