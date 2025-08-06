class Codecompass < Formula
  desc "A CLI tool that helps you navigate and understand large codebases."
  homepage "https://github.com/xeoncross/codecompass"
  url "https://github.com/xeoncross/codecompass/archive/refs/tags/v0.0.1.tar.gz" # Placeholder: Update with each release
  sha256 "0000000000000000000000000000000000000000000000000000000000000000" # Placeholder: Update with each release

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    # Basic test to ensure the binary runs and outputs something
    assert_match "CodeCompass", shell_output("#{bin}/codecompass --version")
  end
end
