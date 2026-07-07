class Dida < Formula
  desc "JSON-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.2.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.6/dida_v0.2.6_darwin_arm64.tar.gz"
      sha256 "4af818f143b891ed3d55b297de9f52641f22d953e17aa5f7b2cd2804c6a567fc"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.6/dida_v0.2.6_darwin_amd64.tar.gz"
      sha256 "e60f730dd01343eaa35f79a701a77b944d9d6440236bf8bc7dcfdaae78398dc9"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.6/dida_v0.2.6_linux_arm64.tar.gz"
      sha256 "aa56c2f60b9b019d89fe684fe5ad3c1bab7eb3df0d4624cefb0462419bed4cdf"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.6/dida_v0.2.6_linux_amd64.tar.gz"
      sha256 "06b7237d69e7701f997501278645f2199080ba572145a9d831971d367c19a11e"
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
