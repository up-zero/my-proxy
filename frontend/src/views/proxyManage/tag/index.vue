<template>
  <div class="p-page">
    <a-form :inline="true" :model="state.query" class="m-search">
      <div class="search-row">
        <div>
          <a-button type="primary" @click="toAddPage" class="mr-2">{{ t("common.add") }}</a-button>
          <a-popconfirm
            :title="t('tag.batchDeleteConfirm')"
            :disabled="selectedRowKeys.length === 0"
            @confirm="batchDelItems"
          >
            <a-button type="primary" danger :disabled="selectedRowKeys.length === 0">{{ t("common.delete") }}</a-button>
          </a-popconfirm>
        </div>
        <div class="search-right">
          <a-input
            v-model:value="state.query.name"
            :placeholder="t('tag.inputTagName')"
            style="width: 220px; margin-right: 8px"
            @pressEnter="getList"
          />
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
      :row-selection="{ selectedRowKeys, onChange: onSelectChange }"
      row-key="uuid"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
        <template v-else-if="column.key === 'operation'">
          <a-popconfirm
            v-if="state.list.length"
            :title="t('tag.deleteConfirm')"
            @confirm="delItem(record)"
          >
            <a-button type="link" danger>{{ t("common.delete") }}</a-button>
          </a-popconfirm>
          <a-button type="link" @click="editItem(record)">{{ t("common.edit") }}</a-button>
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { delTag, batchDelTag, getTagList } from "@/api/tag";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { computed, onMounted, reactive, ref } from "vue";
import addBox from "./add.vue";
import { useAppI18n } from "@/i18n";

interface DataItem {
  uuid: string;
  name: string;
  created_at: string;
  updated_at: string;
}

const QUERY = (): any => ({
  name: "",
});

const addBoxRef = ref();
const { t } = useAppI18n();
const selectedRowKeys = ref<string[]>([]);
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  total: 0,
});

onMounted(() => {
  getList();
});

const columns = computed(() => [
  {
    title: t("common.index"),
    dataIndex: "index",
    key: "index",
    width: 80,
  },
  {
    title: t("tag.tagName"),
    dataIndex: "name",
    key: "name",
  },
  {
    title: t("common.createdAt"),
    dataIndex: "created_at",
    key: "created_at",
    width: 220,
  },
  {
    title: t("common.operation"),
    dataIndex: "operation",
    key: "operation",
    width: 180,
  },
]);

async function getList() {
  try {
    state.isLoading = true;
    const res = await getTagList(state.query);
    if (!res.data) {
      state.list = [];
      state.total = 0;
      return;
    }
    state.list = res.data.map((it: any, index: number) => ({
      ...it,
      index: index + 1,
    }));
    state.total = res.data?.length;
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

function onSelectChange(keys: string[]) {
  selectedRowKeys.value = keys;
}

const delItem = (row: DataItem) => {
  delTag({ uuid: row.uuid }).then(() => {
    getList();
    message.success({
      content: t("common.success"),
    });
  });
};

function batchDelItems() {
  if (selectedRowKeys.value.length === 0) {
    message.warning({ content: t("tag.selectAtLeastOne") });
    return;
  }
  batchDelTag({ uuids: selectedRowKeys.value }).then(() => {
    selectedRowKeys.value = [];
    getList();
    message.success({
      content: t("common.success"),
    });
  });
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

  .m-table :deep(.ant-table-tbody) > tr:nth-child(even) {
    background-color: #fafafa;
  }

  .m-table :deep(.ant-table-thead) > tr > th {
    background-color: #fff !important;
  }

  .m-table :deep(.ant-table-tbody) > tr:hover td {
    background-color: #f0f8ff !important;
  }

  .mr-2 {
    margin-right: 10px;
  }
}
</style>

