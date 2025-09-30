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
        <template v-if="column.key === 'created_at'">
          <span> {{ parseTime(record.created_at) }} </span>
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
        </template>
      </template>
    </a-table>

    <addBox ref="addBoxRef" @get-list="getList" />
  </div>
</template>

<script lang="ts" setup>
import { getUserList, delUser } from "@/api/user";
import addBox from "./add.vue";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { onMounted, reactive, ref } from "vue";

interface DataItem {
  uuid: string;
  username: string;
  password: string;
  level: string;
  created_at: string;
  updated_at: string;
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
    dataIndex: "username",
    title: "用户名称",
    key: "username",
  },
  {
    title: "创建时间",
    dataIndex: "created_at",
    key: "created_at",
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
    const res = await getUserList(state.query);
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
  delUser({ uuid: [row.uuid] }).then(() => {
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
