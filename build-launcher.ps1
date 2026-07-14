$ErrorActionPreference = "Stop"

# Change this API address before running this script.

$ApiBase = "http://服务器IP:8000"

if ($ApiBase -notmatch '^https?://') {
    throw "ApiBase must start with http:// or https://. Current value: $ApiBase"
}

$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$LauncherDir = Join-Path $ProjectRoot "desktop_launcher"

if (!(Test-Path (Join-Path $LauncherDir "package.json"))) {
    throw "desktop_launcher/package.json was not found. Put this script in the project root."
}

foreach ($commandName in @("node", "npm", "cargo")) {
    if (!(Get-Command $commandName -ErrorAction SilentlyContinue)) {
        throw "Missing command: $commandName. Install the Windows build environment first."
    }
}

Push-Location $LauncherDir
try {
    $env:DNF_LAUNCHER_API_BASE = $ApiBase.TrimEnd("/")

    Write-Host "API base: $env:DNF_LAUNCHER_API_BASE"
    Write-Host "Installing/checking npm dependencies..."
    npm install

    Write-Host "Building launcher EXE..."
    npm run tauri -- build

    $ExePath = Join-Path $LauncherDir "src-tauri\target\release\dnf-desktop-launcher.exe"
    if (!(Test-Path $ExePath)) {
        throw "Build finished but EXE was not found: $ExePath"
    }

    Write-Host ""
    Write-Host "Build finished:"
    Write-Host $ExePath
} finally {
    Pop-Location
}
