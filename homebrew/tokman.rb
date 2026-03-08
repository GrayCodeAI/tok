# Homebrew formula for TokMan
# Install with: brew install GrayCodeAI/tokman/tokman

class Tokman < Formula
  desc "Token-aware CLI proxy for LLM interactions"
  homepage "https://github.com/GrayCodeAI/tokman"
  version "1.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/GrayCodeAI/tokman/releases/download/v#{version}/tokman-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_AMD64_SHA256"
    end
    on_arm do
      url "https://github.com/GrayCodeAI/tokman/releases/download/v#{version}/tokman-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_ARM64_SHA256"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/GrayCodeAI/tokman/releases/download/v#{version}/tokman-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
    on_arm do
      url "https://github.com/GrayCodeAI/tokman/releases/download/v#{version}/tokman-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
    end
  end

  def install
    bin.install "tokman"
    
    # Install shell completions
    generate_completions_from_executable(bin/"tokman", "completion", shells: [:bash, :zsh, :fish])
    
    # Install man page if available
    man1.install "tokman.1" if File.exist?("tokman.1")
  end

  def caveats
    <<~EOS
      TokMan installed successfully!
      
      To set up shell integration, run:
        tokman init
      
      To view token savings dashboard:
        tokman dashboard
      
      Documentation: https://github.com/GrayCodeAI/tokman#readme
    EOS
  end

  test do
    assert_match "TokMan #{version}", shell_output("#{bin}/tokman --version")
    assert_match "Token-aware CLI proxy", shell_output("#{bin}/tokman --help")
  end
end
