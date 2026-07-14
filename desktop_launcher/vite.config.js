import { defineConfig } from "vite";

export default defineConfig(() => {
  const apiBase = String(process.env.DNF_LAUNCHER_API_BASE || "").replace(/\/+$/, "");
  if (!/^https?:\/\//.test(apiBase)) {
    throw new Error("DNF_LAUNCHER_API_BASE must be set and start with http:// or https://");
  }

  return {
    clearScreen: false,
    define: {
      __DNF_LAUNCHER_API_BASE__: JSON.stringify(apiBase),
    },
    server: {
      host: "127.0.0.1",
      port: 1420,
      strictPort: true,
      watch: {
        ignored: ["**/src-tauri/**"],
      },
    },
  };
});
