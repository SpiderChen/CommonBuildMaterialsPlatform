import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 9245,
    strictPort: false,
    proxy: {
      "/api": "http://127.0.0.1:8088"
    }
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    target: "esnext",
    cssTarget: "chrome120"
  }
});
