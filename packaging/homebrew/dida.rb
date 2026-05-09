class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.4/dida_v0.1.4_darwin_arm64.tar.gz"
      sha256 "232a681e0dd8877341b775eaa7d8eaac3b4917f1af7af50e3998a1e00b019732"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.4/dida_v0.1.4_darwin_amd64.tar.gz"
      sha256 "94cf9990ddc04166edce2e8f43d5e8f7a517c6b486bd79355b85dfda600a678c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.4/dida_v0.1.4_linux_arm64.tar.gz"
      sha256 "ae96330134e3f64346b5465cd433d137ea7265d5a7f8a07dc84f8aa4580e9307"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.4/dida_v0.1.4_linux_amd64.tar.gz"
      sha256 "833f82fea02a4d2096789eee86707943c4330b3eab3b5ddcbf84e1834aa57e33"
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
