use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};
use std::ffi::{CString, OsStr, c_void};
use std::fs;
use std::mem::transmute;
use std::os::windows::ffi::OsStrExt;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};
use windows_sys::Win32::Foundation::{FARPROC, FreeLibrary, HWND};
use windows_sys::Win32::System::LibraryLoader::{GetProcAddress, LoadLibraryW};
use windows_sys::Win32::System::Threading::{
    GetCurrentThread, SetThreadPriority, THREAD_PRIORITY_HIGHEST,
};
use windows_sys::Win32::UI::Input::KeyboardAndMouse::{
    MAPVK_VK_TO_VSC, MapVirtualKeyW, VkKeyScanW,
};
use windows_sys::Win32::UI::Shell::ShellExecuteW;
use windows_sys::Win32::UI::WindowsAndMessaging::{GetForegroundWindow, GetWindowThreadProcessId};

const CONFIG_FILE_NAME: &str = "rapid-fire.json";
const MIN_INTERVAL_MS: u64 = 1;
const MAX_INTERVAL_MS: u64 = 10_000;
const KEY_PRESS_DURATION_MS: u64 = 1;
const INTERCEPTION_INSTALL_ARGS: &str = "/install";
const INTERCEPTION_FILTER_KEY_DOWN: u16 = 1;
const INTERCEPTION_FILTER_KEY_UP: u16 = 2;
const INTERCEPTION_KEY_DOWN: u16 = 0;
const INTERCEPTION_KEY_UP: u16 = 1;
const INTERCEPTION_MAX_KEYBOARD: Device = 10;

type Device = i32;
type InterceptionContext = *mut c_void;
type CreateContextFn = unsafe extern "C" fn() -> InterceptionContext;
type DestroyContextFn = unsafe extern "C" fn(InterceptionContext);
type SetFilterFn =
    unsafe extern "C" fn(InterceptionContext, Option<unsafe extern "C" fn(Device) -> i32>, u16);
type WaitWithTimeoutFn = unsafe extern "C" fn(InterceptionContext, u32) -> Device;
type SendFn =
    unsafe extern "C" fn(InterceptionContext, Device, *const InterceptionKeyStroke, u32) -> i32;
type ReceiveFn =
    unsafe extern "C" fn(InterceptionContext, Device, *mut InterceptionKeyStroke, u32) -> i32;

#[repr(C)]
#[derive(Clone, Copy)]
struct InterceptionKeyStroke {
    code: u16,
    state: u16,
    information: u32,
}

struct InterceptionApi {
    module: isize,
    context: InterceptionContext,
    destroy_context: DestroyContextFn,
    set_filter: SetFilterFn,
    wait_with_timeout: WaitWithTimeoutFn,
    send: SendFn,
    receive: ReceiveFn,
}

unsafe impl Send for InterceptionApi {}

impl Drop for InterceptionApi {
    fn drop(&mut self) {
        unsafe {
            (self.destroy_context)(self.context);
            FreeLibrary(self.module as _);
        }
    }
}

impl InterceptionApi {
    fn new() -> Result<Self, String> {
        let dll = interception_dll_path()
            .ok_or_else(|| format!("未找到 Interception 运行库：{}", interception_dll_hint()))?;
        let module = unsafe { LoadLibraryW(wide(dll.as_os_str()).as_ptr()) };
        if module.is_null() {
            return Err("加载 interception.dll 失败".to_string());
        }
        let create_context: CreateContextFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_create_context",
            )?)
        };
        let destroy_context: DestroyContextFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_destroy_context",
            )?)
        };
        let set_filter: SetFilterFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_set_filter",
            )?)
        };
        let wait_with_timeout: WaitWithTimeoutFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_wait_with_timeout",
            )?)
        };
        let send: SendFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_send",
            )?)
        };
        let receive: ReceiveFn = unsafe {
            transmute(load_raw_symbol_or_free(
                module as isize,
                "interception_receive",
            )?)
        };
        let context = unsafe { create_context() };
        if context.is_null() {
            unsafe {
                FreeLibrary(module);
            }
            return Err("Interception 驱动未就绪，请安装后重启电脑".to_string());
        }
        Ok(Self {
            module: module as isize,
            context,
            destroy_context,
            set_filter,
            wait_with_timeout,
            send,
            receive,
        })
    }
}

fn load_raw_symbol_or_free(module: isize, name: &str) -> Result<FARPROC, String> {
    match unsafe { load_raw_symbol(module, name) } {
        Ok(symbol) => Ok(symbol),
        Err(error) => {
            unsafe {
                FreeLibrary(module as _);
            }
            Err(error)
        }
    }
}

unsafe fn load_raw_symbol(module: isize, name: &str) -> Result<FARPROC, String> {
    let name = CString::new(name).map_err(|error| error.to_string())?;
    let symbol = unsafe { GetProcAddress(module as _, name.as_ptr() as _) };
    let Some(symbol) = symbol else {
        return Err(format!(
            "interception.dll 缺少函数 {}",
            name.to_string_lossy()
        ));
    };
    Ok(Some(symbol))
}

#[derive(Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RapidFireConfig {
    pub key: String,
    pub interval_ms: u64,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RapidFireSnapshot {
    configs: Vec<RapidFireConfig>,
    ready: bool,
    error: Option<String>,
    driver_installable: bool,
    driver_installer_hint: String,
}

#[derive(Clone)]
struct ResolvedConfig {
    public: RapidFireConfig,
    scan_code: u16,
}

struct RapidFireRuntime {
    configs: Mutex<Vec<ResolvedConfig>>,
    pressed_keys: Mutex<HashSet<u16>>,
    target_pid: Arc<AtomicU32>,
    hook_ready: AtomicBool,
    hook_error: Mutex<Option<String>>,
}

pub struct RapidFireState {
    runtime: Arc<RapidFireRuntime>,
    config_path: PathBuf,
    mutation_lock: Mutex<()>,
}

impl RapidFireState {
    pub fn new(config_path: PathBuf, target_pid: Arc<AtomicU32>) -> Result<Self, String> {
        let stored = if config_path.is_file() {
            let content = fs::read_to_string(&config_path)
                .map_err(|error| format!("读取 {} 失败: {}", CONFIG_FILE_NAME, error))?;
            serde_json::from_str::<Vec<RapidFireConfig>>(&content)
                .map_err(|error| format!("{} 格式错误: {}", CONFIG_FILE_NAME, error))?
        } else {
            Vec::new()
        };

        let mut configs = Vec::with_capacity(stored.len());
        for config in stored {
            configs.push(resolve_config(config.key, config.interval_ms)?);
        }

        let runtime = Arc::new(RapidFireRuntime {
            configs: Mutex::new(configs),
            pressed_keys: Mutex::new(HashSet::new()),
            target_pid,
            hook_ready: AtomicBool::new(false),
            hook_error: Mutex::new(None),
        });
        start_interception_worker(Arc::clone(&runtime));
        Ok(Self {
            runtime,
            config_path,
            mutation_lock: Mutex::new(()),
        })
    }

    fn snapshot(&self) -> Result<RapidFireSnapshot, String> {
        let configs = self
            .runtime
            .configs
            .lock()
            .map_err(|_| "无法读取按键连发配置".to_string())?
            .iter()
            .map(|config| config.public.clone())
            .collect();
        let error = self
            .runtime
            .hook_error
            .lock()
            .map_err(|_| "无法读取按键监听状态".to_string())?
            .clone();
        Ok(RapidFireSnapshot {
            configs,
            ready: self.runtime.hook_ready.load(Ordering::Acquire),
            error,
            driver_installable: interception_installer_path().is_some()
                && !self.runtime.hook_ready.load(Ordering::Acquire),
            driver_installer_hint: interception_installer_hint(),
        })
    }

    fn save(&self) -> Result<(), String> {
        let configs: Vec<RapidFireConfig> = self
            .runtime
            .configs
            .lock()
            .map_err(|_| "无法读取按键连发配置".to_string())?
            .iter()
            .map(|config| config.public.clone())
            .collect();
        let content = serde_json::to_string_pretty(&configs).map_err(|error| error.to_string())?;
        fs::write(&self.config_path, content)
            .map_err(|error| format!("保存 {} 失败: {}", CONFIG_FILE_NAME, error))
    }
}

#[tauri::command]
pub fn list_rapid_fire(
    state: tauri::State<'_, RapidFireState>,
) -> Result<RapidFireSnapshot, String> {
    state.snapshot()
}

#[tauri::command]
pub fn add_rapid_fire(
    state: tauri::State<'_, RapidFireState>,
    key: String,
    interval_ms: u64,
) -> Result<RapidFireSnapshot, String> {
    let _mutation = state
        .mutation_lock
        .lock()
        .map_err(|_| "无法修改按键连发配置".to_string())?;
    let config = resolve_config(key, interval_ms)?;
    let config_scan_code = config.scan_code;
    {
        let mut configs = state
            .runtime
            .configs
            .lock()
            .map_err(|_| "无法修改按键连发配置".to_string())?;
        if configs
            .iter()
            .any(|current| current.scan_code == config.scan_code)
        {
            return Err("该按键已存在连发配置".to_string());
        }
        configs.push(config);
    }
    if let Err(error) = state.save() {
        if let Ok(mut configs) = state.runtime.configs.lock() {
            configs.retain(|current| current.scan_code != config_scan_code);
        }
        return Err(error);
    }
    state.snapshot()
}

#[tauri::command]
pub fn remove_rapid_fire(
    state: tauri::State<'_, RapidFireState>,
    key: String,
) -> Result<RapidFireSnapshot, String> {
    let _mutation = state
        .mutation_lock
        .lock()
        .map_err(|_| "无法修改按键连发配置".to_string())?;
    let (_, _, scan_code) = normalize_key(&key)?;
    let removed = {
        let mut configs = state
            .runtime
            .configs
            .lock()
            .map_err(|_| "无法修改按键连发配置".to_string())?;
        let Some(index) = configs
            .iter()
            .position(|current| current.scan_code == scan_code)
        else {
            return Err("未找到该按键的连发配置".to_string());
        };
        configs.remove(index)
    };
    if let Err(error) = state.save() {
        if let Ok(mut configs) = state.runtime.configs.lock() {
            configs.push(removed);
        }
        return Err(error);
    }
    if let Ok(mut pressed) = state.runtime.pressed_keys.lock() {
        pressed.remove(&scan_code);
    }
    state.snapshot()
}

#[tauri::command]
pub fn install_interception_driver(
    state: tauri::State<'_, RapidFireState>,
) -> Result<RapidFireSnapshot, String> {
    let installer = interception_installer_path().ok_or_else(|| {
        format!(
            "未找到驱动安装程序，请确认官方 Interception 目录已放到 {}",
            interception_installer_hint()
        )
    })?;
    let directory = installer
        .parent()
        .map(PathBuf::from)
        .ok_or_else(|| "无法获取驱动安装程序目录".to_string())?;
    run_installer_as_admin(&installer, &directory)?;
    if let Ok(mut error) = state.runtime.hook_error.lock() {
        *error = Some("驱动安装程序已启动，请完成后重启电脑".to_string());
    }
    state.snapshot()
}

fn interception_installer_hint() -> String {
    "启动器同目录\\start\\Interception\\command line installer\\install-interception.exe"
        .to_string()
}

fn interception_dll_hint() -> String {
    "启动器同目录\\start\\Interception\\library\\x64\\interception.dll".to_string()
}

fn interception_root_path() -> Option<PathBuf> {
    Some(
        std::env::current_exe()
            .ok()?
            .parent()
            .map(PathBuf::from)?
            .join("start")
            .join("Interception"),
    )
}

fn interception_installer_path() -> Option<PathBuf> {
    let installer = interception_root_path()?
        .join("command line installer")
        .join("install-interception.exe");
    installer.is_file().then_some(installer)
}

fn interception_dll_path() -> Option<PathBuf> {
    let dll = interception_root_path()?
        .join("library")
        .join("x64")
        .join("interception.dll");
    dll.is_file().then_some(dll)
}

fn wide(value: &OsStr) -> Vec<u16> {
    value.encode_wide().chain(std::iter::once(0)).collect()
}

fn run_installer_as_admin(installer: &Path, directory: &Path) -> Result<(), String> {
    let operation = wide(OsStr::new("runas"));
    let file = wide(installer.as_os_str());
    let parameters = wide(OsStr::new(INTERCEPTION_INSTALL_ARGS));
    let directory = wide(directory.as_os_str());
    let result = unsafe {
        ShellExecuteW(
            std::ptr::null_mut::<std::ffi::c_void>() as HWND,
            operation.as_ptr(),
            file.as_ptr(),
            parameters.as_ptr(),
            directory.as_ptr(),
            1,
        )
    } as isize;
    if result <= 32 {
        return Err(format!(
            "启动驱动安装程序失败，ShellExecuteW 返回 {}",
            result
        ));
    }
    Ok(())
}

fn resolve_config(key: String, interval_ms: u64) -> Result<ResolvedConfig, String> {
    if !(MIN_INTERVAL_MS..=MAX_INTERVAL_MS).contains(&interval_ms) {
        return Err(format!(
            "连发间隔必须在 {} 到 {} 毫秒之间",
            MIN_INTERVAL_MS, MAX_INTERVAL_MS
        ));
    }
    let (key, _, scan_code) = normalize_key(&key)?;
    Ok(ResolvedConfig {
        public: RapidFireConfig { key, interval_ms },
        scan_code,
    })
}

fn normalize_key(value: &str) -> Result<(String, u16, u16), String> {
    let mut characters = value.trim().chars();
    let Some(character) = characters.next() else {
        return Err("请输入一个按键".to_string());
    };
    if characters.next().is_some() || !character.is_ascii_graphic() {
        return Err("仅支持可直接输入的单个英文、数字或符号按键".to_string());
    }
    let normalized = if character.is_ascii_alphabetic() {
        character.to_ascii_uppercase()
    } else {
        character
    };
    let key_result = unsafe { VkKeyScanW(normalized as u16) };
    if key_result == -1 {
        return Err("当前键盘布局无法识别该按键".to_string());
    }
    let vk = (key_result as u16 & 0x00ff) as u32;
    let scan_code = unsafe { MapVirtualKeyW(vk, MAPVK_VK_TO_VSC) } as u16;
    if scan_code == 0 {
        return Err("无法获取该按键的扫描码".to_string());
    }
    Ok((normalized.to_string(), vk as u16, scan_code))
}

fn is_target_foreground(runtime: &RapidFireRuntime) -> bool {
    let target_pid = runtime.target_pid.load(Ordering::Acquire);
    if target_pid == 0 {
        return false;
    }
    let foreground = unsafe { GetForegroundWindow() };
    if foreground.is_null() {
        return false;
    }
    let mut foreground_pid = 0_u32;
    unsafe {
        GetWindowThreadProcessId(foreground, &mut foreground_pid);
    }
    foreground_pid == target_pid
}

fn start_interception_worker(runtime: Arc<RapidFireRuntime>) {
    thread::spawn(move || {
        unsafe {
            SetThreadPriority(GetCurrentThread(), THREAD_PRIORITY_HIGHEST);
        }
        let context = match InterceptionApi::new() {
            Ok(context) => context,
            Err(message) => {
                if let Ok(mut error) = runtime.hook_error.lock() {
                    *error = Some(message);
                }
                return;
            }
        };
        unsafe {
            (context.set_filter)(
                context.context,
                Some(is_keyboard_device),
                INTERCEPTION_FILTER_KEY_DOWN | INTERCEPTION_FILTER_KEY_UP,
            );
        }

        runtime.hook_ready.store(true, Ordering::Release);
        if let Ok(mut error) = runtime.hook_error.lock() {
            *error = None;
        }

        let mut send_device: Device = 1;
        let mut next_fire: HashMap<u16, Instant> = HashMap::new();
        loop {
            let device = unsafe { (context.wait_with_timeout)(context.context, 1) };
            if is_keyboard_device_id(device) {
                let mut stroke = InterceptionKeyStroke {
                    code: 0,
                    state: INTERCEPTION_KEY_UP,
                    information: 0,
                };
                if unsafe { (context.receive)(context.context, device, &mut stroke, 1) } > 0 {
                    send_device = device;
                    handle_keyboard_stroke(&context, &runtime, device, stroke);
                }
            }
            send_due_repeats(&context, &runtime, send_device, &mut next_fire);
        }
    });
}

unsafe extern "C" fn is_keyboard_device(device: Device) -> i32 {
    is_keyboard_device_id(device) as i32
}

fn is_keyboard_device_id(device: Device) -> bool {
    (1..=INTERCEPTION_MAX_KEYBOARD).contains(&device)
}

fn handle_keyboard_stroke(
    context: &InterceptionApi,
    runtime: &RapidFireRuntime,
    device: Device,
    stroke: InterceptionKeyStroke,
) {
    let code = stroke.code;
    let is_key_up = stroke.state & INTERCEPTION_KEY_UP != 0;
    let is_configured = runtime
        .configs
        .lock()
        .map(|configs| configs.iter().any(|config| config.scan_code == code))
        .unwrap_or(false);

    if !is_target_foreground(runtime) || !is_configured {
        forward_stroke(context, device, stroke);
        return;
    }

    if is_key_up {
        if let Ok(mut pressed) = runtime.pressed_keys.lock() {
            pressed.remove(&code);
        }
    } else if let Ok(mut pressed) = runtime.pressed_keys.lock() {
        pressed.insert(code);
    }
}

fn forward_stroke(context: &InterceptionApi, device: Device, stroke: InterceptionKeyStroke) {
    unsafe {
        (context.send)(context.context, device, &stroke, 1);
    }
}

fn send_due_repeats(
    context: &InterceptionApi,
    runtime: &RapidFireRuntime,
    device: Device,
    next_fire: &mut HashMap<u16, Instant>,
) {
    if !is_target_foreground(runtime) {
        if let Ok(mut pressed) = runtime.pressed_keys.lock() {
            pressed.clear();
        }
        next_fire.clear();
        return;
    }

    let configs = runtime
        .configs
        .lock()
        .map(|configs| configs.clone())
        .unwrap_or_default();
    let pressed = runtime
        .pressed_keys
        .lock()
        .map(|keys| keys.clone())
        .unwrap_or_default();
    let active_keys: HashSet<u16> = configs
        .iter()
        .filter(|config| pressed.contains(&config.scan_code))
        .map(|config| config.scan_code)
        .collect();
    next_fire.retain(|scan_code, _| active_keys.contains(scan_code));

    let now = Instant::now();
    for config in configs {
        if !active_keys.contains(&config.scan_code) {
            continue;
        }
        let deadline = next_fire.entry(config.scan_code).or_insert(now);
        if now >= *deadline {
            send_key(context, device, config.scan_code);
            *deadline = Instant::now() + Duration::from_millis(config.public.interval_ms);
        }
    }
}

fn send_key(context: &InterceptionApi, device: Device, scan_code: u16) {
    let down = InterceptionKeyStroke {
        code: scan_code,
        state: INTERCEPTION_KEY_DOWN,
        information: 0,
    };
    unsafe {
        (context.send)(context.context, device, &down, 1);
    }
    thread::sleep(Duration::from_millis(KEY_PRESS_DURATION_MS));
    let up = InterceptionKeyStroke {
        code: scan_code,
        state: INTERCEPTION_KEY_UP,
        information: 0,
    };
    unsafe {
        (context.send)(context.context, device, &up, 1);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_runtime(target_pid: u32) -> RapidFireRuntime {
        RapidFireRuntime {
            configs: Mutex::new(Vec::new()),
            pressed_keys: Mutex::new(HashSet::new()),
            target_pid: Arc::new(AtomicU32::new(target_pid)),
            hook_ready: AtomicBool::new(false),
            hook_error: Mutex::new(None),
        }
    }

    #[test]
    fn normalizes_ascii_letter() {
        let (key, _, scan_code) = normalize_key("j").expect("J should be supported");
        assert_eq!(key, "J");
        assert_eq!(scan_code, 0x24);
    }

    #[test]
    fn rejects_unsupported_key_values_and_intervals() {
        assert!(normalize_key("").is_err());
        assert!(normalize_key("AB").is_err());
        assert!(normalize_key("空").is_err());
        assert!(resolve_config("J".to_string(), 0).is_err());
        assert!(resolve_config("J".to_string(), 10_001).is_err());
    }

    #[test]
    fn rejects_unbound_or_non_foreground_process() {
        assert!(!is_target_foreground(&test_runtime(0)));
        assert!(!is_target_foreground(&test_runtime(u32::MAX)));
    }
}
