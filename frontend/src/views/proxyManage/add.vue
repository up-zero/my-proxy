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
          <a-select-option value="TCP_UDP">TCP/UDP</a-select-option>
          <a-select-option value="HTTP">HTTP</a-select-option>
          <a-select-option value="SOCKS5">SOCKS5</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item ref="listen_address" :label="t('proxy.listenAddress')" name="listen_address">
        <a-input v-model:value="ruleForm.listen_address" placeholder="default 0.0.0.0"/>
      </a-form-item>
      <a-form-item ref="listen_port" :label="t('proxy.listenPort')" name="listen_port">
        <a-input v-model:value="ruleForm.listen_port" :placeholder="t('proxy.inputListenPort')" />
      </a-form-item>
      <a-form-item v-if="isHttpType" ref="http_mode" :label="t('proxy.httpMode')" name="http_mode">
        <a-radio-group v-model:value="ruleForm.http_mode">
          <a-radio value="dynamic">{{ t("proxy.httpModeDynamic") }}</a-radio>
          <a-radio value="fixed">{{ t("proxy.httpModeFixed") }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item v-if="showTargetFields" ref="target_address" :label="t('proxy.targetAddress')" name="target_address">
        <a-input v-model:value="ruleForm.target_address" :placeholder="t('proxy.inputTargetAddress')" />
      </a-form-item>
      <a-form-item v-if="showTargetFields" ref="target_port" :label="t('proxy.targetPort')" name="target_port">
        <a-input v-model:value="ruleForm.target_port" :placeholder="t('proxy.inputTargetPort')" />
      </a-form-item>
      <a-form-item v-if="isHttpFixedType" ref="upstream_scheme" :label="t('proxy.upstreamScheme')" name="upstream_scheme">
        <a-select v-model:value="ruleForm.upstream_scheme">
          <a-select-option value="http">http</a-select-option>
          <a-select-option value="https">https</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item v-if="isHttpFixedType" :label="t('proxy.description')">
        <span class="form-tip">{{ t("proxy.httpFixedTip") }}</span>
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
      <a-form-item v-if="isHttpDynamicType" :label="t('proxy.httpAuth')">
        <span class="form-tip">{{ t("proxy.httpAuthTip") }}</span>
      </a-form-item>
      <a-form-item v-if="isHttpDynamicType" :label="t('proxy.httpUsername')" name="http_username">
        <a-input v-model:value="ruleForm.http_username" :placeholder="t('proxy.inputHttpUsername')" />
      </a-form-item>
      <a-form-item v-if="isHttpDynamicType" :label="t('proxy.httpPassword')" name="http_password">
        <a-input-password v-model:value="ruleForm.http_password" :placeholder="t('proxy.inputHttpPassword')" />
      </a-form-item>
      <a-form-item v-if="isHttpDynamicType" :label="t('proxy.description')">
        <span class="form-tip">{{ t("proxy.httpTip") }}</span>
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
  upstream_scheme: string;
  http_mode: string;
  socks5_username: string;
  socks5_password: string;
  http_username: string;
  http_password: string;
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
  upstream_scheme: "http",
  http_mode: "dynamic",
  socks5_username: "",
  socks5_password: "",
  http_username: "",
  http_password: "",
});

const formSize = ref("default");
const ruleFormRef = ref();
const ruleForm = ref<RuleForm>(createForm());
const tagList = ref([] as any[]);
const isCopy = ref(false);
const modalTitle = computed(() =>
  isCopy.value ? t("proxy.copyProxy") : ruleForm.value.uuid ? t("proxy.editProxy") : t("proxy.addProxy")
);
const isSocks5Type = computed(() => ruleForm.value.type === "SOCKS5");
const isHttpType = computed(() => ruleForm.value.type === "HTTP");
// HTTP 固定转发：目标地址与上游协议由服务端指定
const isHttpFixedType = computed(() => isHttpType.value && ruleForm.value.http_mode === "fixed");
// HTTP 动态代理：客户端自行配置代理地址，可访问任意目标
const isHttpDynamicType = computed(() => isHttpType.value && ruleForm.value.http_mode !== "fixed");
// 动态代理（SOCKS5、HTTP 动态代理）：无需配置目标地址和目标端口
const isDynamicType = computed(() => isSocks5Type.value || isHttpDynamicType.value);
const showTargetFields = computed(() => !isDynamicType.value);

const rules = computed(() => ({
  name: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  listen_port: [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  target_address: isDynamicType.value ? [] : [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  target_port: isDynamicType.value ? [] : [{ required: true, message: t("password.pleaseInput"), trigger: "blur" }],
  type: [{ required: true, message: t("password.pleaseInput"), trigger: "change" }],
}));

watch(
  () => ruleForm.value.type,
  (type) => {
    if (type === "SOCKS5") {
      ruleForm.value.target_address = "";
      ruleForm.value.target_port = "";
      ruleForm.value.upstream_scheme = "http";
      ruleForm.value.http_mode = "dynamic";
      ruleForm.value.http_username = "";
      ruleForm.value.http_password = "";
      ruleFormRef.value?.clearValidate?.(["target_address", "target_port"]);
    } else if (type === "HTTP") {
      ruleForm.value.socks5_username = "";
      ruleForm.value.socks5_password = "";
      if (ruleForm.value.http_mode !== "fixed") {
        // 动态代理无需目标地址和目标端口
        ruleForm.value.http_mode = "dynamic";
        ruleForm.value.target_address = "";
        ruleForm.value.target_port = "";
      }
      ruleFormRef.value?.clearValidate?.(["target_address", "target_port"]);
    } else {
      ruleForm.value.socks5_username = "";
      ruleForm.value.socks5_password = "";
      ruleForm.value.http_username = "";
      ruleForm.value.http_password = "";
    }
  }
);

// 切换 HTTP 代理模式时同步清理/补全字段
watch(
  () => ruleForm.value.http_mode,
  (mode) => {
    if (!isHttpType.value) return;
    if (mode === "fixed") {
      if (!ruleForm.value.upstream_scheme) {
        ruleForm.value.upstream_scheme = "http";
      }
      return;
    }
    ruleForm.value.target_address = "";
    ruleForm.value.target_port = "";
    ruleFormRef.value?.clearValidate?.(["target_address", "target_port"]);
  }
);

// 目标端口为 443 时，默认使用 https 作为上游协议
watch(
  () => ruleForm.value.target_port,
  (port) => {
    if (isHttpFixedType.value && port === "443") {
      ruleForm.value.upstream_scheme = "https";
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
        target_address: isDynamicType.value ? "" : ruleForm.value.target_address,
        target_port: isDynamicType.value ? "" : ruleForm.value.target_port,
        upstream_scheme: isHttpFixedType.value ? ruleForm.value.upstream_scheme : "",
        socks5_username: isSocks5Type.value ? ruleForm.value.socks5_username : "",
        socks5_password: isSocks5Type.value ? ruleForm.value.socks5_password : "",
        http_username: isHttpDynamicType.value ? ruleForm.value.http_username : "",
        http_password: isHttpDynamicType.value ? ruleForm.value.http_password : "",
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

// init 初始化弹窗；copy=true 为复制模式：带出当前配置但不携带 uuid，保存时调用新增接口
const init = async (row?: RuleForm, copy = false) => {
  await loadTags();
  isCopy.value = !!row && copy;
  if (row) {
    ruleForm.value = {
      ...createForm(),
      ...row,
      uuid: isCopy.value ? "" : row.uuid,
      tag_uuid_list: row.tag_uuid_list || [],
      // HTTP 类型：填写了目标地址即为固定转发
      http_mode: row.type === "HTTP" && row.target_address ? "fixed" : "dynamic",
      upstream_scheme: row.upstream_scheme || (row.target_port === "443" ? "https" : "http"),
    };
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
