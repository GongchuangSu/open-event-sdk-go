#!/usr/bin/env bash
# openevent CLI 安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/GongchuangSu/open-event-sdk-go/main/scripts/install.sh | bash
set -euo pipefail

REPO="GongchuangSu/open-event-sdk-go"
BINARY="openevent"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux)  echo "linux" ;;
    darwin) echo "darwin" ;;
    *)      echo "unsupported" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             echo "unsupported" ;;
  esac
}

get_latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
    grep '"tag_name"' | head -1 | sed -E 's/.*"v?([^"]+)".*/\1/'
}

main() {
  local os arch version url tmp_dir

  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
    echo "错误: 不支持的平台 $(uname -s)/$(uname -m)" >&2
    exit 1
  fi

  echo "检测到平台: ${os}/${arch}"

  if [ -n "${VERSION:-}" ]; then
    version="$VERSION"
  else
    echo "正在获取最新版本..."
    version="$(get_latest_version)"
    if [ -z "$version" ]; then
      echo "错误: 无法获取最新版本号" >&2
      exit 1
    fi
  fi

  echo "安装版本: v${version}"

  url="https://github.com/${REPO}/releases/download/v${version}/${BINARY}_${version}_${os}_${arch}.tar.gz"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "正在下载 ${url}..."
  if ! curl -fsSL "$url" -o "${tmp_dir}/archive.tar.gz"; then
    echo "错误: 下载失败，请检查版本号和网络连接" >&2
    exit 1
  fi

  tar -xzf "${tmp_dir}/archive.tar.gz" -C "$tmp_dir"

  if [ ! -f "${tmp_dir}/${BINARY}" ]; then
    echo "错误: 解压后未找到 ${BINARY} 二进制文件" >&2
    exit 1
  fi

  chmod +x "${tmp_dir}/${BINARY}"

  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    echo "需要 sudo 权限安装到 ${INSTALL_DIR}"
    sudo mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  fi

  echo ""
  echo "✔ ${BINARY} v${version} 已安装到 ${INSTALL_DIR}/${BINARY}"
  echo ""
  echo "快速开始:"
  echo "  ${BINARY} listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET"
  echo ""
  echo "查看帮助:"
  echo "  ${BINARY} --help"
}

main "$@"
