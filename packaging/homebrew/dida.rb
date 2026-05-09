class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.10"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.10/dida_v0.1.10_darwin_arm64.tar.gz"
      sha256 "519432415703fd729fce1cd08706682cc2bc9eaef24ae7bc8a979aabfd9a82ac"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.10/dida_v0.1.10_darwin_amd64.tar.gz"
      sha256 "0458a08dba040f61712b6592bb0f48ffb7a3f2fb45b75c13f5b1dbf19e1c86a2"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.10/dida_v0.1.10_linux_arm64.tar.gz"
      sha256 "6a35e2510cb06e601ea8d8267ed2a24178cd0c6b7b6f0a9817fda3f5e5a9e8b4"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.10/dida_v0.1.10_linux_amd64.tar.gz"
      sha256 "c9d87eb77e8b4a488c89ea331b7ded2cf480dd42d7ab34cef21bc0b630761d9a"
    end
  end

  def install
    bin.install "dida"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dida version")
    assert_match "\"ok\": true", shell_output("#{bin}/dida doctor --json")
  end
end
