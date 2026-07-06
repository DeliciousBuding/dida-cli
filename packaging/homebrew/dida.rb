class Dida < Formula
  desc "JSON-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.2.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.5/dida_v0.2.5_darwin_arm64.tar.gz"
      sha256 "92a23ef9f75ccd21a1770e70ab4f2f33f1e6b3aba81595b3230812f26dec122c"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.5/dida_v0.2.5_darwin_amd64.tar.gz"
      sha256 "a8f373853dc503488249cd03716ab81e4da87702e3f9fe894a4d2c38ea7d9345"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.5/dida_v0.2.5_linux_arm64.tar.gz"
      sha256 "dbaa02791f2e3d63ecd02ea9cec8681c190589c7fbedf846e05e7e1f4acfc916"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.5/dida_v0.2.5_linux_amd64.tar.gz"
      sha256 "4879b1249bc1121c2b6f0c4999f7a9d708018951fb6ff361e75e762eb5fc1799"
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
