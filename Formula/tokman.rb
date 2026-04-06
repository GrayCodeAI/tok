class Tokman < Formula
  desc "Token-aware CLI proxy with advanced quality analysis"
  homepage "https://github.com/GrayCodeAI/tokman"
  url "https://github.com/GrayCodeAI/tokman/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "" # Will be calculated after creating v0.1.0 release
  license "MIT"
  head "https://github.com/GrayCodeAI/tokman.git", branch: "main"

  depends_on "go" => :build

  def install
    # Build with version injection
    system "make", "build"
    bin.install "tokman"

    # Install shell completions
    generate_completions_from_executable(bin/"tokman", "completion")
  end

  test do
    assert_match "tokman", shell_output("#{bin}/tokman --version")
    
    # Test basic functionality
    output = shell_output("#{bin}/tokman --help")
    assert_match "Token-aware CLI proxy", output
  end
end
