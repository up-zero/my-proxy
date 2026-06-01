<template>
  <a-modal v-model:open="showbox" :title="modalTitle" width="720px" center>
    <a-alert
      type="info"
      show-icon
      :message="t('trafficPolicy.limitTip')"
      style="margin-bottom: 16px"
    />
    <a-form
      ref="ruleFormRef"
      :model="ruleForm"
      :rules="rules"
      :label-col="{ flex: '120px' }"
      :wrapper-col="{ flex: 1 }"
      status-icon
      class="traffic-policy-form"
    >
      <a-form-item :label="t('trafficPolicy.policyName')" name="name">
        <a-input v-model:value="ruleForm.name" :placeholder="t('trafficPolicy.inputPolicyName')" />
      </a-form-item>

      <a-form-item :label="t('trafficPolicy.scopeType')" name="scope_type">
        <a-select v-model:value="ruleForm.scope_type" @change="onScopeTypeChange">
          <a-select-option value="ALL">{{ t('trafficPolicy.scopeAll') }}</a-select-option>
          <a-select-option value="TAG">{{ t('trafficPolicy.scopeTag') }}</a-select-option>
          <a-select-option value="PROXY">{{ t('trafficPolicy.scopeProxy') }}</a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item v-if="ruleForm.scope_type === 'TAG'" :label="t('trafficPolicy.scopeValue')" name="scope_value">
        <a-select v-model:value="ruleForm.scope_value" :placeholder="t('trafficPolicy.selectScope')" allow-clear>
          <a-select-option v-for="tag in tagList" :key="tag.uuid" :value="tag.uuid">{{ tag.name }}</a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item v-if="ruleForm.scope_type === 'PROXY'" :label="t('trafficPolicy.scopeValue')" name="scope_value">
        <a-select v-model:value="ruleForm.scope_value" :placeholder="t('trafficPolicy.selectScope')" show-search allow-clear option-filter-prop="label">
          <a-select-option v-for="proxy in proxyList" :key="proxy.uuid" :value="proxy.uuid" :label="proxy.name">
            {{ proxy.name }}（{{ proxy.listen_port }}）
          </a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item :label="t('trafficPolicy.limitConfig')" required :validate-status="limitError ? 'error' : ''" :help="limitError">
        <div class="limit-item">
          <span class="limit-label">{{ t('trafficPolicy.outboundLimit') }}</span>
          <a-input-group compact class="limit-control">
            <a-input
                v-model:value="ruleForm.outbound_limit_value"
                class="limit-value"
                placeholder="50"
            />
            <a-select
                v-model:value="ruleForm.outbound_limit_unit"
                class="limit-unit"
            >
              <a-select-option
                  v-for="unit in rateUnits"
                  :key="unit"
                  :value="unit"
              >
                {{ unit }}/s
              </a-select-option>
            </a-select>
          </a-input-group>
        </div>

        <div class="limit-item">
          <span class="limit-label">{{ t('trafficPolicy.maxConnections') }}</span>
          <a-input
              v-model:value="ruleForm.max_connections"
              placeholder="300"
          />
        </div>

        <div class="limit-item">
          <span class="limit-label">{{ t('trafficPolicy.periodQuota') }}</span>
          <a-input-group compact class="limit-control">
            <a-input
                v-model:value="ruleForm.period_quota_value"
                class="limit-value"
                placeholder="800"
            />
            <a-select
                v-model:value="ruleForm.period_quota_unit"
                class="limit-unit"
            >
              <a-select-option
                  v-for="unit in quotaUnits"
                  :key="unit"
                  :value="unit"
              >
                {{ unit }}
              </a-select-option>
            </a-select>
          </a-input-group>
        </div>
      </a-form-item>

      <a-form-item :label="t('trafficPolicy.overLimitAction')" name="over_limit_action_list">
        <a-checkbox-group v-model:value="ruleForm.over_limit_action_list">
          <a-checkbox value="SLOWDOWN">{{ t('trafficPolicy.actionSlowdown') }}</a-checkbox>
          <a-checkbox value="ALERT">{{ t('trafficPolicy.actionAlert') }}</a-checkbox>
        </a-checkbox-group>
      </a-form-item>
    </a-form>
    <template #footer>
      <a-button type="primary" @click="submitForm(ruleFormRef)">{{ t("common.confirm") }}</a-button>
      <a-button @click="cancel">{{ t("common.cancel") }}</a-button>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
import { addTrafficPolicy, editTrafficPolicy } from "@/api/trafficPolicy";
import { useAppI18n } from "@/i18n";
import { message } from "ant-design-vue";
import { computed, ref } from "vue";

interface RuleForm {
  uuid: string;
  name: string;
  scope_type: string;
  scope_value: string;
  outbound_limit_value: string;
  outbound_limit_unit: string;
  max_connections: string;
  period_quota_value: string;
  period_quota_unit: string;
  over_limit_action_list: string[];
}

const props = defineProps<{
  tagList: Array<{ uuid: string; name: string }>;
  proxyList: Array<{ uuid: string; name: string; listen_port: string }>;
}>();
const emit = defineEmits(["getList"]);
const { t } = useAppI18n();
const ruleFormRef = ref();
const showbox = ref(false);
const limitError = ref("");
const rateUnits = ["KB", "MB", "GB"];
const quotaUnits = ["KB", "MB", "GB", "TB"];

const createForm = (): RuleForm => ({
  uuid: "",
  name: "",
  scope_type: "ALL",
  scope_value: "",
  outbound_limit_value: "",
  outbound_limit_unit: "MB",
  max_connections: "",
  period_quota_value: "",
  period_quota_unit: "GB",
  over_limit_action_list: ["ALERT"],
});

const ruleForm = ref<RuleForm>(createForm());
const tagList = computed(() => props.tagList || []);
const proxyList = computed(() => props.proxyList || []);
const modalTitle = computed(() => (ruleForm.value.uuid ? t("trafficPolicy.editPolicy") : t("trafficPolicy.addPolicy")));

const rules = computed(() => ({
  name: [{ required: true, message: t("trafficPolicy.inputPolicyName"), trigger: "blur" }],
  scope_type: [{ required: true, message: t("trafficPolicy.selectScopeType"), trigger: "change" }],
  scope_value: [{ required: ruleForm.value.scope_type !== "ALL", message: t("trafficPolicy.selectScope"), trigger: "change" }],
  over_limit_action_list: [{ required: true, type: "array", min: 1, message: t("trafficPolicy.selectAction"), trigger: "change" }],
}));

function onScopeTypeChange() {
  ruleForm.value.scope_value = "";
}

function parseLimit(value?: string, defaultUnit = "MB") {
  const raw = String(value || "").trim().toUpperCase();
  const match = raw.match(/([0-9]+(?:\.[0-9]+)?)\s*(KB|MB|GB|TB)?/);
  return {
    value: match?.[1] || "",
    unit: match?.[2] || defaultUnit,
  };
}

function buildPayload() {
  const form = ruleForm.value;
  return {
    uuid: form.uuid,
    name: form.name,
    scope_type: form.scope_type,
    scope_value: form.scope_value,
    outbound_limit: form.outbound_limit_value ? `${form.outbound_limit_value}${form.outbound_limit_unit}/s` : "",
    max_connections: form.max_connections,
    period_quota: form.period_quota_value ? `${form.period_quota_value}${form.period_quota_unit}` : "",
    over_limit_action_list: form.over_limit_action_list,
  };
}

function validateLimitConfig() {
  limitError.value = "";
  if (!ruleForm.value.outbound_limit_value && !ruleForm.value.max_connections && !ruleForm.value.period_quota_value) {
    limitError.value = t("trafficPolicy.limitRequired");
    return false;
  }
  return true;
}

const submitForm = async (formEl: any | undefined) => {
  if (!formEl) return;
  try {
    await formEl.validate();
    if (!validateLimitConfig()) return;
    const payload = buildPayload();
    const request = payload.uuid ? editTrafficPolicy(payload) : addTrafficPolicy(payload);
    await request;
    message.success(t("common.success"));
    cancel();
    emit("getList");
  } catch {
    // form validation failed
  }
};

const resetForm = () => {
  ruleFormRef.value?.resetFields();
  ruleForm.value = createForm();
  limitError.value = "";
};

const cancel = () => {
  resetForm();
  showbox.value = false;
};

const init = (row?: any) => {
  const next = createForm();
  if (row) {
    const outbound = parseLimit(row.outbound_limit, "MB");
    const quota = parseLimit(row.period_quota, "GB");
    Object.assign(next, {
      uuid: row.uuid || "",
      name: row.name || "",
      scope_type: row.scope_type || "ALL",
      scope_value: row.scope_value || "",
      outbound_limit_value: outbound.value,
      outbound_limit_unit: outbound.unit,
      max_connections: row.max_connections || "",
      period_quota_value: quota.value,
      period_quota_unit: quota.unit,
      over_limit_action_list: row.over_limit_action_list?.length
        ? row.over_limit_action_list
        : String(row.over_limit_action || "").split(",").filter(Boolean),
    });
  }
  ruleForm.value = next;
  limitError.value = "";
  showbox.value = true;
};

defineExpose({ init });
</script>

<style scoped lang="less">
.traffic-policy-form :deep(.ant-form-item-control-input-content) {
  min-width: 0;
}

.limit-grid {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.limit-item {
  display: flex;
  align-items: center;
  min-width: 0;
  margin-bottom: 12px;
}

.limit-item:last-child {
  margin-bottom: 0;
}

.limit-label {
  flex-shrink: 0;
  margin-right: 8px;
  color: #595959;
  white-space: nowrap;
}

.limit-control {
  flex: 1;
  display: flex;
}

.limit-item .ant-input {
  flex: 1;
}

.limit-item :deep(.ant-input-group-compact) {
  display: flex;
  width: 100%;
}

.limit-value {
  flex: 1;
  min-width: 0;
  width: auto !important;
}

.limit-unit {
  flex: 0 0 80px;
  width: 80px !important;
}

@media (max-width: 1200px) {
  .limit-grid {
    grid-template-columns: 1fr;
  }
}
</style>
