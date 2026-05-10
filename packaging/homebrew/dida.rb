class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.14"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.14/dida_v0.1.14_darwin_arm64.tar.gz"
      sha256 "b51a5d2e6253aab9731b2168c2768b2feab18309a7fe52b62258fee506c2cbd8"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.14/dida_v0.1.14_darwin_amd64.tar.gz"
      sha256 "bfb239c314b22435ce1ca62b6e17fb6e8028e7185386c76593e92b78dab5dfda"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.14/dida_v0.1.14_linux_arm64.tar.gz"
      sha256 "f2713f64dbef4505db43851ad1f7b9866fdd1489421136f9eb8db23b74ac1bb0"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.14/dida_v0.1.14_linux_amd64.tar.gz"
      sha256 "9e6d89eab9270e5e2c529578dbba15a7bf73ba3f53fbf6f6a66a187448358544"
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
