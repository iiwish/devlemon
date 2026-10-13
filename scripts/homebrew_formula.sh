#!/usr/bin/env bash
# 根据发布版本号与 checksums.txt 生成 Homebrew 公式 (iiwish/homebrew-tap: Formula/devlemon.rb)
# 用法: scripts/homebrew_formula.sh v0.2.12 checksums.txt > devlemon.rb
set -euo pipefail

TAG="${1:?usage: homebrew_formula.sh <tag> <checksums.txt>}"
CHECKSUMS="${2:?usage: homebrew_formula.sh <tag> <checksums.txt>}"
VERSION="${TAG#v}"

if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "invalid version: $TAG" >&2
  exit 1
fi

# 从 checksums.txt 取出指定平台压缩包的 sha256，缺失或格式不对时直接失败，绝不生成残缺公式
sha() {
  local file="devlemon-${TAG}-$1.tar.gz"
  local sum
  sum=$(awk -v f="$file" '$2 == f { print $1 }' "$CHECKSUMS")
  if [[ ! "$sum" =~ ^[0-9a-f]{64}$ ]]; then
    echo "missing sha256 for $file" >&2
    exit 1
  fi
  echo "$sum"
}

DARWIN_ARM64=$(sha darwin-arm64)
DARWIN_AMD64=$(sha darwin-amd64)
LINUX_ARM64=$(sha linux-arm64)
LINUX_AMD64=$(sha linux-amd64)

cat <<EOF
class Devlemon < Formula
  desc "The intelligent, context-aware disk cleanup tool built for developers and the AI era"
  homepage "https://github.com/iiwish/devlemon"
  version "${VERSION}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/iiwish/devlemon/releases/download/v#{version}/devlemon-v#{version}-darwin-arm64.tar.gz"
      sha256 "${DARWIN_ARM64}"
    else
      url "https://github.com/iiwish/devlemon/releases/download/v#{version}/devlemon-v#{version}-darwin-amd64.tar.gz"
      sha256 "${DARWIN_AMD64}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/iiwish/devlemon/releases/download/v#{version}/devlemon-v#{version}-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM64}"
    else
      url "https://github.com/iiwish/devlemon/releases/download/v#{version}/devlemon-v#{version}-linux-amd64.tar.gz"
      sha256 "${LINUX_AMD64}"
    end
  end

  def install
    bin.install "devlemon"
    bin.install_symlink "devlemon" => "dl"
  end

  test do
    system "#{bin}/devlemon", "version"
  end
end
EOF
