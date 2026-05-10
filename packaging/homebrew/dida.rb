class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.16"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.16/dida_v0.1.16_darwin_arm64.tar.gz"
      sha256 "418daecccd8815a404929a097b1cbd0c73fbb219c5e050dbe8e7201dddbd5da5"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.16/dida_v0.1.16_darwin_amd64.tar.gz"
      sha256 "77c1a16f525144d53d13ec3ff014b2ef72d907f90efa345822c0f0d3c458b63f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.16/dida_v0.1.16_linux_arm64.tar.gz"
      sha256 "387f2a349ae6503be7ea4b0943b166393295236d1542029f312f61b16e9442dc"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.16/dida_v0.1.16_linux_amd64.tar.gz"
      sha256 "7150e2f9121b55b5a2c619620e9f9e650f6b9e375e480a2656582ff0054a0131"
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
