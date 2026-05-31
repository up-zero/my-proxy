<template>
  <a-modal v-model:open="showbox" :title="modalTitle" width="500px" center>
    <a-form
      ref="ruleFormRef"
      style="max-width: 600px"
      :model="ruleForm"
      :rules="rules"
      class="demo-ruleForm"
      :size="formSize"
      status-icon
      :label-col="{ flex: '110px' }"
      :wrapper-col="{ flex: 1 }"
    >
      <a-form-item :label="t('tag.tagName')" name="name">
        <a-input v-model:value="ruleForm.name" />
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
import { addTag, editTag } from "@/api/tag";
import { useAppI18n } from "@/i18n";
import { message } from "ant-design-vue";
import { computed, ref } from "vue";

interface RuleForm {
  uuid: string;
  name: string;
}

const emit = defineEmits(["getList"]);
const { t } = useAppI18n();
const formSize = ref("default");
const ruleFormRef = ref();
const showbox = ref(false);

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
});

const ruleForm = ref<RuleForm>(createForm());
const modalTitle = computed(() => (ruleForm.value.uuid ? t("tag.editTag") : t("tag.addTag")));

const rules = computed(() => ({
  name: [{ required: true, message: t("tag.inputTagName"), trigger: "blur" }],
}));

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      const request = ruleForm.value.uuid ? editTag(ruleForm.value) : addTag(ruleForm.value);
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
.demo-ruleForm :deep(.ant-form-item-label),
.demo-ruleForm :deep(.ant-form-item-label > label) {
  white-space: nowrap;
}
</style>

