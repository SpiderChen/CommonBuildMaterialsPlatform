import {
  AlertCircle,
  ArrowLeft,
  ArrowDown,
  ArrowUp,
  BarChart3,
  Bell,
  Building2,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock,
  ClipboardCheck,
  Copy,
  CreditCard,
  Download,
  Factory,
  FileSignature,
  Filter,
  Link2,
  ListChecks,
  Map as MapIcon,
  MapPin,
  Package,
  Pencil,
  PlayCircle,
  Plus,
  Printer,
  ReceiptText,
  RefreshCw,
  Route,
  Scale,
  Search,
  ShieldCheck,
  Truck,
  Users,
  X
} from "lucide-react";
import L from "leaflet";
import { type ChangeEvent, type FormEvent, type MouseEvent as ReactMouseEvent, type ReactNode, useEffect, useMemo, useRef, useState } from "react";
import "leaflet/dist/leaflet.css";
import {
  ActionGroup,
  ActionDialogScope,
  BusinessWorkflowStatus,
  BusinessWorkflowTimeline,
  Button as UiButton,
  ButtonLink,
  Card,
  ChipList,
  ChipButton,
  ContextMenu,
  type ContextMenuItem,
  DataTable,
  type DataTableContextMenuHelpers,
  Dialog,
  DialogContent,
  DialogForm,
  EmptyState,
  Field,
  FormActions,
  HeroDateField,
  IconButton,
  IconField,
  InlineForm,
  KeyValueTable,
  KpiCard,
  LayoutRegion,
  nameOf,
  Panel,
  productLabel,
  QuickForm,
  MetricList,
  SectionGrid,
  SectionHeader,
  SelectInput,
  SelectableCard,
  SimpleTable,
  SplitRow,
  ScopedActionDialog as ActionDialog,
  StatusChip,
  statusLabel,
  SystemForm,
  TextAreaInput,
  TextInput,
  WorkflowForm,
  buildDataTableRowContextMenu,
  findBusinessWorkflowItems,
  useMessage,
  useMessageBox
} from "../components";
import { api } from "../services/api";
import { activeDictionaryOptions, dictionaryLabel, type DictionaryOption } from "../services/dictionaries";
import { hasPermission, permissionsForRole } from "../services/permissions";
import { browserFilePayload } from "../utils/filePayload";
import { sensitiveActionPrompt, type SensitiveActionPrompt } from "../utils/sensitiveActions";
import type {
  AuditLog,
  ApprovalFlow,
  ApprovalTask,
  BackupDrill,
  BackupInfo,
  BootstrapData,
  Carrier,
  CollectionDispatch,
  CollectionTemplate,
  Contract,
  ContractAttachment,
  Customer,
  CustomerBlacklist,
  CustomerComplaint,
  CustomerContact,
  CustomerProfile,
  DataDictionary,
  DashboardData,
  DeliveryNote,
  DeliverySign,
  DeliverySignAttachment,
  DeliverySignLink,
  DeviceCredential,
  DeviceProtocolFrame,
  DispatchOrder,
  DispatchSchedule,
  Driver,
  DispatchCenterOverview,
  DispatchCenterProductionTask,
  DispatchCenterQueueItem,
  DispatchCenterSiteProgress,
  DispatchCenterVehicle,
  FinanceOverview,
  FieldPolicy,
  GatewayOverview,
  GatewayRoute,
  GeoFence,
  GeoFenceEvent,
  Company,
  Department,
  IntegrationEndpoint,
  InventoryBatchTrace,
  InventoryItem,
  InventoryStocktake,
  InventoryTransfer,
  LatestLocation,
  LicensePackage,
  LocationBatchReportResponse,
  LocationReportPayload,
  MasterDataExport,
  MasterDataImportResult,
  ManagementReports,
  MapProviderConfig,
  MFAEnrollment,
  Material,
  MixDesign,
  MixDesignPlantProfile,
  ModuleInfo,
  NotificationItem,
  OIDCProvider,
  ProcurementOverview,
  OrganizationOverview,
  OrganizationNode,
  Payable,
  Payment,
  Plant,
  PlantBufferFlow,
  PlantBufferLocation,
  PortalOverview,
  PricePolicy,
  PricingQuote,
  Product,
  ProductionBatch,
  ProductionPlan,
  ProductionOverview,
  ProductionTask,
  PluginInfo,
  PluginRun,
  Project,
  QualityInspection,
  QualityOverview,
  QualitySample,
  RawMaterialInspection,
  Receivable,
  RuleDefinition,
  RuleOverview,
  Role,
  SalesOrder,
  SalesInvoice,
  ScaleDeviceEvent,
  ScaleTicket,
  ScaleWeightRecord,
  SecurityPolicy,
  Site,
  SCIMProvider,
  StockYard,
  StockYardFlow,
  StockYardPile,
  Statement,
  SupplierStatement,
  SystemBundle,
  TaxRate,
  TicketVoidLog,
  TrackReplay,
  TicketPrintLog,
  TransportSettlement,
  TransportSettlementItem,
  User,
  UpdatePackage,
  Vehicle,
  VehicleAlarm,
  VehicleDevice,
  VehicleLocationEvent,
  WorkflowDefinition,
  WorkflowCatalog,
  WorkflowCatalogEvent,
  WorkflowCatalogField,
  WorkflowCatalogTrigger,
  WorkflowEvent,
  WorkflowEventPreview,
  WorkflowDelivery,
  WorkflowInboxItem,
  WorkflowInstance,
  WorkflowLog,
  WorkflowOutbox,
  WorkflowSubscription,
  WorkflowTask,
  WorkflowOverview,
  IntegrationOverview
} from "../services/types";

export type ERPWorkbenchSection =
  | "overview"
  | "master-customers"
  | "customer-risk"
  | "master-projects"
  | "master-products"
  | "sales-pricing"
  | "master-materials"
  | "master-sites"
  | "master-plants"
  | "stock-yards"
  | "master-drivers"
  | "master-vehicles"
  | "master-carriers"
  | "portal-customer"
  | "portal-driver"
  | "orders"
  | "production"
  | "production-plans"
  | "production-tasks"
  | "production-batches"
  | "production-reports"
  | "dispatch"
  | "dispatch-schedules"
  | "dispatch-queue"
  | "map-center"
  | "weighbridge"
  | "delivery"
  | "delivery-signs"
  | "settlement"
  | "contracts"
  | "raw-material-receipts"
  | "inventory-transfers"
  | "inventory-stocktakes"
  | "raw-material-inspections"
  | "finance"
  | "finance-receivables"
  | "finance-invoices"
  | "finance-collections"
  | "finance-suppliers"
  | "finance-carriers"
  | "reports"
  | "approval-center"
  | "system-org"
  | "system-license"
  | "system-maintenance"
  | "system-gateway"
  | "system-security"
  | "system-identity"
  | "system-plugins"
  | "system-rules"
  | "system-integrations"
  | "system-menu"
  | "system-dictionaries"
  | "system-users"
  | "system-roles"
  | "system-workflows"
  | "system-audit";

export type WorkbenchMenuItem = {
  key: string;
  path: string;
  name: string;
  label: string;
  group: string;
  groupLabel: string;
  icon: string;
  sort: number;
  permissionMark: string;
  affix: boolean;
  noCache: boolean;
  breadcrumb: boolean;
};

type WorkbenchData = {
  dashboard: DashboardData | null;
  reports: ManagementReports | null;
  dispatch: DispatchCenterOverview | null;
  portal: PortalOverview | null;
  production: ProductionOverview | null;
  procurement: ProcurementOverview | null;
  quality: QualityOverview | null;
  finance: FinanceOverview | null;
  rules: RuleOverview | null;
  integrations: IntegrationOverview | null;
  org: OrganizationOverview | null;
  system: SystemBundle | null;
  modules: ModuleInfo[];
  systemUsers: User[];
  systemRoles: Role[];
  auditLogs: AuditLog[];
  contracts: Contract[];
  orders: SalesOrder[];
  dispatchOrders: DispatchOrder[];
  dispatchSchedules: DispatchSchedule[];
  carrierSettlements: TransportSettlement[];
  carrierSettlementItems: TransportSettlementItem[];
  latestLocations: LatestLocation[];
  alarms: VehicleAlarm[];
  geoFences: GeoFence[];
  tickets: ScaleTicket[];
  ticketPrintLogs: TicketPrintLog[];
  ticketVoidLogs: TicketVoidLog[];
  weightRecords: ScaleWeightRecord[];
  scaleDeviceEvents: ScaleDeviceEvent[];
  deliveryNotes: DeliveryNote[];
  signs: DeliverySign[];
  signLinks: DeliverySignLink[];
  signAttachments: DeliverySignAttachment[];
  portalComplaints: CustomerComplaint[];
  statements: Statement[];
  approvals: ApprovalTask[];
};

type MasterKind = "customer" | "project" | "product" | "material" | "site" | "plant" | "driver" | "vehicle" | "carrier";
type MasterRecord = Customer | Project | Product | Material | Site | Plant | Driver | Vehicle | Carrier;
type OrganizationTreeRow = OrganizationNode & { depth: number };
type ProductionDialogMode = "detail" | "create-plan" | "edit-plan" | "tasks" | "batch" | "report" | "cancel-plan";
type ProductionPlanAction = {
  label: string;
  mode: ProductionDialogMode;
  icon: ReactNode;
  variant?: "primary" | "soft";
  disabled?: boolean;
};

function isCustomerStatementClosed(status?: string) {
  return status === "confirmed" || status === "invoiced";
}

function customerStatementClosedLabel(status?: string) {
  if (status === "invoiced") return "已开票";
  if (status === "confirmed") return "已确认";
  return "";
}

type WorkflowPresetDraft = {
  label: string;
  code: string;
  name: string;
  resource: string;
  eventType: string;
  conditions: string;
  steps: string;
  description?: string;
  variables?: WorkflowCatalogField[];
  triggers?: WorkflowCatalogTrigger[];
};

const workflowConditionOperatorOptions = [
  { value: "equals", label: "等于" },
  { value: "not_equals", label: "不等于" },
  { value: "contains", label: "包含" },
  { value: "not_contains", label: "不包含" },
  { value: "greater_than", label: "大于" },
  { value: "greater_or_equal", label: "大于等于" },
  { value: "less_than", label: "小于" },
  { value: "less_or_equal", label: "小于等于" },
  { value: "exists", label: "有值" },
  { value: "missing", label: "无值" }
];

const dictionaryTypePresets = [
  { type: "product_line", label: "产品类型" },
  { type: "dispatch_status", label: "调度状态" },
  { type: "ticket_type", label: "磅单类型" },
  { type: "invoice_type", label: "发票类型" },
  { type: "quality_status", label: "质量状态" },
  { type: "quality_result", label: "质检结论" },
  { type: "yard_type", label: "堆场类型" },
  { type: "buffer_type", label: "筒仓类型" },
  { type: "vehicle_type", label: "车辆类型" },
  { type: "carrier_settle_mode", label: "承运结算方式" },
  { type: "plant_status", label: "生产线状态" },
  { type: "resource_status", label: "资源状态" },
  { type: "shift_type", label: "生产班次" },
  { type: "delivery_channel", label: "签收通道" },
  { type: "payment_method", label: "付款方式" },
  { type: "org_company_level", label: "公司层级" },
  { type: "data_scope", label: "数据范围" },
  { type: "config_status", label: "配置状态" },
  { type: "account_status", label: "账号状态" },
  { type: "laboratory_test_type", label: "实验类型" },
  { type: "severity_level", label: "异常等级" },
  { type: "sample_source_type", label: "样品来源" },
  { type: "settlement_status", label: "结算状态" }
];

const carrierSettleModeFallbackOptions: DictionaryOption[] = [
  { code: "monthly", label: "月结" },
  { code: "per_trip", label: "按趟结算" },
  { code: "per_ton", label: "按吨结算" }
];

const organizationCompanyKinds = ["headquarters", "regional", "subsidiary", "company"];

const emptyData: WorkbenchData = {
  dashboard: null,
  reports: null,
  dispatch: null,
  portal: null,
  production: null,
  procurement: null,
  quality: null,
  finance: null,
  rules: null,
  integrations: null,
  org: null,
  system: null,
  modules: [],
  systemUsers: [],
  systemRoles: [],
  auditLogs: [],
  contracts: [],
  orders: [],
  dispatchOrders: [],
  dispatchSchedules: [],
  carrierSettlements: [],
  carrierSettlementItems: [],
  latestLocations: [],
  alarms: [],
  geoFences: [],
  tickets: [],
  ticketPrintLogs: [],
  ticketVoidLogs: [],
  weightRecords: [],
  scaleDeviceEvents: [],
  deliveryNotes: [],
  signs: [],
  signLinks: [],
  signAttachments: [],
  portalComplaints: [],
  statements: [],
  approvals: []
};

const today = new Date().toISOString().slice(0, 10);

function list<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

function money(value: number | undefined) {
  return Math.round(value || 0).toLocaleString();
}

function fieldNumber(value: string | number | undefined, fallbackValue = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallbackValue;
}

function qty(value: number | undefined) {
  return Number(value || 0).toLocaleString();
}

function percent(value: number | undefined) {
  return `${Math.round(value || 0)}%`;
}

function firstId(items: Array<{ id: number }> | undefined) {
  return items?.[0]?.id || 0;
}

function recordId(item: unknown) {
  if (item && typeof item === "object" && "id" in item) {
    const value = (item as { id?: unknown }).id;
    return typeof value === "number" ? value : fieldNumber(String(value || ""));
  }
  return 0;
}

function recordName(item: unknown, fallbackValue: string) {
  if (item && typeof item === "object") {
    const value = item as { name?: unknown; supplierName?: unknown; orderNo?: unknown };
    const name = value.name || value.supplierName || value.orderNo;
    return typeof name === "string" && name ? name : fallbackValue;
  }
  return fallbackValue;
}

function shortDateTime(value: string | undefined) {
  return value ? value.slice(5, 16) : "未计划";
}

function apiDateTime(value: string) {
  if (!value) return "";
  const normalized = value.replace("T", " ");
  return normalized.length === 16 ? `${normalized}:00` : normalized;
}

function dispatchSearchText(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation) {
  return Object.values(item).join(" ").toLowerCase();
}

function matchesDispatchSearch(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation, keyword: string) {
  return keyword === "" || dispatchSearchText(item).includes(keyword.toLowerCase());
}

function dispatchStatusValues(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation) {
  const row = item as Record<string, unknown>;
  return ["status", "onlineStatus", "transportStatus", "availableStatus"].map((field) => row[field]).filter((value): value is string => typeof value === "string" && value !== "");
}

function matchesDispatchStatus(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation, status: string) {
  return status === "all" || dispatchStatusValues(item).includes(status);
}

function dispatchStatusOptions(items: Array<DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation>) {
  return Array.from(new Set(items.flatMap(dispatchStatusValues))).sort((a, b) => statusLabel(a).localeCompare(statusLabel(b), "zh-CN"));
}

function dispatchSiteOptions(progress: DispatchCenterSiteProgress[]) {
  const sites = new Map<number, string>();
  progress.forEach((item) => {
    if (item.siteId) {
      sites.set(item.siteId, item.siteName || `站点 ${item.siteId}`);
    }
  });
  return Array.from(sites, ([id, name]) => ({ id, name }));
}

function dispatchStageClass(status: string | undefined) {
  switch (status) {
    case "assigned":
    case "accepted":
    case "arrived_site":
    case "waiting_load":
      return "queue";
    case "loading":
      return "loading";
    case "loaded":
    case "departed":
    case "in_transit":
      return "transit";
    case "arrived":
    case "arrived_project":
    case "unloading":
    case "signed":
      return "arrive";
    default:
      return "ready";
  }
}

function dispatchRoutePosition(status: string | undefined) {
  switch (status) {
    case "assigned":
    case "accepted":
      return 12;
    case "arrived_site":
    case "waiting_load":
      return 18;
    case "loading":
      return 24;
    case "loaded":
      return 34;
    case "departed":
      return 44;
    case "in_transit":
      return 66;
    case "arrived_project":
      return 84;
    case "unloading":
    case "signed":
      return 92;
    default:
      return 8;
  }
}

function dispatchLaneVehicleGroups(items: DispatchCenterQueueItem[]) {
  const sorted = [...items].sort((a, b) => dispatchRoutePosition(a.status) - dispatchRoutePosition(b.status));
  const groups: Array<{ key: string; position: number; lastPosition: number; items: DispatchCenterQueueItem[] }> = [];

  sorted.forEach((item) => {
    const position = dispatchRoutePosition(item.status);
    const last = groups[groups.length - 1];
    if (last && position - last.lastPosition <= 8) {
      last.items.push(item);
      last.lastPosition = position;
      last.position = Math.round(last.items.reduce((sum, current) => sum + dispatchRoutePosition(current.status), 0) / last.items.length);
      return;
    }
    groups.push({ key: String(item.dispatchId), position, lastPosition: position, items: [item] });
  });

  return groups.map((group) => ({ key: group.key, position: group.position, items: group.items }));
}

function etaText(item: DispatchCenterQueueItem) {
  if (!item.etaMinutes) {
    return shortDateTime(item.eta || item.plannedEta);
  }
  return `${Math.round(item.etaMinutes)} 分钟`;
}

const fallbackMapCenter: [number, number] = [30.5728, 104.0668];
const siteFenceRadiusMin = 50;
const siteFenceRadiusMax = 3000;
const siteFenceRadiusStep = 50;
const siteFenceRangeOptions = [
  { label: "小", value: 150 },
  { label: "标准", value: 300 },
  { label: "较大", value: 600 },
  { label: "很大", value: 1000 }
];

type SiteAddressLookupState = "" | "loading" | "success" | "failed";

type ReverseGeocodeResult = {
  display_name?: string;
  name?: string;
  address?: Record<string, string | undefined>;
};

function isValidLocation(item: LatestLocation) {
  return Number.isFinite(item.latitude) && Number.isFinite(item.longitude) && Math.abs(item.latitude) <= 90 && Math.abs(item.longitude) <= 180;
}

function isValidCoordinate(longitude: number, latitude: number) {
  return Number.isFinite(latitude) && Number.isFinite(longitude) && Math.abs(latitude) <= 90 && Math.abs(longitude) <= 180 && (latitude !== 0 || longitude !== 0);
}

function normalizedFenceRadius(value: number | string | undefined) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 300;
}

function editableFenceRadius(value: number | string | undefined) {
  return Math.min(siteFenceRadiusMax, Math.max(siteFenceRadiusMin, Math.round(normalizedFenceRadius(value) / siteFenceRadiusStep) * siteFenceRadiusStep));
}

function siteFenceRangeLabel(radius: number) {
  if (radius <= 200) return "小";
  if (radius <= 450) return "标准";
  if (radius <= 800) return "较大";
  return "很大";
}

function compactAddressParts(parts: Array<string | undefined>) {
  return parts
    .map((part) => (part || "").trim())
    .filter(Boolean)
    .filter((part, index, list) => list.indexOf(part) === index);
}

function formatReverseGeocodeAddress(result: ReverseGeocodeResult) {
  const address = result.address || {};
  const composed = compactAddressParts([
    address.state,
    address.city || address.town || address.village,
    address.city_district || address.district || address.county,
    address.suburb || address.neighbourhood,
    address.road,
    address.house_number || address.building || result.name
  ]);
  if (composed.length >= 2) return composed.join("");
  return (result.display_name || result.name || "").trim();
}

async function reverseGeocodeAddress(longitude: number, latitude: number, signal: AbortSignal) {
  const url = new URL("https://nominatim.openstreetmap.org/reverse");
  url.searchParams.set("format", "jsonv2");
  url.searchParams.set("lat", String(latitude));
  url.searchParams.set("lon", String(longitude));
  url.searchParams.set("zoom", "18");
  url.searchParams.set("addressdetails", "1");
  url.searchParams.set("accept-language", "zh-CN,zh,en");
  const response = await fetch(url.toString(), {
    headers: { Accept: "application/json" },
    signal
  });
  if (!response.ok) {
    throw new Error("reverse geocode failed");
  }
  return formatReverseGeocodeAddress(await response.json() as ReverseGeocodeResult);
}

function escapeHtml(value: string) {
  return value.replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "\"": "&quot;",
    "'": "&#39;"
  }[char] || char));
}

function providerSubdomains(provider: MapProviderConfig | null) {
  return provider?.subdomains?.length ? provider.subdomains : ["a", "b", "c"];
}

function VehicleLocationMap({
  locations,
  fences,
  provider,
  selectedVehicleId,
  onSelectVehicle
}: {
  locations: LatestLocation[];
  fences: GeoFence[];
  provider: MapProviderConfig | null;
  selectedVehicleId: number | null;
  onSelectVehicle: (vehicleId: number) => void;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markerLayerRef = useRef<L.LayerGroup | null>(null);
  const fenceLayerRef = useRef<L.LayerGroup | null>(null);
  const tileLayerRef = useRef<L.TileLayer | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;
    const map = L.map(containerRef.current, {
      attributionControl: false,
      zoomControl: false
    }).setView(fallbackMapCenter, 11);
    L.control.zoom({ position: "bottomright" }).addTo(map);
    L.control.attribution({ prefix: false, position: "bottomleft" }).addTo(map);
    fenceLayerRef.current = L.layerGroup().addTo(map);
    markerLayerRef.current = L.layerGroup().addTo(map);
    mapRef.current = map;
    const resizeTimer = window.setTimeout(() => map.invalidateSize(), 0);
    const resizeObserver = new ResizeObserver(() => {
      window.requestAnimationFrame(() => map.invalidateSize());
    });
    resizeObserver.observe(containerRef.current);
    return () => {
      window.clearTimeout(resizeTimer);
      resizeObserver.disconnect();
      map.remove();
      mapRef.current = null;
      markerLayerRef.current = null;
      fenceLayerRef.current = null;
      tileLayerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    if (tileLayerRef.current) {
      tileLayerRef.current.removeFrom(map);
    }
    const tileUrl = provider?.tileUrl || "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png";
    tileLayerRef.current = L.tileLayer(tileUrl, {
      attribution: provider?.attribution || "OpenStreetMap",
      maxZoom: provider?.maxZoom || 18,
      minZoom: provider?.minZoom || 3,
      subdomains: providerSubdomains(provider)
    }).addTo(map);
  }, [provider]);

  useEffect(() => {
    const map = mapRef.current;
    const markerLayer = markerLayerRef.current;
    const fenceLayer = fenceLayerRef.current;
    if (!map || !markerLayer || !fenceLayer) return;
    markerLayer.clearLayers();
    fenceLayer.clearLayers();
    const validLocations = locations.filter(isValidLocation);
    const selected = validLocations.find((item) => item.vehicleId === selectedVehicleId);
    const bounds = L.latLngBounds([]);

    fences.filter((item) => item.status !== "inactive").forEach((fence) => {
      const color = fence.type === "site" ? "#2458a6" : "#087f7b";
      if (fence.shape === "polygon" && fence.polygon?.length >= 3) {
        const points = fence.polygon
          .filter((point) => isValidCoordinate(point.longitude, point.latitude))
          .map((point) => [point.latitude, point.longitude] as [number, number]);
        if (points.length < 3) return;
        const polygon = L.polygon(points, {
          color,
          fillColor: color,
          fillOpacity: 0.12,
          weight: 2
        }).bindTooltip(escapeHtml(fence.name || "电子围栏"), { direction: "top" });
        polygon.addTo(fenceLayer);
        bounds.extend(polygon.getBounds());
        return;
      }
      if (!isValidCoordinate(fence.longitude, fence.latitude)) return;
      const circle = L.circle([fence.latitude, fence.longitude], {
        radius: normalizedFenceRadius(fence.radius),
        color,
        fillColor: color,
        fillOpacity: 0.1,
        weight: 2
      }).bindTooltip(`${escapeHtml(fence.name || "电子围栏")} · ${qty(normalizedFenceRadius(fence.radius))}m`, { direction: "top" });
      circle.addTo(fenceLayer);
      bounds.extend(circle.getBounds());
    });

    validLocations.forEach((item) => {
      const selectedClass = item.vehicleId === selectedVehicleId ? " selected" : "";
      const marker = L.marker([item.latitude, item.longitude], {
        icon: L.divIcon({
          className: "vehicle-map-marker-shell",
          html: `<span class="vehicle-map-marker${selectedClass}"><span>${escapeHtml(item.plateNo)}</span></span>`,
          iconAnchor: [42, 32],
          iconSize: [84, 34]
        })
      });
      marker
        .bindTooltip(`${escapeHtml(item.plateNo)} · ${qty(item.speed)} km/h · ${shortDateTime(item.lastLocationTime)}`, { direction: "top" })
        .on("click", () => onSelectVehicle(item.vehicleId))
        .addTo(markerLayer);
      bounds.extend([item.latitude, item.longitude]);
    });

    if (selected) {
      map.setView([selected.latitude, selected.longitude], Math.max(map.getZoom(), 13), { animate: true });
      return;
    }
    if (bounds.isValid() && validLocations.length === 1 && fences.length === 0) {
      const [item] = validLocations;
      map.setView([item.latitude, item.longitude], 13, { animate: true });
      return;
    }
    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 13 });
      return;
    }
    map.setView(fallbackMapCenter, 10, { animate: true });
  }, [fences, locations, onSelectVehicle, selectedVehicleId]);

  return <div className="vehicle-leaflet-map" ref={containerRef} />;
}

function SiteFenceMap({
  longitude,
  latitude,
  radius = 0,
  provider,
  onCenterChange
}: {
  longitude: number;
  latitude: number;
  radius?: number;
  provider: MapProviderConfig | null;
  onCenterChange: (longitude: number, latitude: number) => void;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);
  const tileLayerRef = useRef<L.TileLayer | null>(null);
  const onCenterChangeRef = useRef(onCenterChange);

  useEffect(() => {
    onCenterChangeRef.current = onCenterChange;
  }, [onCenterChange]);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;
    const map = L.map(containerRef.current, {
      attributionControl: false,
      zoomControl: false
    }).setView(fallbackMapCenter, 11);
    L.control.zoom({ position: "bottomright" }).addTo(map);
    layerRef.current = L.layerGroup().addTo(map);
    map.on("click", (event) => {
      onCenterChangeRef.current(Number(event.latlng.lng.toFixed(6)), Number(event.latlng.lat.toFixed(6)));
    });
    mapRef.current = map;
    const resizeTimer = window.setTimeout(() => map.invalidateSize(), 0);
    const resizeObserver = new ResizeObserver(() => {
      window.requestAnimationFrame(() => map.invalidateSize());
    });
    resizeObserver.observe(containerRef.current);
    return () => {
      window.clearTimeout(resizeTimer);
      resizeObserver.disconnect();
      map.remove();
      mapRef.current = null;
      layerRef.current = null;
      tileLayerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    if (tileLayerRef.current) {
      tileLayerRef.current.removeFrom(map);
    }
    const tileUrl = provider?.tileUrl || "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png";
    tileLayerRef.current = L.tileLayer(tileUrl, {
      attribution: provider?.attribution || "OpenStreetMap",
      maxZoom: provider?.maxZoom || 18,
      minZoom: provider?.minZoom || 3,
      subdomains: providerSubdomains(provider)
    }).addTo(map);
  }, [provider]);

  useEffect(() => {
    const map = mapRef.current;
    const layer = layerRef.current;
    if (!map || !layer) return;
    layer.clearLayers();
    if (!isValidCoordinate(longitude, latitude)) {
      map.setView(fallbackMapCenter, 10, { animate: true });
      return;
    }
    const center: [number, number] = [latitude, longitude];
    L.marker(center).addTo(layer);
    if (radius > 0) {
      const circle = L.circle(center, {
        radius: normalizedFenceRadius(radius),
        color: "#2458a6",
        fillColor: "#2458a6",
        fillOpacity: 0.1,
        weight: 2
      }).addTo(layer);
      map.fitBounds(circle.getBounds().pad(0.2), { animate: true, maxZoom: 16 });
      return;
    }
    map.setView(center, Math.max(map.getZoom(), 15), { animate: true });
  }, [latitude, longitude, radius]);

  return <div className="site-fence-map" ref={containerRef} tabIndex={0} />;
}

function receivableBalance(finance: FinanceOverview | null) {
  return list(finance?.receivables).reduce((sum, item) => sum + Math.max(0, item.amount - item.receivedAmount), 0);
}

export function ERPWorkbenchView({
  bootstrap,
  menuItems,
  selectedSiteId,
  section,
  onChanged
}: {
  bootstrap: BootstrapData | null;
  menuItems: WorkbenchMenuItem[];
  selectedSiteId: number;
  section: ERPWorkbenchSection;
  onChanged: () => void;
}) {
  const [data, setData] = useState<WorkbenchData>(emptyData);
  const [mapConfig, setMapConfig] = useState<MapProviderConfig | null>(null);
  const [error, setError] = useState("");
const [loading, setLoading] = useState(true);
const [siteFilter, setSiteFilter] = useState("all");
const [dispatchSearch, setDispatchSearch] = useState("");
const [dispatchStatusFilter, setDispatchStatusFilter] = useState("all");
const [dispatchVehicleGroup, setDispatchVehicleGroup] = useState("queue");
const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null);
  const [selectedVehicleId, setSelectedVehicleId] = useState<number | null>(null);
  const [geoFenceForm, setGeoFenceForm] = useState({
    id: "",
    name: "",
    type: "site",
    siteId: "",
    projectId: "",
    longitude: "",
    latitude: "",
    radius: "300",
    shape: "circle",
    polygon: "",
    status: "active"
  });
  const [geoFenceEvents, setGeoFenceEvents] = useState<GeoFenceEvent[]>([]);
  const [trackReplay, setTrackReplay] = useState<TrackReplay | null>(null);
  const [trackEvents, setTrackEvents] = useState<VehicleLocationEvent[]>([]);
  const [locationBatchForm, setLocationBatchForm] = useState({
    deviceNo: "",
    plateNo: "",
    longitude: "",
    latitude: "",
    speed: "",
    direction: "",
    mileage: "",
    accStatus: "1",
    sourceType: "erp-console"
  });
  const [locationBatchResult, setLocationBatchResult] = useState<LocationBatchReportResponse | null>(null);
  const [dispatchQty, setDispatchQty] = useState("");
  const [dispatchScheduleForm, setDispatchScheduleForm] = useState({
    siteId: "",
    vehicleId: "",
    driverId: "",
    carrierId: "",
    shiftDate: today,
    shift: "早班",
    capacityQty: "36",
    status: "active"
  });
  const [carrierSettlementForm, setCarrierSettlementForm] = useState({
    carrierId: "",
    period: today.slice(0, 7),
    ratePerTrip: "",
    ratePerUnit: ""
  });
  const [dispatchActionError, setDispatchActionError] = useState("");
  const [dispatchSubmitting, setDispatchSubmitting] = useState(false);
  const [dispatchDialogOpen, setDispatchDialogOpen] = useState(false);
  const [productionQueueOrder, setProductionQueueOrder] = useState<Record<number, number>>({});
  const [actionBusy, setActionBusy] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionDialogId, setActionDialogId] = useState<string | null>(null);
  const [collapsedOrgNodeIds, setCollapsedOrgNodeIds] = useState<Set<string>>(() => new Set());
  const message = useMessage();
  const { showError, confirmMessage } = useMessageBox();
  const projectLocationPickerRef = useRef<HTMLDivElement | null>(null);
  const siteFencePickerRef = useRef<HTMLDivElement | null>(null);
  const reverseGeocodeRequestRef = useRef(0);
  const reverseGeocodeAbortRef = useRef<AbortController | null>(null);
  const [projectLocationDialogOpen, setProjectLocationDialogOpen] = useState(false);
  const [siteFenceDialogOpen, setSiteFenceDialogOpen] = useState(false);
  const [projectAddressLookupState, setProjectAddressLookupState] = useState<SiteAddressLookupState>("");
  const [siteAddressLookupState, setSiteAddressLookupState] = useState<SiteAddressLookupState>("");
  const [approvalComment, setApprovalComment] = useState("同意");
  const [editingMaster, setEditingMaster] = useState<{ kind: MasterKind; id: number } | null>(null);
  const [masterDialogKind, setMasterDialogKind] = useState<MasterKind | null>(null);
  const [orderDialogOpen, setOrderDialogOpen] = useState(false);
  const [productionDialogMode, setProductionDialogMode] = useState<ProductionDialogMode | null>(null);
  const [productionReportPlan, setProductionReportPlan] = useState<ProductionPlan | null>(null);
  const [productionReportBatches, setProductionReportBatches] = useState<ProductionBatch[] | null>(null);
  const [bufferDialogMode, setBufferDialogMode] = useState<"create" | "edit" | "transfer" | "adjust" | null>(null);
  const [plantCardMenu, setPlantCardMenu] = useState<{ plantId: number; x: number; y: number } | null>(null);
  const [yardDialogMode, setYardDialogMode] = useState<"yard" | "yard-edit" | "pile" | "pile-edit" | "receipt" | "adjust" | null>(null);
  const [stockYardDetailDialog, setStockYardDetailDialog] = useState<"flows" | "quality" | "stocktakes" | null>(null);
  const [stockYardMenu, setStockYardMenu] = useState<{ yardId: number; x: number; y: number } | null>(null);
  const [stockYardPileMenu, setStockYardPileMenu] = useState<{ pileId: number; x: number; y: number } | null>(null);
  const [printDeliveryNoteId, setPrintDeliveryNoteId] = useState<number | null>(null);
  const [masterForm, setMasterForm] = useState({
    customerName: "",
    customerContact: "",
    customerPhone: "",
    projectCustomerId: "",
    projectName: "",
    projectAddress: "",
    projectLongitude: "",
    projectLatitude: "",
    productName: "",
    productSpec: "",
    productPrice: "",
    materialName: "",
    materialSpec: "",
    materialSafeStock: "",
    siteName: "",
    siteCode: "",
    siteAddress: "",
    siteCompanyId: "",
    siteLongitude: "",
    siteLatitude: "",
    siteFenceRadius: "",
    plantSiteId: "",
    plantName: "",
    plantCode: "",
    plantCapacity: "",
    plantStatus: "",
    driverName: "",
    driverPhone: "",
    carrierName: "",
    carrierContact: "",
    carrierPhone: "",
    carrierSettleMode: "",
    carrierStatus: "",
    vehicleInternalNo: "",
    vehiclePlate: "",
    vehicleType: "",
    vehicleCapacity: "",
    vehicleDriverId: "",
    vehicleSiteId: "",
    vehicleDeviceNo: ""
  });
  const [customerContactForm, setCustomerContactForm] = useState({
    id: "",
    customerId: "",
    name: "",
    phone: "",
    role: "业务联系人",
    isDefault: "true"
  });
  const [customerProfileForm, setCustomerProfileForm] = useState({
    customerId: "",
    grade: "A",
    riskLevel: "low",
    creditScore: "90",
    tags: ""
  });
  const [customerBlacklistForm, setCustomerBlacklistForm] = useState({
    customerId: "",
    reason: "",
    scope: "sales_order",
    severity: "high",
    blockOrders: "true",
    blockDispatch: "false"
  });
  const [customerComplaintForm, setCustomerComplaintForm] = useState({
    customerId: "",
    projectId: "",
    title: "",
    content: "",
    level: "medium",
    owner: "",
    slaHours: "24",
    resolution: ""
  });
  const [portalComplaintForm, setPortalComplaintForm] = useState({
    customerId: "",
    projectId: "",
    title: "",
    content: "",
    level: "medium"
  });
  const [portalExceptionForm, setPortalExceptionForm] = useState({
    dispatchId: "",
    exception: "",
    level: "medium",
    alarmType: "delay"
  });
  const [taxRateForm, setTaxRateForm] = useState({
    id: "",
    name: "",
    rate: "0.06",
    scope: "sales",
    status: "active"
  });
  const [pricePolicyForm, setPricePolicyForm] = useState({
    id: "",
    customerId: "",
    projectId: "",
    productId: "",
    customerGrade: "A",
    region: "",
    minQuantity: "0",
    maxQuantity: "0",
    floorPrice: "",
    salePrice: "",
    promotionName: "",
    promotionType: "",
    promotionValue: "0",
    priority: "10",
    taxRateId: "",
    effectiveFrom: today,
    effectiveTo: "",
    status: "active"
  });
  const [pricingEvalForm, setPricingEvalForm] = useState({
    customerId: "",
    projectId: "",
    productId: "",
    planTime: today,
    planQuantity: "",
    unitPrice: ""
  });
  const [pricingQuote, setPricingQuote] = useState<PricingQuote | null>(null);
  const [masterBulkForm, setMasterBulkForm] = useState({
    resource: "customers",
    mode: "create",
    rowsJson: "[]"
  });
  const [masterExportResult, setMasterExportResult] = useState<MasterDataExport | null>(null);
  const [masterImportResult, setMasterImportResult] = useState<MasterDataImportResult | null>(null);
  const [bufferForm, setBufferForm] = useState({
    bufferId: "",
    plantId: "",
    yardPileId: "",
    code: "",
    name: "",
    type: "",
    materialId: "",
    capacity: "",
    unit: "t",
    warningQty: "",
    transferQty: "",
    actualQty: "",
    moistureRate: "",
    qualityStatus: "",
    status: "",
    remark: ""
  });
  const [yardForm, setYardForm] = useState({
    yardId: "",
    pileId: "",
    siteId: "",
    code: "",
    name: "",
    type: "",
    area: "",
    materialId: "",
    supplierId: "",
    batchNo: "",
    capacity: "",
    currentQty: "",
    warningQty: "",
    unit: "t",
    moistureRate: "",
    qualityStatus: "",
    status: "",
    receiptQty: "",
    actualQty: "",
    remark: ""
  });
  const [orderForm, setOrderForm] = useState({
    customerId: "",
    projectId: "",
    productId: "",
    siteId: "",
    planQuantity: "",
    unitPrice: "",
    planTime: today,
    contact: "",
    phone: ""
  });
  const [contractForm, setContractForm] = useState({
    contractId: "",
    customerId: "",
    projectId: "",
    productId: "",
    name: "",
    validFrom: today,
    validTo: "",
    quantity: "",
    unitPrice: "",
    reason: "",
    attachmentName: "",
    attachmentFileType: "contract_pdf",
    attachmentUrl: "",
    attachmentChecksum: ""
  });
  const [contractAttachmentCache, setContractAttachmentCache] = useState<Record<number, ContractAttachment[]>>({});
  const [contractAttachmentLoadingId, setContractAttachmentLoadingId] = useState<number | null>(null);
  const [productionForm, setProductionForm] = useState({
    orderId: "",
    planId: "",
    taskId: "",
    plantId: "",
    planDate: today,
    shift: "早班",
    planQuantity: "",
    adjustPlanQuantity: "",
    taskQty: "",
    batchQty: "",
    batchQuality: "passed"
  });
  const [qualityInspectionForm, setQualityInspectionForm] = useState({
    batchId: "",
    inspector: "",
    slump: "",
    temperature: "",
    remark: ""
  });
  const [qualitySampleForm, setQualitySampleForm] = useState({
    sampleId: "",
    strength: "",
    result: "passed",
    testedAt: "",
    remark: ""
  });
  const [rawInspectionForm, setRawInspectionForm] = useState({
    receiptId: "",
    inspector: "",
    sampleNo: "",
    remark: ""
  });
  const [rawInspectionReviewForm, setRawInspectionReviewForm] = useState({
    inspectionId: "",
    moisture: "",
    mudContent: "",
    fineness: "",
    result: "passed",
    remark: ""
  });
  const [procurementForm, setProcurementForm] = useState({
    purchaseOrderId: "",
    supplierId: "",
    siteId: "",
    materialId: "",
    plateNo: "",
    grossWeight: "",
    tareWeight: ""
  });
  const [stocktakeForm, setStocktakeForm] = useState({
    siteId: "",
    materialId: "",
    actualQty: "",
    unit: "t",
    remark: ""
  });
  const [deliveryForm, setDeliveryForm] = useState({
    dispatchId: "",
    ticketId: "",
    channel: "qr",
    phone: "",
    expiresAt: ""
  });
  const [signAttachmentForm, setSignAttachmentForm] = useState({
    signId: "",
    fileName: "",
    fileType: "image/jpeg",
    url: "",
    checksum: "",
    uploadedBy: ""
  });
  const [ticketForm, setTicketForm] = useState({
    mode: "product_out",
    dispatchId: "",
    transferId: "",
    relatedTicketId: "",
    siteId: "",
    materialId: "",
    plateNo: "",
    grossWeight: "",
    tareWeight: "",
    unit: "t",
    remark: ""
  });
  const [financeForm, setFinanceForm] = useState({
    receivableId: "",
    receiptAmount: "",
    planAmount: "",
    planDueDate: today,
    statementId: "",
    invoiceCategory: "blue_vat_special",
    invoiceId: "",
    redLetterInfoId: "",
    redReason: "",
    collectionTaskId: "",
    collectionTemplateId: "",
    collectionRemark: "",
    supplierId: "",
    supplierStatementId: "",
    payableId: "",
    paymentAmount: "",
    paymentMethod: "bank"
  });
  const [collectionTemplateForm, setCollectionTemplateForm] = useState({
    name: "",
    level: "warning",
    channel: "sms",
    content: "",
    enabled: "true"
  });
  const [licenseImportText, setLicenseImportText] = useState("");
  const [licenseIssueForm, setLicenseIssueForm] = useState({
    licenseId: "",
    customerName: "",
    watermark: "",
    expiresAt: today,
    edition: "Enterprise",
    modules: "",
    maxSites: "",
    maxVehicles: "",
    issuer: "CBMP License Center",
    privateKey: ""
  });
  const [licenseRenewForm, setLicenseRenewForm] = useState({
    packageId: "",
    licenseId: "",
    expiresAt: "",
    edition: "",
    modules: "",
    maxSites: "",
    maxVehicles: "",
    issuer: "",
    privateKey: ""
  });
  const [licenseRevokeForm, setLicenseRevokeForm] = useState({
    licenseId: "",
    reason: ""
  });
  const [updatePackageForm, setUpdatePackageForm] = useState({
    version: "",
    component: "server",
    channel: "stable",
    status: "available",
    packageType: "full",
    baseVersion: "",
    rollbackVersion: "",
    artifactFileName: "",
    artifactContentType: "application/octet-stream",
    artifactContentBase64: "",
    targetArtifactSha256: "",
    remark: ""
  });
  const [gatewayRouteForm, setGatewayRouteForm] = useState({
    id: "",
    name: "",
    pathPrefix: "",
    stableUpstream: "",
    canaryUpstream: "",
    canaryPercent: "0",
    readTimeoutSec: "120",
    status: "active"
  });
  const [gatewayCanaryForm, setGatewayCanaryForm] = useState({
    routeId: "",
    canaryPercent: "0",
    canaryUpstream: ""
  });
  const [ssoProviderForm, setSsoProviderForm] = useState({
    id: "",
    name: "",
    code: "",
    issuer: "",
    clientId: "",
    clientSecret: "",
    authUrl: "",
    tokenUrl: "",
    userInfoUrl: "",
    jwksUrl: "",
    redirectUri: "",
    scopes: "openid\nprofile\nemail",
    usernameClaim: "preferred_username",
    displayNameClaim: "name",
    roleCode: "customer",
    companyId: "",
    siteId: "",
    autoProvision: "true",
    status: "enabled"
  });
  const [scimProviderForm, setScimProviderForm] = useState({
    id: "",
    name: "",
    code: "",
    bearerToken: "",
    companyId: "",
    siteId: "",
    defaultRoleCode: "customer",
    status: "enabled"
  });
  const [securityPolicyForm, setSecurityPolicyForm] = useState({
    id: "",
    name: "",
    type: "",
    value: "",
    enabled: "true",
    remark: ""
  });
  const [deviceCredentialForm, setDeviceCredentialForm] = useState({
    id: "",
    deviceNo: "",
    deviceKey: "",
    scopes: "location:report",
    status: "active"
  });
  const [integrationEndpointForm, setIntegrationEndpointForm] = useState({
    id: "",
    name: "",
    type: "collection_sms",
    protocol: "rest/http",
    url: "",
    status: "disabled"
  });
  const [ruleDefinitionForm, setRuleDefinitionForm] = useState({
    id: "",
    code: "",
    name: "",
    category: "vehicle",
    metric: "speed",
    operator: ">",
    threshold: "80",
    level: "warning",
    enabled: "true",
    notifyRoles: "dispatcher",
    description: ""
  });
  const [pluginInstallForm, setPluginInstallForm] = useState({
    id: "",
    name: "",
    type: "integration",
    status: "installed",
    version: "1.0.0",
    checksum: "",
    signature: "",
    permissions: "",
    runtime: "node",
    entrypoint: "",
    sandboxRuntime: "node",
    sandboxTimeoutMs: "30000",
    sandboxNetwork: "false",
    sandboxFilesystem: "none",
    sandboxMaxMemoryMb: "128"
  });
  const [pluginRunForm, setPluginRunForm] = useState({
    pluginId: "",
    action: "",
    permission: "",
    input: "{}"
  });
  const [editingUserId, setEditingUserId] = useState<number | null>(null);
  const [userForm, setUserForm] = useState({
    username: "",
    displayName: "",
    password: "",
    roleCode: "dispatcher",
    companyId: "",
    siteId: "",
    customerId: "",
    driverId: "",
    status: "active"
  });
  const [mfaEnrollment, setMfaEnrollment] = useState<MFAEnrollment | null>(null);
  const [mfaCodes, setMfaCodes] = useState<Record<number, string>>({});
  const [editingRoleId, setEditingRoleId] = useState<number | null>(null);
  const [roleForm, setRoleForm] = useState({
    code: "",
    name: "",
    dataScope: "site",
    permissions: ""
  });
  const [fieldPolicyForm, setFieldPolicyForm] = useState({
    id: "",
    roleCode: "dispatcher",
    resource: "customers",
    field: "phone",
    mask: "phone",
    remark: ""
  });
const [editingDictionaryId, setEditingDictionaryId] = useState<number | null>(null);
const [dictionaryForm, setDictionaryForm] = useState({
    type: "product_line",
    code: "",
    label: "",
    sort: "10",
    status: "active"
});
const [dictionaryFilters, setDictionaryFilters] = useState({ keyword: "", type: "all", status: "all" });
const [editingWorkflowId, setEditingWorkflowId] = useState<number | null>(null);
const [workflowForm, setWorkflowForm] = useState({
    code: "",
    name: "",
    category: "approval",
    resource: "",
    status: "active",
    version: "1",
    triggerEventType: "",
    triggerResource: "",
    triggerConditions: "",
    steps: ""
});
const [approvalFlowForm, setApprovalFlowForm] = useState({
    id: "",
    code: "",
    name: "",
    resource: "",
    steps: "[{\"seq\":1,\"roleCode\":\"boss\",\"action\":\"approve\"}]",
    status: "active"
});
const [workflowPresetCode, setWorkflowPresetCode] = useState("");
const [workflowEventForm, setWorkflowEventForm] = useState({
    eventType: "",
    source: "manual",
    eventKey: "",
    actor: "",
    resource: "",
    resourceId: "",
    resourceNo: "",
    title: "",
    reason: "",
    variables: ""
});
const [workflowEventPreview, setWorkflowEventPreview] = useState<WorkflowEventPreview | null>(null);
const [workflowEventResolution, setWorkflowEventResolution] = useState("已人工确认，无需再触发流程");
const [workflowCancelReason, setWorkflowCancelReason] = useState("业务已撤销，终止流程");
const [workflowReassignRole, setWorkflowReassignRole] = useState("boss");
const [workflowReassignReason, setWorkflowReassignReason] = useState("运行中改派");
const [workflowRuntimeFilters, setWorkflowRuntimeFilters] = useState({
    kind: "all",
    resource: "all",
    status: "all",
    roleCode: "all",
    issue: "all"
});
const [editingWorkflowSubscriptionId, setEditingWorkflowSubscriptionId] = useState<number | null>(null);
const [workflowSubscriptionForm, setWorkflowSubscriptionForm] = useState({
    code: "",
    name: "",
    eventType: "",
    resource: "",
    definitionCode: "",
    targetType: "webhook",
	    endpoint: "",
	    secret: "",
	    retryLimit: "3",
	    timeoutSeconds: "5",
	    status: "active"
	});
const [menuFilters, setMenuFilters] = useState({ keyword: "", permission: "all", cache: "all" });
const [menuLabelDialog, setMenuLabelDialog] = useState<{ key: string; currentLabel: string; title: string } | null>(null);
const [menuLabelMenu, setMenuLabelMenu] = useState<{ key: string; currentLabel: string; title: string; x: number; y: number } | null>(null);
const [menuLabelForm, setMenuLabelForm] = useState({ label: "" });
const [orgForm, setOrgForm] = useState({
    companyName: "",
    companyCode: "",
    companyParentId: "",
    companyLevel: "",
    companyRegion: "",
    departmentCompanyId: "",
    departmentParentId: "0",
    departmentName: "",
    departmentCode: ""
  });
  const allSiteOptions = useMemo(() => list(bootstrap?.sites), [bootstrap]);
  const scopedSiteOptions = useMemo(() => selectedSiteId ? allSiteOptions.filter((item) => item.id === selectedSiteId) : allSiteOptions, [allSiteOptions, selectedSiteId]);
  const defaultSiteId = scopedSiteOptions[0]?.id || allSiteOptions[0]?.id || 0;
  const scopedOrders = useMemo(() => data.orders.filter((item) => matchesCurrentSite(item.siteId)), [data.orders, selectedSiteId]);
  const scopedOrderIds = useMemo(() => new Set(scopedOrders.map((item) => item.id)), [scopedOrders]);
  const scopedDispatchOrders = useMemo(() => data.dispatchOrders.filter((item) => matchesCurrentSite(item.siteId)), [data.dispatchOrders, selectedSiteId]);
  const scopedTickets = useMemo(() => data.tickets.filter((item) => matchesCurrentSite(item.siteId)), [data.tickets, selectedSiteId]);
  const scopedWeightRecords = useMemo(() => {
    if (!selectedSiteId) return data.weightRecords;
    const ticketIds = new Set(scopedTickets.map((item) => item.id));
    return data.weightRecords.filter((item) => ticketIds.has(item.ticketId));
  }, [data.weightRecords, scopedTickets, selectedSiteId]);
  const scopedSigns = useMemo(() => data.signs.filter((item) => matchesCurrentOrder(item.orderId)), [data.signs, scopedOrderIds, selectedSiteId]);
  const scopedSignLinks = useMemo(() => data.signLinks.filter((item) => matchesCurrentOrder(item.orderId)), [data.signLinks, scopedOrderIds, selectedSiteId]);
  const scopedSignAttachments = useMemo(() => {
    const signIds = new Set(scopedSigns.map((item) => item.id));
    const dispatchIds = new Set(scopedDispatchOrders.map((item) => item.id));
    return data.signAttachments.filter((item) => signIds.has(item.signId) || dispatchIds.has(item.dispatchId));
  }, [data.signAttachments, scopedDispatchOrders, scopedSigns]);
  const scopedDeliveryNotes = useMemo(() => data.deliveryNotes.filter((item) => matchesCurrentOrder(item.orderId)), [data.deliveryNotes, scopedOrderIds, selectedSiteId]);
  const dispatchById = useMemo(() => new Map(data.dispatchOrders.map((item) => [item.id, item])), [data.dispatchOrders]);
  const ticketById = useMemo(() => new Map(data.tickets.map((item) => [item.id, item])), [data.tickets]);
  const orderById = useMemo(() => new Map(data.orders.map((item) => [item.id, item])), [data.orders]);
  const signByDispatchId = useMemo(() => new Map(data.signs.map((item) => [item.dispatchId, item])), [data.signs]);
  const linkByDispatchId = useMemo(() => {
    const map = new Map<number, DeliverySignLink>();
    data.signLinks.forEach((item) => {
      const previous = map.get(item.dispatchId);
      if (!previous || item.id > previous.id) {
        map.set(item.dispatchId, item);
      }
    });
    return map;
  }, [data.signLinks]);
  const currentPermissions = useMemo(() => permissionsForRole(bootstrap?.roles, bootstrap?.user.roleCode), [bootstrap]);

  async function load() {
    setLoading(true);
    setError("");
    try {
      const guarded = <T,>(permission: string, request: () => Promise<T>, fallbackValue: T) => (
        hasPermission(currentPermissions, permission) ? request() : Promise.resolve(fallbackValue)
      );
      const shouldLoadSystem = section === "system-menu" || section === "system-license" || section === "system-maintenance" || section === "system-gateway" || section === "system-security" || section === "system-identity" || section === "system-plugins" || section === "system-dictionaries" || section === "system-users" || section === "system-roles" || section === "system-workflows" || section === "system-audit" || section === "approval-center";
      const shouldLoadWorkflows = section === "overview" || shouldLoadSystem || section === "orders" || section === "production" || section === "production-plans" || section === "production-tasks" || section === "production-batches" || section === "production-reports" || section === "master-plants" || section === "stock-yards" || section === "customer-risk" || section === "settlement" || section === "contracts" || section === "inventory-transfers" || section === "inventory-stocktakes" || section === "raw-material-inspections" || section === "weighbridge" || section === "delivery" || section === "delivery-signs" || section === "finance" || section === "finance-receivables" || section === "finance-invoices" || section === "finance-collections" || section === "finance-suppliers" || section === "finance-carriers";
      const shouldLoadOrg = section === "system-org";
      const shouldLoadQuality = section === "production" || section === "production-batches" || section === "stock-yards" || section === "raw-material-receipts" || section === "raw-material-inspections" || section === "reports";
      const shouldLoadDispatchDetails = section === "dispatch" || section === "dispatch-schedules" || section === "dispatch-queue" || section === "settlement" || section === "finance-carriers" || section === "reports";
      const shouldLoadPortal = section === "overview" || section === "dispatch" || section === "delivery" || section === "settlement" || section === "portal-customer" || section === "portal-driver";
      const shouldLoadDeliveryDetails = section === "delivery" || section === "delivery-signs" || section === "settlement";
      const shouldLoadSystemRuntime = shouldLoadSystem || section === "map-center";
      const shouldLoadRules = shouldLoadSystemRuntime || section === "system-rules";
      const shouldLoadIntegrations = shouldLoadSystemRuntime || section === "system-integrations";
      const [dashboard, reports, dispatch, portal, portalComplaints, latestLocations, alarms, geoFences, production, procurement, quality, finance, rules, integrations, contracts, orders, dispatchOrders, dispatchSchedules, carrierSettlementBundle, tickets, ticketPrintLogs, ticketVoidLogs, weightRecords, scaleDeviceEvents, deliveryNotes, signs, signLinks, signAttachments, statements, approvals, nextMapConfig, org, system, workflowOverviewData, workflowCatalogData, directWorkflowInbox, directWorkflowInstances, directWorkflowTasks, directWorkflowLogs, directWorkflowEvents, directWorkflowOutbox, directWorkflowDeliveries, directBackups, directBackupDrills, directGateway, systemUsers, systemRoles, directDictionaries, auditLogs, modules] = await Promise.all([
        guarded("dashboard:read", () => api.dashboard(), null as DashboardData | null),
        guarded("report:read", () => api.reports(), null as ManagementReports | null),
        guarded("dispatch:read", () => api.dispatchCenterOverview(), null as DispatchCenterOverview | null),
        shouldLoadPortal ? guarded("bootstrap:read", () => api.portalOverview(), null as PortalOverview | null) : Promise.resolve(null),
        shouldLoadPortal ? guarded("customer:read", () => api.portalComplaints(), [] as CustomerComplaint[]) : Promise.resolve([] as CustomerComplaint[]),
        guarded("vehicle:read", () => api.latestLocations(), [] as LatestLocation[]),
        guarded("vehicle:read", () => api.alarms(), [] as VehicleAlarm[]),
        guarded("vehicle:read", () => api.geoFences(), [] as GeoFence[]),
        guarded("production:read", () => api.productionOverview(), null as ProductionOverview | null),
        guarded("procurement:read", () => api.procurementOverview(), null as ProcurementOverview | null),
        shouldLoadQuality ? guarded("quality:read", () => api.qualityOverview(), null as QualityOverview | null) : Promise.resolve(null),
        guarded("finance:read", () => api.financeOverview(), null as FinanceOverview | null),
        shouldLoadRules ? guarded("rule:read", () => api.rulesOverview(), null as RuleOverview | null) : Promise.resolve(null),
        shouldLoadIntegrations ? guarded("integration:read", () => api.integrationsOverview(), null as IntegrationOverview | null) : Promise.resolve(null),
        guarded("contract:read", () => api.contracts(), [] as Contract[]),
        guarded("order:read", () => api.orders(), [] as SalesOrder[]),
        guarded("dispatch:read", () => api.dispatchOrders(), [] as DispatchOrder[]),
        shouldLoadDispatchDetails ? guarded("dispatch:read", () => api.dispatchSchedules(), [] as DispatchSchedule[]) : Promise.resolve([] as DispatchSchedule[]),
        shouldLoadDispatchDetails ? guarded("dispatch:read", () => api.carrierSettlements(), { settlements: [] as TransportSettlement[], items: [] as TransportSettlementItem[] }) : Promise.resolve({ settlements: [] as TransportSettlement[], items: [] as TransportSettlementItem[] }),
        guarded("ticket:read", () => api.tickets(), [] as ScaleTicket[]),
        guarded("ticket:read", () => api.ticketPrintLogs(), [] as TicketPrintLog[]),
        guarded("ticket:read", () => api.ticketVoidLogs(), [] as TicketVoidLog[]),
        guarded("ticket:read", () => api.weightRecords(), [] as ScaleWeightRecord[]),
        guarded("ticket:read", () => api.scaleDeviceEvents(), [] as ScaleDeviceEvent[]),
        guarded("delivery:read", () => api.deliveryNotes(), [] as DeliveryNote[]),
        guarded("delivery:read", () => api.signs(), [] as DeliverySign[]),
        guarded("delivery:read", () => api.signLinks(), [] as DeliverySignLink[]),
        shouldLoadDeliveryDetails ? guarded("delivery:read", () => api.signAttachments(), [] as DeliverySignAttachment[]) : Promise.resolve([] as DeliverySignAttachment[]),
        guarded("statement:read", () => api.statements(), [] as Statement[]),
        guarded("approval:read", () => api.approvals(), [] as ApprovalTask[]),
        guarded("system:read", () => api.mapConfig(), null as MapProviderConfig | null),
        shouldLoadOrg ? guarded("org:read", () => api.orgOverview(), null as OrganizationOverview | null) : Promise.resolve(null),
        shouldLoadSystem ? guarded("system:read", () => api.systemBundle(), null as SystemBundle | null) : Promise.resolve(null),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowOverview(), null as WorkflowOverview | null) : Promise.resolve(null),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowCatalog(), null as WorkflowCatalog | null) : Promise.resolve(null),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowInbox(), [] as WorkflowInboxItem[]) : Promise.resolve([] as WorkflowInboxItem[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowInstances(), [] as WorkflowInstance[]) : Promise.resolve([] as WorkflowInstance[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowTasks(), [] as WorkflowTask[]) : Promise.resolve([] as WorkflowTask[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowLogs(), [] as WorkflowLog[]) : Promise.resolve([] as WorkflowLog[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowEvents(), [] as WorkflowEvent[]) : Promise.resolve([] as WorkflowEvent[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowOutbox(), [] as WorkflowOutbox[]) : Promise.resolve([] as WorkflowOutbox[]),
        shouldLoadWorkflows ? guarded("approval:read", () => api.workflowDeliveries(), [] as WorkflowDelivery[]) : Promise.resolve([] as WorkflowDelivery[]),
        shouldLoadSystem ? guarded("system:read", () => api.listBackups(), [] as BackupInfo[]) : Promise.resolve([] as BackupInfo[]),
        shouldLoadSystem ? guarded("system:read", () => api.listBackupDrills(), [] as BackupDrill[]) : Promise.resolve([] as BackupDrill[]),
        shouldLoadSystem ? guarded("system:read", () => api.gatewayOverview(), null as GatewayOverview | null) : Promise.resolve(null),
        shouldLoadSystem ? guarded("system:read", () => api.systemUsers(), [] as User[]) : Promise.resolve([] as User[]),
        shouldLoadSystem ? guarded("system:read", () => api.systemRoles(), [] as Role[]) : Promise.resolve([] as Role[]),
        section === "system-dictionaries" ? guarded("system:read", () => api.systemDictionaries(), [] as DataDictionary[]) : Promise.resolve([] as DataDictionary[]),
        section === "system-audit" ? guarded("system:read", () => api.auditLogs(), [] as AuditLog[]) : Promise.resolve([] as AuditLog[]),
        section === "system-maintenance" ? guarded("system:read", () => api.systemModules(), [] as ModuleInfo[]) : Promise.resolve([] as ModuleInfo[])
      ]);
      const resolvedSystem = ({ ...(system || {}) } as SystemBundle);
      if (workflowOverviewData || shouldLoadWorkflows) {
        const baseWorkflow = workflowOverviewData || resolvedSystem.workflows || { definitions: [], instances: [], tasks: [], inbox: [], events: [], logs: [], outbox: [], subscriptions: [], deliveries: [] };
        resolvedSystem.workflows = {
          ...baseWorkflow,
          catalog: workflowCatalogData || baseWorkflow.catalog,
          inbox: shouldLoadWorkflows ? list(directWorkflowInbox) : list(baseWorkflow.inbox),
          instances: shouldLoadWorkflows ? list(directWorkflowInstances) : list(baseWorkflow.instances),
          tasks: shouldLoadWorkflows ? list(directWorkflowTasks) : list(baseWorkflow.tasks),
          logs: shouldLoadWorkflows ? list(directWorkflowLogs) : list(baseWorkflow.logs),
          events: shouldLoadWorkflows ? list(directWorkflowEvents) : list(baseWorkflow.events),
          outbox: shouldLoadWorkflows ? list(directWorkflowOutbox) : list(baseWorkflow.outbox),
          deliveries: shouldLoadWorkflows ? list(directWorkflowDeliveries) : list(baseWorkflow.deliveries)
        };
      }
      if (directDictionaries.length) resolvedSystem.dictionaries = directDictionaries;
      if (directBackups.length) resolvedSystem.backups = directBackups;
      if (directBackupDrills.length) resolvedSystem.backupDrills = directBackupDrills;
      if (directGateway) resolvedSystem.gateway = directGateway;
      setMapConfig(nextMapConfig);
      setData({
        dashboard,
        reports,
        dispatch,
        portal,
        production,
        procurement,
        quality,
        finance,
        rules,
        integrations,
        org,
        system: resolvedSystem,
        modules: list(modules),
        systemUsers,
        systemRoles,
        auditLogs: list(auditLogs),
        contracts: list(contracts),
        orders: list(orders),
        dispatchOrders: list(dispatchOrders),
        dispatchSchedules: list(dispatchSchedules),
        carrierSettlements: list(carrierSettlementBundle.settlements),
        carrierSettlementItems: list(carrierSettlementBundle.items),
        latestLocations: list(latestLocations),
        alarms: list(alarms),
        geoFences: list(geoFences),
        tickets: list(tickets),
        ticketPrintLogs: list(ticketPrintLogs),
        ticketVoidLogs: list(ticketVoidLogs),
        weightRecords: list(weightRecords),
        scaleDeviceEvents: list(scaleDeviceEvents),
        deliveryNotes: list(deliveryNotes),
        signs: list(signs),
        signLinks: list(signLinks),
        signAttachments: list(signAttachments).length ? list(signAttachments) : list(portal?.signAttachments),
        portalComplaints: list(portalComplaints).length ? list(portalComplaints) : list(portal?.complaints),
        statements: list(statements),
        approvals: list(approvals)
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "ERP 数据加载失败");
      setData(emptyData);
      setMapConfig(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : "ERP 数据加载失败"));
  }, [section, currentPermissions]);

  useEffect(() => {
    if (error) {
      showError(error, "ERP 数据加载失败");
    }
  }, [error, showError]);

  useEffect(() => {
    if (actionError) {
      showError(actionError, "操作失败");
    }
  }, [actionError, showError]);

  useEffect(() => {
    if (dispatchActionError) {
      showError(dispatchActionError, "派车失败");
    }
  }, [dispatchActionError, showError]);

  useEffect(() => {
    const contractId = fieldNumber(contractForm.contractId);
    if (!contractId) {
      setContractAttachmentLoadingId(null);
      return;
    }
    if (Object.prototype.hasOwnProperty.call(contractAttachmentCache, contractId)) {
      return;
    }
    let cancelled = false;
    setContractAttachmentLoadingId(contractId);
    api.contractAttachments(contractId)
      .then((items) => {
        if (cancelled) return;
        setContractAttachmentCache((value) => ({ ...value, [contractId]: list(items) }));
      })
      .catch((err) => {
        if (cancelled) return;
        showError(err, "合同附件加载失败");
        setContractAttachmentCache((value) => ({ ...value, [contractId]: [] }));
      })
      .finally(() => {
        if (cancelled) return;
        setContractAttachmentLoadingId((value) => (value === contractId ? null : value));
      });
    return () => {
      cancelled = true;
    };
  }, [contractAttachmentCache, contractForm.contractId, showError]);

  useEffect(() => () => {
    reverseGeocodeAbortRef.current?.abort();
  }, []);

  useEffect(() => {
    setSiteFilter(selectedSiteId ? String(selectedSiteId) : "all");
  }, [selectedSiteId]);

  useEffect(() => {
    const allProgress = list(data.dispatch?.siteProgress).filter((item) => matchesCurrentSite(item.siteId));
    const allVehicles = list(data.dispatch?.availableVehicles).filter((item) => matchesCurrentSite(item.siteId));
    const progress = allProgress.filter((item) => siteFilter === "all" || String(item.siteId) === siteFilter);
    const vehicles = allVehicles.filter((item) => siteFilter === "all" || String(item.siteId) === siteFilter);
    const siteIds = new Set(allProgress.map((item) => String(item.siteId)));
    if (!selectedSiteId && siteFilter !== "all" && !siteIds.has(siteFilter)) {
      setSiteFilter("all");
    }
    if (progress.length && !progress.some((item) => item.orderId === selectedOrderId)) {
      setSelectedOrderId(progress[0].orderId);
    }
    if (!progress.length && selectedOrderId !== null) {
      setSelectedOrderId(null);
    }
    if (vehicles.length && !vehicles.some((item) => item.vehicleId === selectedVehicleId)) {
      setSelectedVehicleId(vehicles[0].vehicleId);
    }
    if (!vehicles.length && selectedVehicleId !== null) {
      setSelectedVehicleId(null);
    }
  }, [data.dispatch, selectedOrderId, selectedVehicleId, siteFilter, selectedSiteId]);

	  useEffect(() => {
	    const siteId = String(defaultSiteId);
	    const defaultResourceStatus = firstDictionaryCode("resource_status", "active");
	    setMasterForm((value) => ({
      ...value,
      projectCustomerId: value.projectCustomerId || String(firstId(bootstrap?.customers)),
      siteCompanyId: value.siteCompanyId || String(firstId(bootstrap?.companies)),
      vehicleDriverId: value.vehicleDriverId || String(firstId(bootstrap?.drivers)),
      vehicleSiteId: selectedSiteId ? siteId : value.vehicleSiteId || siteId,
      carrierSettleMode: value.carrierSettleMode || "monthly",
      carrierStatus: value.carrierStatus || defaultResourceStatus
    }));
    setBufferForm((value) => ({
      ...value,
      plantId: value.plantId || String(firstId(list(bootstrap?.plants).filter((item) => !selectedSiteId || item.siteId === selectedSiteId))),
      materialId: value.materialId || String(firstId(bootstrap?.materials))
    }));
    setOrderForm((value) => ({
      ...value,
      customerId: value.customerId || String(firstId(bootstrap?.customers)),
      projectId: value.projectId || String(firstId(bootstrap?.projects)),
      productId: value.productId || String(firstId(bootstrap?.products)),
      siteId: selectedSiteId ? siteId : value.siteId || siteId
    }));
    setContractForm((value) => ({
      ...value,
      customerId: value.customerId || String(firstId(bootstrap?.customers)),
      projectId: value.projectId || String(firstId(bootstrap?.projects)),
      productId: value.productId || String(firstId(bootstrap?.products))
    }));
    setUserForm((value) => ({
      ...value,
      companyId: value.companyId || String(firstId(bootstrap?.companies)),
      siteId: selectedSiteId ? siteId : value.siteId || siteId,
      customerId: value.customerId || "0",
      driverId: value.driverId || "0"
    }));
    setOrgForm((value) => ({
      ...value,
      companyParentId: value.companyParentId || String(firstId(bootstrap?.companies)),
      departmentCompanyId: value.departmentCompanyId || String(firstId(bootstrap?.companies))
    }));
  }, [bootstrap, defaultSiteId, selectedSiteId]);

  useEffect(() => {
    const supplier = supplierOptions()[0];
    const purchaseRequestIds = new Set(list(data.procurement?.requests).filter((item) => matchesCurrentSite(item.siteId)).map((item) => item.id));
    const purchaseOrder = list(data.procurement?.orders).filter((item) => !selectedSiteId || !item.requestId || purchaseRequestIds.has(item.requestId))[0];
    const receivable = openReceivables()[0] || list(data.finance?.receivables)[0];
    const contract = activeContracts()[0];
    const statement = invoiceableStatements()[0] || list(data.finance?.statements)[0] || data.statements[0];
    const invoice = redOffsetCandidateInvoices()[0] || list(data.finance?.invoices)[0];
    const redLetter = list(data.finance?.redLetterInfos).find((item) => item.status === "approved") || list(data.finance?.redLetterInfos)[0];
    const collectionTask = openCollectionTasks()[0];
    const collectionTemplate = list(data.finance?.collectionTemplates).find((item) => item.enabled) || list(data.finance?.collectionTemplates)[0];
    const supplierStatement = list(data.finance?.supplierStatements).find((item) => item.status !== "approved") || list(data.finance?.supplierStatements)[0];
    const payable = openPayables()[0];
    setProcurementForm((value) => ({
      ...value,
      purchaseOrderId: value.purchaseOrderId || String(purchaseOrder?.id || ""),
      supplierId: value.supplierId || String(recordId(supplier)),
      siteId: selectedSiteId ? String(defaultSiteId) : value.siteId || String(defaultSiteId),
      materialId: value.materialId || String(firstId(bootstrap?.materials))
    }));
    setContractForm((value) => ({
      ...value,
      contractId: value.contractId || String(contract?.id || ""),
      customerId: value.customerId || String(firstId(bootstrap?.customers)),
      projectId: value.projectId || String(firstId(bootstrap?.projects)),
      productId: value.productId || String(firstId(bootstrap?.products)),
      unitPrice: value.unitPrice || String(bootstrap?.products?.find((item) => item.id === firstId(bootstrap?.products))?.basePrice || "")
    }));
    setFinanceForm((value) => ({
      ...value,
      receivableId: value.receivableId || String(receivable?.id || ""),
      receiptAmount: value.receiptAmount || String(Math.max(0, (receivable?.amount || 0) - (receivable?.receivedAmount || 0)) || ""),
      planAmount: value.planAmount || String(Math.max(0, (receivable?.amount || 0) - (receivable?.receivedAmount || 0)) || ""),
      statementId: value.statementId || String(statement?.id || ""),
      invoiceId: value.invoiceId || String(invoice?.id || ""),
      redLetterInfoId: value.redLetterInfoId || String(redLetter?.id || ""),
      collectionTaskId: value.collectionTaskId || String(collectionTask?.id || ""),
      collectionTemplateId: value.collectionTemplateId || String(collectionTemplate?.id || ""),
      supplierId: value.supplierId || String(recordId(supplier)),
      supplierStatementId: value.supplierStatementId || String(supplierStatement?.id || ""),
      payableId: value.payableId || String(payable?.id || ""),
      paymentAmount: value.paymentAmount || String(Math.max(0, (payable?.amount || 0) - (payable?.paidAmount || 0)) || "")
    }));
  }, [bootstrap, data.contracts, data.procurement, data.finance, data.statements, defaultSiteId, scopedTickets, selectedSiteId]);

  useEffect(() => {
    const plans = list(data.production?.plans).filter((item) => matchesCurrentSite(item.siteId));
    const tasks = list(data.production?.tasks).filter((item) => matchesCurrentSite(item.siteId) && item.status !== "cancelled" && item.status !== "completed");
    const orders = scopedOrders.filter((item) => productionOrderRemaining(item, plans) > 0 && ["approved", "scheduled", "dispatching"].includes(item.status));
    setProductionForm((value) => {
      const selectedOrder = orders.find((item) => String(item.id) === value.orderId) || orders[0];
      const selectedPlan = plans.find((item) => String(item.id) === value.planId) || plans[0];
      const selectedTask = tasks.find((item) => String(item.id) === value.taskId) || tasks[0];
      const selectedPlant = selectedPlan ? productionPlanPlant(selectedPlan) : firstProductionPlant(selectedOrder?.siteId || selectedSiteId || undefined);
      return {
        ...value,
        orderId: selectedOrder ? String(selectedOrder.id) : "",
        planId: selectedPlan ? String(selectedPlan.id) : "",
        taskId: selectedTask ? String(selectedTask.id) : "",
        plantId: selectedPlant ? String(selectedPlant.id) : "",
        planQuantity: value.planQuantity || (selectedOrder ? String(productionOrderRemaining(selectedOrder, plans)) : ""),
        adjustPlanQuantity: value.adjustPlanQuantity || (selectedPlan ? String(selectedPlan.planQuantity) : ""),
        taskQty: value.taskQty || (selectedPlan ? String(Math.max(0, selectedPlan.planQuantity - selectedPlan.plannedTaskQty)) : ""),
        batchQty: value.batchQty || (selectedTask ? String(Math.max(0, selectedTask.planQty - selectedTask.producedQty)) : "")
      };
    });
  }, [data.production, scopedOrders, selectedSiteId]);

  useEffect(() => {
    const dispatch = scopedDispatchOrders.find((item) => String(item.id) === deliveryForm.dispatchId) || scopedDispatchOrders[0];
    const dispatchTickets = dispatch ? scopedTickets.filter((item) => item.dispatchId === dispatch.id) : [];
    const ticket = dispatchTickets.find((item) => String(item.id) === deliveryForm.ticketId) || dispatchTickets[0];
    const order = dispatch ? orderById.get(dispatch.orderId) : undefined;
    const nextDispatchId = dispatch ? String(dispatch.id) : "";
    const nextTicketId = ticket ? String(ticket.id) : "";
    const nextPhone = deliveryForm.phone || order?.phone || "";
    if (deliveryForm.dispatchId !== nextDispatchId || deliveryForm.ticketId !== nextTicketId || deliveryForm.phone !== nextPhone) {
      setDeliveryForm((value) => ({
        ...value,
        dispatchId: nextDispatchId,
        ticketId: nextTicketId,
        phone: nextPhone
      }));
    }
  }, [deliveryForm.dispatchId, deliveryForm.ticketId, deliveryForm.phone, orderById, scopedDispatchOrders, scopedTickets]);

  const activeOrders = useMemo(() => scopedOrders.filter((item) => !["completed", "cancelled"].includes(item.status)), [scopedOrders]);
  const openApprovals = useMemo(() => data.approvals.filter((item) => item.status !== "approved" && item.status !== "rejected"), [data.approvals]);

  function matchesCurrentSite(siteId: number | string | null | undefined) {
    return !selectedSiteId || Number(siteId || 0) === selectedSiteId;
  }

  function matchesCurrentOrder(orderId: number | string | null | undefined) {
    return !selectedSiteId || scopedOrderIds.has(Number(orderId || 0));
  }

  function siteOptions() {
    return scopedSiteOptions;
  }

  function activeSiteFence(siteId: number) {
    return data.geoFences.find((item) => item.type === "site" && item.siteId === siteId && item.status !== "inactive");
  }

  function siteEnableStatus(status?: string) {
    return status === "disabled" || status === "inactive" || status === "retired" ? "disabled" : "active";
  }

  function renderSiteField(label: string, value: string | number | undefined, onChange?: (value: string) => void, inputName?: string) {
    const sites = siteOptions();
    const siteValue = String(fieldNumber(String(value || ""), defaultSiteId));
    if (sites.length <= 1) {
      return (
        <Field label={label} className="site-scope-lock">
          {inputName ? <TextInput type="hidden" name={inputName} value={siteValue} /> : null}
          <TextInput value={nameOf(bootstrap?.sites, fieldNumber(siteValue)) || "操作T,"} readOnly />
        </Field>
      );
    }
    return (
      <Field label={label}>
        {inputName && !onChange ? (
          <SelectInput name={inputName} defaultValue={siteValue}>
            {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        ) : (
          <SelectInput value={siteValue} onChange={(event) => onChange?.(event.target.value)}>
            {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        )}
      </Field>
    );
  }

  function supplierOptions() {
    return list(data.procurement?.suppliers as Array<{ id: number; name?: string; supplierName?: string; orderNo?: string }> | undefined);
  }

  function openReceivables() {
    return list(data.finance?.receivables).filter((item) => item.status !== "paid" && item.amount > item.receivedAmount);
  }

  function activeContracts() {
    return data.contracts.filter((item) => item.status !== "archived");
  }

  function contractAttachments(contractId: number) {
    return contractAttachmentCache[contractId] || [];
  }

  function renderContractAttachmentList(contractId: number, attachments: ContractAttachment[]) {
    if (contractAttachmentLoadingId === contractId) {
      return <span>附件加载中...</span>;
    }
    if (!attachments.length) {
      return <span>暂无附件</span>;
    }
    return attachments.map((item) => <span key={item.id}>{item.fileName} / {item.status}</span>);
  }

  function invoiceableStatements() {
    const invoicedStatementIds = new Set(list(data.finance?.invoices).filter((item) => item.invoiceType !== "red").map((item) => item.statementId));
    return list(data.finance?.statements).filter((item) => item.status === "confirmed" && !invoicedStatementIds.has(item.id));
  }

  function redOffsetCandidateInvoices() {
    return list(data.finance?.invoices).filter((item) => (
      item.invoiceType !== "red" &&
      item.status !== "red_offset" &&
      item.taxStatus === "submitted" &&
      !!item.taxControlNo
    ));
  }

  function openCollectionTasks() {
    return list(data.finance?.collectionTasks).filter((item) => item.status !== "handled" && item.status !== "closed");
  }

  function openPayables() {
    return list(data.finance?.payables).filter((item) => item.status !== "paid" && item.amount > item.paidAmount);
  }

  function payableBalance(item: Payable | undefined) {
    return Math.max(0, (item?.amount || 0) - (item?.paidAmount || 0));
  }

  function productionOrderRemaining(order: SalesOrder, plans: ProductionPlan[]) {
    const planned = plans
      .filter((item) => item.orderId === order.id && item.status !== "cancelled")
      .reduce((sum, item) => sum + item.planQuantity, 0);
    return Math.max(0, Math.round((order.planQuantity - planned) * 100) / 100);
  }

  function productionTaskRemaining(task: ProductionTask | undefined) {
    if (!task) return 0;
    return Math.max(0, Math.round((task.planQty - task.producedQty) * 100) / 100);
  }

  function productionTaskPriority(task: ProductionTask) {
    if (task.status === "running" || task.status === "producing") return 0;
    if (task.status === "pending") return 1;
    return 2;
  }

  function sortProductionTasksForAction(items: ProductionTask[]) {
    return [...items].sort((left, right) => (
      productionTaskPriority(left) - productionTaskPriority(right)
      || productionTaskRemaining(right) - productionTaskRemaining(left)
      || left.id - right.id
    ));
  }

  function productionBatchReportDate(batch: ProductionBatch) {
    return (batch.completedAt || batch.startedAt || "").slice(0, 10);
  }

  function productionPlanTaskRemaining(plan: ProductionPlan | undefined) {
    if (!plan) return 0;
    return Math.max(0, Math.round((plan.planQuantity - plan.plannedTaskQty) * 100) / 100);
  }

  function productionPlanCanIssueTask(plan: ProductionPlan | undefined) {
    if (!plan || plan.status === "cancelled" || plan.status === "completed") return false;
    return productionPlanTaskRemaining(plan) > 0 && plan.recipeStatus !== "missing";
  }

  function activeProductionTasksForPlan(plan: ProductionPlan | undefined) {
    if (!plan) return [];
    return sortProductionTasksForAction(list(data.production?.tasks)
      .filter((item) => matchesCurrentSite(item.siteId))
      .filter((item) => item.planId === plan.id && item.status !== "cancelled" && item.status !== "completed")
      .filter((item) => productionTaskRemaining(item) > 0));
  }

  function productionPlanReports(plan: ProductionPlan | undefined) {
    if (!plan) return [];
    return list(data.production?.reports).filter((item) => item.siteId === plan.siteId && item.reportDate === plan.planDate);
  }

  function productionReportKey(plan: ProductionPlan | undefined) {
    return plan ? `${plan.siteId}:${plan.planDate}` : "";
  }

  function productionReportDayPlans(plan: ProductionPlan | undefined, overridePlan = plan) {
    if (!plan) return [];
    const key = productionReportKey(plan);
    const reportPlans = new Map<number, ProductionPlan>();
    list(data.production?.plans)
      .filter((item) => matchesCurrentSite(item.siteId))
      .filter((item) => item.status !== "cancelled" && productionReportKey(item) === key)
      .forEach((item) => reportPlans.set(item.id, item));
    if (overridePlan && overridePlan.status !== "cancelled" && productionReportKey(overridePlan) === key) {
      reportPlans.set(overridePlan.id, overridePlan);
    }
    return Array.from(reportPlans.values()).sort((left, right) => left.id - right.id);
  }

  function productionReportDayReady(plan: ProductionPlan | undefined, overridePlan = plan) {
    if (!plan || plan.status === "cancelled" || productionPlanReports(plan).length) return false;
    const reportPlans = productionReportDayPlans(plan, overridePlan);
    return reportPlans.length > 0
      && reportPlans.every((item) => item.remainingQty <= 0 || item.status === "completed")
      && reportPlans.some((item) => item.producedQty > 0);
  }

  function productionPlanNeedsReport(plan: ProductionPlan | undefined) {
    return productionReportDayReady(plan);
  }

  function productionReportAnchorPlan(plan: ProductionPlan | undefined) {
    if (!productionReportDayReady(plan, plan)) return undefined;
    const key = productionReportKey(plan);
    const reportPlans = new Map<number, ProductionPlan>();
    list(data.production?.plans)
      .filter((item) => matchesCurrentSite(item.siteId))
      .forEach((item) => reportPlans.set(item.id, item));
    if (plan) {
      reportPlans.set(plan.id, plan);
    }
    return Array.from(reportPlans.values())
      .filter((item) => productionReportDayReady(item, plan) && productionReportKey(item) === key)
      .sort((left, right) => left.id - right.id)[0];
  }

  function productionPlanCanOpenReport(plan: ProductionPlan | undefined) {
    const reportPlan = productionReportAnchorPlan(plan);
    return reportPlan?.id === plan?.id;
  }

  function productionPlanCanCancel(plan: ProductionPlan | undefined) {
    return !!plan && plan.status !== "cancelled" && plan.status !== "completed" && plan.producedQty <= 0;
  }

  function productionPlanCancelReason(plan: ProductionPlan | undefined) {
    if (!plan) return "";
    if (plan.status === "cancelled") return "该计划已取消，无需重复取消。";
    if (plan.status === "completed") return "该计划已完成，不能取消。";
    if (plan.producedQty > 0) return "该计划已有生产批次，不能直接取消。";
    return "";
  }

  function productionPlanNextAction(plan: ProductionPlan): ProductionPlanAction {
    if (plan.status === "cancelled") {
      return { label: "查看详情", mode: "detail", icon: <Search size={14} />, disabled: false };
    }
    if (productionPlanCanOpenReport(plan)) {
      return { label: "生成日报", mode: "report", icon: <ClipboardCheck size={14} />, variant: "primary" };
    }
    if (plan.status === "completed") {
      return { label: "查看详情", mode: "detail", icon: <Search size={14} />, disabled: false };
    }
    if (activeProductionTasksForPlan(plan).length) {
      return { label: "登记批次", mode: "batch", icon: <Plus size={14} />, variant: "primary" };
    }
    if (productionPlanCanIssueTask(plan)) {
      return {
        label: "下达任务",
        mode: "tasks",
        icon: <PlayCircle size={14} />,
        variant: "primary"
      };
    }
    return { label: "查看详情", mode: "detail", icon: <Search size={14} /> };
  }

  function productionPlants() {
    const overviewPlants = list(data.production?.plants);
    return overviewPlants.length ? overviewPlants : list(bootstrap?.plants);
  }

  function productionPlantOptions(siteId: number | undefined) {
    return productionPlants().filter((item) => (!siteId || item.siteId === siteId) && ["running", "active"].includes(item.status));
  }

  function firstProductionPlant(siteId: number | undefined) {
    return productionPlantOptions(siteId)[0];
  }

  function productionOrderPlanDate(order: SalesOrder | undefined) {
    const planTime = order?.planTime?.trim();
    return planTime ? planTime.slice(0, 10) : today;
  }

  function preferredProductionPlantForOrder(order: SalesOrder | undefined, planDate = today) {
    const siteId = order?.siteId || selectedSiteId || undefined;
    const options = productionPlantOptions(siteId);
    if (!order) return options[0];
    const mix = currentApprovedMix(order.productId, order.siteId);
    return options.find((item) => currentProductionProfile(mix?.id, item.id, planDate)) || options[0];
  }

  function productionPlantLabel(plant: Plant | undefined) {
    return plant ? `${plant.name} · ${plant.code}` : "-";
  }

  function vehicleInternalNo(vehicleId: number | undefined, fallbackValue: string) {
    const vehicle = list(bootstrap?.vehicles).find((item) => item.id === vehicleId);
    return vehicle?.internalNo || fallbackValue;
  }

  function dispatchVehicleTitle(item: DispatchCenterQueueItem | DispatchCenterVehicle | LatestLocation) {
    return vehicleInternalNo(item.vehicleId, "queueNo" in item && item.queueNo ? item.queueNo : item.plateNo);
  }

  function dispatchVehicleMeta(item: DispatchCenterQueueItem | DispatchCenterVehicle | LatestLocation) {
    const parts = [item.plateNo];
    if ("driverName" in item && item.driverName) parts.push(item.driverName);
    if ("queueNo" in item && item.queueNo) parts.push(item.queueNo);
    return parts.filter(Boolean).join(" / ");
  }

  function productionQueueStatusRank(status: string | undefined) {
    switch (status) {
      case "loading":
        return 1;
      case "arrived_site":
      case "waiting_load":
        return 2;
      case "accepted":
        return 3;
      case "assigned":
        return 4;
      case "loaded":
      case "departed":
      case "in_transit":
        return 5;
      default:
        return 9;
    }
  }

  function sortedProductionQueue(items: DispatchCenterQueueItem[]) {
    return [...items].sort((left, right) => {
      const leftOrder = productionQueueOrder[left.dispatchId];
      const rightOrder = productionQueueOrder[right.dispatchId];
      if (leftOrder !== undefined || rightOrder !== undefined) {
        return (leftOrder ?? Number.MAX_SAFE_INTEGER) - (rightOrder ?? Number.MAX_SAFE_INTEGER);
      }
      const statusDiff = productionQueueStatusRank(left.status) - productionQueueStatusRank(right.status);
      if (statusDiff) return statusDiff;
      if (left.eta !== right.eta) return left.eta.localeCompare(right.eta);
      return left.dispatchId - right.dispatchId;
    });
  }

  function setProductionQueueSequence(items: DispatchCenterQueueItem[]) {
    setProductionQueueOrder((current) => {
      const next = { ...current };
      items.forEach((item, index) => {
        next[item.dispatchId] = index;
      });
      return next;
    });
  }

  function moveProductionQueueItem(items: DispatchCenterQueueItem[], dispatchId: number, offset: number) {
    const ordered = sortedProductionQueue(items);
    const index = ordered.findIndex((item) => item.dispatchId === dispatchId);
    const target = index + offset;
    if (index < 0 || target < 0 || target >= ordered.length) return;
    const next = [...ordered];
    [next[index], next[target]] = [next[target], next[index]];
    setProductionQueueSequence(next);
  }

  function prioritizeProductionQueueItem(items: DispatchCenterQueueItem[], dispatchId: number) {
    const ordered = sortedProductionQueue(items);
    const index = ordered.findIndex((item) => item.dispatchId === dispatchId);
    if (index <= 0) return;
    const [item] = ordered.splice(index, 1);
    setProductionQueueSequence([item, ...ordered]);
  }

  function productionPlanPlant(plan: ProductionPlan | undefined) {
    if (!plan) return undefined;
    return productionPlants().find((item) => item.id === plan.plantId) || productionPlants().find((item) => item.code === plan.plantCode) || firstProductionPlant(plan.siteId);
  }

  function productionTaskPlant(task: ProductionTask | undefined) {
    if (!task) return undefined;
    return productionPlants().find((item) => item.id === task.plantId) || productionPlants().find((item) => item.code === task.plantCode) || firstProductionPlant(task.siteId);
  }

  function productionMixLabel(mix: MixDesign | undefined) {
    return mix ? `${mix.code} ${mix.version}` : "-";
  }

  function productionMixProfiles() {
    const overviewProfiles = list(data.production?.mixDesignPlantProfiles);
    return overviewProfiles.length ? overviewProfiles : list(bootstrap?.mixDesignPlantProfiles);
  }

  function productionProfileById(id: number | undefined) {
    if (!id) return undefined;
    return productionMixProfiles().find((item) => item.id === id);
  }

  function productionProfileLabel(profile: MixDesignPlantProfile | undefined) {
    return profile ? `${profile.code} ${profile.version}` : "-";
  }

  function productionProfileMatchesDate(profile: MixDesignPlantProfile, planDate: string | undefined) {
    if (!planDate) return true;
    if (profile.effectiveFrom && profile.effectiveFrom > planDate) return false;
    if (profile.effectiveTo && profile.effectiveTo < planDate) return false;
    return true;
  }

  function currentProductionProfile(mixDesignId: number | undefined, plantId: number | undefined, planDate: string | undefined) {
    if (!mixDesignId || !plantId) return undefined;
    const profiles = productionMixProfiles()
      .filter((item) => item.mixDesignId === mixDesignId && item.plantId === plantId && item.status === "approved")
      .filter((item) => productionProfileMatchesDate(item, planDate));
    return profiles.find((item) => item.isCurrent) || profiles[0];
  }

  function productionMixMatchesSite(mix: MixDesign, siteId: number | undefined) {
    return mix.siteId === siteId || mix.siteId === 0 || !siteId;
  }

  function currentApprovedMix(productId: number | undefined, siteId: number | undefined) {
    if (!productId) return undefined;
    const mixes = list(data.production?.mixDesigns);
    return mixes.find((item) => item.productId === productId && productionMixMatchesSite(item, siteId) && item.status === "approved" && item.isCurrent)
      || mixes.find((item) => item.productId === productId && productionMixMatchesSite(item, siteId) && item.status === "approved");
  }

  function productionLineRecipe(plant: Plant) {
    const mixDesigns = list(data.production?.mixDesigns);
    const tasks = list(data.production?.tasks)
      .filter((item) => item.status !== "cancelled" && item.status !== "completed")
      .filter((item) => productionTaskPlant(item)?.id === plant.id)
      .sort((a, b) => {
        const priority = (value: ProductionTask) => value.status === "running" ? 0 : value.status === "pending" ? 1 : 2;
        const diff = priority(a) - priority(b);
        return diff || b.id - a.id;
      });
    const task = tasks[0];
    if (task) {
      const plan = list(data.production?.plans).find((item) => item.id === task.planId);
      const mix = mixDesigns.find((item) => item.id === task.mixDesignId) || currentApprovedMix(task.productId, task.siteId);
      return {
        mix,
        profile: productionProfileById(task.mixProfileId) || productionProfileById(plan?.mixProfileId) || currentProductionProfile(mix?.id || task.mixDesignId, plant.id, plan?.planDate),
        productId: task.productId,
        source: task.status === "running" ? "正在生产" : "已下达任务",
        task,
        plan
      };
    }
    const plans = list(data.production?.plans)
      .filter((item) => item.status !== "cancelled" && item.status !== "completed")
      .filter((item) => productionPlanPlant(item)?.id === plant.id)
      .sort((a, b) => {
        const priority = (value: ProductionPlan) => value.status === "producing" || value.status === "running" ? 0 : 1;
        const diff = priority(a) - priority(b);
        return diff || b.id - a.id;
      });
    const plan = plans[0];
    if (plan) {
      const mix = mixDesigns.find((item) => item.id === plan.mixDesignId) || currentApprovedMix(plan.productId, plan.siteId);
      return {
        mix,
        profile: productionProfileById(plan.mixProfileId) || currentProductionProfile(mix?.id || plan.mixDesignId, plant.id, plan.planDate),
        productId: plan.productId,
        source: plan.status === "producing" || plan.status === "running" ? "计划生产中" : "待生产计划",
        plan
      };
    }
    return { mix: undefined, profile: undefined, productId: 0, source: "暂无生产计划" };
  }

  function availableRoles() {
    return data.systemRoles.length ? data.systemRoles : list(bootstrap?.roles);
  }

  function roleName(code: string | undefined) {
    return availableRoles().find((item) => item.code === code)?.name || code || "-";
  }

  function rolePermissions() {
    return roleForm.permissions.split(/\r?\n|,/).map((item) => item.trim()).filter(Boolean);
  }

  function organizationData(): OrganizationOverview {
    const companies = data.org?.companies?.length ? data.org.companies : list(bootstrap?.companies);
    const departments = data.org?.departments?.length ? data.org.departments : list(bootstrap?.departments);
    const sites = data.org?.sites?.length ? data.org.sites : list(bootstrap?.sites);
    const group = data.org?.group || bootstrap?.groupProfile || {
      name: bootstrap?.license?.customerName || "建材集团",
      code: "GROUP",
      edition: "集团版",
      headquartersCompanyId: firstId(companies),
      operatingMode: "集团总部统管",
      dataArchitecture: "group-company-department"
    };
    const metrics = data.org?.metrics || data.dashboard?.organization || {
      companyCount: companies.length,
      activeCompanyCount: companies.filter((item) => item.status === "active").length,
      siteCount: sites.length,
      runningSiteCount: sites.filter((item) => item.status === "running" || item.status === "active").length,
      departmentCount: departments.length,
      userCount: list(data.system?.security?.users).length
    };
    const orgNodes = list(data.org?.nodes);
    const nodes = (orgNodes.length ? orgNodes : fallbackOrganizationNodes(group, companies, departments))
      .filter((node) => node.kind !== "site" && !node.siteId);
    return {
      group,
      metrics,
      nodes,
      companies,
      departments,
      sites
    };
  }

  function fallbackOrganizationNodes(group: OrganizationOverview["group"], companies: Company[], departments: Department[]): OrganizationNode[] {
    const rootId = `group:${group.code || "GROUP"}`;
    return [
      { id: rootId, parentId: "", kind: "group", name: group.name, code: group.code, region: "", status: "active", companyId: 0, siteId: 0 },
      ...companies.map((item) => ({
        id: `company:${item.id}`,
        parentId: item.parentId ? `company:${item.parentId}` : rootId,
        kind: item.level || "company",
        name: item.name,
        code: item.code,
        region: item.region || "",
        status: item.status,
        companyId: item.id,
        siteId: 0
      })),
      ...departments.map((item) => ({
        id: `department:${item.id}`,
        parentId: item.parentId ? `department:${item.parentId}` : `company:${item.companyId}`,
        kind: "department",
        name: item.name,
        code: item.code,
        region: "",
        status: item.status,
        companyId: item.companyId,
        siteId: 0
      }))
    ];
  }

  function orgLevelLabel(value: string | undefined) {
    const rawValue = value || "";
    const fallback = ({
      headquarters: "集团总部",
      regional: "区域公司",
      subsidiary: "分公司/站点公司",
      company: "公司",
      group: "集团",
      site: "站点",
      department: "部门",
      customer: "客户",
      driver: "司机",
      device: "设备"
    } as Record<string, string>)[rawValue] || rawValue || "-";
    if (organizationCompanyKinds.includes(rawValue)) {
      return dictionaryValueLabel("org_company_level", rawValue, fallback);
    }
    return dictionaryValueLabel("data_scope", rawValue, fallback);
  }

  function orgNodeRank(node: OrganizationNode) {
    if (node.kind === "group") return 0;
    if (organizationCompanyKinds.includes(node.kind)) return 1;
    if (node.kind === "department") return 2;
    return 9;
  }

  function isOrganizationCompanyNode(node: OrganizationNode) {
    return organizationCompanyKinds.includes(node.kind) && !node.siteId && Boolean(node.companyId);
  }

  function orgNodeNumericId(node: OrganizationNode) {
    return Number(String(node.id).split(":")[1] || 0);
  }

  function canCreateChildCompany(node: OrganizationNode) {
    return node.kind === "group" || isOrganizationCompanyNode(node);
  }

  function canCreateChildDepartment(node: OrganizationNode) {
    return isOrganizationCompanyNode(node) || node.kind === "department";
  }

  function organizationBoundSites(companyId: number, sites: Site[]) {
    return sites.filter((site) => site.companyId === companyId);
  }

  function organizationBoundSiteTags(companyId: number, sites: Site[]) {
    const boundSites = organizationBoundSites(companyId, sites);
    if (!boundSites.length) return null;
    const visibleSites = boundSites.slice(0, 4);
    return (
      <span className="org-bound-site-tags" aria-label="已绑定站点">
        {visibleSites.map((site) => (
          <span className="org-bound-site-chip" key={site.id} title={site.address || site.code || site.name}>
            {site.name}
          </span>
        ))}
        {boundSites.length > visibleSites.length ? <span className="org-bound-site-chip is-more">+{boundSites.length - visibleSites.length}</span> : null}
      </span>
    );
  }

  function organizationRowSearchText(node: OrganizationNode, org: OrganizationOverview) {
    const companyName = nameOf(org.companies, node.companyId);
    const boundSiteText = organizationCompanyKinds.includes(node.kind)
      ? organizationBoundSites(node.companyId, org.sites).map((site) => `${site.name} ${site.code} ${site.address || ""}`).join(" ")
      : "";
    return [orgLevelLabel(node.kind), node.name, node.code, node.region, companyName, boundSiteText].filter(Boolean).join(" ");
  }

  function orgNodeRelation(node: OrganizationNode, org: OrganizationOverview) {
    if (node.kind === "group") return node.region || "-";
    const companyName = nameOf(org.companies, node.companyId);
    if (organizationCompanyKinds.includes(node.kind)) {
      return (
        <span className="org-table-relation">
          <b>{node.region || org.group.name || "-"}</b>
          {organizationBoundSiteTags(node.companyId, org.sites)}
        </span>
      );
    }
    if (node.kind === "department") return companyName || "-";
    return node.region || companyName || org.group.name || "-";
  }

  function organizationChildMap(nodes: OrganizationNode[]) {
    const map = new Map<string, OrganizationNode[]>();
    nodes.forEach((node) => {
      const key = node.parentId || "";
      map.set(key, [...(map.get(key) || []), node]);
    });
    map.forEach((items) => items.sort((a, b) => orgNodeRank(a) - orgNodeRank(b) || a.name.localeCompare(b.name)));
    return map;
  }

  function organizationTreeRows(nodes: OrganizationNode[], childMap: Map<string, OrganizationNode[]>, collapsedNodeIds: Set<string>): OrganizationTreeRow[] {
    const rows: OrganizationTreeRow[] = [];
    const seen = new Set<string>();
    const nodeIds = new Set(nodes.map((node) => node.id));
    const roots = childMap.get("")?.length
      ? childMap.get("") || []
      : nodes.filter((node) => !node.parentId || !nodeIds.has(node.parentId));

    function markDescendantsSeen(node: OrganizationNode) {
      (childMap.get(node.id) || []).forEach((child) => {
        if (seen.has(child.id)) return;
        seen.add(child.id);
        markDescendantsSeen(child);
      });
    }

    function visit(node: OrganizationNode, depth: number) {
      if (seen.has(node.id)) return;
      seen.add(node.id);
      rows.push({ ...node, depth });
      if (collapsedNodeIds.has(node.id)) {
        markDescendantsSeen(node);
        return;
      }
      (childMap.get(node.id) || []).forEach((child) => visit(child, depth + 1));
    }

    roots.forEach((node) => visit(node, 0));
    nodes.forEach((node) => {
      if (!seen.has(node.id)) visit(node, 0);
    });
    return rows;
  }

  function toggleOrgNodeCollapsed(nodeId: string) {
    setCollapsedOrgNodeIds((current) => {
      const next = new Set(current);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }

  function orgNodeStatusAction(node: OrganizationNode) {
    const canWriteOrg = hasPermission(currentPermissions, "org:write");
    if (node.kind === "department" && node.companyId) {
      return (
        <UiButton className="org-node-action" size="sm" disabled={actionBusy !== "" || !canWriteOrg} onClick={() => handleOrgStatus("departments", Number(node.id.split(":")[1]), node.status === "active" ? "disabled" : "active")}>
          {node.status === "active" ? "禁用" : "启用"}
        </UiButton>
      );
    }
    if (node.companyId && !node.siteId && organizationCompanyKinds.includes(node.kind)) {
      return (
        <UiButton className="org-node-action" size="sm" disabled={actionBusy !== "" || !canWriteOrg} onClick={() => handleOrgStatus("companies", node.companyId, node.status === "active" ? "disabled" : "active")}>
          {node.status === "active" ? "禁用" : "启用"}
        </UiButton>
      );
    }
    return null;
  }

  function orgNodeIcon(node: OrganizationNode) {
    if (node.kind === "site") return <Factory size={15} />;
    if (node.kind === "department") return <Users size={15} />;
    return <Building2 size={15} />;
  }

  function nextOrgCode(prefix: string, codes: Array<string | undefined>) {
    const usedCodes = new Set(codes.map((code) => (code || "").trim().toUpperCase()).filter(Boolean));
    for (let index = usedCodes.size + 1; index < usedCodes.size + 1000; index += 1) {
      const candidate = `${prefix}-${String(index).padStart(3, "0")}`;
      if (!usedCodes.has(candidate)) return candidate;
    }
    return `${prefix}-${Date.now()}`;
  }

  function resetOrgForm() {
    const org = organizationData();
    const companyId = firstId(org.companies);
    const departmentCodes = org.departments.filter((item) => item.companyId === companyId).map((item) => item.code);
    setOrgForm({
      companyName: "",
      companyCode: nextOrgCode("BRANCH", org.companies.map((item) => item.code)),
      companyParentId: String(companyId),
      companyLevel: firstDictionaryCode("org_company_level", "subsidiary"),
      companyRegion: "",
      departmentCompanyId: String(companyId),
      departmentParentId: "0",
      departmentName: "",
      departmentCode: nextOrgCode("OPS", departmentCodes)
    });
  }

  function openCreateChildCompanyDialog(parent: OrganizationNode) {
    const org = organizationData();
    const parentCompanyId = isOrganizationCompanyNode(parent) ? parent.companyId : 0;
    const defaultCompanyId = firstId(org.companies);
    const companyLevel = parent.kind === "group" ? "regional" : "subsidiary";
    const prefix = parent.kind === "group" ? "REGION" : "BRANCH";
    setOrgForm((form) => ({
      ...form,
      companyName: "",
      companyCode: nextOrgCode(prefix, org.companies.map((item) => item.code)),
      companyParentId: String(parentCompanyId),
      companyLevel,
      companyRegion: parent.region || form.companyRegion || "",
      departmentCompanyId: String(parentCompanyId || defaultCompanyId),
      departmentParentId: "0"
    }));
    setActionDialogId("org-company-create");
  }

  function openCreateChildDepartmentDialog(parent: OrganizationNode) {
    const org = organizationData();
    const companyId = parent.kind === "department"
      ? parent.companyId
      : isOrganizationCompanyNode(parent)
        ? parent.companyId
        : firstId(org.companies);
    const parentDepartmentId = parent.kind === "department" ? orgNodeNumericId(parent) : 0;
    const departmentCodes = org.departments.filter((item) => item.companyId === companyId).map((item) => item.code);
    setOrgForm((form) => ({
      ...form,
      departmentCompanyId: String(companyId),
      departmentParentId: String(parentDepartmentId),
      departmentName: "",
      departmentCode: nextOrgCode("DEPT", departmentCodes),
      companyParentId: String(companyId || firstId(org.companies))
    }));
    setActionDialogId("org-department-create");
  }

  async function handleCreateCompany(event?: FormEvent) {
    event?.preventDefault();
    await runBusinessAction("org-company-save", "公司已创建", async () => {
      await api.createCompany({
        name: orgForm.companyName,
        code: orgForm.companyCode,
        parentId: fieldNumber(orgForm.companyParentId),
        level: orgForm.companyLevel,
        region: orgForm.companyRegion,
        status: "active"
      });
      resetOrgForm();
      closeActionDialog("org-company-create");
    });
  }

  async function handleCreateDepartment(event?: FormEvent) {
    event?.preventDefault();
    await runBusinessAction("org-department-save", "部门已创建", async () => {
      await api.createDepartment({
        companyId: fieldNumber(orgForm.departmentCompanyId),
        parentId: fieldNumber(orgForm.departmentParentId),
        name: orgForm.departmentName,
        code: orgForm.departmentCode,
        status: "active"
      });
      resetOrgForm();
      closeActionDialog("org-department-create");
    });
  }

  async function handleOrgStatus(resource: "companies" | "departments", id: number, status: string) {
    await runBusinessAction(`org-status-${resource}-${id}`, "组织状态已更新", () => api.updateOrgStatus(resource, id, status));
  }

  function resetUserForm() {
    setEditingUserId(null);
    setUserForm({
      username: "",
      displayName: "",
      password: "",
      roleCode: availableRoles()[0]?.code || "dispatcher",
      companyId: String(firstId(bootstrap?.companies)),
      siteId: String(defaultSiteId),
      customerId: "0",
      driverId: "0",
      status: "active"
    });
  }

  function startUserEdit(item: User) {
    setEditingUserId(item.id);
    setUserForm({
      username: item.username,
      displayName: item.displayName,
      password: "",
      roleCode: item.roleCode,
      companyId: String(item.companyId || firstId(bootstrap?.companies)),
      siteId: String(item.siteId || 0),
      customerId: String(item.customerId || 0),
      driverId: String(item.driverId || 0),
      status: item.status || "active"
    });
  }

  async function handleSaveUser(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-user-save", editingUserId ? "用户已更新" : "用户已创建", async () => {
      await api.saveSystemUser({
        id: editingUserId || undefined,
        username: userForm.username,
        displayName: userForm.displayName,
        password: userForm.password || undefined,
        roleCode: userForm.roleCode,
        companyId: fieldNumber(userForm.companyId),
        siteId: fieldNumber(userForm.siteId),
        customerId: fieldNumber(userForm.customerId),
        driverId: fieldNumber(userForm.driverId),
        status: userForm.status
      });
      resetUserForm();
    });
  }

  async function handleUserStatus(item: User, status: string) {
    await runBusinessAction(`system-user-status-${item.id}`, "用户状态已更新", () => api.setSystemUserStatus(item.id, status));
  }

  async function handleEnrollUserMFA(item: User) {
    await runBusinessAction(`system-user-mfa-enroll-${item.id}`, "MFA 密钥已生成", async () => {
      const enrollment = await api.enrollMFA(item.id);
      setMfaEnrollment(enrollment);
      setMfaCodes((codes) => ({ ...codes, [item.id]: "" }));
    }, null);
  }

  async function handleEnableUserMFA(item: User) {
    const code = (mfaCodes[item.id] || "").trim();
    if (!code) {
      setActionError("请输入 MFA 动态码");
      return;
    }
    await runBusinessAction(`system-user-mfa-enable-${item.id}`, "MFA 已启用", async () => {
      await api.enableMFA(item.id, code);
      setMfaEnrollment(null);
      setMfaCodes((codes) => ({ ...codes, [item.id]: "" }));
    });
  }

  async function handleDisableUserMFA(item: User) {
    await runBusinessAction(`system-user-mfa-${item.id}`, "MFA 已关闭", async () => {
      await api.disableMFA(item.id);
      setMfaEnrollment((enrollment) => enrollment?.user.id === item.id ? null : enrollment);
      setMfaCodes((codes) => ({ ...codes, [item.id]: "" }));
    });
  }

  function resetRoleForm() {
    setEditingRoleId(null);
    setRoleForm({
      code: "",
      name: "",
      dataScope: "site",
      permissions: ""
    });
  }

  function startRoleEdit(item: Role) {
    setEditingRoleId(item.id);
    setRoleForm({
      code: item.code,
      name: item.name,
      dataScope: item.dataScope === "platform" ? "group" : item.dataScope || "group",
      permissions: list(item.permissions).join("\n")
    });
  }

  async function handleSaveRole(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-role-save", editingRoleId ? "角色已更新" : "角色已创建", async () => {
      await api.saveSystemRole({
        id: editingRoleId || undefined,
        code: roleForm.code,
        name: roleForm.name,
        dataScope: roleForm.dataScope,
        permissions: rolePermissions()
      });
      resetRoleForm();
    });
  }

  async function handleDeleteRole(item: Role) {
    await runBusinessAction(`system-role-delete-${item.id}`, "角色已删除", () => api.deleteSystemRole(item.id));
  }

  function resetFieldPolicyForm() {
    setFieldPolicyForm({
      id: "",
      roleCode: availableRoles()[0]?.code || "dispatcher",
      resource: "customers",
      field: "phone",
      mask: "phone",
      remark: ""
    });
  }

  function startFieldPolicyEdit(item: FieldPolicy) {
    setFieldPolicyForm({
      id: String(item.id),
      roleCode: item.roleCode,
      resource: item.resource,
      field: item.field,
      mask: item.mask || "phone",
      remark: item.remark || ""
    });
  }

  async function handleCreateFieldPolicy(event: FormEvent) {
    event.preventDefault();
    const policyId = fieldNumber(fieldPolicyForm.id);
    await runBusinessAction(`system-field-policy-${policyId || "create"}`, policyId ? "字段策略已更新" : "字段策略已创建", async () => {
      const payload = {
        roleCode: fieldPolicyForm.roleCode,
        resource: fieldPolicyForm.resource,
        field: fieldPolicyForm.field,
        mask: fieldPolicyForm.mask,
        remark: fieldPolicyForm.remark
      };
      if (policyId) {
        await api.updateFieldPolicy(policyId, payload);
      } else {
        await api.createFieldPolicy(payload);
      }
      resetFieldPolicyForm();
    });
  }

  async function handleToggleFieldPolicy(item: FieldPolicy) {
    await runBusinessAction(`system-field-policy-toggle-${item.id}`, "字段策略状态已更新", () => api.toggleFieldPolicy(item.id, !item.enabled));
  }

  async function handleDeleteFieldPolicy(item: FieldPolicy) {
    await runBusinessAction(`system-field-policy-delete-${item.id}`, "字段策略已删除", () => api.deleteFieldPolicy(item.id));
  }

  function allDictionaries() {
    const systemDictionaries = list(data.system?.dictionaries);
    return systemDictionaries.length ? systemDictionaries : list(bootstrap?.dictionaries);
  }

  function dictionaryTypeLabel(type: string | undefined) {
    const value = type || "";
    return dictionaryTypePresets.find((item) => item.type === value)?.label || value || "-";
  }

  function dictionaryOptions(type: string) {
    return activeDictionaryOptions(allDictionaries(), type);
  }

  function dictionaryOptionsWithFallback(type: string, fallbackItems: DictionaryOption[]) {
    return activeDictionaryOptions(allDictionaries(), type, fallbackItems);
  }

  function dictionaryValueLabel(type: string, code: string | null | undefined, fallbackLabel?: string) {
    return dictionaryLabel(allDictionaries(), type, code, fallbackLabel);
  }

  function firstDictionaryCode(type: string, fallbackCode: string) {
    return dictionaryOptions(type)[0]?.code || fallbackCode;
  }

  function nextDictionarySort(type: string) {
    return String(allDictionaries()
      .filter((item) => item.type === type)
      .reduce((max, item) => Math.max(max, item.sort || 0), 0) + 1);
  }

  function resetDictionaryForm(type = dictionaryFilters.type !== "all" ? dictionaryFilters.type : "product_line") {
    setEditingDictionaryId(null);
    setDictionaryForm({
      type,
      code: "",
      label: "",
      sort: nextDictionarySort(type),
      status: "active"
    });
  }

  function startDictionaryEdit(item: DataDictionary) {
    setEditingDictionaryId(item.id);
    setDictionaryForm({
      type: item.type,
      code: item.code,
      label: item.label,
      sort: String(item.sort || nextDictionarySort(item.type)),
      status: item.status || "active"
    });
  }

  async function handleSaveDictionary(event: FormEvent) {
    event.preventDefault();
    const savedType = dictionaryForm.type;
    await runBusinessAction("system-dictionary-save", editingDictionaryId ? "字典项已更新" : "字典项已创建", async () => {
      await api.saveDictionary({
        id: editingDictionaryId || undefined,
        type: savedType,
        code: dictionaryForm.code,
        label: dictionaryForm.label,
        sort: fieldNumber(dictionaryForm.sort, 1),
        status: dictionaryForm.status
      });
      setDictionaryFilters((value) => ({ ...value, type: savedType }));
      resetDictionaryForm(savedType);
    });
  }

  async function handleDictionaryStatus(item: DataDictionary, status: string) {
    await runBusinessAction(`system-dictionary-status-${item.id}`, "字典状态已更新", () => api.setDictionaryStatus(item.id, status));
  }

  async function handleDeleteDictionary(item: DataDictionary) {
    await runBusinessAction(`system-dictionary-delete-${item.id}`, "字典项已删除", () => api.deleteDictionary(item.id));
  }

  function menuLabelActionKey(key: string) {
    return `system-menu-label-${key}`;
  }

  function openMenuLabelDialog(key: string, currentLabel: string, title: string) {
    setMenuLabelDialog({ key, currentLabel, title });
    setMenuLabelMenu(null);
    setMenuLabelForm({ label: currentLabel });
    setActionError("");
  }

  async function handleSaveMenuLabel(event: FormEvent) {
    event.preventDefault();
    if (!menuLabelDialog) return;
    const label = menuLabelForm.label.trim();
    if (!label) {
      setActionError("菜单显示名称不能为空");
      return;
    }
    await runBusinessAction(menuLabelActionKey(menuLabelDialog.key), "菜单名称已保存", async () => {
      await api.saveMenuLabel({ key: menuLabelDialog.key, label });
      setMenuLabelDialog(null);
    }, null);
  }

  async function handleResetMenuLabel(key: string) {
    await runBusinessAction(menuLabelActionKey(key), "菜单名称已恢复默认", async () => {
      await api.resetMenuLabel(key);
      setMenuLabelDialog(null);
      setMenuLabelMenu(null);
    }, null);
  }

  function workflowOverview(): WorkflowOverview {
    return data.system?.workflows || { definitions: [], instances: [], tasks: [], inbox: [], events: [], logs: [], outbox: [], subscriptions: [], deliveries: [] };
  }

  function workflowDefinitions() {
    return list(workflowOverview().definitions);
  }

  function workflowTasks() {
    return list(workflowOverview().tasks);
  }

  function workflowInbox() {
    return list(workflowOverview().inbox);
  }

  function workflowInstances() {
    return list(workflowOverview().instances);
  }

  function workflowEvents() {
    return list(workflowOverview().events);
  }

  function workflowLogs() {
    return list(workflowOverview().logs);
  }

  function workflowOutbox() {
    return list(workflowOverview().outbox);
  }

  function workflowSubscriptions() {
    return list(workflowOverview().subscriptions);
  }

  function workflowDeliveries() {
    return list(workflowOverview().deliveries);
  }

  function workflowStepsFromText() {
    return workflowForm.steps
      .split(/\r?\n/)
      .map((line, index) => {
        const [seqRaw, roleCodeRaw, actionRaw, nameRaw, slaHoursRaw] = line.split("|").map((item) => item.trim());
        const seq = fieldNumber(seqRaw, index + 1);
        const roleCode = roleCodeRaw || availableRoles()[0]?.code || "boss";
        return {
          seq,
          code: `${workflowForm.code}.step.${seq}`,
          name: nameRaw || `步骤 ${seq}`,
          type: workflowForm.category || "approval",
          roleCode,
          action: actionRaw || "approve",
          slaHours: Math.max(0, fieldNumber(slaHoursRaw))
        };
      })
      .filter((item) => item.roleCode);
  }

  function workflowStepsText(item: WorkflowDefinition) {
    return list(item.steps)
      .map((step) => `${step.seq}|${step.roleCode}|${step.action || "approve"}|${step.name || `步骤 ${step.seq}`}|${step.slaHours || ""}`)
      .join("\n");
  }

  function workflowConditionsFromText() {
    return workflowForm.triggerConditions
      .split(/\r?\n/)
      .map((line) => {
        const [field, operator, value] = line.split("|").map((item) => item.trim());
        return { field, operator: operator || "equals", value: value || "" };
      })
      .filter((item) => item.field);
  }

  function workflowConditionsText(item: WorkflowDefinition) {
    return list(item.trigger?.conditions)
      .map((condition) => `${condition.field}|${condition.operator || "equals"}|${condition.value || ""}`)
      .join("\n");
  }

  function workflowConditionsToText(conditions: Array<{ field: string; operator?: string; value?: string }>) {
    return conditions
      .map((condition) => `${condition.field}|${condition.operator || "equals"}|${condition.value || ""}`)
      .join("\n");
  }

  function setWorkflowCondition(index: number, patch: Partial<{ field: string; operator: string; value: string }>) {
    const next = workflowConditionsFromText().map((condition, conditionIndex) => {
      if (conditionIndex !== index) return condition;
      return { ...condition, ...patch };
    });
    setWorkflowForm((form) => ({ ...form, triggerConditions: workflowConditionsToText(next) }));
  }

  function addWorkflowCondition() {
    const next = [...workflowConditionsFromText(), { field: "eventType", operator: "equals", value: workflowForm.triggerEventType || "" }];
    setWorkflowForm((form) => ({ ...form, triggerConditions: workflowConditionsToText(next) }));
  }

  function removeWorkflowCondition(index: number) {
    const next = workflowConditionsFromText().filter((_, conditionIndex) => conditionIndex !== index);
    setWorkflowForm((form) => ({ ...form, triggerConditions: workflowConditionsToText(next) }));
  }

  function workflowStepsToText(steps: Array<{ seq: number; roleCode: string; action?: string; name?: string; slaHours?: number }>) {
    return steps
      .map((step, index) => `${index + 1}|${step.roleCode}|${step.action || "approve"}|${step.name || `步骤 ${index + 1}`}|${step.slaHours || ""}`)
      .join("\n");
  }

	  function workflowTemplatePresets(): WorkflowPresetDraft[] {
	    const catalogEvents = list(workflowOverview().catalog?.events);
	    if (!catalogEvents.length) return [];
	    return catalogEvents.map((event) => ({
	      label: event.label,
	      code: event.code,
	      name: event.name,
	      resource: event.resource,
      eventType: event.eventType,
      description: event.description,
      variables: event.variables || [],
      triggers: event.triggers || [],
      conditions: workflowConditionsToText(event.conditions || []),
	      steps: workflowStepsToText((event.steps || []).map((step) => ({
	        seq: step.seq,
        roleCode: step.roleCode,
        action: step.action,
        name: step.name,
        slaHours: step.slaHours
      })))
	    }));
	  }

	  function selectedWorkflowPreset() {
	    const presets = workflowTemplatePresets();
	    return presets.find((preset) => preset.code === workflowPresetCode)
	      || presets.find((preset) => preset.eventType === workflowForm.triggerEventType && preset.resource === workflowForm.resource)
	      || null;
	  }

		  function workflowResourceOptionList() {
		    const resources = list(workflowOverview().catalog?.resources);
		    return resources;
		  }

		  function workflowOutboxEventOptionList() {
		    const outboxEvents = list(workflowOverview().catalog?.outboxEvents);
		    const base = outboxEvents;
		    const current = workflowSubscriptionForm.eventType;
		    if (current && !base.some((item) => item.eventType === current)) {
		      return [{ eventType: current, label: current, description: "当前自定义出口事件", payloadFields: [] }, ...base];
		    }
		    return base;
		  }

		  function selectedWorkflowOutboxEvent() {
		    return workflowOutboxEventOptionList().find((item) => item.eventType === workflowSubscriptionForm.eventType) || null;
		  }

		  function workflowConditionFieldOptionList() {
		    const fields = list(workflowOverview().catalog?.conditionFields);
		    const base = fields.map((field) => ({ value: field.key, label: field.label }));
	    const variables = list(selectedWorkflowPreset()?.variables).map((field) => ({ value: field.key, label: field.label || field.key }));
	    const seen = new Set<string>();
	    return [...base, ...variables].filter((item) => {
	      if (seen.has(item.value)) return false;
	      seen.add(item.value);
	      return true;
	    });
	  }

	  function workflowResourceLabel(value: string) {
	    return workflowResourceOptionList().find((option) => option.value === value)?.label || value || "-";
	  }

	  function workflowOutboxEventLabel(value: string) {
	    return workflowOutboxEventOptionList().find((option) => option.eventType === value)?.label || value || "-";
	  }

  function workflowDraftSteps() {
    const steps = workflowStepsFromText();
    if (steps.length) {
      return steps;
    }
    return [{
      seq: 1,
      code: `${workflowForm.code}.step.1`,
      name: "审批确认",
      type: workflowForm.category || "approval",
      roleCode: availableRoles()[0]?.code || "boss",
      action: "approve",
      slaHours: 24
    }];
  }

  function setWorkflowStep(index: number, patch: Partial<{ roleCode: string; action: string; name: string; slaHours: number }>) {
    const next = workflowDraftSteps().map((step, stepIndex) => {
      if (stepIndex !== index) return step;
      return { ...step, ...patch };
    });
    setWorkflowForm((form) => ({ ...form, steps: workflowStepsToText(next) }));
  }

  function addWorkflowStep(roleCode = availableRoles()[0]?.code || "boss") {
    const next = [...workflowDraftSteps(), {
      seq: workflowDraftSteps().length + 1,
      code: `${workflowForm.code}.step.${workflowDraftSteps().length + 1}`,
      name: `步骤 ${workflowDraftSteps().length + 1}`,
      type: workflowForm.category || "approval",
      roleCode,
      action: "approve",
      slaHours: 24
    }];
    setWorkflowForm((form) => ({ ...form, steps: workflowStepsToText(next) }));
  }

  function removeWorkflowStep(index: number) {
    const next = workflowDraftSteps().filter((_, stepIndex) => stepIndex !== index);
    setWorkflowForm((form) => ({ ...form, steps: workflowStepsToText(next.length ? next : workflowDraftSteps().slice(0, 1)) }));
  }

	  function applyWorkflowPreset(preset: WorkflowPresetDraft, resetEditing = true) {
	    if (resetEditing) setEditingWorkflowId(null);
	    setWorkflowPresetCode(preset.code);
	    setWorkflowForm({
      code: preset.code,
      name: preset.name,
      category: "approval",
      resource: preset.resource,
      status: "active",
      version: "1",
      triggerEventType: preset.eventType,
      triggerResource: preset.resource,
      triggerConditions: preset.conditions,
	      steps: preset.steps
	    });
	  }

	  function selectWorkflowPreset(code: string) {
	    const preset = workflowTemplatePresets().find((item) => item.code === code);
	    setWorkflowPresetCode(code);
	    if (preset) applyWorkflowPreset(preset);
	  }

  function resetWorkflowForm() {
    setEditingWorkflowId(null);
    const preset = workflowTemplatePresets()[0];
    if (preset) {
      applyWorkflowPreset(preset, false);
      return;
    }
    setWorkflowPresetCode("");
    setWorkflowForm({
      code: "",
      name: "",
      category: "approval",
      resource: "",
      status: "active",
      version: "1",
      triggerEventType: "",
      triggerResource: "",
      triggerConditions: "",
      steps: ""
    });
  }

	  function startWorkflowEdit(item: WorkflowDefinition) {
	    setEditingWorkflowId(item.id);
	    const matchedPreset = workflowTemplatePresets().find((preset) => preset.code === item.code || (preset.eventType === item.trigger?.eventType && preset.resource === item.resource));
	    setWorkflowPresetCode(matchedPreset?.code || item.code);
	    setWorkflowForm({
      code: item.code,
      name: item.name,
      category: item.category || "approval",
      resource: item.resource || "",
      status: item.status || "active",
      version: String(item.version || 1),
      triggerEventType: item.trigger?.eventType || `${item.resource || ""}.submitted`,
      triggerResource: item.trigger?.resource || item.resource || "",
      triggerConditions: workflowConditionsText(item),
      steps: workflowStepsText(item)
    });
  }

  function workflowDefinitionPayload() {
    return {
      id: editingWorkflowId || undefined,
      code: workflowForm.code,
      name: workflowForm.name,
      category: workflowForm.category,
      resource: workflowForm.resource,
      status: workflowForm.status,
      version: fieldNumber(workflowForm.version, 1),
      trigger: {
        eventType: workflowForm.triggerEventType,
        resource: workflowForm.triggerResource || workflowForm.resource,
        conditions: workflowConditionsFromText()
      },
      steps: workflowStepsFromText()
    };
  }

  async function handleSaveWorkflow(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-workflow-save", editingWorkflowId ? "工作流已更新" : "工作流已创建", async () => {
      await api.saveWorkflowDefinition(workflowDefinitionPayload());
      resetWorkflowForm();
    });
  }

  async function handlePublishWorkflowVersion() {
    if (!editingWorkflowId) return;
    await runBusinessAction("system-workflow-version-publish", "工作流新版本已发布", async () => {
      await api.publishWorkflowDefinitionVersion(editingWorkflowId, workflowDefinitionPayload());
      resetWorkflowForm();
    });
  }

  async function handleRollbackWorkflowVersion(id: number) {
    await runBusinessAction(`system-workflow-version-rollback-${id}`, "工作流版本已回滚", () => api.rollbackWorkflowDefinitionVersion(id));
  }

  function resetApprovalFlowForm(item?: ApprovalFlow) {
    const roleCode = availableRoles()[0]?.code || "boss";
    setApprovalFlowForm({
      id: item ? String(item.id) : "",
      code: item?.code || "",
      name: item?.name || "",
      resource: item?.resource || "sales_order",
      steps: JSON.stringify(item?.steps?.length ? item.steps : [{ seq: 1, roleCode, action: "approve" }], null, 2),
      status: item?.status || "active"
    });
  }

  async function handleSaveApprovalFlow(event: FormEvent) {
    event.preventDefault();
    let steps: ApprovalFlow["steps"] = [];
    try {
      const parsed = JSON.parse(approvalFlowForm.steps || "[]");
      steps = Array.isArray(parsed) ? parsed as ApprovalFlow["steps"] : [];
    } catch {
      setActionError("审批步骤必须是 JSON 数组");
      return;
    }
    await runBusinessAction("system-approval-flow-save", approvalFlowForm.id ? "审批流已更新" : "审批流已创建", async () => {
      await api.saveApprovalFlow({
        id: fieldNumber(approvalFlowForm.id) || undefined,
        code: approvalFlowForm.code.trim(),
        name: approvalFlowForm.name.trim(),
        resource: approvalFlowForm.resource.trim(),
        steps,
        status: approvalFlowForm.status
      });
      resetApprovalFlowForm();
    });
  }

  async function handleApprovalFlowStatus(item: ApprovalFlow, status: string) {
    await runBusinessAction(`system-approval-flow-status-${item.id}-${status}`, "审批流状态已更新", () => api.setApprovalFlowStatus(item.id, status));
  }

  async function handleDeleteApprovalFlow(item: ApprovalFlow) {
    await runBusinessAction(`system-approval-flow-delete-${item.id}`, "审批流已删除", () => api.deleteApprovalFlow(item.id));
  }

  function resetWorkflowSubscriptionForm() {
    setEditingWorkflowSubscriptionId(null);
    const eventType = workflowOutboxEventOptionList()[0]?.eventType || "";
    setWorkflowSubscriptionForm({
      code: "",
      name: "",
      eventType,
      resource: "",
      definitionCode: "",
      targetType: "webhook",
      endpoint: "",
      secret: "",
      retryLimit: "3",
      timeoutSeconds: "5",
      status: "active"
    });
  }

  function startWorkflowSubscriptionEdit(item: WorkflowSubscription) {
    setEditingWorkflowSubscriptionId(item.id);
    setWorkflowSubscriptionForm({
      code: item.code,
      name: item.name,
      eventType: item.eventType || "workflow.*",
      resource: item.resource || "",
      definitionCode: item.definitionCode || "",
      targetType: item.targetType || "webhook",
	      endpoint: item.endpoint,
	      secret: item.secret || "",
	      retryLimit: String(item.retryLimit || 3),
	      timeoutSeconds: String(item.timeoutSeconds || 5),
	      status: item.status || "active"
	    });
  }

  function workflowSubscriptionPayload() {
    return {
      id: editingWorkflowSubscriptionId || undefined,
      code: workflowSubscriptionForm.code,
      name: workflowSubscriptionForm.name,
      eventType: workflowSubscriptionForm.eventType,
      resource: workflowSubscriptionForm.resource || undefined,
      definitionCode: workflowSubscriptionForm.definitionCode || undefined,
      targetType: workflowSubscriptionForm.targetType,
	      endpoint: workflowSubscriptionForm.endpoint.trim(),
	      secret: workflowSubscriptionForm.secret || undefined,
	      retryLimit: fieldNumber(workflowSubscriptionForm.retryLimit, 3),
	      timeoutSeconds: fieldNumber(workflowSubscriptionForm.timeoutSeconds, 5),
	      status: workflowSubscriptionForm.status
	    };
  }

  function validateWorkflowSubscriptionForm() {
    const endpoint = workflowSubscriptionForm.endpoint.trim();
    if (!endpoint) return "请输入 Webhook 地址";
    if (/^mock:\/\//i.test(endpoint)) return "Webhook 地址不能使用 mock:// 模拟端点";
    if (!/^https?:\/\//i.test(endpoint)) return "Webhook 地址必须使用 http:// 或 https://";
    return "";
  }

  async function handleSaveWorkflowSubscription(event: FormEvent) {
    event.preventDefault();
    const validationError = validateWorkflowSubscriptionForm();
    if (validationError) {
      setActionError(validationError);
      return;
    }
    await runBusinessAction("system-workflow-subscription-save", editingWorkflowSubscriptionId ? "事件订阅已更新" : "事件订阅已创建", async () => {
      await api.saveWorkflowSubscription(workflowSubscriptionPayload());
      resetWorkflowSubscriptionForm();
    });
  }

  async function handleWorkflowSubscriptionStatus(item: WorkflowSubscription, status: string) {
    await runBusinessAction(`system-workflow-subscription-status-${item.id}`, "事件订阅状态已更新", () => api.setWorkflowSubscriptionStatus(item.id, status));
  }

  async function handleDeleteWorkflowSubscription(item: WorkflowSubscription) {
    await runBusinessAction(`system-workflow-subscription-delete-${item.id}`, "事件订阅已删除", () => api.deleteWorkflowSubscription(item.id));
  }

  async function dispatchWorkflowDelivery(id: number) {
    await runBusinessAction(`workflow-delivery-dispatch-${id}`, "工作流事件已投递", () => api.dispatchWorkflowDelivery(id));
  }

  async function dispatchDueWorkflowDeliveries() {
    const prompt = sensitiveActionPrompt("workflow-delivery-dispatch-due", "批量投递到期工作流事件");
    if (prompt) {
      const confirmed = await confirmMessage({
        title: prompt.title,
        message: prompt.message,
        tone: "warning",
        confirmLabel: prompt.confirmLabel,
        confirmVariant: prompt.confirmVariant
      });
      if (!confirmed) return;
    }
    setActionBusy("workflow-delivery-dispatch-due");
    setActionError("");
    try {
      const batch = await api.dispatchDueWorkflowDeliveries(50);
      message.success(`已调度 ${batch.dispatched} 条，成功 ${batch.succeeded} 条，失败 ${batch.failed} 条`);
      onChanged();
      await load();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setActionBusy("");
    }
  }

  async function runWorkflowAutomation() {
    const prompt = sensitiveActionPrompt("workflow-automation-run", "执行工作流自动化");
    if (prompt) {
      const confirmed = await confirmMessage({
        title: prompt.title,
        message: prompt.message,
        tone: "warning",
        confirmLabel: prompt.confirmLabel,
        confirmVariant: prompt.confirmVariant
      });
      if (!confirmed) return;
    }
    setActionBusy("workflow-automation-run");
    setActionError("");
    try {
      const run = await api.runWorkflowAutomation(50, 50);
      message.success(`投递成功 ${run.deliveries.succeeded} 条，失败 ${run.deliveries.failed} 条，升级 ${run.escalated} 条`);
      onChanged();
      await load();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setActionBusy("");
    }
  }

  function workflowEventVariablesFromText() {
    return workflowEventForm.variables
      .split(/\r?\n/)
      .map((line) => {
        const separator = line.includes("=") ? "=" : "|";
        const [key, ...rest] = line.split(separator);
        return [key?.trim() || "", rest.join(separator).trim()] as const;
      })
      .filter(([key]) => key)
      .reduce<Record<string, string>>((out, [key, value]) => {
        out[key] = value;
        return out;
      }, {});
  }

  function workflowEventPayload() {
    return {
      eventType: workflowEventForm.eventType,
      source: workflowEventForm.source,
      eventKey: workflowEventForm.eventKey,
      actor: workflowEventForm.actor,
      resource: workflowEventForm.resource,
      resourceId: fieldNumber(workflowEventForm.resourceId),
      resourceNo: workflowEventForm.resourceNo,
      title: workflowEventForm.title,
      reason: workflowEventForm.reason,
      variables: workflowEventVariablesFromText()
    };
  }

  function updateWorkflowEventForm(patch: Partial<typeof workflowEventForm>) {
    setWorkflowEventPreview(null);
    setWorkflowEventForm((current) => ({ ...current, ...patch }));
  }

  function resetWorkflowEventForm() {
    setWorkflowEventPreview(null);
    const preset = selectedWorkflowPreset() || workflowTemplatePresets()[0];
    setWorkflowEventForm({
      eventType: preset?.eventType || "",
      source: "manual",
      eventKey: "",
      actor: "",
      resource: preset?.resource || "",
      resourceId: "",
      resourceNo: "",
      title: "",
      reason: "",
      variables: ""
    });
  }

  async function handlePreviewWorkflowEvent() {
    setActionBusy("system-workflow-event-preview");
    setActionError("");
    try {
      setWorkflowEventPreview(await api.previewWorkflowEvent(workflowEventPayload()));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setActionBusy("");
    }
  }

  async function handlePublishWorkflowEvent(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-workflow-event-publish", "工作流事件已发布", async () => {
      await api.publishWorkflowEvent(workflowEventPayload());
      setWorkflowEventPreview(null);
    });
  }

  async function confirmSensitiveAction(prompt: SensitiveActionPrompt | null) {
    if (!prompt) {
      return true;
    }
    return confirmMessage({
      title: prompt.title,
      message: prompt.message,
      tone: "warning",
      confirmLabel: prompt.confirmLabel,
      confirmVariant: prompt.confirmVariant
    });
  }

  async function runBusinessAction(label: string, success: string, action: () => Promise<unknown>, prompt: SensitiveActionPrompt | null = sensitiveActionPrompt(label, success)) {
    if (!(await confirmSensitiveAction(prompt))) {
      return;
    }
    setActionBusy(label);
    setActionError("");
    try {
      await action();
      message.success(success);
      onChanged();
      await load();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setActionBusy("");
    }
  }

  function triggerDownloadURL(url: string, fileName: string) {
    const target = url.trim();
    if (!target) {
      throw new Error("下载地址为空");
    }
    const parsed = new URL(target, window.location.href);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      throw new Error("下载地址必须是 http(s) 地址");
    }
    const link = document.createElement("a");
    link.href = parsed.href;
    link.target = "_blank";
    link.rel = "noopener";
    link.download = fileName.trim() || "invoice.pdf";
    document.body.appendChild(link);
    link.click();
    link.remove();
  }

  function triggerFileDownload(content: BlobPart, fileName: string, contentType: string) {
    const blob = new Blob([content], { type: contentType || "application/octet-stream" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = fileName.trim() || "download.bin";
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  }

  function base64BlobPart(value: string) {
    const binary = window.atob(value.trim());
    const bytes = new Uint8Array(binary.length);
    for (let index = 0; index < binary.length; index += 1) {
      bytes[index] = binary.charCodeAt(index);
    }
    return bytes;
  }

  async function downloadInvoiceFile(id: number) {
    const download = await api.downloadInvoice(id);
    triggerDownloadURL(download.url, download.fileName);
  }

  async function downloadLicensePackageFile(id: number) {
    const download = await api.downloadLicensePackage(id);
    triggerFileDownload(JSON.stringify(download.package, null, 2), download.fileName, download.contentType || "application/json");
  }

  async function downloadUpdatePackageFile(id: number) {
    const download = await api.downloadUpdate(id);
    triggerFileDownload(
      base64BlobPart(download.artifactContentBase64),
      download.artifactFileName || download.fileName,
      download.artifactContentType || download.contentType || "application/octet-stream"
    );
  }

  function resetUpdatePackageForm() {
    setUpdatePackageForm({
      version: "",
      component: "server",
      channel: "stable",
      status: "available",
      packageType: "full",
      baseVersion: "",
      rollbackVersion: "",
      artifactFileName: "",
      artifactContentType: "application/octet-stream",
      artifactContentBase64: "",
      targetArtifactSha256: "",
      remark: ""
    });
  }

  async function handleUpdatePackageFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setActionError("");
    try {
      const payload = await browserFilePayload(file);
      setUpdatePackageForm((value) => ({
        ...value,
        artifactFileName: payload.fileName,
        artifactContentType: payload.fileType,
        artifactContentBase64: payload.base64,
        targetArtifactSha256: value.packageType === "delta" ? value.targetArtifactSha256 || payload.checksum : payload.checksum
      }));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "读取更新包文件失败");
    }
  }

  async function handlePublishUpdatePackage(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-update-publish", "更新包已发布", () => api.publishUpdate({
      version: updatePackageForm.version.trim(),
      component: updatePackageForm.component,
      channel: updatePackageForm.channel,
      status: updatePackageForm.status,
      packageType: updatePackageForm.packageType,
      baseVersion: updatePackageForm.baseVersion.trim(),
      rollbackVersion: updatePackageForm.rollbackVersion.trim(),
      artifactFileName: updatePackageForm.artifactFileName.trim(),
      artifactContentType: updatePackageForm.artifactContentType.trim(),
      artifactContentBase64: updatePackageForm.artifactContentBase64.trim(),
      targetArtifactSha256: updatePackageForm.targetArtifactSha256.trim(),
      remark: updatePackageForm.remark.trim()
    }));
  }

  async function handleCreateBackup() {
    await runBusinessAction("system-backup-create", "备份已创建", () => api.createBackup());
  }

  async function handleRunBackupDrill() {
    await runBusinessAction("system-backup-drill", "备份演练已完成", () => api.runBackupDrill());
  }

  async function handleEvaluateRules() {
    await runBusinessAction("system-rules-evaluate", "规则已重新评估", () => api.evaluateRules());
  }

  function resetGatewayRouteForm() {
    setGatewayRouteForm({
      id: "",
      name: "",
      pathPrefix: "",
      stableUpstream: "",
      canaryUpstream: "",
      canaryPercent: "0",
      readTimeoutSec: "120",
      status: "active"
    });
  }

  function startGatewayRouteEdit(item: GatewayRoute) {
    setGatewayRouteForm({
      id: String(item.id),
      name: item.name || "",
      pathPrefix: item.pathPrefix || "",
      stableUpstream: item.stableUpstream || "",
      canaryUpstream: item.canaryUpstream || "",
      canaryPercent: String(item.canaryPercent || 0),
      readTimeoutSec: String(item.readTimeoutSec || 120),
      status: item.status || "active"
    });
  }

  function startGatewayCanary(item: GatewayRoute) {
    setGatewayCanaryForm({
      routeId: String(item.id),
      canaryPercent: String(item.canaryPercent || 0),
      canaryUpstream: item.canaryUpstream || ""
    });
  }

  async function handleSaveGatewayRoute(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-gateway-route-${gatewayRouteForm.id || "new"}`, "网关路由已保存", () => api.saveGatewayRoute({
      id: fieldNumber(gatewayRouteForm.id),
      name: gatewayRouteForm.name.trim(),
      pathPrefix: gatewayRouteForm.pathPrefix.trim(),
      stableUpstream: gatewayRouteForm.stableUpstream.trim(),
      canaryUpstream: gatewayRouteForm.canaryUpstream.trim(),
      canaryPercent: fieldNumber(gatewayRouteForm.canaryPercent),
      readTimeoutSec: fieldNumber(gatewayRouteForm.readTimeoutSec, 120),
      status: gatewayRouteForm.status
    }));
  }

  async function handleSetGatewayCanary(event: FormEvent) {
    event.preventDefault();
    const routeId = fieldNumber(gatewayCanaryForm.routeId);
    if (!routeId) {
      setActionError("请选择网关路由");
      return;
    }
    await runBusinessAction(`system-gateway-canary-${routeId}`, "网关灰度已更新", () => api.setGatewayCanary(routeId, fieldNumber(gatewayCanaryForm.canaryPercent), gatewayCanaryForm.canaryUpstream.trim()));
  }

  async function handleDeleteGatewayRoute(item: GatewayRoute) {
    await runBusinessAction(`system-gateway-delete-${item.id}`, "网关路由已删除", () => api.deleteGatewayRoute(item.id));
  }

  async function handleReloadGateway() {
    await runBusinessAction("system-gateway-reload", "网关已重载", () => api.reloadGateway());
  }

  async function handleToggleModule(item: ModuleInfo) {
    await runBusinessAction(`system-module-${item.code}`, "模块状态已更新", () => api.setSystemModuleEnabled(item.code, !item.enabled));
  }

  function integrationListFromText(value: string) {
    return value.split(/\r?\n|,|，/).map((item) => item.trim()).filter(Boolean);
  }

  function resetSSOProviderForm() {
    setSsoProviderForm({
      id: "",
      name: "",
      code: "",
      issuer: "",
      clientId: "",
      clientSecret: "",
      authUrl: "",
      tokenUrl: "",
      userInfoUrl: "",
      jwksUrl: "",
      redirectUri: "",
      scopes: "openid\nprofile\nemail",
      usernameClaim: "preferred_username",
      displayNameClaim: "name",
      roleCode: availableRoles()[0]?.code || "customer",
      companyId: String(firstId(bootstrap?.companies)),
      siteId: String(defaultSiteId || 0),
      autoProvision: "true",
      status: "enabled"
    });
  }

  function startSSOProviderEdit(item: OIDCProvider) {
    setSsoProviderForm({
      id: String(item.id),
      name: item.name || "",
      code: item.code || "",
      issuer: item.issuer || "",
      clientId: item.clientId || "",
      clientSecret: "",
      authUrl: item.authUrl || "",
      tokenUrl: item.tokenUrl || "",
      userInfoUrl: item.userInfoUrl || "",
      jwksUrl: item.jwksUrl || "",
      redirectUri: item.redirectUri || "",
      scopes: list(item.scopes).join("\n"),
      usernameClaim: item.usernameClaim || "preferred_username",
      displayNameClaim: item.displayNameClaim || "name",
      roleCode: item.roleCode || availableRoles()[0]?.code || "customer",
      companyId: String(item.companyId || firstId(bootstrap?.companies)),
      siteId: String(item.siteId || 0),
      autoProvision: item.autoProvision ? "true" : "false",
      status: item.status || "enabled"
    });
  }

  async function handleSaveSSOProvider(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-sso-provider-${ssoProviderForm.id || "new"}`, "SSO 提供商已保存", async () => {
      await api.saveSSOProvider({
        id: fieldNumber(ssoProviderForm.id) || undefined,
        name: ssoProviderForm.name.trim(),
        code: ssoProviderForm.code.trim(),
        issuer: ssoProviderForm.issuer.trim(),
        clientId: ssoProviderForm.clientId.trim(),
        clientSecret: ssoProviderForm.clientSecret.trim() || undefined,
        authUrl: ssoProviderForm.authUrl.trim(),
        tokenUrl: ssoProviderForm.tokenUrl.trim(),
        userInfoUrl: ssoProviderForm.userInfoUrl.trim(),
        jwksUrl: ssoProviderForm.jwksUrl.trim(),
        redirectUri: ssoProviderForm.redirectUri.trim(),
        scopes: integrationListFromText(ssoProviderForm.scopes),
        usernameClaim: ssoProviderForm.usernameClaim.trim(),
        displayNameClaim: ssoProviderForm.displayNameClaim.trim(),
        roleCode: ssoProviderForm.roleCode,
        companyId: fieldNumber(ssoProviderForm.companyId),
        siteId: fieldNumber(ssoProviderForm.siteId),
        autoProvision: ssoProviderForm.autoProvision === "true",
        status: ssoProviderForm.status
      });
      resetSSOProviderForm();
    });
  }

  async function handleSSOProviderStatus(item: OIDCProvider) {
    const status = item.status === "enabled" ? "disabled" : "enabled";
    await runBusinessAction(`system-sso-provider-status-${item.id}`, "SSO 提供商状态已提交", () => api.setSSOProviderStatus(item.id, status));
  }

  async function handleDeleteSSOProvider(item: OIDCProvider) {
    await runBusinessAction(`system-sso-provider-delete-${item.id}`, "SSO 提供商已删除", () => api.deleteSSOProvider(item.id));
  }

  function resetSCIMProviderForm() {
    setScimProviderForm({
      id: "",
      name: "",
      code: "",
      bearerToken: "",
      companyId: String(firstId(bootstrap?.companies)),
      siteId: String(defaultSiteId || 0),
      defaultRoleCode: availableRoles()[0]?.code || "customer",
      status: "enabled"
    });
  }

  function startSCIMProviderEdit(item: SCIMProvider) {
    setScimProviderForm({
      id: String(item.id),
      name: item.name || "",
      code: item.code || "",
      bearerToken: "",
      companyId: String(item.companyId || firstId(bootstrap?.companies)),
      siteId: String(item.siteId || 0),
      defaultRoleCode: item.defaultRoleCode || availableRoles()[0]?.code || "customer",
      status: item.status || "enabled"
    });
  }

  async function handleSaveSCIMProvider(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-scim-provider-${scimProviderForm.id || "new"}`, "SCIM 提供商已保存", async () => {
      await api.saveSCIMProvider({
        id: fieldNumber(scimProviderForm.id) || undefined,
        name: scimProviderForm.name.trim(),
        code: scimProviderForm.code.trim(),
        bearerToken: scimProviderForm.bearerToken.trim() || undefined,
        companyId: fieldNumber(scimProviderForm.companyId),
        siteId: fieldNumber(scimProviderForm.siteId),
        defaultRoleCode: scimProviderForm.defaultRoleCode,
        status: scimProviderForm.status
      });
      resetSCIMProviderForm();
    });
  }

  async function handleSCIMProviderStatus(item: SCIMProvider) {
    const status = item.status === "enabled" ? "disabled" : "enabled";
    await runBusinessAction(`system-scim-provider-status-${item.id}`, "SCIM 提供商状态已提交", () => api.setSCIMProviderStatus(item.id, status));
  }

  async function handleDeleteSCIMProvider(item: SCIMProvider) {
    await runBusinessAction(`system-scim-provider-delete-${item.id}`, "SCIM 提供商已删除", () => api.deleteSCIMProvider(item.id));
  }

  function resetSecurityPolicyForm() {
    setSecurityPolicyForm({
      id: "",
      name: "",
      type: "",
      value: "",
      enabled: "true",
      remark: ""
    });
  }

  function startSecurityPolicyEdit(item: SecurityPolicy) {
    setSecurityPolicyForm({
      id: String(item.id),
      name: item.name || "",
      type: item.type || "",
      value: item.value || "",
      enabled: item.enabled ? "true" : "false",
      remark: item.remark || ""
    });
  }

  async function handleSaveSecurityPolicy(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-security-policy-${securityPolicyForm.id || "new"}`, "安全策略已保存", async () => {
      await api.saveSecurityPolicy({
        id: fieldNumber(securityPolicyForm.id) || undefined,
        name: securityPolicyForm.name.trim(),
        type: securityPolicyForm.type.trim(),
        value: securityPolicyForm.value.trim(),
        enabled: securityPolicyForm.enabled === "true",
        remark: securityPolicyForm.remark.trim()
      });
      resetSecurityPolicyForm();
    });
  }

  async function handleToggleSecurityPolicy(item: SecurityPolicy) {
    await runBusinessAction(`system-security-policy-toggle-${item.id}`, "安全策略状态已更新", () => api.toggleSecurityPolicy(item.id, !item.enabled));
  }

  async function handleDeleteSecurityPolicy(item: SecurityPolicy) {
    await runBusinessAction(`system-security-policy-delete-${item.id}`, "安全策略已删除", () => api.deleteSecurityPolicy(item.id));
  }

  function resetDeviceCredentialForm() {
    setDeviceCredentialForm({
      id: "",
      deviceNo: "",
      deviceKey: "",
      scopes: "location:report",
      status: "active"
    });
  }

  function startDeviceCredentialEdit(item: DeviceCredential) {
    setDeviceCredentialForm({
      id: String(item.id),
      deviceNo: item.deviceNo || "",
      deviceKey: "",
      scopes: list(item.scopes).join("\n"),
      status: item.status || "active"
    });
  }

  async function handleSaveDeviceCredential(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-device-credential-${deviceCredentialForm.id || "new"}`, "设备凭证已保存", async () => {
      await api.saveDeviceCredential({
        id: fieldNumber(deviceCredentialForm.id) || undefined,
        deviceNo: deviceCredentialForm.deviceNo.trim(),
        deviceKey: deviceCredentialForm.deviceKey.trim() || undefined,
        scopes: integrationListFromText(deviceCredentialForm.scopes),
        status: deviceCredentialForm.status
      });
      resetDeviceCredentialForm();
    });
  }

  async function handleDeviceCredentialStatus(item: DeviceCredential) {
    const status = item.status === "active" ? "disabled" : "active";
    await runBusinessAction(`system-device-credential-status-${item.id}`, "设备凭证状态已更新", () => api.setDeviceCredentialStatus(item.id, status));
  }

  async function handleDeleteDeviceCredential(item: DeviceCredential) {
    await runBusinessAction(`system-device-credential-delete-${item.id}`, "设备凭证已删除", () => api.deleteDeviceCredential(item.id));
  }

  function resetIntegrationEndpointForm() {
    setIntegrationEndpointForm({
      id: "",
      name: "",
      type: "collection_sms",
      protocol: "rest/http",
      url: "",
      status: "disabled"
    });
  }

  function startIntegrationEndpointEdit(item: IntegrationEndpoint) {
    setIntegrationEndpointForm({
      id: String(item.id),
      name: item.name || "",
      type: item.type || "collection_sms",
      protocol: item.protocol || "rest/http",
      url: item.url || "",
      status: item.status || "disabled"
    });
  }

  async function handleSaveIntegrationEndpoint(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-integration-endpoint-${integrationEndpointForm.id || "new"}`, "集成端点已保存", async () => {
      await api.saveIntegrationEndpoint({
        id: fieldNumber(integrationEndpointForm.id) || undefined,
        name: integrationEndpointForm.name.trim(),
        type: integrationEndpointForm.type.trim(),
        protocol: integrationEndpointForm.protocol.trim(),
        url: integrationEndpointForm.url.trim(),
        status: integrationEndpointForm.status
      });
      resetIntegrationEndpointForm();
    });
  }

  async function handleIntegrationEndpointStatus(item: IntegrationEndpoint) {
    const status = item.status === "disabled" ? "online" : "disabled";
    await runBusinessAction(`system-integration-endpoint-status-${item.id}`, "集成端点状态已更新", () => api.setIntegrationEndpointStatus(item.id, status));
  }

  async function handleDeleteIntegrationEndpoint(item: IntegrationEndpoint) {
    await runBusinessAction(`system-integration-endpoint-delete-${item.id}`, "集成端点已删除", () => api.deleteIntegrationEndpoint(item.id));
  }

  function resetRuleDefinitionForm() {
    setRuleDefinitionForm({
      id: "",
      code: "",
      name: "",
      category: "vehicle",
      metric: "speed",
      operator: ">",
      threshold: "80",
      level: "warning",
      enabled: "true",
      notifyRoles: "dispatcher",
      description: ""
    });
  }

  function startRuleDefinitionEdit(item: RuleDefinition) {
    setRuleDefinitionForm({
      id: String(item.id),
      code: item.code || "",
      name: item.name || "",
      category: item.category || "vehicle",
      metric: item.metric || "speed",
      operator: item.operator || ">",
      threshold: String(item.threshold ?? 0),
      level: item.level || "warning",
      enabled: item.enabled ? "true" : "false",
      notifyRoles: list(item.notifyRoles).join("\n"),
      description: item.description || ""
    });
  }

  async function handleSaveRuleDefinition(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-rule-definition-${ruleDefinitionForm.id || "new"}`, "自动化规则已保存", async () => {
      await api.saveRuleDefinition({
        id: fieldNumber(ruleDefinitionForm.id) || undefined,
        code: ruleDefinitionForm.code.trim(),
        name: ruleDefinitionForm.name.trim(),
        category: ruleDefinitionForm.category.trim(),
        metric: ruleDefinitionForm.metric.trim(),
        operator: ruleDefinitionForm.operator.trim(),
        threshold: fieldNumber(ruleDefinitionForm.threshold),
        level: ruleDefinitionForm.level,
        enabled: ruleDefinitionForm.enabled === "true",
        notifyRoles: integrationListFromText(ruleDefinitionForm.notifyRoles),
        description: ruleDefinitionForm.description.trim()
      });
      resetRuleDefinitionForm();
    });
  }

  async function handleRuleDefinitionStatus(item: RuleDefinition) {
    await runBusinessAction(`system-rule-definition-status-${item.id}`, "自动化规则状态已更新", () => api.setRuleDefinitionStatus(item.id, !item.enabled));
  }

  async function handleDeleteRuleDefinition(item: RuleDefinition) {
    await runBusinessAction(`system-rule-definition-delete-${item.id}`, "自动化规则已删除", () => api.deleteRuleDefinition(item.id));
  }

  function resetPluginInstallForm() {
    setPluginInstallForm({
      id: "",
      name: "",
      type: "integration",
      status: "installed",
      version: "1.0.0",
      checksum: "",
      signature: "",
      permissions: "",
      runtime: "node",
      entrypoint: "",
      sandboxRuntime: "node",
      sandboxTimeoutMs: "30000",
      sandboxNetwork: "false",
      sandboxFilesystem: "none",
      sandboxMaxMemoryMb: "128"
    });
  }

  async function handleInstallPlugin(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-plugin-install", "插件已安装", async () => {
      await api.installPlugin({
        id: pluginInstallForm.id.trim(),
        name: pluginInstallForm.name.trim(),
        type: pluginInstallForm.type.trim(),
        status: pluginInstallForm.status,
        version: pluginInstallForm.version.trim(),
        checksum: pluginInstallForm.checksum.trim(),
        signature: pluginInstallForm.signature.trim(),
        permissions: integrationListFromText(pluginInstallForm.permissions),
        runtime: pluginInstallForm.runtime.trim(),
        entrypoint: pluginInstallForm.entrypoint.trim(),
        sandbox: {
          runtime: pluginInstallForm.sandboxRuntime.trim(),
          timeoutMs: fieldNumber(pluginInstallForm.sandboxTimeoutMs, 30000),
          network: pluginInstallForm.sandboxNetwork === "true",
          filesystem: pluginInstallForm.sandboxFilesystem.trim(),
          maxMemoryMb: fieldNumber(pluginInstallForm.sandboxMaxMemoryMb, 128)
        }
      });
      resetPluginInstallForm();
    });
  }

  function startPluginRun(item: PluginInfo) {
    setPluginRunForm({
      pluginId: item.id,
      action: list(item.permissions)[0] || "",
      permission: list(item.permissions)[0] || "",
      input: "{}"
    });
  }

  async function handleRunPlugin(event: FormEvent) {
    event.preventDefault();
    if (!pluginRunForm.pluginId) {
      setActionError("请选择插件");
      return;
    }
    let input: Record<string, unknown>;
    try {
      input = pluginRunForm.input.trim() ? JSON.parse(pluginRunForm.input) as Record<string, unknown> : {};
    } catch {
      setActionError("插件输入必须是 JSON 对象");
      return;
    }
    await runBusinessAction(`system-plugin-run-${pluginRunForm.pluginId}`, "插件已运行", () => api.runPlugin(pluginRunForm.pluginId, {
      action: pluginRunForm.action.trim(),
      permission: pluginRunForm.permission.trim(),
      input
    }));
  }

  async function handleVerifyPlugin(item: PluginInfo) {
    await runBusinessAction(`system-plugin-verify-${item.id}`, "插件验签已通过", async () => {
      const result = await api.verifyPlugin(item.id);
      if (!result.valid) {
        throw new Error("插件验签未通过");
      }
    });
  }

  async function handlePluginStatus(item: PluginInfo) {
    const status = item.status === "enabled" ? "disabled" : "enabled";
    await runBusinessAction(`system-plugin-status-${item.id}`, status === "enabled" ? "插件已启用" : "插件已停用", () => api.setPluginStatus(item.id, status));
  }

  function licenseModulesFromText(value: string) {
    return value.split(/\r?\n|,|，/).map((item) => item.trim()).filter(Boolean);
  }

  function resetLicenseIssueForm() {
    setLicenseIssueForm({
      licenseId: "",
      customerName: bootstrap?.license?.customerName && bootstrap.license.lastVerificationState !== "missing" ? bootstrap.license.customerName : "",
      watermark: "",
      expiresAt: today,
      edition: "Enterprise",
      modules: list(bootstrap?.modules).filter((item) => item.enabled).map((item) => item.code).join("\n"),
      maxSites: String(Math.max(1, list(bootstrap?.sites).length)),
      maxVehicles: String(Math.max(1, list(bootstrap?.vehicles).length)),
      issuer: "CBMP License Center",
      privateKey: ""
    });
  }

  function startRenewLicensePackage(item: LicensePackage) {
    setLicenseRenewForm({
      packageId: String(item.id),
      licenseId: "",
      expiresAt: item.expiresAt || "",
      edition: item.edition || "",
      modules: list(item.modules).join("\n"),
      maxSites: item.maxSites ? String(item.maxSites) : "",
      maxVehicles: item.maxVehicles ? String(item.maxVehicles) : "",
      issuer: item.issuer || "",
      privateKey: ""
    });
  }

  function startRevokeLicensePackage(item: LicensePackage) {
    setLicenseRevokeForm({
      licenseId: item.licenseId,
      reason: ""
    });
  }

  async function handleImportLicensePackage(event: FormEvent) {
    event.preventDefault();
    let payload: Record<string, unknown>;
    try {
      payload = JSON.parse(licenseImportText) as Record<string, unknown>;
    } catch {
      setActionError("授权包 JSON 格式错误");
      return;
    }
    await runBusinessAction("system-license-import", "授权包已导入", async () => {
      await api.importLicensePackage(payload);
      setLicenseImportText("");
    });
  }

  async function handleIssueLicense(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("system-license-issue", "授权包已签发", () => api.issueLicense({
      licenseId: licenseIssueForm.licenseId.trim(),
      customerName: licenseIssueForm.customerName.trim(),
      watermark: licenseIssueForm.watermark.trim(),
      expiresAt: licenseIssueForm.expiresAt,
      edition: licenseIssueForm.edition.trim(),
      modules: licenseModulesFromText(licenseIssueForm.modules),
      maxSites: fieldNumber(licenseIssueForm.maxSites),
      maxVehicles: fieldNumber(licenseIssueForm.maxVehicles),
      issuer: licenseIssueForm.issuer.trim(),
      privateKey: licenseIssueForm.privateKey.trim()
    }));
  }

  async function handleRenewLicensePackage(event: FormEvent) {
    event.preventDefault();
    const packageId = fieldNumber(licenseRenewForm.packageId);
    if (!packageId) {
      setActionError("请选择要续期的授权包");
      return;
    }
    await runBusinessAction(`system-license-renew-${packageId}`, "授权包已续期", () => api.renewLicensePackage(packageId, {
      licenseId: licenseRenewForm.licenseId.trim(),
      expiresAt: licenseRenewForm.expiresAt,
      edition: licenseRenewForm.edition.trim(),
      modules: licenseModulesFromText(licenseRenewForm.modules),
      maxSites: fieldNumber(licenseRenewForm.maxSites),
      maxVehicles: fieldNumber(licenseRenewForm.maxVehicles),
      issuer: licenseRenewForm.issuer.trim(),
      privateKey: licenseRenewForm.privateKey.trim()
    }));
  }

  async function handleRevokeLicense(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`system-license-revoke-${licenseRevokeForm.licenseId}`, "授权已吊销", () => api.revokeLicense(licenseRevokeForm.licenseId, licenseRevokeForm.reason));
  }

  function normalizeApprovalKey(value: string | number | null | undefined) {
    return String(value || "").trim().toLowerCase().replace(/[\s-]+/g, "_");
  }

  function approvalsFor(resourceNames: string | string[], resourceId?: number | string | null, resourceNo?: string | null) {
    const resources = new Set(list(Array.isArray(resourceNames) ? resourceNames : [resourceNames]).map(normalizeApprovalKey));
    const targetId = Number(resourceId || 0);
    const targetNo = normalizeApprovalKey(resourceNo);
    return openApprovals
      .filter((item) => {
        const resourceMatched = resources.has(normalizeApprovalKey(item.resource));
        const idMatched = targetId > 0 && Number(item.resourceId || 0) === targetId;
        const noMatched = targetNo !== "" && normalizeApprovalKey(item.resourceNo) === targetNo;
        return resourceMatched && (idMatched || noMatched);
      })
      .sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));
  }

  function approvalFor(resourceNames: string | string[], resourceId?: number | string | null, resourceNo?: string | null) {
    return approvalsFor(resourceNames, resourceId, resourceNo)[0] || null;
  }

  function approvalStatus(task: ApprovalTask | null, baseStatus: ReactNode) {
    if (!task) return baseStatus;
    return (
      <span className="approval-inline-status">
        {baseStatus}
        <span className="approval-inline-badge">待审：{roleName(task.currentRole)}</span>
      </span>
    );
  }

  function approvalActionBlock(task: ApprovalTask | null) {
    if (!task) return null;
    return (
      <div className="workflow-approval-block">
        <div className="workflow-approval-head">
          <b>{task.title}</b>
          <StatusChip value={task.status} />
        </div>
        <span>{task.resourceNo} / {task.flowName}</span>
        <span>当前节点：第 {task.currentStep} 步 / {roleName(task.currentRole)}</span>
        {task.applicant ? <span>申请人：{task.applicant}</span> : null}
        {task.reason ? <span>原因：{task.reason}</span> : null}
        <Field label="审批意见">
          <TextInput value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
        </Field>
        <ActionGroup className="compact-actions">
          <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-approve-${task.id}`, "审批已通过", () => api.actApproval(task.id, "approve", approvalComment))}>通过</UiButton>
          <UiButton disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-reject-${task.id}`, "审批已驳回", () => api.actApproval(task.id, "reject", approvalComment))}>驳回</UiButton>
        </ActionGroup>
      </div>
    );
  }

  function workflowItemsFor(resourceNames: string | string[], resourceId?: number | string | null, resourceNo?: string | null) {
    return findBusinessWorkflowItems(workflowOverview(), resourceNames, resourceId, resourceNo);
  }

  function canActWorkflowTask(task: WorkflowTask | null) {
    if (!task || task.status !== "pending") return false;
    const roleCode = bootstrap?.user.roleCode || "";
    return task.roleCode === roleCode || hasPermission(currentPermissions, "*");
  }

  function workflowStatusFor(resourceNames: string | string[], resourceId: number | string | null | undefined, resourceNo: string | null | undefined, baseStatus: ReactNode) {
    return (
      <BusinessWorkflowStatus
        overview={workflowOverview()}
        resourceNames={resourceNames}
        resourceId={resourceId}
        resourceNo={resourceNo}
        baseStatus={baseStatus}
        roleLabel={roleName}
      />
    );
  }

  async function handleBusinessWorkflowTaskAction(task: WorkflowTask, action: "approve" | "reject") {
    await runBusinessAction(`business-workflow-${action}-${task.id}`, action === "approve" ? "工作流任务已通过" : "工作流任务已驳回", () => api.actWorkflowTask(task.id, action, approvalComment));
  }

  function workflowTimelineBlock(resourceNames: string | string[], resourceId?: number | string | null, resourceNo?: string | null, emptyText = "暂无工作流") {
    return (
      <BusinessWorkflowTimeline
        overview={workflowOverview()}
        resourceNames={resourceNames}
        resourceId={resourceId}
        resourceNo={resourceNo}
        emptyText={emptyText}
        roleLabel={roleName}
        comment={approvalComment}
        commentLabel="审批意见"
        busy={actionBusy !== ""}
        canActTask={canActWorkflowTask}
        onCommentChange={setApprovalComment}
        onActTask={handleBusinessWorkflowTaskAction}
      />
    );
  }

  function masterResource(kind: MasterKind) {
    return ({
      customer: "customers",
      project: "projects",
      product: "products",
      material: "materials",
      site: "sites",
	      plant: "plants",
	      driver: "drivers",
	      vehicle: "vehicles",
	      carrier: "carriers"
    } as Record<string, string>)[kind] || kind;
  }

  function vehicleDeviceFor(vehicleId: number | undefined) {
    if (!vehicleId) return undefined;
    return list(bootstrap?.vehicleDevices).find((item) => item.vehicleId === vehicleId);
  }

  function latestLocationFor(vehicleId: number | undefined) {
    if (!vehicleId) return undefined;
    return visibleLatestLocations.find((item) => item.vehicleId === vehicleId);
  }

  function vehicleDevicePayload(vehicleId: number) {
    return {
      vehicleId,
      deviceNo: masterForm.vehicleDeviceNo.trim(),
      protocol: "gps-forwarder",
      vendor: "GPS 转发器",
      status: "active"
    };
  }

  async function syncVehicleDevice(vehicleId: number) {
    const existing = vehicleDeviceFor(vehicleId);
    const payload = vehicleDevicePayload(vehicleId);
    if (!payload.deviceNo) {
      if (existing) {
        await api.deleteMasterResource<VehicleDevice>("vehicle-devices", existing.id);
      }
      return;
    }
    if (existing) {
      await api.updateMasterResource<VehicleDevice>("vehicle-devices", existing.id, payload);
      return;
    }
    await api.createVehicleDevice(payload);
  }

  function nextVehicleInternalNo() {
    const nextId = Math.max(0, ...list(bootstrap?.vehicles).map((item) => item.id || 0)) + 1;
    return `V${String(nextId).padStart(3, "0")}`;
  }

  function openBufferDialog(mode: "create" | "edit" | "transfer" | "adjust", buffer?: PlantBufferLocation, plant?: Plant) {
    const plants = list(bootstrap?.plants).filter((item) => matchesCurrentSite(item.siteId));
    const selectedPlant = plant || plants.find((item) => item.id === buffer?.plantId) || plants[0];
    const materialId = buffer?.materialId || firstId(bootstrap?.materials);
    const defaultBufferType = firstDictionaryCode("buffer_type", "aggregate_bin");
    const defaultQualityStatus = firstDictionaryCode("quality_status", "passed");
    const defaultResourceStatus = firstDictionaryCode("resource_status", "active");
    setBufferForm({
      bufferId: buffer ? String(buffer.id) : "",
      plantId: selectedPlant ? String(selectedPlant.id) : "",
      yardPileId: "",
      code: buffer?.code || "",
      name: buffer?.name || "",
      type: buffer?.type || defaultBufferType,
      materialId: materialId ? String(materialId) : "",
      capacity: buffer ? String(buffer.capacity || "") : "",
      unit: buffer?.unit || "t",
      warningQty: buffer ? String(buffer.warningQty || "") : "",
      transferQty: "",
      actualQty: buffer ? String(buffer.currentQty || 0) : "",
      moistureRate: buffer ? String(buffer.moistureRate || 0) : "",
      qualityStatus: buffer?.qualityStatus || defaultQualityStatus,
      status: buffer?.status || defaultResourceStatus,
      remark: ""
    });
    setBufferDialogMode(mode);
  }

  function plantBufferCodeDuplicated(code: string, exceptID = 0) {
    const normalized = code.trim().toLowerCase();
    if (!normalized) {
      return false;
    }
    const bufferSource = list(data.production?.plantBufferLocations);
    return bufferSource.some((item) => item.id !== exceptID && item.code.trim().toLowerCase() === normalized);
  }

  async function handleBufferSubmit(event: FormEvent) {
    event.preventDefault();
    if (!bufferDialogMode) return;
    if ((bufferDialogMode === "create" || bufferDialogMode === "edit") && plantBufferCodeDuplicated(bufferForm.code, fieldNumber(bufferForm.bufferId))) {
      setActionError("筒仓编码已存在。采集程序按编码匹配仓位，请换一个唯一编码。");
      return;
    }
    const bufferActionLabels = {
      create: "创建筒仓",
      edit: "更新筒仓",
      transfer: "筒仓转仓",
      adjust: "筒仓库存调整"
    };
    if (!(await confirmSensitiveAction(sensitiveActionPrompt(`buffer-${bufferDialogMode}`, bufferActionLabels[bufferDialogMode])))) {
      return;
    }
    setActionBusy(`buffer-${bufferDialogMode}`);
    setActionError("");
    try {
      if (bufferDialogMode === "create") {
        const plant = list(bootstrap?.plants).find((item) => item.id === fieldNumber(bufferForm.plantId));
        await api.createPlantBufferLocation({
          siteId: plant?.siteId || defaultSiteId,
          plantId: fieldNumber(bufferForm.plantId),
          code: bufferForm.code,
          name: bufferForm.name,
          type: bufferForm.type,
          materialId: fieldNumber(bufferForm.materialId),
          allowedMaterialIds: [fieldNumber(bufferForm.materialId)].filter(Boolean),
          capacity: fieldNumber(bufferForm.capacity),
          unit: bufferForm.unit,
          warningQty: fieldNumber(bufferForm.warningQty),
          qualityStatus: bufferForm.qualityStatus,
          status: bufferForm.status
        });
        message.success("筒仓已创建");
      } else if (bufferDialogMode === "edit") {
        const plant = list(bootstrap?.plants).find((item) => item.id === fieldNumber(bufferForm.plantId));
        await api.updateMasterResource<PlantBufferLocation>("plant-buffer-locations", fieldNumber(bufferForm.bufferId), {
          siteId: plant?.siteId || defaultSiteId,
          plantId: fieldNumber(bufferForm.plantId),
          code: bufferForm.code,
          name: bufferForm.name,
          type: bufferForm.type,
          materialId: fieldNumber(bufferForm.materialId),
          allowedMaterialIds: [fieldNumber(bufferForm.materialId)].filter(Boolean),
          capacity: fieldNumber(bufferForm.capacity),
          unit: bufferForm.unit,
          warningQty: fieldNumber(bufferForm.warningQty),
          qualityStatus: bufferForm.qualityStatus,
          status: bufferForm.status
        });
        message.success("筒仓已更新");
      } else if (bufferDialogMode === "transfer") {
        const yardPileId = fieldNumber(bufferForm.yardPileId);
        await api.createPlantBufferTransfer({
          bufferId: fieldNumber(bufferForm.bufferId),
          yardPileId: yardPileId || undefined,
          materialId: fieldNumber(bufferForm.materialId),
          quantity: fieldNumber(bufferForm.transferQty),
          unit: bufferForm.unit,
          remark: bufferForm.remark
        });
        message.success("筒仓上料已记录");
      } else {
	        await api.createPlantBufferAdjustment({
	          bufferId: fieldNumber(bufferForm.bufferId),
	          actualQty: fieldNumber(bufferForm.actualQty),
	          moistureRate: fieldNumber(bufferForm.moistureRate),
	          qualityStatus: bufferForm.qualityStatus,
	          status: bufferForm.status,
	          remark: bufferForm.remark
	        });
	        message.success("筒仓盘点已提交");
      }
      setBufferDialogMode(null);
      refreshData();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "筒仓操作失败");
    } finally {
      setActionBusy("");
    }
  }

  async function handleDeletePlantBuffer(buffer: PlantBufferLocation) {
    await runBusinessAction(
      `plant-buffer-delete-${buffer.id}`,
      "筒仓已删除",
      () => api.deleteMasterResource<PlantBufferLocation>("plant-buffer-locations", buffer.id)
    );
  }

  function openYardDialog(mode: "yard" | "yard-edit" | "pile" | "pile-edit" | "receipt" | "adjust", yard?: StockYard, pile?: StockYardPile) {
    const yards = list(data.procurement?.stockYards);
    const piles = list(data.procurement?.stockYardPiles);
    const visibleYards = yards.filter((item) => matchesCurrentSite(item.siteId));
    const visiblePiles = piles.filter((item) => matchesCurrentSite(item.siteId));
    const selectedYard = yard || visibleYards.find((item) => item.id === pile?.yardId) || visibleYards[0];
    const shouldUseExistingPile = mode === "pile-edit" || mode === "receipt" || mode === "adjust";
    const selectedPile = shouldUseExistingPile ? pile || visiblePiles[0] : undefined;
    const materialId = selectedPile?.materialId || firstId(bootstrap?.materials);
    const supplierId = selectedPile?.supplierId || recordId(supplierOptions()[0]);
    const isYardMode = mode === "yard" || mode === "yard-edit";
    const isYardEdit = mode === "yard-edit";
    const isPileEdit = mode === "pile-edit";
    const defaultYardType = firstDictionaryCode("yard_type", "aggregate_yard");
    const defaultQualityStatus = firstDictionaryCode("quality_status", "passed");
    const defaultResourceStatus = firstDictionaryCode("resource_status", "active");
    setYardForm({
      yardId: selectedYard ? String(selectedYard.id) : "",
      pileId: selectedPile ? String(selectedPile.id) : "",
      siteId: String(selectedYard?.siteId || defaultSiteId),
      code: isYardMode ? (isYardEdit ? selectedYard?.code || "" : "") : (isPileEdit ? selectedPile?.code || "" : ""),
      name: isYardMode ? (isYardEdit ? selectedYard?.name || "" : "") : (isPileEdit ? selectedPile?.name || "" : ""),
      type: selectedYard?.type || defaultYardType,
      area: selectedYard?.area || "",
      materialId: materialId ? String(materialId) : "",
      supplierId: supplierId ? String(supplierId) : "",
      batchNo: selectedPile?.batchNo || "",
      capacity: isYardMode ? (isYardEdit ? String(selectedYard?.capacity || "") : "") : (isPileEdit ? String(selectedPile?.capacity || "") : ""),
      currentQty: selectedPile ? String(selectedPile.currentQty || 0) : "",
      warningQty: selectedPile ? String(selectedPile.warningQty || "") : "",
      unit: selectedPile?.unit || selectedYard?.unit || "t",
      moistureRate: selectedPile ? String(selectedPile.moistureRate || 0) : "",
      qualityStatus: selectedPile?.qualityStatus || defaultQualityStatus,
      status: selectedPile?.status || selectedYard?.status || defaultResourceStatus,
      receiptQty: "",
      actualQty: selectedPile ? String(selectedPile.currentQty || 0) : "",
      remark: ""
    });
    setYardDialogMode(mode);
  }

  async function handleYardSubmit(event: FormEvent) {
    event.preventDefault();
    if (!yardDialogMode) return;
    const yardActionLabels = {
      yard: "创建堆场",
      "yard-edit": "更新堆场",
      pile: "创建堆位",
      "pile-edit": "更新堆位",
      receipt: "堆位入库",
      adjust: "堆位库存调整"
    };
    if (!(await confirmSensitiveAction(sensitiveActionPrompt(`yard-${yardDialogMode}`, yardActionLabels[yardDialogMode])))) {
      return;
    }
    setActionBusy(`yard-${yardDialogMode}`);
    setActionError("");
    try {
      if (yardDialogMode === "yard") {
        await api.createStockYard({
          siteId: fieldNumber(yardForm.siteId, defaultSiteId),
          code: yardForm.code,
          name: yardForm.name,
          type: yardForm.type,
          area: yardForm.area,
          capacity: fieldNumber(yardForm.capacity),
          unit: yardForm.unit,
          status: yardForm.status
        });
        message.success("堆场已创建");
      } else if (yardDialogMode === "yard-edit") {
        await api.updateMasterResource<StockYard>("stock-yards", fieldNumber(yardForm.yardId), {
          siteId: fieldNumber(yardForm.siteId, defaultSiteId),
          code: yardForm.code,
          name: yardForm.name,
          type: yardForm.type,
          area: yardForm.area,
          capacity: fieldNumber(yardForm.capacity),
          unit: yardForm.unit,
          status: yardForm.status
        });
        message.success("堆场已更新");
      } else if (yardDialogMode === "pile") {
        await api.createStockYardPile({
          siteId: fieldNumber(yardForm.siteId, defaultSiteId),
          yardId: fieldNumber(yardForm.yardId),
          code: yardForm.code,
          name: yardForm.name,
          materialId: fieldNumber(yardForm.materialId),
          supplierId: fieldNumber(yardForm.supplierId),
          batchNo: yardForm.batchNo,
          capacity: fieldNumber(yardForm.capacity),
          unit: yardForm.unit,
          warningQty: fieldNumber(yardForm.warningQty),
          moistureRate: fieldNumber(yardForm.moistureRate),
          qualityStatus: yardForm.qualityStatus,
          status: yardForm.status
        });
        message.success("堆位已创建");
      } else if (yardDialogMode === "pile-edit") {
        await api.updateMasterResource<StockYardPile>("stock-yard-piles", fieldNumber(yardForm.pileId), {
          siteId: fieldNumber(yardForm.siteId, defaultSiteId),
          yardId: fieldNumber(yardForm.yardId),
          code: yardForm.code,
          name: yardForm.name,
          materialId: fieldNumber(yardForm.materialId),
          supplierId: fieldNumber(yardForm.supplierId),
          batchNo: yardForm.batchNo,
          capacity: fieldNumber(yardForm.capacity),
          unit: yardForm.unit,
          warningQty: fieldNumber(yardForm.warningQty),
          moistureRate: fieldNumber(yardForm.moistureRate),
          qualityStatus: yardForm.qualityStatus,
          status: yardForm.status
        });
        message.success("堆位已更新");
      } else if (yardDialogMode === "receipt") {
        await api.createStockYardReceipt({
          pileId: fieldNumber(yardForm.pileId),
          materialId: fieldNumber(yardForm.materialId),
          supplierId: fieldNumber(yardForm.supplierId),
          batchNo: yardForm.batchNo,
          quantity: fieldNumber(yardForm.receiptQty),
          unit: yardForm.unit,
          moistureRate: fieldNumber(yardForm.moistureRate),
          qualityStatus: yardForm.qualityStatus,
          remark: yardForm.remark
        });
        message.success("堆场入场已记录");
      } else {
	        await api.createStockYardAdjustment({
	          pileId: fieldNumber(yardForm.pileId),
	          actualQty: fieldNumber(yardForm.actualQty),
	          moistureRate: fieldNumber(yardForm.moistureRate),
	          qualityStatus: yardForm.qualityStatus,
	          status: yardForm.status,
	          remark: yardForm.remark
	        });
	        message.success("堆位盘点已提交");
      }
      setYardDialogMode(null);
      refreshData();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "堆场操作失败");
    } finally {
      setActionBusy("");
    }
  }

  async function handleDeleteStockYard(yard: StockYard) {
    await runBusinessAction(
      `stock-yard-delete-${yard.id}`,
      "堆场已删除",
      () => api.deleteMasterResource<StockYard>("stock-yards", yard.id)
    );
  }

  async function handleDeleteStockYardPile(pile: StockYardPile) {
    await runBusinessAction(
      `stock-yard-pile-delete-${pile.id}`,
      "堆位已删除",
      () => api.deleteMasterResource<StockYardPile>("stock-yard-piles", pile.id)
    );
  }

  async function handleStockYardStatus(yard: StockYard) {
    const nextStatus = yard.status === "disabled" ? "active" : "disabled";
    await runBusinessAction(
      `stock-yard-status-${yard.id}`,
      nextStatus === "active" ? "堆场已启用" : "堆场已停用",
      () => api.updateMasterResource<StockYard>("stock-yards", yard.id, {
        siteId: yard.siteId,
        code: yard.code,
        name: yard.name,
        type: yard.type,
        area: yard.area,
        capacity: yard.capacity,
        unit: yard.unit,
        status: nextStatus,
        gatewayDeviceNo: yard.gatewayDeviceNo
      })
    );
  }

  async function handleStockYardPileStatus(pile: StockYardPile) {
    const nextStatus = pile.status === "disabled" ? "active" : "disabled";
    await runBusinessAction(
      `stock-yard-pile-status-${pile.id}`,
      nextStatus === "active" ? "堆位已启用" : "堆位已停用",
      () => api.updateMasterResource<StockYardPile>("stock-yard-piles", pile.id, {
        siteId: pile.siteId,
        yardId: pile.yardId,
        code: pile.code,
        name: pile.name,
        materialId: pile.materialId,
        supplierId: pile.supplierId,
        batchNo: pile.batchNo,
        capacity: pile.capacity,
        unit: pile.unit,
        warningQty: pile.warningQty,
        moistureRate: pile.moistureRate,
        qualityStatus: pile.qualityStatus,
        status: nextStatus,
        gatewayDeviceNo: pile.gatewayDeviceNo,
        gatewayChannel: pile.gatewayChannel,
        gatewayProtocol: pile.gatewayProtocol
      })
    );
  }

  async function handleMasterSubmit(kind: MasterKind, event: FormEvent) {
    event.preventDefault();
    const payloads: Record<string, Record<string, unknown>> = {
      customer: {
        name: masterForm.customerName,
        contact: masterForm.customerContact,
        phone: masterForm.customerPhone,
        creditLimit: 100000,
        paymentTerm: 30
      },
      project: {
        customerId: fieldNumber(masterForm.projectCustomerId),
        name: masterForm.projectName,
        address: masterForm.projectAddress,
        longitude: fieldNumber(masterForm.projectLongitude),
        latitude: fieldNumber(masterForm.projectLatitude)
      },
      product: {
        name: masterForm.productName,
        spec: masterForm.productSpec,
        unit: "t",
        basePrice: fieldNumber(masterForm.productPrice, 380),
        costPrice: Math.round(fieldNumber(masterForm.productPrice, 380) * 0.72),
        requiresMix: true
      },
      material: {
        name: masterForm.materialName,
        spec: masterForm.materialSpec,
        unit: "t",
        safeStock: fieldNumber(masterForm.materialSafeStock, 100)
      },
      site: {
        companyId: fieldNumber(masterForm.siteCompanyId, firstId(bootstrap?.companies)),
        name: masterForm.siteName,
        code: masterForm.siteCode,
        address: masterForm.siteAddress,
        longitude: fieldNumber(masterForm.siteLongitude),
        latitude: fieldNumber(masterForm.siteLatitude),
        fenceRadius: normalizedFenceRadius(masterForm.siteFenceRadius)
      },
      plant: {
        siteId: fieldNumber(masterForm.plantSiteId, defaultSiteId),
        name: masterForm.plantName,
        code: masterForm.plantCode,
        capacity: masterForm.plantCapacity,
        status: masterForm.plantStatus
      },
      driver: {
        name: masterForm.driverName,
        phone: masterForm.driverPhone
      },
      carrier: {
        name: masterForm.carrierName,
        contact: masterForm.carrierContact,
        phone: masterForm.carrierPhone,
        settleMode: masterForm.carrierSettleMode || "monthly",
        status: masterForm.carrierStatus || firstDictionaryCode("resource_status", "active")
      },
      vehicle: {
        internalNo: masterForm.vehicleInternalNo,
        plateNo: masterForm.vehiclePlate,
        vehicleType: masterForm.vehicleType,
        capacity: masterForm.vehicleCapacity,
        driverId: fieldNumber(masterForm.vehicleDriverId),
        siteId: fieldNumber(masterForm.vehicleSiteId),
        carrier: bootstrap?.carriers[0]?.name || ""
      }
    };
    const payload = payloads[kind];
    const editing = editingMaster?.kind === kind ? editingMaster : null;
    const criticalMasterNames: Partial<Record<MasterKind, string>> = {
      customer: "客户资料",
      driver: "司机资料",
      vehicle: "车辆资料和车载设备绑定",
      carrier: "承运商资料"
    };
    const masterPrompt = criticalMasterNames[kind]
      ? sensitiveActionPrompt(`master-${kind}`, `${editing ? "更新" : "创建"}${criticalMasterNames[kind]}`)
      : sensitiveActionPrompt(`master-${kind}`, editing ? "基础资料已更新" : "基础资料已保存");
    await runBusinessAction(`master-${kind}`, editing ? "基础资料已更新" : "基础资料已保存", async () => {
      if (editing) {
        await api.updateMasterResource(masterResource(kind), editing.id, payload);
        if (kind === "vehicle") {
          await syncVehicleDevice(editing.id);
        }
      } else {
        const createActions: Record<MasterKind, () => Promise<unknown>> = {
          customer: () => api.createCustomer(payload as never),
          project: () => api.createProject(payload as never),
          product: () => api.createProduct(payload as never),
          material: () => api.createMaterial(payload as never),
          site: () => api.createSite(payload as never),
          plant: () => api.createPlant(payload as never),
          driver: () => api.createDriver(payload as never),
          carrier: () => api.createCarrier(payload as never),
          vehicle: async () => {
            const vehicle = await api.createVehicle(payload as never);
            await syncVehicleDevice(vehicle.id);
            return vehicle;
          }
        };
        await createActions[kind]();
      }
      setEditingMaster(null);
      setMasterDialogKind(null);
    }, masterPrompt);
  }

  function resetCustomerDomainForms(customerId = firstId(bootstrap?.customers)) {
    const customerIdText = customerId ? String(customerId) : "";
    setCustomerContactForm({ id: "", customerId: customerIdText, name: "", phone: "", role: "业务联系人", isDefault: "true" });
    setCustomerProfileForm({ customerId: customerIdText, grade: "A", riskLevel: "low", creditScore: "90", tags: "" });
    setCustomerBlacklistForm({ customerId: customerIdText, reason: "", scope: "sales_order", severity: "high", blockOrders: "true", blockDispatch: "false" });
    setCustomerComplaintForm({ customerId: customerIdText, projectId: "", title: "", content: "", level: "medium", owner: "", slaHours: "24", resolution: "" });
  }

  function startEditCustomerContact(item: CustomerContact) {
    setCustomerContactForm({
      id: String(item.id),
      customerId: String(item.customerId),
      name: item.name || "",
      phone: item.phone || "",
      role: item.role || "业务联系人",
      isDefault: item.isDefault ? "true" : "false"
    });
  }

  async function handleSaveCustomerContact(event: FormEvent) {
    event.preventDefault();
    const contactId = fieldNumber(customerContactForm.id);
    await runBusinessAction(`customer-contact-${contactId || "create"}`, contactId ? "客户联系人已更新" : "客户联系人已创建", async () => {
      const payload = {
        customerId: fieldNumber(customerContactForm.customerId),
        name: customerContactForm.name.trim(),
        phone: customerContactForm.phone.trim(),
        role: customerContactForm.role.trim(),
        isDefault: customerContactForm.isDefault === "true",
        status: "active"
      };
      if (contactId) {
        await api.updateMasterResource<CustomerContact>("customer-contacts", contactId, payload);
      } else {
        await api.createCustomerContact(payload);
      }
      resetCustomerDomainForms(fieldNumber(customerContactForm.customerId));
    });
  }

  async function handleSetDefaultCustomerContact(item: CustomerContact) {
    await runBusinessAction(`customer-contact-default-${item.id}`, "默认联系人已更新", () => api.setDefaultCustomerContact(item.id));
  }

  async function handleDeleteCustomerContact(item: CustomerContact) {
    await runBusinessAction(`customer-contact-delete-${item.id}`, "客户联系人已删除", () => api.deleteMasterResource<CustomerContact>("customer-contacts", item.id));
  }

  async function handleCreateCustomerProfile(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("customer-profile-create", "客户档案已保存", async () => {
      await api.createCustomerProfile({
        customerId: fieldNumber(customerProfileForm.customerId),
        grade: customerProfileForm.grade.trim(),
        riskLevel: customerProfileForm.riskLevel.trim(),
        creditScore: fieldNumber(customerProfileForm.creditScore),
        tags: customerProfileForm.tags.split(/\r?\n|,|，/).map((item) => item.trim()).filter(Boolean),
        status: "active"
      });
      resetCustomerDomainForms(fieldNumber(customerProfileForm.customerId));
    });
  }

  async function handleEvaluateCustomerProfiles() {
    await runBusinessAction("customer-profile-evaluate", "客户风险档案已重算", () => api.evaluateCustomerProfiles());
  }

  async function handleCreateCustomerBlacklist(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("customer-blacklist-create", "客户黑名单已提交", async () => {
      await api.createCustomerBlacklist({
        customerId: fieldNumber(customerBlacklistForm.customerId),
        reason: customerBlacklistForm.reason.trim(),
        scope: customerBlacklistForm.scope,
        severity: customerBlacklistForm.severity,
        blockOrders: customerBlacklistForm.blockOrders === "true",
        blockDispatch: customerBlacklistForm.blockDispatch === "true"
      });
      resetCustomerDomainForms(fieldNumber(customerBlacklistForm.customerId));
    });
  }

  async function handleReleaseCustomerBlacklist(item: CustomerBlacklist) {
    await runBusinessAction(`customer-blacklist-release-${item.id}`, "客户黑名单解除已提交", () => api.releaseCustomerBlacklist(item.id));
  }

  async function handleCreateCustomerComplaint(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("customer-complaint-create", "客户投诉已创建", async () => {
      await api.createCustomerComplaint({
        customerId: fieldNumber(customerComplaintForm.customerId),
        projectId: fieldNumber(customerComplaintForm.projectId),
        title: customerComplaintForm.title.trim(),
        content: customerComplaintForm.content.trim(),
        level: customerComplaintForm.level,
        owner: customerComplaintForm.owner.trim(),
        slaHours: fieldNumber(customerComplaintForm.slaHours, 24)
      });
      resetCustomerDomainForms(fieldNumber(customerComplaintForm.customerId));
    });
  }

  async function handleCloseCustomerComplaint(event: FormEvent<HTMLFormElement>, item: CustomerComplaint) {
    event.preventDefault();
    const resolution = String(new FormData(event.currentTarget).get("resolution") || "").trim();
    if (!resolution) {
      setActionError("请输入处理结果");
      return;
    }
    await runBusinessAction(`customer-complaint-close-${item.id}`, "客户投诉已关闭", () => api.closeCustomerComplaint(item.id, resolution));
  }

  function resetPricingForms(productId = firstId(bootstrap?.products)) {
    const customerId = firstId(bootstrap?.customers);
    const projectId = list(bootstrap?.projects).find((item) => item.customerId === customerId)?.id || firstId(bootstrap?.projects);
    const product = list(bootstrap?.products).find((item) => item.id === productId) || list(bootstrap?.products)[0];
    const taxRateId = firstId(bootstrap?.taxRates);
    setTaxRateForm({ id: "", name: "", rate: "0.06", scope: "sales", status: "active" });
    setPricePolicyForm({
      id: "",
      customerId: customerId ? String(customerId) : "",
      projectId: projectId ? String(projectId) : "",
      productId: product ? String(product.id) : "",
      customerGrade: "A",
      region: "",
      minQuantity: "0",
      maxQuantity: "0",
      floorPrice: product ? String(Math.max(0, product.basePrice - 20)) : "",
      salePrice: product ? String(product.basePrice) : "",
      promotionName: "",
      promotionType: "",
      promotionValue: "0",
      priority: "10",
      taxRateId: taxRateId ? String(taxRateId) : "",
      effectiveFrom: today,
      effectiveTo: "",
      status: "active"
    });
    setPricingEvalForm({
      customerId: customerId ? String(customerId) : "",
      projectId: projectId ? String(projectId) : "",
      productId: product ? String(product.id) : "",
      planTime: today,
      planQuantity: "",
      unitPrice: product ? String(product.basePrice) : ""
    });
  }

  function startEditTaxRate(item: TaxRate) {
    setTaxRateForm({
      id: String(item.id),
      name: item.name,
      rate: String(item.rate),
      scope: item.scope || "sales",
      status: item.status || "active"
    });
  }

  function startEditPricePolicy(item: PricePolicy) {
    setPricePolicyForm({
      id: String(item.id),
      customerId: item.customerId ? String(item.customerId) : "",
      projectId: item.projectId ? String(item.projectId) : "",
      productId: item.productId ? String(item.productId) : "",
      customerGrade: item.customerGrade || "",
      region: item.region || "",
      minQuantity: String(item.minQuantity || 0),
      maxQuantity: String(item.maxQuantity || 0),
      floorPrice: String(item.floorPrice || 0),
      salePrice: String(item.salePrice || 0),
      promotionName: item.promotionName || "",
      promotionType: item.promotionType || "",
      promotionValue: String(item.promotionValue || 0),
      priority: String(item.priority || 10),
      taxRateId: item.taxRateId ? String(item.taxRateId) : "",
      effectiveFrom: item.effectiveFrom || today,
      effectiveTo: item.effectiveTo || "",
      status: item.status || "active"
    });
  }

  async function handleCreateTaxRate(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`master-tax-rate-${taxRateForm.id || "new"}`, "税率已保存", async () => {
      const payload = {
        name: taxRateForm.name.trim(),
        rate: fieldNumber(taxRateForm.rate),
        scope: taxRateForm.scope.trim(),
        status: taxRateForm.status
      };
      const id = fieldNumber(taxRateForm.id);
      if (id) {
        await api.updateMasterResource<TaxRate>("tax-rates", id, payload);
      } else {
        await api.createTaxRate(payload);
      }
      setTaxRateForm({ id: "", name: "", rate: "0.06", scope: "sales", status: "active" });
    });
  }

  async function handleCreatePricePolicy(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction(`master-price-policy-${pricePolicyForm.id || "new"}`, "价格政策已保存", async () => {
      const payload = {
        customerId: fieldNumber(pricePolicyForm.customerId),
        projectId: fieldNumber(pricePolicyForm.projectId),
        productId: fieldNumber(pricePolicyForm.productId),
        customerGrade: pricePolicyForm.customerGrade.trim(),
        region: pricePolicyForm.region.trim(),
        minQuantity: fieldNumber(pricePolicyForm.minQuantity),
        maxQuantity: fieldNumber(pricePolicyForm.maxQuantity),
        floorPrice: fieldNumber(pricePolicyForm.floorPrice),
        salePrice: fieldNumber(pricePolicyForm.salePrice),
        promotionName: pricePolicyForm.promotionName.trim(),
        promotionType: pricePolicyForm.promotionType.trim(),
        promotionValue: fieldNumber(pricePolicyForm.promotionValue),
        priority: fieldNumber(pricePolicyForm.priority, 10),
        taxRateId: fieldNumber(pricePolicyForm.taxRateId),
        effectiveFrom: pricePolicyForm.effectiveFrom,
        effectiveTo: pricePolicyForm.effectiveTo,
        status: pricePolicyForm.status
      };
      const id = fieldNumber(pricePolicyForm.id);
      if (id) {
        await api.updateMasterResource<PricePolicy>("price-policies", id, payload);
      } else {
        await api.createPricePolicy(payload);
      }
      resetPricingForms(fieldNumber(pricePolicyForm.productId));
    });
  }

  async function handleDeleteTaxRate(item: TaxRate) {
    await runBusinessAction(`master-tax-rate-delete-${item.id}`, "税率已删除", () => api.deleteMasterResource<TaxRate>("tax-rates", item.id));
  }

  async function handleDeletePricePolicy(item: PricePolicy) {
    await runBusinessAction(`master-price-policy-delete-${item.id}`, "价格政策已删除", () => api.deleteMasterResource<PricePolicy>("price-policies", item.id));
  }

  async function handleEvaluatePricing(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("master-pricing-evaluate", "报价已计算", async () => {
      const quote = await api.evaluatePricing({
        customerId: fieldNumber(pricingEvalForm.customerId),
        projectId: fieldNumber(pricingEvalForm.projectId),
        productId: fieldNumber(pricingEvalForm.productId),
        planTime: pricingEvalForm.planTime,
        planQuantity: fieldNumber(pricingEvalForm.planQuantity),
        unitPrice: fieldNumber(pricingEvalForm.unitPrice)
      });
      setPricingQuote(quote);
    }, null);
  }

  async function handleExportMasterData(resource = masterBulkForm.resource) {
    await runBusinessAction(`master-export-${resource}`, "主数据已导出", async () => {
      const result = await api.exportMasterData(resource);
      setMasterExportResult(result);
      triggerFileDownload(JSON.stringify(result, null, 2), `${resource}-export.json`, "application/json");
    }, null);
  }

  async function handleImportMasterData(event: FormEvent) {
    event.preventDefault();
    let rows: Record<string, unknown>[];
    try {
      const parsed = JSON.parse(masterBulkForm.rowsJson);
      if (!Array.isArray(parsed)) {
        throw new Error("导入数据必须是数组");
      }
      rows = parsed as Record<string, unknown>[];
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "导入 JSON 解析失败");
      return;
    }
    await runBusinessAction("master-import", "主数据已导入", async () => {
      const result = await api.importMasterData({ resource: masterBulkForm.resource, mode: masterBulkForm.mode, rows });
      setMasterImportResult(result);
    });
  }

	  function resetMasterForm(kind: MasterKind) {
	    const defaultResourceStatus = firstDictionaryCode("resource_status", "active");
	    setMasterForm((form) => {
      switch (kind) {
        case "customer":
          return { ...form, customerName: "", customerContact: "", customerPhone: "" };
        case "project":
          return { ...form, projectCustomerId: String(firstId(bootstrap?.customers)), projectName: "", projectAddress: "", projectLongitude: "", projectLatitude: "" };
        case "product":
          return { ...form, productName: "", productSpec: "", productPrice: "" };
        case "material":
          return { ...form, materialName: "", materialSpec: "", materialSafeStock: "" };
        case "site":
          return { ...form, siteName: "", siteCode: "", siteAddress: "", siteCompanyId: String(firstId(bootstrap?.companies)), siteLongitude: "", siteLatitude: "", siteFenceRadius: "" };
        case "plant":
          return { ...form, plantSiteId: String(defaultSiteId), plantName: "", plantCode: "", plantCapacity: "", plantStatus: firstDictionaryCode("plant_status", "running") };
        case "driver":
          return { ...form, driverName: "", driverPhone: "" };
        case "carrier":
          return { ...form, carrierName: "", carrierContact: "", carrierPhone: "", carrierSettleMode: "monthly", carrierStatus: defaultResourceStatus };
	        case "vehicle":
          return {
            ...form,
            vehicleInternalNo: nextVehicleInternalNo(),
            vehiclePlate: "",
            vehicleType: firstDictionaryCode("vehicle_type", "搅拌车"),
            vehicleCapacity: "",
            vehicleDriverId: String(firstId(bootstrap?.drivers)),
            vehicleSiteId: String(defaultSiteId),
            vehicleDeviceNo: ""
          };
      }
    });
  }

  function openMasterCreateDialog(kind: MasterKind) {
    setEditingMaster(null);
    setProjectLocationDialogOpen(false);
    setSiteFenceDialogOpen(false);
    setProjectAddressLookupState("");
    setSiteAddressLookupState("");
    resetMasterForm(kind);
    setMasterDialogKind(kind);
  }

  function clearMasterEdit() {
    setProjectLocationDialogOpen(false);
    setSiteFenceDialogOpen(false);
    setProjectAddressLookupState("");
    setSiteAddressLookupState("");
    setEditingMaster(null);
    setMasterDialogKind(null);
  }

  function startMasterEdit(kind: MasterKind, item: MasterRecord) {
    setProjectLocationDialogOpen(false);
    setSiteFenceDialogOpen(false);
    setProjectAddressLookupState("");
    setSiteAddressLookupState("");
    setEditingMaster({ kind, id: item.id });
    switch (kind) {
      case "customer": {
        const value = item as Customer;
        setMasterForm((form) => ({
          ...form,
          customerName: value.name,
          customerContact: value.contact || "",
          customerPhone: value.phone || ""
        }));
        break;
      }
      case "project": {
        const value = item as Project;
        setMasterForm((form) => ({
          ...form,
          projectCustomerId: String(value.customerId),
          projectName: value.name,
          projectAddress: value.address || "",
          projectLongitude: String(value.longitude || ""),
          projectLatitude: String(value.latitude || "")
        }));
        break;
      }
      case "product": {
        const value = item as Product;
        setMasterForm((form) => ({
          ...form,
          productName: value.name,
          productSpec: value.spec || "",
          productPrice: String(value.basePrice || "")
        }));
        break;
      }
      case "material": {
        const value = item as Material;
        setMasterForm((form) => ({
          ...form,
          materialName: value.name,
          materialSpec: value.spec || "",
          materialSafeStock: String(value.safeStock || "")
        }));
        break;
      }
      case "site": {
        const value = item as Site;
        const fence = activeSiteFence(value.id);
        setMasterForm((form) => ({
          ...form,
          siteName: value.name,
          siteCode: value.code,
          siteAddress: value.address || "",
          siteCompanyId: String(value.companyId || firstId(bootstrap?.companies)),
          siteLongitude: String(value.longitude || fence?.longitude || ""),
          siteLatitude: String(value.latitude || fence?.latitude || ""),
          siteFenceRadius: String(fence?.radius || value.fenceRadius || 300)
        }));
        break;
      }
      case "plant": {
        const value = item as Plant;
        setMasterForm((form) => ({
          ...form,
          plantSiteId: String(value.siteId || defaultSiteId),
          plantName: value.name,
          plantCode: value.code,
          plantCapacity: value.capacity || "",
          plantStatus: value.status || firstDictionaryCode("plant_status", "running")
        }));
        break;
      }
      case "driver": {
        const value = item as Driver;
        setMasterForm((form) => ({
          ...form,
          driverName: value.name,
          driverPhone: value.phone || ""
        }));
        break;
      }
      case "carrier": {
        const value = item as Carrier;
        setMasterForm((form) => ({
          ...form,
          carrierName: value.name,
          carrierContact: value.contact || "",
          carrierPhone: value.phone || "",
          carrierSettleMode: value.settleMode || "monthly",
          carrierStatus: value.status || firstDictionaryCode("resource_status", "active")
        }));
        break;
      }
	      case "vehicle": {
        const value = item as Vehicle;
        const device = vehicleDeviceFor(value.id);
        setMasterForm((form) => ({
          ...form,
          vehicleInternalNo: value.internalNo || "",
          vehiclePlate: value.plateNo,
          vehicleType: value.vehicleType || "",
          vehicleCapacity: value.capacity || "",
          vehicleDriverId: String(value.driverId || ""),
          vehicleSiteId: String(value.siteId || ""),
          vehicleDeviceNo: device?.deviceNo || ""
        }));
        break;
      }
    }
    setMasterDialogKind(kind);
  }

  async function handleDeleteMaster(kind: MasterKind, id: number) {
    await runBusinessAction(`master-delete-${kind}-${id}`, "基础资料已删除", () => api.deleteMasterResource(masterResource(kind), id));
  }

  function masterFormButton(kind: MasterKind, label: string, disabled = false) {
    const editing = editingMaster?.kind === kind;
    return (
      <FormActions spanAll>
        <UiButton variant="primary" type="submit" icon={editing ? <CheckCircle2 size={15} /> : <Plus size={15} />} disabled={actionBusy !== "" || disabled}>
          {editing ? "保存修改" : label}
        </UiButton>
        <UiButton disabled={actionBusy !== ""} onClick={clearMasterEdit}>取消</UiButton>
      </FormActions>
    );
  }

  function masterRowActions(kind: MasterKind, id: number, item: MasterRecord) {
    return (
      <ActionGroup>
        <UiButton disabled={actionBusy !== ""} onClick={() => startMasterEdit(kind, item)}>编辑</UiButton>
        <UiButton variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeleteMaster(kind, id)}>删除</UiButton>
      </ActionGroup>
    );
  }

  function openProjectLocationDialog() {
    setProjectLocationDialogOpen(true);
    window.setTimeout(() => {
      projectLocationPickerRef.current?.querySelector<HTMLElement>(".site-fence-map")?.focus({ preventScroll: true });
    }, 120);
  }

  function openSiteFenceDialog() {
    setSiteFenceDialogOpen(true);
    window.setTimeout(() => {
      siteFencePickerRef.current?.querySelector<HTMLElement>(".site-fence-map")?.focus({ preventScroll: true });
    }, 120);
  }

  async function fillProjectAddressFromLocation(longitude: number, latitude: number) {
    if (!isValidCoordinate(longitude, latitude) || mapConfig?.offline) {
      setProjectAddressLookupState("failed");
      return;
    }
    reverseGeocodeAbortRef.current?.abort();
    const requestId = reverseGeocodeRequestRef.current + 1;
    reverseGeocodeRequestRef.current = requestId;
    const controller = new AbortController();
    reverseGeocodeAbortRef.current = controller;
    setProjectAddressLookupState("loading");
    try {
      const address = await reverseGeocodeAddress(longitude, latitude, controller.signal);
      if (reverseGeocodeRequestRef.current !== requestId) return;
      if (address) {
        setMasterForm((form) => ({ ...form, projectAddress: address }));
        setProjectAddressLookupState("success");
      } else {
        setProjectAddressLookupState("failed");
      }
    } catch (err) {
      if (controller.signal.aborted) return;
      if (reverseGeocodeRequestRef.current === requestId) {
        setProjectAddressLookupState("failed");
      }
    } finally {
      if (reverseGeocodeRequestRef.current === requestId) {
        reverseGeocodeAbortRef.current = null;
      }
    }
  }

  async function fillSiteAddressFromLocation(longitude: number, latitude: number) {
    if (!isValidCoordinate(longitude, latitude) || mapConfig?.offline) {
      setSiteAddressLookupState("failed");
      return;
    }
    reverseGeocodeAbortRef.current?.abort();
    const requestId = reverseGeocodeRequestRef.current + 1;
    reverseGeocodeRequestRef.current = requestId;
    const controller = new AbortController();
    reverseGeocodeAbortRef.current = controller;
    setSiteAddressLookupState("loading");
    try {
      const address = await reverseGeocodeAddress(longitude, latitude, controller.signal);
      if (reverseGeocodeRequestRef.current !== requestId) return;
      if (address) {
        setMasterForm((form) => ({ ...form, siteAddress: address }));
        setSiteAddressLookupState("success");
      } else {
        setSiteAddressLookupState("failed");
      }
    } catch (err) {
      if (controller.signal.aborted) return;
      if (reverseGeocodeRequestRef.current === requestId) {
        setSiteAddressLookupState("failed");
      }
    } finally {
      if (reverseGeocodeRequestRef.current === requestId) {
        reverseGeocodeAbortRef.current = null;
      }
    }
  }

  async function handleCreateOrder(event: FormEvent) {
    event.preventDefault();
    const product = bootstrap?.products.find((item) => item.id === fieldNumber(orderForm.productId));
    const quantity = fieldNumber(orderForm.planQuantity, 30);
    const unitPrice = fieldNumber(orderForm.unitPrice, product?.basePrice || 380);
    await runBusinessAction("order-create", "销售订单已提交", async () => {
      await api.createOrder({
        customerId: fieldNumber(orderForm.customerId),
        projectId: fieldNumber(orderForm.projectId),
        productId: fieldNumber(orderForm.productId),
        siteId: fieldNumber(orderForm.siteId),
        planQuantity: quantity,
        unit: product?.unit || "t",
        unitPrice,
        planTime: orderForm.planTime,
        contact: orderForm.contact,
        phone: orderForm.phone,
        lines: [{
          productId: fieldNumber(orderForm.productId),
          productLine: product?.line || "asphalt",
          quantity,
          unit: product?.unit || "t",
          unitPrice
        }]
      });
      setOrderDialogOpen(false);
    });
  }

  async function handleCreateContract(event: FormEvent) {
    event.preventDefault();
    const product = bootstrap?.products.find((item) => item.id === fieldNumber(contractForm.productId));
    const quantity = fieldNumber(contractForm.quantity, 1000);
    const unitPrice = fieldNumber(contractForm.unitPrice, product?.basePrice || 380);
    await runBusinessAction("contract-create", "合同草稿已创建", () => api.createContract({
      customerId: fieldNumber(contractForm.customerId),
      projectId: fieldNumber(contractForm.projectId),
      name: contractForm.name,
      validFrom: contractForm.validFrom,
      validTo: contractForm.validTo,
      creditPolicy: "按合同和授信执行",
      changeReason: contractForm.reason,
      totalAmount: quantity * unitPrice,
      items: [{
        productId: fieldNumber(contractForm.productId),
        quantity,
        unit: product?.unit || "t",
        unitPrice
      }]
    }));
  }

  async function handleSubmitContract() {
    const id = fieldNumber(contractForm.contractId);
    await runBusinessAction(`contract-submit-${id}`, "合同已提交审批", () => api.submitContract(id, contractForm.reason));
  }

  async function handleReviseContract() {
    const id = fieldNumber(contractForm.contractId);
    const product = bootstrap?.products.find((item) => item.id === fieldNumber(contractForm.productId));
    const quantity = fieldNumber(contractForm.quantity, 1000);
    const unitPrice = fieldNumber(contractForm.unitPrice, product?.basePrice || 380);
    await runBusinessAction(`contract-revise-${id}`, "合同新版本已生成", () => api.reviseContract(id, {
      name: contractForm.name,
      validFrom: contractForm.validFrom,
      validTo: contractForm.validTo,
      changeReason: contractForm.reason,
      totalAmount: quantity * unitPrice,
      items: [{
        productId: fieldNumber(contractForm.productId),
        quantity,
        unit: product?.unit || "t",
        unitPrice
      }]
    }));
  }

  async function handleContractAttachmentFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setActionError("");
    try {
      const payload = await browserFilePayload(file);
      setContractForm((value) => ({
        ...value,
        attachmentName: payload.fileName,
        attachmentFileType: payload.fileType,
        attachmentUrl: payload.url,
        attachmentChecksum: payload.checksum
      }));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "读取合同附件失败");
    }
  }

  async function handleCreateContractAttachment() {
    const id = fieldNumber(contractForm.contractId);
    await runBusinessAction(`contract-attachment-${id}`, "合同附件已归档", async () => {
      const created = await api.createContractAttachment(id, {
        fileName: contractForm.attachmentName,
        fileType: contractForm.attachmentFileType,
        url: contractForm.attachmentUrl,
        checksum: contractForm.attachmentChecksum
      });
      setContractAttachmentCache((value) => {
        const current = value[id] || [];
        return { ...value, [id]: [...current.filter((item) => item.id !== created.id), created] };
      });
      setContractForm((value) => ({
        ...value,
        attachmentName: "",
        attachmentFileType: "contract_pdf",
        attachmentUrl: "",
        attachmentChecksum: ""
      }));
    });
  }

  async function handleCreateProductionPlan(event: FormEvent) {
    event.preventDefault();
    const orderId = fieldNumber(productionForm.orderId);
    const quantity = fieldNumber(productionForm.planQuantity);
    let createdPlan: ProductionPlan | null = null;
    await runBusinessAction("production-plan-create", "生产计划已生成", async () => {
      const plan = await api.createProductionPlan({
        orderId,
        plantId: fieldNumber(productionForm.plantId),
        planQuantity: quantity,
        planDate: productionForm.planDate,
        shift: productionForm.shift
      });
      createdPlan = plan;
      setProductionForm((value) => ({
        ...value,
        planId: String(plan.id),
        plantId: String(plan.plantId || value.plantId),
        adjustPlanQuantity: String(plan.planQuantity),
        taskQty: String(plan.remainingQty || plan.planQuantity),
        planQuantity: ""
      }));
    });
    if (createdPlan) {
      openProductionDialog("tasks", createdPlan);
    }
  }

  async function handleUpdateProductionPlan(event: FormEvent) {
    event.preventDefault();
    const planId = fieldNumber(productionForm.planId);
    await runBusinessAction(`production-plan-update-${planId}`, "生产计划已调整", () => api.updateProductionPlan(planId, {
      plantId: fieldNumber(productionForm.plantId),
      planQuantity: fieldNumber(productionForm.adjustPlanQuantity),
      planDate: productionForm.planDate,
      shift: productionForm.shift
    }).then(() => setProductionDialogMode(null)));
  }

  async function handleAutoProductionTasks() {
    const planId = fieldNumber(productionForm.planId);
    const plan = list(data.production?.plans).find((item) => item.id === planId);
    if (!productionPlanCanIssueTask(plan)) return;
    const taskQty = productionPlanTaskRemaining(plan);
    let createdTask: ProductionTask | undefined;
    await runBusinessAction(`production-task-auto-${planId}`, "生产任务已自动下达", async () => {
      const tasks = await api.autoCreateProductionTasks(planId, { taskQty });
      createdTask = tasks[0];
      setProductionForm((value) => ({ ...value, taskId: tasks[0] ? String(tasks[0].id) : value.taskId, taskQty: "" }));
    });
    if (createdTask) {
      openProductionDialog("batch", undefined, createdTask);
    } else {
      setProductionDialogMode(null);
    }
  }

  async function handleCreateProductionTask(event: FormEvent) {
    event.preventDefault();
    const planId = fieldNumber(productionForm.planId);
    const plan = list(data.production?.plans).find((item) => item.id === planId);
    const taskQty = fieldNumber(productionForm.taskQty);
    if (!productionPlanCanIssueTask(plan) || taskQty <= 0 || taskQty > productionPlanTaskRemaining(plan)) return;
    let createdTask: ProductionTask | null = null;
    await runBusinessAction(`production-task-create-${planId}`, "生产任务已下达", async () => {
      const task = await api.createProductionTask(planId, { planQty: taskQty });
      createdTask = task;
      setProductionForm((value) => ({ ...value, taskId: String(task.id), batchQty: String(productionTaskRemaining(task)), taskQty: "" }));
    });
    if (createdTask) {
      openProductionDialog("batch", undefined, createdTask);
    }
  }

  async function handleCreateProductionBatch(event?: FormEvent) {
    event?.preventDefault();
    const taskId = fieldNumber(productionForm.taskId);
    const planId = fieldNumber(productionForm.planId);
    const quantity = fieldNumber(productionForm.batchQty);
    const activeBatchTasks = sortProductionTasksForAction(list(data.production?.tasks)
      .filter((item) => matchesCurrentSite(item.siteId))
      .filter((item) => item.status !== "cancelled" && item.status !== "completed")
      .filter((item) => productionTaskRemaining(item) > 0));
    const scopedBatchTasks = planId ? activeBatchTasks.filter((item) => item.planId === planId) : activeBatchTasks;
    const task = (scopedBatchTasks.length ? scopedBatchTasks : activeBatchTasks).find((item) => item.id === taskId)
      || scopedBatchTasks[0]
      || activeBatchTasks[0];
    const plan = task ? list(data.production?.plans).find((item) => item.id === task.planId) : undefined;
    const taskRemaining = productionTaskRemaining(task);
    if (!task || taskRemaining <= 0 || quantity <= 0 || quantity > taskRemaining) return;
    const planWillComplete = !!plan && quantity > 0 && quantity >= plan.remainingQty;
    let createdBatch: ProductionBatch | null = null;
    await runBusinessAction(`production-batch-create-${task.id}`, "生产批次已登记", async () => {
      const batch = await api.createProductionBatch(task.id, {
        quantity,
        qualityStatus: productionForm.batchQuality,
        status: productionForm.batchQuality === "passed" ? "released" : "produced"
      });
      createdBatch = batch;
    });
    const batch = createdBatch as ProductionBatch | null;
    if (batch && planWillComplete && plan && batch.qualityStatus === "passed") {
      const producedQty = Math.min(plan.planQuantity, plan.producedQty + batch.quantity);
      const remainingQty = Math.max(0, plan.remainingQty - batch.quantity);
      const reportBatches = [
        ...list(data.production?.batches).filter((item) => item.planId === plan.id),
        batch
      ];
      openProductionDialog("report", {
        ...plan,
        producedQty,
        remainingQty,
        progress: plan.planQuantity > 0 ? Math.round((producedQty / plan.planQuantity) * 100) : plan.progress,
        status: remainingQty <= 0 ? "completed" : plan.status
      }, undefined, reportBatches);
    } else if (batch) {
      setProductionDialogMode(null);
    }
  }

  function startQualityInspection(batch: ProductionBatch) {
    setQualityInspectionForm({
      batchId: String(batch.id),
      inspector: "",
      slump: "",
      temperature: "",
      remark: ""
    });
  }

  async function handleCreateQualityInspection(event: FormEvent) {
    event.preventDefault();
    const batchId = fieldNumber(qualityInspectionForm.batchId);
    if (!batchId) return;
    await runBusinessAction(`quality-inspection-create-${batchId}`, "生产质检单已创建", () => api.createQualityInspection({
      batchId,
      inspector: qualityInspectionForm.inspector.trim(),
      slump: qualityInspectionForm.slump.trim(),
      temperature: fieldNumber(qualityInspectionForm.temperature),
      remark: qualityInspectionForm.remark.trim()
    }));
  }

  function startQualitySampleTest(item: QualitySample) {
    setQualitySampleForm({
      sampleId: String(item.id),
      strength: item.strength ? String(item.strength) : "",
      result: item.result === "failed" ? "failed" : "passed",
      testedAt: item.testedAt || "",
      remark: item.remark || ""
    });
  }

  async function handleTestQualitySample(event: FormEvent) {
    event.preventDefault();
    const sampleId = fieldNumber(qualitySampleForm.sampleId);
    if (!sampleId) return;
    await runBusinessAction(`quality-sample-test-${sampleId}`, "试样检测已登记", () => api.testQualitySample(sampleId, {
      strength: fieldNumber(qualitySampleForm.strength),
      result: qualitySampleForm.result,
      testedAt: qualitySampleForm.testedAt.trim(),
      remark: qualitySampleForm.remark.trim()
    }));
  }

  async function handleCancelProductionPlan() {
    const planId = fieldNumber(productionForm.planId);
    await runBusinessAction(`production-plan-cancel-${planId}`, "生产计划已取消", () => api.cancelProductionPlan(planId).then(() => setProductionDialogMode(null)));
  }

  async function handleGenerateProductionReport() {
    const selectedPlan = productionReportPlan || list(data.production?.plans).filter((item) => matchesCurrentSite(item.siteId)).find((item) => String(item.id) === productionForm.planId);
    const plan = selectedPlan;
    if (!plan || !productionPlanNeedsReport(plan)) return;
    await runBusinessAction(`production-report-${plan.siteId}-${plan.planDate}`, "生产日报已生成", () => api.generateProductionReport({
      siteId: plan.siteId,
      reportDate: plan.planDate
    }).then(() => {
      setProductionReportPlan(null);
      setProductionReportBatches(null);
      setProductionDialogMode(null);
    }));
  }

  function openProductionDialog(mode: ProductionDialogMode, plan?: ProductionPlan, task?: ProductionTask, reportBatches?: ProductionBatch[]) {
    const plans = list(data.production?.plans).filter((item) => matchesCurrentSite(item.siteId));
    const tasks = list(data.production?.tasks).filter((item) => matchesCurrentSite(item.siteId));
    const openOrders = scopedOrders.filter((item) => productionOrderRemaining(item, plans) > 0 && ["approved", "scheduled", "dispatching"].includes(item.status));
    const selectableTasks = tasks
      .filter((item) => item.status !== "completed" && item.status !== "cancelled")
      .filter((item) => productionTaskRemaining(item) > 0);
    const selectedPlanFromTask = task ? plans.find((item) => item.id === task.planId) : undefined;
    const selectedPlanCandidate = plan || selectedPlanFromTask || plans.find((item) => String(item.id) === productionForm.planId) || plans[0];
    const selectedPlan = mode === "create-plan" ? undefined
      : mode === "report" ? selectedPlanCandidate
      : mode === "tasks" && !productionPlanCanIssueTask(selectedPlanCandidate) ? plans.find(productionPlanCanIssueTask)
      : selectedPlanCandidate;
    const batchTaskCandidates = mode === "batch" && selectedPlan
      ? selectableTasks.filter((item) => item.planId === selectedPlan.id)
      : selectableTasks;
    const selectedTask = mode === "create-plan" ? undefined : task
      || batchTaskCandidates.find((item) => String(item.id) === productionForm.taskId)
      || batchTaskCandidates[0]
      || (mode === "batch" ? undefined : tasks.find((item) => String(item.id) === productionForm.taskId) || tasks[0]);
    const selectedOrder = openOrders.find((item) => String(item.id) === productionForm.orderId) || openOrders[0];
    const createPlanDate = productionOrderPlanDate(selectedOrder);
    const selectedPlant = mode === "create-plan"
      ? preferredProductionPlantForOrder(selectedOrder, createPlanDate)
      : productionTaskPlant(selectedTask) || productionPlanPlant(selectedPlan) || firstProductionPlant(selectedPlan?.siteId || selectedOrder?.siteId || selectedSiteId || undefined);
    const defaultShift = firstDictionaryCode("shift_type", "早班");
    const qualityOptions = dictionaryOptions("quality_status");
    const defaultQualityStatus = qualityOptions.some((item) => item.code === "passed") ? "passed" : qualityOptions[0]?.code || "passed";
    setProductionForm((value) => ({
      ...value,
      orderId: selectedOrder ? String(selectedOrder.id) : value.orderId,
      planId: selectedPlan ? String(selectedPlan.id) : value.planId,
      taskId: selectedTask ? String(selectedTask.id) : value.taskId,
      plantId: selectedPlant ? String(selectedPlant.id) : value.plantId,
      planDate: mode === "create-plan" ? createPlanDate : selectedPlan?.planDate || value.planDate || today,
      shift: mode === "create-plan" ? defaultShift : selectedPlan?.shift || value.shift || defaultShift,
      planQuantity: selectedOrder ? String(productionOrderRemaining(selectedOrder, plans)) : value.planQuantity,
      adjustPlanQuantity: selectedPlan ? String(selectedPlan.planQuantity) : value.adjustPlanQuantity,
      taskQty: selectedPlan ? String(productionPlanTaskRemaining(selectedPlan)) : value.taskQty,
      batchQty: selectedTask ? String(productionTaskRemaining(selectedTask)) : value.batchQty,
      batchQuality: mode === "batch" ? defaultQualityStatus : value.batchQuality || defaultQualityStatus
    }));
    setProductionReportPlan(mode === "report" ? selectedPlan || null : null);
    setProductionReportBatches(mode === "report" ? reportBatches || (selectedPlan ? list(data.production?.batches).filter((item) => item.planId === selectedPlan.id) : null) : null);
    setProductionDialogMode(mode);
  }

  function openOrderDialog() {
    const product = bootstrap?.products[0];
    setOrderForm((form) => ({
      ...form,
      customerId: String(firstId(bootstrap?.customers)),
      projectId: String(firstId(bootstrap?.projects)),
      productId: String(firstId(bootstrap?.products)),
      siteId: String(defaultSiteId),
      planQuantity: form.planQuantity || "",
      unitPrice: String(product?.basePrice || form.unitPrice || ""),
      planTime: form.planTime || today
    }));
    setOrderDialogOpen(true);
  }

  async function handleCreateRawReceipt(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("procurement-receipt", "原料入库已生成", () => api.createRawMaterialReceipt({
      purchaseOrderId: fieldNumber(procurementForm.purchaseOrderId),
      supplierId: fieldNumber(procurementForm.supplierId),
      siteId: fieldNumber(procurementForm.siteId),
      materialId: fieldNumber(procurementForm.materialId),
      plateNo: procurementForm.plateNo,
      grossWeight: fieldNumber(procurementForm.grossWeight),
      tareWeight: fieldNumber(procurementForm.tareWeight)
    }));
  }

  function startRawMaterialInspection(receipt: NonNullable<QualityOverview["receipts"]>[number]) {
    setRawInspectionForm({
      receiptId: String(receipt.id),
      inspector: "",
      sampleNo: "",
      remark: ""
    });
  }

  async function handleCreateRawMaterialInspection(event: FormEvent) {
    event.preventDefault();
    const receiptId = fieldNumber(rawInspectionForm.receiptId);
    if (!receiptId) return;
    await runBusinessAction(`quality-raw-inspection-create-${receiptId}`, "原材质检单已创建", () => api.createRawMaterialInspection({
      receiptId,
      inspector: rawInspectionForm.inspector.trim(),
      sampleNo: rawInspectionForm.sampleNo.trim(),
      remark: rawInspectionForm.remark.trim()
    }));
  }

  function resetStocktakeForm(item?: InventoryItem) {
    setStocktakeForm({
      siteId: String(item?.siteId || selectedSiteId || defaultSiteId || firstId(bootstrap?.sites)),
      materialId: String(item?.materialId || firstId(bootstrap?.materials)),
      actualQty: item ? String(item.quantity) : "",
      unit: item?.unit || "t",
      remark: ""
    });
  }

  async function handleCreateInventoryStocktake(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("inventory-stocktake-create", "库存盘点已提交复核", async () => {
      await api.createInventoryStocktake({
        siteId: fieldNumber(stocktakeForm.siteId),
        materialId: fieldNumber(stocktakeForm.materialId),
        actualQty: fieldNumber(stocktakeForm.actualQty),
        unit: stocktakeForm.unit.trim() || "t",
        remark: stocktakeForm.remark.trim()
      });
      resetStocktakeForm();
    });
  }

  async function handleReviewInventoryStocktake(item: InventoryStocktake) {
    await runBusinessAction(`inventory-stocktake-review-${item.id}`, "库存盘点已复核", () => api.reviewInventoryStocktake(item.id));
  }

  function startRawMaterialInspectionReview(item: RawMaterialInspection) {
    setRawInspectionReviewForm({
      inspectionId: String(item.id),
      moisture: item.moisture ? String(item.moisture) : "",
      mudContent: item.mudContent ? String(item.mudContent) : "",
      fineness: item.fineness || "",
      result: item.result === "failed" ? "failed" : "passed",
      remark: item.remark || ""
    });
  }

  async function handleReviewRawMaterialInspection(event: FormEvent) {
    event.preventDefault();
    const inspectionId = fieldNumber(rawInspectionReviewForm.inspectionId);
    if (!inspectionId) return;
    await runBusinessAction(`quality-raw-inspection-review-${inspectionId}`, "原材质检已复核", () => api.reviewRawMaterialInspection(inspectionId, {
      moisture: fieldNumber(rawInspectionReviewForm.moisture),
      mudContent: fieldNumber(rawInspectionReviewForm.mudContent),
      fineness: rawInspectionReviewForm.fineness.trim(),
      result: rawInspectionReviewForm.result,
      remark: rawInspectionReviewForm.remark.trim()
    }));
  }

  function resetTicketForm(mode = ticketForm.mode) {
    const dispatch = scopedDispatchOrders.find((item) => !["void", "cancelled"].includes(item.status)) || scopedDispatchOrders[0];
    const transfer = list(data.procurement?.transfers).find((item) => item.status === "completed");
    setTicketForm({
      mode,
      dispatchId: dispatch ? String(dispatch.id) : "",
      transferId: transfer ? String(transfer.id) : "",
      relatedTicketId: "",
      siteId: String(selectedSiteId || defaultSiteId || firstId(bootstrap?.sites)),
      materialId: String(firstId(bootstrap?.materials)),
      plateNo: "",
      grossWeight: "",
      tareWeight: "",
      unit: "t",
      remark: ""
    });
  }

  async function handleCreateWeighbridgeTicket(event: FormEvent) {
    event.preventDefault();
    const payload: Partial<ScaleTicket> = {
      dispatchId: fieldNumber(ticketForm.dispatchId),
      transferId: fieldNumber(ticketForm.transferId),
      relatedTicketId: fieldNumber(ticketForm.relatedTicketId),
      siteId: fieldNumber(ticketForm.siteId),
      materialId: fieldNumber(ticketForm.materialId),
      plateNo: ticketForm.plateNo.trim(),
      grossWeight: fieldNumber(ticketForm.grossWeight),
      tareWeight: fieldNumber(ticketForm.tareWeight),
      unit: ticketForm.unit.trim() || "t",
      remark: ticketForm.remark.trim()
    };
    await runBusinessAction(`weighbridge-ticket-create-${ticketForm.mode}`, "过磅记录已创建", async () => {
      if (ticketForm.mode === "inventory_transfer") {
        await api.createTransferTicket({ ...payload, transferId: fieldNumber(ticketForm.transferId) });
      } else if (ticketForm.mode === "product_return") {
        await api.createReturnTicket({ ...payload, dispatchId: fieldNumber(ticketForm.dispatchId) });
      } else if (ticketForm.mode === "waste_out") {
        await api.createWasteTicket({ ...payload, siteId: fieldNumber(ticketForm.siteId), materialId: fieldNumber(ticketForm.materialId) });
      } else {
        await api.createTicket(payload);
      }
      resetTicketForm(ticketForm.mode);
    });
  }

  async function handleTicketVoidReview(id: number, approved: boolean) {
    await runBusinessAction(`ticket-void-review-${id}-${approved ? "approve" : "reject"}`, approved ? "过磅记录作废已通过" : "过磅记录作废已驳回", () => api.approveTicketVoid(id, approved));
  }

  async function handleCreateReceipt(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-receipt", "收款已登记", () => api.createReceipt({
      receivableId: fieldNumber(financeForm.receivableId),
      amount: fieldNumber(financeForm.receiptAmount),
      method: "bank"
    }));
  }

  async function handleCreatePaymentPlan(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-plan", "付款计划已创建", () => api.createPaymentPlan({
      receivableId: fieldNumber(financeForm.receivableId),
      amount: fieldNumber(financeForm.planAmount),
      dueDate: financeForm.planDueDate,
      method: "bank"
    }));
  }

  async function handleCreateInvoice(event: FormEvent) {
    event.preventDefault();
    const defaultInvoiceType = dictionaryOptions("invoice_type")[0]?.code || "blue_vat_special";
    await runBusinessAction("finance-invoice-create", "发票已创建", () => api.createInvoice(fieldNumber(financeForm.statementId), financeForm.invoiceCategory || defaultInvoiceType));
  }

  async function handleSubmitSelectedInvoice() {
    const id = fieldNumber(financeForm.invoiceId);
    await runBusinessAction(`finance-invoice-submit-${id}`, "发票已提交税务", () => api.submitTaxInvoice(id));
  }

  async function handleDownloadSelectedInvoice() {
    const id = fieldNumber(financeForm.invoiceId);
    await runBusinessAction(`finance-invoice-download-${id}`, "发票文件已打开下载", () => downloadInvoiceFile(id));
  }

  async function handleCreateRedLetterInfo(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-red-letter-create", "红字信息表已申请", () => api.createRedLetterInfo(fieldNumber(financeForm.invoiceId), financeForm.redReason));
  }

  async function handleApproveSelectedRedLetter() {
    const id = fieldNumber(financeForm.redLetterInfoId);
    await runBusinessAction(`finance-red-letter-approve-${id}`, "红字信息表已审批", () => api.approveRedLetterInfo(id));
  }

  async function handleRedOffsetSelectedInvoice() {
    const id = fieldNumber(financeForm.invoiceId);
    await runBusinessAction(`finance-red-offset-${id}`, "红字发票已生成", () => api.redOffsetInvoice(id, financeForm.redReason, fieldNumber(financeForm.redLetterInfoId)));
  }

  async function handleGenerateCollections() {
    await runBusinessAction("finance-collections-generate", "催收任务已生成", () => api.generateCollectionTasks());
  }

  async function handleCreateCollectionTemplate(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-collection-template-create", "催收模板已保存", () => api.createCollectionTemplate({
      name: collectionTemplateForm.name.trim(),
      level: collectionTemplateForm.level,
      channel: collectionTemplateForm.channel,
      content: collectionTemplateForm.content.trim(),
      enabled: collectionTemplateForm.enabled === "true"
    }));
  }

  async function handleSendSelectedCollection() {
    const id = fieldNumber(financeForm.collectionTaskId);
    await runBusinessAction(`finance-collection-send-${id}`, "催收短信已发送", () => api.sendCollectionTask(id, fieldNumber(financeForm.collectionTemplateId), ""));
  }

  async function handleCloseSelectedCollection() {
    const id = fieldNumber(financeForm.collectionTaskId);
    await runBusinessAction(`finance-collection-close-${id}`, "催收任务已关闭", () => api.handleCollectionTask(id, financeForm.collectionRemark));
  }

  async function handleCreateSupplierStatement(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-supplier-statement-create", "供应商对账单已生成", () => api.createSupplierStatement({
      supplierId: fieldNumber(financeForm.supplierId),
      period: today,
      amount: 0
    }));
  }

  async function handleApproveSelectedSupplierStatement() {
    const id = fieldNumber(financeForm.supplierStatementId);
    await runBusinessAction(`finance-supplier-statement-approve-${id}`, "供应商对账单已确认", () => api.approveSupplierStatement(id));
  }

  async function handleCreateSupplierPayment(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("finance-payment-create", "供应商付款已登记", () => api.createPayment({
      payableId: fieldNumber(financeForm.payableId),
      amount: fieldNumber(financeForm.paymentAmount),
      method: financeForm.paymentMethod
    }));
  }

  async function handleCreateDispatchSchedule(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("dispatch-schedule-create", "车辆排班已生成", () => api.createDispatchSchedule({
      siteId: fieldNumber(dispatchScheduleForm.siteId),
      vehicleId: fieldNumber(dispatchScheduleForm.vehicleId),
      driverId: fieldNumber(dispatchScheduleForm.driverId),
      carrierId: fieldNumber(dispatchScheduleForm.carrierId),
      shiftDate: dispatchScheduleForm.shiftDate || today,
      shift: dispatchScheduleForm.shift.trim(),
      capacityQty: fieldNumber(dispatchScheduleForm.capacityQty),
      status: dispatchScheduleForm.status
    }));
  }

  async function handleGenerateCarrierSettlement(event: FormEvent) {
    event.preventDefault();
    const carrierId = fieldNumber(carrierSettlementForm.carrierId);
    const ratePerTrip = fieldNumber(carrierSettlementForm.ratePerTrip);
    const ratePerUnit = fieldNumber(carrierSettlementForm.ratePerUnit);
    await runBusinessAction("carrier-settlement-generate", "承运结算单已生成", () => api.generateCarrierSettlement({
      carrierId: carrierId || undefined,
      period: carrierSettlementForm.period || today.slice(0, 7),
      ratePerTrip: ratePerTrip || undefined,
      ratePerUnit: ratePerUnit || undefined
    }));
  }

  async function handleQuickDispatch(event: FormEvent) {
    event.preventDefault();
    const order = list(data.dispatch?.siteProgress).find((item) => item.orderId === selectedOrderId) || list(data.dispatch?.siteProgress)[0];
    const vehicle = list(data.dispatch?.availableVehicles).find((item) => item.vehicleId === selectedVehicleId) || list(data.dispatch?.availableVehicles)[0];
    if (!order || !vehicle) {
      setDispatchActionError("请选择供货订单和可用车辆");
      return;
    }
    const planQuantity = Number(dispatchQty || Math.min(36, order.remainingQty || order.planQuantity || 0));
    if (!Number.isFinite(planQuantity) || planQuantity <= 0) {
      setDispatchActionError("派车方量必须大于 0");
      return;
    }
    if (!(await confirmSensitiveAction(sensitiveActionPrompt("dispatch-create", "创建派车单")))) {
      return;
    }
    setDispatchSubmitting(true);
    setDispatchActionError("");
    try {
      await api.createDispatch({
        orderId: order.orderId,
        vehicleId: vehicle.vehicleId,
        driverId: vehicle.driverId,
        planQuantity
      });
      setDispatchQty("");
      setDispatchDialogOpen(false);
      onChanged();
      await load();
    } catch (err) {
      setDispatchActionError(err instanceof Error ? err.message : "派车失败");
    } finally {
      setDispatchSubmitting(false);
    }
  }

  async function handleAdvanceDispatch(item: DispatchCenterQueueItem) {
    if (!(await confirmSensitiveAction(sensitiveActionPrompt(`dispatch-advance-${item.dispatchId}`, "推进调度状态")))) {
      return;
    }
    setDispatchSubmitting(true);
    setDispatchActionError("");
    try {
      await api.advanceDispatch(item.dispatchId);
      onChanged();
      await load();
    } catch (err) {
      setDispatchActionError(err instanceof Error ? err.message : "调度状态推进失败");
    } finally {
      setDispatchSubmitting(false);
    }
  }

  async function handleProductionQueueStatus(item: DispatchCenterQueueItem, status: string) {
    if (!(await confirmSensitiveAction(sensitiveActionPrompt(`production-queue-status-${item.dispatchId}-${status}`, "生产队列状态下发")))) {
      return;
    }
    setDispatchSubmitting(true);
    setDispatchActionError("");
    try {
      await api.advanceDispatch(item.dispatchId, status);
      onChanged();
      await load();
    } catch (err) {
      setDispatchActionError(err instanceof Error ? err.message : "生产队列状态下发失败");
    } finally {
      setDispatchSubmitting(false);
    }
  }

  async function handleCreateDeliveryNote(event: FormEvent) {
    event.preventDefault();
    const dispatchId = fieldNumber(deliveryForm.dispatchId);
    if (!dispatchId) {
      setActionError("请选择派车单");
      return;
    }
    await runBusinessAction("delivery-note-create", "送货单已生成", () => api.createDeliveryNote({
      dispatchId,
      ticketId: fieldNumber(deliveryForm.ticketId)
    }));
  }

  async function handleCreateDeliveryNoteLink(event: FormEvent, note: DeliveryNote) {
    event.preventDefault();
    await runBusinessAction(`delivery-note-link-${note.id}`, "签收二维码链接已生成", () => api.createDeliveryNoteSignLink(note.id, {
      channel: deliveryForm.channel || "qr",
      phone: deliveryForm.phone,
      expiresAt: apiDateTime(deliveryForm.expiresAt)
    }));
  }

  async function handleDeliveryNoteStatus(note: DeliveryNote, status: string) {
    await runBusinessAction(`delivery-note-status-${note.id}-${status}`, status === "void" ? "送货单已作废" : "送货单已重开", () => api.updateDeliveryNoteStatus(note.id, status));
  }

  async function handleSignAttachmentFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setActionError("");
    try {
      const payload = await browserFilePayload(file);
      setSignAttachmentForm((value) => ({
        ...value,
        fileName: payload.fileName,
        fileType: payload.fileType,
        url: payload.url,
        checksum: payload.checksum
      }));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "读取签收附件失败");
    }
  }

  async function handleReprintDeliveryNote(note: DeliveryNote) {
    setPrintDeliveryNoteId(note.id);
    await runBusinessAction(`delivery-note-reprint-${note.id}`, "送货单补打已记录", () => api.reprintDeliveryNote(note.id));
    window.setTimeout(() => window.print(), 80);
  }

  async function handleAddSignAttachment(event: FormEvent, sign: DeliverySign) {
    event.preventDefault();
    const signId = fieldNumber(signAttachmentForm.signId, sign.id);
    await runBusinessAction(`delivery-sign-attachment-${signId}`, "签收附件已归档", async () => {
      await api.addSignAttachment(signId, {
        dispatchId: sign.dispatchId,
        ticketId: sign.ticketId,
        fileName: signAttachmentForm.fileName.trim(),
        fileType: signAttachmentForm.fileType.trim(),
        url: signAttachmentForm.url.trim(),
        checksum: signAttachmentForm.checksum.trim(),
        uploadedBy: signAttachmentForm.uploadedBy.trim()
      });
      setSignAttachmentForm((value) => ({
        ...value,
        fileName: "",
        fileType: "image/jpeg",
        url: "",
        checksum: ""
      }));
    });
  }

  async function handleCreatePortalComplaint(event: FormEvent) {
    event.preventDefault();
    await runBusinessAction("portal-complaint-create", "门户投诉已提交", () => api.createPortalComplaint({
      customerId: fieldNumber(portalComplaintForm.customerId),
      projectId: fieldNumber(portalComplaintForm.projectId),
      title: portalComplaintForm.title.trim(),
      content: portalComplaintForm.content.trim(),
      level: portalComplaintForm.level
    }));
  }

  async function handleReportPortalDispatchException(event: FormEvent) {
    event.preventDefault();
    const dispatchId = fieldNumber(portalExceptionForm.dispatchId);
    if (!dispatchId) return;
    await runBusinessAction(`portal-dispatch-exception-${dispatchId}`, "派车异常已上报", () => api.reportPortalDispatchException(dispatchId, {
      exception: portalExceptionForm.exception.trim(),
      level: portalExceptionForm.level,
      alarmType: portalExceptionForm.alarmType
    }));
  }

  const visibleLatestLocations = data.latestLocations.length ? data.latestLocations : list(data.dispatch?.latestLocations);
  const moduleDispatchQueue = list(data.dispatch?.vehicleQueue).filter((item) => matchesCurrentSite(item.siteId));

  if (loading) {
    return <Panel>ERP 数据加载中...</Panel>;
  }

  function closeActionDialog(id: string) {
    setActionDialogId((current) => (current === id ? null : current));
  }

  return (
    <ActionDialogScope
      activeId={actionDialogId}
      closeDisabled={actionBusy !== ""}
      onActiveIdChange={setActionDialogId}
      onBeforeOpen={() => {
        setActionError("");
      }}
    >
      <div className="view-stack">
        {error ? (
          <Panel>
            <UiButton onClick={load}>重新加载</UiButton>
          </Panel>
        ) : null}
        {renderMasterDialog()}
        {renderBufferDialog()}
        {renderYardDialog()}
        {renderOrderDialog()}
        {renderProductionDialog()}
        {section === "overview" ? renderOverview() : null}
        {section === "master-customers" ? renderMasterCustomers() : null}
        {section === "customer-risk" ? renderMasterCustomers() : null}
        {section === "master-projects" ? renderMasterProjects() : null}
        {section === "master-products" ? renderMasterProducts() : null}
        {section === "sales-pricing" ? renderMasterProducts() : null}
        {section === "master-materials" ? renderMasterMaterials() : null}
	        {section === "master-sites" ? renderMasterSites() : null}
        {section === "master-plants" ? renderMasterPlants() : null}
        {section === "master-drivers" ? renderMasterDrivers() : null}
        {section === "master-vehicles" ? renderMasterVehicles() : null}
        {section === "master-carriers" ? renderMasterCarriers() : null}
        {section === "orders" ? renderOrders() : null}
        {section === "production" ? renderProduction() : null}
        {section === "production-plans" ? renderProduction() : null}
        {section === "production-tasks" ? renderProduction() : null}
        {section === "production-batches" ? renderProduction() : null}
        {section === "production-reports" ? renderProduction() : null}
        {section === "dispatch" ? renderDispatch() : null}
        {section === "dispatch-schedules" ? renderDispatch() : null}
        {section === "dispatch-queue" ? renderDispatch() : null}
        {section === "map-center" ? renderMapCenter() : null}
        {section === "weighbridge" ? renderWeighbridge() : null}
        {section === "delivery" ? renderDelivery() : null}
        {section === "delivery-signs" ? renderDelivery() : null}
        {section === "portal-customer" ? renderPortal() : null}
        {section === "portal-driver" ? renderPortal() : null}
        {section === "settlement" ? renderSettlement() : null}
        {section === "contracts" ? renderContracts() : null}
        {section === "stock-yards" ? renderProcurement() : null}
        {section === "raw-material-receipts" ? renderProcurement() : null}
        {section === "inventory-transfers" ? renderInventoryTransfers() : null}
        {section === "inventory-stocktakes" ? renderProcurement() : null}
        {section === "raw-material-inspections" ? renderProcurement() : null}
        {section === "finance" ? renderFinance() : null}
        {section === "finance-receivables" ? renderFinance() : null}
        {section === "finance-invoices" ? renderFinance() : null}
        {section === "finance-collections" ? renderFinance() : null}
        {section === "finance-suppliers" ? renderFinance() : null}
        {section === "finance-carriers" ? renderSettlement() : null}
        {section === "reports" ? renderReports() : null}
        {section === "approval-center" ? renderApprovalCenter() : null}
        {section === "system-org" ? renderOrganizationManagement() : null}
        {section === "system-license" ? renderLicenseManagement() : null}
        {section === "system-maintenance" ? renderSystemMaintenance() : null}
        {section === "system-gateway" ? renderSystemMaintenance() : null}
        {section === "system-security" ? renderSystemMaintenance() : null}
        {section === "system-identity" ? renderSystemMaintenance() : null}
        {section === "system-plugins" ? renderSystemMaintenance() : null}
        {section === "system-rules" ? renderSystemMaintenance() : null}
        {section === "system-integrations" ? renderSystemMaintenance() : null}
        {section === "system-menu" ? renderMenuManagement() : null}
        {section === "system-dictionaries" ? renderDictionaryManagement() : null}
        {section === "system-users" ? renderUserManagement() : null}
        {section === "system-roles" ? renderRoleManagement() : null}
        {section === "system-workflows" ? renderWorkflowManagement() : null}
        {section === "system-audit" ? renderAuditLogManagement() : null}
      </div>
    </ActionDialogScope>
  );

  function refreshData() {
    onChanged();
    void load();
  }

  function reloadButton() {
    return (
      <UiButton onClick={refreshData}>
        刷新
      </UiButton>
    );
  }

  function masterEntityName(kind: MasterKind) {
    return ({
      customer: "客户",
      project: "项目",
      product: "产品",
      material: "物料",
      site: "站点",
      plant: "生产线",
      driver: "司机",
      vehicle: "车辆",
	      carrier: "承运商"
    } as Record<MasterKind, string>)[kind];
  }

  function renderMasterDialog() {
    if (!masterDialogKind) {
      return null;
    }
    const editing = editingMaster?.kind === masterDialogKind;
    return (
      <Dialog
        open
        title={<>{editing ? "编辑" : "新增"}{masterEntityName(masterDialogKind)}</>}
        ariaLabel={`${editing ? "编辑" : "新增"}${masterEntityName(masterDialogKind)}`}
        className="master-dialog"
        closeDisabled={actionBusy !== ""}
        onClose={clearMasterEdit}
      >
        {renderMasterForm(masterDialogKind)}
      </Dialog>
    );
  }

  function renderBufferDialog() {
    if (!bufferDialogMode) {
      return null;
    }
    const plants = list(bootstrap?.plants).filter((item) => matchesCurrentSite(item.siteId));
    const bufferSource = list(data.production?.plantBufferLocations);
    const buffers = bufferSource.filter((item) => matchesCurrentSite(item.siteId));
    const materials = bootstrap?.materials || [];
    const selectedBuffer = buffers.find((item) => item.id === fieldNumber(bufferForm.bufferId));
    const yardPileSource = list(data.procurement?.stockYardPiles);
    const transferMaterialId = fieldNumber(bufferForm.materialId, selectedBuffer?.materialId || 0);
    const transferYardPiles = yardPileSource.filter((item) => (
      matchesCurrentSite(item.siteId)
      && item.currentQty > 0
      && (!transferMaterialId || item.materialId === transferMaterialId)
      && (item.qualityStatus === "passed" || item.qualityStatus === "")
      && (item.status === "active" || item.status === "running")
    ));
    const isBufferEditForm = bufferDialogMode === "create" || bufferDialogMode === "edit";
    const duplicateBufferCode = isBufferEditForm && plantBufferCodeDuplicated(bufferForm.code, fieldNumber(bufferForm.bufferId));
    const submitDisabled = actionBusy !== "" || (bufferDialogMode === "create" ? !plants.length : !buffers.length);
    const title = bufferDialogMode === "create" ? "新增筒仓" : bufferDialogMode === "edit" ? "编辑筒仓" : bufferDialogMode === "transfer" ? "堆场补料至筒仓" : "筒仓盘点校正";
    const bufferTypeOptions = dictionaryOptions("buffer_type");
    const qualityStatusOptions = dictionaryOptions("quality_status");
    const resourceStatusOptions = dictionaryOptions("resource_status");
    return (
      <Dialog open title={title} closeDisabled={actionBusy !== ""} onClose={() => setBufferDialogMode(null)}>
          <DialogForm onSubmit={handleBufferSubmit}>
            {isBufferEditForm ? (
              <>
                <Field label="生产线">
                  <SelectInput value={bufferForm.plantId} onChange={(event) => setBufferForm({ ...bufferForm, plantId: event.target.value })}>
                    {plants.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.code}</option>)}
                  </SelectInput>
                </Field>
                <Field label="筒仓编码"><TextInput value={bufferForm.code} onChange={(event) => setBufferForm({ ...bufferForm, code: event.target.value })} /></Field>
                <Field label="筒仓名称"><TextInput value={bufferForm.name} onChange={(event) => setBufferForm({ ...bufferForm, name: event.target.value })} /></Field>
                <Field label="筒仓类型">
                  <SelectInput value={bufferForm.type} onChange={(event) => setBufferForm({ ...bufferForm, type: event.target.value })}>
                    {bufferTypeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="物料">
                  <SelectInput value={bufferForm.materialId} onChange={(event) => setBufferForm({ ...bufferForm, materialId: event.target.value })}>
                    {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                <Field label="容量"><TextInput type="number" value={bufferForm.capacity} onChange={(event) => setBufferForm({ ...bufferForm, capacity: event.target.value })} /></Field>
                <Field label="单位"><TextInput value={bufferForm.unit} onChange={(event) => setBufferForm({ ...bufferForm, unit: event.target.value })} /></Field>
                <Field label="低位阈值"><TextInput type="number" value={bufferForm.warningQty} onChange={(event) => setBufferForm({ ...bufferForm, warningQty: event.target.value })} /></Field>
              </>
            ) : (
              <>
                <Field label="目标筒仓" spanAll>
                  <SelectInput value={bufferForm.bufferId} onChange={(event) => {
                    const next = buffers.find((item) => item.id === fieldNumber(event.target.value));
                    setBufferForm({
                      ...bufferForm,
                      bufferId: event.target.value,
                      materialId: next?.materialId ? String(next.materialId) : bufferForm.materialId,
                      yardPileId: "",
                      actualQty: next ? String(next.currentQty) : bufferForm.actualQty,
                      moistureRate: next ? String(next.moistureRate || 0) : bufferForm.moistureRate,
                      qualityStatus: next?.qualityStatus || bufferForm.qualityStatus,
                      status: next?.status || bufferForm.status
                    });
                  }}>
                    {buffers.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.code}</option>)}
                  </SelectInput>
                </Field>
                <Field label="物料">
                  <SelectInput value={bufferForm.materialId} disabled={bufferDialogMode === "adjust"} onChange={(event) => setBufferForm({ ...bufferForm, materialId: event.target.value })}>
                    {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                {bufferDialogMode === "transfer" ? (
                  <>
                    <Field label="来源堆场堆位">
                      <SelectInput value={bufferForm.yardPileId} onChange={(event) => {
                        const next = transferYardPiles.find((item) => item.id === fieldNumber(event.target.value));
                        setBufferForm({
                          ...bufferForm,
                          yardPileId: event.target.value,
                          materialId: next?.materialId ? String(next.materialId) : bufferForm.materialId,
                          unit: next?.unit || bufferForm.unit,
                          moistureRate: next ? String(next.moistureRate || 0) : bufferForm.moistureRate,
                          qualityStatus: next?.qualityStatus || bufferForm.qualityStatus
                        });
                      }}>
                        <option value="">不关联堆位，仅记录筒仓补料</option>
                        {transferYardPiles.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.code} / {qty(item.currentQty)} {item.unit}</option>)}
                      </SelectInput>
                    </Field>
                    <Field label="补料数量"><TextInput type="number" value={bufferForm.transferQty} onChange={(event) => setBufferForm({ ...bufferForm, transferQty: event.target.value })} /></Field>
                  </>
                ) : (
                  <Field label="实测数量"><TextInput type="number" value={bufferForm.actualQty} onChange={(event) => setBufferForm({ ...bufferForm, actualQty: event.target.value })} /></Field>
                )}
                <Field label="含水率 %"><TextInput type="number" value={bufferForm.moistureRate} onChange={(event) => setBufferForm({ ...bufferForm, moistureRate: event.target.value })} /></Field>
                <Field label="质量状态">
                  <SelectInput value={bufferForm.qualityStatus} onChange={(event) => setBufferForm({ ...bufferForm, qualityStatus: event.target.value })}>
                    {qualityStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="状态">
                  <SelectInput value={bufferForm.status} onChange={(event) => setBufferForm({ ...bufferForm, status: event.target.value })}>
                    {resourceStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="备注" spanAll><TextInput value={bufferForm.remark} onChange={(event) => setBufferForm({ ...bufferForm, remark: event.target.value })} /></Field>
                {selectedBuffer ? <p className="muted span-all">目标筒仓余额 {qty(selectedBuffer.currentQty)} {selectedBuffer.unit} / 容量 {qty(selectedBuffer.capacity)} {selectedBuffer.unit}</p> : null}
              </>
            )}
            <FormActions spanAll>
              <UiButton variant="primary" type="submit" disabled={submitDisabled}>
                {actionBusy !== "" ? "操作中..." : title}
              </UiButton>
            </FormActions>
          </DialogForm>
      </Dialog>
    );
  }

  function renderYardDialog() {
    if (!yardDialogMode) {
      return null;
    }
    const sites = siteOptions();
    const materials = bootstrap?.materials || [];
    const suppliers = supplierOptions();
    const yardSource = list(data.procurement?.stockYards);
    const pileSource = list(data.procurement?.stockYardPiles);
    const yards = yardSource.filter((item) => matchesCurrentSite(item.siteId));
    const piles = pileSource.filter((item) => matchesCurrentSite(item.siteId));
    const selectedPile = piles.find((item) => item.id === fieldNumber(yardForm.pileId));
    const isYardForm = yardDialogMode === "yard" || yardDialogMode === "yard-edit";
    const isPileForm = yardDialogMode === "pile" || yardDialogMode === "pile-edit";
    const title = yardDialogMode === "yard" ? "新增堆场" : yardDialogMode === "yard-edit" ? "编辑堆场" : yardDialogMode === "pile" ? "新增堆位" : yardDialogMode === "pile-edit" ? "编辑堆位" : yardDialogMode === "receipt" ? "堆场入场" : "堆位盘点";
    const yardTypeOptions = dictionaryOptions("yard_type");
    const qualityStatusOptions = dictionaryOptions("quality_status");
    const resourceStatusOptions = dictionaryOptions("resource_status");
    return (
      <Dialog open title={title} size="lg" className="action-dialog yard-form-dialog" closeDisabled={actionBusy !== ""} onClose={() => setYardDialogMode(null)}>
          <DialogForm onSubmit={handleYardSubmit}>
            {isYardForm ? (
              <>
                {renderSiteField("站点", yardForm.siteId, (siteId) => setYardForm({ ...yardForm, siteId }))}
                <Field label="堆场编码"><TextInput value={yardForm.code} onChange={(event) => setYardForm({ ...yardForm, code: event.target.value })} /></Field>
                <Field label="堆场名称"><TextInput value={yardForm.name} onChange={(event) => setYardForm({ ...yardForm, name: event.target.value })} /></Field>
                <Field label="类型">
                  <SelectInput value={yardForm.type} onChange={(event) => setYardForm({ ...yardForm, type: event.target.value })}>
                    {yardTypeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="区域"><TextInput value={yardForm.area} onChange={(event) => setYardForm({ ...yardForm, area: event.target.value })} /></Field>
                <Field label="容量"><TextInput type="number" value={yardForm.capacity} onChange={(event) => setYardForm({ ...yardForm, capacity: event.target.value })} /></Field>
                <Field label="单位"><TextInput value={yardForm.unit} onChange={(event) => setYardForm({ ...yardForm, unit: event.target.value })} /></Field>
              </>
            ) : isPileForm ? (
              <>
                <Field label="堆场" spanAll>
                  <SelectInput value={yardForm.yardId} onChange={(event) => {
                    const yard = yards.find((item) => item.id === fieldNumber(event.target.value));
                    setYardForm({
                      ...yardForm,
                      yardId: event.target.value,
                      siteId: yard ? String(yard.siteId) : yardForm.siteId,
                      unit: yard?.unit || yardForm.unit
                    });
                  }}>
                    {yards.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.code}</option>)}
                  </SelectInput>
                </Field>
                <Field label="堆位编码"><TextInput value={yardForm.code} onChange={(event) => setYardForm({ ...yardForm, code: event.target.value })} /></Field>
                <Field label="堆位名称"><TextInput value={yardForm.name} onChange={(event) => setYardForm({ ...yardForm, name: event.target.value })} /></Field>
                <Field label="物料">
                  <SelectInput value={yardForm.materialId} onChange={(event) => setYardForm({ ...yardForm, materialId: event.target.value })}>
                    {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                <Field label="供应商">
                  <SelectInput value={yardForm.supplierId} onChange={(event) => setYardForm({ ...yardForm, supplierId: event.target.value })}>
                    {suppliers.map((item) => <option key={recordId(item)} value={recordId(item)}>{recordName(item, `供应商 ${recordId(item)}`)}</option>)}
                  </SelectInput>
                </Field>
                <Field label="批次"><TextInput value={yardForm.batchNo} onChange={(event) => setYardForm({ ...yardForm, batchNo: event.target.value })} /></Field>
                <Field label="容量"><TextInput type="number" value={yardForm.capacity} onChange={(event) => setYardForm({ ...yardForm, capacity: event.target.value })} /></Field>
                <Field label="低位阈值"><TextInput type="number" value={yardForm.warningQty} onChange={(event) => setYardForm({ ...yardForm, warningQty: event.target.value })} /></Field>
              </>
            ) : (
              <>
                <Field label="堆位" spanAll>
                  <SelectInput value={yardForm.pileId} onChange={(event) => {
                    const pile = piles.find((item) => item.id === fieldNumber(event.target.value));
                    setYardForm({
                      ...yardForm,
                      pileId: event.target.value,
                      yardId: pile ? String(pile.yardId) : yardForm.yardId,
                      materialId: pile?.materialId ? String(pile.materialId) : yardForm.materialId,
                      supplierId: pile?.supplierId ? String(pile.supplierId) : yardForm.supplierId,
                      batchNo: pile?.batchNo || yardForm.batchNo,
                      currentQty: String(pile?.currentQty || 0),
                      actualQty: String(pile?.currentQty || 0),
                      unit: pile?.unit || yardForm.unit,
                      moistureRate: String(pile?.moistureRate || 0),
                      qualityStatus: pile?.qualityStatus || yardForm.qualityStatus,
                      status: pile?.status || yardForm.status
                    });
                  }}>
                    {piles.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.code}</option>)}
                  </SelectInput>
                </Field>
                <Field label="物料">
                  <SelectInput value={yardForm.materialId} disabled={yardDialogMode === "adjust"} onChange={(event) => setYardForm({ ...yardForm, materialId: event.target.value })}>
                    {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                {yardDialogMode === "receipt" ? (
                  <>
                    <Field label="供应商">
                      <SelectInput value={yardForm.supplierId} onChange={(event) => setYardForm({ ...yardForm, supplierId: event.target.value })}>
                        {suppliers.map((item) => <option key={recordId(item)} value={recordId(item)}>{recordName(item, `供应商 ${recordId(item)}`)}</option>)}
                      </SelectInput>
                    </Field>
                    <Field label="入场数量"><TextInput type="number" value={yardForm.receiptQty} onChange={(event) => setYardForm({ ...yardForm, receiptQty: event.target.value })} /></Field>
                    <Field label="批次"><TextInput value={yardForm.batchNo} onChange={(event) => setYardForm({ ...yardForm, batchNo: event.target.value })} /></Field>
                  </>
                ) : (
                  <Field label="实测数量"><TextInput type="number" value={yardForm.actualQty} onChange={(event) => setYardForm({ ...yardForm, actualQty: event.target.value })} /></Field>
                )}
                <Field label="含水率 %"><TextInput type="number" value={yardForm.moistureRate} onChange={(event) => setYardForm({ ...yardForm, moistureRate: event.target.value })} /></Field>
                <Field label="质量状态">
                  <SelectInput value={yardForm.qualityStatus} onChange={(event) => setYardForm({ ...yardForm, qualityStatus: event.target.value })}>
                    {qualityStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="状态">
                  <SelectInput value={yardForm.status} onChange={(event) => setYardForm({ ...yardForm, status: event.target.value })}>
                    {resourceStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                  </SelectInput>
                </Field>
                <Field label="备注" spanAll><TextInput value={yardForm.remark} onChange={(event) => setYardForm({ ...yardForm, remark: event.target.value })} /></Field>
                {selectedPile ? <p className="muted span-all">当前余额 {qty(selectedPile.currentQty)} {selectedPile.unit} / 容量 {qty(selectedPile.capacity)} {selectedPile.unit}</p> : null}
              </>
            )}
            {isYardForm || isPileForm ? (
              <Field label="状态">
                <SelectInput value={yardForm.status} onChange={(event) => setYardForm({ ...yardForm, status: event.target.value })}>
                  {resourceStatusOptions.filter((item) => item.code !== "cleaning").map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
            ) : null}
            <FormActions spanAll>
              <UiButton variant="primary" type="submit" disabled={actionBusy !== "" || (isYardForm ? !sites.length : isPileForm ? !yards.length || !materials.length : !piles.length)}>
                {actionBusy !== "" ? "操作中..." : title}
              </UiButton>
            </FormActions>
          </DialogForm>
      </Dialog>
    );
  }

  function renderMasterForm(kind: MasterKind) {
    const customers = bootstrap?.customers || [];
    const companies = bootstrap?.companies || [];
    const materials = bootstrap?.materials || [];
    const sites = siteOptions();
    const drivers = bootstrap?.drivers || [];
    const siteFenceRadiusValue = editableFenceRadius(masterForm.siteFenceRadius);
    const projectLongitudeValue = fieldNumber(masterForm.projectLongitude);
    const projectLatitudeValue = fieldNumber(masterForm.projectLatitude);
    const projectLocationReady = isValidCoordinate(projectLongitudeValue, projectLatitudeValue);
    const projectAddressPreview = masterForm.projectAddress.trim();
    const siteAddressPreview = masterForm.siteAddress.trim();
    const editingPlant = editingMaster?.kind === "plant" ? productionPlants().find((item) => item.id === editingMaster.id) : undefined;
    const editingPlantRecipe = editingPlant ? productionLineRecipe(editingPlant) : undefined;
    const plantStatusOptions = dictionaryOptions("plant_status");
    const vehicleTypeOptions = dictionaryOptions("vehicle_type");
    const resourceStatusOptions = dictionaryOptionsWithFallback("resource_status", [
      { code: "active", label: "启用" },
      { code: "disabled", label: "禁用" }
    ]);
	    const carrierSettleModeOptions = dictionaryOptionsWithFallback("carrier_settle_mode", carrierSettleModeFallbackOptions);
    switch (kind) {
      case "customer":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("customer", event)}>
            <Field label="客户名称"><TextInput value={masterForm.customerName} onChange={(event) => setMasterForm({ ...masterForm, customerName: event.target.value })} /></Field>
            <Field label="联系人"><TextInput value={masterForm.customerContact} onChange={(event) => setMasterForm({ ...masterForm, customerContact: event.target.value })} /></Field>
            <Field label="电话" spanAll><TextInput value={masterForm.customerPhone} onChange={(event) => setMasterForm({ ...masterForm, customerPhone: event.target.value })} /></Field>
            {masterFormButton("customer", "新增客户")}
          </DialogForm>
        );
      case "project":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("project", event)}>
            <Field label="客户">
              <SelectInput value={masterForm.projectCustomerId} onChange={(event) => setMasterForm({ ...masterForm, projectCustomerId: event.target.value })}>
                {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            <Field label="项目名称"><TextInput value={masterForm.projectName} onChange={(event) => setMasterForm({ ...masterForm, projectName: event.target.value })} /></Field>
            <Field label="项目地址" spanAll className="site-address-field">
              <div className="site-address-locate-row">
                <TextInput value={masterForm.projectAddress} onChange={(event) => {
                  setProjectAddressLookupState("");
                  setMasterForm({ ...masterForm, projectAddress: event.target.value });
                }} />
                <IconButton className="site-address-locate-button" icon={<MapPin size={16} />} label="地图定位" onClick={openProjectLocationDialog} />
              </div>
            </Field>
            <Dialog
              open={projectLocationDialogOpen}
              title="选择项目位置"
              ariaLabel="选择项目位置"
              size="xl"
              className="site-fence-dialog"
              backdropClassName="site-fence-dialog-backdrop"
              onClose={() => setProjectLocationDialogOpen(false)}
              footer={<UiButton variant="primary" onClick={() => setProjectLocationDialogOpen(false)}>完成</UiButton>}
            >
              <div ref={projectLocationPickerRef} className="site-fence-picker site-fence-picker-dialog">
                <div className={`site-fence-address-preview is-${projectAddressLookupState || (projectLocationReady ? "success" : "idle")}`}>
                  <span>{projectAddressLookupState === "loading" ? "地址识别中" : "地址预览"}</span>
                  <b>
                    {projectAddressLookupState === "loading"
                      ? "正在识别地址..."
                      : projectAddressLookupState === "failed"
                        ? "未识别到地址，可手动填写"
                        : projectAddressPreview || "点击地图选择项目位置"}
                  </b>
                  {projectLocationReady ? <b>坐标 {projectLongitudeValue.toFixed(6)}, {projectLatitudeValue.toFixed(6)}</b> : null}
                </div>
                <SiteFenceMap
                  longitude={projectLongitudeValue}
                  latitude={projectLatitudeValue}
                  provider={mapConfig}
                  onCenterChange={(longitude, latitude) => {
                    setMasterForm((form) => ({
                      ...form,
                      projectLongitude: String(longitude),
                      projectLatitude: String(latitude)
                    }));
                    void fillProjectAddressFromLocation(longitude, latitude);
                  }}
                />
              </div>
            </Dialog>
            {masterFormButton("project", "新增项目", !customers.length)}
          </DialogForm>
        );
      case "product":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("product", event)}>
            <Field label="产品名称"><TextInput value={masterForm.productName} onChange={(event) => setMasterForm({ ...masterForm, productName: event.target.value })} /></Field>
            <Field label="规格"><TextInput value={masterForm.productSpec} onChange={(event) => setMasterForm({ ...masterForm, productSpec: event.target.value })} /></Field>
            <Field label="基准价" spanAll><TextInput type="number" value={masterForm.productPrice} onChange={(event) => setMasterForm({ ...masterForm, productPrice: event.target.value })} /></Field>
            {masterFormButton("product", "新增产品")}
          </DialogForm>
        );
      case "material":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("material", event)}>
            <Field label="物料名称"><TextInput value={masterForm.materialName} onChange={(event) => setMasterForm({ ...masterForm, materialName: event.target.value })} /></Field>
            <Field label="规格"><TextInput value={masterForm.materialSpec} onChange={(event) => setMasterForm({ ...masterForm, materialSpec: event.target.value })} /></Field>
            <Field label="安全库存" spanAll><TextInput type="number" value={masterForm.materialSafeStock} onChange={(event) => setMasterForm({ ...masterForm, materialSafeStock: event.target.value })} /></Field>
            {masterFormButton("material", "新增物料")}
          </DialogForm>
        );
      case "site":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("site", event)}>
            <Field label="站点名称"><TextInput value={masterForm.siteName} onChange={(event) => setMasterForm({ ...masterForm, siteName: event.target.value })} /></Field>
            <Field label="编码"><TextInput value={masterForm.siteCode} onChange={(event) => setMasterForm({ ...masterForm, siteCode: event.target.value })} /></Field>
            <Field label="归属公司">
              <SelectInput value={masterForm.siteCompanyId} onChange={(event) => setMasterForm({ ...masterForm, siteCompanyId: event.target.value })}>
                {companies.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            <Field label="地址" spanAll className="site-address-field">
              <div className="site-address-locate-row">
                <TextInput value={masterForm.siteAddress} onChange={(event) => setMasterForm({ ...masterForm, siteAddress: event.target.value })} />
                <IconButton className="site-address-locate-button" icon={<MapPin size={16} />} label="定位电子围栏" onClick={openSiteFenceDialog} />
              </div>
            </Field>
            <Dialog
              open={siteFenceDialogOpen}
              title="选择电子围栏位置"
              ariaLabel="选择电子围栏位置"
              size="xl"
              className="site-fence-dialog"
              backdropClassName="site-fence-dialog-backdrop"
              onClose={() => setSiteFenceDialogOpen(false)}
              footer={<UiButton variant="primary" onClick={() => setSiteFenceDialogOpen(false)}>完成</UiButton>}
            >
              <div ref={siteFencePickerRef} className="site-fence-picker site-fence-picker-dialog">
                <div className="site-fence-range-panel">
                  <div className="site-fence-range-head">
                    <span>覆盖范围</span>
                    <b>{siteFenceRangeLabel(siteFenceRadiusValue)}</b>
                  </div>
                  <div className="site-fence-range-options">
                    {siteFenceRangeOptions.map((option) => (
                      <ChipButton
                        key={option.value}
                        active={siteFenceRadiusValue === option.value}
                        onClick={() => setMasterForm((form) => ({ ...form, siteFenceRadius: String(option.value) }))}
                      >
                        {option.label}
                      </ChipButton>
                    ))}
                  </div>
                  <input
                    aria-label="调整覆盖范围"
                    className="site-fence-range-slider"
                    max={siteFenceRadiusMax}
                    min={siteFenceRadiusMin}
                    step={siteFenceRadiusStep}
                    type="range"
                    value={siteFenceRadiusValue}
                    onChange={(event) => setMasterForm((form) => ({ ...form, siteFenceRadius: event.target.value }))}
                  />
                </div>
                {(siteAddressLookupState || siteAddressPreview) ? (
                  <div className={`site-fence-address-preview is-${siteAddressLookupState || "idle"}`}>
                    <span>{siteAddressLookupState === "loading" ? "地址识别中" : "地址预览"}</span>
                    <b>{siteAddressLookupState === "loading" ? "正在识别地址..." : siteAddressLookupState === "failed" ? "未识别到地址，可手动填写" : siteAddressPreview}</b>
                  </div>
                ) : null}
                <SiteFenceMap
                  longitude={fieldNumber(masterForm.siteLongitude)}
                  latitude={fieldNumber(masterForm.siteLatitude)}
                  radius={siteFenceRadiusValue}
                  provider={mapConfig}
                  onCenterChange={(longitude, latitude) => {
                    setMasterForm((form) => ({
                      ...form,
                      siteLongitude: String(longitude),
                      siteLatitude: String(latitude)
                    }));
                    void fillSiteAddressFromLocation(longitude, latitude);
                  }}
                />
              </div>
            </Dialog>
            {masterFormButton("site", "新增站点", !companies.length)}
          </DialogForm>
        );
      case "plant":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("plant", event)}>
            {editingPlantRecipe ? (
              <MetricList compact className="span-all">
                <div><span>当前配比</span><b>{productionMixLabel(editingPlantRecipe.mix)}</b></div>
                <div><span>适配配比</span><b>{productionProfileLabel(editingPlantRecipe.profile)}</b></div>
                <div><span>来源</span><b>{editingPlantRecipe.source}</b></div>
                <div><span>产品</span><b>{editingPlantRecipe.productId ? productLabel(bootstrap, editingPlantRecipe.productId) : "-"}</b></div>
                <div><span>任务/计划</span><b>{editingPlantRecipe.task ? editingPlantRecipe.task.taskNo : editingPlantRecipe.plan ? editingPlantRecipe.plan.planNo : "-"}</b></div>
              </MetricList>
            ) : null}
            {renderSiteField("站点", masterForm.plantSiteId, (siteId) => setMasterForm({ ...masterForm, plantSiteId: siteId }))}
            <Field label="生产线名称"><TextInput value={masterForm.plantName} onChange={(event) => setMasterForm({ ...masterForm, plantName: event.target.value })} /></Field>
            <Field label="编码"><TextInput value={masterForm.plantCode} onChange={(event) => setMasterForm({ ...masterForm, plantCode: event.target.value })} /></Field>
            <Field label="产能"><TextInput value={masterForm.plantCapacity} onChange={(event) => setMasterForm({ ...masterForm, plantCapacity: event.target.value })} /></Field>
            <Field label="状态">
              <SelectInput value={masterForm.plantStatus} onChange={(event) => setMasterForm({ ...masterForm, plantStatus: event.target.value })}>
                {plantStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            {masterFormButton("plant", "新增生产线", !sites.length)}
          </DialogForm>
        );
      case "driver":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("driver", event)}>
            <Field label="司机"><TextInput value={masterForm.driverName} onChange={(event) => setMasterForm({ ...masterForm, driverName: event.target.value })} /></Field>
            <Field label="电话"><TextInput value={masterForm.driverPhone} onChange={(event) => setMasterForm({ ...masterForm, driverPhone: event.target.value })} /></Field>
            {masterFormButton("driver", "新增司机")}
          </DialogForm>
        );
      case "carrier":
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("carrier", event)}>
            <Field label="承运商名称"><TextInput value={masterForm.carrierName} onChange={(event) => setMasterForm({ ...masterForm, carrierName: event.target.value })} /></Field>
            <Field label="联系人"><TextInput value={masterForm.carrierContact} onChange={(event) => setMasterForm({ ...masterForm, carrierContact: event.target.value })} /></Field>
            <Field label="电话"><TextInput value={masterForm.carrierPhone} onChange={(event) => setMasterForm({ ...masterForm, carrierPhone: event.target.value })} /></Field>
            <Field label="结算方式">
              <SelectInput value={masterForm.carrierSettleMode} onChange={(event) => setMasterForm({ ...masterForm, carrierSettleMode: event.target.value })}>
                {carrierSettleModeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="状态" spanAll>
              <SelectInput value={masterForm.carrierStatus} onChange={(event) => setMasterForm({ ...masterForm, carrierStatus: event.target.value })}>
                {resourceStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            {masterFormButton("carrier", "新增承运商")}
          </DialogForm>
        );
      case "vehicle":
        const editingVehicle = editingMaster?.kind === "vehicle" ? (bootstrap?.vehicles || []).find((item) => item.id === editingMaster.id) : undefined;
        const editingVehicleDevice = vehicleDeviceFor(editingVehicle?.id);
        const editingVehicleLocation = latestLocationFor(editingVehicle?.id);
        return (
          <DialogForm onSubmit={(event) => handleMasterSubmit("vehicle", event)}>
            <Field label="自编号"><TextInput value={masterForm.vehicleInternalNo} onChange={(event) => setMasterForm({ ...masterForm, vehicleInternalNo: event.target.value })} /></Field>
            <Field label="车牌"><TextInput value={masterForm.vehiclePlate} onChange={(event) => setMasterForm({ ...masterForm, vehiclePlate: event.target.value })} /></Field>
            <Field label="车型">
              <SelectInput value={masterForm.vehicleType} onChange={(event) => setMasterForm({ ...masterForm, vehicleType: event.target.value })}>
                {vehicleTypeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="司机">
              <SelectInput value={masterForm.vehicleDriverId} onChange={(event) => setMasterForm({ ...masterForm, vehicleDriverId: event.target.value })}>
                {drivers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            {renderSiteField("站点", masterForm.vehicleSiteId, (siteId) => setMasterForm({ ...masterForm, vehicleSiteId: siteId }))}
            <Field label="容量" spanAll><TextInput value={masterForm.vehicleCapacity} onChange={(event) => setMasterForm({ ...masterForm, vehicleCapacity: event.target.value })} /></Field>
            <Field label="GPS 设备号" spanAll><TextInput value={masterForm.vehicleDeviceNo} onChange={(event) => setMasterForm({ ...masterForm, vehicleDeviceNo: event.target.value })} placeholder="与 GPS 转发器上报的 deviceNo 对应" /></Field>
            {editingVehicleDevice || editingVehicleLocation ? (
              <MetricList compact className="span-all">
                <div><span>设备状态</span><b>{editingVehicleDevice ? <StatusChip value={editingVehicleDevice.status} /> : "-"}</b></div>
                <div><span>最后接入</span><b>{shortDateTime(editingVehicleDevice?.lastSeenAt) || "-"}</b></div>
                <div><span>最近定位</span><b>{shortDateTime(editingVehicleLocation?.lastLocationTime) || "-"}</b></div>
                <div><span>坐标</span><b>{editingVehicleLocation ? `${editingVehicleLocation.latitude.toFixed(4)}, ${editingVehicleLocation.longitude.toFixed(4)}` : "-"}</b></div>
              </MetricList>
            ) : null}
            {masterFormButton("vehicle", "新增车辆", !drivers.length || !sites.length)}
          </DialogForm>
        );
    }
  }

  function renderOrderDialog() {
    if (!orderDialogOpen) {
      return null;
    }
    const customers = bootstrap?.customers || [];
    const projects = bootstrap?.projects || [];
    const products = bootstrap?.products || [];
    const sites = siteOptions();
    const disabled = actionBusy !== "" || !customers.length || !projects.length || !products.length || !sites.length;
    return (
      <Dialog open title="新增销售订单" className="order-dialog" closeDisabled={actionBusy !== ""} onClose={() => setOrderDialogOpen(false)}>
          <DialogForm onSubmit={handleCreateOrder}>
            <Field label="客户">
              <SelectInput value={orderForm.customerId} onChange={(event) => setOrderForm({ ...orderForm, customerId: event.target.value })}>
                {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            <Field label="项目">
              <SelectInput value={orderForm.projectId} onChange={(event) => setOrderForm({ ...orderForm, projectId: event.target.value })}>
                {projects.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            <Field label="产品">
              <SelectInput value={orderForm.productId} onChange={(event) => setOrderForm({ ...orderForm, productId: event.target.value })}>
                {products.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
              </SelectInput>
            </Field>
            {renderSiteField("站点", orderForm.siteId, (siteId) => setOrderForm({ ...orderForm, siteId }))}
            <Field label="计划方量"><TextInput type="number" min="0" step="0.5" value={orderForm.planQuantity} onChange={(event) => setOrderForm({ ...orderForm, planQuantity: event.target.value })} /></Field>
            <Field label="单价"><TextInput type="number" min="0" step="1" value={orderForm.unitPrice} onChange={(event) => setOrderForm({ ...orderForm, unitPrice: event.target.value })} /></Field>
            <HeroDateField className="span-all" label="计划时间" value={orderForm.planTime} onChange={(planTime) => setOrderForm({ ...orderForm, planTime })} />
            <FormActions spanAll>
              <UiButton disabled={actionBusy !== ""} onClick={() => setOrderDialogOpen(false)}>取消</UiButton>
              <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={disabled}>提交订单</UiButton>
            </FormActions>
          </DialogForm>
      </Dialog>
    );
  }

  function renderProductionDialog() {
    if (!productionDialogMode) {
      return null;
    }
    const plans = list(data.production?.plans).filter((item) => matchesCurrentSite(item.siteId));
    const tasks = list(data.production?.tasks).filter((item) => matchesCurrentSite(item.siteId));
    const openOrders = scopedOrders.filter((item) => productionOrderRemaining(item, plans) > 0 && ["approved", "scheduled", "dispatching"].includes(item.status));
    const selectedPlanFromData = plans.find((item) => String(item.id) === productionForm.planId);
    const selectedPlan = productionDialogMode === "report" && productionReportPlan ? productionReportPlan : selectedPlanFromData;
    const activeTasks = tasks.filter((item) => item.status !== "cancelled" && item.status !== "completed");
    const taskSelectablePlans = plans.filter(productionPlanCanIssueTask);
    const batchCandidateTasks = sortProductionTasksForAction(activeTasks.filter((item) => productionTaskRemaining(item) > 0));
    const scopedBatchTasks = productionDialogMode === "batch" && selectedPlan ? batchCandidateTasks.filter((item) => item.planId === selectedPlan.id) : batchCandidateTasks;
    const batchSelectableTasks = productionDialogMode === "batch" && selectedPlan ? scopedBatchTasks : batchCandidateTasks;
    const selectedTask = productionDialogMode === "batch"
      ? batchSelectableTasks.find((item) => String(item.id) === productionForm.taskId) || batchSelectableTasks[0]
      : activeTasks.find((item) => String(item.id) === productionForm.taskId) || tasks.find((item) => String(item.id) === productionForm.taskId);
    const selectedOrder = openOrders.find((item) => String(item.id) === productionForm.orderId);
    const selectedTaskPlan = selectedTask ? plans.find((item) => item.id === selectedTask.planId) : undefined;
    const selectedPlantSiteId = selectedOrder?.siteId || selectedPlan?.siteId || selectedTaskPlan?.siteId || selectedTask?.siteId;
    const selectedPlantOptions = productionPlantOptions(selectedPlantSiteId);
    const selectedPlant = selectedPlantOptions.find((item) => String(item.id) === productionForm.plantId)
      || (productionDialogMode === "create-plan" ? undefined : productionPlanPlant(selectedPlan))
      || (productionDialogMode === "create-plan" ? undefined : productionTaskPlant(selectedTask))
      || firstProductionPlant(selectedPlantSiteId);
    const selectedPlanTasks = selectedPlan ? tasks.filter((item) => item.planId === selectedPlan.id) : [];
    const selectedPlanBatches = productionDialogMode === "report" && productionReportBatches
      ? productionReportBatches
      : selectedPlan ? list(data.production?.batches).filter((item) => item.planId === selectedPlan.id) : [];
    const selectedPlanReports = selectedPlan ? list(data.production?.reports).filter((item) => item.siteId === selectedPlan.siteId && item.reportDate === selectedPlan.planDate) : [];
    const selectedReportPlans = productionReportDayPlans(selectedPlan);
    const reportBatchMap = new Map<number, ProductionBatch>();
    if (selectedPlan) {
      list(data.production?.batches)
        .filter((item) => item.siteId === selectedPlan.siteId && productionBatchReportDate(item) === selectedPlan.planDate)
        .forEach((item) => reportBatchMap.set(item.id, item));
      if (productionDialogMode === "report" && productionReportBatches) {
        productionReportBatches
          .filter((item) => item.siteId === selectedPlan.siteId && productionBatchReportDate(item) === selectedPlan.planDate)
          .forEach((item) => reportBatchMap.set(item.id, item));
      }
    }
    const selectedReportBatches = Array.from(reportBatchMap.values());
    const selectedOrderRemaining = selectedOrder ? productionOrderRemaining(selectedOrder, plans) : 0;
    const selectedPlanTaskRemaining = productionPlanTaskRemaining(selectedPlan);
    const selectedTaskRemaining = productionTaskRemaining(selectedTask);
    const planQuantityValue = fieldNumber(productionForm.planQuantity);
    const adjustedPlanQuantityValue = fieldNumber(productionForm.adjustPlanQuantity);
    const taskQtyValue = fieldNumber(productionForm.taskQty);
    const batchQtyValue = fieldNumber(productionForm.batchQty);
    const selectedReportPlannedQty = selectedReportPlans.reduce((sum, item) => sum + item.planQuantity, 0);
    const selectedReportProducedQty = selectedReportBatches.reduce((sum, item) => sum + item.quantity, 0);
    const selectedReportPassedBatchCount = selectedReportBatches.filter((item) => item.qualityStatus === "passed").length;
    const selectedReportPendingBatchCount = selectedReportBatches.filter((item) => item.qualityStatus !== "passed").length;
    const close = () => {
      setProductionReportPlan(null);
      setProductionReportBatches(null);
      setProductionDialogMode(null);
    };
    const title = {
      detail: "生产计划详情",
      "create-plan": "新建生产计划",
      "edit-plan": "调整生产计划",
      tasks: "下达生产任务",
      batch: "登记生产批次",
      report: "生成生产日报",
      "cancel-plan": "取消生产计划"
    }[productionDialogMode];
    const shiftOptions = dictionaryOptions("shift_type");
    const qualityStatusOptions = dictionaryOptions("quality_status");
    function renderProductionDialogBody() {
      switch (productionDialogMode) {
        case "detail":
          return (
            <DialogContent className="dialog-form production-detail-dialog detail-dialog-body">
              {selectedPlan ? (
                <>
                  <MetricList compact className="detail-summary-grid">
                    <div><span>计划号</span><b>{selectedPlan.planNo}</b></div>
                    <div><span>订单</span><b>{data.orders.find((order) => order.id === selectedPlan.orderId)?.orderNo || selectedPlan.orderId}</b></div>
                    <div><span>站点</span><b>{nameOf(bootstrap?.sites, selectedPlan.siteId)}</b></div>
                    <div><span>生产线</span><b>{productionPlantLabel(productionPlanPlant(selectedPlan))}</b></div>
                    <div><span>产品</span><b>{productLabel(bootstrap, selectedPlan.productId)}</b></div>
                    <div><span>日期/班次</span><b>{selectedPlan.planDate} · {dictionaryValueLabel("shift_type", selectedPlan.shift, selectedPlan.shift)}</b></div>
                    <div><span>状态</span><b>{workflowStatusFor(["production_plan", "productionPlan"], selectedPlan.id, selectedPlan.planNo, <StatusChip value={selectedPlan.status} />)}</b></div>
                    <div><span>计划/已产</span><b>{qty(selectedPlan.planQuantity)} / {qty(selectedPlan.producedQty)}</b></div>
                    <div><span>剩余</span><b>{qty(selectedPlan.remainingQty)}</b></div>
                    <div><span>任务已下达</span><b>{qty(selectedPlan.plannedTaskQty)}</b></div>
                    <div><span>进度</span><b>{percent(selectedPlan.progress)}</b></div>
                    <div><span>产能</span><b><StatusChip value={selectedPlan.capacityStatus} /></b></div>
                    <div><span>库存</span><b><StatusChip value={selectedPlan.inventoryStatus} /></b></div>
                    <div><span>配比匹配</span><b><StatusChip value={selectedPlan.recipeStatus} /></b></div>
                  </MetricList>
                  {selectedPlan.riskReason ? <p className="muted detail-alert">{selectedPlan.riskReason}</p> : null}
                  {workflowTimelineBlock(["production_plan", "productionPlan"], selectedPlan.id, selectedPlan.planNo, "当前生产计划暂无工作流实例")}
                  <FormActions className="detail-toolbar">
                    <UiButton icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || selectedPlan.status === "completed" || selectedPlan.status === "cancelled"} onClick={() => openProductionDialog("edit-plan", selectedPlan)}>调整计划</UiButton>
                    <UiButton icon={<PlayCircle size={14} />} disabled={actionBusy !== "" || !productionPlanCanIssueTask(selectedPlan)} onClick={() => openProductionDialog("tasks", selectedPlan)}>下达任务</UiButton>
                    <UiButton icon={<Plus size={14} />} disabled={actionBusy !== "" || !activeProductionTasksForPlan(selectedPlan).length} onClick={() => openProductionDialog("batch", selectedPlan)}>登记批次</UiButton>
                    <UiButton icon={<ClipboardCheck size={14} />} disabled={actionBusy !== "" || !productionPlanCanOpenReport(selectedPlan)} onClick={() => openProductionDialog("report", selectedPlan)}>生成日报</UiButton>
                    <UiButton variant="danger" icon={<X size={14} />} disabled={actionBusy !== "" || !productionPlanCanCancel(selectedPlan)} onClick={() => openProductionDialog("cancel-plan", selectedPlan)}>取消计划</UiButton>
                  </FormActions>
                  <div className="production-detail-grid detail-table-grid">
                    <div>
                      <h4>生产任务</h4>
                      <SimpleTable
                        data={selectedPlanTasks}
                        rowKey={(item) => item.id}
                        columns={[
                          { key: "taskNo", title: "任务", render: (item) => item.taskNo },
                          { key: "quantity", title: "计划/已产", render: (item) => `${qty(item.planQty)} / ${qty(item.producedQty)}` },
                          { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                          {
                            key: "action",
                            title: "操作",
                            align: "right",
                            width: "72px",
                            render: (item) => (
                              <UiButton
                                size="sm"
                                icon={<Plus size={12} />}
                                disabled={actionBusy !== "" || item.status === "completed" || item.status === "cancelled" || productionTaskRemaining(item) <= 0}
                                onClick={() => openProductionDialog("batch", selectedPlan, item)}
                              >
                                登记
                              </UiButton>
                            )
                          }
                        ]}
                        emptyText="暂无任务"
                      />
                    </div>
                    <div>
                      <h4>生产批次</h4>
                      <SimpleTable
                        data={selectedPlanBatches}
                        rowKey={(item) => item.id}
                        columns={[
                          { key: "batchNo", title: "批次", render: (item) => item.batchNo },
                          { key: "quantity", title: "方量", render: (item) => qty(item.quantity) },
                          { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> }
                        ]}
                        emptyText="暂无批次"
                      />
                    </div>
                  </div>
                  {selectedPlanReports.length ? (
                    <MetricList compact className="detail-summary-grid detail-report-grid">
                      {selectedPlanReports.map((item) => (
                        <div key={item.id}><span>{item.reportNo}</span><b>{qty(item.producedQty)} / {item.batchCount} 批</b></div>
                      ))}
                    </MetricList>
                  ) : null}
                  <FormActions className="detail-footer">
                    <UiButton disabled={actionBusy !== ""} onClick={close}>关闭</UiButton>
                  </FormActions>
                </>
              ) : (
                null
              )}
            </DialogContent>
          );
        case "create-plan":
          return (
            <DialogForm onSubmit={handleCreateProductionPlan}>
              <Field label="订单">
	                <SelectInput
	                  value={productionForm.orderId}
	                  disabled={!openOrders.length}
	                  onChange={(event) => {
                    const order = openOrders.find((item) => String(item.id) === event.target.value);
                    const planDate = productionOrderPlanDate(order);
                    const plant = preferredProductionPlantForOrder(order, planDate);
                    setProductionForm((value) => ({
                      ...value,
                      orderId: event.target.value,
                      plantId: plant ? String(plant.id) : "",
                      planDate,
                      planQuantity: order ? String(productionOrderRemaining(order, plans)) : ""
                    }));
                  }}
                >
                  {openOrders.map((item) => (
                    <option key={item.id} value={item.id}>{item.orderNo} · {productLabel(bootstrap, item.productId)} · 未计划 {qty(productionOrderRemaining(item, plans))}</option>
                  ))}
	                </SelectInput>
	              </Field>
	              {!openOrders.length ? <p className="muted detail-alert span-all">当前没有待计划订单，请先确认订单状态或切换站点。</p> : null}
	              <Field label="生产线">
	                <SelectInput
	                  value={productionForm.plantId || (selectedPlant ? String(selectedPlant.id) : "")}
	                  onChange={(event) => setProductionForm((value) => ({ ...value, plantId: event.target.value }))}
	                  disabled={!selectedOrder || !selectedPlantOptions.length}
	                >
                  {selectedPlantOptions.map((item) => (
                    <option key={item.id} value={item.id}>{productionPlantLabel(item)}</option>
                  ))}
	                </SelectInput>
	              </Field>
	              {selectedOrder && !selectedPlantOptions.length ? <p className="muted detail-alert span-all">当前订单站点没有可用生产线，请先维护生产线后再生成计划。</p> : null}
	              <Field label="计划方量"><TextInput type="number" min="0" max={selectedOrderRemaining || undefined} step="0.5" value={productionForm.planQuantity} disabled={!selectedOrder} onChange={(event) => setProductionForm((value) => ({ ...value, planQuantity: event.target.value }))} /></Field>
              <HeroDateField label="计划日期" value={productionForm.planDate} disabled={!selectedOrder} onChange={(planDate) => setProductionForm((value) => ({ ...value, planDate }))} />
              <Field label="班次">
                <SelectInput value={productionForm.shift} disabled={!selectedOrder} onChange={(event) => setProductionForm((value) => ({ ...value, shift: event.target.value }))}>
                  {shiftOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              <FormActions spanAll>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !selectedOrder || !selectedPlantOptions.length || planQuantityValue <= 0 || planQuantityValue > selectedOrderRemaining}>生成计划</UiButton>
              </FormActions>
            </DialogForm>
          );
        case "edit-plan":
          return (
            <DialogForm onSubmit={handleUpdateProductionPlan}>
              <Field label="生产计划">
	                <SelectInput
	                  value={productionForm.planId}
	                  disabled={!plans.length}
	                  onChange={(event) => {
	                    const plan = plans.find((item) => String(item.id) === event.target.value);
	                    const plant = productionPlanPlant(plan);
                    setProductionForm((value) => ({
                      ...value,
                      planId: event.target.value,
                      plantId: plant ? String(plant.id) : "",
                      planDate: plan?.planDate || value.planDate,
                      shift: plan?.shift || value.shift,
                      adjustPlanQuantity: plan ? String(plan.planQuantity) : value.adjustPlanQuantity,
                      taskQty: plan ? String(productionPlanTaskRemaining(plan)) : value.taskQty
                    }));
                  }}
                >
                  {plans.map((item) => <option key={item.id} value={item.id}>{item.planNo} · {productLabel(bootstrap, item.productId)}</option>)}
	                </SelectInput>
	              </Field>
	              {!plans.length ? <p className="muted detail-alert span-all">当前没有可调整的生产计划，请先生成计划。</p> : null}
	              <Field label="生产线">
	                <SelectInput
                  value={productionForm.plantId || (selectedPlant ? String(selectedPlant.id) : "")}
                  onChange={(event) => setProductionForm((value) => ({ ...value, plantId: event.target.value }))}
                  disabled={!selectedPlantOptions.length || !selectedPlan || selectedPlan.producedQty > 0 || selectedPlan.plannedTaskQty > 0}
                >
                  {selectedPlantOptions.map((item) => (
                    <option key={item.id} value={item.id}>{productionPlantLabel(item)}</option>
                  ))}
	                </SelectInput>
	              </Field>
	              <Field label="计划方量"><TextInput type="number" min={selectedPlan?.producedQty || 0} step="0.5" value={productionForm.adjustPlanQuantity} disabled={!selectedPlan || selectedPlan.status === "completed" || selectedPlan.status === "cancelled"} onChange={(event) => setProductionForm((value) => ({ ...value, adjustPlanQuantity: event.target.value }))} /></Field>
	              <HeroDateField label="计划日期" value={productionForm.planDate} disabled={!selectedPlan || selectedPlan.status === "completed" || selectedPlan.status === "cancelled"} onChange={(planDate) => setProductionForm((value) => ({ ...value, planDate }))} />
	              <Field label="班次">
	                <SelectInput value={productionForm.shift} disabled={!selectedPlan || selectedPlan.status === "completed" || selectedPlan.status === "cancelled"} onChange={(event) => setProductionForm((value) => ({ ...value, shift: event.target.value }))}>
                  {shiftOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              {selectedPlan ? (
                <MetricList compact className="span-all">
                  <div><span>已生产</span><b>{qty(selectedPlan.producedQty)}</b></div>
                  <div><span>已下达任务</span><b>{qty(selectedPlan.plannedTaskQty)}</b></div>
                  <div><span>状态</span><b>{workflowStatusFor(["production_plan", "productionPlan"], selectedPlan.id, selectedPlan.planNo, <StatusChip value={selectedPlan.status} />)}</b></div>
                </MetricList>
              ) : null}
              <FormActions spanAll>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !selectedPlan || !selectedPlantOptions.length || adjustedPlanQuantityValue <= 0 || adjustedPlanQuantityValue < (selectedPlan?.producedQty || 0) || selectedPlan.status === "completed" || selectedPlan.status === "cancelled"}>保存调整</UiButton>
              </FormActions>
            </DialogForm>
          );
	        case "tasks":
	          return (
	            <DialogForm onSubmit={handleCreateProductionTask}>
	              <Field label="生产计划">
	                <SelectInput
	                  value={productionForm.planId}
	                  disabled={!taskSelectablePlans.length}
	                  onChange={(event) => {
	                    const plan = plans.find((item) => String(item.id) === event.target.value);
	                    const plant = productionPlanPlant(plan);
	                    setProductionForm((value) => ({
                      ...value,
                      planId: event.target.value,
                      plantId: plant ? String(plant.id) : value.plantId,
	                      taskQty: plan ? String(productionPlanTaskRemaining(plan)) : value.taskQty
	                    }));
	                  }}
	                >
	                  {taskSelectablePlans.map((item) => <option key={item.id} value={item.id}>{item.planNo} · 未下达 {qty(productionPlanTaskRemaining(item))}</option>)}
	                </SelectInput>
	              </Field>
	              {!taskSelectablePlans.length ? <p className="muted detail-alert span-all">当前没有可下达任务的生产计划，请先生成计划或处理配比、产能、库存检查。</p> : null}
	              <Field label="任务方量"><TextInput type="number" min="0" max={selectedPlanTaskRemaining || undefined} step="0.5" value={productionForm.taskQty} disabled={!productionPlanCanIssueTask(selectedPlan)} onChange={(event) => setProductionForm((value) => ({ ...value, taskQty: event.target.value }))} /></Field>
	              {selectedPlan ? (
	                <>
	                  <MetricList compact className="span-all">
	                    <div><span>生产线</span><b>{productionPlantLabel(productionPlanPlant(selectedPlan))}</b></div>
	                    <div><span>计划方量</span><b>{qty(selectedPlan.planQuantity)}</b></div>
	                    <div><span>未下达</span><b>{qty(selectedPlanTaskRemaining)}</b></div>
	                    <div><span>产能</span><b><StatusChip value={selectedPlan.capacityStatus} /></b></div>
	                    <div><span>库存</span><b><StatusChip value={selectedPlan.inventoryStatus} /></b></div>
	                    <div><span>配比匹配</span><b><StatusChip value={selectedPlan.recipeStatus} /></b></div>
                  </MetricList>
                  {selectedPlan.riskReason ? <p className="muted detail-alert span-all">{selectedPlan.riskReason}</p> : null}
                </>
              ) : null}
              <FormActions spanAll>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton icon={<Factory size={14} />} disabled={actionBusy !== "" || !productionPlanCanIssueTask(selectedPlan)} onClick={handleAutoProductionTasks}>一键下达剩余</UiButton>
                <UiButton variant="primary" type="submit" icon={<PlayCircle size={14} />} disabled={actionBusy !== "" || !productionPlanCanIssueTask(selectedPlan) || taskQtyValue <= 0 || taskQtyValue > selectedPlanTaskRemaining}>按方量下达</UiButton>
              </FormActions>
            </DialogForm>
          );
        case "batch":
          return (
            <DialogForm onSubmit={handleCreateProductionBatch}>
              <Field label="生产任务">
                <SelectInput
                  value={selectedTask ? String(selectedTask.id) : ""}
                  disabled={!batchSelectableTasks.length}
                  onChange={(event) => {
                    const task = batchSelectableTasks.find((item) => String(item.id) === event.target.value);
                    const plan = task ? plans.find((item) => item.id === task.planId) : undefined;
                    const plant = productionTaskPlant(task);
                    setProductionForm((value) => ({
                      ...value,
                      taskId: event.target.value,
                      planId: plan ? String(plan.id) : value.planId,
                      plantId: plant ? String(plant.id) : value.plantId,
                      batchQty: task ? String(productionTaskRemaining(task)) : value.batchQty
                    }));
                  }}
                >
                  {batchSelectableTasks.map((item) => <option key={item.id} value={item.id}>{item.taskNo} · {productLabel(bootstrap, item.productId)} · 未生产 {qty(productionTaskRemaining(item))}</option>)}
                </SelectInput>
              </Field>
              {!batchSelectableTasks.length ? <p className="muted detail-alert span-all">当前计划没有可登记的生产任务，请先下达任务或选择其他计划。</p> : null}
              <Field label="批次数量"><TextInput type="number" min="0" max={selectedTaskRemaining || undefined} step="0.5" value={productionForm.batchQty} disabled={!selectedTask} onChange={(event) => setProductionForm((value) => ({ ...value, batchQty: event.target.value }))} /></Field>
              <Field label="质量状态">
                <SelectInput value={productionForm.batchQuality} disabled={!selectedTask} onChange={(event) => setProductionForm((value) => ({ ...value, batchQuality: event.target.value }))}>
                  {qualityStatusOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              {selectedTask ? (
                <MetricList compact className="span-all">
                  <div><span>生产计划</span><b>{selectedTaskPlan?.planNo || selectedTask.planId}</b></div>
                  <div><span>生产线</span><b>{productionPlantLabel(productionTaskPlant(selectedTask))}</b></div>
                  <div><span>任务方量</span><b>{qty(selectedTask.planQty)}</b></div>
                  <div><span>已生产</span><b>{qty(selectedTask.producedQty)}</b></div>
                  <div><span>未生产</span><b>{qty(selectedTaskRemaining)}</b></div>
                </MetricList>
              ) : null}
              <FormActions spanAll>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !selectedTask || selectedTaskRemaining <= 0 || batchQtyValue <= 0 || batchQtyValue > selectedTaskRemaining}>登记批次</UiButton>
              </FormActions>
            </DialogForm>
          );
        case "report":
          return (
            <DialogContent className="dialog-form">
              {selectedPlan ? (
                <MetricList compact>
                  <div><span>站点</span><b>{nameOf(bootstrap?.sites, selectedPlan.siteId)}</b></div>
                  <div><span>日期</span><b>{selectedPlan.planDate}</b></div>
                  <div><span>计划数</span><b>{selectedReportPlans.length} 个</b></div>
                  <div><span>计划/已产</span><b>{qty(selectedReportPlannedQty)} / {qty(selectedReportProducedQty)}</b></div>
                  <div><span>批次</span><b>{selectedReportBatches.length} 批</b></div>
                  <div><span>合格</span><b>{selectedReportPassedBatchCount} 批</b></div>
                  <div><span>待处理</span><b>{selectedReportPendingBatchCount} 批</b></div>
                </MetricList>
              ) : null}
              {selectedReportPendingBatchCount > 0 ? <p className="muted detail-alert">存在待检或异常批次，日报会记录为待处理质量。</p> : null}
              <FormActions>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton variant="primary" icon={<ClipboardCheck size={14} />} disabled={actionBusy !== "" || !productionPlanNeedsReport(selectedPlan)} onClick={handleGenerateProductionReport}>生成日报</UiButton>
              </FormActions>
            </DialogContent>
          );
        case "cancel-plan":
          return (
            <DialogContent className="dialog-form">
	              {selectedPlan ? (
	                <>
	                  <MetricList compact>
	                    <div><span>计划</span><b>{selectedPlan.planNo}</b></div>
	                    <div><span>已生产</span><b>{qty(selectedPlan.producedQty)}</b></div>
	                    <div><span>状态</span><b>{workflowStatusFor(["production_plan", "productionPlan"], selectedPlan.id, selectedPlan.planNo, <StatusChip value={selectedPlan.status} />)}</b></div>
	                  </MetricList>
	                  {workflowTimelineBlock(["production_plan", "productionPlan"], selectedPlan.id, selectedPlan.planNo, "当前生产计划暂无工作流实例")}
	                  {productionPlanCancelReason(selectedPlan) ? <p className="muted detail-alert">{productionPlanCancelReason(selectedPlan)}</p> : null}
	                </>
	              ) : null}
              <FormActions>
                <UiButton disabled={actionBusy !== ""} onClick={close}>取消</UiButton>
                <UiButton variant="danger" icon={<X size={14} />} disabled={actionBusy !== "" || !productionPlanCanCancel(selectedPlan)} onClick={handleCancelProductionPlan}>确认取消</UiButton>
              </FormActions>
            </DialogContent>
          );
      }
    }

    return (
      <Dialog
        open
        title={title}
        size={productionDialogMode === "detail" ? "xl" : "lg"}
        className={`production-dialog ${productionDialogMode === "detail" ? "detail-dialog production-detail-modal" : ""}`}
        closeDisabled={actionBusy !== ""}
        onClose={close}
      >
        {renderProductionDialogBody()}
      </Dialog>
    );
  }

  function renderOverview() {
    const operating = data.dashboard?.operating;
    const orderCount = selectedSiteId ? scopedOrders.length : operating?.orderCount || scopedOrders.length;
    const plannedQty = selectedSiteId ? scopedOrders.reduce((sum, item) => sum + item.planQuantity, 0) : operating?.plannedQty;
    const signedQty = selectedSiteId ? scopedOrders.reduce((sum, item) => sum + item.signedQty, 0) : operating?.signedQty;
    const receivable = operating?.receivableBalance || receivableBalance(data.finance);
    const dispatchKpis = data.dispatch?.kpis;
    const workflowInboxItems = workflowInbox();
    const overdueWorkflowInboxItems = workflowInboxItems.filter((item) => item.overdue);

    function renderWorkflowInboxItem(item: WorkflowInboxItem) {
      return (
        <div className="workflow-business-block" key={item.task.id}>
          <SplitRow>
            <div>
              <b>{item.instance.title || item.instance.resourceNo || item.task.taskNo}</b>
              <span className="block-text muted">{workflowResourceLabel(item.instance.resource)} / {item.instance.definitionName || item.instance.definitionCode}</span>
            </div>
            <StatusChip value={item.overdue ? "failed" : item.task.status} />
          </SplitRow>
          <span>节点：{item.task.stepName || item.task.stepCode} / {roleName(item.task.roleCode)}</span>
          <span>{item.task.dueAt ? `到期：${item.task.dueAt}` : `创建：${item.task.createdAt}`}</span>
          <ActionDialog id={`workflow-inbox-${item.task.id}`} title="处理工作流" buttonLabel="处理" triggerIcon={<CheckCircle2 size={13} />}>
            <div className="finance-hidden-actions">
              <div className="finance-action-block">
                <b>{item.instance.instanceNo}</b>
                <span>{item.instance.reason || item.instance.title || "-"}</span>
                <span>{workflowResourceLabel(item.instance.resource)} / {item.instance.resourceNo || `#${item.instance.resourceId}`}</span>
              </div>
              <Field label="审批意见">
                <TextInput value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
              </Field>
              <ActionGroup className="compact-actions">
                <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => handleBusinessWorkflowTaskAction(item.task, "approve")}>通过</UiButton>
                <UiButton disabled={actionBusy !== ""} onClick={() => handleBusinessWorkflowTaskAction(item.task, "reject")}>驳回</UiButton>
              </ActionGroup>
            </div>
          </ActionDialog>
        </div>
      );
    }

    return (
      <>
        <LayoutRegion as="section" className="kpi-grid">
          <KpiCard label="订单数" value={orderCount || 0} />
          <KpiCard label="计划吨位" value={qty(plannedQty)} suffix="t" />
          <KpiCard label="签收吨位" value={qty(signedQty)} suffix="t" />
          <KpiCard label="应收余额" value={money(receivable)} suffix="元" />
          <KpiCard label="毛利率" value={percent(operating?.grossMargin)} />
          <KpiCard label="流程待办" value={workflowInboxItems.length} />
        </LayoutRegion>

        <SectionGrid>
          <Panel as="div" className="span-7">{ordersTable(scopedOrders.slice(0, 8), false)}</Panel>
          <Panel as="div" className="span-5">
            <SplitRow>
              <div>
                <h3>调度摘要</h3>
              </div>
              <Truck size={20} />
            </SplitRow>
            <MetricList compact>
              <div><span>在线车辆</span><b>{dispatchKpis?.onlineVehicles || 0}/{dispatchKpis?.totalVehicles || 0}</b></div>
              <div><span>排队车辆</span><b>{dispatchKpis?.queueVehicles || 0}</b></div>
              <div><span>装料中</span><b>{dispatchKpis?.loadingVehicles || 0}</b></div>
              <div><span>运输中</span><b>{dispatchKpis?.inTransitVehicles || 0}</b></div>
              <div><span>活跃派车</span><b>{dispatchKpis?.activeDispatches || moduleDispatchQueue.length}</b></div>
	            </MetricList>
		        </Panel>
          <Panel as="div" className="span-12">
            <SplitRow>
              <div>
                <h3>我的工作流待办</h3>
                <span className="muted">当前角色可处理 {workflowInboxItems.length} 条，逾期 {overdueWorkflowInboxItems.length} 条</span>
              </div>
              <ListChecks size={20} />
            </SplitRow>
            <div className="workflow-business-log">
              {workflowInboxItems.slice(0, 6).map(renderWorkflowInboxItem)}
              {!workflowInboxItems.length ? <span>暂无待处理工作流</span> : null}
            </div>
          </Panel>
			      </SectionGrid>
      </>
    );
  }

  function renderMasterCustomers() {
    const customers = bootstrap?.customers || [];
    const contacts = list(bootstrap?.customerContacts);
    const profiles = list(bootstrap?.customerProfiles);
    const blacklists = list(bootstrap?.customerBlacklists);
    const complaints = list(bootstrap?.customerComplaints);
    const projects = list(bootstrap?.projects);
    const contactFormView = (
      <DialogForm onSubmit={handleSaveCustomerContact}>
        <Field label="客户">
          <SelectInput value={customerContactForm.customerId} onChange={(event) => setCustomerContactForm({ ...customerContactForm, customerId: event.target.value })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="姓名"><TextInput value={customerContactForm.name} onChange={(event) => setCustomerContactForm({ ...customerContactForm, name: event.target.value })} /></Field>
        <Field label="电话"><TextInput value={customerContactForm.phone} onChange={(event) => setCustomerContactForm({ ...customerContactForm, phone: event.target.value })} /></Field>
        <Field label="角色"><TextInput value={customerContactForm.role} onChange={(event) => setCustomerContactForm({ ...customerContactForm, role: event.target.value })} /></Field>
        <Field label="默认联系人">
          <SelectInput value={customerContactForm.isDefault} onChange={(event) => setCustomerContactForm({ ...customerContactForm, isDefault: event.target.value })}>
            <option value="true">是</option>
            <option value="false">否</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={customerContactForm.id ? <CheckCircle2 size={14} /> : <Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(customerContactForm.customerId) || !customerContactForm.name.trim() || !customerContactForm.phone.trim()}>{customerContactForm.id ? "保存修改" : "保存联系人"}</UiButton>
        </FormActions>
      </DialogForm>
    );
    const profileFormView = (
      <DialogForm onSubmit={handleCreateCustomerProfile}>
        <Field label="客户">
          <SelectInput value={customerProfileForm.customerId} onChange={(event) => setCustomerProfileForm({ ...customerProfileForm, customerId: event.target.value })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="等级"><TextInput value={customerProfileForm.grade} onChange={(event) => setCustomerProfileForm({ ...customerProfileForm, grade: event.target.value })} /></Field>
        <Field label="风险">
          <SelectInput value={customerProfileForm.riskLevel} onChange={(event) => setCustomerProfileForm({ ...customerProfileForm, riskLevel: event.target.value })}>
            <option value="low">low</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
          </SelectInput>
        </Field>
        <Field label="信用分"><TextInput type="number" min="0" max="100" value={customerProfileForm.creditScore} onChange={(event) => setCustomerProfileForm({ ...customerProfileForm, creditScore: event.target.value })} /></Field>
        <Field label="标签" spanAll><TextAreaInput value={customerProfileForm.tags} onChange={(event) => setCustomerProfileForm({ ...customerProfileForm, tags: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !fieldNumber(customerProfileForm.customerId)}>保存档案</UiButton>
        </FormActions>
      </DialogForm>
    );
    const blacklistFormView = (
      <DialogForm onSubmit={handleCreateCustomerBlacklist}>
        <Field label="客户">
          <SelectInput value={customerBlacklistForm.customerId} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, customerId: event.target.value })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="范围">
          <SelectInput value={customerBlacklistForm.scope} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, scope: event.target.value })}>
            <option value="sales_order">订单</option>
            <option value="dispatch">调度</option>
            <option value="all">全部</option>
          </SelectInput>
        </Field>
        <Field label="等级">
          <SelectInput value={customerBlacklistForm.severity} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, severity: event.target.value })}>
            <option value="high">high</option>
            <option value="medium">medium</option>
            <option value="low">low</option>
          </SelectInput>
        </Field>
        <Field label="阻断订单">
          <SelectInput value={customerBlacklistForm.blockOrders} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, blockOrders: event.target.value })}>
            <option value="true">是</option>
            <option value="false">否</option>
          </SelectInput>
        </Field>
        <Field label="阻断调度">
          <SelectInput value={customerBlacklistForm.blockDispatch} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, blockDispatch: event.target.value })}>
            <option value="false">否</option>
            <option value="true">是</option>
          </SelectInput>
        </Field>
        <Field label="原因" spanAll><TextAreaInput value={customerBlacklistForm.reason} onChange={(event) => setCustomerBlacklistForm({ ...customerBlacklistForm, reason: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<AlertCircle size={14} />} disabled={actionBusy !== "" || !fieldNumber(customerBlacklistForm.customerId) || !customerBlacklistForm.reason.trim()}>提交黑名单</UiButton>
        </FormActions>
      </DialogForm>
    );
    const complaintFormView = (
      <DialogForm onSubmit={handleCreateCustomerComplaint}>
        <Field label="客户">
          <SelectInput value={customerComplaintForm.customerId} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, customerId: event.target.value, projectId: "" })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="项目">
          <SelectInput value={customerComplaintForm.projectId} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, projectId: event.target.value })}>
            <option value="">不关联</option>
            {projects.filter((project) => !customerComplaintForm.customerId || String(project.customerId) === customerComplaintForm.customerId).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="标题"><TextInput value={customerComplaintForm.title} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, title: event.target.value })} /></Field>
        <Field label="等级">
          <SelectInput value={customerComplaintForm.level} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, level: event.target.value })}>
            <option value="critical">critical</option>
            <option value="high">high</option>
            <option value="medium">medium</option>
            <option value="low">low</option>
          </SelectInput>
        </Field>
        <Field label="负责人"><TextInput value={customerComplaintForm.owner} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, owner: event.target.value })} /></Field>
        <Field label="SLA 小时"><TextInput type="number" min="1" value={customerComplaintForm.slaHours} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, slaHours: event.target.value })} /></Field>
        <Field label="内容" spanAll><TextAreaInput value={customerComplaintForm.content} onChange={(event) => setCustomerComplaintForm({ ...customerComplaintForm, content: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(customerComplaintForm.customerId) || !customerComplaintForm.title.trim()}>创建投诉</UiButton>
        </FormActions>
      </DialogForm>
    );
    if (section === "customer-risk") {
      const activeBlacklists = blacklists.filter((item) => item.status !== "released");
      const openComplaints = complaints.filter((item) => item.status !== "closed");
      return (
        <Panel className="master-customers-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>客户风控</b>
              <span>{profiles.length} 份风险档案 / {activeBlacklists.length} 个黑名单 / {openComplaints.length} 个未关闭投诉</span>
            </div>
            <ActionGroup>
              <ActionDialog id="customer-profile-create-page" title="维护客户风险档案" buttonLabel="维护档案" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={() => resetCustomerDomainForms()}>
                {profileFormView}
              </ActionDialog>
              <ActionDialog id="customer-blacklist-create-page" title="提交客户黑名单" buttonLabel="提交黑名单" triggerIcon={<AlertCircle size={13} />} onOpen={() => resetCustomerDomainForms()}>
                {blacklistFormView}
              </ActionDialog>
              <ActionDialog id="customer-complaint-create-page" title="创建客户投诉" buttonLabel="创建投诉" triggerIcon={<Plus size={13} />} onOpen={() => resetCustomerDomainForms()}>
                {complaintFormView}
              </ActionDialog>
              <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={handleEvaluateCustomerProfiles}>重算风险</UiButton>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>客户</span><b>{customers.length}</b></div>
            <div><span>风险档案</span><b>{profiles.length}</b></div>
            <div><span>高风险</span><b>{profiles.filter((item) => item.riskLevel === "high").length}</b></div>
            <div><span>黑名单</span><b>{activeBlacklists.length}</b></div>
            <div><span>未关闭投诉</span><b>{openComplaints.length}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="客户风险档案"
              data={profiles}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "customer", title: "客户", render: (item) => item.customerName || nameOf(customers, item.customerId) },
                { key: "grade", title: "等级", render: (item) => <><b>{item.grade}</b><span className="block-text muted">{item.creditScore}</span></> },
                { key: "risk", title: "风险", render: (item) => <StatusChip value={item.riskLevel} /> },
                { key: "tags", title: "标签", render: (item) => <ChipList compact>{list(item.tags).slice(0, 4).map((tag) => <span key={tag}>{tag}</span>)}</ChipList> },
                { key: "updated", title: "更新时间", render: (item) => shortDateTime(item.updatedAt) },
                { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
              ]}
              emptyText="暂无客户风险档案"
            />
            <DataTable
              title="客户黑名单"
              data={blacklists}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "customer", title: "客户", render: (item) => item.customerName || nameOf(customers, item.customerId) },
                { key: "reason", title: "原因", render: (item) => item.reason || "-" },
                { key: "scope", title: "范围", render: (item) => item.scope || "-" },
                { key: "severity", title: "等级", render: (item) => <StatusChip value={item.severity} /> },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["customer_blacklist"], item.id, item.customerName || String(item.customerId), <StatusChip value={item.status} />) },
                { key: "created", title: "创建", render: (item) => shortDateTime(item.createdAt) },
                { key: "actions", title: "操作", width: "120px", render: (item: CustomerBlacklist) => <UiButton size="sm" disabled={actionBusy !== "" || item.status === "released"} onClick={() => handleReleaseCustomerBlacklist(item)}>解除</UiButton> }
              ]}
              emptyText="暂无客户黑名单"
            />
          </SectionGrid>
          <DataTable
            title="客户投诉"
            data={complaints}
            rowKey={(item) => item.id}
            pageSize={10}
            onRefresh={refreshData}
            columns={[
              { key: "complaint", title: "投诉", render: (item) => <><b>{item.complaintNo}</b><span className="block-text muted">{item.title}</span></> },
              { key: "customer", title: "客户", render: (item) => nameOf(customers, item.customerId) },
              { key: "level", title: "等级", render: (item) => <StatusChip value={item.level} /> },
              { key: "sla", title: "SLA", render: (item) => <><StatusChip value={item.slaStatus} /><span className="block-text muted">{shortDateTime(item.dueAt)}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "120px",
                render: (item: CustomerComplaint) => item.status === "closed" ? <span className="muted">{item.resolution || "已关闭"}</span> : (
                  <ActionDialog id={`customer-complaint-close-page-${item.id}`} title="关闭客户投诉" buttonLabel="关闭">
                    <InlineForm onSubmit={(event) => handleCloseCustomerComplaint(event, item)}>
                      <Field label="处理结果"><TextInput name="resolution" defaultValue={item.resolution || ""} /></Field>
                      <UiButton type="submit" variant="primary" disabled={actionBusy !== ""}>关闭投诉</UiButton>
                    </InlineForm>
                  </ActionDialog>
                )
              }
            ]}
            emptyText="暂无客户投诉"
          />
        </Panel>
      );
    }
    return (
      <Panel className="master-customers-panel">
        <DataTable
          data={customers}
          rowKey={(item) => item.id}
          emptyText="暂无客户"
          pageSize={8}
          onRefresh={refreshData}
          columnSettingsKey="master-customers"
          columnSettingsLabel="列设置"
          rowContextMenu={buildDataTableRowContextMenu<Customer>({
            actions: [
              { key: "focus-customer", label: "只看该客户", onSelect: (item, helpers) => helpers.searchText(item.name) }
            ],
            copyFields: [
              { key: "name", label: "客户名称", value: (item) => item.name },
              { key: "contact", label: "联系人", value: (item) => item.contact },
              { key: "phone", label: "电话", value: (item) => item.phone },
              { key: "credit", label: "授信", value: (item) => money(item.creditLimit) }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("customer")}>新增客户</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "客户", render: (item) => item.name },
            { key: "contact", title: "联系人", render: (item) => item.contact },
            { key: "phone", title: "电话", render: (item) => item.phone },
            { key: "creditLimit", title: "授信", render: (item) => money(item.creditLimit) },
            { key: "receivable", title: "应收", render: (item) => money(item.receivable) },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", locked: true, render: (item) => masterRowActions("customer", item.id, item) }
          ]}
        />
        <SectionGrid className="finance-list-page">
          <DataTable
            title="客户联系人"
            data={contacts}
            rowKey={(item) => item.id}
            pageSize={6}
            headerLeftAction={(
              <ActionDialog id="customer-contact-create" title="新增客户联系人" buttonLabel="新增联系人" triggerIcon={<Plus size={13} />} onOpen={() => resetCustomerDomainForms()}>
                {contactFormView}
              </ActionDialog>
            )}
            columns={[
              { key: "customer", title: "客户", render: (item) => nameOf(customers, item.customerId) },
              { key: "name", title: "联系人", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.role}</span></> },
              { key: "phone", title: "电话", render: (item) => item.phone },
              { key: "default", title: "默认", render: (item) => <StatusChip value={item.isDefault ? "active" : "disabled"} /> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "220px",
                render: (item: CustomerContact) => (
                  <ActionGroup className="compact-actions">
                    <UiButton size="sm" disabled={actionBusy !== "" || item.isDefault || item.status !== "active"} onClick={() => handleSetDefaultCustomerContact(item)}>设默认</UiButton>
                    <ActionDialog id={`customer-contact-edit-${item.id}`} title="编辑客户联系人" buttonLabel="编辑" onOpen={() => startEditCustomerContact(item)}>
                      {contactFormView}
                    </ActionDialog>
                    <UiButton size="sm" variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeleteCustomerContact(item)}>删除</UiButton>
                  </ActionGroup>
                )
              }
            ]}
            emptyText="暂无客户联系人"
          />
          <DataTable
            title="客户风险档案"
            data={profiles}
            rowKey={(item) => item.id}
            pageSize={6}
            headerLeftAction={(
              <ActionGroup>
                <ActionDialog id="customer-profile-create" title="维护客户风险档案" buttonLabel="维护档案" triggerIcon={<Plus size={13} />} onOpen={() => resetCustomerDomainForms()}>
                  {profileFormView}
                </ActionDialog>
                <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={handleEvaluateCustomerProfiles}>重算风险</UiButton>
              </ActionGroup>
            )}
            columns={[
              { key: "customer", title: "客户", render: (item) => item.customerName || nameOf(customers, item.customerId) },
              { key: "grade", title: "等级", render: (item) => <><b>{item.grade}</b><span className="block-text muted">{item.creditScore}</span></> },
              { key: "risk", title: "风险", render: (item) => <StatusChip value={item.riskLevel} /> },
              { key: "tags", title: "标签", render: (item) => <ChipList compact>{list(item.tags).slice(0, 4).map((tag) => <span key={tag}>{tag}</span>)}</ChipList> },
              { key: "updated", title: "更新时间", render: (item) => shortDateTime(item.updatedAt) },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
            ]}
            emptyText="暂无客户风险档案"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="客户黑名单"
            data={blacklists}
            rowKey={(item) => item.id}
            pageSize={6}
            headerLeftAction={(
              <ActionDialog id="customer-blacklist-create" title="提交客户黑名单" buttonLabel="提交黑名单" triggerIcon={<AlertCircle size={13} />} onOpen={() => resetCustomerDomainForms()}>
                {blacklistFormView}
              </ActionDialog>
            )}
            columns={[
              { key: "customer", title: "客户", render: (item) => item.customerName || nameOf(customers, item.customerId) },
              { key: "reason", title: "原因", render: (item) => item.reason || "-" },
              { key: "scope", title: "范围", render: (item) => item.scope || "-" },
              { key: "severity", title: "等级", render: (item) => <StatusChip value={item.severity} /> },
              { key: "status", title: "状态", render: (item) => workflowStatusFor(["customer_blacklist"], item.id, item.customerName || String(item.customerId), <StatusChip value={item.status} />) },
              { key: "created", title: "创建", render: (item) => shortDateTime(item.createdAt) },
              { key: "actions", title: "操作", width: "120px", render: (item: CustomerBlacklist) => <UiButton size="sm" disabled={actionBusy !== "" || item.status === "released"} onClick={() => handleReleaseCustomerBlacklist(item)}>解除</UiButton> }
            ]}
            emptyText="暂无客户黑名单"
          />
          <DataTable
            title="客户投诉"
            data={complaints}
            rowKey={(item) => item.id}
            pageSize={6}
            headerLeftAction={(
              <ActionDialog id="customer-complaint-create" title="创建客户投诉" buttonLabel="创建投诉" triggerIcon={<Plus size={13} />} onOpen={() => resetCustomerDomainForms()}>
                {complaintFormView}
              </ActionDialog>
            )}
            columns={[
              { key: "complaint", title: "投诉", render: (item) => <><b>{item.complaintNo}</b><span className="block-text muted">{item.title}</span></> },
              { key: "customer", title: "客户", render: (item) => nameOf(customers, item.customerId) },
              { key: "level", title: "等级", render: (item) => <StatusChip value={item.level} /> },
              { key: "sla", title: "SLA", render: (item) => <><StatusChip value={item.slaStatus} /><span className="block-text muted">{shortDateTime(item.dueAt)}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "120px",
                render: (item: CustomerComplaint) => item.status === "closed" ? <span className="muted">{item.resolution || "已关闭"}</span> : (
                  <ActionDialog id={`customer-complaint-close-${item.id}`} title="关闭客户投诉" buttonLabel="关闭">
                    <InlineForm onSubmit={(event) => handleCloseCustomerComplaint(event, item)}>
                      <Field label="处理结果"><TextInput name="resolution" defaultValue={item.resolution || ""} /></Field>
                      <UiButton type="submit" variant="primary" disabled={actionBusy !== ""}>关闭投诉</UiButton>
                    </InlineForm>
                  </ActionDialog>
                )
              }
            ]}
            emptyText="暂无客户投诉"
          />
        </SectionGrid>
      </Panel>
    );
  }

  function renderMasterProjects() {
    const customers = bootstrap?.customers || [];
    const projects = bootstrap?.projects || [];
    return (
      <Panel>
        <DataTable
          data={projects}
          rowKey={(item) => item.id}
          emptyText="暂无项目"
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Project>({
            actions: [
              { key: "focus-project", label: "只看该项目", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-customer", label: "只看该客户", onSelect: (item, helpers) => helpers.searchText(nameOf(customers, item.customerId)) }
            ],
            copyFields: [
              { key: "project", label: "项目名称", value: (item) => item.name },
              { key: "customer", label: "客户", value: (item) => nameOf(customers, item.customerId) },
              { key: "address", label: "地址", value: (item) => item.address },
              { key: "phone", label: "电话", value: (item) => item.phone }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} disabled={!customers.length} onClick={() => openMasterCreateDialog("project")}>新增项目</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "项目", render: (item) => item.name },
            { key: "customer", title: "客户", render: (item) => nameOf(customers, item.customerId) },
            { key: "address", title: "地址", render: (item) => item.address },
            { key: "contact", title: "联系人", render: (item) => item.contact },
            { key: "phone", title: "电话", render: (item) => item.phone },
	            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("project", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderMasterProducts() {
    const products = bootstrap?.products || [];
    const customers = list(bootstrap?.customers);
    const projects = list(bootstrap?.projects);
    const taxRates = list(bootstrap?.taxRates);
    const pricePolicies = list(bootstrap?.pricePolicies);
    const masterBulkResources = ["customers", "products", "materials"];
    const taxRateFormView = (
      <DialogForm onSubmit={handleCreateTaxRate}>
        <Field label="名称"><TextInput value={taxRateForm.name} onChange={(event) => setTaxRateForm({ ...taxRateForm, name: event.target.value })} /></Field>
        <Field label="税率"><TextInput type="number" min="0" step="0.01" value={taxRateForm.rate} onChange={(event) => setTaxRateForm({ ...taxRateForm, rate: event.target.value })} /></Field>
        <Field label="范围"><TextInput value={taxRateForm.scope} onChange={(event) => setTaxRateForm({ ...taxRateForm, scope: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={taxRateForm.status} onChange={(event) => setTaxRateForm({ ...taxRateForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !taxRateForm.name.trim()}>{taxRateForm.id ? "保存税率" : "创建税率"}</UiButton>
        </FormActions>
      </DialogForm>
    );
    const pricePolicyFormView = (
      <DialogForm onSubmit={handleCreatePricePolicy}>
        <Field label="客户">
          <SelectInput value={pricePolicyForm.customerId} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, customerId: event.target.value, projectId: "" })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="项目">
          <SelectInput value={pricePolicyForm.projectId} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, projectId: event.target.value })}>
            <option value="">全部项目</option>
            {projects.filter((project) => !pricePolicyForm.customerId || String(project.customerId) === pricePolicyForm.customerId).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="产品">
          <SelectInput value={pricePolicyForm.productId} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, productId: event.target.value })}>
            {products.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.spec}</option>)}
          </SelectInput>
        </Field>
        <Field label="客户等级"><TextInput value={pricePolicyForm.customerGrade} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, customerGrade: event.target.value })} /></Field>
        <Field label="区域"><TextInput value={pricePolicyForm.region} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, region: event.target.value })} /></Field>
        <Field label="最低量"><TextInput type="number" min="0" value={pricePolicyForm.minQuantity} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, minQuantity: event.target.value })} /></Field>
        <Field label="最高量"><TextInput type="number" min="0" value={pricePolicyForm.maxQuantity} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, maxQuantity: event.target.value })} /></Field>
        <Field label="底价"><TextInput type="number" min="0" value={pricePolicyForm.floorPrice} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, floorPrice: event.target.value })} /></Field>
        <Field label="售价"><TextInput type="number" min="0" value={pricePolicyForm.salePrice} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, salePrice: event.target.value })} /></Field>
        <Field label="税率">
          <SelectInput value={pricePolicyForm.taxRateId} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, taxRateId: event.target.value })}>
            <option value="">不关联</option>
            {taxRates.map((item) => <option key={item.id} value={item.id}>{item.name} / {percent(item.rate * 100)}</option>)}
          </SelectInput>
        </Field>
        <Field label="优先级"><TextInput type="number" value={pricePolicyForm.priority} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, priority: event.target.value })} /></Field>
        <HeroDateField label="生效日期" value={pricePolicyForm.effectiveFrom} onChange={(effectiveFrom) => setPricePolicyForm({ ...pricePolicyForm, effectiveFrom })} />
        <HeroDateField label="失效日期" value={pricePolicyForm.effectiveTo} onChange={(effectiveTo) => setPricePolicyForm({ ...pricePolicyForm, effectiveTo })} />
        <Field label="促销名称"><TextInput value={pricePolicyForm.promotionName} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, promotionName: event.target.value })} /></Field>
        <Field label="促销类型"><TextInput value={pricePolicyForm.promotionType} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, promotionType: event.target.value })} /></Field>
        <Field label="促销值"><TextInput type="number" value={pricePolicyForm.promotionValue} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, promotionValue: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={pricePolicyForm.status} onChange={(event) => setPricePolicyForm({ ...pricePolicyForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(pricePolicyForm.productId) || !fieldNumber(pricePolicyForm.salePrice)}>{pricePolicyForm.id ? "保存价格政策" : "创建价格政策"}</UiButton>
        </FormActions>
      </DialogForm>
    );
    const pricingEvalFormView = (
      <DialogForm onSubmit={handleEvaluatePricing}>
        <Field label="客户">
          <SelectInput value={pricingEvalForm.customerId} onChange={(event) => setPricingEvalForm({ ...pricingEvalForm, customerId: event.target.value, projectId: "" })}>
            {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="项目">
          <SelectInput value={pricingEvalForm.projectId} onChange={(event) => setPricingEvalForm({ ...pricingEvalForm, projectId: event.target.value })}>
            <option value="">不关联</option>
            {projects.filter((project) => !pricingEvalForm.customerId || String(project.customerId) === pricingEvalForm.customerId).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="产品">
          <SelectInput value={pricingEvalForm.productId} onChange={(event) => setPricingEvalForm({ ...pricingEvalForm, productId: event.target.value })}>
            {products.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.spec}</option>)}
          </SelectInput>
        </Field>
        <HeroDateField label="计划时间" value={pricingEvalForm.planTime} onChange={(planTime) => setPricingEvalForm({ ...pricingEvalForm, planTime })} />
        <Field label="计划量"><TextInput type="number" min="0" value={pricingEvalForm.planQuantity} onChange={(event) => setPricingEvalForm({ ...pricingEvalForm, planQuantity: event.target.value })} /></Field>
        <Field label="输入单价"><TextInput type="number" min="0" value={pricingEvalForm.unitPrice} onChange={(event) => setPricingEvalForm({ ...pricingEvalForm, unitPrice: event.target.value })} /></Field>
        {pricingQuote ? (
          <MetricList compact className="span-all">
            <div><span>报价</span><b>{money(pricingQuote.unitPrice)}</b></div>
            <div><span>底价</span><b>{money(pricingQuote.floorPrice)}</b></div>
            <div><span>税率</span><b>{pricingQuote.taxRateName || "-"}</b></div>
            <div><span>审批</span><b>{pricingQuote.approvalRequired ? "需要" : "不需要"}</b></div>
          </MetricList>
        ) : null}
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Search size={14} />} disabled={actionBusy !== "" || !fieldNumber(pricingEvalForm.customerId) || !fieldNumber(pricingEvalForm.productId)}>评估报价</UiButton>
        </FormActions>
      </DialogForm>
    );
    const masterBulkFormView = (
      <DialogForm onSubmit={handleImportMasterData}>
        <Field label="资源">
          <SelectInput value={masterBulkForm.resource} onChange={(event) => setMasterBulkForm({ ...masterBulkForm, resource: event.target.value })}>
            {masterBulkResources.map((item) => <option key={item} value={item}>{item}</option>)}
          </SelectInput>
        </Field>
        <Field label="模式">
          <SelectInput value={masterBulkForm.mode} onChange={(event) => setMasterBulkForm({ ...masterBulkForm, mode: event.target.value })}>
            <option value="create">create</option>
            <option value="upsert">upsert</option>
          </SelectInput>
        </Field>
        <Field label="JSON 行数组" spanAll><TextAreaInput value={masterBulkForm.rowsJson} onChange={(event) => setMasterBulkForm({ ...masterBulkForm, rowsJson: event.target.value })} /></Field>
        {masterImportResult ? <span className="span-all muted">导入结果：新增 {masterImportResult.created} / 更新 {masterImportResult.updated} / 错误 {masterImportResult.errors.length}</span> : null}
        {masterExportResult ? <span className="span-all muted">最近导出：{masterExportResult.resource} / {masterExportResult.count} 行</span> : null}
        <FormActions spanAll>
          <UiButton disabled={actionBusy !== ""} onClick={() => handleExportMasterData(masterBulkForm.resource)}>导出</UiButton>
          <UiButton variant="primary" type="submit" icon={<ArrowUp size={14} />} disabled={actionBusy !== "" || !masterBulkForm.rowsJson.trim()}>导入</UiButton>
        </FormActions>
      </DialogForm>
    );
    if (section === "sales-pricing") {
      return (
        <Panel>
          <SectionHeader className="panel-head-compact">
            <div>
              <b>销售定价</b>
              <span>{pricePolicies.length} 条价格政策 / {taxRates.length} 个税率</span>
            </div>
            <ActionGroup>
              <ActionDialog id="sales-pricing-tax-rate-create" title="创建税率" buttonLabel="创建税率" triggerIcon={<Plus size={13} />} onOpen={() => setTaxRateForm({ id: "", name: "", rate: "0.06", scope: "sales", status: "active" })}>
                {taxRateFormView}
              </ActionDialog>
              <ActionDialog id="sales-pricing-policy-create" title="创建价格政策" buttonLabel="价格政策" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={() => resetPricingForms()}>
                {pricePolicyFormView}
              </ActionDialog>
              <ActionDialog id="sales-pricing-evaluate" title="价格评估" buttonLabel="价格评估" triggerIcon={<Search size={13} />} onOpen={() => resetPricingForms()}>
                {pricingEvalFormView}
              </ActionDialog>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>产品</span><b>{products.length}</b></div>
            <div><span>价格政策</span><b>{pricePolicies.length}</b></div>
            <div><span>启用政策</span><b>{pricePolicies.filter((item) => item.status === "active").length}</b></div>
            <div><span>税率</span><b>{taxRates.length}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="价格政策"
              data={pricePolicies}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "target", title: "适用", render: (item: PricePolicy) => <><b>{productLabel(bootstrap, item.productId)}</b><span className="block-text muted">{nameOf(customers, item.customerId) || item.customerGrade || "全部客户"}</span></> },
                { key: "qty", title: "数量", render: (item) => `${qty(item.minQuantity)} - ${item.maxQuantity ? qty(item.maxQuantity) : "不限"}` },
                { key: "price", title: "价格", render: (item) => <><span>{money(item.salePrice)}</span><span className="block-text muted">底价 {money(item.floorPrice)}</span></> },
                { key: "tax", title: "税率", render: (item) => taxRates.find((tax) => tax.id === item.taxRateId)?.name || "-" },
                { key: "effective", title: "有效期", render: (item) => `${item.effectiveFrom || "-"} / ${item.effectiveTo || "-"}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "160px",
                  render: (item: PricePolicy) => (
                    <ActionGroup>
                      <ActionDialog id={`sales-pricing-policy-edit-${item.id}`} title="编辑价格政策" buttonLabel="编辑" onOpen={() => startEditPricePolicy(item)}>
                        {pricePolicyFormView}
                      </ActionDialog>
                      <UiButton size="sm" variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeletePricePolicy(item)}>删除</UiButton>
                    </ActionGroup>
                  )
                }
              ]}
              emptyText="暂无价格政策"
            />
            <DataTable
              title="税率"
              data={taxRates}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "name", title: "名称", render: (item: TaxRate) => item.name },
                { key: "rate", title: "税率", render: (item) => percent(item.rate * 100) },
                { key: "scope", title: "范围", render: (item) => item.scope || "-" },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "140px",
                  render: (item: TaxRate) => (
                    <ActionGroup>
                      <ActionDialog id={`sales-pricing-tax-rate-edit-${item.id}`} title="编辑税率" buttonLabel="编辑" onOpen={() => startEditTaxRate(item)}>
                        {taxRateFormView}
                      </ActionDialog>
                      <UiButton size="sm" variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeleteTaxRate(item)}>删除</UiButton>
                    </ActionGroup>
                  )
                }
              ]}
              emptyText="暂无税率"
            />
          </SectionGrid>
        </Panel>
      );
    }
    return (
      <Panel>
        <DataTable
          data={products}
          rowKey={(item) => item.id}
          emptyText="暂无产品"
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Product>({
            actions: [
              { key: "focus-product", label: "只看该产品", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-spec", label: "只看同规格", onSelect: (item, helpers) => helpers.searchText(item.spec) }
            ],
            copyFields: [
              { key: "name", label: "产品名称", value: (item) => item.name },
              { key: "spec", label: "规格", value: (item) => item.spec },
              { key: "unit", label: "单位", value: (item) => item.unit },
              { key: "price", label: "基准价", value: (item) => money(item.basePrice) }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("product")}>新增产品</UiButton>
              <ActionDialog id="master-tax-rate-create" title="创建税率" buttonLabel="创建税率" triggerIcon={<Plus size={13} />} onOpen={() => setTaxRateForm({ id: "", name: "", rate: "0.06", scope: "sales", status: "active" })}>
                {taxRateFormView}
              </ActionDialog>
              <ActionDialog id="master-price-policy-create" title="创建价格政策" buttonLabel="价格政策" triggerIcon={<Plus size={13} />} onOpen={() => resetPricingForms()}>
                {pricePolicyFormView}
              </ActionDialog>
              <ActionDialog id="master-pricing-evaluate" title="价格评估" buttonLabel="价格评估" triggerIcon={<Search size={13} />} onOpen={() => resetPricingForms()}>
                {pricingEvalFormView}
              </ActionDialog>
              <ActionDialog id="master-bulk-tools" title="主数据导入导出" buttonLabel="导入导出" triggerIcon={<Download size={13} />}>
                {masterBulkFormView}
              </ActionDialog>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "产品", render: (item) => item.name },
            { key: "spec", title: "规格", render: (item) => item.spec },
            { key: "unit", title: "单位", render: (item) => item.unit },
            { key: "basePrice", title: "基准价", render: (item) => money(item.basePrice) },
            { key: "costPrice", title: "成本价", render: (item) => money(item.costPrice) },
	            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("product", item.id, item) }
          ]}
        />
        <SectionGrid className="finance-list-page">
          <DataTable
            title="税率"
            data={taxRates}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "name", title: "名称", render: (item: TaxRate) => item.name },
              { key: "rate", title: "税率", render: (item) => percent(item.rate * 100) },
              { key: "scope", title: "范围", render: (item) => item.scope || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "140px",
                render: (item: TaxRate) => (
                  <ActionGroup>
                    <ActionDialog id={`master-tax-rate-edit-${item.id}`} title="编辑税率" buttonLabel="编辑" onOpen={() => startEditTaxRate(item)}>
                      {taxRateFormView}
                    </ActionDialog>
                    <UiButton size="sm" variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeleteTaxRate(item)}>删除</UiButton>
                  </ActionGroup>
                )
              }
            ]}
            emptyText="暂无税率"
          />
          <DataTable
            title="价格政策"
            data={pricePolicies}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "target", title: "适用", render: (item: PricePolicy) => <><b>{productLabel(bootstrap, item.productId)}</b><span className="block-text muted">{nameOf(customers, item.customerId) || item.customerGrade || "全部客户"}</span></> },
              { key: "qty", title: "数量", render: (item) => `${qty(item.minQuantity)} - ${item.maxQuantity ? qty(item.maxQuantity) : "不限"}` },
              { key: "price", title: "价格", render: (item) => <><span>{money(item.salePrice)}</span><span className="block-text muted">底价 {money(item.floorPrice)}</span></> },
              { key: "tax", title: "税率", render: (item) => taxRates.find((tax) => tax.id === item.taxRateId)?.name || "-" },
              { key: "effective", title: "有效期", render: (item) => `${item.effectiveFrom || "-"} / ${item.effectiveTo || "-"}` },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "160px",
                render: (item: PricePolicy) => (
                  <ActionGroup>
                    <ActionDialog id={`master-price-policy-edit-${item.id}`} title="编辑价格政策" buttonLabel="编辑" onOpen={() => startEditPricePolicy(item)}>
                      {pricePolicyFormView}
                    </ActionDialog>
                    <UiButton size="sm" variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeletePricePolicy(item)}>删除</UiButton>
                  </ActionGroup>
                )
              }
            ]}
            emptyText="暂无价格政策"
          />
        </SectionGrid>
      </Panel>
    );
  }

  function renderMasterMaterials() {
    const materials = bootstrap?.materials || [];
    return (
      <Panel>
        <DataTable
          data={materials}
          rowKey={(item) => item.id}
          emptyText="暂无物料"
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Material>({
            actions: [
              { key: "focus-material", label: "只看该物料", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-spec", label: "只看同规格", onSelect: (item, helpers) => helpers.searchText(item.spec) }
            ],
            copyFields: [
              { key: "name", label: "物料名称", value: (item) => item.name },
              { key: "spec", label: "规格", value: (item) => item.spec },
              { key: "unit", label: "单位", value: (item) => item.unit },
              { key: "safe", label: "安全库存", value: (item) => qty(item.safeStock) }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("material")}>新增物料</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "物料", render: (item) => item.name },
            { key: "spec", title: "规格", render: (item) => item.spec },
            { key: "unit", title: "单位", render: (item) => item.unit },
            { key: "safeStock", title: "安全库存", render: (item) => qty(item.safeStock) },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("material", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderMasterSites() {
    const sites = siteOptions();
    return (
      <Panel>
        <DataTable
          data={sites}
          rowKey={(item) => item.id}
          emptyText="暂无站点"
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Site>({
            actions: [
              { key: "focus-site", label: "只看该站点", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-company", label: "只看该公司", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.companies, item.companyId)) }
            ],
            copyFields: [
              { key: "name", label: "站点名称", value: (item) => item.name },
              { key: "code", label: "站点编码", value: (item) => item.code },
              { key: "address", label: "站点地址", value: (item) => item.address },
              { key: "coordinate", label: "站点坐标", value: (item) => item.longitude && item.latitude ? `${item.longitude}, ${item.latitude}` : "" }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("site")}>新增站点</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "站点", render: (item) => item.name },
            { key: "code", title: "编码", render: (item) => item.code },
            { key: "company", title: "公司", render: (item) => nameOf(bootstrap?.companies, item.companyId) || "-" },
            { key: "address", title: "地址", render: (item) => item.address },
            { key: "fence", title: "电子围栏", render: (item) => {
              const fence = activeSiteFence(item.id);
              const longitude = fence?.longitude || item.longitude;
              const latitude = fence?.latitude || item.latitude;
              return isValidCoordinate(longitude, latitude) ? "已设置" : "未设置";
            } },
            { key: "status", title: "启用状态", render: (item) => <StatusChip value={siteEnableStatus(item.status)} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("site", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderMasterPlants() {
    const plants = list(bootstrap?.plants).filter((item) => matchesCurrentSite(item.siteId));
    const sites = siteOptions();
    const bufferSource = list(data.production?.plantBufferLocations);
    const buffers = bufferSource.filter((item) => matchesCurrentSite(item.siteId));
    const materials = bootstrap?.materials || [];
    const buffersByPlantId = new Map<number, PlantBufferLocation[]>();
    buffers.forEach((item) => {
      const plantBuffers = buffersByPlantId.get(item.plantId) || [];
      plantBuffers.push(item);
      buffersByPlantId.set(item.plantId, plantBuffers);
    });
    function plantStats(plant: Plant) {
      const plantBuffers = buffersByPlantId.get(plant.id) || [];
      const issueBuffers = plantBuffers.filter((item) => item.currentQty <= item.warningQty || item.qualityStatus === "blocked" || item.status === "disabled");
      const issueIds = new Set(issueBuffers.map((item) => item.id));
      const orderedBuffers = [...issueBuffers, ...plantBuffers.filter((item) => !issueIds.has(item.id))];
      return {
        plantBuffers,
        issueBuffers,
        orderedBuffers,
        reportedAt: plant.lastFrameAt || plantBuffers.find((item) => item.lastReportedAt)?.lastReportedAt || "",
        gatewayStatus: plant.gatewayStatus || (plant.gatewayDeviceNo ? "registered" : "not_connected")
      };
    }

    function openPlantBufferAction(plant: Plant, mode: "create" | "edit" | "transfer" | "adjust", buffer?: PlantBufferLocation) {
      openBufferDialog(mode, buffer, plant);
    }

    function openPlantCardMenu(event: ReactMouseEvent, plant: Plant) {
      event.preventDefault();
      event.stopPropagation();
      setPlantCardMenu({ plantId: plant.id, x: event.clientX, y: event.clientY });
    }

    const plantMenuPlant = plantCardMenu ? plants.find((item) => item.id === plantCardMenu.plantId) : undefined;
    const plantCardContextItems: ContextMenuItem[] = plantMenuPlant ? [
      {
        key: "add-plant-buffer",
        label: "添加筒仓",
        icon: <Plus size={14} />,
        disabled: actionBusy !== "" || !materials.length,
        onSelect: () => openPlantBufferAction(plantMenuPlant, "create")
      },
      {
        key: "edit-plant",
        label: "编辑生产线",
        icon: <Pencil size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => startMasterEdit("plant", plantMenuPlant)
      }
    ] : [];

    function renderPlantBufferCards(plant: Plant, plantBuffers: PlantBufferLocation[]) {
      return (
        <div className="plant-dialog-buffer-list">
	          {plantBuffers.map((buffer) => {
	            const isLowStock = buffer.currentQty <= buffer.warningQty;
	            const isBlocked = buffer.qualityStatus === "blocked" || buffer.status === "disabled";
	            const fillPct = buffer.capacity > 0 ? Math.min(100, Math.max(0, (buffer.currentQty / buffer.capacity) * 100)) : 0;
	            const adjustmentWorkflow = workflowItemsFor(["plant_buffer_adjustment"], buffer.id, buffer.code);
	            const hasPendingAdjustmentWorkflow = adjustmentWorkflow.instances.some((item) => item.status === "pending");
	            return (
	              <Card
                className={`plant-dialog-buffer-card ${isBlocked ? "danger" : isLowStock ? "warning" : ""}`}
                key={buffer.id}
                role="button"
                tabIndex={0}
                aria-label={`编辑${buffer.name}`}
                onClick={(event) => {
                  event.stopPropagation();
                  openPlantBufferAction(plant, "edit", buffer);
                }}
                onKeyDown={(event) => {
                  event.stopPropagation();
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    openPlantBufferAction(plant, "edit", buffer);
                  }
                }}
              >
                <div className="plant-buffer-title">
                  <b>{buffer.name}</b>
                  <span>{nameOf(materials, buffer.materialId) || "-"} / {dictionaryValueLabel("buffer_type", buffer.type, buffer.type)}</span>
                </div>
                <div className="plant-buffer-stock">
                  <div className="plant-buffer-stock-head">
                    <b>{qty(buffer.currentQty)} / {qty(buffer.capacity)} {buffer.unit}</b>
                    <span>{Math.round(fillPct)}%</span>
                  </div>
                  <div className="plant-buffer-progress" aria-hidden="true">
                    <span style={{ width: `${fillPct}%` }} />
                  </div>
                  <span>{isLowStock ? "低于阈值" : `阈值 ${qty(buffer.warningQty)} ${buffer.unit}`} / 含水率 {qty(buffer.moistureRate)}%</span>
                </div>
	                <div className="plant-simple-buffer-status">
	                  {workflowStatusFor(["plant_buffer_adjustment"], buffer.id, buffer.code, <StatusChip value={isLowStock && !isBlocked ? "warning" : buffer.qualityStatus} />)}
	                  <StatusChip value={buffer.status} />
	                </div>
	                <ActionGroup className="plant-dialog-buffer-actions">
                  <UiButton
                    variant={isLowStock ? "primary" : "soft"}
                    disabled={actionBusy !== ""}
                    onClick={(event) => {
                      event.stopPropagation();
                      openPlantBufferAction(plant, "transfer", buffer);
                    }}
                  >
                    上料
                  </UiButton>
	                  <UiButton
	                    disabled={actionBusy !== "" || hasPendingAdjustmentWorkflow}
	                    onClick={(event) => {
	                      event.stopPropagation();
	                      openPlantBufferAction(plant, "adjust", buffer);
                    }}
                  >
                    盘点
                  </UiButton>
                  <UiButton
                    variant="danger"
                    disabled={actionBusy !== ""}
                    onClick={(event) => {
                      event.stopPropagation();
                      void handleDeletePlantBuffer(buffer);
                    }}
                  >
                    删除
                  </UiButton>
                </ActionGroup>
              </Card>
            );
          })}
          {!plantBuffers.length ? (
            <div className="plant-simple-empty">
              <span>暂无筒仓</span>
              <UiButton
                disabled={actionBusy !== "" || !materials.length}
                onClick={(event) => {
                  event.stopPropagation();
                  openPlantBufferAction(plant, "create");
                }}
              >
                添加
              </UiButton>
            </div>
          ) : null}
        </div>
      );
    }

    return (
      <>
        <Panel className="plant-dialog-home">
          <div className="plant-dialog-home-head">
            <div>
              <h3>生产线</h3>
            </div>
            <ActionGroup>
              <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              <UiButton variant="primary" icon={<Plus size={15} />} disabled={!sites.length} onClick={() => openMasterCreateDialog("plant")}>新增生产线</UiButton>
            </ActionGroup>
          </div>
          {plants.length ? (
            <div className="plant-dialog-grid">
              {plants.map((plant) => {
                const stats = plantStats(plant);
                const recipe = productionLineRecipe(plant);
                return (
                  <Card
                    className={`plant-dialog-card ${stats.issueBuffers.length ? "warning" : ""}`}
                    key={plant.id}
                    role="button"
                    tabIndex={0}
                    aria-label={`编辑生产线${plant.name}`}
                    onClick={() => startMasterEdit("plant", plant)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        startMasterEdit("plant", plant);
                      }
                    }}
                    onContextMenu={(event) => openPlantCardMenu(event, plant)}
                  >
                    <div className="plant-dialog-card-head">
                      <span className="plant-line-icon"><Factory size={18} /></span>
                      <div>
                        <h3>{plant.name}</h3>
                        <p>{plant.code} / {nameOf(sites, plant.siteId) || "未关联站点"}</p>
                      </div>
                      <IconButton
                        className="plant-card-add-button"
                        icon={<Plus size={16} />}
                        label={`给${plant.name}添加筒仓`}
                        variant="primary"
                        disabled={actionBusy !== "" || !materials.length}
                        onClick={(event) => {
                          event.stopPropagation();
                          openPlantBufferAction(plant, "create");
                        }}
                      />
                    </div>
                    <div className="plant-active-mix">
                      <span>正在使用配比</span>
                      <b>{productionMixLabel(recipe.mix)}</b>
                      {recipe.profile ? <small>适配：{productionProfileLabel(recipe.profile)}</small> : null}
                      <small>
                        {recipe.source}
                        {recipe.productId ? ` / ${productLabel(bootstrap, recipe.productId)}` : ""}
                        {recipe.task ? ` / ${recipe.task.taskNo}` : recipe.plan ? ` / ${recipe.plan.planNo}` : ""}
                      </small>
                    </div>
                    <div className="plant-dialog-card-metrics">
                      <div><span>筒仓</span><b>{stats.plantBuffers.length}</b></div>
                      <div><span>异常</span><b>{stats.issueBuffers.length}</b></div>
                      <div><span>网关</span><b><StatusChip value={stats.gatewayStatus} /></b></div>
                      <div><span>上报</span><b>{stats.reportedAt ? shortDateTime(stats.reportedAt) : "-"}</b></div>
                    </div>
                    {plant.gatewayError ? (
                      <UiButton
                        className="plant-gateway-error-button"
                        icon={<AlertCircle size={14} />}
                        onClick={(event) => {
                          event.stopPropagation();
                          showError(plant.gatewayError, "网关异常", "网关异常");
                        }}
                      >
                        查看网关异常
                      </UiButton>
                    ) : null}
                    {renderPlantBufferCards(plant, stats.orderedBuffers)}
                  </Card>
                );
              })}
            </div>
          ) : (
            <div className="plant-empty-state">
              <div>
                <h3>暂无生产线</h3>
              </div>
              <UiButton variant="primary" icon={<Plus size={15} />} disabled={!sites.length} onClick={() => openMasterCreateDialog("plant")}>新增生产线</UiButton>
            </div>
          )}
          {plantCardMenu && plantMenuPlant ? (
            <ContextMenu
              items={plantCardContextItems}
              position={{ x: plantCardMenu.x, y: plantCardMenu.y }}
              label="生产线操作"
              onClose={() => setPlantCardMenu(null)}
            />
          ) : null}
        </Panel>
      </>
    );
  }

  function renderMasterDrivers() {
    const drivers = bootstrap?.drivers || [];
    return (
	      <Panel>
	        <DataTable
	          data={drivers}
	          rowKey={(item) => item.id}
	          emptyText="暂无司机"
	          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Driver>({
            actions: [
              { key: "focus-driver", label: "只看该司机", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-phone", label: "只看该手机号", onSelect: (item, helpers) => helpers.searchText(item.phone) }
            ],
            copyFields: [
              { key: "name", label: "司机姓名", value: (item) => item.name },
              { key: "phone", label: "手机号", value: (item) => item.phone },
              { key: "license", label: "证号", value: (item) => item.licenseNo },
              { key: "expire", label: "证件到期", value: (item) => item.licenseExpire }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("driver")}>新增司机</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "司机姓名", render: (item) => item.name },
            { key: "phone", title: "手机号", render: (item) => item.phone },
            { key: "licenseNo", title: "证件号", render: (item) => item.licenseNo },
            { key: "licenseExpire", title: "证件到期", render: (item) => item.licenseExpire },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("driver", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderMasterCarriers() {
    const carriers = bootstrap?.carriers || [];
    const settleModeOptions = dictionaryOptionsWithFallback("carrier_settle_mode", carrierSettleModeFallbackOptions);
    const settleModeLabel = (value: string) => settleModeOptions.find((item) => item.code === value)?.label || value || "-";
    return (
	      <Panel>
	        <DataTable
	          data={carriers}
	          rowKey={(item) => item.id}
	          emptyText="暂无承运商"
	          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Carrier>({
            actions: [
              { key: "focus-carrier", label: "只看该承运商", onSelect: (item, helpers) => helpers.searchText(item.name) },
              { key: "focus-phone", label: "只看该电话", onSelect: (item, helpers) => helpers.searchText(item.phone) }
            ],
            copyFields: [
              { key: "name", label: "承运商", value: (item) => item.name },
              { key: "contact", label: "联系人", value: (item) => item.contact },
              { key: "phone", label: "电话", value: (item) => item.phone },
              { key: "settleMode", label: "结算方式", value: (item) => settleModeLabel(item.settleMode) }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} onClick={() => openMasterCreateDialog("carrier")}>新增承运商</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "name", title: "承运商", render: (item) => item.name },
            { key: "contact", title: "联系人", render: (item) => item.contact || "-" },
            { key: "phone", title: "电话", render: (item) => item.phone || "-" },
            { key: "settleMode", title: "结算方式", render: (item) => settleModeLabel(item.settleMode) },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("carrier", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderMasterVehicles() {
    const vehicles = list(bootstrap?.vehicles).filter((item) => matchesCurrentSite(item.siteId));
    const drivers = bootstrap?.drivers || [];
    const sites = siteOptions();
    return (
      <Panel>
        <DataTable
          data={vehicles}
          rowKey={(item) => item.id}
          emptyText="暂无数据"
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<Vehicle>({
            actions: [
              { key: "focus-vehicle", label: "只看该车辆", onSelect: (item, helpers) => helpers.searchText(item.plateNo) },
              { key: "focus-driver", label: "只看该司机", onSelect: (item, helpers) => helpers.searchText(nameOf(drivers, item.driverId)) },
              { key: "focus-site", label: "只看该站点", onSelect: (item, helpers) => helpers.searchText(nameOf(sites, item.siteId)) }
            ],
            copyFields: [
              { key: "internalNo", label: "自编号", value: (item) => item.internalNo },
              { key: "plate", label: "车牌号", value: (item) => item.plateNo },
              { key: "driver", label: "司机", value: (item) => nameOf(drivers, item.driverId) },
              { key: "site", label: "站点", value: (item) => nameOf(sites, item.siteId) },
              { key: "type", label: "车辆类型", value: (item) => dictionaryValueLabel("vehicle_type", item.vehicleType, item.vehicleType) },
              { key: "device", label: "GPS 设备", value: (item) => vehicleDeviceFor(item.id)?.deviceNo || "-" }
            ]
          })}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={15} />} disabled={!drivers.length || !sites.length} onClick={() => openMasterCreateDialog("vehicle")}>新增车辆</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "internalNo", title: "自编号", render: (item) => item.internalNo || "-" },
            { key: "plateNo", title: "车牌", render: (item) => item.plateNo },
            { key: "vehicleType", title: "车型", render: (item) => dictionaryValueLabel("vehicle_type", item.vehicleType, item.vehicleType) },
            { key: "driver", title: "司机", render: (item) => nameOf(drivers, item.driverId) },
            { key: "site", title: "站点", render: (item) => nameOf(sites, item.siteId) },
            {
              key: "gpsDevice",
              title: "GPS 设备",
              render: (item) => {
                const device = vehicleDeviceFor(item.id);
                if (!device) return <span className="muted">未绑定</span>;
                return <b>{device.deviceNo}</b>;
              }
            },
            { key: "lastLocation", title: "最近定位", render: (item) => shortDateTime(latestLocationFor(item.id)?.lastLocationTime || vehicleDeviceFor(item.id)?.lastSeenAt) || "-" },
            { key: "onlineStatus", title: "在线状态", render: (item) => <StatusChip value={item.onlineStatus} /> },
            { key: "businessStatus", title: "业务状态", render: (item) => <StatusChip value={item.businessStatus} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("vehicle", item.id, item) }
          ]}
        />
      </Panel>
    );
  }

  function renderOrders() {
    const customers = bootstrap?.customers || [];
    const projects = bootstrap?.projects || [];
    const products = bootstrap?.products || [];
    const sites = siteOptions();
    const canCreateOrder = customers.length > 0 && projects.length > 0 && products.length > 0 && sites.length > 0;
    const orderSummary = [
      { label: "订单总数", value: scopedOrders.length },
      { label: "进行中", value: activeOrders.length },
      { label: "计划方量", value: qty(scopedOrders.reduce((sum, item) => sum + item.planQuantity, 0)) },
      { label: "已签收", value: qty(scopedOrders.reduce((sum, item) => sum + item.signedQty, 0)) },
      { label: "订单金额", value: money(scopedOrders.reduce((sum, item) => sum + item.totalAmount, 0)) }
    ];

    return (
      <Panel className="sales-orders-panel">
        <div className="sales-order-summary" aria-label="订单概览">
          {orderSummary.map((item) => (
            <div className="sales-order-summary-item" key={item.label}>
              <span>{item.label}</span>
              <b>{item.value}</b>
            </div>
          ))}
        </div>
        {ordersTable(scopedOrders, true, (
          <ActionGroup>
            <UiButton variant="primary" icon={<Plus size={14} />} disabled={!canCreateOrder} onClick={openOrderDialog}>新增订单</UiButton>
          </ActionGroup>
        ))}
      </Panel>
    );
  }

  function renderProduction() {
    const plans = list(data.production?.plans).filter((item) => matchesCurrentSite(item.siteId));
    const tasks = list(data.production?.tasks).filter((item) => matchesCurrentSite(item.siteId));
    const batches = list(data.production?.batches).filter((item) => matchesCurrentSite(item.siteId));
    const plants = productionPlants().filter((item) => matchesCurrentSite(item.siteId));
    const plantBufferFlows = [...list(data.production?.plantBufferFlows)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const inventoryBatchTraces = [...list(data.production?.traces)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const productionQueueRows = list(data.dispatch?.vehicleQueue)
      .filter((item) => matchesCurrentSite(item.siteId))
      .filter((item) => ["assigned", "accepted", "arrived_site", "waiting_load", "loading"].includes(item.status));
    const activeTasks = tasks.filter((item) => item.status !== "cancelled" && item.status !== "completed");
    const openProductionOrders = scopedOrders.filter((item) => productionOrderRemaining(item, plans) > 0 && ["approved", "scheduled", "dispatching"].includes(item.status));
    const hasOpenOrders = openProductionOrders.length > 0;
    const taskReadyPlans = plans.filter(productionPlanCanIssueTask);
    const batchReadyTasks = sortProductionTasksForAction(activeTasks.filter((item) => productionTaskRemaining(item) > 0));
    const reportReadyPlans = plans.filter(productionPlanCanOpenReport);
    const qualityBatchIds = new Set(batches.map((item) => item.id));
    const qualityInspections = list(data.quality?.inspections).filter((item) => matchesCurrentSite(item.siteId));
    const inspectedBatchIds = new Set(qualityInspections.map((item) => item.batchId));
    const qualitySamples = list(data.quality?.samples).filter((item) => qualityBatchIds.has(item.batchId));
    const qualityInspectionCandidates = batches.filter((item) => !inspectedBatchIds.has(item.id));
    const qualityResultOptions = dictionaryOptionsWithFallback("quality_result", [
      { code: "passed", label: "合格" },
      { code: "failed", label: "不合格" }
    ]);
    const qualityInspectionFormView = (
      <DialogForm onSubmit={handleCreateQualityInspection}>
        <Field label="生产批次">
          <SelectInput value={qualityInspectionForm.batchId} onChange={(event) => setQualityInspectionForm({ ...qualityInspectionForm, batchId: event.target.value })}>
            {qualityInspectionCandidates.map((item) => <option key={item.id} value={item.id}>{item.batchNo} / {productLabel(bootstrap, item.productId)} / {qty(item.quantity)}</option>)}
          </SelectInput>
        </Field>
        <Field label="检验员"><TextInput value={qualityInspectionForm.inspector} onChange={(event) => setQualityInspectionForm({ ...qualityInspectionForm, inspector: event.target.value })} /></Field>
        <Field label="坍落度/油石比"><TextInput value={qualityInspectionForm.slump} onChange={(event) => setQualityInspectionForm({ ...qualityInspectionForm, slump: event.target.value })} /></Field>
        <Field label="温度"><TextInput type="number" value={qualityInspectionForm.temperature} onChange={(event) => setQualityInspectionForm({ ...qualityInspectionForm, temperature: event.target.value })} /></Field>
        <Field label="备注" spanAll><TextAreaInput value={qualityInspectionForm.remark} onChange={(event) => setQualityInspectionForm({ ...qualityInspectionForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(qualityInspectionForm.batchId)}>创建质检单</UiButton>
        </FormActions>
      </DialogForm>
    );
    const qualitySampleFormView = (
      <DialogForm onSubmit={handleTestQualitySample}>
        <Field label="强度"><TextInput type="number" min="0" step="0.1" value={qualitySampleForm.strength} onChange={(event) => setQualitySampleForm({ ...qualitySampleForm, strength: event.target.value })} /></Field>
        <Field label="结论">
          <SelectInput value={qualitySampleForm.result} onChange={(event) => setQualitySampleForm({ ...qualitySampleForm, result: event.target.value })}>
            {qualityResultOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
          </SelectInput>
        </Field>
        <HeroDateField className="span-all" label="检测时间" mode="date-time" value={qualitySampleForm.testedAt} onChange={(testedAt) => setQualitySampleForm({ ...qualitySampleForm, testedAt })} />
        <Field label="备注" spanAll><TextAreaInput value={qualitySampleForm.remark} onChange={(event) => setQualitySampleForm({ ...qualitySampleForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !fieldNumber(qualitySampleForm.sampleId) || fieldNumber(qualitySampleForm.strength) <= 0}>登记检测</UiButton>
        </FormActions>
      </DialogForm>
    );

    function renderProductionFlowStrip() {
      const firstOpenOrder = openProductionOrders[0];
      const taskPlan = taskReadyPlans[0];
      const batchTask = batchReadyTasks[0];
      const batchPlan = batchTask ? plans.find((item) => item.id === batchTask.planId) : undefined;
      const reportPlan = reportReadyPlans[0];
      const steps = [
        {
          key: "plan",
          title: "1 生成计划",
          metric: `${openProductionOrders.length} 单待计划`,
          hint: firstOpenOrder ? `${firstOpenOrder.orderNo} · 未计划 ${qty(productionOrderRemaining(firstOpenOrder, plans))}` : "订单已计划完成",
          action: "新建计划",
          icon: <Plus size={14} />,
          disabled: !firstOpenOrder,
          onClick: () => openProductionDialog("create-plan")
        },
        {
          key: "task",
          title: "2 下达任务",
          metric: `${taskReadyPlans.length} 个计划可下达`,
          hint: taskPlan ? `${taskPlan.planNo} · 未下达 ${qty(productionPlanTaskRemaining(taskPlan))}` : "暂无待下达任务的计划",
          action: "下达任务",
          icon: <PlayCircle size={14} />,
          disabled: !taskPlan,
          onClick: () => taskPlan && openProductionDialog("tasks", taskPlan)
        },
        {
          key: "batch",
          title: "3 登记批次",
          metric: `${batchReadyTasks.length} 个任务可生产`,
          hint: batchTask ? `${batchTask.taskNo} · 未生产 ${qty(productionTaskRemaining(batchTask))}` : "暂无可登记批次的任务",
          action: "登记批次",
          icon: <Factory size={14} />,
          disabled: !batchTask,
          onClick: () => batchTask && openProductionDialog("batch", batchPlan, batchTask)
        },
        {
          key: "report",
          title: "4 生成日报",
          metric: `${reportReadyPlans.length} 个日报待汇总`,
          hint: reportPlan ? `${nameOf(bootstrap?.sites, reportPlan.siteId)} · ${reportPlan.planDate}` : "当天生产已汇总或暂无产量",
          action: "生成日报",
          icon: <ClipboardCheck size={14} />,
          disabled: !reportPlan,
          onClick: () => reportPlan && openProductionDialog("report", reportPlan)
        }
      ];
      const nextStep = [steps[2], steps[3], steps[1], steps[0]].find((step) => !step.disabled);

      return (
        <div className="production-flow-strip">
          <div className="production-flow-title">
            <b>生产流程</b>
            <span>{nextStep ? `建议先做：${nextStep.title.replace(/^\d+\s*/, "")}，${nextStep.hint}` : "暂无待处理生产事项"}</span>
            {nextStep ? (
              <UiButton size="sm" variant="primary" icon={nextStep.icon} disabled={actionBusy !== ""} onClick={nextStep.onClick}>
                {nextStep.action}
              </UiButton>
            ) : (
              <UiButton size="sm" icon={<RefreshCw size={14} />} onClick={refreshData}>刷新</UiButton>
            )}
          </div>
          <div className="production-flow-steps">
            {steps.map((step) => (
              <div className={`production-flow-step ${step.disabled ? "" : "ready"} ${nextStep?.key === step.key ? "current" : ""}`} key={step.key}>
                <div className="production-flow-step-main">
                  <span>{step.title}</span>
                  <b>{step.metric}</b>
                  <small>{step.hint}</small>
                </div>
                <UiButton size="sm" variant={nextStep?.key === step.key ? "primary" : "soft"} icon={step.icon} disabled={actionBusy !== "" || step.disabled} onClick={step.onClick}>
                  {step.action}
                </UiButton>
              </div>
            ))}
          </div>
        </div>
      );
    }

    function renderProductionPlanActions(item: ProductionPlan) {
      const nextAction = productionPlanNextAction(item);
      return (
        <ActionGroup as="span" className="production-plan-actions">
          {nextAction.mode !== "detail" ? (
            <UiButton size="sm" variant={nextAction.variant || "soft"} icon={nextAction.icon} disabled={actionBusy !== "" || nextAction.disabled} onClick={() => openProductionDialog(nextAction.mode, item)}>
              {nextAction.label}
            </UiButton>
          ) : null}
          <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => openProductionDialog("detail", item)}>
            详情
          </UiButton>
        </ActionGroup>
      );
    }

    function productionPlanRowContextMenu(item: ProductionPlan, _index: number, helpers: DataTableContextMenuHelpers<ProductionPlan>): ContextMenuItem[] {
      const nextAction = productionPlanNextAction(item);
      const orderNo = data.orders.find((order) => order.id === item.orderId)?.orderNo || String(item.orderId);
      const planActions: Array<{
        key: string;
        label: string;
        icon: ReactNode;
        mode: ProductionDialogMode;
        disabled: boolean;
      }> = [
        {
          key: "generate-production-report",
          label: "生成生产日报",
          icon: <ClipboardCheck size={14} />,
          mode: "report",
          disabled: !productionPlanCanOpenReport(item)
        },
        {
          key: "register-production-batch",
          label: "登记生产批次",
          icon: <Factory size={14} />,
          mode: "batch",
          disabled: !activeProductionTasksForPlan(item).length
        },
        {
          key: "issue-production-task",
          label: "下达生产任务",
          icon: <PlayCircle size={14} />,
          mode: "tasks",
          disabled: !productionPlanCanIssueTask(item)
        }
      ];
      const orderedPlanActions = [
        ...planActions.filter((action) => action.mode === nextAction.mode),
        ...planActions.filter((action) => action.mode !== nextAction.mode)
      ].filter((action) => !action.disabled);

      return [
        ...orderedPlanActions.map((action) => ({
          key: action.key,
          label: action.label,
          icon: action.icon,
          onSelect: () => openProductionDialog(action.mode, item)
        })),
        { key: "open-plan-detail", label: "打开计划详情", icon: <Search size={14} />, onSelect: () => openProductionDialog("detail", item) },
        { key: "production-plan-focus-separator", type: "separator" as const },
        { key: "focus-order", label: "只看该订单", icon: <Filter size={14} />, onSelect: () => helpers.searchText(orderNo) },
        { key: "focus-plant", label: "只看该生产线", icon: <Filter size={14} />, onSelect: () => helpers.searchText(productionPlantLabel(productionPlanPlant(item))) },
        { key: "production-plan-copy-separator", type: "separator" as const },
        { key: "copy-plan", label: "复制计划号", icon: <Copy size={14} />, onSelect: () => helpers.copyText(item.planNo, "计划号") },
        { key: "copy-order", label: "复制订单号", icon: <Copy size={14} />, onSelect: () => helpers.copyText(orderNo, "订单号") },
        { key: "copy-plant", label: "复制生产线", icon: <Copy size={14} />, onSelect: () => helpers.copyText(productionPlantLabel(productionPlanPlant(item)), "生产线") },
        { key: "copy-date", label: "复制计划日期", icon: <Copy size={14} />, onSelect: () => helpers.copyText(item.planDate, "计划日期") }
      ];
    }

    function renderProductionQueueRow(item: DispatchCenterQueueItem, lineQueue: DispatchCenterQueueItem[], index: number) {
      const loading = item.status === "loading";
      return (
        <div className={`production-queue-row ${loading ? "loading" : ""}`} key={item.dispatchId}>
          <span className="production-queue-index">{index + 1}</span>
          <div className="production-queue-vehicle">
            <b>{dispatchVehicleTitle(item)}</b>
            <small>{dispatchVehicleMeta(item)} · {item.projectName}</small>
          </div>
          <StatusChip value={item.status} />
          <ActionGroup className="production-queue-actions">
            <IconButton icon={<ArrowUp size={13} />} label="置顶" disabled={index === 0 || dispatchSubmitting} onClick={() => prioritizeProductionQueueItem(lineQueue, item.dispatchId)} />
            <IconButton icon={<ArrowUp size={13} />} label="上移" disabled={index === 0 || dispatchSubmitting} onClick={() => moveProductionQueueItem(lineQueue, item.dispatchId, -1)} />
            <IconButton icon={<ArrowDown size={13} />} label="下移" disabled={index === lineQueue.length - 1 || dispatchSubmitting} onClick={() => moveProductionQueueItem(lineQueue, item.dispatchId, 1)} />
            <UiButton
              size="sm"
              variant="primary"
              icon={<PlayCircle size={12} />}
              disabled={dispatchSubmitting}
              onClick={() => handleProductionQueueStatus(item, loading ? "loaded" : "loading")}
            >
              {loading ? "装完" : "下发装料"}
            </UiButton>
          </ActionGroup>
        </div>
      );
    }

    function renderProductionLineBoard() {
      const loadingTotal = productionQueueRows.filter((item) => item.status === "loading").length;
      const waitingTotal = productionQueueRows.length - loadingTotal;
      const lineModels = plants.map((plant) => ({
        plant,
        tasks: sortProductionTasksForAction(activeTasks
          .filter((task) => productionTaskPlant(task)?.id === plant.id)
          .filter((task) => productionTaskRemaining(task) > 0)),
        queue: [] as DispatchCenterQueueItem[]
      }));
      const lineModelByPlantId = new Map(lineModels.map((item) => [item.plant.id, item]));
      const unmatchedQueue: DispatchCenterQueueItem[] = [];

      productionQueueRows.forEach((item) => {
        const task = sortProductionTasksForAction(activeTasks.filter((candidate) => candidate.orderId === item.orderId && candidate.siteId === item.siteId))[0];
        const plant = productionTaskPlant(task);
        const model = plant ? lineModelByPlantId.get(plant.id) : undefined;
        if (model) {
          model.queue.push(item);
        } else {
          unmatchedQueue.push(item);
        }
      });

      function renderLineCard({
        key,
        title,
        meta,
        status,
        tasks,
        queue,
        unmatched = false
      }: {
        key: string | number;
        title: string;
        meta: string;
        status?: string;
        tasks: ProductionTask[];
        queue: DispatchCenterQueueItem[];
        unmatched?: boolean;
      }) {
        const lineQueue = sortedProductionQueue(queue);
        const loadingRows = lineQueue.filter((item) => item.status === "loading");
        const waitingRows = lineQueue.filter((item) => item.status !== "loading");
        return (
          <div className={`production-line-card ${unmatched ? "unmatched" : ""}`} key={key}>
            <div className="production-line-card-head">
              <div className="production-line-card-title">
                <b>{title}</b>
                <span>{meta}</span>
              </div>
              {status ? <StatusChip value={status} /> : null}
            </div>
            <div className="production-line-card-stats">
              <span><b>{tasks.length}</b><small>生产任务</small></span>
              <span><b>{loadingRows.length}</b><small>装料中</small></span>
              <span><b>{waitingRows.length}</b><small>等待中</small></span>
            </div>
            <div className="production-line-task-summary">
              {tasks.map((task) => {
                const taskPlan = plans.find((item) => item.id === task.planId);
                return (
                  <div className="production-line-task-item" key={task.id}>
                    <span>{task.taskNo} / {productLabel(bootstrap, task.productId)} / 剩余 {qty(productionTaskRemaining(task))}</span>
                    <UiButton size="sm" icon={<Plus size={12} />} disabled={actionBusy !== "" || productionTaskRemaining(task) <= 0} onClick={() => openProductionDialog("batch", taskPlan, task)}>登记</UiButton>
                  </div>
                );
              })}
              {!tasks.length ? <span>{unmatched ? "未匹配到生产任务" : "暂无生产任务"}</span> : null}
            </div>
            <div className="production-line-queue-columns">
              <div className="production-line-queue-column">
                <b>装料中</b>
                {loadingRows.map((item) => renderProductionQueueRow(item, lineQueue, lineQueue.findIndex((row) => row.dispatchId === item.dispatchId)))}
                {!loadingRows.length ? <span className="production-line-empty">暂无车辆</span> : null}
              </div>
              <div className="production-line-queue-column">
                <b>等待中</b>
                {waitingRows.map((item) => renderProductionQueueRow(item, lineQueue, lineQueue.findIndex((row) => row.dispatchId === item.dispatchId)))}
                {!waitingRows.length ? <span className="production-line-empty">暂无车辆</span> : null}
              </div>
            </div>
          </div>
        );
      }

      return (
        <div className="production-line-board">
          <div className="production-line-board-head">
            <div className="production-line-board-title">
              <h3>生产线车辆看板</h3>
              <span>按生产线汇总当前装料与等待车辆</span>
            </div>
            <div className="production-line-board-stats" aria-label="生产线车辆汇总">
              <span><b>{plants.length}</b><small>生产线</small></span>
              <span><b>{loadingTotal}</b><small>装料</small></span>
              <span><b>{waitingTotal}</b><small>等待</small></span>
              {unmatchedQueue.length ? <span><b>{unmatchedQueue.length}</b><small>未匹配</small></span> : null}
            </div>
            <UiButton icon={<RefreshCw size={14} />} onClick={refreshData} disabled={dispatchSubmitting}>刷新</UiButton>
          </div>
          <div className="production-line-grid">
            {lineModels.map((model) => renderLineCard({
              key: model.plant.id,
              title: productionPlantLabel(model.plant),
              meta: `${nameOf(bootstrap?.sites, model.plant.siteId)} / ${model.plant.capacity || "-"}`,
              status: model.plant.gatewayStatus || model.plant.status,
              tasks: model.tasks,
              queue: model.queue
            }))}
            {unmatchedQueue.length ? renderLineCard({
              key: "unmatched",
              title: "未匹配生产线",
              meta: "等待关联生产任务",
              status: "pending",
              tasks: [],
              queue: unmatchedQueue,
              unmatched: true
            }) : null}
            {!plants.length ? (
              <EmptyState className="dispatch-empty-state" title="暂无生产线">
                当前站点没有可用生产线
              </EmptyState>
            ) : null}
          </div>
        </div>
      );
    }

    if (section === "production-plans") {
      const plannedQty = plans.reduce((sum, item) => sum + item.planQuantity, 0);
      const producedQty = plans.reduce((sum, item) => sum + item.producedQty, 0);
      const pendingPlans = plans.filter((item) => item.status !== "completed" && item.status !== "cancelled");
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board production-workbench">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="生产计划范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{plans.length} 个计划</span>
                <span className={`inventory-scope-chip ${taskReadyPlans.length ? "warning" : ""}`}>{taskReadyPlans.length ? `${taskReadyPlans.length} 个待下达` : "任务已下达"}</span>
              </div>
              <ActionGroup>
                <UiButton variant="primary" icon={<Plus size={14} />} disabled={actionBusy !== "" || !hasOpenOrders} onClick={() => openProductionDialog("create-plan")}>新建计划</UiButton>
                <ButtonLink icon={<ListChecks size={15} />} href="/fulfillment/production/tasks">生产任务</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>计划量</span><b>{qty(plannedQty)}</b><small>{pendingPlans.length} 个未完成计划</small></div>
              <div><span>已产量</span><b>{qty(producedQty)}</b><small>{percent(plannedQty ? producedQty / plannedQty * 100 : 0)}</small></div>
              <div><span>待计划订单</span><b>{openProductionOrders.length}</b><small>可生成生产计划</small></div>
              <div><span>待下达</span><b>{taskReadyPlans.length}</b><small>计划转任务</small></div>
              <div><span>日报待汇总</span><b>{reportReadyPlans.length}</b><small>按站点/日期生成</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<ProductionPlan>
              title="生产计划"
              data={plans}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              rowContextMenu={productionPlanRowContextMenu}
              columns={[
                { key: "plan", title: "计划", render: (item) => <b>{item.planNo}</b> },
                { key: "order", title: "订单", render: (item) => data.orders.find((order) => order.id === item.orderId)?.orderNo || item.orderId },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "plant", title: "生产线", render: (item) => productionPlantLabel(productionPlanPlant(item)) },
                { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
                { key: "date", title: "日期/班次", render: (item) => `${item.planDate} · ${dictionaryValueLabel("shift_type", item.shift, item.shift)}` },
                { key: "qty", title: "计划/已产/剩余", render: (item) => `${qty(item.planQuantity)} / ${qty(item.producedQty)} / ${qty(item.remainingQty)}` },
                { key: "progress", title: "进度", render: (item) => percent(item.progress) },
                { key: "related", title: "任务/批次", render: (item) => `${tasks.filter((task) => task.planId === item.id).length} / ${batches.filter((batch) => batch.planId === item.id).length}` },
                { key: "risk", title: "检查", render: (item) => <ActionGroup as="span"><StatusChip value={item.capacityStatus} /><StatusChip value={item.inventoryStatus} /><StatusChip value={item.recipeStatus} /></ActionGroup> },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["production_plan", "productionPlan"], item.id, item.planNo, <StatusChip value={item.status} />) },
                { key: "actions", title: "操作", render: renderProductionPlanActions }
              ]}
              emptyText="暂无生产计划"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "production-tasks") {
      const taskQty = tasks.reduce((sum, item) => sum + item.planQty, 0);
      const taskProducedQty = tasks.reduce((sum, item) => sum + item.producedQty, 0);
      const firstTaskPlan = taskReadyPlans[0];
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board production-workbench">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="生产任务范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{tasks.length} 个任务</span>
                <span className={`inventory-scope-chip ${batchReadyTasks.length ? "warning" : ""}`}>{batchReadyTasks.length ? `${batchReadyTasks.length} 个可生产` : "暂无待生产任务"}</span>
              </div>
              <ActionGroup>
                <UiButton variant="primary" icon={<PlayCircle size={14} />} disabled={actionBusy !== "" || !firstTaskPlan} onClick={() => firstTaskPlan && openProductionDialog("tasks", firstTaskPlan)}>下达任务</UiButton>
                <ButtonLink icon={<Factory size={15} />} href="/fulfillment/production/batches">生产批次</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>任务量</span><b>{qty(taskQty)}</b><small>{activeTasks.length} 个进行中</small></div>
              <div><span>已产量</span><b>{qty(taskProducedQty)}</b><small>{percent(taskQty ? taskProducedQty / taskQty * 100 : 0)}</small></div>
              <div><span>待下达计划</span><b>{taskReadyPlans.length}</b><small>计划转任务</small></div>
              <div><span>可登记批次</span><b>{batchReadyTasks.length}</b><small>任务转批次</small></div>
              <div><span>生产线</span><b>{plants.length}</b><small>当前站点范围</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<ProductionTask>
              title="生产任务"
              data={tasks}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "task", title: "任务", render: (item) => <><b>{item.taskNo}</b><span className="block-text muted">{plans.find((plan) => plan.id === item.planId)?.planNo || `计划 #${item.planId}`}</span></> },
                { key: "order", title: "订单", render: (item) => data.orders.find((order) => order.id === item.orderId)?.orderNo || item.orderId },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "plant", title: "生产线", render: (item) => productionPlantLabel(productionTaskPlant(item)) },
                { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
                { key: "qty", title: "任务/已产/剩余", render: (item) => `${qty(item.planQty)} / ${qty(item.producedQty)} / ${qty(productionTaskRemaining(item))}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "time", title: "开始/完成", render: (item) => `${shortDateTime(item.startedAt)} / ${shortDateTime(item.completedAt)}` },
                {
                  key: "actions",
                  title: "操作",
                  width: "160px",
                  render: (item) => {
                    const plan = plans.find((candidate) => candidate.id === item.planId);
                    return (
                      <ActionGroup className="compact-actions">
                        <UiButton size="sm" variant="primary" disabled={actionBusy !== "" || productionTaskRemaining(item) <= 0} onClick={() => openProductionDialog("batch", plan, item)}>登记批次</UiButton>
                        {plan ? <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => openProductionDialog("detail", plan)}>计划</UiButton> : null}
                      </ActionGroup>
                    );
                  }
                }
              ]}
              emptyText="暂无生产任务"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "production-batches") {
      const batchQty = batches.reduce((sum, item) => sum + item.quantity, 0);
      const firstBatchTask = batchReadyTasks[0];
      const firstBatchPlan = firstBatchTask ? plans.find((item) => item.id === firstBatchTask.planId) : undefined;
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board production-workbench">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="生产批次范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{batches.length} 个批次</span>
                <span className={`inventory-scope-chip ${qualityInspectionCandidates.length ? "warning" : ""}`}>{qualityInspectionCandidates.length ? `${qualityInspectionCandidates.length} 个待质检` : "质检已覆盖"}</span>
              </div>
              <ActionGroup>
                <UiButton variant="primary" icon={<Plus size={14} />} disabled={actionBusy !== "" || !firstBatchTask} onClick={() => firstBatchTask && openProductionDialog("batch", firstBatchPlan, firstBatchTask)}>登记批次</UiButton>
                <ActionDialog
                  id="quality-inspection-create-batches-page"
                  title="创建生产质检单"
                  buttonLabel="创建质检"
                  triggerIcon={<ClipboardCheck size={13} />}
                  disabled={!qualityInspectionCandidates.length}
                  onOpen={() => qualityInspectionCandidates[0] && startQualityInspection(qualityInspectionCandidates[0])}
                >
                  {qualityInspectionFormView}
                </ActionDialog>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>批次数</span><b>{batches.length}</b><small>{qty(batchQty)} 总产量</small></div>
              <div><span>可生产任务</span><b>{batchReadyTasks.length}</b><small>任务剩余量未清</small></div>
              <div><span>待质检批次</span><b>{qualityInspectionCandidates.length}</b><small>批次转质检</small></div>
              <div><span>质检单</span><b>{qualityInspections.length}</b><small>生产质量记录</small></div>
              <div><span>追溯</span><b>{inventoryBatchTraces.length}</b><small>原料批次链路</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<ProductionBatch>
              title="生产批次"
              data={batches}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "batch", title: "批次", render: (item) => <><b>{item.batchNo}</b><span className="block-text muted">{tasks.find((task) => task.id === item.taskId)?.taskNo || `任务 #${item.taskId}`}</span></> },
                { key: "order", title: "订单", render: (item) => data.orders.find((order) => order.id === item.orderId)?.orderNo || item.orderId },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "plant", title: "生产线", render: (item) => productionPlantLabel(plants.find((plant) => plant.id === item.plantId)) },
                { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
                { key: "qty", title: "数量", render: (item) => qty(item.quantity) },
                { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "time", title: "开始/完成", render: (item) => `${shortDateTime(item.startedAt)} / ${shortDateTime(item.completedAt)}` },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => qualityInspectionCandidates.some((candidate) => candidate.id === item.id) ? (
                    <ActionDialog id={`quality-inspection-create-batch-${item.id}`} title="创建生产质检单" buttonLabel="质检" onOpen={() => startQualityInspection(item)}>
                      {qualityInspectionFormView}
                    </ActionDialog>
                  ) : <span className="muted">已建质检</span>
                }
              ]}
              emptyText="暂无生产批次"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<InventoryBatchTrace>
              title="批次追溯"
              data={inventoryBatchTraces}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "trace", title: "追溯号", render: (item) => <><b>{item.traceNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                { key: "batch", title: "生产批次", render: (item) => <><b>{item.productionBatchNo || `#${item.productionBatchId}`}</b><span className="block-text muted">原料批次 {item.batchNo || "-"}</span></> },
                { key: "receipt", title: "入库单/库存", render: (item) => `入库 #${item.rawReceiptId || "-"} / 库存 #${item.inventoryItemId || "-"}` },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                { key: "supplier", title: "供应商", render: (item) => recordName(supplierOptions().find((supplier) => recordId(supplier) === item.supplierId), `供应商 ${item.supplierId || "-"}`) },
                { key: "quantity", title: "用量", render: (item) => `${qty(item.quantity)} ${item.unit || "t"}` }
              ]}
              emptyText="暂无批次追溯"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "production-reports") {
      const productionReports = [...list(data.production?.reports)]
        .filter((item) => matchesCurrentSite(item.siteId))
        .sort((a, b) => (b.reportDate || "").localeCompare(a.reportDate || "") || b.id - a.id);
      const reportProducedQty = productionReports.reduce((sum, item) => sum + item.producedQty, 0);
      const reportPlan = reportReadyPlans[0];
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board production-workbench">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="生产日报范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{productionReports.length} 张日报</span>
                <span className={`inventory-scope-chip ${reportReadyPlans.length ? "warning" : ""}`}>{reportReadyPlans.length ? `${reportReadyPlans.length} 个待汇总` : "日报已汇总"}</span>
              </div>
              <ActionGroup>
                <UiButton variant="primary" icon={<ClipboardCheck size={14} />} disabled={actionBusy !== "" || !reportPlan} onClick={() => reportPlan && openProductionDialog("report", reportPlan)}>生成日报</UiButton>
                <ButtonLink icon={<ClipboardCheck size={15} />} href="/fulfillment/production/plans">生产计划</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>日报</span><b>{productionReports.length}</b><small>当前站点范围</small></div>
              <div><span>汇总产量</span><b>{qty(reportProducedQty)}</b><small>{productionReports.reduce((sum, item) => sum + item.batchCount, 0)} 个批次</small></div>
              <div><span>待汇总</span><b>{reportReadyPlans.length}</b><small>计划已生产未汇总</small></div>
              <div><span>质检通过</span><b>{productionReports.reduce((sum, item) => sum + item.qualityPassed, 0)}</b><small>日报质检统计</small></div>
              <div><span>质检待定</span><b>{productionReports.reduce((sum, item) => sum + item.qualityPending, 0)}</b><small>需跟进</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="生产日报"
              data={productionReports}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "report", title: "日报", render: (item) => <><b>{item.reportNo}</b><span className="block-text muted">{item.reportDate}</span></> },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "qty", title: "计划/实产", render: (item) => `${qty(item.plannedQty)} / ${qty(item.producedQty)}` },
                { key: "batch", title: "批次", render: (item) => item.batchCount },
                { key: "quality", title: "质检", render: (item) => `通过 ${item.qualityPassed} / 待定 ${item.qualityPending}` },
                { key: "cost", title: "材料成本", render: (item) => money(item.materialCost) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "generated", title: "生成时间", render: (item) => shortDateTime(item.generatedAt) }
              ]}
              emptyText="暂无生产日报"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <Panel className="production-workbench">
        {renderProductionFlowStrip()}
        {renderProductionLineBoard()}
        <DataTable<ProductionPlan>
          data={plans}
          rowKey={(item) => item.id}
          emptyText="暂无数据"
          pageSize={10}
          onRefresh={refreshData}
          rowContextMenu={productionPlanRowContextMenu}
          headerLeftAction={(
            <ActionGroup>
              <UiButton variant="primary" icon={<Plus size={14} />} disabled={actionBusy !== "" || !hasOpenOrders} onClick={() => openProductionDialog("create-plan")}>新建计划</UiButton>
            </ActionGroup>
          )}
          columns={[
            { key: "plan", title: "计划", render: (item) => item.planNo },
            { key: "order", title: "订单", render: (item) => data.orders.find((order) => order.id === item.orderId)?.orderNo || item.orderId },
            { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
            { key: "plant", title: "生产线", render: (item) => productionPlantLabel(productionPlanPlant(item)) },
            { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
            {
              key: "recipe",
              title: "配比",
              render: (item) => {
                const mix = list(data.production?.mixDesigns).find((mixItem) => mixItem.id === item.mixDesignId) || currentApprovedMix(item.productId, item.siteId);
                const profile = productionProfileById(item.mixProfileId) || currentProductionProfile(mix?.id || item.mixDesignId, item.plantId, item.planDate);
                return `${productionMixLabel(mix)}${profile ? ` / ${productionProfileLabel(profile)}` : ""}`;
              }
            },
            { key: "date", title: "日期/班次", render: (item) => `${item.planDate} · ${dictionaryValueLabel("shift_type", item.shift, item.shift)}` },
            { key: "qty", title: "计划/已产/剩余", render: (item) => `${qty(item.planQuantity)} / ${qty(item.producedQty)} / ${qty(item.remainingQty)}` },
            { key: "progress", title: "进度", render: (item) => percent(item.progress) },
            { key: "related", title: "任务/批次", render: (item) => `${tasks.filter((task) => task.planId === item.id).length} / ${batches.filter((batch) => batch.planId === item.id).length}` },
            { key: "risk", title: "检查", render: (item) => <ActionGroup as="span"><StatusChip value={item.capacityStatus} /><StatusChip value={item.inventoryStatus} /><StatusChip value={item.recipeStatus} /></ActionGroup> },
            { key: "status", title: "状态", render: (item) => workflowStatusFor(["production_plan", "productionPlan"], item.id, item.planNo, <StatusChip value={item.status} />) },
            {
              key: "actions",
              title: "操作",
              render: renderProductionPlanActions
            }
          ]}
        />
        <section className="production-detail-dialog-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <h3>辅助明细</h3>
            </div>
            <ActionGroup className="production-detail-dialog-actions">
              <ActionDialog
                id="production-quality-inspections-dialog"
                title="生产质检单"
                buttonLabel={`生产质检单 ${qualityInspections.length}`}
                triggerIcon={<ClipboardCheck size={13} />}
                size="xl"
                className="production-detail-dialog"
                bodyClassName="production-detail-dialog-body"
              >
                <DataTable
                  title="生产质检单"
                  data={qualityInspections}
                  rowKey={(item) => item.id}
                  pageSize={6}
                  headerLeftAction={(
                    <ActionDialog
                      id="quality-inspection-create"
                      title="创建生产质检单"
                      buttonLabel="创建质检"
                      triggerIcon={<ClipboardCheck size={13} />}
                      disabled={!qualityInspectionCandidates.length}
                      onOpen={() => qualityInspectionCandidates[0] && startQualityInspection(qualityInspectionCandidates[0])}
                    >
                      {qualityInspectionFormView}
                    </ActionDialog>
                  )}
                  columns={[
                    { key: "inspection", title: "质检单", render: (item) => <><b>{item.inspectionNo}</b><span className="block-text muted">{item.batchNo}</span></> },
                    { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                    { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
                    { key: "inspector", title: "检验员", render: (item) => item.inspector || "-" },
                    { key: "samples", title: "样品", render: (item) => `${qualitySamples.filter((sample) => sample.inspectionId === item.id && sample.status === "completed").length}/${item.sampleCount}` },
                    { key: "result", title: "结论", render: (item) => <StatusChip value={item.result} /> },
                    { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                    { key: "createdAt", title: "创建时间", render: (item) => shortDateTime(item.createdAt) }
                  ]}
                  emptyText="暂无生产质检单"
                />
              </ActionDialog>
              <ActionDialog
                id="production-batch-traces-dialog"
                title="批次追溯"
                buttonLabel={`批次追溯 ${inventoryBatchTraces.length}`}
                triggerIcon={<Package size={13} />}
                size="xl"
                className="production-detail-dialog"
                bodyClassName="production-detail-dialog-body"
              >
                <DataTable<InventoryBatchTrace>
                  title="批次追溯"
                  data={inventoryBatchTraces}
                  rowKey={(item) => item.id}
                  pageSize={6}
                  columns={[
                    { key: "trace", title: "追溯号", render: (item) => <><b>{item.traceNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                    { key: "batch", title: "生产批次", render: (item) => <><b>{item.productionBatchNo || `#${item.productionBatchId}`}</b><span className="block-text muted">原料批次 {item.batchNo || "-"}</span></> },
                    { key: "receipt", title: "入库单/库存", render: (item) => `入库 #${item.rawReceiptId || "-"} / 库存 #${item.inventoryItemId || "-"}` },
                    { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                    { key: "supplier", title: "供应商", render: (item) => recordName(supplierOptions().find((supplier) => recordId(supplier) === item.supplierId), `供应商 ${item.supplierId || "-"}`) },
                    { key: "location", title: "仓位", render: (item) => `${item.warehouse || "-"} / ${item.silo || "-"}` },
                    { key: "quantity", title: "用量", render: (item) => `${qty(item.quantity)} ${item.unit || "t"}` }
                  ]}
                  emptyText="暂无批次追溯"
                />
              </ActionDialog>
              <ActionDialog
                id="production-buffer-flows-dialog"
                title="筒仓流水"
                buttonLabel={`筒仓流水 ${plantBufferFlows.length}`}
                triggerIcon={<Factory size={13} />}
                size="xl"
                className="production-detail-dialog"
                bodyClassName="production-detail-dialog-body"
              >
                <DataTable<PlantBufferFlow>
                  title="筒仓流水"
                  data={plantBufferFlows}
                  rowKey={(item) => item.id}
                  pageSize={6}
                  columns={[
                    { key: "flow", title: "流水", render: (item) => <><b>{item.flowNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                    { key: "plant", title: "生产线/筒仓", render: (item) => <><b>{productionPlantLabel(plants.find((plant) => plant.id === item.plantId))}</b><span className="block-text muted">{item.bufferCode || `#${item.bufferId}`}</span></> },
                    { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                    { key: "direction", title: "方向", render: (item) => <StatusChip value={item.direction} /> },
                    { key: "qty", title: "数量/结存", render: (item) => `${qty(item.quantity)} / ${qty(item.balanceQty)} ${item.unit}` },
                    { key: "quality", title: "质量/含水", render: (item) => <><StatusChip value={item.qualityStatus} /><span className="block-text muted">{item.moistureRate || 0}%</span></> },
                    { key: "source", title: "来源", render: (item) => `${item.sourceType || "-"}${item.sourceId ? ` #${item.sourceId}` : ""}` },
                    { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
                    { key: "remark", title: "备注", render: (item) => item.remark || "-" }
                  ]}
                  emptyText="暂无筒仓流水"
                />
              </ActionDialog>
              <ActionDialog
                id="production-sample-tests-dialog"
                title="试样检测"
                buttonLabel={`试样检测 ${qualitySamples.length}`}
                triggerIcon={<ListChecks size={13} />}
                size="xl"
                className="production-detail-dialog"
                bodyClassName="production-detail-dialog-body"
              >
                <DataTable
                  title="试样检测"
                  data={qualitySamples}
                  rowKey={(item) => item.id}
                  pageSize={6}
                  columns={[
                    { key: "sample", title: "试样", render: (item) => <><b>{item.sampleNo}</b><span className="block-text muted">{item.ageDays} 天 / {item.sampleType}</span></> },
                    { key: "batch", title: "批次", render: (item) => batches.find((batch) => batch.id === item.batchId)?.batchNo || item.batchId },
                    { key: "planned", title: "计划检测", render: (item) => shortDateTime(item.plannedTestAt) },
                    { key: "strength", title: "强度", render: (item) => item.strength ? `${item.strength}` : "-" },
                    { key: "result", title: "结论", render: (item) => <StatusChip value={item.result} /> },
                    { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                    {
                      key: "actions",
                      title: "操作",
                      width: "110px",
                      render: (item: QualitySample) => (
                        <ActionDialog id={`quality-sample-test-${item.id}`} title="登记试样检测" buttonLabel={item.status === "completed" ? "复测" : "检测"} onOpen={() => startQualitySampleTest(item)}>
                          {qualitySampleFormView}
                        </ActionDialog>
                      )
                    }
                  ]}
                  emptyText="暂无试样"
                />
              </ActionDialog>
            </ActionGroup>
          </SectionHeader>
        </section>
      </Panel>
    );
  }

  function geoFencePolygonFromText(value: string) {
    return value
      .split(/\r?\n/)
      .map((line) => line.split(/[,，\s]+/).map((item) => item.trim()).filter(Boolean))
      .map(([longitude, latitude]) => ({ longitude: Number(longitude), latitude: Number(latitude) }))
      .filter((point) => Number.isFinite(point.longitude) && Number.isFinite(point.latitude));
  }

  function geoFencePolygonText(item: GeoFence) {
    return list(item.polygon).map((point) => `${point.longitude},${point.latitude}`).join("\n");
  }

  function resetGeoFenceForm() {
    const firstSite = selectedSiteId || firstId(siteOptions());
    const site = list(bootstrap?.sites).find((item) => item.id === firstSite);
    setGeoFenceForm({
      id: "",
      name: site ? `${site.name}围栏` : "",
      type: "site",
      siteId: firstSite ? String(firstSite) : "",
      projectId: "",
      longitude: site?.longitude ? String(site.longitude) : "",
      latitude: site?.latitude ? String(site.latitude) : "",
      radius: "300",
      shape: "circle",
      polygon: "",
      status: "active"
    });
  }

  function startGeoFenceEdit(item: GeoFence) {
    setGeoFenceForm({
      id: String(item.id),
      name: item.name || "",
      type: item.type || "site",
      siteId: item.siteId ? String(item.siteId) : "",
      projectId: item.projectId ? String(item.projectId) : "",
      longitude: item.longitude ? String(item.longitude) : "",
      latitude: item.latitude ? String(item.latitude) : "",
      radius: item.radius ? String(item.radius) : "300",
      shape: item.shape || "circle",
      polygon: geoFencePolygonText(item),
      status: item.status || "active"
    });
  }

  async function handleSaveGeoFence(event: FormEvent) {
    event.preventDefault();
    const payload: Partial<GeoFence> = {
      name: geoFenceForm.name.trim(),
      type: geoFenceForm.type,
      siteId: fieldNumber(geoFenceForm.siteId),
      projectId: fieldNumber(geoFenceForm.projectId),
      longitude: Number(geoFenceForm.longitude || 0),
      latitude: Number(geoFenceForm.latitude || 0),
      radius: fieldNumber(geoFenceForm.radius, 300),
      shape: geoFenceForm.shape,
      polygon: geoFenceForm.shape === "polygon" ? geoFencePolygonFromText(geoFenceForm.polygon) : [],
      status: geoFenceForm.status
    };
    const fenceId = fieldNumber(geoFenceForm.id);
    await runBusinessAction(`map-fence-save-${fenceId || "new"}`, fenceId ? "围栏已更新" : "围栏已新增", () => (
      fenceId ? api.updateGeoFence(fenceId, payload) : api.createGeoFence(payload)
    ));
  }

  async function loadGeoFenceEvents(fenceId?: number, vehicleId?: number) {
    await runBusinessAction(`map-fence-events-${fenceId || vehicleId || "all"}`, "围栏事件已加载", async () => {
      const events = await api.geoFenceEvents({ fenceId, vehicleId, limit: 50 });
      setGeoFenceEvents(events);
    }, null);
  }

  async function loadTrackReplay(vehicleId: number) {
    await runBusinessAction(`map-track-replay-${vehicleId}`, "轨迹回放已加载", async () => {
      const [replay, events] = await Promise.all([
        api.trackReplay(vehicleId),
        api.vehicleTrack({ vehicleId })
      ]);
      setTrackReplay(replay);
      setTrackEvents(events);
      setGeoFenceEvents(list(replay.fenceEvents));
    }, null);
  }

  async function handleMapAlarm(event: FormEvent<HTMLFormElement>, item: VehicleAlarm) {
    event.preventDefault();
    const remark = String(new FormData(event.currentTarget).get("remark") || "").trim();
    if (!remark) {
      setActionError("请输入处理备注");
      return;
    }
    await runBusinessAction(`map-alarm-handle-${item.id}`, "车辆告警已处理", () => api.handleAlarm(item.id, remark));
  }

  async function handleReportLocationBatch(event: FormEvent) {
    event.preventDefault();
    const report: LocationReportPayload = {
      deviceNo: locationBatchForm.deviceNo.trim() || undefined,
      plateNo: locationBatchForm.plateNo.trim(),
      longitude: fieldNumber(locationBatchForm.longitude),
      latitude: fieldNumber(locationBatchForm.latitude),
      speed: locationBatchForm.speed ? fieldNumber(locationBatchForm.speed) : undefined,
      direction: locationBatchForm.direction ? fieldNumber(locationBatchForm.direction) : undefined,
      mileage: locationBatchForm.mileage ? fieldNumber(locationBatchForm.mileage) : undefined,
      accStatus: locationBatchForm.accStatus ? fieldNumber(locationBatchForm.accStatus) : undefined,
      locationTime: new Date().toISOString(),
      sourceType: locationBatchForm.sourceType.trim() || "erp-console"
    };
    await runBusinessAction("map-location-batch-report", "定位批量上报已提交", async () => {
      const result = await api.reportLocationBatch([report]);
      setLocationBatchResult(result);
    });
  }

  function renderMapCenter() {
    const allLocations = visibleLatestLocations.filter((item) => matchesCurrentSite(item.currentSiteId)).filter(isValidLocation);
    const keyword = dispatchSearch.trim();
    const currentSiteFilter = selectedSiteId ? String(selectedSiteId) : siteFilter;
    const siteOptionsById = new Map<number, string>();
    siteOptions().forEach((item) => siteOptionsById.set(item.id, item.name));
    list(data.dispatch?.siteProgress).filter((item) => matchesCurrentSite(item.siteId)).forEach((item) => {
      if (item.siteId) siteOptionsById.set(item.siteId, item.siteName || `T, ${item.siteId}`);
    });
    allLocations.forEach((item) => {
      if (item.currentSiteId && !siteOptionsById.has(item.currentSiteId)) {
        siteOptionsById.set(item.currentSiteId, nameOf(bootstrap?.sites, item.currentSiteId));
      }
    });
    const mapSiteOptions = Array.from(siteOptionsById, ([id, name]) => ({ id, name }));
    const statusOptions = dispatchStatusOptions(allLocations);
    const locations = allLocations.filter((item) => (
      currentSiteFilter === "all" || String(item.currentSiteId) === currentSiteFilter
    ) && matchesDispatchSearch(item, keyword) && matchesDispatchStatus(item, dispatchStatusFilter));
    const locationByVehicleId = new Map(allLocations.map((item) => [item.vehicleId, item]));
    const mapAlarms = data.alarms.filter((item) => {
      const location = locationByVehicleId.get(item.vehicleId);
      return currentSiteFilter === "all" || !location?.currentSiteId || String(location.currentSiteId) === currentSiteFilter;
    }).sort((a, b) => {
      const statusRank = (value: VehicleAlarm) => value.status === "open" || value.status === "active" || value.status === "pending" ? 0 : 1;
      return statusRank(a) - statusRank(b) || b.id - a.id;
    });
    const siteFences = data.geoFences.filter((item) => (
      item.status !== "inactive" && (currentSiteFilter === "all" || !item.siteId || String(item.siteId) === currentSiteFilter)
    ));
    const onlineLocations = locations.filter((item) => item.onlineStatus === "online");
    const selectedLocation = locations.find((item) => item.vehicleId === selectedVehicleId) || locations[0] || null;
    const selectedReplay = trackReplay && trackReplay.vehicleId === selectedLocation?.vehicleId ? trackReplay : null;
    const selectedTrackRows = selectedReplay ? [...(trackEvents.length ? trackEvents : list(selectedReplay.points))]
      .filter((item) => item.vehicleId === selectedReplay.vehicleId)
      .sort((a, b) => (b.locationTime || b.receiveTime || "").localeCompare(a.locationTime || a.receiveTime || "")) : [];
    const matchesMapIntegration = (...values: Array<string | undefined>) => values.some((value) => /gps|iot|map|vehicle|location|定位|车辆/i.test(value || ""));
    const vehicleRules = list(data.rules?.rules).filter((item) => matchesMapIntegration(item.category, item.code, item.name, item.metric, item.description));
    const mapNotifications = [...list(data.rules?.notifications)].filter((item) => matchesMapIntegration(item.title, item.content, item.targetRole, item.channel)).sort((a, b) => b.id - a.id);
    const mapIntegrationEndpoints = list(data.integrations?.endpoints).filter((item) => matchesMapIntegration(item.name, item.type, item.protocol, item.url));
    const mapProtocolFrames = [...list(data.integrations?.protocolFrames)]
      .filter((item) => matchesMapIntegration(item.channel, item.protocol, item.deviceNo, item.parsedResource, item.actor))
      .sort((a, b) => b.id - a.id);
    const geoFenceFormView = (
      <SystemForm onSubmit={handleSaveGeoFence}>
        <Field label="名称"><TextInput value={geoFenceForm.name} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, name: event.target.value })} /></Field>
        <Field label="类型">
          <SelectInput value={geoFenceForm.type} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, type: event.target.value })}>
            <option value="site">站点</option>
            <option value="project">项目</option>
            <option value="yard">自定义</option>
          </SelectInput>
        </Field>
        <Field label="站点">
          <SelectInput value={geoFenceForm.siteId} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, siteId: event.target.value })}>
            <option value="">不关联</option>
            {siteOptions().map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="项目">
          <SelectInput value={geoFenceForm.projectId} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, projectId: event.target.value })}>
            <option value="">不关联</option>
            {list(bootstrap?.projects).map((project) => <option key={project.id} value={project.id}>{project.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="形状">
          <SelectInput value={geoFenceForm.shape} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, shape: event.target.value })}>
            <option value="circle">圆形</option>
            <option value="polygon">多边形</option>
          </SelectInput>
        </Field>
        <Field label="经度"><TextInput value={geoFenceForm.longitude} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, longitude: event.target.value })} /></Field>
        <Field label="纬度"><TextInput value={geoFenceForm.latitude} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, latitude: event.target.value })} /></Field>
        <Field label="半径"><TextInput type="number" min="1" value={geoFenceForm.radius} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, radius: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={geoFenceForm.status} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="inactive">inactive</option>
          </SelectInput>
        </Field>
        <Field label="多边形坐标" spanAll><TextAreaInput value={geoFenceForm.polygon} onChange={(event) => setGeoFenceForm({ ...geoFenceForm, polygon: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !geoFenceForm.name.trim()}>保存围栏</UiButton>
        </FormActions>
      </SystemForm>
    );

    return (
      <Panel className="map-center-view map-center-shell">
        <div className="map-center-toolbar">
          <div>
            <h3>明细</h3>
          </div>
          <ActionGroup>
            <IconField className="compact-field" icon={<Search size={14} />} label="车牌 / 工地 / 状态">
              <TextInput
                className="compact-input"
                value={dispatchSearch}
                onChange={(event) => setDispatchSearch(event.target.value)}
                placeholder="搜索"
              />
            </IconField>
            <SelectInput className="compact-select wide" value={currentSiteFilter} onChange={(event) => setSiteFilter(event.target.value)}>
              {!selectedSiteId ? <option value="all">全部</option> : null}
              {mapSiteOptions.map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
            </SelectInput>
            <SelectInput className="compact-select" value={dispatchStatusFilter} onChange={(event) => setDispatchStatusFilter(event.target.value)}>
              <option value="all">全部</option>
              {statusOptions.map((status) => <option key={status} value={status}>{statusLabel(status)}</option>)}
            </SelectInput>
            <ActionDialog id="map-fence-create" title="新增围栏" buttonLabel="新增围栏" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetGeoFenceForm}>
              {geoFenceFormView}
            </ActionDialog>
            <ActionDialog
              id="map-location-batch-report"
              title="批量定位上报"
              buttonLabel="定位上报"
              triggerIcon={<MapPin size={13} />}
              onOpen={() => {
                setLocationBatchForm({
                  deviceNo: "",
                  plateNo: selectedLocation?.plateNo || "",
                  longitude: selectedLocation ? String(selectedLocation.longitude) : "",
                  latitude: selectedLocation ? String(selectedLocation.latitude) : "",
                  speed: selectedLocation ? String(selectedLocation.speed || "") : "",
                  direction: selectedLocation ? String(selectedLocation.direction || "") : "",
                  mileage: "",
                  accStatus: "1",
                  sourceType: "erp-console"
                });
              }}
            >
              <DialogForm onSubmit={handleReportLocationBatch}>
                <Field label="车牌"><TextInput value={locationBatchForm.plateNo} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, plateNo: event.target.value })} required /></Field>
                <Field label="设备号"><TextInput value={locationBatchForm.deviceNo} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, deviceNo: event.target.value })} /></Field>
                <Field label="经度"><TextInput type="number" step="0.000001" value={locationBatchForm.longitude} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, longitude: event.target.value })} required /></Field>
                <Field label="纬度"><TextInput type="number" step="0.000001" value={locationBatchForm.latitude} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, latitude: event.target.value })} required /></Field>
                <Field label="速度"><TextInput type="number" min="0" step="0.1" value={locationBatchForm.speed} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, speed: event.target.value })} /></Field>
                <Field label="方向"><TextInput type="number" min="0" step="1" value={locationBatchForm.direction} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, direction: event.target.value })} /></Field>
                <Field label="里程"><TextInput type="number" min="0" step="0.1" value={locationBatchForm.mileage} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, mileage: event.target.value })} /></Field>
                <Field label="ACC"><TextInput type="number" min="0" step="1" value={locationBatchForm.accStatus} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, accStatus: event.target.value })} /></Field>
                <Field label="来源"><TextInput value={locationBatchForm.sourceType} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, sourceType: event.target.value })} /></Field>
                {locationBatchResult ? (
                  <MetricList compact className="span-all">
                    <div><span>总数</span><b>{locationBatchResult.total}</b></div>
                    <div><span>接受</span><b>{locationBatchResult.accepted}</b></div>
                    <div><span>拒绝</span><b>{locationBatchResult.rejected}</b></div>
                  </MetricList>
                ) : null}
                <FormActions spanAll>
                  <UiButton variant="primary" type="submit" icon={<MapPin size={14} />} disabled={actionBusy !== "" || !locationBatchForm.plateNo || !locationBatchForm.longitude || !locationBatchForm.latitude}>提交上报</UiButton>
                </FormActions>
              </DialogForm>
            </ActionDialog>
            <UiButton icon={<RefreshCw size={14} />} onClick={() => { onChanged(); load(); }} title="刷新">刷新</UiButton>
          </ActionGroup>
        </div>
        <div className="map-center-grid">
          <div className="map-canvas-panel">
            <VehicleLocationMap
              locations={locations}
              fences={siteFences}
              provider={mapConfig}
              selectedVehicleId={selectedLocation?.vehicleId || null}
              onSelectVehicle={(vehicleId) => setSelectedVehicleId(vehicleId)}
            />
            <div className="map-floating-summary">
              <MapIcon size={16} />
              <span>{mapConfig?.coordinateSystem || "wgs84"}</span>
              <StatusChip value={mapConfig?.offline ? "offline" : "active"} />
            </div>
          </div>
          <LayoutRegion as="aside" className="map-side-panel">
            <div className="map-provider-list">
              <div><span>地图服务</span><b>{mapConfig?.name || "OpenStreetMap"}</b></div>
              <div><span>服务商</span><b>{mapConfig?.provider || "osm"}</b></div>
              <div><span>坐标系</span><b>{mapConfig?.coordinateSystem || "wgs84"}</b></div>
              <div><span>在线车辆</span><b>{onlineLocations.length}/{locations.length}</b></div>
              <div><span>围栏</span><b>{siteFences.length}</b></div>
              <div><span>定位规则</span><b>{vehicleRules.filter((item) => item.enabled).length}/{vehicleRules.length}</b></div>
              <div><span>协议帧</span><b>{mapProtocolFrames.length}</b></div>
            </div>
            <div className="map-location-list">
              {locations.map((item) => (
                <SelectableCard
                  className="side-card map-location-item"
                  key={item.vehicleId}
                  selected={item.vehicleId === selectedLocation?.vehicleId}
                  onClick={() => setSelectedVehicleId(item.vehicleId)}
                >
                  <div className="side-card-head">
                    <div>
                      <b>{dispatchVehicleTitle(item)}</b>
                      <small>{nameOf(bootstrap?.sites, item.currentSiteId)} / {item.transportStatus || "idle"}</small>
                    </div>
                    <StatusChip value={item.onlineStatus} />
                  </div>
                  <div className="side-card-body">
                    <div className="map-provider-list">
                      <div><span>速度</span><b>{qty(item.speed)} km/h</b></div>
                      <div><span>方向</span><b>{qty(item.direction)}°</b></div>
                      <div><span>坐标</span><b>{item.latitude.toFixed(4)}, {item.longitude.toFixed(4)}</b></div>
                      <div><span>更新时间</span><b>{shortDateTime(item.lastLocationTime)}</b></div>
                    </div>
                    <ActionGroup>
                      <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => loadTrackReplay(item.vehicleId)}>轨迹</UiButton>
                      <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => loadGeoFenceEvents(undefined, item.vehicleId)}>事件</UiButton>
                    </ActionGroup>
                  </div>
                </SelectableCard>
              ))}
              {!locations.length ? null : null}
            </div>
          </LayoutRegion>
        </div>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="电子围栏"
            data={siteFences}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            columns={[
              { key: "name", title: "围栏", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type} / {item.shape}</span></> },
              { key: "master", title: "关联", render: (item) => item.siteId ? nameOf(bootstrap?.sites, item.siteId) : item.projectId ? nameOf(bootstrap?.projects, item.projectId) : "-" },
              { key: "center", title: "中心", render: (item) => `${qty(item.latitude)}, ${qty(item.longitude)}` },
              { key: "radius", title: "半径", render: (item) => item.shape === "circle" ? `${qty(item.radius)}m` : `${list(item.polygon).length} 点` },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "260px",
                render: (item) => (
                  <ActionGroup>
                    <ActionDialog id={`map-fence-edit-${item.id}`} title="编辑围栏" buttonLabel="编辑" onOpen={() => startGeoFenceEdit(item)}>
                      {geoFenceFormView}
                    </ActionDialog>
                    <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => loadGeoFenceEvents(item.id)}>事件</UiButton>
                    <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`map-fence-archive-${item.id}`, "围栏已归档", () => api.archiveGeoFence(item.id))}>归档</UiButton>
                  </ActionGroup>
                )
              }
            ]}
            emptyText="暂无围栏"
          />
          <DataTable
            title="围栏事件"
            data={geoFenceEvents}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "vehicle", title: "车辆", render: (item) => vehicleInternalNo(item.vehicleId, visibleLatestLocations.find((location) => location.vehicleId === item.vehicleId)?.plateNo || nameOf(bootstrap?.vehicles, item.vehicleId)) },
              { key: "fence", title: "围栏", render: (item) => data.geoFences.find((fence) => fence.id === item.fenceId)?.name || item.fenceId },
              { key: "event", title: "事件", render: (item) => <StatusChip value={item.eventType} /> },
              { key: "dispatch", title: "调度", render: (item) => item.dispatchId || "-" },
              { key: "time", title: "时间", render: (item) => shortDateTime(item.eventTime) }
            ]}
            emptyText="暂无围栏事件"
          />
          <DataTable
            title="车辆告警"
            data={mapAlarms}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "vehicle", title: "车辆", render: (item) => vehicleInternalNo(item.vehicleId, locationByVehicleId.get(item.vehicleId)?.plateNo || nameOf(bootstrap?.vehicles, item.vehicleId)) },
              { key: "type", title: "类型", render: (item) => <><b>{item.alarmType}</b><span className="block-text muted">{item.level}</span></> },
              { key: "message", title: "内容", render: (item) => item.message || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) },
              {
                key: "actions",
                title: "操作",
                width: "120px",
                render: (item: VehicleAlarm) => item.status === "handled" ? (
                  <span className="muted">{item.handledBy || "已处理"}</span>
                ) : (
                  <ActionDialog id={`map-alarm-${item.id}`} title="处理车辆告警" buttonLabel="处理" triggerIcon={<AlertCircle size={13} />}>
                    <InlineForm onSubmit={(event) => handleMapAlarm(event, item)}>
                      <Field label="处理备注"><TextInput name="remark" defaultValue="" /></Field>
                      <UiButton type="submit" variant="primary" disabled={actionBusy !== ""}>确认处理</UiButton>
                    </InlineForm>
                  </ActionDialog>
                )
              }
            ]}
            emptyText="暂无车辆告警"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="定位规则"
            data={vehicleRules}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<RuleDefinition>({
              actions: [
                { key: "focus-metric", label: "只看该指标", onSelect: (item, helpers) => helpers.searchText(item.metric) },
                { key: "focus-level", label: "只看该等级", onSelect: (item, helpers) => helpers.searchText(item.level) }
              ],
              copyFields: [
                { key: "code", label: "规则编码", value: (item) => item.code },
                { key: "condition", label: "规则条件", value: (item) => `${item.metric || "-"} ${item.operator || ""} ${item.threshold ?? ""}` },
                { key: "description", label: "说明", value: (item) => item.description }
              ]
            })}
            columns={[
              { key: "rule", title: "规则", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code} / {item.category}</span></> },
              { key: "condition", title: "条件", render: (item) => `${item.metric || "-"} ${item.operator || ""} ${item.threshold ?? ""}` },
              { key: "level", title: "等级", render: (item) => <StatusChip value={item.level || "info"} /> },
              { key: "notify", title: "通知角色", render: (item) => list(item.notifyRoles).join(" / ") || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> }
            ]}
            emptyText="暂无定位规则"
          />
          <DataTable
            title="定位通知"
            data={mapNotifications}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<NotificationItem>({
              actions: [
                { key: "focus-role", label: "只看该角色", onSelect: (item, helpers) => helpers.searchText(roleName(item.targetRole)) },
                { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
              ],
              copyFields: [
                { key: "title", label: "通知标题", value: (item) => item.title },
                { key: "content", label: "通知内容", value: (item) => item.content }
              ]
            })}
            columns={[
              { key: "title", title: "通知", render: (item) => <><b>{item.title}</b><span className="block-text muted">{item.content || "-"}</span></> },
              { key: "target", title: "目标", render: (item) => <><span>{roleName(item.targetRole)}</span><span className="block-text muted">{item.channel}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无定位通知"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="定位集成端点"
            data={mapIntegrationEndpoints}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<IntegrationEndpoint>({
              actions: [
                { key: "focus-type", label: "只看该类型", onSelect: (item, helpers) => helpers.searchText(item.type) },
                { key: "focus-protocol", label: "只看该协议", onSelect: (item, helpers) => helpers.searchText(item.protocol) }
              ],
              copyFields: [
                { key: "name", label: "端点名称", value: (item) => item.name },
                { key: "url", label: "URL", value: (item) => item.url }
              ]
            })}
            columns={[
              { key: "endpoint", title: "端点", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type} / {item.protocol}</span></> },
              { key: "url", title: "地址", render: (item) => item.url || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastSyncAt", title: "最近同步", render: (item) => shortDateTime(item.lastSyncAt) }
            ]}
            emptyText="暂无定位集成端点"
          />
          <DataTable
            title="定位协议帧"
            data={mapProtocolFrames}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<DeviceProtocolFrame>({
              actions: [
                { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceNo) },
                { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
              ],
              copyFields: [
                { key: "frame", label: "帧编号", value: (item) => item.frameNo },
                { key: "raw", label: "原始报文", value: (item) => item.raw }
              ]
            })}
            columns={[
              { key: "frame", title: "帧", render: (item) => <><b>{item.frameNo}</b><span className="block-text muted">{item.channel} / {item.protocol}</span></> },
              { key: "device", title: "设备", render: (item) => item.deviceNo || "-" },
              { key: "parsed", title: "解析对象", render: (item) => item.parsedResource ? `${item.parsedResource}#${item.parsedId || 0}` : "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "error", title: "错误", render: (item) => item.error || "-" },
              { key: "receivedAt", title: "接收时间", render: (item) => shortDateTime(item.receivedAt) }
            ]}
            emptyText="暂无定位协议帧"
          />
        </SectionGrid>
        {selectedReplay ? (
          <MetricList compact className="system-summary-grid">
            <div><span>轨迹车辆</span><b>{selectedReplay.plateNo}</b></div>
            <div><span>距离</span><b>{qty(selectedReplay.distanceKm)} km</b></div>
            <div><span>时长</span><b>{qty(selectedReplay.durationMinutes)} 分钟</b></div>
            <div><span>均速</span><b>{qty(selectedReplay.averageSpeed)} km/h</b></div>
            <div><span>最高速度</span><b>{qty(selectedReplay.maxSpeed)} km/h</b></div>
            <div><span>轨迹点</span><b>{list(selectedReplay.compressedPoints).length}/{list(selectedReplay.points).length}</b></div>
          </MetricList>
        ) : null}
        {selectedReplay ? (
          <DataTable
            title="轨迹明细"
            data={selectedTrackRows}
            rowKey={(item) => item.id || `${item.vehicleId}-${item.locationTime}-${item.longitude}-${item.latitude}`}
            pageSize={8}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<VehicleLocationEvent>({
              actions: [
                { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceId) },
                { key: "focus-source", label: "只看该来源", onSelect: (item, helpers) => helpers.searchText(item.sourceType) }
              ],
              copyFields: [
                { key: "plate", label: "车牌", value: (item) => item.plateNo },
                { key: "device", label: "设备", value: (item) => item.deviceId },
                { key: "location", label: "坐标", value: (item) => `${item.latitude}, ${item.longitude}` }
              ]
            })}
            columns={[
              { key: "time", title: "定位时间", render: (item) => <><b>{shortDateTime(item.locationTime)}</b><span className="block-text muted">{shortDateTime(item.receiveTime)}</span></> },
              { key: "device", title: "设备", render: (item) => <><span>{item.deviceId || "-"}</span><span className="block-text muted">{item.sourceType || "-"}</span></> },
              { key: "coordinate", title: "坐标", render: (item) => `${qty(item.latitude)}, ${qty(item.longitude)}` },
              { key: "speed", title: "速度/方向", render: (item) => `${qty(item.speed)} km/h / ${qty(item.direction)}°` },
              { key: "mileage", title: "里程", render: (item) => `${qty(item.mileage)} km` },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.isAbnormal ? item.abnormalType || "abnormal" : item.onlineStatus || "normal"} /> },
              { key: "address", title: "地址", render: (item) => item.address || "-" }
            ]}
            emptyText="暂无轨迹明细"
          />
        ) : null}
      </Panel>
    );
  }

  function renderDispatch() {
    const dispatch = data.dispatch;
    const kpis = dispatch?.kpis;
    const allProgress = list(dispatch?.siteProgress).filter((item) => matchesCurrentSite(item.siteId));
    const keyword = dispatchSearch.trim();
    const currentSiteFilter = selectedSiteId ? String(selectedSiteId) : siteFilter;
    const sites = dispatchSiteOptions(allProgress);
    const allVehicles = list(dispatch?.availableVehicles).filter((item) => matchesCurrentSite(item.siteId));
    const allQueueRows = list(dispatch?.vehicleQueue).filter((item) => matchesCurrentSite(item.siteId));
    const allLocations = list(dispatch?.latestLocations).filter((item) => matchesCurrentSite(item.currentSiteId));
    const statusOptions = dispatchStatusOptions([...allProgress, ...allVehicles, ...allQueueRows, ...allLocations]);
    const progressRows = allProgress.filter((item) => (currentSiteFilter === "all" || String(item.siteId) === currentSiteFilter) && matchesDispatchSearch(item, keyword) && matchesDispatchStatus(item, dispatchStatusFilter));
    const vehicles = allVehicles.filter((item) => (currentSiteFilter === "all" || String(item.siteId) === currentSiteFilter) && matchesDispatchSearch(item, keyword) && matchesDispatchStatus(item, dispatchStatusFilter));
    const hasDispatchFilters = keyword !== "" || currentSiteFilter !== "all" || dispatchStatusFilter !== "all";
    const vehicleOptions = vehicles.length || hasDispatchFilters ? vehicles : allVehicles;
    const queueRows = allQueueRows.filter((item) => (currentSiteFilter === "all" || String(item.siteId) === currentSiteFilter) && matchesDispatchSearch(item, keyword) && matchesDispatchStatus(item, dispatchStatusFilter));
    const locations = allLocations.filter((item) => (currentSiteFilter === "all" || String(item.currentSiteId) === currentSiteFilter) && matchesDispatchStatus(item, dispatchStatusFilter));
    const selectedOrder = progressRows.find((item) => item.orderId === selectedOrderId) || progressRows[0] || allProgress.find((item) => item.orderId === selectedOrderId) || allProgress[0];
    const selectedVehicle = vehicleOptions.find((item) => item.vehicleId === selectedVehicleId) || vehicleOptions[0];
    const selectedQueue = selectedOrder ? queueRows.filter((item) => item.orderId === selectedOrder.orderId) : queueRows;
    const orderOptions = progressRows.length || hasDispatchFilters ? progressRows : allProgress;
    const displayOrders = progressRows.length || hasDispatchFilters ? progressRows : allProgress;
    const defaultDispatchQty = selectedOrder ? Math.min(36, selectedOrder.remainingQty || selectedOrder.planQuantity || 0) : 0;
    const dispatchLinePlants = productionPlants().filter((item) => matchesCurrentSite(item.siteId));
    const dispatchLineTasks = list(data.production?.tasks).filter((item) => matchesCurrentSite(item.siteId) && item.status !== "cancelled" && item.status !== "completed");
    const dispatchLineRows = dispatchLinePlants.map((plant) => {
      const orderIds = new Set(dispatchLineTasks.filter((task) => productionTaskPlant(task)?.id === plant.id).map((task) => task.orderId));
      const lineQueue = queueRows.filter((item) => item.siteId === plant.siteId && orderIds.has(item.orderId));
      return {
        key: String(plant.id),
        label: plant.name || plant.code,
        loading: lineQueue.filter((item) => item.status === "loading"),
        waiting: lineQueue.filter((item) => item.status !== "loading")
      };
    });
    const matchedLineDispatchIds = new Set(dispatchLineRows.flatMap((item) => [...item.loading, ...item.waiting]).map((item) => item.dispatchId));
    const unmatchedLineQueue = queueRows.filter((item) => !matchedLineDispatchIds.has(item.dispatchId) && ["assigned", "accepted", "arrived_site", "waiting_load", "loading"].includes(item.status));
    if (unmatchedLineQueue.length) {
      dispatchLineRows.push({
        key: "unmatched",
        label: "未配",
        loading: unmatchedLineQueue.filter((item) => item.status === "loading"),
        waiting: unmatchedLineQueue.filter((item) => item.status !== "loading")
      });
    }
    const returningVehicles = locations.filter((item) => item.transportStatus === "returning");
    const repairVehicles = allVehicles.filter((item) => item.onlineStatus !== "online" || item.businessStatus === "maintenance" || item.businessStatus === "repair");
    const noOrderVehicles = vehicleOptions.filter((item) => !queueRows.some((queue) => queue.vehicleId === item.vehicleId));
    const vehicleGroups = [
      { key: "queue", label: "排队", items: queueRows },
      { key: "returning", label: "回厂", items: returningVehicles },
      { key: "task", label: "任务", items: queueRows.filter((item) => item.status !== "signed") },
      { key: "repair", label: "休修", items: repairVehicles },
      { key: "no-order", label: "无单", items: noOrderVehicles }
    ];
    const activeVehicleGroup = vehicleGroups.find((item) => item.key === dispatchVehicleGroup) || vehicleGroups[0];
    const scheduleVehicles = list(bootstrap?.vehicles).filter((item) => matchesCurrentSite(item.siteId));
    const scheduleRows = data.dispatchSchedules.filter((item) => matchesCurrentSite(item.siteId)).sort((a, b) => b.id - a.id);
    const rawPortalDispatchRows = list(data.portal?.dispatches);
    const portalDispatchRows = rawPortalDispatchRows.filter((item) => !selectedSiteId || matchesCurrentSite(item.siteId)).sort((a, b) => b.id - a.id);
    const dispatchScheduleFormView = (
      <DialogForm onSubmit={handleCreateDispatchSchedule}>
        <Field label="站点">
          <SelectInput value={dispatchScheduleForm.siteId} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, siteId: event.target.value })}>
            {siteOptions().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="车辆">
          <SelectInput value={dispatchScheduleForm.vehicleId} onChange={(event) => {
            const vehicle = scheduleVehicles.find((item) => item.id === fieldNumber(event.target.value));
            setDispatchScheduleForm({
              ...dispatchScheduleForm,
              vehicleId: event.target.value,
              driverId: vehicle?.driverId ? String(vehicle.driverId) : dispatchScheduleForm.driverId
            });
          }}>
            {scheduleVehicles.map((item) => <option key={item.id} value={item.id}>{item.internalNo || item.plateNo} / {item.plateNo}</option>)}
          </SelectInput>
        </Field>
        <Field label="司机">
          <SelectInput value={dispatchScheduleForm.driverId} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, driverId: event.target.value })}>
            <option value="">不指定</option>
            {list(bootstrap?.drivers).map((item) => <option key={item.id} value={item.id}>{item.name} / {item.phone}</option>)}
          </SelectInput>
        </Field>
        <Field label="承运商">
          <SelectInput value={dispatchScheduleForm.carrierId} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, carrierId: event.target.value })}>
            <option value="">不指定</option>
            {list(bootstrap?.carriers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <HeroDateField label="排班日期" value={dispatchScheduleForm.shiftDate} onChange={(shiftDate) => setDispatchScheduleForm({ ...dispatchScheduleForm, shiftDate })} />
        <Field label="班次"><TextInput value={dispatchScheduleForm.shift} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, shift: event.target.value })} /></Field>
        <Field label="运力"><TextInput type="number" min="0" step="0.1" value={dispatchScheduleForm.capacityQty} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, capacityQty: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={dispatchScheduleForm.status} onChange={(event) => setDispatchScheduleForm({ ...dispatchScheduleForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="planned">planned</option>
            <option value="inactive">inactive</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Truck size={14} />} disabled={actionBusy !== "" || !fieldNumber(dispatchScheduleForm.siteId) || !fieldNumber(dispatchScheduleForm.vehicleId)}>保存排班</UiButton>
        </FormActions>
      </DialogForm>
    );
    const portalExceptionFormView = (
      <DialogForm onSubmit={handleReportPortalDispatchException}>
        <Field label="派车单">
          <SelectInput value={portalExceptionForm.dispatchId} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, dispatchId: event.target.value })}>
            {portalDispatchRows.map((item) => <option key={item.id} value={item.id}>{item.dispatchNo} / {nameOf(bootstrap?.vehicles, item.vehicleId)}</option>)}
          </SelectInput>
        </Field>
        <Field label="等级">
          <SelectInput value={portalExceptionForm.level} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, level: event.target.value })}>
            <option value="low">low</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
          </SelectInput>
        </Field>
        <Field label="类型"><TextInput value={portalExceptionForm.alarmType} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, alarmType: event.target.value })} /></Field>
        <Field label="异常说明" spanAll><TextAreaInput value={portalExceptionForm.exception} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, exception: event.target.value })} required /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<AlertCircle size={14} />} disabled={actionBusy !== "" || !fieldNumber(portalExceptionForm.dispatchId) || !portalExceptionForm.exception.trim()}>上报异常</UiButton>
        </FormActions>
      </DialogForm>
    );

    function renderLineVehiclePill(item: DispatchCenterQueueItem) {
      return (
        <button
          className={`dispatch-line-vehicle ${item.vehicleId === selectedVehicleId ? "selected" : ""}`}
          key={item.dispatchId}
          type="button"
          title={`${dispatchVehicleMeta(item)} / ${statusLabel(item.status)}`}
          onClick={() => {
            setSelectedVehicleId(item.vehicleId);
            setSelectedOrderId(item.orderId);
          }}
        >
          {dispatchVehicleTitle(item)}
        </button>
      );
    }

    function renderVehiclePoolCell(item: DispatchCenterQueueItem | DispatchCenterVehicle | LatestLocation) {
      const key = "dispatchId" in item ? item.dispatchId : item.vehicleId;
      const active = item.vehicleId === selectedVehicleId;
      const meta = "dispatchId" in item ? item.queueNo || statusLabel(item.status) : "speed" in item ? `${qty(item.speed)} km/h` : item.driverName || item.siteName || item.carrier;
      return (
        <button
          className={`dispatch-pool-cell ${active ? "selected" : ""}`}
          key={`${activeVehicleGroup.key}-${key}`}
          type="button"
          title={dispatchVehicleMeta(item)}
          onClick={() => {
            setSelectedVehicleId(item.vehicleId);
            if ("orderId" in item) setSelectedOrderId(item.orderId);
          }}
        >
          <b>{dispatchVehicleTitle(item)}</b>
          <span>{meta || "-"}</span>
        </button>
      );
    }

    function quickDispatchForm() {
      return (
        <QuickForm className="dispatch-quick-form" onSubmit={handleQuickDispatch}>
          <Field label="订单">
            <SelectInput value={selectedOrder?.orderId || ""} onChange={(event) => setSelectedOrderId(Number(event.target.value))}>
              {orderOptions.map((item) => (
                <option key={item.orderId} value={item.orderId}>{item.orderNo} / {item.projectName}</option>
              ))}
            </SelectInput>
          </Field>
          <Field label="车辆">
            <SelectInput value={selectedVehicle?.vehicleId || ""} onChange={(event) => setSelectedVehicleId(Number(event.target.value))}>
              {vehicleOptions.map((item) => (
                <option key={item.vehicleId} value={item.vehicleId}>{dispatchVehicleTitle(item)} / {item.plateNo} / {item.driverName}</option>
              ))}
            </SelectInput>
          </Field>
          <div className="dispatch-inline-action">
            <Field label="派车数量">
              <TextInput
                type="number"
                min="0"
                step="0.5"
                value={dispatchQty}
                onChange={(event) => setDispatchQty(event.target.value)}
                placeholder={defaultDispatchQty ? String(defaultDispatchQty) : "0"}
              />
            </Field>
            <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={dispatchSubmitting || !selectedOrder || !selectedVehicle}>派车</UiButton>
          </div>
          <div className="quick-form-summary">
            <span>可派 {qty(selectedOrder?.remainingQty)} {selectedOrder?.unit || "t"}</span>
            <span>{selectedVehicle ? `${dispatchVehicleTitle(selectedVehicle)} / ${selectedVehicle.plateNo} / ${selectedVehicle.driverName}` : "未选择车辆"}</span>
          </div>
        </QuickForm>
      );
    }

    if (section === "dispatch-schedules") {
      const activeSchedules = scheduleRows.filter((item) => item.status === "active");
      const plannedCapacity = scheduleRows.reduce((sum, item) => sum + item.capacityQty, 0);
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="调度排班范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{scheduleRows.length} 条排班</span>
                <span className="inventory-scope-chip">{activeSchedules.length} 条启用</span>
              </div>
              <ActionGroup>
                <ActionDialog id="dispatch-schedule-create-page" title="新增车辆排班" buttonLabel="新增排班" triggerIcon={<Truck size={14} />} triggerVariant="primary">
                  {dispatchScheduleFormView}
                </ActionDialog>
                <ButtonLink icon={<Route size={15} />} href="/fulfillment/dispatch/queue">装料队列</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>排班车辆</span><b>{scheduleRows.length}</b><small>当前站点范围</small></div>
              <div><span>启用</span><b>{activeSchedules.length}</b><small>可参与调度</small></div>
              <div><span>计划运力</span><b>{qty(plannedCapacity)}</b><small>排班容量合计</small></div>
              <div><span>在线车辆</span><b>{kpis?.onlineVehicles || 0}</b><small>{kpis?.totalVehicles || 0} 台总车辆</small></div>
              <div><span>排队</span><b>{queueRows.length}</b><small>装料队列车辆</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DispatchSchedule>
              title="车辆排班"
              data={scheduleRows}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "schedule", title: "排班", render: (item) => <><b>{item.shiftDate}</b><span className="block-text muted">{item.shift || "-"}</span></> },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "vehicle", title: "车辆", render: (item) => nameOf(bootstrap?.vehicles, item.vehicleId) || `车辆 #${item.vehicleId}` },
                { key: "driver", title: "司机", render: (item) => item.driverId ? nameOf(bootstrap?.drivers, item.driverId) : "-" },
                { key: "carrier", title: "承运商", render: (item) => item.carrierId ? nameOf(bootstrap?.carriers, item.carrierId) : "-" },
                { key: "capacity", title: "运力", render: (item) => qty(item.capacityQty) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) }
              ]}
              emptyText="暂无车辆排班"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "dispatch-queue") {
      const loadingRows = queueRows.filter((item) => item.status === "loading");
      const waitingRows = queueRows.filter((item) => item.status !== "loading");
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="装料队列范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{queueRows.length} 台排队车</span>
                <span className={`inventory-scope-chip ${loadingRows.length ? "warning" : ""}`}>{loadingRows.length} 台装料中</span>
              </div>
              <ActionGroup>
                <UiButton variant="primary" icon={<Plus size={14} />} disabled={dispatchSubmitting || !selectedOrder || !selectedVehicle} onClick={() => setDispatchDialogOpen(true)}>快速派车</UiButton>
                <ButtonLink icon={<ClipboardCheck size={15} />} href="/fulfillment/dispatch/schedules">调度排班</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData} disabled={dispatchSubmitting}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>排队车辆</span><b>{queueRows.length}</b><small>{displayOrders.length} 个订单</small></div>
              <div><span>装料中</span><b>{loadingRows.length}</b><small>可下发装料完成</small></div>
              <div><span>等待中</span><b>{waitingRows.length}</b><small>等待装料</small></div>
              <div><span>在线车辆</span><b>{kpis?.onlineVehicles || 0}</b><small>车辆池</small></div>
              <div><span>运输中</span><b>{kpis?.inTransitVehicles || 0}</b><small>执行中派车</small></div>
            </div>
            {quickDispatchForm()}
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DispatchCenterQueueItem>
              title="装料队列"
              data={queueRows}
              rowKey={(item) => item.dispatchId}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "queue", title: "队列", render: (item) => <><b>{item.queueNo || item.dispatchNo}</b><span className="block-text muted">{etaText(item)}</span></> },
                { key: "order", title: "订单/项目", render: (item) => <><span>{item.orderNo}</span><span className="block-text muted">{item.projectName}</span></> },
                { key: "vehicle", title: "车辆", render: (item) => <><b>{dispatchVehicleTitle(item)}</b><span className="block-text muted">{dispatchVehicleMeta(item)}</span></> },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "position", title: "位置", render: (item) => queueRows.findIndex((row) => row.dispatchId === item.dispatchId) + 1 },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "220px",
                  render: (item) => {
                    const lineQueue = sortedProductionQueue(queueRows.filter((row) => row.orderId === item.orderId));
                    const index = lineQueue.findIndex((row) => row.dispatchId === item.dispatchId);
                    return (
                      <ActionGroup className="compact-actions">
                        <UiButton size="sm" disabled={dispatchSubmitting || index <= 0} onClick={() => moveProductionQueueItem(lineQueue, item.dispatchId, -1)}>上移</UiButton>
                        <UiButton size="sm" disabled={dispatchSubmitting || index < 0 || index === lineQueue.length - 1} onClick={() => moveProductionQueueItem(lineQueue, item.dispatchId, 1)}>下移</UiButton>
                        <UiButton size="sm" variant="primary" disabled={dispatchSubmitting} onClick={() => handleProductionQueueStatus(item, item.status === "loading" ? "loaded" : "loading")}>{item.status === "loading" ? "装完" : "装料"}</UiButton>
                        <UiButton size="sm" disabled={dispatchSubmitting} onClick={() => handleAdvanceDispatch(item)}>下号</UiButton>
                      </ActionGroup>
                    );
                  }
                }
              ]}
              emptyText="暂无装料队列"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <Panel className="dispatch-center dispatch-board">
        <div className="dispatch-board-toolbar">
          <div className="dispatch-toolbar-title">
            <span className="dispatch-toolbar-icon"><Route size={16} /></span>
            <div>
              <h3>调度明细</h3>
              <p>{displayOrders.length} 个订单 / {vehicleOptions.length} 台可用车 / {queueRows.length} 台排队车</p>
            </div>
          </div>
          <div className="dispatch-board-controls">
            <div className="dispatch-kpi-strip">
              <div className="compact-kpi"><span>在线车辆</span><b>{kpis?.onlineVehicles || 0}/{kpis?.totalVehicles || 0}</b></div>
              <div className="compact-kpi"><span>排队车辆</span><b>{kpis?.queueVehicles || 0}</b></div>
              <div className="compact-kpi"><span>装料中</span><b>{kpis?.loadingVehicles || 0}</b></div>
              <div className="compact-kpi"><span>运输中</span><b>{kpis?.inTransitVehicles || 0}</b></div>
              <div className="compact-kpi"><span>执行调度</span><b>{kpis?.activeDispatches || 0}</b></div>
            </div>
            <IconField className="compact-field" icon={<Search size={14} />} label="订单 / 工地 / 车牌">
              <TextInput
                className="compact-input"
                value={dispatchSearch}
                onChange={(event) => setDispatchSearch(event.target.value)}
                placeholder="搜索"
              />
            </IconField>
            <SelectInput className="compact-select wide" value={currentSiteFilter} onChange={(event) => setSiteFilter(event.target.value)}>
              {!selectedSiteId ? <option value="all">全部</option> : null}
              {sites.map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
            </SelectInput>
            <SelectInput className="compact-select" value={dispatchStatusFilter} onChange={(event) => setDispatchStatusFilter(event.target.value)}>
              <option value="all">全部</option>
              {statusOptions.map((status) => <option key={status} value={status}>{statusLabel(status)}</option>)}
            </SelectInput>
            <UiButton
              variant="primary"
              icon={<Plus size={14} />}
              onClick={() => setDispatchDialogOpen(true)}
              disabled={dispatchSubmitting || !selectedOrder || !selectedVehicle}
              title="快速派车"
            >
              快速派车
            </UiButton>
            <UiButton icon={<RefreshCw size={14} />} onClick={refreshData} disabled={dispatchSubmitting} title="刷新">刷新</UiButton>
          </div>
        </div>

        <div className="dispatch-board-grid">
          <LayoutRegion as="aside" className="dispatch-yard">
            <div className="dispatch-console">
              <div className="dispatch-line-table">
                <div className="dispatch-line-header">
                  <span>产线</span>
                  <span>装料</span>
                  <span>准备</span>
                </div>
                {dispatchLineRows.map((line) => (
                  <div className="dispatch-line-row" key={line.key}>
                    <b>{line.label}</b>
                    <div className="dispatch-line-vehicles">
                      {line.loading.slice(0, 4).map(renderLineVehiclePill)}
                      {!line.loading.length ? <span className="dispatch-line-empty">-</span> : null}
                    </div>
                    <div className="dispatch-line-vehicles">
                      {line.waiting.slice(0, 4).map(renderLineVehiclePill)}
                      {!line.waiting.length ? <span className="dispatch-line-empty">-</span> : null}
                    </div>
                  </div>
                ))}
                {!dispatchLineRows.length ? (
                  <div className="dispatch-line-row muted">
                    <b>暂无</b>
                    <span>无生产线</span>
                    <span>无车辆</span>
                  </div>
                ) : null}
              </div>

              <div className="dispatch-pool">
                <div className="dispatch-pool-tabs">
                  {vehicleGroups.map((group) => (
                    <button
                      className={`dispatch-pool-tab ${activeVehicleGroup.key === group.key ? "active" : ""}`}
                      key={group.key}
                      type="button"
                      onClick={() => setDispatchVehicleGroup(group.key)}
                    >
                      <span>{group.label}</span>
                      <b>{group.items.length}</b>
                    </button>
                  ))}
                </div>
                <div className="dispatch-pool-grid">
                  {activeVehicleGroup.items.slice(0, 60).map(renderVehiclePoolCell)}
                  {!activeVehicleGroup.items.length ? <span className="dispatch-pool-empty">暂无车辆</span> : null}
                </div>
              </div>

              <div className="dispatch-console-actions">
                <TextInput
                  className="dispatch-console-input"
                  value={dispatchSearch}
                  onChange={(event) => setDispatchSearch(event.target.value)}
                  placeholder="车号"
                />
                <UiButton size="sm" onClick={() => setDispatchVehicleGroup("queue")}>查询</UiButton>
                <UiButton size="sm" variant="primary" onClick={() => setDispatchDialogOpen(true)} disabled={dispatchSubmitting || !selectedOrder || !selectedVehicle}>上号</UiButton>
                <UiButton size="sm" onClick={() => selectedQueue[0] ? handleAdvanceDispatch(selectedQueue[0]) : undefined} disabled={dispatchSubmitting || !selectedQueue.length}>下号</UiButton>
                <UiButton size="sm" onClick={() => setDispatchVehicleGroup("no-order")}>排班</UiButton>
                <UiButton size="sm" onClick={() => setDispatchVehicleGroup("queue")}>排队</UiButton>
                <UiButton size="sm" onClick={refreshData}>日志</UiButton>
                <UiButton size="sm" onClick={() => setDispatchVehicleGroup("repair")}>黑名单</UiButton>
              </div>
            </div>

          </LayoutRegion>

          <LayoutRegion as="main" className="dispatch-visual-main">
            <SectionHeader className="panel-head-compact">
              <div>
                <b>订单路线</b>
                <span>{displayOrders.length} 个订单 / {queueRows.length} 台车</span>
              </div>
              <div className="dispatch-legend">
                <span><i className="legend-dot ready" />待派</span>
                <span><i className="legend-dot queue" />排队</span>
                <span><i className="legend-dot production" />装料</span>
                <span><i className="legend-dot transit" />运输</span>
                <span><i className="legend-dot arrive" />到达</span>
              </div>
            </SectionHeader>

            <div className="dispatch-card-grid">
              {displayOrders.map((order) => {
                const itemQueue = queueRows.filter((item) => item.orderId === order.orderId);
                const laneVehicleGroups = dispatchLaneVehicleGroups(itemQueue);
                const nextEta = itemQueue[0] ? etaText(itemQueue[0]) : shortDateTime(order.nextEta);
                return (
                  <SelectableCard
                    key={order.orderId}
                    className="dispatch-visual-card"
                    selected={selectedOrder?.orderId === order.orderId}
                    onClick={() => setSelectedOrderId(order.orderId)}
                  >
                    <div className="visual-card-head">
                      <div>
                        <b>{order.orderNo}</b>
                        <span>{order.customerName} / {order.projectName}</span>
                      </div>
                      <StatusChip value={order.status} />
                    </div>

                    <div className="visual-card-metrics">
                      <div><span>计划</span><b>{qty(order.planQuantity)}</b></div>
                      <div><span>剩余</span><b>{qty(order.remainingQty)}</b></div>
                      <div><span>已装</span><b>{qty(order.loadedQty)}</b></div>
                      <div><span>已签</span><b>{qty(order.signedQty)}</b></div>
                    </div>

                    <div className="dispatch-lane">
                      <div className="lane-line">
                        <span className="lane-node lane-origin" title={order.siteName} role="img" aria-label={`起点：${order.siteName}`}>
                          <Factory size={13} aria-hidden="true" />
                        </span>
                        <span className="lane-node lane-destination" title={order.projectName} role="img" aria-label={`终点：${order.projectName}`}>
                          <MapPin size={13} aria-hidden="true" />
                        </span>
                        {laneVehicleGroups.slice(0, 6).map((group) => {
                          const firstVehicle = group.items[0];
                          return (
                            <span
                              key={group.key}
                              className={`lane-vehicle ${dispatchStageClass(firstVehicle.status)}${group.items.length > 1 ? " grouped" : ""}`}
                              style={{ left: `${group.position}%` }}
                              title={group.items.map((item) => `${dispatchVehicleTitle(item)} / ${item.dispatchNo} / ${item.driverName} / ${statusLabel(item.status)}`).join("\n")}
                            >
                              <span className="lane-vehicle-label">{dispatchVehicleTitle(firstVehicle)}</span>
                              {group.items.length > 1 ? <b className="lane-vehicle-count">+{group.items.length - 1}</b> : null}
                            </span>
                          );
                        })}
                      </div>
                      <div className="lane-stage-labels">
                        <span>派车</span>
                        <span>到站</span>
                        <span>装料</span>
                        <span>出站</span>
                        <span>运输</span>
                        <span>签收</span>
                      </div>
                    </div>

                    <div className="visual-card-foot">
                      <span>{order.productName} / 生产 {percent(order.producedPercent)} / 派车 {percent(order.dispatchedPercent)}</span>
                      <span className="eta-pill"><Clock size={12} />{nextEta}</span>
                    </div>
                  </SelectableCard>
                );
              })}
              {!displayOrders.length ? (
                <EmptyState className="empty-visual-board" title="暂无订单路线">
                  调整搜索、站点或状态筛选
                </EmptyState>
              ) : null}
            </div>
          </LayoutRegion>

        </div>

        <SectionGrid as="div" className="finance-list-page dispatch-detail-grid">
          <div className="span-6">
            <DataTable
              title="车辆排班"
              data={scheduleRows}
              rowKey={(item) => item.id}
              pageSize={6}
              headerLeftAction={(
                <ActionDialog
                  id="dispatch-schedule-create"
                  title="新增车辆排班"
                  buttonLabel="新增排班"
                  triggerIcon={<Plus size={13} />}
                  onOpen={() => {
                    const vehicle = scheduleVehicles.find((item) => item.id === selectedVehicleId) || scheduleVehicles[0];
                    setDispatchScheduleForm({
                      siteId: String(vehicle?.siteId || selectedSiteId || defaultSiteId || firstId(bootstrap?.sites)),
                      vehicleId: String(vehicle?.id || ""),
                      driverId: vehicle?.driverId ? String(vehicle.driverId) : "",
                      carrierId: "",
                      shiftDate: today,
                      shift: "早班",
                      capacityQty: String(vehicle?.capacity || "36"),
                      status: "active"
                    });
                  }}
                >
                  {dispatchScheduleFormView}
                </ActionDialog>
              )}
              columns={[
                { key: "scheduleNo", title: "排班单", render: (item) => <><b>{item.scheduleNo}</b><span className="block-text muted">{item.shiftDate} / {item.shift}</span></> },
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "vehicle", title: "车辆", render: (item) => nameOf(bootstrap?.vehicles, item.vehicleId) },
                { key: "driver", title: "司机", render: (item) => nameOf(bootstrap?.drivers, item.driverId) },
                { key: "capacity", title: "运力", render: (item) => `${qty(item.assignedQty)} / ${qty(item.capacityQty)}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
              ]}
              emptyText="暂无车辆排班"
            />
          </div>
          <div className="span-6">
            <DataTable
              title="门户派车异常"
              data={portalDispatchRows}
              rowKey={(item) => item.id}
              pageSize={6}
              headerLeftAction={(
                <ActionDialog
                  id="portal-dispatch-exception"
                  title="上报派车异常"
                  buttonLabel="异常上报"
                  triggerIcon={<AlertCircle size={13} />}
                  onOpen={() => setPortalExceptionForm({
                    dispatchId: String(portalDispatchRows[0]?.id || ""),
                    exception: "",
                    level: "medium",
                    alarmType: "delay"
                  })}
                  disabled={!portalDispatchRows.length}
                >
                  {portalExceptionFormView}
                </ActionDialog>
              )}
              columns={[
                { key: "dispatchNo", title: "派车单", render: (item) => <b>{item.dispatchNo}</b> },
                { key: "vehicle", title: "车辆", render: (item) => nameOf(bootstrap?.vehicles, item.vehicleId) },
                { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
                { key: "exception", title: "异常", render: (item) => item.exception || "-" },
                { key: "eta", title: "ETA", render: (item) => shortDateTime(item.eta) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
              ]}
              emptyText="暂无门户派车"
            />
          </div>
        </SectionGrid>

        {dispatchDialogOpen ? (
          <Dialog
            open
            title="快速派车"
            description={selectedOrder ? `${selectedOrder.orderNo} / ${selectedOrder.projectName}` : "选择订单和车辆后提交派车"}
            className="dispatch-quick-dialog"
            bodyClassName="dispatch-quick-dialog-body"
            closeDisabled={dispatchSubmitting}
            onClose={() => setDispatchDialogOpen(false)}
          >
            {quickDispatchForm()}
          </Dialog>
        ) : null}
      </Panel>
    );
  }

  function renderWeighbridge() {
    const scopedTicketIds = new Set(scopedTickets.map((item) => item.id));
    const voidLogs = data.ticketVoidLogs.filter((item) => scopedTickets.some((ticket) => ticket.id === item.ticketId));
    const pendingVoidLogs = voidLogs.filter((item) => item.status === "pending");
    const printLogs = data.ticketPrintLogs.filter((item) => scopedTicketIds.has(item.ticketId));
    const deviceEvents = data.scaleDeviceEvents.filter((item) => !selectedSiteId || !item.ticketId || scopedTicketIds.has(item.ticketId));
    const completedTransfers = list(data.procurement?.transfers).filter((item) => item.status === "completed" && (!selectedSiteId || item.fromSiteId === selectedSiteId || item.toSiteId === selectedSiteId));
    const grossWeight = fieldNumber(ticketForm.grossWeight);
    const tareWeight = fieldNumber(ticketForm.tareWeight);
    const netWeight = grossWeight - tareWeight;
    const canCreateTicket = netWeight > 0 && (
      ticketForm.mode === "inventory_transfer" ? fieldNumber(ticketForm.transferId) > 0
        : ticketForm.mode === "waste_out" ? fieldNumber(ticketForm.siteId) > 0 && fieldNumber(ticketForm.materialId) > 0
          : fieldNumber(ticketForm.dispatchId) > 0
    );
    const weightTypeLabel = (value: string) => {
      if (value === "gross") {
        return "-";
      }
      if (value === "tare") {
        return "-";
      }
      return value || "-";
    };
    const weightsForTicket = (ticketId: number) => scopedWeightRecords.filter((item) => item.ticketId === ticketId);
    const voidLogsForTicket = (ticketId: number) => voidLogs.filter((item) => item.ticketId === ticketId);
    const pendingVoidLogForTicket = (ticketId: number) => pendingVoidLogs.find((item) => item.ticketId === ticketId);

    async function handleInlineVoidRequest(event: FormEvent<HTMLFormElement>, ticket: ScaleTicket) {
      event.preventDefault();
      const reason = String(new FormData(event.currentTarget).get("reason") || "").trim();
      if (!reason) {
        return;
      }
      await runBusinessAction(`ticket-void-request-${ticket.id}`, "过磅记录作废已提交复核", () => api.requestTicketVoid(ticket.id, reason));
    }

    const ticketFormView = (
      <DialogForm onSubmit={handleCreateWeighbridgeTicket}>
        <Field label="票据类型">
          <SelectInput value={ticketForm.mode} onChange={(event) => resetTicketForm(event.target.value)}>
            <option value="product_out">成品出厂</option>
            <option value="inventory_transfer">库存调拨</option>
            <option value="product_return">工地退料</option>
            <option value="waste_out">废料出库</option>
          </SelectInput>
        </Field>
        {ticketForm.mode === "inventory_transfer" ? (
          <Field label="调拨单">
            <SelectInput value={ticketForm.transferId} onChange={(event) => setTicketForm({ ...ticketForm, transferId: event.target.value })}>
              {completedTransfers.map((item) => <option key={item.id} value={item.id}>{item.transferNo} / {nameOf(bootstrap?.materials, item.materialId)} / {qty(item.quantity)} {item.unit}</option>)}
            </SelectInput>
          </Field>
        ) : null}
        {ticketForm.mode === "product_out" || ticketForm.mode === "product_return" ? (
          <Field label="派车单">
            <SelectInput value={ticketForm.dispatchId} onChange={(event) => setTicketForm({ ...ticketForm, dispatchId: event.target.value })}>
              {scopedDispatchOrders.map((item) => <option key={item.id} value={item.id}>{item.dispatchNo} / {nameOf(bootstrap?.vehicles, item.vehicleId)} / {item.productName || productLabel(bootstrap, item.productId)}</option>)}
            </SelectInput>
          </Field>
        ) : null}
        {ticketForm.mode === "product_return" ? (
          <Field label="原出厂票">
            <SelectInput value={ticketForm.relatedTicketId} onChange={(event) => setTicketForm({ ...ticketForm, relatedTicketId: event.target.value })}>
              <option value="">不关联</option>
              {scopedTickets.filter((item) => item.ticketType === "product_out").map((item) => <option key={item.id} value={item.id}>{item.ticketNo} / {item.plateNo}</option>)}
            </SelectInput>
          </Field>
        ) : null}
        {ticketForm.mode === "waste_out" ? (
          <>
            <Field label="站点">
              <SelectInput value={ticketForm.siteId} onChange={(event) => setTicketForm({ ...ticketForm, siteId: event.target.value })}>
                {siteOptions().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
            <Field label="物料">
              <SelectInput value={ticketForm.materialId} onChange={(event) => setTicketForm({ ...ticketForm, materialId: event.target.value })}>
                {(bootstrap?.materials || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </SelectInput>
            </Field>
          </>
        ) : null}
        <Field label="车牌"><TextInput value={ticketForm.plateNo} onChange={(event) => setTicketForm({ ...ticketForm, plateNo: event.target.value })} /></Field>
        <Field label="毛重"><TextInput type="number" min="0" step="0.01" value={ticketForm.grossWeight} onChange={(event) => setTicketForm({ ...ticketForm, grossWeight: event.target.value })} /></Field>
        <Field label="皮重"><TextInput type="number" min="0" step="0.01" value={ticketForm.tareWeight} onChange={(event) => setTicketForm({ ...ticketForm, tareWeight: event.target.value })} /></Field>
        <Field label="单位"><TextInput value={ticketForm.unit} onChange={(event) => setTicketForm({ ...ticketForm, unit: event.target.value })} /></Field>
        <Field label="备注" spanAll><TextAreaInput value={ticketForm.remark} onChange={(event) => setTicketForm({ ...ticketForm, remark: event.target.value })} /></Field>
        <MetricList compact className="span-all">
          <div><span>净重</span><b>{qty(Math.max(0, netWeight))} {ticketForm.unit || "t"}</b></div>
          <div><span>类型</span><b>{ticketForm.mode}</b></div>
        </MetricList>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Scale size={14} />} disabled={actionBusy !== "" || !canCreateTicket}>创建过磅记录</UiButton>
        </FormActions>
      </DialogForm>
    );

    return (
      <SectionGrid>
        <Panel as="div" className="span-12">
          <DataTable
            data={scopedTickets}
            rowKey={(item) => item.id}
            pageSize={12}
            emptyText="暂无过磅记录"
            headerLeftAction={(
              <ActionDialog id="weighbridge-ticket-create" title="创建过磅记录" buttonLabel="创建磅单" triggerIcon={<Plus size={13} />} onOpen={() => resetTicketForm("product_out")}>
                {ticketFormView}
              </ActionDialog>
            )}
            headerAction={<span className="muted">{scopedTickets.length} 条记录 / {scopedWeightRecords.length} 条称重流水</span>}
            columns={[
              { key: "ticketNo", title: "磅单", render: (item) => <b>{item.ticketNo}</b> },
              { key: "plateNo", title: "车牌", render: (item) => item.plateNo },
              { key: "ticketType", title: "类型", render: (item) => item.ticketType || "-" },
              { key: "weight", title: "重量", render: (item) => `${qty(item.grossWeight)} / ${qty(item.tareWeight)} / ${qty(item.netWeight)} ${item.unit}` },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) },
              { key: "signStatus", title: "签收", render: (item) => <StatusChip value={item.signStatus} /> },
              { key: "settlementStatus", title: "结算", render: (item) => <StatusChip value={item.settlementStatus} /> },
	              {
	                key: "status",
	                title: "状态",
	                render: (item) => {
	                  const pendingVoidLog = pendingVoidLogForTicket(item.id);
	                  if (pendingVoidLog) {
	                    return workflowStatusFor(["ticket_void", "ticketVoid"], pendingVoidLog.id, item.ticketNo, <StatusChip value={pendingVoidLog.status} />);
	                  }
	                  return workflowStatusFor(["scale_ticket", "ticket", "weighbridge_ticket"], item.id, item.ticketNo, approvalStatus(approvalFor(["scale_ticket", "ticket", "weighbridge_ticket"], item.id, item.ticketNo), <StatusChip value={item.status} />));
	                }
	              },
              { key: "actions", title: "操作", render: (item) => {
	                const ticketWeights = weightsForTicket(item.id);
	                const ticketVoidLogs = voidLogsForTicket(item.id);
		                const pendingVoidLog = pendingVoidLogForTicket(item.id);
		                const task = approvalFor(["scale_ticket", "ticket", "weighbridge_ticket"], item.id, item.ticketNo);
		                const workflow = workflowItemsFor(["scale_ticket", "ticket", "weighbridge_ticket"], item.id, item.ticketNo);
		                const voidWorkflow = pendingVoidLog ? workflowItemsFor(["ticket_void", "ticketVoid"], pendingVoidLog.id, item.ticketNo) : null;
		                return (
	                  <ActionDialog id={`weighbridge-action-${item.id}`} title="过磅操作">
                    <div className="weighbridge-hidden-actions">
                      <ActionGroup className="compact-actions">
                        <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`ticket-reprint-${item.id}`, "过磅记录已补打", () => api.reprintTicket(item.id))}>补打</UiButton>
                      </ActionGroup>
	                      {workflowTimelineBlock(["scale_ticket", "ticket", "weighbridge_ticket"], item.id, item.ticketNo, "当前磅单暂无工作流实例")}
	                      {!workflow.instances.length ? approvalActionBlock(task) : null}
                      {pendingVoidLog ? (
	                        <div className="weighbridge-action-block">
	                          <b>作废复核</b>
	                          <span>{pendingVoidLog.reason} / {shortDateTime(pendingVoidLog.createdAt)} / {workflowStatusFor(["ticket_void", "ticketVoid"], pendingVoidLog.id, item.ticketNo, <StatusChip value={pendingVoidLog.status} />)}</span>
	                          {workflowTimelineBlock(["ticket_void", "ticketVoid"], pendingVoidLog.id, item.ticketNo, "当前作废申请暂无工作流实例")}
	                          {!voidWorkflow?.instances.length ? (
	                            <ActionGroup className="compact-actions">
	                              <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => handleTicketVoidReview(item.id, true)}>通过</UiButton>
	                              <UiButton disabled={actionBusy !== ""} onClick={() => handleTicketVoidReview(item.id, false)}>驳回</UiButton>
	                            </ActionGroup>
	                          ) : null}
	                        </div>
	                      ) : null}
                      {item.status !== "void" && !pendingVoidLog ? (
                        <InlineForm onSubmit={(event) => handleInlineVoidRequest(event, item)}>
                          <Field label="作废原因">
                            <TextInput name="reason" defaultValue="" />
                          </Field>
                          <UiButton type="submit" icon={<Scale size={13} />} disabled={actionBusy !== ""}>提交作废</UiButton>
                        </InlineForm>
                      ) : null}
                      <div className="weighbridge-action-block">
                        <b>称重流水</b>
                        {ticketWeights.map((weight) => (
                          <span key={weight.id}>{weightTypeLabel(weight.weightType)} {qty(weight.weight)} kg / {shortDateTime(weight.createdAt)}</span>
                        ))}
                        {!ticketWeights.length ? <span>暂无流水</span> : null}
                      </div>
                      <div className="weighbridge-action-block">
                        <b>作废记录</b>
                        {ticketVoidLogs.map((log) => (
                          <span key={log.id}>{log.reason || "-"} / {log.approvedBy || "未复核"} / {log.status}</span>
                        ))}
                        {!ticketVoidLogs.length ? <span>暂无记录</span> : null}
                      </div>
                    </div>
                  </ActionDialog>
                );
              } }
            ]}
          />
        </Panel>
        <Panel as="div" className="span-6">
          <DataTable
            title="补打日志"
            data={printLogs}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "ticket", title: "磅单", render: (item) => ticketById.get(item.ticketId)?.ticketNo || item.ticketId },
              { key: "plate", title: "车牌", render: (item) => ticketById.get(item.ticketId)?.plateNo || "-" },
              { key: "printedBy", title: "补打人", render: (item) => item.printedBy || "-" },
              { key: "printedAt", title: "时间", render: (item) => shortDateTime(item.printedAt) }
            ]}
            emptyText="暂无补打日志"
          />
        </Panel>
        <Panel as="div" className="span-6">
          <DataTable
            title="设备事件"
            data={deviceEvents}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "event", title: "事件", render: (item) => <><b>{item.eventNo}</b><span className="block-text muted">{item.deviceCode}</span></> },
              { key: "plate", title: "车牌", render: (item) => item.plateNo || item.recognizedPlateNo || "-" },
              { key: "weight", title: "重量", render: (item) => `${weightTypeLabel(item.weightType)} ${qty(item.weight)} kg` },
              { key: "stable", title: "稳定", render: (item) => <StatusChip value={item.stable ? "active" : "warning"} /> },
              { key: "cheat", title: "异常", render: (item) => item.cheatFlag ? item.cheatReason || "异常" : "-" },
              { key: "receivedAt", title: "时间", render: (item) => shortDateTime(item.receivedAt) }
            ]}
            emptyText="暂无设备事件"
          />
        </Panel>
      </SectionGrid>
    );
  }

  function renderDelivery() {
    const notes = [...scopedDeliveryNotes].sort((a, b) => b.id - a.id);
    const noteDispatchIds = new Set(data.deliveryNotes.map((item) => item.dispatchId));
    const pendingDispatches = scopedDispatchOrders.filter((item) => !noteDispatchIds.has(item.id));
    const selectedDispatchId = fieldNumber(deliveryForm.dispatchId);
    const ticketOptions = scopedTickets.filter((item) => item.dispatchId === selectedDispatchId);
    const printNote = (printDeliveryNoteId ? data.deliveryNotes.find((item) => item.id === printDeliveryNoteId) : null) || notes[0] || null;
    const deliveryChannelOptions = dictionaryOptions("delivery_channel");
    const portalCustomerOptions = list(bootstrap?.customers).filter((item) => !selectedSiteId || scopedOrders.some((order) => order.customerId === item.id));
    const portalComplaints = (data.portalComplaints.length ? data.portalComplaints : list(data.portal?.complaints)).sort((a, b) => b.id - a.id);
    const portalProjectOptionsFor = (customerId = fieldNumber(portalComplaintForm.customerId)) => list(bootstrap?.projects).filter((item) => {
      if (customerId && item.customerId !== customerId) return false;
      return !selectedSiteId || scopedOrders.some((order) => order.projectId === item.id);
    });
    const portalProjectOptions = portalProjectOptionsFor();

    function dispatchLabel(item: DispatchOrder | undefined) {
      if (!item) return "-";
      const vehicle = nameOf(bootstrap?.vehicles, item.vehicleId);
      return `${item.dispatchNo} / ${vehicle} / ${item.productName || productLabel(bootstrap, item.productId)}`;
    }

    function noteOrderLabel(note: DeliveryNote) {
      const order = orderById.get(note.orderId);
      if (!order) return "-";
      return `${nameOf(bootstrap?.customers, order.customerId)} / ${nameOf(bootstrap?.projects, order.projectId)}`;
    }

    function deliveryStatusChip(status: string) {
      const labels = { issued: "已开单", pending: "待签收", signed: "已签收", void: "已作废", cancelled: "已取消" } as Record<string, string>;
      const tone = status === "signed" ? "success" : status === "pending" ? "warning" : status === "void" || status === "cancelled" ? "danger" : "primary";
      return <span className={`status-chip ${tone}`}>{labels[status] || status || "-"}</span>;
    }

    function deliveryPrintSheet(note: DeliveryNote | null) {
      if (!note) return null;
      const dispatch = dispatchById.get(note.dispatchId);
      const order = orderById.get(note.orderId);
      const ticket = ticketById.get(note.ticketId);
      const sign = signByDispatchId.get(note.dispatchId);
      return (
        <LayoutRegion as="section" className="delivery-print-sheet" aria-label="送货单">
          <LayoutRegion as="header">
            <h2>送货单</h2>
            <span>{note.noteNo}</span>
          </LayoutRegion>
          <div className="delivery-print-grid">
            <span>送货单号：{note.noteNo}</span>
            <span>状态：{deliveryStatusChip(note.status)}</span>
            <span>订单：{order?.orderNo || "-"}</span>
            <span>客户/项目：{noteOrderLabel(note)}</span>
            <span>调度单：{dispatch?.dispatchNo || "-"}</span>
            <span>车牌：{ticket?.plateNo || (dispatch ? nameOf(bootstrap?.vehicles, dispatch.vehicleId) : "-")}</span>
            <span>产品：{dispatch?.productName || (order ? productLabel(bootstrap, order.productId) : "-")}</span>
            <span>数量：{qty(ticket?.netWeight || dispatch?.loadedQty || sign?.signedQty)} {ticket?.unit || order?.unit || "t"}</span>
            <span>签收人：{sign?.signer || "-"}</span>
            <span>开单时间：{shortDateTime(note.createdAt)}</span>
            <span className="span-all">收货地址：{order?.receiveAddress || "-"}</span>
            <span className="span-all">二维码：{note.qrCode || "-"}</span>
            <span>过磅票：{ticket?.ticketNo || "-"}</span>
            <span>签收时间：{sign?.signedAt ? shortDateTime(sign.signedAt) : "-"}</span>
          </div>
        </LayoutRegion>
      );
    }

    type DeliveryLedgerRow =
      | { id: string; kind: "note"; sort: number; note: DeliveryNote }
      | { id: string; kind: "pending"; sort: number; dispatch: DispatchOrder }
      | { id: string; kind: "sign"; sort: number; sign: DeliverySign }
      | { id: string; kind: "link"; sort: number; link: DeliverySignLink };
    const deliveryRows: DeliveryLedgerRow[] = [
      ...notes.map((note) => ({ id: `note-${note.id}`, kind: "note" as const, sort: 400000 + note.id, note })),
      ...pendingDispatches.map((dispatch) => ({ id: `pending-${dispatch.id}`, kind: "pending" as const, sort: 300000 + dispatch.id, dispatch })),
      ...[...scopedSigns].sort((a, b) => b.id - a.id).map((sign) => ({ id: `sign-${sign.id}`, kind: "sign" as const, sort: 200000 + sign.id, sign })),
      ...[...scopedSignLinks].sort((a, b) => b.id - a.id).map((link) => ({ id: `link-${link.id}`, kind: "link" as const, sort: 100000 + link.id, link }))
    ].sort((a, b) => b.sort - a.sort);

    if (section === "delivery-signs") {
      const signs = [...scopedSigns].sort((a, b) => b.id - a.id);
      const links = [...scopedSignLinks].sort((a, b) => b.id - a.id);
      const attachments = [...scopedSignAttachments].sort((a, b) => b.id - a.id);
      const pendingLinks = links.filter((item) => item.status !== "used" && item.status !== "expired");
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board delivery-notes-page">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="签收归档范围">
                <span className="inventory-scope-chip strong">{selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点"}</span>
                <span className="inventory-scope-chip">{signs.length} 条签收</span>
                <span className={`inventory-scope-chip ${pendingLinks.length ? "warning" : ""}`}>{pendingLinks.length} 条待用链接</span>
                <span className="inventory-scope-chip">{attachments.length} 个附件</span>
              </div>
              <ActionGroup>
                <ButtonLink icon={<ReceiptText size={15} />} href="/fulfillment/delivery">送货单</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>签收记录</span><b>{signs.length}</b><small>当前站点范围</small></div>
              <div><span>待复核</span><b>{signs.filter((item) => item.reviewStatus !== "reviewed").length}</b><small>签收复核状态</small></div>
              <div><span>签收链接</span><b>{links.length}</b><small>待用 {pendingLinks.length}</small></div>
              <div><span>附件</span><b>{attachments.length}</b><small>照片/签名归档</small></div>
              <div><span>签收量</span><b>{qty(signs.reduce((sum, item) => sum + item.signedQty, 0))}</b><small>已确认数量</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DeliverySign>
              title="签收记录"
              data={signs}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "sign", title: "签收单", render: (item) => <><b>{item.signNo}</b><span className="block-text muted">{shortDateTime(item.signedAt)}</span></> },
                { key: "dispatch", title: "派车", render: (item) => dispatchLabel(dispatchById.get(item.dispatchId)) },
                { key: "customer", title: "客户项目", render: (item) => {
                  const order = orderById.get(item.orderId);
                  return order ? `${nameOf(bootstrap?.customers, order.customerId)} / ${nameOf(bootstrap?.projects, order.projectId)}` : "-";
                } },
                { key: "product", title: "产品", render: (item) => item.productName || productLabel(bootstrap, item.productId) },
                { key: "qty", title: "签收量", render: (item) => qty(item.signedQty) },
                { key: "signer", title: "签收人", render: (item) => `${item.signer || "-"} / ${item.phone || "-"}` },
                { key: "location", title: "定位", render: (item) => item.latitude || item.longitude ? `${item.latitude}, ${item.longitude}` : "-" },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["delivery_sign"], item.id, item.signNo, <StatusChip value={item.reviewStatus || "signed"} />) },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => {
                    const signAttachments = attachments.filter((attachment) => attachment.signId === item.id);
                    return (
                      <ActionDialog
                        id={`delivery-sign-archive-${item.id}`}
                        title="签收归档"
                        buttonLabel="归档"
                        onOpen={() => setSignAttachmentForm({
                          signId: String(item.id),
                          fileName: "",
                          fileType: "image/jpeg",
                          url: "",
                          checksum: "",
                          uploadedBy: bootstrap?.user.displayName || bootstrap?.user.username || ""
                        })}
                      >
                        <div className="delivery-action-block">
                          <b>{item.signNo} / {item.signer || "-"}</b>
                          <span>{item.productName} / {qty(item.signedQty)} / {item.signedAt || "-"}</span>
                          <span>{item.remark || "-"}</span>
                          {workflowTimelineBlock(["delivery_sign"], item.id, item.signNo, "当前签收暂无工作流实例")}
                          <DialogForm className="compact-dialog-form" onSubmit={(event) => handleAddSignAttachment(event, item)}>
                            <Field label="附件文件"><input type="file" onChange={handleSignAttachmentFile} /></Field>
                            <Field label="附件名"><TextInput value={signAttachmentForm.fileName} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, fileName: event.target.value })} required /></Field>
                            <Field label="类型"><TextInput value={signAttachmentForm.fileType} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, fileType: event.target.value })} /></Field>
                            <Field label="校验值"><TextInput value={signAttachmentForm.checksum} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, checksum: event.target.value })} /></Field>
                            <FormActions spanAll>
                              <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== "" || !signAttachmentForm.fileName.trim() || !signAttachmentForm.url.trim()}>归档附件</UiButton>
                            </FormActions>
                          </DialogForm>
                          <div className="finance-action-block">
                            <b>附件</b>
                            {signAttachments.map((attachment) => <span key={attachment.id}>{attachment.fileName} / {attachment.fileType || "-"} / {attachment.uploadedBy || "-"}</span>)}
                            {!signAttachments.length ? <span>暂无附件</span> : null}
                          </div>
                        </div>
                      </ActionDialog>
                    );
                  }
                }
              ]}
              emptyText="暂无签收记录"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DeliverySignLink>
              title="签收链接"
              data={links}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "link", title: "链接", render: (item) => <><b>{item.linkNo}</b><span className="block-text muted">{item.url || "-"}</span></> },
                { key: "dispatch", title: "派车", render: (item) => dispatchLabel(dispatchById.get(item.dispatchId)) },
                { key: "product", title: "产品", render: (item) => item.productName || productLabel(bootstrap, item.productId) },
                { key: "target", title: "目标", render: (item) => `${item.channel} / ${item.phone || "-"}` },
                { key: "sent", title: "发送/使用", render: (item) => `${shortDateTime(item.sentAt)} / ${shortDateTime(item.usedAt)}` },
                { key: "expires", title: "过期", render: (item) => shortDateTime(item.expiresAt) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
              ]}
              emptyText="暂无签收链接"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DeliverySignAttachment>
              title="签收附件"
              data={attachments}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "file", title: "附件", render: (item) => <><b>{item.fileName}</b><span className="block-text muted">{item.fileType || "-"}</span></> },
                { key: "sign", title: "签收单", render: (item) => signs.find((sign) => sign.id === item.signId)?.signNo || `签收 #${item.signId}` },
                { key: "dispatch", title: "派车", render: (item) => dispatchLabel(dispatchById.get(item.dispatchId)) },
                { key: "checksum", title: "校验值", render: (item) => item.checksum || "-" },
                { key: "uploaded", title: "上传", render: (item) => `${item.uploadedBy || "-"} / ${shortDateTime(item.uploadedAt)}` },
                { key: "url", title: "地址", render: (item) => item.url || "-" }
              ]}
              emptyText="暂无签收附件"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <Panel className="delivery-notes-page">
        {deliveryPrintSheet(printNote)}
        <DataTable
          data={deliveryRows}
          rowKey={(item) => item.id}
          pageSize={12}
          onRefresh={refreshData}
          emptyText={loading ? "加载中..." : "暂无送货单数据"}
          rowContextMenu={buildDataTableRowContextMenu<DeliveryLedgerRow>({
            actions: [
              {
                key: "preview-note",
                label: "设为打印预览",
                disabled: (item) => item.kind !== "note",
                onSelect: (item) => {
                  if (item.kind === "note") setPrintDeliveryNoteId(item.note.id);
                }
              },
              {
                key: "focus-dispatch",
                label: "只看该派车单",
                onSelect: (item, helpers) => {
                  if (item.kind === "note") helpers.searchText(dispatchById.get(item.note.dispatchId)?.dispatchNo || String(item.note.dispatchId));
                  if (item.kind === "pending") helpers.searchText(item.dispatch.dispatchNo);
                  if (item.kind === "sign") helpers.searchText(dispatchById.get(item.sign.dispatchId)?.dispatchNo || item.sign.signNo);
                  if (item.kind === "link") helpers.searchText(dispatchById.get(item.link.dispatchId)?.dispatchNo || item.link.linkNo);
                }
              }
            ],
            copyFields: [
              { key: "no", label: "当前单号", value: (item) => {
                if (item.kind === "note") return item.note.noteNo;
                if (item.kind === "pending") return item.dispatch.dispatchNo;
                if (item.kind === "sign") return item.sign.signNo;
                return item.link.linkNo;
              } },
              { key: "dispatch", label: "派车信息", value: (item) => {
                if (item.kind === "note") return dispatchLabel(dispatchById.get(item.note.dispatchId));
                if (item.kind === "pending") return dispatchLabel(item.dispatch);
                if (item.kind === "sign") return dispatchLabel(dispatchById.get(item.sign.dispatchId));
                return dispatchLabel(dispatchById.get(item.link.dispatchId));
              } },
              { key: "customer", label: "客户项目", value: (item) => {
                if (item.kind === "note") return noteOrderLabel(item.note);
                const orderId = item.kind === "pending" ? item.dispatch.orderId : item.kind === "sign" ? item.sign.orderId : item.link.orderId;
                const order = orderById.get(orderId);
                return order ? `${nameOf(bootstrap?.customers, order.customerId)} / ${nameOf(bootstrap?.projects, order.projectId)}` : "";
              } },
              { key: "target", label: "签收目标", value: (item) => {
                if (item.kind === "sign") return `${item.sign.signer || "-"} / ${item.sign.phone || "-"}`;
                if (item.kind === "link") return `${item.link.channel} / ${item.link.phone || "-"}`;
                return "";
              } }
            ]
          })}
          headerLeftAction={
            <ActionDialog
              id="delivery-note-create"
              title="新增送货单"
              buttonLabel="生成送货单"
              onOpen={() => setDeliveryForm((form) => ({ ...form, dispatchId: String(scopedDispatchOrders[0]?.id || ""), ticketId: "" }))}
            >
              <DialogForm onSubmit={handleCreateDeliveryNote}>
                <Field label="派车单">
                  <SelectInput value={deliveryForm.dispatchId} onChange={(event) => setDeliveryForm({ ...deliveryForm, dispatchId: event.target.value })}>
                    {scopedDispatchOrders.map((item) => <option key={item.id} value={item.id}>{dispatchLabel(item)}</option>)}
                  </SelectInput>
                </Field>
                <Field label="磅单">
                  <SelectInput value={deliveryForm.ticketId} onChange={(event) => setDeliveryForm({ ...deliveryForm, ticketId: event.target.value })}>
                    <option value="">不关联磅单</option>
                    {ticketOptions.map((item) => <option key={item.id} value={item.id}>{item.ticketNo} / {qty(item.netWeight)} {item.unit}</option>)}
                  </SelectInput>
                </Field>
                <FormActions>
                  <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(deliveryForm.dispatchId)}>保存送货单</UiButton>
                </FormActions>
              </DialogForm>
            </ActionDialog>
          }
          columns={[
            { key: "type", title: "类型", render: (item) => ({ note: "送货单", pending: "待开单", sign: "签收", link: "链接" }[item.kind]) },
            {
              key: "no",
              title: "单号",
              render: (item) => {
                if (item.kind === "note") return <b>{item.note.noteNo}</b>;
                if (item.kind === "pending") return <b>{item.dispatch.dispatchNo}</b>;
                if (item.kind === "sign") return <b>{item.sign.signNo}</b>;
                return <b>{item.link.linkNo}</b>;
              }
            },
            {
              key: "dispatch",
              title: "派车",
              render: (item) => {
                if (item.kind === "note") return dispatchLabel(dispatchById.get(item.note.dispatchId));
                if (item.kind === "pending") return dispatchLabel(item.dispatch);
                if (item.kind === "sign") return dispatchLabel(dispatchById.get(item.sign.dispatchId));
                return dispatchLabel(dispatchById.get(item.link.dispatchId));
              }
            },
            {
              key: "customer",
              title: "客户项目",
              render: (item) => {
                if (item.kind === "note") return noteOrderLabel(item.note);
                const orderId = item.kind === "pending" ? item.dispatch.orderId : item.kind === "sign" ? item.sign.orderId : item.link.orderId;
                const order = orderById.get(orderId);
                return order ? `${nameOf(bootstrap?.customers, order.customerId)} / ${nameOf(bootstrap?.projects, order.projectId)}` : "-";
              }
            },
            {
              key: "qty",
              title: "数量",
              render: (item) => {
                if (item.kind === "note") {
                  const ticket = ticketById.get(item.note.ticketId);
                  return ticket ? `${ticket.ticketNo} / ${qty(ticket.netWeight)} ${ticket.unit}` : "未关联磅单";
                }
                if (item.kind === "pending") return `${qty(item.dispatch.planQuantity)} / ${item.dispatch.productName || productLabel(bootstrap, item.dispatch.productId)}`;
                if (item.kind === "sign") return `${qty(item.sign.signedQty)} / ${item.sign.signer || "-"}`;
                return `${item.link.channel} / ${item.link.phone || "-"}`;
              }
            },
            {
              key: "status",
              title: "状态",
              width: "110px",
              render: (item) => {
	                if (item.kind === "note") return workflowStatusFor(["delivery_note", "delivery_notes", "delivery"], item.note.id, item.note.noteNo, approvalStatus(approvalFor(["delivery_note", "delivery_notes", "delivery"], item.note.id, item.note.noteNo), deliveryStatusChip(item.note.status)));
                if (item.kind === "pending") return <StatusChip value={item.dispatch.status} />;
                if (item.kind === "sign") return workflowStatusFor(["delivery_sign"], item.sign.id, item.sign.signNo, <StatusChip value={item.sign.reviewStatus || "signed"} />);
                return <StatusChip value={item.link.status} />;
              }
            },
            {
              key: "actions",
              title: "操作",
              width: "120px",
              render: (item) => {
	                if (item.kind === "note") {
	                  const dispatch = dispatchById.get(item.note.dispatchId);
	                  const order = orderById.get(item.note.orderId);
	                  const sign = signByDispatchId.get(item.note.dispatchId);
	                  const workflow = workflowItemsFor(["delivery_note", "delivery_notes", "delivery"], item.note.id, item.note.noteNo);
	                  return (
                    <ActionDialog
                      id={`delivery-note-action-${item.note.id}`}
                      title="送货单操作"
                      onOpen={() => setDeliveryForm((form) => ({ ...form, dispatchId: String(item.note.dispatchId), ticketId: String(item.note.ticketId || ""), phone: order?.phone || form.phone }))}
                    >
                      <div className="delivery-action-stack">
                        <div className="delivery-action-block">
                          <b>{dispatchLabel(dispatch)}</b>
                          <span>{noteOrderLabel(item.note)} / {order?.receiveAddress || "-"}</span>
                          <span>{deliveryStatusChip(item.note.status)}</span>
                          {sign ? <span>已签收：{sign.signer || "-"}</span> : <span>未签收</span>}
                        </div>
	                        {workflowTimelineBlock(["delivery_note", "delivery_notes", "delivery"], item.note.id, item.note.noteNo, "当前送货单暂无工作流实例")}
	                        {!workflow.instances.length ? approvalActionBlock(approvalFor(["delivery_note", "delivery_notes", "delivery"], item.note.id, item.note.noteNo)) : null}
                        <DialogForm className="compact-dialog-form" onSubmit={(event) => handleCreateDeliveryNoteLink(event, item.note)}>
                          <Field label="渠道">
                            <SelectInput value={deliveryForm.channel} onChange={(event) => setDeliveryForm({ ...deliveryForm, channel: event.target.value })}>
                              {deliveryChannelOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                            </SelectInput>
                          </Field>
                          <Field label="手机号"><TextInput value={deliveryForm.phone} onChange={(event) => setDeliveryForm({ ...deliveryForm, phone: event.target.value })} /></Field>
                          <HeroDateField className="span-all" label="过期时间" mode="date-time" value={deliveryForm.expiresAt} onChange={(expiresAt) => setDeliveryForm({ ...deliveryForm, expiresAt })} />
                          <FormActions>
                            <UiButton type="submit" icon={<Link2 size={14} />} disabled={actionBusy !== "" || item.note.status === "void"}>生成签收链接</UiButton>
                          </FormActions>
                        </DialogForm>
                        <ActionGroup className="compact-actions">
                          <UiButton icon={<Printer size={14} />} disabled={actionBusy !== ""} onClick={() => handleReprintDeliveryNote(item.note)}>补打</UiButton>
                          {item.note.status === "void" || item.note.status === "cancelled" ? (
                            <UiButton icon={<RefreshCw size={14} />} disabled={actionBusy !== ""} onClick={() => handleDeliveryNoteStatus(item.note, "reopen")}>恢复</UiButton>
                          ) : (
                            <UiButton icon={<X size={14} />} disabled={actionBusy !== "" || item.note.status === "signed"} onClick={() => handleDeliveryNoteStatus(item.note, "void")}>作废</UiButton>
                          )}
                        </ActionGroup>
                      </div>
                    </ActionDialog>
                  );
                }
                if (item.kind === "pending") {
                  const ticketOptionsForDispatch = scopedTickets.filter((ticket) => ticket.dispatchId === item.dispatch.id);
                  return (
                    <ActionDialog
                      id={`delivery-pending-${item.dispatch.id}`}
                      title="生成送货单"
                      onOpen={() => setDeliveryForm((form) => ({ ...form, dispatchId: String(item.dispatch.id), ticketId: String(ticketOptionsForDispatch[0]?.id || "") }))}
                    >
                      <div className="delivery-action-block">
                        <b>{dispatchLabel(item.dispatch)}</b>
                        <span>{nameOf(bootstrap?.projects, item.dispatch.projectId)} / 计划 {qty(item.dispatch.planQuantity)}</span>
                      </div>
                      <DialogForm className="compact-dialog-form" onSubmit={handleCreateDeliveryNote}>
                        <Field label="磅单">
                          <SelectInput value={deliveryForm.ticketId} onChange={(event) => setDeliveryForm({ ...deliveryForm, ticketId: event.target.value })}>
                            <option value="">不关联磅单</option>
                            {ticketOptionsForDispatch.map((ticket) => <option key={ticket.id} value={ticket.id}>{ticket.ticketNo} / {qty(ticket.netWeight)} {ticket.unit}</option>)}
                          </SelectInput>
                        </Field>
                        <FormActions>
                          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== ""}>保存送货单</UiButton>
                        </FormActions>
                      </DialogForm>
                    </ActionDialog>
                  );
                }
                if (item.kind === "sign") {
                  const attachments = scopedSignAttachments.filter((attachment) => attachment.signId === item.sign.id);
                  return (
                    <ActionDialog
                      id={`delivery-sign-${item.sign.id}`}
                      title="签收详情"
                      onOpen={() => setSignAttachmentForm({
                        signId: String(item.sign.id),
                        fileName: "",
                        fileType: "image/jpeg",
                        url: "",
                        checksum: "",
                        uploadedBy: bootstrap?.user.displayName || bootstrap?.user.username || ""
                      })}
                    >
                      <div className="delivery-action-block">
                        <b>{item.sign.signer || "未填签收人"} / {item.sign.phone || "-"}</b>
                        <span>{item.sign.productName} / {qty(item.sign.signedQty)}</span>
	                        <span>{item.sign.latitude}, {item.sign.longitude} / {item.sign.signedAt || "-"}</span>
	                        <span>{item.sign.signNo}</span>
	                        <span>复核：{item.sign.reviewedBy || "-"} / {item.sign.reviewedAt || "-"}</span>
	                        <span>{item.sign.remark || "-"}</span>
	                        {workflowTimelineBlock(["delivery_sign"], item.sign.id, item.sign.signNo, "当前签收暂无工作流实例")}
                        <DialogForm className="compact-dialog-form" onSubmit={(event) => handleAddSignAttachment(event, item.sign)}>
                          <Field label="附件文件"><input type="file" onChange={handleSignAttachmentFile} /></Field>
                          <Field label="附件名"><TextInput value={signAttachmentForm.fileName} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, fileName: event.target.value })} required /></Field>
                          <Field label="类型"><TextInput value={signAttachmentForm.fileType} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, fileType: event.target.value })} /></Field>
                          <Field label="校验值"><TextInput value={signAttachmentForm.checksum} onChange={(event) => setSignAttachmentForm({ ...signAttachmentForm, checksum: event.target.value })} /></Field>
                          <FormActions spanAll>
                            <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== "" || !signAttachmentForm.fileName.trim() || !signAttachmentForm.url.trim()}>归档附件</UiButton>
                          </FormActions>
                        </DialogForm>
                        <div className="finance-action-block">
                          <b>附件</b>
                          {attachments.map((attachment) => <span key={attachment.id}>{attachment.fileName} / {attachment.fileType || "-"} / {attachment.uploadedBy || "-"}</span>)}
                          {!attachments.length ? <span>暂无附件</span> : null}
                        </div>
	                      </div>
                    </ActionDialog>
                  );
                }
                return (
                  <ActionDialog id={`delivery-link-${item.link.id}`} title="签收链接">
                    <div className="delivery-action-block">
                      <b>{item.link.channel} / {item.link.phone || "-"}</b>
                      <span>{item.link.url || "-"}</span>
                      <span>{item.link.status}</span>
                      <span>{item.link.expiresAt || "-"}</span>
                    </div>
                  </ActionDialog>
                );
              }
            }
          ]}
        />
        <DataTable
          title="客户门户投诉"
          data={portalComplaints}
          rowKey={(item) => item.id}
          pageSize={6}
          headerLeftAction={(
            <ActionDialog
              id="portal-complaint-create"
              title="提交门户投诉"
              buttonLabel="新增投诉"
              triggerIcon={<Plus size={13} />}
              onOpen={() => {
                const customerId = portalCustomerOptions[0]?.id || 0;
                const project = portalProjectOptionsFor(customerId)[0];
                setPortalComplaintForm({
                  customerId: String(customerId || ""),
                  projectId: String(project?.id || ""),
                  title: "",
                  content: "",
                  level: "medium"
                });
              }}
            >
              <DialogForm onSubmit={handleCreatePortalComplaint}>
                <Field label="客户">
                  <SelectInput value={portalComplaintForm.customerId} onChange={(event) => {
                    const customerId = fieldNumber(event.target.value);
                    setPortalComplaintForm({
                      ...portalComplaintForm,
                      customerId: event.target.value,
                      projectId: String(portalProjectOptionsFor(customerId)[0]?.id || "")
                    });
                  }}>
                    <option value="">当前门户客户</option>
                    {portalCustomerOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                <Field label="项目">
                  <SelectInput value={portalComplaintForm.projectId} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, projectId: event.target.value })}>
                    <option value="">由后端按客户归属选择</option>
                    {portalProjectOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                  </SelectInput>
                </Field>
                <Field label="等级">
                  <SelectInput value={portalComplaintForm.level} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, level: event.target.value })}>
                    <option value="low">low</option>
                    <option value="medium">medium</option>
                    <option value="high">high</option>
                  </SelectInput>
                </Field>
                <Field label="标题" spanAll><TextInput value={portalComplaintForm.title} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, title: event.target.value })} required /></Field>
                <Field label="内容" spanAll><TextAreaInput value={portalComplaintForm.content} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, content: event.target.value })} /></Field>
                <FormActions spanAll>
                  <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !portalComplaintForm.title.trim()}>提交投诉</UiButton>
                </FormActions>
              </DialogForm>
            </ActionDialog>
          )}
          columns={[
            { key: "title", title: "投诉", render: (item) => <><b>{item.title}</b><span className="block-text muted">{item.content || "-"}</span></> },
            { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
            { key: "level", title: "等级", render: (item) => <StatusChip value={item.level} /> },
            { key: "owner", title: "负责人", render: (item) => item.owner || "-" },
            { key: "sla", title: "SLA", render: (item) => `${item.slaHours || 0}h / ${item.slaStatus || "-"}` },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
          ]}
          emptyText="暂无门户投诉"
        />
      </Panel>
    );
  }

  function renderContracts() {
    const contracts = activeContracts();
    const selectedContract = contracts.find((item) => item.id === fieldNumber(contractForm.contractId));
    const attachments = selectedContract ? contractAttachments(selectedContract.id) : [];
    const products = bootstrap?.products || [];

    function selectContract(item: Contract) {
      setContractForm((value) => ({
        ...value,
        contractId: String(item.id),
        customerId: String(item.customerId),
        projectId: String(item.projectId),
        productId: String(item.items?.[0]?.productId || value.productId),
        name: item.name,
        validFrom: item.validFrom || value.validFrom,
        validTo: item.validTo || value.validTo,
        quantity: String(item.items?.[0]?.quantity || value.quantity),
        unitPrice: String(item.items?.[0]?.unitPrice || value.unitPrice),
        reason: item.changeReason || value.reason,
        attachmentName: "",
        attachmentFileType: "contract_pdf",
        attachmentUrl: "",
        attachmentChecksum: ""
      }));
    }

    return (
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12">
          <DataTable
            data={contracts}
            rowKey={(item) => item.id}
            pageSize={12}
            emptyText="暂无合同"
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<Contract>({
              actions: [
                { key: "focus-contract", label: "只看该合同", onSelect: (item, helpers) => helpers.searchText(item.contractNo) },
                { key: "focus-customer", label: "只看该客户", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.customers, item.customerId)) },
                { key: "focus-project", label: "只看该项目", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.projects, item.projectId)) }
              ],
              copyFields: [
                { key: "contract", label: "合同编号", value: (item) => item.contractNo },
                { key: "name", label: "合同名称", value: (item) => item.name },
                { key: "customer", label: "客户", value: (item) => nameOf(bootstrap?.customers, item.customerId) },
                { key: "project", label: "项目", value: (item) => nameOf(bootstrap?.projects, item.projectId) },
                { key: "amount", label: "合同金额", value: (item) => money(item.totalAmount) }
              ]
            })}
            headerAction={(
              <div className="finance-header-actions">
                <span className="muted">{contracts.length} 份合同</span>
                <ActionDialog id="contract-management-actions" title="合同操作">
                  <div className="finance-hidden-actions">
                    <Field label="合同">
                      <SelectInput value={contractForm.contractId} onChange={(event) => {
                        const contract = contracts.find((item) => item.id === fieldNumber(event.target.value));
                        if (contract) {
                          selectContract(contract);
                          return;
                        }
                        setContractForm({ ...contractForm, contractId: event.target.value });
                      }}>
                        <option value="">选择合同</option>
                        {contracts.map((item) => <option key={item.id} value={item.id}>{item.contractNo} / {item.name}</option>)}
                      </SelectInput>
                    </Field>
                    {selectedContract ? (
                      <div className="finance-action-block">
                        <b>{selectedContract.contractNo} / v{selectedContract.version}</b>
                        <span>{nameOf(bootstrap?.customers, selectedContract.customerId)} / {money(selectedContract.totalAmount)} / {selectedContract.status}</span>
                        {workflowTimelineBlock(["contract", "contract_version"], selectedContract.id, selectedContract.contractNo, "当前合同暂无工作流实例")}
                        <ActionGroup className="compact-actions">
                          <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== "" || selectedContract.status === "pending_approval" || selectedContract.status === "active"} onClick={handleSubmitContract}>提审</UiButton>
                          <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={handleReviseContract}>修订</UiButton>
                        </ActionGroup>
                        {renderContractAttachmentList(selectedContract.id, attachments)}
                      </div>
                    ) : null}
                    <InlineForm onSubmit={handleCreateContract}>
                      <Field label="客户">
                        <SelectInput value={contractForm.customerId} onChange={(event) => setContractForm({ ...contractForm, customerId: event.target.value })}>
                          {list(bootstrap?.customers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="项目">
                        <SelectInput value={contractForm.projectId} onChange={(event) => setContractForm({ ...contractForm, projectId: event.target.value })}>
                          {list(bootstrap?.projects).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="产品">
                        <SelectInput value={contractForm.productId} onChange={(event) => setContractForm({ ...contractForm, productId: event.target.value })}>
                          {products.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.spec}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="合同名称"><TextInput value={contractForm.name} onChange={(event) => setContractForm({ ...contractForm, name: event.target.value })} /></Field>
                      <HeroDateField label="开始日期" value={contractForm.validFrom} onChange={(validFrom) => setContractForm({ ...contractForm, validFrom })} />
                      <HeroDateField label="结束日期" value={contractForm.validTo} onChange={(validTo) => setContractForm({ ...contractForm, validTo })} />
                      <Field label="数量"><TextInput type="number" value={contractForm.quantity} onChange={(event) => setContractForm({ ...contractForm, quantity: event.target.value })} /></Field>
                      <Field label="单价"><TextInput type="number" value={contractForm.unitPrice} onChange={(event) => setContractForm({ ...contractForm, unitPrice: event.target.value })} /></Field>
                      <Field label="原因"><TextInput value={contractForm.reason} onChange={(event) => setContractForm({ ...contractForm, reason: event.target.value })} /></Field>
                      <Field label="附件文件"><input type="file" onChange={handleContractAttachmentFile} /></Field>
                      <Field label="附件名"><TextInput value={contractForm.attachmentName} onChange={(event) => setContractForm({ ...contractForm, attachmentName: event.target.value })} /></Field>
                      <ActionGroup className="compact-actions">
                        <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !products.length}>创建合同</UiButton>
                        <UiButton icon={<Plus size={13} />} disabled={actionBusy !== "" || !fieldNumber(contractForm.contractId) || !contractForm.attachmentName || !contractForm.attachmentUrl} onClick={handleCreateContractAttachment}>归档附件</UiButton>
                      </ActionGroup>
                    </InlineForm>
                  </div>
                </ActionDialog>
              </div>
            )}
            columns={[
              { key: "contractNo", title: "合同编号", render: (item) => <b>{item.contractNo}</b> },
              { key: "name", title: "合同名称", render: (item) => item.name },
              { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) },
              { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
              { key: "version", title: "版本", render: (item) => `v${item.version}` },
              { key: "period", title: "有效期", render: (item) => `${item.validFrom} ~ ${item.validTo}` },
              { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
              { key: "status", title: "状态", render: (item) => workflowStatusFor(["contract", "contract_version"], item.id, item.contractNo, <StatusChip value={item.status} />) },
              { key: "actions", title: "操作", render: (item) => <UiButton icon={<Search size={13} />} onClick={() => selectContract(item)}>载入</UiButton> }
            ]}
          />
        </Panel>
      </SectionGrid>
    );
  }

  function renderPortal() {
    const portal = data.portal;
    const portalDispatches = [...list(portal?.dispatches)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const portalOrders = [...list(portal?.orders)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const portalStatements = [...list(portal?.statements)]
      .sort((a, b) => b.id - a.id);
    const portalInvoices = [...list(portal?.invoices)]
      .sort((a, b) => b.id - a.id);
    const portalSigns = [...list(portal?.signs)]
      .sort((a, b) => b.id - a.id);
    const portalSignLinks = [...list(portal?.signLinks)]
      .sort((a, b) => b.id - a.id);
    const portalAlarms = [...list(portal?.alarms)]
      .sort((a, b) => b.id - a.id);
    const portalComplaints = [...(data.portalComplaints.length ? data.portalComplaints : list(portal?.complaints))]
      .sort((a, b) => b.id - a.id);
    const portalProjectIds = new Set([
      ...portalOrders.map((item) => item.projectId),
      ...portalStatements.map((item) => item.projectId),
      ...portalComplaints.map((item) => item.projectId)
    ].filter(Boolean));
    const portalCustomerIds = new Set([
      ...portalOrders.map((item) => item.customerId),
      ...portalStatements.map((item) => item.customerId),
      ...portalComplaints.map((item) => item.customerId),
      bootstrap?.user.customerId || 0
    ].filter(Boolean));
    const portalProjectOptions = list(bootstrap?.projects).filter((item) => portalProjectIds.has(item.id));
    const portalCustomerOptions = list(bootstrap?.customers).filter((item) => portalCustomerIds.has(item.id));
    const portalProjectOptionsForCustomer = (customerId = fieldNumber(portalComplaintForm.customerId)) => portalProjectOptions.filter((item) => !customerId || item.customerId === customerId);
    const openStatements = portalStatements.filter((item) => !isCustomerStatementClosed(item.status));
    const activeDispatches = portalDispatches.filter((item) => !["completed", "signed", "cancelled", "void"].includes(item.status));
    const activeSignLinks = portalSignLinks.filter((item) => item.status !== "used" && item.status !== "expired");
    const vehicleForDispatch = (item: DispatchOrder) => list(bootstrap?.vehicles).find((vehicle) => vehicle.id === item.vehicleId);
    const orderForDispatch = (item: DispatchOrder) => orderById.get(item.orderId) || portalOrders.find((order) => order.id === item.orderId);
    const driverSignForDispatch = (dispatchId: number) => portalSigns.find((item) => item.dispatchId === dispatchId);

    const complaintFormView = (
      <DialogForm onSubmit={handleCreatePortalComplaint}>
        <Field label="客户">
          <SelectInput value={portalComplaintForm.customerId} onChange={(event) => {
            const customerId = fieldNumber(event.target.value);
            setPortalComplaintForm({
              ...portalComplaintForm,
              customerId: event.target.value,
              projectId: String(portalProjectOptionsForCustomer(customerId)[0]?.id || "")
            });
          }}>
            <option value="">当前门户客户</option>
            {portalCustomerOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="项目">
          <SelectInput value={portalComplaintForm.projectId} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, projectId: event.target.value })}>
            <option value="">由后端按客户归属选择</option>
            {portalProjectOptionsForCustomer().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="等级">
          <SelectInput value={portalComplaintForm.level} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, level: event.target.value })}>
            <option value="low">low</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
          </SelectInput>
        </Field>
        <Field label="标题" spanAll><TextInput value={portalComplaintForm.title} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, title: event.target.value })} required /></Field>
        <Field label="内容" spanAll><TextAreaInput value={portalComplaintForm.content} onChange={(event) => setPortalComplaintForm({ ...portalComplaintForm, content: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !portalComplaintForm.title.trim()}>提交投诉</UiButton>
        </FormActions>
      </DialogForm>
    );

    const exceptionFormView = (
      <DialogForm onSubmit={handleReportPortalDispatchException}>
        <Field label="派车单">
          <SelectInput value={portalExceptionForm.dispatchId} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, dispatchId: event.target.value })}>
            {portalDispatches.map((item) => <option key={item.id} value={item.id}>{item.dispatchNo} / {nameOf(bootstrap?.vehicles, item.vehicleId)}</option>)}
          </SelectInput>
        </Field>
        <Field label="等级">
          <SelectInput value={portalExceptionForm.level} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, level: event.target.value })}>
            <option value="low">low</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
          </SelectInput>
        </Field>
        <Field label="类型"><TextInput value={portalExceptionForm.alarmType} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, alarmType: event.target.value })} /></Field>
        <Field label="异常说明" spanAll><TextAreaInput value={portalExceptionForm.exception} onChange={(event) => setPortalExceptionForm({ ...portalExceptionForm, exception: event.target.value })} required /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<AlertCircle size={14} />} disabled={actionBusy !== "" || !fieldNumber(portalExceptionForm.dispatchId) || !portalExceptionForm.exception.trim()}>上报异常</UiButton>
        </FormActions>
      </DialogForm>
    );

    const locationReportFormView = (
      <DialogForm onSubmit={handleReportLocationBatch}>
        <Field label="车牌"><TextInput value={locationBatchForm.plateNo} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, plateNo: event.target.value })} required /></Field>
        <Field label="设备号"><TextInput value={locationBatchForm.deviceNo} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, deviceNo: event.target.value })} /></Field>
        <Field label="经度"><TextInput type="number" step="0.000001" value={locationBatchForm.longitude} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, longitude: event.target.value })} required /></Field>
        <Field label="纬度"><TextInput type="number" step="0.000001" value={locationBatchForm.latitude} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, latitude: event.target.value })} required /></Field>
        <Field label="速度"><TextInput type="number" min="0" step="0.1" value={locationBatchForm.speed} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, speed: event.target.value })} /></Field>
        <Field label="方向"><TextInput type="number" min="0" step="1" value={locationBatchForm.direction} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, direction: event.target.value })} /></Field>
        <Field label="里程"><TextInput type="number" min="0" step="0.1" value={locationBatchForm.mileage} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, mileage: event.target.value })} /></Field>
        <Field label="来源"><TextInput value={locationBatchForm.sourceType} onChange={(event) => setLocationBatchForm({ ...locationBatchForm, sourceType: event.target.value })} /></Field>
        {locationBatchResult ? (
          <div className="map-location-result">
            <div><span>总数</span><b>{locationBatchResult.total}</b></div>
            <div><span>接受</span><b>{locationBatchResult.accepted}</b></div>
            <div><span>拒绝</span><b>{locationBatchResult.rejected}</b></div>
          </div>
        ) : null}
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<MapPin size={14} />} disabled={actionBusy !== "" || !locationBatchForm.plateNo || !locationBatchForm.longitude || !locationBatchForm.latitude}>提交上报</UiButton>
        </FormActions>
      </DialogForm>
    );

    if (section === "portal-driver") {
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="司机门户范围">
                <span className="inventory-scope-chip strong">{bootstrap?.user.displayName || bootstrap?.user.username || "司机"}</span>
                <span className="inventory-scope-chip">{portalDispatches.length} 张派车单</span>
                <span className={`inventory-scope-chip ${activeDispatches.length ? "warning" : ""}`}>{activeDispatches.length} 张执行中</span>
                <span className="inventory-scope-chip">{portalAlarms.length} 条告警</span>
              </div>
              <ActionGroup>
                <ActionDialog
                  id="driver-portal-exception"
                  title="上报派车异常"
                  buttonLabel="上报异常"
                  triggerIcon={<AlertCircle size={13} />}
                  triggerVariant="primary"
                  disabled={!portalDispatches.length}
                  onOpen={() => setPortalExceptionForm({
                    dispatchId: String(activeDispatches[0]?.id || portalDispatches[0]?.id || ""),
                    level: "medium",
                    alarmType: "driver_exception",
                    exception: ""
                  })}
                >
                  {exceptionFormView}
                </ActionDialog>
                <ActionDialog
                  id="driver-portal-location"
                  title="上报车辆定位"
                  buttonLabel="上报定位"
                  triggerIcon={<MapPin size={13} />}
                  onOpen={() => {
                    const dispatch = activeDispatches[0] || portalDispatches[0];
                    const vehicle = dispatch ? vehicleForDispatch(dispatch) : undefined;
                    const latest = dispatch ? data.latestLocations.find((item) => item.vehicleId === dispatch.vehicleId) : undefined;
                    setLocationBatchForm({
                      plateNo: vehicle?.plateNo || "",
                      deviceNo: "",
                      longitude: latest?.longitude ? String(latest.longitude) : "",
                      latitude: latest?.latitude ? String(latest.latitude) : "",
                      speed: latest?.speed ? String(latest.speed) : "",
                      direction: latest?.direction ? String(latest.direction) : "",
                      mileage: "",
                      accStatus: "",
                      sourceType: "driver_portal"
                    });
                    setLocationBatchResult(null);
                  }}
                >
                  {locationReportFormView}
                </ActionDialog>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>派车任务</span><b>{portalDispatches.length}</b><small>当前账号范围</small></div>
              <div><span>执行中</span><b>{activeDispatches.length}</b><small>待装料/在途/待签收</small></div>
              <div><span>已签收</span><b>{portalSigns.length}</b><small>{qty(portalSigns.reduce((sum, item) => sum + item.signedQty, 0))}</small></div>
              <div><span>异常告警</span><b>{portalAlarms.length}</b><small>司机上报和规则告警</small></div>
              <div><span>签收链接</span><b>{activeSignLinks.length}</b><small>待使用链接</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<DispatchOrder>
              title="我的派车单"
              data={portalDispatches}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "dispatch", title: "派车单", render: (item) => <><b>{item.dispatchNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                { key: "vehicle", title: "车辆/司机", render: (item) => `${nameOf(bootstrap?.vehicles, item.vehicleId)} / ${nameOf(bootstrap?.drivers, item.driverId)}` },
                { key: "project", title: "客户项目", render: (item) => {
                  const order = orderForDispatch(item);
                  return `${order ? nameOf(bootstrap?.customers, order.customerId) : "-"} / ${nameOf(bootstrap?.projects, item.projectId)}`;
                } },
                { key: "product", title: "产品", render: (item) => item.productName || productLabel(bootstrap, item.productId) },
                { key: "qty", title: "计划/装载/签收", render: (item) => `${qty(item.planQuantity)} / ${qty(item.loadedQty)} / ${qty(driverSignForDispatch(item.id)?.signedQty || item.signedQty)}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => (
                    <ActionDialog
                      id={`driver-portal-exception-${item.id}`}
                      title="上报派车异常"
                      buttonLabel="异常"
                      onOpen={() => setPortalExceptionForm({
                        dispatchId: String(item.id),
                        level: "medium",
                        alarmType: "driver_exception",
                        exception: item.exception || ""
                      })}
                    >
                      {exceptionFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无派车任务"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<VehicleAlarm>
              title="车辆告警"
              data={portalAlarms}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "alarm", title: "告警", render: (item) => <><b>ALM-{item.id}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                { key: "vehicle", title: "车辆", render: (item) => nameOf(bootstrap?.vehicles, item.vehicleId) || `车辆 #${item.vehicleId}` },
                { key: "type", title: "类型/等级", render: (item) => `${item.alarmType} / ${item.level}` },
                { key: "message", title: "内容", render: (item) => item.message || "-" },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
              ]}
              emptyText="暂无车辆告警"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12 department-board finance-department-board">
          <div className="department-board-head">
            <div>
              <h3>客户门户</h3>
              <p>客户订单、送货签收、客户对账、税票下载和投诉反馈</p>
            </div>
            <ActionGroup>
              <ActionDialog
                id="customer-portal-complaint-create"
                title="提交客户投诉"
                buttonLabel="提交投诉"
                triggerIcon={<Plus size={13} />}
                triggerVariant="primary"
                onOpen={() => {
                  const customerId = portalCustomerOptions[0]?.id || bootstrap?.user.customerId || 0;
                  const project = portalProjectOptionsForCustomer(customerId)[0];
                  setPortalComplaintForm({
                    customerId: String(customerId || ""),
                    projectId: String(project?.id || ""),
                    level: "medium",
                    title: "",
                    content: ""
                  });
                }}
              >
                {complaintFormView}
              </ActionDialog>
              <ButtonLink icon={<ReceiptText size={15} />} href="/finance/statements">客户对账</ButtonLink>
              <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
            </ActionGroup>
          </div>
          <div className="finance-focus-grid">
            <Card className="finance-focus-card"><span>订单</span><b>{portalOrders.length}</b><small>客户可见订单</small></Card>
            <Card className="finance-focus-card"><span>待确认对账</span><b>{openStatements.length}</b><small>{money(openStatements.reduce((sum, item) => sum + item.totalAmount, 0))}</small></Card>
            <Card className="finance-focus-card"><span>发票</span><b>{portalInvoices.length}</b><small>{money(portalInvoices.reduce((sum, item) => sum + item.amount, 0))}</small></Card>
            <Card className="finance-focus-card"><span>签收</span><b>{portalSigns.length}</b><small>{qty(portalSigns.reduce((sum, item) => sum + item.signedQty, 0))}</small></Card>
          </div>
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable<SalesOrder>
            title="我的订单"
            data={portalOrders}
            rowKey={(item) => item.id}
            pageSize={10}
            onRefresh={refreshData}
            columns={[
              { key: "order", title: "订单", render: (item) => <><b>{item.orderNo}</b><span className="block-text muted">{shortDateTime(item.planTime)}</span></> },
              { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
              { key: "product", title: "产品", render: (item) => productLabel(bootstrap, item.productId) },
              { key: "qty", title: "计划/发货/签收", render: (item) => `${qty(item.planQuantity)} / ${qty(item.dispatchedQty)} / ${qty(item.signedQty)}` },
              { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
            ]}
            emptyText="暂无客户订单"
          />
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable<Statement>
            title="客户对账"
            data={portalStatements}
            rowKey={(item) => item.id}
            pageSize={10}
            onRefresh={refreshData}
            columns={[
              { key: "statement", title: "对账单", render: (item) => <><b>{item.statementNo}</b><span className="block-text muted">{item.period}</span></> },
              { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
              { key: "qty", title: "方量", render: (item) => qty(item.totalQty) },
              { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
              { key: "status", title: "状态", render: (item) => workflowStatusFor(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo, <StatusChip value={item.status} />) },
              {
                key: "actions",
                title: "操作",
                width: "120px",
                render: (item) => isCustomerStatementClosed(item.status) ? (
                  <span className="muted">{customerStatementClosedLabel(item.status)}</span>
                ) : (
                  <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`customer-portal-statement-${item.id}`, "对账单已确认", () => api.confirmStatement(item.id))}>确认</UiButton>
                )
              }
            ]}
            emptyText="暂无客户对账单"
          />
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable<SalesInvoice>
            title="客户发票"
            data={portalInvoices}
            rowKey={(item) => item.id}
            pageSize={8}
            columns={[
              { key: "invoice", title: "发票", render: (item) => <><b>{item.invoiceNo}</b><span className="block-text muted">{shortDateTime(item.issuedAt)}</span></> },
              { key: "statement", title: "对账单", render: (item) => portalStatements.find((statement) => statement.id === item.statementId)?.statementNo || `对账单 #${item.statementId}` },
              { key: "amount", title: "金额/税额", render: (item) => `${money(item.amount)} / ${money(item.taxAmount)}` },
              { key: "status", title: "税控状态", render: (item) => <StatusChip value={item.taxStatus} /> },
              { key: "actions", title: "操作", width: "120px", render: (item) => <UiButton size="sm" icon={<Download size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`customer-portal-invoice-download-${item.id}`, "发票文件已打开下载", () => downloadInvoiceFile(item.id))}>下载</UiButton> }
            ]}
            emptyText="暂无客户发票"
          />
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable<CustomerComplaint>
            title="投诉反馈"
            data={portalComplaints}
            rowKey={(item) => item.id}
            pageSize={8}
            columns={[
              { key: "complaint", title: "投诉", render: (item) => <><b>{item.complaintNo}</b><span className="block-text muted">{item.title}</span></> },
              { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
              { key: "level", title: "等级/SLA", render: (item) => `${item.level || "-"} / ${item.slaStatus || "-"}` },
              { key: "owner", title: "负责人", render: (item) => item.owner || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无投诉反馈"
          />
        </Panel>
      </SectionGrid>
    );
  }

  function renderSettlement() {
    const contracts = activeContracts();
    const selectedContract = contracts.find((item) => item.id === fieldNumber(contractForm.contractId)) || contracts[0];
    const attachments = selectedContract ? contractAttachments(selectedContract.id) : [];
    const products = bootstrap?.products || [];
    const statements = data.statements.length ? data.statements : list(data.finance?.statements);
    const statementInvoices = (statementId: number) => list(data.finance?.invoices).filter((item) => item.statementId === statementId);
    const carrierSettlements = data.carrierSettlements.sort((a, b) => b.id - a.id);
    const carrierSettlementItems = (settlementId: number) => data.carrierSettlementItems.filter((item) => item.settlementId === settlementId);
    const carrierSettlementFormView = (
      <DialogForm onSubmit={handleGenerateCarrierSettlement}>
        <Field label="承运商">
          <SelectInput value={carrierSettlementForm.carrierId} onChange={(event) => setCarrierSettlementForm({ ...carrierSettlementForm, carrierId: event.target.value })}>
            <option value="">全部承运商</option>
            {list(bootstrap?.carriers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="期间"><TextInput value={carrierSettlementForm.period} onChange={(event) => setCarrierSettlementForm({ ...carrierSettlementForm, period: event.target.value })} /></Field>
        <Field label="按趟费率"><TextInput type="number" min="0" step="1" value={carrierSettlementForm.ratePerTrip} onChange={(event) => setCarrierSettlementForm({ ...carrierSettlementForm, ratePerTrip: event.target.value })} /></Field>
        <Field label="按量费率"><TextInput type="number" min="0" step="0.01" value={carrierSettlementForm.ratePerUnit} onChange={(event) => setCarrierSettlementForm({ ...carrierSettlementForm, ratePerUnit: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<ReceiptText size={14} />} disabled={actionBusy !== ""}>生成承运结算</UiButton>
        </FormActions>
      </DialogForm>
    );

    if (section === "finance-carriers") {
      const settlementAmount = carrierSettlements.reduce((sum, item) => sum + item.amount, 0);
      const settlementTrips = carrierSettlements.reduce((sum, item) => sum + item.tripCount, 0);
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board finance-department-board">
            <div className="department-board-head">
              <div>
                <h3>承运结算</h3>
                <p>承运商运单汇总、结算生成和结算明细核对</p>
              </div>
              <ActionGroup>
                <ActionDialog
                  id="carrier-settlement-generate-page"
                  title="生成承运结算"
                  buttonLabel="生成承运结算"
                  triggerIcon={<ReceiptText size={13} />}
                  triggerVariant="primary"
                  onOpen={() => setCarrierSettlementForm({
                    carrierId: String(list(bootstrap?.carriers)[0]?.id || ""),
                    period: today.slice(0, 7),
                    ratePerTrip: "",
                    ratePerUnit: ""
                  })}
                >
                  {carrierSettlementFormView}
                </ActionDialog>
                <ButtonLink icon={<Truck size={15} />} href="/resources/carriers">承运商</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </div>
            <div className="finance-focus-grid">
              <Card className="finance-focus-card"><span>结算单</span><b>{carrierSettlements.length}</b><small>{data.carrierSettlementItems.length} 条运单明细</small></Card>
              <Card className="finance-focus-card"><span>结算金额</span><b>{money(settlementAmount)}</b><small>{settlementTrips} 趟</small></Card>
              <Card className="finance-focus-card"><span>承运商</span><b>{list(bootstrap?.carriers).length}</b><small>可生成结算</small></Card>
              <Card className="finance-focus-card"><span>待确认</span><b>{carrierSettlements.filter((item) => item.status !== "approved" && item.status !== "closed").length}</b><small>待流程或人工核对</small></Card>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<TransportSettlement>
              title="承运商结算"
              data={carrierSettlements}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "settlementNo", title: "结算单", render: (item) => <b>{item.settlementNo}</b> },
                { key: "carrier", title: "承运商", render: (item) => nameOf(bootstrap?.carriers, item.carrierId) },
                { key: "period", title: "期间", render: (item) => item.period },
                { key: "trips", title: "趟次", render: (item) => qty(item.tripCount) },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["transport_settlement", "carrier_settlement"], item.id, item.settlementNo, <StatusChip value={item.status} />) },
                {
                  key: "actions",
                  title: "明细",
                  width: "110px",
                  render: (item) => {
                    const items = carrierSettlementItems(item.id);
                    return (
                      <ActionDialog id={`carrier-settlement-page-${item.id}`} title="承运结算明细" buttonLabel="明细">
                        <div className="finance-hidden-actions">
                          <div className="finance-action-block">
                            <b>{item.settlementNo}</b>
                            <span>{nameOf(bootstrap?.carriers, item.carrierId)} / {item.period} / {money(item.amount)}</span>
                          </div>
                          {workflowTimelineBlock(["transport_settlement", "carrier_settlement"], item.id, item.settlementNo, "当前承运结算暂无工作流实例")}
                          {items.map((line) => (
                            <div className="finance-action-block" key={line.id}>
                              <b>{line.dispatchNo}</b>
                              <span>{nameOf(bootstrap?.vehicles, line.vehicleId)} / {nameOf(bootstrap?.drivers, line.driverId)} / {qty(line.quantity)} / {money(line.amount)}</span>
                            </div>
                          ))}
                          {!items.length ? <span className="muted">暂无运单明细</span> : null}
                        </div>
                      </ActionDialog>
                    );
                  }
                }
              ]}
              emptyText="暂无承运结算"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<TransportSettlementItem>
              title="承运运单明细"
              data={data.carrierSettlementItems}
              rowKey={(item) => item.id}
              pageSize={12}
              columns={[
                { key: "dispatch", title: "派车单", render: (item) => <b>{item.dispatchNo}</b> },
                { key: "settlement", title: "结算单", render: (item) => carrierSettlements.find((settlement) => settlement.id === item.settlementId)?.settlementNo || `结算 #${item.settlementId}` },
                { key: "carrier", title: "承运商", render: (item) => nameOf(bootstrap?.carriers, item.carrierId) },
                { key: "vehicle", title: "车辆", render: (item) => nameOf(bootstrap?.vehicles, item.vehicleId) },
                { key: "driver", title: "司机", render: (item) => nameOf(bootstrap?.drivers, item.driverId) },
                { key: "qty", title: "数量", render: (item) => qty(item.quantity) },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) }
              ]}
              emptyText="暂无承运运单明细"
            />
          </Panel>
        </SectionGrid>
      );
    }

    function selectContract(item: Contract) {
      setContractForm((value) => ({
        ...value,
        contractId: String(item.id),
        customerId: String(item.customerId),
        projectId: String(item.projectId),
        productId: String(item.items?.[0]?.productId || value.productId),
        name: item.name,
        validFrom: item.validFrom || value.validFrom,
        validTo: item.validTo || value.validTo,
        quantity: String(item.items?.[0]?.quantity || value.quantity),
        unitPrice: String(item.items?.[0]?.unitPrice || value.unitPrice),
        reason: item.changeReason || value.reason,
        attachmentName: "",
        attachmentFileType: "contract_pdf",
        attachmentUrl: "",
        attachmentChecksum: ""
      }));
    }

    return (
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12">
          <DataTable
            data={statements}
            rowKey={(item) => item.id}
            pageSize={12}
            emptyText="暂无对账单"
            rowContextMenu={buildDataTableRowContextMenu<Statement>({
              actions: [
                { key: "focus-customer", label: "只看该客户", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.customers, item.customerId)) },
                { key: "focus-project", label: "只看该项目", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.projects, item.projectId)) },
                { key: "focus-period", label: "只看该期间", onSelect: (item, helpers) => helpers.searchText(item.period) }
              ],
              copyFields: [
                { key: "statement", label: "对账单号", value: (item) => item.statementNo },
                { key: "customer", label: "客户", value: (item) => nameOf(bootstrap?.customers, item.customerId) },
                { key: "project", label: "项目", value: (item) => nameOf(bootstrap?.projects, item.projectId) },
                { key: "amount", label: "金额", value: (item) => money(item.totalAmount) }
              ]
            })}
            headerAction={
              <div className="finance-header-actions">
                <span className="muted">{statements.length} 张对账单 / {contracts.length} 份合同</span>
                <ActionDialog id="settlement-contract-actions" title="合同操作">
                  <div className="finance-hidden-actions">
                    <Field label="合同">
                      <SelectInput value={contractForm.contractId} onChange={(event) => {
                        const contract = contracts.find((item) => item.id === fieldNumber(event.target.value));
                        if (contract) {
                          selectContract(contract);
                          return;
                        }
                        setContractForm({ ...contractForm, contractId: event.target.value });
                      }}>
                        <option value="">选择合同</option>
                        {contracts.map((item) => <option key={item.id} value={item.id}>{item.contractNo} / {item.name}</option>)}
                      </SelectInput>
                    </Field>
                    {selectedContract ? (
	                      <div className="finance-action-block">
	                        <b>{selectedContract.contractNo} / v{selectedContract.version}</b>
	                        <span>{nameOf(bootstrap?.customers, selectedContract.customerId)} / {money(selectedContract.totalAmount)} / {selectedContract.status}</span>
	                        {workflowTimelineBlock(["contract", "contract_version"], selectedContract.id, selectedContract.contractNo, "当前合同暂无工作流实例")}
	                        <ActionGroup className="compact-actions">
                          <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== "" || selectedContract.status === "pending_approval" || selectedContract.status === "active"} onClick={handleSubmitContract}>提审</UiButton>
                          <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={handleReviseContract}>修订</UiButton>
                        </ActionGroup>
                        {renderContractAttachmentList(selectedContract.id, attachments)}
                      </div>
                    ) : null}
                    <InlineForm onSubmit={handleCreateContract}>
                      <Field label="客户">
                        <SelectInput value={contractForm.customerId} onChange={(event) => setContractForm({ ...contractForm, customerId: event.target.value })}>
                          {list(bootstrap?.customers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="项目">
                        <SelectInput value={contractForm.projectId} onChange={(event) => setContractForm({ ...contractForm, projectId: event.target.value })}>
                          {list(bootstrap?.projects).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="产品">
                        <SelectInput value={contractForm.productId} onChange={(event) => setContractForm({ ...contractForm, productId: event.target.value })}>
                          {products.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.spec}</option>)}
                        </SelectInput>
                      </Field>
                      <Field label="合同名称"><TextInput value={contractForm.name} onChange={(event) => setContractForm({ ...contractForm, name: event.target.value })} /></Field>
                      <HeroDateField label="开始日期" value={contractForm.validFrom} onChange={(validFrom) => setContractForm({ ...contractForm, validFrom })} />
                      <HeroDateField label="结束日期" value={contractForm.validTo} onChange={(validTo) => setContractForm({ ...contractForm, validTo })} />
                      <Field label="数量"><TextInput type="number" value={contractForm.quantity} onChange={(event) => setContractForm({ ...contractForm, quantity: event.target.value })} /></Field>
                      <Field label="单价"><TextInput type="number" value={contractForm.unitPrice} onChange={(event) => setContractForm({ ...contractForm, unitPrice: event.target.value })} /></Field>
                      <Field label="原因"><TextInput value={contractForm.reason} onChange={(event) => setContractForm({ ...contractForm, reason: event.target.value })} /></Field>
                      <Field label="附件文件"><input type="file" onChange={handleContractAttachmentFile} /></Field>
                      <Field label="附件名"><TextInput value={contractForm.attachmentName} onChange={(event) => setContractForm({ ...contractForm, attachmentName: event.target.value })} /></Field>
                      <ActionGroup className="compact-actions">
                        <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !products.length}>创建合同</UiButton>
                        <UiButton icon={<Plus size={13} />} disabled={actionBusy !== "" || !fieldNumber(contractForm.contractId) || !contractForm.attachmentName || !contractForm.attachmentUrl} onClick={handleCreateContractAttachment}>归档附件</UiButton>
                      </ActionGroup>
                    </InlineForm>
                  </div>
                </ActionDialog>
              </div>
            }
            columns={[
              { key: "statementNo", title: "对账单", render: (item) => <b>{item.statementNo}</b> },
              { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) },
              { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
              { key: "period", title: "期间", render: (item) => item.period },
              { key: "qty", title: "方量", render: (item) => qty(item.totalQty) },
              { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
	              {
	                key: "status",
	                title: "状态",
	                render: (item) => workflowStatusFor(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo, approvalStatus(approvalFor(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo), <StatusChip value={item.status} />))
	              },
	              { key: "actions", title: "操作", render: (item) => {
	                const invoices = statementInvoices(item.id);
	                const task = approvalFor(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo);
	                const workflow = workflowItemsFor(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo);
	                return (
                  <ActionDialog id={`settlement-statement-${item.id}`} title="对账操作">
                    <div className="finance-hidden-actions">
                      <div className="finance-action-block">
                        <b>对账操作</b>
                        <span>{item.period} / {qty(item.totalQty)} 方 / {money(item.totalAmount)}</span>
	                        {isCustomerStatementClosed(item.status) ? <span>{customerStatementClosedLabel(item.status)}</span> : workflow.instances.length ? <span>流程处理中</span> : (
	                          <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`statement-${item.id}`, "对账单已确认", () => api.confirmStatement(item.id))}>确认对账</UiButton>
	                        )}
                      </div>
	                      {workflowTimelineBlock(["statement", "customer_statement", "finance_statement"], item.id, item.statementNo, "当前对账单暂无工作流实例")}
	                      {!workflow.instances.length ? approvalActionBlock(task) : null}
                      <div className="finance-action-block">
                        <b>发票</b>
                        {invoices.map((invoice) => <span key={invoice.id}>{invoice.invoiceNo} / {money(invoice.amount)} / {invoice.taxStatus}</span>)}
                        {!invoices.length ? <span>暂无发票</span> : null}
                      </div>
                    </div>
                  </ActionDialog>
                );
              } }
            ]}
          />
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable
            title="承运商结算"
            data={carrierSettlements}
            rowKey={(item) => item.id}
            pageSize={8}
            emptyText="暂无承运结算"
            headerLeftAction={(
              <ActionDialog
                id="carrier-settlement-generate"
                title="生成承运结算"
                buttonLabel="生成承运结算"
                triggerIcon={<ReceiptText size={13} />}
                onOpen={() => setCarrierSettlementForm({
                  carrierId: String(list(bootstrap?.carriers)[0]?.id || ""),
                  period: today.slice(0, 7),
                  ratePerTrip: "",
                  ratePerUnit: ""
                })}
              >
                {carrierSettlementFormView}
              </ActionDialog>
            )}
            headerAction={<span className="muted">{carrierSettlements.length} 张结算单 / {data.carrierSettlementItems.length} 条运单明细</span>}
            columns={[
              { key: "settlementNo", title: "结算单", render: (item) => <b>{item.settlementNo}</b> },
              { key: "carrier", title: "承运商", render: (item) => nameOf(bootstrap?.carriers, item.carrierId) },
              { key: "period", title: "期间", render: (item) => item.period },
              { key: "trips", title: "趟次", render: (item) => qty(item.tripCount) },
              { key: "amount", title: "金额", render: (item) => money(item.amount) },
              { key: "status", title: "状态", render: (item) => workflowStatusFor(["transport_settlement", "carrier_settlement"], item.id, item.settlementNo, <StatusChip value={item.status} />) },
              {
                key: "actions",
                title: "明细",
                render: (item) => {
                  const items = carrierSettlementItems(item.id);
                  return (
                    <ActionDialog id={`carrier-settlement-${item.id}`} title="承运结算明细" buttonLabel="明细">
                      <div className="finance-hidden-actions">
                        <div className="finance-action-block">
                          <b>{item.settlementNo}</b>
                          <span>{nameOf(bootstrap?.carriers, item.carrierId)} / {item.period} / {money(item.amount)}</span>
                        </div>
                        {items.map((line) => (
                          <div className="finance-action-block" key={line.id}>
                            <b>{line.dispatchNo}</b>
                            <span>{nameOf(bootstrap?.vehicles, line.vehicleId)} / {nameOf(bootstrap?.drivers, line.driverId)} / {qty(line.quantity)} / {money(line.amount)}</span>
                          </div>
                        ))}
                        {!items.length ? <span className="muted">暂无运单明细</span> : null}
                      </div>
                    </ActionDialog>
                  );
                }
              }
            ]}
          />
        </Panel>
      </SectionGrid>
    );
  }

  function renderProcurement() {
    const procurement = data.procurement;
    const suppliers = supplierOptions();
    const purchaseRequests = list(procurement?.requests).filter((item) => matchesCurrentSite(item.siteId));
    const purchaseRequestIds = new Set(purchaseRequests.map((item) => item.id));
    const purchaseOrders = list(procurement?.orders).filter((item) => !selectedSiteId || !item.requestId || purchaseRequestIds.has(item.requestId));
    const inventoryItems = list(procurement?.inventory).filter((item) => matchesCurrentSite(item.siteId));
    const stocktakes = list(procurement?.stocktakes).filter((item) => matchesCurrentSite(item.siteId)).sort((a, b) => b.id - a.id);
    const stockYards = list(procurement?.stockYards).filter((item) => matchesCurrentSite(item.siteId));
    const stockYardPiles = list(procurement?.stockYardPiles).filter((item) => matchesCurrentSite(item.siteId));
    const stockYardFlows = [...list(procurement?.stockYardFlows)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const inventoryFlows = [...list(procurement?.flows)]
      .filter((item) => matchesCurrentSite(item.siteId))
      .sort((a, b) => b.id - a.id);
    const transferCount = list(procurement?.transfers).filter((item) => !selectedSiteId || item.fromSiteId === selectedSiteId || item.toSiteId === selectedSiteId).length;
    const plantBufferSource = list(data.production?.plantBufferLocations);
    const plantBuffers = plantBufferSource.filter((item) => matchesCurrentSite(item.siteId));
    const isPileIssue = (item: typeof stockYardPiles[number]) => item.currentQty <= item.warningQty || item.qualityStatus === "blocked" || item.status === "disabled";
    const stockYardCards = stockYards.map((yard) => {
      const piles = stockYardPiles.filter((item) => item.yardId === yard.id);
      const balance = piles.reduce((sum, item) => sum + item.currentQty, 0);
      const pileCapacity = piles.reduce((sum, item) => sum + item.capacity, 0);
      const issuePiles = piles.filter(isPileIssue);
      const materialIds = Array.from(new Set(piles.map((item) => item.materialId).filter(Boolean)));
      const materialSummary = materialIds.map((id) => nameOf(bootstrap?.materials, id) || `物料 #${id}`).slice(0, 3).join(" / ");
      const unit = piles[0]?.unit || yard.unit || "t";
      return {
        yard,
        piles,
        balance,
        capacity: yard.capacity || pileCapacity,
        issuePiles,
        materialSummary,
        unit
      };
    }).sort((left, right) => right.issuePiles.length - left.issuePiles.length || right.balance - left.balance);
    const inventoryBalance = inventoryItems.reduce((sum, item) => sum + item.quantity, 0);
    const yardBalance = stockYardPiles.reduce((sum, item) => sum + item.currentQty, 0);
    const lineBufferBalance = plantBuffers.reduce((sum, item) => sum + item.currentQty, 0);
    const lowInventoryCount = inventoryItems.filter((item) => item.availableStatus === "warning" || item.qualityStatus === "blocked").length;
    const lowYardCount = stockYardPiles.filter((item) => item.currentQty <= item.warningQty || item.qualityStatus === "blocked" || item.status === "disabled").length;
    const lowLineBufferCount = plantBuffers.filter((item) => item.currentQty <= item.warningQty || item.qualityStatus === "blocked" || item.status === "disabled").length;
    const pendingReceiptOrders = purchaseOrders.filter((item) => !["closed", "cancelled", "completed"].includes(item.status)).length;
    const riskCount = lowInventoryCount + lowYardCount + lowLineBufferCount;
    const inventoryScopeLabel = selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点";

    const rawReceipts = (list(data.quality?.receipts).length ? list(data.quality?.receipts) : list(procurement?.receipts)).filter((item) => matchesCurrentSite(item.siteId));
    const rawInspections = list(data.quality?.rawInspections).filter((item) => matchesCurrentSite(item.siteId));
    const inspectedReceiptIds = new Set(rawInspections.map((item) => item.receiptId));
    const rawInspectionCandidates = rawReceipts.filter((item) => !inspectedReceiptIds.has(item.id));
    const rawQualityResultOptions = dictionaryOptionsWithFallback("quality_result", [
      { code: "passed", label: "合格" },
      { code: "failed", label: "不合格" }
    ]);
    const purchaseRequestForOrder = (orderId: number) => {
      const order = purchaseOrders.find((item) => item.id === orderId);
      return order ? purchaseRequests.find((request) => request.id === order.requestId) : undefined;
    };
    const startRawReceipt = (order = purchaseOrders[0]) => {
      const request = order ? purchaseRequests.find((item) => item.id === order.requestId) : undefined;
      setProcurementForm({
        purchaseOrderId: order ? String(order.id) : "",
        supplierId: order?.supplierId ? String(order.supplierId) : String(recordId(suppliers[0]) || ""),
        siteId: String(request?.siteId || selectedSiteId || defaultSiteId || firstId(bootstrap?.sites)),
        materialId: String(order?.materialId || request?.materialId || firstId(bootstrap?.materials)),
        plateNo: "",
        grossWeight: "",
        tareWeight: ""
      });
    };
    const rawReceiptFormView = (
      <DialogForm onSubmit={handleCreateRawReceipt}>
        <Field label="采购单">
          <SelectInput value={procurementForm.purchaseOrderId} onChange={(event) => {
            const order = purchaseOrders.find((item) => String(item.id) === event.target.value);
            const request = order ? purchaseRequests.find((item) => item.id === order.requestId) : undefined;
            setProcurementForm({
              ...procurementForm,
              purchaseOrderId: event.target.value,
              supplierId: order?.supplierId ? String(order.supplierId) : procurementForm.supplierId,
              siteId: String(request?.siteId || procurementForm.siteId || selectedSiteId || defaultSiteId),
              materialId: String(order?.materialId || request?.materialId || procurementForm.materialId)
            });
          }}>
            {purchaseOrders.map((item) => <option key={item.id} value={item.id}>{item.orderNo} / {nameOf(bootstrap?.materials, item.materialId)} / {qty(item.quantity)} {item.unit}</option>)}
          </SelectInput>
        </Field>
        <Field label="供应商">
          <SelectInput value={procurementForm.supplierId} onChange={(event) => setProcurementForm({ ...procurementForm, supplierId: event.target.value })}>
            {suppliers.map((item) => <option key={recordId(item)} value={recordId(item)}>{recordName(item, "-")}</option>)}
          </SelectInput>
        </Field>
        <Field label="站点">
          <SelectInput value={procurementForm.siteId} onChange={(event) => setProcurementForm({ ...procurementForm, siteId: event.target.value })}>
            {siteOptions().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="物料">
          <SelectInput value={procurementForm.materialId} onChange={(event) => setProcurementForm({ ...procurementForm, materialId: event.target.value })}>
            {(bootstrap?.materials || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="车牌号"><TextInput value={procurementForm.plateNo} onChange={(event) => setProcurementForm({ ...procurementForm, plateNo: event.target.value })} /></Field>
        <Field label="毛重"><TextInput type="number" min="0" step="0.01" value={procurementForm.grossWeight} onChange={(event) => setProcurementForm({ ...procurementForm, grossWeight: event.target.value })} /></Field>
        <Field label="皮重"><TextInput type="number" min="0" step="0.01" value={procurementForm.tareWeight} onChange={(event) => setProcurementForm({ ...procurementForm, tareWeight: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton
            variant="primary"
            type="submit"
            icon={<Plus size={14} />}
            disabled={actionBusy !== "" || !fieldNumber(procurementForm.purchaseOrderId) || !fieldNumber(procurementForm.siteId) || !fieldNumber(procurementForm.materialId)}
          >
            生成入库单
          </UiButton>
        </FormActions>
      </DialogForm>
    );
    const rawInspectionFormView = (
      <DialogForm onSubmit={handleCreateRawMaterialInspection}>
        <Field label="入库单">
          <SelectInput value={rawInspectionForm.receiptId} onChange={(event) => setRawInspectionForm({ ...rawInspectionForm, receiptId: event.target.value })}>
            {rawInspectionCandidates.map((item) => <option key={item.id} value={item.id}>{item.receiptNo} / {nameOf(bootstrap?.materials, item.materialId)} / {qty(item.netWeight)}</option>)}
          </SelectInput>
        </Field>
        <Field label="检验员"><TextInput value={rawInspectionForm.inspector} onChange={(event) => setRawInspectionForm({ ...rawInspectionForm, inspector: event.target.value })} /></Field>
        <Field label="样品号"><TextInput value={rawInspectionForm.sampleNo} onChange={(event) => setRawInspectionForm({ ...rawInspectionForm, sampleNo: event.target.value })} /></Field>
        <Field label="备注" spanAll><TextAreaInput value={rawInspectionForm.remark} onChange={(event) => setRawInspectionForm({ ...rawInspectionForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !fieldNumber(rawInspectionForm.receiptId)}>创建原材质检</UiButton>
        </FormActions>
      </DialogForm>
    );
    const rawInspectionReviewFormView = (
      <DialogForm onSubmit={handleReviewRawMaterialInspection}>
        <Field label="水分"><TextInput type="number" min="0" step="0.01" value={rawInspectionReviewForm.moisture} onChange={(event) => setRawInspectionReviewForm({ ...rawInspectionReviewForm, moisture: event.target.value })} /></Field>
        <Field label="含泥量"><TextInput type="number" min="0" step="0.01" value={rawInspectionReviewForm.mudContent} onChange={(event) => setRawInspectionReviewForm({ ...rawInspectionReviewForm, mudContent: event.target.value })} /></Field>
        <Field label="细度"><TextInput value={rawInspectionReviewForm.fineness} onChange={(event) => setRawInspectionReviewForm({ ...rawInspectionReviewForm, fineness: event.target.value })} /></Field>
        <Field label="结论">
          <SelectInput value={rawInspectionReviewForm.result} onChange={(event) => setRawInspectionReviewForm({ ...rawInspectionReviewForm, result: event.target.value })}>
            {rawQualityResultOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
          </SelectInput>
        </Field>
        <Field label="备注" spanAll><TextAreaInput value={rawInspectionReviewForm.remark} onChange={(event) => setRawInspectionReviewForm({ ...rawInspectionReviewForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !fieldNumber(rawInspectionReviewForm.inspectionId)}>提交复核</UiButton>
        </FormActions>
      </DialogForm>
    );
    const stocktakeFormView = (
      <DialogForm onSubmit={handleCreateInventoryStocktake}>
        <Field label="站点">
          <SelectInput value={stocktakeForm.siteId} onChange={(event) => setStocktakeForm({ ...stocktakeForm, siteId: event.target.value })}>
            {siteOptions().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="物料">
          <SelectInput value={stocktakeForm.materialId} onChange={(event) => {
            const materialId = Number(event.target.value);
            const item = inventoryItems.find((candidate) => candidate.materialId === materialId && String(candidate.siteId) === stocktakeForm.siteId);
            setStocktakeForm({ ...stocktakeForm, materialId: event.target.value, actualQty: item ? String(item.quantity) : stocktakeForm.actualQty, unit: item?.unit || stocktakeForm.unit });
          }}>
            {(bootstrap?.materials || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="实际数量"><TextInput type="number" min="0" step="0.01" value={stocktakeForm.actualQty} onChange={(event) => setStocktakeForm({ ...stocktakeForm, actualQty: event.target.value })} /></Field>
        <Field label="单位"><TextInput value={stocktakeForm.unit} onChange={(event) => setStocktakeForm({ ...stocktakeForm, unit: event.target.value })} /></Field>
        <Field label="备注" spanAll><TextAreaInput value={stocktakeForm.remark} onChange={(event) => setStocktakeForm({ ...stocktakeForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<ClipboardCheck size={14} />} disabled={actionBusy !== "" || !fieldNumber(stocktakeForm.siteId) || !fieldNumber(stocktakeForm.materialId)}>提交盘点</UiButton>
        </FormActions>
      </DialogForm>
    );
    function openStockYardPileMenu(event: ReactMouseEvent, pile: StockYardPile) {
      event.preventDefault();
      event.stopPropagation();
      setStockYardMenu(null);
      setStockYardPileMenu({ pileId: pile.id, x: event.clientX, y: event.clientY });
    }
    function openStockYardMenu(event: ReactMouseEvent, yard: StockYard) {
      event.preventDefault();
      event.stopPropagation();
      setStockYardPileMenu(null);
      setStockYardMenu({ yardId: yard.id, x: event.clientX, y: event.clientY });
    }
    const stockYardMenuYard = stockYardMenu ? stockYards.find((item) => item.id === stockYardMenu.yardId) : undefined;
    const stockYardContextItems: ContextMenuItem[] = stockYardMenuYard ? [
      {
        key: "edit-yard",
        label: "编辑堆场",
        icon: <Pencil size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => openYardDialog("yard-edit", stockYardMenuYard)
      },
      {
        key: "add-pile",
        label: "新增堆位",
        icon: <Plus size={14} />,
        disabled: actionBusy !== "" || !bootstrap?.materials?.length,
        onSelect: () => openYardDialog("pile", stockYardMenuYard)
      },
      {
        key: "toggle-yard-status",
        label: stockYardMenuYard.status === "disabled" ? "启用堆场" : "停用堆场",
        icon: stockYardMenuYard.status === "disabled" ? <CheckCircle2 size={14} /> : <X size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => void handleStockYardStatus(stockYardMenuYard)
      },
      { key: "yard-danger-separator", type: "separator" },
      {
        key: "delete-yard",
        label: "删除堆场",
        icon: <X size={14} />,
        danger: true,
        disabled: actionBusy !== "",
        onSelect: () => void handleDeleteStockYard(stockYardMenuYard)
      }
    ] : [];
    const stockYardPileMenuPile = stockYardPileMenu ? stockYardPiles.find((item) => item.id === stockYardPileMenu.pileId) : undefined;
    const stockYardPileMenuYard = stockYardPileMenuPile ? stockYards.find((item) => item.id === stockYardPileMenuPile.yardId) : undefined;
    const stockYardPileMenuWorkflow = stockYardPileMenuPile ? workflowItemsFor(["stock_yard_adjustment"], stockYardPileMenuPile.id, stockYardPileMenuPile.code) : undefined;
    const stockYardPileMenuHasPendingAdjustmentWorkflow = !!stockYardPileMenuWorkflow?.instances.some((item) => item.status === "pending");
    const stockYardPileContextItems: ContextMenuItem[] = stockYardPileMenuPile ? [
      {
        key: "edit-pile",
        label: "编辑堆位",
        icon: <Pencil size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => openYardDialog("pile-edit", stockYardPileMenuYard, stockYardPileMenuPile)
      },
      {
        key: "receipt-pile",
        label: "入场",
        icon: <ArrowDown size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => openYardDialog("receipt", stockYardPileMenuYard, stockYardPileMenuPile)
      },
      {
        key: "adjust-pile",
        label: "盘点",
        icon: <ListChecks size={14} />,
        disabled: actionBusy !== "" || stockYardPileMenuHasPendingAdjustmentWorkflow,
        onSelect: () => openYardDialog("adjust", stockYardPileMenuYard, stockYardPileMenuPile)
      },
      {
        key: "toggle-pile-status",
        label: stockYardPileMenuPile.status === "disabled" ? "启用堆位" : "停用堆位",
        icon: stockYardPileMenuPile.status === "disabled" ? <CheckCircle2 size={14} /> : <X size={14} />,
        disabled: actionBusy !== "",
        onSelect: () => void handleStockYardPileStatus(stockYardPileMenuPile)
      },
      { key: "pile-danger-separator", type: "separator" },
      {
        key: "delete-pile",
        label: "删除堆位",
        icon: <X size={14} />,
        danger: true,
        disabled: actionBusy !== "",
        onSelect: () => void handleDeleteStockYardPile(stockYardPileMenuPile)
      }
    ] : [];

    if (section === "raw-material-receipts") {
      const receiptNetWeight = rawReceipts.reduce((sum, item) => sum + item.netWeight, 0);
      const pendingOrders = purchaseOrders.filter((item) => !["closed", "cancelled", "completed"].includes(item.status));
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board inventory-department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="原料收料范围">
                <span className="inventory-scope-chip strong">{inventoryScopeLabel}</span>
                <span className="inventory-scope-chip">{pendingOrders.length} 张待收采购单</span>
                <span className="inventory-scope-chip">{rawReceipts.length} 张入库单</span>
                <span className={`inventory-scope-chip ${rawInspectionCandidates.length ? "warning" : ""}`}>{rawInspectionCandidates.length ? `${rawInspectionCandidates.length} 张待检` : "质检已覆盖"}</span>
              </div>
              <ActionGroup>
                <ActionDialog
                  id="raw-receipt-create-page"
                  title="生成原料入库单"
                  buttonLabel="生成入库"
                  triggerIcon={<Plus size={14} />}
                  triggerVariant="primary"
                  disabled={!purchaseOrders.length}
                  onOpen={() => startRawReceipt(pendingOrders[0] || purchaseOrders[0])}
                >
                  {rawReceiptFormView}
                </ActionDialog>
                <ButtonLink icon={<ClipboardCheck size={15} />} href="/quality/raw-material-inspections">原材质检</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>待收采购单</span><b>{pendingOrders.length}</b><small>{purchaseOrders.length} 张采购单</small></div>
              <div><span>入库单</span><b>{rawReceipts.length}</b><small>{qty(receiptNetWeight)} t 净重</small></div>
              <div><span>待检入库</span><b>{rawInspectionCandidates.length}</b><small>入库后进入原材质检</small></div>
              <div><span>库存流水</span><b>{inventoryFlows.length}</b><small>入库/调拨/盘点流转</small></div>
              <div><span>库存批次</span><b>{inventoryItems.length}</b><small>{qty(inventoryBalance)} t</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="待收采购单"
              data={pendingOrders}
              rowKey={(item) => item.id}
              pageSize={8}
              onRefresh={refreshData}
              columns={[
                { key: "order", title: "采购单", render: (item) => <><b>{item.orderNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                { key: "request", title: "申请", render: (item) => {
                  const request = purchaseRequestForOrder(item.id);
                  return <><span>{request?.requestNo || `申请 #${item.requestId}`}</span><span className="block-text muted">{request ? nameOf(bootstrap?.sites, request.siteId) : "-"}</span></>;
                } },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                { key: "quantity", title: "数量", render: (item) => `${qty(item.quantity)} ${item.unit}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => (
                    <ActionDialog id={`raw-receipt-create-${item.id}`} title="生成原料入库单" buttonLabel="收料" onOpen={() => startRawReceipt(item)}>
                      {rawReceiptFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无待收采购单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="原料入库单"
              data={rawReceipts}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "receipt", title: "入库单", render: (item) => <><b>{item.receiptNo}</b><span className="block-text muted">{item.plateNo || "-"}</span></> },
                { key: "order", title: "采购单", render: (item) => purchaseOrders.find((order) => order.id === item.purchaseOrderId)?.orderNo || `采购单 #${item.purchaseOrderId}` },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                { key: "weight", title: "毛/皮/净", render: (item) => `${qty(item.grossWeight)} / ${qty(item.tareWeight)} / ${qty(item.netWeight)} t` },
                { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => rawInspectionCandidates.some((candidate) => candidate.id === item.id) ? (
                    <ActionDialog id={`raw-inspection-create-from-receipt-${item.id}`} title="创建原材质检单" buttonLabel="质检" onOpen={() => startRawMaterialInspection(item)}>
                      {rawInspectionFormView}
                    </ActionDialog>
                  ) : <span className="muted">已建质检</span>
                }
              ]}
              emptyText="暂无原料入库单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="库存流水"
              data={inventoryFlows}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "flow", title: "流水", render: (item) => <><b>{item.flowNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
                { key: "direction", title: "方向", render: (item) => <StatusChip value={item.direction} /> },
                { key: "qty", title: "数量/结存", render: (item) => `${qty(item.quantity)} / ${qty(item.balanceQty)}` },
                { key: "source", title: "来源", render: (item) => `${item.sourceType || "-"}${item.sourceId ? ` #${item.sourceId}` : ""}` },
                { key: "remark", title: "备注", render: (item) => item.remark || "-" }
              ]}
              emptyText="暂无库存流水"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "raw-material-inspections") {
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board inventory-department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="原材质检范围">
                <span className="inventory-scope-chip strong">{inventoryScopeLabel}</span>
                <span className="inventory-scope-chip">{rawInspectionCandidates.length} 张待检入库</span>
                <span className="inventory-scope-chip">{rawInspections.length} 张质检单</span>
                <span className={`inventory-scope-chip ${rawInspections.some((item) => item.result === "failed") ? "warning" : ""}`}>{rawInspections.filter((item) => item.result === "failed").length} 张不合格</span>
              </div>
              <ActionGroup>
                <ActionDialog
                  id="raw-inspection-create-page"
                  title="创建原材质检单"
                  buttonLabel="创建质检"
                  triggerIcon={<ClipboardCheck size={13} />}
                  triggerVariant="primary"
                  disabled={!rawInspectionCandidates.length}
                  onOpen={() => rawInspectionCandidates[0] && startRawMaterialInspection(rawInspectionCandidates[0])}
                >
                  {rawInspectionFormView}
                </ActionDialog>
                <ButtonLink icon={<ReceiptText size={15} />} href="/resources/inventory/receipts">原料收料</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>待检入库</span><b>{rawInspectionCandidates.length}</b><small>入库后未建质检</small></div>
              <div><span>质检单</span><b>{rawInspections.length}</b><small>待复核 {rawInspections.filter((item) => item.status !== "completed").length}</small></div>
              <div><span>合格</span><b>{rawInspections.filter((item) => item.result === "passed").length}</b><small>已复核原材</small></div>
              <div><span>不合格</span><b>{rawInspections.filter((item) => item.result === "failed").length}</b><small>需业务处理</small></div>
              <div><span>入库单</span><b>{rawReceipts.length}</b><small>当前站点范围</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="原料入库待检"
              data={rawInspectionCandidates}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "receipt", title: "入库单", render: (item) => <><b>{item.receiptNo}</b><span className="block-text muted">{item.plateNo || "-"}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "weight", title: "净重", render: (item) => `${qty(item.netWeight)} t` },
                { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "110px",
                  render: (item) => (
                    <ActionDialog id={`raw-inspection-create-page-${item.id}`} title="创建原材质检单" buttonLabel="创建" onOpen={() => startRawMaterialInspection(item)}>
                      {rawInspectionFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无待检入库单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="原材质检单"
              data={rawInspections}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "inspection", title: "质检单", render: (item) => <><b>{item.inspectionNo}</b><span className="block-text muted">{item.receiptNo}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "metrics", title: "指标", render: (item) => `${item.moisture || "-"} / ${item.mudContent || "-"} / ${item.fineness || "-"}` },
                { key: "result", title: "结论", render: (item) => <StatusChip value={item.result} /> },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["raw_material_inspection"], item.id, item.inspectionNo, <StatusChip value={item.status} />) },
                {
                  key: "actions",
                  title: "操作",
                  width: "110px",
                  render: (item: RawMaterialInspection) => (
                    <ActionDialog id={`raw-inspection-review-page-${item.id}`} title="复核原材质检" buttonLabel={item.status === "completed" ? "复核" : "提交"} onOpen={() => startRawMaterialInspectionReview(item)}>
                      {rawInspectionReviewFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无原材质检单"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "inventory-stocktakes") {
      const pendingStocktakes = stocktakes.filter((item) => item.status !== "completed");
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board inventory-department-board">
            <SectionHeader className="inventory-command-bar">
              <div className="inventory-scope-strip" aria-label="库存盘点范围">
                <span className="inventory-scope-chip strong">{inventoryScopeLabel}</span>
                <span className="inventory-scope-chip">{stocktakes.length} 张盘点单</span>
                <span className={`inventory-scope-chip ${pendingStocktakes.length ? "warning" : ""}`}>{pendingStocktakes.length ? `${pendingStocktakes.length} 张待复核` : "盘点闭环"}</span>
              </div>
              <ActionGroup>
                <ActionDialog id="inventory-stocktake-create-page" title="提交库存盘点" buttonLabel="提交盘点" triggerIcon={<ClipboardCheck size={13} />} triggerVariant="primary" onOpen={() => resetStocktakeForm(inventoryItems[0])}>
                  {stocktakeFormView}
                </ActionDialog>
                <ButtonLink icon={<Package size={15} />} href="/resources/stock-yards">堆场管理</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </SectionHeader>
            <div className="inventory-kpi-strip">
              <div><span>库存批次</span><b>{inventoryItems.length}</b><small>{qty(inventoryBalance)} t</small></div>
              <div><span>盘点单</span><b>{stocktakes.length}</b><small>当前站点范围</small></div>
              <div><span>待复核</span><b>{pendingStocktakes.length}</b><small>需要确认差异</small></div>
              <div><span>差异合计</span><b>{qty(stocktakes.reduce((sum, item) => sum + item.diffQty, 0))}</b><small>账实差异</small></div>
              <div><span>已完成</span><b>{stocktakes.filter((item) => item.status === "completed").length}</b><small>复核闭环</small></div>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="库存盘点"
              data={stocktakes}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "stocktake", title: "盘点单", render: (item) => <><b>{item.stocktakeNo}</b><span className="block-text muted">{nameOf(bootstrap?.sites, item.siteId)}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "qty", title: "账面 / 实际 / 差异", render: (item) => `${qty(item.bookQty)} / ${qty(item.actualQty)} / ${qty(item.diffQty)} ${item.unit}` },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["inventory_stocktake"], item.id, item.stocktakeNo, <StatusChip value={item.status} />) },
                { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) },
                {
                  key: "actions",
                  title: "操作",
                  width: "110px",
                  render: (item: InventoryStocktake) => <UiButton size="sm" disabled={actionBusy !== "" || item.status === "completed"} onClick={() => handleReviewInventoryStocktake(item)}>复核</UiButton>
                }
              ]}
              emptyText="暂无库存盘点"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <>
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12 department-board inventory-department-board">
          <SectionHeader className="inventory-command-bar">
            <div className="inventory-scope-strip" aria-label="区域">
              <span className="inventory-scope-chip strong">{inventoryScopeLabel}</span>
              <span className="inventory-scope-chip">{stockYards.length} 个堆场</span>
              <span className="inventory-scope-chip">{stockYardPiles.length} 个堆位</span>
              <span className={`inventory-scope-chip ${riskCount ? "warning" : ""}`}>{riskCount ? `${riskCount} 项需处理` : "运行正常"}</span>
            </div>
            <ActionGroup>
              <UiButton icon={<Plus size={15} />} disabled={!siteOptions().length} onClick={() => openYardDialog("yard")}>新增堆场</UiButton>
              <UiButton icon={<Route size={15} />} onClick={() => setStockYardDetailDialog("flows")}>堆位流水{stockYardFlows.length ? ` ${stockYardFlows.length}` : ""}</UiButton>
              <UiButton icon={<ClipboardCheck size={15} />} onClick={() => setStockYardDetailDialog("quality")}>原材质检{rawInspectionCandidates.length ? ` ${rawInspectionCandidates.length}` : ""}</UiButton>
              <UiButton icon={<ListChecks size={15} />} onClick={() => setStockYardDetailDialog("stocktakes")}>库存盘点{stocktakes.length ? ` ${stocktakes.length}` : ""}</UiButton>
              <ButtonLink icon={<Route size={15} />} href="/resources/inventory/transfers">库存调拨详情{transferCount ? ` ${transferCount}` : ""}</ButtonLink>
              <ButtonLink icon={<Scale size={15} />} href="/fulfillment/weighbridge">过磅记录</ButtonLink>
              <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
            </ActionGroup>
          </SectionHeader>
          <div className="inventory-kpi-strip">
            <div><span>库存</span><b>{qty(inventoryBalance)} t</b><small>{inventoryItems.length} 个批次</small></div>
            <div><span>堆场</span><b>{qty(yardBalance)} t</b><small>{stockYards.length} 个堆场 / {stockYardPiles.length} 个堆位</small></div>
            <div><span>筒仓</span><b>{qty(lineBufferBalance)} t</b><small>{plantBuffers.length} 个筒仓</small></div>
            <div><span>风险</span><b>{riskCount}</b><small>库存 {lowInventoryCount} / 堆位 {lowYardCount} / 筒仓 {lowLineBufferCount}</small></div>
            <div><span>待收料</span><b>{pendingReceiptOrders}</b><small>{purchaseOrders.length} 个采购单</small></div>
          </div>
          <div className="inventory-yard-card-list">
	            {stockYardCards.map((card) => (
	              <Card className={`inventory-yard-card ${card.issuePiles.length ? "warning" : ""}`} key={card.yard.id} onContextMenu={(event) => openStockYardMenu(event, card.yard)}>
	                <div className="inventory-yard-card-head">
	                  <div>
	                    <b>{card.yard.name}</b>
	                    <span>{card.yard.code} / {card.yard.area || "未设置区域"} / {nameOf(bootstrap?.sites, card.yard.siteId)}</span>
	                  </div>
	                  <div className="inventory-yard-card-head-actions">
	                    <StatusChip value={card.issuePiles.length ? "warning" : card.yard.status} />
	                    <IconButton
	                      className="inventory-card-add-button"
	                      icon={<Plus size={15} />}
                      label={`新增${card.yard.name}堆位`}
                      variant="primary"
                      disabled={actionBusy !== "" || !bootstrap?.materials?.length}
                      onClick={() => openYardDialog("pile", card.yard)}
                    />
                  </div>
                </div>
                <div className="inventory-yard-card-metrics">
                  <div><span>库存</span><b>{qty(card.balance)} {card.unit}</b></div>
                  <div><span>堆位</span><b>{card.piles.length} 个</b></div>
                  <div><span>容量</span><b>{percent(card.capacity ? card.balance / card.capacity * 100 : 0)}</b></div>
                  <div><span>风险</span><b>{card.issuePiles.length}</b></div>
                </div>
                <div className="inventory-yard-card-body">
                  <div>
                    <span>物料</span>
                    <b>{card.materialSummary || "暂无物料"}</b>
                  </div>
                  <div>
                    <span>网关</span>
                    <b>{card.yard.gatewayDeviceNo || "未接入"}</b>
                  </div>
                  <div>
                    <span>上报</span>
                    <b>{card.yard.lastReportedAt ? shortDateTime(card.yard.lastReportedAt) : "暂无"}</b>
                  </div>
                </div>
                <div className="inventory-yard-pile-list">
		                  {card.piles.map((pile) => {
		                    const issue = isPileIssue(pile);
		                    return (
                      <div
                        className={`inventory-yard-pile-row clickable ${issue ? "issue" : ""}`}
                        key={pile.id}
                        role="button"
                        tabIndex={0}
                        aria-label={`编辑堆位 ${pile.name}`}
                        onClick={() => openYardDialog("pile-edit", card.yard, pile)}
	                        onKeyDown={(event) => {
	                          if (event.key === "Enter" || event.key === " ") {
	                            event.preventDefault();
	                            openYardDialog("pile-edit", card.yard, pile);
	                          }
	                        }}
	                        onContextMenu={(event) => openStockYardPileMenu(event, pile)}
	                      >
                        <div>
                          <b>{pile.name}</b>
                          <span>{pile.code} / {nameOf(bootstrap?.materials, pile.materialId) || "-"} / {pile.batchNo || "-"}</span>
                        </div>
                        <div className="inventory-yard-pile-balance">
                          <b>{qty(pile.currentQty)} {pile.unit}</b>
                          <span>{percent(pile.capacity ? pile.currentQty / pile.capacity * 100 : 0)} / 阈值 {qty(pile.warningQty)} {pile.unit}</span>
                        </div>
                        <div>
                          <b>{recordName(suppliers.find((supplier) => recordId(supplier) === pile.supplierId), "-")}</b>
                          <span>{pile.yardCode || "-"}</span>
                        </div>
			                        {workflowStatusFor(["stock_yard_adjustment"], pile.id, pile.code, <StatusChip value={issue ? "warning" : pile.status} />)}
	                      </div>
                    );
                  })}
                  {!card.piles.length ? <div className="inventory-through-empty large">暂无堆位</div> : null}
                </div>
              </Card>
            ))}
            {!stockYardCards.length ? <div className="inventory-through-empty large">暂无堆场</div> : null}
		          </div>
		        </Panel>
		      </SectionGrid>
        {stockYardMenu && stockYardMenuYard ? (
          <ContextMenu
            items={stockYardContextItems}
            position={{ x: stockYardMenu.x, y: stockYardMenu.y }}
            label="堆场操作"
            width={196}
            onClose={() => setStockYardMenu(null)}
          />
        ) : null}
        {stockYardPileMenu && stockYardPileMenuPile ? (
          <ContextMenu
            items={stockYardPileContextItems}
            position={{ x: stockYardPileMenu.x, y: stockYardPileMenu.y }}
            label="堆位操作"
            width={196}
            onClose={() => setStockYardPileMenu(null)}
          />
        ) : null}
        <Dialog open={stockYardDetailDialog === "flows"} title="堆位流水" size="wide" onClose={() => setStockYardDetailDialog(null)}>
          <DataTable<StockYardFlow>
            title="堆位流水"
            data={stockYardFlows}
            rowKey={(item) => item.id}
            pageSize={8}
            columns={[
              { key: "flow", title: "流水", render: (item) => <><b>{item.flowNo}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></> },
              {
                key: "yard",
                title: "堆场/堆位",
                render: (item) => {
                  const yard = stockYards.find((candidate) => candidate.id === item.yardId);
                  const pile = stockYardPiles.find((candidate) => candidate.id === item.pileId);
                  return <><b>{yard?.name || item.pileCode || "-"}</b><span className="block-text muted">{pile?.name || item.pileCode || `#${item.pileId}`}</span></>;
                }
              },
              { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) || `物料 #${item.materialId}` },
              { key: "direction", title: "方向", render: (item) => <StatusChip value={item.direction} /> },
              { key: "qty", title: "数量/结存", render: (item) => `${qty(item.quantity)} / ${qty(item.balanceQty)} ${item.unit}` },
              { key: "quality", title: "质量/含水", render: (item) => <><StatusChip value={item.qualityStatus} /><span className="block-text muted">{item.moistureRate || 0}%</span></> },
              { key: "source", title: "来源", render: (item) => `${item.sourceType || "-"}${item.sourceId ? ` #${item.sourceId}` : ""}` },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
              { key: "remark", title: "备注", render: (item) => item.remark || "-" }
            ]}
            emptyText="暂无堆位流水"
          />
        </Dialog>
        <Dialog open={stockYardDetailDialog === "quality"} title="原材质量检验" size="wide" onClose={() => setStockYardDetailDialog(null)}>
          <SectionHeader className="panel-head-compact">
            <div>
              <b>原材质量检验</b>
              <span>{rawInspections.length} 张质检单 / {rawInspectionCandidates.length} 张入库单待检</span>
            </div>
            <ActionGroup>
              <ActionDialog
                id="raw-inspection-create"
                title="创建原材质检单"
                buttonLabel="创建质检"
                triggerIcon={<ClipboardCheck size={13} />}
                disabled={!rawInspectionCandidates.length}
                onOpen={() => rawInspectionCandidates[0] && startRawMaterialInspection(rawInspectionCandidates[0])}
              >
                {rawInspectionFormView}
              </ActionDialog>
            </ActionGroup>
          </SectionHeader>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="原料入库待检"
              data={rawInspectionCandidates}
              rowKey={(item) => item.id}
              pageSize={6}
              columns={[
                { key: "receipt", title: "入库单", render: (item) => <><b>{item.receiptNo}</b><span className="block-text muted">{item.plateNo || "-"}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "weight", title: "净重", render: (item) => `${qty(item.netWeight)} t` },
                { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "110px",
                  render: (item) => (
                    <ActionDialog id={`raw-inspection-create-${item.id}`} title="创建原材质检单" buttonLabel="创建" onOpen={() => startRawMaterialInspection(item)}>
                      {rawInspectionFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无待检入库单"
            />
            <DataTable
              title="原材质检单"
              data={rawInspections}
              rowKey={(item) => item.id}
              pageSize={6}
              columns={[
                { key: "inspection", title: "质检单", render: (item) => <><b>{item.inspectionNo}</b><span className="block-text muted">{item.receiptNo}</span></> },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "supplier", title: "供应商", render: (item) => recordName(suppliers.find((supplier) => recordId(supplier) === item.supplierId), "-") },
                { key: "metrics", title: "指标", render: (item) => `${item.moisture || "-"} / ${item.mudContent || "-"} / ${item.fineness || "-"}` },
                { key: "result", title: "结论", render: (item) => <StatusChip value={item.result} /> },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["raw_material_inspection"], item.id, item.inspectionNo, <StatusChip value={item.status} />) },
                {
                  key: "actions",
                  title: "操作",
                  width: "110px",
                  render: (item: RawMaterialInspection) => (
                    <ActionDialog id={`raw-inspection-review-${item.id}`} title="复核原材质检" buttonLabel={item.status === "completed" ? "复核" : "提交"} onOpen={() => startRawMaterialInspectionReview(item)}>
                      {rawInspectionReviewFormView}
                    </ActionDialog>
                  )
                }
              ]}
              emptyText="暂无原材质检单"
            />
          </SectionGrid>
        </Dialog>
        <Dialog open={stockYardDetailDialog === "stocktakes"} title="库存盘点" size="wide" onClose={() => setStockYardDetailDialog(null)}>
          <DataTable
            title="库存盘点"
            data={stocktakes}
            rowKey={(item) => item.id}
            pageSize={8}
            headerLeftAction={(
              <ActionDialog id="inventory-stocktake-create" title="提交库存盘点" buttonLabel="提交盘点" triggerIcon={<ClipboardCheck size={13} />} onOpen={() => resetStocktakeForm(inventoryItems[0])}>
                {stocktakeFormView}
              </ActionDialog>
            )}
            columns={[
              { key: "stocktake", title: "盘点单", render: (item) => <><b>{item.stocktakeNo}</b><span className="block-text muted">{nameOf(bootstrap?.sites, item.siteId)}</span></> },
              { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
              { key: "qty", title: "账面 / 实际 / 差异", render: (item) => `${qty(item.bookQty)} / ${qty(item.actualQty)} / ${qty(item.diffQty)} ${item.unit}` },
              { key: "status", title: "状态", render: (item) => workflowStatusFor(["inventory_stocktake"], item.id, item.stocktakeNo, <StatusChip value={item.status} />) },
              { key: "created", title: "创建时间", render: (item) => shortDateTime(item.createdAt) },
              {
                key: "actions",
                title: "操作",
                width: "110px",
                render: (item: InventoryStocktake) => <UiButton size="sm" disabled={actionBusy !== "" || item.status === "completed"} onClick={() => handleReviewInventoryStocktake(item)}>复核</UiButton>
              }
            ]}
            emptyText="暂无库存盘点"
          />
        </Dialog>
      </>
    );
  }

  function renderInventoryTransfers() {
    const procurement = data.procurement;
    const transfers = list(procurement?.transfers).filter((item) => !selectedSiteId || item.fromSiteId === selectedSiteId || item.toSiteId === selectedSiteId);
    const pendingTransfers = transfers.filter((item) => item.status !== "completed").length;
    const approvedTransfers = transfers.filter((item) => item.status === "approved").length;
    const completedTransfers = transfers.filter((item) => item.status === "completed").length;
    const transferQuantity = transfers.reduce((sum, item) => sum + item.quantity, 0);
    const transferUnit = transfers.find((item) => item.unit)?.unit || "t";
    const inventoryScopeLabel = selectedSiteId ? `${nameOf(bootstrap?.sites, selectedSiteId)} 相关调拨` : "全部站点库存调拨";

    async function handleCreateInventoryTransfer(event: FormEvent<HTMLFormElement>) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      await runBusinessAction("inventory-transfer-create", "库存调拨已提交审批", () => api.createInventoryTransfer({
        fromSiteId: fieldNumber(String(form.get("fromSiteId") || "")),
        toSiteId: fieldNumber(String(form.get("toSiteId") || "")),
        materialId: fieldNumber(String(form.get("materialId") || "")),
        quantity: fieldNumber(String(form.get("quantity") || "")),
        unit: String(form.get("unit") || "t"),
        remark: String(form.get("remark") || "")
      }));
    }

    async function handleCompleteInventoryTransfer(item: InventoryTransfer) {
      await runBusinessAction(`inventory-transfer-complete-${item.id}`, "库存调拨已完成", () => api.completeInventoryTransfer(item.id));
    }

    return (
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12 inventory-transfer-detail-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>库存调拨详情</b>
              <span>{inventoryScopeLabel}</span>
            </div>
            <ActionGroup>
              <ButtonLink icon={<ArrowLeft size={15} />} href="/resources/stock-yards">返回堆场管理</ButtonLink>
              <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
            </ActionGroup>
          </SectionHeader>
          <div className="inventory-kpi-strip">
            <div><span>调拨单</span><b>{transfers.length}</b><small>{inventoryScopeLabel}</small></div>
            <div><span>待处理</span><b>{pendingTransfers}</b><small>待审批 / 待完成</small></div>
            <div><span>已批准</span><b>{approvedTransfers}</b><small>可完成调拨</small></div>
            <div><span>已完成</span><b>{completedTransfers}</b><small>完成闭环</small></div>
            <div><span>数量</span><b>{qty(transferQuantity)} {transferUnit}</b><small>当前筛选合计</small></div>
          </div>
          <div className="inventory-transfer-table-block">
            <DataTable<InventoryTransfer>
              title="调拨单"
              data={transfers}
              rowKey={(item) => item.id}
              pageSize={8}
              emptyText="暂无库存调拨"
              columnSettingsKey="inventory-transfer-detail"
              headerLeftAction={(
                <ActionDialog id="inventory-transfer-create" title="新增库存调拨">
                  <DialogForm className="compact-dialog-form" onSubmit={handleCreateInventoryTransfer}>
                    <Field label="调出站点">
                      <SelectInput name="fromSiteId" defaultValue={selectedSiteId || siteOptions()[0]?.id || ""}>
                        {siteOptions().map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
                      </SelectInput>
                    </Field>
                    <Field label="调入站点">
                      <SelectInput name="toSiteId" defaultValue={selectedSiteId ? siteOptions().find((site) => site.id !== selectedSiteId)?.id || "" : siteOptions()[1]?.id || ""}>
                        {siteOptions().map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
                      </SelectInput>
                    </Field>
                    <Field label="物料">
                      <SelectInput name="materialId" defaultValue={bootstrap?.materials?.[0]?.id || ""}>
                        {list(bootstrap?.materials).map((material) => <option key={material.id} value={material.id}>{material.name}</option>)}
                      </SelectInput>
                    </Field>
                    <Field label="数量"><TextInput name="quantity" type="number" min="0" step="1" defaultValue="20" /></Field>
                    <Field label="单位"><TextInput name="unit" defaultValue="t" /></Field>
                    <Field label="备注"><TextInput name="remark" defaultValue="库存调拨" /></Field>
                    <FormActions spanAll>
                      <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || siteOptions().length < 2 || !bootstrap?.materials?.length}>提交调拨</UiButton>
                    </FormActions>
                  </DialogForm>
                </ActionDialog>
              )}
              headerAction={<span className="muted">{transfers.length} 张调拨单</span>}
              columns={[
                { key: "transferNo", title: "调拨单", render: (item) => <b>{item.transferNo}</b> },
                { key: "route", title: "站点", render: (item) => `${nameOf(bootstrap?.sites, item.fromSiteId)} -> ${nameOf(bootstrap?.sites, item.toSiteId)}` },
                { key: "material", title: "物料", render: (item) => nameOf(bootstrap?.materials, item.materialId) },
                { key: "quantity", title: "数量", render: (item) => `${qty(item.quantity)} ${item.unit}` },
                { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["inventory_transfer"], item.id, item.transferNo, <StatusChip value={item.status} />) },
                { key: "actions", title: "操作", render: (item) => (
                  <ActionDialog id={`inventory-transfer-${item.id}`} title="库存调拨流程" buttonLabel="流程">
                    <div className="finance-hidden-actions">
                      <div className="finance-action-block">
                        <b>{item.transferNo}</b>
                        <span>{nameOf(bootstrap?.sites, item.fromSiteId)} {"->"} {nameOf(bootstrap?.sites, item.toSiteId)}</span>
                        <span>{nameOf(bootstrap?.materials, item.materialId)} / {qty(item.quantity)} {item.unit}</span>
                      </div>
                      {workflowTimelineBlock(["inventory_transfer"], item.id, item.transferNo, "当前调拨单暂无工作流实例")}
                      {item.status === "approved" ? <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => handleCompleteInventoryTransfer(item)}>完成调拨</UiButton> : null}
                    </div>
                  </ActionDialog>
                ) }
              ]}
            />
          </div>
        </Panel>
      </SectionGrid>
    );
  }

  function renderFinance() {
    const finance = data.finance;
    const receivables = list(finance?.receivables);
    const payables = list(finance?.payables);
    const paymentPlans = list(finance?.paymentPlans).filter((item) => item.status !== "settled");
    const invoices: SalesInvoice[] = list(finance?.invoices);
    const redLetters = list(finance?.redLetterInfos);
    const invoiceTypeOptions = dictionaryOptions("invoice_type");
    const paymentMethodOptions = dictionaryOptions("payment_method");
    const allCollectionTasks = list(finance?.collectionTasks);
    const collectionTasks = openCollectionTasks();
    const collectionTemplates = list(finance?.collectionTemplates);
    const collectionDispatches = [...list(finance?.collectionDispatches)].sort((a, b) => b.id - a.id);
    const supplierStatements: SupplierStatement[] = list(finance?.supplierStatements);
    const payments: Payment[] = list(finance?.payments);
    const suppliers = supplierOptions();
    const statements = list(finance?.statements).length ? list(finance?.statements) : data.statements;
    const ledgerRows = [
      ...receivables.map((item) => ({ id: `receivable-${item.id}`, kind: "receivable" as const, item })),
      ...payables.map((item) => ({ id: `payable-${item.id}`, kind: "payable" as const, item }))
    ];

    const receivableOpenBalance = (item: Receivable) => Math.max(0, item.amount - item.receivedAmount);
    const statementLabel = (id: number) => statements.find((item) => item.id === id)?.statementNo || `对账单 #${id}`;
    const supplierName = (id: number) => recordName(suppliers.find((supplier) => recordId(supplier) === id), `供应商 ${id}`);
    const supplierStatementLabel = (id: number) => supplierStatements.find((item) => item.id === id)?.statementNo || `供应商对账 #${id}`;
    const receiptList = (id: number) => list(finance?.receipts).filter((item) => item.receivableId === id);
    const planList = (id: number) => paymentPlans.filter((item) => item.receivableId === id);
    const invoiceList = (item: Receivable) => invoices.filter((invoice) => invoice.statementId === item.statementId || invoice.id === item.invoiceId);
    const collectionList = (id: number) => collectionTasks.filter((item) => item.receivableId === id);
    const paymentList = (id: number) => payments.filter((item) => item.payableId === id);
    const supplierStatementFor = (item: Payable) => supplierStatements.find((statement) => statement.id === item.supplierStatementId);
    const collectionTaskLabel = (taskId: number) => allCollectionTasks.find((item) => item.id === taskId)?.taskNo || `催收任务 #${taskId}`;
    const collectionTemplateLabel = (templateId: number) => collectionTemplates.find((item) => item.id === templateId)?.name || `模板 #${templateId}`;
    const firstTemplate = collectionTemplates.find((item) => item.enabled) || collectionTemplates[0];
    const salesContracts = activeContracts();
    const openReceivableAmount = receivables.reduce((sum, item) => sum + receivableOpenBalance(item), 0);
    const openPayableAmount = payables.reduce((sum, item) => sum + payableBalance(item), 0);
    const pendingCustomerStatements = statements.filter((item) => !isCustomerStatementClosed(item.status)).length;
    const pendingSupplierStatements = supplierStatements.filter((item) => item.status !== "approved").length;
    const payableWithoutStatement = payables.filter((item) => !item.supplierStatementId).length;
    const invoiceableStatementRows = statements.filter((item) => item.status === "confirmed" && !invoices.some((invoice) => invoice.statementId === item.id && invoice.invoiceType !== "red"));
    const uninvoicedStatements = invoiceableStatementRows.length;
    const sortedInvoices = [...invoices].sort((a, b) => b.id - a.id);
    const sortedRedLetters = [...redLetters].sort((a, b) => b.id - a.id);
    const taxPendingInvoices = invoices.filter((item) => item.taxStatus !== "submitted" && item.taxStatus !== "success" && item.taxStatus !== "synced");
    const redLetterPendingCount = redLetters.filter((item) => item.status !== "approved" && item.status !== "used").length;

    async function handleInlineReceipt(event: FormEvent<HTMLFormElement>, item: Receivable) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const amount = fieldNumber(String(form.get("amount") || ""), receivableOpenBalance(item));
      await runBusinessAction(`finance-receipt-${item.id}`, "收款已登记", () => api.createReceipt({ receivableId: item.id, amount, method: "bank" }));
    }

    async function handleInlinePaymentPlan(event: FormEvent<HTMLFormElement>, item: Receivable) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      await runBusinessAction(`finance-plan-${item.id}`, "付款计划已创建", () => api.createPaymentPlan({
        receivableId: item.id,
        amount: fieldNumber(String(form.get("amount") || ""), receivableOpenBalance(item)),
        dueDate: String(form.get("dueDate") || today),
        method: "bank"
      }));
    }

    async function handleInlineInvoice(event: FormEvent<HTMLFormElement>, item: Receivable) {
      event.preventDefault();
      const category = String(new FormData(event.currentTarget).get("category") || "blue_vat_special");
      await runBusinessAction(`finance-invoice-create-${item.statementId}`, "发票已创建", () => api.createInvoice(item.statementId, category));
    }

    async function handleStatementInvoice(event: FormEvent<HTMLFormElement>, item: Statement) {
      event.preventDefault();
      const category = String(new FormData(event.currentTarget).get("category") || "blue_vat_special");
      await runBusinessAction(`finance-invoice-create-${item.id}`, "发票已创建", () => api.createInvoice(item.id, category));
    }

    async function handleInlineRedLetter(event: FormEvent<HTMLFormElement>, invoice: SalesInvoice) {
      event.preventDefault();
      const reason = String(new FormData(event.currentTarget).get("reason") || "");
      await runBusinessAction(`finance-red-letter-create-${invoice.id}`, "红字信息表已申请", () => api.createRedLetterInfo(invoice.id, reason));
    }

    async function handleInlineRedOffset(event: FormEvent<HTMLFormElement>, invoice: SalesInvoice) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const reason = String(form.get("reason") || "");
      const redLetterId = fieldNumber(String(form.get("redLetterId") || ""));
      await runBusinessAction(`finance-red-offset-${invoice.id}`, "红字发票已生成", () => api.redOffsetInvoice(invoice.id, reason, redLetterId));
    }

    async function handleInlineSendCollection(event: FormEvent<HTMLFormElement>, taskId: number) {
      event.preventDefault();
      const templateId = fieldNumber(String(new FormData(event.currentTarget).get("templateId") || ""), firstTemplate?.id || 0);
      await runBusinessAction(`finance-collection-send-${taskId}`, "催收消息已发送", () => api.sendCollectionTask(taskId, templateId, ""));
    }

    async function handleInlineCloseCollection(event: FormEvent<HTMLFormElement>, taskId: number) {
      event.preventDefault();
      const remark = String(new FormData(event.currentTarget).get("remark") || "");
      await runBusinessAction(`finance-collection-close-${taskId}`, "催收任务已操作", () => api.handleCollectionTask(taskId, remark));
    }

    async function handleInlineSupplierStatement(event: FormEvent<HTMLFormElement>, item: Payable) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      await runBusinessAction(`finance-supplier-statement-create-${item.supplierId}`, "供应商对账单已生成", () => api.createSupplierStatement({
        supplierId: item.supplierId,
        period: String(form.get("period") || today),
        amount: fieldNumber(String(form.get("amount") || ""), payableBalance(item))
      }));
    }

    async function handleInlineSupplierPayment(event: FormEvent<HTMLFormElement>, item: Payable) {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      await runBusinessAction(`finance-payment-create-${item.id}`, "供应商付款已登记", () => api.createPayment({
        payableId: item.id,
        amount: fieldNumber(String(form.get("amount") || ""), payableBalance(item)),
        method: String(form.get("method") || "bank")
      }));
    }

    if (section === "finance-receivables") {
      const overdueReceivables = receivables.filter((item) => receivableOpenBalance(item) > 0 && item.dueDate && item.dueDate < today);
      const receipts = list(finance?.receipts).sort((a, b) => b.id - a.id);
      const openPlanAmount = paymentPlans.reduce((sum, item) => sum + item.amount, 0);
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board finance-department-board">
            <div className="department-board-head">
              <div>
                <h3>应收收款</h3>
                <p>应收单、收款登记、付款计划和计划结清</p>
              </div>
              <ActionGroup>
                <ButtonLink icon={<ReceiptText size={15} />} href="/finance/invoices">税票管理</ButtonLink>
                <ButtonLink icon={<Bell size={15} />} href="/finance/collections">催收管理</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </div>
            <div className="finance-focus-grid">
              <Card className="finance-focus-card">
                <span>应收余额</span>
                <b>{money(openReceivableAmount)}</b>
                <small>{receivables.length} 笔应收</small>
              </Card>
              <Card className="finance-focus-card">
                <span>逾期应收</span>
                <b>{overdueReceivables.length}</b>
                <small>{money(overdueReceivables.reduce((sum, item) => sum + receivableOpenBalance(item), 0))}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>付款计划</span>
                <b>{paymentPlans.length}</b>
                <small>{money(openPlanAmount)}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>收款记录</span>
                <b>{receipts.length}</b>
                <small>{money(receipts.reduce((sum, item) => sum + item.amount, 0))}</small>
              </Card>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<Receivable>
              title="应收单"
              data={receivables}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "bill", title: "应收单", render: (item) => <><b>{item.billNo}</b><span className="block-text muted">{statementLabel(item.statementId)}</span></> },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "balance", title: "已收/余额", render: (item) => `${money(item.receivedAmount)} / ${money(receivableOpenBalance(item))}` },
                { key: "dueDate", title: "到期日", render: (item) => item.dueDate || "-" },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["receivable", "finance_receivable"], item.id, item.billNo, <StatusChip value={item.status} />) },
                {
                  key: "actions",
                  title: "操作",
                  width: "160px",
                  render: (item) => {
                    const relatedReceipts = receiptList(item.id);
                    const relatedPlans = planList(item.id);
                    const relatedInvoices = invoiceList(item);
                    const workflow = workflowItemsFor(["receivable", "finance_receivable"], item.id, item.billNo);
                    const task = approvalFor(["receivable", "finance_receivable"], item.id, item.billNo);
                    return (
                      <ActionDialog id={`finance-receivable-page-${item.id}`} title="应收操作">
                        <div className="finance-hidden-actions">
                          <div className="finance-action-block">
                            <b>{item.billNo}</b>
                            <span>{nameOf(bootstrap?.customers, item.customerId)} / 余额 {money(receivableOpenBalance(item))}</span>
                          </div>
                          <InlineForm onSubmit={(event) => handleInlineReceipt(event, item)}>
                            <Field label="收款金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={receivableOpenBalance(item)} /></Field>
                            <UiButton variant="primary" type="submit" icon={<CreditCard size={13} />} disabled={actionBusy !== "" || receivableOpenBalance(item) <= 0}>登记收款</UiButton>
                          </InlineForm>
                          <InlineForm onSubmit={(event) => handleInlinePaymentPlan(event, item)}>
                            <Field label="计划金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={receivableOpenBalance(item)} /></Field>
                            <HeroDateField label="计划日期" name="dueDate" defaultValue={item.dueDate || today} />
                            <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== "" || receivableOpenBalance(item) <= 0}>创建计划</UiButton>
                          </InlineForm>
                          {workflowTimelineBlock(["receivable", "finance_receivable"], item.id, item.billNo, "当前应收单暂无工作流实例")}
                          {!workflow.instances.length ? approvalActionBlock(task) : null}
                          <div className="finance-action-block">
                            <b>收款</b>
                            {relatedReceipts.map((receipt) => <span key={receipt.id}>{receipt.receiptNo} / {money(receipt.amount)} / {receipt.receivedAt}</span>)}
                            {!relatedReceipts.length ? <span>暂无收款</span> : null}
                          </div>
                          <div className="finance-action-block">
                            <b>付款计划</b>
                            {relatedPlans.map((plan) => (
                              <span key={plan.id}>
                                {plan.planNo} / {money(plan.amount)} / {plan.dueDate} / {plan.status}
                                {plan.status !== "settled" ? (
                                  <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`payment-plan-${plan.id}`, "付款计划已结清", () => api.settlePaymentPlan(plan.id))}>结清</UiButton>
                                ) : null}
                              </span>
                            ))}
                            {!relatedPlans.length ? <span>暂无计划</span> : null}
                          </div>
                          <div className="finance-action-block">
                            <b>发票</b>
                            {relatedInvoices.map((invoice) => <span key={invoice.id}>{invoice.invoiceNo} / {money(invoice.amount)} / {invoice.taxStatus}</span>)}
                            {!relatedInvoices.length ? <span>暂无发票</span> : null}
                          </div>
                        </div>
                      </ActionDialog>
                    );
                  }
                }
              ]}
              emptyText="暂无应收单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="收款记录"
              data={receipts}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "receipt", title: "收款单", render: (item) => <><b>{item.receiptNo}</b><span className="block-text muted">{item.receivedAt}</span></> },
                { key: "receivable", title: "应收单", render: (item) => receivables.find((receivable) => receivable.id === item.receivableId)?.billNo || `应收 #${item.receivableId}` },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "method", title: "方式", render: (item) => item.method || "-" },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> }
              ]}
              emptyText="暂无收款记录"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "finance-invoices") {
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board finance-department-board">
            <div className="department-board-head">
              <div>
                <h3>税票管理</h3>
                <p>对账单开票、税控提交、红字信息表审批和红冲</p>
              </div>
              <ActionGroup>
                <ButtonLink icon={<ReceiptText size={15} />} href="/finance/statements">客户对账</ButtonLink>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </div>
            <div className="finance-focus-grid">
              <Card className="finance-focus-card">
                <span>待开票</span>
                <b>{invoiceableStatementRows.length}</b>
                <small>已确认客户对账单</small>
              </Card>
              <Card className="finance-focus-card">
                <span>发票</span>
                <b>{invoices.length}</b>
                <small>待税控 {taxPendingInvoices.length}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>红字信息表</span>
                <b>{redLetters.length}</b>
                <small>待处理 {redLetterPendingCount}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>税额</span>
                <b>{money(invoices.reduce((sum, item) => sum + item.taxAmount, 0))}</b>
                <small>发票金额 {money(invoices.reduce((sum, item) => sum + item.amount, 0))}</small>
              </Card>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<Statement>
              title="待开票对账单"
              data={invoiceableStatementRows}
              rowKey={(item) => item.id}
              pageSize={8}
              onRefresh={refreshData}
              columns={[
                { key: "statement", title: "对账单", render: (item) => <><b>{item.statementNo}</b><span className="block-text muted">{shortDateTime(item.confirmedAt)}</span></> },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
                { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) || "-" },
                { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                {
                  key: "actions",
                  title: "操作",
                  width: "240px",
                  render: (item) => (
                    <InlineForm onSubmit={(event) => handleStatementInvoice(event, item)}>
                      <Field label="发票类型">
                        <SelectInput name="category" defaultValue={invoiceTypeOptions[0]?.code || "blue_vat_special"}>
                          {invoiceTypeOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                        </SelectInput>
                      </Field>
                      <UiButton type="submit" icon={<ReceiptText size={13} />} disabled={actionBusy !== ""}>开票</UiButton>
                    </InlineForm>
                  )
                }
              ]}
              emptyText="暂无待开票对账单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<SalesInvoice>
              title="发票"
              data={sortedInvoices}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "invoice", title: "发票", render: (item) => <><b>{item.invoiceNo}</b><span className="block-text muted">{item.taxControlNo || item.redLetterInfoNo || "-"}</span></> },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
                { key: "statement", title: "对账单", render: (item) => statementLabel(item.statementId) },
                { key: "amount", title: "金额/税额", render: (item) => `${money(item.amount)} / ${money(item.taxAmount)}` },
                { key: "type", title: "类型", render: (item) => dictionaryValueLabel("invoice_type", item.invoiceCategory || item.invoiceType, item.invoiceCategory || item.invoiceType) },
                { key: "status", title: "状态", render: (item) => <><StatusChip value={item.taxStatus} /><StatusChip value={item.status} /></> },
                { key: "issued", title: "开票时间", render: (item) => item.issuedAt ? shortDateTime(item.issuedAt) : "-" },
                {
                  key: "actions",
                  title: "操作",
                  width: "160px",
                  render: (item) => {
                    const invoiceRedLetters = redLetters.filter((red) => red.originalInvoiceId === item.id);
                    const approvedRedLetters = invoiceRedLetters.filter((red) => red.status === "approved");
                    return (
                      <ActionDialog id={`finance-invoice-page-${item.id}`} title="发票操作">
                        <div className="finance-hidden-actions">
                          <div className="finance-action-block">
                            <b>{item.invoiceNo}</b>
                            <span>{nameOf(bootstrap?.customers, item.customerId)} / {money(item.amount)} / {item.taxStatus}</span>
                          </div>
                          <ActionGroup className="compact-actions">
                            <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-invoice-submit-${item.id}`, "发票已提交税控", () => api.submitTaxInvoice(item.id))}>提交税控</UiButton>
                            <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-invoice-download-${item.id}`, "发票文件已打开下载", () => downloadInvoiceFile(item.id))}>下载</UiButton>
                          </ActionGroup>
                          {item.invoiceType !== "red" ? (
                            <InlineForm onSubmit={(event) => handleInlineRedLetter(event, item)}>
                              <Field label="红字原因"><TextInput name="reason" required /></Field>
                              <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== ""}>申请红字</UiButton>
                            </InlineForm>
                          ) : null}
                          {invoiceRedLetters.map((red) => {
                            const redWorkflow = workflowItemsFor(["red_letter_info", "redLetterInfo"], red.id, red.infoNo);
                            return (
                              <div className="finance-nested-block" key={red.id}>
                                <span>{red.infoNo} / {workflowStatusFor(["red_letter_info", "redLetterInfo"], red.id, red.infoNo, <StatusChip value={red.status} />)}</span>
                                {workflowTimelineBlock(["red_letter_info", "redLetterInfo"], red.id, red.infoNo, "当前红字信息表暂无工作流实例")}
                                {red.status !== "approved" && !redWorkflow.instances.length ? <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-red-letter-approve-${red.id}`, "红字信息表已审批", () => api.approveRedLetterInfo(red.id))}>审批</UiButton> : null}
                              </div>
                            );
                          })}
                          {approvedRedLetters.length ? (
                            <InlineForm onSubmit={(event) => handleInlineRedOffset(event, item)}>
                              <Field label="红字信息表">
                                <SelectInput name="redLetterId" defaultValue={approvedRedLetters[0]?.id || ""}>
                                  {approvedRedLetters.map((red) => <option key={red.id} value={red.id}>{red.infoNo}</option>)}
                                </SelectInput>
                              </Field>
                              <Field label="红字原因"><TextInput name="reason" required /></Field>
                              <UiButton type="submit" icon={<ReceiptText size={13} />} disabled={actionBusy !== ""}>红冲</UiButton>
                            </InlineForm>
                          ) : null}
                        </div>
                      </ActionDialog>
                    );
                  }
                }
              ]}
              emptyText="暂无发票"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="红字信息表"
              data={sortedRedLetters}
              rowKey={(item) => item.id}
              pageSize={8}
              onRefresh={refreshData}
              columns={[
                { key: "info", title: "信息表", render: (item) => <><b>{item.infoNo}</b><span className="block-text muted">{item.originalInvoiceNo}</span></> },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
                { key: "amount", title: "金额/税额", render: (item) => `${money(item.amount)} / ${money(item.taxAmount)}` },
                { key: "reason", title: "原因", render: (item) => item.reason || "-" },
                { key: "requested", title: "申请时间", render: (item) => item.requestedAt ? shortDateTime(item.requestedAt) : "-" },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["red_letter_info", "redLetterInfo"], item.id, item.infoNo, <StatusChip value={item.status} />) },
                {
                  key: "actions",
                  title: "操作",
                  width: "120px",
                  render: (item) => {
                    const redWorkflow = workflowItemsFor(["red_letter_info", "redLetterInfo"], item.id, item.infoNo);
                    return item.status !== "approved" && !redWorkflow.instances.length ? (
                      <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-red-letter-approve-${item.id}`, "红字信息表已审批", () => api.approveRedLetterInfo(item.id))}>审批</UiButton>
                    ) : <span className="muted">-</span>;
                  }
                }
              ]}
              emptyText="暂无红字信息表"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "finance-collections") {
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board finance-department-board">
            <div className="department-board-head">
              <div>
                <h3>催收管理</h3>
                <p>应收逾期、催收模板、发送流水和关闭处理</p>
              </div>
              <ActionGroup>
                <UiButton icon={<Plus size={14} />} disabled={actionBusy !== ""} onClick={handleGenerateCollections}>生成催收</UiButton>
                <ActionDialog
                  id="finance-collection-template-create-page"
                  title="新增催收模板"
                  buttonLabel="新增模板"
                  triggerIcon={<FileSignature size={13} />}
                  triggerVariant="primary"
                  onOpen={() => setCollectionTemplateForm({
                    name: "",
                    level: "warning",
                    channel: collectionTemplates[0]?.channel || "sms",
                    content: "",
                    enabled: "true"
                  })}
                >
                  <DialogForm onSubmit={handleCreateCollectionTemplate}>
                    <Field label="名称"><TextInput value={collectionTemplateForm.name} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, name: event.target.value })} required /></Field>
                    <Field label="等级">
                      <SelectInput value={collectionTemplateForm.level} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, level: event.target.value })}>
                        <option value="notice">notice</option>
                        <option value="warning">warning</option>
                        <option value="urgent">urgent</option>
                      </SelectInput>
                    </Field>
                    <Field label="渠道">
                      <SelectInput value={collectionTemplateForm.channel} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, channel: event.target.value })}>
                        <option value="sms">sms</option>
                        <option value="wechat">wechat</option>
                        <option value="email">email</option>
                        <option value="phone">phone</option>
                      </SelectInput>
                    </Field>
                    <Field label="启用">
                      <SelectInput value={collectionTemplateForm.enabled} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, enabled: event.target.value })}>
                        <option value="true">启用</option>
                        <option value="false">停用</option>
                      </SelectInput>
                    </Field>
                    <Field label="内容" spanAll><TextAreaInput value={collectionTemplateForm.content} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, content: event.target.value })} required /></Field>
                    <FormActions spanAll>
                      <UiButton variant="primary" type="submit" icon={<FileSignature size={14} />} disabled={actionBusy !== "" || !collectionTemplateForm.name.trim() || !collectionTemplateForm.content.trim()}>保存模板</UiButton>
                    </FormActions>
                  </DialogForm>
                </ActionDialog>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </div>
            <div className="finance-focus-grid">
              <Card className="finance-focus-card">
                <span>待催收任务</span>
                <b>{collectionTasks.length}</b>
                <small>全部任务 {allCollectionTasks.length}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>逾期应收</span>
                <b>{money(openReceivableAmount)}</b>
                <small>{receivables.length} 笔应收</small>
              </Card>
              <Card className="finance-focus-card">
                <span>模板</span>
                <b>{collectionTemplates.length}</b>
                <small>启用 {collectionTemplates.filter((item) => item.enabled).length}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>发送流水</span>
                <b>{collectionDispatches.length}</b>
                <small>失败 {collectionDispatches.filter((item) => item.status === "failed").length}</small>
              </Card>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="催收任务"
              data={allCollectionTasks}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "task", title: "任务", render: (item) => <><b>{item.taskNo}</b><span className="block-text muted">{nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}`}</span></> },
                { key: "receivable", title: "应收", render: (item) => <><span>{statementLabel(receivables.find((receivable) => receivable.id === item.receivableId)?.statementId || 0)}</span><span className="block-text muted">逾期 {item.overdueDays || 0} 天</span></> },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "level", title: "等级", render: (item) => <StatusChip value={item.level} /> },
                { key: "send", title: "发送", render: (item) => `${item.sendCount || 0} 次 / ${shortDateTime(item.lastSentAt)}` },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "actions", title: "操作", width: "240px", render: (item) => (
                  <ActionDialog id={`finance-collection-task-${item.id}`} title="催收任务操作">
                    <div className="finance-hidden-actions">
                      <div className="finance-action-block">
                        <b>{item.taskNo}</b>
                        <span>{nameOf(bootstrap?.customers, item.customerId)} / {money(item.amount)} / 逾期 {item.overdueDays || 0} 天</span>
                      </div>
                      <InlineForm onSubmit={(event) => handleInlineSendCollection(event, item.id)}>
                        <Field label="模板">
                          <SelectInput name="templateId" defaultValue={firstTemplate?.id || ""}>
                            {collectionTemplates.map((template) => <option key={template.id} value={template.id}>{template.name} / {template.channel}</option>)}
                          </SelectInput>
                        </Field>
                        <UiButton type="submit" icon={<FileSignature size={13} />} disabled={actionBusy !== "" || !collectionTemplates.length}>发送</UiButton>
                      </InlineForm>
                      <InlineForm onSubmit={(event) => handleInlineCloseCollection(event, item.id)}>
                        <Field label="处理备注"><TextInput name="remark" required /></Field>
                        <UiButton type="submit" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== "" || item.status === "closed"}>关闭</UiButton>
                      </InlineForm>
                    </div>
                  </ActionDialog>
                ) }
              ]}
              emptyText="暂无催收任务"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<CollectionDispatch>
              title="催收发送流水"
              data={collectionDispatches}
              rowKey={(item) => item.id}
              pageSize={10}
              columns={[
                { key: "dispatch", title: "发送流水", render: (item) => <><b>{item.dispatchNo}</b><span className="block-text muted">{shortDateTime(item.sentAt)}</span></> },
                { key: "task", title: "催收任务", render: (item) => collectionTaskLabel(item.taskId) },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
                { key: "template", title: "模板/渠道", render: (item) => <><b>{collectionTemplateLabel(item.templateId)}</b><span className="block-text muted">{item.channel || "-"}</span></> },
                { key: "target", title: "目标", render: (item) => item.target || "-" },
                { key: "provider", title: "供应商回执", render: (item) => <><span>{item.providerMessageId || item.providerRequestId || "-"}</span><span className="block-text muted">{item.callbackAt ? shortDateTime(item.callbackAt) : "未回执"}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "error", title: "错误", render: (item) => item.error || "-" }
              ]}
              emptyText="暂无催收发送流水"
            />
          </Panel>
        </SectionGrid>
      );
    }

    if (section === "finance-suppliers") {
      return (
        <SectionGrid className="finance-list-page">
          <Panel as="div" className="span-12 department-board finance-department-board">
            <div className="department-board-head">
              <div>
                <h3>供应商对账</h3>
                <p>应付、供应商对账单、付款登记和审批状态</p>
              </div>
              <ActionGroup>
                <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
              </ActionGroup>
            </div>
            <div className="finance-focus-grid">
              <Card className="finance-focus-card">
                <span>供应商对账单</span>
                <b>{supplierStatements.length}</b>
                <small>待确认 {pendingSupplierStatements}</small>
              </Card>
              <Card className="finance-focus-card">
                <span>应付余额</span>
                <b>{money(openPayableAmount)}</b>
                <small>{payables.length} 笔应付</small>
              </Card>
              <Card className="finance-focus-card">
                <span>未对账</span>
                <b>{payableWithoutStatement}</b>
                <small>可直接生成供应商对账单</small>
              </Card>
              <Card className="finance-focus-card">
                <span>已付款</span>
                <b>{payments.length}</b>
                <small>{money(payments.reduce((sum, item) => sum + item.amount, 0))}</small>
              </Card>
            </div>
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable
              title="供应商对账单"
              data={supplierStatements}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "statement", title: "对账单", render: (item) => <><b>{item.statementNo}</b><span className="block-text muted">{supplierName(item.supplierId)}</span></> },
                { key: "period", title: "期间", render: (item) => item.period || "-" },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["supplier_statement", "supplierStatement"], item.id, item.statementNo, <StatusChip value={item.status} />) },
                { key: "approvedAt", title: "确认时间", render: (item) => shortDateTime(item.approvedAt) },
                { key: "actions", title: "操作", width: "140px", render: (item) => {
                  const statementWorkflow = workflowItemsFor(["supplier_statement", "supplierStatement"], item.id, item.statementNo);
                  return item.status !== "approved" && !statementWorkflow.instances.length ? (
                    <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-supplier-statement-approve-${item.id}`, "供应商对账单已确认", () => api.approveSupplierStatement(item.id))}>确认对账</UiButton>
                  ) : <span className="muted">-</span>;
                } }
              ]}
              emptyText="暂无供应商对账单"
            />
          </Panel>
          <Panel as="div" className="span-12">
            <DataTable<Payable>
              title="供应商应付"
              data={payables}
              rowKey={(item) => item.id}
              pageSize={12}
              onRefresh={refreshData}
              columns={[
                { key: "bill", title: "应付单", render: (item) => <><b>{item.billNo}</b><span className="block-text muted">{supplierName(item.supplierId)}</span></> },
                { key: "source", title: "来源", render: (item) => supplierStatementLabel(item.supplierStatementId) },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "balance", title: "已付/余额", render: (item) => `${money(item.paidAmount)} / ${money(payableBalance(item))}` },
                { key: "dueDate", title: "到期日", render: (item) => item.dueDate || "-" },
                { key: "status", title: "状态", render: (item) => workflowStatusFor(["payable", "finance_payable"], item.id, item.billNo, <StatusChip value={item.status} />) },
                { key: "actions", title: "操作", width: "160px", render: (item) => {
                  const statement = supplierStatementFor(item);
                  const relatedPayments = paymentList(item.id);
                  return (
                    <ActionDialog id={`finance-supplier-payable-${item.id}`} title="供应商应付操作">
                      <div className="finance-hidden-actions">
                        <div className="finance-action-block">
                          <b>{item.billNo}</b>
                          <span>{supplierName(item.supplierId)} / {money(payableBalance(item))}</span>
                        </div>
                        <div className="finance-action-block">
                          <b>供应商对账</b>
                          {statement ? (
                            <span>{statement.statementNo} / {money(statement.amount)} / {statement.status}</span>
                          ) : (
                            <InlineForm onSubmit={(event) => handleInlineSupplierStatement(event, item)}>
                              <Field label="期间"><TextInput name="period" defaultValue={today} /></Field>
                              <Field label="金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={payableBalance(item)} /></Field>
                              <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== ""}>生成对账</UiButton>
                            </InlineForm>
                          )}
                        </div>
                        <InlineForm onSubmit={(event) => handleInlineSupplierPayment(event, item)}>
                          <Field label="付款金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={payableBalance(item)} /></Field>
                          <Field label="方式">
                            <SelectInput name="method" defaultValue={paymentMethodOptions[0]?.code || "bank"}>
                              {paymentMethodOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                            </SelectInput>
                          </Field>
                          <UiButton type="submit" icon={<CreditCard size={13} />} disabled={actionBusy !== "" || payableBalance(item) <= 0}>登记付款</UiButton>
                        </InlineForm>
                        <div className="finance-action-block">
                          <b>付款</b>
                          {relatedPayments.map((payment) => <span key={payment.id}>{payment.paymentNo} / {money(payment.amount)} / {payment.method} / {payment.paidAt}</span>)}
                          {!relatedPayments.length ? <span>暂无付款</span> : null}
                        </div>
                      </div>
                    </ActionDialog>
                  );
                } }
              ]}
              emptyText="暂无供应商应付"
            />
          </Panel>
        </SectionGrid>
      );
    }

    return (
      <SectionGrid className="finance-list-page">
        <Panel as="div" className="span-12 department-board finance-department-board">
          <div className="department-board-head">
            <div>
              <h3>财务对账看板</h3>
            </div>
            <ActionGroup>
              <ButtonLink icon={<ReceiptText size={15} />} href="/finance/statements">客户对账</ButtonLink>
              <UiButton icon={<RefreshCw size={15} />} onClick={refreshData}>刷新</UiButton>
            </ActionGroup>
          </div>
          <div className="finance-focus-grid">
            <Card className="finance-focus-card">
              <span>客户对账</span>
              <b>{statements.length} 张对账单</b>
              <small>{salesContracts.length} 份有效合同 / 待确认 {pendingCustomerStatements} / 未开票 {uninvoicedStatements}</small>
            </Card>
            <Card className="finance-focus-card">
              <span>应收余额</span>
              <b>{money(openReceivableAmount)}</b>
              <small>{receivables.length} 笔应收 / 催收任务 {collectionTasks.length}</small>
            </Card>
            <Card className="finance-focus-card">
              <span>供应商对账</span>
              <b>{supplierStatements.length} 张对账单</b>
              <small>待确认 {pendingSupplierStatements} / 未对账 {payableWithoutStatement}</small>
            </Card>
            <Card className="finance-focus-card">
              <span>应付余额</span>
              <b>{money(openPayableAmount)}</b>
              <small>{payables.length} 笔应付 / 已登记付款 {payments.length}</small>
            </Card>
          </div>
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable
            data={ledgerRows}
            rowKey={(row) => row.id}
            pageSize={12}
            emptyText="暂无财务往来"
            rowContextMenu={buildDataTableRowContextMenu<(typeof ledgerRows)[number]>({
              actions: [
                {
                  key: "focus-counterparty",
                  label: "只看该往来方",
                  onSelect: (row, helpers) => helpers.searchText(row.kind === "receivable" ? nameOf(bootstrap?.customers, row.item.customerId) : supplierName(row.item.supplierId))
                },
                {
                  key: "focus-source",
                  label: "只看来源单据",
                  onSelect: (row, helpers) => helpers.searchText(row.kind === "receivable" ? statementLabel(row.item.statementId) : supplierStatementLabel(row.item.supplierStatementId))
                }
              ],
              copyFields: [
                { key: "bill", label: "往来单号", value: (row) => row.item.billNo },
                { key: "counterparty", label: "往来方", value: (row) => row.kind === "receivable" ? nameOf(bootstrap?.customers, row.item.customerId) : supplierName(row.item.supplierId) },
                { key: "source", label: "来源单据", value: (row) => row.kind === "receivable" ? statementLabel(row.item.statementId) : supplierStatementLabel(row.item.supplierStatementId) },
                { key: "balance", label: "余额", value: (row) => row.kind === "receivable" ? money(receivableOpenBalance(row.item)) : money(payableBalance(row.item)) }
              ]
            })}
            headerLeftAction={
              <ActionGroup>
                <UiButton icon={<Plus size={14} />} disabled={actionBusy !== ""} onClick={handleGenerateCollections}>生成催收</UiButton>
                <ActionDialog
                  id="finance-collection-template-create"
                  title="新增催收模板"
                  buttonLabel="新增模板"
                  triggerIcon={<FileSignature size={13} />}
                  onOpen={() => setCollectionTemplateForm({
                    name: "",
                    level: "warning",
                    channel: collectionTemplates[0]?.channel || "sms",
                    content: "",
                    enabled: "true"
                  })}
                >
                  <DialogForm onSubmit={handleCreateCollectionTemplate}>
                    <Field label="名称"><TextInput value={collectionTemplateForm.name} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, name: event.target.value })} required /></Field>
                    <Field label="等级">
                      <SelectInput value={collectionTemplateForm.level} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, level: event.target.value })}>
                        <option value="notice">notice</option>
                        <option value="warning">warning</option>
                        <option value="urgent">urgent</option>
                      </SelectInput>
                    </Field>
                    <Field label="渠道">
                      <SelectInput value={collectionTemplateForm.channel} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, channel: event.target.value })}>
                        <option value="sms">sms</option>
                        <option value="wechat">wechat</option>
                        <option value="email">email</option>
                        <option value="phone">phone</option>
                      </SelectInput>
                    </Field>
                    <Field label="启用">
                      <SelectInput value={collectionTemplateForm.enabled} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, enabled: event.target.value })}>
                        <option value="true">启用</option>
                        <option value="false">停用</option>
                      </SelectInput>
                    </Field>
                    <Field label="内容" spanAll><TextAreaInput value={collectionTemplateForm.content} onChange={(event) => setCollectionTemplateForm({ ...collectionTemplateForm, content: event.target.value })} required /></Field>
                    <FormActions spanAll>
                      <UiButton variant="primary" type="submit" icon={<FileSignature size={14} />} disabled={actionBusy !== "" || !collectionTemplateForm.name.trim() || !collectionTemplateForm.content.trim()}>保存模板</UiButton>
                    </FormActions>
                  </DialogForm>
                </ActionDialog>
              </ActionGroup>
            }
            headerAction={
              <div className="finance-header-actions">
                <span className="muted">应收 {money(receivableBalance(finance))} / 应付 {money(payables.reduce((sum, item) => sum + payableBalance(item), 0))} / 催收 {collectionTasks.length}</span>
              </div>
            }
            columns={[
              { key: "kind", title: "类型", render: (row) => row.kind === "receivable" ? "应收" : "应付" },
              { key: "bill", title: "单据", render: (row) => <b>{row.item.billNo}</b> },
              { key: "counterparty", title: "往来方", render: (row) => row.kind === "receivable" ? nameOf(bootstrap?.customers, row.item.customerId) : supplierName(row.item.supplierId) },
              { key: "source", title: "来源", render: (row) => row.kind === "receivable" ? statementLabel(row.item.statementId) : supplierStatementLabel(row.item.supplierStatementId) },
              { key: "amount", title: "金额", render: (row) => money(row.item.amount) },
              { key: "settled", title: "已结/余额", render: (row) => row.kind === "receivable" ? `${money(row.item.receivedAmount)} / ${money(receivableOpenBalance(row.item))}` : `${money(row.item.paidAmount)} / ${money(payableBalance(row.item))}` },
              { key: "dueDate", title: "到期日", render: (row) => row.item.dueDate },
	              {
	                key: "status",
	                title: "状态",
	                render: (row) => {
	                  const resources = row.kind === "receivable" ? ["receivable", "finance_receivable"] : ["payable", "finance_payable"];
	                  return workflowStatusFor(resources, row.item.id, row.item.billNo, approvalStatus(approvalFor(resources, row.item.id, row.item.billNo), <StatusChip value={row.item.status} />));
	                }
	              },
              { key: "actions", title: "操作", render: (row) => {
                if (row.kind === "receivable") {
                  const item = row.item;
                  const task = approvalFor(["receivable", "finance_receivable"], item.id, item.billNo);
	                  const relatedReceipts = receiptList(item.id);
	                  const relatedPlans = planList(item.id);
	                  const relatedInvoices = invoiceList(item);
	                  const relatedCollections = collectionList(item.id);
	                  const workflow = workflowItemsFor(["receivable", "finance_receivable"], item.id, item.billNo);
	                  return (
                    <ActionDialog id={`finance-receivable-${item.id}`} title="应收操作">
                      <div className="finance-hidden-actions">
                        <InlineForm onSubmit={(event) => handleInlineReceipt(event, item)}>
                          <Field label="收款金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={receivableOpenBalance(item)} /></Field>
                          <UiButton variant="primary" type="submit" icon={<CreditCard size={13} />} disabled={actionBusy !== "" || receivableOpenBalance(item) <= 0}>登记收款</UiButton>
                        </InlineForm>
                        <InlineForm onSubmit={(event) => handleInlinePaymentPlan(event, item)}>
                          <Field label="计划金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={receivableOpenBalance(item)} /></Field>
                          <HeroDateField label="计划日期" name="dueDate" defaultValue={item.dueDate || today} />
                          <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== "" || receivableOpenBalance(item) <= 0}>创建计划</UiButton>
                        </InlineForm>
                        {item.statementId ? (
                          <InlineForm onSubmit={(event) => handleInlineInvoice(event, item)}>
                            <Field label="发票类型">
                              <SelectInput name="category" defaultValue={invoiceTypeOptions[0]?.code || "blue_vat_special"}>
                                {invoiceTypeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                              </SelectInput>
                            </Field>
                            <UiButton type="submit" icon={<ReceiptText size={13} />} disabled={actionBusy !== ""}>开票</UiButton>
                          </InlineForm>
                        ) : null}
	                        {workflowTimelineBlock(["receivable", "finance_receivable"], item.id, item.billNo, "当前应收单暂无工作流实例")}
	                        {!workflow.instances.length ? approvalActionBlock(task) : null}
                        <div className="finance-action-block">
                          <b>收款</b>
                          {relatedReceipts.map((receipt) => <span key={receipt.id}>{receipt.receiptNo} / {money(receipt.amount)} / {receipt.receivedAt}</span>)}
                          {!relatedReceipts.length ? <span>暂无收款</span> : null}
                        </div>
                        <div className="finance-action-block">
                          <b>付款计划</b>
                          {relatedPlans.map((plan) => (
                            <span key={plan.id}>
                              {plan.planNo} / {money(plan.amount)} / {plan.dueDate} / {plan.status}
                              {plan.status !== "settled" ? (
                                <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`payment-plan-${plan.id}`, "付款计划已结清", () => api.settlePaymentPlan(plan.id))}>结清</UiButton>
                              ) : null}
                            </span>
                          ))}
                          {!relatedPlans.length ? <span>暂无计划</span> : null}
                        </div>
                        <div className="finance-action-block">
                          <b>发票</b>
                          {relatedInvoices.map((invoice) => {
                            const invoiceRedLetters = redLetters.filter((red) => red.originalInvoiceId === invoice.id);
                            const approvedRedLetters = invoiceRedLetters.filter((red) => red.status === "approved");
                            return (
                              <div className="finance-nested-block" key={invoice.id}>
                                <span>{invoice.invoiceNo} / {money(invoice.amount)} / {invoice.taxStatus}</span>
                                <ActionGroup className="compact-actions">
                                  <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-invoice-submit-${invoice.id}`, "发票已提交税控", () => api.submitTaxInvoice(invoice.id))}>提交税控</UiButton>
                                  <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-invoice-download-${invoice.id}`, "发票文件已打开下载", () => downloadInvoiceFile(invoice.id))}>下载</UiButton>
                                </ActionGroup>
                                <InlineForm onSubmit={(event) => handleInlineRedLetter(event, invoice)}>
                                  <Field label="红字原因"><TextInput name="reason" required /></Field>
                                  <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== ""}>申请红字</UiButton>
                                </InlineForm>
	                                {invoiceRedLetters.map((red) => {
	                                  const redWorkflow = workflowItemsFor(["red_letter_info", "redLetterInfo"], red.id, red.infoNo);
	                                  return (
	                                    <div className="finance-nested-block" key={red.id}>
	                                      <span>{red.infoNo} / {workflowStatusFor(["red_letter_info", "redLetterInfo"], red.id, red.infoNo, <StatusChip value={red.status} />)}</span>
	                                      {workflowTimelineBlock(["red_letter_info", "redLetterInfo"], red.id, red.infoNo, "当前红字信息表暂无工作流实例")}
	                                      {red.status !== "approved" && !redWorkflow.instances.length ? <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-red-letter-approve-${red.id}`, "红字信息表已审批", () => api.approveRedLetterInfo(red.id))}>审批</UiButton> : null}
	                                    </div>
	                                  );
	                                })}
                                {approvedRedLetters.length ? (
                                  <InlineForm onSubmit={(event) => handleInlineRedOffset(event, invoice)}>
                                    <Field label="红字信息表">
                                      <SelectInput name="redLetterId" defaultValue={approvedRedLetters[0]?.id || ""}>
                                        {approvedRedLetters.map((red) => <option key={red.id} value={red.id}>{red.infoNo}</option>)}
                                      </SelectInput>
                                    </Field>
                                    <Field label="红字原因"><TextInput name="reason" required /></Field>
                                    <UiButton type="submit" icon={<ReceiptText size={13} />} disabled={actionBusy !== ""}>红冲</UiButton>
                                  </InlineForm>
                                ) : null}
                              </div>
                            );
                          })}
                          {!relatedInvoices.length ? <span>暂无发票</span> : null}
                        </div>
                        <div className="finance-action-block">
                          <b>催收</b>
                          {relatedCollections.map((task) => (
                            <div className="finance-nested-block" key={task.id}>
                              <span>{task.taskNo} / 逾期 {task.overdueDays} 天 / 已发 {task.sendCount}</span>
                              <InlineForm onSubmit={(event) => handleInlineSendCollection(event, task.id)}>
                                <Field label="模板">
                                  <SelectInput name="templateId" defaultValue={firstTemplate?.id || ""}>
                                    {collectionTemplates.map((template) => <option key={template.id} value={template.id}>{template.name} / {template.channel}</option>)}
                                  </SelectInput>
                                </Field>
                                <UiButton type="submit" icon={<FileSignature size={13} />} disabled={actionBusy !== "" || !collectionTemplates.length}>发送</UiButton>
                              </InlineForm>
                              <InlineForm onSubmit={(event) => handleInlineCloseCollection(event, task.id)}>
                                <Field label="备注"><TextInput name="remark" required /></Field>
                                <UiButton type="submit" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""}>关闭</UiButton>
                              </InlineForm>
                            </div>
                          ))}
                          {!relatedCollections.length ? <span>暂无催收</span> : null}
                        </div>
                      </div>
                    </ActionDialog>
                  );
                }

	                const item = row.item;
		                const task = approvalFor(["payable", "finance_payable"], item.id, item.billNo);
		                const statement = supplierStatementFor(item);
		                const relatedPayments = paymentList(item.id);
		                const workflow = workflowItemsFor(["payable", "finance_payable"], item.id, item.billNo);
		                const statementWorkflow = statement ? workflowItemsFor(["supplier_statement", "supplierStatement"], statement.id, statement.statementNo) : null;
		                return (
	                  <ActionDialog id={`finance-payable-${item.id}`} title="应付操作">
	                    <div className="finance-hidden-actions">
	                      {workflowTimelineBlock(["payable", "finance_payable"], item.id, item.billNo, "当前应付单暂无工作流实例")}
	                      {!workflow.instances.length ? approvalActionBlock(task) : null}
                      <div className="finance-action-block">
                        <b>供应商对账</b>
	                        {statement ? (
	                          <>
	                            <span>{statement.statementNo} / {money(statement.amount)} / {workflowStatusFor(["supplier_statement", "supplierStatement"], statement.id, statement.statementNo, <StatusChip value={statement.status} />)}</span>
	                            {workflowTimelineBlock(["supplier_statement", "supplierStatement"], statement.id, statement.statementNo, "当前供应商对账单暂无工作流实例")}
	                            {statement.status !== "approved" && !statementWorkflow?.instances.length ? (
	                              <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`finance-supplier-statement-approve-${statement.id}`, "供应商对账单已确认", () => api.approveSupplierStatement(statement.id))}>确认对账</UiButton>
	                            ) : null}
	                          </>
                        ) : (
                          <InlineForm onSubmit={(event) => handleInlineSupplierStatement(event, item)}>
                            <Field label="期间"><TextInput name="period" defaultValue={today} /></Field>
                            <Field label="金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={payableBalance(item)} /></Field>
                            <UiButton type="submit" icon={<Plus size={13} />} disabled={actionBusy !== ""}>生成对账</UiButton>
                          </InlineForm>
                        )}
                      </div>
                      <InlineForm onSubmit={(event) => handleInlineSupplierPayment(event, item)}>
                        <Field label="付款金额"><TextInput name="amount" type="number" min="0" step="1" defaultValue={payableBalance(item)} /></Field>
                        <Field label="方式">
                          <SelectInput name="method" defaultValue={paymentMethodOptions[0]?.code || "bank"}>
                            {paymentMethodOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                          </SelectInput>
                        </Field>
                        <UiButton type="submit" icon={<CreditCard size={13} />} disabled={actionBusy !== "" || payableBalance(item) <= 0}>登记付款</UiButton>
                      </InlineForm>
                      <div className="finance-action-block">
                        <b>付款</b>
                        {relatedPayments.map((payment) => <span key={payment.id}>{payment.paymentNo} / {money(payment.amount)} / {payment.method} / {payment.paidAt}</span>)}
                        {!relatedPayments.length ? <span>暂无付款</span> : null}
                      </div>
                    </div>
                  </ActionDialog>
                );
              } }
            ]}
          />
        </Panel>
        <Panel as="div" className="span-12">
          <DataTable<CollectionDispatch>
            title="催收发送流水"
            data={collectionDispatches}
            rowKey={(item) => item.id}
            pageSize={8}
            columns={[
              { key: "dispatch", title: "发送流水", render: (item) => <><b>{item.dispatchNo}</b><span className="block-text muted">{shortDateTime(item.sentAt)}</span></> },
              { key: "task", title: "催收任务", render: (item) => collectionTaskLabel(item.taskId) },
              { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) || `客户 #${item.customerId}` },
              { key: "template", title: "模板/渠道", render: (item) => <><b>{collectionTemplateLabel(item.templateId)}</b><span className="block-text muted">{item.channel || "-"}</span></> },
              { key: "target", title: "目标", render: (item) => item.target || "-" },
              { key: "provider", title: "供应商回执", render: (item) => <><span>{item.providerMessageId || item.providerRequestId || "-"}</span><span className="block-text muted">{item.callbackAt ? shortDateTime(item.callbackAt) : "未回执"}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "error", title: "错误", render: (item) => item.error || "-" },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
            ]}
            emptyText="暂无催收发送流水"
          />
        </Panel>
      </SectionGrid>
    );
  }

  function renderReports() {
    const reports = data.reports;
    const inventoryWarnings = list(reports?.inventoryWarnings).filter((item) => matchesCurrentSite(item.siteId));
    const projectProfit = list(reports?.projectProfit);
    const customerAging = list(reports?.customerAging);
    const agingBuckets = list(reports?.agingBuckets);
    const energy = list(reports?.energy).filter((item) => matchesCurrentSite(item.siteId));
    const customerStatements = list(reports?.customerStatements);
    const suppliers = supplierOptions();
    const operating = reports?.operating;
    const quality = reports?.quality;
    const reportScope = selectedSiteId ? nameOf(bootstrap?.sites, selectedSiteId) : "全部站点";
    const hasAnyReportData = Boolean(
      operating ||
      quality ||
      inventoryWarnings.length ||
      projectProfit.length ||
      customerAging.length ||
      agingBuckets.length ||
      energy.length ||
      customerStatements.length
    );

    type ReportMetricRow = {
      id: string;
      item: string;
      value: string;
      unit: string;
      note: string;
    };

    function reportTotal<T>(items: T[], selector: (item: T) => number) {
      return items.reduce((sum, item) => sum + selector(item), 0);
    }

    function supplierName(supplierId: number) {
      if (!supplierId) return "-";
      return recordName(suppliers.find((supplier) => recordId(supplier) === supplierId), `供应商 ${supplierId}`);
    }

    const operatingRows: ReportMetricRow[] = operating ? [
      { id: "order-count", item: "订单数", value: qty(operating.orderCount), unit: "单", note: "订单数量" },
      { id: "planned-qty", item: "计划吨位", value: qty(operating.plannedQty), unit: "t", note: "" },
      { id: "signed-qty", item: "签收吨位", value: qty(operating.signedQty), unit: "t", note: "已签收吨位" },
      { id: "revenue", item: "营业收入", value: money(operating.revenue), unit: "元", note: "主营业务收入" },
      { id: "material-cost", item: "材料成本", value: money(operating.materialCost), unit: "元", note: "" },
      { id: "transport-cost", item: "运输成本", value: money(operating.transportCost), unit: "元", note: "运输费用" },
      { id: "total-cost", item: "总成本", value: money(operating.totalCost), unit: "元", note: "" },
      { id: "gross-profit", item: "毛利", value: money(operating.grossProfit), unit: "元", note: `毛利率 ${percent(operating.grossMargin)}` }
    ] : [];
    const qualityRows: ReportMetricRow[] = quality ? [
      { id: "quality-inspections", item: "质量检验", value: qty(quality.inspections), unit: "次", note: `合格 ${qty(quality.passed)} / 待处理 ${qty(quality.pending)} / 不合格 ${qty(quality.failed)}` },
      { id: "quality-pass-rate", item: "合格率", value: percent(quality.passRate), unit: "%", note: "" },
      { id: "quality-samples", item: "样品检测", value: qty(quality.samples), unit: "次", note: `合格 ${qty(quality.samplePassed)} / 待处理 ${qty(quality.samplePending)} / 不合格 ${qty(quality.sampleFailed)}` },
      { id: "lab-tests", item: "试验次数", value: qty(quality.laboratoryTests), unit: "次", note: "" },
      { id: "quality-exceptions", item: "质量异常", value: qty(quality.openExceptionCount), unit: "项", note: "待处理质量异常" }
    ] : [];
    const projectRevenue = reportTotal(projectProfit, (item) => item.revenue);
    const projectCost = reportTotal(projectProfit, (item) => item.cost);
    const projectGrossProfit = reportTotal(projectProfit, (item) => item.profit);
    const customerOverdue = reportTotal(customerAging, (item) => item.overdueTotal);
    const statementAmount = reportTotal(customerStatements, (item) => item.totalAmount);
    const statementQty = reportTotal(customerStatements, (item) => item.totalQty);
    const producedQty = reportTotal(energy, (item) => item.producedQty);
    const estimatedPowerKwh = reportTotal(energy, (item) => item.estimatedPowerKwh);

    return (
      <LayoutRegion as="section" className="reports-page">
        <Panel as="div" className="report-sheet">
          <div className="report-sheet-head">
            <div>
              <p className="report-kicker">管理报表</p>
              <h3>经营总览</h3>
            </div>
            <UiButton icon={<RefreshCw size={15} />} onClick={() => void refreshData()} disabled={loading}>刷新</UiButton>
          </div>
          <KeyValueTable
            className="report-form-table"
            rows={[
              [
                { label: "口径", value: "经营" },
                { label: "范围", value: reportScope },
                { label: "日期", value: today }
              ],
              [
                { label: "模块", value: "订单 / 生产 / 签收 / 财务 / 质量" },
                { label: "状态", value: loading ? "加载中" : hasAnyReportData ? "已生成" : "暂无数据" },
                { label: "单位", value: "人民币 / t" }
              ]
            ]}
          />
          <KeyValueTable
            className="report-summary-table"
            rows={[
              [
                { label: "营业收入", value: money(operating?.revenue) },
                { label: "总成本", value: money(operating?.totalCost) },
                { label: "毛利率", value: percent(operating?.grossMargin) }
              ],
              [
                { label: "应收余额", value: money(operating?.receivableBalance) },
                { label: "逾期应收", value: money(operating?.overdueReceivable || customerOverdue) },
                { label: "库存预警", value: `${qty(operating?.inventoryWarningCount || inventoryWarnings.length)} 项` }
              ],
              [
                { label: "计划方量", value: `${qty(operating?.plannedQty)} m3` },
                { label: "签收方量", value: `${qty(operating?.signedQty)} m3` },
                { label: "质量异常", value: `${qty(operating?.openQualityIssues || quality?.openExceptionCount)} 项` }
              ]
            ]}
          />
        </Panel>

        <SectionGrid as="div" className="report-section-grid">
          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="经营指标"
              data={operatingRows}
              rowKey={(item) => item.id}
              showPagination={false}
              emptyText={loading ? "加载中..." : "暂无经营指标"}
              columns={[
                { key: "item", title: "指标", render: (item) => <b>{item.item}</b> },
                { key: "value", title: "数值", render: (item) => item.value },
                { key: "unit", title: "单位", render: (item) => item.unit }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="项目毛利"
              data={projectProfit}
              rowKey={(item) => item.id}
              showPagination={false}
              headerLeftAction={<span className="report-section-total">收入 {money(projectRevenue)} / 成本 {money(projectCost)} / 毛利 {money(projectGrossProfit)}</span>}
              emptyText={loading ? "加载中..." : "暂无项目毛利数据"}
              columns={[
                { key: "period", title: "期间", render: (item) => item.period },
                { key: "project", title: "项目", render: (item) => <b>{nameOf(bootstrap?.projects, item.projectId)}</b> },
                { key: "revenue", title: "收入", render: (item) => money(item.revenue) },
                { key: "cost", title: "成本", render: (item) => money(item.cost) },
                { key: "profit", title: "毛利", width: "140px", align: "right", render: (item) => money(item.profit) },
                { key: "margin", title: "毛利率", width: "100px", align: "right", render: (item) => percent(item.margin) }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="客户账龄"
              data={customerAging}
              rowKey={(item) => item.customerId}
              showPagination={false}
              headerLeftAction={<span className="report-section-total">逾期合计 {money(customerOverdue)}</span>}
              emptyText={loading ? "加载中..." : "暂无账龄数据"}
              columns={[
                { key: "customer", title: "客户", render: (item) => <b>{item.customerName}</b> },
                { key: "current", title: "未到期", render: (item) => money(item.current) },
                { key: "overdue1", title: "1-30天", width: "120px", align: "right", render: (item) => money(item.overdue1To30) },
                { key: "overdue31", title: "31-60天", width: "120px", align: "right", render: (item) => money(item.overdue31To60) },
                { key: "overdue61", title: "61-90天", width: "120px", align: "right", render: (item) => money(item.overdue61To90) },
                { key: "overdue90", title: "90天以上", render: (item) => money(item.overdueOver90) },
                { key: "total", title: "合计", render: (item) => money(item.total) },
                { key: "risk", title: "风险", render: (item) => <StatusChip value={item.overdueTotal > 0 ? "warning" : "active"} /> }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-6 report-section">
            <DataTable
              title="账龄分段"
              data={agingBuckets}
              rowKey={(item) => item.bucket}
              showPagination={false}
              emptyText={loading ? "加载中..." : "暂无账龄分段"}
              columns={[
                { key: "label", title: "分段", render: (item) => <b>{item.label}</b> },
                { key: "count", title: "数量", render: (item) => qty(item.count) },
                { key: "amount", title: "金额", render: (item) => money(item.amount) },
                { key: "overdueAmount", title: "逾期金额", render: (item) => money(item.overdueAmount) }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-6 report-section">
            <DataTable
              title="质量指标"
              data={qualityRows}
              rowKey={(item) => item.id}
              showPagination={false}
              emptyText={loading ? "加载中..." : "暂无质量指标"}
              columns={[
                { key: "item", title: "指标", render: (item) => <b>{item.item}</b> },
                { key: "value", title: "数值", render: (item) => item.value },
                { key: "unit", title: "单位", render: (item) => item.unit }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="库存预警"
              data={inventoryWarnings}
              rowKey={(item) => item.id}
              showPagination={false}
              emptyText={loading ? "加载中..." : "暂无库存预警"}
              columns={[
                { key: "site", title: "站点", render: (item) => nameOf(bootstrap?.sites, item.siteId) },
                { key: "material", title: "物料", render: (item) => <b>{nameOf(bootstrap?.materials, item.materialId)}</b> },
                { key: "location", title: "位置", render: (item) => `${item.warehouse || "-"} / ${item.silo || "-"}` },
                { key: "batch", title: "批次", render: (item) => item.batchNo || "-" },
                { key: "supplier", title: "供应商", render: (item) => supplierName(item.supplierId) },
                { key: "quantity", title: "数量", render: (item) => `${qty(item.quantity)} ${item.unit}` },
                { key: "quality", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
                { key: "available", title: "可用", render: (item) => <StatusChip value={item.availableStatus} /> },
                { key: "updatedAt", title: "更新时间", render: (item) => shortDateTime(item.updatedAt) }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="客户对账"
              data={customerStatements}
              rowKey={(item) => item.id}
              showPagination={false}
              headerLeftAction={<span className="report-section-total">方量 {qty(statementQty)} m3 / 金额 {money(statementAmount)}</span>}
              emptyText={loading ? "加载中..." : "暂无客户对账数据"}
              columns={[
                { key: "statementNo", title: "对账单", render: (item) => <b>{item.statementNo}</b> },
                { key: "customer", title: "客户", render: (item) => nameOf(bootstrap?.customers, item.customerId) },
                { key: "project", title: "项目", render: (item) => nameOf(bootstrap?.projects, item.projectId) },
                { key: "period", title: "期间", render: (item) => item.period },
                { key: "qty", title: "方量", render: (item) => qty(item.totalQty) },
                { key: "amount", title: "金额", render: (item) => money(item.totalAmount) },
                { key: "status", title: "状态", width: "100px", align: "center", render: (item) => <StatusChip value={item.status} /> },
                { key: "confirmed", title: "确认信息", width: "170px", render: (item) => item.confirmedBy ? `${item.confirmedBy} / ${item.confirmedAt || "-"}` : "-" }
              ]}
            />
          </Panel>

          <Panel as="div" className="span-12 report-section">
            <DataTable
              title="生产能耗"
              data={energy}
              rowKey={(item) => item.siteId}
              showPagination={false}
              headerLeftAction={<span className="report-section-total">产量 {qty(producedQty)} m3 / 估算电耗 {qty(estimatedPowerKwh)} kWh</span>}
              emptyText={loading ? "加载中..." : "暂无生产能耗数据"}
              columns={[
                { key: "site", title: "站点", render: (item) => <b>{nameOf(bootstrap?.sites, item.siteId)}</b> },
                { key: "producedQty", title: "产量", render: (item) => qty(item.producedQty) },
                { key: "batchCount", title: "批次", render: (item) => qty(item.batchCount) },
                { key: "materialUsageQty", title: "材料用量", render: (item) => qty(item.materialUsageQty) },
                { key: "materialCost", title: "材料成本", render: (item) => money(item.materialCost) },
                { key: "unitMaterialCost", title: "单方材料", render: (item) => money(item.unitMaterialCost) },
                { key: "estimatedPowerKwh", title: "估算电耗", render: (item) => `${qty(item.estimatedPowerKwh)} kWh` },
                { key: "unitPowerKwh", title: "单方电耗", render: (item) => `${qty(item.unitPowerKwh)} kWh` }
              ]}
            />
          </Panel>
        </SectionGrid>
      </LayoutRegion>
    );
  }

  function renderApprovalCenter() {
    const currentRoleCode = bootstrap?.user.roleCode || "";
    const pendingApprovals = data.approvals.filter((item) => item.status === "pending" || item.status === "pending_approval");
    const processingApprovals = openApprovals.filter((item) => item.status !== "pending" && item.status !== "pending_approval");
    const myRoleApprovals = openApprovals.filter((item) => item.currentRole === currentRoleCode);
    const completedApprovals = data.approvals.filter((item) => item.status === "approved" || item.status === "rejected");
    const overdueApprovals = openApprovals.filter((item) => {
      const createdAt = item.createdAt ? new Date(item.createdAt).getTime() : 0;
      return createdAt > 0 && Date.now() - createdAt > 24 * 60 * 60 * 1000;
    });
    const approvalRows = [...openApprovals].sort((a, b) => {
      const roleWeightA = a.currentRole === currentRoleCode ? 0 : 1;
      const roleWeightB = b.currentRole === currentRoleCode ? 0 : 1;
      return roleWeightA - roleWeightB || (b.createdAt || "").localeCompare(a.createdAt || "");
    });

    return (
      <Panel className="approval-center-view">
        <MetricList compact className="approval-summary-grid">
          <div>
            <span>待处理</span>
            <b>{openApprovals.length}</b>
          </div>
          <div>
            <span>当前角色</span>
            <b>{myRoleApprovals.length}</b>
          </div>
          <div>
            <span>处理中</span>
            <b>{processingApprovals.length}</b>
          </div>
          <div>
            <span>超 24 小时</span>
            <b>{overdueApprovals.length}</b>
          </div>
          <div>
            <span>已完成</span>
            <b>{completedApprovals.length}</b>
          </div>
        </MetricList>
        <DataTable
          data={approvalRows}
          rowKey={(item) => item.id}
          pageSize={12}
          onRefresh={refreshData}
          emptyText={loading ? "加载中..." : "暂无待处理审批"}
          searchPlaceholder="搜索审批 / 单号 / 申请人"
          searchText={(item) => [item.title, item.taskNo, item.resourceNo, item.resource, item.applicant, item.flowName, roleName(item.currentRole), item.reason].filter(Boolean).join(" ")}
          rowContextMenu={buildDataTableRowContextMenu<ApprovalTask>({
            actions: [
              {
                key: "focus-role",
                label: "只看当前节点",
                onSelect: (item, helpers) => helpers.searchText(roleName(item.currentRole))
              },
              {
                key: "focus-resource",
                label: "按业务对象筛选",
                onSelect: (item, helpers) => helpers.searchText(item.resourceNo)
              }
            ],
            copyFields: [
              { key: "taskNo", label: "审批编号", value: (item) => item.taskNo },
              { key: "title", label: "审批标题", value: (item) => item.title },
              { key: "resource", label: "业务对象", value: (item) => `${item.resourceNo} / ${item.resource}` },
              { key: "node", label: "当前节点", value: (item) => `${item.currentStep}. ${roleName(item.currentRole)}` },
              { key: "status", label: "状态", value: (item) => item.status }
            ]
          })}
          columns={[
            {
              key: "title",
              title: "审批事项",
              render: (item) => (
                <span className="approval-task-title">
                  <b>{item.title}</b>
                  <span className="block-text muted">{item.taskNo} / {item.flowName}</span>
                </span>
              )
            },
            {
              key: "resource",
              title: "业务对象",
              render: (item) => <><b>{item.resourceNo}</b><span className="block-text muted">{item.resource}</span></>
            },
            {
              key: "role",
              title: "当前节点",
              render: (item) => <><b>{roleName(item.currentRole)}</b><span className="block-text muted">第 {item.currentStep} 步</span></>
            },
            {
              key: "applicant",
              title: "申请人",
              render: (item) => <><b>{item.applicant || "-"}</b><span className="block-text muted">{shortDateTime(item.createdAt)}</span></>
            },
            {
              key: "status",
              title: "状态",
              render: (item) => <StatusChip value={item.status} />
            },
            {
              key: "actions",
              title: "处理",
              width: "120px",
              render: (item) => (
                <ActionDialog id={`approval-action-${item.id}`} title="审批处理" buttonLabel="处理" triggerIcon={<CheckCircle2 size={13} />}>
                  <div className="finance-action-block">
                    <b>{item.title}</b>
                    <span>{item.resourceNo} / {item.resource}</span>
                    <span>{item.flowName} / 第 {item.currentStep} 步 / {roleName(item.currentRole)}</span>
                    <span>申请人：{item.applicant || "-"}</span>
                    <span>提交时间：{shortDateTime(item.createdAt)}</span>
                    {item.reason ? <span>原因：{item.reason}</span> : null}
                  </div>
                  <Field label="审批意见">
                    <TextInput value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
                  </Field>
                  <ActionGroup>
                    <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-approve-${item.id}`, "审批已通过", () => api.actApproval(item.id, "approve", approvalComment))}>通过</UiButton>
                    <UiButton disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-reject-${item.id}`, "审批已驳回", () => api.actApproval(item.id, "reject", approvalComment))}>驳回</UiButton>
                  </ActionGroup>
                </ActionDialog>
              )
            }
          ]}
        />
      </Panel>
    );
  }

  function renderLicenseManagement() {
    const system = data.system;
    const verification = system?.licenseVerified;
    const activeLicense = verification?.license;
    const licensePackages = [...list(system?.licensePackages)].sort((a, b) => b.id - a.id);
    const licenseIssues = list(system?.licenseIssues);
    const revocations = list(system?.licenseRevocations);
    const licenseIssueFormView = (
      <SystemForm onSubmit={handleIssueLicense}>
        <Field label="授权编号"><TextInput value={licenseIssueForm.licenseId} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, licenseId: event.target.value })} /></Field>
        <Field label="客户名称"><TextInput value={licenseIssueForm.customerName} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, customerName: event.target.value })} /></Field>
        <Field label="水印"><TextInput value={licenseIssueForm.watermark} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, watermark: event.target.value })} /></Field>
        <Field label="到期日"><TextInput type="date" value={licenseIssueForm.expiresAt} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, expiresAt: event.target.value })} /></Field>
        <Field label="版本"><TextInput value={licenseIssueForm.edition} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, edition: event.target.value })} /></Field>
        <Field label="站点额度"><TextInput type="number" min="1" value={licenseIssueForm.maxSites} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, maxSites: event.target.value })} /></Field>
        <Field label="车辆额度"><TextInput type="number" min="1" value={licenseIssueForm.maxVehicles} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, maxVehicles: event.target.value })} /></Field>
        <Field label="签发方"><TextInput value={licenseIssueForm.issuer} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, issuer: event.target.value })} /></Field>
        <Field label="模块" spanAll><TextAreaInput value={licenseIssueForm.modules} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, modules: event.target.value })} /></Field>
        <Field label="签发私钥" spanAll><TextAreaInput value={licenseIssueForm.privateKey} onChange={(event) => setLicenseIssueForm({ ...licenseIssueForm, privateKey: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<FileSignature size={14} />} disabled={actionBusy !== ""}>签发授权包</UiButton>
        </FormActions>
      </SystemForm>
    );
    const licenseRenewFormView = (
      <SystemForm onSubmit={handleRenewLicensePackage}>
        <Field label="新授权编号"><TextInput value={licenseRenewForm.licenseId} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, licenseId: event.target.value })} /></Field>
        <Field label="新到期日"><TextInput type="date" value={licenseRenewForm.expiresAt} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, expiresAt: event.target.value })} /></Field>
        <Field label="版本"><TextInput value={licenseRenewForm.edition} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, edition: event.target.value })} /></Field>
        <Field label="站点额度"><TextInput type="number" min="1" value={licenseRenewForm.maxSites} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, maxSites: event.target.value })} /></Field>
        <Field label="车辆额度"><TextInput type="number" min="1" value={licenseRenewForm.maxVehicles} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, maxVehicles: event.target.value })} /></Field>
        <Field label="签发方"><TextInput value={licenseRenewForm.issuer} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, issuer: event.target.value })} /></Field>
        <Field label="模块" spanAll><TextAreaInput value={licenseRenewForm.modules} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, modules: event.target.value })} /></Field>
        <Field label="签发私钥" spanAll><TextAreaInput value={licenseRenewForm.privateKey} onChange={(event) => setLicenseRenewForm({ ...licenseRenewForm, privateKey: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<RefreshCw size={14} />} disabled={actionBusy !== ""}>续期授权包</UiButton>
        </FormActions>
      </SystemForm>
    );
    const licenseRevokeFormView = (
      <SystemForm onSubmit={handleRevokeLicense}>
        <Field label="授权编号"><TextInput value={licenseRevokeForm.licenseId} onChange={(event) => setLicenseRevokeForm({ ...licenseRevokeForm, licenseId: event.target.value })} /></Field>
        <Field label="吊销原因" spanAll><TextAreaInput value={licenseRevokeForm.reason} onChange={(event) => setLicenseRevokeForm({ ...licenseRevokeForm, reason: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" disabled={actionBusy !== "" || !licenseRevokeForm.licenseId.trim()}>吊销授权</UiButton>
        </FormActions>
      </SystemForm>
    );
    return (
      <Panel className="system-management-view">
        <SectionHeader className="panel-head-compact">
          <div>
            <b>授权管理</b>
            <span>{activeLicense?.licenseId || "未导入授权包"}</span>
          </div>
          <ActionGroup>
            <ActionDialog id="system-license-issue" title="签发授权包" buttonLabel="签发授权" triggerIcon={<FileSignature size={13} />} triggerVariant="primary" onOpen={resetLicenseIssueForm}>
              {licenseIssueFormView}
            </ActionDialog>
            {reloadButton()}
          </ActionGroup>
        </SectionHeader>
        <MetricList compact className="system-summary-grid">
          <div><span>校验状态</span><b><StatusChip value={verification?.valid ? "valid" : "missing"} /></b></div>
          <div><span>客户</span><b>{activeLicense?.customerName || "-"}</b></div>
          <div><span>到期日</span><b>{activeLicense?.expiresAt || "-"}</b></div>
          <div><span>授权包</span><b>{licensePackages.length}</b></div>
          <div><span>签发记录</span><b>{licenseIssues.length}</b></div>
          <div><span>吊销记录</span><b>{revocations.length}</b></div>
        </MetricList>
        <SystemForm onSubmit={handleImportLicensePackage}>
          <Field label="授权包 JSON" spanAll>
            <TextAreaInput
              value={licenseImportText}
              onChange={(event) => setLicenseImportText(event.target.value)}
              placeholder="粘贴运营平台下载的授权包 JSON"
            />
          </Field>
          <FormActions spanAll>
            <UiButton type="button" onClick={() => setLicenseImportText("")} disabled={actionBusy !== "" || !licenseImportText}>清空</UiButton>
            <UiButton variant="primary" type="submit" icon={<FileSignature size={14} />} disabled={actionBusy !== "" || !licenseImportText.trim()}>导入授权包</UiButton>
          </FormActions>
        </SystemForm>
        <DataTable
          title="授权包历史"
          data={licensePackages}
          rowKey={(item) => item.id}
          pageSize={8}
          onRefresh={refreshData}
          columns={[
            { key: "license", title: "授权", render: (item) => <><b>{item.licenseId}</b><span className="block-text muted">{item.customerName}</span></> },
            { key: "edition", title: "版本", render: (item) => <><span>{item.edition}</span><span className="block-text muted">{list(item.modules).join(" / ")}</span></> },
            { key: "quota", title: "额度", render: (item) => <><span>{item.maxSites} 站</span><span className="block-text muted">{item.maxVehicles} 车</span></> },
            { key: "expiresAt", title: "到期日", render: (item) => item.expiresAt },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status || item.lastVerificationState} /> },
            { key: "fingerprint", title: "公钥指纹", render: (item) => item.publicKeyFingerprint || "-" },
            {
              key: "actions",
              title: "操作",
              width: "230px",
              render: (item) => (
                <ActionGroup>
                  <UiButton
                    size="sm"
                    icon={<Download size={13} />}
                    disabled={actionBusy !== ""}
                    onClick={() => runBusinessAction(`system-license-download-${item.id}`, "授权包已下载", () => downloadLicensePackageFile(item.id))}
                  >
                    下载
                  </UiButton>
                  <ActionDialog id={`system-license-renew-${item.id}`} title="续期授权包" buttonLabel="续期" onOpen={() => startRenewLicensePackage(item)}>
                    {licenseRenewFormView}
                  </ActionDialog>
                  <ActionDialog id={`system-license-revoke-${item.id}`} title="吊销授权" buttonLabel="吊销" onOpen={() => startRevokeLicensePackage(item)}>
                    {licenseRevokeFormView}
                  </ActionDialog>
                </ActionGroup>
              )
            }
          ]}
          emptyText="暂无授权包"
        />
        <SectionGrid className="finance-list-page">
          <DataTable
            title="签发记录"
            data={licenseIssues}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "issue", title: "签发", render: (item) => <><b>{item.issueNo}</b><span className="block-text muted">{item.licenseId}</span></> },
              { key: "customer", title: "客户", render: (item) => item.customerName },
              { key: "expiresAt", title: "到期日", render: (item) => item.expiresAt },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
            ]}
            emptyText="暂无签发记录"
          />
          <DataTable
            title="吊销记录"
            data={revocations}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "revoke", title: "吊销", render: (item) => <><b>{item.revokeNo}</b><span className="block-text muted">{item.licenseId}</span></> },
              { key: "reason", title: "原因", render: (item) => item.reason || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "revokedAt", title: "时间", render: (item) => shortDateTime(item.revokedAt) },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
            ]}
            emptyText="暂无吊销记录"
          />
        </SectionGrid>
      </Panel>
    );
  }

  function renderSystemMaintenance() {
    const system = data.system;
    const modules = data.modules.length ? data.modules : list(bootstrap?.modules);
    const updates = [...list(system?.updates)].sort((a, b) => b.id - a.id);
    const security = system?.security;
    const securityPolicies = list(security?.policies);
    const deviceCredentials = list(security?.deviceCredentials);
    const activeSessions = list(security?.sessions);
    const securityReport = security?.report;
    const backups = [...list(system?.backups)].sort((a, b) => (b.createdAt || "").localeCompare(a.createdAt || ""));
    const backupDrills = [...list(system?.backupDrills)].sort((a, b) => b.id - a.id);
    const gateway = system?.gateway;
    const gatewayRoutes = list(gateway?.routes);
    const gatewayEvents = list(gateway?.events);
    const reloadPlan = gateway?.reloadPlan;
    const ssoProviders = list(system?.security?.ssoProviders);
    const scimProviders = list(system?.security?.scimProviders);
    const scimEvents = [...list(system?.security?.scimEvents)].sort((a, b) => b.id - a.id);
    const plugins = list(system?.plugins);
    const pluginRuns = [...list(system?.pluginRuns)].sort((a, b) => b.id - a.id);
    const ruleOverview = data.rules;
    const integrationOverview = data.integrations;
    const runtime = system?.runtime;
    const rules = list(ruleOverview?.rules);
    const notifications = [...list(ruleOverview?.notifications)].sort((a, b) => b.id - a.id);
    const integrationEndpoints = list(integrationOverview?.endpoints);
    const protocolFrames = [...list(integrationOverview?.protocolFrames)].sort((a, b) => b.id - a.id);
    const ruleCount = rules.length;
    const notificationCount = notifications.length;
    const integrationEndpointCount = integrationEndpoints.length;
    const protocolFrameCount = protocolFrames.length;
    const roleOptions = availableRoles();
    const companyOptions = list(bootstrap?.companies);
    const siteOptions = list(bootstrap?.sites);
    const bytesText = (value: number | undefined) => {
      const size = Number(value || 0);
      if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
      if (size >= 1024) return `${Math.ceil(size / 1024)} KB`;
      return `${size} B`;
    };
    const updatePublishForm = (
      <SystemForm onSubmit={handlePublishUpdatePackage}>
        <Field label="版本"><TextInput value={updatePackageForm.version} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, version: event.target.value })} /></Field>
        <Field label="组件">
          <SelectInput value={updatePackageForm.component} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, component: event.target.value })}>
            <option value="server">server</option>
            <option value="client">client</option>
            <option value="all">all</option>
          </SelectInput>
        </Field>
        <Field label="通道"><TextInput value={updatePackageForm.channel} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, channel: event.target.value })} /></Field>
        <Field label="状态"><TextInput value={updatePackageForm.status} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, status: event.target.value })} /></Field>
        <Field label="类型">
          <SelectInput value={updatePackageForm.packageType} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, packageType: event.target.value })}>
            <option value="full">full</option>
            <option value="delta">delta</option>
          </SelectInput>
        </Field>
        <Field label="基线版本"><TextInput value={updatePackageForm.baseVersion} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, baseVersion: event.target.value })} /></Field>
        <Field label="回滚版本"><TextInput value={updatePackageForm.rollbackVersion} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, rollbackVersion: event.target.value })} /></Field>
        <Field label="更新文件"><input type="file" onChange={handleUpdatePackageFile} /></Field>
        <Field label="文件名"><TextInput value={updatePackageForm.artifactFileName} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, artifactFileName: event.target.value })} /></Field>
        <Field label="内容类型"><TextInput value={updatePackageForm.artifactContentType} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, artifactContentType: event.target.value })} /></Field>
        <Field label="目标 SHA256"><TextInput value={updatePackageForm.targetArtifactSha256} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, targetArtifactSha256: event.target.value })} /></Field>
        <Field label="Artifact Base64" spanAll><TextAreaInput value={updatePackageForm.artifactContentBase64} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, artifactContentBase64: event.target.value })} /></Field>
        <Field label="备注" spanAll><TextAreaInput value={updatePackageForm.remark} onChange={(event) => setUpdatePackageForm({ ...updatePackageForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<ArrowUp size={14} />} disabled={actionBusy !== "" || !updatePackageForm.version.trim() || !updatePackageForm.artifactContentBase64.trim()}>发布更新包</UiButton>
        </FormActions>
      </SystemForm>
    );
    const updateActions = (item: UpdatePackage) => (
      <ActionGroup>
        <UiButton size="sm" icon={<Download size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`system-update-download-${item.id}`, "更新包已下载", () => downloadUpdatePackageFile(item.id))}>下载</UiButton>
        <UiButton size="sm" disabled={actionBusy !== "" || item.status === "installed"} onClick={() => runBusinessAction(`system-update-apply-${item.id}`, "更新包已应用", () => api.applyUpdate(item.id))}>应用</UiButton>
        <UiButton size="sm" disabled={actionBusy !== "" || !item.rollbackVersion} onClick={() => runBusinessAction(`system-update-rollback-${item.id}`, "更新包已回滚", () => api.rollbackUpdate(item.id))}>回滚</UiButton>
      </ActionGroup>
    );
    const backupActions = (item: BackupInfo) => (
      <ActionGroup>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`system-backup-restore-${item.name}`, "备份已恢复", () => api.restoreBackup(item.name))}>恢复</UiButton>
      </ActionGroup>
    );
    const moduleActions = (item: ModuleInfo) => (
      <ActionGroup>
        <UiButton size="sm" disabled={actionBusy !== "" || !item.hotPlug} onClick={() => handleToggleModule(item)}>{item.enabled ? "停用" : "启用"}</UiButton>
      </ActionGroup>
    );
    const gatewayRouteFormView = (
      <SystemForm onSubmit={handleSaveGatewayRoute}>
        <Field label="名称"><TextInput value={gatewayRouteForm.name} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, name: event.target.value })} /></Field>
        <Field label="路径前缀"><TextInput value={gatewayRouteForm.pathPrefix} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, pathPrefix: event.target.value })} /></Field>
        <Field label="稳定上游"><TextInput value={gatewayRouteForm.stableUpstream} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, stableUpstream: event.target.value })} /></Field>
        <Field label="灰度上游"><TextInput value={gatewayRouteForm.canaryUpstream} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, canaryUpstream: event.target.value })} /></Field>
        <Field label="灰度比例"><TextInput type="number" min="0" max="100" value={gatewayRouteForm.canaryPercent} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, canaryPercent: event.target.value })} /></Field>
        <Field label="超时秒数"><TextInput type="number" min="1" value={gatewayRouteForm.readTimeoutSec} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, readTimeoutSec: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={gatewayRouteForm.status} onChange={(event) => setGatewayRouteForm({ ...gatewayRouteForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !gatewayRouteForm.pathPrefix.trim() || !gatewayRouteForm.stableUpstream.trim()}>保存路由</UiButton>
        </FormActions>
      </SystemForm>
    );
    const gatewayCanaryFormView = (
      <SystemForm onSubmit={handleSetGatewayCanary}>
        <Field label="灰度比例"><TextInput type="number" min="0" max="100" value={gatewayCanaryForm.canaryPercent} onChange={(event) => setGatewayCanaryForm({ ...gatewayCanaryForm, canaryPercent: event.target.value })} /></Field>
        <Field label="灰度上游"><TextInput value={gatewayCanaryForm.canaryUpstream} onChange={(event) => setGatewayCanaryForm({ ...gatewayCanaryForm, canaryUpstream: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<RefreshCw size={14} />} disabled={actionBusy !== ""}>更新灰度</UiButton>
        </FormActions>
      </SystemForm>
    );
    const ssoProviderFormView = (
      <SystemForm onSubmit={handleSaveSSOProvider}>
        <Field label="名称"><TextInput value={ssoProviderForm.name} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, name: event.target.value })} /></Field>
        <Field label="编码"><TextInput value={ssoProviderForm.code} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, code: event.target.value })} /></Field>
        <Field label="Issuer"><TextInput value={ssoProviderForm.issuer} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, issuer: event.target.value })} /></Field>
        <Field label="Client ID"><TextInput value={ssoProviderForm.clientId} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, clientId: event.target.value })} /></Field>
        <Field label="Client Secret"><TextInput type="password" value={ssoProviderForm.clientSecret} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, clientSecret: event.target.value })} /></Field>
        <Field label="Auth URL"><TextInput value={ssoProviderForm.authUrl} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, authUrl: event.target.value })} /></Field>
        <Field label="Token URL"><TextInput value={ssoProviderForm.tokenUrl} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, tokenUrl: event.target.value })} /></Field>
        <Field label="UserInfo URL"><TextInput value={ssoProviderForm.userInfoUrl} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, userInfoUrl: event.target.value })} /></Field>
        <Field label="JWKS URL"><TextInput value={ssoProviderForm.jwksUrl} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, jwksUrl: event.target.value })} /></Field>
        <Field label="Redirect URI"><TextInput value={ssoProviderForm.redirectUri} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, redirectUri: event.target.value })} /></Field>
        <Field label="默认角色">
          <SelectInput value={ssoProviderForm.roleCode} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, roleCode: event.target.value })}>
            {roleOptions.map((item) => <option key={item.code} value={item.code}>{item.name} / {item.code}</option>)}
          </SelectInput>
        </Field>
        <Field label="公司">
          <SelectInput value={ssoProviderForm.companyId} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, companyId: event.target.value })}>
            <option value="0">不绑定</option>
            {companyOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="站点">
          <SelectInput value={ssoProviderForm.siteId} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, siteId: event.target.value })}>
            <option value="0">不绑定</option>
            {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="自动开户">
          <SelectInput value={ssoProviderForm.autoProvision} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, autoProvision: event.target.value })}>
            <option value="true">启用</option>
            <option value="false">关闭</option>
          </SelectInput>
        </Field>
        <Field label="状态">
          <SelectInput value={ssoProviderForm.status} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, status: event.target.value })}>
            <option value="enabled">enabled</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <Field label="用户名 Claim"><TextInput value={ssoProviderForm.usernameClaim} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, usernameClaim: event.target.value })} /></Field>
        <Field label="显示名 Claim"><TextInput value={ssoProviderForm.displayNameClaim} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, displayNameClaim: event.target.value })} /></Field>
        <Field label="Scopes" spanAll><TextAreaInput value={ssoProviderForm.scopes} onChange={(event) => setSsoProviderForm({ ...ssoProviderForm, scopes: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetSSOProviderForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !ssoProviderForm.name.trim() || !ssoProviderForm.code.trim()}>保存 SSO</UiButton>
        </FormActions>
      </SystemForm>
    );
    const scimProviderFormView = (
      <SystemForm onSubmit={handleSaveSCIMProvider}>
        <Field label="名称"><TextInput value={scimProviderForm.name} onChange={(event) => setScimProviderForm({ ...scimProviderForm, name: event.target.value })} /></Field>
        <Field label="编码"><TextInput value={scimProviderForm.code} onChange={(event) => setScimProviderForm({ ...scimProviderForm, code: event.target.value })} /></Field>
        <Field label="Bearer Token"><TextInput type="password" value={scimProviderForm.bearerToken} onChange={(event) => setScimProviderForm({ ...scimProviderForm, bearerToken: event.target.value })} /></Field>
        <Field label="默认角色">
          <SelectInput value={scimProviderForm.defaultRoleCode} onChange={(event) => setScimProviderForm({ ...scimProviderForm, defaultRoleCode: event.target.value })}>
            {roleOptions.map((item) => <option key={item.code} value={item.code}>{item.name} / {item.code}</option>)}
          </SelectInput>
        </Field>
        <Field label="公司">
          <SelectInput value={scimProviderForm.companyId} onChange={(event) => setScimProviderForm({ ...scimProviderForm, companyId: event.target.value })}>
            <option value="0">默认公司</option>
            {companyOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="站点">
          <SelectInput value={scimProviderForm.siteId} onChange={(event) => setScimProviderForm({ ...scimProviderForm, siteId: event.target.value })}>
            <option value="0">不绑定</option>
            {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </SelectInput>
        </Field>
        <Field label="状态">
          <SelectInput value={scimProviderForm.status} onChange={(event) => setScimProviderForm({ ...scimProviderForm, status: event.target.value })}>
            <option value="enabled">enabled</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <FormActions spanAll>
          <UiButton onClick={resetSCIMProviderForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !scimProviderForm.name.trim() || !scimProviderForm.code.trim()}>保存 SCIM</UiButton>
        </FormActions>
      </SystemForm>
    );
    const securityPolicyFormView = (
      <SystemForm onSubmit={handleSaveSecurityPolicy}>
        <Field label="名称"><TextInput value={securityPolicyForm.name} onChange={(event) => setSecurityPolicyForm({ ...securityPolicyForm, name: event.target.value })} /></Field>
        <Field label="类型"><TextInput value={securityPolicyForm.type} onChange={(event) => setSecurityPolicyForm({ ...securityPolicyForm, type: event.target.value })} /></Field>
        <Field label="策略值"><TextInput value={securityPolicyForm.value} onChange={(event) => setSecurityPolicyForm({ ...securityPolicyForm, value: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={securityPolicyForm.enabled} onChange={(event) => setSecurityPolicyForm({ ...securityPolicyForm, enabled: event.target.value })}>
            <option value="true">启用</option>
            <option value="false">停用</option>
          </SelectInput>
        </Field>
        <Field label="备注" spanAll><TextAreaInput value={securityPolicyForm.remark} onChange={(event) => setSecurityPolicyForm({ ...securityPolicyForm, remark: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetSecurityPolicyForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<ShieldCheck size={14} />} disabled={actionBusy !== "" || !securityPolicyForm.name.trim() || !securityPolicyForm.type.trim()}>保存安全策略</UiButton>
        </FormActions>
      </SystemForm>
    );
    const deviceCredentialFormView = (
      <SystemForm onSubmit={handleSaveDeviceCredential}>
        <Field label="设备号"><TextInput value={deviceCredentialForm.deviceNo} onChange={(event) => setDeviceCredentialForm({ ...deviceCredentialForm, deviceNo: event.target.value })} /></Field>
        <Field label="设备密钥"><TextInput type="password" value={deviceCredentialForm.deviceKey} onChange={(event) => setDeviceCredentialForm({ ...deviceCredentialForm, deviceKey: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={deviceCredentialForm.status} onChange={(event) => setDeviceCredentialForm({ ...deviceCredentialForm, status: event.target.value })}>
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <Field label="权限范围" spanAll><TextAreaInput value={deviceCredentialForm.scopes} onChange={(event) => setDeviceCredentialForm({ ...deviceCredentialForm, scopes: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetDeviceCredentialForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<ShieldCheck size={14} />} disabled={actionBusy !== "" || !deviceCredentialForm.deviceNo.trim() || (!deviceCredentialForm.id && !deviceCredentialForm.deviceKey.trim()) || !integrationListFromText(deviceCredentialForm.scopes).length}>保存设备凭证</UiButton>
        </FormActions>
      </SystemForm>
    );
    const integrationEndpointFormView = (
      <SystemForm onSubmit={handleSaveIntegrationEndpoint}>
        <Field label="名称"><TextInput value={integrationEndpointForm.name} onChange={(event) => setIntegrationEndpointForm({ ...integrationEndpointForm, name: event.target.value })} /></Field>
        <Field label="类型">
          <SelectInput value={integrationEndpointForm.type} onChange={(event) => setIntegrationEndpointForm({ ...integrationEndpointForm, type: event.target.value })}>
            <option value="collection_sms">collection_sms</option>
            <option value="collection_wecom">collection_wecom</option>
            <option value="tax_gateway">tax_gateway</option>
            <option value="workflow_webhook">workflow_webhook</option>
            <option value="monitoring">monitoring</option>
          </SelectInput>
        </Field>
        <Field label="协议"><TextInput value={integrationEndpointForm.protocol} onChange={(event) => setIntegrationEndpointForm({ ...integrationEndpointForm, protocol: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={integrationEndpointForm.status} onChange={(event) => setIntegrationEndpointForm({ ...integrationEndpointForm, status: event.target.value })}>
            <option value="online">online</option>
            <option value="degraded">degraded</option>
            <option value="offline">offline</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <Field label="URL" spanAll><TextInput value={integrationEndpointForm.url} onChange={(event) => setIntegrationEndpointForm({ ...integrationEndpointForm, url: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetIntegrationEndpointForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !integrationEndpointForm.name.trim() || !integrationEndpointForm.type.trim() || (integrationEndpointForm.status !== "disabled" && !integrationEndpointForm.url.trim())}>保存端点</UiButton>
        </FormActions>
      </SystemForm>
    );
    const ruleDefinitionFormView = (
      <SystemForm onSubmit={handleSaveRuleDefinition}>
        <Field label="规则编码"><TextInput value={ruleDefinitionForm.code} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, code: event.target.value })} /></Field>
        <Field label="规则名称"><TextInput value={ruleDefinitionForm.name} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, name: event.target.value })} /></Field>
        <Field label="分类"><TextInput value={ruleDefinitionForm.category} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, category: event.target.value })} /></Field>
        <Field label="指标">
          <SelectInput value={ruleDefinitionForm.metric} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, metric: event.target.value })}>
            <option value="speed">speed</option>
            <option value="offline_minutes">offline_minutes</option>
          </SelectInput>
        </Field>
        <Field label="操作符">
          <SelectInput value={ruleDefinitionForm.operator} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, operator: event.target.value })}>
            <option value=">">&gt;</option>
            <option value=">=">&gt;=</option>
            <option value="=">=</option>
          </SelectInput>
        </Field>
        <Field label="阈值"><TextInput type="number" min="0" value={ruleDefinitionForm.threshold} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, threshold: event.target.value })} /></Field>
        <Field label="等级">
          <SelectInput value={ruleDefinitionForm.level} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, level: event.target.value })}>
            <option value="info">info</option>
            <option value="warning">warning</option>
            <option value="critical">critical</option>
          </SelectInput>
        </Field>
        <Field label="状态">
          <SelectInput value={ruleDefinitionForm.enabled} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, enabled: event.target.value })}>
            <option value="true">启用</option>
            <option value="false">停用</option>
          </SelectInput>
        </Field>
        <Field label="通知角色" spanAll><TextAreaInput value={ruleDefinitionForm.notifyRoles} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, notifyRoles: event.target.value })} /></Field>
        <Field label="说明" spanAll><TextAreaInput value={ruleDefinitionForm.description} onChange={(event) => setRuleDefinitionForm({ ...ruleDefinitionForm, description: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetRuleDefinitionForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !ruleDefinitionForm.code.trim() || !ruleDefinitionForm.name.trim() || !ruleDefinitionForm.metric.trim()}>保存规则</UiButton>
        </FormActions>
      </SystemForm>
    );
    const pluginInstallFormView = (
      <SystemForm onSubmit={handleInstallPlugin}>
        <Field label="插件 ID"><TextInput value={pluginInstallForm.id} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, id: event.target.value })} /></Field>
        <Field label="名称"><TextInput value={pluginInstallForm.name} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, name: event.target.value })} /></Field>
        <Field label="类型"><TextInput value={pluginInstallForm.type} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, type: event.target.value })} /></Field>
        <Field label="版本"><TextInput value={pluginInstallForm.version} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, version: event.target.value })} /></Field>
        <Field label="状态">
          <SelectInput value={pluginInstallForm.status} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, status: event.target.value })}>
            <option value="installed">installed</option>
            <option value="enabled">enabled</option>
            <option value="disabled">disabled</option>
          </SelectInput>
        </Field>
        <Field label="运行时"><TextInput value={pluginInstallForm.runtime} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, runtime: event.target.value })} /></Field>
        <Field label="入口"><TextInput value={pluginInstallForm.entrypoint} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, entrypoint: event.target.value })} /></Field>
        <Field label="Checksum"><TextInput value={pluginInstallForm.checksum} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, checksum: event.target.value })} /></Field>
        <Field label="Signature"><TextInput value={pluginInstallForm.signature} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, signature: event.target.value })} /></Field>
        <Field label="沙箱运行时"><TextInput value={pluginInstallForm.sandboxRuntime} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, sandboxRuntime: event.target.value })} /></Field>
        <Field label="超时 ms"><TextInput type="number" min="1" value={pluginInstallForm.sandboxTimeoutMs} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, sandboxTimeoutMs: event.target.value })} /></Field>
        <Field label="网络">
          <SelectInput value={pluginInstallForm.sandboxNetwork} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, sandboxNetwork: event.target.value })}>
            <option value="false">关闭</option>
            <option value="true">启用</option>
          </SelectInput>
        </Field>
        <Field label="文件系统"><TextInput value={pluginInstallForm.sandboxFilesystem} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, sandboxFilesystem: event.target.value })} /></Field>
        <Field label="内存 MB"><TextInput type="number" min="1" value={pluginInstallForm.sandboxMaxMemoryMb} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, sandboxMaxMemoryMb: event.target.value })} /></Field>
        <Field label="权限" spanAll><TextAreaInput value={pluginInstallForm.permissions} onChange={(event) => setPluginInstallForm({ ...pluginInstallForm, permissions: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton onClick={resetPluginInstallForm} disabled={actionBusy !== ""}>重置</UiButton>
          <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !pluginInstallForm.id.trim() || !pluginInstallForm.checksum.trim() || !pluginInstallForm.signature.trim()}>安装插件</UiButton>
        </FormActions>
      </SystemForm>
    );
    const pluginRunFormView = (
      <SystemForm onSubmit={handleRunPlugin}>
        <Field label="插件">
          <SelectInput value={pluginRunForm.pluginId} onChange={(event) => setPluginRunForm({ ...pluginRunForm, pluginId: event.target.value })}>
            {plugins.map((item) => <option key={item.id} value={item.id}>{item.name} / {item.id}</option>)}
          </SelectInput>
        </Field>
        <Field label="权限"><TextInput value={pluginRunForm.permission} onChange={(event) => setPluginRunForm({ ...pluginRunForm, permission: event.target.value })} /></Field>
        <Field label="动作"><TextInput value={pluginRunForm.action} onChange={(event) => setPluginRunForm({ ...pluginRunForm, action: event.target.value })} /></Field>
        <Field label="输入 JSON" spanAll><TextAreaInput value={pluginRunForm.input} onChange={(event) => setPluginRunForm({ ...pluginRunForm, input: event.target.value })} /></Field>
        <FormActions spanAll>
          <UiButton variant="primary" type="submit" icon={<PlayCircle size={14} />} disabled={actionBusy !== "" || !pluginRunForm.pluginId}>运行插件</UiButton>
        </FormActions>
      </SystemForm>
    );
    const gatewayActions = (item: GatewayRoute) => (
      <ActionGroup>
        <ActionDialog id={`system-gateway-route-${item.id}`} title="编辑网关路由" buttonLabel="编辑" onOpen={() => startGatewayRouteEdit(item)}>
          {gatewayRouteFormView}
        </ActionDialog>
        <ActionDialog id={`system-gateway-canary-${item.id}`} title="灰度发布" buttonLabel="灰度" onOpen={() => startGatewayCanary(item)}>
          {gatewayCanaryFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`system-gateway-drain-${item.id}`, item.drainEnabled ? "网关排空已停止" : "网关排空已启动", () => api.setGatewayDrain(item.id, !item.drainEnabled))}>{item.drainEnabled ? "停排空" : "排空"}</UiButton>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`system-gateway-status-${item.id}`, item.status === "active" ? "网关路由已禁用" : "网关路由已启用", () => api.setGatewayStatus(item.id, item.status === "active" ? "disabled" : "active"))}>{item.status === "active" ? "禁用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "active"} onClick={() => handleDeleteGatewayRoute(item)}>删除</UiButton>
      </ActionGroup>
    );
    const ssoActions = (item: OIDCProvider) => (
      <ActionGroup>
        <ActionDialog id={`system-sso-provider-${item.id}`} title="编辑 SSO 提供商" buttonLabel="编辑" onOpen={() => startSSOProviderEdit(item)}>
          {ssoProviderFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleSSOProviderStatus(item)}>{item.status === "enabled" ? "停用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "enabled"} onClick={() => handleDeleteSSOProvider(item)}>删除</UiButton>
      </ActionGroup>
    );
    const scimActions = (item: SCIMProvider) => (
      <ActionGroup>
        <ActionDialog id={`system-scim-provider-${item.id}`} title="编辑 SCIM 提供商" buttonLabel="编辑" onOpen={() => startSCIMProviderEdit(item)}>
          {scimProviderFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleSCIMProviderStatus(item)}>{item.status === "enabled" ? "停用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "enabled"} onClick={() => handleDeleteSCIMProvider(item)}>删除</UiButton>
      </ActionGroup>
    );
    const securityPolicyActions = (item: SecurityPolicy) => (
      <ActionGroup>
        <ActionDialog id={`system-security-policy-${item.id}`} title="编辑安全策略" buttonLabel="编辑" onOpen={() => startSecurityPolicyEdit(item)}>
          {securityPolicyFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleToggleSecurityPolicy(item)}>{item.enabled ? "停用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.enabled} onClick={() => handleDeleteSecurityPolicy(item)}>删除</UiButton>
      </ActionGroup>
    );
    const deviceCredentialActions = (item: DeviceCredential) => (
      <ActionGroup>
        <ActionDialog id={`system-device-credential-${item.id}`} title="编辑设备凭证" buttonLabel="编辑" onOpen={() => startDeviceCredentialEdit(item)}>
          {deviceCredentialFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleDeviceCredentialStatus(item)}>{item.status === "active" ? "停用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "active"} onClick={() => handleDeleteDeviceCredential(item)}>删除</UiButton>
      </ActionGroup>
    );
    const integrationEndpointActions = (item: IntegrationEndpoint) => (
      <ActionGroup>
        <ActionDialog id={`system-integration-endpoint-${item.id}`} title="编辑集成端点" buttonLabel="编辑" onOpen={() => startIntegrationEndpointEdit(item)}>
          {integrationEndpointFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== "" || (item.status === "disabled" && !item.url)} onClick={() => handleIntegrationEndpointStatus(item)}>{item.status === "disabled" ? "启用" : "停用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status !== "disabled"} onClick={() => handleDeleteIntegrationEndpoint(item)}>删除</UiButton>
      </ActionGroup>
    );
    const ruleDefinitionActions = (item: RuleDefinition) => (
      <ActionGroup>
        <ActionDialog id={`system-rule-definition-${item.id}`} title="编辑自动化规则" buttonLabel="编辑" onOpen={() => startRuleDefinitionEdit(item)}>
          {ruleDefinitionFormView}
        </ActionDialog>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleRuleDefinitionStatus(item)}>{item.enabled ? "停用" : "启用"}</UiButton>
        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.enabled} onClick={() => handleDeleteRuleDefinition(item)}>删除</UiButton>
      </ActionGroup>
    );
    const pluginActions = (item: PluginInfo) => (
      <ActionGroup>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleVerifyPlugin(item)}>验签</UiButton>
        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handlePluginStatus(item)}>{item.status === "enabled" ? "停用" : "启用"}</UiButton>
        <ActionDialog id={`system-plugin-run-${item.id}`} title="运行插件" buttonLabel="运行" triggerIcon={<PlayCircle size={13} />} onOpen={() => startPluginRun(item)}>
          {pluginRunFormView}
        </ActionDialog>
      </ActionGroup>
    );
    if (section === "system-gateway") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>网关路由</b>
              <span>{gatewayRoutes.length} 条路由 / {gatewayEvents.length} 条事件</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-gateway-route-create" title="新增网关路由" buttonLabel="新增路由" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetGatewayRouteForm}>
                {gatewayRouteFormView}
              </ActionDialog>
              <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== "" || !reloadPlan?.valid} onClick={handleReloadGateway}>重载网关</UiButton>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>路由</span><b>{gatewayRoutes.length}</b></div>
            <div><span>启用</span><b>{gatewayRoutes.filter((item) => item.status === "active").length}</b></div>
            <div><span>排空中</span><b>{gatewayRoutes.filter((item) => item.drainEnabled).length}</b></div>
            <div><span>待重载</span><b>{reloadPlan?.reloadRequired ? "是" : "否"}</b></div>
            <div><span>配置有效</span><b>{reloadPlan?.valid ? "是" : "否"}</b></div>
            <div><span>事件</span><b>{gatewayEvents.length}</b></div>
          </MetricList>
          <DataTable
            title="网关路由"
            data={gatewayRoutes}
            rowKey={(item) => item.id}
            pageSize={10}
            onRefresh={refreshData}
            columns={[
              { key: "route", title: "路由", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.pathPrefix}</span></> },
              { key: "upstream", title: "上游", render: (item) => <><span>{item.stableUpstream}</span><span className="block-text muted">{item.canaryPercent ? `${item.canaryPercent}% -> ${item.canaryUpstream}` : "无灰度"}</span></> },
              { key: "drain", title: "排空", render: (item) => <><StatusChip value={item.drainEnabled ? "active" : "disabled"} /><span className="block-text muted">{shortDateTime(item.drainUntil)}</span></> },
              { key: "timeout", title: "超时", render: (item) => `${item.readTimeoutSec || 0}s` },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "updatedAt", title: "更新时间", render: (item) => shortDateTime(item.updatedAt) },
              { key: "actions", title: "操作", width: "310px", render: gatewayActions }
            ]}
            emptyText="暂无网关路由"
          />
          <SectionGrid className="finance-list-page">
            <DataTable
              title="网关事件"
              data={gatewayEvents}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "event", title: "事件", render: (item) => <><b>{item.eventNo}</b><span className="block-text muted">{item.routeName || "-"}</span></> },
                { key: "action", title: "动作", render: (item) => item.action },
                { key: "detail", title: "详情", render: (item) => item.detail || "-" },
                { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
                { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
              ]}
              emptyText="暂无网关事件"
            />
            <DataTable
              title="排空检查"
              data={list(reloadPlan?.drainChecks)}
              rowKey={(item) => item.routeId}
              pageSize={8}
              columns={[
                { key: "route", title: "路由", render: (item) => <><b>{item.routeName}</b><span className="block-text muted">{item.pathPrefix}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "remaining", title: "剩余", render: (item) => `${item.remainingSeconds || 0}s` },
                { key: "probe", title: "探针", render: (item) => item.probePath || "-" }
              ]}
              emptyText="暂无排空检查"
            />
          </SectionGrid>
        </Panel>
      );
    }
    if (section === "system-security") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>安全策略</b>
              <span>{securityPolicies.length} 条策略 / {deviceCredentials.length} 个设备凭证</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-security-policy-create" title="新增安全策略" buttonLabel="新增策略" triggerIcon={<ShieldCheck size={13} />} triggerVariant="primary" onOpen={resetSecurityPolicyForm}>
                {securityPolicyFormView}
              </ActionDialog>
              <ActionDialog id="system-device-credential-create" title="新增设备凭证" buttonLabel="新增设备凭证" triggerIcon={<ShieldCheck size={13} />} onOpen={resetDeviceCredentialForm}>
                {deviceCredentialFormView}
              </ActionDialog>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>风险等级</span><b>{securityReport?.riskLevel || "-"}</b></div>
            <div><span>启用策略</span><b>{securityPolicies.filter((item) => item.enabled).length}</b></div>
            <div><span>设备凭证</span><b>{deviceCredentials.length}</b></div>
            <div><span>活动会话</span><b>{activeSessions.length}</b></div>
            <div><span>MFA 覆盖</span><b>{Math.round(securityReport?.mfaCoverage || 0)}%</b></div>
            <div><span>失败登录</span><b>{securityReport?.failedLoginLast24h || 0}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="安全策略"
              data={securityPolicies}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              rowContextMenu={buildDataTableRowContextMenu<SecurityPolicy>({
                actions: [
                  { key: "focus-type", label: "只看该类型", onSelect: (item, helpers) => helpers.searchText(item.type) },
                  { key: "focus-status", label: "只看启停状态", onSelect: (item, helpers) => helpers.searchText(item.enabled ? "启用" : "停用") }
                ],
                copyFields: [
                  { key: "name", label: "策略名称", value: (item) => item.name },
                  { key: "type", label: "策略类型", value: (item) => item.type },
                  { key: "value", label: "策略值", value: (item) => item.value }
                ]
              })}
              columns={[
                { key: "policy", title: "策略", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type}</span></> },
                { key: "value", title: "值", render: (item) => item.value || "-" },
                { key: "enabled", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
                { key: "remark", title: "备注", render: (item) => item.remark || "-" },
                { key: "actions", title: "操作", width: "130px", render: securityPolicyActions }
              ]}
              emptyText="暂无安全策略"
            />
            <DataTable
              title="设备凭证"
              data={deviceCredentials}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              rowContextMenu={buildDataTableRowContextMenu<DeviceCredential>({
                actions: [
                  { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceNo) },
                  { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
                ],
                copyFields: [
                  { key: "device", label: "设备号", value: (item) => item.deviceNo },
                  { key: "scopes", label: "权限范围", value: (item) => list(item.scopes).join(", ") },
                  { key: "status", label: "状态", value: (item) => item.status }
                ]
              })}
              columns={[
                { key: "device", title: "设备", render: (item) => <><b>{item.deviceNo}</b><span className="block-text muted">{list(item.scopes).join(" / ") || "-"}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "lastUsedAt", title: "最近使用", render: (item) => shortDateTime(item.lastUsedAt) },
                { key: "actions", title: "操作", width: "130px", render: deviceCredentialActions }
              ]}
              emptyText="暂无设备凭证"
            />
          </SectionGrid>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="活动会话"
              data={activeSessions}
              rowKey={(item) => `${item.userId}-${item.createdAt}-${item.ip}`}
              pageSize={8}
              columns={[
                { key: "user", title: "用户", render: (item) => <><b>{item.displayName || item.username}</b><span className="block-text muted">{item.username} / {item.roleCode}</span></> },
                { key: "ip", title: "来源", render: (item) => <><span>{item.ip || "-"}</span><span className="block-text muted">{item.userAgent || "-"}</span></> },
                { key: "lastSeenAt", title: "最近活跃", render: (item) => shortDateTime(item.lastSeenAt) },
                { key: "expiresAt", title: "过期", render: (item) => shortDateTime(item.expiresAt) }
              ]}
              emptyText="暂无活动会话"
            />
            <DataTable
              title="安全建议"
              data={list(securityReport?.recommendations).map((text, index) => ({ id: index + 1, text }))}
              rowKey={(item) => item.id}
              pageSize={8}
              columns={[
                { key: "recommendation", title: "建议", render: (item) => item.text }
              ]}
              emptyText="暂无安全建议"
            />
          </SectionGrid>
        </Panel>
      );
    }
    if (section === "system-identity") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>身份集成</b>
              <span>{ssoProviders.length} 个 SSO / {scimProviders.length} 个 SCIM</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-sso-provider-create" title="新增 SSO 提供商" buttonLabel="新增 SSO" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetSSOProviderForm}>
                {ssoProviderFormView}
              </ActionDialog>
              <ActionDialog id="system-scim-provider-create" title="新增 SCIM 提供商" buttonLabel="新增 SCIM" triggerIcon={<Plus size={13} />} onOpen={resetSCIMProviderForm}>
                {scimProviderFormView}
              </ActionDialog>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>SSO</span><b>{ssoProviders.length}</b></div>
            <div><span>已启用 SSO</span><b>{ssoProviders.filter((item) => item.status === "enabled").length}</b></div>
            <div><span>SCIM</span><b>{scimProviders.length}</b></div>
            <div><span>已启用 SCIM</span><b>{scimProviders.filter((item) => item.status === "enabled").length}</b></div>
            <div><span>24h 同步</span><b>{securityReport?.scimEventsLast24h || 0}</b></div>
            <div><span>同步事件</span><b>{scimEvents.length}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="SSO 提供商"
              data={ssoProviders}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "provider", title: "提供商", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code}</span></> },
                { key: "issuer", title: "Issuer", render: (item) => item.issuer || "-" },
                { key: "role", title: "默认角色", render: (item) => roleName(item.roleCode) },
                { key: "scope", title: "组织", render: (item) => <><span>{nameOf(companyOptions, item.companyId) || "-"}</span><span className="block-text muted">{nameOf(siteOptions, item.siteId) || "全部站点"}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "lastLoginAt", title: "最近登录", render: (item) => shortDateTime(item.lastLoginAt) },
                { key: "actions", title: "操作", width: "170px", render: ssoActions }
              ]}
              emptyText="暂无 SSO 提供商"
            />
            <DataTable
              title="SCIM 提供商"
              data={scimProviders}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "provider", title: "提供商", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code}</span></> },
                { key: "role", title: "默认角色", render: (item) => roleName(item.defaultRoleCode) },
                { key: "scope", title: "组织", render: (item) => <><span>{nameOf(companyOptions, item.companyId) || "-"}</span><span className="block-text muted">{nameOf(siteOptions, item.siteId) || "全部站点"}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "lastSyncAt", title: "最近同步", render: (item) => shortDateTime(item.lastSyncAt) },
                { key: "actions", title: "操作", width: "170px", render: scimActions }
              ]}
              emptyText="暂无 SCIM 提供商"
            />
          </SectionGrid>
          <DataTable
            title="SCIM 同步事件"
            data={scimEvents}
            rowKey={(item) => item.id}
            pageSize={10}
            columns={[
              { key: "event", title: "事件", render: (item) => <><b>{item.eventNo}</b><span className="block-text muted">{item.providerCode}</span></> },
              { key: "user", title: "用户", render: (item) => item.username || item.userId || "-" },
              { key: "action", title: "动作", render: (item) => item.action },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "detail", title: "详情", render: (item) => item.detail || "-" },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无 SCIM 同步事件"
          />
        </Panel>
      );
    }
    if (section === "system-plugins") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>插件管理</b>
              <span>{plugins.length} 个插件 / {pluginRuns.length} 条运行记录</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-plugin-install" title="安装插件" buttonLabel="安装插件" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetPluginInstallForm}>
                {pluginInstallFormView}
              </ActionDialog>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>插件</span><b>{plugins.length}</b></div>
            <div><span>启用</span><b>{plugins.filter((item) => item.status === "enabled").length}</b></div>
            <div><span>禁用</span><b>{plugins.filter((item) => item.status !== "enabled").length}</b></div>
            <div><span>运行记录</span><b>{pluginRuns.length}</b></div>
          </MetricList>
          <DataTable
            title="插件"
            data={plugins}
            rowKey={(item) => item.id}
            pageSize={10}
            onRefresh={refreshData}
            columns={[
              { key: "plugin", title: "插件", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.id} / {item.version}</span></> },
              { key: "type", title: "类型", render: (item) => item.type || "-" },
              { key: "runtime", title: "运行时", render: (item) => item.runtime || item.sandbox?.runtime || "-" },
              { key: "permissions", title: "权限", render: (item) => <ChipList compact>{list(item.permissions).slice(0, 4).map((permission) => <span key={permission}>{permission}</span>)}</ChipList> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastRunAt", title: "最近运行", render: (item) => shortDateTime(item.lastRunAt) },
              { key: "actions", title: "操作", width: "220px", render: pluginActions }
            ]}
            emptyText="暂无插件"
          />
          <DataTable
            title="插件运行记录"
            data={pluginRuns}
            rowKey={(item) => item.id}
            pageSize={10}
            columns={[
              { key: "run", title: "运行", render: (item: PluginRun) => <><b>{item.runNo}</b><span className="block-text muted">{item.pluginName || item.pluginId}</span></> },
              { key: "action", title: "动作", render: (item) => <><span>{item.action || "-"}</span><span className="block-text muted">{item.permission || "-"}</span></> },
              { key: "runtime", title: "运行时", render: (item) => item.runtime || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "duration", title: "耗时", render: (item) => `${item.durationMs || 0}ms` },
              { key: "output", title: "输出", render: (item) => item.error || item.output || "-" },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
              { key: "completedAt", title: "完成时间", render: (item) => shortDateTime(item.completedAt) }
            ]}
            emptyText="暂无插件运行记录"
          />
        </Panel>
      );
    }
    if (section === "system-rules") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>规则中心</b>
              <span>{ruleCount} 条规则 / {notificationCount} 条通知</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-rule-definition-create" title="新增自动化规则" buttonLabel="新增规则" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetRuleDefinitionForm}>
                {ruleDefinitionFormView}
              </ActionDialog>
              <UiButton disabled={actionBusy !== ""} onClick={handleEvaluateRules}>规则评估</UiButton>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>规则</span><b>{ruleCount}</b></div>
            <div><span>启用</span><b>{rules.filter((item) => item.enabled).length}</b></div>
            <div><span>禁用</span><b>{rules.filter((item) => !item.enabled).length}</b></div>
            <div><span>通知</span><b>{notificationCount}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="自动化规则"
              data={rules}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              rowContextMenu={buildDataTableRowContextMenu<RuleDefinition>({
                actions: [
                  { key: "focus-category", label: "只看该分类", onSelect: (item, helpers) => helpers.searchText(item.category) },
                  { key: "focus-level", label: "只看该等级", onSelect: (item, helpers) => helpers.searchText(item.level) }
                ],
                copyFields: [
                  { key: "code", label: "规则编码", value: (item) => item.code },
                  { key: "metric", label: "指标", value: (item) => item.metric },
                  { key: "description", label: "说明", value: (item) => item.description }
                ]
              })}
              columns={[
                { key: "rule", title: "规则", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code} / {item.category}</span></> },
                { key: "condition", title: "条件", render: (item) => `${item.metric || "-"} ${item.operator || ""} ${item.threshold ?? ""}` },
                { key: "level", title: "等级", render: (item) => <StatusChip value={item.level || "info"} /> },
                { key: "notify", title: "通知角色", render: (item) => list(item.notifyRoles).join(" / ") || "-" },
                { key: "enabled", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
                { key: "actions", title: "操作", width: "130px", render: ruleDefinitionActions }
              ]}
              emptyText="暂无自动化规则"
            />
            <DataTable
              title="规则通知"
              data={notifications}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              columns={[
                { key: "title", title: "通知", render: (item) => <><b>{item.title}</b><span className="block-text muted">{item.content || "-"}</span></> },
                { key: "target", title: "目标", render: (item) => <><span>{roleName(item.targetRole)}</span><span className="block-text muted">{item.channel}</span></> },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
              ]}
              emptyText="暂无规则通知"
            />
          </SectionGrid>
        </Panel>
      );
    }
    if (section === "system-integrations") {
      return (
        <Panel className="system-management-view">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>集成端点</b>
              <span>{integrationEndpointCount} 个端点 / {protocolFrameCount} 条协议帧</span>
            </div>
            <ActionGroup>
              <ActionDialog id="system-integration-endpoint-create" title="新增集成端点" buttonLabel="新增端点" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={resetIntegrationEndpointForm}>
                {integrationEndpointFormView}
              </ActionDialog>
              {reloadButton()}
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="system-summary-grid">
            <div><span>端点</span><b>{integrationEndpointCount}</b></div>
            <div><span>在线</span><b>{integrationEndpoints.filter((item) => item.status === "online").length}</b></div>
            <div><span>禁用</span><b>{integrationEndpoints.filter((item) => item.status === "disabled").length}</b></div>
            <div><span>协议帧</span><b>{protocolFrameCount}</b></div>
          </MetricList>
          <SectionGrid className="finance-list-page">
            <DataTable
              title="集成端点"
              data={integrationEndpoints}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              rowContextMenu={buildDataTableRowContextMenu<IntegrationEndpoint>({
                actions: [
                  { key: "focus-type", label: "只看该类型", onSelect: (item, helpers) => helpers.searchText(item.type) },
                  { key: "focus-protocol", label: "只看该协议", onSelect: (item, helpers) => helpers.searchText(item.protocol) }
                ],
                copyFields: [
                  { key: "name", label: "端点名称", value: (item) => item.name },
                  { key: "url", label: "URL", value: (item) => item.url }
                ]
              })}
              columns={[
                { key: "endpoint", title: "端点", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type} / {item.protocol}</span></> },
                { key: "url", title: "地址", render: (item) => item.url || "-" },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "lastSyncAt", title: "最近同步", render: (item) => shortDateTime(item.lastSyncAt) },
                { key: "actions", title: "操作", width: "130px", render: integrationEndpointActions }
              ]}
              emptyText="暂无集成端点"
            />
            <DataTable
              title="协议帧"
              data={protocolFrames}
              rowKey={(item) => item.id}
              pageSize={10}
              onRefresh={refreshData}
              rowContextMenu={buildDataTableRowContextMenu<DeviceProtocolFrame>({
                actions: [
                  { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceNo) },
                  { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
                ],
                copyFields: [
                  { key: "frame", label: "帧编号", value: (item) => item.frameNo },
                  { key: "raw", label: "原始报文", value: (item) => item.raw }
                ]
              })}
              columns={[
                { key: "frame", title: "帧", render: (item) => <><b>{item.frameNo}</b><span className="block-text muted">{item.channel} / {item.protocol}</span></> },
                { key: "device", title: "设备", render: (item) => item.deviceNo || "-" },
                { key: "parsed", title: "解析对象", render: (item) => item.parsedResource ? `${item.parsedResource}#${item.parsedId || 0}` : "-" },
                { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
                { key: "error", title: "错误", render: (item) => item.error || "-" },
                { key: "receivedAt", title: "接收时间", render: (item) => shortDateTime(item.receivedAt) }
              ]}
              emptyText="暂无协议帧"
            />
          </SectionGrid>
        </Panel>
      );
    }
    return (
      <Panel className="system-management-view">
        <SectionHeader className="panel-head-compact">
          <div>
            <b>系统维护</b>
            <span>{updates.length} 个更新包 / {backups.length} 个备份</span>
          </div>
          <ActionGroup>
            <ActionDialog id="system-update-publish" title="发布更新包" buttonLabel="发布更新" triggerIcon={<RefreshCw size={13} />} triggerVariant="primary" onOpen={resetUpdatePackageForm}>
              {updatePublishForm}
            </ActionDialog>
            <ActionDialog id="system-gateway-route-create" title="新增网关路由" buttonLabel="新增路由" triggerIcon={<Plus size={13} />} onOpen={resetGatewayRouteForm}>
              {gatewayRouteFormView}
            </ActionDialog>
            <ActionDialog id="system-security-policy-create" title="新增安全策略" buttonLabel="新增策略" triggerIcon={<ShieldCheck size={13} />} onOpen={resetSecurityPolicyForm}>
              {securityPolicyFormView}
            </ActionDialog>
            <ActionDialog id="system-device-credential-create" title="新增设备凭证" buttonLabel="新增设备凭证" triggerIcon={<ShieldCheck size={13} />} onOpen={resetDeviceCredentialForm}>
              {deviceCredentialFormView}
            </ActionDialog>
            <ActionDialog id="system-integration-endpoint-create" title="新增集成端点" buttonLabel="新增端点" triggerIcon={<Plus size={13} />} onOpen={resetIntegrationEndpointForm}>
              {integrationEndpointFormView}
            </ActionDialog>
            <ActionDialog id="system-rule-definition-create" title="新增自动化规则" buttonLabel="新增规则" triggerIcon={<Plus size={13} />} onOpen={resetRuleDefinitionForm}>
              {ruleDefinitionFormView}
            </ActionDialog>
            <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== "" || !reloadPlan?.valid} onClick={handleReloadGateway}>重载网关</UiButton>
            <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={handleCreateBackup}>创建备份</UiButton>
            <UiButton disabled={actionBusy !== ""} onClick={handleRunBackupDrill}>备份演练</UiButton>
            <UiButton disabled={actionBusy !== ""} onClick={handleEvaluateRules}>规则评估</UiButton>
            {reloadButton()}
          </ActionGroup>
        </SectionHeader>
        <MetricList compact className="system-summary-grid">
          <div><span>更新包</span><b>{updates.length}</b></div>
          <div><span>业务模块</span><b>{modules.filter((item) => item.enabled).length}/{modules.length}</b></div>
          <div><span>已安装</span><b>{updates.filter((item) => item.status === "installed").length}</b></div>
          <div><span>可用更新</span><b>{updates.filter((item) => item.status === "available").length}</b></div>
          <div><span>备份</span><b>{backups.length}</b></div>
          <div><span>网关路由</span><b>{gatewayRoutes.length}</b></div>
          <div><span>待重载</span><b>{reloadPlan?.reloadRequired ? "是" : "否"}</b></div>
          <div><span>规则</span><b>{ruleCount}</b></div>
          <div><span>通知</span><b>{notificationCount}</b></div>
          <div><span>集成端点</span><b>{integrationEndpointCount}</b></div>
          <div><span>协议帧</span><b>{protocolFrameCount}</b></div>
          <div><span>安全风险</span><b>{securityReport?.riskLevel || "-"}</b></div>
          <div><span>活动会话</span><b>{activeSessions.length}</b></div>
        </MetricList>
        <KeyValueTable
          className="report-form-table"
          rows={[
            [
              { label: "存储", value: runtime?.storage || "-" },
              { label: "业务表", value: `${runtime?.businessTables || "-"} / ${runtime?.businessTableCount || 0}` },
              { label: "领域表", value: `${runtime?.domainTables || "-"} / ${runtime?.domainResourceCount || 0}` }
            ],
            [
              { label: "Redis", value: `${runtime?.redis || "-"} ${runtime?.redisAddr ? `(${runtime.redisAddr})` : ""}` },
              { label: "RabbitMQ", value: `${runtime?.rabbitmq || "-"} ${runtime?.rabbitUrl ? `(${runtime.rabbitUrl})` : ""}` },
              { label: "事件总线", value: runtime?.eventBus || "-" }
            ],
            [
              { label: "税控网关", value: `${runtime?.taxGateway || "-"} / ${runtime?.taxGatewayProvider || "-"}` },
              { label: "地图", value: `${runtime?.mapProvider || "-"} / ${runtime?.mapTiles || "-"}` },
              { label: "设备网关", value: `${list(runtime?.deviceGateways).filter((item) => item.status === "active").length}/${list(runtime?.deviceGateways).length}` }
            ]
          ]}
        />
        <DataTable
          title="业务模块"
          data={modules}
          rowKey={(item) => item.code}
          pageSize={8}
          onRefresh={refreshData}
          rowContextMenu={buildDataTableRowContextMenu<ModuleInfo>({
            actions: [
              { key: "focus-area", label: "只看该领域", onSelect: (item, helpers) => helpers.searchText(item.area) },
              { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.enabled ? "启用" : "停用") }
            ],
            copyFields: [
              { key: "code", label: "模块编码", value: (item) => item.code },
              { key: "name", label: "模块名称", value: (item) => item.name },
              { key: "area", label: "业务领域", value: (item) => item.area }
            ]
          })}
          columns={[
            { key: "module", title: "模块", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code} / {item.version}</span></> },
            { key: "area", title: "领域", render: (item) => item.area || "-" },
            { key: "description", title: "说明", render: (item) => item.description || "-" },
            { key: "hotPlug", title: "热插拔", render: (item) => <StatusChip value={item.hotPlug ? "active" : "disabled"} /> },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
            { key: "actions", title: "操作", width: "100px", render: moduleActions }
          ]}
          emptyText="暂无业务模块"
        />
        <SectionGrid className="finance-list-page">
          <DataTable
            title="安全策略"
            data={securityPolicies}
            rowKey={(item) => item.id}
            pageSize={6}
            rowContextMenu={buildDataTableRowContextMenu<SecurityPolicy>({
              actions: [
                { key: "focus-type", label: "只看该类型", onSelect: (item, helpers) => helpers.searchText(item.type) },
                { key: "focus-status", label: "只看启停状态", onSelect: (item, helpers) => helpers.searchText(item.enabled ? "启用" : "停用") }
              ],
              copyFields: [
                { key: "name", label: "策略名称", value: (item) => item.name },
                { key: "type", label: "策略类型", value: (item) => item.type },
                { key: "value", label: "策略值", value: (item) => item.value }
              ]
            })}
            columns={[
              { key: "policy", title: "策略", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type}</span></> },
              { key: "value", title: "值", render: (item) => item.value || "-" },
              { key: "enabled", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
              { key: "remark", title: "备注", render: (item) => item.remark || "-" },
              { key: "actions", title: "操作", width: "130px", render: securityPolicyActions }
            ]}
            emptyText="暂无安全策略"
          />
          <DataTable
            title="设备凭证"
            data={deviceCredentials}
            rowKey={(item) => item.id}
            pageSize={6}
            rowContextMenu={buildDataTableRowContextMenu<DeviceCredential>({
              actions: [
                { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceNo) },
                { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
              ],
              copyFields: [
                { key: "device", label: "设备号", value: (item) => item.deviceNo },
                { key: "scopes", label: "权限范围", value: (item) => list(item.scopes).join(", ") },
                { key: "status", label: "状态", value: (item) => item.status }
              ]
            })}
            columns={[
              { key: "device", title: "设备", render: (item) => <><b>{item.deviceNo}</b><span className="block-text muted">{list(item.scopes).join(" / ") || "-"}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastUsedAt", title: "最近使用", render: (item) => shortDateTime(item.lastUsedAt) },
              { key: "actions", title: "操作", width: "130px", render: deviceCredentialActions }
            ]}
            emptyText="暂无设备凭证"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="活动会话"
            data={activeSessions}
            rowKey={(item) => `${item.userId}-${item.createdAt}-${item.ip}`}
            pageSize={6}
            columns={[
              { key: "user", title: "用户", render: (item) => <><b>{item.displayName || item.username}</b><span className="block-text muted">{item.username} / {item.roleCode}</span></> },
              { key: "ip", title: "来源", render: (item) => <><span>{item.ip || "-"}</span><span className="block-text muted">{item.userAgent || "-"}</span></> },
              { key: "lastSeenAt", title: "最近活跃", render: (item) => shortDateTime(item.lastSeenAt) },
              { key: "expiresAt", title: "过期", render: (item) => shortDateTime(item.expiresAt) }
            ]}
            emptyText="暂无活动会话"
          />
          <DataTable
            title="安全建议"
            data={list(securityReport?.recommendations).map((text, index) => ({ id: index + 1, text }))}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "recommendation", title: "建议", render: (item) => item.text }
            ]}
            emptyText="暂无安全建议"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="自动化规则"
            data={rules}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<RuleDefinition>({
              actions: [
                { key: "focus-category", label: "只看该分类", onSelect: (item, helpers) => helpers.searchText(item.category) },
                { key: "focus-level", label: "只看该等级", onSelect: (item, helpers) => helpers.searchText(item.level) }
              ],
              copyFields: [
                { key: "code", label: "规则编码", value: (item) => item.code },
                { key: "metric", label: "指标", value: (item) => item.metric },
                { key: "description", label: "说明", value: (item) => item.description }
              ]
            })}
            columns={[
              { key: "rule", title: "规则", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code} / {item.category}</span></> },
              { key: "condition", title: "条件", render: (item) => `${item.metric || "-"} ${item.operator || ""} ${item.threshold ?? ""}` },
              { key: "level", title: "等级", render: (item) => <StatusChip value={item.level || "info"} /> },
              { key: "notify", title: "通知角色", render: (item) => list(item.notifyRoles).join(" / ") || "-" },
              { key: "enabled", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
              { key: "actions", title: "操作", width: "130px", render: ruleDefinitionActions }
            ]}
            emptyText="暂无自动化规则"
          />
          <DataTable
            title="规则通知"
            data={notifications}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            columns={[
              { key: "title", title: "通知", render: (item) => <><b>{item.title}</b><span className="block-text muted">{item.content || "-"}</span></> },
              { key: "target", title: "目标", render: (item) => <><span>{roleName(item.targetRole)}</span><span className="block-text muted">{item.channel}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无规则通知"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="集成端点"
            data={integrationEndpoints}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<IntegrationEndpoint>({
              actions: [
                { key: "focus-type", label: "只看该类型", onSelect: (item, helpers) => helpers.searchText(item.type) },
                { key: "focus-protocol", label: "只看该协议", onSelect: (item, helpers) => helpers.searchText(item.protocol) }
              ],
              copyFields: [
                { key: "name", label: "端点名称", value: (item) => item.name },
                { key: "url", label: "URL", value: (item) => item.url }
              ]
            })}
            columns={[
              { key: "endpoint", title: "端点", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.type} / {item.protocol}</span></> },
              { key: "url", title: "地址", render: (item) => item.url || "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastSyncAt", title: "最近同步", render: (item) => shortDateTime(item.lastSyncAt) },
              { key: "actions", title: "操作", width: "130px", render: integrationEndpointActions }
            ]}
            emptyText="暂无集成端点"
          />
          <DataTable
            title="协议帧"
            data={protocolFrames}
            rowKey={(item) => item.id}
            pageSize={6}
            onRefresh={refreshData}
            rowContextMenu={buildDataTableRowContextMenu<DeviceProtocolFrame>({
              actions: [
                { key: "focus-device", label: "只看该设备", onSelect: (item, helpers) => helpers.searchText(item.deviceNo) },
                { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) }
              ],
              copyFields: [
                { key: "frame", label: "帧编号", value: (item) => item.frameNo },
                { key: "raw", label: "原始报文", value: (item) => item.raw }
              ]
            })}
            columns={[
              { key: "frame", title: "帧", render: (item) => <><b>{item.frameNo}</b><span className="block-text muted">{item.channel} / {item.protocol}</span></> },
              { key: "device", title: "设备", render: (item) => item.deviceNo || "-" },
              { key: "parsed", title: "解析对象", render: (item) => item.parsedResource ? `${item.parsedResource}#${item.parsedId || 0}` : "-" },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "error", title: "错误", render: (item) => item.error || "-" },
              { key: "receivedAt", title: "接收时间", render: (item) => shortDateTime(item.receivedAt) }
            ]}
            emptyText="暂无协议帧"
          />
        </SectionGrid>
        <DataTable
          title="系统更新包"
          data={updates}
          rowKey={(item) => item.id}
          pageSize={8}
          onRefresh={refreshData}
          columns={[
            { key: "version", title: "版本", render: (item) => <><b>{item.component} {item.version}</b><span className="block-text muted">{item.channel} / {item.packageType || "full"}</span></> },
            { key: "artifact", title: "文件", render: (item) => <><span>{item.artifactFileName || item.fileName || "-"}</span><span className="block-text muted">{bytesText(item.artifactSizeBytes || item.sizeBytes)}</span></> },
            { key: "checksum", title: "校验", render: (item) => item.artifactSha256 || item.checksum || "-" },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "downloads", title: "下载", render: (item) => <><span>{item.downloadCount || 0}</span><span className="block-text muted">{shortDateTime(item.lastDownloadedAt)}</span></> },
            { key: "actions", title: "操作", width: "240px", render: updateActions }
          ]}
          emptyText="暂无更新包"
        />
        <DataTable
          title="网关路由"
          data={gatewayRoutes}
          rowKey={(item) => item.id}
          pageSize={8}
          onRefresh={refreshData}
          columns={[
            { key: "route", title: "路由", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.pathPrefix}</span></> },
            { key: "upstream", title: "上游", render: (item) => <><span>{item.stableUpstream}</span><span className="block-text muted">{item.canaryPercent ? `${item.canaryPercent}% -> ${item.canaryUpstream}` : "无灰度"}</span></> },
            { key: "drain", title: "排空", render: (item) => <><StatusChip value={item.drainEnabled ? "active" : "disabled"} /><span className="block-text muted">{shortDateTime(item.drainUntil)}</span></> },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "updatedAt", title: "更新时间", render: (item) => shortDateTime(item.updatedAt) },
            { key: "actions", title: "操作", width: "310px", render: gatewayActions }
          ]}
          emptyText="暂无网关路由"
        />
        <SectionGrid className="finance-list-page">
          <DataTable
            title="网关事件"
            data={gatewayEvents}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "event", title: "事件", render: (item) => <><b>{item.eventNo}</b><span className="block-text muted">{item.routeName || "-"}</span></> },
              { key: "action", title: "动作", render: (item) => item.action },
              { key: "detail", title: "详情", render: (item) => item.detail || "-" },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无网关事件"
          />
          <DataTable
            title="排空检查"
            data={list(reloadPlan?.drainChecks)}
            rowKey={(item) => item.routeId}
            pageSize={6}
            columns={[
              { key: "route", title: "路由", render: (item) => <><b>{item.routeName}</b><span className="block-text muted">{item.pathPrefix}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "remaining", title: "剩余", render: (item) => `${item.remainingSeconds || 0}s` },
              { key: "probe", title: "探针", render: (item) => item.probePath || "-" }
            ]}
            emptyText="暂无排空检查"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="备份"
            data={backups}
            rowKey={(item) => item.name}
            pageSize={6}
            columns={[
              { key: "name", title: "备份", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.path}</span></> },
              { key: "size", title: "大小", render: (item) => bytesText(item.size) },
              { key: "createdAt", title: "创建时间", render: (item) => shortDateTime(item.createdAt) },
              { key: "actions", title: "操作", width: "110px", render: backupActions }
            ]}
            emptyText="暂无备份"
          />
          <DataTable
            title="备份演练"
            data={backupDrills}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "drill", title: "演练", render: (item: BackupDrill) => <><b>{item.drillNo}</b><span className="block-text muted">{item.backupName || "-"}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "duration", title: "耗时", render: (item) => `${Math.round((item.durationMs || 0) / 1000)}s` },
              { key: "checks", title: "校验", render: (item) => list(item.checks).join(" / ") || item.error || "-" },
              { key: "actor", title: "操作人", render: (item) => item.actor || "-" }
            ]}
            emptyText="暂无备份演练"
          />
        </SectionGrid>
        <SectionHeader className="panel-head-compact">
          <div>
            <b>身份集成和插件</b>
            <span>{ssoProviders.length} 个 SSO / {scimProviders.length} 个 SCIM / {plugins.length} 个插件</span>
          </div>
          <ActionGroup>
            <ActionDialog id="system-sso-provider-create" title="新增 SSO 提供商" buttonLabel="新增 SSO" triggerIcon={<Plus size={13} />} onOpen={resetSSOProviderForm}>
              {ssoProviderFormView}
            </ActionDialog>
            <ActionDialog id="system-scim-provider-create" title="新增 SCIM 提供商" buttonLabel="新增 SCIM" triggerIcon={<Plus size={13} />} onOpen={resetSCIMProviderForm}>
              {scimProviderFormView}
            </ActionDialog>
            <ActionDialog id="system-plugin-install" title="安装插件" buttonLabel="安装插件" triggerIcon={<Plus size={13} />} onOpen={resetPluginInstallForm}>
              {pluginInstallFormView}
            </ActionDialog>
          </ActionGroup>
        </SectionHeader>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="SSO 提供商"
            data={ssoProviders}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "provider", title: "提供商", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code}</span></> },
              { key: "issuer", title: "Issuer", render: (item) => item.issuer || "-" },
              { key: "role", title: "默认角色", render: (item) => roleName(item.roleCode) },
              { key: "scope", title: "组织", render: (item) => <><span>{nameOf(companyOptions, item.companyId) || "-"}</span><span className="block-text muted">{nameOf(siteOptions, item.siteId) || "全部站点"}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastLoginAt", title: "最近登录", render: (item) => shortDateTime(item.lastLoginAt) },
              { key: "actions", title: "操作", width: "170px", render: ssoActions }
            ]}
            emptyText="暂无 SSO 提供商"
          />
          <DataTable
            title="SCIM 提供商"
            data={scimProviders}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "provider", title: "提供商", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code}</span></> },
              { key: "role", title: "默认角色", render: (item) => roleName(item.defaultRoleCode) },
              { key: "scope", title: "组织", render: (item) => <><span>{nameOf(companyOptions, item.companyId) || "-"}</span><span className="block-text muted">{nameOf(siteOptions, item.siteId) || "全部站点"}</span></> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastSyncAt", title: "最近同步", render: (item) => shortDateTime(item.lastSyncAt) },
              { key: "actions", title: "操作", width: "170px", render: scimActions }
            ]}
            emptyText="暂无 SCIM 提供商"
          />
        </SectionGrid>
        <SectionGrid className="finance-list-page">
          <DataTable
            title="SCIM 同步事件"
            data={scimEvents}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "event", title: "事件", render: (item) => <><b>{item.eventNo}</b><span className="block-text muted">{item.providerCode}</span></> },
              { key: "user", title: "用户", render: (item) => item.username || item.userId || "-" },
              { key: "action", title: "动作", render: (item) => item.action },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "detail", title: "详情", render: (item) => item.detail || "-" },
              { key: "createdAt", title: "时间", render: (item) => shortDateTime(item.createdAt) }
            ]}
            emptyText="暂无 SCIM 同步事件"
          />
          <DataTable
            title="插件"
            data={plugins}
            rowKey={(item) => item.id}
            pageSize={6}
            columns={[
              { key: "plugin", title: "插件", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.id} / {item.version}</span></> },
              { key: "type", title: "类型", render: (item) => item.type || "-" },
              { key: "runtime", title: "运行时", render: (item) => item.runtime || item.sandbox?.runtime || "-" },
              { key: "permissions", title: "权限", render: (item) => <ChipList compact>{list(item.permissions).slice(0, 4).map((permission) => <span key={permission}>{permission}</span>)}</ChipList> },
              { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
              { key: "lastRunAt", title: "最近运行", render: (item) => shortDateTime(item.lastRunAt) },
              { key: "actions", title: "操作", width: "220px", render: pluginActions }
            ]}
            emptyText="暂无插件"
          />
        </SectionGrid>
        <DataTable
          title="插件运行记录"
          data={pluginRuns}
          rowKey={(item) => item.id}
          pageSize={8}
          columns={[
            { key: "run", title: "运行", render: (item: PluginRun) => <><b>{item.runNo}</b><span className="block-text muted">{item.pluginName || item.pluginId}</span></> },
            { key: "action", title: "动作", render: (item) => <><span>{item.action || "-"}</span><span className="block-text muted">{item.permission || "-"}</span></> },
            { key: "runtime", title: "运行时", render: (item) => item.runtime || "-" },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "duration", title: "耗时", render: (item) => `${item.durationMs || 0}ms` },
            { key: "output", title: "输出", render: (item) => item.error || item.output || "-" },
            { key: "actor", title: "操作人", render: (item) => item.actor || "-" },
            { key: "completedAt", title: "完成时间", render: (item) => shortDateTime(item.completedAt) }
          ]}
          emptyText="暂无插件运行记录"
        />
      </Panel>
    );
  }

  function renderOrganizationManagement() {
    const org = organizationData();
    const companyOptions = org.companies.length ? org.companies : list(bootstrap?.companies);
    const departmentOptions = org.departments.filter((item) => String(item.companyId) === String(orgForm.departmentCompanyId));
    const childMap = organizationChildMap(org.nodes);
    const orgRows = organizationTreeRows(org.nodes, childMap, collapsedOrgNodeIds);
    const companyLevelOptions = dictionaryOptions("org_company_level");
    const canWriteOrg = hasPermission(currentPermissions, "org:write");
    const orgRowContextMenu = (item: OrganizationTreeRow, _index: number, helpers: DataTableContextMenuHelpers<OrganizationTreeRow>): ContextMenuItem[] => {
      const items: ContextMenuItem[] = [];
      if (canWriteOrg && actionBusy === "" && canCreateChildCompany(item)) {
        items.push({
          key: "org-add-child-company",
          label: item.kind === "group" ? "新增下级公司" : "新增子公司",
          icon: <Building2 size={14} />,
          onSelect: () => openCreateChildCompanyDialog(item)
        });
      }
      if (canWriteOrg && actionBusy === "" && canCreateChildDepartment(item)) {
        items.push({
          key: "org-add-child-department",
          label: item.kind === "department" ? "新增子部门" : "新增直属部门",
          icon: <Users size={14} />,
          onSelect: () => openCreateChildDepartmentDialog(item)
        });
      }
      if (items.length) {
        items.push({ key: "org-add-separator", type: "separator" });
      }
      return [
        ...items,
        { key: "focus-node", label: "只看该组织", icon: <Search size={14} />, onSelect: () => helpers.searchText(item.name) },
        { key: "focus-kind", label: "只看该类型", icon: <Filter size={14} />, onSelect: () => helpers.searchText(orgLevelLabel(item.kind)) },
        { key: "copy-name", label: "复制组织名称", icon: <Copy size={14} />, onSelect: () => helpers.copyText(item.name, "组织名称") },
        { key: "copy-code", label: "复制组织编码", icon: <Copy size={14} />, disabled: !item.code, onSelect: () => helpers.copyText(item.code, "组织编码") }
      ];
    };

    return (
      <Panel className="system-management-view org-management-view">
        <DataTable
          data={orgRows}
          rowKey={(item) => item.id}
          pageSize={12}
          autoFilters={false}
          searchPlaceholder="搜索组织 / 编码 / 站点"
          searchText={(item) => organizationRowSearchText(item, org)}
          onRefresh={refreshData}
          rowContextMenu={orgRowContextMenu}
          headerLeftAction={(
            <ActionGroup>
              <ActionDialog
                id="org-company-create"
                title="新增公司"
                buttonLabel="新增公司"
                triggerIcon={<Plus size={13} />}
                triggerVariant="primary"
                showTrigger={false}
                disabled={!canWriteOrg || actionBusy !== ""}
                onOpen={resetOrgForm}
              >
                <SystemForm onSubmit={handleCreateCompany}>
                  <Field label="名称"><TextInput value={orgForm.companyName} onChange={(event) => setOrgForm({ ...orgForm, companyName: event.target.value })} /></Field>
                  <Field label="编码"><TextInput value={orgForm.companyCode} onChange={(event) => setOrgForm({ ...orgForm, companyCode: event.target.value })} /></Field>
                  <Field label="上级">
                    <SelectInput value={orgForm.companyParentId} onChange={(event) => setOrgForm({ ...orgForm, companyParentId: event.target.value })}>
                      <option value="0">无</option>
                      {companyOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                    </SelectInput>
                  </Field>
                  <Field label="层级">
                    <SelectInput value={orgForm.companyLevel} onChange={(event) => setOrgForm({ ...orgForm, companyLevel: event.target.value })}>
                      {companyLevelOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                    </SelectInput>
                  </Field>
                  <Field label="区域"><TextInput value={orgForm.companyRegion} onChange={(event) => setOrgForm({ ...orgForm, companyRegion: event.target.value })} /></Field>
                  <FormActions spanAll>
                    <UiButton onClick={resetOrgForm} disabled={actionBusy !== ""}>重置</UiButton>
                    <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== ""}>新增公司</UiButton>
                  </FormActions>
                </SystemForm>
              </ActionDialog>
              <ActionDialog
                id="org-department-create"
                title={fieldNumber(orgForm.departmentParentId) ? "新增子部门" : "新增直属部门"}
                buttonLabel="新增直属部门"
                triggerIcon={<Plus size={13} />}
                showTrigger={false}
                disabled={!canWriteOrg || actionBusy !== "" || !companyOptions.length}
                onOpen={resetOrgForm}
              >
                <SystemForm onSubmit={handleCreateDepartment}>
                  <Field label="公司">
                    <SelectInput value={orgForm.departmentCompanyId} onChange={(event) => setOrgForm({ ...orgForm, departmentCompanyId: event.target.value, departmentParentId: "0" })}>
                      {companyOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                    </SelectInput>
                  </Field>
                  <Field label="上级部门">
                    <SelectInput value={orgForm.departmentParentId} onChange={(event) => setOrgForm({ ...orgForm, departmentParentId: event.target.value })}>
                      <option value="0">公司直属</option>
                      {departmentOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                    </SelectInput>
                  </Field>
                  <Field label="名称"><TextInput value={orgForm.departmentName} onChange={(event) => setOrgForm({ ...orgForm, departmentName: event.target.value })} /></Field>
                  <Field label="编码"><TextInput value={orgForm.departmentCode} onChange={(event) => setOrgForm({ ...orgForm, departmentCode: event.target.value })} /></Field>
                  <FormActions spanAll>
                    <UiButton onClick={resetOrgForm} disabled={actionBusy !== ""}>重置</UiButton>
                    <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== "" || !companyOptions.length}>
                      {fieldNumber(orgForm.departmentParentId) ? "新增子部门" : "新增直属部门"}
                    </UiButton>
                  </FormActions>
                </SystemForm>
              </ActionDialog>
            </ActionGroup>
          )}
          columns={[
            { key: "kind", title: "类型", width: "110px", render: (item) => orgLevelLabel(item.kind) },
            {
              key: "name",
              title: "名称",
              render: (item) => {
                const hasChildren = Boolean(childMap.get(item.id)?.length);
                const collapsed = collapsedOrgNodeIds.has(item.id);
                return (
                  <div className="org-table-name" style={{ paddingLeft: `${Math.max(0, item.depth) * 18}px` }}>
                    <button
                      className={`org-node-toggle${hasChildren ? "" : " is-empty"}`}
                      type="button"
                      disabled={!hasChildren}
                      aria-label={collapsed ? "展开下级组织" : "折叠下级组织"}
                      onClick={(event) => {
                        event.stopPropagation();
                        if (hasChildren) toggleOrgNodeCollapsed(item.id);
                      }}
                    >
                      {hasChildren ? (collapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />) : null}
                    </button>
                    <span className="org-node-icon compact">{orgNodeIcon(item)}</span>
                    <span className="org-table-label">
                      <b>{item.name}</b>
                      <span className="block-text muted">{item.code || "-"}</span>
                    </span>
                  </div>
                );
              }
            },
            { key: "relation", title: "归属 / 区域", render: (item) => orgNodeRelation(item, org) },
            { key: "children", title: "下级", render: (item) => childMap.get(item.id)?.length || 0 },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status || "active"} /> },
            {
              key: "actions",
              title: "操作",
              width: "120px",
              render: (item) => orgNodeStatusAction(item) || <span className="muted">-</span>
            }
          ]}
        />
      </Panel>
    );
  }

  function renderMenuManagement() {
    const allMenuItems = [...menuItems].sort((a, b) => a.sort - b.sort);
    const canWriteMenuLabels = hasPermission(currentPermissions, "system:write");
    const menuKeyword = menuFilters.keyword.trim().toLowerCase();
    const groupOptions = Array.from(new Map(allMenuItems.map((item) => [item.group, item.groupLabel])));
    const permissionOptions = Array.from(new Set(allMenuItems.map((item) => item.permissionMark).filter(Boolean))).sort();
    const orderedItems = allMenuItems.filter((item) => {
      const matchesKeyword = !menuKeyword || [
        item.label,
        item.name,
        item.path,
        item.groupLabel,
        item.permissionMark,
        item.icon
      ].join(" ").toLowerCase().includes(menuKeyword);
      const matchesPermission = menuFilters.permission === "all" || item.permissionMark === menuFilters.permission;
      const matchesCache = menuFilters.cache === "all"
        || (menuFilters.cache === "noCache" && item.noCache)
        || (menuFilters.cache === "cache" && !item.noCache)
        || (menuFilters.cache === "affix" && item.affix)
        || (menuFilters.cache === "breadcrumb" && item.breadcrumb);
      return matchesKeyword && matchesPermission && matchesCache;
    });
    const topMenuCards = groupOptions.map(([group, label]) => {
      const items = allMenuItems.filter((item) => item.group === group);
      const visibleItems = orderedItems.filter((item) => item.group === group);
      return {
        group,
        label,
        visibleItems,
        totalCount: items.length,
        visibleCount: visibleItems.length,
        permissionCount: new Set(items.map((item) => item.permissionMark).filter(Boolean)).size,
        affixCount: items.filter((item) => item.affix).length
      };
    }).filter((card) => card.visibleCount > 0);
    function openMenuLabelContextMenu(event: ReactMouseEvent, key: string, currentLabel: string, title: string) {
      event.preventDefault();
      event.stopPropagation();
      setMenuLabelMenu({ key, currentLabel, title, x: event.clientX, y: event.clientY });
    }
    const menuLabelContextItems: ContextMenuItem[] = menuLabelMenu ? [
      {
        key: "rename-menu-label",
        label: "修改名称",
        icon: <Pencil size={14} />,
        disabled: !canWriteMenuLabels || actionBusy !== "",
        onSelect: () => openMenuLabelDialog(menuLabelMenu.key, menuLabelMenu.currentLabel, menuLabelMenu.title)
      },
      {
        key: "reset-menu-label",
        label: "恢复默认名称",
        icon: <RefreshCw size={14} />,
        disabled: !canWriteMenuLabels || actionBusy !== "",
        onSelect: () => handleResetMenuLabel(menuLabelMenu.key)
      }
    ] : [];

    return (
      <Panel className="system-management-view menu-management-view">
        <SplitRow className="menu-filter-row">
          <div>
            <h3>菜单配置</h3>
          </div>
          <div className="menu-filter-strip">
            <IconField className="compact-field" icon={<Search size={14} />} label="菜单 / 路由 / 权限">
              <TextInput
                className="compact-input"
                value={menuFilters.keyword}
                onChange={(event) => setMenuFilters((value) => ({ ...value, keyword: event.target.value }))}
                placeholder="搜索"
              />
            </IconField>
            <SelectInput className="compact-select wide" value={menuFilters.permission} onChange={(event) => setMenuFilters((value) => ({ ...value, permission: event.target.value }))}>
              <option value="all">全部权限</option>
              {permissionOptions.map((permission) => <option key={permission} value={permission}>{permission}</option>)}
            </SelectInput>
            <SelectInput className="compact-select" value={menuFilters.cache} onChange={(event) => setMenuFilters((value) => ({ ...value, cache: event.target.value }))}>
              <option value="all">全部</option>
              <option value="cache">缓存</option>
              <option value="noCache">不缓存</option>
              <option value="affix">固定</option>
              <option value="breadcrumb">面包屑</option>
            </SelectInput>
          </div>
        </SplitRow>

        <div className="menu-top-card-grid" aria-label="顶级菜单配置">
          {topMenuCards.map((card) => (
            <section className="menu-top-card menu-top-menu-card" key={card.group}>
              <div
                className="menu-top-card-header"
                onContextMenu={(event) => openMenuLabelContextMenu(event, `group:${card.group}`, card.label, "修改菜单组名称")}
              >
                <span className="menu-top-card-main">
                  <b>{card.label}</b>
                  <span>{card.visibleCount}/{card.totalCount} 项</span>
                </span>
                <span className="menu-top-card-meta">
                  <span>{card.permissionCount} 权限</span>
                  {card.affixCount ? <span>{card.affixCount} 固定</span> : null}
                </span>
              </div>
              <div className="menu-top-card-items">
                {card.visibleItems.map((item) => (
                  <div
                    className="menu-top-card-item"
                    key={item.key}
                    onContextMenu={(event) => openMenuLabelContextMenu(event, item.key, item.label, "修改菜单名称")}
                  >
                    <span className="menu-top-card-item-index">{item.sort}</span>
                    <span className="menu-top-card-item-main">
                      <b>{item.label}</b>
                      <span>{item.name} / {item.path}</span>
                    </span>
                    <span className="menu-top-card-item-meta">
                      <span>{item.permissionMark}</span>
                      {item.noCache ? <span>no-cache</span> : null}
                      {item.affix ? <span>affix</span> : null}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ))}
          {!topMenuCards.length ? <div className="menu-tree-empty menu-top-card-empty">暂无匹配菜单</div> : null}
        </div>
        {menuLabelMenu ? (
          <ContextMenu
            items={menuLabelContextItems}
            position={{ x: menuLabelMenu.x, y: menuLabelMenu.y }}
            label="菜单配置操作"
            onClose={() => setMenuLabelMenu(null)}
          />
        ) : null}
        <Dialog
          open={Boolean(menuLabelDialog)}
          title={menuLabelDialog?.title || "修改菜单名称"}
          description="只修改显示名称，不新增、删除或调整菜单路由。"
          closeDisabled={actionBusy !== ""}
          feedback={actionError || undefined}
          feedbackTone="error"
          onClose={() => setMenuLabelDialog(null)}
        >
          <DialogForm onSubmit={handleSaveMenuLabel}>
            <Field label="当前名称">
              <TextInput value={menuLabelDialog?.currentLabel || ""} disabled />
            </Field>
            <Field label="显示名称">
              <TextInput
                value={menuLabelForm.label}
                onChange={(event) => setMenuLabelForm({ label: event.target.value })}
                autoFocus
                maxLength={32}
              />
            </Field>
            <FormActions>
              <UiButton type="button" disabled={actionBusy !== ""} onClick={() => setMenuLabelDialog(null)}>取消</UiButton>
              <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !menuLabelForm.label.trim()}>
                保存名称
              </UiButton>
            </FormActions>
          </DialogForm>
        </Dialog>
      </Panel>
    );
  }

  function renderDictionaryManagement() {
    const dictionaries = allDictionaries().slice().sort((a, b) => a.type.localeCompare(b.type) || a.sort - b.sort || a.code.localeCompare(b.code));
    const dictionaryKeyword = dictionaryFilters.keyword.trim().toLowerCase();
    const typeOptions = Array.from(new Set([
      ...dictionaryTypePresets.map((item) => item.type),
      ...dictionaries.map((item) => item.type)
    ])).filter(Boolean).sort((a, b) => dictionaryTypeLabel(a).localeCompare(dictionaryTypeLabel(b)));
    const requestedType = dictionaryFilters.type !== "all" ? dictionaryFilters.type : "";
    const selectedDictionaryType = typeOptions.includes(requestedType) ? requestedType : typeOptions[0] || "product_line";
    const selectedTypeKey = selectedDictionaryType.replace(/[^a-zA-Z0-9_-]/g, "-");
    const dictionaryTypeRows = typeOptions.map((type) => {
      const entries = dictionaries.filter((item) => item.type === type);
      return {
        type,
        label: dictionaryTypeLabel(type),
        entries,
        activeCount: entries.filter((item) => item.status === "active").length,
        disabledCount: entries.filter((item) => item.status === "disabled").length
      };
    });
    const selectedTypeRow = dictionaryTypeRows.find((item) => item.type === selectedDictionaryType);
    const selectedTypeEntries = selectedTypeRow?.entries || [];
    const filteredDictionaries = selectedTypeEntries.filter((item) => {
      const matchesKeyword = !dictionaryKeyword || [
        item.code,
        item.label,
        item.status
      ].join(" ").toLowerCase().includes(dictionaryKeyword);
      const matchesStatus = dictionaryFilters.status === "all" || item.status === dictionaryFilters.status;
      return matchesKeyword && matchesStatus;
    });
    const categoryRows = dictionaryTypeRows.filter((item) => {
      const matchesKeyword = !dictionaryKeyword || [
        item.type,
        item.label,
        ...item.entries.map((entry) => `${entry.code} ${entry.label} ${entry.status}`)
      ].join(" ").toLowerCase().includes(dictionaryKeyword);
      const matchesStatus = dictionaryFilters.status === "all" || item.entries.some((entry) => entry.status === dictionaryFilters.status);
      return matchesKeyword && matchesStatus;
    });
    const enabledCount = dictionaries.filter((item) => item.status === "active").length;
    const disabledCount = dictionaries.filter((item) => item.status === "disabled").length;
    const configStatusOptions = dictionaryOptions("config_status");
    const isDictionaryDetail = dictionaryFilters.type !== "all" && Boolean(selectedTypeRow);

    function openDictionaryType(type: string) {
      setEditingDictionaryId(null);
      setDictionaryFilters({ keyword: "", type, status: "all" });
    }

    function closeDictionaryType() {
      setEditingDictionaryId(null);
      setDictionaryFilters({ keyword: "", type: "all", status: "all" });
    }

    function dictionaryEditorForm(lockType = false) {
      return (
        <SystemForm onSubmit={handleSaveDictionary}>
          <Field label="字典类型">
            <TextInput
              list="dictionary-type-options"
              disabled={lockType}
              value={dictionaryForm.type}
              onChange={(event) => setDictionaryForm({ ...dictionaryForm, type: event.target.value })}
            />
          </Field>
          <datalist id="dictionary-type-options">
            {typeOptions.map((type) => <option key={type} value={type}>{dictionaryTypeLabel(type)}</option>)}
          </datalist>
          <Field label="类型编码"><TextInput value={dictionaryForm.code} onChange={(event) => setDictionaryForm({ ...dictionaryForm, code: event.target.value })} /></Field>
          <Field label="显示名称"><TextInput value={dictionaryForm.label} onChange={(event) => setDictionaryForm({ ...dictionaryForm, label: event.target.value })} /></Field>
          <Field label="排序"><TextInput type="number" value={dictionaryForm.sort} onChange={(event) => setDictionaryForm({ ...dictionaryForm, sort: event.target.value })} /></Field>
          <Field label="状态">
            <SelectInput value={dictionaryForm.status} onChange={(event) => setDictionaryForm({ ...dictionaryForm, status: event.target.value })}>
              {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
            </SelectInput>
          </Field>
          {!lockType ? (
            <ChipList className="span-all dictionary-type-presets">
              {dictionaryTypePresets.map((preset) => (
                <ChipButton
                  className="permission-chip-button"
                  key={preset.type}
                  onClick={() => setDictionaryForm((form) => ({
                    ...form,
                    type: preset.type,
                    sort: form.type === preset.type ? form.sort : nextDictionarySort(preset.type)
                  }))}
                >
                  {preset.label}
                </ChipButton>
              ))}
            </ChipList>
          ) : null}
          <FormActions spanAll>
            <UiButton onClick={() => resetDictionaryForm(lockType ? selectedDictionaryType : dictionaryForm.type)} disabled={actionBusy !== ""}>重置</UiButton>
            <UiButton variant="primary" type="submit" icon={<Plus size={14} />} disabled={actionBusy !== ""}>{editingDictionaryId ? "保存字典项" : "新增字典项"}</UiButton>
          </FormActions>
        </SystemForm>
      );
    }

    if (!isDictionaryDetail || !selectedTypeRow) {
      return (
        <Panel className="system-management-view dictionary-management-view">
          <div className="dictionary-overview">
            <SectionHeader className="panel-head-compact dictionary-overview-head">
              <div>
                <b>字典大类</b>
                <span>{typeOptions.length} 类 / {dictionaries.length} 项</span>
              </div>
              <ActionGroup>
                <ActionDialog id="dictionary-create-type" title="新增字典大类" buttonLabel="新增大类" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={() => resetDictionaryForm("custom_type")}>
                  {dictionaryEditorForm()}
                </ActionDialog>
                <ActionDialog id="dictionary-summary-overview" title="字典汇总" buttonLabel="汇总" triggerIcon={<ListChecks size={13} />}>
                  <MetricList compact className="system-summary-grid dictionary-summary-grid">
                    <div><span>字典类型</span><b>{typeOptions.length}</b></div>
                    <div><span>字典项</span><b>{dictionaries.length}</b></div>
                    <div><span>启用</span><b>{enabledCount}</b></div>
                    <div><span>禁用</span><b>{disabledCount}</b></div>
                  </MetricList>
                </ActionDialog>
              </ActionGroup>
            </SectionHeader>
            <MetricList compact className="dictionary-overview-summary">
              <div><span>大类</span><b>{typeOptions.length}</b></div>
              <div><span>子项</span><b>{dictionaries.length}</b></div>
              <div><span>启用</span><b>{enabledCount}</b></div>
              <div><span>禁用</span><b>{disabledCount}</b></div>
            </MetricList>
            <DataTable
              title="字典大类列表"
              data={categoryRows}
              rowKey={(item) => item.type}
              pageSize={12}
              pageSizeOptions={[12, 20, 50]}
              autoFilters={false}
              searchable={false}
              onRefresh={refreshData}
              emptyText={loading ? "加载中..." : "暂无匹配的大类"}
              rowContextMenu={buildDataTableRowContextMenu<(typeof categoryRows)[number]>({
                actions: [
                  { key: "open-type", label: "进入该字典大类", onSelect: (item) => openDictionaryType(item.type) },
                  { key: "focus-type", label: "只看该大类", onSelect: (item, helpers) => helpers.searchText(item.label) }
                ],
                copyFields: [
                  { key: "label", label: "大类名称", value: (item) => item.label },
                  { key: "type", label: "大类编码", value: (item) => item.type },
                  { key: "total", label: "子项数量", value: (item) => item.entries.length },
                  { key: "status", label: "启停数量", value: (item) => `启用 ${item.activeCount} / 禁用 ${item.disabledCount}` }
                ]
              })}
              headerLeftAction={(
                <div className="menu-filter-strip dictionary-filter-strip">
                  <IconField className="compact-field" icon={<Search size={14} />} label="大类 / 子项">
                    <TextInput
                      className="compact-input"
                      value={dictionaryFilters.keyword}
                      onChange={(event) => setDictionaryFilters((value) => ({ ...value, keyword: event.target.value }))}
                      placeholder="搜索大类 / 子项"
                    />
                  </IconField>
                  <SelectInput className="compact-select" value={dictionaryFilters.status} onChange={(event) => setDictionaryFilters((value) => ({ ...value, status: event.target.value }))}>
                    <option value="all">全部状态</option>
                    {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                  </SelectInput>
                </div>
              )}
              columns={[
                {
                  key: "category",
                  title: "字典大类",
                  render: (item) => (
                    <button type="button" className="dictionary-category-link" onClick={() => openDictionaryType(item.type)}>
                      <b>{item.label}</b>
                      <span className="block-text muted">{item.type}</span>
                    </button>
                  )
                },
                { key: "total", title: "子项", width: "90px", render: (item) => item.entries.length },
                {
                  key: "status-count",
                  title: "子项状态",
                  width: "180px",
                  render: (item) => (
                    <span className="dictionary-category-status">
                      <span>启用 {item.activeCount}</span>
                      <span>禁用 {item.disabledCount}</span>
                    </span>
                  )
                },
                {
                  key: "status",
                  title: "状态",
                  width: "110px",
                  render: (item) => <StatusChip value={item.disabledCount ? "warning" : "active"} />
                },
                {
                  key: "samples",
                  title: "全部子项",
                  render: (item) => item.entries.length ? (
                    <span className="dictionary-entry-preview-list">
                      {item.entries.map((entry) => (
                        <span key={entry.id} title={`${entry.label} / ${entry.code}`}>
                          <b>{entry.label}</b>
                          <small>{entry.code}</small>
                        </span>
                      ))}
                    </span>
                  ) : <span className="block-text muted">暂无子项</span>
                },
                {
                  key: "actions",
                  title: "操作",
                  width: "130px",
                  render: (item) => (
                    <UiButton variant="primary" onClick={() => openDictionaryType(item.type)}>进入子项</UiButton>
                  )
                }
              ]}
            />
          </div>
        </Panel>
      );
    }

    return (
      <Panel className="system-management-view dictionary-management-view">
        <div className="dictionary-detail-view">
          <SectionHeader className="panel-head-compact dictionary-detail-head">
            <div className="dictionary-detail-title">
              <UiButton className="dictionary-back-button" icon={<ArrowLeft size={14} />} onClick={closeDictionaryType}>返回大类</UiButton>
              <span>
                <b>{selectedTypeRow.label}</b>
                <small>{selectedDictionaryType} / {selectedTypeEntries.length} 项</small>
              </span>
            </div>
            <ActionGroup>
              <ActionDialog id={`dictionary-create-entry-${selectedTypeKey}`} title="新增字典子项" buttonLabel="新增子项" triggerIcon={<Plus size={13} />} triggerVariant="primary" onOpen={() => resetDictionaryForm(selectedDictionaryType)}>
                {dictionaryEditorForm(true)}
              </ActionDialog>
              <ActionDialog id="dictionary-summary-detail" title="字典汇总" buttonLabel="汇总" triggerIcon={<ListChecks size={13} />}>
                <MetricList compact className="system-summary-grid dictionary-summary-grid">
                  <div><span>字典类型</span><b>{typeOptions.length}</b></div>
                  <div><span>字典项</span><b>{dictionaries.length}</b></div>
                  <div><span>启用</span><b>{enabledCount}</b></div>
                  <div><span>禁用</span><b>{disabledCount}</b></div>
                </MetricList>
              </ActionDialog>
            </ActionGroup>
          </SectionHeader>
          <MetricList compact className="dictionary-detail-summary">
            <div><span>当前子项</span><b>{selectedTypeEntries.length}</b></div>
            <div><span>启用</span><b>{selectedTypeRow.activeCount}</b></div>
            <div><span>禁用</span><b>{selectedTypeRow.disabledCount}</b></div>
          </MetricList>
          <DataTable
            data={filteredDictionaries}
            rowKey={(item) => item.id}
            pageSize={12}
            pageSizeOptions={[12, 20, 50]}
            onRefresh={refreshData}
            emptyText={loading ? "加载中..." : "暂无字典子项"}
            rowContextMenu={buildDataTableRowContextMenu<DataDictionary>({
              actions: [
                { key: "focus-status", label: "只看该状态", onSelect: (item, helpers) => helpers.searchText(item.status) },
                { key: "focus-code", label: "只看该编码", onSelect: (item, helpers) => helpers.searchText(item.code) }
              ],
              copyFields: [
                { key: "type", label: "字典类型", value: (item) => item.type },
                { key: "code", label: "字典编码", value: (item) => item.code },
                { key: "label", label: "显示名称", value: (item) => item.label },
                { key: "status", label: "状态", value: (item) => item.status }
              ]
            })}
            headerLeftAction={(
              <div className="menu-filter-strip dictionary-filter-strip">
                <IconField className="compact-field" icon={<Search size={14} />} label="编码 / 名称">
                  <TextInput
                    className="compact-input"
                    value={dictionaryFilters.keyword}
                    onChange={(event) => setDictionaryFilters((value) => ({ ...value, keyword: event.target.value }))}
                    placeholder="搜索当前大类"
                  />
                </IconField>
                <SelectInput className="compact-select" value={dictionaryFilters.status} onChange={(event) => setDictionaryFilters((value) => ({ ...value, status: event.target.value }))}>
                  <option value="all">全部状态</option>
                  {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
                </SelectInput>
              </div>
            )}
            columns={[
              {
                key: "entry",
                title: "字典子项",
                render: (item) => <><b>{item.label}</b><span className="block-text muted">{item.code}</span></>
              },
              { key: "sort", title: "排序", width: "90px", render: (item) => item.sort },
              { key: "status", title: "状态", width: "110px", render: (item) => <StatusChip value={item.status} /> },
              {
                key: "actions",
                title: "操作",
                width: "160px",
                render: (item) => (
                  <ActionDialog id={`dictionary-${item.id}`} title="编辑字典子项" onOpen={() => startDictionaryEdit(item)}>
                    {dictionaryEditorForm(true)}
                    <ActionGroup>
                      <UiButton
                        disabled={actionBusy !== ""}
                        variant={item.status === "active" ? "danger" : "soft"}
                        onClick={() => handleDictionaryStatus(item, item.status === "active" ? "disabled" : "active")}
                      >
                        {item.status === "active" ? "禁用" : "启用"}
                      </UiButton>
                      {item.status !== "draft" ? (
                        <UiButton disabled={actionBusy !== ""} onClick={() => handleDictionaryStatus(item, "draft")}>转草稿</UiButton>
                      ) : null}
                      <UiButton variant="danger" disabled={actionBusy !== "" || item.status === "active"} onClick={() => handleDeleteDictionary(item)}>删除</UiButton>
                    </ActionGroup>
                  </ActionDialog>
                )
              }
            ]}
          />
        </div>
      </Panel>
    );
  }

  function renderUserManagement() {
    const users = data.systemUsers.length ? data.systemUsers : list(data.system?.security?.users);
    const roles = availableRoles();
    const accountStatusOptions = dictionaryOptions("account_status");
    function userEditorForm() {
      return (
        <SystemForm onSubmit={handleSaveUser}>
          <Field label="用户名"><TextInput value={userForm.username} onChange={(event) => setUserForm({ ...userForm, username: event.target.value })} /></Field>
          <Field label="显示名"><TextInput value={userForm.displayName} onChange={(event) => setUserForm({ ...userForm, displayName: event.target.value })} /></Field>
          <Field label="密码"><TextInput type="password" value={userForm.password} required={!editingUserId} onChange={(event) => setUserForm({ ...userForm, password: event.target.value })} /></Field>
          <Field label="角色">
            <SelectInput value={userForm.roleCode} onChange={(event) => setUserForm({ ...userForm, roleCode: event.target.value })}>
              {roles.map((item) => <option key={item.code} value={item.code}>{item.name} / {item.code}</option>)}
            </SelectInput>
          </Field>
          <Field label="状态">
            <SelectInput value={userForm.status} onChange={(event) => setUserForm({ ...userForm, status: event.target.value })}>
              {accountStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="公司">
            <SelectInput value={userForm.companyId} onChange={(event) => setUserForm({ ...userForm, companyId: event.target.value })}>
              <option value="0">不限定</option>
              {list(bootstrap?.companies).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </SelectInput>
          </Field>
          <Field label="站点">
            <SelectInput value={userForm.siteId} onChange={(event) => setUserForm({ ...userForm, siteId: event.target.value })}>
              <option value="0">不限定</option>
              {siteOptions().map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </SelectInput>
          </Field>
          <Field label="客户">
            <SelectInput value={userForm.customerId} onChange={(event) => setUserForm({ ...userForm, customerId: event.target.value })}>
              <option value="0">不限定</option>
              {list(bootstrap?.customers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </SelectInput>
          </Field>
          <Field label="司机">
            <SelectInput value={userForm.driverId} onChange={(event) => setUserForm({ ...userForm, driverId: event.target.value })}>
              <option value="0">不限定</option>
              {list(bootstrap?.drivers).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </SelectInput>
          </Field>
          <FormActions spanAll>
            <UiButton onClick={resetUserForm} disabled={actionBusy !== ""}>重置</UiButton>
            <UiButton variant="primary" type="submit" icon={<Users size={14} />} disabled={actionBusy !== "" || !roles.length}>{editingUserId ? "保存用户" : "新增用户"}</UiButton>
          </FormActions>
        </SystemForm>
      );
    }
    return (
      <Panel className="system-management-view">
        <DataTable
          data={users}
          rowKey={(item) => item.id}
          pageSize={10}
          filterPlacement="header"
          filterMultiple
          rowContextMenu={buildDataTableRowContextMenu<User>({
            actions: [
              { key: "focus-role", label: "只看该角色", onSelect: (item, helpers) => helpers.searchText(roleName(item.roleCode)) },
              { key: "focus-company", label: "只看该公司", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.companies, item.companyId)) },
              { key: "focus-site", label: "只看该站点", onSelect: (item, helpers) => helpers.searchText(nameOf(bootstrap?.sites, item.siteId)) }
            ],
            copyFields: [
              { key: "username", label: "用户名", value: (item) => item.username },
              { key: "display", label: "显示名", value: (item) => item.displayName },
              { key: "role", label: "角色", value: (item) => `${roleName(item.roleCode)} / ${item.roleCode}` },
              { key: "scope", label: "数据范围", value: (item) => `${nameOf(bootstrap?.companies, item.companyId)} / ${nameOf(bootstrap?.sites, item.siteId)}` }
            ]
          })}
          headerLeftAction={<ActionDialog id="system-user-create" title="新增用户" onOpen={resetUserForm}>{userEditorForm()}</ActionDialog>}
          columns={[
            { key: "user", title: "用户", render: (item) => <><b>{item.displayName}</b><span className="block-text muted">{item.username}</span></> },
            { key: "role", title: "角色", filterKeys: ["roleCode", "currentRole"], render: (item) => <><b>{roleName(item.roleCode)}</b><span className="block-text muted">{item.roleCode}</span></> },
            {
              key: "scope",
              title: "范围",
              filterKeys: ["companyId", "siteId", "currentSiteId"],
              filterLabels: {
                companyId: (item) => nameOf(bootstrap?.companies, item.companyId),
                siteId: (item) => nameOf(bootstrap?.sites, item.siteId),
                customerId: (item) => nameOf(bootstrap?.customers, item.customerId),
                driverId: (item) => nameOf(bootstrap?.drivers, item.driverId)
              },
              render: (item) => <><span>{nameOf(bootstrap?.companies, item.companyId)}</span><span className="block-text muted">{nameOf(bootstrap?.sites, item.siteId)}</span></>
            },
            { key: "mfa", title: "MFA", render: (item) => <StatusChip value={item.mfaEnabled ? "active" : "disabled"} /> },
            { key: "status", title: "状态", filterKeys: ["status"], render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => (
              <ActionDialog id={`system-user-${item.id}`} title="用户操作" onOpen={() => startUserEdit(item)}>
                {userEditorForm()}
                {mfaEnrollment?.user.id === item.id ? (
                  <div className="finance-action-block">
                    <b>MFA 密钥</b>
                    <span>{mfaEnrollment.secret}</span>
                    <span>{mfaEnrollment.otpauthUrl}</span>
                    <Field label="动态码">
                      <TextInput value={mfaCodes[item.id] || ""} onChange={(event) => setMfaCodes((codes) => ({ ...codes, [item.id]: event.target.value }))} />
                    </Field>
                    <UiButton variant="primary" disabled={actionBusy !== "" || !(mfaCodes[item.id] || "").trim()} onClick={() => handleEnableUserMFA(item)}>启用 MFA</UiButton>
                  </div>
                ) : null}
                <ActionGroup>
                  <UiButton disabled={actionBusy !== ""} onClick={() => handleUserStatus(item, item.status === "active" ? "disabled" : "active")}>{item.status === "active" ? "禁用" : "启用"}</UiButton>
                  {item.mfaEnabled ? <UiButton disabled={actionBusy !== ""} onClick={() => handleDisableUserMFA(item)}>关闭 MFA</UiButton> : <UiButton disabled={actionBusy !== ""} onClick={() => handleEnrollUserMFA(item)}>登记 MFA</UiButton>}
                </ActionGroup>
              </ActionDialog>
            ) }
          ]}
        />
      </Panel>
    );
  }

  function renderWorkflowManagement() {
    const definitions = workflowDefinitions();
    const tasks = workflowTasks();
    const instances = workflowInstances();
    const events = [...workflowEvents()].sort((a, b) => b.id - a.id);
    const logs = [...workflowLogs()].sort((a, b) => b.id - a.id);
    const outbox = [...workflowOutbox()].sort((a, b) => b.id - a.id);
    const subscriptions = [...workflowSubscriptions()].sort((a, b) => b.id - a.id);
    const deliveries = [...workflowDeliveries()].sort((a, b) => b.id - a.id);
    const pendingOutbox = outbox.filter((item) => item.status === "pending");
    const pendingDeliveries = deliveries.filter((item) => item.status === "pending" || item.status === "failed");
    const unresolvedEvents = events.filter(workflowEventNeedsRecovery);
    const pendingTasks = tasks.filter((item) => item.status === "pending");
    const overdueWorkflowTasks = pendingTasks.filter(workflowTaskOverdue);
    const activeInstances = instances.filter((item) => item.status === "pending");
    const catalogEvents = list(workflowOverview().catalog?.events);
    const integratedCatalogEvents = catalogEvents.filter((item) => item.integration?.status === "ready" || item.integration?.status === "event_only");
    const writeBackCatalogEvents = catalogEvents.filter((item) => item.integration?.resultPolicy === "write_back");
    const workflowDefinitionCodes = Array.from(new Set(definitions.map(workflowDefinitionGroupKey)));
    const historicalDefinitions = definitions.filter((item) => !workflowDefinitionIsActiveVersion(item));
    const currentDefinitions = definitions.filter((item) => workflowDefinitionActiveVersion(item).id === item.id);
    const deliveryIssues = deliveries.filter((item) => item.status === "failed" || item.status === "dead");
    const outboxIssues = outbox.filter((item) => item.status === "failed" || item.status === "processing");
    const draftSteps = workflowDraftSteps();
    const draftConditions = workflowConditionsFromText();
    const configStatusOptions = dictionaryOptions("config_status");
    const approvalFlows = [...list(data.system?.approvalFlows)].sort((a, b) => b.id - a.id);
    const workflowAllRows = [
      ...definitions.map((definition) => ({ id: `definition-${definition.id}`, kind: "definition" as const, definition })),
      ...events.map((event) => ({ id: `event-${event.id}`, kind: "event" as const, event })),
      ...pendingTasks.map((task) => ({ id: `task-${task.id}`, kind: "task" as const, task })),
      ...instances.map((instance) => ({ id: `instance-${instance.id}`, kind: "instance" as const, instance }))
    ];
    const workflowRows = workflowAllRows.filter(workflowRowMatchesRuntimeFilters);
    const workflowRuntimeStatusOptions = Array.from(new Set(workflowAllRows.map(workflowRowStatus).filter(Boolean))).sort();

    function workflowEventResourceLabel(event: WorkflowEvent) {
      const resource = workflowResourceLabel(event.resource);
      const target = event.resourceNo || (event.resourceId ? `#${event.resourceId}` : "-");
      return `${resource} ${target}`;
    }

    function workflowEventMatchedLabel(event: WorkflowEvent) {
      const matched = list(event.matchedDefinitions);
      return matched.length ? matched.join(" / ") : "-";
    }

    function workflowEventNeedsRecovery(event: WorkflowEvent) {
      return event.status === "failed" || event.status === "ignored";
    }

    function workflowCatalogIntegrationText(event: WorkflowCatalogEvent) {
      const integration = event.integration;
      if (!integration) return "未上报集成状态";
      const triggerText = `${list(event.triggers).length} 个触发入口`;
      const policyText = integration.resultPolicy === "event_only" ? "仅事件通知" : "结果回写";
      if (integration.resultPolicy === "event_only") return `${triggerText} / ${policyText}`;
      return `${triggerText} / ${policyText}${integration.hasResultHandler ? " / 已接业务" : " / 缺少处理器"}`;
    }

    function workflowDefinitionGroupKey(definition: WorkflowDefinition) {
      return `${definition.category || "approval"}:${definition.code}`;
    }

    function workflowDefinitionVersions(definition: WorkflowDefinition) {
      const key = workflowDefinitionGroupKey(definition);
      return definitions
        .filter((item) => workflowDefinitionGroupKey(item) === key)
        .sort((left, right) => (right.version || 0) - (left.version || 0) || right.id - left.id);
    }

    function workflowDefinitionActiveVersion(definition: WorkflowDefinition) {
      const versions = workflowDefinitionVersions(definition);
      return versions.find((item) => item.status === "active") || versions[0] || definition;
    }

    function workflowDefinitionIsActiveVersion(definition: WorkflowDefinition) {
      const active = workflowDefinitionActiveVersion(definition);
      return definition.id === active.id && definition.status === "active";
    }

    function workflowDefinitionVersionNode(definition: WorkflowDefinition) {
      const versions = workflowDefinitionVersions(definition);
      const active = workflowDefinitionActiveVersion(definition);
      return (
        <>
          <b>v{definition.version || 1} {workflowDefinitionIsActiveVersion(definition) ? "当前" : "历史"}</b>
          <span className="block-text muted">{versions.length > 1 ? `共 ${versions.length} 版 / 当前 v${active.version || 1}` : definition.category}</span>
        </>
      );
    }

    function workflowDefinitionVersionPanel(definition: WorkflowDefinition) {
      const versions = workflowDefinitionVersions(definition);
      const active = workflowDefinitionActiveVersion(definition);
      if (versions.length <= 1) return null;
      return (
        <div className="workflow-definition-versions">
          <b>版本组：{definition.code} / 当前 v{active.version || 1}</b>
          {versions.map((item) => (
            <div className="workflow-definition-version-row" key={item.id}>
              <span><b>v{item.version || 1}</b> / {statusLabel(item.status)}</span>
              <span>{item.name}</span>
              <span>{list(item.steps).map((step) => roleName(step.roleCode)).join(" → ") || "无步骤"}</span>
            </div>
          ))}
        </div>
      );
    }

    function workflowRowKindLabel(kind: (typeof workflowAllRows)[number]["kind"]) {
      const labels = {
        definition: "流程定义",
        event: "业务事件",
        task: "待办任务",
        instance: "流程实例"
      };
      return labels[kind] || kind;
    }

    function workflowRowResource(item: (typeof workflowAllRows)[number]) {
      if (item.kind === "definition") return item.definition.resource;
      if (item.kind === "event") return item.event.resource;
      if (item.kind === "task") return item.task.resource;
      return item.instance.resource;
    }

    function workflowRowResourceText(item: (typeof workflowAllRows)[number]) {
      const resource = workflowRowResource(item);
      if (item.kind === "definition") return workflowResourceLabel(resource);
      if (item.kind === "event") return workflowEventResourceLabel(item.event);
      if (item.kind === "task") return `${workflowResourceLabel(resource)} #${item.task.resourceId}`;
      return `${workflowResourceLabel(resource)} ${item.instance.resourceNo || (item.instance.resourceId ? `#${item.instance.resourceId}` : "-")}`;
    }

    function workflowRowStatus(item: (typeof workflowAllRows)[number]) {
      if (item.kind === "definition") return item.definition.status;
      if (item.kind === "event") return item.event.status;
      if (item.kind === "task") return item.task.status;
      return item.instance.status;
    }

    function workflowRowRoleCodes(item: (typeof workflowAllRows)[number]) {
      if (item.kind === "definition") return list(item.definition.steps).map((step) => step.roleCode).filter(Boolean);
      if (item.kind === "task") return [item.task.roleCode].filter(Boolean);
      if (item.kind === "instance") return [item.instance.currentRole].filter(Boolean);
      return [];
    }

    function workflowRowMatchesRuntimeIssue(item: (typeof workflowAllRows)[number]) {
      switch (workflowRuntimeFilters.issue) {
      case "running":
        return (item.kind === "task" && item.task.status === "pending") || (item.kind === "instance" && item.instance.status === "pending");
      case "overdue":
        return item.kind === "task" && workflowTaskOverdue(item.task);
      case "recovery":
        return item.kind === "event" && workflowEventNeedsRecovery(item.event);
      case "unmatched":
        return item.kind === "event" && list(item.event.matchedDefinitions).length === 0;
      case "result_failed":
        return item.kind === "instance" && workflowInstanceLogs(item.instance.id).some((log) => log.action === "result_failed");
      case "history":
        return item.kind === "definition" && !workflowDefinitionIsActiveVersion(item.definition);
      default:
        return true;
      }
    }

    function workflowRowMatchesRuntimeFilters(item: (typeof workflowAllRows)[number]) {
      if (workflowRuntimeFilters.kind !== "all" && item.kind !== workflowRuntimeFilters.kind) return false;
      if (workflowRuntimeFilters.resource !== "all" && workflowRowResource(item) !== workflowRuntimeFilters.resource) return false;
      if (workflowRuntimeFilters.status !== "all" && workflowRowStatus(item) !== workflowRuntimeFilters.status) return false;
      if (workflowRuntimeFilters.roleCode !== "all" && !workflowRowRoleCodes(item).includes(workflowRuntimeFilters.roleCode)) return false;
      return workflowRowMatchesRuntimeIssue(item);
    }

    function updateWorkflowRuntimeFilter(key: keyof typeof workflowRuntimeFilters, value: string) {
      setWorkflowRuntimeFilters((filters) => ({ ...filters, [key]: value }));
    }

    function resetWorkflowRuntimeFilters() {
      setWorkflowRuntimeFilters({ kind: "all", resource: "all", status: "all", roleCode: "all", issue: "all" });
    }

    function workflowRuntimeFilterBar() {
      return (
        <ActionGroup className="workflow-runtime-filters">
          <SelectInput value={workflowRuntimeFilters.kind} onChange={(event) => updateWorkflowRuntimeFilter("kind", event.target.value)}>
            <option value="all">全部类型</option>
            <option value="definition">流程定义</option>
            <option value="event">业务事件</option>
            <option value="task">待办任务</option>
            <option value="instance">流程实例</option>
          </SelectInput>
          <SelectInput value={workflowRuntimeFilters.resource} onChange={(event) => updateWorkflowRuntimeFilter("resource", event.target.value)}>
            <option value="all">全部对象</option>
            {workflowResourceOptionList().map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
          </SelectInput>
          <SelectInput value={workflowRuntimeFilters.status} onChange={(event) => updateWorkflowRuntimeFilter("status", event.target.value)}>
            <option value="all">全部状态</option>
            {workflowRuntimeStatusOptions.map((status) => <option key={status} value={status}>{statusLabel(status)}</option>)}
          </SelectInput>
          <SelectInput value={workflowRuntimeFilters.roleCode} onChange={(event) => updateWorkflowRuntimeFilter("roleCode", event.target.value)}>
            <option value="all">全部角色</option>
            {availableRoles().map((role) => <option key={role.code} value={role.code}>{role.name}</option>)}
          </SelectInput>
          <SelectInput value={workflowRuntimeFilters.issue} onChange={(event) => updateWorkflowRuntimeFilter("issue", event.target.value)}>
            <option value="all">全部问题</option>
            <option value="running">运行中</option>
            <option value="overdue">逾期待办</option>
            <option value="recovery">需恢复事件</option>
            <option value="unmatched">未匹配事件</option>
            <option value="result_failed">回写失败</option>
            <option value="history">历史版本</option>
          </SelectInput>
          <span className="workflow-runtime-filter-count">{workflowRows.length}/{workflowAllRows.length}</span>
          <UiButton size="sm" disabled={actionBusy !== ""} onClick={resetWorkflowRuntimeFilters}>重置</UiButton>
        </ActionGroup>
      );
    }

    function workflowActionLabel(action: string) {
      const labels: Record<string, string> = {
        instance_started: "实例启动",
        task_created: "任务生成",
        task_approved: "任务通过",
        task_rejected: "任务驳回",
        task_cancelled: "任务取消",
        instance_approved: "实例通过",
        instance_rejected: "实例驳回",
        instance_cancelled: "实例取消",
        result_applied: "结果同步",
        result_failed: "结果失败"
      };
      return labels[action] || action;
    }

    function workflowInstanceLogs(instanceId: number) {
      return logs.filter((item) => item.instanceId === instanceId).sort((a, b) => a.id - b.id);
    }

    async function replayWorkflowEvent(event: WorkflowEvent) {
      await runBusinessAction(`workflow-event-replay-${event.id}`, "工作流事件已重放", () => api.replayWorkflowEvent(event.id));
    }

    async function resolveWorkflowEvent(event: WorkflowEvent) {
      await runBusinessAction(`workflow-event-resolve-${event.id}`, "工作流事件已标记处理", () => api.resolveWorkflowEvent(event.id, workflowEventResolution));
    }

    async function cancelWorkflowInstance(id: number) {
      await runBusinessAction(`workflow-instance-cancel-${id}`, "工作流实例已取消", () => api.cancelWorkflowInstance(id, workflowCancelReason));
    }

    async function actWorkflowTask(id: number, action: "approve" | "reject") {
      await runBusinessAction(`workflow-task-${action}-${id}`, action === "approve" ? "工作流任务已通过" : "工作流任务已驳回", () => api.actWorkflowTask(id, action, approvalComment));
    }

    async function reassignWorkflowTask(id: number) {
      await runBusinessAction(`workflow-task-reassign-${id}`, "工作流任务已改派", () => api.reassignWorkflowTask(id, workflowReassignRole, workflowReassignReason));
    }

    async function escalateWorkflowTask(id: number) {
      await runBusinessAction(`workflow-task-escalate-${id}`, "工作流任务已升级", () => api.escalateWorkflowTask(id, workflowReassignRole, workflowReassignReason));
    }

    async function acknowledgeWorkflowOutbox(id: number) {
      await runBusinessAction(`workflow-outbox-ack-${id}`, "出口事件已确认", () => api.acknowledgeWorkflowOutbox(id));
    }

    async function claimWorkflowOutbox(id: number) {
      await runBusinessAction(`workflow-outbox-claim-${id}`, "出口事件已领取", () => api.claimWorkflowOutbox(id, "erp-console"));
    }

    async function failWorkflowOutbox(id: number) {
      await runBusinessAction(`workflow-outbox-fail-${id}`, "出口事件已标记失败", () => api.failWorkflowOutbox(id, "控制台标记投递失败", 5));
    }

    async function resetWorkflowOutbox(id: number) {
      await runBusinessAction(`workflow-outbox-reset-${id}`, "出口事件已重置", () => api.resetWorkflowOutbox(id));
    }

    async function resetWorkflowDelivery(id: number) {
      await runBusinessAction(`workflow-delivery-reset-${id}`, "投递记录已重置", () => api.resetWorkflowDelivery(id));
    }

    function workflowOutboxStatusText(item: WorkflowOutbox) {
      const labels: Record<string, string> = {
        pending: "待投递",
        processing: "投递中",
        failed: "失败",
        sent: "已确认"
      };
      return labels[item.status] || statusLabel(item.status);
    }

    function workflowOutboxDetail(item: WorkflowOutbox) {
      if (item.status === "processing") return `消费者 ${item.claimedBy || "-"} / ${item.claimedAt || item.updatedAt}`;
      if (item.status === "failed") return item.nextAttemptAt ? `下次重试 ${item.nextAttemptAt}` : item.lastError || "等待重试";
      if (item.status === "sent") return item.acknowledgedAt ? `确认 ${item.acknowledgedAt}` : item.updatedAt;
      return item.updatedAt || item.createdAt;
    }

    function workflowOutboxActions(item: WorkflowOutbox) {
      if (item.status === "processing") {
        return (
          <ActionGroup>
            <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => acknowledgeWorkflowOutbox(item.id)}>确认</UiButton>
            <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => failWorkflowOutbox(item.id)}>失败</UiButton>
          </ActionGroup>
        );
      }
      if (item.status === "sent") {
        return <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => resetWorkflowOutbox(item.id)}>重投</UiButton>;
      }
      if (item.status === "failed") {
        return <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => resetWorkflowOutbox(item.id)}>重试</UiButton>;
      }
      return <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => claimWorkflowOutbox(item.id)}>领取</UiButton>;
    }

    function workflowDeliveryStatusText(item: WorkflowDelivery) {
      const labels: Record<string, string> = {
        pending: "待投递",
        processing: "投递中",
        succeeded: "成功",
        failed: "失败",
        dead: "终止"
      };
      return labels[item.status] || statusLabel(item.status);
    }

    function workflowDeliveryDetail(item: WorkflowDelivery) {
      if (item.status === "succeeded") return item.completedAt || item.updatedAt;
      if (item.status === "failed") return item.nextAttemptAt ? `下次重试 ${item.nextAttemptAt}` : item.lastError || "等待重试";
      if (item.status === "dead") return item.lastError || "超过重试次数";
      return item.lastAttemptAt || item.updatedAt || item.createdAt;
    }

    function workflowDeliveryActions(item: WorkflowDelivery) {
      if (item.status === "processing") {
        return <span>-</span>;
      }
      if (item.status === "succeeded" || item.status === "dead") {
        return <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => resetWorkflowDelivery(item.id)}>重置</UiButton>;
      }
      if (item.status === "failed") {
        return (
          <ActionGroup>
            <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => dispatchWorkflowDelivery(item.id)}>投递</UiButton>
            <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => resetWorkflowDelivery(item.id)}>重置</UiButton>
          </ActionGroup>
        );
      }
      return <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => dispatchWorkflowDelivery(item.id)}>投递</UiButton>;
    }

    function workflowSubscriptionFormPanel() {
      const selectedOutboxEvent = selectedWorkflowOutboxEvent();
      const payloadFields = list(selectedOutboxEvent?.payloadFields);
      return (
        <SystemForm onSubmit={handleSaveWorkflowSubscription}>
          <Field label="名称"><TextInput value={workflowSubscriptionForm.name} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, name: event.target.value })} /></Field>
          <Field label="出口事件">
            <SelectInput value={workflowSubscriptionForm.eventType} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, eventType: event.target.value })}>
              {workflowOutboxEventOptionList().map((item) => <option key={item.eventType} value={item.eventType}>{item.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="Webhook 地址"><TextInput value={workflowSubscriptionForm.endpoint} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, endpoint: event.target.value })} placeholder="https://erp.customer.example/workflow-webhook" /></Field>
          <Field label="状态">
            <SelectInput value={workflowSubscriptionForm.status} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, status: event.target.value })}>
              {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
            </SelectInput>
          </Field>
          {selectedOutboxEvent ? (
            <div className="workflow-template-hint span-all">
              <b>{selectedOutboxEvent.label}</b>
              <span>{selectedOutboxEvent.description || selectedOutboxEvent.eventType}</span>
              {payloadFields.length ? <span>字段：{payloadFields.map((field) => field.label || field.key).join("、")}</span> : null}
            </div>
          ) : null}
          <details className="workflow-advanced span-all">
            <summary>高级设置</summary>
            <div className="workflow-inline-fields">
              <Field label="订阅编码"><TextInput value={workflowSubscriptionForm.code} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, code: event.target.value })} /></Field>
              <Field label="业务对象">
                <SelectInput value={workflowSubscriptionForm.resource} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, resource: event.target.value })}>
                  <option value="">全部</option>
                  {workflowResourceOptionList().map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
                </SelectInput>
              </Field>
              <Field label="流程">
                <SelectInput value={workflowSubscriptionForm.definitionCode} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, definitionCode: event.target.value })}>
                  <option value="">全部流程</option>
                  {definitions.map((item) => <option key={item.id} value={item.code}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="重试次数"><TextInput value={workflowSubscriptionForm.retryLimit} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, retryLimit: event.target.value })} /></Field>
              <Field label="超时秒数"><TextInput value={workflowSubscriptionForm.timeoutSeconds} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, timeoutSeconds: event.target.value })} /></Field>
              <Field label="签名密钥"><TextInput value={workflowSubscriptionForm.secret} onChange={(event) => setWorkflowSubscriptionForm({ ...workflowSubscriptionForm, secret: event.target.value })} /></Field>
            </div>
          </details>
          <FormActions spanAll>
            {editingWorkflowSubscriptionId ? <UiButton type="button" onClick={resetWorkflowSubscriptionForm} disabled={actionBusy !== ""}>重置</UiButton> : null}
            <UiButton variant="primary" type="submit" disabled={actionBusy !== ""}>{editingWorkflowSubscriptionId ? "保存通知" : "新增通知"}</UiButton>
          </FormActions>
        </SystemForm>
      );
    }

    function workflowEventPreviewPanel() {
      if (!workflowEventPreview) return null;
      return (
        <div className="workflow-preview span-all">
          <div className="workflow-preview-summary">
            <b>{workflowEventPreview.willStart > 0 ? `将启动 ${workflowEventPreview.willStart} 个流程` : "不会启动新流程"}</b>
            <span>{workflowEventPreview.matchedDefinitions.length ? workflowEventPreview.matchedDefinitions.join(" / ") : "未匹配流程"}</span>
          </div>
          {list(workflowEventPreview.warnings).map((warning) => <span className="workflow-preview-warning" key={warning}>{warning}</span>)}
          {workflowEventPreview.matches.length ? (
            <div className="workflow-list-table workflow-preview-table">
              <div className="workflow-list-head">
                <span>流程</span>
                <span>步骤</span>
                <span>首节点</span>
                <span>结果</span>
              </div>
              {workflowEventPreview.matches.map((match) => (
                <div className="workflow-list-row" key={match.definitionCode}>
                  <b>{match.definitionName || match.definitionCode}</b>
                  <span>{match.stepCount}</span>
                  <span>{match.firstRole ? roleName(match.firstRole) : "-"}</span>
                  <span className={match.willStart ? "workflow-preview-ok" : "workflow-preview-muted"}>{match.willStart ? "可启动" : match.reason || "不会启动"}</span>
                </div>
              ))}
            </div>
          ) : null}
        </div>
      );
    }

    function workflowTaskOverdue(task: { status: string; dueAt?: string }) {
      if (task.status !== "pending" || !task.dueAt) return false;
      const dueAt = Date.parse(task.dueAt.replace(" ", "T"));
      return Number.isFinite(dueAt) && dueAt < Date.now();
    }

    function workflowTaskDueLabel(task: { status: string; dueAt?: string }) {
      if (!task.dueAt) return "无 SLA";
      return workflowTaskOverdue(task) ? `超时 ${task.dueAt}` : `到期 ${task.dueAt}`;
    }

    function workflowEventPublisherForm() {
      return (
        <SystemForm onSubmit={handlePublishWorkflowEvent}>
          <Field label="事件类型"><TextInput value={workflowEventForm.eventType} onChange={(event) => updateWorkflowEventForm({ eventType: event.target.value })} /></Field>
          <Field label="业务对象">
            <SelectInput value={workflowEventForm.resource} onChange={(event) => updateWorkflowEventForm({ resource: event.target.value, eventType: workflowEventForm.eventType || `${event.target.value}.submitted` })}>
              {workflowResourceOptionList().map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="事件来源"><TextInput value={workflowEventForm.source} onChange={(event) => updateWorkflowEventForm({ source: event.target.value })} /></Field>
          <Field label="发起人"><TextInput value={workflowEventForm.actor} onChange={(event) => updateWorkflowEventForm({ actor: event.target.value })} /></Field>
          <Field label="去重键"><TextInput value={workflowEventForm.eventKey} onChange={(event) => updateWorkflowEventForm({ eventKey: event.target.value })} /></Field>
          <Field label="业务 ID"><TextInput value={workflowEventForm.resourceId} onChange={(event) => updateWorkflowEventForm({ resourceId: event.target.value })} /></Field>
          <Field label="业务编号"><TextInput value={workflowEventForm.resourceNo} onChange={(event) => updateWorkflowEventForm({ resourceNo: event.target.value })} /></Field>
          <Field label="标题"><TextInput value={workflowEventForm.title} onChange={(event) => updateWorkflowEventForm({ title: event.target.value })} /></Field>
          <Field label="原因"><TextInput value={workflowEventForm.reason} onChange={(event) => updateWorkflowEventForm({ reason: event.target.value })} /></Field>
          <Field className="span-all" label="变量">
            <TextAreaInput value={workflowEventForm.variables} onChange={(event) => updateWorkflowEventForm({ variables: event.target.value })} />
          </Field>
          {workflowEventPreviewPanel()}
          <FormActions spanAll>
            <UiButton type="button" onClick={handlePreviewWorkflowEvent} disabled={actionBusy !== ""}>预览匹配</UiButton>
            <UiButton variant="primary" type="submit" icon={<PlayCircle size={14} />} disabled={actionBusy !== ""}>发布事件</UiButton>
          </FormActions>
        </SystemForm>
      );
    }

    function workflowEditorForm() {
      const currentPreset = selectedWorkflowPreset();
      const presetVariables = list(currentPreset?.variables);
      const presetTriggers = list(currentPreset?.triggers);
      return (
        <WorkflowForm onSubmit={handleSaveWorkflow}>
          <div className="workflow-inline-fields">
            <Field spanAll label="业务场景">
              <SelectInput value={workflowPresetCode} onChange={(event) => selectWorkflowPreset(event.target.value)}>
                {workflowTemplatePresets().map((preset) => (
                  <option key={preset.code} value={preset.code}>{preset.label}</option>
                ))}
              </SelectInput>
            </Field>
            {currentPreset ? (
              <div className="workflow-template-hint span-all">
                <b>{currentPreset.label}</b>
                <span>{currentPreset.description || `${workflowResourceLabel(currentPreset.resource)} / ${currentPreset.eventType}`}</span>
              </div>
            ) : null}
            <Field label="流程名称"><TextInput value={workflowForm.name} onChange={(event) => setWorkflowForm({ ...workflowForm, name: event.target.value })} /></Field>
            <Field label="业务对象">
              <SelectInput value={workflowForm.resource} onChange={(event) => setWorkflowForm({ ...workflowForm, resource: event.target.value })}>
                {workflowResourceOptionList().map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="状态">
              <SelectInput value={workflowForm.status} onChange={(event) => setWorkflowForm({ ...workflowForm, status: event.target.value })}>
                {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
              </SelectInput>
            </Field>
          </div>
          <div className="workflow-list-table workflow-step-table">
            <div className="workflow-list-head">
              <span>步骤</span>
              <span>角色</span>
              <span>名称</span>
              <span>SLA小时</span>
              <span>操作</span>
            </div>
            {draftSteps.map((step, index) => (
              <div className="workflow-list-row" key={`${step.seq}-${index}`}>
                <span className="step-index">{index + 1}</span>
                <SelectInput value={step.roleCode} onChange={(event) => setWorkflowStep(index, { roleCode: event.target.value })}>
                  {availableRoles().map((role) => <option key={role.code} value={role.code}>{role.name}</option>)}
                </SelectInput>
                <TextInput value={step.name} onChange={(event) => setWorkflowStep(index, { name: event.target.value })} />
                <TextInput value={String(step.slaHours || "")} onChange={(event) => setWorkflowStep(index, { slaHours: Math.max(0, fieldNumber(event.target.value)) })} />
                <UiButton disabled={draftSteps.length <= 1} onClick={() => removeWorkflowStep(index)}>删除</UiButton>
              </div>
            ))}
            <UiButton className="workflow-add-row" variant="ghost" icon={<Plus size={14} />} onClick={() => addWorkflowStep()}>
              添加审批步骤
            </UiButton>
          </div>
          <details className="workflow-advanced">
            <summary>高级设置</summary>
            <div className="workflow-inline-fields">
              <Field label="流程编码"><TextInput value={workflowForm.code} onChange={(event) => setWorkflowForm({ ...workflowForm, code: event.target.value })} /></Field>
              <Field label="类型"><TextInput value={workflowForm.category} onChange={(event) => setWorkflowForm({ ...workflowForm, category: event.target.value })} /></Field>
              <Field label="版本"><TextInput value={workflowForm.version} onChange={(event) => setWorkflowForm({ ...workflowForm, version: event.target.value })} /></Field>
              <Field label="触发事件"><TextInput value={workflowForm.triggerEventType} onChange={(event) => setWorkflowForm({ ...workflowForm, triggerEventType: event.target.value })} /></Field>
              <Field label="触发资源"><TextInput value={workflowForm.triggerResource} onChange={(event) => setWorkflowForm({ ...workflowForm, triggerResource: event.target.value })} /></Field>
              {presetVariables.length ? (
                <div className="workflow-template-hint">
                  <b>可用变量</b>
                  <span>{presetVariables.map((field) => field.label || field.key).join("、")}</span>
                </div>
              ) : null}
              {presetTriggers.length ? (
                <div className="workflow-template-hint span-all">
                  <b>触发入口</b>
                  <span>{presetTriggers.map((trigger) => `${trigger.module} / ${trigger.action} / ${trigger.method} ${trigger.path}`).join("；")}</span>
                </div>
              ) : null}
              <div className="workflow-list-table workflow-condition-table span-all">
                <div className="workflow-list-head">
                  <span>字段</span>
                  <span>判断</span>
                  <span>值</span>
                  <span>操作</span>
                </div>
                {draftConditions.length ? draftConditions.map((condition, index) => (
                  <div className="workflow-list-row" key={`${condition.field}-${index}`}>
                    <SelectInput value={condition.field} onChange={(event) => setWorkflowCondition(index, { field: event.target.value })}>
                      {!workflowConditionFieldOptionList().some((option) => option.value === condition.field) && condition.field ? <option value={condition.field}>{condition.field}</option> : null}
                      {workflowConditionFieldOptionList().map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                    </SelectInput>
                    <SelectInput value={condition.operator || "equals"} onChange={(event) => setWorkflowCondition(index, { operator: event.target.value })}>
                      {workflowConditionOperatorOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                    </SelectInput>
                    <TextInput value={condition.value || ""} disabled={condition.operator === "exists" || condition.operator === "missing"} onChange={(event) => setWorkflowCondition(index, { value: event.target.value })} />
                    <UiButton disabled={actionBusy !== ""} onClick={() => removeWorkflowCondition(index)}>删除</UiButton>
                  </div>
                )) : (
                  <div className="workflow-list-row workflow-empty-row">
                    <span>无附加条件</span>
                    <span>-</span>
                    <span>-</span>
                    <span>-</span>
                  </div>
                )}
                <UiButton className="workflow-add-row" variant="ghost" icon={<Plus size={14} />} onClick={addWorkflowCondition}>
                  添加触发条件
                </UiButton>
              </div>
            </div>
          </details>
          <FormActions>
            {editingWorkflowId ? (
              <UiButton type="button" onClick={() => handleRollbackWorkflowVersion(editingWorkflowId)} disabled={actionBusy !== ""}>回滚此版</UiButton>
            ) : (
              <UiButton type="button" onClick={resetWorkflowForm} disabled={actionBusy !== ""}>重置</UiButton>
            )}
            {editingWorkflowId ? <UiButton type="button" onClick={handlePublishWorkflowVersion} disabled={actionBusy !== ""}>发布新版</UiButton> : null}
            <UiButton variant="primary" type="submit" icon={<Route size={14} />} disabled={actionBusy !== ""}>保存流程</UiButton>
          </FormActions>
        </WorkflowForm>
      );
    }

    function approvalFlowFormPanel() {
      return (
        <SystemForm onSubmit={handleSaveApprovalFlow}>
          <Field label="编码"><TextInput value={approvalFlowForm.code} onChange={(event) => setApprovalFlowForm({ ...approvalFlowForm, code: event.target.value })} /></Field>
          <Field label="名称"><TextInput value={approvalFlowForm.name} onChange={(event) => setApprovalFlowForm({ ...approvalFlowForm, name: event.target.value })} /></Field>
          <Field label="业务对象">
            <SelectInput value={approvalFlowForm.resource} onChange={(event) => setApprovalFlowForm({ ...approvalFlowForm, resource: event.target.value })}>
              {workflowResourceOptionList().map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="状态">
            <SelectInput value={approvalFlowForm.status} onChange={(event) => setApprovalFlowForm({ ...approvalFlowForm, status: event.target.value })}>
              {configStatusOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="步骤 JSON" spanAll><TextAreaInput value={approvalFlowForm.steps} onChange={(event) => setApprovalFlowForm({ ...approvalFlowForm, steps: event.target.value })} /></Field>
          <FormActions spanAll>
            <UiButton onClick={() => resetApprovalFlowForm()} disabled={actionBusy !== ""}>重置</UiButton>
            <UiButton variant="primary" type="submit" icon={<CheckCircle2 size={14} />} disabled={actionBusy !== "" || !approvalFlowForm.code.trim() || !approvalFlowForm.name.trim()}>保存审批流</UiButton>
          </FormActions>
        </SystemForm>
      );
    }

    return (
      <Panel className="system-management-view workflow-management-view workflow-simple-view">
        <div className="workflow-simple-shell">
          <div className="workflow-simple-header">
            <div>
              <h3>工作流</h3>
              <span>业务页面负责发起和审批，这里只管审批规则、异常和外部通知。</span>
            </div>
            <ActionGroup>
              <ActionDialog id="workflow-create" title="新增审批规则" buttonLabel="新增规则" triggerIcon={<Plus size={14} />} triggerVariant="primary" onOpen={resetWorkflowForm}>
                {workflowEditorForm()}
              </ActionDialog>
              <UiButton disabled={actionBusy !== "" || (pendingDeliveries.length === 0 && overdueWorkflowTasks.length === 0)} onClick={runWorkflowAutomation}>自动处理</UiButton>
            </ActionGroup>
          </div>

          <MetricList compact className="workflow-simple-metrics">
            <div><span>审批规则</span><b>{currentDefinitions.length}</b></div>
            <div><span>待处理</span><b>{pendingTasks.length}</b></div>
            <div><span>逾期</span><b>{overdueWorkflowTasks.length}</b></div>
            <div><span>异常</span><b>{unresolvedEvents.length + deliveryIssues.length + outboxIssues.length}</b></div>
            <div><span>通知目标</span><b>{subscriptions.filter((item) => item.status === "active").length}/{subscriptions.length}</b></div>
          </MetricList>

          <div className="workflow-simple-grid">
            <section className="workflow-simple-card workflow-simple-card-wide">
              <div className="workflow-simple-card-head">
                <div>
                  <b>审批规则</b>
                  <span>选业务场景，配置谁来审核。</span>
                </div>
              </div>
              <div className="workflow-simple-list">
                {currentDefinitions.map((definition) => (
                  <div className="workflow-simple-row" key={definition.id}>
                    <span><b>{definition.name}</b><small>{workflowResourceLabel(definition.resource)} / {list(definition.steps).map((step) => roleName(step.roleCode)).join(" → ") || "无审批步骤"}</small></span>
                    {workflowDefinitionVersionNode(definition)}
                    <StatusChip value={definition.status} />
                    <ActionDialog id={`workflow-simple-definition-${definition.id}`} title="编辑审批规则" buttonLabel="编辑" onOpen={() => startWorkflowEdit(definition)}>
                      {workflowDefinitionVersionPanel(definition)}
                      {workflowEditorForm()}
                    </ActionDialog>
                  </div>
                ))}
                {!currentDefinitions.length ? <span className="workflow-simple-empty">暂无审批规则</span> : null}
              </div>
            </section>

            <section className="workflow-simple-card workflow-simple-card-wide">
              <div className="workflow-simple-card-head">
                <div>
                  <b>审批流配置</b>
                  <span>旧版审批流资源和步骤配置。</span>
                </div>
                <ActionDialog id="approval-flow-create" title="新增审批流" buttonLabel="新增审批流" triggerIcon={<Plus size={13} />} onOpen={() => resetApprovalFlowForm()}>
                  {approvalFlowFormPanel()}
                </ActionDialog>
              </div>
              <div className="workflow-simple-list">
                {approvalFlows.map((flow) => (
                  <div className="workflow-simple-row" key={flow.id}>
                    <span><b>{flow.name}</b><small>{workflowResourceLabel(flow.resource)} / {flow.code}</small></span>
                    <span>{list(flow.steps).map((step) => roleName(step.roleCode)).join(" → ") || "无审批步骤"}</span>
                    <StatusChip value={flow.status} />
                    <ActionDialog id={`approval-flow-${flow.id}`} title="编辑审批流" buttonLabel="编辑" onOpen={() => resetApprovalFlowForm(flow)}>
                      {approvalFlowFormPanel()}
                      <ActionGroup>
                        <UiButton
                          disabled={actionBusy !== ""}
                          variant={flow.status === "active" ? "danger" : "soft"}
                          onClick={() => handleApprovalFlowStatus(flow, flow.status === "active" ? "disabled" : "active")}
                        >
                          {flow.status === "active" ? "禁用" : "启用"}
                        </UiButton>
                        {flow.status !== "draft" ? (
                          <UiButton disabled={actionBusy !== ""} onClick={() => handleApprovalFlowStatus(flow, "draft")}>转草稿</UiButton>
                        ) : null}
                        <UiButton variant="danger" disabled={actionBusy !== "" || flow.status === "active"} onClick={() => handleDeleteApprovalFlow(flow)}>删除</UiButton>
                      </ActionGroup>
                    </ActionDialog>
                  </div>
                ))}
                {!approvalFlows.length ? <span className="workflow-simple-empty">暂无审批流配置</span> : null}
              </div>
            </section>

            <section className="workflow-simple-card">
              <div className="workflow-simple-card-head">
                <div>
                  <b>待处理</b>
                  <span>需要当前角色处理的任务。</span>
                </div>
              </div>
              <div className="workflow-simple-list">
                {pendingTasks.slice(0, 6).map((task) => (
                  <div className="workflow-simple-row workflow-simple-row-compact" key={task.id}>
                    <span><b>{task.stepName || task.taskNo}</b><small>{task.definitionCode} / {workflowResourceLabel(task.resource)} #{task.resourceId}</small></span>
                    <span className={workflowTaskOverdue(task) ? "workflow-due-overdue" : "muted"}>{workflowTaskDueLabel(task)}</span>
                    <ActionDialog id={`workflow-simple-task-${task.id}`} title="处理待办" buttonLabel="处理">
                      <div className="finance-action-block">
                        <b>{task.taskNo}</b>
                        <span>{task.definitionCode} / {roleName(task.roleCode)}</span>
                        <span className={workflowTaskOverdue(task) ? "workflow-due-overdue" : ""}>{workflowTaskDueLabel(task)}</span>
                        <TextInput value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
                        <ActionGroup>
                          <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => actWorkflowTask(task.id, "approve")}>通过</UiButton>
                          <UiButton disabled={actionBusy !== ""} onClick={() => actWorkflowTask(task.id, "reject")}>驳回</UiButton>
                        </ActionGroup>
                      </div>
                    </ActionDialog>
                  </div>
                ))}
                {!pendingTasks.length ? <span className="workflow-simple-empty">暂无待处理任务</span> : null}
              </div>
            </section>

            <section className="workflow-simple-card">
              <div className="workflow-simple-card-head">
                <div>
                  <b>异常</b>
                  <span>只显示需要人工看一眼的问题。</span>
                </div>
              </div>
              <div className="workflow-simple-list">
                {unresolvedEvents.slice(0, 4).map((event) => (
                  <div className="workflow-simple-row workflow-simple-row-compact" key={event.id}>
                    <span><b>{event.title || event.eventNo}</b><small>{workflowEventResourceLabel(event)} / {workflowEventMatchedLabel(event)}</small></span>
                    <StatusChip value={event.status} />
                    <ActionDialog id={`workflow-simple-event-${event.id}`} title="处理异常事件" buttonLabel="处理">
                      <div className="finance-action-block">
                        <b>{event.title || event.eventNo}</b>
                        <span>{event.eventType} / {workflowEventResourceLabel(event)}</span>
                        <span>原因：{event.reason || event.error || "-"}</span>
                        <TextAreaInput value={workflowEventResolution} onChange={(changeEvent) => setWorkflowEventResolution(changeEvent.target.value)} />
                        <ActionGroup>
                          <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={() => replayWorkflowEvent(event)}>重放</UiButton>
                          <UiButton disabled={actionBusy !== ""} onClick={() => resolveWorkflowEvent(event)}>标记已处理</UiButton>
                        </ActionGroup>
                      </div>
                    </ActionDialog>
                  </div>
                ))}
                {deliveryIssues.slice(0, 4).map((item) => (
                  <div className="workflow-simple-row workflow-simple-row-compact" key={`delivery-${item.id}`}>
                    <span><b>{item.deliveryNo}</b><small>{item.subscriptionName || item.subscriptionCode} / {workflowDeliveryDetail(item)}</small></span>
                    <StatusChip value={item.status} />
                    {workflowDeliveryActions(item)}
                  </div>
                ))}
                {outboxIssues.slice(0, 3).map((item) => (
                  <div className="workflow-simple-row workflow-simple-row-compact" key={`outbox-${item.id}`}>
                    <span><b>{item.outboxNo}</b><small>{item.eventType} / {workflowOutboxDetail(item)}</small></span>
                    <StatusChip value={item.status} />
                    {workflowOutboxActions(item)}
                  </div>
                ))}
                {!unresolvedEvents.length && !deliveryIssues.length && !outboxIssues.length ? <span className="workflow-simple-empty">暂无异常</span> : null}
              </div>
            </section>
          </div>

          <section className="workflow-simple-card">
            <div className="workflow-simple-card-head">
              <div>
                <b>外部通知</b>
                <span>把审批结果推送到外部系统。</span>
              </div>
              <ActionDialog id="workflow-subscription-create" title="新增通知" buttonLabel="新增通知" triggerIcon={<Link2 size={14} />} onOpen={resetWorkflowSubscriptionForm}>
                {workflowSubscriptionFormPanel()}
              </ActionDialog>
            </div>
            <div className="workflow-simple-list">
              {subscriptions.slice(0, 8).map((item) => (
                <div className="workflow-simple-row" key={item.id}>
                  <span><b>{item.name}</b><small>{workflowOutboxEventLabel(item.eventType)} / {item.endpoint}</small></span>
                  <span>{item.resource ? workflowResourceLabel(item.resource) : "全部对象"} / {item.definitionCode || "全部流程"}</span>
                  <StatusChip value={item.status} />
                  <ActionGroup>
                    <ActionDialog id={`workflow-subscription-${item.id}`} title="编辑通知" buttonLabel="编辑" onOpen={() => startWorkflowSubscriptionEdit(item)}>
                      {workflowSubscriptionFormPanel()}
                    </ActionDialog>
                    <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleWorkflowSubscriptionStatus(item, item.status === "active" ? "disabled" : "active")}>{item.status === "active" ? "停用" : "启用"}</UiButton>
                    <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "active" || item.status === "enabled"} onClick={() => handleDeleteWorkflowSubscription(item)}>删除</UiButton>
                  </ActionGroup>
                </div>
              ))}
              {!subscriptions.length ? <span className="workflow-simple-empty">暂无外部通知</span> : null}
            </div>
          </section>

          <details className="workflow-simple-advanced">
            <summary>高级诊断</summary>
            <div className="workflow-simple-advanced-body">
              <DataTable
          data={workflowRows}
          rowKey={(item) => item.id}
          pageSize={12}
          rowContextMenu={buildDataTableRowContextMenu<(typeof workflowRows)[number]>({
            actions: [
              {
                key: "focus-resource",
                label: "只看业务对象",
                onSelect: (item, helpers) => {
                  if (item.kind === "definition") helpers.searchText(item.definition.resource);
                  if (item.kind === "event") helpers.searchText(item.event.resource);
                  if (item.kind === "task") helpers.searchText(item.task.resource);
                  if (item.kind === "instance") helpers.searchText(item.instance.resource);
                }
              },
              {
                key: "focus-status",
                label: "只看当前状态",
                onSelect: (item, helpers) => helpers.searchText(item.kind === "definition" ? item.definition.status : item.kind === "event" ? item.event.status : item.kind === "task" ? item.task.status : item.instance.status)
              }
            ],
            copyFields: [
              { key: "name", label: "流程名称", value: (item) => item.kind === "definition" ? item.definition.name : item.kind === "event" ? item.event.title || item.event.eventNo : item.kind === "task" ? item.task.taskNo : item.instance.title || item.instance.instanceNo },
              { key: "code", label: "流程编码", value: (item) => item.kind === "definition" ? item.definition.code : item.kind === "event" ? item.event.eventNo : item.kind === "task" ? item.task.definitionCode : item.instance.definitionName },
              { key: "resource", label: "业务对象", value: (item) => workflowRowResourceText(item) },
              { key: "status", label: "状态", value: (item) => item.kind === "definition" ? item.definition.status : item.kind === "event" ? item.event.status : item.kind === "task" ? item.task.status : item.instance.status }
            ]
          })}
          headerLeftAction={workflowRuntimeFilterBar()}
	          headerAction={(
	            <ActionGroup>
	              <ActionDialog id="workflow-create" title="新增流程" onOpen={resetWorkflowForm}>
	                {workflowEditorForm()}
	              </ActionDialog>
              <ActionDialog id="workflow-event-publish" title="发布事件" onOpen={resetWorkflowEventForm}>
                {workflowEventPublisherForm()}
              </ActionDialog>
	              <ActionDialog id="workflow-outbox" title="事件出口">
	                <div className="finance-action-block">
	                  <b>订阅目标 {subscriptions.length}</b>
	                  {workflowSubscriptionFormPanel()}
		                  {subscriptions.slice(0, 6).map((item) => (
		                    <div className="workflow-subscription-row" key={item.id}>
		                      <span><b>{item.name}</b><span className="block-text muted">{item.code} / {workflowOutboxEventLabel(item.eventType)} / {item.eventType}</span></span>
		                      <span>{item.resource ? workflowResourceLabel(item.resource) : "全部对象"} / {item.definitionCode || "全部流程"} / {item.endpoint}</span>
	                      <StatusChip value={item.status} />
	                      <ActionGroup>
	                        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => startWorkflowSubscriptionEdit(item)}>编辑</UiButton>
	                        <UiButton size="sm" disabled={actionBusy !== ""} onClick={() => handleWorkflowSubscriptionStatus(item, item.status === "active" ? "disabled" : "active")}>{item.status === "active" ? "停用" : "启用"}</UiButton>
	                        <UiButton size="sm" variant="danger" disabled={actionBusy !== "" || item.status === "active" || item.status === "enabled"} onClick={() => handleDeleteWorkflowSubscription(item)}>删除</UiButton>
	                      </ActionGroup>
	                    </div>
	                  ))}
		                </div>
		                <div className="finance-action-block">
		                  <b>投递记录 {pendingDeliveries.length} 待处理 / 逾期任务 {overdueWorkflowTasks.length} / 全部 {deliveries.length}</b>
		                  <ActionGroup>
		                    <UiButton size="sm" disabled={actionBusy !== "" || (pendingDeliveries.length === 0 && overdueWorkflowTasks.length === 0)} onClick={runWorkflowAutomation}>运行自动化</UiButton>
		                  </ActionGroup>
		                  {deliveries.slice(0, 12).map((item) => (
		                    <div className="workflow-delivery-row" key={item.id}>
		                      <span><b>{item.deliveryNo}</b> / {item.subscriptionName || item.subscriptionCode}</span>
	                      <span>{item.eventType} / {item.outboxNo}</span>
	                      <span>{workflowDeliveryStatusText(item)} / {workflowDeliveryDetail(item)}{item.lastError ? ` / ${item.lastError}` : ""}</span>
	                      {workflowDeliveryActions(item)}
	                    </div>
	                  ))}
	                  {!deliveries.length ? <span>暂无投递记录</span> : null}
	                </div>
	                <div className="finance-action-block">
	                  <b>原始事件 {pendingOutbox.length} 待处理 / 全部 {outbox.length}</b>
	                  {outbox.slice(0, 12).map((item) => (
	                    <div className="workflow-outbox-row" key={item.id}>
	                      <span><b>{item.outboxNo}</b> / {item.eventType}</span>
	                      <span>{item.definitionCode || "-"} / {item.resource || "-"} #{item.resourceId || "-"}</span>
	                      <span>{workflowOutboxStatusText(item)} / {workflowOutboxDetail(item)}{item.lastError ? ` / ${item.lastError}` : ""}</span>
	                      {workflowOutboxActions(item)}
	                    </div>
	                  ))}
	                  {!outbox.length ? <span>暂无出口事件</span> : null}
	                </div>
	              </ActionDialog>
              <ActionDialog id="workflow-summary" title="流程汇总">
                <MetricList compact className="system-summary-grid">
                  <div><span>流程编码</span><b>{workflowDefinitionCodes.length}</b></div>
                  <div><span>版本</span><b>{definitions.length}</b></div>
                  <div><span>事件</span><b>{events.length}</b></div>
                  <div><span>实例</span><b>{activeInstances.length}</b></div>
                  <div><span>待办</span><b>{pendingTasks.length}</b></div>
	                  <div><span>待处理事件</span><b>{unresolvedEvents.length}</b></div>
	                  <div><span>历史版本</span><b>{historicalDefinitions.length}</b></div>
	                  <div><span>待投递</span><b>{pendingOutbox.length}</b></div>
	                  <div><span>启用</span><b>{definitions.filter((item) => item.status === "active").length}</b></div>
	                  <div><span>事件接入</span><b>{integratedCatalogEvents.length}/{catalogEvents.length}</b></div>
	                  <div><span>结果回写</span><b>{writeBackCatalogEvents.filter((item) => item.integration?.hasResultHandler).length}/{writeBackCatalogEvents.length}</b></div>
	                </MetricList>
	                <div className="finance-action-block">
	                  <b>事件接入覆盖</b>
	                  {catalogEvents.slice(0, 12).map((event) => (
	                    <div className="workflow-subscription-row" key={event.eventType}>
	                      <span><b>{event.label}</b><span className="block-text muted">{event.eventType} / {workflowResourceLabel(event.resource)}</span></span>
	                      <span>{workflowCatalogIntegrationText(event)}</span>
	                      <StatusChip value={event.integration?.status || "missing_trigger"} />
	                    </div>
	                  ))}
	                  {!catalogEvents.length ? <span>暂无事件目录</span> : null}
	                </div>
	              </ActionDialog>
            </ActionGroup>
          )}
          columns={[
            { key: "type", title: "类型", render: (item) => workflowRowKindLabel(item.kind) },
            {
              key: "name",
              title: "名称",
              render: (item) => {
                if (item.kind === "definition") return <><b>{item.definition.name}</b><span className="block-text muted">{item.definition.code}</span></>;
                if (item.kind === "event") return <><b>{item.event.title || item.event.eventNo}</b><span className="block-text muted">{item.event.eventNo} / {item.event.eventType}</span></>;
                if (item.kind === "task") return <><b>{item.task.taskNo}</b><span className="block-text muted">{item.task.definitionCode}</span></>;
                return <><b>{item.instance.title || item.instance.instanceNo}</b><span className="block-text muted">{item.instance.definitionName}</span></>;
              }
            },
            {
              key: "resource",
              title: "业务对象",
              render: (item) => {
                return workflowRowResourceText(item);
              }
            },
            {
              key: "current",
              title: "当前节点",
              render: (item) => {
                if (item.kind === "definition") return workflowDefinitionVersionNode(item.definition);
                if (item.kind === "event") return workflowEventMatchedLabel(item.event);
                if (item.kind === "task") return <><b>{roleName(item.task.roleCode)}</b><span className={`block-text ${workflowTaskOverdue(item.task) ? "workflow-due-overdue" : "muted"}`}>{workflowTaskDueLabel(item.task)}</span></>;
                return item.instance.currentRole ? `${item.instance.currentStep}. ${roleName(item.instance.currentRole)}` : "-";
              }
            },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.kind === "definition" ? item.definition.status : item.kind === "event" ? item.event.status : item.kind === "task" ? item.task.status : item.instance.status} /> },
            {
              key: "actions",
              title: "操作",
              width: "120px",
              render: (item) => {
                if (item.kind === "definition") {
                  return (
                    <ActionDialog id={`workflow-definition-${item.definition.id}`} title="编辑流程" onOpen={() => startWorkflowEdit(item.definition)}>
                      {workflowDefinitionVersionPanel(item.definition)}
                      {workflowEditorForm()}
                    </ActionDialog>
                  );
                }
                if (item.kind === "task") {
                  return (
                    <ActionDialog id={`workflow-task-${item.task.id}`} title="流程任务" onOpen={() => setWorkflowReassignRole(item.task.roleCode)}>
                      <div className="finance-action-block">
                        <b>{item.task.taskNo}</b>
                        <span>{item.task.definitionCode} / {item.task.resource} #{item.task.resourceId}</span>
	                        <span>{item.task.status}</span>
	                        <span>{roleName(item.task.roleCode)} / {item.task.stepName}</span>
	                        <span className={workflowTaskOverdue(item.task) ? "workflow-due-overdue" : ""}>{workflowTaskDueLabel(item.task)}</span>
	                        {item.task.escalatedAt ? <span>已升级：{roleName(item.task.escalatedFromRole)} → {roleName(item.task.roleCode)} / {item.task.escalatedAt}</span> : null}
	                        {item.task.status === "pending" ? (
	                          <>
	                            <TextInput value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
                            <ActionGroup>
                              <UiButton variant="primary" icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => actWorkflowTask(item.task.id, "approve")}>通过</UiButton>
                              <UiButton disabled={actionBusy !== ""} onClick={() => actWorkflowTask(item.task.id, "reject")}>驳回</UiButton>
                            </ActionGroup>
                            <SelectInput value={workflowReassignRole} onChange={(event) => setWorkflowReassignRole(event.target.value)}>
                              {availableRoles().map((role) => <option key={role.code} value={role.code}>{role.name}</option>)}
                            </SelectInput>
	                            <TextInput value={workflowReassignReason} onChange={(event) => setWorkflowReassignReason(event.target.value)} />
	                            <ActionGroup>
	                              <UiButton disabled={actionBusy !== "" || workflowReassignRole === item.task.roleCode} onClick={() => reassignWorkflowTask(item.task.id)}>改派</UiButton>
	                              <UiButton disabled={actionBusy !== "" || workflowReassignRole === item.task.roleCode || !workflowTaskOverdue(item.task)} onClick={() => escalateWorkflowTask(item.task.id)}>升级</UiButton>
	                            </ActionGroup>
	                          </>
	                        ) : null}
                      </div>
                    </ActionDialog>
                  );
                }
                if (item.kind === "event") {
                  return (
                    <ActionDialog id={`workflow-event-${item.event.id}`} title="流程事件">
                      <div className="finance-action-block">
                        <b>{item.event.title || item.event.eventNo}</b>
                        <span>{item.event.eventNo} / {item.event.eventType}</span>
                        <span>来源：{item.event.source || "-"}{item.event.eventKey ? ` / ${item.event.eventKey}` : ""}</span>
                        <span>{workflowEventResourceLabel(item.event)}</span>
                        <span>匹配流程：{workflowEventMatchedLabel(item.event)}</span>
                        <span>发起人：{item.event.actor || "-"}</span>
                        <span>原因：{item.event.reason || item.event.error || "-"}</span>
                        {item.event.replayOfId ? <span>重放来源：#{item.event.replayOfId}</span> : null}
                        {item.event.recoveredByEventId ? <span>恢复事件：#{item.event.recoveredByEventId}</span> : null}
                        {item.event.resolution ? <span>处理说明：{item.event.resolution}</span> : null}
                        {item.event.resolvedAt ? <span>处理人：{item.event.resolvedBy || "-"} / {item.event.resolvedAt}</span> : null}
                        {Object.entries(item.event.variables || {}).length ? (
                          <ChipList compact>
                            {Object.entries(item.event.variables || {}).slice(0, 8).map(([key, value]) => <span key={key}>{key}: {value || "-"}</span>)}
                          </ChipList>
                        ) : null}
                        {workflowEventNeedsRecovery(item.event) ? (
                          <TextAreaInput value={workflowEventResolution} onChange={(event) => setWorkflowEventResolution(event.target.value)} />
                        ) : null}
                        {workflowEventNeedsRecovery(item.event) ? (
                          <ActionGroup>
                            <UiButton icon={<RefreshCw size={13} />} disabled={actionBusy !== ""} onClick={() => replayWorkflowEvent(item.event)}>重放事件</UiButton>
                            <UiButton disabled={actionBusy !== ""} onClick={() => resolveWorkflowEvent(item.event)}>标记已处理</UiButton>
                          </ActionGroup>
                        ) : null}
                      </div>
                    </ActionDialog>
                  );
                }
                return (
                  <ActionDialog id={`workflow-instance-${item.instance.id}`} title="流程实例">
                    <div className="finance-action-block">
                      <b>{item.instance.title || item.instance.instanceNo}</b>
                      <span>{item.instance.definitionName}</span>
                      <span>{item.instance.status}</span>
                      <span>{item.instance.applicant || "-"}</span>
                      {item.instance.triggerEventId ? <span>触发事件：#{item.instance.triggerEventId}</span> : null}
                      {workflowInstanceLogs(item.instance.id).length ? (
                        <ChipList compact>
                          {workflowInstanceLogs(item.instance.id).map((log) => (
                            <span key={log.id}>{log.createdAt} / {workflowActionLabel(log.action)} / {log.actor || "-"}</span>
                          ))}
                        </ChipList>
                      ) : <span>暂无运行日志</span>}
                      {item.instance.status === "pending" ? (
                        <>
                          <TextAreaInput value={workflowCancelReason} onChange={(event) => setWorkflowCancelReason(event.target.value)} />
                          <ActionGroup>
                            <UiButton disabled={actionBusy !== ""} onClick={() => cancelWorkflowInstance(item.instance.id)}>取消实例</UiButton>
                          </ActionGroup>
                        </>
                      ) : null}
                    </div>
                  </ActionDialog>
                );
              }
            }
          ]}
        />
            </div>
          </details>
        </div>
      </Panel>
    );
  }

  function renderAuditLogManagement() {
    const auditLogs = [...data.auditLogs].sort((a, b) => b.id - a.id);
    const actionTypes = Array.from(new Set(auditLogs.map((item) => item.action).filter(Boolean)));
    const resourceTypes = Array.from(new Set(auditLogs.map((item) => item.resource).filter(Boolean)));

    return (
      <Panel className="system-management-view audit-log-view">
        <MetricList compact className="system-summary-grid">
          <div><span>审计事件</span><b>{auditLogs.length}</b></div>
          <div><span>操作人</span><b>{new Set(auditLogs.map((item) => item.user).filter(Boolean)).size}</b></div>
          <div><span>动作类型</span><b>{actionTypes.length}</b></div>
          <div><span>资源类型</span><b>{resourceTypes.length}</b></div>
        </MetricList>
        <DataTable
          data={auditLogs}
          rowKey={(item) => item.id}
          pageSize={14}
          onRefresh={refreshData}
          searchPlaceholder="搜索操作人 / 动作 / 资源 / 详情 / IP"
          rowContextMenu={buildDataTableRowContextMenu<AuditLog>({
            actions: [
              { key: "focus-user", label: "只看该操作人", icon: <Search size={14} />, onSelect: (item, helpers) => helpers.searchText(item.user) },
              { key: "focus-action", label: "只看该动作", icon: <Filter size={14} />, onSelect: (item, helpers) => helpers.searchText(item.action) },
              { key: "focus-resource", label: "只看该资源", icon: <Filter size={14} />, onSelect: (item, helpers) => helpers.searchText(item.resource) }
            ],
            copyFields: [
              { key: "user", label: "操作人", value: (item) => item.user },
              { key: "action", label: "动作", value: (item) => item.action },
              { key: "resource", label: "资源", value: (item) => `${item.resource}#${item.resourceId || 0}` },
              { key: "detail", label: "详情", value: (item) => item.detail },
              { key: "ip", label: "IP", value: (item) => item.ip }
            ]
          })}
          columns={[
            { key: "createdAt", title: "时间", render: (item) => <><b>{shortDateTime(item.createdAt)}</b><span className="block-text muted">#{item.id}</span></> },
            { key: "user", title: "操作人", render: (item) => item.user || "-" },
            { key: "action", title: "动作", render: (item) => item.action || "-" },
            { key: "resource", title: "资源", render: (item) => <><b>{item.resource || "-"}</b><span className="block-text muted">{item.resourceId ? `ID ${item.resourceId}` : "-"}</span></> },
            { key: "detail", title: "详情", render: (item) => item.detail || "-" },
            { key: "ip", title: "IP", render: (item) => item.ip || "-" }
          ]}
          emptyText="暂无审计日志"
        />
      </Panel>
    );
  }

  function renderRoleManagement() {
    const roles = availableRoles();
    const fieldPolicies = list(data.system?.security?.fieldPolicies);
    const permissionCatalog = Array.from(new Set([
      ...roles.flatMap((item) => list(item.permissions)),
      "system:*",
      "master:*",
      "order:read",
      "production:read",
      "dispatch:*",
      "delivery:read",
      "delivery:write",
      "ticket:read",
      "vehicle:read",
      "statement:read",
      "procurement:read",
      "report:read",
      "finance:read",
      "quality:read",
      "org:read",
      "approval:*"
    ])).filter(Boolean);
    const dataScopeOptions = dictionaryOptions("data_scope");
    const fieldPolicyResources = [
      { code: "customers", label: "客户" },
      { code: "customerContacts", label: "客户联系人" },
      { code: "projects", label: "项目" },
      { code: "orders", label: "订单" },
      { code: "deliverySigns", label: "签收" },
      { code: "drivers", label: "司机" },
      { code: "*", label: "全部支持资源" }
    ];
    const fieldPolicyFields = [
      { code: "phone", label: "手机号" },
      { code: "licenseNo", label: "驾驶证号" },
      { code: "*", label: "全部支持字段" }
    ];
    const fieldPolicyMasks = [
      { code: "phone", label: "手机号脱敏" },
      { code: "code", label: "证件号脱敏" },
      { code: "redact", label: "完全隐藏" }
    ];
    const fieldPolicyResourceLabel = (value: string) => fieldPolicyResources.find((item) => item.code === value)?.label || value || "-";
    const fieldPolicyFieldLabel = (value: string) => fieldPolicyFields.find((item) => item.code === value)?.label || value || "-";
    const fieldPolicyMaskLabel = (value: string) => fieldPolicyMasks.find((item) => item.code === value)?.label || value || "-";

    function roleEditorForm() {
      return (
        <SystemForm onSubmit={handleSaveRole}>
          <Field label="角色编码"><TextInput value={roleForm.code} onChange={(event) => setRoleForm({ ...roleForm, code: event.target.value })} /></Field>
          <Field label="角色名称"><TextInput value={roleForm.name} onChange={(event) => setRoleForm({ ...roleForm, name: event.target.value })} /></Field>
          <Field label="数据范围">
            <SelectInput value={roleForm.dataScope} onChange={(event) => setRoleForm({ ...roleForm, dataScope: event.target.value })}>
              {dataScopeOptions.map((option) => <option key={option.code} value={option.code}>{option.label}</option>)}
            </SelectInput>
          </Field>
          <Field label="权限">
            <TextAreaInput value={roleForm.permissions} onChange={(event) => setRoleForm({ ...roleForm, permissions: event.target.value })} />
          </Field>
          <ChipList className="span-all">
            {permissionCatalog.slice(0, 18).map((permission) => (
              <ChipButton
                className="permission-chip-button"
                key={permission}
                onClick={() => setRoleForm((form) => ({ ...form, permissions: Array.from(new Set([...rolePermissions(), permission])).join("\n") }))}
              >
                {permission}
              </ChipButton>
            ))}
          </ChipList>
          <FormActions spanAll>
            <UiButton onClick={resetRoleForm} disabled={actionBusy !== ""}>重置</UiButton>
            <UiButton variant="primary" type="submit" icon={<ShieldCheck size={14} />} disabled={actionBusy !== ""}>{editingRoleId ? "保存角色" : "新增角色"}</UiButton>
          </FormActions>
        </SystemForm>
      );
    }

    return (
      <Panel className="system-management-view">
        <DataTable
          data={roles}
          rowKey={(item) => item.id}
          pageSize={10}
          rowContextMenu={buildDataTableRowContextMenu<Role>({
            actions: [
              { key: "focus-scope", label: "只看该数据范围", onSelect: (item, helpers) => helpers.searchText(orgLevelLabel(item.dataScope)) },
              { key: "focus-role", label: "只看该角色", onSelect: (item, helpers) => helpers.searchText(item.name) }
            ],
            copyFields: [
              { key: "code", label: "角色编码", value: (item) => item.code },
              { key: "name", label: "角色名称", value: (item) => item.name },
              { key: "scope", label: "数据范围", value: (item) => orgLevelLabel(item.dataScope) },
              { key: "permissions", label: "权限", value: (item) => list(item.permissions).join("\n") }
            ]
          })}
          headerLeftAction={<ActionDialog id="system-role-create" title="新增角色" onOpen={resetRoleForm}>{roleEditorForm()}</ActionDialog>}
          columns={[
            { key: "role", title: "角色", render: (item) => <><b>{item.name}</b><span className="block-text muted">{item.code}</span></> },
            { key: "scope", title: "范围", render: (item) => orgLevelLabel(item.dataScope) },
            { key: "permissionCount", title: "权限数", render: (item) => list(item.permissions).length },
            { key: "permissions", title: "权限", render: (item) => (
              <ChipList compact>
                {list(item.permissions).slice(0, 6).map((permission) => <span key={permission}>{permission}</span>)}
                {list(item.permissions).length > 6 ? <span>+{list(item.permissions).length - 6}</span> : null}
              </ChipList>
            ) },
            {
              key: "actions",
              title: "操作",
              render: (item) => (
                <ActionGroup className="compact-actions">
                  <ActionDialog id={`system-role-${item.id}`} title="编辑角色" onOpen={() => startRoleEdit(item)}>
                    {roleEditorForm()}
                  </ActionDialog>
                  <UiButton variant="danger" disabled={actionBusy !== "" || item.code === "boss"} onClick={() => handleDeleteRole(item)}>删除</UiButton>
                </ActionGroup>
              )
            }
          ]}
        />
        <div className="system-management-section">
          <SplitRow>
            <div>
              <h3>字段脱敏策略</h3>
              <span className="muted">按角色、资源和字段控制后端数据脱敏</span>
            </div>
            <ShieldCheck size={20} />
          </SplitRow>
          <SystemForm onSubmit={handleCreateFieldPolicy}>
            <Field label="角色">
              <SelectInput value={fieldPolicyForm.roleCode} onChange={(event) => setFieldPolicyForm({ ...fieldPolicyForm, roleCode: event.target.value })}>
                {roles.map((item) => <option key={item.code} value={item.code}>{item.name} / {item.code}</option>)}
              </SelectInput>
            </Field>
            <Field label="资源">
              <SelectInput value={fieldPolicyForm.resource} onChange={(event) => setFieldPolicyForm({ ...fieldPolicyForm, resource: event.target.value })}>
                {fieldPolicyResources.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="字段">
              <SelectInput value={fieldPolicyForm.field} onChange={(event) => setFieldPolicyForm({ ...fieldPolicyForm, field: event.target.value })}>
                {fieldPolicyFields.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="脱敏方式">
              <SelectInput value={fieldPolicyForm.mask} onChange={(event) => setFieldPolicyForm({ ...fieldPolicyForm, mask: event.target.value })}>
                {fieldPolicyMasks.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
              </SelectInput>
            </Field>
            <Field label="备注" className="span-all">
              <TextAreaInput value={fieldPolicyForm.remark} onChange={(event) => setFieldPolicyForm({ ...fieldPolicyForm, remark: event.target.value })} />
            </Field>
            <FormActions spanAll>
              <UiButton onClick={resetFieldPolicyForm} disabled={actionBusy !== ""}>重置</UiButton>
              <UiButton variant="primary" type="submit" icon={<ShieldCheck size={14} />} disabled={actionBusy !== "" || !fieldPolicyForm.roleCode}>{fieldPolicyForm.id ? "保存策略" : "新增策略"}</UiButton>
            </FormActions>
          </SystemForm>
          <DataTable
            data={fieldPolicies}
            rowKey={(item) => item.id}
            emptyText="暂无字段策略"
            pageSize={8}
            rowContextMenu={buildDataTableRowContextMenu<FieldPolicy>({
              actions: [
                { key: "focus-role", label: "只看该角色", onSelect: (item, helpers) => helpers.searchText(roleName(item.roleCode)) },
                { key: "focus-resource", label: "只看该资源", onSelect: (item, helpers) => helpers.searchText(fieldPolicyResourceLabel(item.resource)) }
              ],
              copyFields: [
                { key: "role", label: "角色", value: (item) => roleName(item.roleCode) },
                { key: "resource", label: "资源", value: (item) => item.resource },
                { key: "field", label: "字段", value: (item) => item.field },
                { key: "mask", label: "脱敏方式", value: (item) => item.mask }
              ]
            })}
            columns={[
              { key: "role", title: "角色", render: (item) => <><b>{roleName(item.roleCode)}</b><span className="block-text muted">{item.roleCode}</span></> },
              { key: "resource", title: "资源", render: (item) => <><b>{fieldPolicyResourceLabel(item.resource)}</b><span className="block-text muted">{item.resource}</span></> },
              { key: "field", title: "字段", render: (item) => <><b>{fieldPolicyFieldLabel(item.field)}</b><span className="block-text muted">{item.field}</span></> },
              { key: "mask", title: "脱敏", render: (item) => fieldPolicyMaskLabel(item.mask) },
              { key: "enabled", title: "状态", render: (item) => <StatusChip value={item.enabled ? "active" : "disabled"} /> },
              { key: "remark", title: "备注", render: (item) => item.remark || "-" },
              {
                key: "actions",
                title: "操作",
              render: (item) => (
                <ActionGroup className="compact-actions">
                  <UiButton disabled={actionBusy !== ""} onClick={() => startFieldPolicyEdit(item)}>编辑</UiButton>
                  <UiButton disabled={actionBusy !== ""} onClick={() => handleToggleFieldPolicy(item)}>{item.enabled ? "停用" : "启用"}</UiButton>
                  <UiButton variant="danger" disabled={actionBusy !== ""} onClick={() => handleDeleteFieldPolicy(item)}>删除</UiButton>
                </ActionGroup>
              )
            }
            ]}
          />
        </div>
      </Panel>
    );
  }

  function ordersTable(orders: SalesOrder[], withActions = false, headerLeftAction?: ReactNode) {
    return (
      <DataTable
        data={orders}
        rowKey={(order) => order.id}
        emptyText="暂无数据"
        pageSize={withActions ? 10 : 8}
        onRefresh={withActions ? refreshData : undefined}
        showPagination={withActions}
        rowContextMenu={buildDataTableRowContextMenu<SalesOrder>({
          actions: [
            { key: "focus-customer", label: "只看该客户", onSelect: (order, helpers) => helpers.searchText(nameOf(bootstrap?.customers, order.customerId)) },
            { key: "focus-project", label: "只看该项目", onSelect: (order, helpers) => helpers.searchText(nameOf(bootstrap?.projects, order.projectId)) },
            { key: "focus-product", label: "只看该产品", onSelect: (order, helpers) => helpers.searchText(productLabel(bootstrap, order.productId)) }
          ],
          copyFields: [
            { key: "order", label: "订单号", value: (order) => order.orderNo },
            { key: "customer", label: "客户", value: (order) => nameOf(bootstrap?.customers, order.customerId) },
            { key: "project", label: "项目", value: (order) => nameOf(bootstrap?.projects, order.projectId) },
            { key: "address", label: "收货地址", value: (order) => order.receiveAddress }
          ]
        })}
        headerLeftAction={headerLeftAction}
        headerAction={withActions ? undefined : <BarChart3 size={18} />}
        columns={[
          { key: "orderNo", title: "订单号", render: (order) => order.orderNo },
          {
            key: "customerProject",
            title: "客户 / 项目",
            render: (order) => (
              <>
                {nameOf(bootstrap?.customers, order.customerId)}
                <span className="block-text muted">{nameOf(bootstrap?.projects, order.projectId)}</span>
              </>
            )
          },
          { key: "product", title: "产品", render: (order) => productLabel(bootstrap, order.productId) },
          { key: "planQuantity", title: "计划量", render: (order) => `${qty(order.planQuantity)} ${order.unit}` },
	          {
	            key: "status",
	            title: "状态",
	            render: (order) => workflowStatusFor(["sales_order", "sales_orders", "order"], order.id, order.orderNo, approvalStatus(approvalFor(["sales_order", "sales_orders", "order"], order.id, order.orderNo), <StatusChip value={order.status} />))
	          },
	          ...(withActions ? [{
	            key: "actions",
	            title: "操作",
	            render: (order: SalesOrder) => {
	              const task = approvalFor(["sales_order", "sales_orders", "order"], order.id, order.orderNo);
	              const workflow = workflowItemsFor(["sales_order", "sales_orders", "order"], order.id, order.orderNo);
	              if (!task && !workflow.instances.length && order.status !== "submitted") return <span className="muted">-</span>;
	              return (
	                <ActionDialog id={`order-approval-${order.id}`} title="订单流程" buttonLabel={task || workflow.instances.length ? "流程" : "操作"} triggerIcon={<CheckCircle2 size={13} />}>
	                  <div className="finance-hidden-actions">
	                    <div className="finance-action-block">
	                      <b>{order.orderNo}</b>
                      <span>{nameOf(bootstrap?.customers, order.customerId)} / {nameOf(bootstrap?.projects, order.projectId)}</span>
	                      <span>{productLabel(bootstrap, order.productId)} / {qty(order.planQuantity)} {order.unit} / {money(order.totalAmount)}</span>
	                      <span>{order.receiveAddress || "-"}</span>
	                    </div>
	                    {workflowTimelineBlock(["sales_order", "sales_orders", "order"], order.id, order.orderNo, "当前订单暂无工作流实例")}
	                    {!workflow.instances.length ? approvalActionBlock(task) : null}
	                    {!task && !workflow.instances.length && order.status === "submitted" ? (
	                      <div className="finance-action-block">
                        <b>订单审批</b>
                        <span>当前订单未关联工作流待办，按订单状态直接审批。</span>
                        <UiButton icon={<CheckCircle2 size={13} />} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`order-approve-${order.id}`, "订单已审批", () => api.approveOrder(order.id))}>通过</UiButton>
                      </div>
                    ) : null}
                  </div>
                </ActionDialog>
              );
            }
          }] : [])
        ]}
      />
    );
  }

}
