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
      :label-col="{ span: 4 }"
    >
      <a-form-item ref="name" label="名称" name="name" laba-position="top">
        <a-input v-model:value="ruleForm.name" />
      </a-form-item>
      <a-form-item label="标签" name="tag_uuid_list">
        <a-select v-model:value="ruleForm.tag_uuid_list" mode="multiple" allow-clear placeholder="请选择标签">
          <a-select-option v-for="item in tagList" :key="item.uuid" :value="item.uuid">
            {{ item.name }}
          </a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="type" label="类型" name="type">
        <a-select v-model:value="ruleForm.type" placeholder="请选择代理类型">
          <a-select-option value="TCP">TCP</a-select-option>
          <a-select-option value="UDP">UDP</a-select-option>
          <a-select-option value="HTTP">HTTP</a-select-option>
          <a-select-option value="SOCKS5">SOCKS5</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="listen_address" label="监听地址" name="listen_address">
        <a-input v-model:value="ruleForm.listen_address" placeholder="default 0.0.0.0"/>
      </a-form-item>
      <a-form-item ref="listen_port" label="监听端口" name="listen_port">
        <a-input v-model:value="ruleForm.listen_port" />
      </a-form-item>
      <a-form-item v-if="!isSocks5Type" ref="target_address" label="目标地址" name="target_address">
        <a-input v-model:value="ruleForm.target_address" />
      </a-form-item>
      <a-form-item v-if="!isSocks5Type" ref="target_port" label="目标端口" name="target_port">
        <a-input v-model:value="ruleForm.target_port" />
      </a-form-item>
      <a-form-item v-else label="说明">
        <span class="form-tip">SOCKS5 为动态代理，无需配置目标地址和目标端口。</span>
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
import { computed, ref, watch } from "vue";
import { addProxy, editProxy } from "@/api/proxy";
import { getTagList } from "@/api/tag";

interface RuleForm {
  uuid: string;
  name: string;
  tag_uuid_list: string[];
  type: string;
  listen_address: string;
  listen_port: string;
  target_address: string;
  target_port: string;
  state?: any;
}

const emit = defineEmits(["getList"]);

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
  tag_uuid_list: [],
  type: "",
  listen_address: "",
  listen_port: "",
  target_address: "",
  target_port: "",
});

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>(createForm());
const tagList = ref([] as any[]);
const modalTitle = computed(() => (ruleForm.value.uuid ? "编辑代理" : "添加代理"));
const isSocks5Type = computed(() => ruleForm.value.type === "SOCKS5");

const rules = computed(() => ({
  name: [{ required: true, message: "请输入", trigger: "blur" }],
  listen_port: [{ required: true, message: "请输入", trigger: "blur" }],
  target_address: isSocks5Type.value ? [] : [{ required: true, message: "请输入", trigger: "blur" }],
  target_port: isSocks5Type.value ? [] : [{ required: true, message: "请输入", trigger: "blur" }],
  type: [{ required: true, message: "请输入", trigger: "change" }],
}));

watch(
  () => ruleForm.value.type,
  (type) => {
    if (type === "SOCKS5") {
      ruleForm.value.target_address = "";
      ruleForm.value.target_port = "";
      ruleFormRef.value?.clearValidate?.(["target_address", "target_port"]);
    }
  }
);

const loadTags = async () => {
  const res = await getTagList({});
  tagList.value = res.data || [];
};

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      const payload = {
        ...ruleForm.value,
        target_address: isSocks5Type.value ? "" : ruleForm.value.target_address,
        target_port: isSocks5Type.value ? "" : ruleForm.value.target_port,
      };
      if (ruleForm.value.uuid) {
        editProxy(payload).then(() => {
          message.success("操作成功");
          cancel();
          emit("getList");
        });
      } else {
        addProxy(payload).then(() => {
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
  ruleFormRef.value?.resetFields();
  ruleForm.value = createForm();
};

const cancel = () => {
  resetForm();
  showbox.value = false;
};

const showbox = ref(false);

const init = async (row?: RuleForm) => {
  await loadTags();
  if (row) {
    ruleForm.value = { ...createForm(), ...row, tag_uuid_list: row.tag_uuid_list || [] };
  } else {
    ruleForm.value = createForm();
  }

  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less">
.form-tip {
  color: rgba(0, 0, 0, 0.45);
}
</style>
