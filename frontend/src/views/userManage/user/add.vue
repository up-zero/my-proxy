<template>
  <a-modal v-model:open="showbox" title="添加代理" width="500px" center>
    <a-form
      ref="ruleFormRef"
      style="max-width: 600px"
      :model="ruleForm"
      :rules="rules"
      laba-width="auto"
      class="demo-ruleForm"
      :size="formSize"
      status-icon
      :label-col="{ span: 4 }"
    >
      <a-form-item   label="用户名" name="username" laba-position="top">
        <a-input v-model:value="ruleForm.username" />
      </a-form-item>
        <a-form-item   label="密码" name="password" laba-position="top">
        <a-input v-model:value="ruleForm.password" />
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)">
          确定
        </a-button>
        <a-button @click="cancel">取消</a-button>
      </div>
    </template>
  </a-modal>
</template>
<script lang="ts" setup>
import { message } from "ant-design-vue";
import { ref, reactive } from "vue";
import { addUser, editUser } from "@/api/user";

interface RuleForm {
  uuid?: string;
  username: string;
  password:string
}

const emit = defineEmits(["getList"]);

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>({
  uuid: "",
  username: "",
  password:""
  
});

const rules = reactive({
  username: [{ required: true, message: "请输入", trigger: "blur" }],
  password: [{ required: true, message: "请输入", trigger: "blur" }],
  
});
const props = defineProps(["comfirmApi"]);
const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      if (ruleForm.value.uuid) {
        editUser(ruleForm.value).then(() => {
          message.success("操作成功");
          cancel();
          emit("getList");
        });
      } else {
        addUser(ruleForm.value).then(() => {
          message.success("操作成功");
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

const init = (row: RuleForm) => {
  if (row) {
    ruleForm.value = { ...row };
  } else {
    ruleForm.value = {} as RuleForm;
  }

  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less"></style>
