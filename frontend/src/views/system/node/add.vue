<template>
  <a-modal v-model:open="showbox" :title="modalTitle" width="560px" center>
    <a-form
      ref="ruleFormRef"
      :model="ruleForm"
      :rules="rules"
      class="node-form"
      status-icon
      :label-col="{ flex: '110px' }"
      :wrapper-col="{ flex: 1 }"
    >
      <a-form-item :label="t('node.name')" name="name">
        <a-input v-model:value="ruleForm.name" :disabled="ruleForm.is_local" :placeholder="t('node.namePlaceholder')" />
      </a-form-item>
      <a-form-item :label="t('node.address')" name="address">
        <a-input v-model:value="ruleForm.address" :placeholder="t('node.addressPlaceholder')" />
      </a-form-item>
      <a-form-item :label="t('node.secretKey')" name="secret_key">
        <a-input-password v-model:value="ruleForm.secret_key" :placeholder="t('node.secretKeyPlaceholder')" />
      </a-form-item>
      <a-form-item :label="t('common.enabled')" name="enabled">
        <a-switch v-model:checked="ruleForm.enabled" :disabled="ruleForm.is_local" />
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)">{{ t("common.confirm") }}</a-button>
        <a-button @click="cancel">{{ t("common.cancel") }}</a-button>
      </div>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import { createNode, updateNode } from "@/api/node";
import { useAppI18n } from "@/i18n";
import { message } from "ant-design-vue";
import { computed, ref } from "vue";

interface RuleForm {
  uuid: string;
  name: string;
  address: string;
  secret_key: string;
  enabled: boolean;
  is_local: boolean;
}

const emit = defineEmits(["getList"]);
const { t } = useAppI18n();
const ruleFormRef = ref();
const showbox = ref(false);

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
  address: "",
  secret_key: "",
  enabled: true,
  is_local: false,
});

const ruleForm = ref<RuleForm>(createForm());
const modalTitle = computed(() => (ruleForm.value.uuid ? t("node.editNode") : t("node.addNode")));

const rules = computed(() => ({
  name: [{ required: true, message: t("node.nameRequired"), trigger: "blur" }],
  address: [{ required: true, message: t("node.addressRequired"), trigger: "blur" }],
  secret_key: [{ required: true, message: t("node.secretKeyRequired"), trigger: "blur" }],
}));

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      const request = ruleForm.value.uuid
        ? updateNode({
            uuid: ruleForm.value.uuid,
            name: ruleForm.value.name,
            address: ruleForm.value.address,
            secret_key: ruleForm.value.secret_key,
            enabled: ruleForm.value.enabled,
          })
        : createNode({
            name: ruleForm.value.name,
            address: ruleForm.value.address,
            secret_key: ruleForm.value.secret_key,
            enabled: ruleForm.value.enabled,
          });
      request.then(() => {
        message.success(t("common.success"));
        cancel();
        emit("getList");
      });
    })
    .catch(() => {
      console.log("error submit!");
    });
};

const resetForm = () => {
  ruleFormRef.value?.resetFields();
  ruleForm.value = createForm();
};

const cancel = () => {
  resetForm();
  showbox.value = false;
};

const init = (row?: RuleForm) => {
  ruleForm.value = row ? { ...createForm(), ...row } : createForm();
  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less">
.node-form :deep(.ant-form-item-label),
.node-form :deep(.ant-form-item-label > label) {
  white-space: nowrap;
}
</style>
