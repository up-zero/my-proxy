<template>
  <a-upload
    :show-upload-list="false"
    :before-upload="beforeUpload"
    :custom-request="customRequest"
  >
    <a-button type="primary" ghost style="margin-right: 10px;">{{ t("common.import") }}</a-button>
  </a-upload>
</template>

<script setup lang="ts">
import { importProxy } from "@/api/proxy";
import { useAppI18n } from "@/i18n";
import { message } from "ant-design-vue";
 

const emit = defineEmits(["success"]);
const { t } = useAppI18n();

const beforeUpload = () => {
  return true;
};

const customRequest = async (options: { file: any; onSuccess: any; onError: any; }) => {
  const { file, onSuccess, onError } = options;

  const formData = new FormData();
  formData.append("file", file);

  try {
    const res = await importProxy(formData);

    message.success(t("proxy.importSuccess"));
    onSuccess(res);
    emit("success"); // 通知父组件刷新列表
  } catch (e) {
    message.error(t("proxy.importFailed"));
    onError(e);
  }
};
</script>
