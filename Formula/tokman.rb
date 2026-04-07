# typed: false
# frozen_string_literal: true

# Homebrew formula for TokMan
# Token-aware CLI proxy with 31-layer compression pipeline
#
# To install:
#   brew install GrayCodeAI/tap/tokman
#
# Or from this formula:
#   brew install --build-from-source Formula/tokman.rb

class Tokman < Formula
  desc "Token-aware CLI proxy with 31-layer compression pipeline for AI coding assistants"
  homepage "https://github.com/GrayCodeAI/tokman"
  url "https://github.com/GrayCodeAI/tokman/archive/refs/tags/v0.28.2.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "MIT"
  head "https://github.com/GrayCodeAI/tokman.git", branch: "main"

  # Bottles will be added once CI builds them
  # bottle do
  #   sha256 cellar: :any_skip_relocation, arm64_sonoma: "PLACEHOLDER"
  #   sha256 cellar: :any_skip_relocation, sonoma:       "PLACEHOLDER"
  #   sha256 cellar: :any_skip_relocation, x86_64_linux: "PLACEHOLDER"
  # end

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X main.commit=#{tap.user}
      -X main.date=#{time.iso8601}
    ]

    system "go", "build",
           *std_go_args(ldflags:),
           "./cmd/tokman"

    # Install shell completions
    generate_completions_from_executable(bin/"tokman", "completion")
  end

  def post_install
    # Create default config directory
    (var/"tokman").mkpath
  end

  def caveats
    <<~EOS
      TokMan has been installed! To get started:

        # Initialize for your AI tool
        tokman init -g                    # Claude Code
        tokman init -g --cursor           # Cursor
        tokman init --all                 # All detected tools

        # Verify installation
        tokman doctor

        # View token savings
        tokman gain

      Configuration: #{etc}/tokman/config.toml
      Database: #{var}/tokman/tokman.db

      For more information:
        https://github.com/GrayCodeAI/tokman
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/tokman --version")

    # Test doctor command
    output = shell_output("#{bin}/tokman doctor 2>&1", 0)
    assert_match(/TokMan|doctor|check/i, output)

    # Test basic filtering
    input = "line1\nline2\nline3\nline4\nline5"
    output = pipe_output("#{bin}/tokman filter --stdin", input)
    assert_predicate output.length, :positive?
  end
end
