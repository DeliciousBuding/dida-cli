class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.6/dida_v0.1.6_darwin_arm64.tar.gz"
      sha256 "0a5adb34d680c4283359d2fc29849e5a2ee02b1928aa6559d881648777ca695d"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.6/dida_v0.1.6_darwin_amd64.tar.gz"
      sha256 "85dc02c7b5e6c3ed6cdb6117eb2dd68494031d8ba8a79003719c8b9b79c79056"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.6/dida_v0.1.6_linux_arm64.tar.gz"
      sha256 "c0c108b5cb734e1313e353399e8d9864216ab8ab5dbd731e2966edb0645f40b9"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.6/dida_v0.1.6_linux_amd64.tar.gz"
      sha256 "9d909b99b547f4556d73d04e1cf882a76ad8f52225671ad349bb732540f4a6f0"
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
