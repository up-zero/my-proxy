<template>
  <Table ref="innerRef" v-bind="$attrs" :columns="fitColumns" :scroll="fitScroll" :row-selection="fitRowSelection">
    <template v-for="(_, name) in $slots" #[name]="slotProps">
      <slot :name="name" v-bind="slotProps || {}" />
    </template>
  </Table>
</template>

<script lang="ts" setup>
/**
 * 全局 a-table 包装组件（在 main.ts 中注册为 ATable，覆盖 antd 默认实现），
 * 用于修复「固定表头（scroll.y）时表头与内容错位」的问题，所有页面统一生效。
 *
 * 错位原因：
 * 1. 开启 scroll.y 后，antd 会把表头与内容渲染成两个独立的 table：
 *    表头列宽取的是内容列“测量后的整数宽度”，而内容列宽是浏览器按比例缩放后的小数宽度，
 *    当列宽合计与 scroll.x 不一致（浏览器会等比缩放列宽）时，取整差异逐列累积形成错位；
 * 2. 表头会额外留出一个「纵向滚动条宽度」的占位单元格，
 *    若实际滚动条宽度与表格内部测量值不一致，也会造成错位（滚动条样式见 assets/less/global.less）。
 *
 * 处理方式：让列宽合计与 scroll.x 严格相等，避免浏览器等比缩放列宽：
 * - 未指定宽度的列（自适应列）平分剩余空间；
 * - 所有列都指定了宽度时，把剩余空间均摊到每一列。
 */
import { Table } from "ant-design-vue";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

defineOptions({ name: "ATable", inheritAttrs: false });

const props = defineProps<{
  columns?: any[];
  scroll?: any;
  rowSelection?: any;
}>();

/** 勾选列默认宽度（未显式指定时参与列宽合计） */
const SELECTION_COLUMN_WIDTH = 52;
/** 未指定宽度的列（自适应列）在可视区偏窄时的最小宽度 */
const FLEX_COLUMN_MIN_WIDTH = 200;

const innerRef = ref<any>();
/** 表格可视区宽度（不含纵向滚动条） */
const viewportWidth = ref(0);
let viewportObserver: ResizeObserver | null = null;

const measureViewport = () => {
  const root: HTMLElement | undefined = innerRef.value?.$el;
  if (!root) return;
  const bodyEl = root.querySelector<HTMLElement>(".ant-table-body");
  viewportWidth.value = bodyEl ? bodyEl.clientWidth : root.clientWidth;
};

onMounted(() => {
  measureViewport();
  const root: HTMLElement | undefined = innerRef.value?.$el;
  if (root && typeof ResizeObserver !== "undefined") {
    viewportObserver = new ResizeObserver(() => measureViewport());
    viewportObserver.observe(root);
  }
});

onBeforeUnmount(() => {
  viewportObserver?.disconnect();
  viewportObserver = null;
});

/** 仅在固定表头（scroll.y）且列配置可计算（无分组表头）时启用 */
const layout = computed(() => {
  const columns = props.columns;
  if (!props.scroll?.y || !Array.isArray(columns) || columns.length === 0) return null;
  if (columns.some((col: any) => !col || col.children?.length)) return null;

  const widths: number[] = [];
  const flexIndexes: number[] = [];
  let fixedWidth = 0;
  columns.forEach((col: any, index: number) => {
    const width = Number(col.width);
    if (Number.isFinite(width) && width > 0) {
      widths[index] = width;
      fixedWidth += width;
    } else {
      widths[index] = 0;
      flexIndexes.push(index);
    }
  });
  const selectionWidth = props.rowSelection
    ? Number(props.rowSelection.columnWidth) || SELECTION_COLUMN_WIDTH
    : 0;
  return {
    widths,
    flexIndexes,
    selectionWidth,
    baseWidth: fixedWidth + selectionWidth + flexIndexes.length * FLEX_COLUMN_MIN_WIDTH,
  };
});

/** 表格宽度：列宽合计与可视区宽度的较大值（列宽合计大于可视区时横向滚动，不做等比缩放） */
const scrollX = computed(() => {
  const info = layout.value;
  if (!info) return undefined;
  return Math.max(info.baseWidth, viewportWidth.value);
});

const fitColumns = computed(() => {
  const columns = props.columns;
  const info = layout.value;
  if (!columns || !info) return columns;

  const extra = (scrollX.value ?? info.baseWidth) - info.baseWidth;
  const widths = [...info.widths];
  if (info.flexIndexes.length > 0) {
    // 未指定宽度的列平分剩余空间（保持整数宽度）
    const average = Math.floor(extra / info.flexIndexes.length);
    const remainder = extra - average * info.flexIndexes.length;
    info.flexIndexes.forEach((index, i) => {
      widths[index] = FLEX_COLUMN_MIN_WIDTH + average + (i === 0 ? remainder : 0);
    });
  } else {
    // 所有列都指定了宽度：按比例填充剩余空间（与浏览器缩放列宽的效果保持一致）
    const total = widths.reduce((sum, width) => sum + width, 0);
    const target = (scrollX.value ?? info.baseWidth) - info.selectionWidth;
    let used = 0;
    for (let i = 0; i < widths.length; i++) {
      widths[i] = Math.max(Math.floor((widths[i] * target) / total), widths[i]);
      used += widths[i];
    }
    // 取整后的余数补到最后一列，保证列宽合计与表格宽度完全相等
    widths[widths.length - 1] += target - used;
  }
  return columns.map((col: any, index: number) => ({ ...col, width: widths[index] }));
});

const fitScroll = computed(() => {
  const x = scrollX.value;
  return x === undefined ? props.scroll : { ...props.scroll, x };
});

const fitRowSelection = computed(() => {
  const rowSelection = props.rowSelection;
  if (!rowSelection || !layout.value) return rowSelection;
  return { ...rowSelection, columnWidth: rowSelection.columnWidth ?? SELECTION_COLUMN_WIDTH };
});
</script>
