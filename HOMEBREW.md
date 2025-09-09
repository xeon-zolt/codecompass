# 🧭 CodeCompass Homebrew Package

This guide covers installing and publishing CodeCompass via Homebrew.

## 📦 Installation

### Option 1: Install from Official Tap (Recommended)
Once published to the official tap:
```bash
brew tap xeon-zolt/codecompass
brew install codecompass
```

### Option 2: Install from URL (Development)
```bash
# Install directly from GitHub release
brew install https://github.com/xeon-zolt/codecompass/releases/download/v1.0.0/codecompass-v1.0.0.tar.gz
```

### Option 3: Install from Local Build
```bash
# Clone and build locally
git clone https://github.com/xeon-zolt/codecompass.git
cd codecompass
go build -o codecompass
sudo cp codecompass /usr/local/bin/
```

## 🚀 Post-Installation Setup

After installation, CodeCompass will show helpful setup information:

```
🧭 CodeCompass is now installed! 

Basic Usage:
  codecompass --help              # Show all available options
  codecompass --summary           # Quick repository overview
  codecompass --authors           # Author leaderboard
  
Advanced Features:
  codecompass --quality           # Comprehensive quality analysis
  codecompass --trends            # Trend analysis with charts
  codecompass --hotspots          # Detect high-risk code areas
  codecompass --team              # Team performance metrics
  
Configuration:
  Sample config: /opt/homebrew/etc/codecompass/codecompass.rc.example
  Copy to your repo: cp /opt/homebrew/etc/codecompass/codecompass.rc.example .codecompass.rc
  
Optional Dependencies:
  • Install ESLint for JavaScript analysis: npm install -g eslint
  • Install Ruff for Python analysis: pip install ruff
```

## 🔧 Configuration

1. **Copy sample configuration:**
   ```bash
   cp /opt/homebrew/etc/codecompass/codecompass.rc.example .codecompass.rc
   ```

2. **Generate new configuration:**
   ```bash
   codecompass --generate-config
   ```

3. **View current configuration:**
   ```bash
   codecompass --show-config
   ```

## 🧪 Verification

Test your installation:
```bash
# Check version
codecompass --version

# Run basic analysis
codecompass --summary

# Test advanced features
codecompass --quality
codecompass --trends
codecompass --hotspots
codecompass --team
```

## 📋 Dependencies

### Required
- **Git**: Repository analysis
- **Go**: Build dependency (handled by Homebrew)

### Optional (for enhanced functionality)
- **Node.js**: For ESLint JavaScript/TypeScript analysis
  ```bash
  brew install node
  npm install -g eslint
  ```
- **Python 3**: For Ruff Python analysis
  ```bash
  brew install python3
  pip3 install ruff
  ```

## 🚢 Publishing Guide (For Maintainers)

### Step 1: Create a GitHub Release
```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0 - Advanced Analytics & Charts"
git push origin v1.0.0

# Create release on GitHub with tarball
# GitHub will automatically create:
# https://github.com/xeon-zolt/codecompass/archive/refs/tags/v1.0.0.tar.gz
```

### Step 2: Calculate SHA256 Hash
```bash
curl -L https://github.com/xeon-zolt/codecompass/archive/refs/tags/v1.0.0.tar.gz | shasum -a 256
```

### Step 3: Update Formula
Update `Formula/codecompass.rb`:
```ruby
url "https://github.com/xeon-zolt/codecompass/archive/refs/tags/v1.0.0.tar.gz"
sha256 "ACTUAL_SHA256_HASH_HERE"
```

### Step 4: Create Homebrew Tap
```bash
# Create tap repository
brew tap-new xeon-zolt/codecompass

# Add formula to tap
cp Formula/codecompass.rb $(brew --repository)/Library/Taps/xeon-zolt/homebrew-codecompass/Formula/
```

### Step 5: Test Installation
```bash
# Test formula
brew audit --strict codecompass
brew test codecompass

# Install from tap
brew tap xeon-zolt/codecompass
brew install codecompass
```

### Step 6: Submit to Homebrew Core (Optional)
For inclusion in official Homebrew:
```bash
# Fork homebrew-core
git clone https://github.com/Homebrew/homebrew-core.git
cd homebrew-core

# Add formula
cp ../codecompass/Formula/codecompass.rb Formula/
git add Formula/codecompass.rb
git commit -m "codecompass: add formula"

# Create pull request
gh pr create --title "codecompass: add formula" --body "Comprehensive code quality analysis tool"
```

## 🔄 Update Process

### For Users
```bash
# Update Homebrew
brew update

# Upgrade CodeCompass
brew upgrade codecompass
```

### For Maintainers
1. Update version in `Formula/codecompass.rb`
2. Update SHA256 hash
3. Test locally
4. Push to tap repository
5. Users will get updates via `brew upgrade`

## 🐛 Troubleshooting

### Installation Issues
```bash
# Clean and reinstall
brew uninstall codecompass
brew cleanup
brew install codecompass

# Check dependencies
brew deps codecompass
brew missing codecompass
```

### Runtime Issues
```bash
# Check installation
which codecompass
codecompass --version

# Verify dependencies
git --version  # Should work
node --version # For ESLint support
python3 --version # For Ruff support
```

### Configuration Issues
```bash
# Reset configuration
rm .codecompass.rc
codecompass --generate-config

# Check configuration
codecompass --show-config
```

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/xeon-zolt/codecompass/issues)
- **Documentation**: [README.md](README.md)
- **Homebrew Tap**: [xeon-zolt/homebrew-codecompass](https://github.com/xeon-zolt/homebrew-codecompass)

---

🧭 **Navigate Your Code Quality with CodeCompass!**