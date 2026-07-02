import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDir, "..");
const erpRoot = path.resolve(frontendRoot, "..");

const appPath = path.join(frontendRoot, "src", "App.tsx");
const workbenchPath = path.join(frontendRoot, "src", "views", "ERPWorkbenchView.tsx");
const laboratoryPath = path.join(frontendRoot, "src", "views", "LaboratoryView.tsx");
const laboratoryTypesPath = path.join(frontendRoot, "src", "views", "laboratory", "LaboratoryModuleTypes.ts");
const siteSigningPath = path.join(frontendRoot, "src", "views", "SiteSigningModule.tsx");
const securityPath = path.join(erpRoot, "backend", "internal", "appliance", "security.go");

const app = fs.readFileSync(appPath, "utf8");
const workbench = fs.readFileSync(workbenchPath, "utf8");
const laboratory = fs.readFileSync(laboratoryPath, "utf8");
const laboratoryTypes = fs.readFileSync(laboratoryTypesPath, "utf8");
const siteSigning = fs.readFileSync(siteSigningPath, "utf8");
const security = fs.readFileSync(securityPath, "utf8");

function fail(message, values) {
  console.error(message);
  for (const value of values) console.error(`  - ${value}`);
  process.exitCode = 1;
}

function blockBetween(source, startText, endText) {
  const start = source.indexOf(startText);
  if (start < 0) return "";
  const end = source.indexOf(endText, start);
  return end < 0 ? "" : source.slice(start, end);
}

const navBlock = blockBetween(app, "const nav = [", "] as const;");
const navItems = [...navBlock.matchAll(/\{[^\n]*?key: "([^"]+)"[^\n]*?meta: \{([^}]*)\}/g)]
  .map((match) => ({ key: match[1], hidden: /hidden:\s*true/.test(match[2]) }));
const visibleNavKeys = navItems.filter((item) => !item.hidden).map((item) => item.key);
const visibleLaboratoryNavKeys = [...navBlock.matchAll(/\{[^\n]*?key: "([^"]+)"[^\n]*?group: "laboratory"[^\n]*?meta: \{([^}]*)\}/g)]
  .map((match) => ({ key: match[1], hidden: /hidden:\s*true/.test(match[2]) }))
  .filter((item) => !item.hidden && item.key !== "site-signing")
  .map((item) => item.key);

const workbenchBlock = app.match(/const workbenchSections:[^=]+= \[([^\]]+)\]/)?.[1] || "";
const workbenchSections = [...workbenchBlock.matchAll(/"([^"]+)"/g)].map((match) => match[1]);
const workbenchSet = new Set(workbenchSections);

const standaloneBlock = app.match(/standaloneSections = new Set<ViewKey>\(\[([^\]]+)\]\)/)?.[1] || "";
const standaloneSections = new Set([...standaloneBlock.matchAll(/"([^"]+)"/g)].map((match) => match[1]));
const laboratorySections = new Set([...app.matchAll(/active === "([^"]+)" \? <LaboratoryView/g)].map((match) => match[1]));
const directAppSections = new Set([...app.matchAll(/active === "([^"]+)" \? </g)].map((match) => match[1]));

const workbenchRenderBranches = new Set([...workbench.matchAll(/section === "([^"]+)" \? render/g)].map((match) => match[1]));
const workbenchRenderTargets = [...workbench.matchAll(/\{section === "([^"]+)" \? (render[A-Za-z0-9_]+)\(\) : null\}/g)]
  .map((match) => ({ key: match[1], render: match[2] }));
const workbenchConditionalBranches = new Set([...workbench.matchAll(/if \(section === "([^"]+)"\)/g)].map((match) => match[1]));
const menuPermissionMarks = new Set([...security.matchAll(/"([^"]+)":\s+"[^"]+"/g)]
  .map((match) => match[1])
  .filter((key) => !key.startsWith("group:")));
const backendMenuPermissionMap = new Map([...security.matchAll(/"([^"]+)":\s+"([^"]+)"/g)]
  .filter((match) => !match[1].startsWith("group:"))
  .map((match) => [match[1], match[2]]));
const routePermissionFallbackBlock = app.match(/routePermissionFallbacks:[^{]+\{([\s\S]*?)\n\};/)?.[1] || "";
const frontendRoutePermissionMap = new Map([
  ...[...routePermissionFallbackBlock.matchAll(/"([^"]+)":\s+"([^"]+)"/g)].map((match) => [match[1], match[2]]),
  ...[...routePermissionFallbackBlock.matchAll(/\n\s*([a-zA-Z][a-zA-Z0-9_-]*)\s*:\s+"([^"]+)"/g)].map((match) => [match[1], match[2]])
]);
const laboratoryModuleTypeKeys = new Set([...(laboratoryTypes.match(/export type LaboratoryModuleKey =([\s\S]*?);/)?.[1] || "").matchAll(/"([^"]+)"/g)].map((match) => match[1]));
const laboratorySwitchKeys = new Set([...laboratory.matchAll(/case "([^"]+)":/g)].map((match) => match[1]));

const workbenchWithoutRender = workbenchSections.filter((key) => !workbenchRenderBranches.has(key));
const renderTargetsBySection = new Map(workbenchRenderTargets.map((item) => [item.key, item.render]));
const sharedRenderTargets = new Map();
for (const key of workbenchSections) {
  const render = renderTargetsBySection.get(key);
  if (!render) continue;
  if (!sharedRenderTargets.has(render)) sharedRenderTargets.set(render, []);
  sharedRenderTargets.get(render).push(key);
}
function renderPrimarySection(renderName, keys) {
  const inferred = renderName
    .replace(/^render/, "")
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .toLowerCase();
  return keys.includes(inferred) ? inferred : keys[0];
}
const sharedPagesWithoutSpecializedBranch = [];
for (const [render, keys] of sharedRenderTargets.entries()) {
  if (keys.length < 2) continue;
  const primary = renderPrimarySection(render, keys);
  for (const key of keys) {
    if (key === primary) continue;
    if (!workbenchConditionalBranches.has(key)) sharedPagesWithoutSpecializedBranch.push(key);
  }
}
const portalGlobalFallbacks = [...workbench.matchAll(/list\(portal\?\.(\w+)\)\.length\s*\?\s*list\(portal\?\.\1\)\s*:\s*([^)\n;]+)/g)]
  .filter((match) => /\bscoped\w+|\bdata\./.test(match[2]))
  .map((match) => `${match[1]} -> ${match[2].trim()}`);
const portalBootstrapProjectFallbacks = [...workbench.matchAll(/portalProjectOptions\[0\]\?\.id\s*\|\|\s*firstId\(bootstrap\?\.projects\)/g)]
  .map((match) => `offset ${match.index}`);
const loadExpressions = new Map([...workbench.matchAll(/const (shouldLoad[A-Za-z]+) = ([^;]+);/g)]
  .map((match) => [match[1], match[2]]));
const loadConditionSections = (name, seen = new Set()) => {
  if (seen.has(name)) return new Set();
  seen.add(name);
  const expression = loadExpressions.get(name) || "";
  const sections = new Set([...expression.matchAll(/section === "([^"]+)"/g)].map((match) => match[1]));
  for (const ref of expression.matchAll(/\b(shouldLoad[A-Za-z]+)\b/g)) {
    if (ref[1] === name) continue;
    for (const section of loadConditionSections(ref[1], seen)) sections.add(section);
  }
  return sections;
};
const missingLoadSections = (conditionName, sections) => {
  const loaded = loadConditionSections(conditionName);
  return sections.filter((key) => !loaded.has(key)).map((key) => `${conditionName}: ${key}`);
};
const shouldLoadQualitySections = loadConditionSections("shouldLoadQuality");
const qualitySectionsMissingLoader = ["raw-material-receipts", "raw-material-inspections"]
  .filter((key) => !shouldLoadQualitySections.has(key));
const businessSectionsMissingLoader = [
  ...missingLoadSections("shouldLoadDispatchDetails", ["dispatch", "dispatch-schedules", "dispatch-queue", "settlement", "finance-carriers", "reports"]),
  ...missingLoadSections("shouldLoadPortal", ["portal-customer", "portal-driver", "dispatch", "delivery", "settlement"]),
  ...missingLoadSections("shouldLoadDeliveryDetails", ["delivery", "delivery-signs", "settlement"]),
  ...missingLoadSections("shouldLoadRules", ["system-rules", "map-center"]),
  ...missingLoadSections("shouldLoadIntegrations", ["system-integrations", "map-center"]),
  ...missingLoadSections("shouldLoadOrg", ["system-org"]),
  ...missingLoadSections("shouldLoadSystem", [
    "system-menu",
    "system-license",
    "system-maintenance",
    "system-gateway",
    "system-security",
    "system-identity",
    "system-plugins",
    "system-dictionaries",
    "system-users",
    "system-roles",
    "system-workflows",
    "system-audit",
    "approval-center"
  ])
];
const visibleWithoutPageTarget = visibleNavKeys.filter((key) => {
  if (workbenchSet.has(key)) return !workbenchRenderBranches.has(key);
  return !laboratorySections.has(key) && !standaloneSections.has(key) && !directAppSections.has(key);
});
const visibleWithoutPermissionMark = visibleNavKeys.filter((key) => !menuPermissionMarks.has(key));
const visibleWithoutFrontendPermissionFallback = visibleNavKeys.filter((key) => !frontendRoutePermissionMap.has(key));
const frontendBackendPermissionMismatches = visibleNavKeys
  .filter((key) => frontendRoutePermissionMap.has(key) && backendMenuPermissionMap.has(key) && frontendRoutePermissionMap.get(key) !== backendMenuPermissionMap.get(key))
  .map((key) => `${key}: frontend=${frontendRoutePermissionMap.get(key)} backend=${backendMenuPermissionMap.get(key)}`);
const laboratoryViewNavKeys = visibleLaboratoryNavKeys.filter((key) => laboratorySections.has(key));
const laboratoryNavWithoutType = laboratoryViewNavKeys.filter((key) => !laboratoryModuleTypeKeys.has(key));
const laboratoryTypeWithoutNav = [...laboratoryModuleTypeKeys].filter((key) => !visibleLaboratoryNavKeys.includes(key));
const laboratoryNavWithoutSwitch = laboratoryViewNavKeys.filter((key) => !laboratorySwitchKeys.has(key));
const siteSigningMissingLinkFlow = [
  !/api\.signLinks\(\)/.test(siteSigning) ? "missing api.signLinks() load" : "",
  !/api\.createSignLink\(/.test(siteSigning) ? "missing createSignLink action" : "",
  !/签收链接/.test(siteSigning) ? "missing sign link table text" : "",
  !/pendingDispatches/.test(siteSigning) ? "missing pending dispatch list" : ""
].filter(Boolean);

if (!navItems.length) fail("No nav items were parsed from App.tsx.", [appPath]);
if (!workbenchSections.length) fail("No workbench sections were parsed from App.tsx.", [appPath]);
if (!menuPermissionMarks.size) fail("No menu permission marks were parsed from security.go.", [securityPath]);
if (workbenchWithoutRender.length) fail("Workbench sections without render branches:", workbenchWithoutRender);
if (sharedPagesWithoutSpecializedBranch.length) fail("Shared workbench render targets without specialized section branches:", sharedPagesWithoutSpecializedBranch);
if (portalGlobalFallbacks.length) fail("Portal pages must not fall back to app-wide data when portal APIs return empty:", portalGlobalFallbacks);
if (portalBootstrapProjectFallbacks.length) fail("Portal complaint forms must not default to a global bootstrap project:", portalBootstrapProjectFallbacks);
if (qualitySectionsMissingLoader.length) fail("Quality-backed workbench sections missing qualityOverview loading:", qualitySectionsMissingLoader);
if (businessSectionsMissingLoader.length) fail("Workbench sections missing required conditional data loading:", businessSectionsMissingLoader);
if (visibleWithoutPageTarget.length) fail("Visible nav entries without page targets:", visibleWithoutPageTarget);
if (visibleWithoutPermissionMark.length) fail("Visible nav entries without backend menu permission marks:", visibleWithoutPermissionMark);
if (visibleWithoutFrontendPermissionFallback.length) fail("Visible nav entries without frontend route permission fallbacks:", visibleWithoutFrontendPermissionFallback);
if (frontendBackendPermissionMismatches.length) fail("Frontend route permission fallbacks differ from backend menu permission marks:", frontendBackendPermissionMismatches);
if (laboratoryNavWithoutType.length) fail("Laboratory nav entries missing LaboratoryModuleKey types:", laboratoryNavWithoutType);
if (laboratoryTypeWithoutNav.length) fail("LaboratoryModuleKey types without visible nav entries:", laboratoryTypeWithoutNav);
if (laboratoryNavWithoutSwitch.length) fail("Laboratory nav entries missing LaboratoryView render branches:", laboratoryNavWithoutSwitch);
if (siteSigningMissingLinkFlow.length) fail("Site signing page must expose the full dispatch -> link -> sign workflow:", siteSigningMissingLinkFlow);

if (!process.exitCode) {
  console.log(`route coverage ok: ${visibleNavKeys.length} visible nav entries, ${workbenchSections.length} workbench sections`);
  const specialized = [...workbenchConditionalBranches].filter((key) => visibleNavKeys.includes(key)).length;
  console.log(`specialized workbench branches: ${specialized}`);
}
