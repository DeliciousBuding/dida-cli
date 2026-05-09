class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.12"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.12/dida_v0.1.12_darwin_arm64.tar.gz"
      sha256 "5646a5b5764873cf88d01395772e6e982c9f38a4bc92aa94a9fee6fdaa748f90"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.12/dida_v0.1.12_darwin_amd64.tar.gz"
      sha256 "dfe94156a67ccb92c40f52b7f4ff42d564c6e0e5edcbefb0ea5700558b57bf0c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.12/dida_v0.1.12_linux_arm64.tar.gz"
      sha256 "0c86fd5d8af764309de501ddd777151e71f158401fc462017739f3416c89604d"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.12/dida_v0.1.12_linux_amd64.tar.gz"
      sha256 "1c17a1d42b0a324b5f74faabcbb0df3f6837e5381c45b7d44483d219f24e60c2"
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
