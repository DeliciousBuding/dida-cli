class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.15"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.15/dida_v0.1.15_darwin_arm64.tar.gz"
      sha256 "93c1e51c74043d5515c644ac88275f95e295bc39dc613a7245e3221f4abb4d04"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.15/dida_v0.1.15_darwin_amd64.tar.gz"
      sha256 "b0c40599e00b77ffa03ac4cef3f54630b5a68253995a04a7cce20f9d8cc6e14b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.15/dida_v0.1.15_linux_arm64.tar.gz"
      sha256 "7e6318452163626b3e182e984845e9030db3fba26a2a837f63902caf469aad76"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.15/dida_v0.1.15_linux_amd64.tar.gz"
      sha256 "f75934acb92956322ca62470fcbdde7d320e73a5b954085bda9c8cafab55e655"
    end
  end

  def install
    binary = Dir["**/dida"].find { |path| File.file?(path) }
    odie "dida binary not found in release archive" unless binary

    bin.install binary => "dida"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dida version")
    assert_match "\"ok\": true", shell_output("#{bin}/dida doctor --json")
  end
end
