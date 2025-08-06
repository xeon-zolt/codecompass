# CodeCompass 🧭

A comprehensive code quality navigation tool that helps you understand and improve your codebase. CodeCompass combines insights from various sources, including linting results and Git history, to provide detailed leaderboards and metrics.

## 🌟 Features

### 📊 Multiple Leaderboards (Compass Directions)
CodeCompass provides a multi-dimensional view of your codebase, represented by compass directions:

-   **🧭 North: Author Leaderboard** - Identifies authors who introduced the most lint issues.
-   **🧭 South: File Leaderboard** - Highlights files with the highest number of linting problems.
-   **🧭 East: Rule Leaderboard** - Shows the most frequently violated ESLint rules.
-   **🧭 West: Lines of Code (LOC) Leaderboard** - Ranks files by their lines of code.
-   **🧭 NE: Commit Activity Leaderboard** - Displays contributors with the most regular (non-merge) commits.
-   **🧭 NNE: Merge Commit Leaderboard** - Shows contributors with the most merge commits.
-   **🧭 NW: Recent Contributors Leaderboard** - Analyzes recent activity to identify active contributors.
-   **🧭 SE: Code Coverage Leaderboard** - Provides insights into code coverage statistics by file.
-   **🧭 SW: Code Churn Leaderboard** - Pinpoints files with the most frequent changes.
-   **🧭 SSE: Bug Density Leaderboard** - Identifies files with a high bug-fix ratio.
-   **🧭 SSW: Technical Debt Leaderboard** - Analyzes TODO/FIXME/HACK comments to quantify technical debt.
-   **🧭 NNW: Code Complexity Leaderboard** - (Coming Soon) Measures the complexity of your code.
-   **🧭 ENE: Spell Check Leaderboard** - Identifies files with spelling errors.
-   **🧭 Center: Repository Summary** - Provides an overall summary of your repository's health.

### ⚙️ Advanced Configuration
-   **Flexible Ignore Patterns** - Ignore files, authors, rules, and paths.
-   **Performance Tuning** - Configurable concurrency and caching.
-   **Multiple Config Formats** - Supports `.codecompass.rc`, `.codecompass.config`, and `codecompass.config`.
-   **Auto-detection** - Smart detection of coverage files and Git repositories.

### 🚀 Performance Optimized
-   **Conditional Execution** - Only runs ESLint when needed.
-   **Concurrent Processing** - Multi-threaded Git blame operations.
-   **Smart Caching** - Caches Git blame results for better performance.
-   **Memory Efficient** - Handles large repositories efficiently.

## 📦 Installation

### From Source
```bash
git clone https://github.com/xeoncross/codecompass
cd codecompass
go mod tidy
CGO_ENABLED=0 go build -o codecompass main.go
chmod +x codecompass
```

### Homebrew (Coming Soon)
```bash
brew tap xeoncross/tap
brew install codecompass
```

## 🚀 Usage

### Basic Usage

Analyze current directory:
```bash
./codecompass
```

Analyze specific repository:
```bash
./codecompass /path/to/your/project
```

Show only specific leaderboards:
```bash
./codecompass --authors --files --coverage
```

Show all leaderboards:
```bash
./codecompass --all
```

### Generate Configuration File

This creates a `.codecompass.rc` file with all available options:
```bash
./codecompass --generate-config
```

## 📋 Command Line Options

### Leaderboard Options

-   `--authors`: Show author leaderboard (lint issue contributors)
-   `--files`: Show file leaderboard (most problematic files)
-   `--rules`: Show rule leaderboard (most violated rules)
-   `--loc`: Show lines of code leaderboard
-   `--commits`: Show regular commit count leaderboard (non-merges)
-   `--merges`: Show merge commit count leaderboard
-   `--recent`: Show recent contributors leaderboard
-   `--coverage`: Show code coverage leaderboard
-   `--churn`: Show code churn leaderboard
-   `--bugs`: Show bug density leaderboard
-   `--debt`: Show technical debt leaderboard
-   `--complexity`: Show code complexity leaderboard (coming soon)
-   `--spellcheck`: Show spell check leaderboard
-   `--summary`: Show repository summary
-   `--all`: Show all leaderboards

### Configuration Options

-   `--config FILE`: Path to configuration file
-   `--generate-config`: Generate a sample configuration file
-   `--show-config`: Show current configuration and exit
-   `--top N`: Number of entries to show in leaderboards (default: 15)
-   `--ignore RULES`: Comma-separated ESLint rules to ignore
-   `--coverage-file FILE`: Path to coverage file (auto-detected)

### Display Options

-   `--logo`: Show CodeCompass ASCII art
-   `--cache`: Enable caching for better performance (default: true)
-   `--verbose`: Enable verbose output
-   `--quiet`: Suppress non-essential output

### Other Options

-   `-h, --help`: Show this help message
-   `-v, --version`: Show version information

## 📊 Example Output

### Author Leaderboard

```
🧭 North: Author Leaderboard - Most ESLint Issues:
    John Doe (john@company.com) – 45 issues (12 errors, 33 warnings), 8 files, top rule: no-unused-vars (15)
    Jane Smith (jane@company.com) – 32 issues (8 errors, 24 warnings), 12 files, top rule: prefer-const (12)
    Bob Wilson (bob@company.com) – 18 issues (3 errors, 15 warnings), 6 files, top rule: no-console (8)
```

### File Leaderboard

```
🧭 South: File Leaderboard - Most Problematic Files:
    src/utils/helper.js – 23 issues, 3 authors, top rule: complexity (8)
    src/components/Dashboard.tsx – 19 issues, 2 authors, top rule: react-hooks/exhaustive-deps (7)
    src/api/client.js – 15 issues, 4 authors, top rule: no-unused-vars (6)
```

### Code Coverage Leaderboard

```
🧭 SE: Code Coverage Leaderboard - Coverage by File:
    src/utils/validation.js – 45.2% (128/283 lines, 67% functions)
    src/components/Modal.tsx – 52.8% (95/180 lines, 71% functions)
    src/api/auth.js – 61.3% (76/124 lines, 80% functions)

🏆 Files with highest coverage:
src/utils/constants.js – 98.5%
src/types/index.ts – 95.2%
src/config/app.js – 92.1%

📊 Overall Coverage: 73.4% (2,847/3,879 lines covered)
```

## 🎯 Use Cases

### For Development Teams
-   **Code Review Focus** - Identify files and authors that need attention.
-   **Quality Metrics** - Track code quality trends over time.
-   **Onboarding** - Help new team members understand codebase patterns.

### For Project Managers
-   **Technical Debt** - Quantify and prioritize technical debt.
-   **Resource Allocation** - Identify areas that need more development time.
-   **Quality Trends** - Monitor code quality improvements over time.

### For Individual Developers
-   **Personal Metrics** - Track your own code quality contributions.
-   **Learning** - Identify which rules you violate most often.
-   **Improvement** - Focus on specific areas for skill development.

## 🔧 Advanced Usage

### Running Only Non-ESLint Leaderboards

Much faster - skips ESLint execution entirely:
```bash
./codecompass --no-authors --no-files --no-rules --loc --coverage --commits
```

### Custom Configuration

Use specific config file:
```bash
./codecompass --config ./my-project.config
```

Show top 25 entries:
```bash
./codecompass --top 25
```

Focus on specific metrics:
```bash
./codecompass --churn --bugs --debt --no-summary
```

### Integration with CI/CD

Generate report for CI:
```bash
./codecompass --quiet --output ./reports/quality-report.txt
```

Check if quality threshold is met:
```bash
./codecompass --format json --output quality.json
```

### Coverage Integration
The tool automatically detects coverage files in these locations:
-   `coverage/lcov.info`
-   `coverage/coverage.info`
-   `lcov.info`
-   `coverage-final.json`
-   `.nyc_output/coverage-final.json`

Or specify manually:
```bash
./codecompass --coverage-file ./custom/coverage.info
```

## 🛠️ Requirements

-   **Git** - Must be run within a Git repository.
-   **Node.js & npm/yarn** - Required for ESLint execution (if using ESLint-based leaderboards like Authors, Files, Rules).
-   **ESLint** - Must be configured in your project (if using ESLint-based leaderboards).

## 📝 Configuration File Reference

### File Locations (in order of precedence)
1.  `.codecompass.rc`
2.  `.codecompass.config`
3.  `codecompass.config`

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ignore-files` | string[] | `[]` | File patterns to ignore (supports globs) |
| `ignore-authors` | string[] | `[]` | Author patterns to ignore |
| `ignore-rules` | string[] | `[]` | ESLint rules to ignore |
| `ignore-paths` | string[] | `[]` | Path patterns to ignore |
| `max-file-size` | int | `5000` | Max file size in KB (0 = no limit) |
| `min-coverage-threshold` | float | `80.0` | Coverage threshold for warnings |
| `max-concurrent-blame` | int | `4` | Max concurrent Git operations |
| `cache-results` | bool | `true` | Enable result caching |

### Pattern Matching
-   **File patterns**: Support wildcards (`*.js`) and globs (`src/**/*.test.js`).
-   **Author patterns**: Match against name or email (case-insensitive).
-   **Path patterns**: Simple substring matching.

## 🤝 Contributing

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'Add some amazing feature'`).
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

-   ESLint team for the excellent linting tool.
-   Git team for the powerful version control system.
-   Go community for the fantastic tooling ecosystem.


**Made with ❤️ for developers who care about code quality**
