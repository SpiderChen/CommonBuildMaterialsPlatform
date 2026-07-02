import { disableNativeContextMenu, disablePageZoom } from "./disable-page-zoom.js";

disablePageZoom();
disableNativeContextMenu();

const state = {
  summary: null,
  customers: [],
  renewals: [],
  alerts: [],
  packages: [],
  assignments: [],
  auditLogs: [],
  updaterPoll: null,
  updaterReport: null,
  updaterDownload: null,
  probeReport: null,
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
    const [summary, customers, renewals, alerts, packages, assignments, auditLogs] = await Promise.all([
      request("/api/summary"),
      request("/api/customers"),
      request("/api/renewals"),
      request("/api/alerts"),
      request("/api/update-packages"),
      request("/api/assignments"),
      request("/api/audit-logs")
    ]);
    Object.assign(state, { summary, customers, renewals, alerts, packages, assignments, auditLogs, loading: false, error: "" });
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
        <a href="#update-assignments">更新执行</a>
        <a href="#updater-flow">Updater 联调</a>
        <a href="#probe-flow">探针上报</a>
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
                <th>操作</th>
              </tr>
            </thead>
            <tbody>${state.customers.map(customerRow).join("")}</tbody>
          </table>
        </div>
        <div class="subtable-title">
          <h3>授权续费记录</h3>
          <span>${state.renewals.length} 个授权包</span>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>续费单</th>
                <th>客户</th>
                <th>授权包</th>
                <th>额度</th>
                <th>下载</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>${state.renewals.length ? state.renewals.map(renewalRow).join("") : `<tr><td colspan="6"><span class="muted">暂无续费授权包</span></td></tr>`}</tbody>
          </table>
        </div>
      </div>
      <div class="form-stack">
        <form class="panel form-panel" data-form="customer">
          <h2>新增客户部署</h2>
          <label>
            <span>客户名称</span>
            <input name="customerName" required />
          </label>
          <label>
            <span>授权 ID</span>
            <input name="licenseId" required />
          </label>
          <div class="form-grid">
            <label>
              <span>产品</span>
              <input name="productName" value="CommonBuildMaterialsPlatform" />
            </label>
            <label>
              <span>版本</span>
              <input name="edition" value="Enterprise Appliance" />
            </label>
          </div>
          <label>
            <span>ERP 地址</span>
            <input name="serverEndpoint" placeholder="https://erp.customer.example.com" />
          </label>
          <div class="form-grid">
            <label>
              <span>部署模式</span>
              <select name="deploymentMode">
                <option value="private_server">私有化服务器</option>
                <option value="private_cloud">私有云</option>
                <option value="edge_appliance">边缘一体机</option>
              </select>
            </label>
            <label>
              <span>环境</span>
              <select name="environment">
                <option value="production">生产</option>
                <option value="staging">预发</option>
                <option value="test">测试</option>
              </select>
            </label>
          </div>
          <div class="form-grid">
            <label>
              <span>联系人</span>
              <input name="contactName" />
            </label>
            <label>
              <span>联系电话</span>
              <input name="contactPhone" />
            </label>
          </div>
          <label>
            <span>授权到期日</span>
            <input name="expiresAt" type="date" required />
          </label>
          <div class="form-grid">
            <label>
              <span>站点额度</span>
              <input name="maxSites" type="number" min="1" value="1" />
            </label>
            <label>
              <span>车辆额度</span>
              <input name="maxVehicles" type="number" min="1" value="1" />
            </label>
          </div>
          <div class="form-grid">
            <label>
              <span>当前客户端</span>
              <input name="currentClientVersion" placeholder="1.0.0" />
            </label>
            <label>
              <span>当前服务端</span>
              <input name="currentServerVersion" placeholder="1.0.0" />
            </label>
          </div>
          <div class="form-grid">
            <label>
              <span>目标客户端</span>
              <input name="targetClientVersion" placeholder="1.0.0" />
            </label>
            <label>
              <span>目标服务端</span>
              <input name="targetServerVersion" placeholder="1.0.0" />
            </label>
          </div>
          <label>
            <span>Updater Token</span>
            <input name="updaterToken" placeholder="留空自动生成" />
          </label>
          <label>
            <span>开通模块</span>
            <input name="modules" value="erp,production,dispatch" />
          </label>
          <label>
            <span>备注</span>
            <textarea name="notes" rows="2"></textarea>
          </label>
          <button class="primary" type="submit">创建客户台账</button>
        </form>
        <form class="panel form-panel" data-form="renewal">
          <h2>续费登记</h2>
          <label>
            <span>客户</span>
            <select name="customerId" ${state.customers.length ? "" : "disabled"}>${state.customers.map((item) => `<option value="${item.id}">${escapeHTML(item.customerName)}</option>`).join("")}</select>
          </label>
          <label>
            <span>新到期日</span>
            <input name="expiresAt" type="date" required ${state.customers.length ? "" : "disabled"} />
          </label>
          <div class="form-grid">
            <label>
              <span>站点额度</span>
              <input name="maxSites" type="number" min="1" value="20" ${state.customers.length ? "" : "disabled"} />
            </label>
            <label>
              <span>车辆额度</span>
              <input name="maxVehicles" type="number" min="1" value="5000" ${state.customers.length ? "" : "disabled"} />
            </label>
          </div>
          <label>
            <span>版本</span>
            <input name="edition" value="Enterprise Appliance" ${state.customers.length ? "" : "disabled"} />
          </label>
          <label>
            <span>备注</span>
            <textarea name="note" rows="3" ${state.customers.length ? "" : "disabled"}>年度授权续费完成</textarea>
          </label>
          <button class="primary" type="submit" ${state.customers.length ? "" : "disabled"}>登记续费</button>
        </form>
      </div>
    </section>

    <section id="alerts" class="section split">
      <div class="panel wide">
        <div class="panel-title">
          <h2>系统告警</h2>
          <span>${state.alerts.filter((item) => item.status !== "resolved").length} 条待处理</span>
        </div>
        <div class="alert-grid">${state.alerts.map(alertCard).join("")}</div>
      </div>
      <form class="panel form-panel" data-form="alert">
        <h2>新增告警</h2>
        <label>
          <span>客户</span>
          <select name="customerId" ${state.customers.length ? "" : "disabled"}>${state.customers.map((item) => `<option value="${item.id}">${escapeHTML(item.customerName)}</option>`).join("")}</select>
        </label>
        <div class="form-grid">
          <label>
            <span>来源</span>
            <select name="source" ${state.customers.length ? "" : "disabled"}>
              <option value="server">服务端</option>
              <option value="client">客户端</option>
              <option value="license">授权</option>
              <option value="backup">备份</option>
            </select>
          </label>
          <label>
            <span>等级</span>
            <select name="severity" ${state.customers.length ? "" : "disabled"}>
              <option value="critical">严重</option>
              <option value="warning">警告</option>
              <option value="info">提示</option>
            </select>
          </label>
        </div>
        <label>
          <span>标题</span>
          <input name="title" required ${state.customers.length ? "" : "disabled"} />
        </label>
        <label>
          <span>说明</span>
          <textarea name="message" rows="3" required ${state.customers.length ? "" : "disabled"}></textarea>
        </label>
        <label>
          <span>负责人</span>
          <input name="assignee" value="交付运维" ${state.customers.length ? "" : "disabled"} />
        </label>
        <button class="primary" type="submit" ${state.customers.length ? "" : "disabled"}>创建告警</button>
      </form>
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
                <th>分配</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>${state.packages.map(packageRow).join("")}</tbody>
          </table>
        </div>
      </div>
      <div class="form-stack">
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
            <span>制品文件</span>
            <input name="artifact" type="file" required />
          </label>
          <label>
            <span>校验值</span>
            <input name="checksum" placeholder="留空时按制品内容自动计算" />
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
        <form class="panel form-panel" data-form="assignment">
          <h2>指定客户分配</h2>
          <label>
            <span>更新包</span>
            <select name="packageId" ${publishedPackages().length ? "" : "disabled"}>${publishedPackages().map((item) => `<option value="${item.id}">${escapeHTML(item.packageNo)} · ${escapeHTML(item.target)} ${escapeHTML(item.version)}</option>`).join("")}</select>
          </label>
          <label>
            <span>客户</span>
            <select name="customerIds" multiple size="5" ${state.customers.length ? "" : "disabled"}>${state.customers.map((item) => `<option value="${item.id}">${escapeHTML(item.customerName)}</option>`).join("")}</select>
          </label>
          <button class="primary" type="submit" ${publishedPackages().length && state.customers.length ? "" : "disabled"}>分配选中客户</button>
        </form>
      </div>
    </section>

    <section id="update-assignments" class="section">
      <div class="panel">
        <div class="panel-title">
          <h2>更新分配执行</h2>
          <span>${assignmentBoardSummary()}</span>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>客户</th>
                <th>更新包</th>
                <th>状态</th>
                <th>执行进度</th>
                <th>时间</th>
                <th>结果</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>${state.assignments.length ? state.assignments.map(assignmentRow).join("") : `<tr><td colspan="7"><span class="muted">暂无更新分配任务</span></td></tr>`}</tbody>
          </table>
        </div>
      </div>
    </section>

    <section id="updater-flow" class="section split">
      <div class="panel wide">
        <div class="panel-title">
          <h2>Updater 任务闭环</h2>
          <span>${state.updaterPoll?.tasks?.length ?? 0} 个待执行任务</span>
        </div>
        ${updaterPollResult()}
        ${resultPanel("最近下载", state.updaterDownload)}
        ${resultPanel("最近回报", state.updaterReport)}
      </div>
      <div class="form-stack">
        <form class="panel form-panel" data-form="updater-poll">
          <h2>拉取任务</h2>
          <label>
            <span>客户 Token</span>
            <select name="updaterToken" ${state.customers.length ? "" : "disabled"}>${customerTokenOptions()}</select>
          </label>
          <button class="primary" type="submit" ${state.customers.length ? "" : "disabled"}>拉取待更新任务</button>
        </form>
        <form class="panel form-panel" data-form="updater-report">
          <h2>执行回报</h2>
          <label>
            <span>分配任务</span>
            <select name="assignmentId" ${state.assignments.length ? "" : "disabled"}>${assignmentOptions()}</select>
          </label>
          <div class="form-grid">
            <label>
              <span>状态</span>
              <select name="status" ${state.assignments.length ? "" : "disabled"}>
                <option value="running">执行中</option>
                <option value="downloaded">已下载</option>
                <option value="applied">已应用</option>
                <option value="failed">失败</option>
                <option value="rolled_back">已回滚</option>
              </select>
            </label>
            <label>
              <span>进度</span>
              <input name="progress" type="number" min="0" max="100" value="100" ${state.assignments.length ? "" : "disabled"} />
            </label>
          </div>
          <label>
            <span>当前版本</span>
            <input name="currentVersion" ${state.assignments.length ? "" : "disabled"} />
          </label>
          <label>
            <span>Updater 版本</span>
            <input name="updaterVersion" value="ops-console" ${state.assignments.length ? "" : "disabled"} />
          </label>
          <label>
            <span>消息</span>
            <textarea name="message" rows="2" ${state.assignments.length ? "" : "disabled"}>执行回报完成</textarea>
          </label>
          <label>
            <span>错误</span>
            <textarea name="error" rows="2" ${state.assignments.length ? "" : "disabled"}></textarea>
          </label>
          <button class="primary" type="submit" ${state.assignments.length ? "" : "disabled"}>提交执行回报</button>
        </form>
      </div>
    </section>

    <section id="probe-flow" class="section split">
      <div class="panel wide">
        <div class="panel-title">
          <h2>探针告警上报</h2>
          <span>${state.probeReport?.alertNo ? state.probeReport.alertNo : "待上报"}</span>
        </div>
        ${resultPanel("最近探针告警", state.probeReport)}
      </div>
      <form class="panel form-panel" data-form="probe-alert">
        <h2>上报告警</h2>
        <label>
          <span>客户 Token</span>
          <select name="updaterToken" ${state.customers.length ? "" : "disabled"}>${customerTokenOptions()}</select>
        </label>
        <div class="form-grid">
          <label>
            <span>来源</span>
            <select name="source" ${state.customers.length ? "" : "disabled"}>
              <option value="server">服务端</option>
              <option value="client">客户端</option>
              <option value="license">授权</option>
              <option value="backup">备份</option>
            </select>
          </label>
          <label>
            <span>等级</span>
            <select name="severity" ${state.customers.length ? "" : "disabled"}>
              <option value="warning">警告</option>
              <option value="critical">严重</option>
              <option value="info">提示</option>
            </select>
          </label>
        </div>
        <label>
          <span>标题</span>
          <input name="title" required ${state.customers.length ? "" : "disabled"} />
        </label>
        <label>
          <span>说明</span>
          <textarea name="message" rows="3" required ${state.customers.length ? "" : "disabled"}></textarea>
        </label>
        <label class="check-row">
          <input name="autoResolve" type="checkbox" ${state.customers.length ? "" : "disabled"} />
          <span>自动关闭</span>
        </label>
        <label>
          <span>关闭说明</span>
          <textarea name="resolution" rows="2" ${state.customers.length ? "" : "disabled"}></textarea>
        </label>
        <button class="primary" type="submit" ${state.customers.length ? "" : "disabled"}>通过探针入口上报</button>
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

function publishedPackages() {
  return state.packages.filter((item) => item.status === "published");
}

function customerById(id) {
  return state.customers.find((item) => String(item.id) === String(id));
}

function customerByToken(token) {
  return state.customers.find((item) => item.updaterToken === token);
}

function packageById(id) {
  return state.packages.find((item) => String(item.id) === String(id));
}

function assignmentById(id) {
  return state.assignments.find((item) => String(item.id) === String(id));
}

function assignmentTaskNo(item) {
  return `UA${item.id}`;
}

function customerTokenOptions() {
  return state.customers.map((item) => `<option value="${escapeHTML(item.updaterToken)}">${escapeHTML(item.customerName)} · ${escapeHTML(item.updaterToken)}</option>`).join("");
}

function assignmentOptions() {
  return state.assignments.map((item) => {
    const pkg = packageById(item.packageId);
    return `<option value="${item.id}">${escapeHTML(assignmentTaskNo(item))} · ${escapeHTML(customerName(item.customerId))} · ${escapeHTML(pkg?.version || "-")}</option>`;
  }).join("");
}

function customerRow(item) {
  return `
    <tr data-context="customer" data-id="${item.id}">
      <td>
        <b>${escapeHTML(item.customerName)}</b>
        <small>${escapeHTML(item.serverEndpoint)}</small>
      </td>
      <td>
        <span>${escapeHTML(item.licenseId)}</span>
        <small>Updater ${escapeHTML(item.updaterToken || "-")}</small>
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
      <td>
        <div class="table-actions">
          <button data-action="copy-updater-token" data-id="${item.id}">复制 Token</button>
          <button data-action="poll-customer-updates" data-id="${item.id}">拉任务</button>
        </div>
      </td>
    </tr>
  `;
}

function renewalRow(item) {
  return `
    <tr data-context="renewal" data-id="${item.id}">
      <td>
        <b>${escapeHTML(item.renewalNo)}</b>
        <small>${escapeHTML(item.createdAt)}</small>
      </td>
      <td>
        <span>${escapeHTML(customerName(item.customerId))}</span>
        <small>${escapeHTML(item.licenseId)}</small>
      </td>
      <td>
        <span>${escapeHTML(item.licensePackageNo || "-")}</span>
        <small>指纹 ${escapeHTML(item.publicKeyFingerprint || "-")}</small>
      </td>
      <td>
        <span>${escapeHTML(item.newExpiresAt)}</span>
        <small>${item.maxSites} 站 / ${item.maxVehicles} 车</small>
      </td>
      <td>
        <span>${item.downloadCount || 0} 次</span>
        <small>${escapeHTML(item.lastDownloadedAt || "未下载")}</small>
      </td>
      <td>
        <div class="table-actions">
          <button data-action="download-license-package" data-id="${item.id}">下载授权包</button>
        </div>
      </td>
    </tr>
  `;
}

function alertCard(item) {
  return `
    <article class="alert-card ${escapeHTML(item.severity)}" data-context="alert" data-id="${item.id}" data-status="${escapeHTML(item.status)}">
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
    <tr data-context="package" data-id="${item.id}" data-status="${escapeHTML(item.status)}">
      <td>
        <b>${escapeHTML(item.packageNo)}</b>
        <small>${escapeHTML(item.fileName)}</small>
      </td>
      <td>${item.target === "client" ? "客户端" : "服务端"}</td>
      <td>${escapeHTML(item.version)}</td>
      <td>${escapeHTML(item.channel)} · ${item.rolloutPct}%</td>
      <td>${status(item.status)}</td>
      <td>${assignmentSummary(item.id)}</td>
      <td>
        <div class="table-actions">
          ${item.status !== "published" ? `<button data-action="publish-package" data-id="${item.id}">发布</button>` : ""}
          <button data-action="assign-package" data-id="${item.id}" ${state.customers.length ? "" : "disabled"}>分配</button>
        </div>
      </td>
    </tr>
  `;
}

function assignmentSummary(packageId) {
  const items = state.assignments.filter((item) => item.packageId === packageId);
  if (!items.length) return `<span class="muted">未分配</span>`;
  const counts = items.reduce((acc, item) => {
    acc[item.status] = (acc[item.status] || 0) + 1;
    return acc;
  }, {});
  return [
    counts.assigned ? `待执行 ${counts.assigned}` : "",
    counts.running ? `执行中 ${counts.running}` : "",
    counts.applied ? `已应用 ${counts.applied}` : "",
    counts.failed ? `失败 ${counts.failed}` : "",
    counts.rolled_back ? `已回滚 ${counts.rolled_back}` : ""
  ].filter(Boolean).join(" / ") || `${items.length} 个分配`;
}

function assignmentBoardSummary() {
  if (!state.assignments.length) return "0 个任务";
  const active = state.assignments.filter((item) => ["assigned", "running", "downloaded"].includes(item.status)).length;
  const failed = state.assignments.filter((item) => item.status === "failed").length;
  const applied = state.assignments.filter((item) => item.status === "applied").length;
  return `${state.assignments.length} 个任务 / 待执行 ${active} / 已应用 ${applied} / 失败 ${failed}`;
}

function assignmentRow(item) {
  const pkg = state.packages.find((candidate) => candidate.id === item.packageId);
  return `
    <tr data-context="assignment" data-id="${item.id}" data-status="${escapeHTML(item.status)}" data-customer-id="${item.customerId}" data-package-id="${item.packageId}" data-task-no="${escapeHTML(assignmentTaskNo(item))}">
      <td>
        <b>${escapeHTML(customerName(item.customerId))}</b>
        <small>任务 ${escapeHTML(assignmentTaskNo(item))}</small>
      </td>
      <td>
        <span>${escapeHTML(pkg?.packageNo || `更新包 ${item.packageId}`)}</span>
        <small>${escapeHTML([pkg?.target === "server" ? "服务端" : "客户端", pkg?.version, pkg?.channel].filter(Boolean).join(" · "))}</small>
      </td>
      <td>${status(item.status)}</td>
      <td>
        <div class="progress-cell">
          <div class="progress-track"><span style="width: ${assignmentProgress(item)}%"></span></div>
          <small>${assignmentProgress(item)}% · ${escapeHTML(item.step || "-")}</small>
        </div>
      </td>
      <td>
        <span>${escapeHTML(item.updatedAt || item.assignedAt || "-")}</span>
        <small>${escapeHTML(item.downloadedAt ? `下载 ${item.downloadedAt}` : item.appliedAt ? `应用 ${item.appliedAt}` : "等待 updater 回传")}</small>
      </td>
      <td>
        <span>${escapeHTML(item.message || item.error || "-")}</span>
        <small>${escapeHTML(item.updaterVersion ? `Updater ${item.updaterVersion}` : "")}</small>
      </td>
      <td>
        <div class="table-actions">
          <button data-action="download-assignment-package" data-id="${item.id}" ${pkg?.status === "published" ? "" : "disabled"}>下载制品</button>
          <button data-action="report-assignment-applied" data-id="${item.id}">回报已应用</button>
        </div>
      </td>
    </tr>
  `;
}

function assignmentProgress(item) {
  const value = Number(item.progress || 0);
  if (!Number.isFinite(value)) return 0;
  return Math.max(0, Math.min(100, Math.round(value)));
}

function auditItem(item) {
  return `
    <div class="timeline-item" data-context="audit">
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
    running: "执行中",
    downloaded: "已下载",
    applied: "已应用",
    failed: "失败",
    rolled_back: "已回滚"
  };
  return labels[value] || value || "-";
}

function customerName(id) {
  return state.customers.find((item) => item.id === id)?.customerName || `客户 ${id}`;
}

function updaterPollResult() {
  if (!state.updaterPoll) {
    return `<div class="result-empty">-</div>`;
  }
  const tasks = state.updaterPoll.tasks || [];
  return `
    <div class="result-card">
      <div class="result-head">
        <b>${escapeHTML(state.updaterPoll.instance?.customerName || "-")}</b>
        <span>${escapeHTML(state.updaterPoll.instance?.healthStatus || "-")} · ${escapeHTML(state.updaterPoll.instance?.lastHeartbeatAt || "-")}</span>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>任务</th>
              <th>组件</th>
              <th>版本</th>
              <th>制品</th>
              <th>下载</th>
            </tr>
          </thead>
          <tbody>${tasks.length ? tasks.map((item) => `
            <tr data-context="assignment" data-id="${item.id}" data-status="${escapeHTML(item.status)}" data-customer-id="${item.instanceId}" data-package-id="${item.updateId}" data-task-no="${escapeHTML(item.taskNo)}">
              <td><b>${escapeHTML(item.taskNo)}</b><small>${escapeHTML(item.status)}</small></td>
              <td>${escapeHTML(item.component)}</td>
              <td><span>${escapeHTML(item.fromVersion || "-")} -> ${escapeHTML(item.version)}</span></td>
              <td><span>${escapeHTML(item.artifactFileName)}</span><small>${escapeHTML(item.checksum)}</small></td>
              <td><span>${escapeHTML(item.downloadUrl)}</span></td>
            </tr>
          `).join("") : `<tr><td colspan="5"><span class="muted">暂无待执行任务</span></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  `;
}

function resultPanel(title, value) {
  if (!value) return "";
  return `
    <div class="result-card">
      <div class="result-head">
        <b>${escapeHTML(title)}</b>
        <span>${escapeHTML(value.generatedAt || value.updatedAt || value.lastSeenAt || value.status || "")}</span>
      </div>
      <pre>${escapeHTML(JSON.stringify(value, null, 2))}</pre>
    </div>
  `;
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

async function copyText(text) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "true");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
}

async function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const value = String(reader.result || "");
      resolve(value.includes(",") ? value.split(",").pop() : value);
    };
    reader.onerror = () => reject(reader.error || new Error("读取制品文件失败"));
    reader.readAsDataURL(file);
  });
}

function downloadBase64File(data, contentField = "artifactContentBase64") {
  const binary = atob(data[contentField] || "");
  const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
  const blob = new Blob([bytes], { type: data.contentType || data.artifactContentType || "application/octet-stream" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = data.artifactFileName || data.fileName || "download.bin";
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

async function downloadLicensePackage(id) {
  const data = await request(`/api/renewals/${id}/license-package`);
  downloadBase64File(data, "contentBase64");
}

async function pollUpdaterTasks(updaterToken) {
  const result = await request("/api/product-ops/system-updates/tasks", {
    method: "POST",
    body: JSON.stringify({ updaterToken })
  });
  state.updaterPoll = result;
  return result;
}

async function downloadAssignmentPackage(assignmentId) {
  const assignment = assignmentById(assignmentId);
  if (!assignment) throw new Error("更新分配任务不存在");
  const customer = customerById(assignment.customerId);
  if (!customer?.updaterToken) throw new Error("客户缺少 updater token");
  const result = await request(`/api/system/updates/${assignment.packageId}/download?assignmentId=${assignment.id}`, {
    headers: { "X-CBMP-Updater-Token": customer.updaterToken }
  });
  state.updaterDownload = {
    fileName: result.fileName,
    artifactFileName: result.artifactFileName,
    artifactSha256: result.artifactSha256,
    verified: result.verified,
    generatedAt: result.generatedAt,
    manifest: result.manifest
  };
  downloadBase64File(result);
  return result;
}

async function reportAssignment(assignmentId, payload = {}) {
  const assignment = assignmentById(assignmentId);
  if (!assignment) throw new Error("更新分配任务不存在");
  const customer = customerById(assignment.customerId);
  if (!customer?.updaterToken) throw new Error("客户缺少 updater token");
  const pkg = packageById(assignment.packageId);
  const statusValue = payload.status || "applied";
  const result = await request(`/api/product-ops/system-updates/tasks/${assignmentTaskNo(assignment)}/report`, {
    method: "POST",
    body: JSON.stringify({
      updaterToken: customer.updaterToken,
      status: statusValue,
      progress: Number(payload.progress ?? (statusValue === "running" ? 50 : 100)),
      step: payload.step || statusValue,
      message: payload.message || "执行回报完成",
      error: payload.error || "",
      currentVersion: payload.currentVersion || pkg?.version || "",
      updaterVersion: payload.updaterVersion || "ops-console"
    })
  });
  state.updaterReport = result;
  return result;
}

function showToast(message) {
  const existing = document.querySelector(".ops-toast");
  existing?.remove();
  const toast = document.createElement("div");
  toast.className = "ops-toast";
  toast.textContent = message;
  document.body.appendChild(toast);
  window.setTimeout(() => toast.remove(), 1800);
}

async function runAction(action, id) {
  if (action === "reload") {
    await load();
    return;
  }
  if (action === "download-license-package") {
    await downloadLicensePackage(id);
    await load();
    return;
  }
  if (action === "copy-updater-token") {
    const customer = customerById(id);
    await copyText(customer?.updaterToken || "");
    showToast("Updater Token 已复制");
    return;
  }
  if (action === "poll-customer-updates") {
    const customer = customerById(id);
    if (!customer?.updaterToken) throw new Error("客户缺少 updater token");
    await pollUpdaterTasks(customer.updaterToken);
    await load();
    return;
  }
  if (action === "download-assignment-package") {
    await downloadAssignmentPackage(id);
    await load();
    return;
  }
  if (action === "report-assignment-applied") {
    await reportAssignment(id, { status: "applied" });
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
}

function closeContextMenu() {
  document.querySelector(".ops-context-menu")?.remove();
}

function contextMenuItem(label, action, options = {}) {
  return { label, action, disabled: Boolean(options.disabled), danger: Boolean(options.danger) };
}

function contextMenuItems(target) {
  const items = [];
  const contextTarget = target.closest("[data-context]");
  const context = contextTarget?.dataset.context;
  const id = contextTarget?.dataset.id;
  const statusValue = contextTarget?.dataset.status;
  const textTarget = contextTarget || target.closest(".panel, .section, .main") || document.body;
  const text = textTarget.innerText.trim();

  if (context === "alert" && id) {
    items.push(contextMenuItem("确认告警", () => runAction("ack-alert", id), { disabled: statusValue !== "open" }));
    items.push(contextMenuItem("关闭告警", () => runAction("resolve-alert", id), { disabled: statusValue === "resolved", danger: true }));
    items.push(contextMenuItem("separator", null));
  }

  if (context === "package" && id) {
    items.push(contextMenuItem("发布更新包", () => runAction("publish-package", id), { disabled: statusValue === "published" }));
    items.push(contextMenuItem("分配全部客户", () => runAction("assign-package", id), { disabled: !state.customers.length }));
    items.push(contextMenuItem("separator", null));
  }

  if (context === "customer" && id) {
    items.push(contextMenuItem("复制 Updater Token", () => runAction("copy-updater-token", id)));
    items.push(contextMenuItem("拉取更新任务", () => runAction("poll-customer-updates", id)));
    items.push(contextMenuItem("separator", null));
  }

  if (context === "assignment" && id) {
    items.push(contextMenuItem("下载分配制品", () => runAction("download-assignment-package", id), { disabled: statusValue !== "assigned" && statusValue !== "running" && statusValue !== "downloaded" }));
    items.push(contextMenuItem("回报已应用", () => runAction("report-assignment-applied", id), { disabled: statusValue === "applied" }));
    items.push(contextMenuItem("separator", null));
  }

  if (context === "renewal" && id) {
    items.push(contextMenuItem("下载授权包", () => runAction("download-license-package", id)));
    items.push(contextMenuItem("separator", null));
  }

  items.push(
    contextMenuItem(context === "customer" || context === "package" || context === "renewal" ? "复制本行文本" : "复制当前区域", async () => {
      await copyText(text);
      showToast("已复制");
    }, { disabled: !text }),
    contextMenuItem("复制页面地址", async () => {
      await copyText(window.location.href);
      showToast("页面地址已复制");
    }),
    contextMenuItem("刷新数据", () => runAction("reload"))
  );

  return items;
}

function openContextMenu(event) {
  const target = event.target instanceof Element ? event.target : null;
  if (!target) return;
  const items = contextMenuItems(target);
  if (!items.length) return;

  event.preventDefault();
  event.stopPropagation();
  closeContextMenu();

  const menu = document.createElement("div");
  menu.className = "ops-context-menu";
  menu.setAttribute("role", "menu");
  items.forEach((item) => {
    if (!item.action) {
      const separator = document.createElement("div");
      separator.className = "ops-context-menu-separator";
      separator.setAttribute("role", "separator");
      menu.appendChild(separator);
      return;
    }

    const button = document.createElement("button");
    button.type = "button";
    button.textContent = item.label;
    button.disabled = item.disabled;
    if (item.danger) button.className = "danger";
    button.addEventListener("click", async () => {
      closeContextMenu();
      try {
        await item.action();
      } catch (error) {
        state.error = error instanceof Error ? error.message : "操作失败";
        render();
      }
    });
    menu.appendChild(button);
  });
  document.body.appendChild(menu);

  const width = 184;
  const height = Math.min(360, 10 + items.length * 34);
  menu.style.left = `${Math.max(8, Math.min(event.clientX, window.innerWidth - width - 8))}px`;
  menu.style.top = `${Math.max(8, Math.min(event.clientY, window.innerHeight - height - 8))}px`;
}

function runPrimaryContextAction(target) {
  const contextTarget = target.closest("[data-context]");
  const context = contextTarget?.dataset.context;
  if (context !== "alert" && context !== "package" && context !== "renewal" && context !== "customer" && context !== "assignment") return;
  const action = contextMenuItems(target).find((item) => item.action && !item.disabled);
  if (action) {
    void action.action().catch((error) => {
      state.error = error instanceof Error ? error.message : "操作失败";
      render();
    });
  }
}

document.addEventListener("submit", async (event) => {
  const form = event.target;
  if (!(form instanceof HTMLFormElement)) return;
  event.preventDefault();
  const data = new FormData(form);
  try {
    if (form.dataset.form === "customer") {
      await request("/api/customers", {
        method: "POST",
        body: JSON.stringify({
          customerName: data.get("customerName"),
          productName: data.get("productName"),
          licenseId: data.get("licenseId"),
          updaterToken: data.get("updaterToken"),
          edition: data.get("edition"),
          deploymentMode: data.get("deploymentMode"),
          environment: data.get("environment"),
          serverEndpoint: data.get("serverEndpoint"),
          contactName: data.get("contactName"),
          contactPhone: data.get("contactPhone"),
          expiresAt: data.get("expiresAt"),
          maxSites: Number(data.get("maxSites") || 0),
          maxVehicles: Number(data.get("maxVehicles") || 0),
          modules: String(data.get("modules") || "")
            .split(",")
            .map((item) => item.trim())
            .filter(Boolean),
          operator: "ops",
          currentClientVersion: data.get("currentClientVersion"),
          currentServerVersion: data.get("currentServerVersion"),
          targetClientVersion: data.get("targetClientVersion"),
          targetServerVersion: data.get("targetServerVersion"),
          notes: data.get("notes")
        })
      });
      form.reset();
    }
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
    if (form.dataset.form === "alert") {
      await request("/api/alerts", {
        method: "POST",
        body: JSON.stringify({
          customerId: Number(data.get("customerId") || 0),
          source: data.get("source"),
          severity: data.get("severity"),
          title: data.get("title"),
          message: data.get("message"),
          assignee: data.get("assignee"),
          operator: "ops"
        })
      });
      form.reset();
    }
    if (form.dataset.form === "package") {
      const file = form.elements.artifact?.files?.[0];
      if (!file) throw new Error("请选择真实制品文件");
      const artifactContentBase64 = await fileToBase64(file);
      await request("/api/update-packages", {
        method: "POST",
        body: JSON.stringify({
          target: data.get("target"),
          channel: data.get("channel"),
          version: data.get("version"),
          fileName: data.get("fileName") || file.name,
          checksum: data.get("checksum"),
          artifactContentType: file.type || "application/octet-stream",
          artifactContentBase64,
          minVersion: data.get("minVersion"),
          rolloutPct: Number(data.get("rolloutPct") || 0),
          releaseNotes: data.get("releaseNotes")
        })
      });
      form.reset();
    }
    if (form.dataset.form === "assignment") {
      const customerIds = data.getAll("customerIds").map((item) => Number(item)).filter(Boolean);
      if (!customerIds.length) throw new Error("请选择要分配的客户");
      await request(`/api/update-packages/${data.get("packageId")}/assign`, {
        method: "POST",
        body: JSON.stringify({ customerIds })
      });
      form.reset();
    }
    if (form.dataset.form === "updater-poll") {
      state.updaterPoll = await pollUpdaterTasks(String(data.get("updaterToken") || ""));
    }
    if (form.dataset.form === "updater-report") {
      state.updaterReport = await reportAssignment(data.get("assignmentId"), {
        status: data.get("status"),
        progress: Number(data.get("progress") || 0),
        step: data.get("status"),
        currentVersion: data.get("currentVersion"),
        updaterVersion: data.get("updaterVersion"),
        message: data.get("message"),
        error: data.get("error")
      });
    }
    if (form.dataset.form === "probe-alert") {
      state.probeReport = await request("/api/product-ops/alerts/report", {
        method: "POST",
        body: JSON.stringify({
          updaterToken: data.get("updaterToken"),
          source: data.get("source"),
          severity: data.get("severity"),
          title: data.get("title"),
          message: data.get("message"),
          autoResolve: data.get("autoResolve") === "on",
          resolution: data.get("resolution"),
          operator: "probe"
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
  closeContextMenu();
  if (!button) return;
  const action = button.dataset.action;
  const id = button.dataset.id;
  try {
    await runAction(action, id);
  } catch (error) {
    state.error = error instanceof Error ? error.message : "操作失败";
    render();
  }
});

document.addEventListener("contextmenu", openContextMenu);
document.addEventListener("dblclick", (event) => {
  const target = event.target instanceof Element ? event.target : null;
  if (!target || target.closest("button, a, input, textarea, select")) return;
  runPrimaryContextAction(target);
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeContextMenu();
});
window.addEventListener("resize", closeContextMenu);

load();
