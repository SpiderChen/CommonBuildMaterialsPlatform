import {
  BarChart3,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Clock,
  ClipboardCheck,
  CreditCard,
  Factory,
  FileSignature,
  Landmark,
  MapPin,
  Package,
  PlayCircle,
  Plus,
  ReceiptText,
  RefreshCw,
  Route,
  Scale,
  Search,
  ShoppingCart,
  Truck,
  X
} from "lucide-react";
import { type FormEvent, type ReactNode, useEffect, useMemo, useState } from "react";
import { DataTable } from "../components/DataTable";
import { KpiCard } from "../components/KpiCard";
import { nameOf, productLabel } from "../components/names";
import { StatusChip } from "../components/StatusChip";
import { api } from "../services/api";
import type {
  ApprovalTask,
  BootstrapData,
  Customer,
  DashboardData,
  DeliverySign,
  DeliverySignLink,
  Driver,
  DispatchCenterOverview,
  DispatchCenterProductionTask,
  DispatchCenterQueueItem,
  DispatchCenterSiteProgress,
  DispatchCenterVehicle,
  FinanceOverview,
  InventoryItem,
  LatestLocation,
  ManagementReports,
  Material,
  ProcurementOverview,
  Product,
  ProductionOverview,
  Project,
  SalesOrder,
  ScaleTicket,
  ScaleWeightRecord,
  Site,
  Statement,
  Vehicle
} from "../services/types";

export type ERPWorkbenchSection =
  | "overview"
  | "master-customers"
  | "master-projects"
  | "master-products"
  | "master-materials"
  | "master-sites"
  | "master-drivers"
  | "master-vehicles"
  | "master-inventory"
  | "orders"
  | "production"
  | "dispatch"
  | "weighbridge"
  | "delivery"
  | "settlement"
  | "procurement"
  | "finance"
  | "reports"
  | "system";

type WorkbenchData = {
  dashboard: DashboardData | null;
  reports: ManagementReports | null;
  dispatch: DispatchCenterOverview | null;
  production: ProductionOverview | null;
  procurement: ProcurementOverview | null;
  finance: FinanceOverview | null;
  orders: SalesOrder[];
  tickets: ScaleTicket[];
  weightRecords: ScaleWeightRecord[];
  signs: DeliverySign[];
  signLinks: DeliverySignLink[];
  statements: Statement[];
  approvals: ApprovalTask[];
};

type MasterKind = "customer" | "project" | "product" | "material" | "site" | "driver" | "vehicle" | "inventory";
type MasterRecord = Customer | Project | Product | Material | Site | Driver | Vehicle | InventoryItem;

const emptyData: WorkbenchData = {
  dashboard: null,
  reports: null,
  dispatch: null,
  production: null,
  procurement: null,
  finance: null,
  orders: [],
  tickets: [],
  weightRecords: [],
  signs: [],
  signLinks: [],
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
  return value ? value.slice(5, 16) : "待计算";
}

function dispatchSearchText(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation) {
  return Object.values(item).join(" ").toLowerCase();
}

function matchesDispatchSearch(item: DispatchCenterSiteProgress | DispatchCenterQueueItem | DispatchCenterProductionTask | DispatchCenterVehicle | LatestLocation, keyword: string) {
  return keyword === "" || dispatchSearchText(item).includes(keyword.toLowerCase());
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

function nextDispatchAction(status: string | undefined) {
  switch (status) {
    case "assigned":
      return "受理";
    case "accepted":
      return "到厂";
    case "arrived_site":
    case "waiting_load":
      return "装料";
    case "loading":
      return "装完";
    case "loaded":
      return "出厂";
    case "departed":
      return "运输";
    case "in_transit":
      return "到场";
    case "arrived_project":
      return "卸料";
    case "unloading":
      return "签收";
    case "signed":
      return "完成";
    default:
      return "推进";
  }
}

function etaText(item: DispatchCenterQueueItem) {
  if (!item.etaMinutes) {
    return shortDateTime(item.eta || item.plannedEta);
  }
  return `${Math.round(item.etaMinutes)} 分钟`;
}

function receivableBalance(finance: FinanceOverview | null) {
  return list(finance?.receivables).reduce((sum, item) => sum + Math.max(0, item.amount - item.receivedAmount), 0);
}

export function ERPWorkbenchView({
  bootstrap,
  section,
  onChanged
}: {
  bootstrap: BootstrapData | null;
  section: ERPWorkbenchSection;
  onChanged: () => void;
}) {
  const [data, setData] = useState<WorkbenchData>(emptyData);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [siteFilter, setSiteFilter] = useState("all");
  const [dispatchSearch, setDispatchSearch] = useState("");
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null);
  const [selectedVehicleId, setSelectedVehicleId] = useState<number | null>(null);
  const [dispatchQty, setDispatchQty] = useState("");
  const [dispatchActionError, setDispatchActionError] = useState("");
  const [dispatchSubmitting, setDispatchSubmitting] = useState(false);
  const [vehiclesCollapsed, setVehiclesCollapsed] = useState(false);
  const [actionBusy, setActionBusy] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [actionError, setActionError] = useState("");
  const [approvalComment, setApprovalComment] = useState("同意");
  const [editingMaster, setEditingMaster] = useState<{ kind: MasterKind; id: number } | null>(null);
  const [masterDialogKind, setMasterDialogKind] = useState<MasterKind | null>(null);
  const [orderDialogOpen, setOrderDialogOpen] = useState(false);
  const [masterForm, setMasterForm] = useState({
    customerName: "新建客户",
    customerContact: "项目负责人",
    customerPhone: "13800000000",
    projectCustomerId: "",
    projectName: "新建项目",
    projectAddress: "项目地址",
    productName: "C30 商品混凝土",
    productSpec: "C30",
    productPrice: "380",
    materialName: "水泥",
    materialSpec: "P.O 42.5",
    materialSafeStock: "100",
    siteName: "新建站点",
    siteCode: "SITE-NEW",
    siteAddress: "站点地址",
    driverName: "新司机",
    driverPhone: "13800000001",
    vehiclePlate: "川A00000",
    vehicleType: "搅拌车",
    vehicleCapacity: "12",
    vehicleDriverId: "",
    vehicleSiteId: "",
    inventorySiteId: "",
    inventoryMaterialId: "",
    inventoryQuantity: "100",
    inventoryWarehouse: "主仓"
  });
  const [orderForm, setOrderForm] = useState({
    customerId: "",
    projectId: "",
    productId: "",
    siteId: "",
    planQuantity: "30",
    unitPrice: "380",
    planTime: today,
    contact: "",
    phone: ""
  });
  const [procurementForm, setProcurementForm] = useState({
    purchaseOrderId: "",
    supplierId: "",
    siteId: "",
    materialId: "",
    plateNo: "原料车-001",
    grossWeight: "42",
    tareWeight: "16"
  });
  const [financeForm, setFinanceForm] = useState({
    receivableId: "",
    receiptAmount: "",
    planAmount: "",
    planDueDate: today
  });

  async function load() {
    setLoading(true);
    setError("");
    try {
      const [dashboard, reports, dispatch, production, procurement, finance, orders, tickets, weightRecords, signs, signLinks, statements, approvals] = await Promise.all([
        api.dashboard(),
        api.reports(),
        api.dispatchCenterOverview(),
        api.productionOverview(),
        api.procurementOverview(),
        api.financeOverview(),
        api.orders(),
        api.tickets(),
        api.weightRecords(),
        api.signs(),
        api.signLinks(),
        api.statements(),
        api.approvals()
      ]);
      setData({
        dashboard,
        reports,
        dispatch,
        production,
        procurement,
        finance,
        orders: list(orders),
        tickets: list(tickets),
        weightRecords: list(weightRecords),
        signs: list(signs),
        signLinks: list(signLinks),
        statements: list(statements),
        approvals: list(approvals)
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "ERP 数据加载失败");
      setData(emptyData);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : "ERP 数据加载失败"));
  }, [section]);

  useEffect(() => {
    const progress = list(data.dispatch?.siteProgress);
    const vehicles = list(data.dispatch?.availableVehicles);
    const siteIds = new Set(progress.map((item) => String(item.siteId)));
    if (siteFilter !== "all" && !siteIds.has(siteFilter)) {
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
  }, [data.dispatch, selectedOrderId, selectedVehicleId, siteFilter]);

  useEffect(() => {
    setMasterForm((value) => ({
      ...value,
      projectCustomerId: value.projectCustomerId || String(firstId(bootstrap?.customers)),
      vehicleDriverId: value.vehicleDriverId || String(firstId(bootstrap?.drivers)),
      vehicleSiteId: value.vehicleSiteId || String(firstId(bootstrap?.sites)),
      inventorySiteId: value.inventorySiteId || String(firstId(bootstrap?.sites)),
      inventoryMaterialId: value.inventoryMaterialId || String(firstId(bootstrap?.materials))
    }));
    setOrderForm((value) => ({
      ...value,
      customerId: value.customerId || String(firstId(bootstrap?.customers)),
      projectId: value.projectId || String(firstId(bootstrap?.projects)),
      productId: value.productId || String(firstId(bootstrap?.products)),
      siteId: value.siteId || String(firstId(bootstrap?.sites))
    }));
  }, [bootstrap]);

  useEffect(() => {
    const supplier = supplierOptions()[0];
    const purchaseOrder = list(data.procurement?.orders)[0];
    const receivable = openReceivables()[0] || list(data.finance?.receivables)[0];
    setProcurementForm((value) => ({
      ...value,
      purchaseOrderId: value.purchaseOrderId || String(purchaseOrder?.id || ""),
      supplierId: value.supplierId || String(recordId(supplier)),
      siteId: value.siteId || String(firstId(bootstrap?.sites)),
      materialId: value.materialId || String(firstId(bootstrap?.materials))
    }));
    setFinanceForm((value) => ({
      ...value,
      receivableId: value.receivableId || String(receivable?.id || ""),
      receiptAmount: value.receiptAmount || String(Math.max(0, (receivable?.amount || 0) - (receivable?.receivedAmount || 0)) || ""),
      planAmount: value.planAmount || String(Math.max(0, (receivable?.amount || 0) - (receivable?.receivedAmount || 0)) || "")
    }));
  }, [bootstrap, data.procurement, data.finance]);

  const activeOrders = useMemo(() => data.orders.filter((item) => !["completed", "cancelled"].includes(item.status)), [data.orders]);
  const openApprovals = useMemo(() => data.approvals.filter((item) => item.status !== "approved" && item.status !== "rejected"), [data.approvals]);
  const unsignedSignLinks = useMemo(() => data.signLinks.filter((item) => item.status !== "used"), [data.signLinks]);
  const unconfirmedStatements = useMemo(() => data.statements.filter((item) => item.status !== "confirmed"), [data.statements]);
  const financeRisk = list(data.finance?.agingBuckets).reduce((sum, item) => sum + (item.overdueAmount || 0), 0) || data.dashboard?.operating?.overdueReceivable || 0;

  function supplierOptions() {
    return list(data.procurement?.suppliers as Array<{ id: number; name?: string; supplierName?: string; orderNo?: string }> | undefined);
  }

  function openReceivables() {
    return list(data.finance?.receivables).filter((item) => item.status !== "paid" && item.amount > item.receivedAmount);
  }

  async function runBusinessAction(label: string, success: string, action: () => Promise<unknown>) {
    setActionBusy(label);
    setActionMessage("");
    setActionError("");
    try {
      await action();
      setActionMessage(success);
      onChanged();
      await load();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setActionBusy("");
    }
  }

  function masterResource(kind: MasterKind) {
    return ({
      customer: "customers",
      project: "projects",
      product: "products",
      material: "materials",
      site: "sites",
      driver: "drivers",
      vehicle: "vehicles",
      inventory: "inventory"
    } as Record<string, string>)[kind] || kind;
  }

  async function handleMasterSubmit(kind: MasterKind, event: FormEvent) {
    event.preventDefault();
    const companyId = bootstrap?.companies[0]?.id || 1;
    const selectedMaterial = bootstrap?.materials.find((item) => item.id === fieldNumber(masterForm.inventoryMaterialId));
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
        address: masterForm.projectAddress
      },
      product: {
        name: masterForm.productName,
        spec: masterForm.productSpec,
        unit: "m3",
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
        companyId,
        name: masterForm.siteName,
        code: masterForm.siteCode,
        address: masterForm.siteAddress
      },
      driver: {
        name: masterForm.driverName,
        phone: masterForm.driverPhone
      },
      vehicle: {
        plateNo: masterForm.vehiclePlate,
        vehicleType: masterForm.vehicleType,
        capacity: masterForm.vehicleCapacity,
        driverId: fieldNumber(masterForm.vehicleDriverId),
        siteId: fieldNumber(masterForm.vehicleSiteId),
        carrier: bootstrap?.carriers[0]?.name || ""
      },
      inventory: {
        siteId: fieldNumber(masterForm.inventorySiteId),
        materialId: fieldNumber(masterForm.inventoryMaterialId),
        quantity: fieldNumber(masterForm.inventoryQuantity),
        warehouse: masterForm.inventoryWarehouse,
        unit: selectedMaterial?.unit || "t"
      }
    };
    const payload = payloads[kind];
    const editing = editingMaster?.kind === kind ? editingMaster : null;
    await runBusinessAction(`master-${kind}`, editing ? "基础资料已更新" : "基础资料已保存", async () => {
      if (editing) {
        await api.updateMasterResource(masterResource(kind), editing.id, payload);
      } else {
        const createActions: Record<MasterKind, () => Promise<unknown>> = {
          customer: () => api.createCustomer(payload as never),
          project: () => api.createProject(payload as never),
          product: () => api.createProduct(payload as never),
          material: () => api.createMaterial(payload as never),
          site: () => api.createSite(payload as never),
          driver: () => api.createDriver(payload as never),
          vehicle: () => api.createVehicle(payload as never),
          inventory: () => api.createInventoryItem(payload as never)
        };
        await createActions[kind]();
      }
      setEditingMaster(null);
      setMasterDialogKind(null);
    });
  }

  function resetMasterForm(kind: MasterKind) {
    setMasterForm((form) => {
      switch (kind) {
        case "customer":
          return { ...form, customerName: "新建客户", customerContact: "项目负责人", customerPhone: "13800000000" };
        case "project":
          return { ...form, projectCustomerId: String(firstId(bootstrap?.customers)), projectName: "新建项目", projectAddress: "项目地址" };
        case "product":
          return { ...form, productName: "C30 商品混凝土", productSpec: "C30", productPrice: "380" };
        case "material":
          return { ...form, materialName: "水泥", materialSpec: "P.O 42.5", materialSafeStock: "100" };
        case "site":
          return { ...form, siteName: "新建站点", siteCode: "SITE-NEW", siteAddress: "站点地址" };
        case "driver":
          return { ...form, driverName: "新司机", driverPhone: "13800000001" };
        case "vehicle":
          return { ...form, vehiclePlate: "川A00000", vehicleType: "搅拌车", vehicleCapacity: "12", vehicleDriverId: String(firstId(bootstrap?.drivers)), vehicleSiteId: String(firstId(bootstrap?.sites)) };
        case "inventory":
          return { ...form, inventorySiteId: String(firstId(bootstrap?.sites)), inventoryMaterialId: String(firstId(bootstrap?.materials)), inventoryQuantity: "100", inventoryWarehouse: "主仓" };
      }
    });
  }

  function openMasterCreateDialog(kind: MasterKind) {
    setEditingMaster(null);
    resetMasterForm(kind);
    setMasterDialogKind(kind);
  }

  function clearMasterEdit() {
    setEditingMaster(null);
    setMasterDialogKind(null);
  }

  function startMasterEdit(kind: MasterKind, item: MasterRecord) {
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
          projectAddress: value.address || ""
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
        setMasterForm((form) => ({
          ...form,
          siteName: value.name,
          siteCode: value.code,
          siteAddress: value.address || ""
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
      case "vehicle": {
        const value = item as Vehicle;
        setMasterForm((form) => ({
          ...form,
          vehiclePlate: value.plateNo,
          vehicleType: value.vehicleType || "",
          vehicleCapacity: value.capacity || "",
          vehicleDriverId: String(value.driverId || ""),
          vehicleSiteId: String(value.siteId || "")
        }));
        break;
      }
      case "inventory": {
        const value = item as InventoryItem;
        setMasterForm((form) => ({
          ...form,
          inventorySiteId: String(value.siteId),
          inventoryMaterialId: String(value.materialId),
          inventoryQuantity: String(value.quantity || ""),
          inventoryWarehouse: value.warehouse || ""
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
      <div className="form-actions span-all">
        <button className="primary-button icon-button-text" type="submit" disabled={actionBusy !== "" || disabled}>
          {editing ? <CheckCircle2 size={15} /> : <Plus size={15} />}{editing ? "保存修改" : label}
        </button>
        <button className="soft-button" type="button" disabled={actionBusy !== ""} onClick={clearMasterEdit}>取消</button>
      </div>
    );
  }

  function masterRowActions(kind: MasterKind, id: number, item: MasterRecord) {
    return (
      <div className="row-actions">
        <button className="soft-button" type="button" disabled={actionBusy !== ""} onClick={() => startMasterEdit(kind, item)}>编辑</button>
        <button className="soft-button danger-button" type="button" disabled={actionBusy !== ""} onClick={() => handleDeleteMaster(kind, id)}>删除</button>
      </div>
    );
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
        unit: product?.unit || "m3",
        unitPrice,
        planTime: orderForm.planTime,
        contact: orderForm.contact,
        phone: orderForm.phone,
        lines: [{
          productId: fieldNumber(orderForm.productId),
          productLine: product?.line || "concrete",
          quantity,
          unit: product?.unit || "m3",
          unitPrice
        }]
      });
      setOrderDialogOpen(false);
    });
  }

  function openOrderDialog() {
    const product = bootstrap?.products[0];
    setOrderForm((form) => ({
      ...form,
      customerId: String(firstId(bootstrap?.customers)),
      projectId: String(firstId(bootstrap?.projects)),
      productId: String(firstId(bootstrap?.products)),
      siteId: String(firstId(bootstrap?.sites)),
      planQuantity: form.planQuantity || "30",
      unitPrice: String(product?.basePrice || form.unitPrice || "380"),
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
      onChanged();
      await load();
    } catch (err) {
      setDispatchActionError(err instanceof Error ? err.message : "派车失败");
    } finally {
      setDispatchSubmitting(false);
    }
  }

  async function handleAdvanceDispatch(item: DispatchCenterQueueItem) {
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

  const modules = [
    { icon: ShoppingCart, title: "销售订单", value: data.orders.length, detail: `${activeOrders.length} 单执行中`, status: activeOrders.length ? "running" : "completed" },
    { icon: Factory, title: "生产计划", value: list(data.production?.tasks).length, detail: `${list(data.production?.batches).length} 个生产批次`, status: list(data.production?.tasks).length ? "producing" : "completed" },
    { icon: Truck, title: "调度运输", value: data.dispatch?.kpis?.activeDispatches || 0, detail: `${data.dispatch?.kpis?.onlineVehicles || 0}/${data.dispatch?.kpis?.totalVehicles || 0} 车辆在线`, status: "dispatching" },
    { icon: Scale, title: "地磅票据", value: data.tickets.length, detail: `${data.weightRecords.length} 条称重记录`, status: "active" },
    { icon: FileSignature, title: "工地签收", value: data.signs.length, detail: `${unsignedSignLinks.length} 个链接待签收`, status: unsignedSignLinks.length ? "warning" : "signed" },
    { icon: ReceiptText, title: "客户对账", value: data.statements.length, detail: `${unconfirmedStatements.length} 单待确认`, status: unconfirmedStatements.length ? "warning" : "confirmed" },
    { icon: Package, title: "采购库存", value: list(data.procurement?.receipts).length, detail: `${list(data.procurement?.inventory).length} 条库存`, status: "stocked" },
    { icon: Landmark, title: "财务结算", value: money(receivableBalance(data.finance)), detail: `逾期 ${money(financeRisk)}`, status: financeRisk > 0 ? "warning" : "confirmed" },
    { icon: ClipboardCheck, title: "审批中心", value: openApprovals.length, detail: `${data.approvals.length} 条审批记录`, status: openApprovals.length ? "pending_approval" : "approved" }
  ];

  if (loading) {
    return <section className="panel">ERP 数据加载中...</section>;
  }

  return (
    <div className="view-stack">
      {error ? (
        <section className="panel">
          <p className="error-text">{error}</p>
          <button className="soft-button" onClick={load}>重新加载</button>
        </section>
      ) : null}
      {actionMessage || actionError ? (
        <section className={`action-feedback ${actionError ? "error" : "success"}`}>
          <span>{actionError || actionMessage}</span>
        </section>
      ) : null}
      {renderMasterDialog()}
      {renderOrderDialog()}
      {section === "overview" ? renderOverview() : null}
      {section === "master-customers" ? renderMasterCustomers() : null}
      {section === "master-projects" ? renderMasterProjects() : null}
      {section === "master-products" ? renderMasterProducts() : null}
      {section === "master-materials" ? renderMasterMaterials() : null}
      {section === "master-sites" ? renderMasterSites() : null}
      {section === "master-drivers" ? renderMasterDrivers() : null}
      {section === "master-vehicles" ? renderMasterVehicles() : null}
      {section === "master-inventory" ? renderMasterInventory() : null}
      {section === "orders" ? renderOrders() : null}
      {section === "production" ? renderProduction() : null}
      {section === "dispatch" ? renderDispatch() : null}
      {section === "weighbridge" ? renderWeighbridge() : null}
      {section === "delivery" ? renderDelivery() : null}
      {section === "settlement" ? renderSettlement() : null}
      {section === "procurement" ? renderProcurement() : null}
      {section === "finance" ? renderFinance() : null}
      {section === "reports" ? renderReports() : null}
      {section === "system" ? renderSystem() : null}
    </div>
  );

  function reloadButton() {
    return (
      <button className="soft-button" onClick={() => { onChanged(); load(); }}>
        刷新
      </button>
    );
  }

  function masterEntityName(kind: MasterKind) {
    return ({
      customer: "客户",
      project: "项目",
      product: "产品",
      material: "物料",
      site: "站点",
      driver: "司机",
      vehicle: "车辆",
      inventory: "库存"
    } as Record<MasterKind, string>)[kind];
  }

  function renderMasterDialog() {
    if (!masterDialogKind) {
      return null;
    }
    const editing = editingMaster?.kind === masterDialogKind;
    return (
      <div className="modal-backdrop" role="presentation">
        <section className="modal-panel master-dialog" role="dialog" aria-modal="true" aria-label={`${editing ? "编辑" : "新增"}${masterEntityName(masterDialogKind)}`}>
          <div className="modal-head">
            <div>
              <h3>{editing ? "编辑" : "新增"}{masterEntityName(masterDialogKind)}</h3>
              <p className="muted">资料归属当前业务模块，保存后同步刷新台账。</p>
            </div>
            <button className="icon-button" type="button" aria-label="关闭" disabled={actionBusy !== ""} onClick={clearMasterEdit}>
              <X size={18} />
            </button>
          </div>
          {renderMasterForm(masterDialogKind)}
        </section>
      </div>
    );
  }

  function renderMasterForm(kind: MasterKind) {
    const customers = bootstrap?.customers || [];
    const materials = bootstrap?.materials || [];
    const sites = bootstrap?.sites || [];
    const drivers = bootstrap?.drivers || [];
    switch (kind) {
      case "customer":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("customer", event)}>
            <label><span>客户名称</span><input value={masterForm.customerName} onChange={(event) => setMasterForm({ ...masterForm, customerName: event.target.value })} /></label>
            <label><span>联系人</span><input value={masterForm.customerContact} onChange={(event) => setMasterForm({ ...masterForm, customerContact: event.target.value })} /></label>
            <label className="span-all"><span>电话</span><input value={masterForm.customerPhone} onChange={(event) => setMasterForm({ ...masterForm, customerPhone: event.target.value })} /></label>
            {masterFormButton("customer", "新增客户")}
          </form>
        );
      case "project":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("project", event)}>
            <label>
              <span>客户</span>
              <select value={masterForm.projectCustomerId} onChange={(event) => setMasterForm({ ...masterForm, projectCustomerId: event.target.value })}>
                {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>项目名称</span><input value={masterForm.projectName} onChange={(event) => setMasterForm({ ...masterForm, projectName: event.target.value })} /></label>
            <label className="span-all"><span>项目地址</span><input value={masterForm.projectAddress} onChange={(event) => setMasterForm({ ...masterForm, projectAddress: event.target.value })} /></label>
            {masterFormButton("project", "新增项目", !customers.length)}
          </form>
        );
      case "product":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("product", event)}>
            <label><span>产品名称</span><input value={masterForm.productName} onChange={(event) => setMasterForm({ ...masterForm, productName: event.target.value })} /></label>
            <label><span>规格</span><input value={masterForm.productSpec} onChange={(event) => setMasterForm({ ...masterForm, productSpec: event.target.value })} /></label>
            <label className="span-all"><span>基准价</span><input type="number" value={masterForm.productPrice} onChange={(event) => setMasterForm({ ...masterForm, productPrice: event.target.value })} /></label>
            {masterFormButton("product", "新增产品")}
          </form>
        );
      case "material":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("material", event)}>
            <label><span>物料名称</span><input value={masterForm.materialName} onChange={(event) => setMasterForm({ ...masterForm, materialName: event.target.value })} /></label>
            <label><span>规格</span><input value={masterForm.materialSpec} onChange={(event) => setMasterForm({ ...masterForm, materialSpec: event.target.value })} /></label>
            <label className="span-all"><span>安全库存</span><input type="number" value={masterForm.materialSafeStock} onChange={(event) => setMasterForm({ ...masterForm, materialSafeStock: event.target.value })} /></label>
            {masterFormButton("material", "新增物料")}
          </form>
        );
      case "site":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("site", event)}>
            <label><span>站点名称</span><input value={masterForm.siteName} onChange={(event) => setMasterForm({ ...masterForm, siteName: event.target.value })} /></label>
            <label><span>编码</span><input value={masterForm.siteCode} onChange={(event) => setMasterForm({ ...masterForm, siteCode: event.target.value })} /></label>
            <label className="span-all"><span>地址</span><input value={masterForm.siteAddress} onChange={(event) => setMasterForm({ ...masterForm, siteAddress: event.target.value })} /></label>
            {masterFormButton("site", "新增站点")}
          </form>
        );
      case "driver":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("driver", event)}>
            <label><span>司机</span><input value={masterForm.driverName} onChange={(event) => setMasterForm({ ...masterForm, driverName: event.target.value })} /></label>
            <label><span>电话</span><input value={masterForm.driverPhone} onChange={(event) => setMasterForm({ ...masterForm, driverPhone: event.target.value })} /></label>
            {masterFormButton("driver", "新增司机")}
          </form>
        );
      case "vehicle":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("vehicle", event)}>
            <label><span>车牌</span><input value={masterForm.vehiclePlate} onChange={(event) => setMasterForm({ ...masterForm, vehiclePlate: event.target.value })} /></label>
            <label><span>车型</span><input value={masterForm.vehicleType} onChange={(event) => setMasterForm({ ...masterForm, vehicleType: event.target.value })} /></label>
            <label>
              <span>司机</span>
              <select value={masterForm.vehicleDriverId} onChange={(event) => setMasterForm({ ...masterForm, vehicleDriverId: event.target.value })}>
                {drivers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>站点</span>
              <select value={masterForm.vehicleSiteId} onChange={(event) => setMasterForm({ ...masterForm, vehicleSiteId: event.target.value })}>
                {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label className="span-all"><span>容量</span><input value={masterForm.vehicleCapacity} onChange={(event) => setMasterForm({ ...masterForm, vehicleCapacity: event.target.value })} /></label>
            {masterFormButton("vehicle", "新增车辆", !drivers.length || !sites.length)}
          </form>
        );
      case "inventory":
        return (
          <form className="lab-form dialog-form" onSubmit={(event) => handleMasterSubmit("inventory", event)}>
            <label>
              <span>站点</span>
              <select value={masterForm.inventorySiteId} onChange={(event) => setMasterForm({ ...masterForm, inventorySiteId: event.target.value })}>
                {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>物料</span>
              <select value={masterForm.inventoryMaterialId} onChange={(event) => setMasterForm({ ...masterForm, inventoryMaterialId: event.target.value })}>
                {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>仓库</span><input value={masterForm.inventoryWarehouse} onChange={(event) => setMasterForm({ ...masterForm, inventoryWarehouse: event.target.value })} /></label>
            <label><span>数量</span><input type="number" value={masterForm.inventoryQuantity} onChange={(event) => setMasterForm({ ...masterForm, inventoryQuantity: event.target.value })} /></label>
            {masterFormButton("inventory", "新增库存", !sites.length || !materials.length)}
          </form>
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
    const sites = bootstrap?.sites || [];
    const disabled = actionBusy !== "" || !customers.length || !projects.length || !products.length || !sites.length;
    return (
      <div className="modal-backdrop" role="presentation">
        <section className="modal-panel order-dialog" role="dialog" aria-modal="true" aria-label="新增销售订单">
          <div className="modal-head">
            <div>
              <h3>新增销售订单</h3>
              <p className="muted">订单提交后进入生产、调度、地磅、签收和结算链路。</p>
            </div>
            <button className="icon-button" type="button" aria-label="关闭" disabled={actionBusy !== ""} onClick={() => setOrderDialogOpen(false)}>
              <X size={18} />
            </button>
          </div>
          <form className="lab-form dialog-form" onSubmit={handleCreateOrder}>
            <label>
              <span>客户</span>
              <select value={orderForm.customerId} onChange={(event) => setOrderForm({ ...orderForm, customerId: event.target.value })}>
                {customers.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>项目</span>
              <select value={orderForm.projectId} onChange={(event) => setOrderForm({ ...orderForm, projectId: event.target.value })}>
                {projects.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>产品</span>
              <select value={orderForm.productId} onChange={(event) => setOrderForm({ ...orderForm, productId: event.target.value })}>
                {products.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
              </select>
            </label>
            <label>
              <span>站点</span>
              <select value={orderForm.siteId} onChange={(event) => setOrderForm({ ...orderForm, siteId: event.target.value })}>
                {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>计划方量</span><input type="number" min="0" step="0.5" value={orderForm.planQuantity} onChange={(event) => setOrderForm({ ...orderForm, planQuantity: event.target.value })} /></label>
            <label><span>单价</span><input type="number" min="0" step="1" value={orderForm.unitPrice} onChange={(event) => setOrderForm({ ...orderForm, unitPrice: event.target.value })} /></label>
            <label className="span-all"><span>计划时间</span><input type="date" value={orderForm.planTime} onChange={(event) => setOrderForm({ ...orderForm, planTime: event.target.value })} /></label>
            <div className="form-actions span-all">
              <button className="soft-button" type="button" disabled={actionBusy !== ""} onClick={() => setOrderDialogOpen(false)}>取消</button>
              <button className="primary-button icon-button-text" type="submit" disabled={disabled}>
                <Plus size={14} />提交订单
              </button>
            </div>
          </form>
        </section>
      </div>
    );
  }

  function renderOverview() {
    return (
      <>
        <section className="kpi-grid">
          <KpiCard label="订单数" value={data.dashboard?.operating?.orderCount || data.orders.length} />
          <KpiCard label="计划方量" value={qty(data.dashboard?.operating?.plannedQty)} suffix="m3" />
          <KpiCard label="签收方量" value={qty(data.dashboard?.operating?.signedQty)} suffix="m3" />
          <KpiCard label="应收余额" value={money(data.dashboard?.operating?.receivableBalance || receivableBalance(data.finance))} suffix="元" />
          <KpiCard label="毛利率" value={percent(data.dashboard?.operating?.grossMargin)} />
        </section>

        <section className="panel">
          <div className="between">
            <div>
              <h3>ERP 业务模块</h3>
              <p className="muted">销售、生产、调度、地磅、签收、结算、采购、财务和审批都在客户侧 ERP 内。</p>
            </div>
            {reloadButton()}
          </div>
          <div className="capability-grid">
            {modules.map((item) => {
              const Icon = item.icon;
              return (
                <div className="capability-card" key={item.title}>
                  <div className="between">
                    <Icon size={20} />
                    <StatusChip value={item.status} />
                  </div>
                  <strong>{item.title}</strong>
                  <span>{item.detail}</span>
                  <b>{item.value}</b>
                </div>
              );
            })}
          </div>
        </section>

        <section className="grid-12">
          <div className="panel span-7">{ordersTable(data.orders.slice(0, 8), "最近订单")}</div>
          <div className="panel span-5">{dispatchSummary()}</div>
        </section>
      </>
    );
  }

  function renderMasterCustomers() {
    const customers = bootstrap?.customers || [];
    return (
      <section className="panel">
        <DataTable
          title="客户台账"
          data={customers}
          rowKey={(item) => item.id}
          emptyText="暂无客户"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" onClick={() => openMasterCreateDialog("customer")}><Plus size={15} />新增客户</button>
            </div>
          )}
          columns={[
            { key: "name", title: "客户", render: (item) => item.name },
            { key: "contact", title: "联系人", render: (item) => item.contact },
            { key: "phone", title: "电话", render: (item) => item.phone },
            { key: "creditLimit", title: "信用额度", render: (item) => money(item.creditLimit) },
            { key: "receivable", title: "应收", render: (item) => money(item.receivable) },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("customer", item.id, item) }
          ]}
        />
      </section>
    );
  }

  function renderMasterProjects() {
    const customers = bootstrap?.customers || [];
    const projects = bootstrap?.projects || [];
    return (
      <section className="panel">
        <DataTable
          title="项目台账"
          data={projects}
          rowKey={(item) => item.id}
          emptyText="暂无项目"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" disabled={!customers.length} onClick={() => openMasterCreateDialog("project")}><Plus size={15} />新增项目</button>
            </div>
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
      </section>
    );
  }

  function renderMasterProducts() {
    const products = bootstrap?.products || [];
    return (
      <section className="panel">
        <DataTable
          title="产品台账"
          data={products}
          rowKey={(item) => item.id}
          emptyText="暂无产品"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" onClick={() => openMasterCreateDialog("product")}><Plus size={15} />新增产品</button>
            </div>
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
      </section>
    );
  }

  function renderMasterMaterials() {
    const materials = bootstrap?.materials || [];
    return (
      <section className="panel">
        <DataTable
          title="物料台账"
          data={materials}
          rowKey={(item) => item.id}
          emptyText="暂无物料"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" onClick={() => openMasterCreateDialog("material")}><Plus size={15} />新增物料</button>
            </div>
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
      </section>
    );
  }

  function renderMasterSites() {
    const sites = bootstrap?.sites || [];
    return (
      <section className="panel">
        <DataTable
          title="站点台账"
          data={sites}
          rowKey={(item) => item.id}
          emptyText="暂无站点"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" onClick={() => openMasterCreateDialog("site")}><Plus size={15} />新增站点</button>
            </div>
          )}
          columns={[
            { key: "name", title: "站点", render: (item) => item.name },
            { key: "code", title: "编码", render: (item) => item.code },
            { key: "address", title: "地址", render: (item) => item.address },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("site", item.id, item) }
          ]}
        />
      </section>
    );
  }

  function renderMasterDrivers() {
    const drivers = bootstrap?.drivers || [];
    return (
      <section className="panel">
        <DataTable
          title="司机台账"
          data={drivers}
          rowKey={(item) => item.id}
          emptyText="暂无司机"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" onClick={() => openMasterCreateDialog("driver")}><Plus size={15} />新增司机</button>
            </div>
          )}
          columns={[
            { key: "name", title: "司机", render: (item) => item.name },
            { key: "phone", title: "电话", render: (item) => item.phone },
            { key: "licenseNo", title: "证号", render: (item) => item.licenseNo },
            { key: "licenseExpire", title: "到期", render: (item) => item.licenseExpire },
            { key: "status", title: "状态", render: (item) => <StatusChip value={item.status} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("driver", item.id, item) }
          ]}
        />
      </section>
    );
  }

  function renderMasterVehicles() {
    const vehicles = bootstrap?.vehicles || [];
    const drivers = bootstrap?.drivers || [];
    const sites = bootstrap?.sites || [];
    return (
      <section className="panel">
        <DataTable
          title="车辆台账"
          data={vehicles}
          rowKey={(item) => item.id}
          emptyText="暂无车辆"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" disabled={!drivers.length || !sites.length} onClick={() => openMasterCreateDialog("vehicle")}><Plus size={15} />新增车辆</button>
            </div>
          )}
          columns={[
            { key: "plateNo", title: "车牌", render: (item) => item.plateNo },
            { key: "vehicleType", title: "车型", render: (item) => item.vehicleType },
            { key: "driver", title: "司机", render: (item) => nameOf(drivers, item.driverId) },
            { key: "site", title: "站点", render: (item) => nameOf(sites, item.siteId) },
            { key: "onlineStatus", title: "在线", render: (item) => <StatusChip value={item.onlineStatus} /> },
            { key: "businessStatus", title: "业务", render: (item) => <StatusChip value={item.businessStatus} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("vehicle", item.id, item) }
          ]}
        />
      </section>
    );
  }

  function renderMasterInventory() {
    const inventory = bootstrap?.inventory || [];
    const sites = bootstrap?.sites || [];
    const materials = bootstrap?.materials || [];
    return (
      <section className="panel">
        <DataTable
          title="库存台账"
          data={inventory}
          rowKey={(item) => item.id}
          emptyText="暂无库存"
          pageSize={8}
          headerAction={(
            <div className="row-actions">
            {reloadButton()}
            <button className="primary-button icon-button-text" type="button" disabled={!sites.length || !materials.length} onClick={() => openMasterCreateDialog("inventory")}><Plus size={15} />新增库存</button>
            </div>
          )}
          columns={[
            { key: "site", title: "站点", render: (item) => nameOf(sites, item.siteId) },
            { key: "warehouse", title: "仓库", render: (item) => item.warehouse },
            { key: "material", title: "物料", render: (item) => nameOf(materials, item.materialId) },
            { key: "quantity", title: "数量", render: (item) => `${qty(item.quantity)} ${item.unit}` },
            { key: "qualityStatus", title: "质量", render: (item) => <StatusChip value={item.qualityStatus} /> },
            { key: "availableStatus", title: "可用", render: (item) => <StatusChip value={item.availableStatus} /> },
            { key: "actions", title: "操作", render: (item) => masterRowActions("inventory", item.id, item) }
          ]}
        />
      </section>
    );
  }

  function renderOrders() {
    const customers = bootstrap?.customers || [];
    const projects = bootstrap?.projects || [];
    const products = bootstrap?.products || [];
    const sites = bootstrap?.sites || [];
    const canCreateOrder = customers.length > 0 && projects.length > 0 && products.length > 0 && sites.length > 0;
    return (
      <>
        <section className="kpi-grid compact">
          <KpiCard label="订单总数" value={data.orders.length} />
          <KpiCard label="执行中" value={activeOrders.length} />
          <KpiCard label="计划方量" value={qty(data.orders.reduce((sum, item) => sum + item.planQuantity, 0))} />
          <KpiCard label="已签收" value={qty(data.orders.reduce((sum, item) => sum + item.signedQty, 0))} />
          <KpiCard label="订单金额" value={money(data.orders.reduce((sum, item) => sum + item.totalAmount, 0))} />
        </section>
        <section className="panel">
          {ordersTable(data.orders, "销售订单", true, (
            <div className="row-actions">
              {reloadButton()}
              <button className="primary-button icon-button-text" type="button" disabled={!canCreateOrder} onClick={openOrderDialog}>
                <Plus size={14} />新增订单
              </button>
            </div>
          ))}
        </section>
      </>
    );
  }

  function renderProduction() {
    const plans = list(data.production?.plans);
    const tasks = list(data.production?.tasks);
    const batches = list(data.production?.batches);
    return (
      <section className="grid-12">
        <div className="panel span-4">
          <h3>生产指标</h3>
          <div className="metric-list">
            <div><span>生产计划</span><b>{plans.length}</b></div>
            <div><span>生产任务</span><b>{tasks.length}</b></div>
            <div><span>生产批次</span><b>{batches.length}</b></div>
            <div><span>已生产</span><b>{qty(tasks.reduce((sum, item) => sum + item.producedQty, 0))}</b></div>
          </div>
        </div>
        <div className="panel span-8">
          <h3>生产任务</h3>
          <table>
            <thead><tr><th>任务</th><th>站点</th><th>产品</th><th>计划/已产</th><th>状态</th></tr></thead>
            <tbody>
              {tasks.map((item) => (
                <tr key={item.id}>
                  <td>{item.taskNo}</td>
                  <td>{nameOf(bootstrap?.sites, item.siteId)}</td>
                  <td>{productLabel(bootstrap, item.productId)}</td>
                  <td>{qty(item.planQty)} / {qty(item.producedQty)}</td>
                  <td><StatusChip value={item.status} /></td>
                </tr>
              ))}
              {!tasks.length ? <tr><td colSpan={5}>暂无生产任务</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    );
  }

  function renderDispatch() {
    const dispatch = data.dispatch;
    const kpis = dispatch?.kpis;
    const allProgress = list(dispatch?.siteProgress);
    const keyword = dispatchSearch.trim();
    const sites = dispatchSiteOptions(allProgress);
    const progressRows = allProgress.filter((item) => (siteFilter === "all" || String(item.siteId) === siteFilter) && matchesDispatchSearch(item, keyword));
    const allVehicles = list(dispatch?.availableVehicles);
    const vehicles = allVehicles.filter((item) => (siteFilter === "all" || String(item.siteId) === siteFilter) && matchesDispatchSearch(item, keyword));
    const vehicleOptions = vehicles.length ? vehicles : allVehicles;
    const queueRows = list(dispatch?.vehicleQueue).filter((item) => (siteFilter === "all" || String(item.siteId) === siteFilter) && matchesDispatchSearch(item, keyword));
    const productionTasks = list(dispatch?.productionTasks).filter((item) => (siteFilter === "all" || String(item.siteId) === siteFilter) && matchesDispatchSearch(item, keyword));
    const locations = list(dispatch?.latestLocations).filter((item) => siteFilter === "all" || String(item.currentSiteId) === siteFilter);
    const selectedOrder = progressRows.find((item) => item.orderId === selectedOrderId) || progressRows[0] || allProgress.find((item) => item.orderId === selectedOrderId) || allProgress[0];
    const selectedVehicle = vehicleOptions.find((item) => item.vehicleId === selectedVehicleId) || vehicleOptions[0];
    const selectedQueue = selectedOrder ? queueRows.filter((item) => item.orderId === selectedOrder.orderId) : queueRows;
    const orderOptions = progressRows.length ? progressRows : allProgress;
    const defaultDispatchQty = selectedOrder ? Math.min(36, selectedOrder.remainingQty || selectedOrder.planQuantity || 0) : 0;

    return (
      <section className="dispatch-center">
        {dispatchActionError ? (
          <div className="dispatch-error">
            <p className="error-text">{dispatchActionError}</p>
          </div>
        ) : null}
        <div className="panel dispatch-board dispatch-window">
          <div className="dispatch-board-toolbar">
            <div>
              <h3>调度中心</h3>
              <p className="muted">订单、生产、车辆、排队和位置联动调度</p>
            </div>
            <div className="dispatch-board-controls">
              <label className="compact-field">
                <Search size={14} />
                <input
                  className="compact-input"
                  value={dispatchSearch}
                  onChange={(event) => setDispatchSearch(event.target.value)}
                  placeholder="订单 / 工地 / 车辆"
                />
              </label>
              <select className="compact-select wide" value={siteFilter} onChange={(event) => setSiteFilter(event.target.value)}>
                <option value="all">全部站点</option>
                {sites.map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
              </select>
              <button className="soft-button icon-button-text" onClick={() => { onChanged(); load(); }} disabled={dispatchSubmitting} title="刷新调度数据">
                <RefreshCw size={14} />刷新
              </button>
            </div>
            <div className="dispatch-kpi-strip">
              <div className="compact-kpi"><span>在线车辆</span><b>{kpis?.onlineVehicles || 0}/{kpis?.totalVehicles || 0}</b></div>
              <div className="compact-kpi"><span>排队</span><b>{kpis?.queueVehicles || 0}</b></div>
              <div className="compact-kpi"><span>装料</span><b>{kpis?.loadingVehicles || 0}</b></div>
              <div className="compact-kpi"><span>运输</span><b>{kpis?.inTransitVehicles || 0}</b></div>
              <div className="compact-kpi"><span>供货单</span><b>{kpis?.openSupplyOrders || 0}</b></div>
            </div>
          </div>
          <div className={`dispatch-board-grid${vehiclesCollapsed ? " vehicles-collapsed" : ""}`}>
            <aside className={`dispatch-yard${vehiclesCollapsed ? " collapsed" : ""}`}>
              {vehiclesCollapsed ? (
                <button className="rail-toggle" onClick={() => setVehiclesCollapsed(false)} title="展开车辆池">
                  <ChevronRight size={14} />车辆池
                </button>
              ) : (
                <>
                  <div className="panel-head-compact">
                    <div>
                      <b>车辆池</b>
                      <span>{vehicleOptions.length} 台可调车辆</span>
                    </div>
                    <button className="icon-only-button" onClick={() => setVehiclesCollapsed(true)} title="收起车辆池">
                      <ChevronLeft size={14} />
                    </button>
                  </div>
                  <form className="dispatch-quick-form" onSubmit={handleQuickDispatch}>
                    <label>
                      供货订单
                      <select value={selectedOrder?.orderId || ""} onChange={(event) => setSelectedOrderId(Number(event.target.value))}>
                        {orderOptions.map((item) => (
                          <option key={item.orderId} value={item.orderId}>{item.orderNo} / {item.projectName}</option>
                        ))}
                      </select>
                    </label>
                    <label>
                      可用车辆
                      <select value={selectedVehicle?.vehicleId || ""} onChange={(event) => setSelectedVehicleId(Number(event.target.value))}>
                        {vehicleOptions.map((item) => (
                          <option key={item.vehicleId} value={item.vehicleId}>{item.plateNo} / {item.driverName}</option>
                        ))}
                      </select>
                    </label>
                    <div className="dispatch-inline-action">
                      <label>
                        方量
                        <input
                          type="number"
                          min="0"
                          step="0.5"
                          value={dispatchQty}
                          onChange={(event) => setDispatchQty(event.target.value)}
                          placeholder={defaultDispatchQty ? String(defaultDispatchQty) : "0"}
                        />
                      </label>
                      <button className="primary-button icon-button-text" type="submit" disabled={dispatchSubmitting || !selectedOrder || !selectedVehicle}>
                        <Plus size={14} />派车
                      </button>
                    </div>
                    <div className="quick-form-summary">
                      <span>可派 {qty(selectedOrder?.remainingQty)} {selectedOrder?.unit || "m3"}</span>
                      <span>排班余量 {qty(selectedVehicle?.scheduleRemaining)} / {selectedVehicle?.scheduleNo || "未排班"}</span>
                    </div>
                  </form>
                  <div className="yard-section">
                    <div className="yard-head"><b>空闲车辆</b><span>在线优先</span></div>
                    <div className="vehicle-token-grid">
                      {vehicleOptions.map((item) => (
                        <button
                          className={`vehicle-token${selectedVehicle?.vehicleId === item.vehicleId ? " selected" : ""}`}
                          key={item.vehicleId}
                          onClick={() => setSelectedVehicleId(item.vehicleId)}
                          type="button"
                        >
                          <span className={`vehicle-light ${item.onlineStatus === "online" ? "online" : "offline"}`} />
                          <b>{item.plateNo}</b>
                          <small>{item.driverName} / {item.capacity || item.vehicleType}</small>
                          <small>{item.scheduleNo || "未排班"} · 余量 {qty(item.scheduleRemaining)}</small>
                        </button>
                      ))}
                      {!vehicleOptions.length ? <p className="muted">暂无可调车辆</p> : null}
                    </div>
                  </div>
                </>
              )}
            </aside>
            <main className="dispatch-visual-main">
              <div className="production-strip">
                <b>生产联动</b>
                <small>{productionTasks.length ? `${productionTasks.length} 个生产任务待调度，最近任务 ${productionTasks[0].taskNo} / ${productionTasks[0].productName}` : "暂无待联动生产任务"}</small>
              </div>
              <div className="dispatch-card-grid">
                {progressRows.map((item) => {
                  const itemQueue = queueRows.filter((queue) => queue.orderId === item.orderId);
                  return (
                    <button
                      className={`dispatch-visual-card${selectedOrder?.orderId === item.orderId ? " selected" : ""}`}
                      key={item.orderId}
                      type="button"
                      onClick={() => setSelectedOrderId(item.orderId)}
                    >
                      <div className="visual-card-head">
                        <div>
                          <b>{item.orderNo} / {item.customerName}</b>
                          <span>{item.projectName} · {item.productName}</span>
                        </div>
                        <StatusChip value={item.status} />
                      </div>
                      <div className="visual-card-metrics">
                        <div><span>计划</span><b>{qty(item.planQuantity)}</b></div>
                        <div><span>生产</span><b>{percent(item.producedPercent)}</b></div>
                        <div><span>派发</span><b>{percent(item.dispatchedPercent)}</b></div>
                        <div><span>签收</span><b>{percent(item.signedPercent)}</b></div>
                      </div>
                      <div className="dispatch-lane">
                        <div className="lane-line">
                          <span className="lane-node lane-origin"><Factory size={12} />站点</span>
                          <span className="lane-node lane-destination"><MapPin size={12} />工地</span>
                          {itemQueue.length ? itemQueue.slice(0, 4).map((queue, index) => (
                            <span
                              className={`lane-vehicle ${dispatchStageClass(queue.status)}`}
                              key={queue.dispatchId}
                              title={`${queue.plateNo} · ${queue.etaDistanceKm || 0}km · ${etaText(queue)}`}
                              style={{ left: `${dispatchRoutePosition(queue.status)}%`, top: `${7 + (index % 2) * 15}px` }}
                            >
                              <Truck size={12} />{queue.plateNo}
                            </span>
                          )) : <span className="lane-empty-marker">待派车</span>}
                        </div>
                        <div className="lane-stage-labels">
                          <span>{item.siteName}</span><span>排队</span><span>装料</span><span>出厂</span><span>在途</span><span>{item.projectName}</span>
                        </div>
                      </div>
                      <div className="visual-card-foot">
                        <span>{item.siteName} · 剩余 {qty(item.remainingQty)} {item.unit}</span>
                        <span className="eta-pill"><Clock size={12} /> {shortDateTime(item.nextEta)}</span>
                      </div>
                    </button>
                  );
                })}
                {!progressRows.length ? (
                  <div className="empty-visual-board">
                    <p className="muted">暂无匹配的供货任务</p>
                  </div>
                ) : null}
              </div>
            </main>
            <aside className="dispatch-side-panel">
              <div className="side-card">
                <button className="side-card-head" type="button">
                  <span>
                    <b>{selectedOrder?.projectName || "未选择订单"}</b>
                    <small>{selectedOrder?.customerName || "选择供货任务查看明细"}</small>
                  </span>
                  {selectedOrder ? <StatusChip value={selectedOrder.status} /> : null}
                </button>
                {selectedOrder ? (
                  <div className="side-card-body">
                    <div className="dispatch-mini-card">
                      <div className="mini-eta">
                        <Route size={13} />
                        <span>{selectedOrder.siteName} → {selectedOrder.projectName}</span>
                        <em>{shortDateTime(selectedOrder.nextEta)}</em>
                      </div>
                      <div className="visual-card-metrics">
                        <div><span>已装</span><b>{qty(selectedOrder.loadedQty)}</b></div>
                        <div><span>已签</span><b>{qty(selectedOrder.signedQty)}</b></div>
                        <div><span>队列</span><b>{selectedOrder.queueVehicles}</b></div>
                        <div><span>在途</span><b>{selectedOrder.inTransitVehicles}</b></div>
                      </div>
                    </div>
                  </div>
                ) : null}
              </div>
              <div className="side-card">
                <button className="side-card-head" type="button">
                  <span><b>调度队列</b><small>{selectedQueue.length} 单执行中</small></span>
                </button>
                <div className="side-card-body">
                  <div className="dispatch-mini-list">
                    {selectedQueue.slice(0, 8).map((item) => (
                      <div className="dispatch-mini-card" key={item.dispatchId}>
                        <div className="dispatch-mini">
                          <span className={`mini-status ${dispatchStageClass(item.status)}`} />
                          <b>{item.plateNo} / {item.dispatchNo}</b>
                          <small>{item.driverName} · {item.productName}</small>
                        </div>
                        <div className="mini-eta">
                          <Clock size={13} />
                          <span>{etaText(item)} · {item.etaTarget}</span>
                          <em>{item.etaConfidence}</em>
                        </div>
                        <div className="mini-actions">
                          <button className="soft-button icon-button-text" onClick={() => handleAdvanceDispatch(item)} disabled={dispatchSubmitting}>
                            <PlayCircle size={13} />{nextDispatchAction(item.status)}
                          </button>
                        </div>
                      </div>
                    ))}
                    {!selectedQueue.length ? <p className="muted">暂无执行中的派车单</p> : null}
                  </div>
                </div>
              </div>
              <div className="side-card">
                <button className="side-card-head" type="button">
                  <span><b>车辆位置</b><small>{locations.length} 条最新定位</small></span>
                </button>
                <div className="side-card-body">
                  <div className="location-chip-list">
                    {locations.slice(0, 6).map((item) => (
                      <div className="location-chip" key={item.vehicleId}>
                        <b><MapPin size={13} /> {item.plateNo}</b>
                        <StatusChip value={item.onlineStatus} />
                        <span>{qty(item.speed)} km/h · {shortDateTime(item.lastLocationTime)} · {item.latitude.toFixed(4)}, {item.longitude.toFixed(4)}</span>
                      </div>
                    ))}
                    {!locations.length ? <p className="muted">暂无车辆定位</p> : null}
                  </div>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </section>
    );
  }

  function renderWeighbridge() {
    return (
      <section className="grid-12">
        <div className="panel span-8">
          <h3>地磅票据</h3>
          <table>
            <thead><tr><th>票据</th><th>车牌</th><th>净重</th><th>签收</th><th>状态</th></tr></thead>
            <tbody>
              {data.tickets.map((item) => (
                <tr key={item.id}>
                  <td>{item.ticketNo}</td>
                  <td>{item.plateNo}</td>
                  <td>{qty(item.netWeight)} {item.unit}</td>
                  <td><StatusChip value={item.signStatus} /></td>
                  <td><StatusChip value={item.status} /></td>
                </tr>
              ))}
              {!data.tickets.length ? <tr><td colSpan={5}>暂无地磅票据</td></tr> : null}
            </tbody>
          </table>
        </div>
        <div className="panel span-4">
          <h3>称重记录</h3>
          <div className="record-list">
            {data.weightRecords.slice(0, 8).map((item) => (
              <div className="record-card" key={item.id}>
                <strong>{item.plateNo} / {item.weightType}</strong>
                <p>{qty(item.weight)} kg / {item.createdAt}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  function renderDelivery() {
    return (
      <section className="grid-12">
        <div className="panel span-6">
          <h3>工地签收</h3>
          <table>
            <thead><tr><th>签收单</th><th>产品</th><th>签收人</th><th>方量</th><th>时间</th></tr></thead>
            <tbody>
              {data.signs.map((item) => (
                <tr key={item.id}>
                  <td>{item.signNo}</td>
                  <td>{item.productName}</td>
                  <td>{item.signer}</td>
                  <td>{qty(item.signedQty)}</td>
                  <td>{item.signedAt}</td>
                </tr>
              ))}
              {!data.signs.length ? <tr><td colSpan={5}>暂无签收记录</td></tr> : null}
            </tbody>
          </table>
        </div>
        <div className="panel span-6">
          <h3>签收链接</h3>
          <table>
            <thead><tr><th>链接</th><th>渠道</th><th>手机号</th><th>过期时间</th><th>状态</th></tr></thead>
            <tbody>
              {data.signLinks.map((item) => (
                <tr key={item.id}>
                  <td>{item.linkNo}</td>
                  <td>{item.channel}</td>
                  <td>{item.phone}</td>
                  <td>{item.expiresAt}</td>
                  <td><StatusChip value={item.status} /></td>
                </tr>
              ))}
              {!data.signLinks.length ? <tr><td colSpan={5}>暂无签收链接</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    );
  }

  function renderSettlement() {
    return (
      <section className="panel">
        <h3>客户对账</h3>
        <table>
          <thead><tr><th>对账单</th><th>客户</th><th>项目</th><th>周期</th><th>方量</th><th>金额</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {data.statements.map((item) => (
              <tr key={item.id}>
                <td>{item.statementNo}</td>
                <td>{nameOf(bootstrap?.customers, item.customerId)}</td>
                <td>{nameOf(bootstrap?.projects, item.projectId)}</td>
                <td>{item.period}</td>
                <td>{qty(item.totalQty)}</td>
                <td>{money(item.totalAmount)}</td>
                <td><StatusChip value={item.status} /></td>
                <td>
                  {item.status !== "confirmed" ? (
                    <button className="soft-button icon-button-text" type="button" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`statement-${item.id}`, "对账单已确认", () => api.confirmStatement(item.id))}>
                      <CheckCircle2 size={13} />确认
                    </button>
                  ) : <span className="muted">已确认</span>}
                </td>
              </tr>
            ))}
            {!data.statements.length ? <tr><td colSpan={8}>暂无对账单</td></tr> : null}
          </tbody>
        </table>
      </section>
    );
  }

  function renderProcurement() {
    const procurement = data.procurement;
    const suppliers = supplierOptions();
    const purchaseOrders = list(procurement?.orders);
    const sites = bootstrap?.sites || [];
    const materials = bootstrap?.materials || [];
    return (
      <section className="grid-12">
        <div className="panel span-12">
          <h3>原料入库</h3>
          <form className="form-grid four" onSubmit={handleCreateRawReceipt}>
            <label>
              <span>采购订单</span>
              <select value={procurementForm.purchaseOrderId} onChange={(event) => setProcurementForm({ ...procurementForm, purchaseOrderId: event.target.value })}>
                <option value="">无采购订单</option>
                {purchaseOrders.map((item) => <option key={item.id} value={item.id}>{item.orderNo}</option>)}
              </select>
            </label>
            <label>
              <span>供应商</span>
              <select value={procurementForm.supplierId} onChange={(event) => setProcurementForm({ ...procurementForm, supplierId: event.target.value })}>
                {suppliers.map((item, index) => <option key={recordId(item) || index} value={recordId(item)}>{recordName(item, `供应商 ${index + 1}`)}</option>)}
              </select>
            </label>
            <label>
              <span>站点</span>
              <select value={procurementForm.siteId} onChange={(event) => setProcurementForm({ ...procurementForm, siteId: event.target.value })}>
                {sites.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>物料</span>
              <select value={procurementForm.materialId} onChange={(event) => setProcurementForm({ ...procurementForm, materialId: event.target.value })}>
                {materials.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>车牌</span><input value={procurementForm.plateNo} onChange={(event) => setProcurementForm({ ...procurementForm, plateNo: event.target.value })} /></label>
            <label><span>毛重</span><input type="number" min="0" step="0.01" value={procurementForm.grossWeight} onChange={(event) => setProcurementForm({ ...procurementForm, grossWeight: event.target.value })} /></label>
            <label><span>皮重</span><input type="number" min="0" step="0.01" value={procurementForm.tareWeight} onChange={(event) => setProcurementForm({ ...procurementForm, tareWeight: event.target.value })} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={actionBusy !== "" || !suppliers.length || !sites.length || !materials.length}>
              <Plus size={14} />生成入库
            </button>
          </form>
        </div>
        <div className="panel span-6">
          <h3>采购订单</h3>
          <table>
            <thead><tr><th>订单</th><th>物料</th><th>数量</th><th>单价</th><th>状态</th></tr></thead>
            <tbody>
              {list(procurement?.orders).map((item) => (
                <tr key={item.id}>
                  <td>{item.orderNo}</td>
                  <td>{nameOf(bootstrap?.materials, item.materialId)}</td>
                  <td>{qty(item.quantity)} {item.unit}</td>
                  <td>{money(item.unitPrice)}</td>
                  <td><StatusChip value={item.status} /></td>
                </tr>
              ))}
              {!list(procurement?.orders).length ? <tr><td colSpan={5}>暂无采购订单</td></tr> : null}
            </tbody>
          </table>
        </div>
        <div className="panel span-6">
          <h3>原料入库</h3>
          <table>
            <thead><tr><th>入库单</th><th>物料</th><th>车牌</th><th>净重</th><th>质检</th></tr></thead>
            <tbody>
              {list(procurement?.receipts).map((item) => (
                <tr key={item.id}>
                  <td>{item.receiptNo}</td>
                  <td>{nameOf(bootstrap?.materials, item.materialId)}</td>
                  <td>{item.plateNo}</td>
                  <td>{qty(item.netWeight)}</td>
                  <td><StatusChip value={item.qualityStatus} /></td>
                </tr>
              ))}
              {!list(procurement?.receipts).length ? <tr><td colSpan={5}>暂无入库单</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    );
  }

  function renderFinance() {
    const finance = data.finance;
    const receivables = openReceivables();
    const selectedReceivable = receivables.find((item) => item.id === fieldNumber(financeForm.receivableId)) || receivables[0];
    const paymentPlans = list(finance?.paymentPlans).filter((item) => item.status !== "settled");
    return (
      <section className="grid-12">
        <div className="panel span-4">
          <h3>财务指标</h3>
          <div className="metric-list">
            <div><span>应收余额</span><b>{money(receivableBalance(finance))}</b></div>
            <div><span>逾期金额</span><b>{money(financeRisk)}</b></div>
            <div><span>发票</span><b>{list(finance?.invoices).length}</b></div>
            <div><span>收款</span><b>{list(finance?.receipts).length}</b></div>
          </div>
        </div>
        <div className="panel span-8">
          <h3>收款与计划</h3>
          <form className="form-grid four" onSubmit={handleCreateReceipt}>
            <label className="span-two">
              <span>应收账款</span>
              <select value={financeForm.receivableId} onChange={(event) => setFinanceForm({ ...financeForm, receivableId: event.target.value })}>
                {receivables.map((item) => <option key={item.id} value={item.id}>{item.billNo} / {nameOf(bootstrap?.customers, item.customerId)} / 未收 {money(item.amount - item.receivedAmount)}</option>)}
              </select>
            </label>
            <label><span>收款金额</span><input type="number" min="0" step="1" value={financeForm.receiptAmount} onChange={(event) => setFinanceForm({ ...financeForm, receiptAmount: event.target.value })} placeholder={String(Math.max(0, (selectedReceivable?.amount || 0) - (selectedReceivable?.receivedAmount || 0)) || "")} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={actionBusy !== "" || !receivables.length}>
              <CreditCard size={14} />登记收款
            </button>
          </form>
          <div className="lab-form-divider" />
          <form className="form-grid four" onSubmit={handleCreatePaymentPlan}>
            <label className="span-two">
              <span>应收账款</span>
              <select value={financeForm.receivableId} onChange={(event) => setFinanceForm({ ...financeForm, receivableId: event.target.value })}>
                {receivables.map((item) => <option key={item.id} value={item.id}>{item.billNo} / {nameOf(bootstrap?.customers, item.customerId)}</option>)}
              </select>
            </label>
            <label><span>计划金额</span><input type="number" min="0" step="1" value={financeForm.planAmount} onChange={(event) => setFinanceForm({ ...financeForm, planAmount: event.target.value })} /></label>
            <label><span>到期日</span><input type="date" value={financeForm.planDueDate} onChange={(event) => setFinanceForm({ ...financeForm, planDueDate: event.target.value })} /></label>
            <button className="soft-button icon-button-text" type="submit" disabled={actionBusy !== "" || !receivables.length}>
              <Plus size={14} />创建计划
            </button>
          </form>
          {paymentPlans.length ? (
            <div className="row-actions compact-actions">
              {paymentPlans.slice(0, 4).map((item) => (
                <button className="soft-button icon-button-text" type="button" key={item.id} disabled={actionBusy !== ""} onClick={() => runBusinessAction(`payment-plan-${item.id}`, "付款计划已结清", () => api.settlePaymentPlan(item.id))}>
                  <CheckCircle2 size={13} />结清 {item.planNo}
                </button>
              ))}
            </div>
          ) : null}
        </div>
        <div className="panel span-8">
          <h3>应收账款</h3>
          <table>
            <thead><tr><th>账单</th><th>客户</th><th>金额</th><th>已收</th><th>到期日</th><th>状态</th></tr></thead>
            <tbody>
              {list(finance?.receivables).map((item) => (
                <tr key={item.id}>
                  <td>{item.billNo}</td>
                  <td>{nameOf(bootstrap?.customers, item.customerId)}</td>
                  <td>{money(item.amount)}</td>
                  <td>{money(item.receivedAmount)}</td>
                  <td>{item.dueDate}</td>
                  <td><StatusChip value={item.status} /></td>
                </tr>
              ))}
              {!list(finance?.receivables).length ? <tr><td colSpan={6}>暂无应收账款</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    );
  }

  function renderReports() {
    const reports = data.reports;
    return (
      <section className="grid-12">
        <div className="panel span-4">
          <h3>经营分析</h3>
          <div className="metric-list">
            <div><span>收入</span><b>{money(reports?.operating.revenue)}</b></div>
            <div><span>总成本</span><b>{money(reports?.operating.totalCost)}</b></div>
            <div><span>毛利</span><b>{money(reports?.operating.grossProfit)}</b></div>
            <div><span>毛利率</span><b>{percent(reports?.operating.grossMargin)}</b></div>
          </div>
        </div>
        <div className="panel span-4">
          <h3>库存预警</h3>
          <div className="record-list">
            {list(reports?.inventoryWarnings).slice(0, 8).map((item) => (
              <div className="record-card" key={item.id}>
                <strong>{nameOf(bootstrap?.materials, item.materialId)}</strong>
                <p>{nameOf(bootstrap?.sites, item.siteId)} / 当前 {qty(item.quantity)} {item.unit} / {item.availableStatus}</p>
              </div>
            ))}
          </div>
        </div>
        <div className="panel span-4">
          <h3>账龄风险</h3>
          <div className="record-list">
            {list(reports?.customerAging).slice(0, 8).map((item) => (
              <div className="record-card" key={item.customerId}>
                <strong>{item.customerName}</strong>
                <p>逾期 {money(item.overdueTotal)} / 合计 {money(item.total)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  function renderSystem() {
    return (
      <section className="grid-12">
        <div className="panel span-5">
          <h3>待办审批</h3>
          <label className="approval-comment">
            <span>审批意见</span>
            <input value={approvalComment} onChange={(event) => setApprovalComment(event.target.value)} />
          </label>
          <div className="record-list">
            {openApprovals.map((item) => (
              <div className="record-card" key={item.id}>
                <div className="between">
                  <strong>{item.title}</strong>
                  <StatusChip value={item.status} />
                </div>
                <p>{item.resourceNo} / 当前角色 {item.currentRole}</p>
                <div className="row-actions">
                  <button className="primary-button icon-button-text" type="button" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-approve-${item.id}`, "审批已通过", () => api.actApproval(item.id, "approve", approvalComment))}>
                    <CheckCircle2 size={13} />通过
                  </button>
                  <button className="soft-button icon-button-text" type="button" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`approval-reject-${item.id}`, "审批已驳回", () => api.actApproval(item.id, "reject", approvalComment))}>
                    驳回
                  </button>
                </div>
              </div>
            ))}
            {!openApprovals.length ? <p className="muted">暂无待办审批</p> : null}
          </div>
        </div>
        <div className="panel span-7">
          <h3>已开通模块</h3>
          <div className="capability-grid">
            {list(bootstrap?.modules).map((item) => (
              <div className="capability-card" key={item.code}>
                <strong>{item.name}</strong>
                <span>{item.area}</span>
                <StatusChip value={item.enabled ? "active" : "disabled"} />
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  function ordersTable(orders: SalesOrder[], title: string, withActions = false, headerAction?: ReactNode) {
    return (
      <DataTable
        title={title}
        data={orders}
        rowKey={(order) => order.id}
        emptyText="暂无订单"
        pageSize={withActions ? 10 : 8}
        showPagination={withActions}
        headerAction={headerAction || <BarChart3 size={18} />}
        columns={[
          { key: "orderNo", title: "订单", render: (order) => order.orderNo },
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
          { key: "status", title: "状态", render: (order) => <StatusChip value={order.status} /> },
          ...(withActions ? [{
            key: "actions",
            title: "操作",
            render: (order: SalesOrder) => order.status === "submitted" ? (
              <button className="soft-button icon-button-text" type="button" disabled={actionBusy !== ""} onClick={() => runBusinessAction(`order-approve-${order.id}`, "订单已审批", () => api.approveOrder(order.id))}>
                <CheckCircle2 size={13} />审批
              </button>
            ) : <span className="muted">-</span>
          }] : [])
        ]}
      />
    );
  }

  function dispatchSummary() {
    return (
      <>
        <h3>调度现场</h3>
        <div className="metric-list">
          <div><span>在线车辆</span><b>{data.dispatch?.kpis?.onlineVehicles || 0}/{data.dispatch?.kpis?.totalVehicles || 0}</b></div>
          <div><span>排队车辆</span><b>{data.dispatch?.kpis?.queueVehicles || 0}</b></div>
          <div><span>运输中</span><b>{data.dispatch?.kpis?.inTransitVehicles || 0}</b></div>
          <div><span>执行中调度单</span><b>{data.dispatch?.kpis?.activeDispatches || 0}</b></div>
        </div>
      </>
    );
  }
}
