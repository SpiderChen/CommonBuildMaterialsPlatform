import { ClipboardList, FileWarning, FlaskConical, ListChecks, TestTube2, Wrench } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { api } from "../services/api";
import type { LaboratoryOverview } from "../services/types";
import { EquipmentCalibrationModule } from "./laboratory/EquipmentCalibrationModule";
import { ExceptionClosureModule } from "./laboratory/ExceptionClosureModule";
import { LaboratoryKpiStrip } from "./laboratory/LaboratoryKpiStrip";
import type {
  CalibrationForm,
  EquipmentForm,
  ExceptionForm,
  LaboratoryModuleKey,
  MixForm,
  SampleForm,
  SubmitHandler,
  TestForm,
  TrialForm
} from "./laboratory/LaboratoryModuleTypes";
import { normalizeLaboratoryOverview, parseMixMaterials, parseNumber, today } from "./laboratory/laboratoryHelpers";
import { MixDesignModule } from "./laboratory/MixDesignModule";
import { SampleLedgerModule } from "./laboratory/SampleLedgerModule";
import { SampleTestModule } from "./laboratory/SampleTestModule";
import { TrialRunModule } from "./laboratory/TrialRunModule";

const laboratoryModules = [
  { key: "mix-designs", label: "配比版本", icon: FlaskConical },
  { key: "trial-runs", label: "试配记录", icon: ClipboardList },
  { key: "sample-tests", label: "样品试验", icon: TestTube2 },
  { key: "equipment-calibration", label: "仪器校准", icon: Wrench },
  { key: "sample-ledger", label: "样品台账", icon: ListChecks },
  { key: "exceptions", label: "异常闭环", icon: FileWarning }
] satisfies Array<{ key: LaboratoryModuleKey; label: string; icon: typeof FlaskConical }>;

export function LaboratoryView({ onChanged }: { onChanged: () => void }) {
  const [overview, setOverview] = useState<LaboratoryOverview | null>(null);
  const [activeModule, setActiveModule] = useState<LaboratoryModuleKey>("mix-designs");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [mixForm, setMixForm] = useState<MixForm>({
    productId: "1",
    siteId: "1",
    code: "MD-NEW",
    version: "v1",
    strengthGrade: "C30",
    slump: "180mm",
    scope: "站内标准生产配比",
    materials: "1:330:kg/m3\n3:780:kg/m3\n4:1020:kg/m3\n5:8.0:kg/m3"
  });
  const [trialForm, setTrialForm] = useState<TrialForm>({
    mixDesignId: "1",
    strength7d: "32",
    strength28d: "42",
    water: "165",
    sandRate: "42",
    admixtureRate: "1.2",
    result: "passed"
  });
  const [sampleForm, setSampleForm] = useState<SampleForm>({
    siteId: "1",
    productId: "1",
    mixDesignId: "1",
    sampleType: "compressive_strength",
    plannedTestAt: today
  });
  const [testForm, setTestForm] = useState<TestForm>({
    sampleId: "1",
    equipmentId: "1",
    metric: "28d_strength",
    value: "42",
    unit: "MPa",
    result: "passed"
  });
  const [equipmentForm, setEquipmentForm] = useState<EquipmentForm>({
    name: "压力试验机",
    siteId: "1",
    model: "YES-2000",
    serialNo: "LAB-NEW-001",
    calibrationCycleDays: "180",
    lastCalibrationAt: today,
    nextCalibrationAt: "2026-12-31"
  });
  const [calibrationForm, setCalibrationForm] = useState<CalibrationForm>({
    equipmentId: "1",
    result: "passed",
    calibratedAt: today,
    nextDueAt: "2026-12-31",
    certificateNo: "CAL-NEW",
    agency: "计量检测中心"
  });
  const [exceptionForm, setExceptionForm] = useState<ExceptionForm>({
    title: "质量异常",
    severity: "medium",
    responsible: "实验室质检员",
    description: "现场反馈或试验异常待处理"
  });

  async function load() {
    setError("");
    setOverview(normalizeLaboratoryOverview(await api.laboratoryOverview()));
  }

  useEffect(() => {
    load().catch((err: unknown) => setError(err instanceof Error ? err.message : "加载实验室数据失败"));
  }, []);

  useEffect(() => {
    if (!overview) return;
    const firstSite = overview.sites[0]?.id || 1;
    const firstProduct = overview.products[0]?.id || 1;
    const firstMix = overview.mixDesigns[0]?.id || 1;
    const firstSample = overview.samples.find((item) => item.status !== "completed")?.id || overview.samples[0]?.id || 1;
    const firstEquipment = overview.equipment.find((item) => item.status === "active")?.id || overview.equipment[0]?.id || 1;
    setMixForm((value) => ({ ...value, siteId: String(firstSite), productId: String(firstProduct) }));
    setTrialForm((value) => ({ ...value, mixDesignId: String(firstMix) }));
    setSampleForm((value) => ({ ...value, siteId: String(firstSite), productId: String(firstProduct), mixDesignId: String(firstMix) }));
    setTestForm((value) => ({ ...value, sampleId: String(firstSample), equipmentId: String(firstEquipment) }));
    setEquipmentForm((value) => ({ ...value, siteId: String(firstSite) }));
    setCalibrationForm((value) => ({ ...value, equipmentId: String(firstEquipment) }));
  }, [overview?.mixDesigns.length, overview?.samples.length, overview?.equipment.length]);

  async function mutate(label: string, action: () => Promise<unknown>) {
    setBusy(label);
    setError("");
    try {
      await action();
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBusy("");
    }
  }

  const submitMix: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("mix", () => api.createLaboratoryMixDesign({
      productId: parseNumber(mixForm.productId),
      siteId: parseNumber(mixForm.siteId),
      code: mixForm.code,
      version: mixForm.version,
      strengthGrade: mixForm.strengthGrade,
      slump: mixForm.slump,
      scope: mixForm.scope,
      materials: parseMixMaterials(mixForm.materials)
    }));
  };

  const submitTrial: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("trial", () => api.createMixDesignTrialRun(parseNumber(trialForm.mixDesignId), {
      water: parseNumber(trialForm.water),
      sandRate: parseNumber(trialForm.sandRate),
      admixtureRate: parseNumber(trialForm.admixtureRate),
      strength7d: parseNumber(trialForm.strength7d),
      strength28d: parseNumber(trialForm.strength28d),
      result: trialForm.result
    }));
  };

  const submitSample: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("sample", () => api.createLaboratorySample({
      siteId: parseNumber(sampleForm.siteId),
      productId: parseNumber(sampleForm.productId),
      mixDesignId: parseNumber(sampleForm.mixDesignId),
      sampleType: sampleForm.sampleType,
      plannedTestAt: sampleForm.plannedTestAt
    }));
  };

  const submitTest: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("test", async () => {
      const test = await api.createLaboratoryTest(parseNumber(testForm.sampleId), {
        equipmentId: parseNumber(testForm.equipmentId),
        metric: testForm.metric,
        value: parseNumber(testForm.value),
        unit: testForm.unit,
        result: testForm.result
      });
      return api.reviewLaboratoryTest(test.id, { result: testForm.result });
    });
  };

  const submitEquipment: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("equipment", () => api.createLaboratoryEquipment({
      name: equipmentForm.name,
      siteId: parseNumber(equipmentForm.siteId),
      model: equipmentForm.model,
      serialNo: equipmentForm.serialNo,
      calibrationCycleDays: parseNumber(equipmentForm.calibrationCycleDays, 180),
      lastCalibrationAt: equipmentForm.lastCalibrationAt,
      nextCalibrationAt: equipmentForm.nextCalibrationAt
    }));
  };

  const submitCalibration: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("calibration", () => api.createLaboratoryCalibration(parseNumber(calibrationForm.equipmentId), {
      result: calibrationForm.result,
      calibratedAt: calibrationForm.calibratedAt,
      nextDueAt: calibrationForm.nextDueAt,
      certificateNo: calibrationForm.certificateNo,
      agency: calibrationForm.agency
    }));
  };

  const submitException: SubmitHandler = async (event) => {
    event.preventDefault();
    await mutate("exception", () => api.createQualityException({
      title: exceptionForm.title,
      severity: exceptionForm.severity,
      responsible: exceptionForm.responsible,
      description: exceptionForm.description,
      siteId: parseNumber(sampleForm.siteId)
    }));
  };

  const productOptions = overview?.products || [];
  const siteOptions = overview?.sites || [];
  const materials = overview?.materials || [];
  const currentMixes = useMemo(() => overview?.mixDesigns.filter((item) => item.isCurrent && item.status === "approved") || [], [overview]);
  const draftMixes = useMemo(() => overview?.mixDesigns.filter((item) => item.status === "draft" || item.status === "pending_approval") || [], [overview]);
  const openExceptions = overview?.exceptions.filter((item) => item.status !== "closed") || [];

  if (!overview) {
    return (
      <section className="panel">
        {error ? <p className="error-text">{error}</p> : "加载实验室工作台..."}
      </section>
    );
  }

  const loadedOverview = overview;

  function renderModule() {
    switch (activeModule) {
      case "mix-designs":
        return (
          <MixDesignModule
            overview={loadedOverview}
            productOptions={productOptions}
            siteOptions={siteOptions}
            materials={materials}
            mixForm={mixForm}
            setMixForm={setMixForm}
            busy={busy}
            mutate={mutate}
            onReload={load}
            onSubmitMix={submitMix}
          />
        );
      case "trial-runs":
        return (
          <TrialRunModule
            trialForm={trialForm}
            setTrialForm={setTrialForm}
            draftMixes={draftMixes}
            currentMixes={currentMixes}
            trialRuns={loadedOverview.trialRuns}
            busy={busy}
            onSubmitTrial={submitTrial}
          />
        );
      case "sample-tests":
        return (
          <SampleTestModule
            productOptions={productOptions}
            mixDesigns={loadedOverview.mixDesigns}
            samples={loadedOverview.samples}
            equipment={loadedOverview.equipment}
            sampleForm={sampleForm}
            setSampleForm={setSampleForm}
            testForm={testForm}
            setTestForm={setTestForm}
            busy={busy}
            onSubmitSample={submitSample}
            onSubmitTest={submitTest}
          />
        );
      case "equipment-calibration":
        return (
          <EquipmentCalibrationModule
            equipment={loadedOverview.equipment}
            calibrations={loadedOverview.calibrations}
            equipmentForm={equipmentForm}
            setEquipmentForm={setEquipmentForm}
            calibrationForm={calibrationForm}
            setCalibrationForm={setCalibrationForm}
            busy={busy}
            onSubmitEquipment={submitEquipment}
            onSubmitCalibration={submitCalibration}
          />
        );
      case "sample-ledger":
        return (
          <SampleLedgerModule
            samples={loadedOverview.samples}
            tests={loadedOverview.tests}
            productOptions={productOptions}
            materials={materials}
          />
        );
      case "exceptions":
        return (
          <ExceptionClosureModule
            exceptions={loadedOverview.exceptions}
            openExceptions={openExceptions}
            exceptionForm={exceptionForm}
            setExceptionForm={setExceptionForm}
            busy={busy}
            mutate={mutate}
            onSubmitException={submitException}
          />
        );
      default:
        return null;
    }
  }

  return (
    <div className="view-stack laboratory-view">
      <LaboratoryKpiStrip overview={loadedOverview} />
      {error ? <p className="error-text">{error}</p> : null}

      <section className="panel laboratory-module-tabs">
        <div className="tabs" role="tablist" aria-label="实验室模块">
          {laboratoryModules.map((item) => {
            const Icon = item.icon;
            return (
              <button
                className={`tab laboratory-module-tab${activeModule === item.key ? " active" : ""}`}
                type="button"
                key={item.key}
                aria-selected={activeModule === item.key}
                onClick={() => setActiveModule(item.key)}
              >
                <Icon size={16} />{item.label}
              </button>
            );
          })}
        </div>
      </section>

      {renderModule()}
    </div>
  );
}
