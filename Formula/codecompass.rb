class Codecompass < Formula
  desc "A CLI tool that helps you navigate and understand large codebases."
  homepage "https://github.com/xeoncross/codecompass"
  url "https://github.com/xeon-zolt/codecompass/archive/refs/tags/v0.0.1-beta.3.tar.gz" # Placeholder: Update with each release
  sha256 "99e7d4be71d3383faa6d00ead8f94c311b4c57b7f549e51ab2191ca58296e72f" # Placeholder: Update with each release

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    # Basic test to ensure the binary runs and outputs something
    assert_match "CodeCompass", shell_output("#{bin}/codecompass --version")
  end
end
