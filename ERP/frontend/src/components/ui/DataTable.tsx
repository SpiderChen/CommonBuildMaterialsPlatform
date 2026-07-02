import dayjs, { type Dayjs } from "dayjs";
import "dayjs/locale/zh-cn";
import { ArrowDown, ArrowUp, Check, ChevronsLeft, ChevronsRight, ChevronLeft, ChevronRight, Columns3, Copy, FileText, Filter, KeyRound, MousePointerClick, RefreshCw, RotateCcw, Search, X } from "lucide-react";
import { isValidElement, type MouseEvent as ReactMouseEvent, type ReactNode, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { statusLabel } from "./StatusChip";
import { BareButton, Button as UiButton, IconButton } from "./Button";
import { ContextMenu, type ContextMenuItem } from "./ContextMenu";
import { Field, IconField, SelectInput, TextInput } from "./Field";
import { useMessage } from "./Message";
import { copyTextToClipboard } from "./clipboard";

dayjs.locale("zh-cn");

export type DataTableColumn<T> = {
  key: string;
  title: ReactNode;
  render: (row: T, index: number) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string;
  filterKeys?: string[];
  filterLabels?: Record<string, (row: T, index: number) => ReactNode>;
  filterMultiple?: boolean;
  locked?: boolean;
  hideable?: boolean;
  reorderable?: boolean;
};

export type DataTableProps<T> = {
  title?: ReactNode;
  data: T[];
  columns: DataTableColumn<T>[];
  rowKey: (row: T) => string | number;
  emptyText?: string;
  headerLeftAction?: ReactNode;
  headerAction?: ReactNode;
  searchable?: boolean;
  autoFilters?: boolean;
  searchPlaceholder?: string;
  searchText?: (row: T) => string;
  onRefresh?: () => void | Promise<void>;
  refreshDisabled?: boolean;
  refreshTitle?: string;
  pageSize?: number;
  pageSizeOptions?: number[];
  showPagination?: boolean;
  maxAutoFilterOptions?: number;
  filterPlacement?: "toolbar" | "header";
  filterMultiple?: boolean;
  columnSettingsEnabled?: boolean;
  columnSettingsKey?: string;
  columnSettingsLabel?: string;
  rowContextMenu?: (row: T, index: number, helpers: DataTableContextMenuHelpers<T>) => ContextMenuItem[];
  tableContextMenu?: (helpers: DataTableContextMenuTableHelpers) => ContextMenuItem[];
};

export type DataTableContextMenuHelpers<T> = {
  row: T;
  rowIndex: number;
  rowKey: string | number;
  rowText: string;
  rowJson: string;
  cellText: string;
  columnKey?: string;
  columnTitle?: string;
  copyText: (text: string, label?: string) => void;
  searchText: (text: string) => void;
  clearSearchAndFilters: () => void;
  refresh: () => void;
};

export type DataTableContextMenuTableHelpers = {
  copyText: (text: string, label?: string) => void;
  searchText: (text: string) => void;
  clearSearchAndFilters: () => void;
  refresh: () => void;
};

type RowContextMenuState<T> = {
  row: T;
  rowIndex: number;
  position: { x: number; y: number };
  columnKey?: string;
  columnTitle?: string;
  cellText: string;
  actions: RowDomAction[];
};

type RowDomAction = {
  key: string;
  label: string;
  disabled: boolean;
  danger: boolean;
  trigger: HTMLElement;
};

type HeaderFilterMenuState = {
  columnKey: string;
  columnTitle: string;
  position: { x: number; y: number };
};

type DataTableColumnSettings = {
  order: string[];
  hidden: string[];
};

type AutoFilterCandidate = {
  fields: string[];
  label: string;
  columnKeys: string[];
  labelFields?: string[];
  useStatusLabel?: boolean;
};

type AutoFilterOption = {
  value: string;
  label: string;
};

type AutoFilterDefinition<T> = {
  key: string;
  label: string;
  columnKeys: string[];
  kind?: "select" | "dateRange";
  options: AutoFilterOption[];
  matches: (row: T, index: number, value: string) => boolean;
};

type ActiveFilterValue = string | string[];
type ActiveFilterValues = Record<string, ActiveFilterValue>;

const categoricalFilterCandidates: AutoFilterCandidate[] = [
  { fields: ["status"], label: "状态", columnKeys: ["status"], useStatusLabel: true },
  { fields: ["result"], label: "结果", columnKeys: ["result"], useStatusLabel: true },
  { fields: ["severity"], label: "等级", columnKeys: ["severity"], useStatusLabel: true },
  { fields: ["onlineStatus"], label: "在线状态", columnKeys: ["onlineStatus", "status"], useStatusLabel: true },
  { fields: ["transportStatus"], label: "运输状态", columnKeys: ["transportStatus", "status"], useStatusLabel: true },
  { fields: ["availableStatus"], label: "库存状态", columnKeys: ["availableStatus", "status"], useStatusLabel: true },
  { fields: ["siteId", "currentSiteId"], label: "站点", columnKeys: ["site", "currentSite", "currentSiteId"], labelFields: ["siteName", "currentSiteName"] },
  { fields: ["companyId"], label: "公司", columnKeys: ["company"], labelFields: ["companyName"] },
  { fields: ["departmentId"], label: "部门", columnKeys: ["department"], labelFields: ["departmentName"] },
  { fields: ["customerId"], label: "客户", columnKeys: ["customer"], labelFields: ["customerName"] },
  { fields: ["projectId"], label: "项目", columnKeys: ["project"], labelFields: ["projectName"] },
  { fields: ["productId"], label: "产品", columnKeys: ["product"], labelFields: ["productName"] },
  { fields: ["materialId"], label: "物料", columnKeys: ["material"], labelFields: ["materialName"] },
  { fields: ["driverId"], label: "司机", columnKeys: ["driver"], labelFields: ["driverName"] },
  { fields: ["vehicleId"], label: "车辆", columnKeys: ["vehicle", "plateNo"], labelFields: ["plateNo", "vehicleName"] },
  { fields: ["orderId"], label: "订单", columnKeys: ["order", "orderNo"], labelFields: ["orderNo"] },
  { fields: ["dispatchId"], label: "配送单", columnKeys: ["dispatch", "dispatchNo"], labelFields: ["dispatchNo"] },
  { fields: ["planId"], label: "计划", columnKeys: ["plan"] },
  { fields: ["roleCode", "currentRole"], label: "角色", columnKeys: ["role", "current"] },
  { fields: ["resource"], label: "业务对象", columnKeys: ["resource"] },
  { fields: ["kind", "type", "category", "sampleType", "invoiceType", "sourceType", "channel"], label: "类型", columnKeys: ["kind", "type", "category", "resource", "sourceType", "channel"] }
];

const dateFilterCandidates = [
  { fields: ["planDate", "dispatchDate", "deliveryDate", "reportDate", "productionDate"], label: "业务日期" },
  { fields: ["signedAt", "ticketTime", "createdAt", "submittedAt", "approvedAt", "updatedAt"], label: "时间" },
  { fields: ["plannedTestAt", "calibratedAt", "lastCalibrationAt", "nextCalibrationAt", "nextDueAt", "dueDate", "expiresAt", "lastLocationTime"], label: "日期" }
];

const DATE_RANGE_VALUE_PREFIX = "range:";
const ANT_RANGE_DISPLAY_FORMAT = "YYYY-MM-DD HH:mm:ss";
const ANT_RANGE_STORAGE_FORMAT = "YYYY-MM-DDTHH:mm:ss";
const ANT_RANGE_DEFAULT_OPEN_VALUE = [
  dayjs().startOf("day"),
  dayjs().hour(23).minute(59).second(59).millisecond(0)
] as [Dayjs, Dayjs];
const DATE_PANEL_WEEKDAYS = ["一", "二", "三", "四", "五", "六", "日"];
const DATE_PANEL_CELL_COUNT = 42;
const TIME_PARTS = [
  { key: "hour", label: "时", max: 23 },
  { key: "minute", label: "分", max: 59 },
  { key: "second", label: "秒", max: 59 }
] as const;
const TIME_PART_OPTIONS = TIME_PARTS.reduce(
  (options, part) => ({
    ...options,
    [part.key]: Array.from({ length: part.max + 1 }, (_, index) => index)
  }),
  {} as Record<TimePart, number[]>
);
type DateTimeRangeDraft = [Dayjs | null, Dayjs | null];
type RangeSide = "start" | "end";
type TimePart = (typeof TIME_PARTS)[number]["key"];

function nodeText(node: ReactNode): string {
  if (node === null || node === undefined || typeof node === "boolean") {
    return "";
  }

  if (typeof node === "string" || typeof node === "number" || typeof node === "bigint") {
    return String(node);
  }

  if (Array.isArray(node)) {
    return node.map(nodeText).join(" ");
  }

  if (isValidElement<{ children?: ReactNode }>(node)) {
    return nodeText(node.props.children);
  }

  return "";
}

function columnTitleText<T>(column: DataTableColumn<T>) {
  return nodeText(column.title).trim() || column.key;
}

function isColumnHideable<T>(column: DataTableColumn<T>) {
  return column.hideable ?? (column.key !== "actions" && !column.locked);
}

function isColumnReorderable<T>(column: DataTableColumn<T>) {
  return column.reorderable ?? (column.key !== "actions" && !column.locked);
}

function defaultColumnSettings<T>(columns: DataTableColumn<T>[]): DataTableColumnSettings {
  return {
    order: columns.filter(isColumnReorderable).map((column) => column.key),
    hidden: []
  };
}

function dataTableColumnSettingsStorageKey(key: string) {
  return `cbmp:data-table-columns:${key}`;
}

function autoColumnSettingsKey<T>(columns: DataTableColumn<T>[], title?: ReactNode) {
  const path = typeof window !== "undefined" ? window.location.pathname : "table";
  const titleText = nodeText(title).trim() || "table";
  const columnKeys = columns.map((column) => column.key).join(".");
  return `auto:${path}:${titleText}:${columnKeys}`;
}

function readColumnSettings(key?: string): DataTableColumnSettings {
  if (!key || typeof window === "undefined") {
    return { order: [], hidden: [] };
  }

  try {
    const raw = window.localStorage.getItem(dataTableColumnSettingsStorageKey(key));
    if (!raw) return { order: [], hidden: [] };
    const parsed = JSON.parse(raw) as Partial<DataTableColumnSettings>;
    return {
      order: Array.isArray(parsed.order) ? parsed.order.filter((item): item is string => typeof item === "string") : [],
      hidden: Array.isArray(parsed.hidden) ? parsed.hidden.filter((item): item is string => typeof item === "string") : []
    };
  } catch {
    return { order: [], hidden: [] };
  }
}

function writeColumnSettings(key: string, settings: DataTableColumnSettings) {
  if (typeof window === "undefined") return;

  try {
    window.localStorage.setItem(dataTableColumnSettingsStorageKey(key), JSON.stringify(settings));
  } catch {
    // Persisting table preferences is best-effort only.
  }
}

function removeColumnSettings(key: string) {
  if (typeof window === "undefined") return;

  try {
    window.localStorage.removeItem(dataTableColumnSettingsStorageKey(key));
  } catch {
    // Clearing table preferences is best-effort only.
  }
}

function columnSettingsSignature(settings: DataTableColumnSettings) {
  return `${settings.order.join(",")}::${settings.hidden.join(",")}`;
}

function normalizeColumnSettings<T>(columns: DataTableColumn<T>[], settings: DataTableColumnSettings): DataTableColumnSettings {
  const columnsByKey = new Map(columns.map((column) => [column.key, column]));
  const reorderableKeys = columns.filter(isColumnReorderable).map((column) => column.key);
  const reorderableKeySet = new Set(reorderableKeys);
  const hiddenKeySet = new Set<string>();
  const order: string[] = [];

  settings.order.forEach((key) => {
    if (reorderableKeySet.has(key) && !order.includes(key)) {
      order.push(key);
    }
  });

  reorderableKeys.forEach((key) => {
    if (!order.includes(key)) {
      order.push(key);
    }
  });

  settings.hidden.forEach((key) => {
    const column = columnsByKey.get(key);
    if (column && isColumnHideable(column)) {
      hiddenKeySet.add(key);
    }
  });

  if (columns.length && columns.every((column) => hiddenKeySet.has(column.key) || !isColumnHideable(column))) {
    const fallback = columns.find(isColumnHideable);
    if (fallback) {
      hiddenKeySet.delete(fallback.key);
    }
  }

  return {
    order,
    hidden: columns.map((column) => column.key).filter((key) => hiddenKeySet.has(key))
  };
}

function settingsEqual(left: DataTableColumnSettings, right: DataTableColumnSettings) {
  return columnSettingsSignature(left) === columnSettingsSignature(right);
}

function orderColumns<T>(columns: DataTableColumn<T>[], settings: DataTableColumnSettings) {
  const columnsByKey = new Map(columns.map((column) => [column.key, column]));
  const orderedReorderableColumns = settings.order
    .map((key) => columnsByKey.get(key))
    .filter((column): column is DataTableColumn<T> => Boolean(column && isColumnReorderable(column)));
  const queue = [...orderedReorderableColumns];

  return columns.map((column) => {
    if (!isColumnReorderable(column)) {
      return column;
    }

    return queue.shift() || column;
  });
}

function visibleColumns<T>(columns: DataTableColumn<T>[], settings: DataTableColumnSettings) {
  const hidden = new Set(settings.hidden);
  return orderColumns(columns, settings).filter((column) => !hidden.has(column.key));
}

function columnFilterKeys<T>(column: DataTableColumn<T>) {
  return new Set([column.key, ...(column.filterKeys || [])]);
}

function filterBelongsToColumn<T>(filter: AutoFilterDefinition<T>, column: DataTableColumn<T>) {
  const keys = columnFilterKeys(column);
  return keys.has(filter.key) || filter.columnKeys.some((key) => keys.has(key));
}

function activeFilterValueList(value: ActiveFilterValue | undefined) {
  if (Array.isArray(value)) {
    return value.filter(Boolean);
  }
  return value ? [value] : [];
}

function activeFilterSingleValue(value: ActiveFilterValue | undefined) {
  return Array.isArray(value) ? value[0] || "" : value || "";
}

function activeFilterValueIsActive(value: ActiveFilterValue | undefined) {
  return activeFilterValueList(value).length > 0;
}

function activeFilterValueSignature(value: ActiveFilterValue | undefined) {
  return activeFilterValueList(value).join(",");
}

function activeFilterMatches<T>(filter: AutoFilterDefinition<T>, row: T, index: number, value: ActiveFilterValue | undefined) {
  const selectedValues = activeFilterValueList(value);
  if (!selectedValues.length) return true;
  return selectedValues.some((item) => filter.matches(row, index, item));
}

function filtersForColumns<T>(filters: AutoFilterDefinition<T>[], columns: DataTableColumn<T>[]) {
  const grouped = new Map<string, AutoFilterDefinition<T>[]>();

  columns.forEach((column) => {
    const matches = filters.filter((filter) => filterBelongsToColumn(filter, column));
    if (!matches.length) return;
    grouped.set(column.key, matches);
  });

  return grouped;
}

function clampedHeaderFilterPosition(position: { x: number; y: number }, width: number) {
  if (typeof window === "undefined") {
    return position;
  }

  return {
    x: Math.max(8, Math.min(position.x, window.innerWidth - width - 8)),
    y: Math.max(8, Math.min(position.y, window.innerHeight - 80))
  };
}

function stringifyRow(row: unknown) {
  try {
    return JSON.stringify(row, null, 2);
  } catch {
    return String(row);
  }
}

const businessCopyFieldLabels: Array<[string, string]> = [
  ["orderNo", "订单号"],
  ["dispatchNo", "派车单号"],
  ["planNo", "计划号"],
  ["taskNo", "任务号"],
  ["batchNo", "批次号"],
  ["ticketNo", "磅单号"],
  ["noteNo", "送货单号"],
  ["signNo", "签收单号"],
  ["linkNo", "签收链接号"],
  ["statementNo", "对账单号"],
  ["contractNo", "合同号"],
  ["invoiceNo", "发票号"],
  ["receiptNo", "收款单号"],
  ["paymentNo", "付款单号"],
  ["billNo", "往来单号"],
  ["requestNo", "申请单号"],
  ["purchaseNo", "采购单号"],
  ["trialNo", "试配编号"],
  ["sampleNo", "样品编号"],
  ["exceptionNo", "异常编号"],
  ["certificateNo", "证书号"],
  ["plateNo", "车牌号"],
  ["phone", "电话"],
  ["code", "编码"]
];

function rowFieldValue(row: unknown, field: string) {
  if (!row || typeof row !== "object") return "";
  const value = (row as Record<string, unknown>)[field];
  if (value === null || value === undefined) return "";
  return String(value).trim();
}

function businessCopyMenuItems(row: unknown, copyText: (text: string, label?: string) => void): ContextMenuItem[] {
  const seenValues = new Set<string>();
  return businessCopyFieldLabels
    .map(([field, label]) => ({ field, label, value: rowFieldValue(row, field) }))
    .filter((item) => {
      if (!item.value || seenValues.has(item.value)) return false;
      seenValues.add(item.value);
      return true;
    })
    .slice(0, 4)
    .map((item) => ({
      key: `copy-business-${item.field}`,
      label: `复制${item.label}`,
      icon: <Copy size={14} />,
      onSelect: () => copyText(item.value, item.label)
    }));
}

function contextMenuItemText(item: ContextMenuItem) {
  return "type" in item ? "" : nodeText(item.label).trim();
}

function isActionColumn<T>(column: DataTableColumn<T>) {
  const title = nodeText(column.title).trim();
  return column.key === "actions" || title === "操作" || title === "字段" || /操作|动作|Action/i.test(title);
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof Element && Boolean(target.closest("button, a, input, textarea, select, [role='button'], [role='menuitem'], [data-row-action-skip]"));
}

function labelForActionElement(element: HTMLElement, index: number) {
  return (
    element.getAttribute("aria-label")
    || element.getAttribute("title")
    || element.textContent?.replace(/\s+/g, " ").trim()
    || `操作 ${index + 1}`
  );
}

function collectRowActions(rowElement: HTMLTableRowElement | null): RowDomAction[] {
  if (!rowElement) return [];
  const actionCells = Array.from(rowElement.querySelectorAll<HTMLElement>("[data-column-key='actions'], .data-table-action-cell"));
  if (!actionCells.length) return [];
  const candidates = actionCells.flatMap((cell) => Array.from(cell.querySelectorAll<HTMLElement>("button, a[href], [role='button']")));
  const uniqueCandidates = Array.from(new Set(candidates));
  return uniqueCandidates
    .filter((element) => !element.closest("[data-row-action-skip]"))
    .map((element, index) => {
      const disabled = (
        element.hasAttribute("disabled")
        || element.getAttribute("aria-disabled") === "true"
        || element.classList.contains("disabled")
      );
      const label = labelForActionElement(element, index);
      const danger = (
        element.classList.contains("danger")
        || element.className.includes("danger")
        || /删除|移除|作废|驳回|关闭|撤销|退回|Disable|Delete|Remove|Reject|Close/i.test(label)
      );
      return {
        key: `${index}-${label}`,
        label,
        disabled,
        danger,
        trigger: element
      };
    })
    .filter((item) => item.label && item.label !== "-");
}

function clickRowAction(action: RowDomAction) {
  if (action.disabled) return;
  action.trigger.click();
}

function buildSearchPlaceholder() {
  return "搜索";
}

function isDateRangeValue(value: string) {
  return value.startsWith(DATE_RANGE_VALUE_PREFIX);
}

function pad2(value: number) {
  return String(value).padStart(2, "0");
}

function encodeDateRange(start: string, end: string) {
  return `${DATE_RANGE_VALUE_PREFIX}${start}|${end}`;
}

function decodeDateRange(value: string) {
  const body = isDateRangeValue(value) ? value.slice(DATE_RANGE_VALUE_PREFIX.length) : "";
  const [start = "", end = ""] = body.split("|");
  return { start, end };
}

function parseDateRangeDateTime(value: string) {
  if (!value) return null;
  const parsed = dayjs(value.replace(" ", "T"));
  return parsed.isValid() ? parsed : null;
}

function dateRangePickerValue(value: string): [Dayjs, Dayjs] | null {
  const { start, end } = decodeDateRange(value);
  const startValue = parseDateRangeDateTime(start);
  const endValue = parseDateRangeDateTime(end);
  return startValue && endValue ? [startValue, endValue] : null;
}

function dateValueToDateTimeString(value: Dayjs) {
  return value.format(ANT_RANGE_STORAGE_FORMAT);
}

function encodeDateRangePickerValue(value: DateTimeRangeDraft | null) {
  return value?.[0] && value?.[1]
    ? encodeDateRange(dateValueToDateTimeString(value[0]), dateValueToDateTimeString(value[1]))
    : "";
}

function rangeSideIndex(side: RangeSide) {
  return side === "start" ? 0 : 1;
}

function emptyDateTimeRange(): DateTimeRangeDraft {
  return [null, null];
}

function dateTimeRangeDraft(value: [Dayjs, Dayjs] | null): DateTimeRangeDraft {
  return value ? [value[0], value[1]] : emptyDateTimeRange();
}

function completeDateTimeRange(value: DateTimeRangeDraft): [Dayjs, Dayjs] | null {
  return value[0] && value[1] ? [value[0], value[1]] : null;
}

function orderedDateTimeRange(value: DateTimeRangeDraft): [Dayjs, Dayjs] | null {
  const completeRange = completeDateTimeRange(value);
  if (!completeRange) return null;
  return completeRange[1].isBefore(completeRange[0]) ? [completeRange[1], completeRange[0]] : completeRange;
}

function mergeDateWithTime(dateValue: Dayjs, timeValue: Dayjs | null | undefined, fallbackTime: Dayjs) {
  const time = timeValue ?? fallbackTime;
  return dateValue
    .hour(time.hour())
    .minute(time.minute())
    .second(time.second())
    .millisecond(0);
}

function monthPanelStart(value: Dayjs) {
  return value.startOf("month").startOf("day");
}

function initialPanelMonth(value: DateTimeRangeDraft) {
  return monthPanelStart(value[0] ?? value[1] ?? dayjs());
}

function calendarGridDates(month: Dayjs) {
  const firstDay = monthPanelStart(month);
  const mondayOffset = (firstDay.day() + 6) % 7;
  const gridStart = firstDay.subtract(mondayOffset, "day");
  return Array.from({ length: DATE_PANEL_CELL_COUNT }, (_, index) => gridStart.add(index, "day"));
}

function isSameDate(left: Dayjs | null | undefined, right: Dayjs | null | undefined) {
  return Boolean(left && right && left.isSame(right, "day"));
}

function isDateBetween(dateValue: Dayjs, start: Dayjs, end: Dayjs) {
  const current = dateValue.startOf("day");
  const rangeStart = start.startOf("day");
  const rangeEnd = end.startOf("day");
  return (
    (current.isAfter(rangeStart) || current.isSame(rangeStart)) &&
    (current.isBefore(rangeEnd) || current.isSame(rangeEnd))
  );
}

function displayDraftDateTime(value: Dayjs | null) {
  return value ? value.format(ANT_RANGE_DISPLAY_FORMAT) : "未选择";
}

function getTimePartValue(value: Dayjs, part: TimePart) {
  if (part === "hour") return value.hour();
  if (part === "minute") return value.minute();
  return value.second();
}

function setTimePartValue(value: Dayjs, part: TimePart, nextValue: number) {
  if (part === "hour") return value.hour(nextValue);
  if (part === "minute") return value.minute(nextValue);
  return value.second(nextValue);
}

function readField(row: unknown, field: string) {
  if (!row || typeof row !== "object") {
    return undefined;
  }

  return (row as Record<string, unknown>)[field];
}

function normalizeFilterValue(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "";
  }

  if (value instanceof Date) {
    return value.toISOString();
  }

  return String(value);
}

function filterValues(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map(normalizeFilterValue).filter(Boolean);
  }

  const normalized = normalizeFilterValue(value);
  return normalized ? [normalized] : [];
}

function fallbackFilterLabel(value: string, useStatusLabel?: boolean) {
  if (useStatusLabel) {
    return statusLabel(value);
  }

  if (value === "true") return "是";
  if (value === "false") return "否";
  return value.replace(/_/g, " ");
}

function filterOptionLabel<T>(row: T, index: number, value: string, field: string, candidate: AutoFilterCandidate, columns: DataTableColumn<T>[]) {
  const labelFromField = candidate.labelFields
    ?.map((field) => normalizeFilterValue(readField(row, field)).trim())
    .find(Boolean);
  if (labelFromField) return labelFromField;

  const column = columns.find((item) => {
    const keys = columnFilterKeys(item);
    return keys.has(field) || candidate.columnKeys.some((key) => keys.has(key));
  });
  const customLabel = column?.filterLabels?.[field]?.(row, index);
  const customLabelText = customLabel ? nodeText(customLabel).trim() : "";
  if (customLabelText) return customLabelText;

  const rendered = column ? nodeText(column.render(row, index)).trim() : "";
  return rendered || fallbackFilterLabel(value, candidate.useStatusLabel);
}

function isUsefulFilterOption(option: AutoFilterOption) {
  const label = option.label.trim();
  return label !== "" && label !== "-" && label !== "—" && label !== "N/A";
}

function parseDateValue(value: unknown) {
  if (value instanceof Date && Number.isFinite(value.valueOf())) {
    return value;
  }

  if (typeof value !== "string" && typeof value !== "number") {
    return null;
  }

  const parsed = new Date(value);
  return Number.isFinite(parsed.valueOf()) ? parsed : null;
}

function startOfDay(value: Date) {
  return new Date(value.getFullYear(), value.getMonth(), value.getDate());
}

function matchesDateBucket(value: unknown, bucket: string) {
  const parsed = parseDateValue(value);
  if (!parsed) return false;

  if (isDateRangeValue(bucket)) {
    const { start, end } = decodeDateRange(bucket);
    const startValue = parseDateValue(start);
    const endValue = parseDateValue(end);
    const time = parsed.valueOf();
    if (startValue && time < startValue.valueOf()) return false;
    if (endValue && time > endValue.valueOf()) return false;
    return true;
  }

  const day = startOfDay(parsed).valueOf();
  const today = startOfDay(new Date());
  const todayTime = today.valueOf();
  const days = 24 * 60 * 60 * 1000;
  const monthStart = new Date(today.getFullYear(), today.getMonth(), 1).valueOf();
  const nextMonthStart = new Date(today.getFullYear(), today.getMonth() + 1, 1).valueOf();

  if (bucket === "today") return day === todayTime;
  if (bucket === "last7") return day >= todayTime - 6 * days && day <= todayTime;
  if (bucket === "last30") return day >= todayTime - 29 * days && day <= todayTime;
  if (bucket === "thisMonth") return day >= monthStart && day < nextMonthStart;
  if (bucket === "future") return day > todayTime;
  if (bucket === "overdue") return day < todayTime;
  return true;
}

function buildAutoFilters<T>(data: T[], columns: DataTableColumn<T>[], maxOptions: number): AutoFilterDefinition<T>[] {
  const definitions: AutoFilterDefinition<T>[] = [];

  categoricalFilterCandidates.forEach((candidate) => {
    const field = candidate.fields.find((name) => data.some((row) => filterValues(readField(row, name)).length > 0));
    if (!field) return;

    const optionsByValue = new Map<string, string>();
    data.forEach((row, index) => {
      filterValues(readField(row, field)).forEach((value) => {
        if (!optionsByValue.has(value)) {
          optionsByValue.set(value, filterOptionLabel(row, index, value, field, candidate, columns));
        }
      });
    });

    const options = Array.from(optionsByValue, ([value, label]) => ({ value, label }))
      .filter(isUsefulFilterOption)
      .sort((a, b) => a.label.localeCompare(b.label, "zh-CN"));
    const hasBlankRows = data.some((row) => filterValues(readField(row, field)).length === 0);
    const hasHiddenDirtyOptions = optionsByValue.size > options.length;

    if (options.length === 0 || (options.length < 2 && !hasBlankRows && !hasHiddenDirtyOptions) || options.length > maxOptions) {
      return;
    }

    definitions.push({
      key: field,
      label: candidate.label,
      columnKeys: candidate.columnKeys,
      options,
      matches: (row, _index, value) => filterValues(readField(row, field)).includes(value)
    });
  });

  const dateCandidate = dateFilterCandidates
    .map((candidate) => {
      const field = candidate.fields.find((name) => data.some((row) => parseDateValue(readField(row, name))));
      return field ? { ...candidate, field } : null;
    })
    .find(Boolean);

  if (dateCandidate) {
    definitions.push({
      key: dateCandidate.field,
      label: dateCandidate.label,
      columnKeys: [dateCandidate.field],
      kind: "dateRange",
      options: [],
      matches: (row, _index, value) => matchesDateBucket(readField(row, dateCandidate.field), value)
    });
  }

  return definitions;
}

function DateRangeActiveFields({
  activeSide,
  draftRange,
  onActiveSideChange
}: {
  activeSide: RangeSide;
  draftRange: DateTimeRangeDraft;
  onActiveSideChange: (side: RangeSide) => void;
}) {
  return (
    <div className="data-table-range-active-fields">
      <BareButton
        className={`data-table-range-active-field ${activeSide === "start" ? "is-active" : ""}`}
        onClick={() => onActiveSideChange("start")}
      >
        <span>开始时间</span>
        <strong>{displayDraftDateTime(draftRange[0])}</strong>
      </BareButton>
      <span className="data-table-range-field-separator">-</span>
      <BareButton
        className={`data-table-range-active-field ${activeSide === "end" ? "is-active" : ""}`}
        onClick={() => onActiveSideChange("end")}
      >
        <span>结束时间</span>
        <strong>{displayDraftDateTime(draftRange[1])}</strong>
      </BareButton>
    </div>
  );
}

function DateRangeCalendarPanel({
  position,
  month,
  draftRange,
  activeSide,
  onDateSelect,
  onShiftMonth
}: {
  position: "left" | "right";
  month: Dayjs;
  draftRange: DateTimeRangeDraft;
  activeSide: RangeSide;
  onDateSelect: (dateValue: Dayjs) => void;
  onShiftMonth: (offset: number, unit: "month" | "year") => void;
}) {
  const dates = useMemo(() => calendarGridDates(month), [month]);
  const orderedRange = orderedDateTimeRange(draftRange);

  return (
    <div className="data-table-range-calendar">
      <div className="data-table-range-calendar-header">
        <div className="data-table-range-calendar-nav-group">
          {position === "left" ? (
            <>
              <IconButton className="data-table-range-calendar-nav" icon={<ChevronsLeft size={16} />} label="上一年" onClick={() => onShiftMonth(-1, "year")} />
              <IconButton className="data-table-range-calendar-nav" icon={<ChevronLeft size={16} />} label="上一月" onClick={() => onShiftMonth(-1, "month")} />
            </>
          ) : null}
        </div>
        <div className="data-table-range-calendar-title">{month.format("YYYY年M月")}</div>
        <div className="data-table-range-calendar-nav-group">
          {position === "right" ? (
            <>
              <IconButton className="data-table-range-calendar-nav" icon={<ChevronRight size={16} />} label="下一月" onClick={() => onShiftMonth(1, "month")} />
              <IconButton className="data-table-range-calendar-nav" icon={<ChevronsRight size={16} />} label="下一年" onClick={() => onShiftMonth(1, "year")} />
            </>
          ) : null}
        </div>
      </div>

      <div className="data-table-range-calendar-weekdays">
        {DATE_PANEL_WEEKDAYS.map((weekday) => (
          <span key={weekday}>{weekday}</span>
        ))}
      </div>

      <div className="data-table-range-calendar-grid">
        {dates.map((dateValue) => {
          const isCurrentMonth = dateValue.isSame(month, "month");
          const isStart = isSameDate(draftRange[0], dateValue);
          const isEnd = isSameDate(draftRange[1], dateValue);
          const isInRange = orderedRange ? isDateBetween(dateValue, orderedRange[0], orderedRange[1]) : false;
          const isToday = dateValue.isSame(dayjs(), "day");

          return (
            <BareButton
              key={dateValue.format("YYYY-MM-DD")}
              className={[
                "data-table-range-calendar-cell",
                isCurrentMonth ? "" : "is-outside",
                isToday ? "is-today" : "",
                isInRange ? "is-in-range" : "",
                isStart ? "is-start" : "",
                isEnd ? "is-end" : "",
                (activeSide === "start" && isStart) || (activeSide === "end" && isEnd) ? "is-active-side" : ""
              ]
                .filter(Boolean)
                .join(" ")}
              onClick={() => onDateSelect(dateValue)}
            >
              <span>{dateValue.date()}</span>
            </BareButton>
          );
        })}
      </div>
    </div>
  );
}

function DateRangeTimePanel({
  side,
  value,
  active,
  onActivate,
  onTimePartChange
}: {
  side: RangeSide;
  value: Dayjs | null;
  active: boolean;
  onActivate: (side: RangeSide) => void;
  onTimePartChange: (side: RangeSide, part: TimePart, nextValue: number) => void;
}) {
  const selectedTime = value ?? ANT_RANGE_DEFAULT_OPEN_VALUE[rangeSideIndex(side)];
  const selectedRefs = useRef<Partial<Record<string, HTMLButtonElement | null>>>({});

  useEffect(() => {
    TIME_PARTS.forEach((part) => {
      const selectedKey = `${part.key}-${getTimePartValue(selectedTime, part.key)}`;
      selectedRefs.current[selectedKey]?.scrollIntoView({ block: "center" });
    });
  }, [selectedTime]);

  return (
    <div className={`data-table-range-time-panel ${active ? "is-active" : ""}`}>
      <BareButton className="data-table-range-time-title" onClick={() => onActivate(side)}>
        {side === "start" ? "开始时分秒" : "结束时分秒"}
      </BareButton>
      <div className="data-table-range-time-columns">
        {TIME_PARTS.map((part) => (
          <div className="data-table-range-time-column" key={part.key}>
            <span className="data-table-range-time-column-label">{part.label}</span>
            <div className="data-table-range-time-options" role="listbox" aria-label={`${side === "start" ? "开始" : "结束"}${part.label}`}>
              {TIME_PART_OPTIONS[part.key].map((option) => {
                const selected = getTimePartValue(selectedTime, part.key) === option;
                const refKey = `${part.key}-${option}`;

                return (
                  <BareButton
                    key={option}
                    ref={(node) => {
                      if (selected) {
                        selectedRefs.current[refKey] = node;
                      }
                    }}
                    className={`data-table-range-time-option ${selected ? "is-selected" : ""}`}
                    aria-selected={selected}
                    onClick={() => {
                      onActivate(side);
                      onTimePartChange(side, part.key, option);
                    }}
                  >
                    {pad2(option)}
                  </BareButton>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function DateRangeDateTimePanel({
  activeSide,
  draftRange,
  viewMonth,
  onActiveSideChange,
  onApply,
  onClear,
  onDateSelect,
  onShiftMonth,
  onTimePartChange
}: {
  activeSide: RangeSide;
  draftRange: DateTimeRangeDraft;
  viewMonth: Dayjs;
  onActiveSideChange: (side: RangeSide) => void;
  onApply: () => void;
  onClear: () => void;
  onDateSelect: (dateValue: Dayjs) => void;
  onShiftMonth: (offset: number, unit: "month" | "year") => void;
  onTimePartChange: (side: RangeSide, part: TimePart, nextValue: number) => void;
}) {
  const canApply = Boolean(completeDateTimeRange(draftRange));

  return (
    <div className="data-table-range-panel">
      <DateRangeActiveFields activeSide={activeSide} draftRange={draftRange} onActiveSideChange={onActiveSideChange} />
      <div className="data-table-range-panel-body">
        <div className="data-table-range-panel-side">
          <DateRangeCalendarPanel
            position="left"
            month={viewMonth}
            draftRange={draftRange}
            activeSide={activeSide}
            onDateSelect={onDateSelect}
            onShiftMonth={onShiftMonth}
          />
          <DateRangeTimePanel
            side="start"
            value={draftRange[0]}
            active={activeSide === "start"}
            onActivate={onActiveSideChange}
            onTimePartChange={onTimePartChange}
          />
        </div>
        <div className="data-table-range-panel-side">
          <DateRangeCalendarPanel
            position="right"
            month={viewMonth.add(1, "month")}
            draftRange={draftRange}
            activeSide={activeSide}
            onDateSelect={onDateSelect}
            onShiftMonth={onShiftMonth}
          />
          <DateRangeTimePanel
            side="end"
            value={draftRange[1]}
            active={activeSide === "end"}
            onActivate={onActiveSideChange}
            onTimePartChange={onTimePartChange}
          />
        </div>
      </div>
      <div className="data-table-range-panel-footer">
        <UiButton size="sm" onClick={onClear}>清空</UiButton>
        <UiButton size="sm" variant="primary" disabled={!canApply} onClick={onApply}>确定</UiButton>
      </div>
    </div>
  );
}

function DateRangeFilter({
  label,
  value,
  onChange
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  const selectedRange = dateRangePickerValue(value);
  const [open, setOpen] = useState(false);
  const [activeSide, setActiveSide] = useState<RangeSide>("start");
  const [draftRange, setDraftRange] = useState<DateTimeRangeDraft>(() => dateTimeRangeDraft(selectedRange));
  const [viewMonth, setViewMonth] = useState(() => initialPanelMonth(dateTimeRangeDraft(selectedRange)));
  const containerRef = useRef<HTMLDivElement | null>(null);
  const selectedRangeSignature = selectedRange ? `${selectedRange[0].valueOf()}|${selectedRange[1].valueOf()}` : "";

  useEffect(() => {
    if (!open) {
      const nextDraftRange = dateTimeRangeDraft(selectedRange);
      setDraftRange(nextDraftRange);
      setViewMonth(initialPanelMonth(nextDraftRange));
    }
  }, [open, selectedRangeSignature]);

  useEffect(() => {
    if (!open) return;

    function closeOnOutside(event: PointerEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    document.addEventListener("pointerdown", closeOnOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  function openPanel(nextOpen: boolean) {
    if (nextOpen) {
      const nextDraftRange = dateTimeRangeDraft(selectedRange);
      setDraftRange(nextDraftRange);
      setViewMonth(initialPanelMonth(nextDraftRange));
      setActiveSide(nextDraftRange[0] && !nextDraftRange[1] ? "end" : "start");
    }

    setOpen(nextOpen);
  }

  function changeDraftDate(dateValue: Dayjs) {
    const side = activeSide;
    const index = rangeSideIndex(side);

    setDraftRange((current) => {
      const nextRange: DateTimeRangeDraft = [current[0], current[1]];
      nextRange[index] = mergeDateWithTime(dateValue, current[index], ANT_RANGE_DEFAULT_OPEN_VALUE[index]);

      if (side === "start" && nextRange[1] && nextRange[0] && nextRange[0].isAfter(nextRange[1])) {
        nextRange[1] = null;
      }

      if (side === "end" && !nextRange[0] && nextRange[1]) {
        nextRange[0] = mergeDateWithTime(dateValue, ANT_RANGE_DEFAULT_OPEN_VALUE[0], ANT_RANGE_DEFAULT_OPEN_VALUE[0]);
      }

      return nextRange;
    });

    setActiveSide(side === "start" ? "end" : "start");
  }

  function changeDraftTimePart(side: RangeSide, part: TimePart, nextValue: number) {
    const index = rangeSideIndex(side);

    setDraftRange((current) => {
      const nextRange: DateTimeRangeDraft = [current[0], current[1]];
      const baseValue = current[index] ?? mergeDateWithTime(dayjs(), ANT_RANGE_DEFAULT_OPEN_VALUE[index], ANT_RANGE_DEFAULT_OPEN_VALUE[index]);
      nextRange[index] = setTimePartValue(baseValue, part, nextValue).millisecond(0);
      return nextRange;
    });
  }

  function clearRange() {
    setDraftRange(emptyDateTimeRange());
    onChange("");
    setOpen(false);
  }

  function applyRange() {
    const completeRange = orderedDateTimeRange(draftRange);
    if (!completeRange) return;
    onChange(encodeDateRangePickerValue(completeRange));
    setOpen(false);
  }

  const triggerLabel = selectedRange
    ? `${selectedRange[0].format(ANT_RANGE_DISPLAY_FORMAT)} - ${selectedRange[1].format(ANT_RANGE_DISPLAY_FORMAT)}`
    : "开始时间 - 结束时间";

  return (
    <div className="data-table-date-filter" ref={containerRef}>
      <BareButton
        className="data-table-range-trigger"
        aria-expanded={open}
        aria-label={label}
        onClick={() => openPanel(!open)}
      >
        <Filter size={14} />
        <span>{triggerLabel}</span>
      </BareButton>
      {open ? (
        <div className="data-table-date-time-popover">
          <DateRangeDateTimePanel
            activeSide={activeSide}
            draftRange={draftRange}
            viewMonth={viewMonth}
            onActiveSideChange={setActiveSide}
            onApply={applyRange}
            onClear={clearRange}
            onDateSelect={changeDraftDate}
            onShiftMonth={(offset, unit) => setViewMonth((current) => current.add(offset, unit))}
            onTimePartChange={changeDraftTimePart}
          />
        </div>
      ) : null}
    </div>
  );
}

export function DataTable<T>({
  title,
  data,
  columns,
  rowKey,
  emptyText = "暂无数据",
  headerLeftAction,
  headerAction,
  searchable,
  autoFilters = true,
  searchPlaceholder,
  searchText,
  onRefresh,
  refreshDisabled = false,
  refreshTitle = "刷新",
  pageSize = 10,
  pageSizeOptions = [10, 20, 50],
  showPagination = true,
  maxAutoFilterOptions = 60,
  filterPlacement = "header",
  filterMultiple = true,
  columnSettingsEnabled: columnSettingsEnabledProp = true,
  columnSettingsKey,
  columnSettingsLabel = "列设置",
  rowContextMenu,
  tableContextMenu
}: DataTableProps<T>) {
  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(pageSize);
  const [keyword, setKeyword] = useState("");
  const [activeFilterValues, setActiveFilterValues] = useState<ActiveFilterValues>({});
  const [pageInput, setPageInput] = useState("1");
  const [rowMenu, setRowMenu] = useState<RowContextMenuState<T> | null>(null);
  const [tableMenuPosition, setTableMenuPosition] = useState<{ x: number; y: number } | null>(null);
  const [headerFilterMenu, setHeaderFilterMenu] = useState<HeaderFilterMenuState | null>(null);
  const [headerFilterKeyword, setHeaderFilterKeyword] = useState("");
  const [columnSettingsOpen, setColumnSettingsOpen] = useState(false);
  const resolvedColumnSettingsKey = columnSettingsEnabledProp
    ? columnSettingsKey || autoColumnSettingsKey(columns, title)
    : undefined;
  const [columnSettingsState, setColumnSettingsState] = useState<DataTableColumnSettings>(() => readColumnSettings(resolvedColumnSettingsKey));
  const columnSettingsRef = useRef<HTMLDivElement | null>(null);
  const headerFilterMenuRef = useRef<HTMLDivElement | null>(null);
  const message = useMessage();
  const searchEnabled = searchable ?? true;
  const filtersInHeader = filterPlacement === "header";
  const columnSettingsEnabled = Boolean(resolvedColumnSettingsKey) && columns.some((column) => isColumnHideable(column) || isColumnReorderable(column));
  const columnSignature = columns
    .map((column) => `${column.key}:${isColumnHideable(column) ? "h" : "H"}:${isColumnReorderable(column) ? "r" : "R"}`)
    .join("|");
  const normalizedColumnSettings = useMemo(
    () => normalizeColumnSettings(columns, columnSettingsState),
    [columnSettingsState, columnSignature]
  );
  const orderedColumns = useMemo(
    () => columnSettingsEnabled ? orderColumns(columns, normalizedColumnSettings) : columns,
    [columnSettingsEnabled, columns, normalizedColumnSettings]
  );
  const resolvedColumns = useMemo(
    () => columnSettingsEnabled ? visibleColumns(columns, normalizedColumnSettings) : columns,
    [columnSettingsEnabled, columns, normalizedColumnSettings]
  );
  const hiddenColumnKeys = useMemo(
    () => new Set(normalizedColumnSettings.hidden),
    [normalizedColumnSettings]
  );
  const visibleConfigurableColumnCount = orderedColumns.filter((column) => !hiddenColumnKeys.has(column.key) && isColumnHideable(column)).length;
  const resolvedPageSizeOptions = useMemo(() => {
    const values = [pageSize, currentPageSize, ...pageSizeOptions].filter((item) => Number.isFinite(item) && item > 0);
    return Array.from(new Set(values)).sort((left, right) => left - right);
  }, [currentPageSize, pageSize, pageSizeOptions]);
  const resolvedSearchPlaceholder = useMemo(
    () => searchPlaceholder || buildSearchPlaceholder(),
    [searchPlaceholder]
  );
  const filterDefinitions = useMemo(
    () => autoFilters ? buildAutoFilters(data, resolvedColumns, maxAutoFilterOptions) : [],
    [autoFilters, data, maxAutoFilterOptions, resolvedColumns]
  );
  const toolbarFilters = filtersInHeader ? [] : filterDefinitions;
  const columnFilters = useMemo(
    () => filtersInHeader ? filtersForColumns(filterDefinitions, resolvedColumns) : new Map<string, AutoFilterDefinition<T>[]>(),
    [filterDefinitions, filtersInHeader, resolvedColumns]
  );
  const activeFilterSignature = filterDefinitions.map((item) => `${item.key}:${activeFilterValueSignature(activeFilterValues[item.key])}`).join("|");
  const normalizedKeyword = keyword.trim().toLowerCase();
  const filteredData = useMemo(() => {
    const enabledFilters = filterDefinitions.filter((filter) => activeFilterValueIsActive(activeFilterValues[filter.key]));

    return data.filter((row, index) => {
      const matchesFilters = enabledFilters.every((filter) => activeFilterMatches(filter, row, index, activeFilterValues[filter.key]));
      if (!matchesFilters) return false;
      if (!searchEnabled || normalizedKeyword === "") return true;

      const text = searchText
        ? searchText(row)
        : resolvedColumns
          .filter((column) => column.key !== "actions")
          .map((column) => nodeText(column.render(row, index)))
          .join(" ");

      return text.toLowerCase().includes(normalizedKeyword);
    });
  }, [activeFilterValues, data, filterDefinitions, normalizedKeyword, resolvedColumns, searchEnabled, searchText]);
  const total = filteredData.length;
  const totalPages = Math.max(1, Math.ceil(total / currentPageSize));
  const page = Math.min(currentPage, totalPages);
  const paginationEnabled = showPagination;
  const start = paginationEnabled ? (page - 1) * currentPageSize : 0;
  const end = paginationEnabled ? Math.min(start + currentPageSize, total) : total;
  const displayStart = total > 0 ? start + 1 : 0;
  const visibleRows = useMemo(() => filteredData.slice(start, end), [filteredData, start, end]);
  const hasTableRefinements = normalizedKeyword !== "" || Object.values(activeFilterValues).some(activeFilterValueIsActive);
  const normalizedColumnSettingsSignature = columnSettingsSignature(normalizedColumnSettings);

  useEffect(() => {
    setCurrentPageSize(pageSize);
  }, [pageSize]);

  useEffect(() => {
    setColumnSettingsState(readColumnSettings(resolvedColumnSettingsKey));
    setColumnSettingsOpen(false);
  }, [resolvedColumnSettingsKey]);

  useEffect(() => {
    if (!columnSettingsEnabled) return;
    setColumnSettingsState((current) => {
      const normalized = normalizeColumnSettings(columns, current);
      return settingsEqual(current, normalized) ? current : normalized;
    });
  }, [columnSettingsEnabled, columnSignature]);

  useEffect(() => {
    if (!resolvedColumnSettingsKey) return;
    writeColumnSettings(resolvedColumnSettingsKey, normalizedColumnSettings);
  }, [resolvedColumnSettingsKey, normalizedColumnSettingsSignature]);

  useEffect(() => {
    if (!columnSettingsOpen) return;

    function closeByPointer(event: MouseEvent) {
      const target = event.target;
      if (target instanceof Node && columnSettingsRef.current?.contains(target)) {
        return;
      }
      setColumnSettingsOpen(false);
    }

    function closeByKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setColumnSettingsOpen(false);
      }
    }

    function closePanel() {
      setColumnSettingsOpen(false);
    }

    window.addEventListener("mousedown", closeByPointer);
    window.addEventListener("keydown", closeByKey);
    window.addEventListener("resize", closePanel);
    window.addEventListener("blur", closePanel);
    return () => {
      window.removeEventListener("mousedown", closeByPointer);
      window.removeEventListener("keydown", closeByKey);
      window.removeEventListener("resize", closePanel);
      window.removeEventListener("blur", closePanel);
    };
  }, [columnSettingsOpen]);

  useEffect(() => {
    if (!headerFilterMenu) return;

    function closeByPointer(event: PointerEvent) {
      const target = event.target;
      if (target instanceof Node && headerFilterMenuRef.current?.contains(target)) {
        return;
      }
      if (target instanceof Element) {
        const interactiveFilterSurface = target.closest(
          ".data-table-column-filter-trigger, .data-table-filter-popover, .data-table-range-panel, .data-table-date-time-popover"
        );
        if (interactiveFilterSurface) {
          return;
        }
      }
      setHeaderFilterMenu(null);
      setHeaderFilterKeyword("");
    }

    function closeByKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setHeaderFilterMenu(null);
        setHeaderFilterKeyword("");
      }
    }

    function closePanel() {
      setHeaderFilterMenu(null);
      setHeaderFilterKeyword("");
    }

    window.addEventListener("pointerdown", closeByPointer);
    window.addEventListener("resize", closePanel);
    window.addEventListener("keydown", closeByKey);
    return () => {
      window.removeEventListener("pointerdown", closeByPointer);
      window.removeEventListener("resize", closePanel);
      window.removeEventListener("keydown", closeByKey);
    };
  }, [headerFilterMenu]);

  useEffect(() => {
    if (currentPage > totalPages) {
      setCurrentPage(totalPages);
    }
  }, [currentPage, totalPages]);

  useEffect(() => {
    setActiveFilterValues((current) => {
      const availableKeys = new Set(filterDefinitions.map((item) => item.key));
      let changed = false;
      const next: ActiveFilterValues = {};
      Object.entries(current).forEach(([key, value]) => {
        if (availableKeys.has(key) && activeFilterValueIsActive(value)) {
          next[key] = value;
        } else {
          changed = true;
        }
      });
      return changed ? next : current;
    });
  }, [filterDefinitions]);

  useEffect(() => {
    setCurrentPage(1);
  }, [normalizedKeyword, activeFilterSignature]);

  useEffect(() => {
    setPageInput(String(page));
  }, [page]);

  function changePageSize(value: string) {
    setCurrentPageSize(Number(value));
    setCurrentPage(1);
  }

  function jumpToPage(value = pageInput) {
    const next = Number(value);
    if (!Number.isFinite(next) || next < 1) {
      setPageInput(String(page));
      return;
    }

    const target = Math.min(totalPages, Math.max(1, Math.trunc(next)));
    setCurrentPage(target);
    setPageInput(String(target));
  }

  function changeFilter(key: string, value: string) {
    setActiveFilterValues((current) => ({ ...current, [key]: value }));
  }

  function toggleFilterValue(key: string, value: string) {
    setActiveFilterValues((current) => {
      const values = activeFilterValueList(current[key]);
      const nextValues = values.includes(value)
        ? values.filter((item) => item !== value)
        : [...values, value];
      return { ...current, [key]: nextValues };
    });
  }

  function changeDateRange(key: string, value: string) {
    setActiveFilterValues((current) => ({ ...current, [key]: value }));
  }

  function clearFilter(key: string) {
    setActiveFilterValues((current) => ({ ...current, [key]: "" }));
  }

  function clearColumnFilters(filters: AutoFilterDefinition<T>[]) {
    setActiveFilterValues((current) => {
      const next = { ...current };
      filters.forEach((filter) => {
        next[filter.key] = "";
      });
      return next;
    });
  }

  function filterActive(filter: AutoFilterDefinition<T>) {
    return activeFilterValueIsActive(activeFilterValues[filter.key]);
  }

  function openHeaderFilterMenu(event: ReactMouseEvent<HTMLElement>, column: DataTableColumn<T>) {
    const filters = columnFilters.get(column.key) || [];
    if (!filters.length) return;
    event.preventDefault();
    event.stopPropagation();
    const rect = event.currentTarget.getBoundingClientRect();
    setHeaderFilterKeyword("");
    setHeaderFilterMenu((current) => (
      current?.columnKey === column.key
        ? null
        : {
          columnKey: column.key,
          columnTitle: columnTitleText(column),
          position: { x: rect.left, y: rect.bottom + 6 }
        }
    ));
  }

  function renderFilterControl(filter: AutoFilterDefinition<T>, compact = false) {
    if (filter.kind === "dateRange") {
      return (
        <DateRangeFilter
          key={filter.key}
          label={filter.label}
          value={activeFilterSingleValue(activeFilterValues[filter.key])}
          onChange={(nextValue) => changeDateRange(filter.key, nextValue)}
        />
      );
    }

    return (
      <div className="data-table-filter data-table-filter-trigger" key={filter.key}>
        <span className="data-table-filter-label">
          <Filter size={compact ? 12 : 14} />
          <span>{filter.label}</span>
        </span>
        <SelectInput
          className="data-table-filter-select"
          aria-label={filter.label}
          value={activeFilterSingleValue(activeFilterValues[filter.key])}
          onChange={(event) => changeFilter(filter.key, event.target.value)}
        >
          <option value="">全部{filter.label}</option>
          {filter.options.map((option) => (
            <option key={option.value} value={option.value}>{option.label}</option>
          ))}
        </SelectInput>
      </div>
    );
  }

  function renderHeaderFilterMenu() {
    if (!headerFilterMenu) return null;
    const filters = columnFilters.get(headerFilterMenu.columnKey) || [];
    if (!filters.length) return null;
    const column = resolvedColumns.find((item) => item.key === headerFilterMenu.columnKey);
    const columnFilterMultiple = Boolean(column?.filterMultiple ?? filterMultiple);
    const width = filters.length > 1 ? 300 : 248;
    const position = clampedHeaderFilterPosition(headerFilterMenu.position, width);
    const hasActiveFilters = filters.some(filterActive);
    const searchValue = headerFilterKeyword.trim().toLowerCase();
    const optionSearchEnabled = filters.some((filter) => filter.kind !== "dateRange" && filter.options.length > 0);
    const showSectionTitles = filters.length > 1;

    return createPortal(
      <div
        className={`data-table-filter-popover ${filters.length > 1 ? "is-multiple" : "is-single"} ${columnFilterMultiple ? "is-checkable" : ""}`}
        ref={headerFilterMenuRef}
        role="menu"
        aria-label={`筛选${headerFilterMenu.columnTitle}`}
        style={{ left: position.x, top: position.y, width }}
        onClick={(event) => event.stopPropagation()}
        onContextMenu={(event) => {
          event.preventDefault();
          event.stopPropagation();
        }}
      >
        <div className="data-table-filter-popover-body">
          {optionSearchEnabled ? (
            <label className="data-table-filter-popover-search">
              <Search size={13} />
              <input
                aria-label="筛选选项"
                autoFocus
                data-slot="data-table-filter-search-input"
                placeholder="输入关键字"
                value={headerFilterKeyword}
                onChange={(event) => setHeaderFilterKeyword(event.target.value)}
              />
            </label>
          ) : null}
          {hasActiveFilters ? (
            <div className="data-table-filter-popover-actions">
              <BareButton onClick={() => clearColumnFilters(filters)}>清除本列</BareButton>
            </div>
          ) : null}
          {filters.map((filter) => {
            if (filter.kind === "dateRange") {
              return (
                <div className="data-table-filter-popover-section" key={filter.key}>
                  <div className="data-table-filter-popover-section-title">{filter.label}</div>
                  {renderFilterControl(filter, true)}
                </div>
              );
            }

            const currentValue = activeFilterValues[filter.key];
            const selectedValues = activeFilterValueList(currentValue);
            const visibleOptions = searchValue
              ? filter.options.filter((option) => `${option.label} ${option.value}`.toLowerCase().includes(searchValue))
              : filter.options;
            return (
              <div className="data-table-filter-popover-section" key={filter.key}>
                {showSectionTitles ? (
                  <div className="data-table-filter-popover-section-title">{filter.label}</div>
                ) : null}
                <div className="data-table-filter-option-list">
                  <BareButton
                    active={selectedValues.length === 0}
                    className="data-table-filter-option"
                    onClick={() => clearFilter(filter.key)}
                  >
                    {columnFilterMultiple ? (
                      <span className="data-table-filter-option-check" aria-hidden="true">
                        {selectedValues.length === 0 ? <Check size={12} /> : null}
                      </span>
                    ) : null}
                    <span className="data-table-filter-option-label">全部{filter.label}</span>
                  </BareButton>
                  {visibleOptions.map((option) => {
                    const selected = columnFilterMultiple
                      ? selectedValues.includes(option.value)
                      : activeFilterSingleValue(currentValue) === option.value;
                    return (
                      <BareButton
                        active={selected}
                        className="data-table-filter-option"
                        key={option.value}
                        onClick={() => columnFilterMultiple ? toggleFilterValue(filter.key, option.value) : changeFilter(filter.key, option.value)}
                      >
                        {columnFilterMultiple ? (
                          <span className="data-table-filter-option-check" aria-hidden="true">
                            {selected ? <Check size={12} /> : null}
                          </span>
                        ) : null}
                        <span className="data-table-filter-option-label">{option.label}</span>
                      </BareButton>
                    );
                  })}
                  {visibleOptions.length === 0 ? (
                    <div className="data-table-filter-empty">无匹配选项</div>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>
      </div>,
      document.body
    );
  }

  function visibleRowText(row: T, index: number) {
    return resolvedColumns
      .filter((column) => column.key !== "actions")
      .map((column) => nodeText(column.render(row, index)).trim())
      .filter(Boolean)
      .join("\t");
  }

  function canMoveColumn(column: DataTableColumn<T>, direction: -1 | 1) {
    if (!isColumnReorderable(column)) return false;
    const order = normalizedColumnSettings.order;
    const index = order.indexOf(column.key);
    const target = index + direction;
    return index >= 0 && target >= 0 && target < order.length;
  }

  function moveColumn(column: DataTableColumn<T>, direction: -1 | 1) {
    if (!canMoveColumn(column, direction)) return;
    setColumnSettingsState((current) => {
      const normalized = normalizeColumnSettings(columns, current);
      const order = [...normalized.order];
      const index = order.indexOf(column.key);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= order.length) {
        return normalized;
      }
      [order[index], order[target]] = [order[target], order[index]];
      return normalizeColumnSettings(columns, { ...normalized, order });
    });
  }

  function toggleColumnHidden(column: DataTableColumn<T>) {
    if (!isColumnHideable(column)) return;
    setColumnSettingsState((current) => {
      const normalized = normalizeColumnSettings(columns, current);
      const hidden = new Set(normalized.hidden);
      if (hidden.has(column.key)) {
        hidden.delete(column.key);
      } else {
        hidden.add(column.key);
      }
      return normalizeColumnSettings(columns, { ...normalized, hidden: Array.from(hidden) });
    });
  }

  function resetColumnSettings() {
    if (resolvedColumnSettingsKey) {
      removeColumnSettings(resolvedColumnSettingsKey);
    }
    setColumnSettingsState(defaultColumnSettings(columns));
  }

  function resetSearchAndFilters() {
    setKeyword("");
    setActiveFilterValues({});
    setCurrentPage(1);
  }

  function searchFromMenu(text: string) {
    const nextText = text.trim();
    if (!nextText) return;
    setKeyword(nextText);
    setCurrentPage(1);
  }

  function copyFromTable(text: string, label = "内容") {
    if (!text) return;
    void copyTextToClipboard(text)
      .then(() => message.success(`${label}已复制`))
      .catch(() => message.error("复制失败"));
  }

  function refreshFromMenu() {
    if (!onRefresh || refreshDisabled) return;
    void Promise.resolve(onRefresh())
      .then(() => message.success("表格已刷新"))
      .catch(() => message.error("刷新失败"));
  }

  function openRowContextMenu(event: ReactMouseEvent, row: T, rowIndex: number, column?: DataTableColumn<T>) {
    event.preventDefault();
    event.stopPropagation();
    const rowElement = event.currentTarget instanceof HTMLElement
      ? event.currentTarget.closest("tr")
      : null;
    setRowMenu({
      row,
      rowIndex,
      columnKey: column?.key,
      columnTitle: column ? nodeText(column.title).trim() || "单元格" : undefined,
      cellText: column ? nodeText(column.render(row, rowIndex)).trim() : visibleRowText(row, rowIndex),
      position: { x: event.clientX, y: event.clientY },
      actions: collectRowActions(rowElement as HTMLTableRowElement | null)
    });
  }

  function runDefaultRowAction(event: ReactMouseEvent<HTMLTableRowElement>) {
    if (isInteractiveTarget(event.target)) return;
    const actions = collectRowActions(event.currentTarget);
    const primaryAction = actions.find((action) => !action.disabled && !action.danger);
    if (!primaryAction) return;
    event.preventDefault();
    clickRowAction(primaryAction);
  }

  const rowContextMenuItems = useMemo<ContextMenuItem[]>(() => {
    if (!rowMenu) return [];
    const currentRowKey = rowKey(rowMenu.row);
    const rowText = visibleRowText(rowMenu.row, rowMenu.rowIndex);
    const rowJson = stringifyRow(rowMenu.row);
    const helpers: DataTableContextMenuHelpers<T> = {
      row: rowMenu.row,
      rowIndex: rowMenu.rowIndex,
      rowKey: currentRowKey,
      rowText,
      rowJson,
      cellText: rowMenu.cellText,
      columnKey: rowMenu.columnKey,
      columnTitle: rowMenu.columnTitle,
      copyText: copyFromTable,
      searchText: searchFromMenu,
      clearSearchAndFilters: resetSearchAndFilters,
      refresh: refreshFromMenu
    };
    const customItems = rowContextMenu?.(rowMenu.row, rowMenu.rowIndex, helpers) || [];
    const actionItems: ContextMenuItem[] = rowMenu.actions.map((action) => ({
      key: `row-action-${action.key}`,
      label: action.label,
      icon: <MousePointerClick size={14} />,
      disabled: action.disabled,
      danger: action.danger,
      onSelect: () => clickRowAction(action)
    }));
    const businessItems = businessCopyMenuItems(rowMenu.row, copyFromTable);
    const usedLabels = new Set([...customItems, ...actionItems].map(contextMenuItemText).filter(Boolean));
    const uniqueBusinessItems = businessItems.filter((item) => !usedLabels.has(contextMenuItemText(item)));

    return [
      ...(customItems.length ? [...customItems, { key: "custom-separator", type: "separator" as const }] : []),
      ...(actionItems.length ? [...actionItems, { key: "row-action-separator", type: "separator" as const }] : []),
      ...(uniqueBusinessItems.length ? [...uniqueBusinessItems, { key: "business-copy-separator", type: "separator" as const }] : []),
      {
        key: "search-cell",
        label: rowMenu.columnTitle ? `搜索${rowMenu.columnTitle}` : "搜索单元格内容",
        icon: <Search size={14} />,
        disabled: !rowMenu.cellText,
        onSelect: () => searchFromMenu(rowMenu.cellText)
      },
      {
        key: "copy-cell",
        label: rowMenu.columnTitle ? `复制${rowMenu.columnTitle}` : "复制单元格",
        icon: <Copy size={14} />,
        disabled: !rowMenu.cellText,
        onSelect: () => copyFromTable(rowMenu.cellText, rowMenu.columnTitle || "单元格")
      },
      {
        key: "copy-row",
        label: "复制本行文本",
        icon: <FileText size={14} />,
        disabled: !rowText,
        onSelect: () => copyFromTable(rowText, "本行文本")
      },
      {
        key: "copy-row-json",
        label: "复制本行 JSON",
        icon: <FileText size={14} />,
        disabled: !rowJson,
        onSelect: () => copyFromTable(rowJson, "本行 JSON")
      },
      {
        key: "copy-row-key",
        label: "复制行 ID",
        icon: <KeyRound size={14} />,
        disabled: currentRowKey === undefined || currentRowKey === null || String(currentRowKey) === "",
        onSelect: () => copyFromTable(String(currentRowKey), "行 ID")
      },
      { key: "table-separator", type: "separator" as const },
      {
        key: "clear-table-refinements",
        label: "清除搜索/筛选",
        icon: <X size={14} />,
        disabled: !hasTableRefinements,
        onSelect: resetSearchAndFilters
      },
      {
        key: "refresh-table",
        label: refreshTitle,
        icon: <RefreshCw size={14} />,
        disabled: !onRefresh || refreshDisabled,
        onSelect: refreshFromMenu
      }
    ];
  }, [hasTableRefinements, onRefresh, refreshDisabled, refreshTitle, rowContextMenu, rowKey, rowMenu]);
  const tableContextMenuItems = useMemo<ContextMenuItem[]>(() => {
    const helpers: DataTableContextMenuTableHelpers = {
      copyText: copyFromTable,
      searchText: searchFromMenu,
      clearSearchAndFilters: resetSearchAndFilters,
      refresh: refreshFromMenu
    };
    const customItems = tableContextMenu?.(helpers) || [];
    return [
      ...(customItems.length ? [...customItems, { key: "table-custom-separator", type: "separator" as const }] : []),
      ...(columnSettingsEnabled ? [
        {
          key: "column-settings",
          label: columnSettingsLabel,
          icon: <Columns3 size={14} />,
          onSelect: () => setColumnSettingsOpen(true)
        },
        { key: "column-settings-separator", type: "separator" as const }
      ] : []),
      {
        key: "clear-table-refinements",
        label: "清除搜索/筛选",
        icon: <X size={14} />,
        disabled: !hasTableRefinements,
        onSelect: resetSearchAndFilters
      },
      {
        key: "refresh-table",
        label: refreshTitle,
        icon: <RefreshCw size={14} />,
        disabled: !onRefresh || refreshDisabled,
        onSelect: refreshFromMenu
      }
    ];
  }, [columnSettingsEnabled, columnSettingsLabel, hasTableRefinements, onRefresh, refreshDisabled, refreshTitle, tableContextMenu]);

  return (
    <div className="data-table">
      {title || headerLeftAction || headerAction || searchEnabled || toolbarFilters.length || onRefresh || columnSettingsEnabled ? (
        <div className="data-table-head">
          {title ? (
            <div className="data-table-title">
              <h3>{title}</h3>
            </div>
          ) : null}
          <div className="data-table-tools">
            <div className="data-table-filter-tools">
              {toolbarFilters.map((filter) => renderFilterControl(filter))}
              {searchEnabled ? (
                <IconField className="data-table-search" icon={<Search size={15} />} label={resolvedSearchPlaceholder}>
                  <TextInput
                    className="data-table-search-control"
                    value={keyword}
                    onChange={(event) => setKeyword(event.target.value)}
                    placeholder={resolvedSearchPlaceholder}
                    aria-label={resolvedSearchPlaceholder}
                  />
                </IconField>
              ) : null}
            </div>
            <div className="data-table-action-tools">
              {headerLeftAction}
              {onRefresh ? <UiButton className="data-table-refresh-button" icon={<RefreshCw size={15} />} disabled={refreshDisabled} onClick={() => void onRefresh()}>{refreshTitle}</UiButton> : null}
              {columnSettingsEnabled ? (
                <div className="data-table-column-settings" ref={columnSettingsRef}>
                  <UiButton
                    className="data-table-column-button"
                    icon={<Columns3 size={15} />}
                    aria-expanded={columnSettingsOpen}
                    aria-haspopup="menu"
                    onClick={() => setColumnSettingsOpen((open) => !open)}
                  >
                    {columnSettingsLabel}
                  </UiButton>
                  {columnSettingsOpen ? (
                    <div className="data-table-column-panel" role="menu" aria-label={columnSettingsLabel}>
                      <div className="data-table-column-panel-head">
                        <b>{columnSettingsLabel}</b>
                        <BareButton className="data-table-column-reset" onClick={resetColumnSettings}>
                          <RotateCcw size={14} />
                          <span>重置</span>
                        </BareButton>
                      </div>
                      <div className="data-table-column-list">
                        {orderedColumns.filter((column) => isColumnHideable(column) || isColumnReorderable(column)).map((column) => {
                          const titleText = columnTitleText(column);
                          const hidden = hiddenColumnKeys.has(column.key);
                          const hideable = isColumnHideable(column);
                          const reorderable = isColumnReorderable(column);
                          const disableHide = !hideable || (!hidden && visibleConfigurableColumnCount <= 1);

                          return (
                            <div className="data-table-column-item" role="menuitem" key={column.key}>
                              <label className="data-table-column-toggle">
                                <input
                                  type="checkbox"
                                  checked={!hidden}
                                  disabled={disableHide}
                                  onChange={() => toggleColumnHidden(column)}
                                />
                                <span>{titleText}</span>
                              </label>
                              <div className="data-table-column-move">
                                <IconButton
                                  icon={<ArrowUp size={14} />}
                                  label={`上移${titleText}`}
                                  disabled={!reorderable || !canMoveColumn(column, -1)}
                                  onClick={() => moveColumn(column, -1)}
                                />
                                <IconButton
                                  icon={<ArrowDown size={14} />}
                                  label={`下移${titleText}`}
                                  disabled={!reorderable || !canMoveColumn(column, 1)}
                                  onClick={() => moveColumn(column, 1)}
                                />
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : null}
              {headerAction}
            </div>
          </div>
        </div>
      ) : null}
      <div
        className="table-scroll"
        onContextMenu={(event) => {
          if (!onRefresh && !hasTableRefinements && !tableContextMenu && !columnSettingsEnabled) return;
          event.preventDefault();
          event.stopPropagation();
          setRowMenu(null);
          setTableMenuPosition({ x: event.clientX, y: event.clientY });
        }}
      >
        <table>
          <thead>
            <tr>
              {resolvedColumns.map((column) => {
                const filters = columnFilters.get(column.key) || [];
                const hasActiveColumnFilter = filters.some(filterActive);
                const thClassName = [
                  isActionColumn(column) ? "data-table-action-cell" : "",
                  filters.length ? "is-filterable" : "",
                  hasActiveColumnFilter ? "has-active-filter" : ""
                ].filter(Boolean).join(" ") || undefined;

                return (
                  <th
                    className={thClassName}
                    data-column-key={column.key}
                    key={column.key}
                    style={{ width: column.width, textAlign: column.align || "left" }}
                  >
                    <div className="data-table-th-content">
                      <span className="data-table-th-title">{column.title}</span>
                      {filters.length ? (
                        <BareButton
                          active={hasActiveColumnFilter}
                          className="data-table-column-filter-trigger"
                          aria-label={`筛选${columnTitleText(column)}`}
                          onClick={(event) => openHeaderFilterMenu(event, column)}
                        >
                          <Filter size={13} />
                        </BareButton>
                      ) : null}
                    </div>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row, index) => (
              <tr
                className="data-table-row-context-target"
                key={rowKey(row)}
                onContextMenu={(event) => openRowContextMenu(event, row, start + index)}
                onDoubleClick={runDefaultRowAction}
              >
                {resolvedColumns.map((column) => (
                  <td
                    className={isActionColumn(column) ? "data-table-action-cell" : undefined}
                    data-column-key={column.key}
                    key={column.key}
                    style={{ textAlign: column.align || "left" }}
                    onContextMenu={(event) => openRowContextMenu(event, row, start + index, column)}
                  >
                    {column.render(row, start + index)}
                  </td>
                ))}
              </tr>
            ))}
            {!total ? (
              <tr>
                <td className="data-table-empty" colSpan={Math.max(resolvedColumns.length, 1)}>{emptyText}</td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
      {paginationEnabled ? (
        <div className="data-table-pagination">
          <span className="pagination-info">
            {displayStart}-{end} / {total}
          </span>
          <Field className="pagination-size" label="每页">
            <SelectInput value={currentPageSize} onChange={(event) => changePageSize(event.target.value)}>
              {resolvedPageSizeOptions.map((item) => <option key={item} value={item}>{item}</option>)}
            </SelectInput>
          </Field>
          <div className="pagination-steps">
            <IconButton icon={<ChevronLeft size={16} />} label="上一页" disabled={page <= 1} onClick={() => setCurrentPage((value) => Math.max(1, value - 1))} />
            <Field className="pagination-jump" label="第">
              <TextInput
                value={pageInput}
                inputMode="numeric"
                aria-label="跳转页码"
                onBlur={() => jumpToPage()}
                onChange={(event) => setPageInput(event.target.value.replace(/\D/g, ""))}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    event.preventDefault();
                    jumpToPage();
                  }
                }}
              />
              <span>/ {totalPages} 页</span>
            </Field>
            <IconButton icon={<ChevronRight size={16} />} label="下一页" disabled={page >= totalPages} onClick={() => setCurrentPage((value) => Math.min(totalPages, value + 1))} />
          </div>
        </div>
      ) : null}
      {renderHeaderFilterMenu()}
      {rowMenu ? (
        <ContextMenu
          items={rowContextMenuItems}
          label="表格行快捷操作"
          position={rowMenu.position}
          width={204}
          onClose={() => setRowMenu(null)}
        />
      ) : null}
      {tableMenuPosition ? (
        <ContextMenu
          items={tableContextMenuItems}
          label="表格快捷操作"
          position={tableMenuPosition}
          width={188}
          onClose={() => setTableMenuPosition(null)}
        />
      ) : null}
    </div>
  );
}
