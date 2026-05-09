class Dida < Formula
  desc "Agent-first CLI for Dida365 and TickTick"
  homepage "https://github.com/DeliciousBuding/dida-cli"
  version "0.1.11"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.11/dida_v0.1.11_darwin_arm64.tar.gz"
      sha256 "a9a55a4738e9bbc37b27fc45c2e68e582e44dd56d31e5e37e563467a1fd0b653"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.11/dida_v0.1.11_darwin_amd64.tar.gz"
      sha256 "bb68262c8d380073fa91b31d32770cd8bfce0528b54dede28438b7640f64e5cb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.11/dida_v0.1.11_linux_arm64.tar.gz"
      sha256 "17fcebf4b84910e44f5cdf29367d69dba5b5ba4d5c04a7e4ce73bcbc08ffd859"
    else
      url "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.1.11/dida_v0.1.11_linux_amd64.tar.gz"
      sha256 "664cda47bb7194c247171d28e2321ca775fdc6e1fb21bd18d56f816655f82f37"
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
