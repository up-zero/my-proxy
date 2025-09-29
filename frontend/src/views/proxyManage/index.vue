<template>
  <div class="p-page">
    <!-- 搜索 -->

    <a-form :inline="true" :model="state.query" class="m-search">
      <a-button type="primary" @click="toAddPage" style="margin-bottom: 10px"
        >新增</a-button
      >
    </a-form>

    <!-- 表格 -->
    <a-table
      :dataSource="state.list"
      :columns="columns"
      bordered
      :pagination="false"
      class="m-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'state'">
          <span class="state-span danger" v-if="record.state === 'STOPPED'"
            >已停止</span
          >
          <span
            class="state-span success"
            v-else-if="record.state === 'RUNNING'"
            >运行中</span
          >
        </template>
        <template v-else-if="column.key === 'operation'">
          <a-popconfirm
            v-if="state.list.length"
            title="确定删除?"
            @confirm="delItem(record)"
          >
            <a-button type="link">删除</a-button>
          </a-popconfirm>
          <a-button type="link" @click="editItem(record)">编辑</a-button>
          <a-popconfirm
            v-if="record.state === 'STOPPED'"
            title="是否启动?"
            @confirm="startItem(record)"
          >
            <a-button type="link">启动</a-button>
          </a-popconfirm>
          <a-popconfirm
            v-if="record.state === 'RUNNING'"
            title="是否停止?"
            @confirm="stopItem(record)"
          >
            <a-button type="link">停止</a-button>
          </a-popconfirm>
          <a-popconfirm
            v-if="state.list.length"
            title="是否重启?"
            @confirm="restartItem(record)"
          >
            <a-button type="link">重启</a-button>
          </a-popconfirm>
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import {
  getProxyStatus,
  delProxy,
  startProxy,
  stopProxy,
  restartProxy,
} from "@/api/proxy";
import addBox from "./add.vue";

import { message } from "ant-design-vue";
import { onMounted, reactive, ref } from "vue";

interface DataItem {
  uuid: string;
  name: string;
  type: string;
  listen_port: string;
  target_address: string;
  target_port: string;
  state?: any;
}

const QUERY = (): any => ({
  search: "",
  page: 1,
  per_page: 10,
  position: 1,
});
const addBoxRef = ref();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  total: 0,
  checkList: [] as string[],
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
    dataIndex: "name",
    title: "名称",
    key: "name",
  },
  {
    title: "类型",
    dataIndex: "type",
    key: "type",
  },
  {
    title: "监听端口",
    dataIndex: "listen_port",
    key: "listen_port",
  },
  {
    title: "目标地址",
    key: "target_address",
    dataIndex: "target_address",
  },
  {
    title: "目标端口",
    key: "target_port",
    dataIndex: "target_port",
  },
  {
    title: "状态",
    dataIndex: "state",
    key: "state",
    // slots: { customRender: "state" },
  },
  {
    title: "操作",
    dataIndex: "operation",
    key: "operation",
    // slots: { customRender: "operation" },
  },
];
/*****************表格******************* */

// 获取列表
async function getList() {
  try {
    state.isLoading = true;
    const res = await getProxyStatus(state.query);
    if (!res.data) return;
    state.list = res.data.map((it: any, index: number) => {
      return {
        ...it,
        index: index + 1,
      };
    });
    state.total = res.data?.length;
  } finally {
    state.isLoading = false;
  }
}

//  新增
function toAddPage() {
  addBoxRef.value.init();
}
//  详情
function editItem(row: DataItem) {
  addBoxRef.value.init(row);
}

// 删除
const delItem = (row: DataItem) => {
  delProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: "操作成功",
    });
  });
};
const startItem = (row: DataItem) => {
  startProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: "操作成功",
    });
  });
};
const stopItem = (row: DataItem) => {
  stopProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: "操作成功",
    });
  });
};
const restartItem = (row: DataItem) => {
  restartProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: "操作成功",
    });
  });
};
</script>

<style lang="less" scoped>
.p-page {
  .state-span {
    &::before {
      content: "";
      display: inline-block;
      width: 10px;
      height: 10px;
      border-radius: 10px;
      margin-right: 10px;
    }
    &.danger::before {
      background-color: #f56c6c;
    }
    &.success::before {
      background-color: #52c41a;
    }
  }

  .m-page {
    margin-top: 30px;
  }
}
</style>
