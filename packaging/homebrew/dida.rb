class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.7/dida_v0.1.7_darwin_arm64.tar.gz"
      sha256 "11540974763c3c8c8f453a8ff4afbd361598d83db686fb5c65da174ed15e1974"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.7/dida_v0.1.7_darwin_amd64.tar.gz"
      sha256 "ff3e9301ba8a74053387c2fc3427614c42269ab75fdc291d4f94a9eb7da025d8"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.7/dida_v0.1.7_linux_arm64.tar.gz"
      sha256 "b48da00fbb507c988b16e94720a39672c05b3d8afa8f57ea825186fd489292e7"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.7/dida_v0.1.7_linux_amd64.tar.gz"
      sha256 "b60ec5e277b7737207ead5e6d5045ba75ff35b770b6b12b0897624b4ce0b83b2"
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
