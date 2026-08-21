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
$EmbeddedFrontendPath = Join-Path $BuildDir "embedded-frontend.html"

function Convert-FileToDataUrl {
    param(
        [Parameter(Mandatory=$true)][string]$Path
    )

    $extension = [System.IO.Path]::GetExtension($Path).ToLowerInvariant()
    $mime = switch ($extension) {
        ".png" { "image/png" }
        ".jpg" { "image/jpeg" }
        ".jpeg" { "image/jpeg" }
        ".webp" { "image/webp" }
        ".svg" { "image/svg+xml" }
        ".ico" { "image/x-icon" }
        ".woff" { "font/woff" }
        ".woff2" { "font/woff2" }
        default { "application/octet-stream" }
    }
    $bytes = [System.IO.File]::ReadAllBytes($Path)
    return "data:$mime;base64,$([Convert]::ToBase64String($bytes))"
}

function Resolve-DistAssetPath {
    param(
        [Parameter(Mandatory=$true)][string]$DistDir,
        [Parameter(Mandatory=$true)][string]$Reference
    )

    $clean = $Reference.Split("?")[0]
    if ($clean.StartsWith("/")) {
        return Join-Path $DistDir $clean.TrimStart("/")
    }
    return Join-Path $DistDir $clean
}

function Inline-CssUrls {
    param(
        [Parameter(Mandatory=$true)][string]$Css,
        [Parameter(Mandatory=$true)][string]$DistDir
    )

    return [regex]::Replace($Css, 'url\(([''"]?)([^)''"]+)\1\)', {
        param($match)
        $reference = $match.Groups[2].Value
        if ($reference.StartsWith("data:") -or $reference.StartsWith("http://") -or $reference.StartsWith("https://")) {
            return $match.Value
        }
        $assetPath = Resolve-DistAssetPath -DistDir $DistDir -Reference $reference
        if (!(Test-Path $assetPath)) {
            return $match.Value
        }
        return "url(" + (Convert-FileToDataUrl -Path $assetPath) + ")"
    })
}

function New-EmbeddedFrontend {
    param(
        [Parameter(Mandatory=$true)][string]$DistDir,
        [Parameter(Mandatory=$true)][string]$OutputPath
    )

    $indexPath = Join-Path $DistDir "index.html"
    if (!(Test-Path $indexPath)) {
        throw "Frontend dist index was not found: $indexPath"
    }

    $html = [System.IO.File]::ReadAllText($indexPath, [System.Text.Encoding]::UTF8)

    $html = [regex]::Replace($html, '<link\b(?=[^>]*\brel=["'']stylesheet["''])(?=[^>]*\bhref=["'']([^"'']+\.css)["''])[^>]*>', {
        param($match)
        $assetPath = Resolve-DistAssetPath -DistDir $DistDir -Reference $match.Groups[1].Value
        if (!(Test-Path $assetPath)) {
            throw "Stylesheet referenced by index.html was not found: $assetPath"
        }
        $css = [System.IO.File]::ReadAllText($assetPath, [System.Text.Encoding]::UTF8)
        $css = Inline-CssUrls -Css $css -DistDir $DistDir
        return "<style>$css</style>"
    })

    $html = [regex]::Replace($html, '<script\b(?=[^>]*\bsrc=["'']([^"'']+\.js)["''])[^>]*>\s*</script>', {
        param($match)
        $assetPath = Resolve-DistAssetPath -DistDir $DistDir -Reference $match.Groups[1].Value
        if (!(Test-Path $assetPath)) {
            throw "Script referenced by index.html was not found: $assetPath"
        }
        $js = [System.IO.File]::ReadAllText($assetPath, [System.Text.Encoding]::UTF8)
        return "<script type=`"module`">$js</script>"
    })

    $html = [regex]::Replace($html, '(src|href)=["''](/assets/[^"'']+\.(png|jpg|jpeg|webp|svg|ico))["'']', {
        param($match)
        $assetPath = Resolve-DistAssetPath -DistDir $DistDir -Reference $match.Groups[2].Value
        if (!(Test-Path $assetPath)) {
            return $match.Value
        }
        return $match.Groups[1].Value + "=`"" + (Convert-FileToDataUrl -Path $assetPath) + "`""
    })

    $outputDirectory = Split-Path -Parent $OutputPath
    if (!(Test-Path $outputDirectory)) {
        New-Item -ItemType Directory -Path $outputDirectory | Out-Null
    }
    [System.IO.File]::WriteAllText($OutputPath, $html, [System.Text.UTF8Encoding]::new($false))
}

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

function Test-WebView2SdkAvailable {
    if ($env:WEBVIEW2_SDK_DIR) {
        return (Test-Path (Join-Path $env:WEBVIEW2_SDK_DIR "include\WebView2.h"))
    }

    $packageRoot = Join-Path $env:USERPROFILE ".nuget\packages\microsoft.web.webview2"
    if (!(Test-Path $packageRoot)) {
        return $false
    }

    $packageDirectory = Get-ChildItem -LiteralPath $packageRoot -Directory |
        Sort-Object Name -Descending |
        Select-Object -First 1
    if (!$packageDirectory) {
        return $false
    }

    return (Test-Path (Join-Path $packageDirectory.FullName "build\native\include\WebView2.h"))
}

foreach ($commandName in @("dotnet", "node", "npm")) {
    if (!(Get-Command $commandName -ErrorAction SilentlyContinue)) {
        throw "Missing command: $commandName"
    }
}

$VsDevCmd = Resolve-VsDevCmd

if (Test-WebView2SdkAvailable) {
    Write-Host "WebView2 SDK already available."
} else {
    Write-Host "Restoring WebView2 SDK..."
    dotnet restore (Join-Path $CppDir "WebView2SdkRestore.csproj")
    if ($LASTEXITCODE -ne 0) {
        throw "WebView2 SDK restore failed with exit code $LASTEXITCODE"
    }
}

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

$DistDir = Join-Path $LauncherDir "dist"
Write-Host "Embedding frontend resources..."
New-EmbeddedFrontend -DistDir $DistDir -OutputPath $EmbeddedFrontendPath

$CMakeCommand = "cmake -S `"$CppDir`" -B `"$BuildDir`" -G Ninja -DCMAKE_BUILD_TYPE=$Configuration -DFRONTEND_HTML_PATH=`"$EmbeddedFrontendPath`" && cmake --build `"$BuildDir`" --config $Configuration"
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

Write-Host ""
Write-Host "Build finished:"
Write-Host $ExePath
