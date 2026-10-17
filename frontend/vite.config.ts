import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import wails from "@wailsio/runtime/plugins/vite";

// Set by scripts/smoke-browser-loop.sh (browser UAT against the -tags server build); unset for wails3 dev.
const backend = process.env.GIFSMITH_BACKEND;

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
    proxy: backend ? { "/wails": { target: backend, ws: true } } : undefined,
  },
  plugins: [svelte(), wails("./bindings")],
});
