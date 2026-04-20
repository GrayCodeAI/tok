# Homebrew formula for tok.
#
# Install:   brew install --HEAD lakshmanpatel/tok/tok
# Upgrade:   brew upgrade tok
#
# For tagged releases, replace `head` with `url` + `sha256` once a GitHub
# Releases page is publishing signed archives.
class Tok < Formula
  desc "Transparent command-output filter that reduces LLM token consumption"
  homepage "https://github.com/lakshmanpatel/tok"
  license "MIT"
  head "https://github.com/lakshmanpatel/tok.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/tok"
  end

  test do
    assert_match "tok", shell_output("#{bin}/tok --version 2>&1", 0)
  end
end
