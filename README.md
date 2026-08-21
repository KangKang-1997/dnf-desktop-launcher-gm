# DNF Taiwan Desktop Launcher GM

这是一个面向 DNF 台服管理场景的桌面登录器和 GM 管理工具。

- 服务端：Go HTTP API
- 桌面端：C++ WebView2 原生窗口
- 前端：Vite + HTML/CSS/JavaScript

## 目录结构

```text
go_server/                 Go 服务端源码
desktop_launcher/          Web 前端源码
cpp_launcher/              C++ WebView2 桌面壳
build-cpp-launcher.ps1     Windows 桌面端构建脚本
```

关键内置资源：

```text
cpp_launcher/assets/app.ico
desktop_launcher/public/assets/background.jpg
cpp_launcher/assets/interception/install-interception.exe
cpp_launcher/assets/interception/x64/interception.dll
```

桌面端 EXE 会内置前端页面、默认背景图、Interception 安装器和 DLL。用户环境缺失驱动时，可在“按键连发”页面点击“安装驱动”，程序会释放内置安装器并请求管理员权限执行。驱动安装后通常需要重启 Windows。

## 服务端

服务端配置文件固定从二进制同目录读取：

```text
config.json
```

配置模板：

```text
go_server/config.example.json
```

常用字段：

```json
{
  "listen_host": "0.0.0.0",
  "listen_port": 8000,
  "server_host": "127.0.0.1",
  "db_port": 3306,
  "db_user": "数据库用户名",
  "db_password": "数据库密码",
  "game_db_name": "d_taiwan",
  "tool_db_name": "dnf_launcher",
  "db_charset": "utf8",
  "session_secret": "发布时改成随机长字符串",
  "session_ttl_seconds": 86400,
  "cors_origins": ["*"],
  "login_private_key_path": "/home/neople/game/privatekey.pem"
}
```

开发运行：

```powershell
cd go_server
copy config.example.json config.json
go run .\cmd\dnf-server
```

Linux 构建示例：

```powershell
cd go_server
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o .\dist\dnf-server-linux-amd64 .\cmd\dnf-server
```

部署时将服务端二进制、`config.json`、`privatekey.pem` 放到服务端目录。MySQL 与服务端在同机时，`server_host` 可以继续使用 `127.0.0.1`。

Linux 生成登录密钥：

```bash
sudo mkdir -p /home/neople/game
sudo openssl genrsa -out /home/neople/game/privatekey.pem 1024
sudo openssl rsa -in /home/neople/game/privatekey.pem -pubout -out /home/neople/game/publickey.pem
sudo chmod 600 /home/neople/game/privatekey.pem
sudo chmod 644 /home/neople/game/publickey.pem
```

海报文件从服务端二进制同目录读取：

```text
posters/
```

默认公告图片引用为 `/api/posters/sample-*`。需要展示图片时，由维护人员在 `posters/` 中放入匹配文件，也可以在后台改成实际海报地址。

## 桌面端

桌面端构建产物位于：

```text
cpp_launcher/build-ninja/dnf-webview2-launcher.exe
```

构建依赖：

- Windows 10/11
- Visual Studio Build Tools，安装“使用 C++ 的桌面开发”
- CMake
- Ninja
- .NET SDK，用于恢复 WebView2 SDK
- Node.js 和 npm
- Microsoft Edge WebView2 Runtime

构建命令：

```powershell
$env:DNF_LAUNCHER_API_BASE="http://127.0.0.1:8000"
.\build-cpp-launcher.ps1
```

只重编 C++ 壳：

```powershell
.\build-cpp-launcher.ps1 -SkipFrontend
```

构建脚本会：

- 安装或检查前端 npm 依赖
- 使用 `DNF_LAUNCHER_API_BASE` 写入前端 API 地址
- 构建 Vite 前端
- 将前端页面和默认背景图编译进 C++ 资源
- 构建 C++ WebView2 EXE

## 客户端部署布局

实际客户端目录建议保持平铺：

```text
dnf-webview2-launcher.exe
DNF.exe
Script.pvf
```

说明：

- `DNF.exe` 和 `Script.pvf` 与登录器 EXE 同目录。
- 登录器前端和默认背景图已编译进 EXE，客户端不需要 `web/` 或背景图目录。
- 按键连发配置保存到 `%LOCALAPPDATA%\DNFLauncher\rapid-fire.json`。
- 内置 Interception 文件按需释放到 `%LOCALAPPDATA%\DNFLauncher\Interception`。
