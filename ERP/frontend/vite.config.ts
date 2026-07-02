import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_DEV_API_TARGET || "http://127.0.0.1:8088";
const configuredDevPort = process.env.VITE_DEV_PORT || process.env.PORT;
const devPort = Number(configuredDevPort || 9245);

export default defineConfig({
  plugins: [react()],
  server: {
    host: "127.0.0.1",
    port: devPort,
    strictPort: false,
    hmr: configuredDevPort ? {
      host: "127.0.0.1",
      port: devPort,
      protocol: "ws"
    } : undefined,
    proxy: {
      "/api": apiTarget
    }
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    target: "esnext",
    cssTarget: "chrome120"
  }
});
