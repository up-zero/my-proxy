<template>
  <div class="p-page">
    <a-form :inline="true" :model="state.query" class="m-search">
      <div class="search-row">
        <div class="search-left">
          <a-input
            v-model:value="state.query.name"
            :placeholder="t('tag.inputTagName')"
            style="width: 220px; margin-right: 8px"
            @pressEnter="getList"
          />
          <a-button type="primary" @click="getList">{{ t("common.search") }}</a-button>
        </div>
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
      </div>
    </a-form>

    <a-table
      :dataSource="state.list"
      :columns="columns"
      bordered
      :pagination="false"
      class="m-table"
      size="middle"
      :scroll="{ y: 'calc(100vh - 320px)' }"
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

    :deep(.ant-table-tbody) > tr:nth-child(even) {
      background-color: var(--color-table-stripe, #fafafa);
    }

    :deep(.ant-table-thead) > tr > th {
      background-color: var(--color-table-header, #fff) !important;
    }

    :deep(.ant-table-tbody) > tr:hover td {
      background-color: var(--color-table-hover, #f0f8ff) !important;
    }

    :deep(.ant-table-tbody > tr > td) {
      padding-top: 7px;
      padding-bottom: 7px;
    }
    :deep(.ant-table-thead > tr > th) {
      padding-top: 12px;
      padding-bottom: 12px;
    }

    :deep(.ant-table-body) {
      scrollbar-width: thin;
      scrollbar-color: rgba(128, 128, 128, 0.3) transparent;

      &::-webkit-scrollbar {
        width: 6px;
        height: 6px;
      }
      &::-webkit-scrollbar-track {
        background: transparent;
      }
      &::-webkit-scrollbar-thumb {
        background: rgba(128, 128, 128, 0.3);
        border-radius: 3px;
      }
      &::-webkit-scrollbar-thumb:hover {
        background: rgba(128, 128, 128, 0.5);
      }
    }
  }

  .mr-2 {
    margin-right: 10px;
  }
}
</style>

