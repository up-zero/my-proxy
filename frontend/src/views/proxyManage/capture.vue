<template>
  <div class="capture-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <a-button @click="goBack">返回列表</a-button>
        <a-button
          type="primary"
          :danger="state.connected || state.connecting"
          :loading="state.connecting"
          @click="toggleConnection"
        >
          {{ state.connected || state.connecting ? "停止抓包" : "开始抓包" }}
        </a-button>
        <a-button @click="clearPackets" :disabled="!packets.length">清空数据</a-button>
        <a-switch v-model:checked="state.autoScroll" checked-children="自动滚动" un-checked-children="手动滚动" />
      </div>
      <div class="toolbar-right">
        <a-tag :color="connectionColor">{{ connectionText }}</a-tag>
        <span v-if="state.reconnectScheduled" class="reconnect-tip">
          {{ Math.ceil(state.reconnectDelayMs / 1000) }}s 后第 {{ state.reconnectAttempts }} 次重连
        </span>
      </div>
    </div>

    <div class="capture-main">
      <a-alert
        v-if="state.authExpired"
        type="error"
        show-icon
        message="登录已过期，请重新登录后重新开启抓包。"
        style="margin-bottom: 16px"
      />
      <a-alert
        v-else-if="state.errorMessage"
        type="warning"
        show-icon
        :message="state.errorMessage"
        style="margin-bottom: 16px"
      />

      <div class="capture-spin-wrap">
        <a-spin class="capture-spin-root" :spinning="state.loading">
          <div class="capture-body">
      <div class="summary-grid">
        <div class="summary-card">
          <div class="summary-title">任务信息</div>
          <div class="summary-lines">
            <div class="summary-line">
              <span class="label">代理名称</span>
              <span class="value">{{ state.task?.name || routeTask.name || "-" }}</span>
            </div>
            <div class="summary-line">
              <span class="label">代理类型</span>
              <span class="value">{{ state.task?.type || routeTask.type || "-" }}</span>
            </div>
          </div>
        </div>
        <div class="summary-card">
          <div class="summary-title">地址信息</div>
          <div class="summary-lines">
            <div class="summary-line">
              <span class="label">监听地址</span>
              <span class="value">{{ listenEndpoint }}</span>
            </div>
            <div class="summary-line">
              <span class="label">目标地址</span>
              <span class="value">{{ targetEndpoint }}</span>
            </div>
          </div>
        </div>
        <div class="summary-card">
          <div class="summary-title">抓包概览</div>
          <div class="stat-inline-grid">
            <div class="stat-inline-item">
              <span class="stat-inline-label">已收包</span>
              <span class="stat-inline-value">{{ state.totalPackets }}</span>
            </div>
            <div class="stat-inline-item">
              <span class="stat-inline-label">累计字节</span>
              <span class="stat-inline-value">{{ formatBytes(state.totalBytes) }}</span>
            </div>
            <div class="stat-inline-item">
              <span class="stat-inline-label">自动滚动</span>
              <span class="stat-inline-value">{{ state.autoScroll ? "开启" : "关闭" }}</span>
            </div>
            <div class="stat-inline-item">
              <span class="stat-inline-label">数据保留</span>
              <span class="stat-inline-value">最近 {{ MAX_PACKET_COUNT }} 条</span>
            </div>
          </div>
        </div>
      </div>

      <div class="content-grid">
        <div class="packet-list-card">
          <div class="panel-header">
            <span>抓包列表</span>
            <span class="panel-desc">按时间倒序展示最新数据</span>
          </div>
          <div v-if="!packets.length" class="empty-wrap">
            <a-empty description="暂无抓包数据" />
          </div>
          <div v-else ref="listRef" class="packet-list-scroll">
            <div class="packet-list">
              <div
                v-for="item in packets"
                :key="item.id"
                class="packet-item"
                :class="{ active: selectedPacket?.id === item.id }"
                @click="selectPacket(item)"
              >
                <div class="packet-item-head">
                  <div>
                    <a-tag :color="item.direction === 'IN' ? 'green' : 'blue'">{{ item.direction }}</a-tag>
                    <a-tag>{{ item.protocol }}</a-tag>
                    <span class="packet-time">{{ parseTime(item.timestamp) }}</span>
                  </div>
                  <span class="packet-size">{{ formatBytes(item.size) }}</span>
                </div>
                <div class="packet-preview">{{ item.previewLine }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="packet-detail-card">
          <div class="panel-header">
            <span>数据详情</span>
            <span class="panel-desc">支持文本预览和 HEX 查看</span>
          </div>
          <div v-if="!selectedPacket" class="empty-wrap">
            <a-empty description="请选择一条抓包记录" />
          </div>
          <template v-else>
            <div class="detail-meta">
              <div><strong>时间：</strong>{{ parseTime(selectedPacket.timestamp) }}</div>
              <div><strong>方向：</strong>{{ selectedPacket.direction }}</div>
              <div><strong>协议：</strong>{{ selectedPacket.protocol }}</div>
              <div><strong>大小：</strong>{{ formatBytes(selectedPacket.size) }}</div>
            </div>
            <a-tabs>
              <a-tab-pane key="text" tab="文本预览">
                <pre class="payload-pre">{{ selectedPacket.previewText || "(二进制或不可见字符较多，建议查看 HEX)" }}</pre>
              </a-tab-pane>
              <a-tab-pane key="hex" tab="HEX">
                <pre class="payload-pre">{{ selectedPacket.formattedHex }}</pre>
              </a-tab-pane>
            </a-tabs>
          </template>
        </div>
      </div>
          </div>
        </a-spin>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getProxyStatus } from "@/api/proxy";
import { errorHandle } from "@/lib/error";
import { buildCaptureWebSocketUrl } from "@/lib/request";
import { parseTime, toast } from "@/lib/util";
import { message } from "ant-design-vue";
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { onBeforeRouteLeave, useRoute, useRouter } from "vue-router";

interface ProxyTaskInfo {
  uuid: string;
  name: string;
  type: string;
  state: string;
  listen_address: string;
  listen_port: string;
  target_address: string;
  target_port: string;
}

interface PacketMessage {
  task_uuid: string;
  timestamp: number;
  direction: string;
  protocol: string;
  payload: string;
}

interface PacketView extends PacketMessage {
  id: string;
  size: number;
  previewText: string;
  previewLine: string;
  formattedHex: string;
}

const MAX_PACKET_COUNT = 500;
const RECONNECT_BASE_DELAY = 1000;
const RECONNECT_MAX_DELAY = 10000;
const AUTH_CLOSE_CODE = 1008;

const route = useRoute();
const router = useRouter();
const listRef = ref<HTMLDivElement>();
const packets = ref<PacketView[]>([]);
const selectedPacket = ref<PacketView | null>(null);

const routeTask = computed(() => ({
  uuid: String(route.query.task_uuid || ""),
  name: String(route.query.name || ""),
  type: String(route.query.type || ""),
  state: String(route.query.state || ""),
  listenAddress: String(route.query.listen_address || ""),
  listenPort: String(route.query.listen_port || ""),
  targetAddress: String(route.query.target_address || ""),
  targetPort: String(route.query.target_port || ""),
}));

const state = reactive({
  loading: false,
  connecting: false,
  connected: false,
  autoScroll: true,
  authExpired: false,
  reconnectScheduled: false,
  reconnectAttempts: 0,
  reconnectDelayMs: 0,
  totalPackets: 0,
  totalBytes: 0,
  trimmedPackets: 0,
  errorMessage: "",
  lastCloseReason: "",
  task: null as ProxyTaskInfo | null,
});

let captureSocket: WebSocket | null = null;
let reconnectTimer: number | null = null;
let packetSeq = 0;
let pageDestroyed = false;
let manualClose = false;
let authToastShown = false;

const connectionText = computed(() => {
  if (state.authExpired) {
    return "登录已过期";
  }
  if (state.connected) {
    return "抓包中";
  }
  if (state.connecting) {
    return "连接中";
  }
  if (state.reconnectScheduled) {
    return "等待重连";
  }
  return "未连接";
});

const connectionColor = computed(() => {
  if (state.authExpired) {
    return "red";
  }
  if (state.connected) {
    return "green";
  }
  if (state.connecting || state.reconnectScheduled) {
    return "orange";
  }
  return "default";
});

function normalizeListenAddress(address: string) {
  const normalized = address.trim();
  if (!normalized || normalized === "127.0.0.1" || normalized.toLowerCase() === "localhost") {
    return "0.0.0.0";
  }
  return normalized;
}

const listenEndpoint = computed(() => {
  const address = normalizeListenAddress(state.task?.listen_address || routeTask.value.listenAddress || "");
  const port = state.task?.listen_port || routeTask.value.listenPort || "";
  return port ? `${address}:${port}` : address;
});

const targetEndpoint = computed(() => {
  const address = state.task?.target_address || routeTask.value.targetAddress || "-";
  const port = state.task?.target_port || routeTask.value.targetPort || "";
  if (address === "-") {
    return address;
  }
  return port ? `${address}:${port}` : address;
});

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)}KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)}MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)}GB`;
}

function normalizeText(value: string) {
  return value.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, ".");
}

function hexToBytes(hex: string): Uint8Array {
  const normalized = hex.replace(/\s+/g, "");
  if (!normalized || normalized.length % 2 !== 0) {
    return new Uint8Array();
  }

  const bytes = new Uint8Array(normalized.length / 2);
  for (let i = 0; i < normalized.length; i += 2) {
    const value = Number.parseInt(normalized.slice(i, i + 2), 16);
    if (Number.isNaN(value)) {
      return new Uint8Array();
    }
    bytes[i / 2] = value;
  }
  return bytes;
}

function formatHex(hex: string) {
  const normalized = hex.replace(/\s+/g, "");
  const segments: string[] = [];
  for (let i = 0; i < normalized.length; i += 32) {
    const chunk = normalized.slice(i, i + 32);
    segments.push(chunk.match(/.{1,2}/g)?.join(" ") || chunk);
  }
  return segments.join("\n");
}

function decodePreviewText(hex: string) {
  const bytes = hexToBytes(hex);
  if (!bytes.length) {
    return { size: 0, previewText: "", previewLine: "" };
  }

  const decoder = new TextDecoder("utf-8", { fatal: false });
  const text = normalizeText(decoder.decode(bytes));
  const previewLine = text.replace(/\s+/g, " ").slice(0, 160) || "(空内容)";

  return {
    size: bytes.length,
    previewText: text,
    previewLine,
  };
}

function pushPacket(packet: PacketView) {
  packets.value.unshift(packet);
  state.totalPackets += 1;
  state.totalBytes += packet.size;

  if (packets.value.length > MAX_PACKET_COUNT) {
    packets.value.splice(MAX_PACKET_COUNT);
    state.trimmedPackets += 1;
  }

  selectedPacket.value = packet;

  if (state.autoScroll) {
    nextTick(() => {
      listRef.value?.scrollTo({ top: 0, behavior: "smooth" });
    });
  }
}

function selectPacket(packet: PacketView) {
  selectedPacket.value = packet;
}

function clearPackets() {
  packets.value = [];
  selectedPacket.value = null;
  state.totalPackets = 0;
  state.totalBytes = 0;
  state.trimmedPackets = 0;
}

function clearReconnectTimer() {
  if (reconnectTimer !== null) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  state.reconnectScheduled = false;
  state.reconnectDelayMs = 0;
}

function isAuthClose(event: CloseEvent) {
  return event.code === AUTH_CLOSE_CODE || /登录过期|auth/i.test(event.reason || "");
}

async function redirectToLogin() {
  if (authToastShown) {
    return;
  }
  authToastShown = true;
  await toast("登录过期，请重新登录", "error", 1.5);
  router.replace("/login");
}

function closeSocket(reason = "manual close") {
  clearReconnectTimer();
  state.connecting = false;

  if (!captureSocket) {
    state.connected = false;
    return;
  }

  const ws = captureSocket;
  captureSocket = null;
  state.connected = false;

  try {
    ws.close(1000, reason);
  } catch (err) {
    console.warn("capture websocket close failed", err);
  }
}

function cleanupPageConnection() {
  manualClose = true;
  closeSocket("page leave");
}

async function refreshTaskInfo(showError = true) {
  const taskUuid = routeTask.value.uuid;
  if (!taskUuid) {
    state.errorMessage = "缺少 task_uuid，无法开启抓包。";
    return "missing_task_uuid";
  }

  try {
    state.loading = true;
    const res = await getProxyStatus({ uuid: taskUuid });
    const task = Array.isArray(res.data) ? res.data[0] : null;
    state.task = task || null;

    if (!task) {
      state.errorMessage = "抓包任务不存在或已被删除。";
      return "not_found";
    }

    if (task.state !== "RUNNING") {
      state.errorMessage = "当前代理未运行，无法建立抓包连接。";
      return "not_running";
    }

    state.errorMessage = "";
    return "ok";
  } catch (err) {
    if (showError) {
      errorHandle(err, "load capture task failed");
    }
    state.errorMessage = "获取代理状态失败，请稍后重试。";
    return "request_error";
  } finally {
    state.loading = false;
  }
}

function scheduleReconnect() {
  if (pageDestroyed || manualClose || state.authExpired) {
    return;
  }

  clearReconnectTimer();
  state.reconnectAttempts += 1;
  state.reconnectDelayMs = Math.min(
    RECONNECT_BASE_DELAY * 2 ** Math.max(state.reconnectAttempts - 1, 0),
    RECONNECT_MAX_DELAY
  );
  state.reconnectScheduled = true;

  reconnectTimer = window.setTimeout(async () => {
    reconnectTimer = null;
    state.reconnectScheduled = false;
    await connectCapture(true);
  }, state.reconnectDelayMs);
}

function handlePacketMessage(raw: string) {
  try {
    const data = JSON.parse(raw) as PacketMessage;
    const preview = decodePreviewText(data.payload || "");

    pushPacket({
      ...data,
      id: `${Date.now()}-${packetSeq++}`,
      size: preview.size,
      previewText: preview.previewText,
      previewLine: preview.previewLine,
      formattedHex: formatHex(data.payload || ""),
    });
  } catch (err) {
    console.warn("capture packet parse failed", err);
  }
}

function bindSocketEvents(ws: WebSocket) {
  ws.onopen = () => {
    if (captureSocket !== ws) {
      ws.close();
      return;
    }

    state.connecting = false;
    state.connected = true;
    state.reconnectScheduled = false;
    state.reconnectDelayMs = 0;
    state.reconnectAttempts = 0;
    state.lastCloseReason = "";
    state.errorMessage = "";
  };

  ws.onmessage = (event) => {
    if (captureSocket !== ws) {
      return;
    }
    if (typeof event.data === "string") {
      handlePacketMessage(event.data);
    }
  };

  ws.onerror = () => {
    state.connecting = false;
  };

  ws.onclose = async (event) => {
    if (captureSocket !== ws) {
      return;
    }

    captureSocket = null;

    state.connecting = false;
    state.connected = false;
    state.lastCloseReason = event.reason || `连接已关闭(code=${event.code})`;

    if (isAuthClose(event)) {
      state.authExpired = true;
      state.errorMessage = "登录已过期，抓包连接已关闭。";
      clearReconnectTimer();
      await redirectToLogin();
      return;
    }

    if (manualClose || pageDestroyed) {
      return;
    }

    state.errorMessage = "抓包连接已断开，正在尝试自动重连。";
    message.warning("抓包连接已断开，正在尝试重连");
    scheduleReconnect();
  };
}

async function connectCapture(isReconnect = false) {
  if (state.authExpired) {
    return;
  }
  if (captureSocket || state.connecting) {
    return;
  }

  const refreshState = await refreshTaskInfo(!isReconnect);
  if (refreshState !== "ok") {
    if (isReconnect && !state.authExpired && refreshState === "request_error") {
      scheduleReconnect();
    }
    return;
  }

  manualClose = false;
  state.connecting = true;
  state.errorMessage = isReconnect ? "正在重新连接抓包通道..." : "";

  const ws = new WebSocket(buildCaptureWebSocketUrl(routeTask.value.uuid));
  captureSocket = ws;
  bindSocketEvents(ws);
}

function disconnectCapture() {
  manualClose = true;
  state.errorMessage = "";
  closeSocket("manual disconnect");
}

function toggleConnection() {
  if (state.connected || state.connecting) {
    disconnectCapture();
  } else {
    connectCapture();
  }
}

async function initializePage() {
  clearReconnectTimer();
  closeSocket("reinitialize");
  clearPackets();
  state.authExpired = false;
  state.lastCloseReason = "";
  state.errorMessage = "";
  authToastShown = false;
  await connectCapture();
}

function goBack() {
  router.push("/proxyManage/index");
}

function handleWindowLeave() {
  cleanupPageConnection();
}

watch(
  () => routeTask.value.uuid,
  async (next, prev) => {
    if (next && next !== prev) {
      await initializePage();
    }
  }
);

onBeforeRouteLeave(() => {
  cleanupPageConnection();
});

onMounted(async () => {
  window.addEventListener("beforeunload", handleWindowLeave);
  window.addEventListener("pagehide", handleWindowLeave);
  await initializePage();
});

onUnmounted(() => {
  pageDestroyed = true;
  window.removeEventListener("beforeunload", handleWindowLeave);
  window.removeEventListener("pagehide", handleWindowLeave);
  cleanupPageConnection();
});
</script>

<style scoped lang="less">
.capture-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
  overflow: hidden;

  .capture-main {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .capture-spin-wrap {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .capture-spin-root {
    flex: 1;
    min-height: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;

    .toolbar-left,
    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 12px;
      flex-wrap: wrap;
    }

    .reconnect-tip {
      color: rgba(0, 0, 0, 0.45);
      font-size: 12px;
    }
  }

  .capture-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    flex-shrink: 0;
  }

  .summary-card,
  .packet-list-card,
  .packet-detail-card {
    background: #fff;
    border: 1px solid #f0f0f0;
    border-radius: 8px;
    padding: 16px;
  }

  .summary-title {
    font-size: 16px;
    font-weight: 700;
    color: rgba(0, 0, 0, 0.88);
    margin-bottom: 10px;
  }

  .summary-lines {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .summary-line {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: flex-start;

    .label {
      flex-shrink: 0;
      color: rgba(0, 0, 0, 0.45);
      font-size: 12px;
    }

    .value {
      text-align: right;
      font-size: 13px;
      font-weight: 500;
      word-break: break-all;
    }
  }

  .stat-inline-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px 12px;
  }

  .stat-inline-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;

  }

  .stat-inline-label {
    font-size: 12px;
    color: rgba(0, 0, 0, 0.45);
  }

  .stat-inline-value {
    font-size: 13px;
    font-weight: 500;
    min-width: 0;
    word-break: break-all;
  }

  .content-grid {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(320px, 1fr) minmax(420px, 1.2fr);
    gap: 16px;
  }

  .packet-list-card,
  .packet-detail-card {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
    font-weight: 600;

    .panel-desc {
      font-size: 12px;
      color: rgba(0, 0, 0, 0.45);
      font-weight: 400;
    }
  }

  .packet-list-scroll {
    //flex: 1;
    min-height: 0;
    height: 50vh;
    overflow-y: auto;
    scrollbar-gutter: stable;
  }

  .packet-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 4px;
  }

  .packet-item {
    border: 1px solid #f0f0f0;
    border-radius: 6px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover,
    &.active {
      border-color: #1677ff;
      background: #f0f7ff;
    }

    .packet-item-head {
      display: flex;
      justify-content: space-between;
      gap: 8px;
      align-items: center;
      margin-bottom: 8px;
    }

    .packet-time,
    .packet-size {
      font-size: 12px;
      color: rgba(0, 0, 0, 0.45);
    }

    .packet-preview {
      font-family: Consolas, Monaco, monospace;
      font-size: 12px;
      line-height: 1.6;
      color: rgba(0, 0, 0, 0.75);
      word-break: break-all;
      white-space: pre-wrap;
    }
  }

  .detail-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 10px;
    margin-bottom: 12px;
    font-size: 13px;
    flex-shrink: 0;
  }

  .payload-pre {
    min-height: 280px;
    max-height: calc(100vh - 530px);
    overflow: auto;
    scrollbar-gutter: stable;
    background: #0f172a;
    color: #e2e8f0;
    border-radius: 6px;
    padding: 12px;
    font-family: Consolas, Monaco, monospace;
    font-size: 12px;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .empty-wrap {
    flex: 1;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

@media (max-width: 1200px) {
  .capture-page {
    .summary-grid {
      grid-template-columns: 1fr;
    }

    .content-grid {
      grid-template-columns: 1fr;
    }
  }
}
</style>












