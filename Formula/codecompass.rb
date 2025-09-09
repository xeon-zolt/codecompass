class Codecompass < Formula
  desc "Navigate Your Code Quality - A comprehensive code quality analysis tool with advanced analytics"
  homepage "https://github.com/xeon-zolt/codecompass"
  url "https://github.com/xeon-zolt/codecompass/archive/refs/tags/v1.0.0.tar.gz" # Update with each release
  sha256 "PLACEHOLDER_SHA256_HASH" # Calculate with: shasum -a 256 codecompass-v1.0.0.tar.gz
  license "MIT"

  depends_on "go" => :build
  depends_on "git"
  
  # Optional dependencies for enhanced functionality
  uses_from_macos "node" => :optional # For ESLint support
  uses_from_macos "python3" => :optional # For Ruff support

  def install
    # Build with version information
    version_ldflags = "-X main.version=#{version} -X main.buildDate=#{Time.now.strftime('%Y-%m-%d')}"
    system "go", "build", *std_go_args(ldflags: "-s -w #{version_ldflags}")
    
    # Create default config directory
    (etc/"codecompass").mkpath
    
    # Install sample configuration
    system bin/"codecompass", "--generate-config"
    (etc/"codecompass").install ".codecompass.rc" => "codecompass.rc.example"
  end

  def caveats
    <<~EOS
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
        Sample config: #{etc}/codecompass/codecompass.rc.example
        Copy to your repo: cp #{etc}/codecompass/codecompass.rc.example .codecompass.rc
        
      Optional Dependencies:
        • Install ESLint for JavaScript analysis: npm install -g eslint
        • Install Ruff for Python analysis: pip install ruff
    EOS
  end

  test do
    # Test version output
    version_output = shell_output("#{bin}/codecompass --version")
    assert_match "CodeCompass", version_output
    assert_match version.to_s, version_output
    
    # Test help output
    help_output = shell_output("#{bin}/codecompass --help")
    assert_match "Navigate Your Code Quality", help_output
    assert_match "--quality", help_output
    assert_match "--trends", help_output
    assert_match "--hotspots", help_output
    assert_match "--team", help_output
    
    # Test config generation
    system bin/"codecompass", "--generate-config"
    assert_predicate testpath/".codecompass.rc", :exist?
    
    # Test that it can run basic analysis (in a git repo)
    system "git", "init"
    system "git", "config", "user.name", "Test User"
    system "git", "config", "user.email", "test@example.com"
    (testpath/"test.go").write "package main\nfunc main() {}\n"
    system "git", "add", "."
    system "git", "commit", "-m", "Initial commit"
    
    # Test summary command
    summary_output = shell_output("#{bin}/codecompass --summary")
    assert_match "Repository Summary", summary_output
  end
end
