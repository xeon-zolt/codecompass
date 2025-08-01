# ESLint Blame Tool 🔍

A comprehensive code quality analysis tool that combines ESLint results with git blame information to provide detailed insights into your JavaScript/TypeScript codebase. Track who introduced lint issues, analyze code metrics, and identify areas for improvement.

## 🌟 Features

### 📊 Multiple Leaderboards
- **Author Leaderboard** - Who introduced the most lint issues
- **File Leaderboard** - Files with the most problems
- **Rule Leaderboard** - Most violated ESLint rules
- **Lines of Code** - Largest files in your codebase
- **Commit Activity** - Most active contributors
- **Recent Contributors** - Recent activity analysis
- **Code Coverage** - Coverage statistics by file
- **Code Churn** - Files with most frequent changes
- **Bug Density** - Files with highest bug-fix ratio
- **Technical Debt** - TODO/FIXME/HACK comments analysis

### ⚙️ Advanced Configuration
- **Flexible Ignore Patterns** - Ignore files, authors, rules, and paths
- **Performance Tuning** - Configurable concurrency and caching
- **Multiple Config Formats** - Support for various configuration files
- **Auto-detection** - Smart detection of coverage files and git repositories

### 🚀 Performance Optimized
- **Conditional Execution** - Only runs ESLint when needed
- **Concurrent Processing** - Multi-threaded git blame operations
- **Smart Caching** - Caches git blame results for better performance
- **Memory Efficient** - Handles large repositories efficiently

## 📦 Installation

### From Source
```
git clone https://github.com/your-org/eslint-blame
cd eslint-blame
go mod tidy
CGO_ENABLED=0 go build -o eslint-blame main.go
chmod +x eslint-blame
```

### Basic Usage

Analyze current directory

./eslint-blame
Analyze specific repository

./eslint-blame /path/to/your/project
Show only specific leaderboards

./eslint-blame --authors --files --coverage


### Generate Configuration File

This creates a `.eslint-blamerc` file with all available options:


ESLint Blame Configuration File
Lines starting with # are comments
Ignore specific files (supports wildcards and globs)

ignore-files = ".test.js,.spec.js,.d.ts,dist/,build/,node_modules/"
Ignore specific file paths

ignore-paths = "node_modules,coverage,.git,vendor,tmp,.next,out"
Ignore specific authors (email or name patterns)

ignore-authors = "bot@company.com,dependabot,renovate,github-actions"
Additional ESLint rules to ignore beyond command line

ignore-rules = "prefer-const,no-console"
Maximum file size to analyze (in KB, 0 = no limit)

max-file-size = 5000
Minimum coverage threshold for warnings (percentage)

min-coverage-threshold = 80
Maximum concurrent git blame operations

max-concurrent-blame = 4
Cache git blame results for better performance

cache-results = true
Enable git hooks integration (experimental)

enable-git-hooks = false

## 📋 Command Line Options

### Leaderboard Options

--authors Show author leaderboard (default: true)
--files Show file leaderboard (default: true)
--rules Show rule leaderboard (default: true)
--loc Show lines of code leaderboard (default: true)
--commits Show commit count leaderboard (default: true)
--recent Show recent contributors leaderboard (default: true)
--coverage Show code coverage leaderboard (default: true)
--churn Show code churn leaderboard (default: false)
--bugs Show bug density leaderboard (default: false)
--debt Show technical debt leaderboard (default: false)
--summary Show repository summary (default: true)

### Configuration Options

--config FILE Path to configuration file
--generate-config Generate a sample configuration file
--show-config Show current configuration and exit
--top N Number of entries to show (default: 15)
--ignore RULES Comma-separated ESLint rules to ignore
--coverage-file FILE Path to coverage file (auto-detected)


### Output Options

--format FORMAT Output format: console, json, csv (default: console)
--output FILE Output file (default: stdout)
--verbose Enable verbose output
--quiet Suppress non-essential output


### Performance Options

--cache Enable result caching (default: true)


## 📊 Example Output

### Author Leaderboard

🏆 Author Leaderboard - Most ESLint Issues:

    John Doe (john@company.com) – 45 issues (12 errors, 33 warnings), 8 files, top rule: no-unused-vars (15)

    Jane Smith (jane@company.com) – 32 issues (8 errors, 24 warnings), 12 files, top rule: prefer-const (12)

    Bob Wilson (bob@company.com) – 18 issues (3 errors, 15 warnings), 6 files, top rule: no-console (8)

File Leaderboard

📁 File Leaderboard - Most Problematic Files:

    src/utils/helper.js – 23 issues, 3 authors, top rule: complexity (8)

    src/components/Dashboard.tsx – 19 issues, 2 authors, top rule: react-hooks/exhaustive-deps (7)

    src/api/client.js – 15 issues, 4 authors, top rule: no-unused-vars (6)

### Code Coverage Leaderboard

📈 Code Coverage Leaderboard - Coverage by File:

    src/utils/validation.js – 45.2% (128/283 lines, 67% functions)

    src/components/Modal.tsx – 52.8% (95/180 lines, 71% functions)

    src/api/auth.js – 61.3% (76/124 lines, 80% functions)

🏆 Files with highest coverage:
src/utils/constants.js – 98.5%
src/types/index.ts – 95.2%
src/config/app.js – 92.1%

📊 Overall Coverage: 73.4% (2,847/3,879 lines covered)

## 🎯 Use Cases

### For Development Teams
- **Code Review Focus** - Identify files and authors that need attention
- **Quality Metrics** - Track code quality trends over time
- **Onboarding** - Help new team members understand codebase patterns

### For Project Managers
- **Technical Debt** - Quantify and prioritize technical debt
- **Resource Allocation** - Identify areas that need more development time
- **Quality Trends** - Monitor code quality improvements over time

### For Individual Developers
- **Personal Metrics** - Track your own code quality contributions
- **Learning** - Identify which rules you violate most often
- **Improvement** - Focus on specific areas for skill development

## 🔧 Advanced Usage

### Running Only Non-ESLint Leaderboards

Much faster - skips ESLint execution entirely

./eslint-blame --no-authors --no-files --no-rules --loc --coverage --commits

## Custom Configuration

Use specific config file

./eslint-blame --config ./my-project.config
Show top 25 entries

./eslint-blame --top 25
Focus on specific metrics

./eslint-blame --churn --bugs --debt --no-summary


### Integration with CI/CD

Generate report for CI

./eslint-blame --quiet --output ./reports/quality-report.txt
Check if quality threshold is met

./eslint-blame --format json --output quality.json


### Coverage Integration
The tool automatically detects coverage files in these locations:
- `coverage/lcov.info`
- `coverage/coverage.info`
- `lcov.info`
- `coverage-final.json`
- `.nyc_output/coverage-final.json`

Or specify manually:

./eslint-blame --coverage-file ./custom/coverage.info

## 🛠️ Requirements

- **Git** - Must be run within a git repository
- **Node.js & npm/yarn** - For ESLint execution (if using ESLint leaderboards)
- **ESLint** - Configured in your project (if using ESLint leaderboards)

## 📝 Configuration File Reference

### File Locations (in order of precedence)
1. `.eslint-blamerc`
2. `.eslintblamerc`
3. `eslint-blame.config`
4. `.eslint-blame.config`

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ignore-files` | string[] | `[]` | File patterns to ignore (supports globs) |
| `ignore-authors` | string[] | `[]` | Author patterns to ignore |
| `ignore-rules` | string[] | `[]` | ESLint rules to ignore |
| `ignore-paths` | string[] | `[]` | Path patterns to ignore |
| `max-file-size` | int | `5000` | Max file size in KB (0 = no limit) |
| `min-coverage-threshold` | float | `80.0` | Coverage threshold for warnings |
| `max-concurrent-blame` | int | `4` | Max concurrent git operations |
| `cache-results` | bool | `true` | Enable result caching |
| `enable-git-hooks` | bool | `false` | Enable git hooks integration |

### Pattern Matching
- **File patterns**: Support wildcards (`*.js`) and globs (`src/**/*.test.js`)
- **Author patterns**: Match against name or email (case-insensitive)
- **Path patterns**: Simple substring matching

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- ESLint team for the excellent linting tool
- Git team for the powerful version control system
- Go community for the fantastic tooling ecosystem


**Made with ❤️ for developers who care about code quality**
