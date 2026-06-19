export type LicenseInfo = {
  licenseId: string;
  customerName: string;
  watermark: string;
  expiresAt: string;
  edition: string;
  modules: string[];
  maxSites: number;
  maxVehicles: number;
  issuedAt: string;
  issuer: string;
  publicKey?: string;
  publicKeyFingerprint?: string;
  signature: string;
  lastVerifiedAt: string;
  lastVerificationState: string;
  lastVerificationError: string;
};

export type LicensePackage = LicenseInfo & {
  id: number;
  publicKey: string;
  publicKeyFingerprint: string;
  status: string;
  activatedAt: string;
};

export type LicenseIssueRecord = {
  id: number;
  issueNo: string;
  licenseId: string;
  customerName: string;
  watermark: string;
  expiresAt: string;
  edition: string;
  modules: string[];
  maxSites: number;
  maxVehicles: number;
  issuer: string;
  publicKeyFingerprint: string;
  status: string;
  issuedAt: string;
  actor: string;
};

export type LicenseRevocation = {
  id: number;
  revokeNo: string;
  licenseId: string;
  reason: string;
  status: string;
  revokedAt: string;
  actor: string;
};

export type LicensePackageDownload = {
  fileName: string;
  contentType: string;
  package: LicensePackage;
  valid: boolean;
  reason: string;
  fingerprint: string;
  downloadedAt: string;
};

export type LicenseVerification = {
  valid: boolean;
  reason: string;
  license: LicenseInfo;
  moduleCount: number;
  maxSites: number;
  maxVehicles: number;
  signatureType: string;
  fingerprint: string;
  verifiedAt: string;
};

export type LicensePortalOverview = {
  kpis: LicensePortalKPI;
  customers: LicensePortalCustomer[];
  recentEvents: LicensePortalEvent[];
  requiredModules: string[];
};

export type LicensePortalKPI = {
  totalCustomers: number;
  totalPackages: number;
  activePackages: number;
  issuedPackages: number;
  revokedPackages: number;
  expiring30Days: number;
  expiredPackages: number;
  downloadCount: number;
  moduleCoverage: number;
  riskLevel: string;
};

export type LicensePortalCustomer = {
  customerName: string;
  watermark: string;
  licenseId: string;
  edition: string;
  status: string;
  expiresAt: string;
  daysToExpire: number;
  modules: string[];
  moduleCount: number;
  moduleCoverage: number;
  maxSites: number;
  maxVehicles: number;
  packageCount: number;
  latestPackageId: number;
  latestIssuedAt: string;
  latestActivatedAt: string;
  latestDownloadAt: string;
  renewalAvailable: boolean;
  revoked: boolean;
  verificationState: string;
  verificationError: string;
  riskLevel: string;
  lastOperation: string;
  publicKeyFingerprint: string;
};

export type LicensePortalEvent = {
  action: string;
  resource: string;
  licenseId: string;
  detail: string;
  actor: string;
  ip: string;
  createdAt: string;
};

export type ProductInstance = {
  id: number;
  customerName: string;
  licenseId: string;
  watermark: string;
  edition: string;
  deploymentMode: string;
  clientVersion: string;
  serverVersion: string;
  endpoint: string;
  status: string;
  probeToken: string;
  probeEnabled: boolean;
  healthStatus: string;
  lastProbeAt: string;
  licenseExpiresAt: string;
  daysToExpire: number;
  licenseRisk: string;
  renewalAvailable: boolean;
  renewalOwner: string;
  renewalStage: string;
  alertLevel: string;
  lastHeartbeatAt: string;
  latestPackageId: number;
  createdAt: string;
  remark: string;
};

export type SystemAlert = {
  id: number;
  alertNo: string;
  instanceId: number;
  customerName: string;
  severity: string;
  source: string;
  title: string;
  message: string;
  status: string;
  groupKey: string;
  policyNo: string;
  eventCount: number;
  suppressedUntil: string;
  escalationLevel: string;
  escalatedAt: string;
  firstSeenAt: string;
  lastSeenAt: string;
  handledBy: string;
  handledAt: string;
};

export type ProductRenewalTask = {
  id: number;
  taskNo: string;
  instanceId: number;
  customerName: string;
  licenseId: string;
  stage: string;
  status: string;
  owner: string;
  amount: number;
  currency: string;
  dueDate: string;
  nextFollowAt: string;
  riskLevel: string;
  lastContactAt: string;
  closedAt: string;
  createdAt: string;
  remark: string;
};

export type ProductRenewalQuote = {
  id: number;
  quoteNo: string;
  taskId: number;
  instanceId: number;
  customerName: string;
  licenseId: string;
  amount: number;
  currency: string;
  modules: string[];
  newExpiresAt: string;
  status: string;
  preparedBy: string;
  preparedAt: string;
  approvedBy: string;
  approvedAt: string;
  remark: string;
};

export type ProductRenewalContract = {
  id: number;
  contractNo: string;
  taskId: number;
  quoteId: number;
  instanceId: number;
  customerName: string;
  licenseId: string;
  amount: number;
  currency: string;
  status: string;
  signedBy: string;
  signedAt: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductRenewalPayment = {
  id: number;
  paymentNo: string;
  taskId: number;
  contractId: number;
  instanceId: number;
  customerName: string;
  amount: number;
  currency: string;
  method: string;
  status: string;
  paidAt: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductRenewalApproval = {
  id: number;
  approvalNo: string;
  taskId: number;
  quoteId: number;
  contractId: number;
  instanceId: number;
  customerName: string;
  licenseId: string;
  approvalType: string;
  amount: number;
  currency: string;
  status: string;
  currentRole: string;
  requestedBy: string;
  requestedAt: string;
  approvedBy: string;
  approvedAt: string;
  comment: string;
};

export type ProductRenewalInvoice = {
  id: number;
  invoiceNo: string;
  taskId: number;
  contractId: number;
  paymentId: number;
  instanceId: number;
  customerName: string;
  licenseId: string;
  amount: number;
  taxRate: number;
  taxAmount: number;
  invoiceType: string;
  status: string;
  taxStatus: string;
  fileUrl: string;
  createdBy: string;
  createdAt: string;
  issuedAt: string;
  externalRequest: string;
  remark: string;
};

export type ProductRenewalESign = {
  id: number;
  signNo: string;
  taskId: number;
  contractId: number;
  instanceId: number;
  customerName: string;
  licenseId: string;
  signer: string;
  phone: string;
  channel: string;
  status: string;
  linkUrl: string;
  sentBy: string;
  sentAt: string;
  signedAt: string;
  signature: string;
  remark: string;
};

export type ProductRenewalIntegration = {
  id: number;
  integrationNo: string;
  name: string;
  code: string;
  provider: string;
  scenario: string;
  endpoint: string;
  token: string;
  secret: string;
  status: string;
  retryLimit: number;
  timeoutSeconds: number;
  lastSyncAt: string;
  lastError: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductRenewalSyncRecord = {
  id: number;
  syncNo: string;
  integrationId: number;
  integrationNo: string;
  integrationCode: string;
  provider: string;
  scenario: string;
  resourceType: string;
  resourceId: number;
  resourceNo: string;
  taskId: number;
  customerName: string;
  action: string;
  status: string;
  attemptCount: number;
  nextRetryAt: string;
  externalRequestId: string;
  externalStatus: string;
  requestPayload: string;
  responsePayload: string;
  error: string;
  createdAt: string;
  lastAttemptAt: string;
  completedAt: string;
};

export type ProductProbeReport = {
  id: number;
  reportNo: string;
  instanceId: number;
  customerName: string;
  watermark: string;
  component: string;
  clientVersion: string;
  serverVersion: string;
  status: string;
  cpuPercent: number;
  memoryPercent: number;
  diskPercent: number;
  queueBacklog: number;
  errorCount: number;
  message: string;
  reportedAt: string;
  receivedAt: string;
  sourceIp: string;
  alertRaised: boolean;
};

export type ProductTelemetryEvent = {
  id: number;
  eventNo: string;
  instanceId: number;
  customerName: string;
  watermark: string;
  source: string;
  component: string;
  severity: string;
  eventType: string;
  traceId: string;
  spanId: string;
  endpoint: string;
  durationMs: number;
  statusCode: number;
  errorMessage: string;
  message: string;
  occurredAt: string;
  receivedAt: string;
  sourceIp: string;
  alertRaised: boolean;
};

export type ProductMonitoringIntegration = {
  id: number;
  integrationNo: string;
  name: string;
  code: string;
  provider: string;
  endpoint: string;
  token: string;
  status: string;
  lastEventAt: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductAlertRule = {
  id: number;
  ruleNo: string;
  name: string;
  source: string;
  component: string;
  metric: string;
  operator: string;
  threshold: number;
  severity: string;
  status: string;
  notifyChannels: string[];
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductAlertPolicy = {
  id: number;
  policyNo: string;
  name: string;
  source: string;
  component: string;
  metric: string;
  severity: string;
  aggregateWindowMinutes: number;
  suppressMinutes: number;
  escalateAfterMinutes: number;
  escalateTo: string;
  notifyChannels: string[];
  status: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductAlertChannel = {
  id: number;
  channelNo: string;
  name: string;
  code: string;
  type: string;
  endpoint: string;
  token: string;
  secret: string;
  status: string;
  retryLimit: number;
  timeoutSeconds: number;
  lastDeliveredAt: string;
  lastError: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

export type ProductAlertNotification = {
  id: number;
  notificationNo: string;
  alertId: number;
  alertNo: string;
  policyId: number;
  policyNo: string;
  instanceId: number;
  customerName: string;
  action: string;
  severity: string;
  channelId: number;
  channelNo: string;
  channel: string;
  target: string;
  endpoint: string;
  status: string;
  attemptCount: number;
  nextRetryAt: string;
  message: string;
  error: string;
  createdAt: string;
  deliveredAt: string;
};

export type ProductMonitoringEvent = {
  id: number;
  eventNo: string;
  integrationId: number;
  integrationName: string;
  provider: string;
  instanceId: number;
  customerName: string;
  watermark: string;
  source: string;
  component: string;
  metric: string;
  value: number;
  severity: string;
  status: string;
  title: string;
  message: string;
  labels: Record<string, string>;
  occurredAt: string;
  receivedAt: string;
  sourceIp: string;
  alertRaised: boolean;
  matchedRuleNo: string;
};

export type ProductUpdateRolloutItem = {
  id: number;
  instanceId: number;
  customerName: string;
  fromVersion: string;
  toVersion: string;
  status: string;
  message: string;
  startedAt: string;
  appliedAt: string;
  rolledBackAt: string;
};

export type ProductUpdateRollout = {
  id: number;
  rolloutNo: string;
  updateId: number;
  version: string;
  component: string;
  strategy: string;
  status: string;
  totalTargets: number;
  appliedTargets: number;
  failedTargets: number;
  createdBy: string;
  createdAt: string;
  startedAt: string;
  completedAt: string;
  remark: string;
  items: ProductUpdateRolloutItem[];
};

export type ProductUpdateExecutionStep = {
  id: number;
  name: string;
  status: string;
  message: string;
  startedAt: string;
  completedAt: string;
  durationMs: number;
};

export type ProductUpdateExecution = {
  id: number;
  executionNo: string;
  rolloutId: number;
  rolloutNo: string;
  updateId: number;
  instanceId: number;
  customerName: string;
  component: string;
  version: string;
  action: string;
  status: string;
  artifactFileName: string;
  checksumVerified: boolean;
  dryRun: boolean;
  startedBy: string;
  startedAt: string;
  completedAt: string;
  durationMs: number;
  precheckSummary: string;
  result: string;
  error: string;
  steps: ProductUpdateExecutionStep[];
};

export type ProductSystemUpdateTaskLog = {
  id: number;
  status: string;
  progress: number;
  step: string;
  message: string;
  createdAt: string;
};

export type ProductSystemUpdateTask = {
  id: number;
  taskNo: string;
  executionId: number;
  executionNo: string;
  rolloutId: number;
  rolloutNo: string;
  rolloutItemId: number;
  updateId: number;
  instanceId: number;
  customerName: string;
  watermark: string;
  component: string;
  version: string;
  fromVersion: string;
  action: string;
  status: string;
  progress: number;
  artifactFileName: string;
  checksum: string;
  signature: string;
  downloadUrl: string;
  updaterTokenHint: string;
  createdBy: string;
  createdAt: string;
  claimedAt: string;
  startedAt: string;
  completedAt: string;
  lastHeartbeatAt: string;
  result: string;
  error: string;
  remark: string;
  logs: ProductSystemUpdateTaskLog[];
};

export type ProductOpsOverview = {
  kpis: {
    customers: number;
    onlineInstances: number;
    degradedInstances: number;
    expiringLicenses: number;
    openAlerts: number;
    criticalAlerts: number;
    openRenewals: number;
    overdueRenewals: number;
    pendingRenewalQuotes: number;
    pendingRenewalContracts: number;
    paidRenewalAmount: number;
    pendingRenewalApprovals: number;
    issuedRenewalInvoices: number;
    pendingRenewalESigns: number;
    activeRenewalIntegrations: number;
    renewalSyncRecords: number;
    failedRenewalSyncRecords: number;
    pendingRenewalSyncRecords: number;
    probeReports: number;
    unhealthyProbes: number;
    telemetryEvents: number;
    criticalTelemetryEvents: number;
    monitoringIntegrations: number;
    activeAlertRules: number;
    monitoringEvents: number;
    monitoringAlerts: number;
    activeAlertPolicies: number;
    suppressedAlerts: number;
    escalatedAlerts: number;
    activeAlertChannels: number;
    alertNotifications: number;
    failedAlertNotifications: number;
    pendingAlertNotifications: number;
    activeRollouts: number;
    failedRolloutItems: number;
    updateExecutions: number;
    failedUpdateExecutions: number;
    systemUpdateTasks: number;
    runningSystemUpdateTasks: number;
    failedSystemUpdateTasks: number;
    clientUpdatePackages: number;
    serverUpdatePackages: number;
    availableUpdates: number;
  };
  instances: ProductInstance[];
  alerts: SystemAlert[];
  renewalTasks: ProductRenewalTask[];
  renewalQuotes: ProductRenewalQuote[];
  renewalContracts: ProductRenewalContract[];
  renewalPayments: ProductRenewalPayment[];
  renewalApprovals: ProductRenewalApproval[];
  renewalInvoices: ProductRenewalInvoice[];
  renewalESigns: ProductRenewalESign[];
  renewalIntegrations: ProductRenewalIntegration[];
  renewalSyncRecords: ProductRenewalSyncRecord[];
  probeReports: ProductProbeReport[];
  telemetryEvents: ProductTelemetryEvent[];
  monitoringIntegrations: ProductMonitoringIntegration[];
  alertRules: ProductAlertRule[];
  monitoringEvents: ProductMonitoringEvent[];
  alertPolicies: ProductAlertPolicy[];
  alertChannels: ProductAlertChannel[];
  alertNotifications: ProductAlertNotification[];
  updateRollouts: ProductUpdateRollout[];
  updateExecutions: ProductUpdateExecution[];
  systemUpdateTasks: ProductSystemUpdateTask[];
  updates: UpdatePackage[];
  licensePortal: LicensePortalOverview;
  recentEvents: LicensePortalEvent[];
  runtime: Record<string, unknown>;
};

export type User = {
  id: number;
  companyId: number;
  siteId: number;
  customerId: number;
  driverId: number;
  username: string;
  displayName: string;
  roleCode: string;
  status: string;
  mfaEnabled: boolean;
};

export type LoginResult = {
  token?: string;
  user?: User;
  createdAt?: string;
  mfaRequired?: boolean;
  username?: string;
  displayName?: string;
};

export type MFAEnrollment = {
  user: User;
  secret: string;
  otpauthUrl: string;
};

export type OIDCProvider = {
  id: number;
  name: string;
  code: string;
  issuer: string;
  clientId: string;
  clientSecret?: string;
  authUrl: string;
  tokenUrl: string;
  userInfoUrl: string;
  jwksUrl: string;
  redirectUri: string;
  scopes: string[];
  usernameClaim: string;
  displayNameClaim: string;
  roleCode: string;
  companyId: number;
  siteId: number;
  autoProvision: boolean;
  status: string;
  lastLoginAt: string;
};

export type SCIMProvider = {
  id: number;
  name: string;
  code: string;
  bearerToken?: string;
  companyId: number;
  siteId: number;
  defaultRoleCode: string;
  status: string;
  lastSyncAt: string;
  createdAt: string;
};

export type SCIMProvisioningEvent = {
  id: number;
  eventNo: string;
  providerId: number;
  providerCode: string;
  userId: number;
  username: string;
  action: string;
  status: string;
  detail: string;
  createdAt: string;
  actor: string;
  ip: string;
};

export type OIDCLoginStart = {
  provider: OIDCProvider;
  authorizationUrl: string;
  state: string;
  nonce: string;
  expiresAt: string;
};

export type Company = {
  id: number;
  name: string;
  code: string;
  status: string;
};

export type Department = {
  id: number;
  companyId: number;
  name: string;
  code: string;
  parentId: number;
  status: string;
};

export type ModuleInfo = {
  code: string;
  name: string;
  area: string;
  description: string;
  enabled: boolean;
  hotPlug: boolean;
  version: string;
};

export type PluginInfo = {
  id: string;
  name: string;
  type: string;
  status: string;
  version: string;
  checksum: string;
  signature: string;
  permissions: string[];
  runtime: string;
  entrypoint: string;
  sandbox: {
    runtime: string;
    timeoutMs: number;
    network: boolean;
    filesystem: string;
    maxMemoryMb: number;
  };
  lastRunAt: string;
};

export type PluginRun = {
  id: number;
  runNo: string;
  pluginId: string;
  pluginName: string;
  runtime: string;
  action: string;
  permission: string;
  status: string;
  input: string;
  output: string;
  error: string;
  actor: string;
  startedAt: string;
  completedAt: string;
  durationMs: number;
};

export type UpdatePackage = {
  id: number;
  version: string;
  component: string;
  channel: string;
  status: string;
  packageType: string;
  baseVersion: string;
  deltaAlgorithm: string;
  checksum: string;
  signature: string;
  signaturePublicKey: string;
  signatureKeyFingerprint: string;
  fileName: string;
  sizeBytes: number;
  artifactFileName: string;
  artifactContentType: string;
  artifactContentBase64?: string;
  artifactSha256: string;
  artifactSizeBytes: number;
  baseArtifactSha256: string;
  targetArtifactSha256: string;
  rollbackVersion: string;
  publishedBy: string;
  publishedAt: string;
  downloadCount: number;
  lastDownloadedAt: string;
  appliedBy: string;
  appliedAt: string;
  createdAt: string;
  remark: string;
};

export type UpdatePackageDownload = {
  fileName: string;
  contentType: string;
  verified: boolean;
  generatedAt: string;
  artifactFileName: string;
  artifactContentType: string;
  artifactSizeBytes: number;
  artifactSha256: string;
  artifactContentBase64: string;
  manifest: Record<string, string>;
  package: UpdatePackage;
};

export type Site = {
  id: number;
  companyId: number;
  name: string;
  code: string;
  address: string;
  longitude: number;
  latitude: number;
  status: string;
};

export type Warehouse = {
  id: number;
  siteId: number;
  name: string;
  code: string;
  type: string;
  status: string;
};

export type Silo = {
  id: number;
  warehouseId: number;
  materialId: number;
  name: string;
  code: string;
  capacity: number;
  currentQty: number;
  status: string;
};

export type Customer = {
  id: number;
  companyId: number;
  name: string;
  contact: string;
  phone: string;
  creditLimit: number;
  receivable: number;
  paymentTerm: number;
  status: string;
};

export type CustomerContact = {
  id: number;
  customerId: number;
  name: string;
  phone: string;
  role: string;
  isDefault: boolean;
  status: string;
};

export type CustomerBlacklist = {
  id: number;
  customerId: number;
  customerName: string;
  reason: string;
  scope: string;
  severity: string;
  blockOrders: boolean;
  blockDispatch: boolean;
  status: string;
  createdAt: string;
  releasedAt: string;
  actor: string;
};

export type CustomerProfile = {
  id: number;
  customerId: number;
  customerName: string;
  grade: string;
  riskLevel: string;
  creditScore: number;
  tags: string[];
  status: string;
  updatedAt: string;
  actor: string;
};

export type CustomerComplaint = {
  id: number;
  complaintNo: string;
  customerId: number;
  projectId: number;
  title: string;
  content: string;
  level: string;
  status: string;
  owner: string;
  slaHours: number;
  dueAt: string;
  slaStatus: string;
  overdueHours: number;
  createdAt: string;
  closedAt: string;
  resolution: string;
};

export type PortalOverview = {
  dispatches: DispatchOrder[];
  orders: SalesOrder[];
  statements: Statement[];
  invoices: SalesInvoice[];
  signs: DeliverySign[];
  signLinks: DeliverySignLink[];
  signAttachments: DeliverySignAttachment[];
  alarms: VehicleAlarm[];
  complaints: CustomerComplaint[];
};

export type ContractItem = {
  productId: number;
  unit: string;
  quantity: number;
  unitPrice: number;
};

export type Contract = {
  id: number;
  customerId: number;
  projectId: number;
  parentId: number;
  contractNo: string;
  version: number;
  name: string;
  validFrom: string;
  validTo: string;
  creditPolicy: string;
  totalAmount: number;
  usedAmount: number;
  changeReason: string;
  submittedAt: string;
  approvedAt: string;
  approvedBy: string;
  status: string;
  items: ContractItem[];
};

export type ContractAttachment = {
  id: number;
  contractId: number;
  customerId: number;
  fileName: string;
  fileType: string;
  url: string;
  checksum: string;
  status: string;
  uploadedBy: string;
  uploadedAt: string;
};

export type Project = {
  id: number;
  customerId: number;
  name: string;
  address: string;
  contact: string;
  phone: string;
  longitude: number;
  latitude: number;
  status: string;
};

export type Product = {
  id: number;
  line: string;
  name: string;
  spec: string;
  unit: string;
  basePrice: number;
  costPrice: number;
  requiresMix: boolean;
  status: string;
};

export type PricePolicy = {
  id: number;
  customerId: number;
  projectId: number;
  productId: number;
  customerGrade: string;
  region: string;
  minQuantity: number;
  maxQuantity: number;
  floorPrice: number;
  salePrice: number;
  promotionName: string;
  promotionType: string;
  promotionValue: number;
  priority: number;
  taxRateId: number;
  effectiveFrom: string;
  effectiveTo: string;
  status: string;
};

export type TaxRate = {
  id: number;
  name: string;
  rate: number;
  scope: string;
  status: string;
};

export type MasterDataExport = {
  resource: string;
  count: number;
  fields: string[];
  rows: Record<string, unknown>[];
};

export type MasterDataImportResult = {
  resource: string;
  mode: string;
  created: number;
  updated: number;
  errors: string[];
};

export type PricingQuote = {
  customerId: number;
  projectId: number;
  productId: number;
  policyId: number;
  customerGrade: string;
  region: string;
  minQuantity: number;
  maxQuantity: number;
  source: string;
  listPrice: number;
  unitPrice: number;
  floorPrice: number;
  promotionName: string;
  promotionType: string;
  promotionValue: number;
  promotionAmount: number;
  taxRateId: number;
  taxRateName: string;
  taxRate: number;
  belowFloor: boolean;
  approvalRequired: boolean;
  reason: string;
};

export type Material = {
 id: number;
 name: string;
 spec: string;
 unit: string;
 safeStock: number;
 status: string;
};

export type Carrier = {
  id: number;
  name: string;
  contact: string;
  phone: string;
  settleMode: string;
  status: string;
};

export type Vehicle = {
  id: number;
  plateNo: string;
  vehicleType: string;
  capacity: string;
  carrier: string;
  siteId: number;
  driverId: number;
  onlineStatus: string;
  businessStatus: string;
  certExpiresAt: string;
  status: string;
};

export type Driver = {
  id: number;
  name: string;
  phone: string;
  licenseNo: string;
  licenseExpire: string;
  status: string;
};

export type SalesOrder = {
  id: number;
  orderNo: string;
  customerId: number;
  projectId: number;
  productId: number;
  siteId: number;
  productLine: string;
  planQuantity: number;
  unit: string;
  unitPrice: number;
  totalAmount: number;
  lines?: SalesOrderLine[];
  planTime: string;
  receiveAddress: string;
  contact: string;
  phone: string;
  settlementMode: string;
  transportMode: string;
  strengthGrade: string;
  slump: string;
  pouringPart: string;
  pumpMode: string;
  dispatchedQty: number;
  signedQty: number;
  status: string;
  riskFlag: string;
  createdAt: string;
};

export type SalesOrderLine = {
  id: number;
  seq: number;
  productId: number;
  productLine: string;
  productName: string;
  strengthGrade: string;
  slump: string;
  pouringPart: string;
  quantity: number;
  unit: string;
  unitPrice: number;
  floorPrice: number;
  taxRate: number;
  amount: number;
  priceSource: string;
  riskFlag: string;
};

export type ProductionPlan = {
  id: number;
  planNo: string;
  orderId: number;
  siteId: number;
  productId: number;
  planQuantity: number;
  producedQty: number;
  planDate: string;
  shift: string;
  capacityStatus: string;
  inventoryStatus: string;
  recipeStatus: string;
  status: string;
};

export type MixDesignMaterial = {
  materialId: number;
  dosage: number;
  unit: string;
};

export type MixDesign = {
  id: number;
  productId: number;
  siteId: number;
  parentId: number;
  code: string;
  version: string;
  strengthGrade: string;
  slump: string;
  scope: string;
  status: string;
  isCurrent: boolean;
  effectiveFrom: string;
  effectiveTo: string;
  approvedBy: string;
  approvedAt: string;
  retiredAt: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  materials: MixDesignMaterial[];
};

export type MixDesignTrialRun = {
  id: number;
  trialNo: string;
  mixDesignId: number;
  productId: number;
  siteId: number;
  targetStrength: string;
  slump: string;
  water: number;
  sandRate: number;
  admixtureRate: number;
  strength7d: number;
  strength28d: number;
  result: string;
  conclusion: string;
  tester: string;
  testedAt: string;
  createdAt: string;
  remark: string;
};

export type LaboratorySample = {
  id: number;
  sampleNo: string;
  sourceType: string;
  sourceId: number;
  siteId: number;
  productId: number;
  materialId: number;
  mixDesignId: number;
  batchId: number;
  inspectionId: number;
  rawInspectionId: number;
  sampleType: string;
  status: string;
  result: string;
  plannedTestAt: string;
  collectedAt: string;
  createdBy: string;
  remark: string;
};

export type LaboratoryTestRecord = {
  id: number;
  testNo: string;
  sampleId: number;
  equipmentId: number;
  siteId: number;
  testType: string;
  metric: string;
  value: number;
  unit: string;
  result: string;
  status: string;
  tester: string;
  testedAt: string;
  reviewer: string;
  reviewedAt: string;
  remark: string;
};

export type LaboratoryEquipment = {
  id: number;
  equipmentNo: string;
  name: string;
  siteId: number;
  model: string;
  serialNo: string;
  status: string;
  calibrationCycleDays: number;
  lastCalibrationAt: string;
  nextCalibrationAt: string;
  createdAt: string;
  remark: string;
};

export type LaboratoryCalibration = {
  id: number;
  calibrationNo: string;
  equipmentId: number;
  siteId: number;
  result: string;
  calibratedAt: string;
  nextDueAt: string;
  certificateNo: string;
  agency: string;
  operator: string;
  remark: string;
};

export type QualityException = {
  id: number;
  exceptionNo: string;
  sourceType: string;
  sourceId: number;
  siteId: number;
  severity: string;
  title: string;
  description: string;
  status: string;
  responsible: string;
  rootCause: string;
  correctiveAction: string;
  createdAt: string;
  handledAt: string;
  closedBy: string;
};

export type LaboratoryKPI = {
  mixDesigns: number;
  currentMixDesigns: number;
  pendingMixDesigns: number;
  trialRuns: number;
  samples: number;
  pendingSamples: number;
  tests: number;
  pendingReviews: number;
  equipments: number;
  calibrationDue: number;
  calibrationOverdue: number;
  openExceptions: number;
  passRate: number;
};

export type ProductionTask = {
  id: number;
  taskNo: string;
  planId: number;
  orderId: number;
  siteId: number;
  productId: number;
  mixDesignId: number;
  planQty: number;
  producedQty: number;
  status: string;
  startedAt: string;
  completedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type ProductionBatch = {
  id: number;
  batchNo: string;
  taskId: number;
  planId: number;
  orderId: number;
  siteId: number;
  productId: number;
  mixDesignId: number;
  quantity: number;
  plantCode: string;
  operator: string;
  qualityStatus: string;
  status: string;
  startedAt: string;
  completedAt: string;
};

export type ProductionDailyReport = {
  id: number;
  reportNo: string;
  siteId: number;
  reportDate: string;
  plannedQty: number;
  producedQty: number;
  batchCount: number;
  materialCost: number;
  qualityPassed: number;
  qualityPending: number;
  status: string;
  generatedAt: string;
};

export type QualityInspection = {
  id: number;
  inspectionNo: string;
  batchId: number;
  batchNo: string;
  siteId: number;
  productId: number;
  mixDesignId: number;
  inspector: string;
  slump: string;
  temperature: number;
  sampleCount: number;
  result: string;
  status: string;
  remark: string;
  createdAt: string;
  completedAt: string;
};

export type QualitySample = {
  id: number;
  sampleNo: string;
  inspectionId: number;
  batchId: number;
  sampleType: string;
  ageDays: number;
  plannedTestAt: string;
  testedAt: string;
  strength: number;
  result: string;
  status: string;
  remark: string;
};

export type RawMaterialInspection = {
  id: number;
  inspectionNo: string;
  receiptId: number;
  receiptNo: string;
  siteId: number;
  supplierId: number;
  materialId: number;
  inspector: string;
  sampleNo: string;
  moisture: number;
  mudContent: number;
  fineness: string;
  result: string;
  status: string;
  remark: string;
  createdAt: string;
  completedAt: string;
};

export type DispatchOrder = {
  id: number;
  dispatchNo: string;
  orderId: number;
  vehicleId: number;
  driverId: number;
  siteId: number;
  projectId: number;
  lineId: number;
  lineSeq: number;
  productId: number;
  productName: string;
  planQuantity: number;
  loadedQty: number;
  signedQty: number;
  queueNo: string;
  eta: string;
  status: string;
  exception: string;
  createdAt: string;
  updatedAt: string;
};

export type DispatchSchedule = {
  id: number;
  scheduleNo: string;
  siteId: number;
  vehicleId: number;
  driverId: number;
  carrierId: number;
  shiftDate: string;
  shift: string;
  capacityQty: number;
  assignedQty: number;
  status: string;
  createdAt: string;
  updatedAt: string;
};

export type DispatchCenterOverview = {
  kpis: DispatchCenterKPI;
  siteProgress: DispatchCenterSiteProgress[];
  vehicleQueue: DispatchCenterQueueItem[];
  productionTasks: DispatchCenterProductionTask[];
  availableVehicles: DispatchCenterVehicle[];
  latestLocations: LatestLocation[];
};

export type DispatchCenterKPI = {
  totalVehicles: number;
  onlineVehicles: number;
  idleVehicles: number;
  queueVehicles: number;
  loadingVehicles: number;
  inTransitVehicles: number;
  arrivedVehicles: number;
  activeDispatches: number;
  openSupplyOrders: number;
  activeProductionTasks: number;
  vehicleOnlineRate: number;
};

export type DispatchCenterSiteProgress = {
  orderId: number;
  orderNo: string;
  customerId: number;
  customerName: string;
  projectId: number;
  projectName: string;
  siteId: number;
  siteName: string;
  productId: number;
  productName: string;
  unit: string;
  planQuantity: number;
  producedQty: number;
  dispatchedQty: number;
  loadedQty: number;
  signedQty: number;
  remainingQty: number;
  producedPercent: number;
  dispatchedPercent: number;
  loadedPercent: number;
  signedPercent: number;
  activeDispatches: number;
  queueVehicles: number;
  inTransitVehicles: number;
  nextEta: string;
  status: string;
};

export type DispatchCenterQueueItem = {
  dispatchId: number;
  dispatchNo: string;
  orderId: number;
  orderNo: string;
  customerName: string;
  projectId: number;
  projectName: string;
  siteId: number;
  siteName: string;
  productName: string;
  vehicleId: number;
  plateNo: string;
  driverId: number;
  driverName: string;
  queueNo: string;
  eta: string;
  plannedEta: string;
  etaSource: string;
  etaMinutes: number;
  etaDistanceKm: number;
  etaConfidence: string;
  etaTarget: string;
  etaSpeedKph: number;
  status: string;
  planQuantity: number;
  loadedQty: number;
  signedQty: number;
  onlineStatus: string;
  businessStatus: string;
  lastLocationAt: string;
  updatedAt: string;
};

export type DispatchCenterProductionTask = {
  taskId: number;
  taskNo: string;
  planId: number;
  planNo: string;
  orderId: number;
  orderNo: string;
  customerName: string;
  projectId: number;
  projectName: string;
  siteId: number;
  siteName: string;
  productId: number;
  productName: string;
  mixDesignCode: string;
  planQty: number;
  producedQty: number;
  remainingQty: number;
  progress: number;
  status: string;
  startedAt: string;
  updatedAt: string;
};

export type DispatchCenterVehicle = {
  vehicleId: number;
  plateNo: string;
  vehicleType: string;
  capacity: string;
  carrier: string;
  siteId: number;
  siteName: string;
  driverId: number;
  driverName: string;
  onlineStatus: string;
  businessStatus: string;
  scheduleNo: string;
  scheduleRemaining: number;
  lastLocationAt: string;
};

export type ScaleTicket = {
  id: number;
  ticketNo: string;
  ticketType: string;
  dispatchId: number;
  orderId: number;
  siteId: number;
  vehicleId: number;
  plateNo: string;
  grossWeight: number;
  tareWeight: number;
  netWeight: number;
  unit: string;
  snapshotUrl: string;
  printCount: number;
  signStatus: string;
  settlementStatus: string;
  status: string;
  createdAt: string;
  receiptId: number;
  supplierId: number;
  materialId: number;
  transferId: number;
  relatedTicketId: number;
  remark: string;
};

export type ScaleWeightRecord = {
  id: number;
  deviceId: number;
  ticketId: number;
  plateNo: string;
  weight: number;
  weightType: string;
  snapshotUrl: string;
  createdAt: string;
};

export type ScaleDeviceEvent = {
  id: number;
  eventNo: string;
  deviceId: number;
  deviceCode: string;
  ticketId: number;
  plateNo: string;
  recognizedPlateNo: string;
  weight: number;
  weightType: string;
  stable: boolean;
  snapshotUrl: string;
  cheatFlag: boolean;
  cheatReason: string;
  status: string;
  receivedAt: string;
};

export type TicketPrintLog = {
  id: number;
  ticketId: number;
  printedBy: string;
  printedAt: string;
};

export type TicketVoidLog = {
  id: number;
  ticketId: number;
  reason: string;
  approvedBy: string;
  status: string;
  createdAt: string;
};

export type ApprovalTaskAction = {
  seq: number;
  step: number;
  roleCode: string;
  action: string;
  actor: string;
  comment: string;
  actedAt: string;
};

export type ApprovalTask = {
  id: number;
  taskNo: string;
  flowCode: string;
  flowName: string;
  resource: string;
  resourceId: number;
  resourceNo: string;
  title: string;
  applicant: string;
  currentStep: number;
  currentRole: string;
  status: string;
  reason: string;
  createdAt: string;
  updatedAt: string;
  actions: ApprovalTaskAction[];
};

export type ApprovalStep = {
  seq: number;
  roleCode: string;
  action: string;
};

export type ApprovalFlow = {
  id: number;
  code: string;
  name: string;
  resource: string;
  steps: ApprovalStep[];
  status: string;
};

export type DeliverySign = {
  id: number;
  signNo: string;
  dispatchId: number;
  linkId: number;
  ticketId: number;
  orderId: number;
  lineId: number;
  lineSeq: number;
  productId: number;
  productName: string;
  customerId: number;
  projectId: number;
  signer: string;
  phone: string;
  signedQty: number;
  longitude: number;
  latitude: number;
  photo: string;
  signature: string;
  remark: string;
  signedAt: string;
};

export type DeliverySignLink = {
  id: number;
  linkNo: string;
  dispatchId: number;
  ticketId: number;
  orderId: number;
  lineId: number;
  lineSeq: number;
  productId: number;
  productName: string;
  customerId: number;
  projectId: number;
  channel: string;
  phone: string;
  token: string;
  url: string;
  qrCode: string;
  status: string;
  sentAt: string;
  expiresAt: string;
  usedAt: string;
  createdBy: string;
  createdAt: string;
};

export type DeliverySignAttachment = {
  id: number;
  signId: number;
  dispatchId: number;
  ticketId: number;
  fileName: string;
  fileType: string;
  url: string;
  checksum: string;
  uploadedBy: string;
  uploadedAt: string;
};

export type PublicDeliverySignDetail = {
  link: DeliverySignLink;
  dispatch: DispatchOrder;
  ticket: ScaleTicket;
  order: SalesOrder;
  customer: string;
  project: string;
  product: string;
  plateNo: string;
  attachments: DeliverySignAttachment[];
};

export type Statement = {
  id: number;
  statementNo: string;
  customerId: number;
  projectId: number;
  period: string;
  totalQty: number;
  totalAmount: number;
  status: string;
  confirmedBy: string;
  confirmedAt: string;
};

export type PurchaseRequest = {
  id: number;
  requestNo: string;
  siteId: number;
  materialId: number;
  quantity: number;
  unit: string;
  requiredAt: string;
  status: string;
  createdAt: string;
};

export type PurchaseOrder = {
  id: number;
  orderNo: string;
  requestId: number;
  supplierId: number;
  materialId: number;
  quantity: number;
  unitPrice: number;
  unit: string;
  status: string;
  createdAt: string;
};

export type RawMaterialReceipt = {
  id: number;
  receiptNo: string;
  purchaseOrderId: number;
  supplierId: number;
  ticketId: number;
  siteId: number;
  materialId: number;
  plateNo: string;
  grossWeight: number;
  tareWeight: number;
  netWeight: number;
  qualityStatus: string;
  status: string;
  createdAt: string;
};

export type InventoryFlow = {
  id: number;
  flowNo: string;
  siteId: number;
  materialId: number;
  sourceType: string;
  sourceId: number;
  direction: string;
  quantity: number;
  balanceQty: number;
  remark: string;
  createdAt: string;
};

export type InventoryItem = {
  id: number;
  siteId: number;
  warehouse: string;
  silo: string;
  materialId: number;
  batchNo: string;
  rawReceiptId: number;
  supplierId: number;
  quantity: number;
  unit: string;
  qualityStatus: string;
  availableStatus: string;
  updatedAt: string;
};

export type InventoryTransfer = {
  id: number;
  transferNo: string;
  fromSiteId: number;
  toSiteId: number;
  materialId: number;
  quantity: number;
  unit: string;
  status: string;
  remark: string;
  createdAt: string;
  completedAt: string;
};

export type InventoryStocktake = {
  id: number;
  stocktakeNo: string;
  siteId: number;
  materialId: number;
  bookQty: number;
  actualQty: number;
  diffQty: number;
  unit: string;
  status: string;
  remark: string;
  createdAt: string;
  reviewedAt: string;
};

export type InventoryBatchTrace = {
  id: number;
  traceNo: string;
  productionBatchId: number;
  productionBatchNo: string;
  rawReceiptId: number;
  inventoryItemId: number;
  siteId: number;
  materialId: number;
  supplierId: number;
  batchNo: string;
  warehouse: string;
  silo: string;
  quantity: number;
  unit: string;
  createdAt: string;
};

export type SalesInvoice = {
  id: number;
  invoiceNo: string;
  statementId: number;
  customerId: number;
  amount: number;
  taxRate: number;
  taxAmount: number;
  taxControlNo: string;
  taxStatus: string;
  fileUrl: string;
  downloadedAt: string;
  status: string;
  issuedAt: string;
  invoiceType: string;
  invoiceCategory: string;
  originalInvoiceId: number;
  redLetterInfoId: number;
  redLetterInfoNo: string;
  redReason: string;
  redAt: string;
};

export type RedLetterInfo = {
  id: number;
  infoNo: string;
  originalInvoiceId: number;
  originalInvoiceNo: string;
  redInvoiceId: number;
  customerId: number;
  amount: number;
  taxAmount: number;
  reason: string;
  applicant: string;
  status: string;
  taxControlNo: string;
  requestedAt: string;
  approvedBy: string;
  approvedAt: string;
  usedAt: string;
  remark: string;
};

export type TaxGatewaySubmission = {
  id: number;
  submissionNo: string;
  invoiceId: number;
  invoiceNo: string;
  action: string;
  provider: string;
  endpoint: string;
  requestId: string;
  status: string;
  taxControlNo: string;
  fileUrl: string;
  error: string;
  attempt: number;
  durationMs: number;
  submittedAt: string;
  completedAt: string;
  actor: string;
};

export type InvoiceDownload = {
  invoice: SalesInvoice;
  fileName: string;
  url: string;
  downloadedAt: string;
};

export type Receivable = {
  id: number;
  billNo: string;
  customerId: number;
  statementId: number;
  invoiceId: number;
  amount: number;
  receivedAmount: number;
  dueDate: string;
  status: string;
  createdAt: string;
};

export type Receipt = {
  id: number;
  receiptNo: string;
  receivableId: number;
  customerId: number;
  amount: number;
  method: string;
  status: string;
  receivedAt: string;
};

export type PaymentPlan = {
  id: number;
  planNo: string;
  receivableId: number;
  customerId: number;
  amount: number;
  dueDate: string;
  method: string;
  status: string;
  createdAt: string;
  settledAt: string;
  remark: string;
};

export type ReceivableAgingBucket = {
  bucket: string;
  label: string;
  count: number;
  amount: number;
  overdueAmount: number;
};

export type CollectionTask = {
  id: number;
  taskNo: string;
  receivableId: number;
  customerId: number;
  customerName: string;
  amount: number;
  dueDate: string;
  overdueDays: number;
  level: string;
  channel: string;
  status: string;
  message: string;
  templateId: number;
  sendCount: number;
  lastSentAt: string;
  generatedAt: string;
  handledBy: string;
  handledAt: string;
  remark: string;
};

export type CollectionTemplate = {
  id: number;
  code: string;
  name: string;
  level: string;
  channel: string;
  content: string;
  enabled: boolean;
  updatedAt: string;
};

export type CollectionDispatch = {
  id: number;
  dispatchNo: string;
  taskId: number;
  templateId: number;
  customerId: number;
  channel: string;
  target: string;
  content: string;
  endpoint: string;
  providerRequestId: string;
  providerMessageId: string;
  status: string;
  error: string;
  sentAt: string;
  callbackAt: string;
  actor: string;
};

export type TransportSettlement = {
  id: number;
  settlementNo: string;
  carrierId: number;
  period: string;
  tripCount: number;
  amount: number;
  status: string;
};

export type TransportSettlementItem = {
  id: number;
  settlementId: number;
  dispatchId: number;
  dispatchNo: string;
  carrierId: number;
  vehicleId: number;
  driverId: number;
  quantity: number;
  amount: number;
  status: string;
  createdAt: string;
};

export type Payable = {
  id: number;
  billNo: string;
  supplierId: number;
  supplierStatementId: number;
  amount: number;
  paidAmount: number;
  dueDate: string;
  status: string;
};

export type SupplierStatement = {
  id: number;
  statementNo: string;
  supplierId: number;
  period: string;
  amount: number;
  status: string;
  approvedBy: string;
  approvedAt: string;
};

export type ProjectProfit = {
  id: number;
  projectId: number;
  revenue: number;
  cost: number;
  profit: number;
  margin: number;
  period: string;
};

export type LatestLocation = {
  vehicleId: number;
  plateNo: string;
  longitude: number;
  latitude: number;
  speed: number;
  direction: number;
  onlineStatus: string;
  transportStatus: string;
  lastLocationTime: string;
  currentOrderId: number;
  currentProjectId: number;
  currentSiteId: number;
  currentCustomerId: number;
};

export type TrackStopPoint = {
  longitude: number;
  latitude: number;
  startTime: string;
  endTime: string;
  durationMinutes: number;
  address: string;
};

export type VehicleLocationEvent = {
  id: number;
  vehicleId: number;
  plateNo: string;
  driverId: number;
  dispatchId: number;
  deviceId: string;
  sourceType: string;
  longitude: number;
  latitude: number;
  speed: number;
  direction: number;
  mileage: number;
  onlineStatus: string;
  address: string;
  isAbnormal: boolean;
  abnormalType: string;
  locationTime: string;
  receiveTime: string;
};

export type LocationReportPayload = {
  deviceNo?: string;
  plateNo: string;
  longitude: number;
  latitude: number;
  speed?: number;
  direction?: number;
  mileage?: number;
  accStatus?: number;
  locationTime?: string;
  sourceType?: string;
};

export type LocationBatchReportResponse = {
  total: number;
  accepted: number;
  rejected: number;
  results: Array<{
    index: number;
    status: string;
    error?: string;
    location?: VehicleLocationEvent;
  }>;
};

export type TrackReplay = {
  vehicleId: number;
  plateNo: string;
  startTime: string;
  endTime: string;
  distanceKm: number;
  durationMinutes: number;
  averageSpeed: number;
  maxSpeed: number;
  stopCount: number;
  points: VehicleLocationEvent[];
  compressedPoints: VehicleLocationEvent[];
  compression: TrackCompressionSummary;
  stops: TrackStopPoint[];
  fenceEvents: GeoFenceEvent[];
  tickets: ScaleTicket[];
  signs: DeliverySign[];
  exportName: string;
};

export type TrackCompressionSummary = {
  algorithm: string;
  toleranceMeters: number;
  rawPointCount: number;
  compressedPointCount: number;
  compressionRatio: number;
  reductionPercent: number;
  preservedStops: number;
  preservedAbnormalPoints: number;
};

export type GeoPoint = {
  longitude: number;
  latitude: number;
};

export type GeoFence = {
  id: number;
  name: string;
  type: string;
  siteId: number;
  projectId: number;
  longitude: number;
  latitude: number;
  radius: number;
  shape: string;
  polygon: GeoPoint[];
  status: string;
};

export type GeoFenceEvent = {
  id: number;
  vehicleId: number;
  fenceId: number;
  eventType: string;
  dispatchId: number;
  eventTime: string;
};

export type VehicleAlarm = {
  id: number;
  vehicleId: number;
  dispatchId: number;
  alarmType: string;
  level: string;
  message: string;
  status: string;
  createdAt: string;
  handledBy?: string;
  handledAt?: string;
};

export type RuleDefinition = {
  id: number;
  code: string;
  name: string;
  category: string;
  metric: string;
  operator: string;
  threshold: number;
  level: string;
  enabled: boolean;
  notifyRoles: string[];
  description: string;
};

export type NotificationItem = {
  id: number;
  targetRole: string;
  channel: string;
  title: string;
  content: string;
  status: string;
  createdAt: string;
};

export type IntegrationEndpoint = {
  id: number;
  name: string;
  type: string;
  protocol: string;
  url: string;
  status: string;
  lastSyncAt: string;
};

export type VehicleDevice = {
  id: number;
  vehicleId: number;
  deviceNo: string;
  protocol: string;
  vendor: string;
  status: string;
  lastSeenAt: string;
};

export type DeviceCredential = {
  id: number;
  deviceNo: string;
  scopes: string[];
  status: string;
  lastUsedAt: string;
};

export type DeviceProtocolFrame = {
  id: number;
  frameNo: string;
  channel: string;
  protocol: string;
  deviceNo: string;
  raw: string;
  parsedResource: string;
  parsedId: number;
  status: string;
  error: string;
  receivedAt: string;
  actor: string;
};

export type SecurityPolicy = {
  id: number;
  name: string;
  type: string;
  value: string;
  enabled: boolean;
  remark: string;
};

export type SessionPolicy = {
  timeoutMinutes: number;
  maxSessionsPerUser: number;
  ipWhitelistEnabled: boolean;
  allowedIpRanges: string[];
};

export type ActiveSessionSummary = {
  userId: number;
  username: string;
  displayName: string;
  roleCode: string;
  ip: string;
  userAgent: string;
  createdAt: string;
  lastSeenAt: string;
  expiresAt: string;
  ageMinutes: number;
};

export type SecurityReport = {
  totalUsers: number;
  activeUsers: number;
  mfaUsers: number;
  mfaCoverage: number;
  ssoProviders: number;
  scimProviders: number;
  scimEventsLast24h: number;
  enabledSecurityPolicies: number;
  deviceCredentials: number;
  activeSessions: number;
  staleSessions: number;
  expiringSessions: number;
  loginLast24h: number;
  failedLoginLast24h: number;
  auditEventsLast24h: number;
  ipWhitelistEnabled: boolean;
  riskLevel: string;
  recommendations: string[];
};

export type FieldPolicy = {
  id: number;
  roleCode: string;
  resource: string;
  field: string;
  mask: string;
  enabled: boolean;
  remark: string;
};

export type RuntimeStatus = {
  storage: string;
  businessTables: string;
  businessTableCount: number;
  businessProjectionRows: number;
  domainTables: string;
  domainResourceCount: number;
  domainRowCount: number;
  redisAddr: string;
  redis: string;
  rabbitUrl: string;
  rabbitmq: string;
  clickhouseUrl: string;
  clickhouse: string;
  eventBus: string;
  taxGatewayProvider: string;
  taxGatewayUrl: string;
  taxGateway: string;
  mapProvider: string;
  mapTiles: string;
  mapTileUrl: string;
  mapCoordinateSystem: string;
  mapApiKeyConfigured: boolean;
  deviceGateways: DeviceGatewayStatus[];
};

export type MapProviderConfig = {
  provider: string;
  name: string;
  tileUrl: string;
  attribution: string;
  subdomains: string[];
  minZoom: number;
  maxZoom: number;
  coordinateSystem: string;
  offline: boolean;
  apiKeyConfigured: boolean;
  requiresClientKey: boolean;
};

export type DeviceGatewayStatus = {
  name: string;
  kind: string;
  address: string;
  source: string;
  channel: string;
  protocol: string;
  status: string;
  lastError: string;
  acceptedFrames: number;
  rejectedFrames: number;
  lastFrameAt: string;
};

export type GatewayRoute = {
  id: number;
  name: string;
  pathPrefix: string;
  stableUpstream: string;
  canaryUpstream: string;
  canaryPercent: number;
  drainEnabled: boolean;
  drainUntil: string;
  readTimeoutSec: number;
  status: string;
  updatedAt: string;
};

export type GatewayEvent = {
  id: number;
  eventNo: string;
  routeId: number;
  routeName: string;
  action: string;
  detail: string;
  actor: string;
  createdAt: string;
};

export type GatewayDrainCheck = {
  routeId: number;
  routeName: string;
  pathPrefix: string;
  drainEnabled: boolean;
  drainUntil: string;
  active: boolean;
  remainingSeconds: number;
  probePath: string;
  expectedHeader: string;
  status: string;
};

export type GatewayReloadPlan = {
  reloadRequired: boolean;
  valid: boolean;
  configHash: string;
  configPath: string;
  reloadCommand: string;
  generatedAt: string;
  lastReloadAt: string;
  reloadedAt?: string;
  activeRoutes: number;
  drainingRoutes: number;
  automaticReload: boolean;
  steps: string[];
  errors: string[];
  drainChecks: GatewayDrainCheck[];
};

export type GatewayOverview = {
  routes: GatewayRoute[];
  events: GatewayEvent[];
  nginxConfig: string;
  reloadPlan: GatewayReloadPlan;
};

export type BackupInfo = {
  name: string;
  path: string;
  size: number;
  createdAt: string;
};

export type BackupDrill = {
  id: number;
  drillNo: string;
  backupName: string;
  status: string;
  startedAt: string;
  completedAt: string;
  durationMs: number;
  snapshotSize: number;
  schemaVersion: number;
  objectCounts: Record<string, number>;
  checks: string[];
  error: string;
  actor: string;
};

export type ScaleDevice = {
  id: number;
  siteId: number;
  name: string;
  code: string;
  protocol: string;
  ip: string;
  status: string;
};

export type Plant = {
  id: number;
  siteId: number;
  name: string;
  code: string;
  capacity: string;
  interface: string;
  status: string;
};

export type BootstrapData = {
  user: User;
  license: LicenseInfo;
  modules: ModuleInfo[];
  roles: unknown[];
  companies: Company[];
  departments: Department[];
  sites: Site[];
  plants: Plant[];
  customers: Customer[];
  customerContacts: CustomerContact[];
  customerBlacklists: CustomerBlacklist[];
  customerProfiles: CustomerProfile[];
  customerComplaints: CustomerComplaint[];
  pricePolicies: PricePolicy[];
  taxRates: TaxRate[];
  projects: Project[];
  products: Product[];
  materials: Material[];
  carriers: Carrier[];
  vehicles: Vehicle[];
  drivers: Driver[];
  contracts: Contract[];
  contractAttachments: ContractAttachment[];
  dispatchSchedules: DispatchSchedule[];
  productionPlans: ProductionPlan[];
  mixDesigns: MixDesign[];
  mixDesignTrialRuns: MixDesignTrialRun[];
  productionTasks: ProductionTask[];
  productionBatches: ProductionBatch[];
  productionReports: ProductionDailyReport[];
  qualityInspections: QualityInspection[];
  qualitySamples: QualitySample[];
  laboratorySamples: LaboratorySample[];
  laboratoryTests: LaboratoryTestRecord[];
  laboratoryEquipment: LaboratoryEquipment[];
  laboratoryCalibrations: LaboratoryCalibration[];
  qualityExceptions: QualityException[];
  inventory: InventoryItem[];
};

export type DashboardData = {
  kpis: Record<string, number>;
  siteProduction: Record<string, number>;
  customerDebt: Record<string, number>;
  recentOrders: SalesOrder[];
  alarms: VehicleAlarm[];
  operating: OperatingAnalysisReport;
  customerAging: CustomerAgingReport[];
  quality: QualityAnalysisReport;
  energy: ProductionEnergyReport[];
};

export type OperatingAnalysisReport = {
  orderCount: number;
  plannedQty: number;
  signedQty: number;
  revenue: number;
  materialCost: number;
  transportCost: number;
  totalCost: number;
  grossProfit: number;
  grossMargin: number;
  receivableBalance: number;
  overdueReceivable: number;
  inventoryWarningCount: number;
  openQualityIssues: number;
};

export type CustomerAgingReport = {
  customerId: number;
  customerName: string;
  current: number;
  overdue1To30: number;
  overdue31To60: number;
  overdue61To90: number;
  overdueOver90: number;
  total: number;
  overdueTotal: number;
};

export type QualityAnalysisReport = {
  inspections: number;
  passed: number;
  pending: number;
  failed: number;
  passRate: number;
  samples: number;
  samplePassed: number;
  samplePending: number;
  sampleFailed: number;
  laboratoryTests: number;
  openExceptionCount: number;
  openExceptions: QualityInspection[];
  qualityExceptions: QualityException[];
};

export type ProductionEnergyReport = {
  siteId: number;
  producedQty: number;
  batchCount: number;
  materialUsageQty: number;
  materialCost: number;
  estimatedPowerKwh: number;
  unitMaterialCost: number;
  unitPowerKwh: number;
};

export type ManagementReports = {
  projectProfit: ProjectProfit[];
  inventoryWarnings: InventoryItem[];
  vehicleEfficiency: Record<string, unknown>[];
  customerStatements: Statement[];
  operating: OperatingAnalysisReport;
  customerAging: CustomerAgingReport[];
  agingBuckets: ReceivableAgingBucket[];
  quality: QualityAnalysisReport;
  energy: ProductionEnergyReport[];
};

export type ProcurementOverview = {
  requests: PurchaseRequest[];
  orders: PurchaseOrder[];
  receipts: RawMaterialReceipt[];
  flows: InventoryFlow[];
  inventory: InventoryItem[];
  transfers: InventoryTransfer[];
  stocktakes: InventoryStocktake[];
  traces: InventoryBatchTrace[];
  suppliers: unknown[];
  warehouses: Warehouse[];
  silos: Silo[];
};

export type ProductionOverview = {
  plans: ProductionPlan[];
  tasks: ProductionTask[];
  batches: ProductionBatch[];
  reports: ProductionDailyReport[];
  mixDesigns: MixDesign[];
  plants: Plant[];
  traces: InventoryBatchTrace[];
};

export type DataDictionary = {
  id: number;
  type: string;
  code: string;
  label: string;
  sort: number;
  status: string;
};

export type QualityOverview = {
  inspections: QualityInspection[];
  samples: QualitySample[];
  rawInspections: RawMaterialInspection[];
  batches: ProductionBatch[];
  receipts: RawMaterialReceipt[];
  mixDesigns: MixDesign[];
  laboratorySamples: LaboratorySample[];
  laboratoryTests: LaboratoryTestRecord[];
  equipment: LaboratoryEquipment[];
  calibrations: LaboratoryCalibration[];
  qualityExceptions: QualityException[];
  laboratoryKpis: LaboratoryKPI;
};

export type LaboratoryOverview = {
  kpis: LaboratoryKPI;
  mixDesigns: MixDesign[];
  trialRuns: MixDesignTrialRun[];
  qualityInspections: QualityInspection[];
  qualitySamples: QualitySample[];
  rawInspections: RawMaterialInspection[];
  samples: LaboratorySample[];
  tests: LaboratoryTestRecord[];
  equipment: LaboratoryEquipment[];
  calibrations: LaboratoryCalibration[];
  exceptions: QualityException[];
  batches: ProductionBatch[];
  receipts: RawMaterialReceipt[];
  products: Product[];
  materials: Material[];
  sites: Site[];
};

export type FinanceOverview = {
  invoices: SalesInvoice[];
  redLetterInfos: RedLetterInfo[];
  invoiceTypes: DataDictionary[];
  taxGatewaySubmissions: TaxGatewaySubmission[];
  statements: Statement[];
  receivables: Receivable[];
  receipts: Receipt[];
  paymentPlans: PaymentPlan[];
  collectionTasks: CollectionTask[];
  collectionTemplates: CollectionTemplate[];
  collectionDispatches: CollectionDispatch[];
  agingBuckets: ReceivableAgingBucket[];
  supplierStatements: SupplierStatement[];
  payables: Payable[];
  payments: unknown[];
  transportSettlements: TransportSettlement[];
  transportSettlementItems: TransportSettlementItem[];
  costCalcs: unknown[];
  projectProfits: ProjectProfit[];
};

export type RuleOverview = {
  rules: RuleDefinition[];
  alarms: VehicleAlarm[];
  notifications: NotificationItem[];
};

export type IntegrationOverview = {
  endpoints: IntegrationEndpoint[];
  vehicleDevices: VehicleDevice[];
  scaleDevices: ScaleDevice[];
  plants: Plant[];
  protocolFrames: DeviceProtocolFrame[];
};

export type SystemBundle = {
  plugins: PluginInfo[];
  pluginRuns: PluginRun[];
  updates: UpdatePackage[];
  licensePackages: LicensePackage[];
  licenseIssues: LicenseIssueRecord[];
  licenseRevocations: LicenseRevocation[];
  licensePortal: LicensePortalOverview;
  security: {
    policies: SecurityPolicy[];
    fieldPolicies: FieldPolicy[];
    deviceCredentials: DeviceCredential[];
    users: User[];
    ssoProviders: OIDCProvider[];
    scimProviders: SCIMProvider[];
    scimEvents: SCIMProvisioningEvent[];
    sessionPolicy: SessionPolicy;
    sessions: ActiveSessionSummary[];
    report: SecurityReport;
  };
  runtime: RuntimeStatus;
  backups: BackupInfo[];
  backupDrills: BackupDrill[];
  approvalFlows: ApprovalFlow[];
  dictionaries: DataDictionary[];
  licenseVerified: LicenseVerification;
};
