class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.13"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.13/dida_v0.1.13_darwin_arm64.tar.gz"
      sha256 "71605d6c8e656cf5bff71bbb2bc9257e938aef19d8f1d9802d9d10d5df5e9388"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.13/dida_v0.1.13_darwin_amd64.tar.gz"
      sha256 "df94006461f5861be1117e5e1b9246bdcb16650030d201d9e4959e89eee034b0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.13/dida_v0.1.13_linux_arm64.tar.gz"
      sha256 "68852426ffb8775b783965acf256fa0091f91eb44f78bd10d4b5c5b5c2cb3729"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.13/dida_v0.1.13_linux_amd64.tar.gz"
      sha256 "f5d84f523aa71b10a72e3cdb3f7170269df7846704b2c09a930a7df611de8fc5"
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
