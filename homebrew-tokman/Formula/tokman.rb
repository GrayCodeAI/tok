class Tokman < Formula
  desc "Token-aware CLI proxy with advanced quality analysis"
  homepage "https://github.com/GrayCodeAI/tokman"
  url "https://github.com/GrayCodeAI/tokman/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "" # Update with actual SHA256 after release
  license "MIT"
  head "https://github.com/GrayCodeAI/tokman.git", branch: "main"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "tokman"
    generate_completions_from_executable(bin/"tokman", "completion")
  end

  test do
    assert_match "tokman", shell_output("#{bin}/tokman --version")
    assert_match "Token-aware CLI proxy", shell_output("#{bin}/tokman --help")
  end
end
