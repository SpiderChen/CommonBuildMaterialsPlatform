import { FlaskConical, PackageCheck } from "lucide-react";
import { Component, ErrorInfo, FormEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { api, eventURL } from "./services/api";
import type { BootstrapData, OIDCProvider } from "./services/types";
import { LaboratoryView } from "./views/LaboratoryView";
import { ProductOpsView } from "./views/ProductOpsView";
import { PublicSignView } from "./views/PublicSignView";

type ViewKey = "productOps" | "laboratory";

const nav = [
  { key: "productOps", label: "产品运营台", icon: PackageCheck },
  { key: "laboratory", label: "实验室管理", icon: FlaskConical }
] as const;

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
  const [active, setActive] = useState<ViewKey>("productOps");
  const [bootstrap, setBootstrap] = useState<BootstrapData | null>(null);
  const [eventCount, setEventCount] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);
  const [ssoProviders, setSSOProviders] = useState<OIDCProvider[]>([]);

  const currentNav = useMemo(() => nav.find((item) => item.key === active) || nav[0], [active]);

  async function load() {
    setBootstrap(await api.bootstrap());
  }

  useEffect(() => {
    if (!tokenReady) return;
    load().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : "加载失败");
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
      "product_ops.alert.created",
      "product_ops.alert.handled",
      "product_ops.instance.saved",
      "product_ops.renewal.saved",
      "product_ops.renewal.closed",
      "product_ops.renewal.quote.created",
      "product_ops.renewal.contract.created",
      "product_ops.renewal.payment.created",
      "product_ops.renewal.approval.changed",
      "product_ops.renewal.invoice.created",
      "product_ops.renewal.esign.changed",
      "product_ops.renewal.integration.saved",
      "product_ops.renewal.sync.retried",
      "product_ops.renewal.sync.callback",
      "product_ops.probe.reported",
      "product_ops.telemetry.reported",
      "product_ops.monitoring.integration.saved",
      "product_ops.monitoring.rule.saved",
      "product_ops.monitoring.reported",
      "product_ops.alert.channel.saved",
      "product_ops.alert.notification.retried",
      "product_ops.rollout.created",
      "product_ops.rollout.changed",
      "product_ops.rollout.executed",
      "product_ops.system_update.task.created",
      "product_ops.system_update.reported",
      "system.update.published",
      "system.update.changed",
      "system.backup.created",
      "system.backup.restored",
      "system.backup.drill.passed",
      "system.backup.drill.failed",
      "system.gateway.route",
      "system.gateway.canary",
      "system.gateway.drain",
      "system.gateway.status",
      "system.gateway.reload",
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
      setError(err instanceof Error ? err.message : "登录失败");
    }
  }

  async function handleSSO(provider: OIDCProvider) {
    setError("");
    try {
      const start = await api.startSSO(provider.code);
      window.location.assign(start.authorizationUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : "SSO 登录失败");
    }
  }

  if (publicSignToken) {
    return <PublicSignView token={decodeURIComponent(publicSignToken)} />;
  }

  if (!tokenReady) {
    return (
      <main className="login-shell">
        <section className="login-card panel">
          <p className="eyebrow">Product Operations</p>
          <h1>产品运营交付平台</h1>
          <form onSubmit={handleLogin} className="login-form">
            <label>
              <span>账号</span>
              <input value={username} onChange={(event) => setUsername(event.target.value)} />
            </label>
            <label>
              <span>密码</span>
              <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
            </label>
            {mfaRequired ? (
              <label>
                <span>动态码</span>
                <input value={mfaCode} onChange={(event) => setMfaCode(event.target.value)} />
              </label>
            ) : null}
            <button className="primary-button" type="submit">登录平台</button>
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
    <div className="app-shell">
      <aside className="side">
        <div className="brand">
          <PackageCheck size={28} />
          <div>
            <b>产品运营交付平台</b>
            <span>{bootstrap?.license.issuer || "License Center"}</span>
          </div>
        </div>
        <div className="side-nav">
          {nav.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.key}
                className={active === item.key ? "nav-item active" : "nav-item"}
                onClick={() => setActive(item.key)}
              >
                <Icon size={18} />
                {item.label}
              </button>
            );
          })}
        </div>
      </aside>
      <section className="workbench">
        <header className="topbar">
          <div>
            <p className="eyebrow">{bootstrap?.license.edition || "Product Ops Appliance"}</p>
            <h2>{currentNav.label}</h2>
          </div>
          <div className="top-actions">
            <div className="live-pill">实时事件 {eventCount}</div>
            <div className="user-pill">{bootstrap?.user.displayName}</div>
          </div>
        </header>
        <div className="content">
          <div className="tabs">
            {nav.map((item) => (
              <button
                key={item.key}
                className={active === item.key ? "tab active" : "tab"}
                onClick={() => setActive(item.key)}
              >
                {item.label}
              </button>
            ))}
          </div>
          <ViewErrorBoundary viewKey={active}>
            {active === "productOps" ? <ProductOpsView onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
            {active === "laboratory" ? <LaboratoryView onChanged={() => setRefreshKey((value) => value + 1)} /> : null}
          </ViewErrorBoundary>
        </div>
      </section>
    </div>
  );
}
