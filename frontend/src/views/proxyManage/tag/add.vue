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
      :label-col="{ span: 4 }"
    >
      <a-form-item label="标签名称" name="name">
        <a-input v-model:value="ruleForm.name" />
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)">确定</a-button>
        <a-button @click="cancel">取消</a-button>
      </div>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import { addTag, editTag } from "@/api/tag";
import { message } from "ant-design-vue";
import { computed, reactive, ref } from "vue";

interface RuleForm {
  uuid: string;
  name: string;
}

const emit = defineEmits(["getList"]);
const formSize = ref("default");
const ruleFormRef = ref();
const showbox = ref(false);

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
});

const ruleForm = ref<RuleForm>(createForm());
const modalTitle = computed(() => (ruleForm.value.uuid ? "编辑标签" : "新增标签"));

const rules = reactive({
  name: [{ required: true, message: "请输入标签名称", trigger: "blur" }],
});

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      const request = ruleForm.value.uuid ? editTag(ruleForm.value) : addTag(ruleForm.value);
      request.then(() => {
        message.success("操作成功");
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

<style scoped lang="less"></style>

