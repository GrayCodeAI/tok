class Tok < Formula
  desc "Unified token optimization CLI"
  homepage "https://github.com/GrayCodeAI/tok"
  url "https://github.com/GrayCodeAI/tok/archive/refs/tags/v0.29.0.tar.gz"
  sha256 "..." # will be updated on release
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/tok/"
    bin.install "tok"
  end

  test do
    assert_match "tok", shell_output("#{bin}/tok --version")
  end
end
