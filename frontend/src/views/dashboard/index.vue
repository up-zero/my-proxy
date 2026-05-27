<template>
  <div class="dashboard-page">
    <a-spin :spinning="state.loading" class="dashboard-spin">
      <div class="dashboard-summary-grid">
        <div v-for="item in summaryCards" :key="item.key" class="summary-card">
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-sub">{{ item.sub }}</div>
        </div>
      </div>

      <div class="dashboard-body">
        <div class="dashboard-main-column">
          <section class="panel-card traffic-card">
            <div class="panel-header">
              <div>
                <h3>实时速率</h3>
                <p>入站 / 出站速率折线图</p>
              </div>
              <div class="panel-highlight">
                <span>{{ formatRate(state.overview?.summary.inbound_rate || 0) }}</span>
                <em>入站</em>
                <span>{{ formatRate(state.overview?.summary.outbound_rate || 0) }}</span>
                <em>出站</em>
              </div>
            </div>
            <MetricLineChart
              :labels="trafficLabels"
              :series="trafficSeries"
              :valueFormatter="formatRate"
              :yAxisFormatter="formatTrafficValue"
            />
            <div class="chart-legend">
              <span><i class="dot inbound"></i>入站速率</span>
              <span><i class="dot outbound"></i>出站速率</span>
            </div>
          </section>

          <div class="dashboard-chart-row">
            <section class="panel-card compact-card">
              <div class="panel-header">
                <div>
                  <h3>连接数</h3>
                  <p>当前活动连接走势</p>
                </div>
                <div class="panel-metric">{{ formatNumber(state.overview?.summary.total_connections || 0) }}</div>
              </div>
              <MetricLineChart :labels="connectionLabels" :series="connectionSeries" />
            </section>

            <section class="panel-card compact-card">
              <div class="panel-header">
                <div>
                  <h3>系统资源</h3>
                  <p>CPU / 内存折线图</p>
                </div>
                <div class="panel-metric">{{ formatPercent(state.overview?.system.cpu_percent || 0) }}</div>
              </div>
              <MetricLineChart
                :labels="systemLabels"
                :series="systemSeries"
                :valueFormatter="formatPercent"
                :yAxisFormatter="formatPercent"
              />
            </section>
          </div>
        </div>

        <div class="dashboard-side-column">
          <section class="panel-card node-card">
            <div class="panel-header">
              <div>
                <h3>节点负载 Top 6</h3>
                <p>按连接数与实时速率综合排序</p>
              </div>
              <div class="panel-metric small">{{ updatedAtText }}</div>
            </div>
            <div class="node-list">
              <div v-if="!nodeList.length" class="empty-node">暂无代理节点</div>
              <div v-for="(item, index) in nodeList" :key="item.uuid" class="node-item">
                <div class="node-rank">{{ index + 1 }}</div>
                <div class="node-body">
                  <div class="node-head">
                    <div class="node-title">
                      <strong>{{ item.name }}</strong>
                      <a-tag size="small">{{ item.type }}</a-tag>
                    </div>
                    <span :class="['node-state', item.state === 'RUNNING' ? 'running' : 'stopped']">
                      {{ item.state === 'RUNNING' ? '运行中' : '已停止' }}
                    </span>
                  </div>
                  <div class="node-meta">
                    <span>{{ item.tag_list?.length ? item.tag_list.map((tag) => tag.name).join(' / ') : '未打标签' }}</span>
                    <span>{{ item.listen_address }}:{{ item.listen_port }}</span>
                  </div>
                  <div class="node-stats">
                    <span>连接数 {{ formatNumber(item.active_connections) }}</span>
                    <span>入 {{ formatRate(item.inbound_rate) }}</span>
                    <span>出 {{ formatRate(item.outbound_rate) }}</span>
                    <span>总 {{ formatBytes(item.traffic_out) }} / {{ formatBytes(item.traffic_in) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script lang="ts" setup>
import { getDashboardOverview } from "@/api/dashboard";
import MetricLineChart from "@/components/dashboard/MetricLineChart.vue";
import { errorHandle } from "@/lib/error";
import { computed, onMounted, onUnmounted, reactive } from "vue";

interface SummaryMetrics {
  proxy_total: number;
  proxy_running: number;
  proxy_stopped: number;
  total_connections: number;
  total_traffic_in: number;
  total_traffic_out: number;
  inbound_rate: number;
  outbound_rate: number;
  uptime_seconds: number;
  updated_at: number;
}

interface SystemMetrics {
  cpu_percent: number;
  memory_percent: number;
  memory_used: number;
  memory_total: number;
  go_memory_alloc: number;
  goroutines: number;
  sample_interval_ms: number;
}

interface TrafficPoint {
  timestamp: number;
  inbound_rate: number;
  outbound_rate: number;
}

interface ConnectionPoint {
  timestamp: number;
  connections: number;
}

interface SystemPoint {
  timestamp: number;
  cpu_percent: number;
  memory_percent: number;
}

interface NodeLoadMetric {
  uuid: string;
  name: string;
  tag_list: Array<{ uuid: string; name: string }>;
  type: string;
  state: string;
  listen_address: string;
  listen_port: string;
  active_connections: number;
  traffic_in: number;
  traffic_out: number;
  inbound_rate: number;
  outbound_rate: number;
}

interface DashboardOverview {
  summary: SummaryMetrics;
  system: SystemMetrics;
  charts: {
    traffic: TrafficPoint[];
    connections: ConnectionPoint[];
    system: SystemPoint[];
  };
  nodes: NodeLoadMetric[];
}

const POLL_INTERVAL = 3000;

const state = reactive({
  loading: false,
  overview: null as DashboardOverview | null,
});

let timer: number | null = null;

const summaryCards = computed(() => {
  const summary = state.overview?.summary;
  return [
    {
      key: "proxy_total",
      label: "代理总数",
      value: formatNumber(summary?.proxy_total || 0),
      sub: `运行 ${formatNumber(summary?.proxy_running || 0)} / 停止 ${formatNumber(summary?.proxy_stopped || 0)}`,
    },
    {
      key: "connections",
      label: "当前连接数",
      value: formatNumber(summary?.total_connections || 0),
      sub: "活跃 TCP / UDP / HTTP 请求",
    },
    {
      key: "traffic_in",
      label: "累计入站",
      value: formatBytes(summary?.total_traffic_in || 0),
      sub: `实时 ${formatRate(summary?.inbound_rate || 0)}`,
    },
    {
      key: "traffic_out",
      label: "累计出站",
      value: formatBytes(summary?.total_traffic_out || 0),
      sub: `实时 ${formatRate(summary?.outbound_rate || 0)}`,
    },
    {
      key: "cpu",
      label: "CPU",
      value: formatPercent(state.overview?.system.cpu_percent || 0),
      sub: `内存 ${formatPercent(state.overview?.system.memory_percent || 0)}`,
    },
    {
      key: "uptime",
      label: "服务运行时长",
      value: formatUptime(summary?.uptime_seconds || 0),
      sub: updatedAtText.value,
    },
  ];
});

const trafficLabels = computed(() =>
  (state.overview?.charts.traffic || []).map((item) => formatChartTime(item.timestamp))
);
const trafficSeries = computed(() => [
  {
    name: "入站速率",
    color: "#1677ff",
    values: (state.overview?.charts.traffic || []).map((item) => item.inbound_rate),
  },
  {
    name: "出站速率",
    color: "#52c41a",
    values: (state.overview?.charts.traffic || []).map((item) => item.outbound_rate),
  },
]);

const connectionLabels = computed(() =>
  (state.overview?.charts.connections || []).map((item) => formatChartTime(item.timestamp))
);
const connectionSeries = computed(() => [
  {
    name: "连接数",
    color: "#fa8c16",
    values: (state.overview?.charts.connections || []).map((item) => item.connections),
  },
]);

const systemLabels = computed(() =>
  (state.overview?.charts.system || []).map((item) => formatChartTime(item.timestamp))
);
const systemSeries = computed(() => [
  {
    name: "CPU",
    color: "#1677ff",
    values: (state.overview?.charts.system || []).map((item) => item.cpu_percent),
  },
  {
    name: "内存",
    color: "#722ed1",
    values: (state.overview?.charts.system || []).map((item) => item.memory_percent),
  },
]);

const updatedAtText = computed(() => {
  const timestamp = state.overview?.summary.updated_at;
  return timestamp ? `更新于 ${new Date(timestamp).toLocaleTimeString("zh-CN", { hour12: false })}` : "等待采样";
});

const nodeList = computed(() => state.overview?.nodes || []);

function formatBytes(bytes: number) {
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let value = Math.max(bytes, 0);
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${value.toFixed(2)}${units[unitIndex]}`;
}

function formatRate(value: number) {
  return `${formatTrafficValue(value)}/s`;
}

function formatTrafficValue(value: number) {
  return formatBytes(value);
}

function formatPercent(value: number) {
  return `${value.toFixed(2)}%`;
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}

function formatChartTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour12: false,
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatUptime(totalSeconds: number) {
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  if (days > 0) {
    return `${days}天 ${hours}时`;
  }
  if (hours > 0) {
    return `${hours}时 ${minutes}分`;
  }
  return `${minutes}分 ${totalSeconds % 60}秒`;
}

async function loadOverview(showError = true) {
  try {
    state.loading = !state.overview;
    const res = await getDashboardOverview();
    state.overview = res.data || null;
  } catch (err) {
    if (showError) {
      errorHandle(err, "load dashboard overview failed");
    }
  } finally {
    state.loading = false;
  }
}

function startPolling() {
  stopPolling();
  timer = window.setInterval(() => {
    loadOverview(false);
  }, POLL_INTERVAL);
}

function stopPolling() {
  if (timer !== null) {
    window.clearInterval(timer);
    timer = null;
  }
}

onMounted(async () => {
  await loadOverview();
  startPolling();
});

onUnmounted(() => {
  stopPolling();
});
</script>

<style scoped lang="less">
.dashboard-page {
  height: 100%;
  min-height: 0;
  overflow: hidden;

  :deep(.ant-spin-nested-loading),
  :deep(.ant-spin-container) {
    height: 100%;
  }

  :deep(.ant-spin-container) {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
  }

  .dashboard-spin {
    height: 100%;
  }

  .dashboard-summary-grid {
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 16px;
    flex-shrink: 0;
  }

  .summary-card,
  .panel-card {
    border-radius: 18px;
    background: #ffffff;
    box-shadow: 0 10px 40px rgba(15, 23, 42, 0.06);
  }

  .summary-card {
    min-height: 108px;
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  .summary-label {
    font-size: 13px;
    color: #667085;
  }

  .summary-value {
    font-size: 30px;
    font-weight: 700;
    color: #101828;
    line-height: 1.1;
  }

  .summary-sub {
    color: #98a2b3;
    font-size: 12px;
  }

  .dashboard-body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(0, 1.7fr) minmax(360px, 1fr);
    gap: 16px;
  }

  .dashboard-main-column,
  .dashboard-side-column {
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .dashboard-main-column {
    gap: 16px;
  }

  .traffic-card {
    flex: 1.2;
  }

  .dashboard-chart-row {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
  }

  .compact-card,
  .node-card {
    min-height: 0;
  }

  .node-card {
    flex: 1;
  }

  .panel-card {
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 14px;
    gap: 16px;

    h3 {
      margin: 0;
      font-size: 18px;
      color: #101828;
    }

    p {
      margin: 4px 0 0;
      color: #98a2b3;
      font-size: 12px;
    }
  }

  .panel-highlight {
    display: grid;
    grid-template-columns: auto auto;
    column-gap: 12px;
    row-gap: 4px;
    align-items: center;
    justify-items: end;

    span {
      font-weight: 700;
      color: #101828;
    }

    em {
      color: #98a2b3;
      font-size: 12px;
      font-style: normal;
      justify-self: start;
    }
  }

  .panel-metric {
    font-size: 28px;
    font-weight: 700;
    color: #101828;
    line-height: 1;

    &.small {
      font-size: 13px;
      font-weight: 500;
      color: #98a2b3;
      line-height: 1.4;
    }
  }

  .chart-legend {
    display: flex;
    gap: 16px;
    margin-top: 10px;
    color: #667085;
    font-size: 12px;

    .dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      margin-right: 6px;
      vertical-align: middle;
    }

    .inbound {
      background: #1677ff;
    }

    .outbound {
      background: #52c41a;
    }

    .cpu {
      background: #1677ff;
    }

    .memory {
      background: #722ed1;
    }
  }

  .node-list {
    flex: 1;
    display: grid;
    gap: 12px;
    min-height: 0;
    overflow: hidden;
    grid-auto-rows: minmax(0, 1fr);
  }

  .empty-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #98a2b3;
    border-radius: 14px;
    background: #f8fafc;
  }

  .node-item {
    display: flex;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 16px;
    background: #f8fafc;
    min-width: 0;
    min-height: 0;
  }

  .node-rank {
    width: 28px;
    height: 28px;
    border-radius: 999px;
    background: linear-gradient(135deg, #1677ff, #69b1ff);
    color: #fff;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .node-body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    justify-content: center;
  }

  .node-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
  }

  .node-title {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;

    strong {
      color: #101828;
      font-size: 15px;
      max-width: 180px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .node-state {
    font-size: 12px;
    padding: 2px 10px;
    border-radius: 999px;
    flex-shrink: 0;

    &.running {
      color: #039855;
      background: rgba(18, 183, 106, 0.12);
    }

    &.stopped {
      color: #d92d20;
      background: rgba(217, 45, 32, 0.12);
    }
  }

  .node-meta,
  .node-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 14px;
    color: #667085;
    font-size: 12px;
  }
}
</style>


