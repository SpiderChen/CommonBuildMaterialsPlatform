import { nameOf } from "../../components/names";
import type {
  LaboratoryOverview,
  LaboratorySample,
  LaboratoryTestRecord,
  Material,
  MixDesignMaterial,
  Product,
  QualityException
} from "../../services/types";

export const today = new Date().toISOString().slice(0, 10);

export function parseNumber(value: string, fallbackValue = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallbackValue;
}

export function parseMixMaterials(value: string): MixDesignMaterial[] {
  return value
    .split(/[\n,]+/)
    .map((entry) => entry.trim())
    .filter(Boolean)
    .map((entry) => {
      const [materialId, dosage, unit] = entry.split(":").map((part) => part.trim());
      return { materialId: parseNumber(materialId), dosage: parseNumber(dosage), unit: unit || "kg/m3" };
    })
    .filter((item) => item.materialId > 0 && item.dosage > 0);
}

export function arrayOrEmpty<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : [];
}

export function materialSummary(items: MixDesignMaterial[] | null | undefined, materials: Material[]) {
  const safeItems = arrayOrEmpty(items);
  if (!safeItems.length) return "-";
  return safeItems.map((item) => `${nameOf(materials, item.materialId)} ${item.dosage}${item.unit}`).join(" / ");
}

export function productName(products: Product[], id: number) {
  const product = products.find((item) => item.id === id);
  if (!product) return "-";
  return `${product.name} ${product.spec}`;
}

export function latestTestForSample(sample: LaboratorySample, tests: LaboratoryTestRecord[]) {
  return tests.filter((item) => item.sampleId === sample.id).sort((a, b) => b.id - a.id)[0];
}

export function exceptionSummary(items: QualityException[]) {
  const open = items.filter((item) => item.status !== "closed");
  return open.length ? `${open.length} 个待处理` : "已闭环";
}

const emptyLaboratoryKpis: LaboratoryOverview["kpis"] = {
  mixDesigns: 0,
  currentMixDesigns: 0,
  pendingMixDesigns: 0,
  trialRuns: 0,
  samples: 0,
  pendingSamples: 0,
  tests: 0,
  pendingReviews: 0,
  equipments: 0,
  calibrationDue: 0,
  calibrationOverdue: 0,
  openExceptions: 0,
  passRate: 0
};

export function normalizeLaboratoryOverview(payload: LaboratoryOverview): LaboratoryOverview {
  return {
    ...payload,
    kpis: payload.kpis ?? emptyLaboratoryKpis,
    mixDesigns: arrayOrEmpty(payload.mixDesigns).map((mix) => ({ ...mix, materials: arrayOrEmpty(mix.materials) })),
    trialRuns: arrayOrEmpty(payload.trialRuns),
    qualityInspections: arrayOrEmpty(payload.qualityInspections),
    qualitySamples: arrayOrEmpty(payload.qualitySamples),
    rawInspections: arrayOrEmpty(payload.rawInspections),
    samples: arrayOrEmpty(payload.samples),
    tests: arrayOrEmpty(payload.tests),
    equipment: arrayOrEmpty(payload.equipment),
    calibrations: arrayOrEmpty(payload.calibrations),
    exceptions: arrayOrEmpty(payload.exceptions),
    batches: arrayOrEmpty(payload.batches),
    receipts: arrayOrEmpty(payload.receipts),
    products: arrayOrEmpty(payload.products),
    materials: arrayOrEmpty(payload.materials),
    sites: arrayOrEmpty(payload.sites)
  };
}
