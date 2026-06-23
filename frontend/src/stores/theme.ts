import { defineStore } from "pinia";

export type ThemeMode = "light" | "dark";

export default defineStore("theme", {
  persist: true,
  state: () => ({
    mode: "light" as ThemeMode,
  }),
  getters: {
    isDark: (state) => state.mode === "dark",
  },
  actions: {
    toggleTheme() {
      this.mode = this.mode === "light" ? "dark" : "light";
    },
    setTheme(mode: ThemeMode) {
      this.mode = mode;
    },
  },
});
