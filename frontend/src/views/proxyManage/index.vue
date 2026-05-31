<template>
  <div class="p-page">
    <!-- 搜索 -->
    <a-form :model="state.query" class="m-search">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <div>
          <a-button type="primary" @click="toAddPage" class="mr-2">{{ t("common.add") }}</a-button>
          <fileUpload @success="getList" />
          <a-button type="primary" ghost @click="handleExport" :disabled="!selectedRowKeys.length" class="mr-2">{{ t("common.export") }}</a-button>
          <a-button type="primary" danger :disabled="!selectedRowKeys.length" @click="delBatch">{{ t("common.delete") }}</a-button>
        </div>
        <div style="display: flex; align-items: center;">
          <a-input v-model:value="state.query.name" :placeholder="t('proxy.inputProxyName')" style="width: 200px; margin-right: 8px;" @pressEnter="getList"></a-input>
          <a-select v-model:value="state.query.tag_uuid_list" mode="multiple" allow-clear :placeholder="t('proxy.selectTags')" style="width: 240px; margin-right: 8px;">
            <a-select-option v-for="item in state.tagList" :key="item.uuid" :value="item.uuid">
              {{ item.name }}
            </a-select-option>
          </a-select>
          <a-button type="primary" @click="getList">{{ t("common.search") }}</a-button>
        </div>
      </div>
    </a-form>

    <!-- 表格 -->
    <a-table :scroll="{ x: 1560, y: 'calc(100vh - 320px)' }" :dataSource="state.list" :columns="columns" bordered
      table-layout="fixed"
      :pagination="false" rowKey="uuid" class="m-table" :row-selection="{
        selectedRowKeys: selectedRowKeys,
        onChange: onSelectChange,
      }" size="middle">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'state'">
          <span class="state-span danger" v-if="record.state === 'STOPPED'">{{ t("proxy.statusStopped") }}</span>
          <span class="state-span success" v-else-if="record.state === 'RUNNING'">{{ t("proxy.statusRunning") }}</span>
        </template>
        <template v-else-if="column.key === 'tag_list'">
          <div class="tag-list-cell">
            <template v-if="record.tag_list?.length">
              <a-tag v-for="tag in record.tag_list" :key="tag.uuid">{{ tag.name }}</a-tag>
            </template>
            <span v-else>-</span>
          </div>
        </template>
        <template v-else-if="column.key === 'target_address'">
          <span class="single-line-text">{{ record.target_address || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'target_port'">
          <span class="single-line-text">{{ record.target_port || '-' }}</span>
        </template>
        <template v-if="column.key === 'traffic_in'">

          <span class="traf-span">{{ getNum(record.traffic_out) }} / {{ getNum(record.traffic_in) }}</span>

        </template>
        <template v-if="column.key === 'listen_port'">
          <a @click="showQuickAccessModal(record)" class="port-link single-line-text">{{ record.listen_port }}</a>
        </template>
        <template v-else-if="column.key === 'operation'">
          <div class="operation-actions">
            <a-popconfirm v-if="state.list.length" :title="t('proxy.confirmDelete')" @confirm="delItem(record)">
              <a-button type="link">{{ t("proxy.actionDelete") }}</a-button>
            </a-popconfirm>
            <a-button type="link" @click="editItem(record)">{{ t("proxy.actionEdit") }}</a-button>
            <a-popconfirm v-if="record.state === 'STOPPED'" :title="t('proxy.confirmStart')" @confirm="startItem(record)">
              <a-button type="link">{{ t("proxy.actionStart") }}</a-button>
            </a-popconfirm>
            <a-popconfirm v-if="record.state === 'RUNNING'" :title="t('proxy.confirmStop')" @confirm="stopItem(record)">
              <a-button type="link">{{ t("proxy.actionStop") }}</a-button>
            </a-popconfirm>
            <a-popconfirm v-if="state.list.length" :title="t('proxy.confirmRestart')" @confirm="restartItem(record)">
              <a-button type="link">{{ t("proxy.actionRestart") }}</a-button>
            </a-popconfirm>
            <a-button type="link" :disabled="record.state !== 'RUNNING'" @click="captureItem(record)">{{ t("proxy.actionCapture") }}</a-button>
          </div>
        </template>
      </template>
    </a-table>



    <addBox ref="addBoxRef" @get-list="getList" />

    <!-- 快捷访问弹窗 -->
    <a-modal v-model:open="quickAccessModalVisible" :title="t('proxy.quickAccess')" width="735px" center footer="">
      <div class="quick-access-modal" style="margin-top: 15px;">
        <!-- SSH用户名输入框 -->
        <div style="margin-bottom: 16px;">
          <a-form-item :label="t('proxy.sshUsername')" class="quick-access-form-item" style="margin-bottom: 0;">
            <a-input v-model:value="sshUser" @input="saveSshUser" :placeholder="t('proxy.inputSshUsername')" style="width: 300px;">
              <template #addonAfter>
                <a-tooltip :title="t('proxy.sshLoginTip')">
                  <info-circle-outlined />
                </a-tooltip>
              </template>
            </a-input>
          </a-form-item>
        </div>
        
        <a-table :data-source="quickLinks" :columns="linkColumns" bordered size="middle" :pagination="false">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'action'">
              <a-button type="primary" :disabled="record.disabled" @click="openLink(record.url)">{{ record.label }}</a-button>
            </template>
          </template>
        </a-table>
        <div style="margin-top: 24px; text-align: right;">
          <a-button @click="quickAccessModalVisible = false">{{ t("common.close") }}</a-button>
        </div>
      </div>
    </a-modal>
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
import { getTagList } from "@/api/tag";
import addBox from "./add.vue";
import { ExclamationCircleOutlined, InfoCircleOutlined } from "@ant-design/icons-vue";
import { message, Modal } from "ant-design-vue";
import { createVNode, onMounted, reactive, ref, computed } from "vue";
import { downloadJson } from "@/lib/download";
import fileUpload from "./fileUpload.vue";
import { useRouter } from "vue-router";
import { useAppI18n } from "@/i18n";


interface DataItem {
  uuid: string;
  name: string;
  tag_uuid_list?: string[];
  tag_list?: Array<{ uuid: string; name: string }>;
  type: string;
  listen_port: string;
  target_address: string;
  target_port: string;
  state?: any;
}

const QUERY = (): any => ({
  name: "",
  tag_uuid_list: [],
  page: 1,
  per_page: 10,
  position: 1,
});
const selectedRowKeys = ref([] as any);
const addBoxRef = ref();
const router = useRouter();
const { t } = useAppI18n();
const state = reactive({
  isLoading: false,
  query: QUERY(),
  list: [] as any,
  tagList: [] as any,
  total: 0,
  checkList: [] as string[],
});

// 快捷访问弹窗相关
const quickAccessModalVisible = ref(false);
const currentPort = ref("");
const sshUser = ref(localStorage.getItem('sshUser') || '');

// 保存SSH用户到本地存储
const saveSshUser = () => {
  localStorage.setItem('sshUser', sshUser.value);
};

// 链接表格列定义
const linkColumns = computed(() => [
  {
    title: t("common.name"),
    dataIndex: "name",
    key: "name",
    width: 135,
  },
  {
    title: t("common.link"),
    dataIndex: "url",
    key: "url",
    ellipsis: true,
  },
  {
    title: t("common.operation"),
    dataIndex: "action",
    key: "action",
    width: 200,
  },
]);

// 获取当前域
const getCurrentDomain = () => {
  return window.location.hostname;
};

// 快捷链接配置
const quickLinks = computed(() => {
  const domain = getCurrentDomain();
  const currentName = currentRecord.value?.name || "";
  const hasSshUser = !!sshUser.value;
  const hasPort = !!currentPort.value;

  return [
    {
      key: "1",
      name: "Web",
      label: t("proxy.webAccess"),
      url: `http://${domain}:${currentPort.value}`,
      disabled: false
    },
    {
      key: "2",
      name: "MobaXterm(SSH)",
      label: t("proxy.openMobaxterm"),
      url: generateMobaSSHUrl(domain, currentPort.value, currentName),
      disabled: !hasSshUser || !hasPort
    },
    {
      key: "3",
      name: "WinSCP(SFTP)",
      label: t("proxy.openWinScp"),
      url: generateWinScpUrl(domain, currentPort.value),
      disabled: !hasSshUser || !hasPort
    },
  ];
});

// 生成 MobaXterm SSH 连接 URL
const generateMobaSSHUrl = (ip: string, port: string, sessionName: string) => {
  const encodedSessionName = encodeURIComponent(sessionName);
  const encodedIp = encodeURIComponent(ip);
  const user = sshUser.value;
  const suffix = `%25%25%2D1%25%2D1%25%25%25%25%250%250%250%25%25%25%2D1%250%250%250%25%251080%25%250%250%251%25%250%25%25%25%250%25%2D1%25%2D1%250%23MobaFont%2510%250%250%25%2D1%2515%25236%2C236%2C236%2530%2C30%2C30%25180%2C180%2C192%250%25%2D1%250%25%25xterm%25%2D1%250%25%5FStd%5FColors%5F0%5F%2580%2524%250%251%25%2D1%25%3Cnone%3E%25%250%250%25%2D1%25%2D1%230%23%20%23%2D1`;

  return `mobaxterm:${encodedSessionName}%3D%23109%230%25${encodedIp}%25${port}%25${user}${suffix}`;
};

// 生成 WinSCP SFTP 连接 URL（若本地无已保存会话，则以临时会话方式直接打开）
const generateWinScpUrl = (ip: string, port: string) => {
  const user = encodeURIComponent(sshUser.value);
  const host = encodeURIComponent(ip);

  return `winscp-sftp://${user}@${host}:${port}/`;
};

// 打开链接
const openLink = (url: string) => {
  window.open(url, "_blank");
};

// 显示快捷访问弹窗
const currentRecord = ref<any>(null);
const showQuickAccessModal = (record: any) => {
  currentRecord.value = record;
  currentPort.value = record.listen_port;
  quickAccessModalVisible.value = true;
};

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
onMounted(async () => {
  await loadTags();
  getList();
});
const tagColumnWidth = computed(() => {
  const maxTagTextLength = state.list.reduce((max: number, item: any) => {
    const tagText = Array.isArray(item.tag_list)
      ? item.tag_list.map((tag: any) => tag.name).join(" ")
      : "";

    return Math.max(max, tagText.length);
  }, 0);

  if (maxTagTextLength === 0) {
    return 72;
  }

  return Math.min(Math.max(maxTagTextLength * 14 + 24, 120), 180);
});

const columns = computed(() => {
  return [
    {
      title: t("common.index"),
      dataIndex: "index",
      key: "index",
      width: 70,
    },
    {
      dataIndex: "name",
      title: t("common.name"),
      key: "name",
      width: 200,
      ellipsis: true,
    },
    {
      dataIndex: "tag_list",
      title: t("proxy.tags"),
      key: "tag_list",
      width: tagColumnWidth.value,
    },
    {
      title: t("common.type"),
      dataIndex: "type",
      key: "type",
      width: 90,
    },
    {
      title: t("proxy.listenAddress"),
      dataIndex: "listen_address",
      key: "listen_address",
      width: 140,
      ellipsis: true,
    },
    {
      title: t("proxy.listenPort"),
      dataIndex: "listen_port",
      key: "listen_port",
      width: 80,
    },
    {
      title: t("proxy.targetAddress"),
      key: "target_address",
      dataIndex: "target_address",
      width: 140,
      ellipsis: true,
    },
    {
      title: t("proxy.targetPort"),
      key: "target_port",
      dataIndex: "target_port",
      width: 100,
    },
    {
      title: t("common.status"),
      dataIndex: "state",
      key: "state",
      width: 100,
      // slots: { customRender: "state" },
    },
    {
      title: t("proxy.trafficInOut"),
      dataIndex: "traffic_in",
      key: "traffic_in",
      width: 170,
      // slots: { customRender: "state" },
    },
    {
      title: t("common.operation"),
      dataIndex: "operation",
      key: "operation",
      width: 260,
      // slots: { customRender: "operation" },
    },
  ];
});
/*****************表格******************* */

async function loadTags() {
  const res = await getTagList({});
  state.tagList = res.data || [];
}

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
      content: t("common.success"),
    });
  });
};
const startItem = (row: DataItem) => {
  startProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: t("common.success"),
    });
  });
};
const stopItem = (row: DataItem) => {
  stopProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: t("common.success"),
    });
  });
};
const restartItem = (row: DataItem) => {
  restartProxy({ name: row.name }).then(() => {
    getList();
    message.success({
      content: t("common.success"),
    });
  });
};

const captureItem = (row: any) => {
  router.push({
    path: "/proxyManage/capture",
    query: {
      task_uuid: row.uuid,
      name: row.name,
      type: row.type,
      state: row.state,
      listen_address: row.listen_address,
      listen_port: row.listen_port,
      target_address: row.target_address,
      target_port: row.target_port,
    },
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
    title: () => t("proxy.deleteSelectedConfirm"),
    icon: () => createVNode(ExclamationCircleOutlined),
    content: () => "",
    okText: () => t("common.confirm"),
    okType: "danger",
    cancelText: () => t("common.cancel"),
    onOk() {
      delBacthProxy({ uuid: selectedRowKeys.value }).then(() => {
        getList();
        message.success({
          content: t("common.success"),
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

  .m-table :deep(.ant-table-cell) {
    white-space: nowrap;
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

  .single-line-text {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
  }

  .tag-list-cell {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    :deep(.ant-tag) {
      margin-inline-end: 4px;
    }
  }

  .traf-span {
    display: inline-block;
    white-space: nowrap;
  }

  .operation-actions {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;

    :deep(.ant-btn) {
      padding-inline: 8px;
      white-space: nowrap;
    }
  }

  /* 监听端口链接样式 */
  .port-link {
    color: #1890ff;
    text-decoration: none;
    cursor: pointer;
    &:hover {
      color: #40a9ff;
      text-decoration: underline;
    }
  }

  :deep(.quick-access-form-item .ant-form-item-label) {
    flex: 0 0 112px;
    white-space: nowrap;
  }

  :deep(.quick-access-form-item .ant-form-item-label > label) {
    white-space: nowrap;
  }
}
</style>
