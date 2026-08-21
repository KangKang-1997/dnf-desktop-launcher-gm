const isWebView2Runtime = Boolean(window.chrome?.webview);
const webView2Bridge = window.dnfNative || null;

async function invokeNative(command, args = {}) {
  if (webView2Bridge?.invoke) {
    return webView2Bridge.invoke(command, args);
  }
  throw new Error("当前环境不支持本机功能");
}

async function setWindowTitle(title) {
  if (webView2Bridge?.setWindowTitle) {
    return webView2Bridge.setWindowTitle(title);
  }
  return null;
}

async function minimizeWindow() {
  if (webView2Bridge?.minimizeWindow) {
    return webView2Bridge.minimizeWindow();
  }
  throw new Error("当前环境不支持最小化窗口");
}

async function startWindowDrag() {
  if (webView2Bridge?.startWindowDrag) {
    return webView2Bridge.startWindowDrag();
  }
  return null;
}

async function closeWindow() {
  if (webView2Bridge?.closeWindow) {
    return webView2Bridge.closeWindow();
  }
  throw new Error("当前环境不支持关闭窗口");
}

async function revealWindow() {
  if (webView2Bridge?.revealWindow) {
    return webView2Bridge.revealWindow();
  }
  return null;
}

export const native = {
  isAvailable: isWebView2Runtime || Boolean(webView2Bridge),
  runtime: isWebView2Runtime ? "webview2" : "browser",
  invoke: invokeNative,
  setWindowTitle,
  startWindowDrag,
  minimizeWindow,
  closeWindow,
  revealWindow,
};
