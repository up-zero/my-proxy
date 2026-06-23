<template>
  <div class="terminal-page">
    <!-- 终端标签页 -->
    <div v-if="tabs.length > 0" class="terminal-tabs-wrap">
      <div class="terminal-tabs">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          :class="['terminal-tab', { active: activeTabId === tab.id }]"
          @click="switchTab(tab.id)"
        >
          <span class="tab-label">{{ tab.name }}</span>
          <span :class="['tab-status', tab.status]"></span>
          <CloseOutlined
            class="tab-close"
            @click.stop="closeTab(tab.id)"
          />
        </div>
        <!-- 新建终端标签 -->
        <div class="terminal-tab new-tab-btn" @click="showAddDialog = true">
          <PlusOutlined class="new-tab-icon" />
          <span>{{ t("terminal.newTab") }}</span>
        </div>
        <!-- 全局监控开关 -->
        <div class="terminal-tab monitor-toggle" @click.stop>
          <span class="toggle-label">{{ t("terminal.monitor") }}</span>
          <a-switch
            v-model:checked="monitorGlobalEnabled"
            size="small"
            @change="onMonitorToggle"
          />
        </div>
      </div>

      <!-- 终端内容区 -->
      <div class="terminal-content">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          :ref="(el) => setTerminalRef(tab.id, el as HTMLElement)"
          :class="['terminal-instance', { visible: activeTabId === tab.id }]"
        ></div>
      </div>

      <!-- 远程监控面板 -->
      <div v-if="monitorGlobalEnabled" class="monitor-panel">
        <!-- CPU with sparkline -->
        <div class="monitor-item monitor-cpu">
          <span class="monitor-label">CPU</span>
          <svg
            class="cpu-sparkline"
            viewBox="0 0 80 24"
            preserveAspectRatio="none"
          >
            <polyline
              v-if="cpuSparklinePoints"
              :points="cpuSparklinePoints"
              fill="none"
              stroke="#4ec9b0"
              stroke-width="1.5"
              stroke-linejoin="round"
              stroke-linecap="round"
            />
          </svg>
          <span class="monitor-value cpu-value">{{ cpuText }}</span>
        </div>

        <!-- MEM -->
        <div class="monitor-item">
          <span class="monitor-label">MEM</span>
          <div class="monitor-bar-wrap">
            <div
              class="monitor-bar mem"
              :style="{ width: memUsagePercent + '%' }"
            ></div>
          </div>
          <span class="monitor-value">{{ memText }}</span>
        </div>

        <!-- DISKS -->
        <div
          v-for="disk in activeDisks"
          :key="disk.mount"
          class="monitor-item"
        >
          <span class="monitor-label disk-label">DISK {{ disk.mount }}</span>
          <div class="monitor-bar-wrap">
            <div
              class="monitor-bar disk"
              :style="{ width: disk.percent + '%' }"
            ></div>
          </div>
          <span class="monitor-value">{{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="terminal-empty">
      <a-empty :description="t('terminal.noTerminal')">
        <a-button type="primary" @click="showAddDialog = true">
          <PlusOutlined /> {{ t("terminal.newTab") }}
        </a-button>
      </a-empty>
    </div>

    <!-- 新建终端弹窗 -->
    <a-modal
      v-model:open="showAddDialog"
      :title="t('terminal.newTerminal')"
      :footer="null"
      width="520px"
    >
      <a-tabs v-model:activeKey="addTabType">
        <a-tab-pane key="local" :tab="t('terminal.connectLocal')">
          <a-form layout="vertical" class="terminal-form">
            <a-form-item :label="t('terminal.sshPort')">
              <a-input
                v-model:value="localForm.port"
                :placeholder="t('terminal.inputSshPort')"
              />
            </a-form-item>
            <a-form-item :label="t('terminal.sshUsername')">
              <a-input
                v-model:value="localForm.username"
                :placeholder="t('terminal.inputSshUsername')"
              />
            </a-form-item>
            <a-form-item :label="t('terminal.sshPassword')">
              <a-input-password
                v-model:value="localForm.password"
                :placeholder="t('terminal.inputSshPassword')"
              />
            </a-form-item>
            <a-button type="primary" block @click="connectLocal">
              {{ t("terminal.connectLocal") }}
            </a-button>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="proxy" :tab="t('terminal.connectProxy')">
          <a-form layout="vertical" class="terminal-form">
            <a-form-item :label="t('terminal.selectProxy')">
              <a-select
                v-model:value="proxyForm.proxyUuid"
                :placeholder="t('terminal.selectProxy')"
                show-search
                option-filter-prop="label"
                :filter-option="filterProxyOption"
                @change="onProxySelect"
              >
                <a-select-option
                  v-for="p in proxyList"
                  :key="p.uuid"
                  :value="p.uuid"
                  :label="p.name"
                >
                  {{ p.name }} ({{ p.target_address }}:{{ p.target_port }})
                </a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item :label="t('terminal.sshPort')">
              <a-input
                v-model:value="proxyForm.port"
                :placeholder="t('terminal.inputSshPort')"
              />
            </a-form-item>
            <a-form-item :label="t('terminal.sshUsername')">
              <a-input
                v-model:value="proxyForm.username"
                :placeholder="t('terminal.inputSshUsername')"
              />
            </a-form-item>
            <a-form-item :label="t('terminal.sshPassword')">
              <a-input-password
                v-model:value="proxyForm.password"
                :placeholder="t('terminal.inputSshPassword')"
              />
            </a-form-item>
            <a-button
              type="primary"
              block
              :disabled="!proxyForm.proxyUuid"
              @click="connectProxy"
            >
              {{ t("terminal.connectProxy") }}
            </a-button>
          </a-form>
        </a-tab-pane>
      </a-tabs>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onBeforeUnmount, watch, computed } from "vue";
import { useAppI18n } from "@/i18n";
import { PlusOutlined, CloseOutlined } from "@ant-design/icons-vue";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { getProxyStatus } from "@/api/proxy";
import { createTerminalSocket } from "@/api/terminal";
import type { TerminalConnInfo } from "@/api/terminal";

const { t } = useAppI18n();

// ===================== 终端实例管理 =====================
interface TerminalTab {
  id: string;
  name: string;
  host: string;
  port: string;
  username: string;
  password: string;
  status: "connecting" | "connected" | "disconnected" | "error";
  term: Terminal | null;
  fitAddon: FitAddon | null;
  ws: WebSocket | null;
}

const tabs = ref<TerminalTab[]>([]);
const activeTabId = ref<string>("");
const showAddDialog = ref(false);
const addTabType = ref("local");
const terminalRefs = ref<Record<string, HTMLElement>>({});

function setTerminalRef(id: string, el: HTMLElement) {
  if (el) {
    terminalRefs.value[id] = el;
  }
}

// 本地连接表单
const localForm = reactive({
  port: "22",
  username: "root",
  password: "",
});

// 代理连接表单
const proxyForm = reactive({
  proxyUuid: "",
  port: "22",
  username: "root",
  password: "",
});

// 代理列表
const proxyList = ref<any[]>([]);

// ===================== 远程监控 =====================
const monitorGlobalEnabled = ref(false);

// 磁盘信息
interface DiskInfo {
  mount: string;
  total: number;
  used: number;
}

// 监控数据（用 reactive map 确保响应式更新）
interface MonitorSnapshot {
  cpu: number;
  mem_total: number;
  mem_used: number;
  disks: DiskInfo[];
}
const monitorDataMap = reactive<Record<string, MonitorSnapshot>>({});

// CPU 历史数据（最多存 60 个点）
const cpuHistoryMap = reactive<Record<string, number[]>>({});
const MAX_CPU_HISTORY = 60;

// 当前活跃标签的监控数据
const activeMonitorData = computed<MonitorSnapshot | null>(() => {
  return monitorDataMap[activeTabId.value] ?? null;
});

// 当前活跃标签的 CPU 历史
const activeCpuHistory = computed<number[]>(() => {
  return cpuHistoryMap[activeTabId.value] ?? [];
});

// CPU sparkline SVG points（纵轴上限固定 100%，更直观反映实际 CPU 负载）
const cpuSparklinePoints = computed(() => {
  const history = activeCpuHistory.value;
  if (history.length < 2) return "";
  const w = 80, h = 24, pad = 2;
  const max = 100;
  return history
    .map((v, i) => {
      const x = pad + (i / (history.length - 1)) * (w - pad * 2);
      const y = pad + (1 - v / max) * (h - pad * 2);
      return `${x},${y}`;
    })
    .join(" ");
});

// CPU 显示文本
const cpuText = computed(() => {
  const d = activeMonitorData.value;
  return d != null ? d.cpu.toFixed(1) + "%" : "--";
});

const memUsagePercent = computed(() => {
  const d = activeMonitorData.value;
  if (!d || d.mem_total === 0) return 0;
  return Math.min((d.mem_used / d.mem_total) * 100, 100);
});

const memText = computed(() => {
  const d = activeMonitorData.value;
  return d ? formatBytes(d.mem_used) + " / " + formatBytes(d.mem_total) : "--";
});

// 当前活跃标签的磁盘列表（带计算百分比）
const activeDisks = computed(() => {
  const d = activeMonitorData.value;
  if (!d || !d.disks) return [];
  return d.disks.map((dk) => ({
    ...dk,
    percent: dk.total > 0 ? Math.min((dk.used / dk.total) * 100, 100) : 0,
  }));
});

// 格式化字节数（自动选择 TB/GB/MB/KB/B 单位，保留一位小数）
function formatBytes(bytes: number): string {
  if (bytes >= 1099511627776) {
    return (bytes / 1099511627776).toFixed(1) + " TB";
  } else if (bytes >= 1073741824) {
    return (bytes / 1073741824).toFixed(1) + " GB";
  } else if (bytes >= 1048576) {
    return (bytes / 1048576).toFixed(1) + " MB";
  } else if (bytes >= 1024) {
    return (bytes / 1024).toFixed(1) + " KB";
  }
  return bytes + " B";
}

// 全局监控开关切换 — 仅对当前活跃标签生效，切换标签时自动关闭旧标签监控
function onMonitorToggle(checked: boolean) {
  if (checked) {
    // 只对当前活跃标签开启监控
    const activeTab = tabs.value.find(t => t.id === activeTabId.value);
    if (activeTab?.ws && activeTab.ws.readyState === WebSocket.OPEN) {
      activeTab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: true }));
    }
  } else {
    // 关闭所有标签的监控
    for (const tab of tabs.value) {
      if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
        tab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: false }));
      }
    }
  }
}

// ===================== SSH 凭据缓存 =====================
const STORAGE_KEY_LOCAL = "my-proxy:terminal:ssh:local";
const STORAGE_KEY_PROXY_PREFIX = "my-proxy:terminal:ssh:proxy:";

interface SshCredentials {
  port: string;
  username: string;
  password: string;
}

// 加载本地 SSH 凭据
function loadLocalSshCredentials(): SshCredentials | null {
  try {
    const data = localStorage.getItem(STORAGE_KEY_LOCAL);
    return data ? JSON.parse(data) : null;
  } catch {
    return null;
  }
}

// 保存本地 SSH 凭据
function saveLocalSshCredentials(port: string, username: string, password: string) {
  try {
    const credentials: SshCredentials = { port, username, password };
    localStorage.setItem(STORAGE_KEY_LOCAL, JSON.stringify(credentials));
  } catch (e) {
    console.error("Failed to save local SSH credentials:", e);
  }
}

// 加载代理 SSH 凭据
function loadProxySshCredentials(proxyUuid: string): SshCredentials | null {
  try {
    const key = `${STORAGE_KEY_PROXY_PREFIX}${proxyUuid}`;
    const data = localStorage.getItem(key);
    return data ? JSON.parse(data) : null;
  } catch {
    return null;
  }
}

// 保存代理 SSH 凭据
function saveProxySshCredentials(proxyUuid: string, port: string, username: string, password: string) {
  try {
    const credentials: SshCredentials = { port, username, password };
    const key = `${STORAGE_KEY_PROXY_PREFIX}${proxyUuid}`;
    localStorage.setItem(key, JSON.stringify(credentials));
  } catch (e) {
    console.error("Failed to save proxy SSH credentials:", e);
  }
}

// ===================== 获取代理列表 =====================
async function fetchProxyList() {
  try {
    const res = await getProxyStatus({});
    if (res.data) {
      proxyList.value = res.data;
    }
  } catch (e) {
    console.error("Failed to fetch proxy list:", e);
  }
}

// ===================== 创建终端连接 =====================
/**
 * 将 viewport 高度精确收拢到 rows×rowHeight，消除底部余数空白。
 * 双 rAF 确保在 xterm.js 内部布局完全结束后再调整；
 * MutationObserver 兜底，防止后续 xterm 内部重绘时覆写 inline style。
 */
function fitSnap(term: Terminal, fitAddon: FitAddon) {
  fitAddon.fit();
  const rows = term.rows;
  if (rows <= 0 || !term.element) return;

  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      _applyViewportSnap(term, rows);
      _ensureViewportGuard(term);
    });
  });
}

function _applyViewportSnap(term: Terminal, rows: number) {
  if (!term.element) return;
  const screen = term.element.querySelector('.xterm-screen') as HTMLElement;
  const viewport = term.element.querySelector('.xterm-viewport') as HTMLElement;
  if (!screen || !viewport || rows <= 0) return;
  const rowH = screen.offsetHeight / rows;
  if (rowH <= 0) return;
  const targetH = Math.round(rows * rowH);
  (term as any).__vpTargetHeight = targetH; // 缓存目标值，供 guard 回退
  viewport.style.bottom = 'auto';
  viewport.style.height = targetH + 'px';
}

/** MutationObserver 兜底：xterm 内部重绘覆写 viewport style 时自动恢复 */
function _ensureViewportGuard(term: Terminal) {
  if (!term.element) return;
  const viewport = term.element.querySelector('.xterm-viewport') as HTMLElement;
  if (!viewport) return;

  const old: MutationObserver | undefined = (term as any).__vpObserver;
  if (old) old.disconnect();

  const observer = new MutationObserver(() => {
    const expected = (term as any).__vpTargetHeight as number | undefined;
    if (
      expected &&
      viewport.style.height === expected + 'px' &&
      viewport.style.bottom === 'auto'
    ) {
      return; // 样式未被改动，跳过避免死循环
    }
    observer.disconnect();
    if (expected && term.element) {
      viewport.style.bottom = 'auto';
      viewport.style.height = expected + 'px';
    }
    observer.observe(viewport, { attributes: true, attributeFilter: ['style'] });
  });
  observer.observe(viewport, { attributes: true, attributeFilter: ['style'] });
  (term as any).__vpObserver = observer;
}

function _cleanupViewportGuard(term: Terminal) {
  const obs: MutationObserver | undefined = (term as any).__vpObserver;
  if (obs) obs.disconnect();
  delete (term as any).__vpObserver;
  delete (term as any).__vpTargetHeight;
}

function createTerminalInstance(tab: TerminalTab) {
  if (!terminalRefs.value[tab.id]) return;

  const term = new Terminal({
    cursorBlink: true,
    cursorStyle: "bar",
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: "#1e1e1e",
      foreground: "#d4d4d4",
      cursor: "#ffffff",
      selectionBackground: "#264f78",
      black: "#000000",
      red: "#cd3131",
      green: "#0dbc79",
      yellow: "#e5e510",
      blue: "#2472c8",
      magenta: "#bc3fbc",
      cyan: "#11a8cd",
      white: "#e5e5e5",
      brightBlack: "#666666",
      brightRed: "#f14c4c",
      brightGreen: "#23d18b",
      brightYellow: "#f5f543",
      brightBlue: "#3b8eea",
      brightMagenta: "#d670d6",
      brightCyan: "#29b8db",
      brightWhite: "#ffffff",
    },
    allowProposedApi: true,
  });

  const fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  // 等待 DOM 渲染
  nextTick(() => {
    const el = terminalRefs.value[tab.id];
    if (el) {
      term.open(el);
      
      // 立即执行一次 fit，确保初始尺寸正确
      let lastCols = 0;
      let lastRows = 0;
      setTimeout(() => {
        fitSnap(term, fitAddon);
        const { cols, rows } = term;
        lastCols = cols;
        lastRows = rows;
        // 发送初始尺寸到后端
        if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
          tab.ws.send(JSON.stringify({ type: "resize", size: { cols, rows } }));
        }
      }, 100);

      // 监听窗口大小变化（仅活跃标签同步尺寸，避免隐藏标签 resize 触发 shell 重复输出提示符）
      let resizeTimer: ReturnType<typeof setTimeout> | null = null;
      const observer = new ResizeObserver(() => {
        if (resizeTimer) clearTimeout(resizeTimer);
        resizeTimer = setTimeout(() => {
          // 仅活跃标签才 fit 并同步尺寸
          if (activeTabId.value !== tab.id) return;
          fitSnap(term, fitAddon);
          const { cols, rows } = term;
          // 尺寸未变化则跳过，避免无效的 PTY resize 导致 shell 重复输出提示符
          if (cols === lastCols && rows === lastRows) return;
          lastCols = cols;
          lastRows = rows;
          if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
            tab.ws.send(JSON.stringify({ type: "resize", size: { cols, rows } }));
          }
        }, 150);
      });
      observer.observe(el);

      // 存储 observer 以便清理
      (term as any)._resizeObserver = observer;
    }
  });

  tab.term = term;
  tab.fitAddon = fitAddon;

  // 终端输入 -> WebSocket
  term.onData((data) => {
    // 断连/异常时按回车触发重连
    if (
      (tab.status === "disconnected" || tab.status === "error") &&
      data === "\r"
    ) {
      reconnectTab(tab.id);
      return;
    }
    if (tab.ws && tab.ws.readyState === WebSocket.OPEN) {
      tab.ws.send(JSON.stringify({ type: "data", data }));
    }
  });
}

// ===================== 建立 WebSocket 连接 =====================
function connectTab(tab: TerminalTab) {
  if (!tab.term) {
    createTerminalInstance(tab);
  }

  tab.status = "connecting";
  if (tab.term) {
    tab.term.write(`\r\n🔌 ${t("terminal.connecting")}\r\n`);
  }

  const info: TerminalConnInfo = {
    id: tab.id,
    name: tab.name,
    host: tab.host,
    port: tab.port,
    username: tab.username,
    password: tab.password,
  };

  try {
    const ws = createTerminalSocket(info);
    tab.ws = ws;

    ws.onopen = () => {
      tab.status = "connected";
      if (tab.term) {
        tab.term.clear();
        fitSnap(tab.term, tab.fitAddon!);
        // 连接成功后发送当前终端尺寸
        const { cols, rows } = tab.term;
        ws.send(JSON.stringify({ type: "resize", size: { cols, rows } }));
      }
      // 如果全局监控已开启且当前为活跃标签，开启监控（非活跃标签不监控）
      if (monitorGlobalEnabled.value && activeTabId.value === tab.id) {
        ws.send(JSON.stringify({ type: "monitor_toggle", enabled: true }));
      }
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === "data") {
          if (tab.term) {
            tab.term.write(msg.data);
          }
        } else if (msg.type === "monitor") {
          // 解析监控数据 — 写入响应式 map
          if (msg.monitor) {
            const snap: MonitorSnapshot = {
              cpu: msg.monitor.cpu ?? 0,
              mem_total: msg.monitor.mem_total ?? 0,
              mem_used: msg.monitor.mem_used ?? 0,
              disks: (msg.monitor.disks || []).map((d: any) => ({
                mount: d.mount || "?",
                total: d.total ?? 0,
                used: d.used ?? 0,
              })),
            };
            monitorDataMap[tab.id] = snap;

            // 追加 CPU 历史
            if (!cpuHistoryMap[tab.id]) {
              cpuHistoryMap[tab.id] = [];
            }
            cpuHistoryMap[tab.id].push(snap.cpu);
            if (cpuHistoryMap[tab.id].length > MAX_CPU_HISTORY) {
              cpuHistoryMap[tab.id].shift();
            }
          }
        } else if (msg.type === "error") {
          tab.status = "error";
          if (tab.term) {
            tab.term.write(`\r\n❌ ${msg.data}\r\n`);
            tab.term.write(`\x1b[90m${t("terminal.pressEnterReconnect")}\x1b[0m\r\n`);
          }
        }
      } catch {
        // 非 JSON 数据直接写入终端
        if (tab.term) {
          tab.term.write(event.data);
        }
      }
    };

    ws.onclose = () => {
      tab.status = "disconnected";
      if (tab.term) {
        tab.term.write(`\r\n⚠️  ${t("terminal.disconnected")}\r\n`);
        tab.term.write(`\x1b[90m${t("terminal.pressEnterReconnect")}\x1b[0m\r\n`);
      }
    };

    ws.onerror = () => {
      tab.status = "error";
      if (tab.term) {
        tab.term.write(`\r\n❌ ${t("terminal.connectionFailed")}\r\n`);
        tab.term.write(`\x1b[90m${t("terminal.pressEnterReconnect")}\x1b[0m\r\n`);
      }
    };
  } catch (e: any) {
    tab.status = "error";
    if (tab.term) {
      tab.term.write(`\r\n❌ ${e?.message || t("terminal.connectionFailed")}\r\n`);
      tab.term.write(`\x1b[90m${t("terminal.pressEnterReconnect")}\x1b[0m\r\n`);
    }
  }
}

// ===================== 标签页操作 =====================
function addTab(name: string, host: string, port: string, username: string, password: string) {
  const id = `term-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const tab: TerminalTab = {
    id,
    name,
    host,
    port,
    username,
    password,
    status: "disconnected",
    term: null,
    fitAddon: null,
    ws: null,
  };

  const oldActiveId = activeTabId.value;
  tabs.value.push(tab);
  activeTabId.value = id;

  // 停止旧标签上的远程监控（新标签在 ws.onopen 中按需开启）
  if (monitorGlobalEnabled.value && oldActiveId) {
    const oldTab = tabs.value.find(t => t.id === oldActiveId);
    if (oldTab?.ws && oldTab.ws.readyState === WebSocket.OPEN) {
      oldTab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: false }));
    }
  }

  // 延迟创建终端实例，等待 DOM 渲染
  nextTick(() => {
    createTerminalInstance(tab);
    connectTab(tab);
  });
}

function switchTab(id: string) {
  const oldActiveId = activeTabId.value;
  activeTabId.value = id;

  // 切换标签时同步监控：关闭旧标签监控，开启新标签监控
  if (monitorGlobalEnabled.value) {
    const oldTab = tabs.value.find(t => t.id === oldActiveId);
    if (oldTab?.ws && oldTab.ws.readyState === WebSocket.OPEN) {
      oldTab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: false }));
    }
    const newTab = tabs.value.find(t => t.id === id);
    if (newTab?.ws && newTab.ws.readyState === WebSocket.OPEN) {
      newTab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: true }));
    }
  }

  // 切换标签时重新 fit
  nextTick(() => {
    const tab = tabs.value.find((t) => t.id === id);
    if (tab?.fitAddon && tab.term) {
      fitSnap(tab.term, tab.fitAddon);
    }
  });
}

function closeTab(id: string) {
  const idx = tabs.value.findIndex((t) => t.id === id);
  if (idx === -1) return;

  const tab = tabs.value[idx];

  // 清理 WebSocket
  if (tab.ws) {
    tab.ws.close();
    tab.ws = null;
  }

  // 清理终端
  if (tab.term) {
    const observer = (tab.term as any)._resizeObserver;
    if (observer) observer.disconnect();
    _cleanupViewportGuard(tab.term);
    tab.term.dispose();
    tab.term = null;
  }

  // 清理监控数据
  delete monitorDataMap[id];
  delete cpuHistoryMap[id];

  tabs.value.splice(idx, 1);

  // 切换到相邻标签，并同步监控到新活跃标签
  if (activeTabId.value === id) {
    if (tabs.value.length > 0) {
      const newIdx = Math.min(idx, tabs.value.length - 1);
      activeTabId.value = tabs.value[newIdx].id;
      // 如果全局监控开启，启动新活跃标签的监控
      if (monitorGlobalEnabled.value) {
        const newActiveTab = tabs.value[newIdx];
        if (newActiveTab?.ws && newActiveTab.ws.readyState === WebSocket.OPEN) {
          newActiveTab.ws.send(JSON.stringify({ type: "monitor_toggle", enabled: true }));
        }
      }
    } else {
      activeTabId.value = "";
    }
  }
}

// ===================== 重连操作 =====================
function reconnectTab(id: string) {
  const tab = tabs.value.find((t) => t.id === id);
  if (!tab) return;

  // 关闭旧 WebSocket
  if (tab.ws) {
    tab.ws.onclose = null;
    tab.ws.onerror = null;
    tab.ws.onmessage = null;
    tab.ws.close();
    tab.ws = null;
  }

  // 重置终端内容
  if (tab.term) {
    tab.term.clear();
    tab.term.reset();
  }

  // 重新建立连接
  connectTab(tab);

  // 切换后重新 fit
  nextTick(() => {
    if (tab.fitAddon && tab.term) {
      fitSnap(tab.term, tab.fitAddon);
    }
  });
}

// ===================== 连接操作 =====================
function connectLocal() {
  // 保存本地 SSH 凭据
  saveLocalSshCredentials(localForm.port, localForm.username, localForm.password);
  
  addTab(
    `${t("terminal.localhost")} (${localForm.port})`,
    "127.0.0.1",
    localForm.port,
    localForm.username,
    localForm.password
  );
  showAddDialog.value = false;
}

function connectProxy() {
  if (!proxyForm.proxyUuid) return;
  const proxy = proxyList.value.find((p: any) => p.uuid === proxyForm.proxyUuid);
  if (!proxy) return;

  // 代理当前监听端口 > 缓存端口 > 默认 22
  const sshPort = proxy.listen_port || proxyForm.port || "22";

  // 保存代理 SSH 凭据
  saveProxySshCredentials(proxyForm.proxyUuid, sshPort, proxyForm.username, proxyForm.password);

  addTab(
    proxy.name || proxyForm.proxyUuid,
    proxy.listen_address || "127.0.0.1",
    sshPort,
    proxyForm.username,
    proxyForm.password
  );
  showAddDialog.value = false;
}

function onProxySelect(uuid: string) {
  const proxy = proxyList.value.find((p: any) => p.uuid === uuid);
  const cachedCreds = loadProxySshCredentials(uuid);

  // 代理当前监听端口 > 缓存端口 > 默认 22
  proxyForm.port = proxy?.listen_port || cachedCreds?.port || "22";
  proxyForm.username = cachedCreds?.username || "root";
  proxyForm.password = cachedCreds?.password || "";
}

// 代理选项筛选函数
function filterProxyOption(input: string, option: any) {
  const label = option.label?.toLowerCase() || '';
  return label.includes(input.toLowerCase());
}

// ===================== 生命周期 =====================
// 页面加载时恢复本地 SSH 凭据
const cachedLocalCreds = loadLocalSshCredentials();
if (cachedLocalCreds) {
  localForm.port = cachedLocalCreds.port || "22";
  localForm.username = cachedLocalCreds.username || "root";
  localForm.password = cachedLocalCreds.password || "";
}

fetchProxyList();

onBeforeUnmount(() => {
  // 清理所有终端
  for (const tab of tabs.value) {
    if (tab.ws) tab.ws.close();
    if (tab.term) {
      const observer = (tab.term as any)._resizeObserver;
      if (observer) observer.disconnect();
      _cleanupViewportGuard(tab.term);
      tab.term.dispose();
    }
  }
});
</script>

<style scoped lang="less">
.terminal-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}

.terminal-toolbar {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #2d2d2d;
  border-bottom: 1px solid #3c3c3c;
  flex-shrink: 0;
}

.terminal-tabs-wrap {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.terminal-tabs {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
  overflow-x: auto;
  flex-shrink: 0;
  min-height: 36px;

  &::-webkit-scrollbar {
    height: 3px;
  }
  &::-webkit-scrollbar-thumb {
    background: #555;
  }
}

.terminal-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  cursor: pointer;
  color: #cccccc;
  font-size: 13px;
  border-right: 1px solid #3c3c3c;
  white-space: nowrap;
  user-select: none;
  transition: background 0.15s;

  &:hover {
    background: #2a2d2e;
  }

  &.active {
    background: #1e1e1e;
    color: #ffffff;
    border-bottom: 2px solid #007acc;
  }
}

.tab-label {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;

  &.connecting {
    background: #e5a610;
    animation: pulse 1s infinite;
  }
  &.connected {
    background: #0dbc79;
  }
  &.disconnected {
    background: #666666;
  }
  &.error {
    background: #cd3131;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.tab-close {
  font-size: 11px;
  color: var(--color-text-muted, #888);
  padding: 2px;
  border-radius: 3px;

  &:hover {
    color: #fff;
    background: #555;
  }
}

.new-tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  cursor: pointer;
  color: #cccccc;
  font-size: 13px;
  white-space: nowrap;
  user-select: none;
  transition: background 0.15s;
  min-width: auto;

  &:hover {
    background: #2a2d2e;
    color: #ffffff;
  }
}

.new-tab-icon {
  font-size: 12px;
}

.terminal-content {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: #1e1e1e;
}

// 消除 xterm.js 内部边框产生的白线
.terminal-instance :deep(.xterm-viewport) {
  border: none !important;
}

.terminal-instance :deep(.xterm) {
  border: none !important;
}

.terminal-instance {
  position: absolute;
  top: 6px;
  left: 8px;
  right: 0;
  bottom: 0;

  display: none;

  &.visible {
    display: block;
  }
}

.monitor-toggle {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 14px;
  cursor: default;
  border-right: none;

  &:hover {
    background: transparent;
  }
}

.toggle-label {
  font-size: 12px;
  color: var(--color-text-muted, #999);
  white-space: nowrap;
}

.monitor-panel {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 5px 12px;
  background: var(--color-terminal-connect-bg, #ffffff);
  flex-shrink: 0;
  min-height: 34px;
  overflow-x: auto;
  border-top: 1px solid var(--color-terminal-connect-border, #e0e0e0);

  .monitor-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 12px;
    border-right: 1px solid var(--color-terminal-connect-border, #e0e0e0);
    flex-shrink: 0;

    &:last-child {
      border-right: none;
    }
  }

  .monitor-cpu {
    gap: 4px;
  }

  .monitor-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--color-terminal-connect-label, #555);
    white-space: nowrap;
    flex-shrink: 0;

    &.disk-label {
      min-width: 0;
      max-width: 72px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .cpu-sparkline {
    width: 80px;
    height: 24px;
    flex-shrink: 0;
    border-radius: 2px;
    background: var(--color-terminal-connect-input-bg, #f0f0f0);

    polyline {
      vector-effect: non-scaling-stroke;
    }
  }

  .monitor-bar-wrap {
    width: 80px;
    height: 8px;
    background: var(--color-terminal-connect-input-hover, #e8e8e8);
    border-radius: 4px;
    overflow: hidden;
    flex-shrink: 0;
  }

  .monitor-bar {
    height: 100%;
    border-radius: 4px;
    transition: width 0.4s ease;

    &.mem {
      background: #569cd6;
    }
    &.disk {
      background: #ce9178;
    }
  }

  .monitor-value {
    font-size: 11px;
    color: var(--color-terminal-connect-text, #333);
    white-space: nowrap;
    font-family: Consolas, "Courier New", monospace;

    &.cpu-value {
      min-width: 52px;
      text-align: right;
    }
  }
}

.terminal-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e1e1e;
  color: #ffffff;

  :deep(.ant-empty-description) {
    color: #ffffff;
  }
}

.terminal-form {
  padding: 8px 0;
}
</style>
