<template>
  <div class="p-page">
    <!-- 搜索 -->
    <a-form :model="state.query" class="m-search">
      <div style="display: flex; justify-content: space-between;">
        <div>
          <a-button type="primary" @click="toAddPage" class="mr-2">新增</a-button>
          <fileUpload @success="getList" />
          <a-button type="primary" ghost @click="handleExport" :disabled="!selectedRowKeys.length" class="mr-2">导出</a-button>
          <a-button type="primary" danger :disabled="!selectedRowKeys.length" @click="delBatch">删除</a-button>
        </div>
        <div>
          <a-button type="primary" @click="getList">刷新</a-button>
        </div>
      </div>
    </a-form>

    <!-- 表格 -->
    <a-table :scroll="{ y: 'calc(100vh - 320px)' }" :dataSource="state.list" :columns="columns" bordered
      :pagination="false" rowKey="uuid" class="m-table" :row-selection="{
        selectedRowKeys: selectedRowKeys,
        onChange: onSelectChange,
      }" size="middle">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'state'">
          <span class="state-span danger" v-if="record.state === 'STOPPED'">已停止</span>
          <span class="state-span success" v-else-if="record.state === 'RUNNING'">运行中</span>
        </template>
        <template v-if="column.key === 'traffic_in'">

          <span class="traf-span">{{ getNum(record.traffic_out) }} / {{ getNum(record.traffic_in) }}</span>

        </template>
        <template v-else-if="column.key === 'operation'">
          <a-popconfirm v-if="state.list.length" title="确定删除?" @confirm="delItem(record)">
            <a-button type="link">删除</a-button>
          </a-popconfirm>
          <a-button type="link" @click="editItem(record)">编辑</a-button>
          <a-popconfirm v-if="record.state === 'STOPPED'" title="是否启动?" @confirm="startItem(record)">
            <a-button type="link">启动</a-button>
          </a-popconfirm>
          <a-popconfirm v-if="record.state === 'RUNNING'" title="是否停止?" @confirm="stopItem(record)">
            <a-button type="link">停止</a-button>
          </a-popconfirm>
          <a-popconfirm v-if="state.list.length" title="是否重启?" @confirm="restartItem(record)">
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
  delBacthProxy,
  exportProxy,
} from "@/api/proxy";
import addBox from "./add.vue";
import { ExclamationCircleOutlined } from "@ant-design/icons-vue";
import { message, Modal } from "ant-design-vue";
import { createVNode, onMounted, reactive, ref } from "vue";
import { downloadJson } from "@/lib/download";
import fileUpload from "./fileUpload.vue";


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
const selectedRowKeys = ref([] as any);
const addBoxRef = ref();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  total: 0,
  checkList: [] as string[],
});
const getNum = (num:number) => {
  if (num < 1024) {
    return num + 'B'
  } else if (num < 1024 * 1024) {
    return (num / 1024).toFixed(2) + 'KB'
  } else if (num < 1024 * 1024 * 1024) {
    return (num / (1024 * 1024)).toFixed(2) + 'MB'
  } else if (num < 1024 * 1024 * 1024 * 1024) {
    return (num / (1024 * 1024 * 1024)).toFixed(2) + 'GB'
  } else {
    return (num / (1024 * 1024 * 1024 * 1024)).toFixed(2) + 'TB'
  }
}
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
    title: "监听地址",
    dataIndex: "listen_address",
    key: "listen_address",
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
    title: "出/入站流量 ",
    dataIndex: "traffic_in",
    key: "traffic_in",
    // slots: { customRender: "state" },
  },
  {
    title: "操作",
    dataIndex: "operation",
    key: "operation",
    width: 280,
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

const onSelectChange = (selectedRowKeys1: any[]) => {
  console.log("selectedRowKeys changed: ", selectedRowKeys1);
  selectedRowKeys.value = selectedRowKeys1;
};
const handleExport = async () => {
  downloadJson(
    await exportProxy({ uuid: selectedRowKeys.value })

  );


};
const delBatch = () => {
  Modal.confirm({
    title: () => "确定要删除选中的项？",
    icon: () => createVNode(ExclamationCircleOutlined),
    content: () => "",
    okText: () => "确定",
    okType: "danger",
    cancelText: () => "取消",
    onOk() {
      delBacthProxy({ uuid: selectedRowKeys.value }).then(() => {
        getList();
        message.success({
          content: "操作成功",
        });
      });
    },
    onCancel() {
      console.log("Cancel");
    },
  });
};
</script>

<style lang="less" scoped>
.p-page {
  .table {
    height: 30vh;
  }

  .m-search {
    margin-bottom: 10px;
  }

  .m-table :deep(.ant-table-tbody) > tr:nth-child(even) {
    background-color: #fafafa; /* 设置偶数浅色斑马纹 */
  }

  .m-table :deep(.ant-table-thead) > tr > th {
    background-color: #fff !important; /* 表头为白色 */
  }

  .m-table :deep(.ant-table-tbody) > tr:hover td {
    background-color: #f0f8ff !important; /* 鼠标悬停时使用浅蓝色 */
  }

  .mr-2 {
    margin-right: 10px;
  }

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
