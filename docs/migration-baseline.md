# C++ WebView2 + Go Migration Baseline

This document freezes the current behavior before replacing the implementation
layers. The migration target is C++ + WebView2 for the Windows launcher and Go
for the server, while keeping the visible UI and user-facing behavior unchanged.

## Current Project

- Windows launcher: `desktop_launcher`
- Current launcher stack: Tauri 2 + Vite + HTML/CSS/JavaScript
- Current server: `server`
- Current server stack: Python + FastAPI
- Current server config example: `server/config.example.json`
- Current build-time API setting: `build-launcher.ps1` sets `DNF_LAUNCHER_API_BASE`

## Non-Negotiable Compatibility

- Keep the current launcher UI layout, pages, text, controls, and workflows.
- Reuse the existing HTML/CSS/JavaScript in WebView2 instead of redesigning a
  native UI.
- Keep login, registration, remembered login, announcements, poster carousel,
  game launch, PVF MD5 check, GM tools, permission controls, and audit/log pages.
- Keep backend API behavior compatible so the existing frontend can be used as
  the test client during server migration.
- Do not add a self-developed companion `DNF.exe` or DLL in this migration.
- Do not write, generate, or overwrite external game client configuration files.

## IP Boundaries

There are two different IP/configuration concepts:

- Launcher backend API address: compiled into the launcher build through
  `DNF_LAUNCHER_API_BASE`.
- DNF game server address: maintained by deployment staff in the external game
  client config file.

For the current selected existing client, the only game-client field the
launcher may read is:

```ini
[登录配置]
服务器IP=...
```

The file is GBK/ANSI encoded. The launcher may validate that this key exists and
has a plausible IP/host value, but it must not modify the file.

## Existing Frontend API Calls

The current frontend calls these server endpoints and the Go server must remain
compatible with them:

- `GET /health`
- `GET /api/settings`
- `GET /api/posters/{filename}`
- `POST /api/auth/login`
- `POST /api/auth/register`
- `POST /api/auth/change-password`
- `POST /api/auth/admin/change-password`
- `GET /api/pvf/status`
- `GET /api/pvf/items`
- `POST /api/admin/pvf/refresh`
- `PUT /api/admin/pvf/client-md5`
- `GET /api/gm/characters`
- `POST /api/gm/account/resolve`
- `GET /api/gm/character/job-options`
- `POST /api/gm/character/level`
- `POST /api/gm/character/pvp-grade`
- `POST /api/gm/character/pvp-point`
- `POST /api/gm/character/job`
- `POST /api/gm/character/delete`
- `POST /api/gm/character/recover`
- `POST /api/gm/mail/send`
- `POST /api/gm/mail/send-all`
- `POST /api/gm/mail/delete`
- `POST /api/gm/mail/delete-all`
- `POST /api/gm/inventory/query`
- `POST /api/gm/inventory/delete`
- `POST /api/gm/inventory/clear`
- `GET /api/gm/avatar/options`
- `POST /api/gm/avatar/query`
- `POST /api/gm/avatar/hidden`
- `POST /api/gm/cera/query`
- `POST /api/gm/cera/charge`
- `GET /api/gm/events`
- `POST /api/gm/events`
- `DELETE /api/gm/events/{log_id}`
- `POST /api/gm/ban/query`
- `POST /api/gm/ban/set`
- `POST /api/gm/ban/unban`
- `GET /api/admin/permissions`
- `GET /api/admin/accounts`
- `PUT /api/admin/accounts/{uid}/permissions`
- `GET /api/admin/logs`
- `PUT /api/admin/settings/home`
- `POST /api/launcher/direct`

## Launcher Native Commands To Preserve

The WebView2 launcher must provide native equivalents for the current Tauri
commands used by `desktop_launcher/src/main.js`:

- `get_launcher_window_title`
- `get_launcher_background`
- `load_saved_login`
- `save_saved_login`
- `clear_saved_login`
- `is_game_running`
- `launch_game`
- `stop_game`
- `open_url`
- `install_interception_driver`
- `list_rapid_fire`
- `add_rapid_fire`
- `remove_rapid_fire`

## Go Migration Strategy

Phase 1:

- Add a new Go server beside the existing Python server.
- Keep the Python server untouched as the behavior reference.
- Implement config loading compatible with `server/config.example.json`.
- Implement `/health` and API route skeletons.
- Return explicit `501 Not Implemented` for endpoints not yet migrated.

Phase 2:

- Port authentication, session tokens, settings, permissions, and audit logs.
- Use the existing frontend as the compatibility test client.

Phase 3:

- Port GM database operations and PVF cache/search logic.
- Keep SQL behavior compatible with the Python implementation.

Phase 4:

- Replace the launcher shell with C++ + WebView2 while reusing the current
  web UI assets.
- Add read-only `muqing.ini` server-IP validation in the launch path.

## Generic Source Release Rules

- Do not commit real server IPs.
- Do not commit database passwords.
- Do not commit production `config.json`.
- Do not commit external game client binaries or `muqing.ini`.
- Keep `build-launcher.ps1` or an example build script using placeholder API
  addresses for public source.
