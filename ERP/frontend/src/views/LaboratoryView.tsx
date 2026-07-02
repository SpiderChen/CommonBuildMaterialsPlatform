import { useEffect, useMemo, useState } from "react";
import { Button, Panel, useMessageBox } from "../components";
import { api } from "../services/api";
import { hasPermission } from "../services/permissions";
import type { LaboratoryOverview, WorkflowOverview } from "../services/types";
import { sensitiveActionPrompt } from "../utils/sensitiveActions";
import { EquipmentCalibrationModule } from "./laboratory/EquipmentCalibrationModule";
import { ExceptionClosureModule } from "./laboratory/ExceptionClosureModule";
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
import { laboratoryDictionaryOptions, normalizeLaboratoryOverview, parseMixMaterials, parseNumber, today } from "./laboratory/laboratoryHelpers";
import { MixDesignModule } from "./laboratory/MixDesignModule";
import { SampleLedgerModule } from "./laboratory/SampleLedgerModule";
import { SampleTestModule } from "./laboratory/SampleTestModule";
import { TrialRunModule } from "./laboratory/TrialRunModule";

const blankMixForm: MixForm = {
  productId: "",
  siteId: "",
  code: "",
  version: "",
  strengthGrade: "",
  slump: "",
  scope: "",
  materials: []
};

const blankTrialForm: TrialForm = {
  mixDesignId: "",
  strength7d: "",
  strength28d: "",
  water: "",
  sandRate: "",
  admixtureRate: "",
  result: ""
};

const blankSampleForm: SampleForm = {
  siteId: "",
  productId: "",
  mixDesignId: "",
  sampleType: "",
  plannedTestAt: today
};

const blankTestForm: TestForm = {
  sampleId: "",
  equipmentId: "",
  metric: "",
  value: "",
  unit: "",
  result: ""
};

const blankEquipmentForm: EquipmentForm = {
  name: "",
  siteId: "",
  model: "",
  serialNo: "",
  calibrationCycleDays: "",
  lastCalibrationAt: today,
  nextCalibrationAt: ""
};

const blankCalibrationForm: CalibrationForm = {
  equipmentId: "",
  result: "",
  calibratedAt: today,
  nextDueAt: "",
  certificateNo: "",
  agency: ""
};

const blankExceptionForm: ExceptionForm = {
  title: "",
  severity: "",
  responsible: "",
  description: "",
  rootCause: "",
  correctiveAction: ""
};

function keepValidId(value: string, items: Array<{ id: number }>) {
  const id = Number(value);
  if (Number.isFinite(id) && items.some((item) => item.id === id)) return value;
  return items[0]?.id ? String(items[0].id) : "";
}

function keepDictionaryCode(value: string, options: Array<{ code: string }>) {
  return options.some((item) => item.code === value) ? value : options[0]?.code || "";
}

function defaultMixMaterials(overview: LaboratoryOverview) {
  return overview.materials.slice(0, 4).map((material) => ({
    materialId: String(material.id),
    dosage: "",
    unit: material.unit || "kg/t"
  }));
}

function keepMixMaterials(materials: MixForm["materials"], overview: LaboratoryOverview) {
  const materialIds = new Set(overview.materials.map((item) => item.id));
  const validRows = materials.filter((item) => materialIds.has(Number(item.materialId)));
  return validRows.length ? validRows : defaultMixMaterials(overview);
}

export function LaboratoryView({
  activeModule,
  currentRoleCode = "",
  currentPermissions = [],
  onChanged
}: {
  activeModule: LaboratoryModuleKey;
  currentRoleCode?: string;
  currentPermissions?: string[];
  onChanged: () => void;
}) {
  const [overview, setOverview] = useState<LaboratoryOverview | null>(null);
  const [workflowOverview, setWorkflowOverview] = useState<WorkflowOverview | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const { showError, confirmMessage } = useMessageBox();
  const [mixForm, setMixForm] = useState<MixForm>(blankMixForm);
  const [trialForm, setTrialForm] = useState<TrialForm>(blankTrialForm);
  const [sampleForm, setSampleForm] = useState<SampleForm>(blankSampleForm);
  const [testForm, setTestForm] = useState<TestForm>(blankTestForm);
  const [equipmentForm, setEquipmentForm] = useState<EquipmentForm>(blankEquipmentForm);
  const [calibrationForm, setCalibrationForm] = useState<CalibrationForm>(blankCalibrationForm);
  const [exceptionForm, setExceptionForm] = useState<ExceptionForm>(blankExceptionForm);

  async function load() {
    setError("");
    const workflowRequest = hasPermission(currentPermissions, "approval:read")
      ? api.workflowOverview()
      : Promise.resolve(null);
    const [nextOverview, nextWorkflowOverview] = await Promise.all([
      api.laboratoryOverview(),
      workflowRequest
    ]);
    setOverview(normalizeLaboratoryOverview(nextOverview));
    setWorkflowOverview(nextWorkflowOverview);
  }

  useEffect(() => {
    load().catch((err: unknown) => setError(err instanceof Error ? err.message : "加载实验室数据失败"));
  }, []);

  useEffect(() => {
    if (error) {
      showError(error, "加载实验室数据失败");
    }
  }, [error, showError]);

  useEffect(() => {
    if (!overview) return;
    const availableSamples = [overview.samples.find((item) => item.status !== "completed"), ...overview.samples].filter(Boolean) as Array<{ id: number }>;
    const availableEquipment = [overview.equipment.find((item) => item.status === "active"), ...overview.equipment].filter(Boolean) as Array<{ id: number }>;
    const sampleTypeOptions = laboratoryDictionaryOptions(overview.dictionaries, "laboratory_test_type");
    const qualityResultOptions = laboratoryDictionaryOptions(overview.dictionaries, "quality_result");
    const severityOptions = laboratoryDictionaryOptions(overview.dictionaries, "severity_level");
    setMixForm((value) => ({
      ...value,
      siteId: keepValidId(value.siteId, overview.sites),
      productId: keepValidId(value.productId, overview.products),
      materials: keepMixMaterials(value.materials, overview)
    }));
    setTrialForm((value) => ({
      ...value,
      mixDesignId: keepValidId(value.mixDesignId, overview.mixDesigns),
      result: keepDictionaryCode(value.result, qualityResultOptions)
    }));
    setSampleForm((value) => ({
      ...value,
      siteId: keepValidId(value.siteId, overview.sites),
      productId: keepValidId(value.productId, overview.products),
      mixDesignId: keepValidId(value.mixDesignId, overview.mixDesigns),
      sampleType: keepDictionaryCode(value.sampleType, sampleTypeOptions)
    }));
    setTestForm((value) => ({
      ...value,
      sampleId: keepValidId(value.sampleId, availableSamples),
      equipmentId: keepValidId(value.equipmentId, availableEquipment),
      result: keepDictionaryCode(value.result, qualityResultOptions)
    }));
    setEquipmentForm((value) => ({ ...value, siteId: keepValidId(value.siteId, overview.sites) }));
    setCalibrationForm((value) => ({
      ...value,
      equipmentId: keepValidId(value.equipmentId, availableEquipment),
      result: keepDictionaryCode(value.result, qualityResultOptions)
    }));
    setExceptionForm((value) => ({ ...value, severity: keepDictionaryCode(value.severity, severityOptions) }));
  }, [
    overview?.sites.length,
    overview?.products.length,
    overview?.materials.length,
    overview?.mixDesigns.length,
    overview?.samples.length,
    overview?.equipment.length,
    overview?.dictionaries.length
  ]);

  function laboratoryActionName(label: string) {
    if (label.startsWith("workflow-task-approve")) return "通过质量异常工作流任务";
    if (label.startsWith("workflow-task-reject")) return "驳回质量异常工作流任务";
    const names: Record<string, string> = {
      test: "登记并复核试验结果",
      calibration: "登记设备校准",
      exception: "提交质量异常",
      revise: "修订配比",
      approve: "审批配比",
      retire: "停用配比",
      approveProfile: "审批站点配比",
      retireProfile: "停用站点配比",
      "handle-exception": "关闭质量异常"
    };
    return names[label] || label;
  }

  async function mutate(label: string, action: () => Promise<unknown>) {
    const prompt = sensitiveActionPrompt(label, laboratoryActionName(label));
    if (prompt) {
      const confirmed = await confirmMessage({
        title: prompt.title,
        message: prompt.message,
        tone: "warning",
        confirmLabel: prompt.confirmLabel,
        confirmVariant: prompt.confirmVariant
      });
      if (!confirmed) return;
    }
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
  if (!overview) {
    return (
      <Panel>
        {error ? <Button onClick={load}>重新加载</Button> : "加载实验室工作台..."}
      </Panel>
    );
  }

  const loadedOverview = overview;
  const dictionaries = loadedOverview.dictionaries || [];

  function renderModule() {
    switch (activeModule) {
      case "mix-designs":
        return (
          <MixDesignModule
            mode="base"
            overview={loadedOverview}
            productOptions={productOptions}
            siteOptions={siteOptions}
            materialOptions={loadedOverview.materials}
            mixForm={mixForm}
            setMixForm={setMixForm}
            busy={busy}
            mutate={mutate}
            onReload={load}
            onSubmitMix={submitMix}
          />
        );
      case "plant-mix-designs":
        return (
          <MixDesignModule
            mode="plant"
            overview={loadedOverview}
            productOptions={productOptions}
            siteOptions={siteOptions}
            materialOptions={loadedOverview.materials}
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
            dictionaries={dictionaries}
            trialForm={trialForm}
            setTrialForm={setTrialForm}
            draftMixes={draftMixes}
            currentMixes={currentMixes}
            trialRuns={loadedOverview.trialRuns}
            busy={busy}
            onReload={load}
            onSubmitTrial={submitTrial}
          />
        );
      case "sample-tests":
        return (
          <SampleTestModule
            dictionaries={dictionaries}
            productOptions={productOptions}
            siteOptions={siteOptions}
            mixDesigns={loadedOverview.mixDesigns}
            samples={loadedOverview.samples}
            tests={loadedOverview.tests}
            equipment={loadedOverview.equipment}
            sampleForm={sampleForm}
            setSampleForm={setSampleForm}
            testForm={testForm}
            setTestForm={setTestForm}
            busy={busy}
            onReload={load}
            onSubmitSample={submitSample}
            onSubmitTest={submitTest}
          />
        );
      case "equipment-calibration":
        return (
          <EquipmentCalibrationModule
            dictionaries={dictionaries}
            equipment={loadedOverview.equipment}
            calibrations={loadedOverview.calibrations}
            siteOptions={siteOptions}
            equipmentForm={equipmentForm}
            setEquipmentForm={setEquipmentForm}
            calibrationForm={calibrationForm}
            setCalibrationForm={setCalibrationForm}
            busy={busy}
            onReload={load}
            onSubmitEquipment={submitEquipment}
            onSubmitCalibration={submitCalibration}
          />
        );
      case "sample-ledger":
        return (
          <SampleLedgerModule
            dictionaries={dictionaries}
            samples={loadedOverview.samples}
            tests={loadedOverview.tests}
            productOptions={productOptions}
            materials={materials}
            onReload={load}
          />
        );
      case "exceptions":
        return (
	          <ExceptionClosureModule
	            dictionaries={dictionaries}
	            exceptions={loadedOverview.exceptions}
	            workflowOverview={workflowOverview}
	            currentRoleCode={currentRoleCode}
	            currentPermissions={currentPermissions}
	            exceptionForm={exceptionForm}
	            setExceptionForm={setExceptionForm}
            busy={busy}
            mutate={mutate}
            onReload={load}
            onSubmitException={submitException}
          />
        );
      default:
        return null;
    }
  }

  return (
    <div className="view-stack laboratory-view">
      {renderModule()}
    </div>
  );
}
