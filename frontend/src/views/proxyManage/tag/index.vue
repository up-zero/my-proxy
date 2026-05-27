<template>
  <div class="p-page">
    <a-form :inline="true" :model="state.query" class="m-search">
      <a-button type="primary" @click="toAddPage" class="mr-2">新增</a-button>
      <a-input
        v-model:value="state.query.name"
        placeholder="请输入标签名称"
        style="width: 220px; margin-right: 8px"
        @pressEnter="getList"
      />
      <a-button type="primary" @click="getList">搜索</a-button>
    </a-form>

    <a-table :dataSource="state.list" :columns="columns" bordered :pagination="false" class="m-table" size="middle">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'created_at'">
          <span>{{ parseTime(record.created_at) }}</span>
        </template>
        <template v-else-if="column.key === 'operation'">
          <a-popconfirm
            v-if="state.list.length"
            title="确定删除标签?"
            @confirm="delItem(record)"
          >
            <a-button type="link">删除</a-button>
          </a-popconfirm>
          <a-button type="link" @click="editItem(record)">编辑</a-button>
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { delTag, getTagList } from "@/api/tag";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { onMounted, reactive, ref } from "vue";
import addBox from "./add.vue";

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
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  total: 0,
});

onMounted(() => {
  getList();
});

const columns = [
  {
    title: "序号",
    dataIndex: "index",
    key: "index",
    width: 80,
  },
  {
    title: "标签名称",
    dataIndex: "name",
    key: "name",
  },
  {
    title: "创建时间",
    dataIndex: "created_at",
    key: "created_at",
    width: 220,
  },
  {
    title: "操作",
    dataIndex: "operation",
    key: "operation",
    width: 180,
  },
];

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

const delItem = (row: DataItem) => {
  delTag({ uuid: row.uuid }).then(() => {
    getList();
    message.success({
      content: "操作成功",
    });
  });
};
</script>

<style lang="less" scoped>
.p-page {
  .m-search {
    margin-bottom: 10px;
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

