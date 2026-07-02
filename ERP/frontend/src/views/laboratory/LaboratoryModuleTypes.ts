import type { FormEvent } from "react";

export type LaboratoryModuleKey =
  | "mix-designs"
  | "plant-mix-designs"
  | "trial-runs"
  | "sample-tests"
  | "equipment-calibration"
  | "sample-ledger"
  | "exceptions";

export type SubmitHandler = (event: FormEvent<HTMLFormElement>) => void | Promise<void>;

export type MutateAction = (label: string, action: () => Promise<unknown>) => Promise<void>;

export type MixMaterialForm = {
  materialId: string;
  dosage: string;
  unit: string;
};

export type MixForm = {
  productId: string;
  siteId: string;
  code: string;
  version: string;
  strengthGrade: string;
  slump: string;
  scope: string;
  materials: MixMaterialForm[];
};

export type TrialForm = {
  mixDesignId: string;
  strength7d: string;
  strength28d: string;
  water: string;
  sandRate: string;
  admixtureRate: string;
  result: string;
};

export type SampleForm = {
  siteId: string;
  productId: string;
  mixDesignId: string;
  sampleType: string;
  plannedTestAt: string;
};

export type TestForm = {
  sampleId: string;
  equipmentId: string;
  metric: string;
  value: string;
  unit: string;
  result: string;
};

export type EquipmentForm = {
  name: string;
  siteId: string;
  model: string;
  serialNo: string;
  calibrationCycleDays: string;
  lastCalibrationAt: string;
  nextCalibrationAt: string;
};

export type CalibrationForm = {
  equipmentId: string;
  result: string;
  calibratedAt: string;
  nextDueAt: string;
  certificateNo: string;
  agency: string;
};

export type ExceptionForm = {
  title: string;
  severity: string;
  responsible: string;
  description: string;
  rootCause: string;
  correctiveAction: string;
};
