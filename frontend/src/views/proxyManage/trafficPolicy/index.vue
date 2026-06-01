<template>
  <div class="p-page">
    <a-form :model="state.query" class="m-search">
      <div class="search-row">
        <div>
          <a-button type="primary" @click="toAddPage" class="mr-2">{{ t("common.add") }}</a-button>
        </div>
        <div class="search-right">
          <a-input v-model:value="state.query.name" :placeholder="t('trafficPolicy.inputPolicyName')" style="width: 220px; margin-right: 8px" @pressEnter="getList" />
          <a-select v-model:value="state.query.scope_type" allow-clear :placeholder="t('trafficPolicy.scopeType')" style="width: 160px; margin-right: 8px">
            <a-select-option value="ALL">{{ t('trafficPolicy.scopeAll') }}</a-select-option>
            <a-select-option value="TAG">{{ t('trafficPolicy.scopeTag') }}</a-select-option>
            <a-select-option value="PROXY">{{ t('trafficPolicy.scopeProxy') }}</a-select-option>
          </a-select>
          <a-select v-model:value="state.query.status" allow-clear :placeholder="t('common.status')" style="width: 140px; margin-right: 8px">
            <a-select-option value="ENABLED">{{ t('common.enabled') }}</a-select-option>
            <a-select-option value="DISABLED">{{ t('common.disabled') }}</a-select-option>
          </a-select>
          <a-button type="primary" @click="getList">{{ t("common.search") }}</a-button>
        </div>
      </div>
    </a-form>

    <a-table
      :dataSource="state.list"
      :columns="columns"
      bordered
      :pagination="false"
      class="m-table"
      size="middle"
      :scroll="{ x: 1320, y: 'calc(100vh - 320px)' }"
      rowKey="uuid"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'scope'">
          <a-tag :color="scopeColor(record.scope_type)">{{ scopeText(record) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'limits'">
          <div class="limit-cell">
            <a-tag v-if="record.outbound_limit">{{ t('trafficPolicy.outboundLimit') }}: {{ record.outbound_limit }}</a-tag>
            <a-tag v-if="record.max_connections">{{ t('trafficPolicy.maxConnections') }}: {{ record.max_connections }}</a-tag>
            <a-tag v-if="record.period_quota">{{ t('trafficPolicy.periodQuota') }}: {{ record.period_quota }}</a-tag>
          </div>
        </template>
        <template v-else-if="column.key === 'quota_used'">
          <span>{{ record.quota_used || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'over_limit_action'">
          <a-tag v-if="actionList(record).includes('SLOWDOWN')" color="orange">{{ t('trafficPolicy.actionSlowdown') }}</a-tag>
          <a-tag v-if="actionList(record).includes('ALERT')" color="red">{{ t('trafficPolicy.actionAlert') }}</a-tag>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-switch
            :checked="record.status === 'ENABLED'"
            :checked-children="t('common.enabled')"
            :un-checked-children="t('common.disabled')"
            @change="(checked: boolean) => toggleStatus(record, checked)"
          />
        </template>
        <template v-else-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
        <template v-else-if="column.key === 'operation'">
          <div class="operation-actions">
            <a-popconfirm :title="t('trafficPolicy.confirmDelete')" @confirm="delItem(record)">
              <a-button type="link" danger>{{ t("common.delete") }}</a-button>
            </a-popconfirm>
            <a-button type="link" @click="editItem(record)">{{ t("common.edit") }}</a-button>
          </div>
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" :tag-list="state.tagList" :proxy-list="state.proxyList" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { getProxyStatus } from "@/api/proxy";
import { getTagList } from "@/api/tag";
import { deleteTrafficPolicy, disableTrafficPolicy, enableTrafficPolicy, getTrafficPolicyList } from "@/api/trafficPolicy";
import { useAppI18n } from "@/i18n";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { computed, onMounted, reactive, ref } from "vue";
import addBox from "./add.vue";

interface DataItem {
  uuid: string;
  name: string;
  scope_type: string;
  scope_value: string;
  scope_name: string;
  outbound_limit: string;
  max_connections: string;
  period_quota: string;
  quota_used: string;
  over_limit_action: string;
  over_limit_action_list?: string[];
  status: string;
  created_at: number;
}

const QUERY = () => ({
  name: "",
  scope_type: undefined as string | undefined,
  status: undefined as string | undefined,
});

const { t } = useAppI18n();
const addBoxRef = ref();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as DataItem[],
  tagList: [] as any[],
  proxyList: [] as any[],
});

onMounted(async () => {
  await Promise.all([loadTags(), loadProxies()]);
  getList();
});

const columns = computed(() => [
  { title: t("common.index"), dataIndex: "index", key: "index", width: 80 },
  { title: t("trafficPolicy.policyName"), dataIndex: "name", key: "name", width: 180 },
  { title: t("trafficPolicy.scopeType"), dataIndex: "scope", key: "scope", width: 180 },
  { title: t("trafficPolicy.limitConfig"), dataIndex: "limits", key: "limits", width: 340 },
  { title: t("trafficPolicy.quotaUsed"), dataIndex: "quota_used", key: "quota_used", width: 140 },
  { title: t("trafficPolicy.overLimitAction"), dataIndex: "over_limit_action", key: "over_limit_action", width: 160 },
  { title: t("common.status"), dataIndex: "status", key: "status", width: 100 },
  { title: t("common.createdAt"), dataIndex: "created_at", key: "created_at", width: 180 },
  { title: t("common.operation"), dataIndex: "operation", key: "operation", width: 150, fixed: "right" },
]);

async function loadTags() {
  const res = await getTagList({});
  state.tagList = res.data || [];
}

async function loadProxies() {
  const res = await getProxyStatus({});
  state.proxyList = res.data || [];
}

async function getList() {
  try {
    state.isLoading = true;
    const res = await getTrafficPolicyList(state.query);
    state.list = (res.data || []).map((it: DataItem, index: number) => ({ ...it, index: index + 1 }));
  } finally {
    state.isLoading = false;
  }
}

function toAddPage() {
  addBoxRef.value.init();
}

function editItem(row: DataItem) {
  addBoxRef.value.init(row);
}

function actionList(row: DataItem) {
  return row.over_limit_action_list?.length ? row.over_limit_action_list : String(row.over_limit_action || "").split(",").filter(Boolean);
}

function scopeColor(scopeType: string) {
  if (scopeType === "TAG") return "blue";
  if (scopeType === "PROXY") return "purple";
  return "green";
}

function scopeText(row: DataItem) {
  if (row.scope_type === "TAG") return `${t("trafficPolicy.scopeTag")}: ${row.scope_name || row.scope_value}`;
  if (row.scope_type === "PROXY") return `${t("trafficPolicy.scopeProxy")}: ${row.scope_name || row.scope_value}`;
  return t("trafficPolicy.scopeAll");
}

async function toggleStatus(row: DataItem, checked: boolean) {
  if (checked) {
    await enableTrafficPolicy({ uuid: row.uuid });
  } else {
    await disableTrafficPolicy({ uuid: row.uuid });
  }
  message.success(t("common.success"));
  getList();
}

async function delItem(row: DataItem) {
  await deleteTrafficPolicy({ uuid: row.uuid });
  message.success(t("common.success"));
  getList();
}

</script>

<style lang="less" scoped>
.p-page {
  .m-search {
    margin-bottom: 10px;
  }
  .search-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .search-right {
    display: flex;
    align-items: center;
  }
  .mr-2 {
    margin-right: 10px;
  }
  .limit-cell {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }
  .operation-actions {
    display: flex;
    flex-wrap: wrap;
  }
}
</style>


