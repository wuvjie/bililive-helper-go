import { defineConfig, loadEnv } from "vite";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";
import { existsSync, rmSync } from "fs";
import vue from "@vitejs/plugin-vue";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";

const __dirname = dirname(fileURLToPath(import.meta.url));

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());

  return {
    base: "/",
    resolve: {
      alias: {
        "@": resolve(__dirname, "src")
      }
    },
    plugins: [
      vue(),
      AutoImport({
        resolvers: [ElementPlusResolver()],
        imports: ["vue", "vue-router", "pinia"],
        dts: "src/auto-imports.d.ts"
      }),
      Components({
        resolvers: [ElementPlusResolver()],
        dts: "src/components.d.ts"
      }),
      {
        name: "clean-stale-assets",
        buildStart() {
          const assetsDir = resolve(__dirname, "../templates/assets");
          try {
            if (existsSync(assetsDir)) {
              rmSync(assetsDir, { recursive: true });
            }
          } catch {}
        }
      }
    ],
    server: {
      port: 3000,
      proxy: {
        "/api": { target: "http://localhost:5000", changeOrigin: true },
        "/login": { target: "http://localhost:5000", changeOrigin: true },
        "/logout": { target: "http://localhost:5000", changeOrigin: true }
      }
    },
    build: {
      outDir: "../templates",
      emptyOutDir: false,
      rolldownOptions: {
        output: {
          chunkFileNames: "assets/[name]-[hash].js",
          entryFileNames: "assets/[name]-[hash].js",
          assetFileNames: "assets/[name]-[hash].[ext]"
        }
      }
    }
  };
});
