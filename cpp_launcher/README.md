# C++ WebView2 Launcher

This is the Windows desktop shell for the DNF launcher.

Current scope:

- Loads the frontend from `web/index.html` next to the EXE.
- Injects `window.dnfNative` for the frontend bridge.
- Implements window commands, saved login storage, local background loading, URL opening, DNF.exe launch/stop/status, and Script.pvf MD5 verification.
- Implements rapid-fire commands through the Interception runtime embedded in the EXE.

Expected runtime layout:

```text
launcher.exe
web/
  index.html
  assets/
start/
  backgrounds/
DNF.exe
Script.pvf
```

Build prerequisites:

- Visual Studio Build Tools with MSVC C++.
- CMake.
- Microsoft Edge WebView2 Runtime.
- Microsoft.Web.WebView2 NuGet package.

Restore WebView2 SDK:

```powershell
dotnet restore cpp_launcher\WebView2SdkRestore.csproj
```

Example build:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build-cpp-launcher.ps1
```

For C++ only:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build-cpp-launcher.ps1 -SkipFrontend
```

The build script copies `desktop_launcher/dist` to the EXE directory as `web` when the frontend build is not skipped. `cpp_launcher/assets/interception` is compiled into the EXE as resources; at runtime those files are extracted to `%LOCALAPPDATA%\DNFLauncher\Interception` only when the rapid-fire feature needs them.
