import {
  BarChart3,
  Building2,
  ChevronRight,
  ClipboardCheck,
  HardHat,
  Factory,
  FileSignature,
  FlaskConical,
  Home,
  Languages,
  Landmark,
  Layers,
  Menu,
  Package,
  Palette,
  ReceiptText,
  RefreshCw,
  Scale,
  Search,
  Settings,
  ShoppingCart,
  Truck,
  X
} from "lucide-react";
import { Component, CSSProperties, ErrorInfo, FormEvent, MouseEvent as ReactMouseEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { api, eventURL } from "./services/api";
import type { BootstrapData, OIDCProvider } from "./services/types";
import { ERPWorkbenchView } from "./views/ERPWorkbenchView";
import type { ERPWorkbenchSection } from "./views/ERPWorkbenchView";
import { LaboratoryView } from "./views/LaboratoryView";
import { PublicSignView } from "./views/PublicSignView";

type ViewKey = ERPWorkbenchSection | "laboratory";
type Locale = "zh-CN" | "en-US";
type LayoutMode = "side" | "top";
type NavGroupKey = "core" | "fulfillment" | "settlement" | "platform";
type HistoryMode = "push" | "replace" | "none";

const nav = [
  { path: "/dashboard", name: "Dashboard", key: "overview", group: "core", icon: Building2, meta: { title: "overview", icon: "dashboard", affix: true, noCache: true, breadcrumb: true } },
  { path: "/sales/orders", name: "SalesOrders", key: "orders", group: "core", icon: ShoppingCart, meta: { title: "orders", icon: "shopping-cart", breadcrumb: true } },
  { path: "/sales/customers", name: "Customers", key: "master-customers", group: "core", icon: Building2, meta: { title: "masterCustomers", icon: "building", breadcrumb: true } },
  { path: "/sales/projects", name: "Projects", key: "master-projects", group: "core", icon: HardHat, meta: { title: "masterProjects", icon: "hard-hat", breadcrumb: true } },
  { path: "/fulfillment/production", name: "Production", key: "production", group: "fulfillment", icon: Factory, meta: { title: "production", icon: "factory", breadcrumb: true } },
  { path: "/fulfillment/products", name: "Products", key: "master-products", group: "fulfillment", icon: Package, meta: { title: "masterProducts", icon: "package", breadcrumb: true } },
  { path: "/fulfillment/sites", name: "Sites", key: "master-sites", group: "fulfillment", icon: Factory, meta: { title: "masterSites", icon: "factory", breadcrumb: true } },
  { path: "/fulfillment/dispatch", name: "Dispatch", key: "dispatch", group: "fulfillment", icon: Truck, meta: { title: "dispatch", icon: "truck", breadcrumb: true } },
  { path: "/fulfillment/drivers", name: "Drivers", key: "master-drivers", group: "fulfillment", icon: ClipboardCheck, meta: { title: "masterDrivers", icon: "clipboard", breadcrumb: true } },
  { path: "/fulfillment/vehicles", name: "Vehicles", key: "master-vehicles", group: "fulfillment", icon: Truck, meta: { title: "masterVehicles", icon: "truck", breadcrumb: true } },
  { path: "/fulfillment/weighbridge", name: "Weighbridge", key: "weighbridge", group: "fulfillment", icon: Scale, meta: { title: "weighbridge", icon: "scale", breadcrumb: true } },
  { path: "/fulfillment/delivery", name: "DeliverySigning", key: "delivery", group: "fulfillment", icon: FileSignature, meta: { title: "delivery", icon: "signature", breadcrumb: true } },
  { path: "/settlement/statements", name: "Statements", key: "settlement", group: "settlement", icon: ReceiptText, meta: { title: "settlement", icon: "receipt", breadcrumb: true } },
  { path: "/settlement/procurement", name: "Procurement", key: "procurement", group: "settlement", icon: Package, meta: { title: "procurement", icon: "package", breadcrumb: true } },
  { path: "/settlement/procurement/materials", name: "Materials", key: "master-materials", group: "settlement", icon: Layers, meta: { title: "masterMaterials", icon: "stack", breadcrumb: true } },
  { path: "/settlement/procurement/inventory", name: "Inventory", key: "master-inventory", group: "settlement", icon: Scale, meta: { title: "masterInventory", icon: "scale", breadcrumb: true } },
  { path: "/settlement/finance", name: "Finance", key: "finance", group: "settlement", icon: Landmark, meta: { title: "finance", icon: "landmark", breadcrumb: true } },
  { path: "/platform/reports", name: "Reports", key: "reports", group: "platform", icon: BarChart3, meta: { title: "reports", icon: "chart", breadcrumb: true } },
  { path: "/platform/approvals", name: "Approvals", key: "system", group: "platform", icon: ClipboardCheck, meta: { title: "system", icon: "clipboard", breadcrumb: true } },
  { path: "/platform/laboratory", name: "Laboratory", key: "laboratory", group: "platform", icon: FlaskConical, meta: { title: "laboratory", icon: "flask", noCache: true, breadcrumb: true } }
] as const;

const workbenchSections: ViewKey[] = ["overview", "master-customers", "master-projects", "master-products", "master-materials", "master-sites", "master-drivers", "master-vehicles", "master-inventory", "orders", "production", "dispatch", "weighbridge", "delivery", "settlement", "procurement", "finance", "reports", "system"];
const navGroups: NavGroupKey[] = ["core", "fulfillment", "settlement", "platform"];

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
      "master-sites": "站点资料",
      "master-drivers": "司机资料",
      "master-vehicles": "车辆资料",
      "master-inventory": "库存资料",
      orders: "销售订单",
      production: "生产计划",
      dispatch: "调度运输",
      weighbridge: "地磅票据",
      delivery: "工地签收",
      settlement: "客户对账",
      procurement: "采购库存",
      finance: "财务结算",
      reports: "报表分析",
      system: "审批系统",
      laboratory: "实验室管理"
    },
    groups: {
      core: "核心业务",
      fulfillment: "履约中心",
      settlement: "结算财务",
      platform: "平台能力"
    },
    auth: {
      title: "建材 ERP 管理平台",
      username: "账号",
      password: "密码",
      mfa: "动态码",
      login: "登录平台"
    },
    actions: {
      collapse: "收起侧边栏",
      search: "搜索",
      refresh: "刷新",
      theme: "主题色",
      language: "语言",
      liveEvents: "实时事件",
      close: "关闭",
      closeOthers: "关闭其它",
      closeAll: "关闭全部",
      cachedViews: "缓存视图",
      workbench: "ERP 工作台",
      layout: "布局模式",
      sideLayout: "左侧",
      topLayout: "顶部",
      settings: "系统设置",
      closeSettings: "关闭设置"
    },
    errors: {
      loadFailed: "加载失败",
      loginFailed: "登录失败",
      ssoFailed: "SSO 登录失败"
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
      "master-sites": "Sites",
      "master-drivers": "Drivers",
      "master-vehicles": "Vehicles",
      "master-inventory": "Inventory",
      orders: "Sales Orders",
      production: "Production",
      dispatch: "Dispatch",
      weighbridge: "Weighbridge",
      delivery: "Site Signing",
      settlement: "Statements",
      procurement: "Procurement",
      finance: "Finance",
      reports: "Reports",
      system: "Approvals",
      laboratory: "Laboratory"
    },
    groups: {
      core: "Core Business",
      fulfillment: "Fulfillment",
      settlement: "Settlement",
      platform: "Platform"
    },
    auth: {
      title: "Building Materials ERP",
      username: "Username",
      password: "Password",
      mfa: "MFA Code",
      login: "Sign In"
    },
    actions: {
      collapse: "Collapse sidebar",
      search: "Search",
      refresh: "Refresh",
      theme: "Theme",
      language: "Language",
      liveEvents: "Live events",
      close: "Close",
      closeOthers: "Close others",
      closeAll: "Close all",
      cachedViews: "Cached views",
      workbench: "ERP Workspace",
      layout: "Layout",
      sideLayout: "Side",
      topLayout: "Top",
      settings: "Settings",
      closeSettings: "Close settings"
    },
    errors: {
      loadFailed: "Load failed",
      loginFailed: "Login failed",
      ssoFailed: "SSO login failed"
    }
  }
} as const;

const themeOptions = [
  { name: "Element Blue", color: "#1890ff" },
  { name: "ERP Green", color: "#13ce66" },
  { name: "Concrete Orange", color: "#f59e0b" },
  { name: "Graphite", color: "#475569" }
];

type AppRoute = (typeof nav)[number];

function normalizedPath(pathname: string) {
  return pathname.replace(/\/+$/, "") || "/";
}

function routeForPath(pathname: string) {
  const path = normalizedPath(pathname);
  return nav.find((item) => item.path === path);
}

function viewKeyFromPath(pathname: string): ViewKey {
  return routeForPath(pathname)?.key || "overview";
}

function pathForView(key: ViewKey) {
  return nav.find((item) => item.key === key)?.path || "/dashboard";
}

function savedLocale(): Locale {
  return window.localStorage.getItem("erp-locale") === "en-US" ? "en-US" : "zh-CN";
}

function savedThemeColor() {
  return window.localStorage.getItem("erp-theme-color") || themeOptions[0].color;
}

function savedLayoutMode(): LayoutMode {
  return window.localStorage.getItem("erp-layout-mode") === "top" ? "top" : "side";
}

function isAffixRoute(route: AppRoute) {
  return "affix" in route.meta && route.meta.affix;
}

function isNoCacheRoute(route: AppRoute) {
  return "noCache" in route.meta && route.meta.noCache;
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
        <section className="panel view-error-panel">
          <h3>页面加载失败</h3>
          <p className="error-text">{this.state.error.message || "视图渲染异常"}</p>
          <button className="soft-button" onClick={() => this.setState({ error: null })}>重新加载页面</button>
        </section>
      );
    }
    return this.props.children;
  }
}

export function App() {
  const publicSignToken = window.location.pathname.match(/^\/public\/sign\/([^/]+)/)?.[1];
  const [tokenReady, setTokenReady] = useState(Boolean(api.token));
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin123");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaRequired, setMfaRequired] = useState(false);
  const [error, setError] = useState("");
  const [active, setActiveView] = useState<ViewKey>(() => viewKeyFromPath(window.location.pathname));
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [visitedViews, setVisitedViews] = useState<ViewKey[]>(["overview"]);
  const [locale, setLocale] = useState<Locale>(savedLocale);
  const [themeColor, setThemeColor] = useState(savedThemeColor);
  const [layoutMode, setLayoutMode] = useState<LayoutMode>(savedLayoutMode);
  const [tagMenu, setTagMenu] = useState<{ key: ViewKey; x: number; y: number } | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [bootstrap, setBootstrap] = useState<BootstrapData | null>(null);
  const [eventCount, setEventCount] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);
  const [ssoProviders, setSSOProviders] = useState<OIDCProvider[]>([]);

  const currentNav = useMemo(() => nav.find((item) => item.key === active) || nav[0], [active]);
  const cachedViews = useMemo(
    () => visitedViews.filter((key) => !isNoCacheRoute(nav.find((item) => item.key === key) || nav[0])),
    [visitedViews]
  );
  const copy = messages[locale];
  const currentLabel = copy.nav[active];
  const shellStyle = { "--erp-theme-color": themeColor } as CSSProperties;
  const shellClass = [
    "app-shell",
    "admin-layout",
    layoutMode === "top" ? "top-nav-layout" : "side-nav-layout",
    sidebarCollapsed && layoutMode === "side" ? "sidebar-collapsed" : ""
  ].filter(Boolean).join(" ");
  const tagMenuRoute = tagMenu ? nav.find((item) => item.key === tagMenu.key) || nav[0] : null;

  async function load() {
    setBootstrap(await api.bootstrap());
  }

  function setActive(key: ViewKey, historyMode: HistoryMode = "push") {
    setActiveView(key);
    if (publicSignToken || historyMode === "none") return;
    const path = pathForView(key);
    if (normalizedPath(window.location.pathname) === path) return;
    const method = historyMode === "replace" ? "replaceState" : "pushState";
    window.history[method]({ viewKey: key }, "", path);
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
    window.localStorage.setItem("erp-locale", locale);
  }, [locale]);

  useEffect(() => {
    window.localStorage.setItem("erp-theme-color", themeColor);
  }, [themeColor]);

  useEffect(() => {
    window.localStorage.setItem("erp-layout-mode", layoutMode);
  }, [layoutMode]);

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
    if (!settingsOpen) return;
    function closeSettingsByKey(event: KeyboardEvent) {
      if (event.key === "Escape") setSettingsOpen(false);
    }
    window.addEventListener("keydown", closeSettingsByKey);
    return () => window.removeEventListener("keydown", closeSettingsByKey);
  }, [settingsOpen]);

  useEffect(() => {
    if (!tokenReady) return;
    load().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : copy.errors.loadFailed);
      setTokenReady(false);
    });
  }, [tokenReady, refreshKey]);

  useEffect(() => {
    if (tokenReady) return;
    api.ssoProviders().then(setSSOProviders).catch(() => setSSOProviders([]));
  }, [tokenReady]);

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
  }, [tokenReady]);

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
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

  function toggleLocale() {
    setLocale((value) => (value === "zh-CN" ? "en-US" : "zh-CN"));
  }

  function openTagContextMenu(event: ReactMouseEvent<HTMLDivElement>, key: ViewKey) {
    event.preventDefault();
    event.stopPropagation();
    const menuWidth = 132;
    const menuHeight = 112;
    setActive(key);
    setTagMenu({
      key,
      x: Math.max(8, Math.min(event.clientX, window.innerWidth - menuWidth - 8)),
      y: Math.max(8, Math.min(event.clientY, window.innerHeight - menuHeight - 8))
    });
  }

  if (publicSignToken) {
    return <PublicSignView token={decodeURIComponent(publicSignToken)} />;
  }

  if (!tokenReady) {
    return (
      <main className="login-shell">
        <section className="login-card panel">
          <p className="eyebrow">ERP Appliance</p>
          <h1>{copy.auth.title}</h1>
          <form onSubmit={handleLogin} className="login-form">
            <label>
              <span>{copy.auth.username}</span>
              <input value={username} onChange={(event) => setUsername(event.target.value)} />
            </label>
            <label>
              <span>{copy.auth.password}</span>
              <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
            </label>
            {mfaRequired ? (
              <label>
                <span>{copy.auth.mfa}</span>
                <input value={mfaCode} onChange={(event) => setMfaCode(event.target.value)} />
              </label>
            ) : null}
            <button className="primary-button" type="submit">{copy.auth.login}</button>
            {ssoProviders.length ? (
              <div className="row-actions">
                {ssoProviders.map((provider) => (
                  <button className="soft-button" type="button" key={provider.code} onClick={() => handleSSO(provider)}>
                    {provider.name}
                  </button>
                ))}
              </div>
            ) : null}
            {error ? <p className="error-text">{error}</p> : null}
          </form>
        </section>
      </main>
    );
  }

  return (
    <div className={shellClass} style={shellStyle}>
      {layoutMode === "side" ? (
        <aside className="side">
          <div className="brand">
            <Building2 size={28} />
            <div>
              <b>{copy.brandTitle}</b>
              <span>{bootstrap?.license.issuer || copy.workspace}</span>
            </div>
          </div>
          <div className="side-nav">
            {navGroups.map((group) => (
              <div className="sidebar-group" key={group}>
                <div className="sidebar-group-title">{copy.groups[group]}</div>
                {nav.filter((item) => item.group === group).map((item) => {
                  const Icon = item.icon;
                  return (
                    <button
                      key={item.key}
                      className={active === item.key ? "nav-item active" : "nav-item"}
                      onClick={() => setActive(item.key)}
                      title={copy.nav[item.key]}
                    >
                      <Icon size={18} />
                      <span>{copy.nav[item.key]}</span>
                    </button>
                  );
                })}
              </div>
            ))}
          </div>
        </aside>
      ) : null}
      <section className="workbench">
        <header className="topbar admin-navbar">
          <div className="navbar-left">
            {layoutMode === "side" ? (
              <button className="navbar-icon-button" type="button" onClick={() => setSidebarCollapsed((value) => !value)} title={copy.actions.collapse}>
                <Menu size={20} />
              </button>
            ) : null}
            <nav className="breadcrumbs navbar-breadcrumbs" aria-label="面包屑导航">
              <button className="breadcrumb-link" type="button" onClick={() => setActive("overview")}>
                <Home size={15} />
                {copy.actions.workbench}
              </button>
              {active !== "overview" ? (
                <>
                  <ChevronRight className="breadcrumb-separator-icon" size={15} />
                  <span className="breadcrumb-section">{copy.groups[currentNav.group]}</span>
                  <ChevronRight className="breadcrumb-separator-icon" size={15} />
                  <span className="breadcrumb-current">{currentLabel}</span>
                </>
              ) : null}
            </nav>
          </div>
          {layoutMode === "top" ? (
            <nav className="top-nav-menu" aria-label={copy.actions.topLayout}>
              {nav.map((item) => {
                const Icon = item.icon;
                return (
                  <button
                    className={active === item.key ? "top-nav-item active" : "top-nav-item"}
                    key={item.key}
                    type="button"
                    onClick={() => setActive(item.key)}
                    title={copy.nav[item.key]}
                  >
                    <Icon size={15} />
                    <span>{copy.nav[item.key]}</span>
                  </button>
                );
              })}
            </nav>
          ) : null}
          <div className="top-actions navbar-actions">
            <button className="navbar-icon-button" type="button" title={copy.actions.search}>
              <Search size={18} />
            </button>
            <button className="navbar-icon-button" type="button" onClick={() => setRefreshKey((value) => value + 1)} title={copy.actions.refresh}>
              <RefreshCw size={18} />
            </button>
            <button className="navbar-icon-button" type="button" onClick={() => setSettingsOpen(true)} title={copy.actions.settings}>
              <Settings size={18} />
            </button>
            <button className="navbar-text-button" type="button" onClick={toggleLocale} title={copy.actions.language}>
              <Languages size={18} />
              {locale === "zh-CN" ? "中文" : "EN"}
            </button>
            <div className="live-pill">{copy.actions.liveEvents} {eventCount}</div>
            <div className="user-pill">{bootstrap?.user.displayName}</div>
          </div>
        </header>
        <div className="tags-view" data-cached-views={cachedViews.join(",")}>
          <div className="tags-scroll">
            {visitedViews.map((key) => {
              const route = nav.find((item) => item.key === key) || nav[0];
              return (
                <div
                  className={active === key ? "tag-view active" : "tag-view"}
                  key={key}
                  onContextMenu={(event) => openTagContextMenu(event, key)}
                  title={`${route.path} · ${route.name}`}
                >
                  <button className="tag-view-label" type="button" onClick={() => setActive(key)}>
                    {copy.nav[key]}
                  </button>
                </div>
              );
            })}
          </div>
          <div className="tags-actions">
            <span>{copy.actions.cachedViews} {cachedViews.length}</span>
          </div>
        </div>
        {tagMenu && tagMenuRoute ? (
          <div
            className="tag-context-menu"
            style={{ left: tagMenu.x, top: tagMenu.y }}
            onClick={(event) => event.stopPropagation()}
            role="menu"
          >
            <button
              disabled={isAffixRoute(tagMenuRoute)}
              type="button"
              onClick={() => {
                closeTag(tagMenu.key);
                setTagMenu(null);
              }}
            >
              {copy.actions.close}
            </button>
            <button
              type="button"
              onClick={() => {
                closeOtherTags(tagMenu.key);
                setTagMenu(null);
              }}
            >
              {copy.actions.closeOthers}
            </button>
            <button
              type="button"
              onClick={() => {
                closeAllTags();
                setTagMenu(null);
              }}
            >
              {copy.actions.closeAll}
            </button>
          </div>
        ) : null}
        {settingsOpen ? (
          <>
            <button className="settings-mask" type="button" onClick={() => setSettingsOpen(false)} aria-label={copy.actions.closeSettings} />
            <aside className="settings-drawer" aria-label={copy.actions.settings}>
              <div className="settings-header">
                <div>
                  <p className="eyebrow">vue-admin</p>
                  <h3>{copy.actions.settings}</h3>
                </div>
                <button className="navbar-icon-button" type="button" onClick={() => setSettingsOpen(false)} title={copy.actions.closeSettings}>
                  <X size={18} />
                </button>
              </div>
              <section className="settings-section">
                <div className="settings-section-title">
                  <Palette size={16} />
                  <span>{copy.actions.theme}</span>
                </div>
                <div className="settings-theme-list" aria-label={copy.actions.theme}>
                  {themeOptions.map((item) => (
                    <button
                      className={themeColor === item.color ? "theme-dot active" : "theme-dot"}
                      key={item.color}
                      style={{ backgroundColor: item.color }}
                      title={item.name}
                      type="button"
                      onClick={() => setThemeColor(item.color)}
                    />
                  ))}
                </div>
              </section>
              <section className="settings-section">
                <div className="settings-section-title">
                  <Menu size={16} />
                  <span>{copy.actions.layout}</span>
                </div>
                <div className="layout-switcher settings-layout-switcher" role="group" aria-label={copy.actions.layout}>
                  <button className={layoutMode === "side" ? "active" : ""} type="button" onClick={() => setLayoutMode("side")}>
                    {copy.actions.sideLayout}
                  </button>
                  <button className={layoutMode === "top" ? "active" : ""} type="button" onClick={() => setLayoutMode("top")}>
                    {copy.actions.topLayout}
                  </button>
                </div>
              </section>
            </aside>
          </>
        ) : null}
        <div className="content">
          <main className="app-main fade-transform-enter" key={active} data-route-name={currentNav.name} data-no-cache={isNoCacheRoute(currentNav) ? "true" : "false"}>
            <ViewErrorBoundary viewKey={active}>
              {workbenchSections.includes(active) ? <ERPWorkbenchView section={active as ERPWorkbenchSection} bootstrap={bootstrap} onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
              {active === "laboratory" ? <LaboratoryView onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
            </ViewErrorBoundary>
          </main>
        </div>
      </section>
    </div>
  );
}
