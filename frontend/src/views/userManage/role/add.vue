<template>
  <a-modal v-model:open="showbox" :title="modalTitle" width="640px" center :maskClosable="false">
    <a-form
      ref="ruleFormRef"
      :model="ruleForm"
      :rules="rules"
      class="role-form"
      :label-col="{ flex: '100px' }"
      :wrapper-col="{ flex: 1 }"
    >
      <a-form-item :label="t('role.roleName')" name="name">
        <a-input v-model:value="ruleForm.name" :placeholder="t('role.inputRoleName')" :disabled="isAdminRole" />
      </a-form-item>
      <a-form-item :label="t('role.description')" name="description">
        <a-input v-model:value="ruleForm.description" :placeholder="t('role.inputDescription')" />
      </a-form-item>
      <a-form-item :label="t('role.permissions')" name="permissions">
        <div class="perm-groups">
          <div v-for="group in permGroups" :key="group.groupKey" class="perm-group">
            <div class="perm-group-header">
              <a-checkbox
                :checked="isGroupAllChecked(group.perms.map(p => p.key))"
                :indeterminate="isGroupIndeterminate(group.perms.map(p => p.key))"
                @change="(e: any) => onGroupCheckAll(group.perms.map(p => p.key), e.target.checked)"
                :disabled="isAdminRole"
              >
                <strong>{{ getGroupDisplayName(group.groupName) }}</strong>
              </a-checkbox>
            </div>
            <div class="perm-group-items">
              <a-checkbox
                v-for="permItem in group.perms"
                :key="permItem.key"
                :checked="ruleForm.permissions.includes(permItem.key)"
                @change="(e: any) => onPermChange(permItem.key, e.target.checked)"
                :disabled="isAdminRole"
              >
                {{ getPermDisplayName(permItem) }}
              </a-checkbox>
            </div>
          </div>
        </div>
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)" :disabled="isAdminRole">
          {{ t("common.confirm") }}
        </a-button>
        <a-button @click="cancel">{{ t("common.cancel") }}</a-button>
      </div>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import { message } from "ant-design-vue";
import { computed, ref } from "vue";
import { addRole, editRole } from "@/api/role";
import { useAppI18n } from "@/i18n";

interface RoleForm {
  uuid?: string;
  name: string;
  description: string;
  permissions: string[];
  built_in?: boolean;
}

const emit = defineEmits(["get-list"]);
const { t } = useAppI18n();

const ruleFormRef = ref();
const ruleForm = ref<RoleForm>({
  name: "",
  description: "",
  permissions: [],
});

// 权限分组定义（按菜单结构）
const permGroups = [
  { 
    groupKey: "dashboard", 
    groupName: "routes.dashboard",
    perms: [
      { key: "dashboard.view", nameKey: "role.permItems.dashboard" },
    ]
  },
  { 
    groupKey: "proxyManage", 
    groupName: "routes.proxyManage",
    perms: [
      { key: "proxy.view", nameKey: "role.permItems.proxyList" },
      { key: "tag.manage", nameKey: "role.permItems.tagManage" },
      { key: "traffic_policy.manage", nameKey: "role.permItems.trafficPolicy" },
    ]
  },
  { 
    groupKey: "operation", 
    groupName: "routes.operationCenter",
    perms: [
      { key: "alert.view", nameKey: "role.permItems.alertNotify" },
      { key: "audit.view", nameKey: "role.permItems.auditLog" },
      { key: "terminal.view", nameKey: "role.permItems.terminal" },
    ]
  },
  { 
    groupKey: "userManage", 
    groupName: "routes.userManage",
    perms: [
      { key: "user.manage", nameKey: "role.permItems.userList" },
      { key: "role.manage", nameKey: "role.permItems.permPolicy" },
    ]
  },
];

function getPermDisplayName(permItem: { key: string; nameKey: string }): string {
  return t(permItem.nameKey);
}

function getGroupDisplayName(groupName: string): string {
  return t(groupName);
}

const isAdminRole = computed(() => ruleForm.value.built_in === true && ruleForm.value.name === "admin");
const modalTitle = computed(() => (ruleForm.value.uuid ? t("role.editRole") : t("role.addRole")));

const rules = computed(() => ({
  name: [{ required: true, message: t("role.inputRoleName"), trigger: "blur" }],
}));

function isGroupAllChecked(perms: string[]): boolean {
  return perms.every((p) => ruleForm.value.permissions.includes(p));
}

function isGroupIndeterminate(perms: string[]): boolean {
  const checkedCount = perms.filter((p) => ruleForm.value.permissions.includes(p)).length;
  return checkedCount > 0 && checkedCount < perms.length;
}

function onGroupCheckAll(perms: string[], checked: boolean) {
  const newPerms = [...ruleForm.value.permissions];
  for (const p of perms) {
    const idx = newPerms.indexOf(p);
    if (checked && idx === -1) {
      newPerms.push(p);
    } else if (!checked && idx !== -1) {
      newPerms.splice(idx, 1);
    }
  }
  ruleForm.value.permissions = newPerms;
}

function onPermChange(perm: string, checked: boolean) {
  const newPerms = [...ruleForm.value.permissions];
  const idx = newPerms.indexOf(perm);
  if (checked && idx === -1) {
    newPerms.push(perm);
  } else if (!checked && idx !== -1) {
    newPerms.splice(idx, 1);
  }
  ruleForm.value.permissions = newPerms;
}

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl.validate().then(() => {
    const payload = {
      ...ruleForm.value,
    };
    if (ruleForm.value.uuid) {
      editRole(payload).then(() => {
        message.success(t("common.success"));
        cancel();
        emit("get-list");
      });
    } else {
      addRole(payload).then(() => {
        message.success(t("common.success"));
        cancel();
        emit("get-list");
      });
    }
  }).catch(() => {
    console.log("error submit!");
  });
};

const resetForm = () => {
  ruleFormRef.value?.resetFields();
};

const cancel = () => {
  resetForm();
  showbox.value = false;
};

const showbox = ref(false);

const init = (row?: RoleForm) => {
  if (row && row.uuid) {
    // 编辑模式
    let perms: string[] = [];
    if (typeof (row as any).permissions === "string") {
      try {
        perms = JSON.parse((row as any).permissions);
      } catch {
        perms = [];
      }
    } else if (Array.isArray(row.permissions)) {
      perms = row.permissions;
    }
    // 过滤掉已失效的权限
    const validPerms = permGroups.flatMap(g => g.perms.map(p => p.key));
    perms = perms.filter(p => validPerms.includes(p));
    ruleForm.value = {
      uuid: row.uuid,
      name: row.name,
      description: row.description,
      permissions: perms,
      built_in: row.built_in,
    };
  } else {
    ruleForm.value = {
      name: "",
      description: "",
      permissions: [],
    };
  }
  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less">
.role-form :deep(.ant-form-item-label),
.role-form :deep(.ant-form-item-label > label) {
  white-space: nowrap;
}

.perm-groups {
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 12px;
  max-height: 400px;
  overflow-y: auto;
  background: #fafafa;
}

.perm-group {
  margin-bottom: 12px;
  &:last-child {
    margin-bottom: 0;
  }
}

.perm-group-header {
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1px solid #f0f0f0;
}

.perm-group-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  padding-left: 24px;
}
</style>
