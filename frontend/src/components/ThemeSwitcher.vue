<template>
  <div class="theme-switcher">
    <button class="theme-trigger" type="button" @click="toggle" :title="nextLabel">
      <!-- Sun icon (shown in dark mode — click to switch to light) -->
      <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="5"></circle>
        <line x1="12" y1="1" x2="12" y2="3"></line>
        <line x1="12" y1="21" x2="12" y2="23"></line>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
        <line x1="1" y1="12" x2="3" y2="12"></line>
        <line x1="21" y1="12" x2="23" y2="12"></line>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
      </svg>
      <!-- Moon icon (shown in light mode — click to switch to dark) -->
      <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import store from "@/stores";

const themeStore = store.useThemeStore();
const isDark = computed(() => themeStore.isDark);

const nextLabel = computed(() =>
  isDark.value ? "Switch to light theme" : "Switch to dark theme"
);

function toggle() {
  themeStore.toggleTheme();
}
</script>

<style scoped lang="less">
.theme-switcher {
  display: inline-flex;
  align-items: center;

  .theme-trigger {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 999px;
    border: 1px solid var(--color-border-light, #e5e7eb);
    background: var(--color-bg-card-secondary, #f8fafc);
    color: var(--color-text-primary, #344054);
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      border-color: var(--color-btn-hover-border, #cbd5e1);
      background: var(--color-btn-hover-bg, #eef2f7);
    }
  }
}
</style>
