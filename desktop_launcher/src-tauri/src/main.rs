#![cfg_attr(target_os = "windows", windows_subsystem = "windows")]

use base64::{Engine as _, engine::general_purpose::STANDARD as BASE64};
use keyring::{Entry, Error as KeyringError};
use md5::{Digest, Md5};
use serde::{Deserialize, Serialize};
use std::fs;
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Child, Command};
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};
use tauri::{
    AppHandle, Manager, WindowEvent,
    image::Image,
    menu::{MenuBuilder, MenuEvent},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
};

#[cfg(target_os = "windows")]
mod rapid_fire;

#[cfg(target_os = "windows")]
use std::os::windows::process::CommandExt;

const CREDENTIAL_SERVICE: &str = "com.dnf.adventurer.launcher";
const CREDENTIAL_USER: &str = "saved-login";
const BACKGROUND_DIRECTORY: &str = "start/backgrounds";
const MAX_BACKGROUND_SIZE: u64 = 20 * 1024 * 1024;
#[cfg(target_os = "windows")]
const RAPID_FIRE_CONFIG_FILE_NAME: &str = "rapid-fire.json";
#[cfg(target_os = "windows")]
const CREATE_NO_WINDOW: u32 = 0x08000000;
const CLIENT_PVF_MISMATCH_ERROR: &str = "CLIENT_PVF_MISMATCH";

#[derive(Serialize, Deserialize)]
struct SavedLogin {
    account: String,
    password: String,
}

struct GameProcess {
    child: Mutex<Option<Child>>,
    target_pid: Arc<AtomicU32>,
}

#[derive(Default)]
struct AppLifecycle {
    is_quitting: AtomicBool,
}

impl Default for GameProcess {
    fn default() -> Self {
        Self {
            child: Mutex::new(None),
            target_pid: Arc::new(AtomicU32::new(0)),
        }
    }
}

fn executable_path() -> Result<PathBuf, String> {
    std::env::current_exe().map_err(|error| error.to_string())
}

fn launcher_dir() -> Result<PathBuf, String> {
    executable_path()?
        .parent()
        .map(Path::to_path_buf)
        .ok_or_else(|| "无法获取登录器所在目录".to_string())
}

#[cfg(target_os = "windows")]
fn rapid_fire_config_path() -> Result<PathBuf, String> {
    let local_app_data = std::env::var_os("LOCALAPPDATA")
        .map(PathBuf::from)
        .ok_or_else(|| "无法读取 LOCALAPPDATA 目录".to_string())?;
    let config_dir = local_app_data.join("DNFLauncher");
    fs::create_dir_all(&config_dir)
        .map_err(|error| format!("创建按键连发配置目录失败: {}", error))?;
    Ok(config_dir.join(RAPID_FIRE_CONFIG_FILE_NAME))
}

fn launcher_window_title() -> Result<String, String> {
    Ok("地下城与勇士".to_string())
}

fn show_main_window(app: &AppHandle) {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

fn create_tray(app: &AppHandle) -> tauri::Result<()> {
    let menu = MenuBuilder::new(app)
        .text("show-main", "打开主界面")
        .text("quit", "退出")
        .build()?;
    let icon = Image::from_bytes(include_bytes!("../icons/icon.ico"))?;

    TrayIconBuilder::with_id("main-tray")
        .icon(icon)
        .tooltip("地下城与勇士")
        .menu(&menu)
        .show_menu_on_left_click(false)
        .build(app)?;

    Ok(())
}

fn handle_tray_menu_event(app: &AppHandle, event: MenuEvent) {
    match event.id().as_ref() {
        "show-main" => show_main_window(app),
        "quit" => {
            if let Some(lifecycle) = app.try_state::<AppLifecycle>() {
                lifecycle.is_quitting.store(true, Ordering::Release);
            }
            app.exit(0);
        }
        _ => {}
    }
}

fn handle_tray_icon_event(app: &AppHandle, event: TrayIconEvent) {
    match event {
        TrayIconEvent::Click {
            button: MouseButton::Left,
            button_state: MouseButtonState::Up,
            ..
        }
        | TrayIconEvent::DoubleClick {
            button: MouseButton::Left,
            ..
        } => show_main_window(app),
        _ => {}
    }
}

#[tauri::command]
fn get_launcher_background() -> Result<Option<String>, String> {
    let directory = launcher_dir()?.join(BACKGROUND_DIRECTORY);
    if !directory.is_dir() {
        return Ok(None);
    }
    let mut candidates = Vec::new();
    for entry in
        fs::read_dir(&directory).map_err(|error| format!("读取背景图目录失败: {}", error))?
    {
        let path = entry.map_err(|error| error.to_string())?.path();
        if !path.is_file() {
            continue;
        }
        let extension = path
            .extension()
            .map(|value| value.to_string_lossy().to_ascii_lowercase())
            .unwrap_or_default();
        if !matches!(extension.as_str(), "jpg" | "jpeg" | "png" | "webp") {
            continue;
        }
        let file_size = path.metadata().map_err(|error| error.to_string())?.len();
        if file_size == 0 || file_size > MAX_BACKGROUND_SIZE {
            continue;
        }
        candidates.push((path, extension));
    }
    if candidates.is_empty() {
        return Ok(None);
    }
    candidates.sort_by(|left, right| left.0.cmp(&right.0));
    let seed = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|error| error.to_string())?
        .as_nanos();
    let (path, extension) = &candidates[(seed % candidates.len() as u128) as usize];
    let bytes = fs::read(path).map_err(|error| format!("读取背景图失败: {}", error))?;
    let mime = match extension.as_str() {
        "png" => "image/png",
        "webp" => "image/webp",
        _ => "image/jpeg",
    };
    Ok(Some(format!(
        "data:{};base64,{}",
        mime,
        BASE64.encode(bytes)
    )))
}

#[tauri::command]
fn get_launcher_window_title() -> Result<String, String> {
    launcher_window_title()
}

fn credential_entry() -> Result<Entry, String> {
    Entry::new(CREDENTIAL_SERVICE, CREDENTIAL_USER).map_err(|error| error.to_string())
}

#[tauri::command]
fn save_saved_login(account: String, password: String) -> Result<(), String> {
    if account.is_empty() || password.is_empty() {
        return Err("account and password are required".to_string());
    }
    let value = serde_json::to_string(&SavedLogin { account, password })
        .map_err(|error| error.to_string())?;
    credential_entry()?
        .set_password(&value)
        .map_err(|error| error.to_string())
}

#[tauri::command]
fn load_saved_login() -> Result<Option<SavedLogin>, String> {
    match credential_entry()?.get_password() {
        Ok(value) => serde_json::from_str(&value)
            .map(Some)
            .map_err(|error| error.to_string()),
        Err(KeyringError::NoEntry) => Ok(None),
        Err(error) => Err(error.to_string()),
    }
}

#[tauri::command]
fn clear_saved_login() -> Result<(), String> {
    match credential_entry()?.delete_credential() {
        Ok(()) | Err(KeyringError::NoEntry) => Ok(()),
        Err(error) => Err(error.to_string()),
    }
}

fn find_game_executable(directory: &Path) -> Result<PathBuf, String> {
    for name in ["DNF.exe", "dnf.exe"] {
        let candidate = directory.join(name);
        if candidate.is_file() {
            return Ok(candidate);
        }
    }
    Err("未在登录器同目录找到 DNF.exe".to_string())
}

fn file_md5(path: &Path) -> Result<String, String> {
    let mut file = fs::File::open(path).map_err(|error| error.to_string())?;
    let mut digest = Md5::new();
    let mut buffer = vec![0_u8; 1024 * 1024];
    loop {
        let size = file.read(&mut buffer).map_err(|error| error.to_string())?;
        if size == 0 {
            break;
        }
        digest.update(&buffer[..size]);
    }
    Ok(format!("{:X}", digest.finalize()))
}

fn verify_client_pvf(directory: &Path, expected_md5: &str) -> Result<(), String> {
    let expected = expected_md5.trim().to_ascii_uppercase();
    if expected.is_empty() {
        return Ok(());
    }
    let path = directory.join("Script.pvf");
    if !path.is_file() {
        return Err(format!(
            "{}:客户端 Script.pvf 不存在",
            CLIENT_PVF_MISMATCH_ERROR
        ));
    }
    let actual = file_md5(&path).map_err(|error| format!("读取 Script.pvf 失败: {}", error))?;
    if actual != expected {
        return Err(format!(
            "{}:客户端 Script.pvf 校验失败",
            CLIENT_PVF_MISMATCH_ERROR
        ));
    }
    Ok(())
}

#[tauri::command]
fn open_url(url: String) -> Result<(), String> {
    let trimmed = url.trim();
    if !(trimmed.starts_with("http://") || trimmed.starts_with("https://")) {
        return Err("下载地址无效".to_string());
    }
    #[cfg(target_os = "windows")]
    {
        Command::new("cmd")
            .args(["/C", "start", "", trimmed])
            .creation_flags(CREATE_NO_WINDOW)
            .spawn()
            .map_err(|error| format!("打开下载地址失败: {}", error))?;
        return Ok(());
    }
    #[allow(unreachable_code)]
    Err("当前系统不支持打开下载地址".to_string())
}

#[tauri::command]
fn launch_game(
    state: tauri::State<'_, GameProcess>,
    dnf_token: String,
    expected_pvf_md5: String,
) -> Result<(), String> {
    if dnf_token.len() < 16
        || dnf_token.len() > 4096
        || !dnf_token
            .chars()
            .all(|character| character.is_ascii_alphanumeric() || "+/=".contains(character))
    {
        return Err("服务器返回的 DNF 登录参数无效".to_string());
    }

    let directory = launcher_dir()?;
    verify_client_pvf(&directory, &expected_pvf_md5)?;
    let game_executable = find_game_executable(&directory)?;
    let mut command = Command::new(&game_executable);
    command.arg(dnf_token).current_dir(&directory);
    #[cfg(target_os = "windows")]
    command.creation_flags(CREATE_NO_WINDOW);
    let mut process = state
        .child
        .lock()
        .map_err(|_| "无法读取 DNF.exe 运行状态".to_string())?;
    if let Some(child) = process.as_mut()
        && child
            .try_wait()
            .map_err(|error| error.to_string())?
            .is_none()
    {
        return Err("DNF.exe 已在运行".to_string());
    }
    let child = command
        .spawn()
        .map_err(|error| format!("启动 DNF.exe 失败: {}", error))?;
    state.target_pid.store(child.id(), Ordering::Release);
    *process = Some(child);
    Ok(())
}

#[tauri::command]
fn is_game_running(state: tauri::State<'_, GameProcess>) -> Result<bool, String> {
    let mut process = state
        .child
        .lock()
        .map_err(|_| "无法读取 DNF.exe 运行状态".to_string())?;
    let running = match process.as_mut() {
        Some(child) => child
            .try_wait()
            .map_err(|error| error.to_string())?
            .is_none(),
        None => false,
    };
    if !running {
        process.take();
        state.target_pid.store(0, Ordering::Release);
    }
    Ok(running)
}

#[tauri::command]
fn stop_game(state: tauri::State<'_, GameProcess>) -> Result<(), String> {
    let mut process = state
        .child
        .lock()
        .map_err(|_| "无法读取 DNF.exe 运行状态".to_string())?;
    let Some(child) = process.as_mut() else {
        return Err("DNF.exe 未运行".to_string());
    };
    if child
        .try_wait()
        .map_err(|error| error.to_string())?
        .is_some()
    {
        process.take();
        state.target_pid.store(0, Ordering::Release);
        return Ok(());
    }
    child
        .kill()
        .map_err(|error| format!("结束 DNF.exe 失败: {}", error))?;
    child
        .wait()
        .map_err(|error| format!("等待 DNF.exe 退出失败: {}", error))?;
    process.take();
    state.target_pid.store(0, Ordering::Release);
    Ok(())
}

fn main() {
    let game_process = GameProcess::default();
    #[cfg(target_os = "windows")]
    let rapid_fire_state = rapid_fire::RapidFireState::new(
        rapid_fire_config_path().expect("failed to resolve rapid fire config path"),
        Arc::clone(&game_process.target_pid),
    )
    .expect("failed to initialize rapid fire");

    tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| {
            show_main_window(app);
        }))
        .manage(game_process)
        .manage(AppLifecycle::default())
        .manage(rapid_fire_state)
        .setup(|app| {
            create_tray(app.handle())?;
            Ok(())
        })
        .on_menu_event(handle_tray_menu_event)
        .on_tray_icon_event(handle_tray_icon_event)
        .on_window_event(|window, event| {
            if window.label() != "main" {
                return;
            }
            if let WindowEvent::CloseRequested { api, .. } = event {
                let should_quit = window
                    .app_handle()
                    .try_state::<AppLifecycle>()
                    .map(|lifecycle| lifecycle.is_quitting.load(Ordering::Acquire))
                    .unwrap_or(false);
                if !should_quit {
                    api.prevent_close();
                    let _ = window.hide();
                }
            }
        })
        .invoke_handler(tauri::generate_handler![
            get_launcher_window_title,
            get_launcher_background,
            save_saved_login,
            load_saved_login,
            clear_saved_login,
            open_url,
            launch_game,
            is_game_running,
            stop_game,
            rapid_fire::install_interception_driver,
            rapid_fire::list_rapid_fire,
            rapid_fire::add_rapid_fire,
            rapid_fire::remove_rapid_fire
        ])
        .run(tauri::generate_context!())
        .expect("failed to run DNF desktop launcher");
}
