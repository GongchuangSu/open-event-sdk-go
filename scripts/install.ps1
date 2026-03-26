# openevent CLI Windows 安装脚本
# 用法: irm https://raw.githubusercontent.com/GongchuangSu/open-event-sdk-go/main/scripts/install.ps1 | iex
$ErrorActionPreference = "Stop"

$Repo = "GongchuangSu/open-event-sdk-go"
$Binary = "openevent"

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { Write-Error "不支持的架构: $arch"; exit 1 }
    }
}

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name -replace '^v', ''
}

$arch = Get-Arch
Write-Host "检测到平台: windows/$arch"

if ($env:VERSION) {
    $version = $env:VERSION
} else {
    Write-Host "正在获取最新版本..."
    $version = Get-LatestVersion
    if (-not $version) {
        Write-Error "无法获取最新版本号"
        exit 1
    }
}

Write-Host "安装版本: v$version"

$url = "https://github.com/$Repo/releases/download/v$version/${Binary}_${version}_windows_${arch}.zip"
$installDir = "$env:LOCALAPPDATA\openevent"
$tmpZip = Join-Path $env:TEMP "openevent.zip"
$tmpDir = Join-Path $env:TEMP "openevent_extract"

Write-Host "正在下载 $url..."
try {
    Invoke-WebRequest -Uri $url -OutFile $tmpZip -UseBasicParsing
} catch {
    Write-Error "下载失败，请检查版本号和网络连接"
    exit 1
}

if (Test-Path $tmpDir) { Remove-Item -Recurse -Force $tmpDir }
Expand-Archive -Path $tmpZip -DestinationPath $tmpDir

if (-not (Test-Path (Join-Path $tmpDir "$Binary.exe"))) {
    Write-Error "解压后未找到 $Binary.exe"
    exit 1
}

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}
Move-Item -Force (Join-Path $tmpDir "$Binary.exe") (Join-Path $installDir "$Binary.exe")

Remove-Item -Force $tmpZip -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host ""
    Write-Host "已将 $installDir 添加到用户 PATH（重启终端生效）"
}

Write-Host ""
Write-Host "✔ $Binary v$version 已安装到 $installDir\$Binary.exe"
Write-Host ""
Write-Host "快速开始:"
Write-Host "  $Binary listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET"
Write-Host ""
Write-Host "查看帮助:"
Write-Host "  $Binary --help"
