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
      <a-form-item label="分组" name="group_uuid">
        <a-select v-model:value="ruleForm.group_uuid" allow-clear placeholder="请选择分组">
          <a-select-option v-for="item in groupList" :key="item.uuid" :value="item.uuid">
            {{ item.name }}
          </a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="type" label="类型" name="type">
        <a-select v-model:value="ruleForm.type" placeholder="please select">
          <a-select-option value="TCP">TCP</a-select-option>
          <a-select-option value="UDP">UDP</a-select-option>
          <a-select-option value="HTTP">HTTP</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="listen_address" label="监听地址" name="listen_address">
        <a-input v-model:value="ruleForm.listen_address" placeholder="default 0.0.0.0"/>
      </a-form-item>
      <a-form-item ref="listen_port" label="监听端口" name="listen_port">
        <a-input v-model:value="ruleForm.listen_port" />
      </a-form-item>
      <a-form-item ref="target_address" label="目标地址" name="target_address">
        <a-input v-model:value="ruleForm.target_address" />
      </a-form-item>
      <a-form-item ref="target_port" label="目标端口" name="target_port">
        <a-input v-model:value="ruleForm.target_port" />
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
import { computed, reactive, ref } from "vue";
import { addProxy, editProxy } from "@/api/proxy";
import { getGroupList } from "@/api/group";

interface RuleForm {
  uuid: string;
  name: string;
  group_uuid: string;
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
  group_uuid: "",
  type: "",
  listen_address: "",
  listen_port: "",
  target_address: "",
  target_port: "",
});

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>(createForm());
const groupList = ref([] as any[]);
const modalTitle = computed(() => (ruleForm.value.uuid ? "编辑代理" : "添加代理"));

const rules = reactive({
  name: [{ required: true, message: "请输入", trigger: "blur" }],
  listen_port: [{ required: true, message: "请输入", trigger: "blur" }],
  target_address: [{ required: true, message: "请输入", trigger: "blur" }],
  target_port: [{ required: true, message: "请输入", trigger: "blur" }],
  type: [{ required: true, message: "请输入", trigger: "change" }],
});

const loadGroups = async () => {
  const res = await getGroupList({});
  groupList.value = res.data || [];
};

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  await formEl
    .validate()
    .then(() => {
      if (ruleForm.value.uuid) {
        editProxy(ruleForm.value).then(() => {
          message.success("操作成功");
          cancel();
          emit("getList");
        });
      } else {
        addProxy(ruleForm.value).then(() => {
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
  await loadGroups();
  if (row) {
    ruleForm.value = { ...createForm(), ...row };
  } else {
    ruleForm.value = createForm();
  }

  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less"></style>
