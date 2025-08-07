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
	"codecompass/internal/history"
	"codecompass/internal/leaderboard"
	"codecompass/internal/types"
	"codecompass/internal/utils"

	"codecompass/internal/ruff"

	"github.com/charmbracelet/lipgloss"
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

const MINI_COMPASS = "🧭"

var (
	// Styles for various elements
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#878787"))

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	compassArtStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF"))

	logoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	versionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	usageHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0000FF")).
		Bold(true)

	leaderboardTitleStyle = lipgloss.NewStyle().
		Bold(true)
)

func main() {
	var (
		help     = flag.Bool("help", false, "Show help message")
		h        = flag.Bool("h", false, "Show help message (short)")
		version  = flag.Bool("version", false, "Show version information")
		v        = flag.Bool("v", false, "Show version information (short)")
		showLogo = flag.Bool("logo", false, "Show CodeCompass ASCII art")

		// Leaderboard flags (default to false to be opt-in)
		showAuthors    = flag.Bool("authors", false, "Show author leaderboard (lint issue contributors)")
		showFiles      = flag.Bool("files", false, "Show file leaderboard (most problematic files)")
		showRules      = flag.Bool("rules", false, "Show rule leaderboard (most violated rules)")
		showLoc        = flag.Bool("loc", false, "Show lines of code leaderboard")
		showCommits    = flag.Bool("commits", false, "Show regular commit count leaderboard (non-merges)")
		showMerges     = flag.Bool("merges", false, "Show merge commit count leaderboard")
		showRecent     = flag.Bool("recent", false, "Show recent contributors leaderboard")
		showCoverage   = flag.Bool("coverage", false, "Show code coverage leaderboard")
		showChurn      = flag.Bool("churn", false, "Show code churn leaderboard")
		showBugs       = flag.Bool("bugs", false, "Show bug density leaderboard")
		showDebt       = flag.Bool("debt", false, "Show technical debt leaderboard")
		showComplexity = flag.Bool("complexity", false, "Show code complexity leaderboard")
		showSummary    = flag.Bool("summary", false, "Show repository summary")
		showSpellCheck = flag.Bool("spellcheck", false, "Show spell check leaderboard")
		showRuff       = flag.Bool("ruff", false, "Show Ruff (Python) leaderboard")

		showAll = flag.Bool("all", false, "Show all leaderboards")

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

		// History logging flags
		logHistory = flag.Bool("log-history", false, "Enable logging of leaderboard data to CSV files")
		logDir     = flag.String("log-dir", ".codecompass/history", "Directory to save leaderboard CSV logs")
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

	if *generateConfig {
		filename := ".codecompass.rc"
		if err := config.GenerateConfigFile(filename); err != nil {
			log.Fatalf("Failed to generate config file: %v", err)
		}
		fmt.Printf("✅ Generated configuration file: %s\n", filename)
		return
	}

	if *showAll {
		*showAuthors = true
		*showFiles = true
		*showRules = true
		*showLoc = true
		*showCommits = true
		*showMerges = true
		*showRecent = true
		*showCoverage = true
		*showChurn = true
		*showBugs = true
		*showDebt = true
		*showComplexity = true
		*showSummary = true
		*showSpellCheck = true
		*showRuff = true
	}

	// Check if any action was requested by the user.
	actionRequested := *showAuthors || *showFiles || *showRules || *showLoc ||
		*showCommits || *showMerges || *showRecent || *showCoverage || *showChurn ||
		*showBugs || *showDebt || *showComplexity || *showSummary || *showSpellCheck || *showRuff ||
		*showConfig

	// Check if Ruff-based leaderboards are needed
	needsRuff := *showRuff

	// If no action is specified, show usage information and exit.
	if !actionRequested && len(flag.Args()) == 0 {
		showUsage()
		return
	}

	if !*quiet {
		fmt.Print(compassArtStyle.Render(COMPASS_ART))
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
				fmt.Printf("Warning: Failed to load config: %s\n", warningStyle.Render(err.Error()))
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
	var ruffIssues []types.Issue
	var authorStats map[string]*types.AuthorStats
	var fileStats map[string]*types.FileStats
	var ruleStats map[string]*types.RuleStats
	var warningLogs []string

	if needsESLint {
		if !*quiet {
			fmt.Printf("%s %s\n", MINI_COMPASS, lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")).Render("Running ESLint analysis..."))
		}

		// Run ESLint
		eslintIssues, err := eslint.RunESLint(filteredFiles, ignoredRules)
		if err != nil {
			log.Fatal("Failed to run ESLint:", err)
		}
		issues = append(issues, eslintIssues...)

		if !*quiet {
			fmt.Printf("📊 %d lint issues collected.\n", len(eslintIssues))
			if len(ignoredRules) > 0 {
				fmt.Printf("🚫 Ignored ESLint rules: %s\n", strings.Join(ignoredRules, ", "))
			}
		}
	}

	if needsRuff {
		if !*quiet {
			fmt.Printf("%s %s\n", MINI_COMPASS, lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")).Render("Running Ruff analysis..."))
		}

		pythonFiles := []string{}
		for file := range filteredFiles {
			if strings.HasSuffix(file, ".py") {
				pythonFiles = append(pythonFiles, file)
			}
		}

		currentRuffIssues, err := ruff.RunRuff(pythonFiles, cfg.RuffRules, cfg.RuffIgnorePaths)
		if err != nil {
			log.Fatal("Failed to run Ruff:", err)
		}
		ruffIssues = append(ruffIssues, currentRuffIssues...)
		issues = append(issues, currentRuffIssues...)

		if !*quiet {
			fmt.Printf("📊 %d Ruff issues collected.\n", len(currentRuffIssues))
			if len(cfg.RuffRules) > 0 {
				fmt.Printf("🚫 Ruff rules: %s\n", strings.Join(cfg.RuffRules, ", "))
			}
			if len(cfg.RuffIgnorePaths) > 0 {
				fmt.Printf("🚫 Ignored Ruff paths: %s\n", strings.Join(cfg.RuffIgnorePaths, ", "))
			}
		}
	}

	if len(issues) == 0 {
		if !*quiet {
			fmt.Printf("%s %s\n", MINI_COMPASS, successStyle.Render("Clean codebase - no issues found!"))
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

	if !*quiet {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, leaderboardTitleStyle.Render("Code Quality Navigation"))
		fmt.Printf("%s\n", strings.Repeat("─", 50))
	}

	// Generate leaderboards with compass directions
	if *showAuthors && len(issues) > 0 {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#FFFF00")).Render("North: "))
		if needsESLint {
			authorEntries := leaderboard.GenerateAuthorLeaderboard(authorStats, *topN)
			leaderboard.PrintAuthorLeaderboard(authorEntries, *topN)
			if *logHistory {
				if err := history.WriteAuthorLeaderboardCSV(*logDir, authorEntries); err != nil {
					fmt.Printf("❌ Failed to log author leaderboard: %s\n", errorStyle.Render(err.Error()))
				} else if !*quiet {
					fmt.Printf("✅ Author leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		} else {
			fmt.Println("Author leaderboard requires ESLint analysis. Run with --authors flag.")
		}
	}

	if *showFiles {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#FF0000")).Render("South: "))
		if needsESLint {
			fileEntries := leaderboard.GenerateFileLeaderboard(fileStats, *topN)
			leaderboard.PrintFileLeaderboard(fileEntries, *topN)
			if *logHistory {
				if err := history.WriteFileLeaderboardCSV(*logDir, fileEntries); err != nil {
					fmt.Printf("❌ Failed to log file leaderboard: %s\n", errorStyle.Render(err.Error()))
				} else if !*quiet {
					fmt.Printf("✅ File leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		} else {
			fmt.Println("File leaderboard requires ESLint analysis. Run with --files flag.")
		}
	}

	if *showRules {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#0000FF")).Render("East: "))
		if needsESLint {
			ruleEntries := leaderboard.GenerateRuleLeaderboard(ruleStats, *topN)
			leaderboard.PrintRuleLeaderboard(ruleEntries, *topN)
			if *logHistory {
				if err := history.WriteRuleLeaderboardCSV(*logDir, ruleEntries); err != nil {
					fmt.Printf("❌ Failed to log rule leaderboard: %s\n", errorStyle.Render(err.Error()))
				} else if !*quiet {
					fmt.Printf("✅ Rule leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		} else {
			fmt.Println("Rule leaderboard requires ESLint analysis. Run with --rules flag.")
		}
	}

	if *showLoc {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#00FF00")).Render("West: "))
		locEntries := leaderboard.GenerateLinesOfCodeLeaderboard(filteredFiles, *topN)
		leaderboard.PrintLinesOfCodeLeaderboard(locEntries, *topN)
		if *logHistory {
			if err := history.WriteLinesOfCodeLeaderboardCSV(*logDir, locEntries); err != nil {
				fmt.Printf("❌ Failed to log lines of code leaderboard: %s\n", errorStyle.Render(err.Error()))
			} else if !*quiet {
					fmt.Printf("✅ Lines of code leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}

	if *showCommits {
		fmt.Printf("\n%s %s", MINI_COMPASS, leaderboardTitleStyle.Foreground(lipgloss.Color("#FF00FF")).Render("NE: "))
		commitEntries, err := leaderboard.GenerateCommitCountLeaderboard(*topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate commit count leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintCommitCountLeaderboard(commitEntries, *topN)
			if *logHistory {
				if err := history.WriteCommitCountLeaderboardCSV(*logDir, commitEntries); err != nil {
					fmt.Printf("❌ Failed to log commit count leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Commit count leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showRecent {
		fmt.Printf("\n%s %s", MINI_COMPASS, leaderboardTitleStyle.Foreground(lipgloss.Color("#00FFFF")).Render("NW: "))
		recentEntries, err := leaderboard.GenerateRecentContributorsLeaderboard(*topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate recent contributors leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintRecentContributorsLeaderboard(recentEntries, *topN)
			if *logHistory {
				if err := history.WriteRecentContributorsLeaderboardCSV(*logDir, recentEntries); err != nil {
					fmt.Printf("❌ Failed to log recent contributors leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Recent contributors leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showCoverage {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#00FF00")).Render("SE: "))
		coverageEntries, overallCoverage := leaderboard.GenerateCodeCoverageLeaderboard(filteredFiles, *coverageFile, *topN)
		leaderboard.PrintCodeCoverageLeaderboard(coverageEntries, overallCoverage, *topN)
		if *logHistory {
			if err := history.WriteCodeCoverageLeaderboardCSV(*logDir, coverageEntries); err != nil {
				fmt.Printf("❌ Failed to log code coverage leaderboard: %s\n", errorStyle.Render(err.Error()))
			} else if !*quiet {
					fmt.Printf("✅ Code coverage leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}

	if *showChurn {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#FFFF00")).Render("SW: "))
		churnEntries, err := leaderboard.GenerateCodeChurnLeaderboard(filteredFiles, *topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate code churn leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintCodeChurnLeaderboard(churnEntries, *topN)
			if *logHistory {
				if err := history.WriteCodeChurnLeaderboardCSV(*logDir, churnEntries); err != nil {
					fmt.Printf("❌ Failed to log code churn leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Code churn leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showBugs {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#FF0000")).Render("SSE: "))
		bugEntries, err := leaderboard.GenerateBugDensityLeaderboard(filteredFiles, *topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate bug density leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintBugDensityLeaderboard(bugEntries, *topN)
			if *logHistory {
				if err := history.WriteBugDensityLeaderboardCSV(*logDir, bugEntries); err != nil {
					fmt.Printf("❌ Failed to log bug density leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Bug density leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showDebt {
		fmt.Printf("\n%s %s", MINI_COMPASS, leaderboardTitleStyle.Render("SSW: "))
		debtEntries, err := leaderboard.GenerateTechnicalDebtLeaderboard(filteredFiles, *topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate technical debt leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintTechnicalDebtLeaderboard(debtEntries, *topN)
			if *logHistory {
				if err := history.WriteTechnicalDebtLeaderboardCSV(*logDir, debtEntries); err != nil {
					fmt.Printf("❌ Failed to log technical debt leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Technical debt leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showComplexity {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#0000FF")).Render("NNW: "))
		fmt.Printf("Code complexity leaderboard coming soon!\n")
	}

	if *showSpellCheck {
		fmt.Printf("\n🧭 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#00FFFF")).Render("ENE: "))
		spellEntries, spellAuthorStats, err := leaderboard.GenerateSpellCheckLeaderboard(filteredFiles, cfg, *topN)
		if err != nil {
			fmt.Printf("❌ Failed to generate spell check leaderboard: %s\n", errorStyle.Render(err.Error()))
		} else {
			leaderboard.PrintSpellCheckLeaderboard(spellEntries, spellAuthorStats, *topN)
			if *logHistory {
				if err := history.WriteSpellCheckLeaderboardCSV(*logDir, spellEntries); err != nil {
					fmt.Printf("❌ Failed to log spell check leaderboard: %v\n", err)
				} else if !*quiet {
					fmt.Printf("✅ Spell check leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		}
	}

	if *showRuff {
		fmt.Printf("\n\xe2\x90\x80 %s", leaderboardTitleStyle.Foreground(lipgloss.Color("#FFA500")).Render("WNW: "))
		if len(ruffIssues) > 0 {
			ruffRuleEntries := leaderboard.GenerateRuleLeaderboard(ruleStats, *topN)
			leaderboard.PrintRuleLeaderboard(ruffRuleEntries, *topN)
			if *logHistory {
				if err := history.WriteRuleLeaderboardCSV(*logDir, ruffRuleEntries); err != nil {
					fmt.Printf("❌ Failed to log Ruff rule leaderboard: %s\n", errorStyle.Render(err.Error()))
				} else if !*quiet {
					fmt.Printf("✅ Ruff rule leaderboard logged to %s\n", successStyle.Render(*logDir))
				}
			}
		} else {
			fmt.Println("No Ruff issues found.")
		}
	}

	if *showSummary {
		fmt.Printf("\n%s %s", MINI_COMPASS, leaderboardTitleStyle.Render("Center: "))
		leaderboard.GenerateSummaryStats(authorStats, fileStats, ruleStats)
	}

	if len(warningLogs) > 0 && !*quiet {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, warningStyle.Render("Navigation Warnings:"))
		for _, warn := range warningLogs {
			fmt.Printf("  %s\n", infoStyle.Render(warn))
		}
	}

	if *verbose {
		fmt.Printf("\n%s %s\n", MINI_COMPASS, successStyle.Render("Navigation completed successfully!"))
	}
}

func showCompassArt() {
	fmt.Print(compassArtStyle.Render(`
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

	fmt.Println(logoStyle.Render("\nCodeCompass v" + VERSION))
	fmt.Println(infoStyle.Render("Your comprehensive code quality navigation tool"))
	fmt.Println(infoStyle.Render("Navigate your codebase with precision and insight"))
}

func showVersion() {
	fmt.Printf("%s %s v%s\n", MINI_COMPASS, PROJECT_NAME, versionStyle.Render(VERSION))
	fmt.Println(infoStyle.Render("A comprehensive code quality navigation tool"))
	fmt.Println(infoStyle.Render("Navigate your codebase with precision and insight"))
	fmt.Println(infoStyle.Render("Built with Go - https://github.com/your-org/codecompass"))
}

func showUsage() {
	fmt.Print(compassArtStyle.Render(COMPASS_ART))
	fmt.Println(leaderboardTitleStyle.Render("CodeCompass - Navigate Your Code Quality"))
	fmt.Println(usageHeaderStyle.Render("\nUSAGE:"))
	fmt.Printf("  %s [OPTIONS] [DIRECTORY]\n\n", os.Args[0])

	fmt.Println(usageHeaderStyle.Render("ARGUMENTS:"))
	fmt.Println(infoStyle.Render("  DIRECTORY              Target git repository directory (default: current directory)\n"))

	fmt.Println(usageHeaderStyle.Render("COMPASS DIRECTIONS (Leaderboards):"))
	fmt.Printf("  %s North    --authors              Author leaderboard (lint issue contributors)\n", MINI_COMPASS)
	fmt.Printf("  %s South    --files                File leaderboard (most problematic files)\n", MINI_COMPASS)
	fmt.Printf("  %s East     --rules                Rule leaderboard (most violated rules)\n", MINI_COMPASS)
	fmt.Printf("  %s West     --loc                  Lines of code leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s NE       --commits              Regular commit count leaderboard (non-merges)\n", MINI_COMPASS)
	fmt.Printf("  %s NNE      --merges               Merge commit count leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s NW       --recent               Recent contributors leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s SE       --coverage             Code coverage leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s SW       --churn                Code churn leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s SSE      --bugs                 Bug density leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s SSW      --debt                 Technical debt leaderboard\n", MINI_COMPASS)
	fmt.Printf("  %s NNW      --complexity           Code complexity leaderboard (coming soon)\n", MINI_COMPASS)
	fmt.Printf("  %s ENE      --spellcheck           Spell check leaderboard\n", MINI_COMPASS)
		fmt.Printf("  %s WNW      --ruff                   Ruff (Python) leaderboard\n", MINI_COMPASS)
		fmt.Printf("  %s Center   --summary              Repository summary\n\n", MINI_COMPASS)

	fmt.Println(usageHeaderStyle.Render("CONFIGURATION OPTIONS:"))
	fmt.Println(infoStyle.Render("  --config FILE          Path to configuration file (.codecompass.rc)"))
	fmt.Println(infoStyle.Render("  --generate-config      Generate a sample configuration file"))
	fmt.Println(infoStyle.Render("  --show-config          Show current configuration and exit"))
	fmt.Println(infoStyle.Render("  --top N                Number of entries to show in leaderboards (default: 15)"))
	fmt.Println(infoStyle.Render("  --ignore RULES         Comma-separated ESLint rules to ignore"))
	fmt.Println(infoStyle.Render("  --coverage-file FILE   Path to coverage file (auto-detected if not specified)\n"))

	fmt.Println(usageHeaderStyle.Render("DISPLAY OPTIONS:"))
	fmt.Println(infoStyle.Render("  --logo                 Show CodeCompass ASCII art"))
	fmt.Println(infoStyle.Render("  --cache                Enable caching for better performance (default: true)"))
	fmt.Println(infoStyle.Render("  --verbose              Enable verbose output"))
	fmt.Println(infoStyle.Render("  --quiet                Suppress non-essential output\n"))

	fmt.Println(usageHeaderStyle.Render("HISTORY LOGGING OPTIONS:"))
	fmt.Println(infoStyle.Render("  --log-history          Enable logging of leaderboard data to CSV files"))
	fmt.Println(infoStyle.Render("  --log-dir DIR          Directory to save leaderboard CSV logs (default: .codecompass/history)\n"))

	fmt.Println(usageHeaderStyle.Render("OTHER OPTIONS:"))
	fmt.Println(infoStyle.Render("  -h, --help             Show this help message"))
	fmt.Println(infoStyle.Render("  -v, --version          Show version information\n"))

	fmt.Println(usageHeaderStyle.Render("NAVIGATION EXAMPLES:"))
	fmt.Printf("  %s --all                              # Full compass navigation (all leaderboards)\n", os.Args[0])
	fmt.Printf("  %s                                     # Show help message\n", os.Args[0])
	fmt.Printf("  %s /path/to/repo --commits --merges   # Navigate specific repository and compare commits\n", os.Args[0])
	fmt.Printf("  %s --authors --files                  # North & South directions only\n", os.Args[0])
	fmt.Printf("  %s --loc --coverage                   # West & SE directions (no ESLint)\n", os.Args[0])
	fmt.Printf("  %s --generate-config                  # Create .codecompass.rc file\n\n", os.Args[0])

	fmt.Println(usageHeaderStyle.Render("CONFIGURATION FILE:"))
	fmt.Println(infoStyle.Render("  CodeCompass looks for configuration files in this order:"))
	fmt.Println(infoStyle.Render("  1. .codecompass.rc"))
	fmt.Println(infoStyle.Render("  2. .codecompass.config"))
	fmt.Println(infoStyle.Render("  3. codecompass.config\n"))

	fmt.Println(infoStyle.Render("  Use --generate-config to create a sample configuration file."))
}



