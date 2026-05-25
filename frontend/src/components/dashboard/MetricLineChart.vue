<template>
  <div class="metric-line-chart">
    <div v-if="!hasData" class="empty-state">暂无监控数据</div>
    <div v-else ref="chartRef" class="chart-canvas"></div>
  </div>
</template>

<script setup lang="ts">
import * as echarts from "echarts/core";
import { GridComponent, TooltipComponent, type TooltipComponentOption } from "echarts/components";
import { LineChart, type LineSeriesOption } from "echarts/charts";
import { CanvasRenderer } from "echarts/renderers";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";

echarts.use([GridComponent, TooltipComponent, LineChart, CanvasRenderer]);

interface SeriesItem {
  name: string;
  color: string;
  values: number[];
}

type ValueFormatter = (value: number, seriesName?: string) => string;

type EChartsOption = echarts.ComposeOption<TooltipComponentOption | LineSeriesOption>;

const props = defineProps<{
  labels: string[];
  series: SeriesItem[];
  valueFormatter?: ValueFormatter;
  yAxisFormatter?: ValueFormatter;
}>();

const chartRef = ref<HTMLDivElement>();
let chartInstance: echarts.ECharts | null = null;
let resizeObserver: ResizeObserver | null = null;

const flattenedValues = computed(() =>
  props.series.flatMap((item) => item.values).filter((value) => Number.isFinite(value))
);

const hasData = computed(() => flattenedValues.value.length > 0);

function createOption(): EChartsOption {
  return {
    animation: true,
    grid: {
      top: 10,
      left: 8,
      right: 8,
      bottom: 24,
      containLabel: true,
    },
    tooltip: {
      trigger: "axis",
      backgroundColor: "rgba(15, 23, 42, 0.88)",
      borderWidth: 0,
      textStyle: {
        color: "#fff",
      },
      valueFormatter: (value) => {
        const numericValue = typeof value === "number" ? value : Number(value || 0);
        return props.valueFormatter ? props.valueFormatter(numericValue) : `${numericValue}`;
      },
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: props.labels,
      axisLine: {
        lineStyle: {
          color: "#e5e7eb",
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: "#98a2b3",
        fontSize: 11,
        interval: "auto",
      },
    },
    yAxis: {
      type: "value",
      splitNumber: 4,
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: "#98a2b3",
        fontSize: 11,
        formatter: (value: number) => {
          return props.yAxisFormatter ? props.yAxisFormatter(value) : `${value}`;
        },
      },
      splitLine: {
        lineStyle: {
          color: "#e5e7eb",
          type: "dashed",
        },
      },
    },
    series: props.series.map((item) => ({
      name: item.name,
      type: "line",
      data: item.values,
      smooth: true,
      showSymbol: false,
      symbol: "circle",
      symbolSize: 6,
      lineStyle: {
        width: 3,
        color: item.color,
      },
      itemStyle: {
        color: item.color,
      },
      emphasis: {
        focus: "series",
      },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: `${item.color}33` },
          { offset: 1, color: `${item.color}08` },
        ]),
      },
    })),
  };
}

async function renderChart() {
  if (!hasData.value) {
    chartInstance?.clear();
    return;
  }

  await nextTick();
  if (!chartRef.value) {
    return;
  }

  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value, undefined, { renderer: "canvas" });
  }
  chartInstance.setOption(createOption(), true);
  chartInstance.resize();
}

function setupResizeObserver() {
  if (!chartRef.value || resizeObserver) {
    return;
  }
  resizeObserver = new ResizeObserver(() => {
    chartInstance?.resize();
  });
  resizeObserver.observe(chartRef.value);
}

onMounted(async () => {
  await renderChart();
  setupResizeObserver();
});

watch(
  () => [props.labels, props.series],
  async () => {
    await renderChart();
    setupResizeObserver();
  },
  { deep: true }
);

onUnmounted(() => {
  resizeObserver?.disconnect();
  resizeObserver = null;
  chartInstance?.dispose();
  chartInstance = null;
});
</script>

<style scoped lang="less">
.metric-line-chart {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: #98a2b3;
    background: linear-gradient(180deg, rgba(22, 119, 255, 0.03), rgba(82, 196, 26, 0.02));
    border-radius: 12px;
  }

  .chart-canvas {
    flex: 1;
    min-height: 0;
    width: 100%;
  }
}
</style>

