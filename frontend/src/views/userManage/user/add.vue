<template>
  <a-modal v-model:open="showbox" :title="modalTitle" width="500px" center>
    <a-form
      ref="ruleFormRef"
      style="max-width: 600px"
      :model="ruleForm"
      :rules="rules"
      laba-width="auto"
      class="demo-ruleForm"
      :size="formSize"
      status-icon
      :label-col="{ flex: '110px' }"
      :wrapper-col="{ flex: 1 }"
    >
      <a-form-item :label="t('user.username')" name="username" laba-position="top">
        <a-input v-model:value="ruleForm.username" />
      </a-form-item>
      <a-form-item :label="t('user.password')" name="password" laba-position="top">
        <a-input v-model:value="ruleForm.password" />
      </a-form-item>
      <a-form-item :label="t('user.role')" name="role_id" laba-position="top">
        <a-select v-model:value="ruleForm.role_id" :placeholder="t('user.selectRole')">
          <a-select-option v-for="role in roleList" :key="role.uuid" :value="role.uuid">
            {{ getRoleDisplayName(role) }}
          </a-select-option>
        </a-select>
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)">
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
import { addUser, editUser } from "@/api/user";
import { getRoleList } from "@/api/role";
import { useAppI18n } from "@/i18n";

interface RuleForm {
  uuid?: string;
  username: string;
  password: string;
  role_id: string;
}

const emit = defineEmits(["getList"]);
const { t } = useAppI18n();

const formSize = ref("default");
const ruleFormRef = ref();
const roleList = ref<any[]>([]);
const ruleForm = ref<RuleForm>({
  uuid: "",
  username: "",
  password: "",
  role_id: "",
});
const modalTitle = computed(() => (ruleForm.value.uuid ? t("user.editUser") : t("user.addUser")));

const rules = computed(() => ({
  username: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  password: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
}));
const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      if (ruleForm.value.uuid) {
        editUser(ruleForm.value).then(() => {
          message.success(t("common.success"));
          cancel();
          emit("getList");
        });
      } else {
        addUser(ruleForm.value).then(() => {
          message.success(t("common.success"));
          cancel();
          emit("getList");
        });
      }
    })
    .catch(() => {
      console.log("error submit!");
    });
};

const resetForm = () => {
  ruleFormRef.value.resetFields();
};

const cancel = () => {
  resetForm();
  showbox.value = false;
};

const showbox = ref(false);

function getRoleDisplayName(role: any): string {
  const key = `role.roleNames.${role.name}`;
  const translated = t(key);
  return translated !== key ? translated : role.name;
}

async function loadRoles() {
  try {
    const res = await getRoleList({});
    roleList.value = res.data || [];
  } catch {
    roleList.value = [];
  }
}

const init = (row?: RuleForm) => {
  loadRoles();
  if (row && row.uuid) {
    ruleForm.value = { ...row };
  } else {
    ruleForm.value = {} as RuleForm;
  }

  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less">
.demo-ruleForm :deep(.ant-form-item-label),
.demo-ruleForm :deep(.ant-form-item-label > label) {
  white-space: nowrap;
}
</style>
