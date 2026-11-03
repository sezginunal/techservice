import { readFileSync } from "node:fs";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const version = readFileSync(new URL("../VERSION", import.meta.url), "utf8").trim();
if (!/^[0-9]+\.[0-9]+\.[0-9]+$/.test(version)) {
  throw new Error("VERSION must be MAJOR.MINOR.PATCH");
}

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(version),
  },
});
