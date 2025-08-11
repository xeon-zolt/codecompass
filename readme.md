# CodeCompass 🧭

[![Go Report Card](https://goreportcard.com/badge/github.com/xeon-zolt/codecompass)](https://goreportcard.com/report/github.com/xeon-zolt/codecompass) [![Go](https://github.com/xeon-zolt/codecompass/actions/workflows/release.yml/badge.svg)](https://github.com/xeon-zolt/codecompass/actions/workflows/release.yml)

A comprehensive code quality navigation tool that helps you understand and improve your codebase. CodeCompass combines insights from various sources, including linting results and Git history, to provide detailed leaderboards and metrics.

## Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quickstart](#-quickstart)
- [Usage](#-usage)
- [Command Line Options](#-command-line-options)
- [Configuration](#-configuration)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

## 🌟 Features

CodeCompass provides a multi-dimensional view of your codebase, represented by compass directions:

- **🧭 North**: Identifies authors who introduced the most lint issues.
- **🧭 South**: Highlights files with the highest number of linting problems.
- **🧭 East**: Shows the most frequently violated ESLint rules.
- **🧭 West**: Ranks files by their lines of code.
- **🧭 And more**: Provides leaderboards for commit activity, code coverage, code churn, bug density, technical debt, and spell-checking.

## 📦 Installation

### Homebrew (Recommended)

```bash
brew tap xeon-zolt/tap
brew install codecompass
```

The Homebrew formula is available at [Formula/codecompass.rb](Formula/codecompass.rb) and uses the latest release [v0.0.1-beta.3](https://github.com/xeon-zolt/codecompass/archive/refs/tags/v0.0.1-beta.3.tar.gz).

### From Source

```bash
git clone https://github.com/xeon-zolt/codecompass.git
cd codecompass
go mod tidy
CGO_ENABLED=0 go build -o codecompass main.go
chmod +x codecompass
```

## 🚀 Quickstart

### Install via Homebrew

```bash
brew tap xeon-zolt/tap
brew install codecompass
```

### Install from Source

```bash
git clone https://github.com/xeon-zolt/codecompass.git
cd codecompass
go mod tidy
CGO_ENABLED=0 go build -o codecompass main.go
chmod +x codecompass
```

## 🚀 Usage

### Basic Usage

Analyze current directory:
```bash
./codecompass --all
```

Analyze a specific repository:
```bash
./codecompass /path/to/your/project --all
```

Show only specific leaderboards:
```bash
./codecompass --authors --files --coverage
```

### Generate Configuration File

This creates a `.codecompass.rc` file with all available options:

```bash
./codecompass --generate-config
```

## 📋 Command Line Options

| Option | Description |
|--------|-------------|
| `--authors` | Show author leaderboard (lint issue contributors) |
| `--files` | Show file leaderboard (most problematic files) |
| `--rules` | Show rule leaderboard (most violated rules) |
| `--loc` | Show lines of code leaderboard |
| `--commits` | Show regular commit count leaderboard (non-merges) |
| `--merges` | Show merge commit count leaderboard |
| `--recent` | Show recent contributors leaderboard |
| `--coverage` | Show code coverage leaderboard |
| `--churn` | Show code churn leaderboard |
| `--bugs` | Show bug density leaderboard |
| `--debt` | Show technical debt leaderboard |
| `--spellcheck` | Show spell check leaderboard |
| `--summary` | Show repository summary |
| `--all` | Show all leaderboards |

For a full list of options, run `./codecompass --help`.

## ⚙️ Configuration

CodeCompass can be configured via a `.codecompass.rc` file. To generate a sample configuration file, run:

```bash
./codecompass --generate-config
```

The configuration file allows you to ignore files, authors, rules, and paths, as well as set performance-related options.

## 🛠️ Development

To run the tests, use the following command:

```bash
go test ./...
```

## 🤝 Contributing

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'Add some amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
