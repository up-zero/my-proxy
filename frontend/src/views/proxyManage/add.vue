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
      :label-col="{ flex: '118px' }"
      :wrapper-col="{ flex: 1 }"
    >
      <a-form-item ref="name" :label="t('common.name')" name="name" laba-position="top">
        <a-input v-model:value="ruleForm.name" />
      </a-form-item>
      <a-form-item :label="t('proxy.tags')" name="tag_uuid_list">
        <a-select v-model:value="ruleForm.tag_uuid_list" mode="multiple" allow-clear :placeholder="t('proxy.selectTags')">
          <a-select-option v-for="item in tagList" :key="item.uuid" :value="item.uuid">
            {{ item.name }}
          </a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="type" :label="t('common.type')" name="type">
        <a-select v-model:value="ruleForm.type" :placeholder="t('proxy.selectProxyType')">
          <a-select-option value="TCP">TCP</a-select-option>
          <a-select-option value="UDP">UDP</a-select-option>
          <a-select-option value="HTTP">HTTP</a-select-option>
          <a-select-option value="SOCKS5">SOCKS5</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="listen_address" :label="t('proxy.listenAddress')" name="listen_address">
        <a-input v-model:value="ruleForm.listen_address" placeholder="default 0.0.0.0"/>
      </a-form-item>
      <a-form-item ref="listen_port" :label="t('proxy.listenPort')" name="listen_port">
        <a-input v-model:value="ruleForm.listen_port" />
      </a-form-item>
      <a-form-item v-if="!isSocks5Type" ref="target_address" :label="t('proxy.targetAddress')" name="target_address">
        <a-input v-model:value="ruleForm.target_address" />
      </a-form-item>
      <a-form-item v-if="!isSocks5Type" ref="target_port" :label="t('proxy.targetPort')" name="target_port">
        <a-input v-model:value="ruleForm.target_port" />
      </a-form-item>
      <a-form-item v-if="isSocks5Type" :label="t('proxy.socks5Auth')">
        <span class="form-tip">{{ t("proxy.socks5AuthTip") }}</span>
      </a-form-item>
      <a-form-item v-if="isSocks5Type" :label="t('proxy.socks5Username')" name="socks5_username">
        <a-input v-model:value="ruleForm.socks5_username" :placeholder="t('proxy.inputSocks5Username')" />
      </a-form-item>
      <a-form-item v-if="isSocks5Type" :label="t('proxy.socks5Password')" name="socks5_password">
        <a-input-password v-model:value="ruleForm.socks5_password" :placeholder="t('proxy.inputSocks5Password')" />
      </a-form-item>
      <a-form-item v-if="isSocks5Type" :label="t('proxy.description')">
        <span class="form-tip">{{ t("proxy.socks5Tip") }}</span>
      </a-form-item>
    </a-form>
    <template #footer>
      <div class="dialog-footer">
        <a-button type="primary" @click="submitForm(ruleFormRef)">
          {{ t("common.confirm") }}
        </a-button>
        <a-button @click="cancel">{{ t("common.cancel") }}</a-button>
      </div>
    </template>
  </a-modal>
</template>
<script lang="ts" setup>
import { message } from "ant-design-vue";
import { computed, ref, watch } from "vue";
import { addProxy, editProxy } from "@/api/proxy";
import { getTagList } from "@/api/tag";
import { useAppI18n } from "@/i18n";

interface RuleForm {
  uuid: string;
  name: string;
  tag_uuid_list: string[];
  type: string;
  listen_address: string;
  listen_port: string;
  target_address: string;
  target_port: string;
  socks5_username: string;
  socks5_password: string;
  state?: any;
}

const emit = defineEmits(["getList"]);
const { t } = useAppI18n();

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
  tag_uuid_list: [],
  type: "",
  listen_address: "",
  listen_port: "",
  target_address: "",
  target_port: "",
  socks5_username: "",
  socks5_password: "",
});

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>(createForm());
const tagList = ref([] as any[]);
const modalTitle = computed(() => (ruleForm.value.uuid ? t("proxy.editProxy") : t("proxy.addProxy")));
const isSocks5Type = computed(() => ruleForm.value.type === "SOCKS5");

const rules = computed(() => ({
  name: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  listen_port: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  target_address: isSocks5Type.value ? [] : [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  target_port: isSocks5Type.value ? [] : [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  type: [{ required: true, message: t("password.pleaseInput"), trigger: "change" }],
}));

watch(
  () => ruleForm.value.type,
  (type) => {
    if (type === "SOCKS5") {
      ruleForm.value.target_address = "";
      ruleForm.value.target_port = "";
      ruleFormRef.value?.clearValidate?.(["target_address", "target_port"]);
    } else {
      ruleForm.value.socks5_username = "";
      ruleForm.value.socks5_password = "";
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
        socks5_username: isSocks5Type.value ? ruleForm.value.socks5_username : "",
        socks5_password: isSocks5Type.value ? ruleForm.value.socks5_password : "",
      };
      if (ruleForm.value.uuid) {
        editProxy(payload).then(() => {
          message.success(t("common.success"));
          cancel();
          emit("getList");
        });
      } else {
        addProxy(payload).then(() => {
          message.success(t("common.success"));
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
.demo-ruleForm :deep(.ant-form-item-label),
.demo-ruleForm :deep(.ant-form-item-label > label) {
  white-space: nowrap;
}

.form-tip {
  color: rgba(0, 0, 0, 0.45);
}
</style>
