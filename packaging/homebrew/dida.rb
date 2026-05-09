class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.9"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.9/dida_v0.1.9_darwin_arm64.tar.gz"
      sha256 "18f47b8ce033c2e8c58804d00dfbd70053a3587e2eb28d7515fe9ea3f7216639"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.9/dida_v0.1.9_darwin_amd64.tar.gz"
      sha256 "02dd57b3260d58113baa8fcc077b9d2583a089381aec7b008f88cb96ebcc0923"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.9/dida_v0.1.9_linux_arm64.tar.gz"
      sha256 "ae87434a1002153805fdc5e5ecf26a43e5ff937b408b59c93967859ab302680a"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.9/dida_v0.1.9_linux_amd64.tar.gz"
      sha256 "021d0facaf79c05c4aec37e056db95cd372294794edbfb2009a15e90a6d67c0d"
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
