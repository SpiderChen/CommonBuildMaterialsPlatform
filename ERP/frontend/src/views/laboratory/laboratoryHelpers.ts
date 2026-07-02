import { nameOf } from "../../components/names";
import { activeDictionaryOptions, dictionaryLabel } from "../../services/dictionaries";
import type {
  DataDictionary,
  LaboratoryCalibration,
  LaboratoryEquipment,
  LaboratoryOverview,
  LaboratorySample,
  LaboratoryTestRecord,
  Material,
  MixDesign,
  MixDesignMaterial,
  MixDesignTrialRun,
  Product,
  QualityException
} from "../../services/types";
import type { MixMaterialForm } from "./LaboratoryModuleTypes";

export const today = new Date().toISOString().slice(0, 10);

export function laboratoryDictionaryOptions(dictionaries: DataDictionary[] | null | undefined, type: string) {
  return activeDictionaryOptions(dictionaries, type);
}

export function parseNumber(value: string, fallbackValue = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallbackValue;
}

export function parseMixMaterials(items: MixMaterialForm[]): MixDesignMaterial[] {
  return items
    .map((entry) => ({
      materialId: parseNumber(entry.materialId),
      dosage: parseNumber(entry.dosage),
      unit: entry.unit || "kg/t"
    }))
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

export function shortDate(value?: string) {
  return value ? value.slice(0, 10) : "-";
}

export function daysUntil(value?: string) {
  if (!value) return null;
  const target = new Date(value.slice(0, 10));
  const base = new Date(today);
  if (Number.isNaN(target.getTime())) return null;
  return Math.ceil((target.getTime() - base.getTime()) / 86400000);
}

export function percent(part: number, total: number) {
  if (!total) return 0;
  return Math.round((part / total) * 100);
}

export function average(values: number[]) {
  const validValues = values.filter((value) => Number.isFinite(value));
  if (!validValues.length) return 0;
  return Math.round((validValues.reduce((sum, value) => sum + value, 0) / validValues.length) * 10) / 10;
}

export function statusCount<T extends { status: string }>(items: T[], status: string) {
  return items.filter((item) => item.status === status).length;
}

export function resultCount<T extends { result: string }>(items: T[], result: string) {
  return items.filter((item) => item.result === result).length;
}

export function latestTestForSample(sample: LaboratorySample, tests: LaboratoryTestRecord[]) {
  return tests.filter((item) => item.sampleId === sample.id).sort((a, b) => b.id - a.id)[0];
}

export function latestCalibrationForEquipment(equipment: LaboratoryEquipment, calibrations: LaboratoryCalibration[]) {
  return calibrations.filter((item) => item.equipmentId === equipment.id).sort((a, b) => b.id - a.id)[0];
}

export function trialsForMix(mix: MixDesign, trialRuns: MixDesignTrialRun[]) {
  return trialRuns.filter((item) => item.mixDesignId === mix.id).sort((a, b) => b.id - a.id);
}

export function sampleTypeLabel(value: string, dictionaries?: DataDictionary[]) {
  const labels: Record<string, string> = {
    compressive_strength: "马歇尔稳定度",
    marshall_stability: "马歇尔稳定度",
    raw_material: "原材复检",
    slump: "油石比",
    durability: "耐久性",
    chloride: "氯离子"
  };
  return dictionaryLabel(dictionaries, "laboratory_test_type", value, labels[value] || value || "-");
}

export function sourceTypeLabel(value: string, dictionaries?: DataDictionary[]) {
  const labels: Record<string, string> = {
    manual: "手工登记",
    production_batch: "生产批次",
    raw_inspection: "原材验收",
    quality_inspection: "质检抽样"
  };
  return dictionaryLabel(dictionaries, "sample_source_type", value, labels[value] || value || "-");
}

export function sampleSubject(sample: LaboratorySample, products: Product[], materials: Material[]) {
  if (sample.productId) return productName(products, sample.productId);
  if (sample.materialId) return nameOf(materials, sample.materialId);
  return "-";
}

export function calibrationState(equipment: LaboratoryEquipment) {
  const dueDays = daysUntil(equipment.nextCalibrationAt);
  if (dueDays === null) return { label: "未排期", tone: "unknown", detail: "缺少下次校准日期" };
  if (dueDays < 0) return { label: "已逾期", tone: "failed", detail: `逾期 ${Math.abs(dueDays)} 天` };
  if (dueDays <= 30) return { label: "临期", tone: "warning", detail: `${dueDays} 天后到期` };
  return { label: "有效", tone: "active", detail: `${dueDays} 天后到期` };
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
    mixDesignPlantProfiles: arrayOrEmpty(payload.mixDesignPlantProfiles).map((profile) => ({ ...profile, materials: arrayOrEmpty(profile.materials) })),
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
    sites: arrayOrEmpty(payload.sites),
    plants: arrayOrEmpty(payload.plants),
    plantBufferLocations: arrayOrEmpty(payload.plantBufferLocations),
    dictionaries: arrayOrEmpty(payload.dictionaries)
  };
}
