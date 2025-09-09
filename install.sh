#!/bin/bash
set -e

# 🧭 CodeCompass Installation Script
# Navigate Your Code Quality

VERSION="v1.0.0"
BINARY_NAME="codecompass"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/codecompass"

echo "🧭 CodeCompass Installation Script"
echo "=================================="

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux|darwin) ;;
    *) echo "❌ Unsupported OS: $OS"; exit 1 ;;
esac

echo "📋 Detected: $OS ($ARCH)"

# Check dependencies
check_dependency() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ Required dependency '$1' not found"
        return 1
    else
        echo "✅ Found $1"
        return 0
    fi
}

echo ""
echo "🔍 Checking dependencies..."
check_dependency "git" || exit 1
check_dependency "curl" || exit 1

# Optional dependencies
echo ""
echo "🔍 Checking optional dependencies..."
check_dependency "go" && echo "   - Can build from source" || echo "   - Will use prebuilt binary"
check_dependency "node" && echo "   - ESLint support available" || echo "   - ESLint support not available"
check_dependency "python3" && echo "   - Python analysis support available" || echo "   - Python analysis support not available"

echo ""
echo "📦 Installation method:"
echo "1. Build from source (requires Go)"
echo "2. Download prebuilt binary (coming soon)"
echo "3. Use Homebrew (macOS)"

read -p "Choose installation method (1-3): " METHOD

case $METHOD in
    1)
        echo "🔨 Building from source..."
        
        # Check if we're already in the codecompass directory
        if [[ -f "go.mod" ]] && grep -q "codecompass" go.mod; then
            echo "✅ Using current directory (codecompass repository)"
            BUILD_DIR="."
        else
            echo "📥 Cloning repository..."
            BUILD_DIR="/tmp/codecompass-build"
            rm -rf $BUILD_DIR
            git clone https://github.com/xeon-zolt/codecompass.git $BUILD_DIR
            cd $BUILD_DIR
        fi
        
        echo "🏗️ Building binary..."
        if [[ "$BUILD_DIR" != "." ]]; then
            cd $BUILD_DIR
        fi
        
        go build -ldflags "-s -w -X main.version=$VERSION -X main.buildDate=$(date +'%Y-%m-%d')" -o $BINARY_NAME
        
        echo "📋 Installing binary..."
        sudo cp $BINARY_NAME $INSTALL_DIR/
        sudo chmod +x $INSTALL_DIR/$BINARY_NAME
        
        if [[ "$BUILD_DIR" != "." ]]; then
            cd - > /dev/null
            rm -rf $BUILD_DIR
        fi
        ;;
        
    2)
        echo "❌ Prebuilt binaries not yet available"
        echo "💡 Please use method 1 (build from source) or method 3 (Homebrew)"
        exit 1
        ;;
        
    3)
        if [[ "$OS" != "darwin" ]]; then
            echo "❌ Homebrew is only available on macOS"
            echo "💡 Please use method 1 (build from source)"
            exit 1
        fi
        
        if ! command -v brew &> /dev/null; then
            echo "❌ Homebrew not found. Install it from https://brew.sh"
            exit 1
        fi
        
        echo "🍺 Installing via Homebrew..."
        # For now, install from source since tap isn't published yet
        echo "⚠️  Note: Installing from source via Homebrew"
        brew install go
        
        BUILD_DIR="/tmp/codecompass-homebrew"
        rm -rf $BUILD_DIR
        git clone https://github.com/xeon-zolt/codecompass.git $BUILD_DIR
        cd $BUILD_DIR
        
        go build -ldflags "-s -w -X main.version=$VERSION -X main.buildDate=$(date +'%Y-%m-%d')" -o $BINARY_NAME
        cp $BINARY_NAME /usr/local/bin/
        
        cd - > /dev/null
        rm -rf $BUILD_DIR
        ;;
        
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

# Create configuration directory
echo ""
echo "⚙️ Setting up configuration..."
mkdir -p $CONFIG_DIR

# Generate sample configuration
$BINARY_NAME --generate-config > /dev/null 2>&1 || true

if [[ -f ".codecompass.rc" ]]; then
    cp .codecompass.rc $CONFIG_DIR/codecompass.rc.example
    echo "✅ Sample configuration created at $CONFIG_DIR/codecompass.rc.example"
fi

# Verify installation
echo ""
echo "🧪 Verifying installation..."
if command -v $BINARY_NAME &> /dev/null; then
    echo "✅ CodeCompass installed successfully!"
    
    VERSION_OUTPUT=$($BINARY_NAME --version)
    echo "📋 $VERSION_OUTPUT"
else
    echo "❌ Installation failed"
    exit 1
fi

echo ""
echo "🎉 Installation Complete!"
echo "========================"
echo ""
echo "🚀 Quick Start:"
echo "  codecompass --help              # Show all available options"
echo "  codecompass --summary           # Quick repository overview"
echo "  codecompass --quality           # Comprehensive quality analysis"
echo ""
echo "📚 Advanced Features:"
echo "  codecompass --trends            # Trend analysis with charts"
echo "  codecompass --hotspots          # Detect high-risk code areas"
echo "  codecompass --team              # Team performance metrics"
echo ""
echo "⚙️ Configuration:"
echo "  codecompass --generate-config   # Create .codecompass.rc file"
echo "  Sample: $CONFIG_DIR/codecompass.rc.example"
echo ""
echo "📖 Documentation:"
echo "  • README: https://github.com/xeon-zolt/codecompass/blob/main/README.md"
echo "  • Homebrew: https://github.com/xeon-zolt/codecompass/blob/main/HOMEBREW.md"
echo ""
echo "🧭 Navigate Your Code Quality with CodeCompass!"