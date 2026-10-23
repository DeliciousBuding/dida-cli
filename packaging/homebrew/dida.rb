class Dida < Formula
  desc "JSON-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.2.10"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.10/dida_v0.2.10_darwin_arm64.tar.gz"
      sha256 "27333d3303fec37df56d7bece3b690052126d7bb919a90f63df24f2fd11447c5"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.10/dida_v0.2.10_darwin_amd64.tar.gz"
      sha256 "36788a1adb550a54117da35cbdce49c9b3dc0f181dcf557adf868d024f0fd4f0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.10/dida_v0.2.10_linux_arm64.tar.gz"
      sha256 "954bf84ccf2799a678c86c21c3ba1df15ef383371caf67e4d2c680c749d41015"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.10/dida_v0.2.10_linux_amd64.tar.gz"
      sha256 "c64e2cf7bc9814b4d56bf5322fa0facf37a19d576bd5df703f6aaa4dd9d5a0d6"
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
