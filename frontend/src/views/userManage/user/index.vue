<template>
  <div class="p-page">
    <!-- 搜索 -->

    <a-form :inline="true" :model="state.query" class="m-search">
      <a-button type="primary" @click="toAddPage" style="margin-bottom: 10px"
        >{{ t("common.add") }}</a-button
      >
    </a-form>

    <!-- 表格 -->
    <a-table
      :dataSource="state.list"
      :columns="columns"
      bordered
      :pagination="false"
      class="m-table"
      size="middle"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'role_id'">
          <a-tag :color="getRoleColor(record.role_id)">{{ getRoleName(record.role_id) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          <span> {{ parseTime(record.created_at) }} </span>
        </template>
        <template v-else-if="column.key === 'operation'">
          <a-popconfirm
            v-if="state.list.length"
            :title="t('user.deleteConfirm')"
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
import { getUserList, delUser } from "@/api/user";
import { getRoleList } from "@/api/role";
import addBox from "./add.vue";
import { parseTime } from "@/lib/util";
import { message } from "ant-design-vue";
import { computed, onMounted, reactive, ref } from "vue";
import { useAppI18n } from "@/i18n";

interface DataItem {
  uuid: string;
  username: string;
  password: string;
  level: string;
  role_id: string;
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
const { t } = useAppI18n();
const roleMap = ref<Record<string, any>>({});
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  total: 0,
  checkList: [] as string[],
});

onMounted(() => {
  getList();
  loadRoles();
});

async function loadRoles() {
  try {
    const res = await getRoleList({});
    const roles = res.data || [];
    const map: Record<string, any> = {};
    for (const r of roles) {
      map[r.uuid] = r;
    }
    roleMap.value = map;
  } catch {
    roleMap.value = {};
  }
}

function getRoleName(roleId: string): string {
  if (!roleId || !roleMap.value[roleId]) return t("common.none");
  const role = roleMap.value[roleId];
  const key = `role.roleNames.${role.name}`;
  const translated = t(key);
  return translated !== key ? translated : role.name;
}

function getRoleColor(roleId: string): string {
  if (!roleId || !roleMap.value[roleId]) return "";
  const role = roleMap.value[roleId];
  if (role.name === "admin") return "red";
  if (role.name === "ops") return "blue";
  return "default";
}
const columns = computed(() => [
  {
    title: t("common.index"),
    dataIndex: "index",
    key: "index",
    width: 80,
  },
  {
    dataIndex: "username",
    title: t("user.userNameColumn"),
    key: "username",
  },
  {
    title: t("user.role"),
    dataIndex: "role_id",
    key: "role_id",
  },
  {
    title: t("common.createdAt"),
    dataIndex: "created_at",
    key: "created_at",
  },
  {
    title: t("common.operation"),
    dataIndex: "operation",
    key: "operation",
  },
]);
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
      content: t("common.success"),
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
