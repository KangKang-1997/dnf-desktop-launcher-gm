import { invoke } from "@tauri-apps/api/core";
import { getCurrentWindow } from "@tauri-apps/api/window";
import { WebviewWindow } from "@tauri-apps/api/webviewWindow";
import "./styles.css";

const body = document.body;
const backgroundArt = document.querySelector(".background-art");
const loginView = document.querySelector("#loginView");
const homeView = document.querySelector("#homeView");
const accountSummary = document.querySelector("#accountSummary");
const accountSummaryPrefix = document.querySelector("#accountSummaryPrefix");
const currentAccount = document.querySelector("#currentAccount");
const loginForm = document.querySelector("#loginForm");
const loginAccount = document.querySelector("#loginAccount");
const loginPassword = document.querySelector("#loginPassword");
const loginSubmit = document.querySelector("#loginSubmit");
const authStatus = document.querySelector("#authStatus");
const rememberPassword = document.querySelector("#rememberPassword");
const registerForm = document.querySelector("#registerForm");
const registerAccount = document.querySelector("#registerAccount");
const registerPassword = document.querySelector("#registerPassword");
const registerConfirmPassword = document.querySelector("#registerConfirmPassword");
const registerQq = document.querySelector("#registerQq");
const registerSubmit = document.querySelector("#registerSubmit");
const registerStatus = document.querySelector("#registerStatus");
const authTitle = document.querySelector("#authTitle");
const sideDrawer = document.querySelector("#sideDrawer");
const drawerToggle = document.querySelector("#drawerToggle");
const toast = document.querySelector("#toast");
const posterCarousel = document.querySelector("#posterCarousel");
const homeEyebrow = document.querySelector("#homeEyebrow");
const homeTitle = document.querySelector("#homeTitle");
const announcementList = document.querySelector("#announcementList");
const announcementOverlay = document.querySelector("#announcementOverlay");
const announcementTitle = document.querySelector("#announcementTitle");
const announcementContent = document.querySelector("#announcementContent");
const passwordOverlay = document.querySelector("#passwordOverlay");
const passwordResetForm = document.querySelector("#passwordResetForm");
const passwordPanelTitle = document.querySelector("#passwordPanelTitle");
const passwordPanelHint = document.querySelector("#passwordPanelHint");
const accountDetails = document.querySelector("#accountDetails");
const accountDetailUid = document.querySelector("#accountDetailUid");
const logoutButton = document.querySelector("#logoutButton");
const resetAccountField = document.querySelector("#resetAccountField");
const resetAccount = document.querySelector("#resetAccount");
const resetQq = document.querySelector("#resetQq");
const resetVerificationLabel = document.querySelector("#resetVerificationLabel");
const resetNewPassword = document.querySelector("#resetNewPassword");
const resetConfirmPassword = document.querySelector("#resetConfirmPassword");
const resetPasswordStatus = document.querySelector("#resetPasswordStatus");
const resetPasswordSubmit = document.querySelector("#resetPasswordSubmit");
const launchButton = document.querySelector("#launchButton");
const launchButtonTitle = document.querySelector("#launchButtonTitle");
const launchButtonSubtitle = document.querySelector("#launchButtonSubtitle");
const clientState = document.querySelector("#clientState");
const announcementPanel = document.querySelector("#announcementPanel");
const launchArea = document.querySelector("#launchArea");
const characterWorkspace = document.querySelector("#characterWorkspace");
const characterSearch = document.querySelector("#characterSearch");
const characterRefresh = document.querySelector("#characterRefresh");
const characterScope = document.querySelector("#characterScope");
const characterCount = document.querySelector("#characterCount");
const characterStatus = document.querySelector("#characterStatus");
const characterList = document.querySelector("#characterList");
const characterPrev = document.querySelector("#characterPrev");
const characterNext = document.querySelector("#characterNext");
const characterPage = document.querySelector("#characterPage");
const characterDetailEmpty = document.querySelector("#characterDetailEmpty");
const characterDetailBody = document.querySelector("#characterDetailBody");
const characterDetailName = document.querySelector("#characterDetailName");
const characterDetailState = document.querySelector("#characterDetailState");
const characterDetailNo = document.querySelector("#characterDetailNo");
const characterDetailLevel = document.querySelector("#characterDetailLevel");
const characterDetailJob = document.querySelector("#characterDetailJob");
const characterDetailExpertJob = document.querySelector("#characterDetailExpertJob");
const characterDetailPvpGrade = document.querySelector("#characterDetailPvpGrade");
const characterDetailPvpPoint = document.querySelector("#characterDetailPvpPoint");
const characterEditOpen = document.querySelector("#characterEditOpen");
const characterEditorOverlay = document.querySelector("#characterEditorOverlay");
const characterEditorClose = document.querySelector("#characterEditorClose");
const characterEditorIdentity = document.querySelector("#characterEditorIdentity");
const characterEditLevel = document.querySelector("#characterEditLevel");
const characterEditLevelSubmit = document.querySelector("#characterEditLevelSubmit");
const characterEditPvpGrade = document.querySelector("#characterEditPvpGrade");
const characterEditPvpGradeSubmit = document.querySelector("#characterEditPvpGradeSubmit");
const characterEditPvpPoint = document.querySelector("#characterEditPvpPoint");
const characterEditPvpPointSubmit = document.querySelector("#characterEditPvpPointSubmit");
const characterEditJob = document.querySelector("#characterEditJob");
const characterEditGrowType = document.querySelector("#characterEditGrowType");
const characterEditWakeFlag = document.querySelector("#characterEditWakeFlag");
const characterEditExpertJob = document.querySelector("#characterEditExpertJob");
const characterEditJobSubmit = document.querySelector("#characterEditJobSubmit");
const characterDelete = document.querySelector("#characterDelete");
const characterRecover = document.querySelector("#characterRecover");
const characterEditStatus = document.querySelector("#characterEditStatus");
const mailWorkspace = document.querySelector("#mailWorkspace");
const mailForm = document.querySelector("#mailForm");
const mailScope = document.querySelector("#mailScope");
const mailCharacterCount = document.querySelector("#mailCharacterCount");
const mailSelectedRecipient = document.querySelector("#mailSelectedRecipient");
const mailRecipientSearch = document.querySelector("#mailRecipientSearch");
const mailMessage = document.querySelector("#mailMessage");
const mailAttachmentType = document.querySelector("#mailAttachmentType");
const mailItemId = document.querySelector("#mailItemId");
const mailItemCount = document.querySelector("#mailItemCount");
const mailForgeLevel = document.querySelector("#mailForgeLevel");
const mailItemCountLabel = document.querySelector("#mailItemCountLabel");
const mailGold = document.querySelector("#mailGold");
const mailEquipmentFields = document.querySelector("#mailEquipmentFields");
const mailEnhancementLevel = document.querySelector("#mailEnhancementLevel");
const mailAmplifyOption = document.querySelector("#mailAmplifyOption");
const mailAmplifyValue = document.querySelector("#mailAmplifyValue");
const mailSend = document.querySelector("#mailSend");
const mailDelete = document.querySelector("#mailDelete");
const mailGlobalActions = document.querySelector("#mailGlobalActions");
const mailSendAll = document.querySelector("#mailSendAll");
const mailDeleteAll = document.querySelector("#mailDeleteAll");
const mailStatus = document.querySelector("#mailStatus");
const mailCharacterStatus = document.querySelector("#mailCharacterStatus");
const mailCharacterList = document.querySelector("#mailCharacterList");
const mailCharacterPrev = document.querySelector("#mailCharacterPrev");
const mailCharacterNext = document.querySelector("#mailCharacterNext");
const mailCharacterPage = document.querySelector("#mailCharacterPage");
const mailItemStatus = document.querySelector("#mailItemStatus");
const mailItemKeyword = document.querySelector("#mailItemKeyword");
const mailSearchItem = document.querySelector("#mailSearchItem");
const mailItemResults = document.querySelector("#mailItemResults");
const mailItemPrev = document.querySelector("#mailItemPrev");
const mailItemNext = document.querySelector("#mailItemNext");
const mailItemPage = document.querySelector("#mailItemPage");
const mailConfirmOverlay = document.querySelector("#mailConfirmOverlay");
const mailConfirmText = document.querySelector("#mailConfirmText");
const mailConfirmCancel = document.querySelector("#mailConfirmCancel");
const mailConfirmSubmit = document.querySelector("#mailConfirmSubmit");
const inventoryWorkspace = document.querySelector("#inventoryWorkspace");
const inventoryScopeSummary = document.querySelector("#inventoryScopeSummary");
const inventoryItemCount = document.querySelector("#inventoryItemCount");
const inventoryCharacterSearch = document.querySelector("#inventoryCharacterSearch");
const inventoryScope = document.querySelector("#inventoryScope");
const inventoryQuery = document.querySelector("#inventoryQuery");
const inventoryCharacterStatus = document.querySelector("#inventoryCharacterStatus");
const inventoryCharacterList = document.querySelector("#inventoryCharacterList");
const inventoryCharacterPrev = document.querySelector("#inventoryCharacterPrev");
const inventoryCharacterNext = document.querySelector("#inventoryCharacterNext");
const inventoryCharacterPage = document.querySelector("#inventoryCharacterPage");
const inventorySelectedCharacter = document.querySelector("#inventorySelectedCharacter");
const inventoryDelete = document.querySelector("#inventoryDelete");
const inventoryClear = document.querySelector("#inventoryClear");
const inventoryStatus = document.querySelector("#inventoryStatus");
const inventoryList = document.querySelector("#inventoryList");
const inventoryConfirmOverlay = document.querySelector("#inventoryConfirmOverlay");
const inventoryConfirmTitle = document.querySelector("#inventoryConfirmTitle");
const inventoryConfirmText = document.querySelector("#inventoryConfirmText");
const inventoryConfirmCancel = document.querySelector("#inventoryConfirmCancel");
const inventoryConfirmSubmit = document.querySelector("#inventoryConfirmSubmit");
const avatarWorkspace = document.querySelector("#avatarWorkspace");
const avatarSelectionSummary = document.querySelector("#avatarSelectionSummary");
const avatarItemCount = document.querySelector("#avatarItemCount");
const avatarCharacterSearch = document.querySelector("#avatarCharacterSearch");
const avatarHiddenOption = document.querySelector("#avatarHiddenOption");
const avatarQuery = document.querySelector("#avatarQuery");
const avatarApply = document.querySelector("#avatarApply");
const avatarCharacterStatus = document.querySelector("#avatarCharacterStatus");
const avatarCharacterList = document.querySelector("#avatarCharacterList");
const avatarCharacterPrev = document.querySelector("#avatarCharacterPrev");
const avatarCharacterNext = document.querySelector("#avatarCharacterNext");
const avatarCharacterPage = document.querySelector("#avatarCharacterPage");
const avatarSelectedCharacter = document.querySelector("#avatarSelectedCharacter");
const avatarStatus = document.querySelector("#avatarStatus");
const avatarList = document.querySelector("#avatarList");
const rechargeWorkspace = document.querySelector("#rechargeWorkspace");
const rechargeAccount = document.querySelector("#rechargeAccount");
const rechargeCera = document.querySelector("#rechargeCera");
const rechargeCeraPoint = document.querySelector("#rechargeCeraPoint");
const rechargeRefresh = document.querySelector("#rechargeRefresh");
const rechargeForm = document.querySelector("#rechargeForm");
const rechargeType = document.querySelector("#rechargeType");
const rechargeAction = document.querySelector("#rechargeAction");
const rechargeAmount = document.querySelector("#rechargeAmount");
const rechargeSubmit = document.querySelector("#rechargeSubmit");
const rechargeStatus = document.querySelector("#rechargeStatus");
const rapidFireWorkspace = document.querySelector("#rapidFireWorkspace");
const rapidFireNativeStatus = document.querySelector("#rapidFireNativeStatus");
const rapidFireCount = document.querySelector("#rapidFireCount");
const rapidFireForm = document.querySelector("#rapidFireForm");
const rapidFireKey = document.querySelector("#rapidFireKey");
const rapidFireInterval = document.querySelector("#rapidFireInterval");
const rapidFireAdd = document.querySelector("#rapidFireAdd");
const rapidFireInstallDriver = document.querySelector("#rapidFireInstallDriver");
const rapidFireStatus = document.querySelector("#rapidFireStatus");
const rapidFireList = document.querySelector("#rapidFireList");
const gmTargetBar = document.querySelector("#gmTargetBar");
const gmTargetType = document.querySelector("#gmTargetType");
const gmTargetInput = document.querySelector("#gmTargetInput");
const gmTargetResolve = document.querySelector("#gmTargetResolve");
const gmTargetSelected = document.querySelector("#gmTargetSelected");
const eventWorkspace = document.querySelector("#eventWorkspace");
const eventForm = document.querySelector("#eventForm");
const eventSelect = document.querySelector("#eventSelect");
const eventParam1 = document.querySelector("#eventParam1");
const eventParam2 = document.querySelector("#eventParam2");
const eventRefresh = document.querySelector("#eventRefresh");
const eventAdd = document.querySelector("#eventAdd");
const eventDelete = document.querySelector("#eventDelete");
const eventStatus = document.querySelector("#eventStatus");
const eventCount = document.querySelector("#eventCount");
const eventList = document.querySelector("#eventList");
const banWorkspace = document.querySelector("#banWorkspace");
const banTarget = document.querySelector("#banTarget");
const banState = document.querySelector("#banState");
const banStateView = document.querySelector("#banStateView");
const banTypeView = document.querySelector("#banTypeView");
const banStartView = document.querySelector("#banStartView");
const banEndView = document.querySelector("#banEndView");
const banReasonView = document.querySelector("#banReasonView");
const banQuery = document.querySelector("#banQuery");
const banForm = document.querySelector("#banForm");
const banPunishType = document.querySelector("#banPunishType");
const banDays = document.querySelector("#banDays");
const banReason = document.querySelector("#banReason");
const banUnban = document.querySelector("#banUnban");
const banSubmit = document.querySelector("#banSubmit");
const banStatus = document.querySelector("#banStatus");
const permissionWorkspace = document.querySelector("#permissionWorkspace");
const permissionSearchForm = document.querySelector("#permissionSearchForm");
const permissionKeyword = document.querySelector("#permissionKeyword");
const permissionAccountCount = document.querySelector("#permissionAccountCount");
const permissionAccountList = document.querySelector("#permissionAccountList");
const permissionForm = document.querySelector("#permissionForm");
const permissionSelectedAccount = document.querySelector("#permissionSelectedAccount");
const permissionSave = document.querySelector("#permissionSave");
const permissionGrid = document.querySelector("#permissionGrid");
const permissionStatus = document.querySelector("#permissionStatus");
const systemWorkspace = document.querySelector("#systemWorkspace");
const systemHomePanel = document.querySelector("#systemHomePanel");
const systemHomeTitle = document.querySelector("#systemHomeTitle");
const systemHomeEyebrow = document.querySelector("#systemHomeEyebrow");
const systemClientDownloadUrl = document.querySelector("#systemClientDownloadUrl");
const announcementAdd = document.querySelector("#announcementAdd");
const announcementEditorList = document.querySelector("#announcementEditorList");
const systemHomeStatus = document.querySelector("#systemHomeStatus");
const systemHomeSave = document.querySelector("#systemHomeSave");
const pvfRefreshForm = document.querySelector("#pvfRefreshForm");
const pvfPath = document.querySelector("#pvfPath");
const pvfEncode = document.querySelector("#pvfEncode");
const pvfRefresh = document.querySelector("#pvfRefresh");
const pvfStatus = document.querySelector("#pvfStatus");
const pvfLog = document.querySelector("#pvfLog");
const pvfMd5Form = document.querySelector("#pvfMd5Form");
const pvfClientMd5 = document.querySelector("#pvfClientMd5");
const pvfMd5Save = document.querySelector("#pvfMd5Save");
const pvfMd5Status = document.querySelector("#pvfMd5Status");
const logSearchForm = document.querySelector("#logSearchForm");
const logKeyword = document.querySelector("#logKeyword");
const logList = document.querySelector("#logList");
const logPrev = document.querySelector("#logPrev");
const logNext = document.querySelector("#logNext");
const logPage = document.querySelector("#logPage");
const gmConfirmOverlay = document.querySelector("#gmConfirmOverlay");
const gmConfirmTitle = document.querySelector("#gmConfirmTitle");
const gmConfirmText = document.querySelector("#gmConfirmText");
const gmConfirmCancel = document.querySelector("#gmConfirmCancel");
const gmConfirmSubmit = document.querySelector("#gmConfirmSubmit");

const apiBase = __DNF_LAUNCHER_API_BASE__;
const isTauri = "__TAURI_INTERNALS__" in window;
const defaultHome = {
  home_title: "欢迎回来，勇士",
  home_eyebrow: "冒险准备完成",
  client_download_url: "",
  announcements: [
    {
      title: "版本更新公告标题占位",
      summary: "后续可替换为具体版本更新内容",
      content: "这里用于展示版本更新全文，管理员可在系统页修改。",
      poster_url: "/api/posters/sample-1",
    },
    {
      title: "客户端下载说明占位",
      summary: "后续可替换为客户端下载说明",
      content: "这里用于展示客户端下载说明全文，管理员可在系统页修改。",
      poster_url: "/api/posters/sample-2",
    },
    {
      title: "活动与维护通知占位",
      summary: "后续可替换为活动、维护或补偿通知",
      content: "这里用于展示活动、维护或补偿通知全文，管理员可在系统页修改。",
      poster_url: "/api/posters/sample-3",
    },
  ],
};
const sectionPermissions = {
  "邮件": "gm.mail",
  "充值": "gm.cera.charge",
  "角色": "gm.character.edit",
  "背包": "gm.inventory",
  "活动": "gm.event.manage",
  "时装潜能": "gm.avatar.edit",
};
const gmTargetSections = new Set(["角色", "邮件", "充值", "背包", "时装潜能", "封禁"]);
const permissionNames = {
  "gm.mail": "邮件管理",
  "gm.cera.charge": "账号充值",
  "gm.character.edit": "角色修改",
  "gm.inventory": "背包管理",
  "gm.event.manage": "活动管理",
  "gm.avatar.edit": "时装潜能",
};

let accessToken = "";
let currentUser = null;
let toastTimer = 0;
let posterIndex = 0;
let posterTimer = 0;
let rememberedAccount = "";
let passwordPanelMode = "forgot";
let homeSettings = defaultHome;
let gameMonitorTimer = 0;
let gameRunning = false;
let clientUpdateRequired = false;
let clientDownloadUrl = "";
let activeSection = "大厅";
let characterJobOptions = null;
let characterJobOptionsLoading = null;
let inventoryConfirmMode = "";
let avatarOptionsLoaded = false;
let avatarOptionsLoading = null;
const characterState = {
  page: 1,
  limit: 8,
  total: 0,
  characters: [],
  loaded: false,
  loading: false,
  selectedCharacterNo: null,
};
const mailState = {
  page: 1,
  limit: 8,
  total: 0,
  characters: [],
  loaded: false,
  loading: false,
  selectedCharacterNo: null,
  itemPage: 1,
  itemLimit: 10,
  itemTotal: 0,
  selectedItemStackLimit: null,
};
const inventoryState = {
  page: 1,
  limit: 8,
  total: 0,
  characters: [],
  loaded: false,
  loading: false,
  selectedCharacterNo: null,
  selectedSlot: null,
  items: [],
};
const avatarState = {
  page: 1,
  limit: 8,
  total: 0,
  characters: [],
  loaded: false,
  loading: false,
  selectedCharacterNo: null,
  selectedUiIds: [],
  items: [],
};
const rechargeState = {
  loading: false,
};
const rapidFireState = {
  configs: [],
  loaded: false,
  loading: false,
};
const gmState = {
  target: null,
  resolving: false,
  revision: 0,
  abortController: null,
};
const eventState = {
  loaded: false,
  loading: false,
  selectedLogId: null,
  available: [],
  running: [],
};
const permissionState = {
  loaded: false,
  loading: false,
  all: [],
  accounts: [],
  selected: null,
};
const systemState = {
  loaded: false,
  loading: false,
  tab: "home",
  announcements: [],
  logPage: 1,
  logLimit: 30,
  logTotal: 0,
};
let gmConfirmAction = null;

const apiMessages = {
  "Invalid account or password": "账号或密码错误",
  "Account name must use ASCII characters": "账号只能使用英文字符",
  "Account name must contain digits only": "账号只能包含数字",
  "QQ must contain digits only": "QQ 只能包含数字",
  "Account or QQ is incorrect": "账号或注册 QQ 不正确",
  "Missing bearer token": "请先登录",
  "Invalid session token": "登录状态无效，请重新登录",
  "Session expired": "登录已过期，请重新登录",
  "Missing character permission": "当前账号没有角色功能权限",
  "Character not found": "角色不存在或不可操作",
  "Invalid job": "职业无效",
  "Invalid grow type": "转职无效",
  "Invalid expert job": "副职业无效",
  "Invalid PVP grade": "决斗等级无效",
  "Message, item, or gold is required": "邮件内容、物品或金币至少填写一项",
  "Item count must be greater than 0": "物品数量必须大于 0",
  "Mass mail attachment rows must not exceed 10000": "全服邮件附件总计不能超过 10000 行",
  "缺少权限：gm.mail": "当前账号没有邮件管理权限",
  "Invalid inventory scope": "背包范围无效",
  "Inventory row not found": "未找到该角色的背包数据",
  "Inventory blob is empty": "背包数据为空",
  "Inventory blob decompress failed": "背包数据解压失败",
  "Slot is outside inventory range": "物品槽位超出背包范围",
  "No avatar selected": "请选择要修改的时装",
  "Invalid hidden option": "时装潜能选项无效",
  "Selected avatar does not belong to character": "所选时装不属于当前角色",
  "Invalid cera type": "充值类型无效",
  "Invalid cera action": "充值操作无效",
  "Amount must be greater than 0": "充值数量必须大于 0",
  "Account name or UID is required": "未找到当前账号信息",
  "Account not found": "当前账号不存在",
  "Client download URL must use http:// or https://": "客户端下载地址必须以 http:// 或 https:// 开头",
  "Poster URL must use http://, https://, or /api/posters/": "海报地址必须使用 http://、https:// 或 /api/posters/",
  "PVF MD5 must be 32 hex characters": "PVF MD5 必须是 32 位十六进制字符",
};

function showToast(message) {
  window.clearTimeout(toastTimer);
  toast.textContent = message;
  toast.classList.add("show");
  toastTimer = window.setTimeout(() => toast.classList.remove("show"), 2800);
}

function errorMessage(error, fallback) {
  if (error instanceof Error && error.message) return error.message;
  if (typeof error === "string" && error.trim()) return error;
  return fallback;
}

function isAdmin() {
  return currentUser?.user_type === "admin";
}

function currentOperationAccount() {
  return isAdmin() ? gmState.target : currentUser;
}

function operationAccountLabel(fallback = "当前账号") {
  const account = currentOperationAccount();
  return account?.account_name ? `账号 ${account.account_name}` : fallback;
}

function gmCharacterQuery() {
  return isAdmin() && gmState.target ? `&uid=${encodeURIComponent(gmState.target.uid)}` : "";
}

function cancelGmTargetRequests() {
  if (gmState.abortController) gmState.abortController.abort();
  gmState.abortController = new AbortController();
}

function bumpGmTargetRevision() {
  gmState.revision += 1;
  cancelGmTargetRequests();
}

function gmTargetRequestContext() {
  return {
    revision: gmState.revision,
    uid: Number(gmState.target?.uid || 0),
    signal: gmState.abortController?.signal || null,
  };
}

function isCurrentGmTargetContext(context) {
  if (!isAdmin()) return true;
  return (
    context?.revision === gmState.revision &&
    Number(gmState.target?.uid || 0) === Number(context?.uid || 0)
  );
}

function withGmTargetSignal(options, context) {
  if (!isAdmin() || !context?.signal) return options;
  return { ...options, signal: context.signal };
}

function isCanceledRequest(error) {
  return error?.name === "AbortError";
}

function resetGmTarget() {
  bumpGmTargetRevision();
  gmState.target = null;
  gmState.resolving = false;
  gmTargetType.value = "account_name";
  gmTargetInput.value = "";
  gmTargetInput.placeholder = "输入目标账号";
  gmTargetSelected.textContent = "未选择目标";
  gmTargetResolve.disabled = false;
}

function resetTargetFeatureStates() {
  resetCharacterState();
  resetMailState();
  resetInventoryState();
  resetAvatarState();
  resetRechargeState();
  resetBanState();
}

function requireGmTarget(statusElement) {
  if (!isAdmin() || gmState.target) return true;
  if (statusElement) statusElement.textContent = "请先在页面顶部选择目标账号";
  return false;
}

async function loadLauncherWindowTitle() {
  if (!isTauri) return;
  try {
    const title = String(await invoke("get_launcher_window_title") || "").trim();
    if (title) {
      document.title = title;
      await getCurrentWindow().setTitle(title);
    }
  } catch (error) {
    console.error("读取登录器窗口标题失败", error);
  }
}

async function loadLauncherBackground() {
  if (!isTauri) return;
  try {
    const dataUrl = await invoke("get_launcher_background");
    if (!dataUrl) return;
    await new Promise((resolve, reject) => {
      const image = new Image();
      const timeout = window.setTimeout(() => reject(new Error("背景图加载超时")), 8000);
      image.onload = () => {
        window.clearTimeout(timeout);
        resolve();
      };
      image.onerror = () => {
        window.clearTimeout(timeout);
        reject(new Error("背景图格式无效"));
      };
      image.src = dataUrl;
    });
    backgroundArt.style.backgroundImage = `url("${dataUrl}")`;
  } catch (error) {
    console.warn("加载客户端背景图失败，使用内置底图", error);
  }
}

async function api(path, options = {}) {
  const controller = new AbortController();
  const timeoutMs = Number(options.timeoutMs || 8000);
  let timedOut = false;
  const timeout = window.setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeoutMs);
  const externalSignal = options.signal;
  let externalAbortHandler = null;
  if (externalSignal) {
    if (externalSignal.aborted) {
      controller.abort();
    } else {
      externalAbortHandler = () => controller.abort();
      externalSignal.addEventListener("abort", externalAbortHandler, { once: true });
    }
  }
  const headers = { ...(options.headers || {}) };
  if (options.body) headers["Content-Type"] = "application/json";
  if (accessToken) headers.Authorization = `Bearer ${accessToken}`;
  const { timeoutMs: _timeoutMs, signal: _signal, ...fetchOptions } = options;

  let response;
  try {
    response = await fetch(`${apiBase}${path}`, {
      ...fetchOptions,
      headers,
      signal: controller.signal,
    });
  } catch (error) {
    if (error?.name === "AbortError") {
      if (!timedOut && externalSignal?.aborted) {
        const cancelError = new Error("请求已取消");
        cancelError.name = "AbortError";
        throw cancelError;
      }
      throw new Error("连接服务器超时");
    }
    throw new Error("无法连接服务器，请检查服务地址和网络");
  } finally {
    window.clearTimeout(timeout);
    if (externalSignal && externalAbortHandler) {
      externalSignal.removeEventListener("abort", externalAbortHandler);
    }
  }

  const text = await response.text();
  let data = {};
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { detail: text };
    }
  }
  if (!response.ok) {
    const detail = String(data.detail || `HTTP ${response.status}`);
    throw new Error(apiMessages[detail] || detail);
  }
  return data;
}

function setAuthTab(name) {
  document.querySelectorAll("[data-auth-tab]").forEach((button) => {
    button.classList.toggle("active", button.dataset.authTab === name);
  });
  loginForm.hidden = name !== "login";
  registerForm.hidden = name !== "register";
  authTitle.textContent = name === "login" ? "登录游戏账号" : "注册游戏账号";
  authStatus.textContent = "";
  registerStatus.textContent = "";
  document.querySelectorAll("[data-auth-icon]").forEach((icon) => {
    icon.toggleAttribute("hidden", icon.dataset.authIcon !== name);
  });
}

function updateNavigation(permissions) {
  document.querySelectorAll(".nav-item").forEach((button) => {
    const required = sectionPermissions[button.dataset.section];
    const adminOnly = button.hasAttribute("data-admin-only");
    const hideRapidFire = isAdmin() && button.dataset.section === "按键连发";
    button.hidden = hideRapidFire
      || (adminOnly && !isAdmin())
      || Boolean(required && !permissions.includes(required));
  });
}

function setGameButtonMode(running, options = {}) {
  gameRunning = running;
  if (!options.keepUpdateRequired) clientUpdateRequired = false;
  launchButton.classList.toggle("stop-mode", running);
  launchButtonTitle.textContent = running ? "结束游戏" : "启动游戏";
  launchButtonSubtitle.textContent = running ? "STOP GAME" : "START GAME";
}

function setClientUpdateMode() {
  gameRunning = false;
  clientUpdateRequired = true;
  launchButton.classList.remove("stop-mode");
  launchButtonTitle.textContent = "更新";
  launchButtonSubtitle.textContent = "UPDATE";
  clientState.innerHTML = "<i></i> 客户端待更新";
  launchButton.disabled = false;
}

function resetCharacterState() {
  characterState.page = 1;
  characterState.total = 0;
  characterState.characters = [];
  characterState.loaded = false;
  characterState.loading = false;
  characterState.selectedCharacterNo = null;
  characterSearch.value = "";
  characterList.replaceChildren();
  characterStatus.textContent = "进入角色页后加载角色";
  characterCount.textContent = "0 个角色";
  characterPage.textContent = "第 1 / 1 页";
  characterRefresh.disabled = false;
  characterRefresh.textContent = "刷新角色";
  characterDetailEmpty.hidden = false;
  characterDetailBody.hidden = true;
  characterEditOpen.disabled = true;
  characterEditorOverlay.hidden = true;
  characterJobOptions = null;
  characterJobOptionsLoading = null;
}

function resetMailState() {
  mailState.page = 1;
  mailState.total = 0;
  mailState.characters = [];
  mailState.loaded = false;
  mailState.loading = false;
  mailState.selectedCharacterNo = null;
  mailState.itemPage = 1;
  mailState.itemTotal = 0;
  mailState.selectedItemStackLimit = null;
  mailForm.reset();
  mailCharacterList.replaceChildren();
  mailItemResults.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "mail-empty";
  empty.textContent = "可先导入 PVF 后搜索物品";
  mailItemResults.appendChild(empty);
  mailScope.textContent = "当前账号";
  mailCharacterCount.textContent = "0 个角色";
  mailSelectedRecipient.textContent = "未选择收件角色";
  mailStatus.textContent = "搜索并选择收件角色后发送邮件";
  mailCharacterStatus.textContent = "等待加载";
  mailCharacterPage.textContent = "第 1 / 1 页";
  mailCharacterPrev.disabled = true;
  mailCharacterNext.disabled = true;
  mailItemStatus.textContent = "输入名称或 ID 搜索";
  mailItemPage.textContent = "第 1 / 1 页";
  mailItemPrev.disabled = true;
  mailItemNext.disabled = true;
  mailConfirmOverlay.hidden = true;
  updateMailAttachmentFields();
}

function resetInventoryState() {
  inventoryState.page = 1;
  inventoryState.total = 0;
  inventoryState.characters = [];
  inventoryState.loaded = false;
  inventoryState.loading = false;
  inventoryState.selectedCharacterNo = null;
  inventoryState.selectedSlot = null;
  inventoryState.items = [];
  inventoryConfirmMode = "";
  inventoryCharacterSearch.value = "";
  inventoryScope.value = "inventory";
  inventoryCharacterList.replaceChildren();
  inventoryList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "inventory-empty";
  empty.textContent = "尚未查询背包";
  inventoryList.appendChild(empty);
  inventoryScopeSummary.textContent = "物品栏";
  inventoryItemCount.textContent = "0 个物品";
  inventoryCharacterStatus.textContent = "等待加载";
  inventoryCharacterPage.textContent = "第 1 / 1 页";
  inventoryCharacterPrev.disabled = true;
  inventoryCharacterNext.disabled = true;
  inventorySelectedCharacter.textContent = "尚未选择角色";
  inventoryStatus.textContent = "选择角色后查询背包";
  inventoryDelete.disabled = true;
  inventoryConfirmOverlay.hidden = true;
}

function resetAvatarState() {
  avatarState.page = 1;
  avatarState.total = 0;
  avatarState.characters = [];
  avatarState.loaded = false;
  avatarState.loading = false;
  avatarState.selectedCharacterNo = null;
  avatarState.selectedUiIds = [];
  avatarState.items = [];
  avatarOptionsLoaded = false;
  avatarOptionsLoading = null;
  avatarCharacterSearch.value = "";
  avatarHiddenOption.replaceChildren();
  addSelectOption(avatarHiddenOption, 0, "0 - 无");
  avatarCharacterList.replaceChildren();
  avatarList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "inventory-empty";
  empty.textContent = "尚未查询时装";
  avatarList.appendChild(empty);
  avatarSelectionSummary.textContent = "未选择时装";
  avatarItemCount.textContent = "0 件时装";
  avatarCharacterStatus.textContent = "等待加载";
  avatarCharacterPage.textContent = "第 1 / 1 页";
  avatarCharacterPrev.disabled = true;
  avatarCharacterNext.disabled = true;
  avatarSelectedCharacter.textContent = "尚未选择角色";
  avatarStatus.textContent = "选择角色后查询时装";
  avatarApply.disabled = true;
}

function resetRechargeState() {
  rechargeState.loading = false;
  rechargeForm.reset();
  rechargeAccount.textContent = "-";
  rechargeCera.textContent = "-";
  rechargeCeraPoint.textContent = "-";
  rechargeStatus.textContent = "进入页面后自动查询余额";
  rechargeRefresh.disabled = false;
  rechargeSubmit.disabled = false;
}

function resetRapidFireState() {
  rapidFireState.configs = [];
  rapidFireState.loaded = false;
  rapidFireState.loading = false;
  rapidFireForm.reset();
  rapidFireInterval.value = "20";
  rapidFireNativeStatus.textContent = "等待读取";
  rapidFireCount.textContent = "0 个按键";
  rapidFireInstallDriver.hidden = true;
  rapidFireStatus.textContent = "仅游戏内生效";
  rapidFireList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "rapid-fire-empty";
  empty.textContent = "尚未添加连发按键";
  rapidFireList.appendChild(empty);
}

function resetEventState() {
  eventState.loaded = false;
  eventState.loading = false;
  eventState.selectedLogId = null;
  eventState.available = [];
  eventState.running = [];
  eventSelect.replaceChildren();
  eventParam1.value = "1";
  eventParam2.value = "0";
  eventStatus.textContent = "进入页面后加载活动";
  eventCount.textContent = "0 个运行中";
  eventDelete.disabled = true;
  eventList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "gm-admin-empty";
  empty.textContent = "当前没有运行中的活动";
  eventList.appendChild(empty);
}

function resetBanState() {
  banTarget.textContent = gmState.target?.account_name || "未选择目标账号";
  banState.textContent = "等待查询";
  banStateView.textContent = "-";
  banTypeView.textContent = "-";
  banStartView.textContent = "-";
  banEndView.textContent = "-";
  banReasonView.textContent = "-";
  banPunishType.value = "1";
  banDays.value = "365";
  banReason.value = "";
  banStatus.textContent = gmState.target
    ? "目标账号已选择，可查询或执行限制操作"
    : "请先在页面顶部选择目标账号";
}

function resetPermissionState() {
  permissionState.loaded = false;
  permissionState.loading = false;
  permissionState.all = [];
  permissionState.accounts = [];
  permissionState.selected = null;
  permissionKeyword.value = "";
  permissionAccountCount.textContent = "0 个结果";
  permissionSelectedAccount.textContent = "未选择账号";
  permissionSave.disabled = true;
  permissionStatus.textContent = "选择账号后配置可用功能";
  permissionAccountList.replaceChildren();
  permissionGrid.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "gm-admin-empty";
  empty.textContent = "进入页面后加载最近账号";
  permissionAccountList.appendChild(empty);
}

function resetSystemState() {
  systemState.loaded = false;
  systemState.loading = false;
  systemState.tab = "home";
  systemState.announcements = [];
  systemState.logPage = 1;
  systemState.logTotal = 0;
  systemHomeTitle.value = "";
  systemHomeEyebrow.value = "";
  systemClientDownloadUrl.value = "";
  announcementEditorList.replaceChildren();
  systemHomeStatus.textContent = "最多保存 8 条公告";
  pvfPath.value = "";
  pvfEncode.value = "big5";
  pvfClientMd5.value = "";
  pvfStatus.textContent = "尚未读取 PVF 状态";
  pvfMd5Status.textContent = "未启用客户端 PVF 校验";
  pvfLog.hidden = true;
  pvfLog.textContent = "";
  logKeyword.value = "";
  logList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "gm-admin-empty";
  empty.textContent = "进入日志页后加载最近操作";
  logList.appendChild(empty);
  logPage.textContent = "第 1 / 1 页";
  logPrev.disabled = true;
  logNext.disabled = true;
  setSystemTab("home", false);
}

function resetAdminStates() {
  resetEventState();
  resetPermissionState();
  resetSystemState();
  gmConfirmOverlay.hidden = true;
  gmConfirmAction = null;
}

function normalizeCharacterSearch(value) {
  return String(value || "").trim().toLowerCase();
}

function characterJobName(character) {
  return character.job_name || `${character.job}/${character.grow_type}`;
}

function formatNumber(value) {
  return new Intl.NumberFormat("zh-CN").format(Number(value || 0));
}

function addSelectOption(select, value, label) {
  const option = document.createElement("option");
  option.value = String(value);
  option.textContent = label;
  select.appendChild(option);
}

function optionLabel(value, label) {
  return label ? `${value} - ${label}` : String(value);
}

function selectedCharacter() {
  return characterState.characters.find(
    (character) => Number(character.charac_no) === characterState.selectedCharacterNo,
  ) || null;
}

function renderCharacterGrowTypes(selectedGrowType = null) {
  const jobId = Number(characterEditJob.value || "0");
  const job = (characterJobOptions?.jobs || []).find((item) => Number(item.job) === jobId);
  const growTypes = job?.grow_types || [];
  characterEditGrowType.replaceChildren();
  if (!growTypes.length) addSelectOption(characterEditGrowType, 0, "0");
  for (const growType of growTypes) {
    addSelectOption(
      characterEditGrowType,
      growType.grow_type,
      optionLabel(growType.grow_type, growType.name),
    );
  }
  if (selectedGrowType !== null && selectedGrowType !== undefined) {
    characterEditGrowType.value = String(selectedGrowType);
  }
}

function renderCharacterJobOptions(options) {
  characterJobOptions = options || { jobs: [], expert_jobs: [], pvp_ranks: [] };
  characterEditJob.replaceChildren();
  for (const job of characterJobOptions.jobs || []) {
    addSelectOption(characterEditJob, job.job, optionLabel(job.job, job.name));
  }
  if (!characterEditJob.options.length) addSelectOption(characterEditJob, 0, "0");

  characterEditExpertJob.replaceChildren();
  for (const expertJob of characterJobOptions.expert_jobs || []) {
    addSelectOption(
      characterEditExpertJob,
      expertJob.expert_job,
      optionLabel(expertJob.expert_job, expertJob.name),
    );
  }

  characterEditPvpGrade.replaceChildren();
  for (const rank of characterJobOptions.pvp_ranks || []) {
    addSelectOption(characterEditPvpGrade, rank.pvp_grade, rank.name || String(rank.pvp_grade));
  }
  if (!characterEditPvpGrade.options.length) addSelectOption(characterEditPvpGrade, 0, "10级");
  renderCharacterGrowTypes();
}

async function loadCharacterJobOptions() {
  if (characterJobOptions) return true;
  if (characterJobOptionsLoading) return characterJobOptionsLoading;
  characterJobOptionsLoading = (async () => {
    try {
      const data = await api("/api/gm/character/job-options", { method: "GET" });
      renderCharacterJobOptions(data);
      return true;
    } catch (error) {
      characterStatus.textContent = `职业选项加载失败：${errorMessage(error, "未知错误")}`;
      return false;
    } finally {
      characterJobOptionsLoading = null;
    }
  })();
  return characterJobOptionsLoading;
}

function fillCharacterEditor(character) {
  characterEditorIdentity.textContent = `${character.charac_name || character.charac_no} · ${character.charac_no}`;
  characterEditLevel.value = String(character.level ?? "");
  characterEditPvpGrade.value = String(character.pvp_grade || 0);
  characterEditPvpPoint.value = String(character.pvp_point || 0);
  characterEditJob.value = String(character.job || 0);
  renderCharacterGrowTypes(character.grow_type_base);
  characterEditGrowType.value = String(character.grow_type_base || 0);
  characterEditWakeFlag.value = String(character.wake_flag || 0);
  characterEditExpertJob.value = String(character.expert_job || 0);
  characterDelete.disabled = Boolean(character.delete_flag);
  characterRecover.disabled = !character.delete_flag;
}

function refreshCharacterInState(character) {
  if (!character?.charac_no) return;
  const index = characterState.characters.findIndex(
    (item) => Number(item.charac_no) === Number(character.charac_no),
  );
  if (index >= 0) characterState.characters[index] = character;
  selectCharacter(character);
  if (!characterEditorOverlay.hidden) fillCharacterEditor(character);
}

function setCharacterButtonBusy(button, busy, busyText = "") {
  if (busy) {
    button.dataset.idleText = button.textContent;
    button.textContent = busyText || "处理中";
  } else if (button.dataset.idleText) {
    button.textContent = button.dataset.idleText;
    delete button.dataset.idleText;
  }
  button.disabled = busy;
}

function filteredCharacters() {
  const keyword = normalizeCharacterSearch(characterSearch.value);
  if (!keyword) return characterState.characters;
  return characterState.characters.filter((character) => {
    const name = normalizeCharacterSearch(character.charac_name);
    const number = String(character.charac_no || "");
    return name.includes(keyword) || number.includes(keyword);
  });
}

function appendCharacterText(parent, tagName, className, text) {
  const element = document.createElement(tagName);
  element.className = className;
  element.textContent = text;
  parent.appendChild(element);
  return element;
}

function selectCharacter(character) {
  characterState.selectedCharacterNo = Number(character.charac_no || 0) || null;
  characterDetailEmpty.hidden = true;
  characterDetailBody.hidden = false;
  characterDetailName.textContent = character.charac_name || "-";
  characterDetailState.textContent = character.delete_flag ? "已删除" : "正常";
  characterDetailState.classList.toggle("deleted", Boolean(character.delete_flag));
  characterDetailNo.textContent = character.charac_no ?? "-";
  characterDetailLevel.textContent = character.level ?? "-";
  characterDetailJob.textContent = characterJobName(character);
  characterDetailExpertJob.textContent = character.expert_job_name || character.expert_job || "无";
  characterDetailPvpGrade.textContent = character.pvp_grade_name || character.pvp_grade || "0";
  characterDetailPvpPoint.textContent = formatNumber(character.pvp_point);
  characterEditOpen.disabled = false;
  renderCharacterPicker();
}

function renderCharacterRows(characters) {
  characterList.replaceChildren();
  if (!characters.length) {
    const empty = document.createElement("div");
    empty.className = "character-empty";
    empty.textContent = "没有找到角色";
    characterList.appendChild(empty);
    return;
  }

  for (const character of characters) {
    const button = document.createElement("button");
    button.className = "desktop-character-card";
    button.type = "button";
    button.classList.toggle("selected", Number(character.charac_no) === characterState.selectedCharacterNo);
    button.classList.toggle("deleted", Boolean(character.delete_flag));

    const heading = document.createElement("span");
    heading.className = "desktop-character-heading";
    appendCharacterText(heading, "strong", "", character.charac_name || "-");
    appendCharacterText(heading, "em", "", character.delete_flag ? "已删除" : `Lv.${character.level || 0}`);
    button.appendChild(heading);

    const meta = document.createElement("span");
    meta.className = "desktop-character-meta";
    appendCharacterText(meta, "span", "", `编号 ${character.charac_no}`);
    appendCharacterText(meta, "span", "", characterJobName(character));
    appendCharacterText(meta, "span", "", `决斗 ${character.pvp_grade_name || character.pvp_grade || "0"}`);
    button.appendChild(meta);

    button.addEventListener("click", () => selectCharacter(character));
    characterList.appendChild(button);
  }
}

function renderCharacterPicker() {
  const characters = filteredCharacters();
  characterState.total = characters.length;
  const totalPages = Math.max(1, Math.ceil(characterState.total / characterState.limit));
  if (characterState.page > totalPages) characterState.page = totalPages;
  const start = (characterState.page - 1) * characterState.limit;
  const pageCharacters = characters.slice(start, start + characterState.limit);

  characterScope.textContent = operationAccountLabel();
  characterCount.textContent = `${characterState.total} 个角色`;
  characterStatus.textContent = characterState.loaded ? `共 ${characterState.total} 个角色` : "等待加载";
  characterPage.textContent = `第 ${characterState.page} / ${totalPages} 页`;
  characterPrev.disabled = characterState.page <= 1;
  characterNext.disabled = characterState.page >= totalPages;
  renderCharacterRows(pageCharacters);
}

async function loadCharacterPicker(force = false) {
  if (!currentUser || characterState.loading) return;
  if (!requireGmTarget(characterStatus)) return;
  if (characterState.loaded && !force) {
    renderCharacterPicker();
    return;
  }

  const requestContext = gmTargetRequestContext();
  characterState.loading = true;
  characterRefresh.disabled = true;
  characterRefresh.textContent = "刷新中";
  characterStatus.textContent = "角色加载中";
  try {
    const fetchLimit = 500;
    let fetchPage = 1;
    let total = 0;
    const characters = [];
    do {
      const data = await api(
        `/api/gm/characters?page=${fetchPage}&limit=${fetchLimit}&include_deleted=1${gmCharacterQuery()}`,
        withGmTargetSignal({ method: "GET" }, requestContext),
      );
      if (!isCurrentGmTargetContext(requestContext)) return;
      characters.push(...(data.characters || []));
      total = data.total || characters.length;
      fetchPage += 1;
    } while (characters.length < total);
    if (!isCurrentGmTargetContext(requestContext)) return;
    characterState.characters = characters;
    characterState.loaded = true;
    if (force) characterState.page = 1;

    const selected = characters.find(
      (character) => Number(character.charac_no) === characterState.selectedCharacterNo,
    );
    if (selected) selectCharacter(selected);
    else {
      characterState.selectedCharacterNo = null;
      characterDetailEmpty.hidden = false;
      characterDetailBody.hidden = true;
      renderCharacterPicker();
    }
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return;
    characterStatus.textContent = `角色加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) {
      characterState.loading = false;
      characterRefresh.disabled = false;
      characterRefresh.textContent = "刷新角色";
    }
  }
}

function filteredMailCharacters() {
  const keyword = normalizeCharacterSearch(mailRecipientSearch.value);
  if (!keyword) return mailState.characters;
  return mailState.characters.filter((character) => {
    const name = normalizeCharacterSearch(character.charac_name);
    const number = String(character.charac_no || "");
    return name.includes(keyword) || number.includes(keyword);
  });
}

function selectMailCharacter(character) {
  mailState.selectedCharacterNo = Number(character.charac_no || 0) || null;
  mailRecipientSearch.value = character.charac_name || String(character.charac_no || "");
  mailSelectedRecipient.textContent = `收件人：${character.charac_name || character.charac_no}`;
  mailStatus.textContent = `已选择收件人：${character.charac_name || character.charac_no}`;
  renderMailCharacterPicker();
}

function renderMailCharacterRows(characters) {
  mailCharacterList.replaceChildren();
  if (!characters.length) {
    const empty = document.createElement("div");
    empty.className = "mail-empty";
    empty.textContent = "没有找到角色";
    mailCharacterList.appendChild(empty);
    return;
  }
  for (const character of characters) {
    const button = document.createElement("button");
    button.className = "mail-character-row";
    button.type = "button";
    button.classList.toggle(
      "selected",
      Number(character.charac_no) === mailState.selectedCharacterNo,
    );
    const name = document.createElement("strong");
    name.textContent = character.charac_name || "-";
    const detail = document.createElement("span");
    detail.textContent = `Lv.${character.level || 0} · ${characterJobName(character)}`;
    button.append(name, detail);
    button.addEventListener("click", () => selectMailCharacter(character));
    mailCharacterList.appendChild(button);
  }
}

function renderMailCharacterPicker() {
  const characters = filteredMailCharacters();
  mailState.total = characters.length;
  const totalPages = Math.max(1, Math.ceil(mailState.total / mailState.limit));
  if (mailState.page > totalPages) mailState.page = totalPages;
  const start = (mailState.page - 1) * mailState.limit;
  const pageCharacters = characters.slice(start, start + mailState.limit);
  mailScope.textContent = operationAccountLabel();
  mailCharacterCount.textContent = `${mailState.total} 个角色`;
  mailCharacterStatus.textContent = mailState.loaded ? `共 ${mailState.total} 个角色` : "等待加载";
  mailCharacterPage.textContent = `第 ${mailState.page} / ${totalPages} 页`;
  mailCharacterPrev.disabled = mailState.page <= 1;
  mailCharacterNext.disabled = mailState.page >= totalPages;
  renderMailCharacterRows(pageCharacters);
}

async function loadMailCharacters(force = false) {
  if (!currentUser || mailState.loading) return;
  if (!requireGmTarget(mailCharacterStatus)) return;
  if (mailState.loaded && !force) {
    renderMailCharacterPicker();
    return;
  }
  const requestContext = gmTargetRequestContext();
  mailState.loading = true;
  mailCharacterStatus.textContent = "角色加载中";
  try {
    const fetchLimit = 500;
    let fetchPage = 1;
    let total = 0;
    const characters = [];
    do {
      const data = await api(
        `/api/gm/characters?page=${fetchPage}&limit=${fetchLimit}${gmCharacterQuery()}`,
        withGmTargetSignal({ method: "GET" }, requestContext),
      );
      if (!isCurrentGmTargetContext(requestContext)) return;
      characters.push(...(data.characters || []));
      total = data.total || characters.length;
      fetchPage += 1;
    } while (characters.length < total);
    if (!isCurrentGmTargetContext(requestContext)) return;
    mailState.characters = characters;
    mailState.loaded = true;
    if (force) mailState.page = 1;
    const selected = characters.find(
      (character) => Number(character.charac_no) === mailState.selectedCharacterNo,
    );
    if (!selected) {
      mailState.selectedCharacterNo = null;
      mailSelectedRecipient.textContent = "未选择收件角色";
    }
    renderMailCharacterPicker();
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return;
    mailCharacterStatus.textContent = `角色加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) mailState.loading = false;
  }
}

function renderMailItems(data) {
  const items = data.items || [];
  mailState.itemPage = data.page || mailState.itemPage;
  mailState.itemLimit = data.limit || mailState.itemLimit;
  mailState.itemTotal = data.total || 0;
  const totalPages = Math.max(1, Math.ceil(mailState.itemTotal / mailState.itemLimit));
  mailItemStatus.textContent = mailState.itemTotal
    ? `共 ${mailState.itemTotal} 个物品`
    : "点击物品后自动填入物品 ID";
  mailItemPage.textContent = `第 ${mailState.itemPage} / ${totalPages} 页`;
  mailItemPrev.disabled = mailState.itemPage <= 1;
  mailItemNext.disabled = mailState.itemPage >= totalPages;
  mailItemResults.replaceChildren();
  if (!items.length) {
    const empty = document.createElement("div");
    empty.className = "mail-empty";
    empty.textContent = "没有找到物品";
    mailItemResults.appendChild(empty);
    return;
  }
  for (const item of items) {
    const button = document.createElement("button");
    button.className = "mail-item-row";
    button.type = "button";
    const name = document.createElement("strong");
    name.textContent = item.item_name || "-";
    const detail = document.createElement("span");
    const stackText =
      item.item_type === "stackable" && Number(item.stack_limit || 0) > 0
        ? ` · 上限 ${item.stack_limit}`
        : "";
    detail.textContent = `ID ${item.item_id} · ${item.item_type}${stackText}`;
    button.append(name, detail);
    button.addEventListener("click", () => {
      mailItemId.value = item.item_id;
      mailItemKeyword.value = item.item_name || item.item_id;
      mailAttachmentType.value = item.item_type === "equipment" ? "equipment" : "stackable";
      mailState.selectedItemStackLimit =
        item.item_type === "stackable" && Number(item.stack_limit || 0) > 0
          ? Number(item.stack_limit)
          : null;
      updateMailAttachmentFields();
      mailStatus.textContent = `已选择物品：${item.item_name || item.item_id}`;
      mailItemResults.querySelectorAll(".mail-item-row").forEach((row) => {
        row.classList.toggle("selected", row === button);
      });
    });
    mailItemResults.appendChild(button);
  }
}

async function searchMailItems(page = mailState.itemPage) {
  const keyword = mailItemKeyword.value.trim();
  if (!keyword) {
    mailStatus.textContent = "请输入物品名或 ID";
    return;
  }
  mailState.itemPage = Math.max(1, page);
  setCharacterButtonBusy(mailSearchItem, true, "搜索中");
  mailStatus.textContent = "搜索中";
  try {
    const data = await api(
      `/api/pvf/items?keyword=${encodeURIComponent(keyword)}&page=${mailState.itemPage}&limit=${mailState.itemLimit}`,
      { method: "GET" },
    );
    renderMailItems(data);
    mailStatus.textContent = `找到 ${data.total || 0} 个物品`;
  } catch (error) {
    mailStatus.textContent = `搜索失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(mailSearchItem, false);
  }
}

function mailPayload() {
  const isEquipment = mailAttachmentType.value === "equipment";
  return {
    message: mailMessage.value.trim(),
    item_id: Number(mailItemId.value.trim() || "0") || null,
    item_count: isEquipment ? 1 : Number(mailItemCount.value.trim() || "0"),
    gold: Number(mailGold.value.trim() || "0"),
    charac_no: mailState.selectedCharacterNo,
    item_type: mailAttachmentType.value,
    item_grade: isEquipment ? 1 : 0,
    enhancement_level: isEquipment ? Number(mailEnhancementLevel.value.trim() || "0") : 0,
    forge_level: isEquipment ? Number(mailForgeLevel.value.trim() || "0") : 0,
    amplify_option: isEquipment ? Number(mailAmplifyOption.value || "0") : 0,
    amplify_value: isEquipment ? Number(mailAmplifyValue.value.trim() || "0") : 0,
  };
}

function clampMailItemCount() {
  mailItemCount.value = mailItemCount.value.replace(/\D/g, "");
  if (mailAttachmentType.value === "equipment") return;
  const stackLimit = Number(mailState.selectedItemStackLimit || 0);
  if (stackLimit <= 0) {
    mailItemCount.removeAttribute("max");
    return;
  }
  mailItemCount.max = String(stackLimit);
  const currentCount = Number(mailItemCount.value || "0");
  if (currentCount > stackLimit) mailItemCount.value = String(stackLimit);
}

function updateMailAttachmentFields() {
  const isEquipment = mailAttachmentType.value === "equipment";
  mailItemCountLabel.textContent = isEquipment ? "锻造等级" : "物品数量";
  mailItemCount.hidden = isEquipment;
  mailForgeLevel.hidden = !isEquipment;
  mailEquipmentFields.hidden = !isEquipment;
  if (isEquipment) mailItemCount.removeAttribute("max");
  else clampMailItemCount();
}

function inventoryScopeName() {
  return inventoryScope.options[inventoryScope.selectedIndex]?.textContent || "物品栏";
}

function filteredInventoryCharacters() {
  const keyword = normalizeCharacterSearch(inventoryCharacterSearch.value);
  if (!keyword) return inventoryState.characters;
  return inventoryState.characters.filter((character) => {
    const name = normalizeCharacterSearch(character.charac_name);
    const number = String(character.charac_no || "");
    return name.includes(keyword) || number.includes(keyword);
  });
}

function selectInventoryCharacter(character) {
  inventoryState.selectedCharacterNo = Number(character.charac_no || 0) || null;
  inventoryState.selectedSlot = null;
  inventoryCharacterSearch.value = character.charac_name || String(character.charac_no || "");
  inventorySelectedCharacter.textContent = character.charac_name || String(character.charac_no || "");
  inventoryStatus.textContent = `已选择角色：${character.charac_name || character.charac_no}`;
  inventoryDelete.disabled = true;
  renderInventoryCharacterPicker();
}

function renderInventoryCharacterRows(characters) {
  inventoryCharacterList.replaceChildren();
  if (!characters.length) {
    const empty = document.createElement("div");
    empty.className = "inventory-empty";
    empty.textContent = "没有找到角色";
    inventoryCharacterList.appendChild(empty);
    return;
  }
  for (const character of characters) {
    const button = document.createElement("button");
    button.className = "inventory-character-row";
    button.type = "button";
    button.classList.toggle(
      "selected",
      Number(character.charac_no) === inventoryState.selectedCharacterNo,
    );
    const name = document.createElement("strong");
    name.textContent = character.charac_name || "-";
    const detail = document.createElement("span");
    detail.textContent = `Lv.${character.level || 0} · ${characterJobName(character)}`;
    button.append(name, detail);
    button.addEventListener("click", () => selectInventoryCharacter(character));
    inventoryCharacterList.appendChild(button);
  }
}

function renderInventoryCharacterPicker() {
  const characters = filteredInventoryCharacters();
  inventoryState.total = characters.length;
  const totalPages = Math.max(1, Math.ceil(inventoryState.total / inventoryState.limit));
  if (inventoryState.page > totalPages) inventoryState.page = totalPages;
  const start = (inventoryState.page - 1) * inventoryState.limit;
  const pageCharacters = characters.slice(start, start + inventoryState.limit);
  inventoryCharacterStatus.textContent = inventoryState.loaded
    ? `共 ${inventoryState.total} 个角色`
    : "等待加载";
  inventoryCharacterPage.textContent = `第 ${inventoryState.page} / ${totalPages} 页`;
  inventoryCharacterPrev.disabled = inventoryState.page <= 1;
  inventoryCharacterNext.disabled = inventoryState.page >= totalPages;
  renderInventoryCharacterRows(pageCharacters);
}

async function loadInventoryCharacters(force = false) {
  if (!currentUser || inventoryState.loading) return;
  if (!requireGmTarget(inventoryCharacterStatus)) return;
  if (inventoryState.loaded && !force) {
    renderInventoryCharacterPicker();
    return;
  }
  const requestContext = gmTargetRequestContext();
  inventoryState.loading = true;
  inventoryCharacterStatus.textContent = "角色加载中";
  try {
    const fetchLimit = 500;
    let fetchPage = 1;
    let total = 0;
    const characters = [];
    do {
      const data = await api(
        `/api/gm/characters?page=${fetchPage}&limit=${fetchLimit}${gmCharacterQuery()}`,
        withGmTargetSignal({ method: "GET" }, requestContext),
      );
      if (!isCurrentGmTargetContext(requestContext)) return;
      characters.push(...(data.characters || []));
      total = data.total || characters.length;
      fetchPage += 1;
    } while (characters.length < total);
    if (!isCurrentGmTargetContext(requestContext)) return;
    inventoryState.characters = characters;
    inventoryState.loaded = true;
    if (force) inventoryState.page = 1;
    const selected = characters.find(
      (character) => Number(character.charac_no) === inventoryState.selectedCharacterNo,
    );
    if (!selected) {
      inventoryState.selectedCharacterNo = null;
      inventorySelectedCharacter.textContent = "尚未选择角色";
    }
    renderInventoryCharacterPicker();
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return;
    inventoryCharacterStatus.textContent = `角色加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) inventoryState.loading = false;
  }
}

function inventoryItemDetail(item) {
  const parts = [];
  if (item.item_type_name === "装备") {
    if (item.is_seal) parts.push("封装");
    if (item.enhancement_level > 0) parts.push(`强化 +${item.enhancement_level}`);
    parts.push(`品级 ${formatNumber(item.count_or_grade)}`);
    parts.push(`耐久 ${item.durability}`);
    if (item.increase_type) {
      parts.push(`${item.increase_type_name} +${item.increase_value}`);
    }
    if (item.forge_level > 0) parts.push(`锻造 +${item.forge_level}`);
  } else {
    parts.push(`数量 ${formatNumber(item.count_or_grade)}`);
  }
  if (item.orb) parts.push(`宝珠 ${item.orb}`);
  return parts.join(" / ") || "-";
}

function renderInventory(data) {
  const character = data.character || {};
  const items = data.items || [];
  inventoryState.selectedSlot = null;
  inventoryState.items = items;
  if (character.charac_no) {
    inventoryState.selectedCharacterNo = Number(character.charac_no || 0) || null;
    inventoryCharacterSearch.value = character.charac_name || String(character.charac_no);
    inventorySelectedCharacter.textContent = character.charac_name || String(character.charac_no);
  }
  inventoryScopeSummary.textContent = data.scope_name || inventoryScopeName();
  inventoryItemCount.textContent = `${items.length} 个物品`;
  inventoryDelete.disabled = true;
  inventoryList.replaceChildren();
  if (!items.length) {
    const empty = document.createElement("div");
    empty.className = "inventory-empty";
    empty.textContent = "没有找到物品";
    inventoryList.appendChild(empty);
    return;
  }
  for (const item of items) {
    const button = document.createElement("button");
    button.className = "inventory-item-row";
    button.type = "button";
    const heading = document.createElement("span");
    heading.className = "inventory-item-heading";
    const name = document.createElement("strong");
    name.textContent = item.item_name || `未知物品 ${item.item_id}`;
    const slot = document.createElement("em");
    slot.textContent = `槽位 ${item.slot}`;
    heading.append(name, slot);
    const meta = document.createElement("span");
    meta.className = "inventory-item-meta";
    meta.textContent = `ID ${item.item_id} · ${item.item_type_name} · ${inventoryItemDetail(item)}`;
    button.append(heading, meta);
    button.addEventListener("click", () => {
      inventoryState.selectedSlot = item.slot;
      inventoryList.querySelectorAll(".inventory-item-row").forEach((row) => {
        row.classList.toggle("selected", row === button);
      });
      inventoryDelete.disabled = false;
      inventoryStatus.textContent = `已选择槽位 ${item.slot}`;
    });
    inventoryList.appendChild(button);
  }
}

async function queryInventory() {
  const characNo = inventoryState.selectedCharacterNo;
  if (!characNo) {
    inventoryStatus.textContent = "请先从角色列表选择角色";
    return null;
  }
  setCharacterButtonBusy(inventoryQuery, true, "读取中");
  inventoryStatus.textContent = "背包读取中";
  try {
    const data = await api("/api/gm/inventory/query", {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo, scope: inventoryScope.value }),
    });
    renderInventory(data);
    inventoryStatus.textContent = `读取完成：${data.item_count || 0} 个物品`;
    return data;
  } catch (error) {
    inventoryStatus.textContent = `读取失败：${errorMessage(error, "未知错误")}`;
    return null;
  } finally {
    setCharacterButtonBusy(inventoryQuery, false);
  }
}

function renderAvatarOptions(options) {
  avatarHiddenOption.replaceChildren();
  for (const optionInfo of options || []) {
    const optionName = optionInfo.name === "None" ? "无" : optionInfo.name;
    addSelectOption(
      avatarHiddenOption,
      optionInfo.hidden_option,
      `${optionInfo.hidden_option} - ${optionName || optionInfo.hidden_option}`,
    );
  }
  if (!avatarHiddenOption.options.length) addSelectOption(avatarHiddenOption, 0, "0 - 无");
}

async function loadAvatarOptions() {
  if (avatarOptionsLoaded) return true;
  if (avatarOptionsLoading) return avatarOptionsLoading;
  avatarOptionsLoading = (async () => {
    try {
      const data = await api("/api/gm/avatar/options", { method: "GET" });
      renderAvatarOptions(data.options || []);
      avatarOptionsLoaded = true;
      return true;
    } catch (error) {
      avatarStatus.textContent = `潜能选项加载失败：${errorMessage(error, "未知错误")}`;
      return false;
    } finally {
      avatarOptionsLoading = null;
    }
  })();
  return avatarOptionsLoading;
}

function filteredAvatarCharacters() {
  const keyword = normalizeCharacterSearch(avatarCharacterSearch.value);
  if (!keyword) return avatarState.characters;
  return avatarState.characters.filter((character) => {
    const name = normalizeCharacterSearch(character.charac_name);
    const number = String(character.charac_no || "");
    return name.includes(keyword) || number.includes(keyword);
  });
}

function selectAvatarCharacter(character) {
  avatarState.selectedCharacterNo = Number(character.charac_no || 0) || null;
  avatarState.selectedUiIds = [];
  avatarCharacterSearch.value = character.charac_name || String(character.charac_no || "");
  avatarSelectedCharacter.textContent = character.charac_name || String(character.charac_no || "");
  avatarSelectionSummary.textContent = "未选择时装";
  avatarApply.disabled = true;
  avatarStatus.textContent = `已选择角色：${character.charac_name || character.charac_no}`;
  renderAvatarCharacterPicker();
}

function renderAvatarCharacterRows(characters) {
  avatarCharacterList.replaceChildren();
  if (!characters.length) {
    const empty = document.createElement("div");
    empty.className = "inventory-empty";
    empty.textContent = "没有找到角色";
    avatarCharacterList.appendChild(empty);
    return;
  }
  for (const character of characters) {
    const button = document.createElement("button");
    button.className = "inventory-character-row";
    button.type = "button";
    button.classList.toggle(
      "selected",
      Number(character.charac_no) === avatarState.selectedCharacterNo,
    );
    const name = document.createElement("strong");
    name.textContent = character.charac_name || "-";
    const detail = document.createElement("span");
    detail.textContent = `Lv.${character.level || 0} · ${characterJobName(character)}`;
    button.append(name, detail);
    button.addEventListener("click", () => selectAvatarCharacter(character));
    avatarCharacterList.appendChild(button);
  }
}

function renderAvatarCharacterPicker() {
  const characters = filteredAvatarCharacters();
  avatarState.total = characters.length;
  const totalPages = Math.max(1, Math.ceil(avatarState.total / avatarState.limit));
  if (avatarState.page > totalPages) avatarState.page = totalPages;
  const start = (avatarState.page - 1) * avatarState.limit;
  const pageCharacters = characters.slice(start, start + avatarState.limit);
  avatarCharacterStatus.textContent = avatarState.loaded
    ? `共 ${avatarState.total} 个角色`
    : "等待加载";
  avatarCharacterPage.textContent = `第 ${avatarState.page} / ${totalPages} 页`;
  avatarCharacterPrev.disabled = avatarState.page <= 1;
  avatarCharacterNext.disabled = avatarState.page >= totalPages;
  renderAvatarCharacterRows(pageCharacters);
}

async function loadAvatarCharacters(force = false) {
  if (!currentUser || avatarState.loading) return;
  if (!requireGmTarget(avatarCharacterStatus)) return;
  if (avatarState.loaded && !force) {
    renderAvatarCharacterPicker();
    return;
  }
  const requestContext = gmTargetRequestContext();
  avatarState.loading = true;
  avatarCharacterStatus.textContent = "角色加载中";
  try {
    const fetchLimit = 500;
    let fetchPage = 1;
    let total = 0;
    const characters = [];
    do {
      const data = await api(
        `/api/gm/characters?page=${fetchPage}&limit=${fetchLimit}${gmCharacterQuery()}`,
        withGmTargetSignal({ method: "GET" }, requestContext),
      );
      if (!isCurrentGmTargetContext(requestContext)) return;
      characters.push(...(data.characters || []));
      total = data.total || characters.length;
      fetchPage += 1;
    } while (characters.length < total);
    if (!isCurrentGmTargetContext(requestContext)) return;
    avatarState.characters = characters;
    avatarState.loaded = true;
    if (force) avatarState.page = 1;
    const selected = characters.find(
      (character) => Number(character.charac_no) === avatarState.selectedCharacterNo,
    );
    if (!selected) {
      avatarState.selectedCharacterNo = null;
      avatarSelectedCharacter.textContent = "尚未选择角色";
    }
    renderAvatarCharacterPicker();
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return;
    avatarCharacterStatus.textContent = `角色加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) avatarState.loading = false;
  }
}

function renderAvatar(data) {
  const character = data.character || {};
  const items = data.items || [];
  avatarState.selectedUiIds = [];
  avatarState.items = items;
  if (character.charac_no) {
    avatarState.selectedCharacterNo = Number(character.charac_no || 0) || null;
    avatarCharacterSearch.value = character.charac_name || String(character.charac_no);
    avatarSelectedCharacter.textContent = character.charac_name || String(character.charac_no);
  }
  avatarSelectionSummary.textContent = "未选择时装";
  avatarItemCount.textContent = `${items.length} 件时装`;
  avatarApply.disabled = true;
  avatarList.replaceChildren();
  if (!items.length) {
    const empty = document.createElement("div");
    empty.className = "inventory-empty";
    empty.textContent = "没有找到时装";
    avatarList.appendChild(empty);
    return;
  }
  for (const item of items) {
    const button = document.createElement("button");
    button.className = "avatar-item-row";
    button.type = "button";
    const heading = document.createElement("span");
    heading.className = "avatar-item-heading";
    const name = document.createElement("strong");
    name.textContent = item.item_name || `时装 ${item.item_id}`;
    const uiId = document.createElement("em");
    uiId.textContent = `UI ${item.ui_id}`;
    heading.append(name, uiId);
    const hiddenName = item.hidden_name === "None" ? "无" : item.hidden_name;
    const meta = document.createElement("span");
    meta.className = "avatar-item-meta";
    meta.textContent = `ID ${item.item_id} · 当前潜能 ${item.hidden_option} - ${hiddenName || item.hidden_option}`;
    button.append(heading, meta);
    button.addEventListener("click", () => {
      const id = Number(item.ui_id);
      if (avatarState.selectedUiIds.includes(id)) {
        avatarState.selectedUiIds = avatarState.selectedUiIds.filter((value) => value !== id);
      } else {
        avatarState.selectedUiIds.push(id);
      }
      button.classList.toggle("selected", avatarState.selectedUiIds.includes(id));
      avatarSelectionSummary.textContent = avatarState.selectedUiIds.length
        ? `已选择 ${avatarState.selectedUiIds.length} 件`
        : "未选择时装";
      avatarStatus.textContent = `已选择 ${avatarState.selectedUiIds.length} 件时装`;
      avatarApply.disabled = avatarState.selectedUiIds.length === 0;
    });
    avatarList.appendChild(button);
  }
}

async function queryAvatar() {
  const characNo = avatarState.selectedCharacterNo;
  if (!characNo) {
    avatarStatus.textContent = "请先从角色列表选择角色";
    return null;
  }
  setCharacterButtonBusy(avatarQuery, true, "读取中");
  avatarStatus.textContent = "时装读取中";
  try {
    const data = await api("/api/gm/avatar/query", {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo }),
    });
    renderAvatar(data);
    avatarStatus.textContent = `读取完成：${data.item_count || 0} 件时装`;
    return data;
  } catch (error) {
    avatarStatus.textContent = `读取失败：${errorMessage(error, "未知错误")}`;
    return null;
  } finally {
    setCharacterButtonBusy(avatarQuery, false);
  }
}

function rechargeTargetPayload() {
  const account = currentOperationAccount();
  const uid = Number(account?.uid || 0);
  return {
    uid: uid > 0 ? uid : null,
    account_name: String(account?.account_name || "").trim(),
  };
}

function renderRechargeBalance(data) {
  rechargeAccount.textContent = data.account_name || currentOperationAccount()?.account_name || "-";
  rechargeCera.textContent = formatNumber(data.cera);
  rechargeCeraPoint.textContent = formatNumber(data.cera_point);
}

async function queryRechargeBalance() {
  if (!currentUser || rechargeState.loading) return null;
  if (!requireGmTarget(rechargeStatus)) return null;
  const requestContext = gmTargetRequestContext();
  rechargeState.loading = true;
  setCharacterButtonBusy(rechargeRefresh, true, "查询中");
  rechargeStatus.textContent = "正在查询当前账号余额";
  try {
    const data = await api("/api/gm/cera/query", {
      method: "POST",
      body: JSON.stringify(rechargeTargetPayload()),
      signal: requestContext.signal,
    });
    if (!isCurrentGmTargetContext(requestContext)) return null;
    renderRechargeBalance(data);
    rechargeStatus.textContent = "余额已更新";
    return data;
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return null;
    rechargeStatus.textContent = `查询失败：${errorMessage(error, "未知错误")}`;
    return null;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) {
      rechargeState.loading = false;
      setCharacterButtonBusy(rechargeRefresh, false);
    }
  }
}

function renderRapidFire(snapshot) {
  rapidFireState.configs = Array.isArray(snapshot?.configs) ? snapshot.configs : [];
  rapidFireState.loaded = true;
  rapidFireNativeStatus.textContent = snapshot?.error
    ? "监听不可用"
    : snapshot?.ready
      ? "监听已就绪"
      : isTauri
        ? "监听器启动中"
        : "仅 EXE 中生效";
  rapidFireInstallDriver.hidden = !(snapshot?.error && snapshot?.driverInstallable);
  rapidFireCount.textContent = `${rapidFireState.configs.length} 个按键`;
  rapidFireList.replaceChildren();
  if (!rapidFireState.configs.length) {
    const empty = document.createElement("div");
    empty.className = "rapid-fire-empty";
    empty.textContent = "尚未添加连发按键";
    rapidFireList.appendChild(empty);
  } else {
    for (const config of rapidFireState.configs) {
      const row = document.createElement("article");
      row.className = "rapid-fire-row";
      const key = document.createElement("strong");
      key.className = "rapid-fire-key-badge";
      key.textContent = config.key;
      const detail = document.createElement("div");
      const title = document.createElement("strong");
      title.textContent = `每 ${formatNumber(config.intervalMs)} 毫秒触发一次`;
      const hint = document.createElement("span");
      hint.textContent = `按住 ${config.key} 时连续触发`;
      detail.append(title, hint);
      const remove = document.createElement("button");
      remove.type = "button";
      remove.textContent = "删除";
      remove.addEventListener("click", async () => {
        if (rapidFireState.loading) return;
        rapidFireState.loading = true;
        setCharacterButtonBusy(remove, true, "删除中");
        rapidFireStatus.textContent = `正在删除 ${config.key} 键`;
        try {
          let nextSnapshot;
          if (isTauri) {
            nextSnapshot = await invoke("remove_rapid_fire", { key: config.key });
          } else {
            nextSnapshot = {
              configs: rapidFireState.configs.filter((item) => item.key !== config.key),
              ready: false,
              error: null,
            };
          }
          renderRapidFire(nextSnapshot);
          rapidFireStatus.textContent = `${config.key} 键连发已删除`;
        } catch (error) {
          rapidFireStatus.textContent = `删除失败：${errorMessage(error, "未知错误")}`;
        } finally {
          rapidFireState.loading = false;
          setCharacterButtonBusy(remove, false);
        }
      });
      row.append(key, detail, remove);
      rapidFireList.appendChild(row);
    }
  }
  if (snapshot?.error) {
    rapidFireStatus.textContent = snapshot.driverInstallable
      ? `${snapshot.error}。可点击“安装驱动”，完成后重启电脑。`
      : snapshot.error;
  }
}

async function loadRapidFireConfigs(force = false) {
  if (rapidFireState.loading || (rapidFireState.loaded && !force)) return;
  rapidFireState.loading = true;
  rapidFireStatus.textContent = "正在读取连发配置";
  try {
    const snapshot = isTauri
      ? await invoke("list_rapid_fire")
      : { configs: rapidFireState.configs, ready: false, error: null };
    renderRapidFire(snapshot);
    if (!snapshot.error) rapidFireStatus.textContent = "仅游戏内生效";
  } catch (error) {
    rapidFireNativeStatus.textContent = "读取失败";
    rapidFireStatus.textContent = `读取失败：${errorMessage(error, "未知错误")}`;
  } finally {
    rapidFireState.loading = false;
  }
}

function openGmConfirm(title, message, submitText, action) {
  gmConfirmTitle.textContent = title;
  gmConfirmText.textContent = message;
  gmConfirmSubmit.textContent = submitText;
  gmConfirmAction = action;
  gmConfirmOverlay.hidden = false;
}

function closeGmConfirm() {
  gmConfirmOverlay.hidden = true;
  gmConfirmAction = null;
  gmConfirmSubmit.disabled = false;
}

function renderEvents() {
  eventSelect.replaceChildren();
  for (const item of eventState.available) {
    addSelectOption(
      eventSelect,
      item.event_id,
      `${item.event_id} - ${item.event_explain || item.event_name || "活动"}`,
    );
  }
  eventList.replaceChildren();
  eventCount.textContent = `${eventState.running.length} 个运行中`;
  eventDelete.disabled = !eventState.selectedLogId;
  if (!eventState.running.length) {
    const empty = document.createElement("div");
    empty.className = "gm-admin-empty";
    empty.textContent = "当前没有运行中的活动";
    eventList.appendChild(empty);
    return;
  }
  for (const item of eventState.running) {
    const row = document.createElement("button");
    row.type = "button";
    row.className = "gm-event-row";
    row.classList.toggle("selected", Number(item.log_id) === eventState.selectedLogId);
    const id = document.createElement("strong");
    id.textContent = `#${item.log_id}`;
    const name = document.createElement("strong");
    name.textContent = item.event_explain || item.event_name || `活动 ${item.event_id}`;
    const parameter1 = document.createElement("span");
    parameter1.textContent = `参数1 ${item.parameter1}`;
    const parameter2 = document.createElement("span");
    parameter2.textContent = `参数2 ${item.parameter2}`;
    row.append(id, name, parameter1, parameter2);
    row.addEventListener("click", () => {
      eventState.selectedLogId = Number(item.log_id);
      eventStatus.textContent = `已选择活动日志 #${item.log_id}`;
      renderEvents();
    });
    eventList.appendChild(row);
  }
}

async function loadEvents(force = false) {
  if (!currentUser?.permissions?.includes("gm.event.manage")
    || eventState.loading
    || (eventState.loaded && !force)) return;
  eventState.loading = true;
  setCharacterButtonBusy(eventRefresh, true, "刷新中");
  eventStatus.textContent = "活动加载中";
  try {
    const data = await api("/api/gm/events", { method: "GET" });
    eventState.available = data.available || [];
    eventState.running = data.running || [];
    eventState.selectedLogId = null;
    eventState.loaded = true;
    renderEvents();
    eventStatus.textContent = `可用 ${eventState.available.length} 个活动`;
  } catch (error) {
    eventStatus.textContent = `活动加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    eventState.loading = false;
    setCharacterButtonBusy(eventRefresh, false);
  }
}

function renderBanStatus(data) {
  banTarget.textContent = `${data.account_name} · UID ${data.uid}`;
  banState.textContent = data.banned ? "已限制" : "状态正常";
  banStateView.textContent = data.banned ? "已限制" : "正常";
  banTypeView.textContent = data.banned ? data.punish_type_name || "-" : "-";
  banStartView.textContent = data.banned ? data.start_time || "-" : "-";
  banEndView.textContent = data.banned ? data.end_time || "-" : "-";
  banReasonView.textContent = data.banned ? data.reason || "未填写" : "-";
}

async function queryBanStatus() {
  if (!requireGmTarget(banStatus)) return null;
  const requestContext = gmTargetRequestContext();
  setCharacterButtonBusy(banQuery, true, "查询中");
  banStatus.textContent = "正在查询目标账号状态";
  try {
    const data = await api("/api/gm/ban/query", {
      method: "POST",
      body: JSON.stringify(rechargeTargetPayload()),
      signal: requestContext.signal,
    });
    if (!isCurrentGmTargetContext(requestContext)) return null;
    renderBanStatus(data);
    banStatus.textContent = "账号状态已更新";
    return data;
  } catch (error) {
    if (isCanceledRequest(error) || !isCurrentGmTargetContext(requestContext)) return null;
    banStatus.textContent = `查询失败：${errorMessage(error, "未知错误")}`;
    return null;
  } finally {
    if (isCurrentGmTargetContext(requestContext)) setCharacterButtonBusy(banQuery, false);
  }
}

function renderPermissionAccounts() {
  permissionAccountList.replaceChildren();
  permissionAccountCount.textContent = `${permissionState.accounts.length} 个结果`;
  if (!permissionState.accounts.length) {
    const empty = document.createElement("div");
    empty.className = "gm-admin-empty";
    empty.textContent = "没有找到账号";
    permissionAccountList.appendChild(empty);
    return;
  }
  for (const account of permissionState.accounts) {
    const row = document.createElement("button");
    row.type = "button";
    row.className = "gm-account-row";
    row.classList.toggle("selected", Number(account.uid) === Number(permissionState.selected?.uid));
    const name = document.createElement("strong");
    name.textContent = account.account_name;
    const uid = document.createElement("span");
    uid.textContent = `UID ${account.uid}`;
    row.append(name, uid);
    row.addEventListener("click", () => {
      permissionState.selected = account;
      permissionSelectedAccount.textContent = `${account.account_name} · UID ${account.uid}`;
      permissionSave.disabled = false;
      permissionStatus.textContent = "修改后点击保存权限";
      renderPermissionAccounts();
      renderPermissionGrid(account.permissions || []);
    });
    permissionAccountList.appendChild(row);
  }
}

function renderPermissionGrid(selectedPermissions) {
  const selected = new Set(selectedPermissions || []);
  permissionGrid.replaceChildren();
  for (const permission of permissionState.all) {
    const label = document.createElement("label");
    label.className = "gm-permission-item";
    const input = document.createElement("input");
    input.type = "checkbox";
    input.value = permission;
    input.checked = selected.has(permission);
    input.disabled = !permissionState.selected;
    const detail = document.createElement("span");
    const name = document.createElement("strong");
    name.textContent = permissionNames[permission] || permission;
    const code = document.createElement("code");
    code.textContent = permission;
    detail.append(name, code);
    label.append(input, detail);
    permissionGrid.appendChild(label);
  }
}

async function loadPermissionAccounts(force = false) {
  if (!isAdmin() || permissionState.loading || (permissionState.loaded && !force)) return;
  permissionState.loading = true;
  permissionStatus.textContent = "正在加载账号权限";
  try {
    if (!permissionState.all.length) {
      const permissionData = await api("/api/admin/permissions", { method: "GET" });
      permissionState.all = permissionData.permissions || [];
    }
    const keyword = permissionKeyword.value.trim();
    const data = await api(`/api/admin/accounts?keyword=${encodeURIComponent(keyword)}&limit=50`, {
      method: "GET",
    });
    permissionState.accounts = data.accounts || [];
    permissionState.selected = null;
    permissionState.loaded = true;
    permissionSelectedAccount.textContent = "未选择账号";
    permissionSave.disabled = true;
    renderPermissionAccounts();
    renderPermissionGrid([]);
    permissionStatus.textContent = `已加载 ${permissionState.accounts.length} 个账号`;
  } catch (error) {
    permissionStatus.textContent = `权限加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    permissionState.loading = false;
  }
}

function announcementValue(value = {}) {
  return {
    title: String(value.title || ""),
    summary: String(value.summary || ""),
    content: String(value.content || ""),
    poster_url: String(value.poster_url || ""),
  };
}

function renderAnnouncementEditor() {
  announcementEditorList.replaceChildren();
  systemState.announcements.forEach((announcement, index) => {
    const row = document.createElement("div");
    row.className = "gm-announcement-row";
    const title = document.createElement("input");
    title.maxLength = 80;
    title.placeholder = `公告 ${index + 1} 标题`;
    title.value = announcement.title;
    const summary = document.createElement("input");
    summary.maxLength = 160;
    summary.placeholder = "摘要";
    summary.value = announcement.summary;
    const posterUrl = document.createElement("input");
    posterUrl.maxLength = 512;
    posterUrl.placeholder = "海报地址，例如 /api/posters/event 或 /api/posters/event.png";
    posterUrl.value = announcement.poster_url;
    const content = document.createElement("textarea");
    content.maxLength = 2000;
    content.rows = 1;
    content.placeholder = "公告全文";
    content.value = announcement.content;
    content.className = "gm-announcement-content";
    const preview = document.createElement("img");
    preview.className = "gm-poster-preview";
    preview.alt = `公告 ${index + 1} 海报预览`;
    const updatePreview = () => {
      const url = resolvePosterUrl(posterUrl.value.trim());
      preview.hidden = !url;
      if (url) preview.src = url;
    };
    const remove = document.createElement("button");
    remove.type = "button";
    remove.textContent = "×";
    remove.setAttribute("aria-label", `删除公告 ${index + 1}`);
    title.addEventListener("input", () => { announcement.title = title.value; });
    summary.addEventListener("input", () => { announcement.summary = summary.value; });
    posterUrl.addEventListener("input", () => {
      announcement.poster_url = posterUrl.value;
      updatePreview();
    });
    content.addEventListener("input", () => { announcement.content = content.value; });
    remove.addEventListener("click", () => {
      systemState.announcements.splice(index, 1);
      renderAnnouncementEditor();
    });
    preview.addEventListener("error", () => { preview.hidden = true; });
    row.append(title, summary, posterUrl, remove, preview, content);
    updatePreview();
    announcementEditorList.appendChild(row);
  });
  announcementAdd.disabled = systemState.announcements.length >= 8;
}

function renderPvfStatus(data) {
  pvfClientMd5.value = data?.client_pvf_md5 || "";
  pvfMd5Status.textContent = data?.client_pvf_md5
    ? "启动器将校验客户端 Script.pvf"
    : "未启用客户端 PVF 校验";
  if (!data?.loaded) {
    pvfStatus.textContent = "尚未加载 PVF 缓存";
    return;
  }
  pvfPath.value = data.pvf_path || data.path || "";
  pvfEncode.value = data.encode || "big5";
  pvfStatus.textContent = `已加载 ${formatNumber(data.stackable_count)} 个道具、${formatNumber(data.equipment_count)} 件装备 · MD5 ${data.md5 || "-"} · ${data.updated_at || "刚刚"}`;
}

async function loadSystemData(force = false) {
  if (!isAdmin() || systemState.loading || (systemState.loaded && !force)) return;
  systemState.loading = true;
  systemHomeStatus.textContent = "正在读取系统设置";
  try {
    const [settingsData, pvfData] = await Promise.all([
      api("/api/settings", { method: "GET" }),
      api("/api/pvf/status", { method: "GET" }),
    ]);
    const home = normalizeHome(settingsData.home);
    systemHomeTitle.value = home.home_title;
    systemHomeEyebrow.value = home.home_eyebrow;
    systemClientDownloadUrl.value = home.client_download_url || "";
    systemState.announcements = home.announcements.slice(0, 8).map(announcementValue);
    renderAnnouncementEditor();
    renderPvfStatus(pvfData);
    systemState.loaded = true;
    systemHomeStatus.textContent = `已读取 ${systemState.announcements.length} 条公告`;
  } catch (error) {
    systemHomeStatus.textContent = `系统设置加载失败：${errorMessage(error, "未知错误")}`;
  } finally {
    systemState.loading = false;
  }
}

function setSystemTab(tabName, load = true) {
  systemState.tab = tabName;
  document.querySelectorAll("[data-system-tab]").forEach((button) => {
    button.classList.toggle("active", button.dataset.systemTab === tabName);
  });
  document.querySelectorAll("[data-system-panel]").forEach((panel) => {
    const active = panel.dataset.systemPanel === tabName;
    panel.hidden = !active;
    panel.classList.toggle("active", active);
  });
  if (load && tabName === "logs") loadOperationLogs();
}

function renderOperationLogs(logs) {
  logList.replaceChildren();
  if (!logs.length) {
    const empty = document.createElement("div");
    empty.className = "gm-admin-empty";
    empty.textContent = "没有找到操作日志";
    logList.appendChild(empty);
  } else {
    for (const item of logs) {
      const row = document.createElement("article");
      row.className = "gm-log-row";
      for (const value of [
        `#${item.id}`,
        item.action,
        `UID ${item.uid}`,
        item.created_at || "-",
        item.detail || "-",
      ]) {
        const element = document.createElement(value === item.action ? "strong" : "span");
        element.textContent = value;
        if (value === item.detail && item.detail) element.title = item.detail;
        row.appendChild(element);
      }
      logList.appendChild(row);
    }
  }
  const totalPages = Math.max(1, Math.ceil(systemState.logTotal / systemState.logLimit));
  logPage.textContent = `第 ${systemState.logPage} / ${totalPages} 页`;
  logPrev.disabled = systemState.logPage <= 1;
  logNext.disabled = systemState.logPage >= totalPages;
}

async function loadOperationLogs(page = systemState.logPage) {
  if (!isAdmin()) return;
  systemState.logPage = Math.max(1, page);
  logList.replaceChildren();
  const loading = document.createElement("div");
  loading.className = "gm-admin-empty";
  loading.textContent = "操作日志加载中";
  logList.appendChild(loading);
  try {
    const data = await api(`/api/admin/logs?keyword=${encodeURIComponent(logKeyword.value.trim())}&page=${systemState.logPage}&limit=${systemState.logLimit}`, {
      method: "GET",
    });
    systemState.logTotal = data.total || 0;
    renderOperationLogs(data.logs || []);
  } catch (error) {
    loading.textContent = `日志加载失败：${errorMessage(error, "未知错误")}`;
  }
}

function enterSession(user) {
  currentUser = { ...user };
  accessToken = currentUser.access_token || accessToken;
  const accountName = currentUser.account_name || "-";
  const adminMode = isAdmin();

  resetGmTarget();
  resetAdminStates();
  body.classList.toggle("gm-mode", adminMode);
  accountSummaryPrefix.textContent = adminMode ? "GM：" : "账号：";
  mailGlobalActions.hidden = !adminMode;

  currentAccount.textContent = accountName;
  accountDetailUid.value = currentUser.uid ?? "-";
  launchButton.disabled = !currentUser.can_launch;
  setGameButtonMode(false);
  clientState.classList.toggle("locked", !currentUser.can_launch);
  clientState.innerHTML = adminMode ? "<i></i> GM 管理模式" : "<i></i> 客户端就绪";
  updateNavigation(currentUser.permissions || []);

  body.classList.remove("logged-out");
  body.classList.add("logged-in");
  loginView.hidden = true;
  homeView.hidden = false;
  accountSummary.hidden = false;
  activateSection("大厅");
  showToast("登录成功");
  if (isTauri && currentUser.can_launch) {
    gameMonitorTimer = window.setTimeout(monitorGameProcess, 0);
  }
}

function leaveSession() {
  window.clearTimeout(gameMonitorTimer);
  accessToken = "";
  currentUser = null;
  activeSection = "大厅";
  passwordOverlay.hidden = true;
  announcementOverlay.hidden = true;
  gmTargetBar.hidden = true;
  mailGlobalActions.hidden = true;
  accountSummary.hidden = true;
  accountSummary.classList.remove("active");
  homeView.hidden = true;
  loginView.hidden = false;
  body.classList.remove("logged-in");
  body.classList.remove("gm-mode");
  body.classList.remove("gm-target-section");
  body.classList.add("logged-out");
  launchButton.disabled = true;
  setGameButtonMode(false);
  clientState.classList.remove("locked");
  clientState.innerHTML = "<i></i> 等待登录";
  resetGmTarget();
  resetTargetFeatureStates();
  resetRapidFireState();
  resetAdminStates();
  authStatus.textContent = "";
  if (!rememberPassword.checked) loginPassword.value = "";
}

function activateSection(section) {
  activeSection = section;
  document.querySelectorAll(".nav-item").forEach((item) => {
    item.classList.toggle("active", item.dataset.section === section);
  });
  accountSummary.classList.toggle("active", section === "账号");
  const showCharacterWorkspace = section === "角色";
  const showMailWorkspace = section === "邮件";
  const showInventoryWorkspace = section === "背包";
  const showAvatarWorkspace = section === "时装潜能";
  const showRechargeWorkspace = section === "充值";
  const showRapidFireWorkspace = section === "按键连发";
  const showEventWorkspace = section === "活动";
  const showBanWorkspace = section === "封禁";
  const showPermissionWorkspace = section === "权限";
  const showSystemWorkspace = section === "系统";
  const showFeatureWorkspace = showCharacterWorkspace
    || showMailWorkspace
    || showInventoryWorkspace
    || showAvatarWorkspace
    || showRechargeWorkspace
    || showRapidFireWorkspace
    || showEventWorkspace
    || showBanWorkspace
    || showPermissionWorkspace
    || showSystemWorkspace;
  characterWorkspace.hidden = !showCharacterWorkspace;
  mailWorkspace.hidden = !showMailWorkspace;
  inventoryWorkspace.hidden = !showInventoryWorkspace;
  avatarWorkspace.hidden = !showAvatarWorkspace;
  rechargeWorkspace.hidden = !showRechargeWorkspace;
  rapidFireWorkspace.hidden = !showRapidFireWorkspace;
  eventWorkspace.hidden = !showEventWorkspace;
  banWorkspace.hidden = !showBanWorkspace;
  permissionWorkspace.hidden = !showPermissionWorkspace;
  systemWorkspace.hidden = !showSystemWorkspace;
  const showGmTarget = isAdmin() && gmTargetSections.has(section);
  gmTargetBar.hidden = !showGmTarget;
  body.classList.toggle("gm-target-section", showGmTarget);
  announcementPanel.hidden = showFeatureWorkspace;
  launchArea.hidden = showFeatureWorkspace || !currentUser?.can_launch;
  if (showCharacterWorkspace) {
    loadCharacterPicker();
    loadCharacterJobOptions();
  } else if (showMailWorkspace) {
    loadMailCharacters();
  } else if (showInventoryWorkspace) {
    loadInventoryCharacters();
  } else if (showAvatarWorkspace) {
    loadAvatarOptions();
    loadAvatarCharacters();
  } else if (showRechargeWorkspace) {
    rechargeAccount.textContent = currentOperationAccount()?.account_name || "-";
    queryRechargeBalance();
  } else if (showRapidFireWorkspace) {
    loadRapidFireConfigs();
  } else if (showEventWorkspace) {
    loadEvents();
  } else if (showBanWorkspace) {
    banTarget.textContent = gmState.target
      ? `${gmState.target.account_name} · UID ${gmState.target.uid}`
      : "未选择目标账号";
    if (gmState.target) queryBanStatus();
    else banStatus.textContent = "请先在页面顶部选择目标账号";
  } else if (showPermissionWorkspace) {
    loadPermissionAccounts();
  } else if (showSystemWorkspace) {
    loadSystemData();
    setSystemTab(systemState.tab);
  } else if (section !== "大厅" && section !== "账号") {
    showToast(`${section}功能将在后续阶段迁移`);
  }
}

async function clearSavedCredential() {
  rememberedAccount = "";
  rememberPassword.checked = false;
  if (!isTauri) return;
  try {
    await invoke("clear_saved_login");
  } catch (error) {
    console.error("清除已保存密码失败", error);
  }
}

async function loadSavedCredential() {
  if (!isTauri) return;
  try {
    const saved = await invoke("load_saved_login");
    if (!saved) return;
    loginAccount.value = saved.account;
    loginPassword.value = saved.password;
    rememberPassword.checked = true;
    rememberedAccount = saved.account;
  } catch (error) {
    console.error("读取已保存密码失败", error);
  }
}

async function persistLoginChoice() {
  if (!rememberPassword.checked) {
    await clearSavedCredential();
    return;
  }
  if (!isTauri) return;
  await invoke("save_saved_login", {
    account: loginAccount.value.trim(),
    password: loginPassword.value,
  });
  rememberedAccount = loginAccount.value.trim();
}

function normalizeHome(home = {}) {
  const announcements = Array.isArray(home.announcements) && home.announcements.length
    ? home.announcements
    : defaultHome.announcements;
  return {
    home_title: home.home_title || defaultHome.home_title,
    home_eyebrow: home.home_eyebrow || defaultHome.home_eyebrow,
    client_download_url: home.client_download_url || "",
    announcements,
  };
}

const defaultPosterUrls = [
  "/api/posters/sample-1",
  "/api/posters/sample-2",
  "/api/posters/sample-3",
];

function resolvePosterUrl(value) {
  const url = String(value || "").trim();
  if (!url) return "";
  if (/^(https?:|data:)/i.test(url)) return url;
  if (url.startsWith("/api/")) return `${apiBase}${url}`;
  return url;
}

function renderPosterCarousel(announcements) {
  window.clearInterval(posterTimer);
  posterIndex = 0;
  const configured = announcements
    .filter((announcement) => announcement.poster_url)
    .slice(0, 5)
    .map((announcement) => ({ announcement, url: resolvePosterUrl(announcement.poster_url) }));
  const entries = configured.length
    ? configured
    : defaultPosterUrls.map((url, index) => ({ announcement: announcements[index] || null, url: resolvePosterUrl(url) }));
  posterCarousel.replaceChildren();
  const dotsContainer = document.createElement("div");
  dotsContainer.className = "poster-dots";
  dotsContainer.setAttribute("aria-hidden", "true");
  const posters = [];
  const dots = [];
  entries.forEach((entry, index) => {
    const poster = document.createElement("img");
    poster.classList.toggle("active", index === 0);
    poster.alt = entry.announcement?.title || `活动海报 ${index + 1}`;
    poster.src = entry.url;
    poster.addEventListener("error", () => {
      poster.removeAttribute("src");
      poster.alt = "海报加载失败";
    }, { once: true });
    if (entry.announcement) {
      poster.tabIndex = 0;
      poster.setAttribute("role", "button");
      poster.addEventListener("click", () => openAnnouncement(entry.announcement));
      poster.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") openAnnouncement(entry.announcement);
      });
    }
    const dot = document.createElement("span");
    dot.classList.toggle("active", index === 0);
    posters.push(poster);
    dots.push(dot);
    posterCarousel.appendChild(poster);
    dotsContainer.appendChild(dot);
  });
  posterCarousel.appendChild(dotsContainer);
  if (posters.length > 1) {
    posterTimer = window.setInterval(() => {
      posterIndex = (posterIndex + 1) % posters.length;
      posters.forEach((poster, index) => poster.classList.toggle("active", index === posterIndex));
      dots.forEach((dot, index) => dot.classList.toggle("active", index === posterIndex));
    }, 5200);
  }
}

function openAnnouncement(announcement) {
  announcementTitle.textContent = announcement.title || "公告";
  announcementContent.textContent =
    announcement.content || announcement.summary || "暂无公告全文";
  announcementOverlay.hidden = false;
}

function renderHome(home) {
  homeSettings = normalizeHome(home);
  clientDownloadUrl = homeSettings.client_download_url || "";
  homeTitle.textContent = homeSettings.home_title;
  homeEyebrow.textContent = homeSettings.home_eyebrow;
  renderPosterCarousel(homeSettings.announcements);
  announcementList.replaceChildren();

  homeSettings.announcements.forEach((announcement) => {
    const button = document.createElement("button");
    button.type = "button";
    const label = document.createElement("span");
    const badge = document.createElement("b");
    badge.textContent = "公告";
    label.append(badge, document.createTextNode(announcement.title || "未命名公告"));
    const detail = document.createElement("time");
    detail.textContent = "详情";
    button.append(label, detail);
    button.addEventListener("click", () => openAnnouncement(announcement));
    announcementList.appendChild(button);
  });
}

async function loadHomeSettings() {
  try {
    const data = await api("/api/settings", { method: "GET" });
    renderHome(data.home);
  } catch (error) {
    renderHome(defaultHome);
    console.warn("加载大厅配置失败", error);
  }
}

function openPasswordPanel(mode) {
  passwordPanelMode = mode;
  const isAccountMode = mode === "account";
  const isAdminAccountMode = isAccountMode && isAdmin();
  passwordPanelTitle.textContent = isAdminAccountMode
    ? "管理员安全"
    : isAccountMode
      ? "账号安全"
      : "找回密码";
  passwordPanelHint.textContent = isAdminAccountMode
    ? "验证当前密码后设置管理员新密码"
    : isAccountMode
      ? "验证注册 QQ 后设置当前账号的新密码"
      : "使用原账号和注册时填写的 QQ 验证身份";
  accountDetails.hidden = !isAccountMode || isAdminAccountMode;
  logoutButton.hidden = !isAccountMode;
  resetAccountField.hidden = isAccountMode;
  resetAccount.required = !isAccountMode;
  resetAccount.value = isAccountMode && currentUser
    ? currentUser.account_name
    : loginAccount.value.trim();
  resetAccount.readOnly = isAccountMode;
  resetVerificationLabel.textContent = isAdminAccountMode ? "当前密码" : "注册 QQ";
  resetQq.type = isAdminAccountMode ? "password" : "text";
  resetQq.placeholder = isAdminAccountMode ? "请输入当前管理员密码" : "请输入注册时填写的 QQ";
  if (isAdminAccountMode) resetQq.removeAttribute("inputmode");
  else resetQq.setAttribute("inputmode", "numeric");
  resetQq.value = "";
  resetNewPassword.value = "";
  resetConfirmPassword.value = "";
  resetPasswordStatus.textContent = "";
  passwordOverlay.hidden = false;
  window.setTimeout(() => (isAccountMode ? resetQq : resetAccount).focus(), 0);
}

function closePasswordPanel() {
  passwordOverlay.hidden = true;
  accountSummary.classList.remove("active");
  activateSection("大厅");
}

document.querySelectorAll("[data-auth-tab]").forEach((button) => {
  button.addEventListener("click", () => setAuthTab(button.dataset.authTab));
});

loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  loginSubmit.disabled = true;
  loginSubmit.textContent = "登录中";
  authStatus.textContent = "正在验证账号";
  try {
    const data = await api("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        account_name: loginAccount.value.trim(),
        password: loginPassword.value,
      }),
    });
    if (!["game", "admin"].includes(data.user_type)) {
      throw new Error("服务器返回了未知的账号类型");
    }
    accessToken = data.access_token;
    enterSession(data);
    try {
      await persistLoginChoice();
    } catch (error) {
      showToast(`登录成功，但保存密码失败：${errorMessage(error, "未知错误")}`);
    }
  } catch (error) {
    accessToken = "";
    authStatus.textContent = `登录失败：${errorMessage(error, "未知错误")}`;
  } finally {
    loginSubmit.disabled = false;
    loginSubmit.textContent = "登录";
  }
});

[loginAccount, loginPassword].forEach((input) => {
  input.addEventListener("input", () => {
    authStatus.textContent = "";
    if (rememberedAccount && (!loginAccount.value || !loginPassword.value)) {
      clearSavedCredential();
    }
  });
});

rememberPassword.addEventListener("change", () => {
  if (!rememberPassword.checked) clearSavedCredential();
});

[registerAccount, registerQq].forEach((input) => {
  input.addEventListener("input", () => {
    input.value = input.value.replace(/\D/g, "");
    registerStatus.textContent = "";
  });
});

registerForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (registerPassword.value !== registerConfirmPassword.value) {
    registerStatus.textContent = "两次输入的密码不一致";
    return;
  }
  registerSubmit.disabled = true;
  registerSubmit.textContent = "注册中";
  registerStatus.textContent = "正在创建账号";
  try {
    await api("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({
        account_name: registerAccount.value.trim(),
        password: registerPassword.value,
        qq: registerQq.value.trim(),
      }),
    });
    loginAccount.value = registerAccount.value.trim();
    loginPassword.value = registerPassword.value;
    registerForm.reset();
    setAuthTab("login");
    authStatus.textContent = "注册成功，请登录";
  } catch (error) {
    registerStatus.textContent = `注册失败：${errorMessage(error, "未知错误")}`;
  } finally {
    registerSubmit.disabled = false;
    registerSubmit.textContent = "创建账号";
  }
});

document.querySelector("#forgotPassword").addEventListener("click", () => {
  openPasswordPanel("forgot");
});
document.querySelector("#passwordPanelClose").addEventListener("click", closePasswordPanel);
passwordOverlay.addEventListener("click", (event) => {
  if (event.target === passwordOverlay) closePasswordPanel();
});
document.querySelector("#announcementClose").addEventListener("click", () => {
  announcementOverlay.hidden = true;
});
announcementOverlay.addEventListener("click", (event) => {
  if (event.target === announcementOverlay) announcementOverlay.hidden = true;
});

document.addEventListener("keydown", (event) => {
  if (event.key !== "Escape") return;
  if (!gmConfirmOverlay.hidden) closeGmConfirm();
  else if (!announcementOverlay.hidden) announcementOverlay.hidden = true;
  else if (!passwordOverlay.hidden) closePasswordPanel();
});

resetAccount.addEventListener("input", () => {
  resetAccount.value = resetAccount.value.replace(/\D/g, "");
});

resetQq.addEventListener("input", () => {
  if (!(passwordPanelMode === "account" && isAdmin())) {
    resetQq.value = resetQq.value.replace(/\D/g, "");
  }
});

passwordResetForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (resetNewPassword.value !== resetConfirmPassword.value) {
    resetPasswordStatus.textContent = "两次输入的新密码不一致";
    return;
  }

  resetPasswordSubmit.disabled = true;
  resetPasswordSubmit.textContent = "提交中";
  resetPasswordStatus.textContent = "";
  try {
    const adminAccountMode = passwordPanelMode === "account" && isAdmin();
    if (adminAccountMode) {
      await api("/api/auth/admin/change-password", {
        method: "POST",
        body: JSON.stringify({
          current_password: resetQq.value,
          new_password: resetNewPassword.value,
        }),
      });
    } else {
      await api("/api/auth/change-password", {
        method: "POST",
        body: JSON.stringify({
          account_name: resetAccount.value.trim(),
          qq: resetQq.value.trim(),
          new_password: resetNewPassword.value,
        }),
      });
    }
    if (rememberedAccount === resetAccount.value.trim()) {
      await clearSavedCredential();
      loginPassword.value = "";
    }
    resetQq.value = "";
    resetNewPassword.value = "";
    resetConfirmPassword.value = "";
    resetPasswordStatus.textContent = adminAccountMode
      ? "管理员密码已修改，下次请使用新密码登录"
      : "密码已设置，下次请使用新密码登录";
  } catch (error) {
    resetPasswordStatus.textContent = errorMessage(error, "设置失败");
  } finally {
    resetPasswordSubmit.disabled = false;
    resetPasswordSubmit.textContent = "设置新密码";
  }
});

logoutButton.addEventListener("click", () => {
  leaveSession();
  showToast("已退出当前账号");
});

drawerToggle.addEventListener("click", () => {
  const collapsed = sideDrawer.classList.toggle("collapsed");
  homeView.classList.toggle("drawer-collapsed", collapsed);
  drawerToggle.querySelector("span").textContent = collapsed ? ">" : "<";
  drawerToggle.setAttribute("aria-expanded", String(!collapsed));
  drawerToggle.setAttribute("aria-label", collapsed ? "展开功能导航" : "收起功能导航");
});

document.querySelectorAll(".nav-item").forEach((button) => {
  button.addEventListener("click", () => activateSection(button.dataset.section));
});

gmTargetType.addEventListener("change", () => {
  const byUid = gmTargetType.value === "uid";
  gmTargetInput.value = "";
  gmTargetInput.placeholder = byUid ? "输入目标 UID" : "输入目标账号";
  if (byUid) gmTargetInput.setAttribute("inputmode", "numeric");
  else gmTargetInput.removeAttribute("inputmode");
});

gmTargetInput.addEventListener("input", () => {
  if (gmTargetType.value === "uid") {
    gmTargetInput.value = gmTargetInput.value.replace(/\D/g, "");
  }
});

async function resolveGmTarget() {
  if (!isAdmin() || gmState.resolving) return;
  const value = gmTargetInput.value.trim();
  if (!value) {
    gmTargetSelected.textContent = gmTargetType.value === "uid" ? "请输入目标 UID" : "请输入目标账号";
    return;
  }

  const byUid = gmTargetType.value === "uid";
  gmState.resolving = true;
  gmTargetResolve.disabled = true;
  gmTargetResolve.textContent = "查询中";
  try {
    const target = await api("/api/gm/account/resolve", {
      method: "POST",
      body: JSON.stringify({
        uid: byUid ? Number(value) : null,
        account_name: byUid ? "" : value,
      }),
    });
    gmState.target = target;
    bumpGmTargetRevision();
    gmTargetInput.value = target.account_name || value;
    gmTargetType.value = "account_name";
    gmTargetInput.placeholder = "输入目标账号";
    gmTargetInput.removeAttribute("inputmode");
    gmTargetSelected.textContent = `${target.account_name} · UID ${target.uid}`;
    resetTargetFeatureStates();
    activateSection(activeSection);
    showToast(`已选择目标账号 ${target.account_name}`);
  } catch (error) {
    gmState.target = null;
    bumpGmTargetRevision();
    resetTargetFeatureStates();
    gmTargetSelected.textContent = `选择失败：${errorMessage(error, "未知错误")}`;
  } finally {
    gmState.resolving = false;
    gmTargetResolve.disabled = false;
    gmTargetResolve.textContent = "选择目标";
  }
}

gmTargetResolve.addEventListener("click", resolveGmTarget);
gmTargetInput.addEventListener("keydown", (event) => {
  if (event.key !== "Enter") return;
  event.preventDefault();
  resolveGmTarget();
});

[eventParam1, eventParam2].forEach((input) => {
  input.addEventListener("input", () => {
    input.value = input.value.replace(/\D/g, "");
  });
});

eventSelect.addEventListener("change", () => {
  const selected = eventState.available.find(
    (item) => Number(item.event_id) === Number(eventSelect.value),
  );
  const description = selected?.event_explain || selected?.event_name || "";
  eventParam1.value = description.includes("百分比")
    ? "200"
    : description.includes("倍数")
      ? "2"
      : "1";
});

eventRefresh.addEventListener("click", () => loadEvents(true));

eventForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const eventId = Number(eventSelect.value || "0");
  if (!eventId || eventState.loading) {
    eventStatus.textContent = eventId ? "活动操作正在进行" : "请选择活动";
    return;
  }
  setCharacterButtonBusy(eventAdd, true, "添加中");
  eventStatus.textContent = "正在添加活动";
  try {
    await api("/api/gm/events", {
      method: "POST",
      body: JSON.stringify({
        event_id: eventId,
        parameter1: Number(eventParam1.value || "1"),
        parameter2: Number(eventParam2.value || "0"),
      }),
    });
    eventState.loaded = false;
    await loadEvents(true);
    eventStatus.textContent = "活动已添加，通常需要重启游戏服务";
  } catch (error) {
    eventStatus.textContent = `添加失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(eventAdd, false);
  }
});

eventDelete.addEventListener("click", () => {
  const logId = eventState.selectedLogId;
  if (!logId) {
    eventStatus.textContent = "请选择要删除的活动";
    return;
  }
  openGmConfirm(
    "删除运行活动",
    `确定删除活动日志 #${logId}？活动变更后通常需要重启游戏服务。`,
    "确认删除",
    async () => {
      eventStatus.textContent = "正在删除活动";
      try {
        await api(`/api/gm/events/${logId}`, { method: "DELETE" });
        eventState.loaded = false;
        await loadEvents(true);
        eventStatus.textContent = "活动已删除，通常需要重启游戏服务";
      } catch (error) {
        eventStatus.textContent = `删除失败：${errorMessage(error, "未知错误")}`;
      }
    },
  );
});

banDays.addEventListener("input", () => {
  banDays.value = banDays.value.replace(/\D/g, "");
  if (Number(banDays.value || "0") > 3650) banDays.value = "3650";
});

banQuery.addEventListener("click", queryBanStatus);

banForm.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!requireGmTarget(banStatus)) return;
  const days = Number(banDays.value || "0");
  if (days <= 0) {
    banStatus.textContent = "封禁天数必须大于 0";
    return;
  }
  const target = gmState.target;
  openGmConfirm(
    "执行账号封禁",
    `确定对账号 ${target.account_name}（UID ${target.uid}）执行 ${days} 天限制？`,
    "确认封禁",
    async () => {
      setCharacterButtonBusy(banSubmit, true, "执行中");
      banStatus.textContent = "正在提交封禁";
      try {
        const data = await api("/api/gm/ban/set", {
          method: "POST",
          body: JSON.stringify({
            ...rechargeTargetPayload(),
            punish_type: Number(banPunishType.value),
            days,
            reason: banReason.value.trim(),
          }),
        });
        renderBanStatus(data);
        banStatus.textContent = "封禁已完成";
      } catch (error) {
        banStatus.textContent = `封禁失败：${errorMessage(error, "未知错误")}`;
      } finally {
        setCharacterButtonBusy(banSubmit, false);
      }
    },
  );
});

banUnban.addEventListener("click", () => {
  if (!requireGmTarget(banStatus)) return;
  const target = gmState.target;
  openGmConfirm(
    "解除账号限制",
    `确定解除账号 ${target.account_name}（UID ${target.uid}）的当前限制？`,
    "确认解除",
    async () => {
      setCharacterButtonBusy(banUnban, true, "执行中");
      banStatus.textContent = "正在解除限制";
      try {
        const data = await api("/api/gm/ban/unban", {
          method: "POST",
          body: JSON.stringify(rechargeTargetPayload()),
        });
        renderBanStatus(data);
        banStatus.textContent = "限制已解除";
      } catch (error) {
        banStatus.textContent = `解除失败：${errorMessage(error, "未知错误")}`;
      } finally {
        setCharacterButtonBusy(banUnban, false);
      }
    },
  );
});

permissionSearchForm.addEventListener("submit", (event) => {
  event.preventDefault();
  permissionState.loaded = false;
  loadPermissionAccounts(true);
});

permissionForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!permissionState.selected) return;
  const permissions = [...permissionGrid.querySelectorAll('input[type="checkbox"]:checked')]
    .map((input) => input.value);
  setCharacterButtonBusy(permissionSave, true, "保存中");
  permissionStatus.textContent = "正在保存账号权限";
  try {
    const data = await api(`/api/admin/accounts/${permissionState.selected.uid}/permissions`, {
      method: "PUT",
      body: JSON.stringify({ permissions }),
    });
    permissionState.selected.permissions = data.permissions || [];
    renderPermissionGrid(permissionState.selected.permissions);
    permissionStatus.textContent = "账号权限已保存";
  } catch (error) {
    permissionStatus.textContent = `保存失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(permissionSave, false);
  }
});

document.querySelectorAll("[data-system-tab]").forEach((button) => {
  button.addEventListener("click", () => setSystemTab(button.dataset.systemTab));
});

announcementAdd.addEventListener("click", () => {
  if (systemState.announcements.length >= 8) return;
  systemState.announcements.push(announcementValue());
  renderAnnouncementEditor();
});

systemHomePanel.addEventListener("submit", async (event) => {
  event.preventDefault();
  setCharacterButtonBusy(systemHomeSave, true, "保存中");
  systemHomeStatus.textContent = "正在保存大厅设置";
  try {
    const homeData = await api("/api/admin/settings/home", {
      method: "PUT",
      body: JSON.stringify({
        home_title: systemHomeTitle.value.trim(),
        home_eyebrow: systemHomeEyebrow.value.trim(),
        client_download_url: systemClientDownloadUrl.value.trim(),
        announcements: systemState.announcements.map(announcementValue),
      }),
    });
    renderHome(homeData.home);
    systemState.announcements = (homeData.home?.announcements || []).map(announcementValue);
    renderAnnouncementEditor();
    systemHomeStatus.textContent = "大厅内容已保存";
  } catch (error) {
    systemHomeStatus.textContent = `保存失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(systemHomeSave, false);
  }
});

pvfRefreshForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const path = pvfPath.value.trim() || pvfPath.placeholder.trim();
  if (!path) {
    pvfStatus.textContent = "请输入服务端 Script.pvf 路径";
    return;
  }
  pvfPath.value = path;
  setCharacterButtonBusy(pvfRefresh, true, "导入中");
  pvfStatus.textContent = "PVF 读取中，请等待";
  pvfLog.hidden = true;
  try {
    const data = await api("/api/admin/pvf/refresh", {
      method: "POST",
      timeoutMs: 120000,
      body: JSON.stringify({ pvf_path: path, encode: pvfEncode.value }),
    });
    renderPvfStatus({ ...data, loaded: true, pvf_path: data.path, updated_at: "刚刚" });
    if (data.logs?.length) {
      pvfLog.textContent = data.logs.join("\n");
      pvfLog.hidden = false;
    }
    characterJobOptions = null;
    characterJobOptionsLoading = null;
  } catch (error) {
    pvfStatus.textContent = `PVF 导入失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(pvfRefresh, false);
  }
});

pvfClientMd5.addEventListener("input", () => {
  pvfClientMd5.value = pvfClientMd5.value.replace(/[^0-9a-f]/gi, "").toUpperCase();
});

pvfMd5Form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const clientPvfMd5 = pvfClientMd5.value.trim().toUpperCase();
  if (clientPvfMd5 && clientPvfMd5.length !== 32) {
    pvfMd5Status.textContent = "请输入 32 位 MD5，或留空关闭校验";
    return;
  }
  setCharacterButtonBusy(pvfMd5Save, true, "保存中");
  try {
    const data = await api("/api/admin/pvf/client-md5", {
      method: "PUT",
      body: JSON.stringify({ client_pvf_md5: clientPvfMd5 }),
    });
    pvfClientMd5.value = data.client_pvf_md5 || "";
    pvfMd5Status.textContent = data.client_pvf_md5
      ? "客户端 PVF MD5 已更新"
      : "已关闭客户端 PVF 校验";
  } catch (error) {
    pvfMd5Status.textContent = `保存失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(pvfMd5Save, false);
  }
});

logSearchForm.addEventListener("submit", (event) => {
  event.preventDefault();
  loadOperationLogs(1);
});
logPrev.addEventListener("click", () => loadOperationLogs(systemState.logPage - 1));
logNext.addEventListener("click", () => loadOperationLogs(systemState.logPage + 1));

gmConfirmCancel.addEventListener("click", closeGmConfirm);
gmConfirmOverlay.addEventListener("click", (event) => {
  if (event.target === gmConfirmOverlay) closeGmConfirm();
});
gmConfirmSubmit.addEventListener("click", async () => {
  if (!gmConfirmAction) return;
  const action = gmConfirmAction;
  gmConfirmSubmit.disabled = true;
  try {
    await action();
  } finally {
    closeGmConfirm();
  }
});

characterSearch.addEventListener("input", () => {
  characterState.page = 1;
  renderCharacterPicker();
});

characterSearch.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    characterState.page = 1;
    renderCharacterPicker();
  }
});

characterRefresh.addEventListener("click", () => loadCharacterPicker(true));

characterPrev.addEventListener("click", () => {
  characterState.page = Math.max(1, characterState.page - 1);
  renderCharacterPicker();
});

characterNext.addEventListener("click", () => {
  characterState.page += 1;
  renderCharacterPicker();
});

mailRecipientSearch.addEventListener("input", () => {
  mailState.selectedCharacterNo = null;
  mailState.page = 1;
  mailSelectedRecipient.textContent = "未选择收件角色";
  renderMailCharacterPicker();
});

mailRecipientSearch.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    mailState.page = 1;
    renderMailCharacterPicker();
  }
});

mailCharacterPrev.addEventListener("click", () => {
  mailState.page = Math.max(1, mailState.page - 1);
  renderMailCharacterPicker();
});

mailCharacterNext.addEventListener("click", () => {
  mailState.page += 1;
  renderMailCharacterPicker();
});

mailItemId.addEventListener("input", () => {
  mailItemId.value = mailItemId.value.replace(/\D/g, "");
  mailState.selectedItemStackLimit = null;
  clampMailItemCount();
});

mailItemCount.addEventListener("input", () => {
  clampMailItemCount();
});

mailGold.addEventListener("input", () => {
  mailGold.value = mailGold.value.replace(/\D/g, "");
});

mailEnhancementLevel.addEventListener("input", () => {
  mailEnhancementLevel.value = mailEnhancementLevel.value.replace(/\D/g, "");
  if (Number(mailEnhancementLevel.value || "0") > 31) mailEnhancementLevel.value = "31";
});

mailForgeLevel.addEventListener("input", () => {
  mailForgeLevel.value = mailForgeLevel.value.replace(/\D/g, "");
  if (Number(mailForgeLevel.value || "0") > 31) mailForgeLevel.value = "31";
});

mailAmplifyValue.addEventListener("input", () => {
  mailAmplifyValue.value = mailAmplifyValue.value.replace(/\D/g, "");
});

mailAttachmentType.addEventListener("change", () => {
  if (mailAttachmentType.value === "equipment") {
    mailState.selectedItemStackLimit = null;
  }
  updateMailAttachmentFields();
});

mailSearchItem.addEventListener("click", () => searchMailItems(1));

mailItemKeyword.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    searchMailItems(1);
  }
});

mailItemPrev.addEventListener("click", () => searchMailItems(mailState.itemPage - 1));
mailItemNext.addEventListener("click", () => searchMailItems(mailState.itemPage + 1));

mailForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = mailPayload();
  if (!payload.charac_no) {
    mailStatus.textContent = "请先从角色列表选择收件人";
    return;
  }
  if (payload.item_id > 0 && payload.item_count <= 0) {
    mailStatus.textContent = "物品数量必须大于 0";
    return;
  }
  setCharacterButtonBusy(mailSend, true, "发送中");
  mailStatus.textContent = "发送中";
  try {
    const data = await api("/api/gm/mail/send", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    mailStatus.textContent = `发送完成：邮件 ${data.letter_id}`;
  } catch (error) {
    mailStatus.textContent = `发送失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(mailSend, false);
  }
});

mailSendAll.addEventListener("click", () => {
  const { charac_no: _characNo, ...payload } = mailPayload();
  if (payload.item_id > 0 && payload.item_count <= 0) {
    mailStatus.textContent = "物品数量必须大于 0";
    return;
  }
  if (!payload.message && !payload.item_id && payload.gold <= 0) {
    mailStatus.textContent = "请输入邮件内容、物品或金币";
    return;
  }
  openGmConfirm(
    "发送全服邮件",
    "确定将当前邮件发送给全服所有正常角色？角色数量较多时操作可能需要等待。",
    "确认发送",
    async () => {
      setCharacterButtonBusy(mailSendAll, true, "发送中");
      mailStatus.textContent = "全服邮件发送中";
      try {
        const data = await api("/api/gm/mail/send-all", {
          method: "POST",
          timeoutMs: 120000,
          body: JSON.stringify(payload),
        });
        mailStatus.textContent = `全服发送完成：${data.target_count || 0} 个角色`;
      } catch (error) {
        mailStatus.textContent = `全服发送失败：${errorMessage(error, "未知错误")}`;
      } finally {
        setCharacterButtonBusy(mailSendAll, false);
      }
    },
  );
});

mailDeleteAll.addEventListener("click", () => {
  openGmConfirm(
    "清空全服邮件",
    "确定删除全服所有角色邮箱中的全部未删除邮件？该操作不可撤销。",
    "确认清空",
    async () => {
      setCharacterButtonBusy(mailDeleteAll, true, "清空中");
      mailStatus.textContent = "全服邮件删除中";
      try {
        const data = await api("/api/gm/mail/delete-all", {
          method: "POST",
          timeoutMs: 120000,
        });
        mailStatus.textContent = `全服邮件删除完成：${data.deleted_count || 0} 封`;
      } catch (error) {
        mailStatus.textContent = `全服邮件删除失败：${errorMessage(error, "未知错误")}`;
      } finally {
        setCharacterButtonBusy(mailDeleteAll, false);
      }
    },
  );
});

mailDelete.addEventListener("click", () => {
  const characNo = mailState.selectedCharacterNo;
  const characName = mailRecipientSearch.value.trim();
  if (!characNo) {
    mailStatus.textContent = "请先从角色列表选择要删除邮件的角色";
    return;
  }
  mailConfirmText.textContent = `确定删除角色 ${characName || characNo} 邮箱中的所有未删除邮件？`;
  mailConfirmOverlay.hidden = false;
});

mailConfirmCancel.addEventListener("click", () => {
  mailConfirmOverlay.hidden = true;
});

mailConfirmOverlay.addEventListener("click", (event) => {
  if (event.target === mailConfirmOverlay) mailConfirmOverlay.hidden = true;
});

mailConfirmSubmit.addEventListener("click", async () => {
  const characNo = mailState.selectedCharacterNo;
  if (!characNo) {
    mailConfirmOverlay.hidden = true;
    mailStatus.textContent = "请先从角色列表选择要删除邮件的角色";
    return;
  }
  setCharacterButtonBusy(mailConfirmSubmit, true, "清空中");
  mailStatus.textContent = "角色邮件删除中";
  try {
    const data = await api("/api/gm/mail/delete", {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo }),
    });
    mailStatus.textContent = `角色邮件删除完成：${data.deleted_count || 0} 封`;
  } catch (error) {
    mailStatus.textContent = `角色邮件删除失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(mailConfirmSubmit, false);
    mailConfirmOverlay.hidden = true;
  }
});

inventoryCharacterSearch.addEventListener("input", () => {
  inventoryState.selectedCharacterNo = null;
  inventoryState.selectedSlot = null;
  inventoryState.page = 1;
  inventorySelectedCharacter.textContent = "尚未选择角色";
  inventoryDelete.disabled = true;
  renderInventoryCharacterPicker();
});

inventoryCharacterSearch.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    inventoryState.page = 1;
    renderInventoryCharacterPicker();
  }
});

inventoryCharacterPrev.addEventListener("click", () => {
  inventoryState.page = Math.max(1, inventoryState.page - 1);
  renderInventoryCharacterPicker();
});

inventoryCharacterNext.addEventListener("click", () => {
  inventoryState.page += 1;
  renderInventoryCharacterPicker();
});

inventoryScope.addEventListener("change", () => {
  inventoryState.selectedSlot = null;
  inventoryState.items = [];
  inventoryScopeSummary.textContent = inventoryScopeName();
  inventoryItemCount.textContent = "0 个物品";
  inventoryDelete.disabled = true;
  inventoryList.replaceChildren();
  const empty = document.createElement("div");
  empty.className = "inventory-empty";
  empty.textContent = "范围已切换，请重新查询";
  inventoryList.appendChild(empty);
  inventoryStatus.textContent = "范围已切换，请重新查询背包";
});

inventoryQuery.addEventListener("click", queryInventory);

inventoryDelete.addEventListener("click", () => {
  if (!inventoryState.selectedCharacterNo || inventoryState.selectedSlot === null) {
    inventoryStatus.textContent = "请选择要删除的物品槽位";
    return;
  }
  inventoryConfirmMode = "delete";
  inventoryConfirmTitle.textContent = "删除选中物品";
  inventoryConfirmText.textContent = `确定删除槽位 ${inventoryState.selectedSlot} 的物品？`;
  inventoryConfirmSubmit.textContent = "确认删除";
  inventoryConfirmOverlay.hidden = false;
});

inventoryClear.addEventListener("click", () => {
  if (!inventoryState.selectedCharacterNo) {
    inventoryStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  inventoryConfirmMode = "clear";
  inventoryConfirmTitle.textContent = "清空当前栏位";
  inventoryConfirmText.textContent = `确定清空当前${inventoryScopeName()}？`;
  inventoryConfirmSubmit.textContent = "确认清空";
  inventoryConfirmOverlay.hidden = false;
});

inventoryConfirmCancel.addEventListener("click", () => {
  inventoryConfirmMode = "";
  inventoryConfirmOverlay.hidden = true;
});

inventoryConfirmOverlay.addEventListener("click", (event) => {
  if (event.target === inventoryConfirmOverlay) {
    inventoryConfirmMode = "";
    inventoryConfirmOverlay.hidden = true;
  }
});

inventoryConfirmSubmit.addEventListener("click", async () => {
  const characNo = inventoryState.selectedCharacterNo;
  const mode = inventoryConfirmMode;
  if (!characNo || !mode) {
    inventoryConfirmOverlay.hidden = true;
    return;
  }
  if (mode === "delete" && inventoryState.selectedSlot === null) {
    inventoryConfirmOverlay.hidden = true;
    inventoryStatus.textContent = "请选择要删除的物品槽位";
    return;
  }
  setCharacterButtonBusy(inventoryConfirmSubmit, true, mode === "delete" ? "删除中" : "清空中");
  inventoryStatus.textContent = mode === "delete" ? "删除中" : "清空中";
  try {
    const payload = { charac_no: characNo, scope: inventoryScope.value };
    let data;
    if (mode === "delete") {
      payload.slot = inventoryState.selectedSlot;
      data = await api("/api/gm/inventory/delete", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      renderInventory(data);
      inventoryStatus.textContent = `删除完成：槽位 ${data.deleted_slots.join(", ")}`;
    } else {
      data = await api("/api/gm/inventory/clear", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      renderInventory(data);
      inventoryStatus.textContent = `清空完成：${data.deleted_slots.length} 个槽位`;
    }
  } catch (error) {
    inventoryStatus.textContent = `${mode === "delete" ? "删除" : "清空"}失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(inventoryConfirmSubmit, false);
    inventoryConfirmMode = "";
    inventoryConfirmOverlay.hidden = true;
  }
});

avatarCharacterSearch.addEventListener("input", () => {
  avatarState.selectedCharacterNo = null;
  avatarState.selectedUiIds = [];
  avatarState.page = 1;
  avatarSelectedCharacter.textContent = "尚未选择角色";
  avatarSelectionSummary.textContent = "未选择时装";
  avatarApply.disabled = true;
  renderAvatarCharacterPicker();
});

avatarCharacterSearch.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    avatarState.page = 1;
    renderAvatarCharacterPicker();
  }
});

avatarCharacterPrev.addEventListener("click", () => {
  avatarState.page = Math.max(1, avatarState.page - 1);
  renderAvatarCharacterPicker();
});

avatarCharacterNext.addEventListener("click", () => {
  avatarState.page += 1;
  renderAvatarCharacterPicker();
});

avatarQuery.addEventListener("click", queryAvatar);

avatarApply.addEventListener("click", async () => {
  const characNo = avatarState.selectedCharacterNo;
  if (!characNo) {
    avatarStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  if (!avatarState.selectedUiIds.length) {
    avatarStatus.textContent = "请选择要修改的时装";
    return;
  }
  setCharacterButtonBusy(avatarApply, true, "设置中");
  avatarStatus.textContent = "设置潜能中";
  try {
    await api("/api/gm/avatar/hidden", {
      method: "POST",
      body: JSON.stringify({
        charac_no: characNo,
        ui_ids: avatarState.selectedUiIds,
        hidden_option: Number(avatarHiddenOption.value || "0"),
      }),
    });
    await queryAvatar();
    avatarStatus.textContent = "潜能已设置";
  } catch (error) {
    avatarStatus.textContent = `设置失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(avatarApply, false);
    avatarApply.disabled = avatarState.selectedUiIds.length === 0;
  }
});

rechargeAmount.addEventListener("input", () => {
  rechargeAmount.value = rechargeAmount.value.replace(/\D/g, "");
  if (Number(rechargeAmount.value || "0") > 2147483647) {
    rechargeAmount.value = "2147483647";
  }
});

rechargeRefresh.addEventListener("click", queryRechargeBalance);

rechargeForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!currentUser || rechargeState.loading) return;
  if (!requireGmTarget(rechargeStatus)) return;
  const amount = Number(rechargeAmount.value || "0");
  if (rechargeAction.value === "add" && amount <= 0) {
    rechargeStatus.textContent = "充值数量必须大于 0";
    return;
  }

  rechargeState.loading = true;
  setCharacterButtonBusy(rechargeSubmit, true, "提交中");
  rechargeRefresh.disabled = true;
  rechargeStatus.textContent = "正在提交充值";
  try {
    const data = await api("/api/gm/cera/charge", {
      method: "POST",
      body: JSON.stringify({
        ...rechargeTargetPayload(),
        cera_type: rechargeType.value,
        action: rechargeAction.value,
        amount,
      }),
    });
    renderRechargeBalance(data);
    rechargeStatus.textContent = "充值已完成，余额已更新";
    showToast("充值已完成");
  } catch (error) {
    rechargeStatus.textContent = `充值失败：${errorMessage(error, "未知错误")}`;
  } finally {
    rechargeState.loading = false;
    setCharacterButtonBusy(rechargeSubmit, false);
    rechargeRefresh.disabled = false;
  }
});

rapidFireKey.addEventListener("input", () => {
  const printable = [...rapidFireKey.value].filter((character) => {
    const code = character.charCodeAt(0);
    return code >= 33 && code <= 126;
  });
  rapidFireKey.value = (printable.at(-1) || "").toUpperCase();
});

rapidFireInterval.addEventListener("input", () => {
  rapidFireInterval.value = rapidFireInterval.value.replace(/\D/g, "");
  if (Number(rapidFireInterval.value || "0") > 10000) rapidFireInterval.value = "10000";
});

rapidFireInstallDriver.addEventListener("click", async () => {
  if (rapidFireState.loading) return;
  if (!isTauri) {
    rapidFireStatus.textContent = "浏览器预览不执行驱动安装";
    return;
  }
  rapidFireState.loading = true;
  setCharacterButtonBusy(rapidFireInstallDriver, true, "启动中");
  rapidFireStatus.textContent = "正在请求管理员权限";
  try {
    const snapshot = await invoke("install_interception_driver");
    renderRapidFire(snapshot);
    rapidFireStatus.textContent = "驱动安装程序已启动，请完成后重启电脑";
  } catch (error) {
    rapidFireStatus.textContent = `启动安装失败：${errorMessage(error, "未知错误")}`;
  } finally {
    rapidFireState.loading = false;
    setCharacterButtonBusy(rapidFireInstallDriver, false);
  }
});

rapidFireForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (rapidFireState.loading) return;
  const key = rapidFireKey.value.trim().toUpperCase();
  const intervalMs = Number(rapidFireInterval.value || "0");
  if (!/^[!-~]$/.test(key)) {
    rapidFireStatus.textContent = "请输入一个可直接输入的按键";
    return;
  }
  if (intervalMs < 1 || intervalMs > 10000) {
    rapidFireStatus.textContent = "连发间隔必须在 1 到 10000 毫秒之间";
    return;
  }
  if (rapidFireState.configs.some((config) => config.key === key)) {
    rapidFireStatus.textContent = "该按键已存在连发配置";
    return;
  }

  rapidFireState.loading = true;
  setCharacterButtonBusy(rapidFireAdd, true, "添加中");
  rapidFireStatus.textContent = `正在添加 ${key} 键`;
  try {
    const snapshot = isTauri
      ? await invoke("add_rapid_fire", { key, intervalMs })
      : {
          configs: [...rapidFireState.configs, { key, intervalMs }],
          ready: false,
          error: null,
        };
    renderRapidFire(snapshot);
    rapidFireKey.value = "";
    rapidFireStatus.textContent = `${key} 键连发已添加`;
  } catch (error) {
    rapidFireStatus.textContent = `添加失败：${errorMessage(error, "未知错误")}`;
  } finally {
    rapidFireState.loading = false;
    setCharacterButtonBusy(rapidFireAdd, false);
  }
});

characterEditOpen.addEventListener("click", async () => {
  const character = selectedCharacter();
  if (!character) {
    characterStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  characterEditStatus.textContent = "职业选项加载中";
  characterEditorOverlay.hidden = false;
  const loaded = await loadCharacterJobOptions();
  if (!loaded) {
    characterEditStatus.textContent = "职业选项加载失败，请关闭后重试";
    return;
  }
  fillCharacterEditor(character);
  characterEditStatus.textContent = "请选择要修改的内容";
});

characterEditorClose.addEventListener("click", () => {
  characterEditorOverlay.hidden = true;
});

characterEditorOverlay.addEventListener("click", (event) => {
  if (event.target === characterEditorOverlay) characterEditorOverlay.hidden = true;
});

characterEditJob.addEventListener("change", () => renderCharacterGrowTypes());

characterEditLevel.addEventListener("input", () => {
  characterEditLevel.value = characterEditLevel.value.replace(/\D/g, "");
});

characterEditPvpPoint.addEventListener("input", () => {
  characterEditPvpPoint.value = characterEditPvpPoint.value.replace(/\D/g, "");
});

characterEditLevelSubmit.addEventListener("click", async () => {
  const characNo = characterState.selectedCharacterNo;
  const level = Number(characterEditLevel.value.trim() || "0");
  if (!characNo) {
    characterEditStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  if (level <= 0) {
    characterEditStatus.textContent = "等级必须大于 0";
    return;
  }
  setCharacterButtonBusy(characterEditLevelSubmit, true, "修改中");
  characterEditStatus.textContent = "修改中";
  try {
    const data = await api("/api/gm/character/level", {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo, level }),
    });
    refreshCharacterInState(data.character);
    characterEditStatus.textContent = "等级已修改";
    characterStatus.textContent = "等级已修改";
  } catch (error) {
    characterEditStatus.textContent = `修改失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(characterEditLevelSubmit, false);
  }
});

characterEditPvpGradeSubmit.addEventListener("click", async () => {
  const characNo = characterState.selectedCharacterNo;
  if (!characNo) {
    characterEditStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  setCharacterButtonBusy(characterEditPvpGradeSubmit, true, "修改中");
  characterEditStatus.textContent = "修改决斗等级中";
  try {
    const data = await api("/api/gm/character/pvp-grade", {
      method: "POST",
      body: JSON.stringify({
        charac_no: characNo,
        pvp_grade: Number(characterEditPvpGrade.value || "0"),
      }),
    });
    refreshCharacterInState(data.character);
    characterEditStatus.textContent = "决斗等级已修改";
    characterStatus.textContent = "决斗等级已修改";
  } catch (error) {
    characterEditStatus.textContent = `修改失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(characterEditPvpGradeSubmit, false);
  }
});

characterEditPvpPointSubmit.addEventListener("click", async () => {
  const characNo = characterState.selectedCharacterNo;
  const pvpPoint = Number(characterEditPvpPoint.value.trim() || "0");
  if (!characNo) {
    characterEditStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  setCharacterButtonBusy(characterEditPvpPointSubmit, true, "修改中");
  characterEditStatus.textContent = "修改决斗胜点中";
  try {
    const data = await api("/api/gm/character/pvp-point", {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo, pvp_point: pvpPoint }),
    });
    refreshCharacterInState(data.character);
    characterEditStatus.textContent = "决斗胜点已修改";
    characterStatus.textContent = "决斗胜点已修改";
  } catch (error) {
    characterEditStatus.textContent = `修改失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(characterEditPvpPointSubmit, false);
  }
});

characterEditJobSubmit.addEventListener("click", async () => {
  const characNo = characterState.selectedCharacterNo;
  if (!characNo) {
    characterEditStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  setCharacterButtonBusy(characterEditJobSubmit, true, "修改中");
  characterEditStatus.textContent = "修改职业中";
  try {
    const data = await api("/api/gm/character/job", {
      method: "POST",
      body: JSON.stringify({
        charac_no: characNo,
        job: Number(characterEditJob.value || "0"),
        grow_type: Number(characterEditGrowType.value || "0"),
        wake_flag: Number(characterEditWakeFlag.value || "0"),
        expert_job: Number(characterEditExpertJob.value || "0"),
      }),
    });
    refreshCharacterInState(data.character);
    characterEditStatus.textContent = "职业已修改";
    characterStatus.textContent = "职业已修改";
  } catch (error) {
    characterEditStatus.textContent = `修改失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(characterEditJobSubmit, false);
  }
});

async function updateCharacterVisibility(endpoint, button, busyText, doneText) {
  const characNo = characterState.selectedCharacterNo;
  if (!characNo) {
    characterEditStatus.textContent = "请先从角色列表选择角色";
    return;
  }
  setCharacterButtonBusy(button, true, busyText);
  characterEditStatus.textContent = busyText;
  try {
    const data = await api(endpoint, {
      method: "POST",
      body: JSON.stringify({ charac_no: characNo }),
    });
    refreshCharacterInState(data.character);
    characterEditStatus.textContent = doneText;
    characterStatus.textContent = doneText;
  } catch (error) {
    characterEditStatus.textContent = `操作失败：${errorMessage(error, "未知错误")}`;
  } finally {
    setCharacterButtonBusy(button, false);
    const character = selectedCharacter();
    if (character) fillCharacterEditor(character);
  }
}

characterDelete.addEventListener("click", () => updateCharacterVisibility(
  "/api/gm/character/delete",
  characterDelete,
  "删除角色中",
  "角色已标记为已删除",
));

characterRecover.addEventListener("click", () => updateCharacterVisibility(
  "/api/gm/character/recover",
  characterRecover,
  "恢复角色中",
  "角色已恢复为正常",
));

accountSummary.addEventListener("click", () => {
  activateSection("账号");
  openPasswordPanel("account");
});

async function monitorGameProcess() {
  window.clearTimeout(gameMonitorTimer);
  if (!currentUser?.can_launch || !isTauri) return;
  try {
    const running = await invoke("is_game_running");
    if (running) {
      clientState.innerHTML = "<i></i> 运行中";
      setGameButtonMode(true);
      launchButton.disabled = false;
      gameMonitorTimer = window.setTimeout(monitorGameProcess, 2000);
      return;
    }
  } catch (error) {
    console.error("读取 DNF.exe 运行状态失败", error);
  }
  setGameButtonMode(false, { keepUpdateRequired: true });
  if (!clientUpdateRequired) clientState.innerHTML = "<i></i> 客户端就绪";
  launchButton.disabled = false;
}

launchButton.addEventListener("click", async () => {
  if (!currentUser?.can_launch) return;
  if (clientUpdateRequired) {
    if (!clientDownloadUrl) {
      showToast("尚未配置客户端下载地址");
      return;
    }
    try {
      if (isTauri) await invoke("open_url", { url: clientDownloadUrl });
      else window.open(clientDownloadUrl, "_blank", "noopener");
    } catch (error) {
      showToast(`打开下载地址失败：${errorMessage(error, "未知错误")}`);
    }
    return;
  }
  if (gameRunning) {
    launchButton.disabled = true;
    clientState.innerHTML = "<i></i> 正在结束";
    try {
      await invoke("stop_game");
      window.clearTimeout(gameMonitorTimer);
      setGameButtonMode(false);
      clientState.innerHTML = "<i></i> 客户端就绪";
      showToast("DNF.exe 已结束");
    } catch (error) {
      clientState.innerHTML = "<i></i> 结束失败";
      showToast(`结束游戏失败：${errorMessage(error, "未知错误")}`);
    } finally {
      launchButton.disabled = false;
    }
    return;
  }
  launchButton.disabled = true;
  clientState.innerHTML = "<i></i> 正在启动";
  try {
    const data = await api("/api/launcher/direct", { method: "POST" });
    if (!data.dnf_token) throw new Error("服务器未返回 DNF 登录参数");
    if (isTauri) {
      await invoke("launch_game", {
        dnfToken: data.dnf_token,
        expectedPvfMd5: data.client_pvf_md5 || "",
      });
      showToast("DNF.exe 已启动");
      setGameButtonMode(true);
      clientState.innerHTML = "<i></i> 运行中";
      launchButton.disabled = false;
      gameMonitorTimer = window.setTimeout(monitorGameProcess, 1200);
    } else {
      showToast("浏览器预览不执行本机 DNF.exe");
      clientState.innerHTML = "<i></i> 客户端就绪";
      launchButton.disabled = false;
    }
  } catch (error) {
    if (String(errorMessage(error, "")).startsWith("CLIENT_PVF_MISMATCH")) {
      setClientUpdateMode();
      showToast(clientDownloadUrl ? "客户端待更新，请点击更新下载" : "客户端待更新，但未配置下载地址");
      return;
    }
    clientState.innerHTML = "<i></i> 启动失败";
    showToast(`启动失败：${errorMessage(error, "未知错误")}`);
    window.setTimeout(() => {
      if (!currentUser) return;
      launchButton.disabled = false;
      clientState.innerHTML = "<i></i> 客户端就绪";
    }, 1400);
  }
});

document.querySelectorAll("[data-window-action]").forEach((button) => {
  button.addEventListener("click", async () => {
    const action = button.dataset.windowAction;
    if (!isTauri) {
      showToast(action === "close" ? "浏览器预览中无法关闭窗口" : "浏览器预览中无法最小化");
      return;
    }
    const appWindow = getCurrentWindow();
    if (action === "minimize") await appWindow.minimize();
    else if (action === "close") await appWindow.close();
  });
});

async function revealWindow() {
  if (!isTauri) return;
  await configurationReady;
  await new Promise((resolve) => window.requestAnimationFrame(() => window.requestAnimationFrame(resolve)));
  await getCurrentWindow().show();
  try {
    const splashWindow = await WebviewWindow.getByLabel("splash");
    await splashWindow?.close();
  } catch (error) {
    console.warn("关闭启动提示窗口失败", error);
  }
}

renderHome(defaultHome);
const configurationReady = Promise.all([loadLauncherWindowTitle(), loadLauncherBackground()]);
configurationReady.then(loadHomeSettings);
loadSavedCredential();
if (document.readyState === "complete") revealWindow();
else window.addEventListener("load", revealWindow, { once: true });
