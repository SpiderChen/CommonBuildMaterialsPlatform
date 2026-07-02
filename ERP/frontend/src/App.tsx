import {
  BarChart3,
  Bell,
  Building2,
  ChevronDown,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  Copy,
  CreditCard,
  ExternalLink,
  HardHat,
  Factory,
  FileSignature,
  FileWarning,
  FileText,
  FlaskConical,
  Home,
  Languages,
  Landmark,
  Layers,
  LogOut,
  Map as MapIcon,
  Menu,
  Monitor,
  Moon,
  Package,
  Palette,
  ReceiptText,
  RefreshCw,
  Route,
  Scale,
  Search,
  Server,
  Settings,
  ShieldCheck,
  ShoppingCart,
  Sun,
  Truck,
  UserCircle,
  Wrench,
  ListChecks,
  Users,
  X
} from "lucide-react";
import { Component, CSSProperties, ErrorInfo, FormEvent, MouseEvent as ReactMouseEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { api, eventURL } from "./services/api";
import type { BootstrapData, MFAEnrollment, OIDCProvider } from "./services/types";
import { ActionGroup, BareButton, Button, ChipList, ContextMenu, Dialog, Field, IconButton, IconField, LayoutRegion, LoginForm, MessageBox, Panel, SectionGrid, SectionHeader, SelectInput, TextInput, useMessage, useMessageBox } from "./components";
import type { ContextMenuItem } from "./components";
import { copyTextToClipboard } from "./components/ui/clipboard";
import { ERPWorkbenchView } from "./views/ERPWorkbenchView";
import type { ERPWorkbenchSection, WorkbenchMenuItem } from "./views/ERPWorkbenchView";
import { LaboratoryView } from "./views/LaboratoryView";
import type { LaboratoryModuleKey } from "./views/laboratory/LaboratoryModuleTypes";
import { PublicSignView } from "./views/PublicSignView";
import { SiteSigningModule } from "./views/SiteSigningModule";
import { hasPermission, permissionsForRole } from "./services/permissions";
import appIconUrl from "./assets/app-icon.svg";
import { InitialRoute, OpenStandaloneWindow } from "../wailsjs/go/main/DesktopApp";

type AccountRouteKey = "user-profile" | "account-security";
type ViewKey = ERPWorkbenchSection | LaboratoryModuleKey | "site-signing" | AccountRouteKey;
type Locale = "zh-CN" | "zh-TW" | "en-US";
type ConnectionMode = "local" | "server";
type LayoutMode = "side" | "top";
type ResolvedThemeMode = "highlight" | "night";
type ThemeMode = "system" | ResolvedThemeMode;
type NavGroupKey = "production" | "sales" | "laboratory" | "finance" | "fleet" | "system-settings";
type HistoryMode = "push" | "replace" | "none";
type DesktopBridgeWindow = Window & {
  runtime?: unknown;
  go?: {
    main?: {
      DesktopApp?: {
        InitialRoute?: () => Promise<string>;
        OpenStandaloneWindow?: (route: string) => Promise<void>;
      };
    };
  };
};

const nav = [
  { path: "/dashboard", name: "Dashboard", key: "overview", group: "production", icon: Building2, meta: { title: "overview", icon: "dashboard", affix: true, noCache: true, breadcrumb: true } },
  { path: "/fulfillment/production", name: "Production", key: "production", group: "production", icon: Factory, meta: { title: "production", icon: "factory", breadcrumb: true } },
  { path: "/fulfillment/production/plans", name: "ProductionPlans", key: "production-plans", group: "production", icon: ClipboardList, meta: { title: "productionPlans", icon: "clipboard-list", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/production/tasks", name: "ProductionTasks", key: "production-tasks", group: "production", icon: ListChecks, meta: { title: "productionTasks", icon: "list-checks", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/production/batches", name: "ProductionBatches", key: "production-batches", group: "production", icon: Factory, meta: { title: "productionBatches", icon: "factory", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/production/reports", name: "ProductionReports", key: "production-reports", group: "production", icon: BarChart3, meta: { title: "productionReports", icon: "chart", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/dispatch", name: "Dispatch", key: "dispatch", group: "production", icon: Truck, meta: { title: "dispatch", icon: "truck", breadcrumb: true } },
  { path: "/fulfillment/dispatch/schedules", name: "DispatchSchedules", key: "dispatch-schedules", group: "production", icon: ClipboardList, meta: { title: "dispatchSchedules", icon: "clipboard-list", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/dispatch/queue", name: "DispatchQueue", key: "dispatch-queue", group: "production", icon: Route, meta: { title: "dispatchQueue", icon: "route", breadcrumb: true, hidden: true } },
  { path: "/resources/products", name: "Products", key: "master-products", group: "production", icon: Package, meta: { title: "masterProducts", icon: "package", breadcrumb: true } },
  { path: "/resources/materials", name: "Materials", key: "master-materials", group: "production", icon: Layers, meta: { title: "masterMaterials", icon: "stack", breadcrumb: true } },
  { path: "/resources/sites", name: "SiteInfo", key: "master-sites", group: "system-settings", icon: Factory, meta: { title: "masterSites", icon: "factory", breadcrumb: true } },
  { path: "/resources/plants", name: "ProductionLines", key: "master-plants", group: "production", icon: Factory, meta: { title: "masterPlants", icon: "factory", breadcrumb: true } },
  { path: "/resources/stock-yards", name: "StockYards", key: "stock-yards", group: "production", icon: Layers, meta: { title: "stockYards", icon: "stack", breadcrumb: true } },
  { path: "/sales/orders", name: "SalesOrders", key: "orders", group: "sales", icon: ShoppingCart, meta: { title: "orders", icon: "shopping-cart", breadcrumb: true } },
  { path: "/sales/customers", name: "Customers", key: "master-customers", group: "sales", icon: Building2, meta: { title: "masterCustomers", icon: "building", breadcrumb: true } },
  { path: "/sales/customer-risk", name: "CustomerRisk", key: "customer-risk", group: "sales", icon: ShieldCheck, meta: { title: "customerRisk", icon: "shield", breadcrumb: true } },
  { path: "/sales/projects", name: "Projects", key: "master-projects", group: "sales", icon: HardHat, meta: { title: "masterProjects", icon: "hard-hat", breadcrumb: true } },
  { path: "/sales/pricing", name: "SalesPricing", key: "sales-pricing", group: "sales", icon: ReceiptText, meta: { title: "salesPricing", icon: "receipt", breadcrumb: true } },
  { path: "/portal/customer", name: "CustomerPortal", key: "portal-customer", group: "sales", icon: UserCircle, meta: { title: "customerPortal", icon: "user", breadcrumb: true } },
  { path: "/analytics/reports", name: "Reports", key: "reports", group: "sales", icon: BarChart3, meta: { title: "reports", icon: "chart", breadcrumb: true } },
  { path: "/quality/laboratory/exceptions", name: "LaboratoryExceptions", key: "exceptions", group: "laboratory", icon: FileWarning, meta: { title: "laboratoryExceptions", icon: "file-warning", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/site-signing", name: "DeliverySigning", key: "site-signing", group: "laboratory", icon: FileSignature, meta: { title: "site-signing", icon: "signature", breadcrumb: true } },
  { path: "/quality/laboratory/mix-designs", name: "LaboratoryMixes", key: "mix-designs", group: "laboratory", icon: FlaskConical, meta: { title: "laboratoryMixes", icon: "flask", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/plant-mix-designs", name: "LaboratoryPlantMixes", key: "plant-mix-designs", group: "laboratory", icon: Factory, meta: { title: "laboratoryPlantMixes", icon: "factory", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/trial-runs", name: "LaboratoryTrials", key: "trial-runs", group: "laboratory", icon: ClipboardList, meta: { title: "laboratoryTrials", icon: "clipboard-list", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/sample-tests", name: "LaboratorySampleTests", key: "sample-tests", group: "laboratory", icon: FlaskConical, meta: { title: "laboratorySampleTests", icon: "flask", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/equipment-calibration", name: "LaboratoryCalibration", key: "equipment-calibration", group: "laboratory", icon: Wrench, meta: { title: "laboratoryCalibration", icon: "wrench", noCache: true, breadcrumb: true } },
  { path: "/quality/laboratory/sample-ledger", name: "LaboratoryLedger", key: "sample-ledger", group: "laboratory", icon: ListChecks, meta: { title: "laboratoryLedger", icon: "list-checks", noCache: true, breadcrumb: true } },
  { path: "/resources/inventory/receipts", name: "RawMaterialReceipts", key: "raw-material-receipts", group: "production", icon: ReceiptText, meta: { title: "rawMaterialReceipts", icon: "receipt", breadcrumb: true, hidden: true } },
  { path: "/resources/inventory/transfers", name: "InventoryTransfers", key: "inventory-transfers", group: "production", icon: Route, meta: { title: "inventoryTransfers", icon: "route", breadcrumb: true, hidden: true } },
  { path: "/resources/inventory/stocktakes", name: "InventoryStocktakes", key: "inventory-stocktakes", group: "production", icon: ClipboardCheck, meta: { title: "inventoryStocktakes", icon: "clipboard", breadcrumb: true, hidden: true } },
  { path: "/quality/raw-material-inspections", name: "RawMaterialInspections", key: "raw-material-inspections", group: "laboratory", icon: ClipboardCheck, meta: { title: "rawMaterialInspections", icon: "clipboard", breadcrumb: true } },
  { path: "/finance/statements", name: "Statements", key: "settlement", group: "finance", icon: ReceiptText, meta: { title: "settlement", icon: "receipt", breadcrumb: true } },
  { path: "/finance/contracts", name: "Contracts", key: "contracts", group: "finance", icon: FileText, meta: { title: "contracts", icon: "file-text", breadcrumb: true } },
  { path: "/finance/settlement", name: "Finance", key: "finance", group: "finance", icon: Landmark, meta: { title: "finance", icon: "landmark", breadcrumb: true } },
  { path: "/finance/receivables", name: "Receivables", key: "finance-receivables", group: "finance", icon: CreditCard, meta: { title: "receivables", icon: "credit-card", breadcrumb: true, hidden: true } },
  { path: "/finance/invoices", name: "TaxInvoices", key: "finance-invoices", group: "finance", icon: ReceiptText, meta: { title: "taxInvoices", icon: "receipt", breadcrumb: true, hidden: true } },
  { path: "/finance/collections", name: "CollectionManagement", key: "finance-collections", group: "finance", icon: Bell, meta: { title: "collectionManagement", icon: "bell", breadcrumb: true, hidden: true } },
  { path: "/finance/suppliers", name: "SupplierStatements", key: "finance-suppliers", group: "finance", icon: ReceiptText, meta: { title: "supplierStatements", icon: "receipt", breadcrumb: true, hidden: true } },
  { path: "/finance/carrier-settlements", name: "CarrierSettlements", key: "finance-carriers", group: "finance", icon: Truck, meta: { title: "carrierSettlements", icon: "truck", breadcrumb: true, hidden: true } },
  { path: "/resources/drivers", name: "Drivers", key: "master-drivers", group: "fleet", icon: ClipboardCheck, meta: { title: "masterDrivers", icon: "clipboard", breadcrumb: true } },
  { path: "/resources/vehicles", name: "Vehicles", key: "master-vehicles", group: "fleet", icon: Truck, meta: { title: "masterVehicles", icon: "truck", breadcrumb: true } },
  { path: "/resources/carriers", name: "Carriers", key: "master-carriers", group: "fleet", icon: Route, meta: { title: "masterCarriers", icon: "route", breadcrumb: true } },
  { path: "/portal/driver", name: "DriverPortal", key: "portal-driver", group: "fleet", icon: UserCircle, meta: { title: "driverPortal", icon: "user", breadcrumb: true } },
  { path: "/fulfillment/delivery", name: "DeliveryNotes", key: "delivery", group: "fleet", icon: FileText, meta: { title: "delivery", icon: "file-text", breadcrumb: true } },
  { path: "/fulfillment/delivery/signs", name: "DeliverySigns", key: "delivery-signs", group: "fleet", icon: FileSignature, meta: { title: "deliverySigns", icon: "signature", breadcrumb: true, hidden: true } },
  { path: "/fulfillment/map-center", name: "MapCenter", key: "map-center", group: "fleet", icon: MapIcon, meta: { title: "mapCenter", icon: "map", breadcrumb: true } },
  { path: "/fulfillment/weighbridge", name: "Weighbridge", key: "weighbridge", group: "fleet", icon: Scale, meta: { title: "weighbridge", icon: "scale", breadcrumb: true } },
  { path: "/system/org", name: "GroupOrganization", key: "system-org", group: "system-settings", icon: Building2, meta: { title: "groupOrganization", icon: "building", breadcrumb: true } },
  { path: "/system/license", name: "LicenseManagement", key: "system-license", group: "system-settings", icon: FileSignature, meta: { title: "licenseManagement", icon: "signature", breadcrumb: true } },
  { path: "/system/maintenance", name: "SystemMaintenance", key: "system-maintenance", group: "system-settings", icon: Server, meta: { title: "systemMaintenance", icon: "server", breadcrumb: true } },
  { path: "/system/gateway", name: "GatewayRoutes", key: "system-gateway", group: "system-settings", icon: Route, meta: { title: "gatewayRoutes", icon: "route", breadcrumb: true } },
  { path: "/system/security", name: "SecurityPolicies", key: "system-security", group: "system-settings", icon: ShieldCheck, meta: { title: "securityPolicies", icon: "shield", breadcrumb: true } },
  { path: "/system/identity", name: "IdentityIntegration", key: "system-identity", group: "system-settings", icon: Users, meta: { title: "identityIntegration", icon: "users", breadcrumb: true } },
  { path: "/system/plugins", name: "PluginManagement", key: "system-plugins", group: "system-settings", icon: Package, meta: { title: "pluginManagement", icon: "package", breadcrumb: true } },
  { path: "/system/rules", name: "RuleCenter", key: "system-rules", group: "system-settings", icon: ListChecks, meta: { title: "ruleCenter", icon: "list-checks", breadcrumb: true } },
  { path: "/system/integrations", name: "IntegrationEndpoints", key: "system-integrations", group: "system-settings", icon: Server, meta: { title: "integrationEndpoints", icon: "server", breadcrumb: true } },
  { path: "/system/menu", name: "MenuManagement", key: "system-menu", group: "system-settings", icon: Menu, meta: { title: "menuManagement", icon: "menu", breadcrumb: true } },
  { path: "/system/dictionaries", name: "DataDictionary", key: "system-dictionaries", group: "system-settings", icon: ListChecks, meta: { title: "dataDictionary", icon: "list-checks", breadcrumb: true } },
  { path: "/system/users", name: "UserManagement", key: "system-users", group: "system-settings", icon: Users, meta: { title: "userManagement", icon: "users", breadcrumb: true } },
  { path: "/system/roles", name: "RoleManagement", key: "system-roles", group: "system-settings", icon: ShieldCheck, meta: { title: "roleManagement", icon: "shield", breadcrumb: true } },
  { path: "/system/workflows", name: "WorkflowManagement", key: "system-workflows", group: "system-settings", icon: Route, meta: { title: "workflowManagement", icon: "route", breadcrumb: true } },
  { path: "/system/audit", name: "AuditLogs", key: "system-audit", group: "system-settings", icon: ClipboardList, meta: { title: "auditLogs", icon: "clipboard-list", breadcrumb: true } },
  { path: "/system/approvals", name: "ApprovalCenter", key: "approval-center", group: "system-settings", icon: ClipboardCheck, meta: { title: "approvalCenter", icon: "clipboard", breadcrumb: true } },
  { path: "/account/profile", name: "UserProfile", key: "user-profile", group: "system-settings", icon: UserCircle, meta: { title: "userProfile", icon: "user", breadcrumb: true, hidden: true } },
  { path: "/account/security", name: "AccountSecurity", key: "account-security", group: "system-settings", icon: ShieldCheck, meta: { title: "accountSecurity", icon: "shield", breadcrumb: true, hidden: true } }
] as const;

const defaultLaboratoryModule: LaboratoryModuleKey = "exceptions";
const standaloneSections = new Set<ViewKey>(["dispatch", "map-center"]);

const workbenchSections: ViewKey[] = ["overview", "master-customers", "customer-risk", "master-projects", "master-products", "sales-pricing", "portal-customer", "master-materials", "master-sites", "master-plants", "stock-yards", "master-drivers", "master-vehicles", "master-carriers", "portal-driver", "orders", "production", "production-plans", "production-tasks", "production-batches", "production-reports", "dispatch", "dispatch-schedules", "dispatch-queue", "map-center", "weighbridge", "delivery", "delivery-signs", "settlement", "contracts", "raw-material-receipts", "inventory-transfers", "inventory-stocktakes", "raw-material-inspections", "finance", "finance-receivables", "finance-invoices", "finance-collections", "finance-suppliers", "finance-carriers", "reports", "approval-center", "system-org", "system-license", "system-maintenance", "system-gateway", "system-security", "system-identity", "system-plugins", "system-rules", "system-integrations", "system-menu", "system-dictionaries", "system-users", "system-roles", "system-workflows", "system-audit"];
const navGroups: NavGroupKey[] = ["production", "sales", "laboratory", "finance", "fleet", "system-settings"];
const messages = {
  "zh-CN": {
    brandTitle: "建材 ERP 管理平台",
    workspace: "客户工作台",
    nav: {
      overview: "经营总览",
      "master-customers": "客户资料",
      "master-projects": "项目资料",
      "master-products": "产品资料",
      "master-materials": "物料资料",
      "master-sites": "站点信息",
      "master-plants": "生产线管理",
      "stock-yards": "堆场管理",
      "master-drivers": "司机资料",
      "master-vehicles": "车辆资料",
      "master-carriers": "承运商资料",
      orders: "销售订单",
      "customer-risk": "客户风控",
      "sales-pricing": "销售定价",
      "portal-customer": "客户门户",
      production: "生产工作台",
      "production-plans": "生产计划",
      "production-tasks": "生产任务",
      "production-batches": "生产批次",
      "production-reports": "生产日报",
      dispatch: "调度运输",
      "dispatch-schedules": "调度排班",
      "dispatch-queue": "装料队列",
      delivery: "送货单管理",
      "delivery-signs": "签收归档",
      "portal-driver": "司机门户",
      "map-center": "地图中心",
      weighbridge: "过磅记录",
      "site-signing": "工地签收",
      settlement: "客户对账",
      contracts: "合同管理",
      "raw-material-receipts": "原料收料",
      "inventory-transfers": "库存调拨",
      "inventory-stocktakes": "库存盘点",
      "raw-material-inspections": "原材质检",
      finance: "对账结算",
      "finance-receivables": "应收收款",
      "finance-invoices": "税票管理",
      "finance-collections": "催收管理",
      "finance-suppliers": "供应商对账",
      "finance-carriers": "承运结算",
      reports: "报表分析",
      "approval-center": "审批中心",
      "system-org": "集团组织",
      "system-license": "授权管理",
      "system-maintenance": "系统维护",
      "system-gateway": "网关路由",
      "system-security": "安全策略",
      "system-identity": "身份集成",
      "system-plugins": "插件管理",
      "system-rules": "规则中心",
      "system-integrations": "集成端点",
      "system-menu": "菜单管理",
      "system-dictionaries": "数据字典",
      "system-users": "用户管理",
      "system-roles": "角色管理",
      "system-workflows": "工作流管理",
      "system-audit": "审计日志",
      "user-profile": "个人中心",
      "account-security": "账号安全",
      laboratory: "实验室管理",
      "mix-designs": "基础配比",
      "plant-mix-designs": "生产线配比",
      "trial-runs": "试配记录",
      "sample-tests": "样品试验",
      "equipment-calibration": "仪器校准",
      "sample-ledger": "样品台账",
      exceptions: "异常闭环"
    },
    groups: {
      production: "生产",
      sales: "销售",
      laboratory: "实验室",
      finance: "财务",
      fleet: "车队",
      "system-settings": "系统设置"
    },
    auth: {
      title: "建材 ERP 管理平台",
      connectionMode: "连接模式",
      localMode: "本地模式",
      serverMode: "服务端模式",
      localModeHint: "使用本机内置服务和本地数据，不依赖外部服务端。",
      localModeReady: "本地模式无需填写服务端地址，保存后将连接本机内置服务。",
      serverModeHint: "连接局域网、服务器或云端 ERP 服务。",
      server: "服务端地址",
      serverPlaceholder: "例如 192.168.1.10:8088",
      username: "账号",
      password: "密码",
      mfa: "动态码",
      login: "登录平台"
    },
    actions: {
      collapse: "收起侧边栏",
      expand: "展开侧边栏",
      collapseGroup: "收起菜单组",
      expandGroup: "展开菜单组",
      search: "搜索",
      searchPlaceholder: "搜索页面 / 功能",
      noSearchResults: "未找到页面",
      refresh: "刷新",
      theme: "主题色",
      language: "语言",
      liveEvents: "实时事件",
      notifications: "消息通知",
      close: "关闭",
      closeOthers: "关闭其它",
      closeLeft: "关闭左侧",
      closeRight: "关闭右侧",
      closeAll: "关闭全部",
      refreshView: "刷新页面",
      copyPath: "复制路径",
      openPage: "打开页面",
      openStandalone: "独立窗口打开",
      workbench: "ERP 工作台",
      laboratoryHubTitle: "实验室管理",
      siteScope: "当前站点",
      allSites: "全部站点",
      displayMode: "显示模式",
      systemMode: "跟随系统",
      highlightMode: "高亮模式",
      nightMode: "暗夜模式",
      layout: "布局模式",
      sideLayout: "左侧",
      topLayout: "顶部",
      settings: "系统设置",
      closeSettings: "关闭设置",
      serverSettings: "连接设置",
      testServer: "测试连接",
      saveServer: "保存连接",
      resetServer: "恢复本地默认",
      currentServer: "当前 API",
      serverSaved: "连接已保存",
      serverConnected: "连接正常",
      serverReset: "已恢复本地默认"
    },
    errors: {
      loadFailed: "加载失败",
      loginFailed: "登录失败",
      ssoFailed: "SSO 登录失败",
      serverConnectFailed: "服务端连接失败"
    }
  },
  "en-US": {
    brandTitle: "Building Materials ERP",
    workspace: "Customer Workspace",
    nav: {
      overview: "Overview",
      "master-customers": "Customers",
      "master-projects": "Projects",
      "master-products": "Products",
      "master-materials": "Materials",
      "master-sites": "Site Info",
      "master-plants": "Production Lines",
      "stock-yards": "Stock Yards",
      "master-drivers": "Drivers",
      "master-vehicles": "Vehicles",
      "master-carriers": "Carriers",
      orders: "Sales Orders",
      "customer-risk": "Customer Risk",
      "sales-pricing": "Sales Pricing",
      "portal-customer": "Customer Portal",
      production: "Production",
      "production-plans": "Production Plans",
      "production-tasks": "Production Tasks",
      "production-batches": "Production Batches",
      "production-reports": "Production Reports",
      dispatch: "Dispatch",
      "dispatch-schedules": "Dispatch Schedules",
      "dispatch-queue": "Loading Queue",
      delivery: "Delivery Notes",
      "delivery-signs": "Delivery Signs",
      "portal-driver": "Driver Portal",
      "map-center": "Map Center",
      weighbridge: "Weighbridge",
      "site-signing": "Site Signing",
      settlement: "Statements",
      contracts: "Contracts",
      "raw-material-receipts": "Raw Receipts",
      "inventory-transfers": "Inventory Transfers",
      "inventory-stocktakes": "Inventory Stocktakes",
      "raw-material-inspections": "Raw Inspections",
      finance: "Reconciliation",
      "finance-receivables": "Receivables",
      "finance-invoices": "Tax Invoices",
      "finance-collections": "Collections",
      "finance-suppliers": "Supplier Statements",
      "finance-carriers": "Carrier Settlements",
      reports: "Reports",
      "approval-center": "Approvals",
      "system-org": "Group Org",
      "system-license": "License",
      "system-maintenance": "Maintenance",
      "system-gateway": "Gateway",
      "system-security": "Security",
      "system-identity": "Identity",
      "system-plugins": "Plugins",
      "system-rules": "Rules",
      "system-integrations": "Integrations",
      "system-menu": "Menu",
      "system-dictionaries": "Dictionaries",
      "system-users": "Users",
      "system-roles": "Roles",
      "system-workflows": "Workflows",
      "system-audit": "Audit Logs",
      "user-profile": "Profile",
      "account-security": "Account Security",
      laboratory: "Laboratory",
      "mix-designs": "Base Mix Designs",
      "plant-mix-designs": "Production Line Mixes",
      "trial-runs": "Trial Runs",
      "sample-tests": "Sample Tests",
      "equipment-calibration": "Calibration",
      "sample-ledger": "Sample Ledger",
      exceptions: "Quality Exceptions"
    },
    groups: {
      production: "Production",
      sales: "Sales",
      laboratory: "Laboratory",
      finance: "Finance",
      fleet: "Fleet",
      "system-settings": "System Settings"
    },
    auth: {
      title: "Building Materials ERP",
      connectionMode: "Connection mode",
      localMode: "Local mode",
      serverMode: "Server mode",
      localModeHint: "Use the built-in local service and local data without an external server.",
      localModeReady: "Local mode does not require a server URL. Save to use the built-in local service.",
      serverModeHint: "Connect to a LAN, server, or cloud ERP service.",
      server: "Server URL",
      serverPlaceholder: "e.g. 192.168.1.10:8088",
      username: "Username",
      password: "Password",
      mfa: "MFA Code",
      login: "Sign In"
    },
      actions: {
        collapse: "Collapse sidebar",
        expand: "Expand sidebar",
        collapseGroup: "Collapse menu group",
        expandGroup: "Expand menu group",
        search: "Search",
        searchPlaceholder: "Search pages / features",
        noSearchResults: "No pages found",
        refresh: "Refresh",
      theme: "Theme",
      language: "Language",
      liveEvents: "Live events",
      notifications: "Notifications",
      close: "Close",
      closeOthers: "Close others",
      closeLeft: "Close left",
      closeRight: "Close right",
      closeAll: "Close all",
      refreshView: "Refresh page",
      copyPath: "Copy path",
      openPage: "Open page",
      openStandalone: "Open standalone",
      workbench: "ERP Workspace",
      laboratoryHubTitle: "Laboratory Hub",
      siteScope: "Current site",
      allSites: "All sites",
      displayMode: "Display mode",
      systemMode: "System",
      highlightMode: "Highlight",
      nightMode: "Night",
      layout: "Layout",
      sideLayout: "Side",
      topLayout: "Top",
      settings: "Settings",
      closeSettings: "Close settings",
      serverSettings: "Connection settings",
      testServer: "Test connection",
      saveServer: "Save connection",
      resetServer: "Reset local default",
      currentServer: "Current API",
      serverSaved: "Connection saved",
      serverConnected: "Connection OK",
      serverReset: "Local default restored"
    },
    errors: {
      loadFailed: "Load failed",
      loginFailed: "Login failed",
      ssoFailed: "SSO login failed",
      serverConnectFailed: "Server connection failed"
    }
  }
} as const;

const zhTWMessages = {
  ...messages["zh-CN"],
  workspace: "客戶工作台",
  nav: {
    ...messages["zh-CN"].nav,
    overview: "經營總覽",
    "master-customers": "客戶資料",
    "master-projects": "專案資料",
    "master-products": "產品資料",
    "master-materials": "物料資料",
    "master-sites": "站點資訊",
    "master-plants": "生產線管理",
    "stock-yards": "堆場管理",
    "master-drivers": "司機資料",
    "master-vehicles": "車輛資料",
    "master-carriers": "承運商資料",
    orders: "銷售訂單",
    "customer-risk": "客戶風控",
    "sales-pricing": "銷售定價",
    "portal-customer": "客戶門戶",
    production: "生產工作台",
    "production-plans": "生產計劃",
    "production-tasks": "生產任務",
    "production-batches": "生產批次",
    "production-reports": "生產日報",
    dispatch: "調度運輸",
    "dispatch-schedules": "調度排班",
    "dispatch-queue": "裝料隊列",
    delivery: "送貨單管理",
    "delivery-signs": "簽收歸檔",
    "portal-driver": "司機門戶",
    "map-center": "地圖中心",
    weighbridge: "過磅記錄",
    "site-signing": "工地簽收",
    settlement: "客戶對帳",
    contracts: "合約管理",
    "raw-material-receipts": "原料收料",
    "inventory-transfers": "庫存調撥",
    "inventory-stocktakes": "庫存盤點",
    "raw-material-inspections": "原材質檢",
    finance: "對帳結算",
    "finance-receivables": "應收收款",
    "finance-invoices": "稅票管理",
    "finance-collections": "催收管理",
    "finance-suppliers": "供應商對帳",
    "finance-carriers": "承運結算",
    reports: "報表分析",
    "approval-center": "審批中心",
    "system-org": "集團組織",
    "system-license": "授權管理",
    "system-maintenance": "系統維護",
    "system-gateway": "閘道路由",
    "system-security": "安全策略",
    "system-identity": "身份整合",
    "system-plugins": "插件管理",
    "system-rules": "規則中心",
    "system-integrations": "整合端點",
    "system-menu": "選單管理",
    "system-dictionaries": "資料字典",
    "system-users": "使用者管理",
    "system-roles": "角色管理",
    "system-workflows": "工作流管理",
    "system-audit": "審計日誌",
    "user-profile": "個人中心",
    "account-security": "帳號安全",
    laboratory: "實驗室管理",
    "mix-designs": "基礎配比",
    "plant-mix-designs": "生產線配比",
    "trial-runs": "試配記錄",
    "sample-tests": "樣品試驗",
    "equipment-calibration": "儀器校準",
    "sample-ledger": "樣品台帳",
    exceptions: "異常閉環"
  },
  groups: {
    production: "生產",
    sales: "銷售",
    laboratory: "實驗室",
    finance: "財務",
    fleet: "車隊",
    "system-settings": "系統設定"
  },
  auth: {
    ...messages["zh-CN"].auth,
    connectionMode: "連線模式",
    localMode: "本地模式",
    serverMode: "服務端模式",
    localModeHint: "使用本機內建服務和本地資料，不依賴外部服務端。",
    localModeReady: "本地模式無需填寫服務端地址，儲存後將連接本機內建服務。",
    serverModeHint: "連接局域網、伺服器或雲端 ERP 服務。",
    serverPlaceholder: "例如 192.168.1.10:8088",
    username: "帳號",
    password: "密碼",
    login: "登入平台"
  },
  actions: {
    ...messages["zh-CN"].actions,
    collapse: "收起側邊欄",
    expand: "展開側邊欄",
    collapseGroup: "收起選單組",
    expandGroup: "展開選單組",
    searchPlaceholder: "搜尋頁面 / 功能",
    noSearchResults: "未找到頁面",
    notifications: "訊息通知",
    closeOthers: "關閉其它",
    closeLeft: "關閉左側",
    closeRight: "關閉右側",
    closeAll: "關閉全部",
    refreshView: "重新整理頁面",
    copyPath: "複製路徑",
    openPage: "開啟頁面",
    openStandalone: "獨立視窗開啟",
    siteScope: "目前站點",
    allSites: "全部站點",
    displayMode: "顯示模式",
    systemMode: "跟隨系統",
    highlightMode: "高亮模式",
    nightMode: "暗夜模式",
    layout: "佈局模式",
    sideLayout: "左側",
    topLayout: "頂部",
    settings: "系統設定",
    closeSettings: "關閉設定",
    serverSettings: "連線設定",
    testServer: "測試連線",
    saveServer: "儲存連線",
    resetServer: "恢復本地預設",
    currentServer: "目前 API",
    serverSaved: "連線已儲存",
    serverConnected: "連線正常",
    serverReset: "已恢復本地預設"
  },
  errors: {
    ...messages["zh-CN"].errors,
    loadFailed: "載入失敗",
    loginFailed: "登入失敗",
    serverConnectFailed: "服務端連線失敗"
  }
};

const localeMessages = {
  ...messages,
  "zh-TW": zhTWMessages
} as const;

const languageOptions: Array<{ value: Locale; label: string }> = [
  { value: "zh-CN", label: "简体中文" },
  { value: "zh-TW", label: "繁體中文" },
  { value: "en-US", label: "English" }
];

const CONNECTION_MODE_STORAGE_KEY = "erp-connection-mode";
const THEME_MODE_STORAGE_KEY = "erp-theme-mode";
const THEME_MODE_STORAGE_VERSION_KEY = "erp-theme-mode-storage-version";
const THEME_MODE_STORAGE_VERSION = "system-v1";
const SYSTEM_THEME_MEDIA_QUERY = "(prefers-color-scheme: dark)";

const themeOptions = [
  { name: "Neon Cyan", color: "#20f6ff" },
  { name: "Ion Lime", color: "#7dffb2" },
  { name: "Pulse Amber", color: "#ffd166" },
  { name: "Signal Magenta", color: "#ff4fd8" }
];

type AppRoute = (typeof nav)[number];

const routePermissionFallbacks: Partial<Record<ViewKey, string>> = {
  overview: "dashboard:read",
  production: "production:read",
  "production-plans": "production:read",
  "production-tasks": "production:read",
  "production-batches": "production:read",
  "production-reports": "production:read",
  "master-products": "master:read",
  "master-materials": "master:read",
  "master-sites": "master:read",
  "master-plants": "master:read",
  "stock-yards": "procurement:read",
  orders: "order:read",
  "master-customers": "master:read",
  "customer-risk": "master:read",
  "master-projects": "master:read",
  "sales-pricing": "master:read",
  "portal-customer": "customer:read",
  reports: "report:read",
  exceptions: "quality:read",
  "site-signing": "delivery:read&dispatch:read",
  "mix-designs": "quality:read",
  "plant-mix-designs": "quality:read",
  "trial-runs": "quality:read",
  "sample-tests": "quality:read",
  "equipment-calibration": "quality:read",
  "sample-ledger": "quality:read",
  "raw-material-receipts": "procurement:read",
  "inventory-transfers": "procurement:read",
  "inventory-stocktakes": "procurement:read",
  "raw-material-inspections": "quality:read",
  settlement: "statement:read",
  contracts: "contract:read",
  finance: "finance:read",
  "finance-receivables": "finance:read",
  "finance-invoices": "finance:read",
  "finance-collections": "finance:read",
  "finance-suppliers": "finance:read",
  "finance-carriers": "finance:read",
  "master-drivers": "master:read",
  "master-vehicles": "master:read",
  "master-carriers": "master:read",
  "portal-driver": "driver:read",
  dispatch: "dispatch:read",
  "dispatch-schedules": "dispatch:read",
  "dispatch-queue": "dispatch:read",
  delivery: "delivery:read",
  "delivery-signs": "delivery:read",
  "map-center": "vehicle:read",
  weighbridge: "ticket:read",
  "system-org": "org:read",
  "system-license": "system:read",
  "system-maintenance": "system:read",
  "system-gateway": "system:read",
  "system-security": "system:read",
  "system-identity": "system:read",
  "system-plugins": "system:read",
  "system-rules": "rule:read",
  "system-integrations": "integration:read",
  "system-menu": "system:read",
  "system-dictionaries": "system:read",
  "system-users": "system:read",
  "system-roles": "system:read",
  "system-workflows": "system:read",
  "system-audit": "system:read",
  "approval-center": "approval:read"
};

function routePermission(route: AppRoute, permissions: Record<string, string> | undefined) {
  return permissions?.[route.key] || routePermissionFallbacks[route.key] || "";
}

function routeAllowed(route: AppRoute, grants: string[], menuPermissions: Record<string, string> | undefined) {
  const permission = routePermission(route, menuPermissions);
  return permission ? hasPermission(grants, permission) : false;
}

function normalizedPath(pathname: string) {
  return pathname.replace(/\/+$/, "") || "/";
}

function routeForPath(pathname: string) {
  const path = normalizedPath(pathname);
  if (path === "/resources/stock-yards" || path === "/resources/inventory" || path === "/resources/inventory-legacy" || path === "/finance/procurement") {
    return nav.find((item) => item.key === "stock-yards");
  }
  if (path === "/resources/warehouses") {
    return nav.find((item) => item.key === "stock-yards");
  }
  if (path === "/resources/plant-buffers" || path === "/resources/silos") {
    return nav.find((item) => item.key === "master-plants");
  }
  if (path === "/quality/laboratory") {
    return nav.find((item) => item.key === defaultLaboratoryModule);
  }
  return nav.find((item) => item.path === path);
}

function viewKeyFromPath(pathname: string): ViewKey {
  return routeForPath(pathname)?.key || "overview";
}

function pathForView(key: ViewKey) {
  return nav.find((item) => item.key === key)?.path || "/dashboard";
}

function isStandaloneSection(key: ViewKey) {
  return standaloneSections.has(key);
}

function hasDesktopBridge() {
  const desktopWindow = window as DesktopBridgeWindow;
  return Boolean(desktopWindow.runtime && desktopWindow.go?.main?.DesktopApp?.OpenStandaloneWindow);
}

async function openStandaloneSection(key: ViewKey) {
  const path = pathForView(key);
  if (hasDesktopBridge()) {
    await OpenStandaloneWindow(path);
    return true;
  }

  const popup = window.open(path, "_blank", "popup,width=1440,height=900");
  if (!popup) return false;
  popup.opener = null;
  popup.focus();
  return true;
}

function savedLocale(): Locale {
  const saved = window.localStorage.getItem("erp-locale");
  return languageOptions.some((item) => item.value === saved) ? saved as Locale : "zh-CN";
}

function savedConnectionMode(): ConnectionMode {
  const saved = window.localStorage.getItem(CONNECTION_MODE_STORAGE_KEY);
  if (saved === "local" || saved === "server") return saved;
  return api.baseURL() === api.defaultBaseURL() ? "local" : "server";
}

function savedThemeColor() {
  const saved = window.localStorage.getItem("erp-theme-color");
  return themeOptions.some((item) => item.color === saved) ? saved as string : themeOptions[0].color;
}

function savedLayoutMode(): LayoutMode {
  return window.localStorage.getItem("erp-layout-mode") === "top" ? "top" : "side";
}

function savedThemeMode(): ThemeMode {
  const saved = window.localStorage.getItem(THEME_MODE_STORAGE_KEY);
  const storageVersion = window.localStorage.getItem(THEME_MODE_STORAGE_VERSION_KEY);
  const mode: ThemeMode = storageVersion === THEME_MODE_STORAGE_VERSION && isThemeMode(saved) ? saved : "system";
  document.documentElement.dataset.themeMode = resolveThemeMode(mode);
  document.documentElement.dataset.themePreference = mode;
  return mode;
}

function isThemeMode(mode: string | null): mode is ThemeMode {
  return mode === "system" || mode === "highlight" || mode === "night";
}

function resolveSystemThemeMode(): ResolvedThemeMode {
  return window.matchMedia(SYSTEM_THEME_MEDIA_QUERY).matches ? "night" : "highlight";
}

function resolveThemeMode(mode: ThemeMode): ResolvedThemeMode {
  return mode === "system" ? resolveSystemThemeMode() : mode;
}

function savedSelectedSiteId() {
  const value = Number(window.localStorage.getItem("erp-selected-site-id") || "0");
  return Number.isFinite(value) && value > 0 ? value : 0;
}

function isAffixRoute(route: AppRoute) {
  return "affix" in route.meta && route.meta.affix;
}

function isNoCacheRoute(route: AppRoute) {
  return "noCache" in route.meta && route.meta.noCache;
}

function isHiddenRoute(route: AppRoute) {
  return "hidden" in route.meta && Boolean(route.meta.hidden);
}

function isDemoOIDCProvider(provider: OIDCProvider) {
  return [
    provider.name,
    provider.code,
    provider.issuer,
    provider.clientId,
    provider.authUrl
  ].some((value) => /demo|sample|mock|演示/i.test(value || ""));
}

class ViewErrorBoundary extends Component<
  { children: ReactNode; viewKey: string },
  { error: Error | null }
> {
  state: { error: Error | null } = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("View render failed", error, info.componentStack);
  }

  componentDidUpdate(prevProps: { viewKey: string }) {
    if (prevProps.viewKey !== this.props.viewKey && this.state.error) {
      this.setState({ error: null });
    }
  }

  render() {
    if (this.state.error) {
      return (
        <>
          <MessageBox
            open
            title="页面加载失败"
            tone="error"
            message={this.state.error.message || "视图渲染异常"}
            confirmLabel="重新加载页面"
            onClose={() => this.setState({ error: null })}
          />
          <Panel className="view-error-panel">
            <h3>页面加载失败</h3>
            <Button onClick={() => this.setState({ error: null })}>重新加载页面</Button>
          </Panel>
        </>
      );
    }
    return this.props.children;
  }
}

export function App() {
  const publicSignToken = window.location.pathname.match(/^\/public\/sign\/([^/]+)/)?.[1];
  const [tokenReady, setTokenReady] = useState(Boolean(api.token));
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaRequired, setMfaRequired] = useState(false);
  const [error, setError] = useState("");
  const [active, setActiveView] = useState<ViewKey>(() => viewKeyFromPath(window.location.pathname));
  const [desktopRouteReady, setDesktopRouteReady] = useState(() => !hasDesktopBridge());
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [visitedViews, setVisitedViews] = useState<ViewKey[]>(["overview"]);
  const [locale, setLocale] = useState<Locale>(savedLocale);
  const [themeColor, setThemeColor] = useState(savedThemeColor);
  const [layoutMode, setLayoutMode] = useState<LayoutMode>(savedLayoutMode);
  const [themeMode, setThemeMode] = useState<ThemeMode>(savedThemeMode);
  const [systemThemeMode, setSystemThemeMode] = useState<ResolvedThemeMode>(resolveSystemThemeMode);
  const [collapsedNavGroups, setCollapsedNavGroups] = useState<Partial<Record<NavGroupKey, boolean>>>({});
  const [tagMenu, setTagMenu] = useState<{ key: ViewKey; x: number; y: number } | null>(null);
  const [navMenu, setNavMenu] = useState<{ key: ViewKey; group?: NavGroupKey; x: number; y: number } | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [loginServerConfigOpen, setLoginServerConfigOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [pageSearch, setPageSearch] = useState("");
  const [bootstrap, setBootstrap] = useState<BootstrapData | null>(null);
  const [selectedSiteId, setSelectedSiteId] = useState(savedSelectedSiteId);
  const [eventCount, setEventCount] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);
  const [viewReloadKey, setViewReloadKey] = useState(0);
  const [ssoProviders, setSSOProviders] = useState<OIDCProvider[]>([]);
  const [connectionMode, setConnectionMode] = useState<ConnectionMode>(savedConnectionMode);
  const [serverBaseUrl, setServerBaseUrl] = useState(api.baseURL());
  const [serverInput, setServerInput] = useState(api.baseURL());
  const [serverStatus, setServerStatus] = useState("");
  const [serverStatusTone, setServerStatusTone] = useState<"neutral" | "success" | "error">("neutral");
  const [serverTesting, setServerTesting] = useState(false);
  const [accountMfaEnrollment, setAccountMfaEnrollment] = useState<MFAEnrollment | null>(null);
  const [accountMfaCode, setAccountMfaCode] = useState("");
  const [accountProfileForm, setAccountProfileForm] = useState({ displayName: "" });
  const [accountPasswordForm, setAccountPasswordForm] = useState({ currentPassword: "", newPassword: "", confirmPassword: "" });
  const [accountBusy, setAccountBusy] = useState("");
  const [accountActionError, setAccountActionError] = useState("");

  const copy = localeMessages[locale];
  const menuLabelOverrides = bootstrap?.menuLabels;
  function configuredMenuLabel(key: string) {
    return menuLabelOverrides?.[key]?.trim() || "";
  }
  function navLabel(key: ViewKey) {
    const override = configuredMenuLabel(key);
    if (override) return override;
    const labels = copy.nav as Partial<Record<ViewKey, string>>;
    return labels[key] || nav.find((item) => item.key === key)?.name || key;
  }
  function navGroupLabel(key: NavGroupKey) {
    return configuredMenuLabel(`group:${key}`) || copy.groups[key];
  }
  const currentPermissions = useMemo(() => permissionsForRole(bootstrap?.roles, bootstrap?.user.roleCode), [bootstrap]);
  const menuPermissions = bootstrap?.menuPermissions;
  const visibleNav = useMemo(() => (bootstrap ? nav.filter((item) => routeAllowed(item, currentPermissions, menuPermissions)) : nav), [bootstrap, currentPermissions, menuPermissions]);
  const visibleMainNav = useMemo(() => visibleNav.filter((item) => !isHiddenRoute(item)), [visibleNav]);
  const visibleNavKeys = useMemo(() => new Set<ViewKey>(visibleNav.map((item) => item.key)), [visibleNav]);
  const firstVisibleKey = (visibleMainNav[0]?.key || visibleNav[0]?.key || "overview") as ViewKey;
  const currentNav = useMemo(() => nav.find((item) => item.key === active) || nav[0], [active]);
  const cachedViews = useMemo(
    () => visitedViews.filter((key) => visibleNavKeys.has(key) && !isNoCacheRoute(nav.find((item) => item.key === key) || nav[0])),
    [visitedViews, visibleNavKeys]
  );
  const siteOptions = bootstrap?.sites || [];
  const selectedSite = useMemo(() => siteOptions.find((item) => item.id === selectedSiteId) || null, [siteOptions, selectedSiteId]);
  const menuItems = useMemo<WorkbenchMenuItem[]>(
    () => visibleMainNav.map((item, index) => ({
      key: item.key,
      path: item.path,
      name: item.name,
      label: navLabel(item.key),
      group: item.group,
      groupLabel: navGroupLabel(item.group),
      icon: item.meta.icon,
      sort: index + 1,
      permissionMark: routePermission(item, menuPermissions) || "-",
      affix: isAffixRoute(item),
      noCache: isNoCacheRoute(item),
      breadcrumb: Boolean(item.meta.breadcrumb)
    })),
    [copy, menuLabelOverrides, menuPermissions, visibleMainNav]
  );
  const pageSearchKeyword = pageSearch.trim().toLowerCase();
  const pageSearchResults = useMemo(() => {
    if (!pageSearchKeyword) return [];

    return menuItems.filter((item) => [
      item.label,
      item.name,
      item.groupLabel,
      item.path,
      item.permissionMark
    ].join(" ").toLowerCase().includes(pageSearchKeyword)).slice(0, 8);
  }, [menuItems, pageSearchKeyword]);
  const currentLabel = navLabel(active);
  function localeText(zhCN: string, zhTW: string, enUS: string) {
    if (locale === "en-US") return enUS;
    return locale === "zh-TW" ? zhTW : zhCN;
  }
  const isSideLayout = layoutMode === "side";
  const standaloneMode = isStandaloneSection(active);
  const sidebarToggleLabel = sidebarCollapsed ? copy.actions.expand : copy.actions.collapse;
  const resolvedThemeMode: ResolvedThemeMode = themeMode === "system" ? systemThemeMode : themeMode;
  const shellStyle = { "--erp-theme-color": themeColor } as CSSProperties;
  const shellClass = [
    "app-shell",
    "admin-layout",
    `theme-mode-${resolvedThemeMode}`,
    isSideLayout ? "side-nav-layout" : "top-nav-layout",
    standaloneMode ? "standalone-layout" : "",
    isSideLayout && sidebarCollapsed ? "sidebar-collapsed" : ""
  ].filter(Boolean).join(" ");
  const tagMenuRoute = tagMenu ? nav.find((item) => item.key === tagMenu.key) || nav[0] : null;
  const navMenuRoute = navMenu ? nav.find((item) => item.key === navMenu.key) || nav[0] : null;
  const activeViewRenderKey = `${active}-${viewReloadKey}`;
  const currentRole = useMemo(() => bootstrap?.roles.find((item) => item.code === bootstrap.user.roleCode) || null, [bootstrap]);
  const currentCompany = useMemo(() => bootstrap?.companies.find((item) => item.id === bootstrap.user.companyId) || null, [bootstrap]);
  const currentSite = useMemo(() => bootstrap?.sites.find((item) => item.id === bootstrap.user.siteId) || null, [bootstrap]);
  const message = useMessage();
  const { showError } = useMessageBox();

  async function load() {
    setBootstrap(await api.bootstrap());
  }

  useEffect(() => {
    if (!bootstrap?.user) return;
    setAccountProfileForm({ displayName: bootstrap.user.displayName || "" });
  }, [bootstrap?.user.id, bootstrap?.user.displayName]);

  async function runAccountSecurityAction(label: string, action: () => Promise<void>) {
    setAccountBusy(label);
    setAccountActionError("");
    try {
      await action();
    } catch (err) {
      const nextError = err instanceof Error ? err.message : localeText("账号安全操作失败", "帳號安全操作失敗", "Account security action failed");
      setAccountActionError(nextError);
      message.error(nextError);
    } finally {
      setAccountBusy("");
    }
  }

  async function handleUpdateAccountProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const displayName = accountProfileForm.displayName.trim();
    if (!displayName) {
      const nextError = localeText("显示名称不能为空", "顯示名稱不能為空", "Display name is required");
      setAccountActionError(nextError);
      message.error(nextError);
      return;
    }
    await runAccountSecurityAction("profile", async () => {
      await api.updateAccountProfile({ displayName });
      await load();
      message.success(localeText("资料已保存", "資料已儲存", "Profile saved"));
    });
  }

  async function handleChangeAccountPassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const currentPassword = accountPasswordForm.currentPassword;
    const newPassword = accountPasswordForm.newPassword.trim();
    const confirmPassword = accountPasswordForm.confirmPassword.trim();
    if (newPassword.length < 6) {
      const nextError = localeText("新密码至少 6 位", "新密碼至少 6 位", "New password must be at least 6 characters");
      setAccountActionError(nextError);
      message.error(nextError);
      return;
    }
    if (newPassword !== confirmPassword) {
      const nextError = localeText("两次输入的新密码不一致", "兩次輸入的新密碼不一致", "New passwords do not match");
      setAccountActionError(nextError);
      message.error(nextError);
      return;
    }
    await runAccountSecurityAction("password", async () => {
      await api.changeAccountPassword({ currentPassword, newPassword });
      setAccountPasswordForm({ currentPassword: "", newPassword: "", confirmPassword: "" });
      await load();
      message.success(localeText("密码已更新", "密碼已更新", "Password updated"));
    });
  }

  async function handleEnrollAccountMFA() {
    if (!bootstrap?.user.id) return;
    await runAccountSecurityAction("mfa-enroll", async () => {
      const enrollment = await api.enrollMFA(bootstrap.user.id);
      setAccountMfaEnrollment(enrollment);
      setAccountMfaCode("");
      message.success(localeText("MFA 密钥已生成", "MFA 密鑰已產生", "MFA secret generated"));
    });
  }

  async function handleEnableAccountMFA() {
    if (!bootstrap?.user.id) return;
    const code = accountMfaCode.trim();
    if (!code) {
      setAccountActionError(localeText("请输入 MFA 动态码", "請輸入 MFA 動態碼", "Enter the MFA code"));
      return;
    }
    await runAccountSecurityAction("mfa-enable", async () => {
      await api.enableMFA(bootstrap.user.id, code);
      setAccountMfaEnrollment(null);
      setAccountMfaCode("");
      await load();
      message.success(localeText("MFA 已启用", "MFA 已啟用", "MFA enabled"));
    });
  }

  async function handleDisableAccountMFA() {
    if (!bootstrap?.user.id) return;
    await runAccountSecurityAction("mfa-disable", async () => {
      await api.disableMFA(bootstrap.user.id);
      setAccountMfaEnrollment(null);
      setAccountMfaCode("");
      await load();
      message.success(localeText("MFA 已关闭", "MFA 已關閉", "MFA disabled"));
    });
  }

  function activateInCurrentWindow(nextKey: ViewKey, historyMode: HistoryMode = "push") {
    setActiveView(nextKey);
    if (publicSignToken || historyMode === "none") return;
    const path = pathForView(nextKey);
    if (normalizedPath(window.location.pathname) === path) return;
    const method = historyMode === "replace" ? "replaceState" : "pushState";
    window.history[method]({ viewKey: nextKey }, "", path);
  }

  function setActive(key: ViewKey, historyMode: HistoryMode = "push") {
    const nextKey = bootstrap && !visibleNavKeys.has(key) ? firstVisibleKey : key;
    if (!standaloneMode && historyMode === "push" && isStandaloneSection(nextKey)) {
      void openStandaloneSection(nextKey)
        .then((opened) => {
          if (!opened) activateInCurrentWindow(nextKey, historyMode);
        })
        .catch(() => activateInCurrentWindow(nextKey, historyMode));
      return;
    }
    activateInCurrentWindow(nextKey, historyMode);
  }

  function openSearchResult(key: string) {
    setActive(key as ViewKey);
    setPageSearch("");
  }

  useEffect(() => {
    if (publicSignToken) return;
    function syncActiveFromLocation() {
      setActiveView(viewKeyFromPath(window.location.pathname));
    }
    window.addEventListener("popstate", syncActiveFromLocation);
    return () => window.removeEventListener("popstate", syncActiveFromLocation);
  }, [publicSignToken]);

  useEffect(() => {
    if (!hasDesktopBridge()) {
      setDesktopRouteReady(true);
      return;
    }

    let cancelled = false;
    InitialRoute()
      .then((route) => {
        if (cancelled || !route) return;
        const nextKey = viewKeyFromPath(route);
        if (!isStandaloneSection(nextKey)) return;
        setActiveView(nextKey);
        const path = pathForView(nextKey);
        if (normalizedPath(window.location.pathname) !== path) {
          window.history.replaceState({ viewKey: nextKey }, "", path);
        }
      })
      .catch(() => undefined)
      .finally(() => {
        if (!cancelled) setDesktopRouteReady(true);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!tokenReady || !bootstrap || publicSignToken) return;
    setVisitedViews((items) => {
      const filtered = items.filter((key) => visibleNavKeys.has(key));
      const next = filtered.length ? filtered : [firstVisibleKey];
      return next.length === items.length && next.every((key, index) => key === items[index]) ? items : next;
    });
    if (!visibleNavKeys.has(active)) {
      setActive(firstVisibleKey, "replace");
    }
  }, [active, bootstrap, firstVisibleKey, publicSignToken, tokenReady, visibleNavKeys]);

  useEffect(() => {
    if (publicSignToken || active !== defaultLaboratoryModule) return;
    if (normalizedPath(window.location.pathname) === "/quality/laboratory") {
      window.history.replaceState({ viewKey: defaultLaboratoryModule }, "", pathForView(active));
    }
  }, [active, publicSignToken]);

  useEffect(() => {
    window.localStorage.setItem("erp-locale", locale);
  }, [locale]);

  useEffect(() => {
    window.localStorage.setItem("erp-theme-color", themeColor);
  }, [themeColor]);

  useEffect(() => {
    window.localStorage.setItem("erp-layout-mode", layoutMode);
  }, [layoutMode]);

  useEffect(() => {
    window.localStorage.setItem(THEME_MODE_STORAGE_KEY, themeMode);
    window.localStorage.setItem(THEME_MODE_STORAGE_VERSION_KEY, THEME_MODE_STORAGE_VERSION);
    document.documentElement.dataset.themeMode = resolvedThemeMode;
    document.documentElement.dataset.themePreference = themeMode;
  }, [resolvedThemeMode, themeMode]);

  useEffect(() => {
    const mediaQuery = window.matchMedia(SYSTEM_THEME_MEDIA_QUERY);
    const syncSystemTheme = () => {
      setSystemThemeMode(mediaQuery.matches ? "night" : "highlight");
    };

    syncSystemTheme();
    mediaQuery.addEventListener("change", syncSystemTheme);
    return () => mediaQuery.removeEventListener("change", syncSystemTheme);
  }, []);

  useEffect(() => {
    window.localStorage.setItem("erp-selected-site-id", String(selectedSiteId));
  }, [selectedSiteId]);

  useEffect(() => {
    window.localStorage.setItem(CONNECTION_MODE_STORAGE_KEY, connectionMode);
  }, [connectionMode]);

  useEffect(() => {
    if (!bootstrap || selectedSiteId === 0) return;
    if (!bootstrap.sites.some((item) => item.id === selectedSiteId)) {
      setSelectedSiteId(0);
    }
  }, [bootstrap, selectedSiteId]);

  useEffect(() => {
    setVisitedViews((items) => (items.includes(active) ? items : [...items, active]));
  }, [active]);

  useEffect(() => {
    if (!tagMenu) return;
    function closeContextMenu() {
      setTagMenu(null);
    }
    function closeContextMenuByKey(event: KeyboardEvent) {
      if (event.key === "Escape") closeContextMenu();
    }
    window.addEventListener("click", closeContextMenu);
    window.addEventListener("keydown", closeContextMenuByKey);
    return () => {
      window.removeEventListener("click", closeContextMenu);
      window.removeEventListener("keydown", closeContextMenuByKey);
    };
  }, [tagMenu]);

  useEffect(() => {
    if (!userMenuOpen) return;
    function closeUserMenu() {
      setUserMenuOpen(false);
    }
    function closeUserMenuByKey(event: KeyboardEvent) {
      if (event.key === "Escape") closeUserMenu();
    }
    window.addEventListener("click", closeUserMenu);
    window.addEventListener("keydown", closeUserMenuByKey);
    return () => {
      window.removeEventListener("click", closeUserMenu);
      window.removeEventListener("keydown", closeUserMenuByKey);
    };
  }, [userMenuOpen]);

  useEffect(() => {
    if (!settingsOpen) return;
    function closeSettingsByKey(event: KeyboardEvent) {
      if (event.key === "Escape") setSettingsOpen(false);
    }
    window.addEventListener("keydown", closeSettingsByKey);
    return () => window.removeEventListener("keydown", closeSettingsByKey);
  }, [settingsOpen]);

  useEffect(() => {
    if (!settingsOpen) return;
    setServerInput(connectionMode === "local" ? api.defaultBaseURL() : api.baseURL());
    setServerStatus("");
    setServerStatusTone("neutral");
  }, [settingsOpen, connectionMode]);

  useEffect(() => {
    if (!loginServerConfigOpen) return;
    setServerInput(connectionMode === "local" ? api.defaultBaseURL() : api.baseURL());
    setServerStatus("");
    setServerStatusTone("neutral");
  }, [loginServerConfigOpen, connectionMode]);

  useEffect(() => {
    if (!tokenReady) return;
    load().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : copy.errors.loadFailed);
      setTokenReady(false);
    });
  }, [tokenReady, refreshKey]);

  useEffect(() => {
    if (error) {
      showError(error, copy.errors.loadFailed);
    }
  }, [copy.errors.loadFailed, error, showError]);

  useEffect(() => {
    if (tokenReady) return;
    api.ssoProviders()
      .then((providers) => setSSOProviders(providers.filter((provider) => !isDemoOIDCProvider(provider))))
      .catch(() => setSSOProviders([]));
  }, [serverBaseUrl, tokenReady]);

  useEffect(() => {
    if (!tokenReady || !api.token) return;
    const source = new EventSource(eventURL(), { withCredentials: true });
    source.onmessage = () => setEventCount((value) => value + 1);
    const topics = [
      "sales.order.created",
      "sales.order.update",
      "dispatch.order.update",
      "ticket.created",
      "delivery.signed",
      "vehicle.location.update",
      "statement.confirmed",
      "finance.invoice.created",
      "finance.invoice.tax_accepted",
      "finance.invoice.tax_submitted",
      "finance.invoice.tax_callback",
      "finance.invoice.red_submitted",
      "finance.receipt.created",
      "procurement.receipt.created",
      "vehicle.alarm.created",
      "laboratory.mix_design.created",
      "laboratory.mix_design.revised",
      "laboratory.mix_design.approved",
      "laboratory.mix_design.retired",
      "laboratory.mix_design_plant_profile.created",
      "laboratory.mix_design_plant_profile.approved",
      "laboratory.mix_design_plant_profile.retired",
      "laboratory.trial.created",
      "laboratory.sample.created",
      "laboratory.test.created",
      "laboratory.test.reviewed",
      "laboratory.equipment.saved",
      "laboratory.calibration.created",
      "laboratory.exception.created",
      "laboratory.exception.handled"
    ];
    topics.forEach((topic) => {
      source.addEventListener(topic, () => {
        setEventCount((value) => value + 1);
        setRefreshKey((value) => value + 1);
      });
    });
    return () => source.close();
  }, [serverBaseUrl, tokenReady]);

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      const nextServer = api.setBaseURL(connectionMode === "local" ? api.defaultBaseURL() : serverInput);
      setServerBaseUrl(nextServer);
      setServerInput(nextServer);
      const result = await api.login(username, password, mfaCode);
      if (result.mfaRequired) {
        setMfaRequired(true);
        setMfaCode("");
        return;
      }
      setTokenReady(true);
      setMfaRequired(false);
      setMfaCode("");
    } catch (err) {
      setError(err instanceof Error ? err.message : copy.errors.loginFailed);
    }
  }

  async function handleSSO(provider: OIDCProvider) {
    setError("");
    try {
      const start = await api.startSSO(provider.code);
      window.location.assign(start.authorizationUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : copy.errors.ssoFailed);
    }
  }

  function handleLogout() {
    api.token = "";
    window.localStorage.removeItem("cbmp.token");
    setTokenReady(false);
    setBootstrap(null);
    setVisitedViews(["overview"]);
    setActiveView("overview");
    setUserMenuOpen(false);
    setError("");
    if (!publicSignToken && normalizedPath(window.location.pathname) !== "/dashboard") {
      window.history.replaceState({ viewKey: "overview" }, "", "/dashboard");
    }
  }

  function resetSessionForServerChange() {
    api.token = "";
    window.localStorage.removeItem("cbmp.token");
    setTokenReady(false);
    setBootstrap(null);
    setVisitedViews(["overview"]);
    setActiveView("overview");
    setMfaRequired(false);
    setMfaCode("");
    setEventCount(0);
    if (!publicSignToken && normalizedPath(window.location.pathname) !== "/dashboard") {
      window.history.replaceState({ viewKey: "overview" }, "", "/dashboard");
    }
  }

  function setConnectionModeDraft(mode: ConnectionMode) {
    setConnectionMode(mode);
    setServerStatus("");
    setServerStatusTone("neutral");
    if (mode === "local") {
      setServerInput(api.defaultBaseURL());
    } else {
      setServerInput(api.baseURL());
    }
  }

  function connectionDraftURL(value = serverInput) {
    return connectionMode === "local" ? api.defaultBaseURL() : value;
  }

  function saveServerConfig(value = serverInput) {
    setError("");
    try {
      const previous = api.baseURL();
      const next = api.setBaseURL(connectionDraftURL(value));
      setServerBaseUrl(next);
      setServerInput(next);
      setServerStatus(copy.actions.serverSaved);
      setServerStatusTone("success");
      if (next !== previous) {
        resetSessionForServerChange();
      }
    } catch (err) {
      setServerStatus(err instanceof Error ? `${copy.errors.serverConnectFailed}: ${err.message}` : copy.errors.serverConnectFailed);
      setServerStatusTone("error");
    }
  }

  function resetServerConfig() {
    const previous = api.baseURL();
    const next = api.resetBaseURL();
    setConnectionMode("local");
    setServerBaseUrl(next);
    setServerInput(next);
    setServerStatus(copy.actions.serverReset);
    setServerStatusTone("success");
    if (next !== previous) {
      resetSessionForServerChange();
    }
  }

  async function testServerConfig() {
    setServerTesting(true);
    setServerStatus("");
    setServerStatusTone("neutral");
    try {
      await api.testConnection(connectionDraftURL());
      setServerStatus(copy.actions.serverConnected);
      setServerStatusTone("success");
    } catch (err) {
      setServerStatus(err instanceof Error ? `${copy.errors.serverConnectFailed}: ${err.message}` : copy.errors.serverConnectFailed);
      setServerStatusTone("error");
    } finally {
      setServerTesting(false);
    }
  }

  function renderServerConfigPanel() {
    const displayedServerInput = connectionDraftURL();
    const serverMode = connectionMode === "server";
    return (
      <div className="server-config-panel">
        <div className="server-config-mode">
          <span>{copy.auth.connectionMode}</span>
          <div className="settings-segmented server-mode-switcher" aria-label={copy.auth.connectionMode} role="group">
            <BareButton active={connectionMode === "local"} onClick={() => setConnectionModeDraft("local")}>
              <Home size={14} />
              {copy.auth.localMode}
            </BareButton>
            <BareButton active={connectionMode === "server"} onClick={() => setConnectionModeDraft("server")}>
              <Server size={14} />
              {copy.auth.serverMode}
            </BareButton>
          </div>
          <p>{connectionMode === "local" ? copy.auth.localModeHint : copy.auth.serverModeHint}</p>
        </div>
        {serverMode ? (
          <>
            <Field label={copy.auth.server}>
              <TextInput
                value={displayedServerInput}
                onChange={(event) => setServerInput(event.target.value)}
                placeholder={copy.auth.serverPlaceholder}
                autoComplete="url"
              />
            </Field>
            <div className="server-config-current">
              <span>{copy.actions.currentServer}</span>
              <b>{serverBaseUrl}</b>
            </div>
          </>
        ) : (
          <div className="server-config-local-summary">
            <Home size={15} />
            <span>{copy.auth.localModeReady}</span>
          </div>
        )}
        <ActionGroup className="server-config-actions">
          {serverMode ? (
            <Button type="button" onClick={testServerConfig} disabled={serverTesting}>
              {copy.actions.testServer}
            </Button>
          ) : null}
          <Button type="button" variant="primary" onClick={() => saveServerConfig()}>
            {copy.actions.saveServer}
          </Button>
          {serverMode ? (
            <Button type="button" onClick={resetServerConfig}>
              {copy.actions.resetServer}
            </Button>
          ) : null}
        </ActionGroup>
        {serverStatus ? <p className={`server-config-status ${serverStatusTone}`}>{serverStatus}</p> : null}
      </div>
    );
  }

  function closeTag(key: ViewKey) {
    const route = nav.find((item) => item.key === key) || nav[0];
    if (isAffixRoute(route)) return;
    const next = visitedViews.filter((item) => item !== key);
    const safeNext = next.length ? next : ["overview" as ViewKey];
    setVisitedViews(safeNext);
    if (active === key) {
      setActive(safeNext[safeNext.length - 1]);
    }
  }

  function closeOtherTags(target: ViewKey = active) {
    const fixed = visitedViews.filter((key) => isAffixRoute(nav.find((item) => item.key === key) || nav[0]));
    const next = Array.from(new Set([...fixed, target]));
    setVisitedViews(next);
    setActive(target);
  }

  function closeAllTags() {
    const fixed = visitedViews.filter((key) => isAffixRoute(nav.find((item) => item.key === key) || nav[0]));
    const next = fixed.length ? fixed : ["overview" as ViewKey];
    setVisitedViews(next);
    setActive(next[next.length - 1]);
  }

  function closeSideTags(target: ViewKey, side: "left" | "right") {
    const targetIndex = visitedViews.indexOf(target);
    if (targetIndex < 0) return;
    const next = visitedViews.filter((key, index) => {
      const route = nav.find((item) => item.key === key) || nav[0];
      if (isAffixRoute(route) || key === target) return true;
      return side === "left" ? index > targetIndex : index < targetIndex;
    });
    const safeNext = next.length ? next : ["overview" as ViewKey];
    setVisitedViews(safeNext);
    if (!safeNext.includes(active)) {
      setActive(target);
    }
  }

  function refreshView(key: ViewKey = active) {
    setActive(key);
    setViewReloadKey((value) => value + 1);
  }

  function copyAppText(text: string, label: string) {
    void copyTextToClipboard(text)
      .then(() => message.success(localeText(`${label}已复制`, `${label}已複製`, `${label} copied`)))
      .catch(() => message.error(localeText("复制失败", "複製失敗", "Copy failed")));
  }

  function switchSite(value: string) {
    const next = Number(value);
    setSelectedSiteId(Number.isFinite(next) && next > 0 ? next : 0);
  }

  function renderSiteSwitcher(className = "") {
    const classes = ["site-switcher", className].filter(Boolean).join(" ");
    return (
      <IconField className={classes} icon={<Factory size={16} />} label={copy.actions.siteScope} title={`${copy.actions.siteScope}: ${selectedSite?.name || copy.actions.allSites}`}>
        <SelectInput value={selectedSiteId} onChange={(event) => switchSite(event.target.value)} aria-label={copy.actions.siteScope} disabled={!siteOptions.length}>
          <option value={0}>{copy.actions.allSites}</option>
          {siteOptions.map((site) => <option key={site.id} value={site.id}>{site.name}</option>)}
        </SelectInput>
      </IconField>
    );
  }

  function toggleNavGroup(group: NavGroupKey) {
    setCollapsedNavGroups((items) => ({
      ...items,
      [group]: !items[group]
    }));
  }

  function openTagContextMenu(event: ReactMouseEvent<HTMLDivElement>, key: ViewKey) {
    event.preventDefault();
    event.stopPropagation();
    setActive(key);
    setTagMenu({
      key,
      x: event.clientX,
      y: event.clientY
    });
  }

  function openNavContextMenu(event: ReactMouseEvent, key: ViewKey, group?: NavGroupKey) {
    event.preventDefault();
    event.stopPropagation();
    setNavMenu({
      key,
      group,
      x: event.clientX,
      y: event.clientY
    });
  }

  function tagContextMenuItems(key: ViewKey, route: AppRoute): ContextMenuItem[] {
    const targetIndex = visitedViews.indexOf(key);
    const closableLeft = visitedViews.some((item, index) => index < targetIndex && !isAffixRoute(nav.find((routeItem) => routeItem.key === item) || nav[0]));
    const closableRight = visitedViews.some((item, index) => index > targetIndex && !isAffixRoute(nav.find((routeItem) => routeItem.key === item) || nav[0]));
    const hasClosableTag = visitedViews.some((item) => !isAffixRoute(nav.find((routeItem) => routeItem.key === item) || nav[0]));
    return [
      {
        key: "refresh",
        label: copy.actions.refreshView,
        icon: <RefreshCw size={14} />,
        onSelect: () => refreshView(key)
      },
      {
        key: "copy-path",
        label: copy.actions.copyPath,
        icon: <Copy size={14} />,
        onSelect: () => copyAppText(route.path, localeText("路径", "路徑", "Path"))
      },
      {
        key: "open-standalone",
        label: copy.actions.openStandalone,
        icon: <ExternalLink size={14} />,
        disabled: !isStandaloneSection(key),
        onSelect: () => {
          void openStandaloneSection(key);
        }
      },
      { key: "tag-separator", type: "separator" },
      {
        key: "close",
        label: copy.actions.close,
        icon: <X size={14} />,
        disabled: isAffixRoute(route),
        onSelect: () => closeTag(key)
      },
      {
        key: "close-others",
        label: copy.actions.closeOthers,
        icon: <Layers size={14} />,
        disabled: visitedViews.length <= 1,
        onSelect: () => closeOtherTags(key)
      },
      {
        key: "close-left",
        label: copy.actions.closeLeft,
        icon: <ChevronRight size={14} />,
        disabled: !closableLeft,
        onSelect: () => closeSideTags(key, "left")
      },
      {
        key: "close-right",
        label: copy.actions.closeRight,
        icon: <ChevronRight size={14} />,
        disabled: !closableRight,
        onSelect: () => closeSideTags(key, "right")
      },
      {
        key: "close-all",
        label: copy.actions.closeAll,
        icon: <X size={14} />,
        disabled: !hasClosableTag,
        danger: true,
        onSelect: closeAllTags
      }
    ];
  }

  function navContextMenuItems(key: ViewKey, route: AppRoute, group?: NavGroupKey): ContextMenuItem[] {
    const groupCollapsed = group ? Boolean(collapsedNavGroups[group]) : false;
    return [
      {
        key: "open",
        label: copy.actions.openPage,
        icon: <Menu size={14} />,
        onSelect: () => setActive(key)
      },
      {
        key: "refresh",
        label: copy.actions.refreshView,
        icon: <RefreshCw size={14} />,
        onSelect: () => refreshView(key)
      },
      {
        key: "copy-path",
        label: copy.actions.copyPath,
        icon: <Copy size={14} />,
        onSelect: () => copyAppText(route.path, localeText("路径", "路徑", "Path"))
      },
      {
        key: "open-standalone",
        label: copy.actions.openStandalone,
        icon: <ExternalLink size={14} />,
        disabled: !isStandaloneSection(key),
        onSelect: () => {
          void openStandaloneSection(key);
        }
      },
      group ? { key: "nav-group-separator", type: "separator" as const } : { key: "nav-empty-separator", type: "separator" as const },
      {
        key: "toggle-group",
        label: groupCollapsed ? copy.actions.expandGroup : copy.actions.collapseGroup,
        icon: <ChevronRight size={14} />,
        disabled: !group,
        onSelect: () => {
          if (group) toggleNavGroup(group);
        }
      },
      {
        key: "close-others",
        label: copy.actions.closeOthers,
        icon: <Layers size={14} />,
        disabled: !visitedViews.includes(key),
        onSelect: () => closeOtherTags(key)
      }
    ];
  }

  function renderPermissionChips() {
    return (
      <ChipList compact>
        {currentPermissions.map((item) => <span key={item}>{item}</span>)}
        {!currentPermissions.length ? <span>无权限</span> : null}
      </ChipList>
    );
  }

  function renderUserProfilePage() {
    if (!bootstrap) {
      return <Panel className="account-page-view">暂无用户信息</Panel>;
    }
    return (
      <div className="view-stack account-page-view">
        <Panel className="account-hero-panel">
          <div className="profile-summary account-profile-summary">
            <div className="profile-avatar"><UserCircle size={38} /></div>
            <div>
              <b>{bootstrap.user.displayName}</b>
              <span>{bootstrap.user.username} / {currentRole?.name || bootstrap.user.roleCode}</span>
            </div>
          </div>
          <span className="account-status-pill">{bootstrap.user.status}</span>
        </Panel>
        <SectionGrid>
          <Panel className="span-6 account-detail-panel">
            <SectionHeader className="panel-head-compact">
              <div>
                <b>{navLabel("user-profile")}</b>
                <span>{bootstrap.user.username}</span>
              </div>
              <UserCircle size={16} />
            </SectionHeader>
            <form className="account-profile-form" onSubmit={handleUpdateAccountProfile}>
              <Field label="显示名称">
                <TextInput value={accountProfileForm.displayName} onChange={(event) => setAccountProfileForm({ displayName: event.target.value })} />
              </Field>
              <ActionGroup>
                <Button variant="primary" type="submit" disabled={accountBusy !== "" || !accountProfileForm.displayName.trim()}>保存资料</Button>
              </ActionGroup>
            </form>
            {accountActionError ? <p className="form-error">{accountActionError}</p> : null}
            <div className="profile-info-grid">
              <div><span>角色</span><b>{currentRole?.name || bootstrap.user.roleCode}</b></div>
              <div><span>数据范围</span><b>{currentRole?.dataScope || "-"}</b></div>
              <div><span>公司</span><b>{currentCompany?.name || "-"}</b></div>
              <div><span>站点</span><b>{currentSite?.name || "-"}</b></div>
              <div><span>MFA</span><b>{bootstrap.user.mfaEnabled ? "已启用" : "未启用"}</b></div>
              <div><span>状态</span><b>{bootstrap.user.status}</b></div>
            </div>
          </Panel>
          <Panel className="span-6 account-detail-panel">
            <SectionHeader className="panel-head-compact">
              <div>
                <b>权限标识</b>
                <span>{currentPermissions.length} 项</span>
              </div>
              <ShieldCheck size={16} />
            </SectionHeader>
            {renderPermissionChips()}
          </Panel>
        </SectionGrid>
      </div>
    );
  }

  function renderAccountSecurityPage() {
    if (!bootstrap) {
      return <Panel className="account-page-view">暂无账号信息</Panel>;
    }
    return (
      <div className="view-stack account-page-view">
        <Panel className="account-security-panel">
          <SectionHeader className="panel-head-compact account-page-head">
            <div>
              <b>{navLabel("account-security")}</b>
              <span>{bootstrap.user.username}</span>
            </div>
            <ShieldCheck size={18} />
          </SectionHeader>
          <div className="profile-info-grid">
            <div><span>登录账号</span><b>{bootstrap.user.username}</b></div>
            <div><span>显示名称</span><b>{bootstrap.user.displayName}</b></div>
            <div><span>角色</span><b>{currentRole?.name || bootstrap.user.roleCode}</b></div>
            <div><span>MFA</span><b>{bootstrap.user.mfaEnabled ? "已启用" : "未启用"}</b></div>
            <div><span>公司</span><b>{currentCompany?.name || "-"}</b></div>
            <div><span>站点</span><b>{currentSite?.name || "-"}</b></div>
          </div>
        </Panel>
        <Panel className="account-detail-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>登录密码</b>
              <span>当前账号</span>
            </div>
            <ShieldCheck size={16} />
          </SectionHeader>
          <form className="account-password-form" onSubmit={handleChangeAccountPassword}>
            <Field label="当前密码">
              <TextInput type="password" value={accountPasswordForm.currentPassword} onChange={(event) => setAccountPasswordForm((value) => ({ ...value, currentPassword: event.target.value }))} />
            </Field>
            <Field label="新密码">
              <TextInput type="password" value={accountPasswordForm.newPassword} onChange={(event) => setAccountPasswordForm((value) => ({ ...value, newPassword: event.target.value }))} />
            </Field>
            <Field label="确认新密码">
              <TextInput type="password" value={accountPasswordForm.confirmPassword} onChange={(event) => setAccountPasswordForm((value) => ({ ...value, confirmPassword: event.target.value }))} />
            </Field>
            <ActionGroup>
              <Button
                variant="primary"
                type="submit"
                disabled={accountBusy !== "" || !accountPasswordForm.currentPassword || !accountPasswordForm.newPassword || !accountPasswordForm.confirmPassword}
              >
                修改密码
              </Button>
            </ActionGroup>
          </form>
          {accountActionError ? <p className="form-error">{accountActionError}</p> : null}
        </Panel>
        <Panel className="account-detail-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>MFA 认证</b>
              <span>{bootstrap.user.mfaEnabled ? "已启用" : accountMfaEnrollment ? "待验证" : "未启用"}</span>
            </div>
            <ShieldCheck size={16} />
          </SectionHeader>
          <div className="profile-info-grid">
            <div><span>账号</span><b>{bootstrap.user.username}</b></div>
            <div><span>状态</span><b>{bootstrap.user.mfaEnabled ? "已启用" : "未启用"}</b></div>
          </div>
          {accountMfaEnrollment ? (
            <div className="account-mfa-enrollment">
              <Field label="Secret">
                <TextInput value={accountMfaEnrollment.secret} readOnly />
              </Field>
              <Field label="Authenticator URL">
                <TextInput value={accountMfaEnrollment.otpauthUrl} readOnly />
              </Field>
              <Field label="动态码">
                <TextInput value={accountMfaCode} onChange={(event) => setAccountMfaCode(event.target.value)} inputMode="numeric" />
              </Field>
            </div>
          ) : null}
          {accountActionError ? <p className="form-error">{accountActionError}</p> : null}
          <ActionGroup>
            {!bootstrap.user.mfaEnabled ? (
              <>
                <Button type="button" onClick={handleEnrollAccountMFA} disabled={accountBusy !== ""}>{accountMfaEnrollment ? "重新生成密钥" : "生成 MFA 密钥"}</Button>
                <Button variant="primary" type="button" onClick={handleEnableAccountMFA} disabled={accountBusy !== "" || !accountMfaEnrollment || !accountMfaCode.trim()}>启用 MFA</Button>
              </>
            ) : (
              <Button variant="danger" type="button" onClick={handleDisableAccountMFA} disabled={accountBusy !== ""}>关闭 MFA</Button>
            )}
          </ActionGroup>
        </Panel>
        <Panel className="account-detail-panel">
          <SectionHeader className="panel-head-compact">
            <div>
              <b>当前权限</b>
              <span>{currentPermissions.length} 项 permission marks</span>
            </div>
            <ShieldCheck size={16} />
          </SectionHeader>
          {renderPermissionChips()}
        </Panel>
      </div>
    );
  }

  function renderActiveView() {
    return (
      <ViewErrorBoundary viewKey={active}>
        {workbenchSections.includes(active) ? <ERPWorkbenchView section={active as ERPWorkbenchSection} bootstrap={bootstrap} menuItems={menuItems} selectedSiteId={selectedSiteId} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "mix-designs" ? <LaboratoryView activeModule="mix-designs" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "plant-mix-designs" ? <LaboratoryView activeModule="plant-mix-designs" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "trial-runs" ? <LaboratoryView activeModule="trial-runs" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "sample-tests" ? <LaboratoryView activeModule="sample-tests" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "equipment-calibration" ? <LaboratoryView activeModule="equipment-calibration" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "sample-ledger" ? <LaboratoryView activeModule="sample-ledger" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "exceptions" ? <LaboratoryView activeModule="exceptions" currentRoleCode={bootstrap?.user.roleCode || ""} currentPermissions={currentPermissions} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "site-signing" ? <SiteSigningModule selectedSiteId={selectedSiteId} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
        {active === "user-profile" ? renderUserProfilePage() : null}
        {active === "account-security" ? renderAccountSecurityPage() : null}
      </ViewErrorBoundary>
    );
  }

  if (publicSignToken) {
    return <PublicSignView token={decodeURIComponent(publicSignToken)} />;
  }

  if (tokenReady && !desktopRouteReady) {
    return (
      <LayoutRegion as="main" className={`login-shell theme-mode-${resolvedThemeMode}`} data-theme-mode={resolvedThemeMode}>
        <Panel className="login-card">
          <p className="eyebrow">ERP Appliance</p>
          <h1>Loading...</h1>
        </Panel>
      </LayoutRegion>
    );
  }

  if (!tokenReady) {
    return (
      <LayoutRegion as="main" className={`login-shell theme-mode-${resolvedThemeMode}`} data-theme-mode={resolvedThemeMode}>
        <Panel className="login-card">
          <div className="login-corner-actions">
            <Button
              type="button"
              icon={<Server size={14} />}
              onClick={() => setLoginServerConfigOpen((value) => !value)}
              aria-expanded={loginServerConfigOpen}
            >
              {copy.actions.serverSettings}
            </Button>
          </div>
          <Dialog
            open={loginServerConfigOpen}
            title={copy.actions.serverSettings}
            size="sm"
            className="login-server-dialog"
            closeLabel={copy.actions.close}
            onClose={() => setLoginServerConfigOpen(false)}
          >
            {renderServerConfigPanel()}
          </Dialog>
          <p className="eyebrow">ERP Appliance</p>
          <h1>{copy.auth.title}</h1>
          <LoginForm onSubmit={handleLogin}>
            <Field label={copy.auth.username}>
              <TextInput value={username} onChange={(event) => setUsername(event.target.value)} />
            </Field>
            <Field label={copy.auth.password}>
              <TextInput type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
            </Field>
            {mfaRequired ? (
              <Field label={copy.auth.mfa}>
                <TextInput value={mfaCode} onChange={(event) => setMfaCode(event.target.value)} />
              </Field>
            ) : null}
            <Button variant="primary" type="submit">{copy.auth.login}</Button>
            {ssoProviders.length ? (
              <ActionGroup>
                {ssoProviders.map((provider) => (
                  <Button key={provider.code} onClick={() => handleSSO(provider)}>
                    {provider.name}
                  </Button>
                ))}
              </ActionGroup>
            ) : null}
          </LoginForm>
        </Panel>
      </LayoutRegion>
    );
  }

  if (standaloneMode) {
    return (
      <div className={shellClass} data-theme-mode={resolvedThemeMode} style={shellStyle}>
        <LayoutRegion as="section" className="workbench standalone-workbench">
          <div className="content standalone-content">
            <LayoutRegion as="main" className="app-main standalone-main fade-transform-enter" key={activeViewRenderKey} data-route-name={currentNav.name} data-no-cache={isNoCacheRoute(currentNav) ? "true" : "false"}>
              {renderActiveView()}
            </LayoutRegion>
          </div>
        </LayoutRegion>
      </div>
    );
  }

  return (
    <div className={shellClass} data-theme-mode={resolvedThemeMode} style={shellStyle}>
      {isSideLayout ? (
        <LayoutRegion as="aside" className="side" id="main-sidebar">
          <div className="brand">
            <img className="brand-logo" src={appIconUrl} alt="" aria-hidden="true" />
            <div className="brand-body">
              <div className="brand-text">
                <b>{copy.brandTitle}</b>
              </div>
              {renderSiteSwitcher("brand-site-switcher")}
            </div>
          </div>
          <div className="side-nav">
            {navGroups.map((group) => {
              const groupItems = visibleMainNav.filter((item) => item.group === group);
              if (!groupItems.length) return null;
              const groupCollapsed = Boolean(collapsedNavGroups[group]);
              const hideGroupItems = groupCollapsed && !sidebarCollapsed;
              const groupLabel = navGroupLabel(group);
              const groupToggleLabel = `${groupCollapsed ? copy.actions.expandGroup : copy.actions.collapseGroup}: ${groupLabel}`;
              return (
                <div className="sidebar-group" key={group}>
                  <BareButton
                    className="sidebar-group-title"
                    onClick={() => toggleNavGroup(group)}
                    title={groupToggleLabel}
                    aria-label={groupToggleLabel}
                    aria-controls={`sidebar-group-items-${group}`}
                    aria-expanded={!groupCollapsed}
                  >
                    <span>{groupLabel}</span>
                    <ChevronRight className="sidebar-group-chevron" size={14} />
                  </BareButton>
                  <div className="sidebar-group-items" id={`sidebar-group-items-${group}`} hidden={hideGroupItems}>
                    {groupItems.map((item) => {
                      const Icon = item.icon;
                      return (
                        <BareButton
                          key={item.key}
                          className="nav-item"
                          active={active === item.key}
                          onClick={() => setActive(item.key)}
                          onContextMenu={(event) => openNavContextMenu(event, item.key, item.group)}
                          title={navLabel(item.key)}
                        >
                          <Icon size={18} />
                          <span>{navLabel(item.key)}</span>
                        </BareButton>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </LayoutRegion>
      ) : null}
      <LayoutRegion as="section" className="workbench">
        <LayoutRegion as="header" className="topbar admin-navbar">
          <div className="navbar-left">
            {isSideLayout ? (
              <IconButton
                className="navbar-icon-button"
                icon={<Menu size={20} />}
                label={sidebarToggleLabel}
                onClick={() => setSidebarCollapsed((value) => !value)}
                title={sidebarToggleLabel}
                aria-controls="main-sidebar"
                aria-expanded={!sidebarCollapsed}
              />
            ) : (
              <BareButton className="top-brand" onClick={() => setActive("overview")} title={copy.brandTitle}>
                <img className="top-brand-logo" src={appIconUrl} alt="" aria-hidden="true" />
                <span>{copy.brandTitle}</span>
              </BareButton>
            )}
            <nav className="breadcrumbs navbar-breadcrumbs" aria-label="面包屑导航">
              <BareButton className="breadcrumb-link" onClick={() => setActive("overview")}>
                <Home size={15} />
                {copy.actions.workbench}
              </BareButton>
              {active !== "overview" ? (
                <>
                  <ChevronRight className="breadcrumb-separator-icon" size={15} />
                  <span className="breadcrumb-section">{navGroupLabel(currentNav.group)}</span>
                  <ChevronRight className="breadcrumb-separator-icon" size={15} />
                  <span className="breadcrumb-current">{currentLabel}</span>
                </>
              ) : null}
            </nav>
          </div>
          {layoutMode === "top" ? (
            <nav className="top-nav-menu" aria-label={copy.actions.topLayout}>
              {navGroups.map((group) => {
                const groupItems = visibleMainNav.filter((item) => item.group === group);
                if (!groupItems.length) return null;
                const groupLabel = navGroupLabel(group);
                const GroupIcon = groupItems[0].icon;
                const groupActive = groupItems.some((item) => item.key === active);
                const contextMenuItem = groupItems.find((item) => item.key === active) || groupItems[0];
                return (
                  <div className="top-nav-group" key={group}>
                    <BareButton
                      className="top-nav-group-trigger"
                      active={groupActive}
                      aria-haspopup="menu"
                      onContextMenu={(event) => openNavContextMenu(event, contextMenuItem.key, group)}
                      title={groupLabel}
                    >
                      <GroupIcon size={15} />
                      <span>{groupLabel}</span>
                      <ChevronDown size={13} />
                    </BareButton>
                    <div className="top-nav-dropdown" role="menu" aria-label={groupLabel}>
                      {groupItems.map((item) => {
                        const Icon = item.icon;
                        return (
                          <BareButton
                            className="top-nav-dropdown-item"
                            active={active === item.key}
                            key={item.key}
                            role="menuitem"
                            onClick={(event) => {
                              setActive(item.key);
                              event.currentTarget.blur();
                            }}
                            onContextMenu={(event) => openNavContextMenu(event, item.key, item.group)}
                            title={navLabel(item.key)}
                          >
                            <Icon size={15} />
                            <span>{navLabel(item.key)}</span>
                          </BareButton>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </nav>
          ) : null}
          <div className="top-actions navbar-actions">
            <div className="global-search" onClick={(event) => event.stopPropagation()}>
              <IconField className="global-search-box" icon={<Search size={15} />} label={copy.actions.search}>
                <TextInput
                  value={pageSearch}
                  onChange={(event) => setPageSearch(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" && pageSearchResults[0]) {
                      event.preventDefault();
                      openSearchResult(pageSearchResults[0].key);
                    }
                  }}
                  placeholder={copy.actions.searchPlaceholder}
                  aria-label={copy.actions.search}
                />
                {pageSearch ? (
                  <IconButton className="global-search-clear" icon={<X size={14} />} label={copy.actions.close} onClick={() => setPageSearch("")} />
                ) : null}
              </IconField>
              {pageSearchKeyword ? (
                <div className="global-search-panel" role="listbox" aria-label={copy.actions.search}>
                  {pageSearchResults.length ? pageSearchResults.map((item) => {
                    const route = nav.find((candidate) => candidate.key === item.key);
                    const Icon = route?.icon || Menu;
                    return (
                      <BareButton key={item.key} role="option" onClick={() => openSearchResult(item.key)}>
                        <Icon size={15} />
                        <span>
                          <b>{item.label}</b>
                          <small>{item.groupLabel} / {item.path}</small>
                        </span>
                      </BareButton>
                    );
                  }) : (
                    <span className="global-search-empty">{copy.actions.noSearchResults}</span>
                  )}
                </div>
              ) : null}
            </div>
            {isSideLayout ? null : renderSiteSwitcher()}
            <div className="user-menu" onClick={(event) => event.stopPropagation()}>
              <BareButton
                className="user-pill"
                aria-haspopup="menu"
                aria-expanded={userMenuOpen}
                onClick={() => setUserMenuOpen((value) => !value)}
              >
                <UserCircle size={15} />
                <span>{bootstrap?.user.displayName}</span>
                <ChevronDown size={13} />
              </BareButton>
              {userMenuOpen ? (
                <div className="user-dropdown" role="menu">
                  <div className="user-dropdown-head">
                    <b>{bootstrap?.user.displayName}</b>
                    <span>{currentRole?.name || bootstrap?.user.roleCode}</span>
                  </div>
                  <BareButton role="menuitem" className="notification-menu-item" onClick={() => setEventCount(0)}>
                    <Bell size={14} />
                    <span>{copy.actions.notifications}</span>
                    <b className="notification-count">{eventCount}</b>
                  </BareButton>
                  <BareButton role="menuitem" onClick={() => { setActive("user-profile"); setUserMenuOpen(false); }}>
                    <UserCircle size={14} />个人中心
                  </BareButton>
                  <BareButton role="menuitem" onClick={() => { setActive("account-security"); setUserMenuOpen(false); }}>
                    <ShieldCheck size={14} />账号安全
                  </BareButton>
                  <BareButton role="menuitem" onClick={() => { setSettingsOpen(true); setUserMenuOpen(false); }}>
                    <Settings size={14} />系统设置
                  </BareButton>
                  <BareButton className="danger" role="menuitem" onClick={handleLogout}>
                  <LogOut size={14} />退出登录
                  </BareButton>
                </div>
              ) : null}
            </div>
          </div>
        </LayoutRegion>
        <div className="tags-view" data-cached-views={cachedViews.join(",")}>
          <div className="tags-scroll">
            {visitedViews.map((key) => {
              const route = nav.find((item) => item.key === key) || nav[0];
              const closable = !isAffixRoute(route);
              return (
                <div
                  className={`${active === key ? "tag-view active" : "tag-view"}${closable ? " closable" : ""}`}
                  key={key}
                  onContextMenu={(event) => openTagContextMenu(event, key)}
                  title={`${route.path} · ${route.name}`}
                >
                  <BareButton className="tag-view-label" onClick={() => setActive(key)}>
                    {navLabel(key)}
                  </BareButton>
                  {closable ? (
                    <BareButton
                      className="tag-view-close"
                      aria-label={`${copy.actions.close} ${navLabel(key)}`}
                      title={copy.actions.close}
                      onClick={(event) => {
                        event.stopPropagation();
                        closeTag(key);
                      }}
                    >
                      <X size={13} />
                    </BareButton>
                  ) : null}
                </div>
              );
            })}
          </div>
        </div>
        {tagMenu && tagMenuRoute ? (
          <ContextMenu
            items={tagContextMenuItems(tagMenu.key, tagMenuRoute)}
            label="页签快捷操作"
            position={{ x: tagMenu.x, y: tagMenu.y }}
            width={188}
            onClose={() => setTagMenu(null)}
          />
        ) : null}
        {navMenu && navMenuRoute ? (
          <ContextMenu
            items={navContextMenuItems(navMenu.key, navMenuRoute, navMenu.group)}
            label="导航快捷操作"
            position={{ x: navMenu.x, y: navMenu.y }}
            width={196}
            onClose={() => setNavMenu(null)}
          />
        ) : null}
        {settingsOpen ? (
          <>
            <BareButton className="settings-mask" onClick={() => setSettingsOpen(false)} aria-label={copy.actions.closeSettings} />
            <LayoutRegion as="aside" className="settings-drawer" aria-label={copy.actions.settings}>
              <div className="settings-header">
                <div>
                  <p className="eyebrow">{copy.brandTitle}</p>
                  <h3>{copy.actions.settings}</h3>
                </div>
                <IconButton className="navbar-icon-button" icon={<X size={18} />} label={copy.actions.closeSettings} onClick={() => setSettingsOpen(false)} />
              </div>
              <LayoutRegion as="section" className="settings-section">
                <div className="settings-section-title">
                  <Server size={16} />
                  <span>{copy.actions.serverSettings}</span>
                </div>
                {renderServerConfigPanel()}
              </LayoutRegion>
              <LayoutRegion as="section" className="settings-section">
                <div className="settings-section-title">
                  <Palette size={16} />
                  <span>{copy.actions.theme}</span>
                </div>
                <div className="settings-theme-list" aria-label={copy.actions.theme}>
                  {themeOptions.map((item) => (
                    <BareButton
                      className="theme-dot"
                      active={themeColor === item.color}
                      key={item.color}
                      style={{ backgroundColor: item.color }}
                      title={item.name}
                      onClick={() => setThemeColor(item.color)}
                    />
                  ))}
                </div>
              </LayoutRegion>
              <LayoutRegion as="section" className="settings-section">
                <div className="settings-section-title">
                  <Languages size={16} />
                  <span>{copy.actions.language}</span>
                </div>
                <SelectInput
                  className="settings-language-select"
                  value={locale}
                  onChange={(event) => setLocale(event.target.value as Locale)}
                  aria-label={copy.actions.language}
                >
                  {languageOptions.map((item) => (
                    <option key={item.value} value={item.value}>{item.label}</option>
                  ))}
                </SelectInput>
              </LayoutRegion>
              <LayoutRegion as="section" className="settings-section">
                <div className="settings-section-title">
                  <Monitor size={16} />
                  <span>{copy.actions.displayMode}</span>
                </div>
                <div className="layout-switcher settings-segmented settings-mode-switcher" role="group" aria-label={copy.actions.displayMode}>
                  <BareButton active={themeMode === "system"} onClick={() => setThemeMode("system")}>
                    <Monitor size={14} />
                    {copy.actions.systemMode}
                  </BareButton>
                  <BareButton active={themeMode === "highlight"} onClick={() => setThemeMode("highlight")}>
                    <Sun size={14} />
                    {copy.actions.highlightMode}
                  </BareButton>
                  <BareButton active={themeMode === "night"} onClick={() => setThemeMode("night")}>
                    <Moon size={14} />
                    {copy.actions.nightMode}
                  </BareButton>
                </div>
              </LayoutRegion>
              <LayoutRegion as="section" className="settings-section">
                <div className="settings-section-title">
                  <Menu size={16} />
                  <span>{copy.actions.layout}</span>
                </div>
                <div className="layout-switcher settings-segmented settings-layout-switcher" role="group" aria-label={copy.actions.layout}>
                  <BareButton active={layoutMode === "side"} onClick={() => setLayoutMode("side")}>
                    {copy.actions.sideLayout}
                  </BareButton>
                  <BareButton active={layoutMode === "top"} onClick={() => setLayoutMode("top")}>
                    {copy.actions.topLayout}
                  </BareButton>
                </div>
              </LayoutRegion>
            </LayoutRegion>
          </>
        ) : null}
        <div className="content">
          <LayoutRegion as="main" className="app-main fade-transform-enter" key={activeViewRenderKey} data-route-name={currentNav.name} data-no-cache={isNoCacheRoute(currentNav) ? "true" : "false"}>
            {renderActiveView()}
          </LayoutRegion>
        </div>
      </LayoutRegion>
    </div>
  );
}
