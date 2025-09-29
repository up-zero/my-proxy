<template>
  <div class="login-page">
    <div class="login-box">
      <img src="@/assets/logo.png" alt="" class="logo" />

      <a-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="rules"
        :label-col="{span:0}"
        style="text-align: left;"
      >
        <a-form-item name="username">
          <a-input v-model:value="ruleForm.username" placeholder="请输入账号" size="large">
            <template #prefix>
              <user-outlined type="user" />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item name="password">
          <a-input-password
          placeholder="请输入密码"
            v-model:value="ruleForm.password"
            size="large"
            type="password"
          >
            <template #prefix>
              <KeyOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <div>
          <a-button type="primary" @click="handleLogin" class="sub-button">
            登录
          </a-button>
        </div>
      </a-form>
      <p class="copyright">©up-zero</p>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { UserOutlined, KeyOutlined } from "@ant-design/icons-vue";

import { reactive, ref } from "vue";
import store from "@/stores";

interface RuleForm {
  username: string;
  password: string;
}
const loading = ref(false);

const ruleFormRef = ref();
const user = store.useUserStore();
const rules = reactive({
  username: [{ required: true, message: "请输入账号", trigger: "blur" }],
  password: [
    {
      required: true,
      message: "请输入密码",
      trigger: "change",
    },
  ],
});
const ruleForm = reactive<RuleForm>({
  username: "",
  password: "",
});

// 登录
async function handleLogin() {
  if (!ruleFormRef.value) return;
  await ruleFormRef.value
    .validate()
    .then(async () => {
      loading.value = true;
      await user.login({
        username: ruleForm.username,
        password: ruleForm.password,
      });
      loading.value = false;
    })
    .catch((err: any) => {
      console.log("error submit!", err);
    });
}
</script>
<style>
html,
body,
#app {
  width: 100%;
  height: 100%;
}
</style>
<style lang="less" scoped>
.login-page {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: linear-gradient(to top, #87ceeb 0%, #b0e0ff 60%, #e0f7ff 100%);
  .login-box {
    width: 500px;
    max-width: 90%;
    padding: 30px 50px;
    border-radius: 4px;
    background-color: white;
    box-shadow: 0 0 10px #a3bded;
    text-align: center;
    position: relative;
    .copyright {
      position: absolute;
      bottom: -40px;
      left: 50%;
      transform: translateX(-50%);
      white-space: nowrap;
      color: #bbb;
    }
    .logo {
      width: 100px;
      margin-bottom: 30px;
    }
    h3 {
      padding: 20px 0 40px 0;
    }
    .code-img {
      background-color: #f5f5f5;
      width: 100%;
      height: 40px;
      cursor: pointer;
      vertical-align: middle;
    }
    .sub-button {
      width: 100%;
      height: 40px;
      margin-top: 30px;
    }
  }
}
</style>
