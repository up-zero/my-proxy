<template>
  <a-modal v-model:open="showbox" :title="t('password.title')" width="500px" center>
    <a-form ref="ruleFormRef" style="max-width: 600px" :model="ruleForm" :rules="rules" laba-width="auto"
      class="demo-ruleForm" :size="formSize" status-icon :label-col="{ flex: '118px' }" :wrapper-col="{ flex: 1 }">
      <a-form-item :label="t('password.oldPassword')" name="old_password" laba-position="top">
        <a-input-password v-model:value="ruleForm.old_password" />
      </a-form-item>

      <a-form-item :label="t('password.newPassword')" name="new_password">
        <a-input-password v-model:value="ruleForm.new_password" />
      </a-form-item>
      <a-form-item :label="t('password.confirmPassword')" name="re_password">
        <a-input-password v-model:value="ruleForm.re_password" />
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

import { changePassword } from "@/api/user";
import { useAppI18n } from "@/i18n";

interface RuleForm {
  old_password: any;
  new_password: any;
  re_password: any;
}

const { t } = useAppI18n();

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>({
  old_password: "",
  new_password: "",
  re_password: "",
});

const rules = computed(() => ({
  old_password: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  new_password: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  re_password: [
    {
      validator: (r: any, value: string) => {
        if (value === "") {
          return Promise.reject(t("password.pleaseReenterPassword"));
        } else {
          if (value !== ruleForm.value.new_password) {
            return Promise.reject(t("password.passwordMismatch"));
          }
          return Promise.resolve();
        }
      },
      trigger: "blur",
    },
  ],
}));
const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      changePassword(ruleForm.value).then(() => {
        message.success(t("common.success"));
        cancel();
      });
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

const init = () => {

  ruleForm.value = {} as RuleForm;


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
