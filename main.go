package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codecompass/internal/analyzer"
	"codecompass/internal/config"
	"codecompass/internal/eslint"
	"codecompass/internal/git"
	"codecompass/internal/leaderboard"
	"codecompass/internal/types"
	"codecompass/internal/utils"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

const VERSION = "1.0.0"
const PROJECT_NAME = "CodeCompass"

// ASCII compass art
const COMPASS_ART = `
    🧭 CodeCompass 🧭
         N
         ↑
    W ←  +  → E
         ↓
         S

   Navigate Your Code Quality
`

const MINI_COMPASS = `🧭`

func main() {
	var (
		help     = flag.Bool("help", false, "Show help message")
		h        = flag.Bool("h", false, "Show help message (short)")
		version  = flag.Bool("version", false, "Show version information")
		v        = flag.Bool("v", false, "Show version information (short)")
		showLogo = flag.Bool("logo", false, "Show CodeCompass ASCII art")

		// Leaderboard flags
		showAuthors    = flag.Bool("authors", true, "Show author leaderboard")
		showFiles      = flag.Bool("files", true, "Show file leaderboard")
		showRules      = flag.Bool("rules", true, "Show rule leaderboard")
		showLoc        = flag.Bool("loc", true, "Show lines of code leaderboard")
		showCommits    = flag.Bool("commits", true, "Show commit count leaderboard")
		showRecent     = flag.Bool("recent", true, "Show recent contributors leaderboard")
		showCoverage   = flag.Bool("coverage", true, "Show code coverage leaderboard")
		showChurn      = flag.Bool("churn", false, "Show code churn leaderboard")
		showBugs       = flag.Bool("bugs", false, "Show bug density leaderboard")
		showDebt       = flag.Bool("debt", false, "Show technical debt leaderboard")
		showComplexity = flag.Bool("complexity", false, "Show code complexity leaderboard")
		showSummary    = flag.Bool("summary", true, "Show repository summary")

		// Configuration flags
		topN             = flag.Int("top", 15, "Number of entries to show in leaderboards")
		ignoredRulesFlag = flag.String("ignore", "sort-imports,import/order", "Comma-separated list of rules to ignore")
		coverageFile     = flag.String("coverage-file", "", "Path to coverage file (auto-detected if not specified)")
		configFile       = flag.String("config", "", "Path to configuration file")
		generateConfig   = flag.Bool("generate-config", false, "Generate a sample configuration file")
		showConfig       = flag.Bool("show-config", false, "Show current configuration and exit")

		// Advanced flags
		enableCache = flag.Bool("cache", true, "Enable caching for better performance")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		quiet       = flag.Bool("quiet", false, "Suppress non-essential output")
	)

	flag.Usage = showUsage
	flag.Parse()

	if *help || *h {
		showUsage()
		return
	}

	if *version || *v {
		showVersion()
		return
	}

	if *showLogo {
		showCompassArt()
		return
	}

	// Show mini compass at start unless quiet
	if !*quiet {
		fmt.Print(color.CyanString(COMPASS_ART))
	}

	if *generateConfig {
		filename := ".codecompass.rc"
		if err := config.GenerateConfigFile(filename); err != nil {
			log.Fatalf("Failed to generate config file: %v", err)
		}
		fmt.Printf("✅ Generated configuration file: %s\n", filename)
		return
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if *configFile != "" {
		cfg, err = config.LoadConfigFromFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config file %s: %v", *configFile, err)
		}
	} else {
		cfg, err = config.LoadConfig()
		if err != nil {
			if !*quiet {
				fmt.Printf("Warning: Failed to load config: %v\n", err)
			}
			cfg = config.NewConfig()
		}
	}

	if *showConfig {
		cfg.PrintSummary()
		return
	}

	// Apply config overrides
	if *enableCache {
		cfg.CacheResults = *enableCache
	}

	// Parse ignored rules from both config and command line
	ignoredRules := cfg.IgnoredRules
	if *ignoredRulesFlag != "" {
		cmdIgnoredRules := strings.Split(*ignoredRulesFlag, ",")
		for i, rule := range cmdIgnoredRules {
			cmdIgnoredRules[i] = strings.TrimSpace(rule)
		}
		ignoredRules = append(ignoredRules, cmdIgnoredRules...)
	}

	// Handle positional arguments (directory path)
	args := flag.Args()
	if len(args) > 0 {
		targetDir := args[0]
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			log.Fatalf("Failed to resolve path %s: %v", targetDir, err)
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			log.Fatalf("Directory does not exist: %s", absPath)
		}

		if err := os.Chdir(absPath); err != nil {
			log.Fatalf("Failed to change to directory %s: %v", absPath, err)
		}

		if !*quiet {
			fmt.Printf("%s Analyzing repository in: %s\n", MINI_COMPASS, absPath)
		}
	}

	// Validate git repository
	if err := git.ValidateRepository(); err != nil {
		log.Fatal("Not in a git repository. Please run from within a git repository or specify a valid git repository path.")
	}

	// Show configuration summary if verbose
	if *verbose && !*quiet {
		cfg.PrintSummary()
		fmt.Println()
	}

	// Get tracked files
	trackedFiles, err := git.GetTrackedFiles()
	if err != nil {
		log.Fatal("Failed to get tracked files:", err)
	}

	// Filter tracked files based on config
	filteredFiles := make(map[string]bool)
	for file := range trackedFiles {
		if !cfg.ShouldIgnoreFile(file) {
			filteredFiles[file] = true
		}
	}

	if *verbose && !*quiet {
		fmt.Printf("📁 Found %d tracked files (%d after filtering)\n", len(trackedFiles), len(filteredFiles))
	}

	// Check if ESLint-based leaderboards are needed
	needsESLint := *showAuthors || *showFiles || *showRules

	var issues []types.Issue
	var authorStats map[string]*types.AuthorStats
	var fileStats map[string]*types.FileStats
	var ruleStats map[string]*types.RuleStats
	var warningLogs []string

	if needsESLint {
		if !*quiet {
			fmt.Printf("%s %s\n", MINI_COMPASS, color.BlueString("Running ESLint analysis..."))
		}

		// Run ESLint
		issues, err = eslint.RunESLint(filteredFiles, ignoredRules)
		if err != nil {
			log.Fatal("Failed to run ESLint:", err)
		}

		if !*quiet {
			fmt.Printf("📊 %d lint issues collected.\n", len(issues))
			if len(ignoredRules) > 0 {
				fmt.Printf("🚫 Ignored ESLint rules: %s\n", strings.Join(ignoredRules, ", "))
			}
		}

		if len(issues) == 0 {
			if !*quiet {
				fmt.Printf("%s %s\n", MINI_COMPASS, color.GreenString("Clean codebase - no issues found!"))
			}
		} else {
			// Initialize data structures
			authorStats = make(map[string]*types.AuthorStats)
			fileStats = make(map[string]*types.FileStats)
			ruleStats = make(map[string]*types.RuleStats)
			var mu sync.Mutex

			// Process issues
			var bar *progressbar.ProgressBar
			if !*quiet {
				bar = progressbar.Default(int64(len(issues)))
			}

			semaphore := utils.NewSemaphore(cfg.GetConcurrency())
			analyzer := analyzer.New(semaphore, &mu)

			for _, issue := range issues {
				if err := analyzer.ProcessIssueWithConfig(issue, cfg, authorStats, fileStats, ruleStats, &warningLogs); err != nil {
					if *verbose {
						fmt.Printf("Warning: Failed to process issue in %s:%d: %v\n", issue.FilePath, issue.Line, err)
					}
					continue
				}

				if bar != nil {
					bar.Add(1)
				}

				// Small delay to prevent system overload
				time.Sleep(50 * time.Millisecond)
			}

			if bar != nil {
				bar.Finish()
			}
		}
	}

	// Navigation header for leaderboards
	if !*quiet {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, color.New(color.Bold).Sprint("Code Quality Navigation"))
		fmt.Printf("%s\n", strings.Repeat("─", 50))
	}

	// Generate leaderboards with compass directions
	if *showAuthors && len(issues) > 0 {
		fmt.Printf("\n🧭 %s", color.New(color.FgYellow).Sprint("North: "))
		leaderboard.GenerateAuthorLeaderboard(authorStats, *topN)
	}

	if *showFiles && len(issues) > 0 {
		fmt.Printf("\n🧭 %s", color.New(color.FgRed).Sprint("South: "))
		leaderboard.GenerateFileLeaderboard(fileStats, *topN)
	}

	if *showRules && len(issues) > 0 {
		fmt.Printf("\n🧭 %s", color.New(color.FgBlue).Sprint("East: "))
		leaderboard.GenerateRuleLeaderboard(ruleStats, *topN)
	}

	if *showLoc {
		fmt.Printf("\n🧭 %s", color.New(color.FgGreen).Sprint("West: "))
		leaderboard.GenerateLinesOfCodeLeaderboard(filteredFiles, *topN)
	}

	if *showCommits {
		fmt.Printf("\n🧭 %s", color.New(color.FgMagenta).Sprint("NE: "))
		leaderboard.GenerateCommitCountLeaderboard(*topN)
	}

	if *showRecent {
		fmt.Printf("\n🧭 %s", color.New(color.FgCyan).Sprint("NW: "))
		leaderboard.GenerateRecentContributorsLeaderboard(*topN)
	}

	if *showCoverage {
		fmt.Printf("\n🧭 %s", color.New(color.FgHiGreen).Sprint("SE: "))
		leaderboard.GenerateCodeCoverageLeaderboard(filteredFiles, *coverageFile, *topN)
	}

	if *showChurn {
		fmt.Printf("\n🧭 %s", color.New(color.FgHiYellow).Sprint("SW: "))
		if err := leaderboard.GenerateCodeChurnLeaderboard(filteredFiles, *topN); err != nil {
			fmt.Printf("❌ Failed to generate code churn leaderboard: %v\n", err)
		}
	}

	if *showBugs {
		fmt.Printf("\n🧭 %s", color.New(color.FgHiRed).Sprint("SSE: "))
		if err := leaderboard.GenerateBugDensityLeaderboard(filteredFiles, *topN); err != nil {
			fmt.Printf("❌ Failed to generate bug density leaderboard: %v\n", err)
		}
	}

	if *showDebt {
		fmt.Printf("\n🧭 %s", color.New(color.FgHiMagenta).Sprint("SSW: "))
		if err := leaderboard.GenerateTechnicalDebtLeaderboard(filteredFiles, *topN); err != nil {
			fmt.Printf("❌ Failed to generate technical debt leaderboard: %v\n", err)
		}
	}

	if *showComplexity {
		fmt.Printf("\n🧭 %s", color.New(color.FgHiBlue).Sprint("NNE: "))
		fmt.Printf("Code complexity leaderboard coming soon!\n")
	}

	if *showSummary {
		fmt.Printf("\n%s %s", MINI_COMPASS, color.New(color.Bold).Sprint("Center: "))
		leaderboard.GenerateSummaryStats(authorStats, fileStats, ruleStats)
	}

	// Print warnings
	if len(warningLogs) > 0 && !*quiet {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, color.YellowString("Navigation Warnings:"))
		for _, warn := range warningLogs {
			fmt.Printf("  %s\n", color.New(color.FgHiBlack).Sprint(warn))
		}
	}

	if *verbose {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, color.GreenString("Navigation completed successfully!"))
	}
}

func showCompassArt() {
	fmt.Print(color.CyanString(`
        🧭 CodeCompass 🧭
             
            ╭─────╮
         ╭──┤  N  ├──╮
      ╭──┤  ╰──┬──╯  ├──╮
   ╭──┤ NW   ╭─┼─╮   NE ├──╮
╭──┤ W  ├───┤ + ├───┤  E ├──╮
│  ╰──┤ SW   ╰─┼─╯   SE ├──╯│
│     ╰──┤  ╭──┴──╮  ├──╯  │
│        ╰──┤  S  ├──╯     │
│           ╰─────╯        │
╰─ Navigate Your Code ────╯
    Quality with Precision
          
    🎯 Find Issues
    📊 Track Metrics  
    🚀 Improve Quality
    📈 Monitor Progress
`))

	fmt.Println(color.New(color.Bold).Sprint("\nCodeCompass v" + VERSION))
	fmt.Println("Your comprehensive code quality navigation tool")
}

func showVersion() {
	fmt.Printf("%s %s v%s\n", MINI_COMPASS, PROJECT_NAME, VERSION)
	fmt.Printf("A comprehensive code quality navigation tool\n")
	fmt.Printf("Navigate your codebase with precision and insight\n")
	fmt.Printf("Built with Go - https://github.com/your-org/codecompass\n")
}

func showUsage() {
	fmt.Print(color.CyanString(COMPASS_ART))
	fmt.Printf("%s\n", color.New(color.Bold).Sprint("CodeCompass - Navigate Your Code Quality"))
	fmt.Printf("\n%s\n", color.BlueString("USAGE:"))
	fmt.Printf("  %s [OPTIONS] [DIRECTORY]\n\n", os.Args[0])

	fmt.Printf("%s\n", color.BlueString("ARGUMENTS:"))
	fmt.Printf("  DIRECTORY              Target git repository directory (default: current directory)\n\n")

	fmt.Printf("%s\n", color.BlueString("COMPASS DIRECTIONS (Leaderboards):"))
	fmt.Printf("  🧭 North    --authors              Author leaderboard (lint issue contributors)\n")
	fmt.Printf("  🧭 South    --files                File leaderboard (most problematic files)\n")
	fmt.Printf("  🧭 East     --rules                Rule leaderboard (most violated rules)\n")
	fmt.Printf("  🧭 West     --loc                  Lines of code leaderboard\n")
	fmt.Printf("  🧭 NE       --commits              Commit count leaderboard\n")
	fmt.Printf("  🧭 NW       --recent               Recent contributors leaderboard\n")
	fmt.Printf("  🧭 SE       --coverage             Code coverage leaderboard\n")
	fmt.Printf("  🧭 SW       --churn                Code churn leaderboard\n")
	fmt.Printf("  🧭 SSE      --bugs                 Bug density leaderboard\n")
	fmt.Printf("  🧭 SSW      --debt                 Technical debt leaderboard\n")
	fmt.Printf("  🧭 NNE      --complexity           Code complexity leaderboard\n")
	fmt.Printf("  🧭 Center   --summary              Repository summary\n\n")

	fmt.Printf("%s\n", color.BlueString("CONFIGURATION OPTIONS:"))
	fmt.Printf("  --config FILE          Path to configuration file (.codecompass.rc)\n")
	fmt.Printf("  --generate-config      Generate a sample configuration file\n")
	fmt.Printf("  --show-config          Show current configuration and exit\n")
	fmt.Printf("  --top N                Number of entries to show in leaderboards (default: 15)\n")
	fmt.Printf("  --ignore RULES         Comma-separated ESLint rules to ignore\n")
	fmt.Printf("  --coverage-file FILE   Path to coverage file (auto-detected if not specified)\n\n")

	fmt.Printf("%s\n", color.BlueString("DISPLAY OPTIONS:"))
	fmt.Printf("  --logo                 Show CodeCompass ASCII art\n")
	fmt.Printf("  --cache                Enable caching for better performance (default: true)\n")
	fmt.Printf("  --verbose              Enable verbose output\n")
	fmt.Printf("  --quiet                Suppress non-essential output\n\n")

	fmt.Printf("%s\n", color.BlueString("OTHER OPTIONS:"))
	fmt.Printf("  -h, --help             Show this help message\n")
	fmt.Printf("  -v, --version          Show version information\n\n")

	fmt.Printf("%s\n", color.BlueString("NAVIGATION EXAMPLES:"))
	fmt.Printf("  %s                                    # Full compass navigation\n", os.Args[0])
	fmt.Printf("  %s /path/to/repo                      # Navigate specific repository\n", os.Args[0])
	fmt.Printf("  %s --authors --files                  # North & South directions only\n", os.Args[0])
	fmt.Printf("  %s --loc --coverage                   # West & SE directions (no ESLint)\n", os.Args[0])
	fmt.Printf("  %s --churn --bugs --debt              # SW, SSE, SSW (advanced metrics)\n", os.Args[0])
	fmt.Printf("  %s --generate-config                  # Create .codecompass.rc file\n", os.Args[0])
	fmt.Printf("  %s --logo                             # Show compass art\n", os.Args[0])
	fmt.Printf("  %s --quiet --summary                  # Silent summary only\n\n", os.Args[0])

	fmt.Printf("%s\n", color.BlueString("COMPASS READINGS:"))
	fmt.Printf("  📊 Quality readings help you navigate code health\n")
	fmt.Printf("  🎯 Each direction provides different insights\n")
	fmt.Printf("  🚀 Use multiple directions for comprehensive analysis\n")
	fmt.Printf("  📈 Track changes over time to monitor progress\n\n")

	fmt.Printf("%s\n", color.BlueString("CONFIGURATION FILE:"))
	fmt.Printf("  CodeCompass looks for configuration files in this order:\n")
	fmt.Printf("  1. .codecompass.rc\n")
	fmt.Printf("  2. .codecompass.config\n")
	fmt.Printf("  3. codecompass.config\n\n")

	fmt.Printf("  Use --generate-config to create a sample configuration file.\n")
}
