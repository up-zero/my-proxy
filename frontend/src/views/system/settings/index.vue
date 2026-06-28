<template>
  <div class="settings-page">
    <a-spin :spinning="state.isLoading">
      <div class="settings-list">
        <div v-for="item in state.items" :key="item.key" class="settings-item">
          <div class="item-label">
            <span class="label-text">{{ getItemLabel(item.key) }}</span>
            <a-tag v-if="isDefault(item)" color="blue" size="small" class="default-tag">
              {{ t("settings.defaultLabel") }}
            </a-tag>
          </div>
          <div class="item-desc">{{ getItemDesc(item.key) }}</div>
          <div class="item-control">
            <!-- JWT 密钥：文本输入 + 复制图标 -->
            <template v-if="isJwtItem(item.key)">
              <a-input-password
                v-model:value="editTextValues[item.key]"
                style="width: 360px"
              />
              <a-button type="text" size="small" @click="copyToClipboard(editTextValues[item.key])">
                <template #icon><copy-outlined /></template>
              </a-button>
            </template>
            <!-- 服务端口 -->
            <template v-else-if="isPortItem(item.key)">
              <a-input-number
                v-model:value="editValues[item.key]"
                :min="1"
                :max="65535"
                :precision="0"
                style="width: 160px"
              />
              <a-button
                v-if="!isDefault(item)"
                type="link"
                size="small"
                @click="resetItem(item)"
              >
                {{ t("settings.resetDefault") }}
              </a-button>
            </template>
            <!-- 天数 -->
            <template v-else>
              <a-input-number
                v-model:value="editValues[item.key]"
                :min="1"
                :max="3650"
                :precision="0"
                style="width: 160px"
                :addon-after="t('settings.unit.days')"
              />
              <a-button
                v-if="!isDefault(item)"
                type="link"
                size="small"
                @click="resetItem(item)"
              >
                {{ t("settings.resetDefault") }}
              </a-button>
            </template>
          </div>
        </div>
      </div>

      <div class="settings-footer">
        <a-button type="primary" @click="handleSave" :loading="state.saving">
          {{ t("settings.save") }}
        </a-button>
      </div>
    </a-spin>
  </div>
</template>

<script lang="ts" setup>
import { getSystemSettings, updateSystemSettings, type ConfigItem } from "@/api/config";
import { useAppI18n } from "@/i18n";
import { message } from "ant-design-vue";
import { CopyOutlined } from "@ant-design/icons-vue";
import { onMounted, reactive } from "vue";

const { t } = useAppI18n();

// 需要特殊处理的配置项 key
const PORT_CONFIG_KEY = "SERVER_PORT_KEY";
const JWT_CONFIG_KEY = "JWT_SECRET_KEY";

const state = reactive({
  isLoading: false,
  saving: false,
  items: [] as ConfigItem[],
});

// 编辑中的数值（端口、天数）
const editValues: Record<string, number> = reactive({});
// 编辑中的文本值（JWT 密钥等）
const editTextValues: Record<string, string> = reactive({});

onMounted(() => {
  loadSettings();
});

async function loadSettings() {
  try {
    state.isLoading = true;
    const res = await getSystemSettings();
    const items: ConfigItem[] = res.data?.items || [];
    state.items = items;
    // 初始化编辑值
    for (const item of items) {
      if (item.key === JWT_CONFIG_KEY) {
        editTextValues[item.key] = item.value || "";
      } else {
        let valueStr = item.value;
        if (item.key === PORT_CONFIG_KEY) {
          while (valueStr.startsWith(':')) {
            valueStr = valueStr.substring(1);
          }
        }
        editValues[item.key] = parseInt(valueStr) || parseInt(item.default_value) || 0;
      }
    }
  } finally {
    state.isLoading = false;
  }
}

function getItemLabel(key: string): string {
  return t(`settings.items.${key}`);
}

function getItemDesc(key: string): string {
  return t(`settings.desc.${key}`);
}

function isDefault(item: ConfigItem): boolean {
  if (item.key === JWT_CONFIG_KEY) return false;
  const currentVal = editValues[item.key];
  const defaultVal = parseInt(item.default_value);
  return currentVal === defaultVal;
}

function resetItem(item: ConfigItem) {
  editValues[item.key] = parseInt(item.default_value) || 0;
}

function isPortItem(key: string): boolean {
  return key === PORT_CONFIG_KEY;
}

function isJwtItem(key: string): boolean {
  return key === JWT_CONFIG_KEY;
}

function copyToClipboard(text: string) {
  if (!text) return;
  navigator.clipboard.writeText(text).then(() => {
    message.success(t("settings.copySuccess"));
  }).catch(() => {
    message.error(t("settings.copyFail"));
  });
}

async function handleSave() {
  try {
    state.saving = true;
    const items = state.items.map((item) => {
      if (item.key === JWT_CONFIG_KEY) {
        return { key: item.key, value: editTextValues[item.key] ?? item.value };
      }
      return { key: item.key, value: String(editValues[item.key] ?? item.value) };
    });
    await updateSystemSettings(items);
    // JWT 密钥修改后需要重启服务才能生效
    message.success(t("settings.saveSuccess"));
    await loadSettings();
  } finally {
    state.saving = false;
  }
}
</script>

<style scoped lang="less">
.settings-page {
  width: 100%;
}

.settings-header {
  margin-bottom: 24px;
  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text-primary, #101828);
  }
}

.settings-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.settings-item {
  padding: 16px;
  border: 1px solid var(--color-border, #f0f0f0);
  border-radius: 8px;
  background: var(--color-bg-card-secondary, #fafafa);
  transition: border-color 0.2s;

  &:hover {
    border-color: var(--color-item-hover-border, #d9d9d9);
  }

  .item-label {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;

    .label-text {
      font-weight: 500;
      font-size: 14px;
      color: var(--color-text-primary, #101828);
    }

    .default-tag {
      font-size: 11px;
    }
  }

  .item-desc {
    font-size: 12px;
    color: var(--color-text-muted, #888);
    margin-bottom: 12px;
  }

  .item-control {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.settings-footer {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border, #f0f0f0);
}
</style>
