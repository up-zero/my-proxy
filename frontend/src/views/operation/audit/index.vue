<template>
  <div class="p-page">
    <a-form :model="state.query" class="m-search">
      <div class="search-row">
        <div class="search-left">
          <a-input
            v-model:value="state.query.keyword"
            :placeholder="t('audit.keywordPlaceholder')"
            style="width: 280px; margin-right: 8px"
            @pressEnter="getList"
          />
          <a-select
            v-model:value="state.query.module"
            allow-clear
            :placeholder="t('audit.module')"
            style="width: 160px; margin-right: 8px"
          >
            <a-select-option value="AUTH">{{ t('audit.moduleAuth') }}</a-select-option>
            <a-select-option value="PROXY">{{ t('audit.moduleProxy') }}</a-select-option>
            <a-select-option value="TRAFFIC_POLICY">{{ t('audit.moduleTrafficPolicy') }}</a-select-option>
          </a-select>
          <a-select
            v-model:value="state.query.action"
            allow-clear
            :placeholder="t('audit.action')"
            style="width: 150px; margin-right: 8px"
          >
            <a-select-option value="LOGIN">{{ t('audit.actionLogin') }}</a-select-option>
            <a-select-option value="CREATE">{{ t('audit.actionCreate') }}</a-select-option>
            <a-select-option value="UPDATE">{{ t('audit.actionUpdate') }}</a-select-option>
            <a-select-option value="DELETE">{{ t('audit.actionDelete') }}</a-select-option>
            <a-select-option value="START">{{ t('audit.actionStart') }}</a-select-option>
            <a-select-option value="STOP">{{ t('audit.actionStop') }}</a-select-option>
            <a-select-option value="RESTART">{{ t('audit.actionRestart') }}</a-select-option>
            <a-select-option value="ENABLE">{{ t('audit.actionEnable') }}</a-select-option>
            <a-select-option value="DISABLE">{{ t('audit.actionDisable') }}</a-select-option>
            <a-select-option value="IMPORT">{{ t('audit.actionImport') }}</a-select-option>
          </a-select>
          <a-button type="primary" @click="getList">{{ t("common.search") }}</a-button>
        </div>
      </div>
    </a-form>

    <a-table
      :dataSource="state.list"
      :columns="columns"
      bordered
      :pagination="pagination"
      class="m-table"
      size="middle"
      :scroll="{ y: 'calc(100vh - 360px)' }"
      rowKey="uuid"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'module'">
          <a-tag :color="moduleColor(record.module)">{{ moduleText(record.module) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-tag :color="actionColor(record.action)">{{ actionText(record.action) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script lang="ts" setup>
import { getAuditList } from "@/api/audit";
import { useAppI18n } from "@/i18n";
import { parseTime } from "@/lib/util";
import { computed, onMounted, reactive } from "vue";

interface AuditItem {
  uuid: string;
  username: string;
  module: string;
  action: string;
  target: string;
  target_uuid: string;
  detail: string;
  source_ip: string;
  created_at: number;
}

const QUERY = () => ({
  keyword: "",
  module: undefined as string | undefined,
  action: undefined as string | undefined,
  page: 1,
  per_page: readStoredPageSize(),
});

const PAGE_SIZE_KEY = "my-proxy:audit-page-size";
const PAGE_SIZE_OPTIONS = ["20", "50", "100", "200", "500"];

function readStoredPageSize() {
  try {
    const value = Number(localStorage.getItem(PAGE_SIZE_KEY));
    return [20, 50, 100, 200, 500].includes(value) ? value : 20;
  } catch {
    return 20;
  }
}

function savePageSize(pageSize: number) {
  try {
    localStorage.setItem(PAGE_SIZE_KEY, String(pageSize));
  } catch {
    // ignore storage errors
  }
}

const { t } = useAppI18n();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as AuditItem[],
  total: 0,
});

onMounted(() => {
  getList();
});

const columns = computed(() => [
  { title: t("common.index"), dataIndex: "index", key: "index", width: 70 },
  { title: t("audit.username"), dataIndex: "username", key: "username", width: 120 },
  { title: t("audit.module"), dataIndex: "module", key: "module", width: 120 },
  { title: t("audit.action"), dataIndex: "action", key: "action", width: 100 },
  { title: t("audit.target"), dataIndex: "target", key: "target", width: 160 },
  { title: t("audit.detail"), dataIndex: "detail", key: "detail" },
  { title: t("audit.sourceIp"), dataIndex: "source_ip", key: "source_ip", width: 140 },
  { title: t("common.createdAt"), dataIndex: "created_at", key: "created_at", width: 180 },
]);

const pagination = computed(() => ({
  current: state.query.page,
  pageSize: state.query.per_page,
  total: state.total,
  showSizeChanger: true,
  pageSizeOptions: PAGE_SIZE_OPTIONS,
}));

async function getList() {
  try {
    state.isLoading = true;
    const res = await getAuditList(state.query);
    const data = res.data || {};
    const list = Array.isArray(data) ? data : data.list || [];
    state.total = Array.isArray(data) ? list.length : data.count || 0;
    state.list = list.map((it: AuditItem, index: number) => ({
      ...it,
      index: (state.query.page - 1) * state.query.per_page + index + 1,
    }));
  } finally {
    state.isLoading = false;
  }
}

function handleTableChange(pageInfo: any) {
  state.query.page = pageInfo.current;
  state.query.per_page = pageInfo.pageSize;
  savePageSize(state.query.per_page);
  getList();
}

function moduleText(module: string) {
  const map: Record<string, string> = {
    AUTH: t("audit.moduleAuth"),
    PROXY: t("audit.moduleProxy"),
    TRAFFIC_POLICY: t("audit.moduleTrafficPolicy"),
  };
  return map[module] || module || "-";
}

function moduleColor(module: string) {
  const map: Record<string, string> = {
    AUTH: "blue",
    PROXY: "cyan",
    TRAFFIC_POLICY: "purple",
  };
  return map[module] || "default";
}

function actionText(action: string) {
  const map: Record<string, string> = {
    LOGIN: t("audit.actionLogin"),
    CREATE: t("audit.actionCreate"),
    UPDATE: t("audit.actionUpdate"),
    DELETE: t("audit.actionDelete"),
    START: t("audit.actionStart"),
    STOP: t("audit.actionStop"),
    RESTART: t("audit.actionRestart"),
    ENABLE: t("audit.actionEnable"),
    DISABLE: t("audit.actionDisable"),
    IMPORT: t("audit.actionImport"),
  };
  return map[action] || action || "-";
}

function actionColor(action: string) {
  const map: Record<string, string> = {
    LOGIN: "blue",
    CREATE: "green",
    UPDATE: "cyan",
    DELETE: "red",
    START: "green",
    STOP: "orange",
    RESTART: "purple",
    ENABLE: "green",
    DISABLE: "orange",
    IMPORT: "geekblue",
  };
  return map[action] || "default";
}
</script>

<style lang="less" scoped>
.p-page {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  .m-search {
    margin-bottom: 10px;
    flex-shrink: 0;
  }

  .search-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-left {
    display: flex;
    align-items: center;
  }


  .m-table {
    flex: 1;
    min-height: 0;
  }
}
</style>
