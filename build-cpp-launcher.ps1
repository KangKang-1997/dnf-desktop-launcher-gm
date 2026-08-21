param(
    [string]$ApiBase = $env:DNF_LAUNCHER_API_BASE,
    [string]$Configuration = "Release",
    [switch]$SkipFrontend
)

$ErrorActionPreference = "Stop"

if (!$ApiBase) {
    $ApiBase = "http://127.0.0.1:8000"
}

if ($ApiBase -notmatch '^https?://') {
    throw "ApiBase must start with http:// or https://. Current value: $ApiBase"
}

$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$LauncherDir = Join-Path $ProjectRoot "desktop_launcher"
$CppDir = Join-Path $ProjectRoot "cpp_launcher"
$BuildDir = Join-Path $CppDir "build-ninja"

function Resolve-VsDevCmd {
    if ($env:VSINSTALLDIR) {
        $candidate = Join-Path $env:VSINSTALLDIR "Common7\Tools\VsDevCmd.bat"
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    $vswhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
    if (Test-Path $vswhere) {
        $installPath = & $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
        if ($installPath) {
            $candidate = Join-Path $installPath "Common7\Tools\VsDevCmd.bat"
            if (Test-Path $candidate) {
                return $candidate
            }
        }
    }

    foreach ($candidate in @(
        "D:\Microsoft Visual Studio\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Community\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\BuildTools\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Professional\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Enterprise\Common7\Tools\VsDevCmd.bat"
    )) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "VsDevCmd.bat was not found. Install Visual Studio Build Tools with Desktop development with C++."
}

foreach ($commandName in @("dotnet", "node", "npm")) {
    if (!(Get-Command $commandName -ErrorAction SilentlyContinue)) {
        throw "Missing command: $commandName"
    }
}

$VsDevCmd = Resolve-VsDevCmd

Write-Host "Restoring WebView2 SDK..."
dotnet restore (Join-Path $CppDir "WebView2SdkRestore.csproj")

if (!$SkipFrontend) {
    Push-Location $LauncherDir
    try {
        $env:DNF_LAUNCHER_API_BASE = $ApiBase.TrimEnd("/")

        Write-Host "API base: $env:DNF_LAUNCHER_API_BASE"
        Write-Host "Installing/checking npm dependencies..."
        npm install

        Write-Host "Building web frontend..."
        node ".\node_modules\vite\bin\vite.js" build
    } finally {
        Pop-Location
    }
}

$CMakeCommand = "cmake -S `"$CppDir`" -B `"$BuildDir`" -G Ninja && cmake --build `"$BuildDir`" --config $Configuration"
$Cmd = "call `"$VsDevCmd`" -arch=x64 -host_arch=x64 >nul && $CMakeCommand"

Get-Process -Name "dnf-webview2-launcher" -ErrorAction SilentlyContinue | Stop-Process -Force

Write-Host "Building C++ WebView2 launcher..."
cmd.exe /d /c $Cmd
if ($LASTEXITCODE -ne 0) {
    throw "C++ build failed with exit code $LASTEXITCODE"
}

$ExePath = Join-Path $BuildDir "dnf-webview2-launcher.exe"
if (!(Test-Path $ExePath)) {
    throw "Build finished but EXE was not found: $ExePath"
}

$DistDir = Join-Path $LauncherDir "dist"
if (Test-Path $DistDir) {
    $WebDir = Join-Path $BuildDir "web"
    if (Test-Path $WebDir) {
        Remove-Item -LiteralPath $WebDir -Recurse -Force
    }
    Copy-Item -LiteralPath $DistDir -Destination $WebDir -Recurse
}

Write-Host ""
Write-Host "Build finished:"
Write-Host $ExePath
