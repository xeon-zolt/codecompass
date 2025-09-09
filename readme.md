# CodeCompass 🧭

[![Go Report Card](https://goreportcard.com/badge/github.com/xeon/codecompass)](https://goreportcard.com/report/github.com/xeoncross/codecompass)
[![Go](https://github.com/xeoncross/codecompass/actions/workflows/release.yml/badge.svg)](https://github.com/xeoncross/codecompass/actions/workflows/release.yml)

A comprehensive code quality navigation tool that helps you understand and improve your codebase. CodeCompass combines insights from various sources, including linting results and Git history, to provide detailed leaderboards and metrics.

## Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
- [Command Line Options](#-command-line-options)
- [Configuration](#-configuration)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

## 🌟 Features

CodeCompass provides a multi-dimensional view of your codebase, represented by compass directions:

-   **🧭 North**: Identifies authors who introduced the most lint issues.
-   **🧭 South**: Highlights files with the highest number of linting problems.
-   **🧭 East**: Shows the most frequently violated ESLint rules.
-   **🧭 West**: Ranks files by their lines of code.
-   **🧭 And more**: Provides leaderboards for commit activity, code coverage, code churn, bug density, technical debt, and spell-checking.

## 📦 Installation

### 🍺 Homebrew (Recommended - macOS)

```bash
# Install from tap (coming soon)
brew tap xeon-zolt/codecompass
brew install codecompass

# Or install directly (for now)
./install.sh
```

### 🔧 Quick Install Script (All Platforms)

```bash
# Download and run installer
curl -sSL https://raw.githubusercontent.com/xeon-zolt/codecompass/main/install.sh | bash

# Or clone and run
git clone https://github.com/xeon-zolt/codecompass.git
cd codecompass
./install.sh
```

### 🏗️ From Source (Manual)

```bash
git clone https://github.com/xeon-zolt/codecompass.git
cd codecompass
go mod tidy
go build -ldflags "-s -w -X main.version=v1.0.0" -o codecompass
sudo cp codecompass /usr/local/bin/
```

### 📋 Post-Installation

After installation, set up your configuration:
```bash
# Generate configuration file
codecompass --generate-config

# View available options  
codecompass --help

# Test installation
codecompass --version
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
| --- | --- |
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

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'Add some amazing feature'`).
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.