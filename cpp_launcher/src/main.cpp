#include <windows.h>
#include <shellapi.h>
#include <shlwapi.h>
#include <wincrypt.h>
#include <wrl.h>
#include <WebView2.h>
#include <WebView2EnvironmentOptions.h>

#include "resource.h"

#include <algorithm>
#include <atomic>
#include <chrono>
#include <cstdlib>
#include <cwctype>
#include <cstdint>
#include <filesystem>
#include <fstream>
#include <map>
#include <optional>
#include <random>
#include <sstream>
#include <stdexcept>
#include <cstdio>
#include <string>
#include <mutex>
#include <thread>
#include <unordered_map>
#include <unordered_set>
#include <vector>

using Microsoft::WRL::Callback;
using Microsoft::WRL::ComPtr;

namespace fs = std::filesystem;

namespace {

constexpr wchar_t kWindowClassName[] = L"DNFWebView2LauncherWindow";
constexpr wchar_t kWindowTitle[] = L"地下城与勇士";
constexpr wchar_t kClientPVFMismatchError[] = L"CLIENT_PVF_MISMATCH";
constexpr wchar_t kRapidFireConfigFileName[] = L"rapid-fire.json";
constexpr unsigned long long kRapidFireMinIntervalMs = 1;
constexpr unsigned long long kRapidFireMaxIntervalMs = 10000;
constexpr unsigned long long kRapidFirePressDurationMs = 1;
constexpr unsigned short kInterceptionFilterKeyDown = 1;
constexpr unsigned short kInterceptionFilterKeyUp = 2;
constexpr unsigned short kInterceptionKeyDown = 0;
constexpr unsigned short kInterceptionKeyUp = 1;
constexpr int kInterceptionMaxKeyboard = 10;

HWND g_mainWindow = nullptr;
ComPtr<ICoreWebView2Controller> g_controller;
ComPtr<ICoreWebView2> g_webview;
PROCESS_INFORMATION g_gameProcess{};
bool g_hasGameProcess = false;
bool g_windowRevealed = false;
std::mutex g_logMutex;
std::atomic<DWORD> g_targetPid{0};

struct RapidFireConfig {
    std::wstring key;
    unsigned long long intervalMs = 0;
    unsigned short scanCode = 0;
};

std::mutex g_rapidConfigMutex;
std::mutex g_rapidMutationMutex;
std::mutex g_pressedKeysMutex;
std::mutex g_hookErrorMutex;
std::vector<RapidFireConfig> g_rapidConfigs;
std::unordered_set<unsigned short> g_pressedKeys;
std::atomic_bool g_hookReady{false};
std::wstring g_hookError;
std::once_flag g_rapidInitOnce;

int ScaleForDpi(int value, UINT dpi) {
    return MulDiv(value, static_cast<int>(dpi), USER_DEFAULT_SCREEN_DPI);
}

void SetMainWindowOpacity(BYTE alpha) {
    if (!g_mainWindow) return;
    SetLayeredWindowAttributes(g_mainWindow, 0, alpha, LWA_ALPHA);
}

void RevealMainWindow(bool activate) {
    if (!g_windowRevealed) {
        SetMainWindowOpacity(255);
        g_windowRevealed = true;
    }
    ShowWindow(g_mainWindow, SW_SHOW);
    if (activate) SetForegroundWindow(g_mainWindow);
}

std::wstring Utf8ToWide(const std::string& value) {
    if (value.empty()) return L"";
    int length = MultiByteToWideChar(CP_UTF8, 0, value.data(), static_cast<int>(value.size()), nullptr, 0);
    std::wstring result(length, L'\0');
    MultiByteToWideChar(CP_UTF8, 0, value.data(), static_cast<int>(value.size()), result.data(), length);
    return result;
}

std::string WideToUtf8(const std::wstring& value) {
    if (value.empty()) return "";
    int length = WideCharToMultiByte(CP_UTF8, 0, value.data(), static_cast<int>(value.size()), nullptr, 0, nullptr, nullptr);
    std::string result(length, '\0');
    WideCharToMultiByte(CP_UTF8, 0, value.data(), static_cast<int>(value.size()), result.data(), length, nullptr, nullptr);
    return result;
}

std::wstring ExeDirectory() {
    std::wstring buffer(MAX_PATH, L'\0');
    DWORD length = GetModuleFileNameW(nullptr, buffer.data(), static_cast<DWORD>(buffer.size()));
    while (length == buffer.size()) {
        buffer.resize(buffer.size() * 2);
        length = GetModuleFileNameW(nullptr, buffer.data(), static_cast<DWORD>(buffer.size()));
    }
    buffer.resize(length);
    return fs::path(buffer).parent_path().wstring();
}

void LogLine(const std::wstring& message) {
    std::lock_guard<std::mutex> lock(g_logMutex);
    std::ofstream file(fs::path(ExeDirectory()) / L"launcher.log", std::ios::binary | std::ios::app);
    if (!file) return;
    SYSTEMTIME now{};
    GetLocalTime(&now);
    char timestamp[32]{};
    sprintf_s(
        timestamp,
        "[%04u-%02u-%02u %02u:%02u:%02u] ",
        now.wYear,
        now.wMonth,
        now.wDay,
        now.wHour,
        now.wMinute,
        now.wSecond);
    file << timestamp << WideToUtf8(message) << "\r\n";
}

std::wstring HResultText(HRESULT result) {
    wchar_t buffer[32]{};
    swprintf_s(buffer, L"0x%08X", static_cast<unsigned int>(result));
    return buffer;
}

std::wstring JsonEscape(const std::wstring& value) {
    std::wstring result;
    result.reserve(value.size() + 8);
    for (wchar_t ch : value) {
        switch (ch) {
        case L'\\': result += L"\\\\"; break;
        case L'"': result += L"\\\""; break;
        case L'\b': result += L"\\b"; break;
        case L'\f': result += L"\\f"; break;
        case L'\n': result += L"\\n"; break;
        case L'\r': result += L"\\r"; break;
        case L'\t': result += L"\\t"; break;
        default:
            if (ch < 0x20) {
                wchar_t buffer[7]{};
                swprintf_s(buffer, L"\\u%04x", ch);
                result += buffer;
            } else {
                result += ch;
            }
            break;
        }
    }
    return result;
}

std::wstring JsonString(const std::wstring& value) {
    return L"\"" + JsonEscape(value) + L"\"";
}

std::vector<unsigned char> ResourceBytes(int resourceId) {
    HRSRC resource = FindResourceW(nullptr, MAKEINTRESOURCEW(resourceId), RT_RCDATA);
    if (!resource) return {};
    HGLOBAL loaded = LoadResource(nullptr, resource);
    if (!loaded) throw std::runtime_error("加载内置资源失败");
    DWORD size = SizeofResource(nullptr, resource);
    const void* data = LockResource(loaded);
    if (!data || size == 0) throw std::runtime_error("读取内置资源失败");
    const auto* begin = static_cast<const unsigned char*>(data);
    return std::vector<unsigned char>(begin, begin + size);
}

std::wstring EmbeddedFrontendHtml() {
    std::vector<unsigned char> bytes = ResourceBytes(IDR_FRONTEND_HTML);
    if (bytes.empty()) return L"";
    return Utf8ToWide(std::string(reinterpret_cast<const char*>(bytes.data()), bytes.size()));
}

std::optional<std::wstring> ExtractJsonString(const std::wstring& json, const std::wstring& key) {
    const std::wstring marker = L"\"" + key + L"\":";
    size_t pos = json.find(marker);
    if (pos == std::wstring::npos) return std::nullopt;
    pos += marker.size();
    while (pos < json.size() && iswspace(json[pos])) ++pos;
    if (pos >= json.size() || json[pos] != L'"') return std::nullopt;
    ++pos;
    std::wstring result;
    while (pos < json.size()) {
        wchar_t ch = json[pos++];
        if (ch == L'"') return result;
        if (ch != L'\\') {
            result += ch;
            continue;
        }
        if (pos >= json.size()) break;
        wchar_t escaped = json[pos++];
        switch (escaped) {
        case L'"': result += L'"'; break;
        case L'\\': result += L'\\'; break;
        case L'/': result += L'/'; break;
        case L'b': result += L'\b'; break;
        case L'f': result += L'\f'; break;
        case L'n': result += L'\n'; break;
        case L'r': result += L'\r'; break;
        case L't': result += L'\t'; break;
        default: result += escaped; break;
        }
    }
    return std::nullopt;
}

std::optional<long long> ExtractJsonInteger(const std::wstring& json, const std::wstring& key) {
    const std::wstring marker = L"\"" + key + L"\":";
    size_t pos = json.find(marker);
    if (pos == std::wstring::npos) return std::nullopt;
    pos += marker.size();
    while (pos < json.size() && iswspace(json[pos])) ++pos;
    size_t start = pos;
    if (pos < json.size() && json[pos] == L'-') ++pos;
    while (pos < json.size() && iswdigit(json[pos])) ++pos;
    if (pos == start) return std::nullopt;
    return _wtoi64(json.substr(start, pos - start).c_str());
}

void PostNativeResult(long long id, bool ok, const std::wstring& payloadJson) {
    if (!g_webview) return;
    std::wstring message = L"{\"type\":\"native-result\",\"id\":" + std::to_wstring(id) +
        L",\"ok\":" + (ok ? L"true" : L"false") +
        (ok ? L",\"result\":" : L",\"error\":") + payloadJson + L"}";
    g_webview->PostWebMessageAsJson(message.c_str());
}

std::string Base64Encode(const std::vector<unsigned char>& data) {
    if (data.empty()) return "";
    DWORD length = 0;
    CryptBinaryToStringA(data.data(), static_cast<DWORD>(data.size()), CRYPT_STRING_BASE64 | CRYPT_STRING_NOCRLF, nullptr, &length);
    std::string result(length, '\0');
    if (!CryptBinaryToStringA(data.data(), static_cast<DWORD>(data.size()), CRYPT_STRING_BASE64 | CRYPT_STRING_NOCRLF, result.data(), &length)) {
        return "";
    }
    if (!result.empty() && result.back() == '\0') result.pop_back();
    return result;
}

std::vector<unsigned char> ReadFileBytes(const fs::path& path, size_t maxBytes = 0) {
    std::ifstream file(path, std::ios::binary);
    if (!file) return {};
    file.seekg(0, std::ios::end);
    std::streamoff size = file.tellg();
    if (size <= 0 || (maxBytes > 0 && static_cast<size_t>(size) > maxBytes)) return {};
    file.seekg(0, std::ios::beg);
    std::vector<unsigned char> bytes(static_cast<size_t>(size));
    file.read(reinterpret_cast<char*>(bytes.data()), size);
    return bytes;
}

std::wstring LauncherBackground() {
    std::vector<unsigned char> bytes = ResourceBytes(IDR_DEFAULT_BACKGROUND);
    if (bytes.empty()) return L"null";
    std::string dataUrl = "data:image/jpeg;base64," + Base64Encode(bytes);
    return JsonString(Utf8ToWide(dataUrl));
}

fs::path SavedLoginPath() {
    wchar_t* localAppData = nullptr;
    size_t length = 0;
    _wdupenv_s(&localAppData, &length, L"LOCALAPPDATA");
    fs::path base = localAppData ? fs::path(localAppData) : fs::path(ExeDirectory());
    free(localAppData);
    fs::path directory = base / L"DNFLauncher";
    fs::create_directories(directory);
    return directory / L"saved-login.bin";
}

std::vector<unsigned char> ProtectBytes(const std::string& payload) {
    DATA_BLOB input{};
    input.pbData = reinterpret_cast<BYTE*>(const_cast<char*>(payload.data()));
    input.cbData = static_cast<DWORD>(payload.size());
    DATA_BLOB output{};
    if (!CryptProtectData(&input, L"DNFLauncherSavedLogin", nullptr, nullptr, nullptr, 0, &output)) {
        throw std::runtime_error("保存密码失败");
    }
    std::vector<unsigned char> result(output.pbData, output.pbData + output.cbData);
    LocalFree(output.pbData);
    return result;
}

std::string UnprotectBytes(const std::vector<unsigned char>& payload) {
    if (payload.empty()) return "";
    DATA_BLOB input{};
    input.pbData = const_cast<BYTE*>(payload.data());
    input.cbData = static_cast<DWORD>(payload.size());
    DATA_BLOB output{};
    if (!CryptUnprotectData(&input, nullptr, nullptr, nullptr, nullptr, 0, &output)) {
        return "";
    }
    std::string result(reinterpret_cast<char*>(output.pbData), reinterpret_cast<char*>(output.pbData + output.cbData));
    LocalFree(output.pbData);
    return result;
}

void SaveSavedLogin(const std::wstring& account, const std::wstring& password) {
    std::ofstream file(SavedLoginPath(), std::ios::binary | std::ios::trunc);
    std::string payload = WideToUtf8(account) + "\n" + WideToUtf8(password);
    std::vector<unsigned char> protectedPayload = ProtectBytes(payload);
    file.write(reinterpret_cast<const char*>(protectedPayload.data()), static_cast<std::streamsize>(protectedPayload.size()));
}

std::wstring LoadSavedLogin() {
    std::vector<unsigned char> protectedPayload = ReadFileBytes(SavedLoginPath());
    if (protectedPayload.empty()) return L"null";
    std::string payload = UnprotectBytes(protectedPayload);
    size_t split = payload.find('\n');
    if (split == std::string::npos) return L"null";
    std::wstring account = Utf8ToWide(payload.substr(0, split));
    std::wstring password = Utf8ToWide(payload.substr(split + 1));
    return L"{\"account\":" + JsonString(account) + L",\"password\":" + JsonString(password) + L"}";
}

void ClearSavedLogin() {
    std::error_code ignored;
    fs::remove(SavedLoginPath(), ignored);
}

std::optional<fs::path> FindGameExecutable() {
    const fs::path directory = ExeDirectory();
    for (const wchar_t* name : {L"DNF.exe", L"dnf.exe"}) {
        fs::path candidate = directory / name;
        if (fs::is_regular_file(candidate)) return candidate;
    }
    return std::nullopt;
}

std::wstring FileMD5(const fs::path& path) {
    HCRYPTPROV provider = 0;
    HCRYPTHASH hash = 0;
    if (!CryptAcquireContextW(&provider, nullptr, nullptr, PROV_RSA_FULL, CRYPT_VERIFYCONTEXT)) return L"";
    if (!CryptCreateHash(provider, CALG_MD5, 0, 0, &hash)) {
        CryptReleaseContext(provider, 0);
        return L"";
    }
    std::ifstream file(path, std::ios::binary);
    std::vector<unsigned char> buffer(1024 * 1024);
    while (file) {
        file.read(reinterpret_cast<char*>(buffer.data()), static_cast<std::streamsize>(buffer.size()));
        std::streamsize size = file.gcount();
        if (size > 0) CryptHashData(hash, buffer.data(), static_cast<DWORD>(size), 0);
    }
    BYTE digest[16]{};
    DWORD digestSize = sizeof(digest);
    std::wstring result;
    if (CryptGetHashParam(hash, HP_HASHVAL, digest, &digestSize, 0)) {
        wchar_t hex[3]{};
        for (DWORD i = 0; i < digestSize; ++i) {
            swprintf_s(hex, L"%02X", digest[i]);
            result += hex;
        }
    }
    CryptDestroyHash(hash);
    CryptReleaseContext(provider, 0);
    return result;
}

void VerifyClientPVF(const std::wstring& expectedMD5) {
    std::wstring expected = expectedMD5;
    expected.erase(std::remove_if(expected.begin(), expected.end(), iswspace), expected.end());
    std::transform(expected.begin(), expected.end(), expected.begin(), towupper);
    if (expected.empty()) return;
    fs::path path = fs::path(ExeDirectory()) / L"Script.pvf";
    if (!fs::is_regular_file(path)) {
        throw std::runtime_error(WideToUtf8(std::wstring(kClientPVFMismatchError) + L":客户端 Script.pvf 不存在"));
    }
    if (FileMD5(path) != expected) {
        throw std::runtime_error(WideToUtf8(std::wstring(kClientPVFMismatchError) + L":客户端 Script.pvf 校验失败"));
    }
}

bool IsGameRunning() {
    if (!g_hasGameProcess) return false;
    DWORD code = 0;
    if (!GetExitCodeProcess(g_gameProcess.hProcess, &code) || code != STILL_ACTIVE) {
        CloseHandle(g_gameProcess.hProcess);
        CloseHandle(g_gameProcess.hThread);
    g_gameProcess = {};
    g_hasGameProcess = false;
        g_targetPid.store(0, std::memory_order_release);
        return false;
    }
    return true;
}

void StopGame() {
    if (!IsGameRunning()) {
        throw std::runtime_error("DNF.exe 未运行");
    }
    TerminateProcess(g_gameProcess.hProcess, 0);
    WaitForSingleObject(g_gameProcess.hProcess, 5000);
    IsGameRunning();
}

void LaunchGame(const std::wstring& token, const std::wstring& expectedMD5) {
    if (token.size() < 16 || token.size() > 4096) {
        throw std::runtime_error("服务器返回的 DNF 登录参数无效");
    }
    for (wchar_t ch : token) {
        if (!(iswalnum(ch) || ch == L'+' || ch == L'/' || ch == L'=')) {
            throw std::runtime_error("服务器返回的 DNF 登录参数无效");
        }
    }
    VerifyClientPVF(expectedMD5);
    if (IsGameRunning()) {
        throw std::runtime_error("DNF.exe 已在运行");
    }
    auto game = FindGameExecutable();
    if (!game) {
        throw std::runtime_error("未在登录器同目录找到 DNF.exe");
    }
    std::wstring commandLine = L"\"" + game->wstring() + L"\" " + token;
    STARTUPINFOW startup{};
    startup.cb = sizeof(startup);
    PROCESS_INFORMATION process{};
    if (!CreateProcessW(
            game->c_str(),
            commandLine.data(),
            nullptr,
            nullptr,
            FALSE,
            CREATE_NO_WINDOW,
            nullptr,
            fs::path(ExeDirectory()).c_str(),
            &startup,
            &process)) {
        throw std::runtime_error("启动 DNF.exe 失败");
    }
    g_gameProcess = process;
    g_hasGameProcess = true;
    g_targetPid.store(process.dwProcessId, std::memory_order_release);
}

void OpenUrl(const std::wstring& url) {
    if (!(url.rfind(L"http://", 0) == 0 || url.rfind(L"https://", 0) == 0)) {
        throw std::runtime_error("下载地址无效");
    }
    ShellExecuteW(g_mainWindow, L"open", url.c_str(), nullptr, nullptr, SW_SHOWNORMAL);
}

using InterceptionDevice = int;
using InterceptionContext = void*;
using CreateContextFn = InterceptionContext(__cdecl*)();
using DestroyContextFn = void(__cdecl*)(InterceptionContext);
using SetFilterFn = void(__cdecl*)(InterceptionContext, int(__cdecl*)(InterceptionDevice), unsigned short);
using WaitWithTimeoutFn = InterceptionDevice(__cdecl*)(InterceptionContext, unsigned int);

struct InterceptionKeyStroke {
    unsigned short code;
    unsigned short state;
    unsigned int information;
};

using SendFn = int(__cdecl*)(InterceptionContext, InterceptionDevice, const InterceptionKeyStroke*, unsigned int);
using ReceiveFn = int(__cdecl*)(InterceptionContext, InterceptionDevice, InterceptionKeyStroke*, unsigned int);

bool IsKeyboardDeviceId(InterceptionDevice device) {
    return device >= 1 && device <= kInterceptionMaxKeyboard;
}

int __cdecl IsKeyboardDevice(InterceptionDevice device) {
    return IsKeyboardDeviceId(device) ? 1 : 0;
}

fs::path LauncherDataDirectory() {
    DWORD length = GetEnvironmentVariableW(L"LOCALAPPDATA", nullptr, 0);
    if (length == 0) throw std::runtime_error("无法读取 LOCALAPPDATA 目录");
    std::wstring localAppData(length, L'\0');
    GetEnvironmentVariableW(L"LOCALAPPDATA", localAppData.data(), length);
    if (!localAppData.empty() && localAppData.back() == L'\0') localAppData.pop_back();
    fs::path directory = fs::path(localAppData) / L"DNFLauncher";
    std::error_code error;
    fs::create_directories(directory, error);
    if (error) throw std::runtime_error("创建启动器数据目录失败");
    return directory;
}

fs::path RapidFireConfigPath() {
    fs::path directory = LauncherDataDirectory();
    return directory / kRapidFireConfigFileName;
}

std::optional<fs::path> ExtractResourceFile(int resourceId, const fs::path& relativePath) {
    std::vector<unsigned char> bytes = ResourceBytes(resourceId);
    if (bytes.empty()) return std::nullopt;

    fs::path output = LauncherDataDirectory() / relativePath;
    std::error_code error;
    fs::create_directories(output.parent_path(), error);
    if (error) throw std::runtime_error("创建内置资源释放目录失败");

    if (fs::is_regular_file(output, error)) {
        error.clear();
        if (fs::file_size(output, error) == static_cast<std::uintmax_t>(bytes.size()) && !error) {
            return output;
        }
    }

    std::ofstream file(output, std::ios::binary | std::ios::trunc);
    if (!file) throw std::runtime_error("写入内置资源失败");
    file.write(reinterpret_cast<const char*>(bytes.data()), static_cast<std::streamsize>(bytes.size()));
    if (!file) throw std::runtime_error("写入内置资源失败");
    return output;
}

fs::path InterceptionRootPath() {
    return fs::path(ExeDirectory()) / L"Interception";
}

std::wstring InterceptionInstallerHint() {
    return L"%LOCALAPPDATA%\\DNFLauncher\\Interception\\command line installer\\install-interception.exe";
}

std::wstring InterceptionDllHint() {
    return L"%LOCALAPPDATA%\\DNFLauncher\\Interception\\library\\x64\\interception.dll";
}

std::optional<fs::path> InterceptionInstallerPath() {
    auto embedded = ExtractResourceFile(
        IDR_INTERCEPTION_INSTALLER,
        fs::path(L"Interception") / L"command line installer" / L"install-interception.exe");
    if (embedded) return embedded;
    fs::path installer = InterceptionRootPath() / L"command line installer" / L"install-interception.exe";
    if (fs::is_regular_file(installer)) return installer;
    return std::nullopt;
}

std::optional<fs::path> InterceptionDllPath() {
    auto embedded = ExtractResourceFile(
        IDR_INTERCEPTION_DLL_X64,
        fs::path(L"Interception") / L"library" / L"x64" / L"interception.dll");
    if (embedded) return embedded;
    fs::path dll = InterceptionRootPath() / L"library" / L"x64" / L"interception.dll";
    if (fs::is_regular_file(dll)) return dll;
    return std::nullopt;
}

FARPROC LoadInterceptionSymbol(HMODULE module, const char* name) {
    FARPROC symbol = GetProcAddress(module, name);
    if (!symbol) {
        throw std::runtime_error(std::string("interception.dll 缺少函数 ") + name);
    }
    return symbol;
}

struct InterceptionApi {
    HMODULE module = nullptr;
    InterceptionContext context = nullptr;
    DestroyContextFn destroyContext = nullptr;
    SetFilterFn setFilter = nullptr;
    WaitWithTimeoutFn waitWithTimeout = nullptr;
    SendFn send = nullptr;
    ReceiveFn receive = nullptr;

    InterceptionApi() {
        auto dll = InterceptionDllPath();
        if (!dll) {
            throw std::runtime_error("未找到 Interception 运行库：" + WideToUtf8(InterceptionDllHint()));
        }
        module = LoadLibraryW(dll->c_str());
        if (!module) throw std::runtime_error("加载 interception.dll 失败");
        try {
            auto createContext = reinterpret_cast<CreateContextFn>(LoadInterceptionSymbol(module, "interception_create_context"));
            destroyContext = reinterpret_cast<DestroyContextFn>(LoadInterceptionSymbol(module, "interception_destroy_context"));
            setFilter = reinterpret_cast<SetFilterFn>(LoadInterceptionSymbol(module, "interception_set_filter"));
            waitWithTimeout = reinterpret_cast<WaitWithTimeoutFn>(LoadInterceptionSymbol(module, "interception_wait_with_timeout"));
            send = reinterpret_cast<SendFn>(LoadInterceptionSymbol(module, "interception_send"));
            receive = reinterpret_cast<ReceiveFn>(LoadInterceptionSymbol(module, "interception_receive"));
            context = createContext();
            if (!context) throw std::runtime_error("Interception 驱动未安装或未生效");
        } catch (...) {
            if (module) FreeLibrary(module);
            module = nullptr;
            throw;
        }
    }

    ~InterceptionApi() {
        if (context && destroyContext) destroyContext(context);
        if (module) FreeLibrary(module);
    }

    InterceptionApi(const InterceptionApi&) = delete;
    InterceptionApi& operator=(const InterceptionApi&) = delete;
};

RapidFireConfig ResolveRapidFireConfig(const std::wstring& rawKey, unsigned long long intervalMs) {
    if (intervalMs < kRapidFireMinIntervalMs || intervalMs > kRapidFireMaxIntervalMs) {
        throw std::runtime_error(
            "连发间隔必须在 " + std::to_string(kRapidFireMinIntervalMs) +
            " 到 " + std::to_string(kRapidFireMaxIntervalMs) + " 毫秒之间");
    }
    std::wstring key = rawKey;
    key.erase(key.begin(), std::find_if(key.begin(), key.end(), [](wchar_t ch) { return !iswspace(ch); }));
    key.erase(std::find_if(key.rbegin(), key.rend(), [](wchar_t ch) { return !iswspace(ch); }).base(), key.end());
    if (key.size() != 1 || key[0] < 0x21 || key[0] > 0x7e) {
        throw std::runtime_error("仅支持可直接输入的单个英文、数字或符号按键");
    }
    wchar_t normalized = key[0];
    if (normalized >= L'a' && normalized <= L'z') normalized = static_cast<wchar_t>(towupper(normalized));
    SHORT keyResult = VkKeyScanW(normalized);
    if (keyResult == -1) throw std::runtime_error("当前键盘布局无法识别该按键");
    UINT vk = static_cast<unsigned short>(keyResult) & 0x00ff;
    UINT scanCode = MapVirtualKeyW(vk, MAPVK_VK_TO_VSC);
    if (scanCode == 0) throw std::runtime_error("无法获取该按键的扫描码");
    return RapidFireConfig{std::wstring(1, normalized), intervalMs, static_cast<unsigned short>(scanCode)};
}

std::vector<RapidFireConfig> ParseRapidFireConfigs(const std::wstring& json) {
    std::vector<RapidFireConfig> configs;
    size_t pos = 0;
    while (true) {
        size_t keyPos = json.find(L"\"key\"", pos);
        if (keyPos == std::wstring::npos) break;
        std::wstring tail = json.substr(keyPos);
        auto key = ExtractJsonString(tail, L"key");
        auto interval = ExtractJsonInteger(tail, L"intervalMs");
        if (!key || !interval) throw std::runtime_error("rapid-fire.json 格式错误");
        configs.push_back(ResolveRapidFireConfig(*key, static_cast<unsigned long long>(*interval)));
        pos = keyPos + 5;
    }
    return configs;
}

void LoadRapidFireConfigs() {
    fs::path path = RapidFireConfigPath();
    if (!fs::is_regular_file(path)) return;
    auto bytes = ReadFileBytes(path);
    std::wstring json = Utf8ToWide(std::string(bytes.begin(), bytes.end()));
    auto configs = ParseRapidFireConfigs(json);
    std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
    g_rapidConfigs = std::move(configs);
}

std::wstring RapidFireConfigsJsonLocked() {
    std::wstring json = L"[";
    for (size_t i = 0; i < g_rapidConfigs.size(); ++i) {
        if (i > 0) json += L",";
        json += L"{\"key\":" + JsonString(g_rapidConfigs[i].key) +
            L",\"intervalMs\":" + std::to_wstring(g_rapidConfigs[i].intervalMs) + L"}";
    }
    json += L"]";
    return json;
}

void SaveRapidFireConfigs() {
    std::wstring json;
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        json = RapidFireConfigsJsonLocked();
    }
    std::ofstream file(RapidFireConfigPath(), std::ios::binary | std::ios::trunc);
    if (!file) throw std::runtime_error("保存 rapid-fire.json 失败");
    std::string utf8 = WideToUtf8(json);
    file.write(utf8.data(), static_cast<std::streamsize>(utf8.size()));
}

bool IsTargetForeground() {
    DWORD targetPid = g_targetPid.load(std::memory_order_acquire);
    if (targetPid == 0) return false;
    HWND foreground = GetForegroundWindow();
    if (!foreground) return false;
    DWORD foregroundPid = 0;
    GetWindowThreadProcessId(foreground, &foregroundPid);
    return foregroundPid == targetPid;
}

void ForwardStroke(InterceptionApi& api, InterceptionDevice device, const InterceptionKeyStroke& stroke) {
    api.send(api.context, device, &stroke, 1);
}

void SendKey(InterceptionApi& api, InterceptionDevice device, unsigned short scanCode) {
    InterceptionKeyStroke down{scanCode, kInterceptionKeyDown, 0};
    api.send(api.context, device, &down, 1);
    std::this_thread::sleep_for(std::chrono::milliseconds(kRapidFirePressDurationMs));
    InterceptionKeyStroke up{scanCode, kInterceptionKeyUp, 0};
    api.send(api.context, device, &up, 1);
}

void HandleKeyboardStroke(InterceptionApi& api, InterceptionDevice device, const InterceptionKeyStroke& stroke) {
    unsigned short code = stroke.code;
    bool isKeyUp = (stroke.state & kInterceptionKeyUp) != 0;
    bool configured = false;
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        configured = std::any_of(g_rapidConfigs.begin(), g_rapidConfigs.end(), [code](const RapidFireConfig& config) {
            return config.scanCode == code;
        });
    }
    if (!IsTargetForeground() || !configured) {
        ForwardStroke(api, device, stroke);
        return;
    }
    std::lock_guard<std::mutex> lock(g_pressedKeysMutex);
    if (isKeyUp) g_pressedKeys.erase(code);
    else g_pressedKeys.insert(code);
}

void SendDueRepeats(InterceptionApi& api, InterceptionDevice device, std::unordered_map<unsigned short, std::chrono::steady_clock::time_point>& nextFire) {
    if (!IsTargetForeground()) {
        std::lock_guard<std::mutex> lock(g_pressedKeysMutex);
        g_pressedKeys.clear();
        nextFire.clear();
        return;
    }
    std::vector<RapidFireConfig> configs;
    std::unordered_set<unsigned short> pressed;
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        configs = g_rapidConfigs;
    }
    {
        std::lock_guard<std::mutex> lock(g_pressedKeysMutex);
        pressed = g_pressedKeys;
    }
    std::unordered_set<unsigned short> active;
    for (const auto& config : configs) {
        if (pressed.count(config.scanCode) > 0) active.insert(config.scanCode);
    }
    for (auto it = nextFire.begin(); it != nextFire.end();) {
        if (active.count(it->first) == 0) it = nextFire.erase(it);
        else ++it;
    }
    auto now = std::chrono::steady_clock::now();
    for (const auto& config : configs) {
        if (active.count(config.scanCode) == 0) continue;
        auto [it, inserted] = nextFire.emplace(config.scanCode, now);
        if (now >= it->second) {
            SendKey(api, device, config.scanCode);
            it->second = std::chrono::steady_clock::now() + std::chrono::milliseconds(config.intervalMs);
        }
    }
}

void RapidFireWorker() {
    SetThreadPriority(GetCurrentThread(), THREAD_PRIORITY_HIGHEST);
    try {
        InterceptionApi api;
        api.setFilter(api.context, IsKeyboardDevice, kInterceptionFilterKeyDown | kInterceptionFilterKeyUp);
        g_hookReady.store(true, std::memory_order_release);
        {
            std::lock_guard<std::mutex> lock(g_hookErrorMutex);
            g_hookError.clear();
        }
        InterceptionDevice sendDevice = 1;
        std::unordered_map<unsigned short, std::chrono::steady_clock::time_point> nextFire;
        while (true) {
            InterceptionDevice device = api.waitWithTimeout(api.context, 1);
            if (IsKeyboardDeviceId(device)) {
                InterceptionKeyStroke stroke{0, kInterceptionKeyUp, 0};
                if (api.receive(api.context, device, &stroke, 1) > 0) {
                    sendDevice = device;
                    HandleKeyboardStroke(api, device, stroke);
                }
            }
            SendDueRepeats(api, sendDevice, nextFire);
        }
    } catch (const std::exception& error) {
        g_hookReady.store(false, std::memory_order_release);
        std::lock_guard<std::mutex> lock(g_hookErrorMutex);
        g_hookError = Utf8ToWide(error.what());
    }
}

void EnsureRapidFireStarted() {
    std::call_once(g_rapidInitOnce, []() {
        try {
            LoadRapidFireConfigs();
        } catch (const std::exception& error) {
            std::lock_guard<std::mutex> lock(g_hookErrorMutex);
            g_hookError = Utf8ToWide(error.what());
        }
        std::thread(RapidFireWorker).detach();
    });
}

std::wstring RapidFireSnapshot() {
    EnsureRapidFireStarted();
    std::wstring configs;
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        configs = RapidFireConfigsJsonLocked();
    }
    bool ready = g_hookReady.load(std::memory_order_acquire);
    std::wstring error;
    {
        std::lock_guard<std::mutex> lock(g_hookErrorMutex);
        error = g_hookError;
    }
    bool installable = InterceptionInstallerPath().has_value() && !ready;
    return L"{\"configs\":" + configs +
        L",\"ready\":" + std::wstring(ready ? L"true" : L"false") +
        L",\"error\":" + (error.empty() ? L"null" : JsonString(error)) +
        L",\"driverInstallable\":" + std::wstring(installable ? L"true" : L"false") +
        L",\"driverInstallerHint\":" + JsonString(InterceptionInstallerHint()) + L"}";
}

std::wstring AddRapidFire(const std::wstring& key, unsigned long long intervalMs) {
    EnsureRapidFireStarted();
    std::lock_guard<std::mutex> mutation(g_rapidMutationMutex);
    RapidFireConfig config = ResolveRapidFireConfig(key, intervalMs);
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        if (std::any_of(g_rapidConfigs.begin(), g_rapidConfigs.end(), [&](const RapidFireConfig& current) {
                return current.scanCode == config.scanCode;
            })) {
            throw std::runtime_error("该按键已存在连发配置");
        }
        g_rapidConfigs.push_back(config);
    }
    try {
        SaveRapidFireConfigs();
    } catch (...) {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        g_rapidConfigs.erase(std::remove_if(g_rapidConfigs.begin(), g_rapidConfigs.end(), [&](const RapidFireConfig& current) {
            return current.scanCode == config.scanCode;
        }), g_rapidConfigs.end());
        throw;
    }
    return RapidFireSnapshot();
}

std::wstring RemoveRapidFire(const std::wstring& key) {
    EnsureRapidFireStarted();
    std::lock_guard<std::mutex> mutation(g_rapidMutationMutex);
    RapidFireConfig config = ResolveRapidFireConfig(key, kRapidFireMinIntervalMs);
    std::optional<RapidFireConfig> removed;
    {
        std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
        auto it = std::find_if(g_rapidConfigs.begin(), g_rapidConfigs.end(), [&](const RapidFireConfig& current) {
            return current.scanCode == config.scanCode;
        });
        if (it == g_rapidConfigs.end()) throw std::runtime_error("未找到该按键的连发配置");
        removed = *it;
        g_rapidConfigs.erase(it);
    }
    try {
        SaveRapidFireConfigs();
    } catch (...) {
        if (removed) {
            std::lock_guard<std::mutex> lock(g_rapidConfigMutex);
            g_rapidConfigs.push_back(*removed);
        }
        throw;
    }
    {
        std::lock_guard<std::mutex> lock(g_pressedKeysMutex);
        g_pressedKeys.erase(config.scanCode);
    }
    return RapidFireSnapshot();
}

std::wstring InstallInterceptionDriver() {
    EnsureRapidFireStarted();
    auto installer = InterceptionInstallerPath();
    if (!installer) {
        throw std::runtime_error("未找到驱动安装程序，内置 Interception 资源释放失败：" + WideToUtf8(InterceptionInstallerHint()));
    }
    fs::path directory = installer->parent_path();
    HINSTANCE result = ShellExecuteW(
        g_mainWindow,
        L"runas",
        installer->c_str(),
        L"/install",
        directory.c_str(),
        SW_SHOWNORMAL);
    if (reinterpret_cast<INT_PTR>(result) <= 32) {
        throw std::runtime_error("启动驱动安装程序失败，ShellExecuteW 返回 " + std::to_string(reinterpret_cast<INT_PTR>(result)));
    }
    {
        std::lock_guard<std::mutex> lock(g_hookErrorMutex);
        g_hookError = L"驱动安装程序已启动，请完成后重启电脑";
    }
    return RapidFireSnapshot();
}

std::wstring HandleCommand(const std::wstring& command, const std::wstring& json) {
    if (command == L"get_launcher_window_title") return JsonString(kWindowTitle);
    if (command == L"get_launcher_background") return LauncherBackground();
    if (command == L"set_window_title") {
        std::wstring title = ExtractJsonString(json, L"title").value_or(kWindowTitle);
        SetWindowTextW(g_mainWindow, title.c_str());
        return L"null";
    }
    if (command == L"start_window_drag") {
        ReleaseCapture();
        SendMessageW(g_mainWindow, WM_NCLBUTTONDOWN, HTCAPTION, 0);
        return L"null";
    }
    if (command == L"minimize_window") {
        ShowWindow(g_mainWindow, SW_MINIMIZE);
        return L"null";
    }
    if (command == L"close_window") {
        PostMessageW(g_mainWindow, WM_CLOSE, 0, 0);
        return L"null";
    }
    if (command == L"reveal_window") {
        RevealMainWindow(true);
        return L"null";
    }
    if (command == L"save_saved_login") {
        SaveSavedLogin(ExtractJsonString(json, L"account").value_or(L""), ExtractJsonString(json, L"password").value_or(L""));
        return L"null";
    }
    if (command == L"load_saved_login") return LoadSavedLogin();
    if (command == L"clear_saved_login") {
        ClearSavedLogin();
        return L"null";
    }
    if (command == L"open_url") {
        OpenUrl(ExtractJsonString(json, L"url").value_or(L""));
        return L"null";
    }
    if (command == L"launch_game") {
        LaunchGame(ExtractJsonString(json, L"dnfToken").value_or(L""), ExtractJsonString(json, L"expectedPvfMd5").value_or(L""));
        return L"null";
    }
    if (command == L"is_game_running") return IsGameRunning() ? L"true" : L"false";
    if (command == L"stop_game") {
        StopGame();
        return L"null";
    }
    if (command == L"list_rapid_fire") return RapidFireSnapshot();
    if (command == L"add_rapid_fire") {
        return AddRapidFire(
            ExtractJsonString(json, L"key").value_or(L""),
            static_cast<unsigned long long>(ExtractJsonInteger(json, L"intervalMs").value_or(0)));
    }
    if (command == L"remove_rapid_fire") {
        return RemoveRapidFire(ExtractJsonString(json, L"key").value_or(L""));
    }
    if (command == L"install_interception_driver") {
        return InstallInterceptionDriver();
    }
    throw std::runtime_error("未知本机命令");
}

void CaptureWebViewPreview() {
    if (!g_webview) return;
    fs::path previewPath = fs::path(ExeDirectory()) / L"launcher-preview.png";
    IStream* stream = nullptr;
    HRESULT streamResult = SHCreateStreamOnFileEx(
        previewPath.wstring().c_str(),
        STGM_CREATE | STGM_WRITE | STGM_SHARE_EXCLUSIVE,
        FILE_ATTRIBUTE_NORMAL,
        TRUE,
        nullptr,
        &stream);
    LogLine(L"CapturePreview stream result=" + HResultText(streamResult) + L" path=" + previewPath.wstring());
    if (FAILED(streamResult) || !stream) return;

    g_webview->CapturePreview(
        COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT_PNG,
        stream,
        Callback<ICoreWebView2CapturePreviewCompletedHandler>(
            [stream](HRESULT result) -> HRESULT {
                LogLine(L"CapturePreview result=" + HResultText(result));
                stream->Release();
                return S_OK;
            }).Get());
}

void HandleWebMessage(ICoreWebView2WebMessageReceivedEventArgs* args) {
    LPWSTR rawJson = nullptr;
    if (FAILED(args->get_WebMessageAsJson(&rawJson)) || !rawJson) return;
    std::wstring json(rawJson);
    CoTaskMemFree(rawJson);
    long long id = ExtractJsonInteger(json, L"id").value_or(0);
    std::wstring type = ExtractJsonString(json, L"type").value_or(L"");
    if (type == L"native-log") {
        std::wstring message = ExtractJsonString(json, L"message").value_or(json);
        LogLine(L"frontend: " + message);
        if (message.rfind(L"snapshot ", 0) == 0) {
            CaptureWebViewPreview();
        }
        return;
    }
    std::wstring command = ExtractJsonString(json, L"command").value_or(L"");
    LogLine(L"native command: " + command);
    try {
        PostNativeResult(id, true, HandleCommand(command, json));
    } catch (const std::exception& error) {
        LogLine(L"native command failed: " + command + L" error=" + Utf8ToWide(error.what()));
        PostNativeResult(id, false, JsonString(Utf8ToWide(error.what())));
    }
}

const wchar_t* NativeBridgeScript() {
    return LR"JS(
(() => {
  const pending = new Map();
  let nextId = 1;
  function nativeLog(message) {
    try {
      chrome.webview.postMessage({ type: "native-log", message: String(message) });
    } catch (_) {
    }
  }
  window.addEventListener("error", (event) => {
    nativeLog(`error: ${event.message || ""} ${event.filename || ""}:${event.lineno || 0}:${event.colno || 0}`);
  });
  window.addEventListener("unhandledrejection", (event) => {
    nativeLog(`unhandledrejection: ${event.reason?.message || event.reason || ""}`);
  });
  document.addEventListener("DOMContentLoaded", () => nativeLog(`DOMContentLoaded readyState=${document.readyState}`));
  window.addEventListener("load", () => nativeLog(`load readyState=${document.readyState}`));
  window.setTimeout(() => {
    const shell = document.querySelector(".launcher-shell");
    const login = document.querySelector("#loginView");
    const titlebar = document.querySelector(".titlebar");
    const shellRect = shell?.getBoundingClientRect();
    const loginRect = login?.getBoundingClientRect();
    nativeLog(
      `snapshot readyState=${document.readyState} bodyClass=${document.body?.className || ""} ` +
      `text=${(document.body?.innerText || "").slice(0, 120).replace(/\s+/g, " ")} ` +
      `shell=${shellRect ? `${shellRect.width}x${shellRect.height}` : "missing"} ` +
      `login=${loginRect ? `${loginRect.width}x${loginRect.height}` : "missing"} ` +
      `titlebar=${titlebar ? "present" : "missing"}`
    );
  }, 1200);
  function invoke(command, args = {}) {
    const id = nextId++;
    return new Promise((resolve, reject) => {
      pending.set(id, { resolve, reject });
      chrome.webview.postMessage({ type: "native-invoke", id, command, args });
    });
  }
  chrome.webview.addEventListener("message", (event) => {
    const data = event.data || {};
    if (data.type !== "native-result") return;
    const item = pending.get(data.id);
    if (!item) return;
    pending.delete(data.id);
    if (data.ok) item.resolve(data.result);
    else item.reject(new Error(data.error || "Native command failed"));
  });
  window.dnfNative = {
    invoke,
    setWindowTitle: (title) => invoke("set_window_title", { title }),
    startWindowDrag: () => invoke("start_window_drag"),
    minimizeWindow: () => invoke("minimize_window"),
    closeWindow: () => invoke("close_window"),
    revealWindow: () => invoke("reveal_window"),
  };
})();
)JS";
}

void ResizeWebView() {
    if (!g_controller) return;
    RECT bounds{};
    GetClientRect(g_mainWindow, &bounds);
    g_controller->put_Bounds(bounds);
    g_controller->put_IsVisible(TRUE);
}

void AttachDebugLogging() {
    if (!g_webview) return;

    EventRegistrationToken token{};
    g_webview->add_NavigationStarting(
        Callback<ICoreWebView2NavigationStartingEventHandler>(
            [](ICoreWebView2*, ICoreWebView2NavigationStartingEventArgs* args) -> HRESULT {
                LPWSTR uri = nullptr;
                if (SUCCEEDED(args->get_Uri(&uri)) && uri) {
                    LogLine(L"navigation starting: " + std::wstring(uri));
                    CoTaskMemFree(uri);
                } else {
                    LogLine(L"navigation starting");
                }
                return S_OK;
            }).Get(),
        &token);

    g_webview->add_NavigationCompleted(
        Callback<ICoreWebView2NavigationCompletedEventHandler>(
            [](ICoreWebView2*, ICoreWebView2NavigationCompletedEventArgs* args) -> HRESULT {
                BOOL success = FALSE;
                COREWEBVIEW2_WEB_ERROR_STATUS status = COREWEBVIEW2_WEB_ERROR_STATUS_UNKNOWN;
                args->get_IsSuccess(&success);
                args->get_WebErrorStatus(&status);
                LogLine(
                    L"navigation completed: success=" + std::wstring(success ? L"true" : L"false") +
                    L" status=" + std::to_wstring(static_cast<int>(status)));
                ResizeWebView();
                InvalidateRect(g_mainWindow, nullptr, TRUE);
                UpdateWindow(g_mainWindow);
                return S_OK;
            }).Get(),
        &token);

    g_webview->CallDevToolsProtocolMethod(
        L"Runtime.enable",
        L"{}",
        Callback<ICoreWebView2CallDevToolsProtocolMethodCompletedHandler>(
            [](HRESULT result, LPCWSTR response) -> HRESULT {
                LogLine(L"devtools Runtime.enable result=" + HResultText(result) + L" response=" + (response ? response : L""));
                return S_OK;
            }).Get());

    ComPtr<ICoreWebView2DevToolsProtocolEventReceiver> consoleReceiver;
    if (SUCCEEDED(g_webview->GetDevToolsProtocolEventReceiver(L"Runtime.consoleAPICalled", &consoleReceiver)) && consoleReceiver) {
        consoleReceiver->add_DevToolsProtocolEventReceived(
            Callback<ICoreWebView2DevToolsProtocolEventReceivedEventHandler>(
                [](ICoreWebView2*, ICoreWebView2DevToolsProtocolEventReceivedEventArgs* args) -> HRESULT {
                    LPWSTR json = nullptr;
                    if (SUCCEEDED(args->get_ParameterObjectAsJson(&json)) && json) {
                        LogLine(L"js console: " + std::wstring(json));
                        CoTaskMemFree(json);
                    }
                    return S_OK;
                }).Get(),
            &token);
    }

    ComPtr<ICoreWebView2DevToolsProtocolEventReceiver> exceptionReceiver;
    if (SUCCEEDED(g_webview->GetDevToolsProtocolEventReceiver(L"Runtime.exceptionThrown", &exceptionReceiver)) && exceptionReceiver) {
        exceptionReceiver->add_DevToolsProtocolEventReceived(
            Callback<ICoreWebView2DevToolsProtocolEventReceivedEventHandler>(
                [](ICoreWebView2*, ICoreWebView2DevToolsProtocolEventReceivedEventArgs* args) -> HRESULT {
                    LPWSTR json = nullptr;
                    if (SUCCEEDED(args->get_ParameterObjectAsJson(&json)) && json) {
                        LogLine(L"js exception: " + std::wstring(json));
                        CoTaskMemFree(json);
                    }
                    return S_OK;
                }).Get(),
            &token);
    }
}

void InitializeWebView() {
    LogLine(L"InitializeWebView");
    auto options = Microsoft::WRL::Make<CoreWebView2EnvironmentOptions>();
    CreateCoreWebView2EnvironmentWithOptions(
        nullptr,
        nullptr,
        options.Get(),
        Callback<ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler>(
            [](HRESULT result, ICoreWebView2Environment* environment) -> HRESULT {
                LogLine(L"CreateCoreWebView2EnvironmentWithOptions result=" + HResultText(result));
                if (FAILED(result) || !environment) {
                    MessageBoxW(g_mainWindow, L"初始化 WebView2 失败", kWindowTitle, MB_ICONERROR);
                    return S_OK;
                }
                environment->CreateCoreWebView2Controller(
                    g_mainWindow,
                    Callback<ICoreWebView2CreateCoreWebView2ControllerCompletedHandler>(
                        [](HRESULT result, ICoreWebView2Controller* controller) -> HRESULT {
                            LogLine(L"CreateCoreWebView2Controller result=" + HResultText(result));
                            if (FAILED(result) || !controller) {
                                MessageBoxW(g_mainWindow, L"创建 WebView2 控件失败", kWindowTitle, MB_ICONERROR);
                                return S_OK;
                            }
                            g_controller = controller;
                            g_controller->get_CoreWebView2(&g_webview);
                            ComPtr<ICoreWebView2Controller2> controller2;
                            if (SUCCEEDED(g_controller.As(&controller2)) && controller2) {
                                COREWEBVIEW2_COLOR background{};
                                background.A = 255;
                                background.R = 5;
                                background.G = 8;
                                background.B = 19;
                                HRESULT backgroundResult = controller2->put_DefaultBackgroundColor(background);
                                LogLine(L"put_DefaultBackgroundColor result=" + HResultText(backgroundResult));
                            }
                            ResizeWebView();
                            g_controller->put_IsVisible(TRUE);
                            AttachDebugLogging();

                            ComPtr<ICoreWebView2Settings> settings;
                            g_webview->get_Settings(&settings);
                            settings->put_AreDefaultContextMenusEnabled(FALSE);
                            settings->put_AreDevToolsEnabled(FALSE);
                            ComPtr<ICoreWebView2Settings3> settings3;
                            if (SUCCEEDED(settings.As(&settings3))) {
                                settings3->put_AreBrowserAcceleratorKeysEnabled(FALSE);
                            }

                            EventRegistrationToken token{};
                            g_webview->add_WebMessageReceived(
                                Callback<ICoreWebView2WebMessageReceivedEventHandler>(
                                    [](ICoreWebView2*, ICoreWebView2WebMessageReceivedEventArgs* args) -> HRESULT {
                                        HandleWebMessage(args);
                                        return S_OK;
                                    }).Get(),
                                &token);

                            std::wstring frontendHtml = EmbeddedFrontendHtml();
                            if (frontendHtml.empty()) {
                                LogLine(L"embedded frontend not found");
                                MessageBoxW(g_mainWindow, L"内置前端资源缺失", kWindowTitle, MB_ICONERROR);
                                return S_OK;
                            }
                            LogLine(L"frontend: embedded resource");
                            g_webview->AddScriptToExecuteOnDocumentCreated(
                                NativeBridgeScript(),
                                Callback<ICoreWebView2AddScriptToExecuteOnDocumentCreatedCompletedHandler>(
                                    [frontendHtml](HRESULT scriptResult, LPCWSTR) -> HRESULT {
                                        LogLine(L"AddScriptToExecuteOnDocumentCreated result=" + HResultText(scriptResult));
                                        if (FAILED(scriptResult)) {
                                            MessageBoxW(g_mainWindow, L"初始化本机桥接失败", kWindowTitle, MB_ICONERROR);
                                            return S_OK;
                                        }
                                        g_webview->NavigateToString(frontendHtml.c_str());
                                        LogLine(L"NavigateToString called");
                                        ShowWindow(g_mainWindow, SW_SHOW);
                                        SetWindowPos(g_mainWindow, nullptr, 0, 0, 0, 0, SWP_NOMOVE | SWP_NOSIZE | SWP_NOZORDER | SWP_FRAMECHANGED);
                                        ResizeWebView();
                                        return S_OK;
                                    }).Get());
                            return S_OK;
                        }).Get());
                return S_OK;
            }).Get());
}

LRESULT CALLBACK WindowProc(HWND hwnd, UINT message, WPARAM wParam, LPARAM lParam) {
    switch (message) {
    case WM_SIZE:
        ResizeWebView();
        return 0;
    case WM_GETMINMAXINFO: {
        auto* info = reinterpret_cast<MINMAXINFO*>(lParam);
        UINT dpi = GetDpiForWindow(hwnd);
        info->ptMinTrackSize.x = ScaleForDpi(1024, dpi);
        info->ptMinTrackSize.y = ScaleForDpi(576, dpi);
        return 0;
    }
    case WM_DPICHANGED: {
        auto* suggested = reinterpret_cast<RECT*>(lParam);
        SetWindowPos(
            hwnd,
            nullptr,
            suggested->left,
            suggested->top,
            suggested->right - suggested->left,
            suggested->bottom - suggested->top,
            SWP_NOZORDER | SWP_NOACTIVATE);
        ResizeWebView();
        return 0;
    }
    case WM_DESTROY:
        if (IsGameRunning()) StopGame();
        PostQuitMessage(0);
        return 0;
    default:
        return DefWindowProcW(hwnd, message, wParam, lParam);
    }
}

bool RegisterMainWindowClass(HINSTANCE instance) {
    WNDCLASSEXW wc{};
    wc.cbSize = sizeof(wc);
    wc.hInstance = instance;
    wc.lpfnWndProc = WindowProc;
    wc.lpszClassName = kWindowClassName;
    wc.hCursor = LoadCursor(nullptr, IDC_ARROW);
    wc.hIcon = reinterpret_cast<HICON>(LoadImageW(instance, MAKEINTRESOURCEW(IDI_APP_ICON), IMAGE_ICON, 0, 0, LR_DEFAULTSIZE));
    wc.hIconSm = reinterpret_cast<HICON>(LoadImageW(instance, MAKEINTRESOURCEW(IDI_APP_ICON), IMAGE_ICON, GetSystemMetrics(SM_CXSMICON), GetSystemMetrics(SM_CYSMICON), 0));
    wc.hbrBackground = reinterpret_cast<HBRUSH>(GetStockObject(BLACK_BRUSH));
    return RegisterClassExW(&wc) != 0;
}

HWND CreateMainWindow(HINSTANCE instance) {
    UINT dpi = GetDpiForSystem();
    int width = ScaleForDpi(1280, dpi);
    int height = ScaleForDpi(720, dpi);
    int x = (GetSystemMetrics(SM_CXSCREEN) - width) / 2;
    int y = (GetSystemMetrics(SM_CYSCREEN) - height) / 2;
    return CreateWindowExW(
        WS_EX_APPWINDOW | WS_EX_LAYERED,
        kWindowClassName,
        kWindowTitle,
        WS_POPUP | WS_MINIMIZEBOX | WS_CLIPCHILDREN | WS_CLIPSIBLINGS,
        x,
        y,
        width,
        height,
        nullptr,
        nullptr,
        instance,
        nullptr);
}

} // namespace

int APIENTRY wWinMain(HINSTANCE instance, HINSTANCE, LPWSTR, int) {
    SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2);
    CoInitializeEx(nullptr, COINIT_APARTMENTTHREADED);
    if (!RegisterMainWindowClass(instance)) return 1;
    g_mainWindow = CreateMainWindow(instance);
    if (!g_mainWindow) return 1;
    EnsureRapidFireStarted();
    SetMainWindowOpacity(0);
    ShowWindow(g_mainWindow, SW_SHOWNORMAL);
    UpdateWindow(g_mainWindow);
    InitializeWebView();

    MSG message{};
    while (GetMessageW(&message, nullptr, 0, 0)) {
        TranslateMessage(&message);
        DispatchMessageW(&message);
    }
    CoUninitialize();
    return 0;
}
