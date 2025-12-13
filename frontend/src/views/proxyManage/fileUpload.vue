<template>
  <a-upload
    :show-upload-list="false"
    :before-upload="beforeUpload"
    :custom-request="customRequest"
  >
    <a-button type="primary" ghost style="margin-right: 10px;">导入</a-button>
  </a-upload>
</template>

<script setup lang="ts">
import { importProxy } from "@/api/proxy";
import { message } from "ant-design-vue";
 

const emit = defineEmits(["success"]);

const beforeUpload = () => {
  return true;
};

const customRequest = async (options: { file: any; onSuccess: any; onError: any; }) => {
  const { file, onSuccess, onError } = options;

  const formData = new FormData();
  formData.append("file", file);

  try {
    const res = await importProxy(formData);

    message.success("上传成功");
    onSuccess(res);
    emit("success"); // 通知父组件刷新列表
  } catch (e) {
    message.error("上传失败");
    onError(e);
  }
};
</script>
