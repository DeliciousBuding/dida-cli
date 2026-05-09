class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.8"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.8/dida_v0.1.8_darwin_arm64.tar.gz"
      sha256 "6a9d11a8694568718d92d3594c29f26f78863fef0370ec00fa6fed708c6bbc5f"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.8/dida_v0.1.8_darwin_amd64.tar.gz"
      sha256 "5ae10c94319a065cf4f447e62cac5beca6bd3a6c184bbf668a225c30c0b2304b"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.8/dida_v0.1.8_linux_arm64.tar.gz"
      sha256 "5a06405cff7cfb0e6a4adebde24b26c5b85a5b48ab40764190a90d3ebbb08641"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.8/dida_v0.1.8_linux_amd64.tar.gz"
      sha256 "80995375c5a4807958496e5b36d03705e1efa964d2512600a3e3388c7e662b61"
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
