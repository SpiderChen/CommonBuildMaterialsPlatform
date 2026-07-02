import type {
  AuditLog,
  ApprovalFlow,
  ApprovalTask,
  BackupDrill,
  BootstrapData,
  Carrier,
  Company,
  Customer,
  DashboardData,
  Department,
  DeliverySign,
  DeliverySignAttachment,
  DeliverySignLink,
  DeliveryNote,
  DispatchCenterOverview,
  DispatchOrder,
  DispatchSchedule,
  FinanceOverview,
  FieldPolicy,
  GatewayOverview,
  GatewayRoute,
  GeoFence,
  GeoFenceEvent,
  IntegrationEndpoint,
  IntegrationOverview,
  InventoryItem,
  InventoryStocktake,
  InventoryTransfer,
  InvoiceDownload,
  LatestLocation,
  DeviceCredential,
  LaboratoryCalibration,
  LaboratoryEquipment,
  LaboratoryOverview,
  LaboratorySample,
  LaboratoryTestRecord,
  LicenseIssueRecord,
  LicensePackage,
  LicensePackageDownload,
  LicenseRevocation,
  LocationBatchReportResponse,
  LocationReportPayload,
  LoginResult,
  MapProviderConfig,
  MasterDataExport,
  MasterDataImportResult,
  Material,
  ManagementReports,
  MFAEnrollment,
  MixDesign,
  MixDesignPlantProfile,
  MixDesignTrialRun,
  ModuleInfo,
  OIDCLoginStart,
  OIDCProvider,
  OrganizationOverview,
  SCIMProvider,
  SecurityPolicy,
  Payment,
  PaymentPlan,
  Plant,
  PlantBufferFlow,
  PlantBufferLocation,
  PortalOverview,
  Product,
  PricePolicy,
  Project,
  PricingQuote,
  TrackReplay,
  VehicleLocationEvent,
  PluginInfo,
  PluginRun,
  ProductionBatch,
  ProductionDailyReport,
  ProductionOverview,
  ProductionPlan,
  ProductionTask,
  ProcurementOverview,
  PublicDeliverySignDetail,
  QualityInspection,
  QualityException,
  QualityOverview,
  QualitySample,
  RawMaterialInspection,
  RawMaterialReceipt,
  Receivable,
  Receipt,
  RedLetterInfo,
  Role,
  RuleDefinition,
  RuleOverview,
  SalesOrder,
  SalesOrderLine,
  ScaleTicket,
  ScaleDeviceEvent,
  ScaleWeightRecord,
  Site,
  StockYard,
  StockYardFlow,
  StockYardPile,
  Statement,
  SupplierStatement,
  TicketPrintLog,
  TicketVoidLog,
  SystemBundle,
  TaxRate,
  UpdatePackage,
  UpdatePackageDownload,
  User,
  Vehicle,
  VehicleAlarm,
  VehicleDevice,
  Driver,
  CustomerContact,
  CustomerBlacklist,
  CustomerProfile,
  CustomerComplaint,
  Contract,
  ContractAttachment,
  CollectionTask,
  CollectionTemplate,
  CollectionDispatch,
  DataDictionary,
  TransportSettlement,
  TransportSettlementItem,
  WorkflowAutomationRun,
  WorkflowDefinition,
  WorkflowEvent,
  WorkflowEventQuery,
  WorkflowCatalog,
  WorkflowEventPreview,
  WorkflowInstance,
  WorkflowInstanceQuery,
  WorkflowLog,
  WorkflowLogQuery,
  WorkflowTask,
  WorkflowTaskQuery,
  WorkflowDelivery,
  WorkflowDeliveryDispatchBatch,
  WorkflowDeliveryQuery,
  WorkflowInboxItem,
  WorkflowOutbox,
  WorkflowOutboxQuery,
  WorkflowOverview,
  WorkflowSubscription
} from "./types";

const browserOrigin = typeof window !== "undefined" ? window.location.origin : "";
const browserHost = typeof window !== "undefined" ? window.location.hostname : "";
const isWailsHost = browserHost === "wails.localhost" || browserHost.endsWith(".wails.localhost");
const defaultAPIRoot = browserOrigin && !isWailsHost ? `${browserOrigin}/api` : "http://127.0.0.1:8088/api";
const configuredAPIRoot = import.meta.env.VITE_API_BASE_URL || defaultAPIRoot;
const API_ROOT_STORAGE_KEY = "cbmp.apiRoot";

function normalizeAPIRoot(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return normalizeAPIRoot(configuredAPIRoot);
  if (trimmed.startsWith("/")) {
    const root = trimmed.replace(/\/+$/, "") || "/api";
    return root.endsWith("/api") ? root : `${root}/api`;
  }

  const withScheme = !/^[a-z][a-z\d+\-.]*:\/\//i.test(trimmed) ? `http://${trimmed}` : trimmed;
  const parsed = new URL(withScheme);
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("Unsupported server URL protocol");
  }
  parsed.hash = "";
  parsed.search = "";
  parsed.pathname = parsed.pathname.replace(/\/+$/, "");
  if (!parsed.pathname || parsed.pathname === "/") {
    parsed.pathname = "/api";
  } else if (!parsed.pathname.endsWith("/api")) {
    parsed.pathname = `${parsed.pathname}/api`;
  }
  return parsed.toString().replace(/\/+$/, "");
}

function savedAPIRoot() {
  if (typeof window === "undefined") return normalizeAPIRoot(configuredAPIRoot);
  try {
    return normalizeAPIRoot(window.localStorage.getItem(API_ROOT_STORAGE_KEY) || configuredAPIRoot);
  } catch {
    window.localStorage.removeItem(API_ROOT_STORAGE_KEY);
    return normalizeAPIRoot(configuredAPIRoot);
  }
}

let currentAPIRoot = savedAPIRoot();

export function defaultAPIBaseURL() {
  return normalizeAPIRoot(configuredAPIRoot);
}

export function getAPIBaseURL() {
  return currentAPIRoot;
}

export function setAPIBaseURL(value: string) {
  const next = normalizeAPIRoot(value);
  currentAPIRoot = next;
  window.localStorage.setItem(API_ROOT_STORAGE_KEY, next);
  return next;
}

export function resetAPIBaseURL() {
  const next = defaultAPIBaseURL();
  currentAPIRoot = next;
  window.localStorage.removeItem(API_ROOT_STORAGE_KEY);
  return next;
}

type RequestOptions = RequestInit & { anonymous?: boolean };
type QueryValue = string | number | boolean | null | undefined;

function querySuffix(params: object = {}) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params) as Array<[string, QueryValue]>) {
    if (value === undefined || value === null || value === "") continue;
    query.set(key, String(value));
  }
  const text = query.toString();
  return text ? `?${text}` : "";
}

export class APIClient {
  token = localStorage.getItem("cbmp.token") || "";

  baseURL() {
    return getAPIBaseURL();
  }

  defaultBaseURL() {
    return defaultAPIBaseURL();
  }

  setBaseURL(value: string) {
    return setAPIBaseURL(value);
  }

  resetBaseURL() {
    return resetAPIBaseURL();
  }

  async testConnection(value = this.baseURL()) {
    const root = normalizeAPIRoot(value);
    const response = await fetch(`${root}/health`, {
      credentials: "include"
    });
    if (!response.ok) {
      const payload = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(payload.error || response.statusText);
    }
    return response.json().catch(() => ({ status: "ok" }));
  }

  async login(username: string, password: string, mfaCode = "") {
    const result = await this.request<LoginResult>("/auth/login", {
      anonymous: true,
      method: "POST",
      body: JSON.stringify({ username, password, mfaCode })
    });
    if (result.token) {
      this.token = result.token;
      localStorage.setItem("cbmp.token", result.token);
    }
    return result;
  }

  async ssoProviders() {
    return this.request<OIDCProvider[]>("/auth/sso/providers", { anonymous: true });
  }

  async startSSO(providerCode: string) {
    return this.request<OIDCLoginStart>(`/auth/sso/${providerCode}/start`, {
      anonymous: true,
      method: "POST",
      body: "{}"
    });
  }

  async bootstrap() {
    return this.request<BootstrapData>("/bootstrap");
  }

  async updateAccountProfile(payload: { displayName: string }) {
    return this.request<User>("/account/profile", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async changeAccountPassword(payload: { currentPassword: string; newPassword: string }) {
    return this.request<User>("/account/password", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async dashboard() {
    return this.request<DashboardData>("/dashboard");
  }

  async reports() {
    return this.request<ManagementReports>("/reports");
  }

  async dispatchCenterOverview() {
    return this.request<DispatchCenterOverview>("/dispatch-center/overview");
  }

  async orders() {
    return this.request<SalesOrder[]>("/orders");
  }

  async createOrder(payload: Omit<Partial<SalesOrder>, "lines"> & { lines?: Array<Partial<SalesOrderLine> & Pick<SalesOrderLine, "productId" | "quantity">> }) {
    return this.request<SalesOrder>("/orders", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createCustomerContact(payload: Partial<CustomerContact> & Pick<CustomerContact, "customerId" | "name" | "phone">) {
    return this.request<CustomerContact>("/master/customer-contacts", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createCustomer(payload: Partial<Customer> & Pick<Customer, "name">) {
    return this.request<Customer>("/master/customers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createProject(payload: Partial<Project> & Pick<Project, "customerId" | "name">) {
    return this.request<Project>("/master/projects", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createProduct(payload: Partial<Product> & Pick<Product, "name">) {
    return this.request<Product>("/master/products", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createMaterial(payload: Partial<Material> & Pick<Material, "name">) {
    return this.request<Material>("/master/materials", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createDriver(payload: Partial<Driver> & Pick<Driver, "name">) {
    return this.request<Driver>("/master/drivers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createCarrier(payload: Partial<Carrier> & Pick<Carrier, "name">) {
    return this.request<Carrier>("/master/carriers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createVehicle(payload: Partial<Vehicle> & Pick<Vehicle, "plateNo">) {
    return this.request<Vehicle>("/master/vehicles", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createVehicleDevice(payload: Partial<VehicleDevice> & Pick<VehicleDevice, "vehicleId" | "deviceNo">) {
    return this.request<VehicleDevice>("/master/vehicle-devices", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createSite(payload: Partial<Site> & Pick<Site, "name" | "code">) {
    return this.request<Site>("/master/sites", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPlant(payload: Partial<Plant> & Pick<Plant, "siteId" | "name" | "code">) {
    return this.request<Plant>("/master/plants", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPlantBufferLocation(payload: Partial<PlantBufferLocation> & Pick<PlantBufferLocation, "siteId" | "plantId" | "code" | "name">) {
    return this.request<PlantBufferLocation>("/master/plant-buffer-locations", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createStockYard(payload: Partial<StockYard> & Pick<StockYard, "siteId" | "code" | "name">) {
    return this.request<StockYard>("/master/stock-yards", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createStockYardPile(payload: Partial<StockYardPile> & Pick<StockYardPile, "yardId" | "code" | "name">) {
    return this.request<StockYardPile>("/master/stock-yard-piles", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createInventoryItem(payload: Partial<InventoryItem> & Pick<InventoryItem, "siteId" | "materialId">) {
    return this.request<InventoryItem>("/master/inventory", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateMasterResource<T>(resource: string, id: number, payload: Partial<T>) {
    return this.request<T>(`/master/${resource}/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    });
  }

  async deleteMasterResource<T>(resource: string, id: number) {
    return this.request<T>(`/master/${resource}/${id}`, { method: "DELETE" });
  }

  async setDefaultCustomerContact(id: number) {
    return this.request<CustomerContact>(`/master/customer-contacts/${id}/default`, { method: "POST" });
  }

  async createTaxRate(payload: Partial<TaxRate>) {
    return this.request<TaxRate>("/master/tax-rates", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async exportMasterData(resource: string) {
    return this.request<MasterDataExport>(`/master/export?resource=${encodeURIComponent(resource)}`);
  }

  async importMasterData(payload: { resource: string; mode: string; rows: Record<string, unknown>[] }) {
    return this.request<MasterDataImportResult>("/master/import", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPricePolicy(payload: Partial<PricePolicy>) {
    return this.request<PricePolicy>("/master/price-policies", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async evaluatePricing(payload: { customerId: number; projectId: number; productId: number; planTime?: string; planQuantity?: number; unitPrice?: number }) {
    return this.request<PricingQuote>("/master/pricing/evaluate", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createCustomerBlacklist(payload: Partial<CustomerBlacklist> & Pick<CustomerBlacklist, "customerId" | "reason">) {
    return this.request<CustomerBlacklist>("/master/customer-blacklists", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async releaseCustomerBlacklist(id: number) {
    return this.request<CustomerBlacklist | WorkflowInstance>(`/master/customer-blacklists/${id}/release`, { method: "POST" });
  }

  async evaluateCustomerProfiles() {
    return this.request<CustomerProfile[]>("/master/customer-profiles/evaluate", {
      method: "POST",
      body: "{}"
    });
  }

  async createCustomerProfile(payload: Partial<CustomerProfile> & Pick<CustomerProfile, "customerId" | "grade" | "riskLevel">) {
    return this.request<CustomerProfile>("/master/customer-profiles", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createCustomerComplaint(payload: Partial<CustomerComplaint> & Pick<CustomerComplaint, "customerId" | "title">) {
    return this.request<CustomerComplaint>("/master/customer-complaints", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async portalOverview() {
    return this.request<PortalOverview>("/portal/overview");
  }

  async portalComplaints() {
    return this.request<CustomerComplaint[]>("/portal/complaints");
  }

  async createPortalComplaint(payload: Partial<CustomerComplaint> & Pick<CustomerComplaint, "title">) {
    return this.request<CustomerComplaint>("/portal/complaints", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reportPortalDispatchException(id: number, payload: { exception: string; level?: string; alarmType?: string }) {
    return this.request<DispatchOrder>(`/portal/dispatches/${id}/exception`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reportLocationBatch(reports: LocationReportPayload[]) {
    return this.request<LocationBatchReportResponse>("/iot/vehicle/location/batch", {
      method: "POST",
      body: JSON.stringify({ reports })
    });
  }

  async closeCustomerComplaint(id: number, resolution: string) {
    return this.request<CustomerComplaint>(`/master/customer-complaints/${id}/close`, {
      method: "POST",
      body: JSON.stringify({ resolution })
    });
  }

  async createContractAttachment(contractId: number, payload: Partial<ContractAttachment> & Pick<ContractAttachment, "fileName">) {
    return this.request<ContractAttachment>(`/contracts/${contractId}/attachments`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async contractAttachments(contractId: number) {
    return this.request<ContractAttachment[]>(`/contracts/${contractId}/attachments`);
  }

  async contracts() {
    return this.request<Contract[]>("/contracts");
  }

  async createContract(payload: Partial<Contract> & Pick<Contract, "customerId" | "projectId" | "name" | "items">) {
    return this.request<Contract>("/contracts", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async submitContract(id: number, reason: string) {
    return this.request<Contract>(`/contracts/${id}/submit`, {
      method: "POST",
      body: JSON.stringify({ reason })
    });
  }

  async reviseContract(id: number, payload: Partial<Contract>) {
    return this.request<Contract>(`/contracts/${id}/revise`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async approveOrder(id: number) {
    return this.request<SalesOrder>(`/orders/${id}/approve`, { method: "POST" });
  }

  async dispatchOrders() {
    return this.request<DispatchOrder[]>("/dispatch-orders");
  }

  async dispatchSchedules() {
    return this.request<DispatchSchedule[]>("/dispatch-orders/schedules");
  }

  async createDispatchSchedule(payload: Partial<DispatchSchedule> & Pick<DispatchSchedule, "siteId" | "vehicleId">) {
    return this.request<DispatchSchedule>("/dispatch-orders/schedules", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async carrierSettlements() {
    return this.request<{ settlements: TransportSettlement[]; items: TransportSettlementItem[] }>("/dispatch-orders/carrier-settlements");
  }

  async generateCarrierSettlement(payload: { carrierId?: number; period?: string; ratePerTrip?: number; ratePerUnit?: number }) {
    return this.request<{ settlement: TransportSettlement; items: TransportSettlementItem[] }>("/dispatch-orders/carrier-settlements/generate", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createDispatch(payload: Partial<DispatchOrder>) {
    return this.request<DispatchOrder>("/dispatch-orders", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async advanceDispatch(id: number, status?: string) {
    return this.request<DispatchOrder>(`/dispatch-orders/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async tickets() {
    return this.request<ScaleTicket[]>("/weighbridge/tickets");
  }

  async createTicket(payload: Partial<ScaleTicket>) {
    return this.request<ScaleTicket>("/weighbridge/tickets", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createTransferTicket(payload: Partial<ScaleTicket> & Pick<ScaleTicket, "transferId">) {
    return this.request<ScaleTicket>("/weighbridge/tickets/transfer", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createReturnTicket(payload: Partial<ScaleTicket> & Pick<ScaleTicket, "dispatchId">) {
    return this.request<ScaleTicket>("/weighbridge/tickets/return", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createWasteTicket(payload: Partial<ScaleTicket> & Pick<ScaleTicket, "siteId" | "materialId">) {
    return this.request<ScaleTicket>("/weighbridge/tickets/waste", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async ticketPrintLogs() {
    return this.request<TicketPrintLog[]>("/weighbridge/ticket-prints");
  }

  async ticketVoidLogs() {
    return this.request<TicketVoidLog[]>("/weighbridge/ticket-voids");
  }

  async weightRecords() {
    return this.request<ScaleWeightRecord[]>("/weighbridge/weight-records");
  }

  async scaleDeviceEvents() {
    return this.request<ScaleDeviceEvent[]>("/weighbridge/device-events");
  }

  async reprintTicket(id: number) {
    return this.request<TicketPrintLog>(`/weighbridge/tickets/${id}/reprint`, { method: "POST" });
  }

  async requestTicketVoid(id: number, reason: string) {
    return this.request<TicketVoidLog>(`/weighbridge/tickets/${id}/void/request`, {
      method: "POST",
      body: JSON.stringify({ reason })
    });
  }

  async approveTicketVoid(id: number, approved: boolean) {
    return this.request<TicketVoidLog>(`/weighbridge/tickets/${id}/void/approve`, {
      method: "POST",
      body: JSON.stringify({ approved })
    });
  }

  async signs() {
    return this.request<DeliverySign[]>("/delivery/sign");
  }

  async deliveryNotes() {
    return this.request<DeliveryNote[]>("/delivery/notes");
  }

  async createDeliveryNote(payload: { dispatchId: number; ticketId?: number; status?: string }) {
    return this.request<DeliveryNote>("/delivery/notes", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateDeliveryNoteStatus(id: number, status: string) {
    return this.request<DeliveryNote>(`/delivery/notes/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async createDeliveryNoteSignLink(id: number, payload: { channel: string; phone?: string; expiresAt?: string }) {
    return this.request<DeliverySignLink>(`/delivery/notes/${id}/sign-link`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reprintDeliveryNote(id: number) {
    return this.request<DeliveryNote>(`/delivery/notes/${id}/reprint`, {
      method: "POST",
      body: "{}"
    });
  }

  async signDelivery(payload: Partial<DeliverySign> & { attachments?: Partial<DeliverySignAttachment>[] }) {
    return this.request<DeliverySign>("/delivery/sign", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async signLinks() {
    return this.request<DeliverySignLink[]>("/delivery/sign-links");
  }

  async createSignLink(payload: { dispatchId: number; ticketId?: number; channel: string; phone?: string; expiresAt?: string }) {
    return this.request<DeliverySignLink>("/delivery/sign-links", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async signAttachments() {
    return this.request<DeliverySignAttachment[]>("/delivery/sign-attachments");
  }

  async addSignAttachment(signId: number, payload: Partial<DeliverySignAttachment>) {
    return this.request<DeliverySignAttachment>(`/delivery/sign/${signId}/attachments`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async publicSignDetail(token: string) {
    return this.request<PublicDeliverySignDetail>(`/public/delivery-sign/${token}`, { anonymous: true });
  }

  async publicSign(token: string, payload: Partial<DeliverySign> & { attachments?: Partial<DeliverySignAttachment>[] }) {
    return this.request<DeliverySign>(`/public/delivery-sign/${token}`, {
      method: "POST",
      anonymous: true,
      body: JSON.stringify(payload)
    });
  }

  async statements() {
    return this.request<Statement[]>("/statements");
  }

  async confirmStatement(id: number) {
    return this.request<Statement>(`/statements/${id}/confirm`, { method: "POST" });
  }

  async latestLocations() {
    return this.request<LatestLocation[]>("/vehicle/location/latest");
  }

  async trackReplay(vehicleId: number) {
    return this.request<TrackReplay>(`/vehicle/track/replay?vehicleId=${encodeURIComponent(String(vehicleId))}`);
  }

  async vehicleTrack(params: { vehicleId?: number; startTime?: string; endTime?: string } = {}) {
    return this.request<VehicleLocationEvent[]>(`/vehicle/track${querySuffix(params)}`);
  }

  async alarms() {
    return this.request<VehicleAlarm[]>("/vehicle/alarms");
  }

  async geoFences() {
    return this.request<GeoFence[]>("/vehicle/fences");
  }

  async createGeoFence(payload: Partial<GeoFence>) {
    return this.request<GeoFence>("/vehicle/fences", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateGeoFence(id: number, payload: Partial<GeoFence>) {
    return this.request<GeoFence>(`/vehicle/fences/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    });
  }

  async archiveGeoFence(id: number) {
    return this.request<GeoFence>(`/vehicle/fences/${id}`, { method: "DELETE" });
  }

  async geoFenceEvents(params: { vehicleId?: number; fenceId?: number; limit?: number } = {}) {
    const query = new URLSearchParams();
    if (params.vehicleId) query.set("vehicleId", String(params.vehicleId));
    if (params.fenceId) query.set("fenceId", String(params.fenceId));
    if (params.limit) query.set("limit", String(params.limit));
    const suffix = query.toString() ? `?${query.toString()}` : "";
    return this.request<GeoFenceEvent[]>(`/vehicle/fence-events${suffix}`);
  }

  async procurementOverview() {
    return this.request<ProcurementOverview>("/procurement/overview");
  }

  async createRawMaterialReceipt(payload: Partial<RawMaterialReceipt>) {
    return this.request<RawMaterialReceipt>("/procurement/receipts", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createInventoryTransfer(payload: Partial<InventoryTransfer>) {
    return this.request<InventoryTransfer>("/procurement/transfers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async completeInventoryTransfer(id: number) {
    return this.request<InventoryTransfer>(`/procurement/transfers/${id}/complete`, { method: "POST" });
  }

  async createInventoryStocktake(payload: Partial<InventoryStocktake>) {
    return this.request<InventoryStocktake>("/procurement/stocktakes", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reviewInventoryStocktake(id: number) {
    return this.request<InventoryStocktake>(`/procurement/stocktakes/${id}/review`, { method: "POST" });
  }

  async createStockYardReceipt(payload: { pileId?: number; pileCode?: string; materialId?: number; supplierId?: number; batchNo?: string; quantity: number; unit?: string; moistureRate?: number; qualityStatus?: string; remark?: string }) {
    return this.request<StockYardFlow>("/procurement/yard-receipts", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createStockYardAdjustment(payload: { pileId?: number; pileCode?: string; actualQty: number; moistureRate?: number; qualityStatus?: string; status?: string; remark?: string }) {
    return this.request<StockYardFlow | WorkflowInstance>("/procurement/yard-adjustments", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async productionOverview() {
    return this.request<ProductionOverview>("/production-plans/overview");
  }

  async createPlantBufferTransfer(payload: { bufferId?: number; bufferCode?: string; yardPileId?: number; yardPileCode?: string; materialId?: number; quantity: number; unit?: string; remark?: string }) {
    return this.request<PlantBufferFlow>("/production-plans/buffer-transfers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPlantBufferAdjustment(payload: { bufferId?: number; bufferCode?: string; actualQty: number; moistureRate?: number; qualityStatus?: string; status?: string; remark?: string }) {
    return this.request<PlantBufferFlow | WorkflowInstance>("/production-plans/buffer-adjustments", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createProductionPlan(payload: Partial<ProductionPlan>) {
    return this.request<ProductionPlan>("/production-plans", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateProductionPlan(planId: number, payload: Partial<ProductionPlan>) {
    return this.request<ProductionPlan>(`/production-plans/${planId}`, {
      method: "PATCH",
      body: JSON.stringify(payload)
    });
  }

  async cancelProductionPlan(planId: number) {
    return this.request<ProductionPlan>(`/production-plans/${planId}/cancel`, {
      method: "POST"
    });
  }

  async createProductionTask(planId: number, payload: Partial<ProductionTask> = {}) {
    return this.request<ProductionTask>(`/production-plans/${planId}/tasks`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async autoCreateProductionTasks(planId: number, payload: { taskQty?: number } = {}) {
    return this.request<ProductionTask[]>(`/production-plans/${planId}/tasks/auto`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createProductionBatch(taskId: number, payload: Partial<ProductionBatch> = {}) {
    return this.request<ProductionBatch>(`/production-plans/tasks/${taskId}/batches`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async generateProductionReport(payload: { siteId: number; reportDate: string }) {
    return this.request<ProductionDailyReport>("/production-plans/reports/generate", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async qualityOverview() {
    return this.request<QualityOverview>("/quality/overview");
  }

  async laboratoryOverview() {
    return this.request<LaboratoryOverview>("/laboratory/overview");
  }

  async createLaboratoryMixDesign(payload: Partial<MixDesign> & Pick<MixDesign, "productId" | "materials">) {
    return this.request<MixDesign>("/laboratory/mix-designs", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reviseLaboratoryMixDesign(id: number, payload: Partial<MixDesign>) {
    return this.request<MixDesign>(`/laboratory/mix-designs/${id}/revise`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async approveLaboratoryMixDesign(id: number, payload: { trialRunId?: number; effectiveFrom?: string; effectiveTo?: string } = {}) {
    return this.request<MixDesign>(`/laboratory/mix-designs/${id}/approve`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async retireLaboratoryMixDesign(id: number) {
    return this.request<MixDesign>(`/laboratory/mix-designs/${id}/retire`, { method: "POST", body: "{}" });
  }

  async createMixDesignPlantProfile(mixDesignId: number, payload: Partial<MixDesignPlantProfile> & Pick<MixDesignPlantProfile, "plantId" | "materials">) {
    return this.request<MixDesignPlantProfile>(`/laboratory/mix-designs/${mixDesignId}/plant-profiles`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async approveMixDesignPlantProfile(id: number, payload: { effectiveFrom?: string; effectiveTo?: string } = {}) {
    return this.request<MixDesignPlantProfile>(`/laboratory/mix-design-plant-profiles/${id}/approve`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async retireMixDesignPlantProfile(id: number) {
    return this.request<MixDesignPlantProfile>(`/laboratory/mix-design-plant-profiles/${id}/retire`, { method: "POST", body: "{}" });
  }

  async createMixDesignTrialRun(id: number, payload: Partial<MixDesignTrialRun>) {
    return this.request<MixDesignTrialRun>(`/laboratory/mix-designs/${id}/trial-runs`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createLaboratorySample(payload: Partial<LaboratorySample>) {
    return this.request<LaboratorySample>("/laboratory/samples", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createLaboratoryTest(sampleId: number, payload: Partial<LaboratoryTestRecord>) {
    return this.request<LaboratoryTestRecord>(`/laboratory/samples/${sampleId}/tests`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reviewLaboratoryTest(id: number, payload: Partial<LaboratoryTestRecord>) {
    return this.request<LaboratoryTestRecord>(`/laboratory/tests/${id}/review`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createLaboratoryEquipment(payload: Partial<LaboratoryEquipment>) {
    return this.request<LaboratoryEquipment>("/laboratory/equipment", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createLaboratoryCalibration(id: number, payload: Partial<LaboratoryCalibration>) {
    return this.request<LaboratoryCalibration>(`/laboratory/equipment/${id}/calibrations`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createQualityException(payload: Partial<QualityException>) {
    return this.request<QualityException>("/laboratory/exceptions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async handleQualityException(id: number, payload: Partial<QualityException>) {
    return this.request<QualityException>(`/laboratory/exceptions/${id}/handle`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createQualityInspection(payload: Partial<QualityInspection> & Pick<QualityInspection, "batchId">) {
    return this.request<QualityInspection>("/quality/inspections", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async testQualitySample(id: number, payload: Partial<QualitySample>) {
    return this.request<QualitySample>(`/quality/samples/${id}/test`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createRawMaterialInspection(payload: Partial<RawMaterialInspection> & Pick<RawMaterialInspection, "receiptId">) {
    return this.request<RawMaterialInspection>("/quality/raw-inspections", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async reviewRawMaterialInspection(id: number, payload: Partial<RawMaterialInspection>) {
    return this.request<RawMaterialInspection | WorkflowInstance>(`/quality/raw-inspections/${id}/review`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async financeOverview() {
    return this.request<FinanceOverview>("/finance/overview");
  }

  async createInvoice(statementId: number, invoiceCategory = "blue_vat_special") {
    return this.request("/finance/invoices", {
      method: "POST",
      body: JSON.stringify({ statementId, invoiceCategory })
    });
  }

  async submitTaxInvoice(id: number) {
    return this.request(`/finance/invoices/${id}/submit-tax`, { method: "POST" });
  }

  async createRedLetterInfo(originalInvoiceId: number, reason: string) {
    return this.request<RedLetterInfo>("/finance/red-letters", {
      method: "POST",
      body: JSON.stringify({ originalInvoiceId, reason })
    });
  }

  async approveRedLetterInfo(id: number) {
    return this.request<RedLetterInfo>(`/finance/red-letters/${id}/approve`, { method: "POST" });
  }

  async redOffsetInvoice(id: number, reason: string, redLetterInfoId = 0) {
    return this.request(`/finance/invoices/${id}/red-offset`, {
      method: "POST",
      body: JSON.stringify({ reason, redLetterInfoId })
    });
  }

  async downloadInvoice(id: number) {
    return this.request<InvoiceDownload>(`/finance/invoices/${id}/download`);
  }

  async createReceipt(payload: Partial<Receipt> & Pick<Receipt, "receivableId" | "amount">) {
    return this.request<Receipt>("/finance/receipts", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPayment(payload: Partial<Payment> & Pick<Payment, "payableId" | "amount">) {
    return this.request<Payment>("/finance/payments", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createPaymentPlan(payload: Partial<PaymentPlan> & Pick<PaymentPlan, "receivableId" | "amount">) {
    return this.request<PaymentPlan>("/finance/payment-plans", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async settlePaymentPlan(id: number) {
    return this.request<{ paymentPlan: PaymentPlan; receipt: Receipt }>(`/finance/payment-plans/${id}/settle`, {
      method: "POST",
      body: "{}"
    });
  }

  async generateCollectionTasks() {
    return this.request<CollectionTask[]>("/finance/collections/generate", {
      method: "POST",
      body: "{}"
    });
  }

  async createCollectionTemplate(payload: Partial<CollectionTemplate> & Pick<CollectionTemplate, "name" | "level" | "channel" | "content">) {
    return this.request<CollectionTemplate>("/finance/collection-templates", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async sendCollectionTask(id: number, templateId = 0, channel = "") {
    return this.request<CollectionDispatch>(`/finance/collections/${id}/send`, {
      method: "POST",
      body: JSON.stringify({ templateId, channel })
    });
  }

  async handleCollectionTask(id: number, remark: string) {
    return this.request<CollectionTask>(`/finance/collections/${id}/handle`, {
      method: "POST",
      body: JSON.stringify({ remark })
    });
  }

  async createSupplierStatement(payload: Partial<SupplierStatement> & Pick<SupplierStatement, "supplierId">) {
    return this.request<SupplierStatement>("/finance/supplier-statements", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async approveSupplierStatement(id: number) {
    return this.request<SupplierStatement>(`/finance/supplier-statements/${id}/approve`, { method: "POST" });
  }

  async rulesOverview() {
    return this.request<RuleOverview>("/rules");
  }

  async saveRuleDefinition(payload: Partial<RuleDefinition> & Pick<RuleDefinition, "code" | "name" | "metric">) {
    return this.request<RuleDefinition>("/rules/definitions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setRuleDefinitionStatus(id: number, enabled: boolean) {
    return this.request<RuleDefinition>(`/rules/definitions/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ enabled })
    });
  }

  async deleteRuleDefinition(id: number) {
    return this.request<RuleDefinition>(`/rules/definitions/${id}`, { method: "DELETE" });
  }

  async evaluateRules() {
    return this.request("/rules/evaluate", { method: "POST" });
  }

  async handleAlarm(id: number, remark: string) {
    return this.request<VehicleAlarm>(`/rules/alarms/${id}/handle`, {
      method: "POST",
      body: JSON.stringify({ remark })
    });
  }

  async integrationsOverview() {
    return this.request<IntegrationOverview>("/integrations/overview");
  }

  async saveIntegrationEndpoint(payload: Partial<IntegrationEndpoint> & Pick<IntegrationEndpoint, "name" | "type">) {
    return this.request<IntegrationEndpoint>("/integrations/endpoints", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setIntegrationEndpointStatus(id: number, status: string) {
    return this.request<IntegrationEndpoint>(`/integrations/endpoints/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteIntegrationEndpoint(id: number) {
    return this.request<IntegrationEndpoint>(`/integrations/endpoints/${id}`, { method: "DELETE" });
  }

  async approvals() {
    return this.request<ApprovalTask[]>("/approvals");
  }

  async actApproval(id: number, action: "approve" | "reject", comment: string) {
    return this.request<ApprovalTask>(`/approvals/${id}/act`, {
      method: "POST",
      body: JSON.stringify({ action, comment })
    });
  }

  async systemBundle(): Promise<SystemBundle> {
    const [plugins, pluginRuns, updates, licenseVerified, licensePackages, licenseIssues, licenseRevocations, licensePortal, security, runtime, backups, backupDrills, gateway, approvalFlows, workflows, dictionaries] = await Promise.all([
      this.request<PluginInfo[]>("/system/plugins"),
      this.request<PluginRun[]>("/system/plugins/runs"),
      this.request<UpdatePackage[]>("/system/updates"),
      this.request<SystemBundle["licenseVerified"]>("/system/license/verify"),
      this.request<LicensePackage[]>("/system/license/packages"),
      this.request<LicenseIssueRecord[]>("/system/license/issues"),
      this.request<LicenseRevocation[]>("/system/license/revocations"),
      this.request<SystemBundle["licensePortal"]>("/system/license/portal"),
      this.request<SystemBundle["security"]>("/system/security"),
      this.request<SystemBundle["runtime"]>("/system/runtime"),
      this.request<SystemBundle["backups"]>("/system/backups"),
      this.request<BackupDrill[]>("/system/backups/drills"),
      this.request<GatewayOverview>("/system/gateway"),
      this.request<ApprovalFlow[]>("/system/approval-flows"),
      this.request<WorkflowOverview>("/system/workflows"),
      this.request<DataDictionary[]>("/system/dictionaries")
    ]);
    return { plugins, pluginRuns, updates, licenseVerified, licensePackages, licenseIssues, licenseRevocations, licensePortal, security, runtime, backups, backupDrills, gateway, approvalFlows, workflows, dictionaries };
  }

  async systemModules() {
    return this.request<ModuleInfo[]>("/system/modules");
  }

  async setSystemModuleEnabled(code: string, enabled: boolean) {
    return this.request<ModuleInfo>(`/system/modules/${encodeURIComponent(code)}`, {
      method: "POST",
      body: JSON.stringify({ enabled })
    });
  }

  async workflowOverview() {
    return this.request<WorkflowOverview>("/system/workflows");
  }

  async workflowCatalog() {
    return this.request<WorkflowCatalog>("/system/workflows/catalog");
  }

  async workflowInbox() {
    return this.request<WorkflowInboxItem[]>("/system/workflows/inbox");
  }

  async workflowInstances(params: WorkflowInstanceQuery = {}) {
    return this.request<WorkflowInstance[]>(`/system/workflows/instances${querySuffix(params)}`);
  }

  async workflowTasks(params: WorkflowTaskQuery = {}) {
    return this.request<WorkflowTask[]>(`/system/workflows/tasks${querySuffix(params)}`);
  }

  async workflowLogs(params: WorkflowLogQuery = {}) {
    return this.request<WorkflowLog[]>(`/system/workflows/logs${querySuffix(params)}`);
  }

  async systemUsers() {
    return this.request<User[]>("/system/users");
  }

  async systemDictionaries() {
    return this.request<DataDictionary[]>("/system/dictionaries");
  }

  async auditLogs() {
    return this.request<AuditLog[]>("/system/audit");
  }

  async saveMenuLabel(payload: { key: string; label: string }) {
    return this.request<Record<string, string>>("/system/menu-labels", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async resetMenuLabel(key: string) {
    return this.request<Record<string, string>>(`/system/menu-labels/${encodeURIComponent(key)}`, { method: "DELETE" });
  }

  async orgOverview() {
    return this.request<OrganizationOverview>("/system/org");
  }

  async saveSystemUser(payload: Partial<User> & { password?: string; username: string; displayName: string; roleCode: string }) {
    return this.request<User>("/system/users", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setSystemUserStatus(id: number, status: string) {
    return this.request<User>(`/system/users/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async systemRoles() {
    return this.request<Role[]>("/system/roles");
  }

  async saveSystemRole(payload: Partial<Role> & { code: string; name: string }) {
    return this.request<Role>("/system/roles", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async deleteSystemRole(id: number) {
    return this.request<Role>(`/system/roles/${id}`, { method: "DELETE" });
  }

  async saveApprovalFlow(payload: Partial<ApprovalFlow>) {
    return this.request<ApprovalFlow>("/system/approval-flows", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setApprovalFlowStatus(id: number, status: string) {
    return this.request<ApprovalFlow>(`/system/approval-flows/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteApprovalFlow(id: number) {
    return this.request<ApprovalFlow>(`/system/approval-flows/${id}`, { method: "DELETE" });
  }

  async saveWorkflowDefinition(payload: Partial<WorkflowDefinition>) {
    return this.request<WorkflowDefinition>("/system/workflows/definitions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async publishWorkflowDefinitionVersion(id: number, payload: Partial<WorkflowDefinition>) {
    return this.request<WorkflowDefinition>(`/system/workflows/definitions/${id}/publish`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async rollbackWorkflowDefinitionVersion(id: number) {
    return this.request<WorkflowDefinition>(`/system/workflows/definitions/${id}/rollback`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async saveWorkflowSubscription(payload: Partial<WorkflowSubscription>) {
    return this.request<WorkflowSubscription>("/system/workflows/subscriptions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setWorkflowSubscriptionStatus(id: number, status: string) {
    return this.request<WorkflowSubscription>(`/system/workflows/subscriptions/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteWorkflowSubscription(id: number) {
    return this.request<WorkflowSubscription>(`/system/workflows/subscriptions/${id}`, { method: "DELETE" });
  }

  async workflowEvents(params: WorkflowEventQuery = {}) {
    return this.request<WorkflowEvent[]>(`/system/workflows/events${querySuffix(params)}`);
  }

  async workflowOutbox(params: WorkflowOutboxQuery = {}) {
    return this.request<WorkflowOutbox[]>(`/system/workflows/outbox${querySuffix(params)}`);
  }

  async workflowDeliveries(params: WorkflowDeliveryQuery = {}) {
    return this.request<WorkflowDelivery[]>(`/system/workflows/deliveries${querySuffix(params)}`);
  }

  async publishWorkflowEvent(payload: {
    eventType: string;
    source?: string;
    eventKey?: string;
    actor?: string;
    resource: string;
    resourceId?: number;
    resourceNo?: string;
    title?: string;
    reason?: string;
    variables?: Record<string, string>;
  }) {
    return this.request<WorkflowEvent>("/system/workflows/events", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async previewWorkflowEvent(payload: {
    eventType: string;
    source?: string;
    eventKey?: string;
    actor?: string;
    resource: string;
    resourceId?: number;
    resourceNo?: string;
    title?: string;
    reason?: string;
    variables?: Record<string, string>;
  }) {
    return this.request<WorkflowEventPreview>("/system/workflows/events/preview", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async replayWorkflowEvent(id: number) {
    return this.request<WorkflowEvent>(`/system/workflows/events/${id}/replay`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async cancelWorkflowInstance(id: number, reason: string) {
    return this.request<WorkflowInstance>(`/system/workflows/instances/${id}/cancel`, {
      method: "POST",
      body: JSON.stringify({ reason })
    });
  }

  async actWorkflowTask(id: number, action: "approve" | "reject", comment: string) {
    return this.request<WorkflowInstance>(`/system/workflows/tasks/${id}/act`, {
      method: "POST",
      body: JSON.stringify({ action, comment })
    });
  }

  async reassignWorkflowTask(id: number, roleCode: string, reason: string) {
    return this.request<WorkflowInstance>(`/system/workflows/tasks/${id}/reassign`, {
      method: "POST",
      body: JSON.stringify({ roleCode, reason })
    });
  }

  async escalateWorkflowTask(id: number, roleCode: string, reason: string) {
    return this.request<WorkflowInstance>(`/system/workflows/tasks/${id}/escalate`, {
      method: "POST",
      body: JSON.stringify({ roleCode, reason })
    });
  }

  async acknowledgeWorkflowOutbox(id: number) {
    return this.request<WorkflowOutbox>(`/system/workflows/outbox/${id}/ack`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async claimWorkflowOutbox(id: number, consumer = "erp-console") {
    return this.request<WorkflowOutbox>(`/system/workflows/outbox/${id}/claim`, {
      method: "POST",
      body: JSON.stringify({ consumer })
    });
  }

  async failWorkflowOutbox(id: number, error = "投递失败", retryAfterMinutes = 5) {
    return this.request<WorkflowOutbox>(`/system/workflows/outbox/${id}/fail`, {
      method: "POST",
      body: JSON.stringify({ error, retryAfterMinutes })
    });
  }

  async dispatchWorkflowDelivery(id: number) {
    return this.request<WorkflowDelivery>(`/system/workflows/deliveries/${id}/dispatch`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async resetWorkflowDelivery(id: number) {
    return this.request<WorkflowDelivery>(`/system/workflows/deliveries/${id}/reset`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async dispatchDueWorkflowDeliveries(limit = 20) {
    return this.request<WorkflowDeliveryDispatchBatch>(`/system/workflows/deliveries/dispatch-due?limit=${encodeURIComponent(String(limit))}`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async runWorkflowAutomation(deliveryLimit = 20, escalationLimit = 20) {
    return this.request<WorkflowAutomationRun>(`/system/workflows/automation/run?deliveryLimit=${encodeURIComponent(String(deliveryLimit))}&escalationLimit=${encodeURIComponent(String(escalationLimit))}`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async resetWorkflowOutbox(id: number) {
    return this.request<WorkflowOutbox>(`/system/workflows/outbox/${id}/reset`, {
      method: "POST",
      body: JSON.stringify({})
    });
  }

  async resolveWorkflowEvent(id: number, resolution: string) {
    return this.request<WorkflowEvent>(`/system/workflows/events/${id}/resolve`, {
      method: "POST",
      body: JSON.stringify({ resolution })
    });
  }

  async saveDictionary(payload: Partial<DataDictionary>) {
    return this.request<DataDictionary>("/system/dictionaries", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setDictionaryStatus(id: number, status: string) {
    return this.request<DataDictionary>(`/system/dictionaries/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteDictionary(id: number) {
    return this.request<DataDictionary>(`/system/dictionaries/${id}`, { method: "DELETE" });
  }

  async mapConfig() {
    return this.request<MapProviderConfig>("/system/map-config");
  }

  async createBackup() {
    return this.request<SystemBundle["backups"][number]>("/system/backups", { method: "POST" });
  }

  async listBackups() {
    return this.request<SystemBundle["backups"]>("/system/backups");
  }

  async restoreBackup(name: string) {
    return this.request<{ restored: string }>(`/system/backups/${encodeURIComponent(name)}/restore`, { method: "POST" });
  }

  async listBackupDrills() {
    return this.request<BackupDrill[]>("/system/backups/drills");
  }

  async runBackupDrill() {
    return this.request<BackupDrill>("/system/backups/drills", { method: "POST" });
  }

  async importLicensePackage(payload: LicensePackage | Record<string, unknown>) {
    return this.request<LicensePackage>("/system/license/import", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async issueLicense(payload: Record<string, unknown>) {
    return this.request<LicensePackage>("/system/license/issues", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async downloadLicensePackage(id: number) {
    return this.request<LicensePackageDownload>(`/system/license/packages/${id}/download`);
  }

  async renewLicensePackage(id: number, payload: Record<string, unknown>) {
    return this.request<LicensePackage>(`/system/license/packages/${id}/renew`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async revokeLicense(licenseId: string, reason: string) {
    return this.request<LicenseRevocation>("/system/license/revoke", {
      method: "POST",
      body: JSON.stringify({ licenseId, reason })
    });
  }

  async createCompany(payload: Partial<Company> & Pick<Company, "name" | "code">) {
    return this.request<Company>("/system/org/companies", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async createDepartment(payload: Partial<Department> & Pick<Department, "companyId" | "name" | "code">) {
    return this.request<Department>("/system/org/departments", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateOrgStatus(resource: "companies" | "departments", id: number, status: string) {
    return this.request<Company | Department>(`/system/org/${resource}/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async gatewayOverview() {
    return this.request<GatewayOverview>("/system/gateway");
  }

  async saveGatewayRoute(payload: Partial<GatewayRoute>) {
    return this.request<GatewayRoute>("/system/gateway", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setGatewayCanary(id: number, canaryPercent: number, canaryUpstream: string) {
    return this.request<GatewayRoute>(`/system/gateway/routes/${id}/canary`, {
      method: "POST",
      body: JSON.stringify({ canaryPercent, canaryUpstream })
    });
  }

  async setGatewayDrain(id: number, enabled: boolean, durationMs = 300000) {
    return this.request<GatewayRoute>(`/system/gateway/routes/${id}/drain`, {
      method: "POST",
      body: JSON.stringify({ enabled, durationMs })
    });
  }

  async setGatewayStatus(id: number, status: string) {
    return this.request<GatewayRoute>(`/system/gateway/routes/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteGatewayRoute(id: number) {
    return this.request<GatewayRoute>(`/system/gateway/routes/${id}`, {
      method: "DELETE"
    });
  }

  async reloadGateway() {
    return this.request<GatewayOverview["reloadPlan"]>("/system/gateway/reload", { method: "POST" });
  }

  async enrollMFA(userId: number) {
    return this.request<MFAEnrollment>(`/system/mfa/users/${userId}/enroll`, { method: "POST" });
  }

  async enableMFA(userId: number, code: string) {
    return this.request<User>(`/system/mfa/users/${userId}/enable`, {
      method: "POST",
      body: JSON.stringify({ code })
    });
  }

  async disableMFA(userId: number) {
    return this.request<User>(`/system/mfa/users/${userId}/disable`, { method: "POST" });
  }

  async saveSSOProvider(payload: Partial<OIDCProvider>) {
    return this.request<OIDCProvider>("/system/sso/providers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setSSOProviderStatus(id: number, status: string) {
    return this.request<OIDCProvider>(`/system/sso/providers/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteSSOProvider(id: number) {
    return this.request<OIDCProvider>(`/system/sso/providers/${id}`, { method: "DELETE" });
  }

  async saveSCIMProvider(payload: Partial<SCIMProvider>) {
    return this.request<SCIMProvider>("/system/scim/providers", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setSCIMProviderStatus(id: number, status: string) {
    return this.request<SCIMProvider>(`/system/scim/providers/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteSCIMProvider(id: number) {
    return this.request<SCIMProvider>(`/system/scim/providers/${id}`, { method: "DELETE" });
  }

  async saveSecurityPolicy(payload: Partial<SecurityPolicy> & Pick<SecurityPolicy, "name" | "type">) {
    return this.request<SecurityPolicy>("/system/security/policies", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async toggleSecurityPolicy(id: number, enabled: boolean) {
    return this.request<SecurityPolicy>(`/system/security/policies/${id}/toggle`, {
      method: "POST",
      body: JSON.stringify({ enabled })
    });
  }

  async deleteSecurityPolicy(id: number) {
    return this.request<SecurityPolicy>(`/system/security/policies/${id}`, { method: "DELETE" });
  }

  async saveDeviceCredential(payload: Partial<DeviceCredential> & Pick<DeviceCredential, "deviceNo"> & { deviceKey?: string }) {
    return this.request<DeviceCredential>("/system/security/device-credentials", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async setDeviceCredentialStatus(id: number, status: string) {
    return this.request<DeviceCredential>(`/system/security/device-credentials/${id}/status`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async deleteDeviceCredential(id: number) {
    return this.request<DeviceCredential>(`/system/security/device-credentials/${id}`, { method: "DELETE" });
  }

  async createFieldPolicy(payload: Partial<FieldPolicy> & Pick<FieldPolicy, "roleCode" | "resource" | "field">) {
    return this.request<FieldPolicy>("/system/field-policies", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async updateFieldPolicy(id: number, payload: Partial<FieldPolicy> & Pick<FieldPolicy, "roleCode" | "resource" | "field">) {
    return this.request<FieldPolicy>(`/system/field-policies/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    });
  }

  async deleteFieldPolicy(id: number) {
    return this.request<FieldPolicy>(`/system/field-policies/${id}`, { method: "DELETE" });
  }

  async toggleFieldPolicy(id: number, enabled: boolean) {
    return this.request<FieldPolicy>(`/system/field-policies/${id}/toggle`, {
      method: "POST",
      body: JSON.stringify({ enabled })
    });
  }

  async installPlugin(payload: Partial<PluginInfo>) {
    return this.request<PluginInfo>("/system/plugins/install", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async verifyPlugin(pluginId: string) {
    return this.request<{ plugin: PluginInfo; valid: boolean }>(`/system/plugins/${encodeURIComponent(pluginId)}/verify`, { method: "POST" });
  }

  async setPluginStatus(pluginId: string, status: string) {
    return this.request<PluginInfo>(`/system/plugins/${encodeURIComponent(pluginId)}`, {
      method: "POST",
      body: JSON.stringify({ status })
    });
  }

  async runPlugin(pluginId: string, payload: { action?: string; permission: string; input: Record<string, unknown> }) {
    return this.request<PluginRun>(`/system/plugins/${encodeURIComponent(pluginId)}/run`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async applyUpdate(id: number) {
    return this.request<UpdatePackage>(`/system/updates/${id}/apply`, { method: "POST" });
  }

  async publishUpdate(payload: Partial<UpdatePackage>) {
    return this.request<UpdatePackage>("/system/updates", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async rollbackUpdate(id: number) {
    return this.request<UpdatePackage>(`/system/updates/${id}/rollback`, { method: "POST" });
  }

  async downloadUpdate(id: number) {
    return this.request<UpdatePackageDownload>(`/system/updates/${id}/download`);
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const headers = new Headers(options.headers || {});
    headers.set("Content-Type", "application/json");
    if (!options.anonymous && this.token) {
      headers.set("Authorization", `Bearer ${this.token}`);
    }
    const response = await fetch(`${getAPIBaseURL()}${path}`, {
      ...options,
      headers,
      credentials: "include"
    });
    if (!response.ok) {
      const payload = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(payload.error || response.statusText);
    }
    return response.json();
  }
}

export const api = new APIClient();
export const eventURL = () => `${getAPIBaseURL()}/events?token=${encodeURIComponent(api.token)}`;
