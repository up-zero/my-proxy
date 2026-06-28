<template>
  <div class="p-page">
    <!-- 搜索 -->
    <a-form :inline="true" :model="state.query" class="m-search">
      <div class="search-row">
        <div class="search-left">
          <a-input
            v-model:value="state.query.keyword"
            :placeholder="t('node.searchPlaceholder')"
            style="width: 220px; margin-right: 8px"
            @pressEnter="getList"
          />
          <a-button type="primary" @click="getList">{{ t("common.search") }}</a-button>
        </div>
        <div>
          <a-button type="primary" @click="toAddPage" class="mr-2">{{ t("common.add") }}</a-button>
        </div>
      </div>
    </a-form>

    <a-table
      :dataSource="filteredList"
      :columns="columns"
      bordered
      :pagination="false"
      class="m-table"
      size="middle"
      :scroll="{ y: 'calc(100vh - 320px)' }"
      row-key="uuid"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'is_local'">
          <a-tag v-if="record.is_local" color="blue">{{ t("node.localTag") }}</a-tag>
          <span v-else>-</span>
        </template>
        <template v-else-if="column.key === 'secret_key'">
          <span class="secret-masked">••••••••</span>
          <a-button type="link" size="small" @click="copySecret(record.secret_key)">
            <template #icon><copy-outlined /></template>
          </a-button>
        </template>
        <template v-else-if="column.key === 'enabled'">
          <a-tag :color="record.enabled ? 'green' : 'default'">
            {{ record.enabled ? t("common.enabled") : t("common.disabled") }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
        <template v-else-if="column.key === 'updated_at'">
          <span>{{ parseTime(record.updated_at) }}</span>
        </template>
        <template v-else-if="column.key === 'operation'">
          <a-button v-if="!record.is_local" type="link" @click="editItem(record)">{{ t("common.edit") }}</a-button>
          <a-popconfirm
            v-if="!record.is_local"
            :title="t('node.deleteConfirm')"
            @confirm="delItem(record)"
          >
            <a-button type="link" danger>{{ t("common.delete") }}</a-button>
          </a-popconfirm>
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { getNodeList, deleteNode, type NodeItem } from "@/api/node";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { CopyOutlined } from "@ant-design/icons-vue";
import { computed, onMounted, reactive, ref } from "vue";
import addBox from "./add.vue";
import { useAppI18n } from "@/i18n";

const QUERY = (): any => ({
  keyword: "",
});

const addBoxRef = ref();
const { t } = useAppI18n();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as NodeItem[],
});

onMounted(() => {
  getList();
});

const filteredList = computed(() => {
  const kw = state.query.keyword.trim().toLowerCase();
  if (!kw) return state.list;
  return state.list.filter((it) => it.name.toLowerCase().includes(kw));
});

const columns = computed(() => [
  {
    title: t("common.index"),
    dataIndex: "index",
    key: "index",
    width: 80,
  },
  {
    title: t("node.name"),
    dataIndex: "name",
    key: "name",
    width: 160,
  },
  {
    title: t("node.address"),
    dataIndex: "address",
    key: "address",
    width: 220,
  },
  {
    title: t("node.secretKey"),
    dataIndex: "secret_key",
    key: "secret_key",
    width: 180,
  },
  {
    title: t("common.status"),
    dataIndex: "enabled",
    key: "enabled",
    width: 90,
  },
  {
    title: t("node.localTag"),
    dataIndex: "is_local",
    key: "is_local",
    width: 90,
  },
  {
    title: t("common.createdAt"),
    dataIndex: "created_at",
    key: "created_at",
    width: 180,
  },
  {
    title: t("node.updatedAt"),
    dataIndex: "updated_at",
    key: "updated_at",
    width: 180,
  },
  {
    title: t("common.operation"),
    dataIndex: "operation",
    key: "operation",
    width: 150,
  },
]);

async function getList() {
  try {
    state.isLoading = true;
    const res = await getNodeList();
    const list = res.data?.list || [];
    state.list = list.map((it: NodeItem, index: number) => ({
      ...it,
      index: index + 1,
    }));
  } finally {
    state.isLoading = false;
  }
}

function toAddPage() {
  addBoxRef.value.init();
}

function editItem(row: NodeItem) {
  if (row.is_local) return;
  addBoxRef.value.init(row);
}

function delItem(row: NodeItem) {
  deleteNode({ uuid: row.uuid }).then(() => {
    message.success(t("common.success"));
    getList();
  });
}

function copySecret(text: string) {
  if (!text) return;
  navigator.clipboard.writeText(text).then(() => {
    message.success(t("settings.copySuccess"));
  }).catch(() => {
    message.error(t("settings.copyFail"));
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
    flex-shrink: 0;
    margin-bottom: 10px;
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

  .mr-2 {
    margin-right: 10px;
  }

  .secret-masked {
    font-family: monospace;
    letter-spacing: 2px;
    color: var(--color-text-muted, #999);
  }

  .m-table {
    flex: 1;
    min-height: 0;

    :deep(.ant-table-tbody > tr > td) {
      padding-top: 7px;
      padding-bottom: 7px;
      line-height: 32px;
    }
    :deep(.ant-table-thead > tr > th) {
      padding-top: 12px;
      padding-bottom: 12px;
    }

    // 滚动条美化：窄 + 半透明，深色/浅色主题均适配
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
}
</style>
