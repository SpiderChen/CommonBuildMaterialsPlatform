import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(new URL("..", import.meta.url).pathname);
const backendApp = readFileSync(resolve(root, "backend/internal/ops/app.go"), "utf8");
const backendWorkflow = readFileSync(resolve(root, "backend/internal/ops/workflow.go"), "utf8");
const frontendApp = readFileSync(resolve(root, "frontend/src/app.js"), "utf8");

const failures = [];

function requireNeedles(scope, source, checks) {
  for (const check of checks) {
    const missing = check.needles.filter((needle) => !source.includes(needle));
    if (missing.length) {
      failures.push(`${scope}: ${check.name} missing ${missing.join(", ")}`);
    }
  }
}

requireNeedles("backend routes", backendApp, [
  {
    name: "updater task polling",
    needles: ['parts[0] == "product-ops"', 'parts[1] == "system-updates"', 'parts[2] == "tasks"', "PollUpdateTasks"]
  },
  {
    name: "updater task reporting",
    needles: ['parts[4] == "report"', "ReportUpdateTask"]
  },
  {
    name: "assigned package download",
    needles: ['parts[0] == "system"', 'parts[1] == "updates"', 'parts[3] == "download"', "DownloadAssignedPackage"]
  },
  {
    name: "probe alert reporting",
    needles: ['parts[0] == "product-ops"', 'parts[1] == "alerts"', 'parts[2] == "report"', "CreateOrUpdateAlert"]
  }
]);

requireNeedles("backend workflow", backendWorkflow, [
  { name: "task numbers", needles: ["taskNoForAssignment", 'number("UA", id)'] },
  { name: "artifact verification", needles: ["opsUpdatePackageVerified", "ArtifactContentBase64"] },
  { name: "license package signing", needles: ["DownloadRenewalLicensePackage", "signLicensePackageForRenewal"] }
]);

requireNeedles("frontend actions", frontendApp, [
  { name: "updater task polling", needles: ['data-form="updater-poll"', "/api/product-ops/system-updates/tasks", "pollUpdaterTasks"] },
  { name: "updater task reporting", needles: ['data-form="updater-report"', "/api/product-ops/system-updates/tasks/${assignmentTaskNo", "reportAssignment"] },
  { name: "assigned package download", needles: ["downloadAssignmentPackage", "/api/system/updates/${assignment.packageId}/download?assignmentId="] },
  { name: "probe alert reporting", needles: ['data-form="probe-alert"', "/api/product-ops/alerts/report"] },
  { name: "targeted assignment", needles: ['data-form="assignment"', "data.getAll(\"customerIds\")"] },
  { name: "updater and probe pages", needles: ['href="#updater-flow"', 'href="#probe-flow"', 'id="updater-flow"', 'id="probe-flow"'] },
  { name: "row actions", needles: ['data-action="download-assignment-package"', 'data-action="report-assignment-applied"', 'data-action="poll-customer-updates"'] }
]);

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log("operations api coverage ok");
