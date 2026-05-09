class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.5/dida_v0.1.5_darwin_arm64.tar.gz"
      sha256 "96d0f94a5dece568237802db94a71c48e69854947f4ecdf531fa48773640a721"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.5/dida_v0.1.5_darwin_amd64.tar.gz"
      sha256 "e1cd7c3af0d3e243cacc0f66b8d2a0c9181ff1e10785a489fc6af36b36ced383"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.5/dida_v0.1.5_linux_arm64.tar.gz"
      sha256 "1ac00a3c165872fdad9d11bd0a78ea45ea3ea760e89b110b10e5067c68c0b003"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.5/dida_v0.1.5_linux_amd64.tar.gz"
      sha256 "9b0d6a81ea0adeeb6822e94b0daf6cb17abfc8a8d345dc820571814a59c7370a"
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
