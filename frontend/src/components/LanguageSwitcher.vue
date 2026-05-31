<template>
  <div :class="['language-switcher', `language-switcher--${variant}`]">
    <template v-if="variant === 'compact'">
      <a-dropdown trigger="click">
        <button class="compact-trigger" type="button">
          <GlobalOutlined />
          <span>{{ currentShortLabel }}</span>
        </button>
        <template #overlay>
          <a-menu @click="handleMenuClick">
            <a-menu-item v-for="item in languageOptions" :key="item.value">
              <div class="compact-option">
                <span>{{ item.label }}</span>
                <CheckOutlined v-if="item.value === locale" />
              </div>
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </template>
    <template v-else>
      <span v-if="showLabel" class="language-label">{{ t("language.switchLabel") }}</span>
      <a-select :value="locale" size="small" style="width: 132px" @change="handleChange">
        <a-select-option v-for="item in languageOptions" :key="item.value" :value="item.value">
          {{ item.label }}
        </a-select-option>
      </a-select>
    </template>
  </div>
</template>

<script setup lang="ts">
import { CheckOutlined, GlobalOutlined } from "@ant-design/icons-vue";
import { useAppI18n, type AppLocale } from "@/i18n";
import { computed } from "vue";

withDefaults(
  defineProps<{
    showLabel?: boolean;
    variant?: "default" | "compact";
  }>(),
  {
    showLabel: false,
    variant: "default",
  }
);

const { locale, languageOptions, setLocale, t } = useAppI18n();
const currentShortLabel = computed(() => (locale.value === "zh-CN" ? "ZH" : "EN"));

function handleChange(value: AppLocale) {
  setLocale(value);
}

function handleMenuClick({ key }: { key: AppLocale }) {
  setLocale(key);
}
</script>

<style scoped lang="less">
.language-switcher {
  display: inline-flex;
  align-items: center;
  gap: 8px;

  &--compact {
    .compact-trigger {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      height: 34px;
      padding: 0 12px;
      border-radius: 999px;
      border: 1px solid #e5e7eb;
      background: #f8fafc;
      color: #344054;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.2s ease;

      &:hover {
        border-color: #cbd5e1;
        background: #eef2f7;
      }
    }

    .compact-option {
      min-width: 108px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
  }
}

.language-label {
  color: #667085;
  font-size: 12px;
}
</style>

