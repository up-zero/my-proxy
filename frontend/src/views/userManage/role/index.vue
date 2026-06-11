<template>
  <div class="p-page">
    <div class="page-title-row">
      <div class="page-title">
        <h1>{{ t("routes.permPolicy") }}</h1>
        <p>{{ t("role.permPageDesc") }}</p>
      </div>
      <a-button type="primary" @click="toAddPage">{{ t("role.addRole") }}</a-button>
    </div>

    <!-- 角色卡片网格 -->
    <div class="role-grid">
      <div v-for="item in state.list" :key="item.uuid" class="role-card" :class="{ 'is-admin': item.name === 'admin' }">
        <div class="role-card-header">
          <div class="role-name-row">
            <span class="role-name">{{ getRoleDisplayName(item.name) }}</span>
            <a-tag v-if="item.built_in" color="blue">{{ t("role.builtIn") }}</a-tag>
            <a-tag v-else color="green">{{ t("role.custom") }}</a-tag>
          </div>
          <div class="role-actions" v-if="!(item.built_in && item.name === 'admin')">
            <a-button type="link" size="small" @click="editItem(item)">{{ t("common.edit") }}</a-button>
            <a-popconfirm
              v-if="!item.built_in"
              :title="t('role.deleteConfirm')"
              @confirm="delItem(item)"
            >
              <a-button type="link" danger size="small">{{ t("common.delete") }}</a-button>
            </a-popconfirm>
          </div>
        </div>
        <div class="role-desc" v-if="item.description">{{ item.description }}</div>
        <div class="role-perms">
          <a-tag
            v-for="perm in getPermissionList(item)"
            :key="perm"
            size="small"
            class="perm-tag"
          >
            {{ getPermDisplayName(perm) }}
          </a-tag>
          <a-tag v-if="getPermissionList(item).length === 0" color="default">{{ t("common.none") }}</a-tag>
        </div>
      </div>
    </div>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { getRoleList, delRole } from "@/api/role";
import addBox from "./add.vue";
import { message } from "ant-design-vue";
import { onMounted, reactive, ref } from "vue";
import { useAppI18n } from "@/i18n";

interface RoleItem {
  uuid: string;
  name: string;
  description: string;
  built_in: boolean;
  permissions: string;
  created_at: number;
  updated_at: number;
}

const addBoxRef = ref();
const { t } = useAppI18n();
const state = reactive({
  isLoading: false,
  list: [] as RoleItem[],
});

onMounted(() => {
  getList();
});

function getRoleDisplayName(name: string): string {
  const key = `role.roleNames.${name}`;
  const translated = t(key);
  return translated !== key ? translated : name;
}

function getPermissionList(item: RoleItem): string[] {
  if (!item.permissions) return [];
  try {
    return JSON.parse(item.permissions);
  } catch {
    return [];
  }
}

function getPermDisplayName(perm: string): string {
  // 权限名称映射表
  const permNames: Record<string, string> = {
    "dashboard.view": t("role.permItems.dashboard"),
    "proxy.view": t("role.permItems.proxyList"),
    "tag.manage": t("role.permItems.tagManage"),
    "traffic_policy.manage": t("role.permItems.trafficPolicy"),
    "alert.view": t("role.permItems.alertNotify"),
    "audit.view": t("role.permItems.auditLog"),
    "terminal.view": t("role.permItems.terminal"),
    "user.manage": t("role.permItems.userList"),
    "role.manage": t("role.permItems.permPolicy"),
  };
  return permNames[perm] || perm;
}

async function getList() {
  try {
    state.isLoading = true;
    const res = await getRoleList({});
    if (!res.data) return;
    state.list = res.data || [];
  } finally {
    state.isLoading = false;
  }
}

function toAddPage() {
  addBoxRef.value.init();
}

function editItem(row: RoleItem) {
  addBoxRef.value.init(row);
}

const delItem = (row: RoleItem) => {
  delRole({ uuid: [row.uuid] }).then(() => {
    getList();
    message.success({ content: t("common.success") });
  });
};
</script>

<style lang="less" scoped>
.page-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.page-title {
  h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #101828;
  }
  p {
    margin: 6px 0 0;
    color: #98a2b3;
    font-size: 13px;
  }
}

.role-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.role-card {
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 16px 20px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  }

  &.is-admin {
    border-color: #1677ff;
    background: linear-gradient(135deg, #f0f7ff, #fff);
  }
}

.role-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.role-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-name {
  font-size: 16px;
  font-weight: 600;
  color: #101828;
}

.role-desc {
  color: #667085;
  font-size: 13px;
  margin-bottom: 12px;
}

.role-perms {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.perm-tag {
  font-size: 12px;
}
</style>
