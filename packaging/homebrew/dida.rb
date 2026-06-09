class Dida < Formula
  desc "JSON-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.3/dida_v0.2.3_darwin_arm64.tar.gz"
      sha256 "47fee46e764877034fa769bd6a74ff9d9479f43e07605adc73e23eebfc76b39e"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.3/dida_v0.2.3_darwin_amd64.tar.gz"
      sha256 "dd5a653c4c6ddeee44f01e08ffa940acba17fb9edcfe4365bb9a6b52277033ad"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.3/dida_v0.2.3_linux_arm64.tar.gz"
      sha256 "3104fdb4efb9916277828a6b853304c684c14df307b7b1859af0896a465ba55a"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.3/dida_v0.2.3_linux_amd64.tar.gz"
      sha256 "7329965fab555d840c1d8e96c468441301a4df8629f19b1f0b6f8f8b8467605c"
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
