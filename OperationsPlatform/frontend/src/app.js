const state = {
  summary: null,
  customers: [],
  alerts: [],
  packages: [],
  assignments: [],
  auditLogs: [],
  loading: true,
  error: ""
};

async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    }
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || "请求失败");
  }
  return data;
}

async function load() {
  state.loading = true;
  render();
  try {
    const [summary, customers, alerts, packages, assignments, auditLogs] = await Promise.all([
      request("/api/summary"),
      request("/api/customers"),
      request("/api/alerts"),
      request("/api/update-packages"),
      request("/api/assignments"),
      request("/api/audit-logs")
    ]);
    Object.assign(state, { summary, customers, alerts, packages, assignments, auditLogs, loading: false, error: "" });
  } catch (error) {
    state.loading = false;
    state.error = error instanceof Error ? error.message : "加载失败";
  }
  render();
}

function render() {
  const app = document.querySelector("#app");
  app.innerHTML = `
    <aside class="side">
      <div class="brand">
        <div class="brand-mark"><span></span><span></span><span></span></div>
        <div>
          <b>建材产品运营平台</b>
          <small>授权 · 告警 · 更新</small>
        </div>
      </div>
      <nav>
        <a href="#overview">运营概览</a>
        <a href="#customers">客户授权</a>
        <a href="#alerts">系统告警</a>
        <a href="#updates">更新包</a>
        <a href="#audit">审计日志</a>
      </nav>
    </aside>
    <main class="main">
      <header class="topbar">
        <div>
          <p>Internal Operations</p>
          <h1>私有化 ERP 产品运营</h1>
        </div>
        <button class="ghost" data-action="reload">刷新</button>
      </header>
      ${state.error ? `<div class="banner danger">${escapeHTML(state.error)}</div>` : ""}
      ${state.loading ? `<div class="loading">加载中...</div>` : page()}
    </main>
  `;
}

function page() {
  return `
    <section id="overview" class="section">
      <div class="kpis">
        ${kpi("有效客户", state.summary.activeCustomers)}
        ${kpi("45 天内到期", state.summary.expiringLicenses)}
        ${kpi("已过期", state.summary.expiredLicenses)}
        ${kpi("打开告警", state.summary.openAlerts)}
        ${kpi("严重告警", state.summary.criticalAlerts)}
        ${kpi("待更新分配", state.summary.pendingUpdateRollouts)}
      </div>
    </section>

    <section id="customers" class="section split">
      <div class="panel wide">
        <div class="panel-title">
          <h2>客户授权</h2>
          <span>${state.customers.length} 个私有化部署</span>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>客户</th>
                <th>授权</th>
                <th>版本</th>
                <th>健康</th>
                <th>到期日</th>
              </tr>
            </thead>
            <tbody>${state.customers.map(customerRow).join("")}</tbody>
          </table>
        </div>
      </div>
      <form class="panel form-panel" data-form="renewal">
        <h2>续费登记</h2>
        <label>
          <span>客户</span>
          <select name="customerId">${state.customers.map((item) => `<option value="${item.id}">${escapeHTML(item.customerName)}</option>`).join("")}</select>
        </label>
        <label>
          <span>新到期日</span>
          <input name="expiresAt" type="date" required />
        </label>
        <div class="form-grid">
          <label>
            <span>站点额度</span>
            <input name="maxSites" type="number" min="1" value="20" />
          </label>
          <label>
            <span>车辆额度</span>
            <input name="maxVehicles" type="number" min="1" value="5000" />
          </label>
        </div>
        <label>
          <span>版本</span>
          <input name="edition" value="Enterprise Appliance" />
        </label>
        <label>
          <span>备注</span>
          <textarea name="note" rows="3">年度授权续费完成</textarea>
        </label>
        <button class="primary" type="submit">登记续费</button>
      </form>
    </section>

    <section id="alerts" class="section">
      <div class="panel">
        <div class="panel-title">
          <h2>系统告警</h2>
          <span>${state.alerts.filter((item) => item.status !== "resolved").length} 条待处理</span>
        </div>
        <div class="alert-grid">${state.alerts.map(alertCard).join("")}</div>
      </div>
    </section>

    <section id="updates" class="section split">
      <div class="panel wide">
        <div class="panel-title">
          <h2>客户端 / 服务端更新包</h2>
          <span>${state.packages.length} 个包</span>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>包号</th>
                <th>目标</th>
                <th>版本</th>
                <th>通道</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>${state.packages.map(packageRow).join("")}</tbody>
          </table>
        </div>
      </div>
      <form class="panel form-panel" data-form="package">
        <h2>新增更新包</h2>
        <div class="form-grid">
          <label>
            <span>目标</span>
            <select name="target"><option value="client">客户端</option><option value="server">服务端</option></select>
          </label>
          <label>
            <span>通道</span>
            <select name="channel"><option value="stable">stable</option><option value="beta">beta</option></select>
          </label>
        </div>
        <label>
          <span>版本号</span>
          <input name="version" placeholder="1.4.4" required />
        </label>
        <label>
          <span>文件名</span>
          <input name="fileName" placeholder="cbmp-appliance-1.4.4.tar.gz" required />
        </label>
        <label>
          <span>校验值</span>
          <input name="checksum" placeholder="sha256:..." />
        </label>
        <div class="form-grid">
          <label>
            <span>最低版本</span>
            <input name="minVersion" value="1.3.0" />
          </label>
          <label>
            <span>灰度比例</span>
            <input name="rolloutPct" type="number" min="0" max="100" value="25" />
          </label>
        </div>
        <label>
          <span>发布说明</span>
          <textarea name="releaseNotes" rows="3"></textarea>
        </label>
        <button class="primary" type="submit">创建更新包</button>
      </form>
    </section>

    <section id="audit" class="section">
      <div class="panel">
        <div class="panel-title">
          <h2>审计日志</h2>
          <span>${state.auditLogs.length} 条</span>
        </div>
        <div class="timeline">${state.auditLogs.slice(0, 12).map(auditItem).join("")}</div>
      </div>
    </section>
  `;
}

function kpi(label, value) {
  return `<div class="kpi"><span>${label}</span><b>${value ?? 0}</b></div>`;
}

function customerRow(item) {
  return `
    <tr>
      <td>
        <b>${escapeHTML(item.customerName)}</b>
        <small>${escapeHTML(item.serverEndpoint)}</small>
      </td>
      <td>
        <span>${escapeHTML(item.licenseId)}</span>
        <small>${item.maxSites} 站 / ${item.maxVehicles} 车</small>
      </td>
      <td>
        <span>客户端 ${escapeHTML(item.currentClientVersion)}</span>
        <small>服务端 ${escapeHTML(item.currentServerVersion)}</small>
      </td>
      <td>${status(item.healthStatus)}</td>
      <td>
        <span>${escapeHTML(item.expiresAt)}</span>
        <small>${escapeHTML(item.renewalStatus)}</small>
      </td>
    </tr>
  `;
}

function alertCard(item) {
  return `
    <article class="alert-card ${escapeHTML(item.severity)}">
      <div>
        <div class="alert-head">
          ${status(item.severity)}
          ${status(item.status)}
        </div>
        <h3>${escapeHTML(item.title)}</h3>
        <p>${escapeHTML(item.message)}</p>
        <small>${escapeHTML(customerName(item.customerId))} · ${escapeHTML(item.source)} · ${escapeHTML(item.lastSeenAt)}</small>
      </div>
      <div class="row-actions">
        ${item.status === "open" ? `<button data-action="ack-alert" data-id="${item.id}">确认</button>` : ""}
        ${item.status !== "resolved" ? `<button data-action="resolve-alert" data-id="${item.id}">关闭</button>` : ""}
      </div>
    </article>
  `;
}

function packageRow(item) {
  return `
    <tr>
      <td>
        <b>${escapeHTML(item.packageNo)}</b>
        <small>${escapeHTML(item.fileName)}</small>
      </td>
      <td>${item.target === "client" ? "客户端" : "服务端"}</td>
      <td>${escapeHTML(item.version)}</td>
      <td>${escapeHTML(item.channel)} · ${item.rolloutPct}%</td>
      <td>${status(item.status)}</td>
      <td>
        <div class="table-actions">
          ${item.status !== "published" ? `<button data-action="publish-package" data-id="${item.id}">发布</button>` : ""}
          <button data-action="assign-package" data-id="${item.id}">分配</button>
        </div>
      </td>
    </tr>
  `;
}

function auditItem(item) {
  return `
    <div class="timeline-item">
      <span>${escapeHTML(item.createdAt)}</span>
      <b>${escapeHTML(item.action)}</b>
      <p>${escapeHTML(item.target)} · ${escapeHTML(item.detail)}</p>
    </div>
  `;
}

function status(value) {
  return `<span class="status ${escapeHTML(value)}">${escapeHTML(statusLabel(value))}</span>`;
}

function statusLabel(value) {
  const labels = {
    active: "有效",
    expiring: "临期",
    expired: "过期",
    healthy: "健康",
    degraded: "降级",
    critical: "严重",
    warning: "警告",
    info: "提示",
    open: "打开",
    acknowledged: "已确认",
    resolved: "已关闭",
    staged: "待发布",
    published: "已发布",
    assigned: "已分配",
    applied: "已应用"
  };
  return labels[value] || value || "-";
}

function customerName(id) {
  return state.customers.find((item) => item.id === id)?.customerName || `客户 ${id}`;
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

document.addEventListener("submit", async (event) => {
  const form = event.target;
  if (!(form instanceof HTMLFormElement)) return;
  event.preventDefault();
  const data = new FormData(form);
  try {
    if (form.dataset.form === "renewal") {
      await request(`/api/customers/${data.get("customerId")}/renewals`, {
        method: "POST",
        body: JSON.stringify({
          expiresAt: data.get("expiresAt"),
          edition: data.get("edition"),
          maxSites: Number(data.get("maxSites") || 0),
          maxVehicles: Number(data.get("maxVehicles") || 0),
          operator: "ops",
          note: data.get("note")
        })
      });
      form.reset();
    }
    if (form.dataset.form === "package") {
      await request("/api/update-packages", {
        method: "POST",
        body: JSON.stringify({
          target: data.get("target"),
          channel: data.get("channel"),
          version: data.get("version"),
          fileName: data.get("fileName"),
          checksum: data.get("checksum"),
          minVersion: data.get("minVersion"),
          rolloutPct: Number(data.get("rolloutPct") || 0),
          releaseNotes: data.get("releaseNotes")
        })
      });
      form.reset();
    }
    await load();
  } catch (error) {
    state.error = error instanceof Error ? error.message : "操作失败";
    render();
  }
});

document.addEventListener("click", async (event) => {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const action = button.dataset.action;
  const id = button.dataset.id;
  try {
    if (action === "reload") {
      await load();
      return;
    }
    if (action === "ack-alert") {
      await request(`/api/alerts/${id}/ack`, { method: "POST", body: "{}" });
    }
    if (action === "resolve-alert") {
      await request(`/api/alerts/${id}/resolve`, {
        method: "POST",
        body: JSON.stringify({ operator: "ops", resolution: "人工确认处理完成" })
      });
    }
    if (action === "publish-package") {
      await request(`/api/update-packages/${id}/publish`, { method: "POST", body: "{}" });
    }
    if (action === "assign-package") {
      await request(`/api/update-packages/${id}/assign`, {
        method: "POST",
        body: JSON.stringify({ customerIds: state.customers.map((item) => item.id) })
      });
    }
    await load();
  } catch (error) {
    state.error = error instanceof Error ? error.message : "操作失败";
    render();
  }
});

load();
