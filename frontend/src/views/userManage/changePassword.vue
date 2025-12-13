<template>
  <a-modal v-model:open="showbox" title="修改密码" width="500px" center>
    <a-form ref="ruleFormRef" style="max-width: 600px" :model="ruleForm" :rules="rules" laba-width="auto"
      class="demo-ruleForm" :size="formSize" status-icon :label-col="{ span: 4 }">
      <a-form-item label="旧密码" name="old_password" laba-position="top">
        <a-input-password v-model:value="ruleForm.old_password" />
      </a-form-item>

      <a-form-item label="新密码" name="new_password">
        <a-input-password v-model:value="ruleForm.new_password" />
      </a-form-item>
      <a-form-item label="确认密码" name="re_password">
        <a-input-password v-model:value="ruleForm.re_password" />
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

import { changePassword } from "@/api/user";

interface RuleForm {
  old_password: any;
  new_password: any;
  re_password: any;
}

const emit = defineEmits(["getList"]);

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>({
  old_password: "",
  new_password: "",
  re_password: "",
});

const rules = reactive({
  old_password: [{ required: true, message: "请输入", trigger: "blur" }],
  new_password: [{ required: true, message: "请输入", trigger: "blur" }],
  re_password: [
    {
      validator: (r: any, value: string) => {
        if (value === "") {
          return Promise.reject("请再次输入密码");
        } else {
          if (value !== ruleForm.value.new_password) {
            return Promise.reject("两次密码不一致");
          }
          return Promise.resolve();
        }
      },
      trigger: "blur",
    },
  ],
});
const props = defineProps(["comfirmApi"]);
const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      changePassword(ruleForm.value).then(() => {
        message.success("操作成功");
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

<style scoped lang="less"></style>
