import inertia from "@inertiajs/vite";
import { defineConfig } from "vite";
import laravel from "laravel-vite-plugin";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    laravel({
      input: ["resources/js/app.jsx", "resources/css/app.css"],
      publicDirectory: "public",
      buildDirectory: "build",
      refresh: true,
    }),
    inertia(),
    tailwindcss(),
    react(),
  ],
});
