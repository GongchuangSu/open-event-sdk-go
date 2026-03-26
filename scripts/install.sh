#!/usr/bin/env bash
# openevent CLI 安装脚本（macOS / Linux / Windows Git Bash）
# 用法: curl -fsSL https://raw.githubusercontent.com/GongchuangSu/open-event-sdk-go/main/scripts/install.sh | bash
set -euo pipefail

REPO="GongchuangSu/open-event-sdk-go"
BINARY="openevent"

detect_os() {
  local os
  os="$(uname -s)"
  case "$os" in
    Linux)              echo "linux" ;;
    Darwin)             echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *)                  echo "unsupported" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             echo "unsupported" ;;
  esac
}

get_latest_version() {
  # 通过重定向 URL 提取版本号，不依赖 GitHub API（避免限流 403）
  local url
  url="$(curl -fsSI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null)"
  if [ -n "$url" ]; then
    echo "$url" | sed -E 's|.*/v?||'
    return
  fi
  # 回退到 API 方式
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null |
    grep '"tag_name"' | head -1 | sed -E 's/.*"v?([^"]+)".*/\1/'
}

default_install_dir() {
  local os="$1"
  if [ "$os" = "windows" ]; then
    echo "${LOCALAPPDATA:-$HOME}/openevent"
  else
    echo "/usr/local/bin"
  fi
}

main() {
  local os arch version url tmp_dir ext bin_name install_dir

  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
    echo "错误: 不支持的平台 $(uname -s)/$(uname -m)" >&2
    exit 1
  fi

  install_dir="${INSTALL_DIR:-$(default_install_dir "$os")}"

  if [ "$os" = "windows" ]; then
    ext="zip"
    bin_name="${BINARY}.exe"
  else
    ext="tar.gz"
    bin_name="${BINARY}"
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

  url="https://github.com/${REPO}/releases/download/v${version}/${BINARY}_${version}_${os}_${arch}.${ext}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "正在下载 ${url}..."
  if ! curl -fsSL "$url" -o "${tmp_dir}/archive.${ext}"; then
    echo "错误: 下载失败，请检查版本号和网络连接" >&2
    exit 1
  fi

  if [ "$ext" = "zip" ]; then
    if command -v unzip >/dev/null 2>&1; then
      unzip -qo "${tmp_dir}/archive.zip" -d "$tmp_dir"
    elif command -v powershell >/dev/null 2>&1; then
      powershell -NoProfile -Command \
        "Expand-Archive -Force '${tmp_dir}/archive.zip' '${tmp_dir}'"
    else
      echo "错误: 需要 unzip 或 powershell 来解压 .zip 文件" >&2
      exit 1
    fi
  else
    tar -xzf "${tmp_dir}/archive.tar.gz" -C "$tmp_dir"
  fi

  if [ ! -f "${tmp_dir}/${bin_name}" ]; then
    echo "错误: 解压后未找到 ${bin_name}" >&2
    exit 1
  fi

  chmod +x "${tmp_dir}/${bin_name}"

  mkdir -p "$install_dir"

  if [ -w "$install_dir" ]; then
    mv "${tmp_dir}/${bin_name}" "${install_dir}/${bin_name}"
  else
    echo "需要 sudo 权限安装到 ${install_dir}"
    sudo mv "${tmp_dir}/${bin_name}" "${install_dir}/${bin_name}"
  fi

  echo ""
  echo "✔ ${BINARY} v${version} 已安装到 ${install_dir}/${bin_name}"

  if [ "$os" = "windows" ]; then
    case ":${PATH}:" in
      *":${install_dir}:"*) ;;
      *)
        echo ""
        echo "提示: 请将 ${install_dir} 添加到 PATH 环境变量"
        echo "  PowerShell: \$env:Path += \";${install_dir}\""
        echo "  或通过 系统属性 → 环境变量 永久添加"
        ;;
    esac
  fi

  echo ""
  echo "快速开始:"
  echo "  ${BINARY} listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET"
  echo ""
  echo "查看帮助:"
  echo "  ${BINARY} --help"
}

main "$@"
