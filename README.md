# DNF 桌面启动器与管理 API

本项目由 Windows 桌面启动器和 Linux 纯 API 服务端组成。

技术栈：

- 桌面端：Tauri 2、Rust、Vite、原生 HTML/CSS/JavaScript
- 服务端：Python 3、FastAPI、Uvicorn、PyMySQL、OpenSSL
- 数据库：游戏库 `d_taiwan`，工具库 `dnf_launcher`

## 目录

```text
desktop_launcher/          Tauri 桌面启动器
server/                    Linux API 服务端
  run.py                   服务入口
  config.example.json      配置模板
  requirements.txt
  app/
    main.py                ASGI 入口
    api.py                 API 路由与现有 GM 业务实现
    models.py              请求模型
    core/                  配置、数据库、权限、会话安全
    services/              登录、令牌、启动、PVF 服务
  data/
    posters/               公告海报文件
```

## 服务端部署

将 `server` 上传到 `/opt/server`。服务端需要 `openssl` 命令用于生成 DNF 登录令牌，登录密钥私钥也放在该目录：

生成一组新的 2048 位 RSA 登录密钥对：

```bash
cd /opt/server
openssl genrsa -out privatekey.pem 2048
openssl rsa -in privatekey.pem -pubout -out publickey.pem
chmod 600 privatekey.pem
```

生成后：

- `privatekey.pem` 留在服务端，并确保 `config.json` 的 `login_private_key_path` 指向它。
- `publickey.pem` 需要同步到游戏端登录校验使用的位置。
- 每次更换密钥对，都必须同时更新服务端私钥和游戏端公钥。
- 只替换其中一端会导致 DNF 登录令牌校验失败。

创建配置并启动：

```bash
cd /opt/server
cp config.example.json config.json
python3 -m pip install -r requirements.txt
nohup python3 run.py > /opt/server/server.log 2>&1 < /dev/null &
```

服务端仅开放桌面端所需接口：

- HTTP API：`0.0.0.0:8000`
- 健康检查：`GET /health`

首次启动会创建 `dnf_launcher` 工具库及下列表：

```text
admins
roles
account_roles
operation_logs
settings
pvf_meta
pvf_items
pvf_data
```

默认管理员为 `admin/admin`，部署后应立即在桌面端修改密码。

关键配置项：

```json
{
  "server_host": "游戏数据库IP",
  "db_user": "game",
  "db_password": "数据库密码",
  "game_db_name": "d_taiwan",
  "tool_db_name": "dnf_launcher",
  "session_secret": "发布时改成随机长字符串",
  "login_private_key_path": "/opt/server/privatekey.pem"
}
```

## 公告海报

把海报文件放入：

```text
/opt/server/data/posters/
```

支持 JPG、JPEG、PNG、WebP。GM 端公告海报地址填写：

```text
/api/posters/example.jpg
```

也可以填写完整的 `http://` 或 `https://` 图片地址。建议尺寸为 1120×480。

## 客户端目录

启动器和 `DNF.exe` 位于客户端根目录，随机底图固定读取根目录下的 `start/backgrounds`：

```text
DNF客户端/
  启动器.exe
  DNF.exe
  start/
    backgrounds/
      01.jpg
      02.webp
    Interception/
      command line installer/
        install-interception.exe
      library/
        x64/
          interception.dll
```

支持 JPG、JPEG、PNG、WebP。每次启动随机选择一张；目录不存在、为空或图片无效时使用内置底图。建议使用 1920×1080 图片，单张不超过 20MB。

按键连发配置固定保存到 `%LOCALAPPDATA%\DNFLauncher\rapid-fire.json`，不依赖启动器所在目录写权限。当前连发实现使用 Interception 驱动层输入，仅在 `DNF.exe` 前台运行时生效。

将官方 `Interception.zip` 解压后的完整 `Interception` 目录放到客户端 `start` 目录下：

```text
DNF客户端/start/Interception/
```

启动器会固定查找：

```text
DNF客户端/start/Interception/command line installer/install-interception.exe
DNF客户端/start/Interception/library/x64/interception.dll
```

如果检测到安装文件但驱动未就绪，按键连发页会显示“安装驱动”按钮。点击后会请求管理员权限并启动安装程序。安装完成后需要重启电脑；驱动状态正常后不再显示安装按钮。

## Windows 编译环境

编译 EXE 需要完整 Tauri Windows 编译环境。

必需组件：

- Node.js LTS，包含 `npm`
- Rust stable，包含 `cargo`
- Visual Studio Build Tools
  - Desktop development with C++
  - MSVC C++ build tools
  - Windows 10/11 SDK
- Microsoft Edge WebView2 Runtime

推荐安装方式：

```powershell
winget install OpenJS.NodeJS.LTS
winget install Rustlang.Rustup
winget install Microsoft.VisualStudio.2022.BuildTools
winget install Microsoft.EdgeWebView2Runtime
rustup default stable
```

安装 Visual Studio Build Tools 时，需要在安装器中勾选“使用 C++ 的桌面开发”。如果只安装 `npm`，无法编译 Tauri 的 Rust 桌面壳。

首次拉取或清空依赖后，安装前端依赖：

```powershell
cd desktop_launcher
npm install
```

## 桌面端开发与构建

开发运行：

```powershell
cd desktop_launcher
$env:DNF_LAUNCHER_API_BASE="http://服务器IP:8000"
npm run tauri -- dev
```

构建 EXE：

```powershell
# 修改 build-launcher.ps1 顶部的 $ApiBase 后执行
.\build-launcher.ps1
```

构建产物：

```text
desktop_launcher/src-tauri/target/release/dnf-desktop-launcher.exe
```

未设置 `DNF_LAUNCHER_API_BASE`，或者不是 `http://` / `https://` 开头时，构建会直接失败。

最终发布时，将该 EXE 改名后放入客户端根目录即可。启动器会查找同目录下的 `DNF.exe`。
