<template>
  <div class="p-page">
    <a-form :model="state.query" class="m-search">
      <div class="search-row">
        <div class="search-right">
          <a-input v-model:value="state.query.keyword" :placeholder="t('alert.keywordPlaceholder')" style="width: 280px; margin-right: 8px" @pressEnter="getList" />
          <a-select v-model:value="state.query.level" allow-clear :placeholder="t('alert.level')" style="width: 150px; margin-right: 8px">
            <a-select-option value="INFO">INFO</a-select-option>
            <a-select-option value="WARNING">WARNING</a-select-option>
            <a-select-option value="ERROR">ERROR</a-select-option>
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
      :scroll="{ y: 'calc(100vh - 330px)' }"
      rowKey="uuid"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'level'">
          <a-tag :color="levelColor(record.level)">{{ record.level }}</a-tag>
        </template>
        <template v-else-if="column.key === 'source'">
          <a-tag color="blue">{{ sourceText(record.source) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script lang="ts" setup>
import { getAlertList } from "@/api/alert";
import { useAppI18n } from "@/i18n";
import { parseTime } from "@/lib/util";
import { computed, onMounted, reactive } from "vue";

interface AlertItem {
  uuid: string;
  source: string;
  source_uuid: string;
  level: string;
  title: string;
  content: string;
  created_at: number;
}

const QUERY = () => ({
  keyword: "",
  level: undefined as string | undefined,
  page: 1,
  per_page: readStoredPageSize(),
});

const PAGE_SIZE_KEY = "my-proxy:alert-page-size";
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
  list: [] as AlertItem[],
  total: 0,
});

onMounted(() => {
  getList();
});

const columns = computed(() => [
  { title: t("common.index"), dataIndex: "index", key: "index", width: 80 },
  { title: t("alert.level"), dataIndex: "level", key: "level", width: 120 },
  { title: t("alert.source"), dataIndex: "source", key: "source", width: 160 },
  { title: t("alert.title"), dataIndex: "title", key: "title", width: 220 },
  { title: t("alert.content"), dataIndex: "content", key: "content" },
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
    const res = await getAlertList(state.query);
    const data = res.data || {};
    const list = Array.isArray(data) ? data : data.list || [];
    state.total = Array.isArray(data) ? list.length : data.count || 0;
    state.list = list.map((it: AlertItem, index: number) => ({ ...it, index: (state.query.page - 1) * state.query.per_page + index + 1 }));
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

function levelColor(level: string) {
  if (level === "ERROR") return "red";
  if (level === "WARNING") return "orange";
  return "blue";
}

function sourceText(source: string) {
  if (source === "TRAFFIC_POLICY") return t("alert.sourceTrafficPolicy");
  return source || "-";
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
    justify-content: flex-end;
    align-items: center;
  }

  .search-right {
    display: flex;
    align-items: center;
  }

  .m-table {
    flex: 1;
    min-height: 0;
  }
}
</style>


