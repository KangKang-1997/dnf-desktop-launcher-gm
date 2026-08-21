# Desktop WebView2 Architecture

## Scope

The DNF Taiwan Windows launcher is a C++ WebView2 host that loads the web
frontend from `web/index.html`. The frontend is plain Vite HTML/CSS/JavaScript
and talks to native code through `window.dnfNative`.

## Native Bridge

The frontend imports `desktop_launcher/src/native/bridge.js` and calls:

- `native.invoke(command, args)`
- `native.setWindowTitle(title)`
- `native.minimizeWindow()`
- `native.closeWindow()`
- `native.startWindowDrag()`
- `native.revealWindow()`

The C++ host injects `window.dnfNative` before page scripts run. Requests and
responses use WebView2 web messages:

```json
{
  "type": "native-result",
  "id": 1,
  "ok": true,
  "result": null
}
```

Errors use:

```json
{
  "type": "native-result",
  "id": 1,
  "ok": false,
  "error": "message"
}
```

## Native Commands

- Window title, minimize, close, drag, and delayed reveal.
- `get_launcher_window_title`.
- `get_launcher_background` from `start/backgrounds`.
- `save_saved_login`, `load_saved_login`, `clear_saved_login` using Windows DPAPI.
- `open_url`.
- `launch_game`, `is_game_running`, `stop_game`.
- `Script.pvf` MD5 validation before launch.
- `list_rapid_fire`, `add_rapid_fire`, `remove_rapid_fire`.
- `install_interception_driver`.

## Rapid Fire

Interception files are embedded into the EXE as Windows resources:

- `cpp_launcher/assets/interception/install-interception.exe`
- `cpp_launcher/assets/interception/x64/interception.dll`

At runtime they are extracted on demand to:

```text
%LOCALAPPDATA%\DNFLauncher\Interception\
```

If the driver is missing or not active, the rapid-fire page shows an install
button. Clicking it starts the embedded installer with administrator privileges.
The user must restart Windows after driver installation.

## Expected Runtime Layout

```text
dnf-webview2-launcher.exe
web/
  index.html
  assets/
start/
  backgrounds/
DNF.exe
Script.pvf
```
