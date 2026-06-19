import { useEffect, useMemo, useState } from "react";
import { StatusChip } from "../components/StatusChip";
import { api } from "../services/api";
import type { BackupDrill, BackupInfo, BootstrapData, GatewayOverview, GatewayRoute, ProductAlertChannel, ProductAlertNotification, ProductAlertPolicy, ProductAlertRule, ProductInstance, ProductMonitoringIntegration, ProductOpsOverview, ProductRenewalIntegration, ProductRenewalSyncRecord, ProductRenewalTask, ProductSystemUpdateTask, ProductUpdateExecution, ProductUpdateRollout, SystemAlert, UpdatePackage } from "../services/types";

const textEncoder = new TextEncoder();

function bytesToBase64(bytes: Uint8Array) {
  let binary = "";
  bytes.forEach((byte) => { binary += String.fromCharCode(byte); });
  return window.btoa(binary);
}

function base64ToBytes(value: string) {
  const binary = window.atob(value);
  const bytes = new Uint8Array(binary.length);
  for (let index = 0; index < binary.length; index += 1) {
    bytes[index] = binary.charCodeAt(index);
  }
  return bytes;
}

async function sha256Checksum(text: string) {
  const digest = await crypto.subtle.digest("SHA-256", textEncoder.encode(text));
  const hex = Array.from(new Uint8Array(digest)).map((byte) => byte.toString(16).padStart(2, "0")).join("");
  return `sha256:${hex}`;
}

function formatBytes(value: number) {
  if (!value) return "0 B";
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`;
  return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

type GatewayRouteForm = {
  id: string;
  name: string;
  pathPrefix: string;
  stableUpstream: string;
  canaryUpstream: string;
  canaryPercent: string;
  readTimeoutSec: string;
  status: string;
};

export function ProductOpsView({ onChanged }: { onChanged: () => void }) {
  const [overview, setOverview] = useState<ProductOpsOverview | null>(null);
  const [masterData, setMasterData] = useState<BootstrapData | null>(null);
  const [backups, setBackups] = useState<BackupInfo[]>([]);
  const [backupDrills, setBackupDrills] = useState<BackupDrill[]>([]);
  const [backupBusy, setBackupBusy] = useState("");
  const [gateway, setGateway] = useState<GatewayOverview | null>(null);
  const [gatewayBusy, setGatewayBusy] = useState("");
  const [error, setError] = useState("");
  const [instanceForm, setInstanceForm] = useState({
    id: "0",
    customerName: "新交付客户",
    licenseId: "LIC-NEW-2026",
    watermark: "CBMP-NEW",
    edition: "Operations",
    deploymentMode: "private",
    clientVersion: "1.0.1",
    serverVersion: "1.0.1",
    endpoint: "https://cbmp.customer.example",
    status: "online",
    probeToken: "",
    probeEnabled: true,
    healthStatus: "healthy",
    licenseExpiresAt: "2026-12-31",
    renewalOwner: "客户成功",
    renewalStage: "新签",
    alertLevel: "normal",
    remark: ""
  });
  const [customerForm, setCustomerForm] = useState({
    name: "新交付客户",
    contact: "客户联系人",
    phone: "13800018888",
    creditLimit: "1000000",
    paymentTerm: "30"
  });
  const [projectForm, setProjectForm] = useState({
    customerId: "1",
    name: "客户交付项目",
    address: "客户现场地址",
    contact: "",
    phone: ""
  });
  const [productForm, setProductForm] = useState({
    name: "C30 商品混凝土",
    spec: "C30",
    line: "concrete",
    unit: "m3",
    basePrice: "480",
    costPrice: "360"
  });
  const [materialForm, setMaterialForm] = useState({
    name: "水泥 P.O42.5",
    spec: "P.O42.5",
    unit: "t",
    safeStock: "80"
  });
  const [siteForm, setSiteForm] = useState({
    companyId: "1",
    name: "客户交付站",
    code: "SITE-NEW",
    address: "客户现场地址"
  });
  const [inventoryForm, setInventoryForm] = useState({
    siteId: "1",
    materialId: "1",
    warehouse: "主仓",
    silo: "SILO-1",
    quantity: "100"
  });
  const [driverForm, setDriverForm] = useState({
    name: "新司机",
    phone: "13900018888",
    licenseNo: "D-NEW",
    licenseExpire: "2027-12-31"
  });
  const [carrierForm, setCarrierForm] = useState({
    name: "客户现场承运商",
    contact: "调度联系人",
    phone: "13900019999",
    settleMode: "monthly"
  });
  const [vehicleForm, setVehicleForm] = useState({
    plateNo: "粤BNEW01",
    vehicleType: "mixer",
    capacity: "12m3",
    carrier: "客户现场承运商",
    siteId: "1",
    driverId: "1",
    certExpiresAt: "2027-12-31"
  });
  const [alertForm, setAlertForm] = useState({
    instanceId: "0",
    customerName: "",
    severity: "warning",
    source: "server",
    title: "服务端心跳异常",
    message: "客户实例心跳超过阈值未上报"
  });
  const [telemetryForm, setTelemetryForm] = useState({
    instanceId: "0",
    probeToken: "",
    source: "trace",
    component: "server",
    severity: "critical",
    eventType: "http_500",
    traceId: "trace-demo-001",
    endpoint: "/api/product-ops/overview",
    durationMs: "4200",
    statusCode: "500",
    errorMessage: "客户现场接口返回 500",
    message: "客户现场链路追踪异常"
  });
  const [monitoringIntegrationForm, setMonitoringIntegrationForm] = useState({
    name: "客户现场 Prometheus",
    code: "prometheus-site",
    provider: "prometheus",
    endpoint: "/api/product-ops/monitoring/report",
    token: "mon-prometheus-site",
    status: "active",
    remark: "Alertmanager webhook"
  });
  const [alertRuleForm, setAlertRuleForm] = useState({
    name: "服务端 CPU 过高",
    source: "prometheus",
    component: "server",
    metric: "cpu_percent",
    operator: ">=",
    threshold: "85",
    severity: "critical",
    status: "active",
    notifyChannels: "sse webhook",
    remark: "外部监控规则"
  });
  const [alertPolicyForm, setAlertPolicyForm] = useState({
    id: "0",
    policyNo: "",
    name: "严重告警聚合升级",
    source: "all",
    component: "all",
    metric: "all",
    severity: "critical",
    aggregateWindowMinutes: "30",
    suppressMinutes: "10",
    escalateAfterMinutes: "15",
    escalateTo: "on_call_manager",
    notifyChannels: "sse webhook",
    status: "active",
    remark: "严重告警聚合降噪，超时未处理自动升级"
  });
  const [alertChannelForm, setAlertChannelForm] = useState({
    id: "0",
    name: "值班 Webhook",
    code: "webhook",
    type: "webhook",
    endpoint: "mock://success",
    token: "",
    secret: "",
    status: "active",
    retryLimit: "3",
    timeoutSeconds: "3",
    remark: "告警升级通知通道"
  });
  const [monitoringEventForm, setMonitoringEventForm] = useState({
    integrationCode: "prometheus-site",
    integrationToken: "mon-prometheus-site",
    instanceId: "0",
    provider: "prometheus",
    source: "prometheus",
    component: "server",
    metric: "cpu_percent",
    value: "96",
    severity: "warning",
    status: "firing",
    title: "CPU 使用率过高",
    message: "第三方监控触发 CPU 使用率过高"
  });
  const [renewalForm, setRenewalForm] = useState({
    id: "0",
    instanceId: "0",
    customerName: "新交付客户",
    licenseId: "LIC-NEW-2026",
    stage: "待跟进",
    status: "open",
    owner: "客户成功",
    amount: "98000",
    currency: "CNY",
    dueDate: "2026-12-31",
    nextFollowAt: "2026-06-21 10:00:00",
    riskLevel: "warning",
    lastContactAt: "",
    remark: ""
  });
  const [licenseForm, setLicenseForm] = useState({
    customerName: "新交付客户",
    watermark: "CBMP-NEW",
    expiresAt: "2027-12-31",
    edition: "Operations",
    modules: "dashboard license update report",
    maxSites: "5",
    maxVehicles: "30",
    issuer: "CBMP License Center",
    privateKey: ""
  });
  const [updateForm, setUpdateForm] = useState({
    version: "1.0.2",
    component: "client",
    channel: "stable",
    status: "available",
    packageType: "full",
    baseVersion: "1.0.1",
    deltaAlgorithm: "cbmp-copy-v1",
    baseArtifactSha256: "",
    targetArtifactSha256: "",
    checksum: "sha256:client-102",
    signature: "sig:client-102",
    fileName: "cbmp-client-1.0.2.json",
    sizeBytes: "2048",
    artifactContentType: "application/json",
    artifactText: JSON.stringify({ component: "client", version: "1.0.2", release: "客户端运营台修复包" }, null, 2),
    rollbackVersion: "1.0.1",
    remark: "客户端运营台修复包"
  });
  const [commercialForm, setCommercialForm] = useState({
    taskId: "0",
    amount: "66000",
    modules: "license update support",
    newExpiresAt: "2027-12-31",
    paymentAmount: "66000",
    method: "bank",
    remark: "续费商务闭环"
  });
  const [renewalWorkflowForm, setRenewalWorkflowForm] = useState({
    approvalType: "contract",
    currentRole: "boss",
    comment: "续费金额、模块和合同条款确认通过",
    signer: "客户授权代表",
    phone: "13800019999",
    channel: "local_esign",
    invoiceType: "blue_e_invoice",
    taxRate: "0.06",
    invoiceRemark: "续费回款电子发票"
  });
  const [renewalIntegrationForm, setRenewalIntegrationForm] = useState({
    id: "0",
    name: "税控电子发票网关",
    code: "tax_gateway",
    provider: "tax",
    scenario: "tax",
    endpoint: "mock://success",
    token: "",
    secret: "",
    status: "active",
    retryLimit: "3",
    timeoutSeconds: "3",
    remark: "续费发票税控平台"
  });
  const [rolloutForm, setRolloutForm] = useState({
    updateId: "0",
    strategy: "gray",
    targetInstanceIds: [] as string[],
    remark: "客户现场灰度更新批次"
  });
  const [gatewayForm, setGatewayForm] = useState<GatewayRouteForm>({
    id: "0",
    name: "产品运营 API",
    pathPrefix: "/api",
    stableUpstream: "http://cbmp-backend:8088",
    canaryUpstream: "http://cbmp-backend-v2:8088",
    canaryPercent: "0",
    readTimeoutSec: "60",
    status: "active"
  });

  async function load() {
    const [nextOverview, nextBackups, nextBackupDrills, nextGateway, nextMasterData] = await Promise.all([
      api.productOpsOverview(),
      api.listBackups(),
      api.listBackupDrills(),
      api.gatewayOverview(),
      api.bootstrap()
    ]);
    setOverview(nextOverview);
    setBackups(Array.isArray(nextBackups) ? nextBackups : []);
    setBackupDrills(Array.isArray(nextBackupDrills) ? nextBackupDrills : []);
    setGateway(nextGateway);
    setMasterData(nextMasterData);
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : "产品运营数据加载失败"));
  }, []);

  const openAlerts = useMemo(
    () => (overview?.alerts || []).filter((item) => item.status !== "handled" && item.status !== "closed"),
    [overview]
  );
  const openRenewals = useMemo(
    () => (overview?.renewalTasks || []).filter((item) => item.status !== "closed" && item.status !== "cancelled"),
    [overview]
  );
  const clientUpdates = useMemo(
    () => (overview?.updates || []).filter((item) => item.component === "client" || item.component === "all"),
    [overview]
  );
  const serverUpdates = useMemo(
    () => (overview?.updates || []).filter((item) => item.component === "server" || item.component === "all" || !item.component),
    [overview]
  );
  const rolloutUpdates = useMemo(
    () => (overview?.updates || []).filter((item) => item.status !== "draft" && item.status !== "rolled_back"),
    [overview]
  );

  function latestQuote(taskId: number) {
    return (overview?.renewalQuotes || []).find((item) => item.taskId === taskId);
  }

  function latestContract(taskId: number) {
    return (overview?.renewalContracts || []).find((item) => item.taskId === taskId);
  }

  function paidAmount(contractId: number) {
    return (overview?.renewalPayments || [])
      .filter((item) => item.contractId === contractId && item.status === "paid")
      .reduce((sum, item) => sum + (item.amount || 0), 0);
  }

  function latestPayment(contractId: number) {
    return (overview?.renewalPayments || []).find((item) => item.contractId === contractId && item.status === "paid");
  }

  function latestApproval(taskId: number) {
    return (overview?.renewalApprovals || []).find((item) => item.taskId === taskId);
  }

  function latestInvoice(taskId: number) {
    return (overview?.renewalInvoices || []).find((item) => item.taskId === taskId);
  }

  function latestESign(taskId: number) {
    return (overview?.renewalESigns || []).find((item) => item.taskId === taskId);
  }

  async function saveCustomer() {
    setError("");
    try {
      const created = await api.createCustomer({
        companyId: 1,
        name: customerForm.name,
        contact: customerForm.contact,
        phone: customerForm.phone,
        creditLimit: Number(customerForm.creditLimit) || 0,
        paymentTerm: Number(customerForm.paymentTerm) || 0,
        status: "active"
      });
      setProjectForm((value) => ({ ...value, customerId: String(created.id), contact: created.contact, phone: created.phone }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存客户失败");
    }
  }

  async function saveProject() {
    setError("");
    try {
      await api.createProject({
        customerId: Number(projectForm.customerId) || 0,
        name: projectForm.name,
        address: projectForm.address,
        contact: projectForm.contact,
        phone: projectForm.phone,
        status: "active"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存项目失败");
    }
  }

  async function saveProduct() {
    setError("");
    try {
      await api.createProduct({
        name: productForm.name,
        spec: productForm.spec,
        line: productForm.line,
        unit: productForm.unit,
        basePrice: Number(productForm.basePrice) || 0,
        costPrice: Number(productForm.costPrice) || 0,
        status: "active"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存产品失败");
    }
  }

  async function saveMaterial() {
    setError("");
    try {
      const created = await api.createMaterial({
        name: materialForm.name,
        spec: materialForm.spec,
        unit: materialForm.unit,
        safeStock: Number(materialForm.safeStock) || 0,
        status: "active"
      });
      setInventoryForm((value) => ({ ...value, materialId: String(created.id) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存物料失败");
    }
  }

  async function saveSite() {
    setError("");
    try {
      const created = await api.createSite({
        companyId: Number(siteForm.companyId) || 1,
        name: siteForm.name,
        code: siteForm.code,
        address: siteForm.address,
        status: "running"
      });
      setInventoryForm((value) => ({ ...value, siteId: String(created.id) }));
      setVehicleForm((value) => ({ ...value, siteId: String(created.id) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存站点失败");
    }
  }

  async function saveInventoryItem() {
    setError("");
    try {
      await api.createInventoryItem({
        siteId: Number(inventoryForm.siteId) || 0,
        materialId: Number(inventoryForm.materialId) || 0,
        warehouse: inventoryForm.warehouse,
        silo: inventoryForm.silo,
        quantity: Number(inventoryForm.quantity) || 0
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存库存失败");
    }
  }

  async function saveDriver() {
    setError("");
    try {
      const created = await api.createDriver({
        name: driverForm.name,
        phone: driverForm.phone,
        licenseNo: driverForm.licenseNo,
        licenseExpire: driverForm.licenseExpire,
        status: "active"
      });
      setVehicleForm((value) => ({ ...value, driverId: String(created.id) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存司机失败");
    }
  }

  async function saveCarrier() {
    setError("");
    try {
      const created = await api.createCarrier({
        name: carrierForm.name,
        contact: carrierForm.contact,
        phone: carrierForm.phone,
        settleMode: carrierForm.settleMode,
        status: "active"
      });
      setVehicleForm((value) => ({ ...value, carrier: created.name }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存承运商失败");
    }
  }

  async function saveVehicle() {
    setError("");
    try {
      await api.createVehicle({
        plateNo: vehicleForm.plateNo,
        vehicleType: vehicleForm.vehicleType,
        capacity: vehicleForm.capacity,
        carrier: vehicleForm.carrier,
        siteId: Number(vehicleForm.siteId) || 0,
        driverId: Number(vehicleForm.driverId) || 0,
        certExpiresAt: vehicleForm.certExpiresAt,
        status: "active"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存车辆失败");
    }
  }

  function editInstance(instance: ProductInstance) {
    setInstanceForm({
      id: String(instance.id || 0),
      customerName: instance.customerName,
      licenseId: instance.licenseId,
      watermark: instance.watermark,
      edition: instance.edition,
      deploymentMode: instance.deploymentMode || "private",
      clientVersion: instance.clientVersion,
      serverVersion: instance.serverVersion,
      endpoint: instance.endpoint,
      status: instance.status || "online",
      probeToken: instance.probeToken || "",
      probeEnabled: instance.probeEnabled ?? true,
      healthStatus: instance.healthStatus || instance.status || "healthy",
      licenseExpiresAt: instance.licenseExpiresAt,
      renewalOwner: instance.renewalOwner,
      renewalStage: instance.renewalStage,
      alertLevel: instance.alertLevel || "normal",
      remark: instance.remark || ""
    });
    setAlertForm((value) => ({ ...value, instanceId: String(instance.id), customerName: instance.customerName }));
    setTelemetryForm((value) => ({ ...value, instanceId: String(instance.id), probeToken: instance.probeToken || value.probeToken }));
    setLicenseForm((value) => ({
      ...value,
      customerName: instance.customerName,
      watermark: instance.watermark,
      expiresAt: instance.licenseExpiresAt || value.expiresAt,
      edition: instance.edition || value.edition
    }));
    setRenewalForm((value) => ({
      ...value,
      instanceId: String(instance.id),
      customerName: instance.customerName,
      licenseId: instance.licenseId,
      owner: instance.renewalOwner || value.owner,
      stage: instance.renewalStage || value.stage,
      dueDate: instance.licenseExpiresAt || value.dueDate,
      riskLevel: instance.licenseRisk || value.riskLevel
    }));
  }

  function editRenewal(task: ProductRenewalTask) {
    setRenewalForm({
      id: String(task.id || 0),
      instanceId: String(task.instanceId || 0),
      customerName: task.customerName,
      licenseId: task.licenseId,
      stage: task.stage,
      status: task.status || "open",
      owner: task.owner,
      amount: String(task.amount || 0),
      currency: task.currency || "CNY",
      dueDate: task.dueDate,
      nextFollowAt: task.nextFollowAt,
      riskLevel: task.riskLevel || "warning",
      lastContactAt: task.lastContactAt,
      remark: task.remark || ""
    });
  }

  function editRenewalIntegration(item: ProductRenewalIntegration) {
    setRenewalIntegrationForm({
      id: String(item.id || 0),
      name: item.name,
      code: item.code,
      provider: item.provider || item.code,
      scenario: item.scenario || "all",
      endpoint: item.endpoint || "mock://success",
      token: item.token || "",
      secret: item.secret || "",
      status: item.status || "active",
      retryLimit: String(item.retryLimit || 3),
      timeoutSeconds: String(item.timeoutSeconds || 3),
      remark: item.remark || ""
    });
  }

  async function saveInstance() {
    setError("");
    try {
      await api.saveProductInstance({
        id: Number(instanceForm.id) || 0,
        customerName: instanceForm.customerName,
        licenseId: instanceForm.licenseId,
        watermark: instanceForm.watermark,
        edition: instanceForm.edition,
        deploymentMode: instanceForm.deploymentMode,
        clientVersion: instanceForm.clientVersion,
        serverVersion: instanceForm.serverVersion,
        endpoint: instanceForm.endpoint,
        status: instanceForm.status,
        probeToken: instanceForm.probeToken,
        probeEnabled: instanceForm.probeEnabled,
        healthStatus: instanceForm.healthStatus,
        licenseExpiresAt: instanceForm.licenseExpiresAt,
        renewalOwner: instanceForm.renewalOwner,
        renewalStage: instanceForm.renewalStage,
        alertLevel: instanceForm.alertLevel,
        remark: instanceForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "客户实例保存失败");
    }
  }

  async function createAlert() {
    setError("");
    try {
      await api.createSystemAlert({
        instanceId: Number(alertForm.instanceId) || 0,
        customerName: alertForm.customerName,
        severity: alertForm.severity,
        source: alertForm.source,
        title: alertForm.title,
        message: alertForm.message
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "告警创建失败");
    }
  }

  async function handleAlert(alert: SystemAlert) {
    setError("");
    try {
      await api.handleSystemAlert(alert.id, "运营台确认处理");
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "告警处理失败");
    }
  }

  async function reportTelemetryEvent() {
    setError("");
    try {
      const instance = (overview?.instances || []).find((item) => String(item.id) === telemetryForm.instanceId);
      await api.reportTelemetryEvent({
        probeToken: telemetryForm.probeToken || instance?.probeToken || "",
        watermark: instance?.watermark,
        source: telemetryForm.source,
        component: telemetryForm.component,
        severity: telemetryForm.severity,
        eventType: telemetryForm.eventType,
        traceId: telemetryForm.traceId,
        endpoint: telemetryForm.endpoint,
        durationMs: Number(telemetryForm.durationMs) || 0,
        statusCode: Number(telemetryForm.statusCode) || 0,
        errorMessage: telemetryForm.errorMessage,
        message: telemetryForm.message
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "日志/APM/链路事件上报失败");
    }
  }

  function editMonitoringIntegration(item: ProductMonitoringIntegration) {
    setMonitoringIntegrationForm({
      name: item.name,
      code: item.code,
      provider: item.provider || "prometheus",
      endpoint: item.endpoint || "/api/product-ops/monitoring/report",
      token: item.token || "",
      status: item.status || "active",
      remark: item.remark || ""
    });
    setMonitoringEventForm((value) => ({
      ...value,
      integrationCode: item.code,
      integrationToken: item.token,
      provider: item.provider || value.provider,
      source: item.provider || value.source
    }));
  }

  function editAlertRule(item: ProductAlertRule) {
    setAlertRuleForm({
      name: item.name,
      source: item.source || "prometheus",
      component: item.component || "server",
      metric: item.metric,
      operator: item.operator || ">=",
      threshold: String(item.threshold || 0),
      severity: item.severity || "warning",
      status: item.status || "active",
      notifyChannels: (item.notifyChannels || []).join(" "),
      remark: item.remark || ""
    });
  }

  function editAlertPolicy(item: ProductAlertPolicy) {
    setAlertPolicyForm({
      id: String(item.id || 0),
      policyNo: item.policyNo || "",
      name: item.name,
      source: item.source || "all",
      component: item.component || "all",
      metric: item.metric || "all",
      severity: item.severity || "warning",
      aggregateWindowMinutes: String(item.aggregateWindowMinutes || 30),
      suppressMinutes: String(item.suppressMinutes || 0),
      escalateAfterMinutes: String(item.escalateAfterMinutes || 0),
      escalateTo: item.escalateTo || "",
      notifyChannels: (item.notifyChannels || []).join(" "),
      status: item.status || "active",
      remark: item.remark || ""
    });
  }

  function editAlertChannel(item: ProductAlertChannel) {
    setAlertChannelForm({
      id: String(item.id || 0),
      name: item.name,
      code: item.code,
      type: item.type || "webhook",
      endpoint: item.endpoint || "",
      token: item.token || "",
      secret: item.secret || "",
      status: item.status || "active",
      retryLimit: String(item.retryLimit || 3),
      timeoutSeconds: String(item.timeoutSeconds || 3),
      remark: item.remark || ""
    });
  }

  async function saveMonitoringIntegration() {
    setError("");
    try {
      const saved = await api.saveMonitoringIntegration({
        name: monitoringIntegrationForm.name,
        code: monitoringIntegrationForm.code,
        provider: monitoringIntegrationForm.provider,
        endpoint: monitoringIntegrationForm.endpoint,
        token: monitoringIntegrationForm.token,
        status: monitoringIntegrationForm.status,
        remark: monitoringIntegrationForm.remark
      });
      setMonitoringEventForm((value) => ({
        ...value,
        integrationCode: saved.code,
        integrationToken: saved.token,
        provider: saved.provider || value.provider,
        source: saved.provider || value.source
      }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "监控接入保存失败");
    }
  }

  async function saveAlertRule() {
    setError("");
    try {
      await api.saveProductAlertRule({
        name: alertRuleForm.name,
        source: alertRuleForm.source,
        component: alertRuleForm.component,
        metric: alertRuleForm.metric,
        operator: alertRuleForm.operator,
        threshold: Number(alertRuleForm.threshold) || 0,
        severity: alertRuleForm.severity,
        status: alertRuleForm.status,
        notifyChannels: alertRuleForm.notifyChannels.split(/\s+/).filter(Boolean),
        remark: alertRuleForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "告警规则保存失败");
    }
  }

  async function saveAlertPolicy() {
    setError("");
    try {
      await api.saveProductAlertPolicy({
        id: Number(alertPolicyForm.id) || 0,
        policyNo: alertPolicyForm.policyNo,
        name: alertPolicyForm.name,
        source: alertPolicyForm.source,
        component: alertPolicyForm.component,
        metric: alertPolicyForm.metric,
        severity: alertPolicyForm.severity,
        aggregateWindowMinutes: Number(alertPolicyForm.aggregateWindowMinutes) || 0,
        suppressMinutes: Number(alertPolicyForm.suppressMinutes) || 0,
        escalateAfterMinutes: Number(alertPolicyForm.escalateAfterMinutes) || 0,
        escalateTo: alertPolicyForm.escalateTo,
        notifyChannels: alertPolicyForm.notifyChannels.split(/\s+/).filter(Boolean),
        status: alertPolicyForm.status,
        remark: alertPolicyForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "告警策略保存失败");
    }
  }

  async function saveAlertChannel() {
    setError("");
    try {
      await api.saveProductAlertChannel({
        id: Number(alertChannelForm.id) || 0,
        name: alertChannelForm.name,
        code: alertChannelForm.code,
        type: alertChannelForm.type,
        endpoint: alertChannelForm.endpoint,
        token: alertChannelForm.token,
        secret: alertChannelForm.secret,
        status: alertChannelForm.status,
        retryLimit: Number(alertChannelForm.retryLimit) || 3,
        timeoutSeconds: Number(alertChannelForm.timeoutSeconds) || 3,
        remark: alertChannelForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "通知通道保存失败");
    }
  }

  async function retryAlertNotification(item: ProductAlertNotification) {
    setError("");
    try {
      await api.retryAlertNotification(item.id);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "通知重试失败");
    }
  }

  async function saveRenewalIntegration() {
    setError("");
    try {
      await api.saveRenewalIntegration({
        id: Number(renewalIntegrationForm.id) || 0,
        name: renewalIntegrationForm.name,
        code: renewalIntegrationForm.code,
        provider: renewalIntegrationForm.provider,
        scenario: renewalIntegrationForm.scenario,
        endpoint: renewalIntegrationForm.endpoint,
        token: renewalIntegrationForm.token,
        secret: renewalIntegrationForm.secret,
        status: renewalIntegrationForm.status,
        retryLimit: Number(renewalIntegrationForm.retryLimit) || 3,
        timeoutSeconds: Number(renewalIntegrationForm.timeoutSeconds) || 3,
        remark: renewalIntegrationForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费外部集成保存失败");
    }
  }

  async function retryRenewalSyncRecord(item: ProductRenewalSyncRecord) {
    setError("");
    try {
      await api.retryRenewalSyncRecord(item.id);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费同步重试失败");
    }
  }

  async function escalateAlert(alert: SystemAlert) {
    setError("");
    try {
      await api.escalateSystemAlert(alert.id, alertPolicyForm.escalateTo || "manual_on_call", "运营台人工升级");
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "告警升级失败");
    }
  }

  async function reportMonitoringEvent() {
    setError("");
    try {
      const integration = (overview?.monitoringIntegrations || []).find((item) => item.code === monitoringEventForm.integrationCode) || overview?.monitoringIntegrations?.[0];
      const instance = (overview?.instances || []).find((item) => String(item.id) === monitoringEventForm.instanceId) || overview?.instances?.[0];
      await api.reportMonitoringEvent({
        integrationCode: monitoringEventForm.integrationCode || integration?.code,
        integrationToken: monitoringEventForm.integrationToken || integration?.token || "",
        provider: monitoringEventForm.provider || integration?.provider,
        instanceId: Number(monitoringEventForm.instanceId) || instance?.id,
        watermark: instance?.watermark,
        source: monitoringEventForm.source,
        component: monitoringEventForm.component,
        metric: monitoringEventForm.metric,
        value: Number(monitoringEventForm.value) || 0,
        severity: monitoringEventForm.severity,
        status: monitoringEventForm.status,
        title: monitoringEventForm.title,
        message: monitoringEventForm.message,
        labels: {
          watermark: instance?.watermark || "",
          job: monitoringEventForm.component
        }
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "第三方监控事件上报失败");
    }
  }

  async function saveRenewalTask() {
    setError("");
    try {
      await api.saveRenewalTask({
        id: Number(renewalForm.id) || 0,
        instanceId: Number(renewalForm.instanceId) || 0,
        customerName: renewalForm.customerName,
        licenseId: renewalForm.licenseId,
        stage: renewalForm.stage,
        status: renewalForm.status,
        owner: renewalForm.owner,
        amount: Number(renewalForm.amount) || 0,
        currency: renewalForm.currency,
        dueDate: renewalForm.dueDate,
        nextFollowAt: renewalForm.nextFollowAt,
        riskLevel: renewalForm.riskLevel,
        lastContactAt: renewalForm.lastContactAt,
        remark: renewalForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费任务保存失败");
    }
  }

  async function closeRenewalTask(task: ProductRenewalTask) {
    setError("");
    try {
      await api.closeRenewalTask(task.id, "运营台确认续费完成");
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费任务关闭失败");
    }
  }

  async function createRenewalQuote(task: ProductRenewalTask) {
    setError("");
    try {
      await api.createRenewalQuote(task.id, {
        amount: Number(commercialForm.amount) || task.amount,
        currency: task.currency || "CNY",
        modules: commercialForm.modules.split(/\s+/).filter(Boolean),
        newExpiresAt: commercialForm.newExpiresAt,
        remark: commercialForm.remark
      });
      setCommercialForm((value) => ({ ...value, taskId: String(task.id), paymentAmount: String(Number(value.amount) || task.amount || 0) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费报价生成失败");
    }
  }

  async function createRenewalContract(task: ProductRenewalTask) {
    setError("");
    try {
      const quote = latestQuote(task.id);
      await api.createRenewalContract(task.id, {
        quoteId: quote?.id,
        amount: quote?.amount || Number(commercialForm.amount) || task.amount,
        currency: quote?.currency || task.currency || "CNY",
        signedBy: "客户成功确认",
        remark: commercialForm.remark
      });
      setCommercialForm((value) => ({ ...value, taskId: String(task.id), paymentAmount: String(quote?.amount || Number(value.amount) || task.amount || 0) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费合同确认失败");
    }
  }

  async function createRenewalPayment(task: ProductRenewalTask) {
    setError("");
    try {
      const contract = latestContract(task.id);
      if (!contract) {
        setError("请先确认续费合同");
        return;
      }
      const remaining = Math.max((contract.amount || 0) - paidAmount(contract.id), 0);
      await api.createRenewalPayment(task.id, {
        contractId: contract.id,
        amount: Number(commercialForm.paymentAmount) || remaining,
        currency: contract.currency || task.currency || "CNY",
        method: commercialForm.method,
        remark: commercialForm.remark
      });
      setCommercialForm((value) => ({ ...value, taskId: String(task.id) }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费回款登记失败");
    }
  }

  async function submitRenewalApproval(task: ProductRenewalTask) {
    setError("");
    try {
      const quote = latestQuote(task.id);
      const contract = latestContract(task.id);
      await api.submitRenewalApproval(task.id, {
        action: "submit",
        quoteId: quote?.id,
        contractId: contract?.id,
        approvalType: renewalWorkflowForm.approvalType,
        amount: contract?.amount || quote?.amount || task.amount,
        currency: contract?.currency || quote?.currency || task.currency || "CNY",
        currentRole: renewalWorkflowForm.currentRole,
        comment: renewalWorkflowForm.comment
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费审批提交失败");
    }
  }

  async function approveRenewal(task: ProductRenewalTask, action: "approve" | "reject") {
    setError("");
    try {
      const approval = latestApproval(task.id);
      if (!approval || approval.status !== "pending") {
        setError("请先提交待处理的续费审批");
        return;
      }
      await api.submitRenewalApproval(task.id, {
        action,
        approvalId: approval.id,
        comment: renewalWorkflowForm.comment
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费审批处理失败");
    }
  }

  async function sendRenewalESign(task: ProductRenewalTask) {
    setError("");
    try {
      const contract = latestContract(task.id);
      if (!contract) {
        setError("请先确认续费合同");
        return;
      }
      await api.changeRenewalESign(task.id, {
        action: "send",
        contractId: contract.id,
        signer: renewalWorkflowForm.signer,
        phone: renewalWorkflowForm.phone,
        channel: renewalWorkflowForm.channel,
        remark: renewalWorkflowForm.comment
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费电子签发送失败");
    }
  }

  async function completeRenewalESign(task: ProductRenewalTask) {
    setError("");
    try {
      const sign = latestESign(task.id);
      if (!sign || sign.status !== "sent") {
        setError("请先发送待签署的续费电子签");
        return;
      }
      await api.changeRenewalESign(task.id, {
        action: "complete",
        signId: sign.id,
        contractId: sign.contractId,
        signature: `${sign.signer || renewalWorkflowForm.signer} 电子签名`,
        remark: "客户已完成电子签"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费电子签完成失败");
    }
  }

  async function createRenewalInvoice(task: ProductRenewalTask) {
    setError("");
    try {
      const contract = latestContract(task.id);
      if (!contract) {
        setError("请先确认续费合同");
        return;
      }
      const payment = latestPayment(contract.id);
      await api.createRenewalInvoice(task.id, {
        contractId: contract.id,
        paymentId: payment?.id,
        amount: payment?.amount || paidAmount(contract.id) || contract.amount,
        taxRate: Number(renewalWorkflowForm.taxRate) || 0.06,
        invoiceType: renewalWorkflowForm.invoiceType,
        remark: renewalWorkflowForm.invoiceRemark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "续费开票失败");
    }
  }

  async function issueLicense() {
    setError("");
    try {
      await api.issueLicense({
        customerName: licenseForm.customerName,
        watermark: licenseForm.watermark,
        expiresAt: licenseForm.expiresAt,
        edition: licenseForm.edition,
        modules: licenseForm.modules.split(/\s+/).filter(Boolean),
        maxSites: Number(licenseForm.maxSites) || 0,
        maxVehicles: Number(licenseForm.maxVehicles) || 0,
        issuer: licenseForm.issuer,
        privateKey: licenseForm.privateKey
      });
      setLicenseForm((value) => ({ ...value, privateKey: "" }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "授权签发失败");
    }
  }

  async function renewLicense(instance: ProductInstance) {
    if (!instance.latestPackageId) return;
    setError("");
    try {
      await api.renewLicensePackage(instance.latestPackageId, {
        expiresAt: licenseForm.expiresAt,
        edition: licenseForm.edition || instance.edition,
        modules: licenseForm.modules.split(/\s+/).filter(Boolean),
        maxSites: Number(licenseForm.maxSites) || 0,
        maxVehicles: Number(licenseForm.maxVehicles) || 0,
        issuer: licenseForm.issuer,
        privateKey: licenseForm.privateKey
      });
      setLicenseForm((value) => ({ ...value, privateKey: "" }));
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "授权续费失败");
    }
  }

  async function downloadUpdate(update: UpdatePackage) {
    setError("");
    try {
      const download = await api.downloadUpdate(update.id);
      const blob = download.artifactContentBase64
        ? new Blob([base64ToBytes(download.artifactContentBase64)], { type: download.artifactContentType || "application/octet-stream" })
        : new Blob([JSON.stringify(download, null, 2)], { type: download.contentType || "application/json" });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = download.artifactFileName || download.fileName || `cbmp-${update.component || "update"}-${update.version}.json`;
      anchor.click();
      URL.revokeObjectURL(url);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新包下载失败");
    }
  }

  async function applyUpdate(update: UpdatePackage) {
    setError("");
    try {
      await api.applyUpdate(update.id);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新包状态更新失败");
    }
  }

  async function publishUpdate() {
    setError("");
    try {
      const artifactText = updateForm.artifactText.trim();
      const artifactBytes = textEncoder.encode(artifactText);
      const artifactPayload = artifactText
        ? {
          artifactContentBase64: bytesToBase64(artifactBytes),
          artifactContentType: updateForm.artifactContentType,
          artifactFileName: updateForm.fileName,
          artifactSizeBytes: artifactBytes.length,
          checksum: await sha256Checksum(artifactText),
          sizeBytes: artifactBytes.length
        }
        : {
          checksum: updateForm.checksum,
          sizeBytes: Number(updateForm.sizeBytes) || 0
        };
      await api.publishUpdate({
        version: updateForm.version,
        component: updateForm.component,
        channel: updateForm.channel,
        status: updateForm.status,
        packageType: updateForm.packageType,
        baseVersion: updateForm.packageType === "delta" ? updateForm.baseVersion : undefined,
        deltaAlgorithm: updateForm.packageType === "delta" ? updateForm.deltaAlgorithm : undefined,
        checksum: artifactPayload.checksum,
        signature: updateForm.signature,
        fileName: updateForm.fileName,
        sizeBytes: artifactPayload.sizeBytes,
        artifactContentBase64: "artifactContentBase64" in artifactPayload ? artifactPayload.artifactContentBase64 : undefined,
        artifactContentType: "artifactContentType" in artifactPayload ? artifactPayload.artifactContentType : undefined,
        artifactFileName: "artifactFileName" in artifactPayload ? artifactPayload.artifactFileName : undefined,
        artifactSizeBytes: "artifactSizeBytes" in artifactPayload ? artifactPayload.artifactSizeBytes : undefined,
        baseArtifactSha256: updateForm.packageType === "delta" ? updateForm.baseArtifactSha256 : undefined,
        targetArtifactSha256: updateForm.packageType === "delta" ? updateForm.targetArtifactSha256 : undefined,
        rollbackVersion: updateForm.rollbackVersion,
        remark: updateForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新包发布失败");
    }
  }

  async function createRollout() {
    setError("");
    try {
      const updateId = Number(rolloutForm.updateId) || rolloutUpdates[0]?.id || 0;
      if (!updateId) {
        setError("请先发布可用于灰度的更新包");
        return;
      }
      await api.createUpdateRollout({
        updateId,
        strategy: rolloutForm.strategy,
        targetInstanceIds: rolloutForm.targetInstanceIds.map((id) => Number(id)).filter(Boolean),
        remark: rolloutForm.remark
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新灰度批次创建失败");
    }
  }

  async function advanceRollout(rollout: ProductUpdateRollout, action: "apply" | "fail" | "rollback") {
    setError("");
    try {
      const target = rollout.items.find((item) => action === "rollback" ? item.status === "applied" : item.status === "pending" || item.status === "running");
      await api.advanceUpdateRollout(rollout.id, {
        action,
        instanceId: target?.instanceId,
        message: action === "apply" ? "运营台确认应用成功" : action === "fail" ? "运营台标记推进失败" : "运营台确认回滚"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新灰度批次推进失败");
    }
  }

  function rolloutTarget(rollout: ProductUpdateRollout, action: "apply" | "rollback") {
    return rollout.items.find((item) => action === "rollback" ? item.status === "applied" : item.status === "pending" || item.status === "running");
  }

  async function executeRollout(rollout: ProductUpdateRollout, action: "apply" | "rollback", dryRun = false) {
    setError("");
    try {
      const target = rolloutTarget(rollout, action);
      if (!target) {
        setError(action === "rollback" ? "当前批次没有可回滚客户实例" : "当前批次没有待执行客户实例");
        return;
      }
      await api.executeUpdateRollout(rollout.id, {
        action,
        instanceId: target.instanceId,
        dryRun,
        remark: dryRun ? "运营台执行前预检" : action === "apply" ? "运营台受控执行成功" : "运营台受控回滚"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新执行器运行失败");
    }
  }

  async function enqueueSystemUpdateTask(rollout: ProductUpdateRollout, action: "apply" | "rollback") {
    setError("");
    try {
      const target = rolloutTarget(rollout, action);
      if (!target) {
        setError(action === "rollback" ? "当前批次没有可回滚客户实例" : "当前批次没有待下发客户实例");
        return;
      }
      await api.createSystemUpdateTask(rollout.id, {
        action,
        instanceId: target.instanceId,
        remark: action === "rollback" ? "运营台下发端内更新器回滚" : "运营台下发端内更新器更新"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "端内更新器任务下发失败");
    }
  }

  async function confirmSystemUpdateTaskSuccess(task: ProductSystemUpdateTask) {
    setError("");
    try {
      const instance = (overview?.instances || []).find((item) => item.id === task.instanceId);
      if (!instance?.probeToken) {
        setError("客户实例没有配置端内更新器/probe token，无法手动确认回执");
        return;
      }
      if (task.status === "queued") {
        await api.pollSystemUpdateTasks({ updaterToken: instance.probeToken, watermark: task.watermark });
      }
      await api.reportSystemUpdateTask(task.taskNo, {
        updaterToken: instance.probeToken,
        status: task.action === "rollback" ? "rolled_back" : "succeeded",
        progress: 100,
        step: "health_check",
        message: task.action === "rollback" ? "端内更新器已回滚并通过健康检查" : "端内更新器已完成更新并通过健康检查",
        currentVersion: task.action === "rollback" ? task.fromVersion : task.version,
        updaterVersion: "updater-console-manual-confirm"
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "端内更新器手动确认回执失败");
    }
  }

  async function createSystemBackup() {
    setError("");
    setBackupBusy("create");
    try {
      await api.createBackup();
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建加密备份失败");
    } finally {
      setBackupBusy("");
    }
  }

  async function runSystemBackupDrill() {
    setError("");
    setBackupBusy("drill");
    try {
      await api.runBackupDrill();
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "恢复演练失败");
    } finally {
      setBackupBusy("");
    }
  }

  async function restoreSystemBackup(backup: BackupInfo) {
    setError("");
    if (!window.confirm(`确认恢复备份 ${backup.name}？当前平台数据会被该备份覆盖。`)) {
      return;
    }
    setBackupBusy(`restore:${backup.name}`);
    try {
      await api.restoreBackup(backup.name);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "恢复备份失败");
    } finally {
      setBackupBusy("");
    }
  }

  async function saveGatewayRoute() {
    setError("");
    setGatewayBusy("save");
    try {
      await api.saveGatewayRoute({
        id: Number(gatewayForm.id) || 0,
        name: gatewayForm.name,
        pathPrefix: gatewayForm.pathPrefix,
        stableUpstream: gatewayForm.stableUpstream,
        canaryUpstream: gatewayForm.canaryUpstream,
        canaryPercent: Number(gatewayForm.canaryPercent) || 0,
        readTimeoutSec: Number(gatewayForm.readTimeoutSec) || 60,
        status: gatewayForm.status
      });
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存网关路由失败");
    } finally {
      setGatewayBusy("");
    }
  }

  function editGatewayRoute(route: GatewayRoute) {
    setGatewayForm({
      id: String(route.id || 0),
      name: route.name || "",
      pathPrefix: route.pathPrefix || "/api",
      stableUpstream: route.stableUpstream || "",
      canaryUpstream: route.canaryUpstream || "",
      canaryPercent: String(route.canaryPercent || 0),
      readTimeoutSec: String(route.readTimeoutSec || 60),
      status: route.status || "active"
    });
  }

  async function updateGatewayCanary(route: GatewayRoute, percent: number) {
    setError("");
    setGatewayBusy(`canary:${route.id}`);
    try {
      await api.setGatewayCanary(route.id, percent, route.canaryUpstream || gatewayForm.canaryUpstream);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新网关灰度失败");
    } finally {
      setGatewayBusy("");
    }
  }

  async function toggleGatewayDrain(route: GatewayRoute) {
    setError("");
    setGatewayBusy(`drain:${route.id}`);
    try {
      await api.setGatewayDrain(route.id, !route.drainEnabled, 300000);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新网关排空失败");
    } finally {
      setGatewayBusy("");
    }
  }

  async function toggleGatewayStatus(route: GatewayRoute) {
    setError("");
    setGatewayBusy(`status:${route.id}`);
    try {
      await api.setGatewayStatus(route.id, route.status === "active" ? "disabled" : "active");
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新网关状态失败");
    } finally {
      setGatewayBusy("");
    }
  }

  async function reloadGatewayConfig() {
    setError("");
    setGatewayBusy("reload");
    try {
      await api.reloadGateway();
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "记录网关 reload 失败");
    } finally {
      setGatewayBusy("");
    }
  }

  return (
    <div className="view-stack">
      {error ? <div className="panel error-text">{error}</div> : null}
      <section className="panel">
        <div className="between">
          <h3>产品运营总览</h3>
          <StatusChip value={overview?.kpis.criticalAlerts ? "critical" : overview?.kpis.openAlerts ? "warning" : "normal"} />
        </div>
        <div className="metric-list">
          <div><span>客户实例</span><b>{overview?.kpis.customers || 0}</b></div>
          <div><span>在线实例</span><b>{overview?.kpis.onlineInstances || 0}</b></div>
          <div><span>异常实例</span><b>{overview?.kpis.degradedInstances || 0}</b></div>
          <div><span>待续费</span><b>{overview?.kpis.expiringLicenses || 0}</b></div>
          <div><span>开放告警</span><b>{overview?.kpis.openAlerts || 0}</b></div>
          <div><span>严重告警</span><b>{overview?.kpis.criticalAlerts || 0}</b></div>
          <div><span>续费任务</span><b>{overview?.kpis.openRenewals || 0}</b></div>
          <div><span>高危续费</span><b>{overview?.kpis.overdueRenewals || 0}</b></div>
          <div><span>待确认报价</span><b>{overview?.kpis.pendingRenewalQuotes || 0}</b></div>
          <div><span>待回款合同</span><b>{overview?.kpis.pendingRenewalContracts || 0}</b></div>
          <div><span>已回款</span><b>{Math.round(overview?.kpis.paidRenewalAmount || 0).toLocaleString()}</b></div>
          <div><span>待审批</span><b>{overview?.kpis.pendingRenewalApprovals || 0}</b></div>
          <div><span>待电签</span><b>{overview?.kpis.pendingRenewalESigns || 0}</b></div>
          <div><span>已开票</span><b>{overview?.kpis.issuedRenewalInvoices || 0}</b></div>
          <div><span>续费集成</span><b>{overview?.kpis.activeRenewalIntegrations || 0}</b></div>
          <div><span>同步流水</span><b>{overview?.kpis.renewalSyncRecords || 0}</b></div>
          <div><span>同步失败</span><b>{overview?.kpis.failedRenewalSyncRecords || 0}</b></div>
          <div><span>待同步</span><b>{overview?.kpis.pendingRenewalSyncRecords || 0}</b></div>
          <div><span>探针上报</span><b>{overview?.kpis.probeReports || 0}</b></div>
          <div><span>异常探针</span><b>{overview?.kpis.unhealthyProbes || 0}</b></div>
          <div><span>可观测事件</span><b>{overview?.kpis.telemetryEvents || 0}</b></div>
          <div><span>严重事件</span><b>{overview?.kpis.criticalTelemetryEvents || 0}</b></div>
          <div><span>监控接入</span><b>{overview?.kpis.monitoringIntegrations || 0}</b></div>
          <div><span>告警规则</span><b>{overview?.kpis.activeAlertRules || 0}</b></div>
          <div><span>监控事件</span><b>{overview?.kpis.monitoringEvents || 0}</b></div>
          <div><span>外部告警</span><b>{overview?.kpis.monitoringAlerts || 0}</b></div>
          <div><span>告警策略</span><b>{overview?.kpis.activeAlertPolicies || 0}</b></div>
          <div><span>抑制告警</span><b>{overview?.kpis.suppressedAlerts || 0}</b></div>
          <div><span>升级告警</span><b>{overview?.kpis.escalatedAlerts || 0}</b></div>
          <div><span>通知通道</span><b>{overview?.kpis.activeAlertChannels || 0}</b></div>
          <div><span>通知审计</span><b>{overview?.kpis.alertNotifications || 0}</b></div>
          <div><span>通知失败</span><b>{overview?.kpis.failedAlertNotifications || 0}</b></div>
          <div><span>待投递</span><b>{overview?.kpis.pendingAlertNotifications || 0}</b></div>
          <div><span>活跃批次</span><b>{overview?.kpis.activeRollouts || 0}</b></div>
          <div><span>失败目标</span><b>{overview?.kpis.failedRolloutItems || 0}</b></div>
          <div><span>更新执行</span><b>{overview?.kpis.updateExecutions || 0}</b></div>
          <div><span>执行失败</span><b>{overview?.kpis.failedUpdateExecutions || 0}</b></div>
          <div><span>现场任务</span><b>{overview?.kpis.systemUpdateTasks || 0}</b></div>
          <div><span>端内执行中</span><b>{overview?.kpis.runningSystemUpdateTasks || 0}</b></div>
          <div><span>端内失败</span><b>{overview?.kpis.failedSystemUpdateTasks || 0}</b></div>
          <div><span>客户端包</span><b>{overview?.kpis.clientUpdatePackages || 0}</b></div>
          <div><span>服务端包</span><b>{overview?.kpis.serverUpdatePackages || 0}</b></div>
          <div><span>可发布更新</span><b>{overview?.kpis.availableUpdates || 0}</b></div>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>交付初始化基础资料</h3>
          <StatusChip value={(masterData?.customers || []).length ? "active" : "warning"} />
        </div>
        <div className="metric-list">
          <div><span>客户</span><b>{masterData?.customers.length || 0}</b></div>
          <div><span>项目</span><b>{masterData?.projects.length || 0}</b></div>
          <div><span>产品</span><b>{masterData?.products.length || 0}</b></div>
          <div><span>物料</span><b>{masterData?.materials.length || 0}</b></div>
          <div><span>站点</span><b>{masterData?.sites.length || 0}</b></div>
          <div><span>库存</span><b>{masterData?.inventory.length || 0}</b></div>
          <div><span>司机</span><b>{masterData?.drivers.length || 0}</b></div>
          <div><span>车辆</span><b>{masterData?.vehicles.length || 0}</b></div>
        </div>

        <div className="split-grid">
          <div>
            <div className="between compact-row">
              <h4>客户与项目</h4>
              <div className="row-actions">
                <button className="soft-button" onClick={saveCustomer}>保存客户</button>
                <button className="soft-button" onClick={saveProject}>保存项目</button>
              </div>
            </div>
            <div className="form-grid two">
              <label><span>客户名称</span><input value={customerForm.name} onChange={(event) => setCustomerForm({ ...customerForm, name: event.target.value })} /></label>
              <label><span>联系人</span><input value={customerForm.contact} onChange={(event) => setCustomerForm({ ...customerForm, contact: event.target.value })} /></label>
              <label><span>电话</span><input value={customerForm.phone} onChange={(event) => setCustomerForm({ ...customerForm, phone: event.target.value })} /></label>
              <label><span>信用额度</span><input value={customerForm.creditLimit} onChange={(event) => setCustomerForm({ ...customerForm, creditLimit: event.target.value })} /></label>
              <label><span>项目客户</span><select value={projectForm.customerId} onChange={(event) => setProjectForm({ ...projectForm, customerId: event.target.value })}>{(masterData?.customers || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>项目名称</span><input value={projectForm.name} onChange={(event) => setProjectForm({ ...projectForm, name: event.target.value })} /></label>
              <label><span>项目地址</span><input value={projectForm.address} onChange={(event) => setProjectForm({ ...projectForm, address: event.target.value })} /></label>
              <label><span>项目联系人</span><input value={projectForm.contact} onChange={(event) => setProjectForm({ ...projectForm, contact: event.target.value })} /></label>
            </div>
            <table>
              <thead><tr><th>客户</th><th>项目</th><th>状态</th></tr></thead>
              <tbody>
                {(masterData?.customers || []).slice(-4).map((customer) => {
                  const projects = (masterData?.projects || []).filter((project) => project.customerId === customer.id).map((project) => project.name).join(" / ");
                  return <tr key={customer.id}><td>{customer.name}<span className="muted block-text">{customer.contact || "-"} · {customer.phone || "-"}</span></td><td>{projects || "-"}</td><td><StatusChip value={customer.status} /></td></tr>;
                })}
              </tbody>
            </table>
          </div>

          <div>
            <div className="between compact-row">
              <h4>产品物料库存</h4>
              <div className="row-actions">
                <button className="soft-button" onClick={saveProduct}>保存产品</button>
                <button className="soft-button" onClick={saveMaterial}>保存物料</button>
                <button className="soft-button" onClick={saveInventoryItem}>入库</button>
              </div>
            </div>
            <div className="form-grid two">
              <label><span>产品名称</span><input value={productForm.name} onChange={(event) => setProductForm({ ...productForm, name: event.target.value })} /></label>
              <label><span>产品规格</span><input value={productForm.spec} onChange={(event) => setProductForm({ ...productForm, spec: event.target.value })} /></label>
              <label><span>销售单价</span><input value={productForm.basePrice} onChange={(event) => setProductForm({ ...productForm, basePrice: event.target.value })} /></label>
              <label><span>成本单价</span><input value={productForm.costPrice} onChange={(event) => setProductForm({ ...productForm, costPrice: event.target.value })} /></label>
              <label><span>物料名称</span><input value={materialForm.name} onChange={(event) => setMaterialForm({ ...materialForm, name: event.target.value })} /></label>
              <label><span>安全库存</span><input value={materialForm.safeStock} onChange={(event) => setMaterialForm({ ...materialForm, safeStock: event.target.value })} /></label>
              <label><span>库存站点</span><select value={inventoryForm.siteId} onChange={(event) => setInventoryForm({ ...inventoryForm, siteId: event.target.value })}>{(masterData?.sites || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>库存物料</span><select value={inventoryForm.materialId} onChange={(event) => setInventoryForm({ ...inventoryForm, materialId: event.target.value })}>{(masterData?.materials || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>仓库</span><input value={inventoryForm.warehouse} onChange={(event) => setInventoryForm({ ...inventoryForm, warehouse: event.target.value })} /></label>
              <label><span>数量</span><input value={inventoryForm.quantity} onChange={(event) => setInventoryForm({ ...inventoryForm, quantity: event.target.value })} /></label>
            </div>
            <table>
              <thead><tr><th>产品</th><th>物料</th><th>库存</th></tr></thead>
              <tbody>
                {(masterData?.products || []).slice(-4).map((product, index) => {
                  const material = (masterData?.materials || []).slice(-4)[index];
                  const inventory = material ? (masterData?.inventory || []).filter((item) => item.materialId === material.id).reduce((sum, item) => sum + (item.quantity || 0), 0) : 0;
                  return <tr key={product.id}><td>{product.name}<span className="muted block-text">{product.spec || "-"} · {product.unit}</span></td><td>{material?.name || "-"}<span className="muted block-text">{material?.safeStock ? `安全 ${material.safeStock}` : "-"}</span></td><td>{inventory.toLocaleString()} {material?.unit || ""}</td></tr>;
                })}
              </tbody>
            </table>
          </div>
        </div>

        <div className="split-grid">
          <div>
            <div className="between compact-row">
              <h4>站点</h4>
              <button className="soft-button" onClick={saveSite}>保存站点</button>
            </div>
            <div className="form-grid two">
              <label><span>站点名称</span><input value={siteForm.name} onChange={(event) => setSiteForm({ ...siteForm, name: event.target.value })} /></label>
              <label><span>站点编码</span><input value={siteForm.code} onChange={(event) => setSiteForm({ ...siteForm, code: event.target.value })} /></label>
              <label><span>公司</span><select value={siteForm.companyId} onChange={(event) => setSiteForm({ ...siteForm, companyId: event.target.value })}>{(masterData?.companies || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>地址</span><input value={siteForm.address} onChange={(event) => setSiteForm({ ...siteForm, address: event.target.value })} /></label>
            </div>
            <table>
              <thead><tr><th>站点</th><th>地址</th><th>状态</th></tr></thead>
              <tbody>
                {(masterData?.sites || []).slice(-4).map((site) => <tr key={site.id}><td>{site.name}<span className="muted block-text">{site.code}</span></td><td>{site.address || "-"}</td><td><StatusChip value={site.status} /></td></tr>)}
              </tbody>
            </table>
          </div>

          <div>
            <div className="between compact-row">
              <h4>车队</h4>
              <div className="row-actions">
                <button className="soft-button" onClick={saveDriver}>保存司机</button>
                <button className="soft-button" onClick={saveCarrier}>保存承运商</button>
                <button className="soft-button" onClick={saveVehicle}>保存车辆</button>
              </div>
            </div>
            <div className="form-grid two">
              <label><span>司机姓名</span><input value={driverForm.name} onChange={(event) => setDriverForm({ ...driverForm, name: event.target.value })} /></label>
              <label><span>司机电话</span><input value={driverForm.phone} onChange={(event) => setDriverForm({ ...driverForm, phone: event.target.value })} /></label>
              <label><span>承运商</span><input value={carrierForm.name} onChange={(event) => setCarrierForm({ ...carrierForm, name: event.target.value })} /></label>
              <label><span>承运联系人</span><input value={carrierForm.contact} onChange={(event) => setCarrierForm({ ...carrierForm, contact: event.target.value })} /></label>
              <label><span>车牌号</span><input value={vehicleForm.plateNo} onChange={(event) => setVehicleForm({ ...vehicleForm, plateNo: event.target.value })} /></label>
              <label><span>车辆站点</span><select value={vehicleForm.siteId} onChange={(event) => setVehicleForm({ ...vehicleForm, siteId: event.target.value })}>{(masterData?.sites || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>绑定司机</span><select value={vehicleForm.driverId} onChange={(event) => setVehicleForm({ ...vehicleForm, driverId: event.target.value })}>{(masterData?.drivers || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <label><span>容量</span><input value={vehicleForm.capacity} onChange={(event) => setVehicleForm({ ...vehicleForm, capacity: event.target.value })} /></label>
            </div>
            <table>
              <thead><tr><th>车辆</th><th>司机</th><th>承运商</th></tr></thead>
              <tbody>
                {(masterData?.vehicles || []).slice(-4).map((vehicle) => {
                  const driver = (masterData?.drivers || []).find((item) => item.id === vehicle.driverId);
                  return <tr key={vehicle.id}><td>{vehicle.plateNo}<span className="muted block-text">{vehicle.vehicleType || "-"} · {vehicle.capacity || "-"}</span></td><td>{driver?.name || "-"}</td><td>{vehicle.carrier || "-"}</td></tr>;
                })}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>客户实例</h3>
          <button className="soft-button" onClick={saveInstance}>保存实例</button>
        </div>
        <table>
          <thead><tr><th>客户</th><th>授权</th><th>版本</th><th>探针</th><th>续费</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.instances || []).map((instance) => (
              <tr key={instance.id}>
                <td>
                  <b>{instance.customerName}</b>
                  <span className="muted block-text">{instance.endpoint || instance.watermark}</span>
                </td>
                <td>
                  {instance.licenseId || "-"}
                  <span className="muted block-text">{instance.licenseExpiresAt || "-"} · {instance.daysToExpire} 天</span>
                </td>
                <td>
                  客户端 {instance.clientVersion || "-"}
                  <span className="muted block-text">服务端 {instance.serverVersion || "-"}</span>
                </td>
                <td>
                  {instance.lastProbeAt || instance.lastHeartbeatAt || "-"}
                  <span className="muted block-text">{instance.probeEnabled ? "enabled" : "disabled"} · {instance.healthStatus || "-"}</span>
                </td>
                <td>
                  {instance.renewalStage || "-"}
                  <span className="muted block-text">{instance.renewalOwner || "-"}</span>
                </td>
                <td>
                  <StatusChip value={instance.alertLevel || instance.status} />
                  <span className="muted block-text">{instance.status}</span>
                </td>
                <td className="row-actions">
                  <button className="soft-button" onClick={() => editInstance(instance)}>编辑</button>
                  <button className="soft-button" disabled={!instance.latestPackageId} onClick={() => renewLicense(instance)}>续费</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className="form-grid two">
          <label><span>客户名称</span><input value={instanceForm.customerName} onChange={(event) => setInstanceForm({ ...instanceForm, customerName: event.target.value })} /></label>
          <label><span>License ID</span><input value={instanceForm.licenseId} onChange={(event) => setInstanceForm({ ...instanceForm, licenseId: event.target.value })} /></label>
          <label><span>水印</span><input value={instanceForm.watermark} onChange={(event) => setInstanceForm({ ...instanceForm, watermark: event.target.value })} /></label>
          <label><span>版本</span><input value={`${instanceForm.clientVersion}/${instanceForm.serverVersion}`} onChange={(event) => {
            const [clientVersion = "", serverVersion = ""] = event.target.value.split("/");
            setInstanceForm({ ...instanceForm, clientVersion, serverVersion });
          }} /></label>
          <label><span>入口地址</span><input value={instanceForm.endpoint} onChange={(event) => setInstanceForm({ ...instanceForm, endpoint: event.target.value })} /></label>
          <label><span>授权到期</span><input value={instanceForm.licenseExpiresAt} onChange={(event) => setInstanceForm({ ...instanceForm, licenseExpiresAt: event.target.value })} /></label>
          <label><span>续费负责人</span><input value={instanceForm.renewalOwner} onChange={(event) => setInstanceForm({ ...instanceForm, renewalOwner: event.target.value })} /></label>
          <label><span>续费阶段</span><input value={instanceForm.renewalStage} onChange={(event) => setInstanceForm({ ...instanceForm, renewalStage: event.target.value })} /></label>
          <label><span>探针 Token</span><input value={instanceForm.probeToken} onChange={(event) => setInstanceForm({ ...instanceForm, probeToken: event.target.value })} /></label>
          <label><span>探针启用</span><select value={instanceForm.probeEnabled ? "enabled" : "disabled"} onChange={(event) => setInstanceForm({ ...instanceForm, probeEnabled: event.target.value === "enabled" })}><option value="enabled">enabled</option><option value="disabled">disabled</option></select></label>
        </div>
      </section>

      <section className="panel">
        <h3>客户现场探针上报</h3>
        <table>
          <thead><tr><th>报告</th><th>客户</th><th>组件</th><th>状态</th><th>资源</th><th>队列/错误</th><th>接收时间</th></tr></thead>
          <tbody>
            {(overview?.probeReports || []).map((report) => (
              <tr key={report.id}>
                <td>
                  <b>{report.reportNo}</b>
                  <span className="muted block-text">{report.message || report.watermark}</span>
                </td>
                <td>{report.customerName}</td>
                <td>
                  {report.component}
                  <span className="muted block-text">C {report.clientVersion || "-"} / S {report.serverVersion || "-"}</span>
                </td>
                <td><StatusChip value={report.alertRaised ? "critical" : report.status} /></td>
                <td>{Math.round(report.cpuPercent)}% / {Math.round(report.memoryPercent)}% / {Math.round(report.diskPercent)}%</td>
                <td>{report.queueBacklog} / {report.errorCount}</td>
                <td>{report.receivedAt || report.reportedAt}</td>
              </tr>
            ))}
            {!(overview?.probeReports || []).length ? <tr><td colSpan={7}>暂无探针上报</td></tr> : null}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <div className="between">
          <h3>日志 / APM / 链路事件</h3>
          <button className="soft-button" onClick={reportTelemetryEvent}>测试上报</button>
        </div>
        <table>
          <thead><tr><th>事件</th><th>客户</th><th>来源</th><th>链路</th><th>指标</th><th>状态</th><th>接收时间</th></tr></thead>
          <tbody>
            {(overview?.telemetryEvents || []).map((event) => (
              <tr key={event.id}>
                <td>
                  <b>{event.eventNo}</b>
                  <span className="muted block-text">{event.message || event.errorMessage || event.eventType}</span>
                </td>
                <td>{event.customerName}</td>
                <td>{event.source} · {event.component}</td>
                <td>
                  {event.traceId || "-"}
                  <span className="muted block-text">{event.endpoint || "-"}</span>
                </td>
                <td>{event.durationMs || 0}ms / {event.statusCode || "-"}</td>
                <td><StatusChip value={event.alertRaised ? "critical" : event.severity} /></td>
                <td>{event.receivedAt || event.occurredAt}</td>
              </tr>
            ))}
            {!(overview?.telemetryEvents || []).length ? <tr><td colSpan={7}>暂无日志/APM/链路事件</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label>
            <span>客户实例</span>
            <select value={telemetryForm.instanceId} onChange={(event) => {
              const instance = (overview?.instances || []).find((item) => String(item.id) === event.target.value);
              setTelemetryForm({ ...telemetryForm, instanceId: event.target.value, probeToken: instance?.probeToken || telemetryForm.probeToken });
            }}>
              <option value="0">选择实例</option>
              {(overview?.instances || []).map((instance) => <option key={instance.id} value={instance.id}>{instance.customerName}</option>)}
            </select>
          </label>
          <label><span>Token</span><input value={telemetryForm.probeToken} onChange={(event) => setTelemetryForm({ ...telemetryForm, probeToken: event.target.value })} /></label>
          <label><span>来源</span><select value={telemetryForm.source} onChange={(event) => setTelemetryForm({ ...telemetryForm, source: event.target.value })}><option value="log">log</option><option value="apm">apm</option><option value="trace">trace</option><option value="metric">metric</option></select></label>
          <label><span>组件</span><select value={telemetryForm.component} onChange={(event) => setTelemetryForm({ ...telemetryForm, component: event.target.value })}><option value="client">client</option><option value="server">server</option><option value="gateway">gateway</option></select></label>
          <label><span>级别</span><select value={telemetryForm.severity} onChange={(event) => setTelemetryForm({ ...telemetryForm, severity: event.target.value })}><option value="normal">normal</option><option value="warning">warning</option><option value="critical">critical</option></select></label>
          <label><span>事件类型</span><input value={telemetryForm.eventType} onChange={(event) => setTelemetryForm({ ...telemetryForm, eventType: event.target.value })} /></label>
          <label><span>Trace ID</span><input value={telemetryForm.traceId} onChange={(event) => setTelemetryForm({ ...telemetryForm, traceId: event.target.value })} /></label>
          <label><span>Endpoint</span><input value={telemetryForm.endpoint} onChange={(event) => setTelemetryForm({ ...telemetryForm, endpoint: event.target.value })} /></label>
          <label><span>耗时 ms</span><input value={telemetryForm.durationMs} onChange={(event) => setTelemetryForm({ ...telemetryForm, durationMs: event.target.value })} /></label>
          <label><span>状态码</span><input value={telemetryForm.statusCode} onChange={(event) => setTelemetryForm({ ...telemetryForm, statusCode: event.target.value })} /></label>
          <label><span>错误</span><input value={telemetryForm.errorMessage} onChange={(event) => setTelemetryForm({ ...telemetryForm, errorMessage: event.target.value })} /></label>
          <label><span>消息</span><input value={telemetryForm.message} onChange={(event) => setTelemetryForm({ ...telemetryForm, message: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>监控接入 / 告警规则</h3>
          <div className="row-actions">
            <button className="soft-button" onClick={saveMonitoringIntegration}>保存接入</button>
            <button className="soft-button" onClick={saveAlertRule}>保存规则</button>
            <button className="soft-button" onClick={reportMonitoringEvent}>测试外部事件</button>
          </div>
        </div>
        <div className="grid-12">
          <div className="span-6">
            <h4>监控接入</h4>
            <table>
              <thead><tr><th>接入</th><th>来源</th><th>Token</th><th>状态</th><th>操作</th></tr></thead>
              <tbody>
                {(overview?.monitoringIntegrations || []).map((item) => (
                  <tr key={item.id}>
                    <td>
                      <b>{item.name}</b>
                      <span className="muted block-text">{item.integrationNo} · {item.code}</span>
                    </td>
                    <td>{item.provider}<span className="muted block-text">{item.lastEventAt || item.createdAt}</span></td>
                    <td>{item.token || "-"}</td>
                    <td><StatusChip value={item.status} /></td>
                    <td><button className="soft-button" onClick={() => editMonitoringIntegration(item)}>编辑</button></td>
                  </tr>
                ))}
                {!(overview?.monitoringIntegrations || []).length ? <tr><td colSpan={5}>暂无监控接入</td></tr> : null}
              </tbody>
            </table>
          </div>
          <div className="span-6">
            <h4>告警规则</h4>
            <table>
              <thead><tr><th>规则</th><th>范围</th><th>条件</th><th>级别</th><th>操作</th></tr></thead>
              <tbody>
                {(overview?.alertRules || []).map((item) => (
                  <tr key={item.id}>
                    <td>
                      <b>{item.name}</b>
                      <span className="muted block-text">{item.ruleNo}</span>
                    </td>
                    <td>{item.source} · {item.component}</td>
                    <td>{item.metric} {item.operator} {item.threshold}</td>
                    <td><StatusChip value={item.severity || item.status} /></td>
                    <td><button className="soft-button" onClick={() => editAlertRule(item)}>编辑</button></td>
                  </tr>
                ))}
                {!(overview?.alertRules || []).length ? <tr><td colSpan={5}>暂无告警规则</td></tr> : null}
              </tbody>
            </table>
          </div>
        </div>
        <table>
          <thead><tr><th>事件</th><th>客户</th><th>指标</th><th>规则</th><th>状态</th><th>接收时间</th></tr></thead>
          <tbody>
            {(overview?.monitoringEvents || []).map((event) => (
              <tr key={event.id}>
                <td>
                  <b>{event.eventNo}</b>
                  <span className="muted block-text">{event.title || event.message || event.integrationName}</span>
                </td>
                <td>{event.customerName}</td>
                <td>{event.metric} = {event.value}<span className="muted block-text">{event.source} · {event.component}</span></td>
                <td>{event.matchedRuleNo || "-"}</td>
                <td><StatusChip value={event.alertRaised ? "critical" : event.severity} /></td>
                <td>{event.receivedAt || event.occurredAt}</td>
              </tr>
            ))}
            {!(overview?.monitoringEvents || []).length ? <tr><td colSpan={6}>暂无第三方监控事件</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label><span>接入名称</span><input value={monitoringIntegrationForm.name} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, name: event.target.value })} /></label>
          <label><span>接入编码</span><input value={monitoringIntegrationForm.code} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, code: event.target.value })} /></label>
          <label><span>来源</span><select value={monitoringIntegrationForm.provider} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, provider: event.target.value })}><option value="prometheus">prometheus</option><option value="grafana">grafana</option><option value="sentry">sentry</option><option value="zabbix">zabbix</option><option value="custom">custom</option></select></label>
          <label><span>状态</span><select value={monitoringIntegrationForm.status} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, status: event.target.value })}><option value="active">active</option><option value="disabled">disabled</option></select></label>
          <label><span>Token</span><input value={monitoringIntegrationForm.token} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, token: event.target.value })} /></label>
          <label className="span-two"><span>Endpoint</span><input value={monitoringIntegrationForm.endpoint} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, endpoint: event.target.value })} /></label>
          <label><span>接入备注</span><input value={monitoringIntegrationForm.remark} onChange={(event) => setMonitoringIntegrationForm({ ...monitoringIntegrationForm, remark: event.target.value })} /></label>
          <label><span>规则名称</span><input value={alertRuleForm.name} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, name: event.target.value })} /></label>
          <label><span>规则来源</span><input value={alertRuleForm.source} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, source: event.target.value })} /></label>
          <label><span>组件</span><select value={alertRuleForm.component} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, component: event.target.value })}><option value="server">server</option><option value="client">client</option><option value="gateway">gateway</option><option value="all">all</option></select></label>
          <label><span>指标</span><input value={alertRuleForm.metric} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, metric: event.target.value })} /></label>
          <label><span>条件</span><input value={`${alertRuleForm.operator} ${alertRuleForm.threshold}`} onChange={(event) => {
            const [operator = ">=", threshold = "0"] = event.target.value.split(/\s+/);
            setAlertRuleForm({ ...alertRuleForm, operator, threshold });
          }} /></label>
          <label><span>级别</span><select value={alertRuleForm.severity} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, severity: event.target.value })}><option value="warning">warning</option><option value="critical">critical</option><option value="normal">normal</option></select></label>
          <label><span>规则状态</span><select value={alertRuleForm.status} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, status: event.target.value })}><option value="active">active</option><option value="disabled">disabled</option></select></label>
          <label><span>通知通道</span><input value={alertRuleForm.notifyChannels} onChange={(event) => setAlertRuleForm({ ...alertRuleForm, notifyChannels: event.target.value })} /></label>
          <label><span>测试客户</span><select value={monitoringEventForm.instanceId} onChange={(event) => setMonitoringEventForm({ ...monitoringEventForm, instanceId: event.target.value })}><option value="0">选择实例</option>{(overview?.instances || []).map((instance) => <option key={instance.id} value={instance.id}>{instance.customerName}</option>)}</select></label>
          <label><span>测试指标</span><input value={`${monitoringEventForm.metric}/${monitoringEventForm.value}`} onChange={(event) => {
            const [metric = "", value = "0"] = event.target.value.split("/");
            setMonitoringEventForm({ ...monitoringEventForm, metric, value });
          }} /></label>
          <label><span>测试状态</span><select value={monitoringEventForm.status} onChange={(event) => setMonitoringEventForm({ ...monitoringEventForm, status: event.target.value })}><option value="firing">firing</option><option value="resolved">resolved</option><option value="triggered">triggered</option></select></label>
          <label><span>测试消息</span><input value={monitoringEventForm.message} onChange={(event) => setMonitoringEventForm({ ...monitoringEventForm, message: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>告警聚合 / 抑制 / 升级</h3>
          <div className="row-actions">
            <button className="soft-button" onClick={saveAlertChannel}>保存通道</button>
            <button className="soft-button" onClick={saveAlertPolicy}>保存策略</button>
          </div>
        </div>
        <div className="grid-12">
          <div className="span-4">
            <h4>策略配置</h4>
            <table>
              <thead><tr><th>策略</th><th>匹配</th><th>窗口</th><th>升级</th><th>操作</th></tr></thead>
              <tbody>
                {(overview?.alertPolicies || []).map((item) => (
                  <tr key={item.id}>
                    <td>
                      <b>{item.name}</b>
                      <span className="muted block-text">{item.policyNo} · {item.status}</span>
                    </td>
                    <td>{item.source} · {item.component}<span className="muted block-text">{item.metric} · {item.severity}</span></td>
                    <td>{item.aggregateWindowMinutes}m / {item.suppressMinutes}m</td>
                    <td>{item.escalateAfterMinutes ? `${item.escalateAfterMinutes}m -> ${item.escalateTo || "-"}` : "-"}</td>
                    <td><button className="soft-button" onClick={() => editAlertPolicy(item)}>编辑</button></td>
                  </tr>
                ))}
                {!(overview?.alertPolicies || []).length ? <tr><td colSpan={5}>暂无告警策略</td></tr> : null}
              </tbody>
            </table>
          </div>
          <div className="span-4">
            <h4>通知通道</h4>
            <table>
              <thead><tr><th>通道</th><th>类型</th><th>状态</th><th>最近结果</th><th>操作</th></tr></thead>
              <tbody>
                {(overview?.alertChannels || []).map((item) => (
                  <tr key={item.id}>
                    <td>
                      <b>{item.name}</b>
                      <span className="muted block-text">{item.channelNo} · {item.code}</span>
                    </td>
                    <td>{item.type}<span className="muted block-text">{item.endpoint || "-"}</span></td>
                    <td><StatusChip value={item.status} /></td>
                    <td>{item.lastDeliveredAt || item.lastError || "-"}</td>
                    <td><button className="soft-button" onClick={() => editAlertChannel(item)}>编辑</button></td>
                  </tr>
                ))}
                {!(overview?.alertChannels || []).length ? <tr><td colSpan={5}>暂无通知通道</td></tr> : null}
              </tbody>
            </table>
          </div>
          <div className="span-4">
            <h4>通知审计</h4>
            <table>
              <thead><tr><th>通知</th><th>客户</th><th>动作</th><th>通道</th><th>结果</th><th>操作</th></tr></thead>
              <tbody>
                {(overview?.alertNotifications || []).map((item) => (
                  <tr key={item.id}>
                    <td>
                      <b>{item.notificationNo}</b>
                      <span className="muted block-text">{item.alertNo} · {item.message}</span>
                    </td>
                    <td>{item.customerName}</td>
                    <td><StatusChip value={item.action === "escalated" ? "critical" : item.action === "suppressed" ? "warning" : "normal"} /></td>
                    <td>{item.channel}<span className="muted block-text">{item.target || item.endpoint || "-"}</span></td>
                    <td>
                      <StatusChip value={item.status === "failed" ? "critical" : item.status === "pending" ? "warning" : "normal"} />
                      <span className="muted block-text">{item.error || item.deliveredAt || item.nextRetryAt || item.createdAt}</span>
                    </td>
                    <td><button className="soft-button" disabled={item.status !== "failed"} onClick={() => retryAlertNotification(item)}>重试</button></td>
                  </tr>
                ))}
                {!(overview?.alertNotifications || []).length ? <tr><td colSpan={6}>暂无通知审计</td></tr> : null}
              </tbody>
            </table>
          </div>
        </div>
        <div className="form-grid four">
          <label><span>策略名称</span><input value={alertPolicyForm.name} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, name: event.target.value })} /></label>
          <label><span>来源</span><select value={alertPolicyForm.source} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, source: event.target.value })}><option value="all">all</option><option value="server">server</option><option value="client">client</option><option value="license">license</option><option value="probe">probe</option><option value="telemetry">telemetry</option><option value="monitoring">monitoring</option></select></label>
          <label><span>组件</span><select value={alertPolicyForm.component} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, component: event.target.value })}><option value="all">all</option><option value="server">server</option><option value="client">client</option><option value="gateway">gateway</option><option value="license">license</option></select></label>
          <label><span>指标</span><input value={alertPolicyForm.metric} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, metric: event.target.value })} /></label>
          <label><span>级别</span><select value={alertPolicyForm.severity} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, severity: event.target.value })}><option value="normal">normal</option><option value="warning">warning</option><option value="critical">critical</option></select></label>
          <label><span>聚合窗口分钟</span><input value={alertPolicyForm.aggregateWindowMinutes} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, aggregateWindowMinutes: event.target.value })} /></label>
          <label><span>抑制分钟</span><input value={alertPolicyForm.suppressMinutes} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, suppressMinutes: event.target.value })} /></label>
          <label><span>升级分钟</span><input value={alertPolicyForm.escalateAfterMinutes} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, escalateAfterMinutes: event.target.value })} /></label>
          <label><span>升级对象</span><input value={alertPolicyForm.escalateTo} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, escalateTo: event.target.value })} /></label>
          <label><span>通知通道</span><input value={alertPolicyForm.notifyChannels} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, notifyChannels: event.target.value })} /></label>
          <label><span>策略状态</span><select value={alertPolicyForm.status} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, status: event.target.value })}><option value="active">active</option><option value="disabled">disabled</option></select></label>
          <label><span>备注</span><input value={alertPolicyForm.remark} onChange={(event) => setAlertPolicyForm({ ...alertPolicyForm, remark: event.target.value })} /></label>
          <label><span>通道名称</span><input value={alertChannelForm.name} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, name: event.target.value })} /></label>
          <label><span>通道编码</span><input value={alertChannelForm.code} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, code: event.target.value })} /></label>
          <label><span>通道类型</span><select value={alertChannelForm.type} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, type: event.target.value })}><option value="sse">sse</option><option value="local">local</option><option value="webhook">webhook</option><option value="enterprise_wechat">enterprise_wechat</option><option value="sms">sms</option><option value="itsm">itsm</option></select></label>
          <label><span>通道状态</span><select value={alertChannelForm.status} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, status: event.target.value })}><option value="active">active</option><option value="disabled">disabled</option></select></label>
          <label className="span-two"><span>Endpoint</span><input value={alertChannelForm.endpoint} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, endpoint: event.target.value })} /></label>
          <label><span>Token</span><input value={alertChannelForm.token} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, token: event.target.value })} /></label>
          <label><span>Secret</span><input value={alertChannelForm.secret} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, secret: event.target.value })} /></label>
          <label><span>重试次数</span><input value={alertChannelForm.retryLimit} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, retryLimit: event.target.value })} /></label>
          <label><span>超时秒数</span><input value={alertChannelForm.timeoutSeconds} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, timeoutSeconds: event.target.value })} /></label>
          <label className="span-two"><span>通道备注</span><input value={alertChannelForm.remark} onChange={(event) => setAlertChannelForm({ ...alertChannelForm, remark: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>授权续费任务</h3>
          <button className="soft-button" onClick={saveRenewalTask}>保存续费任务</button>
        </div>
        <table>
          <thead><tr><th>任务</th><th>客户</th><th>金额</th><th>到期</th><th>下次跟进</th><th>风险</th><th>操作</th></tr></thead>
          <tbody>
            {openRenewals.map((task) => (
              <tr key={task.id}>
                <td>
                  <b>{task.taskNo}</b>
                  <span className="muted block-text">{task.stage} · {task.owner}</span>
                </td>
                <td>
                  {task.customerName}
                  <span className="muted block-text">{task.licenseId || "-"}</span>
                </td>
                <td>{task.currency || "CNY"} {Math.round(task.amount || 0).toLocaleString()}</td>
                <td>{task.dueDate || "-"}</td>
                <td>{task.nextFollowAt || "-"}</td>
                <td><StatusChip value={task.riskLevel || task.status} /></td>
                <td className="row-actions">
                  <button className="soft-button" onClick={() => editRenewal(task)}>编辑</button>
                  <button className="soft-button" onClick={() => closeRenewalTask(task)}>关闭</button>
                </td>
              </tr>
            ))}
            {!openRenewals.length ? <tr><td colSpan={7}>暂无开放续费任务</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label>
            <span>客户实例</span>
            <select value={renewalForm.instanceId} onChange={(event) => {
              const instance = (overview?.instances || []).find((item) => String(item.id) === event.target.value);
              setRenewalForm({
                ...renewalForm,
                instanceId: event.target.value,
                customerName: instance?.customerName || renewalForm.customerName,
                licenseId: instance?.licenseId || renewalForm.licenseId,
                dueDate: instance?.licenseExpiresAt || renewalForm.dueDate,
                owner: instance?.renewalOwner || renewalForm.owner,
                stage: instance?.renewalStage || renewalForm.stage,
                riskLevel: instance?.licenseRisk || renewalForm.riskLevel
              });
            }}>
              <option value="0">手工指定</option>
              {(overview?.instances || []).map((instance) => <option key={instance.id} value={instance.id}>{instance.customerName}</option>)}
            </select>
          </label>
          <label><span>客户名称</span><input value={renewalForm.customerName} onChange={(event) => setRenewalForm({ ...renewalForm, customerName: event.target.value })} /></label>
          <label><span>License ID</span><input value={renewalForm.licenseId} onChange={(event) => setRenewalForm({ ...renewalForm, licenseId: event.target.value })} /></label>
          <label><span>阶段</span><input value={renewalForm.stage} onChange={(event) => setRenewalForm({ ...renewalForm, stage: event.target.value })} /></label>
          <label><span>负责人</span><input value={renewalForm.owner} onChange={(event) => setRenewalForm({ ...renewalForm, owner: event.target.value })} /></label>
          <label><span>金额</span><input value={renewalForm.amount} onChange={(event) => setRenewalForm({ ...renewalForm, amount: event.target.value })} /></label>
          <label><span>到期日</span><input value={renewalForm.dueDate} onChange={(event) => setRenewalForm({ ...renewalForm, dueDate: event.target.value })} /></label>
          <label><span>下次跟进</span><input value={renewalForm.nextFollowAt} onChange={(event) => setRenewalForm({ ...renewalForm, nextFollowAt: event.target.value })} /></label>
          <label><span>风险</span><select value={renewalForm.riskLevel} onChange={(event) => setRenewalForm({ ...renewalForm, riskLevel: event.target.value })}><option value="normal">normal</option><option value="warning">warning</option><option value="critical">critical</option></select></label>
          <label className="span-all"><span>备注</span><input value={renewalForm.remark} onChange={(event) => setRenewalForm({ ...renewalForm, remark: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>续费报价 / 合同 / 回款</h3>
          <StatusChip value={(overview?.kpis.pendingRenewalContracts || 0) > 0 ? "warning" : "normal"} />
        </div>
        <table>
          <thead><tr><th>续费任务</th><th>报价</th><th>合同</th><th>回款</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.renewalTasks || []).map((task) => {
              const quote = latestQuote(task.id);
              const contract = latestContract(task.id);
              const paid = contract ? paidAmount(contract.id) : 0;
              return (
                <tr key={task.id}>
                  <td>
                    <b>{task.taskNo}</b>
                    <span className="muted block-text">{task.customerName} · {task.owner || "-"}</span>
                  </td>
                  <td>
                    {quote ? `${quote.currency || "CNY"} ${Math.round(quote.amount || 0).toLocaleString()}` : "-"}
                    <span className="muted block-text">{quote?.quoteNo || "未报价"} · {quote?.status || "-"}</span>
                  </td>
                  <td>
                    {contract ? `${contract.currency || "CNY"} ${Math.round(contract.amount || 0).toLocaleString()}` : "-"}
                    <span className="muted block-text">{contract?.contractNo || "未签合同"} · {contract?.status || "-"}</span>
                  </td>
                  <td>
                    {contract ? `${Math.round(paid).toLocaleString()} / ${Math.round(contract.amount || 0).toLocaleString()}` : "-"}
                    <span className="muted block-text">{(overview?.renewalPayments || []).find((item) => item.contractId === contract?.id)?.paymentNo || "未回款"}</span>
                  </td>
                  <td><StatusChip value={task.status === "closed" ? "normal" : task.riskLevel || task.status} /></td>
                  <td className="row-actions">
                    <button className="soft-button" onClick={() => createRenewalQuote(task)}>生成报价</button>
                    <button className="soft-button" disabled={!quote} onClick={() => createRenewalContract(task)}>确认合同</button>
                    <button className="soft-button" disabled={!contract || contract.status === "paid"} onClick={() => createRenewalPayment(task)}>登记回款</button>
                  </td>
                </tr>
              );
            })}
            {!(overview?.renewalTasks || []).length ? <tr><td colSpan={6}>暂无续费报价、合同和回款数据</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label>
            <span>续费任务</span>
            <select value={commercialForm.taskId} onChange={(event) => {
              const task = (overview?.renewalTasks || []).find((item) => String(item.id) === event.target.value);
              setCommercialForm({
                ...commercialForm,
                taskId: event.target.value,
                amount: String(task?.amount || commercialForm.amount),
                paymentAmount: String(task?.amount || commercialForm.paymentAmount)
              });
            }}>
              <option value="0">使用行内任务</option>
              {(overview?.renewalTasks || []).map((task) => <option key={task.id} value={task.id}>{task.taskNo} · {task.customerName}</option>)}
            </select>
          </label>
          <label><span>报价金额</span><input value={commercialForm.amount} onChange={(event) => setCommercialForm({ ...commercialForm, amount: event.target.value })} /></label>
          <label><span>新到期日</span><input value={commercialForm.newExpiresAt} onChange={(event) => setCommercialForm({ ...commercialForm, newExpiresAt: event.target.value })} /></label>
          <label><span>回款金额</span><input value={commercialForm.paymentAmount} onChange={(event) => setCommercialForm({ ...commercialForm, paymentAmount: event.target.value })} /></label>
          <label><span>模块</span><input value={commercialForm.modules} onChange={(event) => setCommercialForm({ ...commercialForm, modules: event.target.value })} /></label>
          <label><span>回款方式</span><select value={commercialForm.method} onChange={(event) => setCommercialForm({ ...commercialForm, method: event.target.value })}><option value="bank">bank</option><option value="wechat">wechat</option><option value="alipay">alipay</option><option value="offline">offline</option></select></label>
          <label className="span-two"><span>备注</span><input value={commercialForm.remark} onChange={(event) => setCommercialForm({ ...commercialForm, remark: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>续费审批 / 电子签 / 发票</h3>
          <StatusChip value={(overview?.kpis.pendingRenewalApprovals || overview?.kpis.pendingRenewalESigns) ? "warning" : "normal"} />
        </div>
        <table>
          <thead><tr><th>续费任务</th><th>审批</th><th>电子签</th><th>发票</th><th>阶段</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.renewalTasks || []).map((task) => {
              const approval = latestApproval(task.id);
              const sign = latestESign(task.id);
              const invoice = latestInvoice(task.id);
              const contract = latestContract(task.id);
              return (
                <tr key={task.id}>
                  <td>
                    <b>{task.taskNo}</b>
                    <span className="muted block-text">{task.customerName} · {task.currency || "CNY"} {Math.round(task.amount || 0).toLocaleString()}</span>
                  </td>
                  <td>
                    {approval?.approvalNo || "-"}
                    <span className="muted block-text">{approval?.approvalType || "未提交"} · {approval?.status || "-"}</span>
                  </td>
                  <td>
                    {sign?.signNo || "-"}
                    <span className="muted block-text">{sign?.signer || "未发送"} · {sign?.status || "-"}</span>
                  </td>
                  <td>
                    {invoice?.invoiceNo || "-"}
                    <span className="muted block-text">{invoice ? `${invoice.invoiceType} · ${invoice.taxStatus}` : "未开票"}</span>
                  </td>
                  <td><StatusChip value={task.status === "closed" ? "normal" : task.riskLevel || task.status} /></td>
                  <td className="row-actions">
                    <button className="soft-button" onClick={() => submitRenewalApproval(task)}>提交审批</button>
                    <button className="soft-button" disabled={!approval || approval.status !== "pending"} onClick={() => approveRenewal(task, "approve")}>审批通过</button>
                    <button className="soft-button" disabled={!contract} onClick={() => sendRenewalESign(task)}>发送电签</button>
                    <button className="soft-button" disabled={!sign || sign.status !== "sent"} onClick={() => completeRenewalESign(task)}>完成电签</button>
                    <button className="soft-button" disabled={!contract} onClick={() => createRenewalInvoice(task)}>开票</button>
                  </td>
                </tr>
              );
            })}
            {!(overview?.renewalTasks || []).length ? <tr><td colSpan={6}>暂无续费审批、电签和开票数据</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label><span>审批类型</span><select value={renewalWorkflowForm.approvalType} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, approvalType: event.target.value })}><option value="quote">quote</option><option value="contract">contract</option><option value="renewal">renewal</option></select></label>
          <label><span>审批角色</span><select value={renewalWorkflowForm.currentRole} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, currentRole: event.target.value })}><option value="boss">boss</option><option value="finance">finance</option><option value="customer_success">customer_success</option></select></label>
          <label><span>签署人</span><input value={renewalWorkflowForm.signer} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, signer: event.target.value })} /></label>
          <label><span>签署手机号</span><input value={renewalWorkflowForm.phone} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, phone: event.target.value })} /></label>
          <label><span>电签通道</span><select value={renewalWorkflowForm.channel} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, channel: event.target.value })}><option value="local_esign">local_esign</option><option value="docusign">docusign</option><option value="fada">fada</option></select></label>
          <label><span>发票类型</span><select value={renewalWorkflowForm.invoiceType} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, invoiceType: event.target.value })}><option value="blue_e_invoice">blue_e_invoice</option><option value="blue_vat_special">blue_vat_special</option><option value="blue_vat_normal">blue_vat_normal</option></select></label>
          <label><span>税率</span><input value={renewalWorkflowForm.taxRate} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, taxRate: event.target.value })} /></label>
          <label><span>审批意见</span><input value={renewalWorkflowForm.comment} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, comment: event.target.value })} /></label>
          <label className="span-all"><span>开票备注</span><input value={renewalWorkflowForm.invoiceRemark} onChange={(event) => setRenewalWorkflowForm({ ...renewalWorkflowForm, invoiceRemark: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>续费外部集成</h3>
          <StatusChip value={(overview?.kpis.failedRenewalSyncRecords || 0) > 0 ? "critical" : "normal"} />
        </div>
        <table>
          <thead><tr><th>集成</th><th>场景</th><th>Endpoint</th><th>最近同步</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.renewalIntegrations || []).map((item) => (
              <tr key={item.id}>
                <td>
                  <b>{item.name}</b>
                  <span className="muted block-text">{item.integrationNo} · {item.code} · {item.provider}</span>
                </td>
                <td>{item.scenario}</td>
                <td>{item.endpoint || "-"}</td>
                <td>
                  {item.lastSyncAt || "-"}
                  <span className="muted block-text">{item.lastError || "无错误"}</span>
                </td>
                <td><StatusChip value={item.status === "active" ? "normal" : item.status} /></td>
                <td><button className="soft-button" onClick={() => editRenewalIntegration(item)}>编辑</button></td>
              </tr>
            ))}
            {!(overview?.renewalIntegrations || []).length ? <tr><td colSpan={6}>暂无续费外部集成</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label><span>名称</span><input value={renewalIntegrationForm.name} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, name: event.target.value })} /></label>
          <label><span>编码</span><input value={renewalIntegrationForm.code} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, code: event.target.value })} /></label>
          <label><span>提供方</span><input value={renewalIntegrationForm.provider} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, provider: event.target.value })} /></label>
          <label><span>场景</span><select value={renewalIntegrationForm.scenario} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, scenario: event.target.value })}><option value="esign">esign</option><option value="payment">payment</option><option value="finance">finance</option><option value="tax">tax</option><option value="all">all</option></select></label>
          <label className="span-two"><span>Endpoint</span><input value={renewalIntegrationForm.endpoint} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, endpoint: event.target.value })} /></label>
          <label><span>状态</span><select value={renewalIntegrationForm.status} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, status: event.target.value })}><option value="active">active</option><option value="paused">paused</option></select></label>
          <label><span>重试次数</span><input value={renewalIntegrationForm.retryLimit} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, retryLimit: event.target.value })} /></label>
          <label><span>Token</span><input value={renewalIntegrationForm.token} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, token: event.target.value })} /></label>
          <label><span>Secret</span><input value={renewalIntegrationForm.secret} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, secret: event.target.value })} /></label>
          <label><span>超时秒数</span><input value={renewalIntegrationForm.timeoutSeconds} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, timeoutSeconds: event.target.value })} /></label>
          <label className="span-two"><span>备注</span><input value={renewalIntegrationForm.remark} onChange={(event) => setRenewalIntegrationForm({ ...renewalIntegrationForm, remark: event.target.value })} /></label>
          <button className="soft-button" onClick={saveRenewalIntegration}>保存集成</button>
        </div>
        <table>
          <thead><tr><th>同步流水</th><th>资源</th><th>客户</th><th>外部状态</th><th>错误/重试</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.renewalSyncRecords || []).map((item) => (
              <tr key={item.id}>
                <td>
                  <b>{item.syncNo}</b>
                  <span className="muted block-text">{item.integrationCode || "未匹配"} · {item.scenario} · {item.action}</span>
                </td>
                <td>
                  {item.resourceType} · {item.resourceNo || item.resourceId}
                  <span className="muted block-text">尝试 {item.attemptCount || 0} 次</span>
                </td>
                <td>{item.customerName}</td>
                <td>
                  <StatusChip value={item.status === "succeeded" ? "normal" : item.status} />
                  <span className="muted block-text">{item.externalStatus || item.externalRequestId || "-"}</span>
                </td>
                <td>
                  {item.error || "-"}
                  <span className="muted block-text">{item.nextRetryAt || item.completedAt || item.lastAttemptAt}</span>
                </td>
                <td><button className="soft-button" disabled={item.status === "succeeded"} onClick={() => retryRenewalSyncRecord(item)}>重试</button></td>
              </tr>
            ))}
            {!(overview?.renewalSyncRecords || []).length ? <tr><td colSpan={6}>暂无续费同步流水</td></tr> : null}
          </tbody>
        </table>
      </section>

      <div className="grid-12">
        <section className="span-6 panel">
          <div className="between">
            <h3>系统异常报警</h3>
            <button className="soft-button" onClick={createAlert}>新建告警</button>
          </div>
          <table>
            <thead><tr><th>告警</th><th>客户</th><th>来源</th><th>聚合</th><th>抑制/升级</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              {openAlerts.map((alert) => (
                <tr key={alert.id}>
                  <td>
                    <b>{alert.title}</b>
                    <span className="muted block-text">{alert.alertNo} · {alert.message}</span>
                  </td>
                  <td>{alert.customerName}</td>
                  <td>{alert.source}</td>
                  <td>
                    {alert.eventCount || 1} 次
                    <span className="muted block-text">{alert.policyNo || "无策略"} · {alert.lastSeenAt}</span>
                  </td>
                  <td>
                    {alert.suppressedUntil || "-"}
                    <span className="muted block-text">{alert.escalationLevel || "未升级"} {alert.escalatedAt || ""}</span>
                  </td>
                  <td><StatusChip value={alert.severity} /></td>
                  <td className="row-actions">
                    <button className="soft-button" onClick={() => escalateAlert(alert)}>升级</button>
                    <button className="soft-button" onClick={() => handleAlert(alert)}>处理</button>
                  </td>
                </tr>
              ))}
              {!openAlerts.length ? <tr><td colSpan={7}>暂无开放告警</td></tr> : null}
            </tbody>
          </table>
          <div className="form-grid two">
            <label>
              <span>客户实例</span>
              <select value={alertForm.instanceId} onChange={(event) => {
                const instance = (overview?.instances || []).find((item) => String(item.id) === event.target.value);
                setAlertForm({ ...alertForm, instanceId: event.target.value, customerName: instance?.customerName || alertForm.customerName });
              }}>
                <option value="0">手工指定</option>
                {(overview?.instances || []).map((instance) => <option key={instance.id} value={instance.id}>{instance.customerName}</option>)}
              </select>
            </label>
            <label><span>客户名称</span><input value={alertForm.customerName} onChange={(event) => setAlertForm({ ...alertForm, customerName: event.target.value })} /></label>
            <label><span>级别</span><select value={alertForm.severity} onChange={(event) => setAlertForm({ ...alertForm, severity: event.target.value })}><option value="warning">warning</option><option value="critical">critical</option><option value="info">info</option></select></label>
            <label><span>来源</span><select value={alertForm.source} onChange={(event) => setAlertForm({ ...alertForm, source: event.target.value })}><option value="client">client</option><option value="server">server</option><option value="license">license</option><option value="gateway">gateway</option></select></label>
            <label><span>标题</span><input value={alertForm.title} onChange={(event) => setAlertForm({ ...alertForm, title: event.target.value })} /></label>
            <label><span>内容</span><input value={alertForm.message} onChange={(event) => setAlertForm({ ...alertForm, message: event.target.value })} /></label>
          </div>
        </section>

        <section className="span-6 panel">
          <div className="between">
            <h3>授权续费</h3>
            <button className="soft-button" onClick={issueLicense}>签发授权</button>
          </div>
          <table>
            <thead><tr><th>客户</th><th>授权</th><th>到期</th><th>风险</th></tr></thead>
            <tbody>
              {(overview?.licensePortal.customers || []).map((customer) => (
                <tr key={customer.licenseId || customer.customerName}>
                  <td>{customer.customerName}</td>
                  <td>{customer.licenseId}</td>
                  <td>{customer.expiresAt}<span className="muted block-text">{customer.daysToExpire} 天</span></td>
                  <td><StatusChip value={customer.riskLevel} /></td>
                </tr>
              ))}
              {!(overview?.licensePortal.customers || []).length ? <tr><td colSpan={4}>暂无已签发授权包</td></tr> : null}
            </tbody>
          </table>
          <div className="form-grid two">
            <label><span>签发客户</span><input value={licenseForm.customerName} onChange={(event) => setLicenseForm({ ...licenseForm, customerName: event.target.value })} /></label>
            <label><span>水印</span><input value={licenseForm.watermark} onChange={(event) => setLicenseForm({ ...licenseForm, watermark: event.target.value })} /></label>
            <label><span>到期日</span><input value={licenseForm.expiresAt} onChange={(event) => setLicenseForm({ ...licenseForm, expiresAt: event.target.value })} /></label>
            <label><span>版本</span><input value={licenseForm.edition} onChange={(event) => setLicenseForm({ ...licenseForm, edition: event.target.value })} /></label>
            <label><span>模块</span><input value={licenseForm.modules} onChange={(event) => setLicenseForm({ ...licenseForm, modules: event.target.value })} /></label>
            <label><span>站点/车辆</span><input value={`${licenseForm.maxSites}/${licenseForm.maxVehicles}`} onChange={(event) => {
              const [maxSites = "0", maxVehicles = "0"] = event.target.value.split("/");
              setLicenseForm({ ...licenseForm, maxSites, maxVehicles });
            }} /></label>
            <label><span>签发方</span><input value={licenseForm.issuer} onChange={(event) => setLicenseForm({ ...licenseForm, issuer: event.target.value })} /></label>
            <label><span>签发私钥</span><input type="password" value={licenseForm.privateKey} onChange={(event) => setLicenseForm({ ...licenseForm, privateKey: event.target.value })} /></label>
          </div>
        </section>
      </div>

      <section className="panel">
        <div className="between">
          <h3>客户端 / 服务端更新包</h3>
          <button className="soft-button" onClick={publishUpdate}>发布更新包</button>
        </div>
        <div className="form-grid four">
          <label><span>版本</span><input value={updateForm.version} onChange={(event) => setUpdateForm({ ...updateForm, version: event.target.value })} /></label>
          <label><span>组件</span><select value={updateForm.component} onChange={(event) => setUpdateForm({ ...updateForm, component: event.target.value })}><option value="client">client</option><option value="server">server</option><option value="all">all</option></select></label>
          <label><span>通道</span><select value={updateForm.channel} onChange={(event) => setUpdateForm({ ...updateForm, channel: event.target.value })}><option value="stable">stable</option><option value="gray">gray</option><option value="beta">beta</option></select></label>
          <label><span>状态</span><select value={updateForm.status} onChange={(event) => setUpdateForm({ ...updateForm, status: event.target.value })}><option value="available">available</option><option value="gray">gray</option><option value="draft">draft</option></select></label>
          <label><span>包类型</span><select value={updateForm.packageType} onChange={(event) => setUpdateForm({ ...updateForm, packageType: event.target.value })}><option value="full">full</option><option value="delta">delta</option></select></label>
          <label><span>Base 版本</span><input value={updateForm.baseVersion} onChange={(event) => setUpdateForm({ ...updateForm, baseVersion: event.target.value })} /></label>
          <label><span>差分算法</span><input value={updateForm.deltaAlgorithm} onChange={(event) => setUpdateForm({ ...updateForm, deltaAlgorithm: event.target.value })} /></label>
          <label><span>Checksum</span><input value={updateForm.checksum} onChange={(event) => setUpdateForm({ ...updateForm, checksum: event.target.value })} /></label>
          <label><span>Signature</span><input value={updateForm.signature} onChange={(event) => setUpdateForm({ ...updateForm, signature: event.target.value })} /></label>
          <label><span>文件名</span><input value={updateForm.fileName} onChange={(event) => setUpdateForm({ ...updateForm, fileName: event.target.value })} /></label>
          <label><span>内容类型</span><input value={updateForm.artifactContentType} onChange={(event) => setUpdateForm({ ...updateForm, artifactContentType: event.target.value })} /></label>
          <label><span>大小/回滚</span><input value={`${updateForm.sizeBytes}/${updateForm.rollbackVersion}`} onChange={(event) => {
            const [sizeBytes = "0", rollbackVersion = ""] = event.target.value.split("/");
            setUpdateForm({ ...updateForm, sizeBytes, rollbackVersion });
          }} /></label>
          <label className="span-two"><span>Base SHA256</span><input value={updateForm.baseArtifactSha256} onChange={(event) => setUpdateForm({ ...updateForm, baseArtifactSha256: event.target.value })} /></label>
          <label className="span-two"><span>目标 SHA256</span><input value={updateForm.targetArtifactSha256} onChange={(event) => setUpdateForm({ ...updateForm, targetArtifactSha256: event.target.value })} /></label>
          <label className="span-all"><span>发布说明</span><input value={updateForm.remark} onChange={(event) => setUpdateForm({ ...updateForm, remark: event.target.value })} /></label>
          <label className="span-all"><span>Artifact / Patch 内容</span><textarea rows={4} value={updateForm.artifactText} onChange={(event) => setUpdateForm({ ...updateForm, artifactText: event.target.value })} /></label>
        </div>
        <div className="grid-12">
          <div className="span-6">
            <h4>客户端</h4>
            <UpdateTable updates={clientUpdates} onDownload={downloadUpdate} onApply={applyUpdate} />
          </div>
          <div className="span-6">
            <h4>服务端</h4>
            <UpdateTable updates={serverUpdates} onDownload={downloadUpdate} onApply={applyUpdate} />
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>更新灰度批次</h3>
          <button className="soft-button" onClick={createRollout}>创建批次</button>
        </div>
        <table>
          <thead><tr><th>批次</th><th>更新包</th><th>目标</th><th>状态</th><th>明细</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.updateRollouts || []).map((rollout) => (
              <tr key={rollout.id}>
                <td>
                  <b>{rollout.rolloutNo}</b>
                  <span className="muted block-text">{rollout.strategy || "gray"} · {rollout.createdBy || "-"}</span>
                </td>
                <td>
                  {rollout.component} {rollout.version}
                  <span className="muted block-text">{rollout.createdAt}</span>
                </td>
                <td>
                  {rollout.appliedTargets}/{rollout.totalTargets}
                  <span className="muted block-text">失败 {rollout.failedTargets || 0}</span>
                </td>
                <td><StatusChip value={rollout.status} /></td>
                <td>
                  {(rollout.items || []).slice(0, 3).map((item) => (
                    <span className="muted block-text" key={item.id}>
                      {item.customerName}: {item.fromVersion || "-"} → {item.toVersion || "-"} · {item.status}
                    </span>
                  ))}
                </td>
                <td className="row-actions">
                  <button className="soft-button" disabled={!rollout.items.some((item) => item.status === "pending" || item.status === "running")} onClick={() => executeRollout(rollout, "apply", true)}>预检</button>
                  <button className="soft-button" disabled={!rollout.items.some((item) => item.status === "pending" || item.status === "running")} onClick={() => executeRollout(rollout, "apply")}>执行一台</button>
                  <button className="soft-button" disabled={!rollout.items.some((item) => item.status === "pending" || item.status === "running")} onClick={() => enqueueSystemUpdateTask(rollout, "apply")}>端内更新</button>
                  <button className="soft-button" disabled={!rollout.items.some((item) => item.status === "applied")} onClick={() => executeRollout(rollout, "rollback")}>回滚一台</button>
                  <button className="soft-button" disabled={!rollout.items.some((item) => item.status === "applied")} onClick={() => enqueueSystemUpdateTask(rollout, "rollback")}>端内回滚</button>
                  <button className="soft-button" disabled={rollout.status === "completed"} onClick={() => advanceRollout(rollout, "fail")}>标记失败</button>
                </td>
              </tr>
            ))}
            {!(overview?.updateRollouts || []).length ? <tr><td colSpan={6}>暂无更新灰度批次</td></tr> : null}
          </tbody>
        </table>
        <div className="form-grid four">
          <label>
            <span>更新包</span>
            <select value={rolloutForm.updateId} onChange={(event) => setRolloutForm({ ...rolloutForm, updateId: event.target.value })}>
              <option value="0">选择更新包</option>
              {rolloutUpdates.map((update) => (
                <option key={update.id} value={update.id}>{update.component} {update.version} · {update.channel}</option>
              ))}
            </select>
          </label>
          <label><span>策略</span><select value={rolloutForm.strategy} onChange={(event) => setRolloutForm({ ...rolloutForm, strategy: event.target.value })}><option value="gray">gray</option><option value="stable">stable</option><option value="hotfix">hotfix</option></select></label>
          <label className="span-two">
            <span>客户实例</span>
            <select
              multiple
              value={rolloutForm.targetInstanceIds}
              onChange={(event) => setRolloutForm({ ...rolloutForm, targetInstanceIds: Array.from(event.target.selectedOptions).map((option) => option.value) })}
            >
              {(overview?.instances || []).map((instance) => (
                <option key={instance.id} value={instance.id}>{instance.customerName} · C {instance.clientVersion || "-"} / S {instance.serverVersion || "-"}</option>
              ))}
            </select>
          </label>
          <label className="span-all"><span>批次备注</span><input value={rolloutForm.remark} onChange={(event) => setRolloutForm({ ...rolloutForm, remark: event.target.value })} /></label>
        </div>
      </section>

      <section className="panel">
        <div className="between">
          <h3>端内系统更新任务</h3>
          <StatusChip value={(overview?.kpis.failedSystemUpdateTasks || 0) > 0 ? "critical" : (overview?.kpis.runningSystemUpdateTasks || 0) > 0 ? "warning" : "normal"} />
        </div>
        <table>
          <thead><tr><th>任务</th><th>客户</th><th>组件版本</th><th>状态</th><th>更新包</th><th>回执日志</th><th>操作</th></tr></thead>
          <tbody>
            {(overview?.systemUpdateTasks || []).map((task: ProductSystemUpdateTask) => (
              <tr key={task.id}>
                <td>
                  <b>{task.taskNo}</b>
                  <span className="muted block-text">{task.rolloutNo || "-"} · {task.executionNo || "-"}</span>
                </td>
                <td>
                  {task.customerName}
                  <span className="muted block-text">{task.watermark || "-"} · token {task.updaterTokenHint || "-"}</span>
                </td>
                <td>
                  {task.component} {task.fromVersion || "-"} → {task.version || "-"}
                  <span className="muted block-text">{task.action} · {task.createdBy || "-"}</span>
                </td>
                <td>
                  <StatusChip value={task.status} />
                  <div className="progress-meter compact-progress">
                    <progress max={100} value={task.progress || 0} />
                    <span>{task.progress || 0}% · 心跳 {task.lastHeartbeatAt || task.claimedAt || "-"}</span>
                  </div>
                </td>
                <td>
                  <span className="mono-text">{task.artifactFileName || "-"}</span>
                  <span className="muted block-text">{task.downloadUrl || "-"}</span>
                </td>
                <td>
                  {(task.logs || []).slice(-3).map((log) => (
                    <span className="muted block-text" key={log.id}>{log.step || log.status} · {log.message || log.status}</span>
                  ))}
                  {task.result || task.error ? <span className="muted block-text">{task.result || task.error}</span> : null}
                </td>
                <td className="row-actions">
                  <button className="soft-button" disabled={!["queued", "assigned", "running"].includes(task.status)} onClick={() => confirmSystemUpdateTaskSuccess(task)}>手动确认回执</button>
                </td>
              </tr>
            ))}
            {!(overview?.systemUpdateTasks || []).length ? <tr><td colSpan={7}>暂无端内系统更新任务</td></tr> : null}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <div className="between">
          <h3>更新执行器</h3>
          <StatusChip value={(overview?.kpis.failedUpdateExecutions || 0) > 0 ? "critical" : "normal"} />
        </div>
        <table>
          <thead><tr><th>执行单</th><th>客户</th><th>组件版本</th><th>动作</th><th>验签</th><th>步骤</th><th>结果</th></tr></thead>
          <tbody>
            {(overview?.updateExecutions || []).map((execution: ProductUpdateExecution) => (
              <tr key={execution.id}>
                <td>
                  <b>{execution.executionNo}</b>
                  <span className="muted block-text">{execution.rolloutNo || "-"} · {execution.startedBy || "-"}</span>
                </td>
                <td>{execution.customerName}</td>
                <td>
                  {execution.component} {execution.version}
                  <span className="muted block-text">{execution.artifactFileName || "-"}</span>
                </td>
                <td>
                  {execution.action}{execution.dryRun ? " dry-run" : ""}
                  <span className="muted block-text">{execution.startedAt || "-"}</span>
                </td>
                <td><StatusChip value={execution.checksumVerified ? "normal" : "critical"} /></td>
                <td>
                  {(execution.steps || []).slice(0, 3).map((step) => (
                    <span className="muted block-text" key={step.id}>{step.name} · {step.status}</span>
                  ))}
                  {(execution.steps || []).length > 3 ? <span className="muted block-text">共 {execution.steps.length} 步</span> : null}
                </td>
                <td>
                  <StatusChip value={execution.status} />
                  <span className="muted block-text">{execution.result || execution.error || execution.precheckSummary || "-"}</span>
                </td>
              </tr>
            ))}
            {!(overview?.updateExecutions || []).length ? <tr><td colSpan={7}>暂无更新执行记录</td></tr> : null}
          </tbody>
        </table>
      </section>

      <GatewayCenter
        gateway={gateway}
        form={gatewayForm}
        busy={gatewayBusy}
        onFormChange={setGatewayForm}
        onSave={saveGatewayRoute}
        onEdit={editGatewayRoute}
        onCanary={updateGatewayCanary}
        onDrain={toggleGatewayDrain}
        onStatus={toggleGatewayStatus}
        onReload={reloadGatewayConfig}
      />

      <BackupCenter
        backups={backups}
        drills={backupDrills}
        busy={backupBusy}
        onCreate={createSystemBackup}
        onDrill={runSystemBackupDrill}
        onRestore={restoreSystemBackup}
      />
    </div>
  );
}

function GatewayCenter({
  gateway,
  form,
  busy,
  onFormChange,
  onSave,
  onEdit,
  onCanary,
  onDrain,
  onStatus,
  onReload
}: {
  gateway: GatewayOverview | null;
  form: GatewayRouteForm;
  busy: string;
  onFormChange: (form: GatewayRouteForm) => void;
  onSave: () => void;
  onEdit: (route: GatewayRoute) => void;
  onCanary: (route: GatewayRoute, percent: number) => void;
  onDrain: (route: GatewayRoute) => void;
  onStatus: (route: GatewayRoute) => void;
  onReload: () => void;
}) {
  const plan = gateway?.reloadPlan;
  return (
    <section className="panel">
      <div className="between">
        <h3>API Gateway 配置中心</h3>
        <div className="row-actions">
          <button className="soft-button" disabled={!!busy || !plan?.valid} onClick={onReload}>{busy === "reload" ? "记录中" : "记录 reload"}</button>
          <StatusChip value={plan?.reloadRequired ? "warning" : plan?.valid === false ? "critical" : "normal"} />
        </div>
      </div>
      <div className="kpi-grid compact">
        <div><span>Active 路由</span><b>{plan?.activeRoutes || 0}</b></div>
        <div><span>排空路由</span><b>{plan?.drainingRoutes || 0}</b></div>
        <div><span>Reload</span><b>{plan?.reloadRequired ? "required" : "synced"}</b></div>
        <div><span>配置 hash</span><b>{plan?.configHash?.slice(0, 10) || "-"}</b></div>
      </div>
      <div className="form-grid four">
        <label><span>路由名称</span><input value={form.name} onChange={(event) => onFormChange({ ...form, name: event.target.value })} /></label>
        <label><span>路径前缀</span><input value={form.pathPrefix} onChange={(event) => onFormChange({ ...form, pathPrefix: event.target.value })} /></label>
        <label className="span-two"><span>稳定 upstream</span><input value={form.stableUpstream} onChange={(event) => onFormChange({ ...form, stableUpstream: event.target.value })} /></label>
        <label className="span-two"><span>灰度 upstream</span><input value={form.canaryUpstream} onChange={(event) => onFormChange({ ...form, canaryUpstream: event.target.value })} /></label>
        <label><span>灰度比例</span><input value={form.canaryPercent} onChange={(event) => onFormChange({ ...form, canaryPercent: event.target.value })} /></label>
        <label><span>超时秒数</span><input value={form.readTimeoutSec} onChange={(event) => onFormChange({ ...form, readTimeoutSec: event.target.value })} /></label>
        <label><span>状态</span><select value={form.status} onChange={(event) => onFormChange({ ...form, status: event.target.value })}><option value="active">active</option><option value="disabled">disabled</option></select></label>
        <label><span>路由 ID</span><input value={form.id} onChange={(event) => onFormChange({ ...form, id: event.target.value })} /></label>
        <label className="span-all"><span>操作</span><button className="soft-button" disabled={!!busy} onClick={onSave}>{busy === "save" ? "保存中" : "保存网关路由"}</button></label>
      </div>
      <div className="grid-12">
        <div className="span-6">
          <h4>路由与灰度</h4>
          <table>
            <thead><tr><th>路由</th><th>Upstream</th><th>排空</th><th>操作</th></tr></thead>
            <tbody>
              {(gateway?.routes || []).map((route) => (
                <tr key={route.id}>
                  <td>
                    <b>{route.name}</b>
                    <span className="muted block-text">{route.pathPrefix} / {route.status}</span>
                    <span className="muted block-text">timeout {route.readTimeoutSec || 60}s</span>
                  </td>
                  <td>
                    <span className="muted block-text">stable {route.stableUpstream}</span>
                    <span className="muted block-text">canary {route.canaryPercent || 0}% {route.canaryUpstream || "-"}</span>
                  </td>
                  <td>
                    <StatusChip value={route.drainEnabled ? "warning" : "normal"} />
                    <span className="muted block-text">{route.drainUntil || "-"}</span>
                  </td>
                  <td className="row-actions">
                    <button className="soft-button" disabled={!!busy} onClick={() => onEdit(route)}>编辑</button>
                    <button className="soft-button" disabled={!!busy} onClick={() => onCanary(route, route.canaryPercent ? 0 : 25)}>{route.canaryPercent ? "灰度清零" : "灰度25%"}</button>
                    <button className="soft-button" disabled={!!busy} onClick={() => onDrain(route)}>{route.drainEnabled ? "停止排空" : "排空5分钟"}</button>
                    <button className="soft-button" disabled={!!busy} onClick={() => onStatus(route)}>{route.status === "active" ? "停用" : "启用"}</button>
                  </td>
                </tr>
              ))}
              {!(gateway?.routes || []).length ? <tr><td colSpan={4}>暂无网关路由</td></tr> : null}
            </tbody>
          </table>
        </div>
        <div className="span-6">
          <h4>Reload 计划</h4>
          <table>
            <tbody>
              <tr><td>配置路径</td><td>{plan?.configPath || "-"}</td></tr>
              <tr><td>Reload 命令</td><td>{plan?.reloadCommand || "-"}</td></tr>
              <tr><td>生成时间</td><td>{plan?.generatedAt || "-"}</td></tr>
              <tr><td>上次 reload</td><td>{plan?.lastReloadAt || "-"}</td></tr>
              <tr><td>校验</td><td>{plan?.valid ? "valid" : (plan?.errors || []).join("; ") || "-"}</td></tr>
            </tbody>
          </table>
          <label className="span-all">
            <span>Nginx 配置片段</span>
            <textarea rows={9} readOnly value={gateway?.nginxConfig || ""} />
          </label>
        </div>
      </div>
      <table>
        <thead><tr><th>事件</th><th>路由</th><th>动作</th><th>详情</th></tr></thead>
        <tbody>
          {(gateway?.events || []).slice(-8).reverse().map((event) => (
            <tr key={event.id}>
              <td>
                <b>{event.eventNo}</b>
                <span className="muted block-text">{event.createdAt} / {event.actor}</span>
              </td>
              <td>{event.routeName}</td>
              <td>{event.action}</td>
              <td>{event.detail}</td>
            </tr>
          ))}
          {!(gateway?.events || []).length ? <tr><td colSpan={4}>暂无网关事件</td></tr> : null}
        </tbody>
      </table>
    </section>
  );
}

function BackupCenter({
  backups,
  drills,
  busy,
  onCreate,
  onDrill,
  onRestore
}: {
  backups: BackupInfo[];
  drills: BackupDrill[];
  busy: string;
  onCreate: () => void;
  onDrill: () => void;
  onRestore: (backup: BackupInfo) => void;
}) {
  const safeBackups = Array.isArray(backups) ? backups : [];
  const safeDrills = Array.isArray(drills) ? drills : [];
  const latestDrill = safeDrills[0];
  return (
    <section className="panel">
      <div className="between">
        <h3>灾备与恢复演练</h3>
        <div className="row-actions">
          <button className="soft-button" disabled={!!busy} onClick={onCreate}>{busy === "create" ? "创建中" : "创建加密备份"}</button>
          <button className="soft-button" disabled={!!busy} onClick={onDrill}>{busy === "drill" ? "演练中" : "运行恢复演练"}</button>
        </div>
      </div>
      <div className="kpi-grid compact">
        <div><span>备份文件</span><b>{safeBackups.length}</b></div>
        <div><span>最新备份</span><b>{safeBackups[0]?.createdAt || "-"}</b></div>
        <div><span>最新演练</span><b>{latestDrill?.status || "-"}</b></div>
        <div><span>快照大小</span><b>{formatBytes(latestDrill?.snapshotSize || safeBackups[0]?.size || 0)}</b></div>
      </div>
      <div className="grid-12">
        <div className="span-6">
          <h4>加密备份</h4>
          <table>
            <thead><tr><th>文件</th><th>大小</th><th>创建时间</th><th>操作</th></tr></thead>
            <tbody>
              {safeBackups.map((backup) => (
                <tr key={backup.name}>
                  <td>
                    <b>{backup.name}</b>
                    <span className="muted block-text">{backup.path}</span>
                  </td>
                  <td>{formatBytes(backup.size)}</td>
                  <td>{backup.createdAt || "-"}</td>
                  <td className="row-actions">
                    <button className="soft-button" disabled={!!busy} onClick={() => onRestore(backup)}>{busy === `restore:${backup.name}` ? "恢复中" : "恢复"}</button>
                  </td>
                </tr>
              ))}
              {!safeBackups.length ? <tr><td colSpan={4}>暂无加密备份</td></tr> : null}
            </tbody>
          </table>
        </div>
        <div className="span-6">
          <h4>恢复演练记录</h4>
          <table>
            <thead><tr><th>演练</th><th>状态</th><th>检查项</th><th>对象</th></tr></thead>
            <tbody>
              {safeDrills.slice(0, 8).map((drill) => (
                <tr key={drill.id}>
                  <td>
                    <b>{drill.drillNo}</b>
                    <span className="muted block-text">{drill.backupName || "-"}</span>
                    <span className="muted block-text">{drill.startedAt || "-"} / {drill.completedAt || "-"}</span>
                  </td>
                  <td><StatusChip value={drill.status === "passed" ? "normal" : "critical"} /></td>
                  <td>
                    {(drill.checks || []).slice(0, 3).map((check) => <span className="muted block-text" key={check}>{check}</span>)}
                    {drill.error ? <span className="muted block-text">{drill.error}</span> : null}
                  </td>
                  <td>
                    {formatBytes(drill.snapshotSize || 0)}
                    <span className="muted block-text">对象 {Object.values(drill.objectCounts || {}).reduce((sum, count) => sum + count, 0)}</span>
                  </td>
                </tr>
              ))}
              {!safeDrills.length ? <tr><td colSpan={4}>暂无恢复演练记录</td></tr> : null}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}

function UpdateTable({ updates, onDownload, onApply }: { updates: UpdatePackage[]; onDownload: (item: UpdatePackage) => void; onApply: (item: UpdatePackage) => void }) {
  return (
    <table>
      <thead><tr><th>版本</th><th>通道</th><th>状态</th><th>发布</th><th>下载</th><th>应用</th><th>操作</th></tr></thead>
      <tbody>
        {updates.map((update) => (
          <tr key={update.id}>
            <td>
              <b>{update.version}</b>
              <span className="muted block-text">{update.artifactFileName || update.fileName || update.remark || update.checksum}</span>
              <span className="muted block-text">{update.artifactSha256 || update.checksum} / {update.artifactSizeBytes || update.sizeBytes || 0} B</span>
              <span className="muted block-text">{update.packageType || "full"}{update.baseVersion ? ` from ${update.baseVersion}` : ""}{update.targetArtifactSha256 ? ` -> ${update.targetArtifactSha256}` : ""}</span>
              <span className="muted block-text">{update.signatureKeyFingerprint ? `ed25519 ${update.signatureKeyFingerprint}` : update.signature || "-"}</span>
            </td>
            <td>{update.channel}</td>
            <td><StatusChip value={update.status} /></td>
            <td>
              {update.publishedBy || "-"}
              <span className="muted block-text">{update.publishedAt || update.createdAt}</span>
            </td>
            <td>
              {update.downloadCount || 0} 次
              <span className="muted block-text">{update.lastDownloadedAt || "-"}</span>
            </td>
            <td>
              {update.appliedBy || "-"}
              <span className="muted block-text">{update.appliedAt || update.rollbackVersion || "-"}</span>
            </td>
            <td className="row-actions">
              <button className="soft-button" onClick={() => onDownload(update)}>下载</button>
              <button className="soft-button" disabled={update.status === "installed"} onClick={() => onApply(update)}>标记发布</button>
            </td>
          </tr>
        ))}
        {!updates.length ? <tr><td colSpan={7}>暂无更新包</td></tr> : null}
      </tbody>
    </table>
  );
}
