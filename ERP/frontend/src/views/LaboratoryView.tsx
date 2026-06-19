import { Ban, CheckCircle2, FlaskConical, Plus, RotateCcw, TestTube2, Wrench, XCircle } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { KpiCard } from "../components/KpiCard";
import { nameOf } from "../components/names";
import { StatusChip } from "../components/StatusChip";
import { api } from "../services/api";
import type {
  LaboratoryOverview,
  LaboratorySample,
  LaboratoryTestRecord,
  Material,
  MixDesignMaterial,
  Product,
  QualityException
} from "../services/types";

const today = new Date().toISOString().slice(0, 10);

function parseNumber(value: string, fallbackValue = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallbackValue;
}

function parseMixMaterials(value: string): MixDesignMaterial[] {
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

function materialSummary(items: MixDesignMaterial[], materials: Material[]) {
  if (!items.length) return "-";
  return items.map((item) => `${nameOf(materials, item.materialId)} ${item.dosage}${item.unit}`).join(" / ");
}

function productName(products: Product[], id: number) {
  const product = products.find((item) => item.id === id);
  if (!product) return "-";
  return `${product.name} ${product.spec}`;
}

function latestTestForSample(sample: LaboratorySample, tests: LaboratoryTestRecord[]) {
  return tests.filter((item) => item.sampleId === sample.id).sort((a, b) => b.id - a.id)[0];
}

function exceptionSummary(items: QualityException[]) {
  const open = items.filter((item) => item.status !== "closed");
  return open.length ? `${open.length} 个待处理` : "已闭环";
}

export function LaboratoryView({ onChanged }: { onChanged: () => void }) {
  const [overview, setOverview] = useState<LaboratoryOverview | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [mixForm, setMixForm] = useState({
    productId: "1",
    siteId: "1",
    code: "MD-NEW",
    version: "v1",
    strengthGrade: "C30",
    slump: "180mm",
    scope: "站内标准生产配比",
    materials: "1:330:kg/m3\n3:780:kg/m3\n4:1020:kg/m3\n5:8.0:kg/m3"
  });
  const [trialForm, setTrialForm] = useState({
    mixDesignId: "1",
    strength7d: "32",
    strength28d: "42",
    water: "165",
    sandRate: "42",
    admixtureRate: "1.2",
    result: "passed"
  });
  const [sampleForm, setSampleForm] = useState({
    siteId: "1",
    productId: "1",
    mixDesignId: "1",
    sampleType: "compressive_strength",
    plannedTestAt: today
  });
  const [testForm, setTestForm] = useState({
    sampleId: "1",
    equipmentId: "1",
    metric: "28d_strength",
    value: "42",
    unit: "MPa",
    result: "passed"
  });
  const [equipmentForm, setEquipmentForm] = useState({
    name: "压力试验机",
    siteId: "1",
    model: "YES-2000",
    serialNo: "LAB-NEW-001",
    calibrationCycleDays: "180",
    lastCalibrationAt: today,
    nextCalibrationAt: "2026-12-31"
  });
  const [calibrationForm, setCalibrationForm] = useState({
    equipmentId: "1",
    result: "passed",
    calibratedAt: today,
    nextDueAt: "2026-12-31",
    certificateNo: "CAL-NEW",
    agency: "计量检测中心"
  });
  const [exceptionForm, setExceptionForm] = useState({
    title: "质量异常",
    severity: "medium",
    responsible: "实验室质检员",
    description: "现场反馈或试验异常待处理"
  });

  async function load() {
    setError("");
    setOverview(await api.laboratoryOverview());
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

  async function submitMix(event: FormEvent<HTMLFormElement>) {
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
  }

  async function submitTrial(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await mutate("trial", () => api.createMixDesignTrialRun(parseNumber(trialForm.mixDesignId), {
      water: parseNumber(trialForm.water),
      sandRate: parseNumber(trialForm.sandRate),
      admixtureRate: parseNumber(trialForm.admixtureRate),
      strength7d: parseNumber(trialForm.strength7d),
      strength28d: parseNumber(trialForm.strength28d),
      result: trialForm.result
    }));
  }

  async function submitSample(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await mutate("sample", () => api.createLaboratorySample({
      siteId: parseNumber(sampleForm.siteId),
      productId: parseNumber(sampleForm.productId),
      mixDesignId: parseNumber(sampleForm.mixDesignId),
      sampleType: sampleForm.sampleType,
      plannedTestAt: sampleForm.plannedTestAt
    }));
  }

  async function submitTest(event: FormEvent<HTMLFormElement>) {
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
  }

  async function submitEquipment(event: FormEvent<HTMLFormElement>) {
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
  }

  async function submitCalibration(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await mutate("calibration", () => api.createLaboratoryCalibration(parseNumber(calibrationForm.equipmentId), {
      result: calibrationForm.result,
      calibratedAt: calibrationForm.calibratedAt,
      nextDueAt: calibrationForm.nextDueAt,
      certificateNo: calibrationForm.certificateNo,
      agency: calibrationForm.agency
    }));
  }

  async function submitException(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await mutate("exception", () => api.createQualityException({
      title: exceptionForm.title,
      severity: exceptionForm.severity,
      responsible: exceptionForm.responsible,
      description: exceptionForm.description,
      siteId: parseNumber(sampleForm.siteId)
    }));
  }

  const productOptions = overview?.products || [];
  const siteOptions = overview?.sites || [];
  const materials = overview?.materials || [];
  const currentMixes = useMemo(() => overview?.mixDesigns.filter((item) => item.isCurrent && item.status === "approved") || [], [overview]);
  const draftMixes = useMemo(() => overview?.mixDesigns.filter((item) => item.status === "draft" || item.status === "pending_approval") || [], [overview]);
  const openExceptions = overview?.exceptions.filter((item) => item.status !== "closed") || [];

  if (!overview) {
    return <section className="panel">加载实验室工作台...</section>;
  }

  return (
    <div className="view-stack laboratory-view">
      <section className="kpi-grid compact">
        <KpiCard label="当前配比" value={overview.kpis.currentMixDesigns} />
        <KpiCard label="待审批配比" value={overview.kpis.pendingMixDesigns} />
        <KpiCard label="样品待试验" value={overview.kpis.pendingSamples} />
        <KpiCard label="试验待复核" value={overview.kpis.pendingReviews} />
        <KpiCard label="仪器临期" value={overview.kpis.calibrationDue + overview.kpis.calibrationOverdue} />
      </section>
      {error ? <p className="error-text">{error}</p> : null}

      <section className="grid-12">
        <div className="panel span-8">
          <div className="between">
            <div>
              <h3>配比版本管理</h3>
              <p className="muted">生产任务默认使用同站点当前生效配比，历史批次保留原配比版本</p>
            </div>
            <button className="soft-button icon-button-text" type="button" onClick={() => load()}>
              <RotateCcw size={16} />刷新
            </button>
          </div>
          <div className="lab-card-grid">
            {overview.mixDesigns.map((mix) => (
              <article className="lab-card" key={mix.id}>
                <div className="between compact-row">
                  <div>
                    <strong>{mix.code} {mix.version}</strong>
                    <p className="muted">{productName(productOptions, mix.productId)} · {mix.strengthGrade || "-"} · {mix.slump || "-"}</p>
                  </div>
                  <StatusChip value={mix.isCurrent ? "active" : mix.status} />
                </div>
                <p className="muted">{nameOf(siteOptions, mix.siteId)} · {mix.scope || "通用配比"}</p>
                <p className="muted">{materialSummary(mix.materials, materials)}</p>
                <div className="row-actions">
                  <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => mutate("approve", () => api.approveLaboratoryMixDesign(mix.id, { effectiveFrom: today }))}>
                    <CheckCircle2 size={15} />审批生效
                  </button>
                  <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => mutate("revise", () => api.reviseLaboratoryMixDesign(mix.id, { version: `${mix.version}-rev` }))}>
                    <Plus size={15} />修订
                  </button>
                  <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => mutate("retire", () => api.retireLaboratoryMixDesign(mix.id))}>
                    <Ban size={15} />停用
                  </button>
                </div>
              </article>
            ))}
          </div>
        </div>

        <div className="panel span-4">
          <h3>新增配比</h3>
          <form className="lab-form" onSubmit={submitMix}>
            <label>
              <span>产品</span>
              <select value={mixForm.productId} onChange={(event) => setMixForm({ ...mixForm, productId: event.target.value })}>
                {productOptions.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
              </select>
            </label>
            <label>
              <span>站点</span>
              <select value={mixForm.siteId} onChange={(event) => setMixForm({ ...mixForm, siteId: event.target.value })}>
                {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>编号</span><input value={mixForm.code} onChange={(event) => setMixForm({ ...mixForm, code: event.target.value })} /></label>
            <label><span>版本</span><input value={mixForm.version} onChange={(event) => setMixForm({ ...mixForm, version: event.target.value })} /></label>
            <label><span>强度等级</span><input value={mixForm.strengthGrade} onChange={(event) => setMixForm({ ...mixForm, strengthGrade: event.target.value })} /></label>
            <label><span>坍落度</span><input value={mixForm.slump} onChange={(event) => setMixForm({ ...mixForm, slump: event.target.value })} /></label>
            <label className="span-all"><span>适用范围</span><input value={mixForm.scope} onChange={(event) => setMixForm({ ...mixForm, scope: event.target.value })} /></label>
            <label className="span-all"><span>材料用量</span><textarea value={mixForm.materials} onChange={(event) => setMixForm({ ...mixForm, materials: event.target.value })} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><Plus size={16} />保存配比</button>
          </form>
        </div>
      </section>

      <section className="grid-12">
        <div className="panel span-4">
          <h3>试配记录</h3>
          <form className="lab-form" onSubmit={submitTrial}>
            <label>
              <span>配比</span>
              <select value={trialForm.mixDesignId} onChange={(event) => setTrialForm({ ...trialForm, mixDesignId: event.target.value })}>
                {[...draftMixes, ...currentMixes].map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
              </select>
            </label>
            <label><span>7d 强度</span><input value={trialForm.strength7d} onChange={(event) => setTrialForm({ ...trialForm, strength7d: event.target.value })} /></label>
            <label><span>28d 强度</span><input value={trialForm.strength28d} onChange={(event) => setTrialForm({ ...trialForm, strength28d: event.target.value })} /></label>
            <label><span>用水量</span><input value={trialForm.water} onChange={(event) => setTrialForm({ ...trialForm, water: event.target.value })} /></label>
            <label><span>砂率</span><input value={trialForm.sandRate} onChange={(event) => setTrialForm({ ...trialForm, sandRate: event.target.value })} /></label>
            <label><span>外加剂率</span><input value={trialForm.admixtureRate} onChange={(event) => setTrialForm({ ...trialForm, admixtureRate: event.target.value })} /></label>
            <label>
              <span>结论</span>
              <select value={trialForm.result} onChange={(event) => setTrialForm({ ...trialForm, result: event.target.value })}>
                <option value="passed">合格</option>
                <option value="failed">不合格</option>
              </select>
            </label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><FlaskConical size={16} />记录试配</button>
          </form>
          <div className="record-list compact-row">
            {overview.trialRuns.slice(0, 4).map((item) => (
              <div className="record-card" key={item.id}>
                <strong>{item.trialNo}</strong>
                <p>{item.targetStrength} · 28d {item.strength28d}MPa · <StatusChip value={item.result} /></p>
              </div>
            ))}
          </div>
        </div>

        <div className="panel span-4">
          <h3>样品试验</h3>
          <form className="lab-form" onSubmit={submitSample}>
            <label>
              <span>产品</span>
              <select value={sampleForm.productId} onChange={(event) => setSampleForm({ ...sampleForm, productId: event.target.value })}>
                {productOptions.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
              </select>
            </label>
            <label>
              <span>配比</span>
              <select value={sampleForm.mixDesignId} onChange={(event) => setSampleForm({ ...sampleForm, mixDesignId: event.target.value })}>
                {overview.mixDesigns.map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
              </select>
            </label>
            <label><span>计划试验</span><input value={sampleForm.plannedTestAt} onChange={(event) => setSampleForm({ ...sampleForm, plannedTestAt: event.target.value })} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><Plus size={16} />登记样品</button>
          </form>
          <form className="lab-form lab-form-divider" onSubmit={submitTest}>
            <label>
              <span>样品</span>
              <select value={testForm.sampleId} onChange={(event) => setTestForm({ ...testForm, sampleId: event.target.value })}>
                {overview.samples.map((item) => <option key={item.id} value={item.id}>{item.sampleNo} · {item.sampleType}</option>)}
              </select>
            </label>
            <label>
              <span>仪器</span>
              <select value={testForm.equipmentId} onChange={(event) => setTestForm({ ...testForm, equipmentId: event.target.value })}>
                {overview.equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>指标</span><input value={testForm.metric} onChange={(event) => setTestForm({ ...testForm, metric: event.target.value })} /></label>
            <label><span>结果值</span><input value={testForm.value} onChange={(event) => setTestForm({ ...testForm, value: event.target.value })} /></label>
            <label>
              <span>判定</span>
              <select value={testForm.result} onChange={(event) => setTestForm({ ...testForm, result: event.target.value })}>
                <option value="passed">合格</option>
                <option value="failed">不合格</option>
              </select>
            </label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><TestTube2 size={16} />试验并复核</button>
          </form>
        </div>

        <div className="panel span-4">
          <h3>仪器校准</h3>
          <form className="lab-form" onSubmit={submitEquipment}>
            <label><span>名称</span><input value={equipmentForm.name} onChange={(event) => setEquipmentForm({ ...equipmentForm, name: event.target.value })} /></label>
            <label><span>型号</span><input value={equipmentForm.model} onChange={(event) => setEquipmentForm({ ...equipmentForm, model: event.target.value })} /></label>
            <label><span>序列号</span><input value={equipmentForm.serialNo} onChange={(event) => setEquipmentForm({ ...equipmentForm, serialNo: event.target.value })} /></label>
            <label><span>下次校准</span><input value={equipmentForm.nextCalibrationAt} onChange={(event) => setEquipmentForm({ ...equipmentForm, nextCalibrationAt: event.target.value })} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><Wrench size={16} />登记仪器</button>
          </form>
          <form className="lab-form lab-form-divider" onSubmit={submitCalibration}>
            <label>
              <span>仪器</span>
              <select value={calibrationForm.equipmentId} onChange={(event) => setCalibrationForm({ ...calibrationForm, equipmentId: event.target.value })}>
                {overview.equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label><span>证书号</span><input value={calibrationForm.certificateNo} onChange={(event) => setCalibrationForm({ ...calibrationForm, certificateNo: event.target.value })} /></label>
            <label><span>下次到期</span><input value={calibrationForm.nextDueAt} onChange={(event) => setCalibrationForm({ ...calibrationForm, nextDueAt: event.target.value })} /></label>
            <button className="soft-button icon-button-text" type="submit" disabled={busy !== ""}><CheckCircle2 size={16} />保存校准</button>
          </form>
        </div>
      </section>

      <section className="grid-12">
        <div className="panel span-8">
          <div className="between">
            <h3>样品台账与试验结果</h3>
            <span className="muted">{overview.samples.length} 个样品 · {overview.tests.length} 条试验</span>
          </div>
          <table>
            <thead>
              <tr><th>样品</th><th>来源</th><th>产品/物料</th><th>计划</th><th>最新试验</th><th>状态</th></tr>
            </thead>
            <tbody>
              {overview.samples.slice(0, 8).map((sample) => {
                const test = latestTestForSample(sample, overview.tests);
                return (
                  <tr key={sample.id}>
                    <td>{sample.sampleNo}</td>
                    <td>{sample.sourceType}</td>
                    <td>{sample.productId ? productName(productOptions, sample.productId) : nameOf(materials, sample.materialId)}</td>
                    <td>{sample.plannedTestAt || "-"}</td>
                    <td>{test ? `${test.metric} ${test.value}${test.unit}` : "-"}</td>
                    <td><StatusChip value={sample.result || sample.status} /></td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div className="panel span-4">
          <div className="between">
            <h3>异常闭环</h3>
            <span className="muted">{exceptionSummary(overview.exceptions)}</span>
          </div>
          <form className="lab-form" onSubmit={submitException}>
            <label><span>标题</span><input value={exceptionForm.title} onChange={(event) => setExceptionForm({ ...exceptionForm, title: event.target.value })} /></label>
            <label>
              <span>等级</span>
              <select value={exceptionForm.severity} onChange={(event) => setExceptionForm({ ...exceptionForm, severity: event.target.value })}>
                <option value="low">低</option>
                <option value="medium">中</option>
                <option value="high">高</option>
              </select>
            </label>
            <label><span>责任人</span><input value={exceptionForm.responsible} onChange={(event) => setExceptionForm({ ...exceptionForm, responsible: event.target.value })} /></label>
            <label className="span-all"><span>描述</span><textarea value={exceptionForm.description} onChange={(event) => setExceptionForm({ ...exceptionForm, description: event.target.value })} /></label>
            <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><XCircle size={16} />创建异常</button>
          </form>
          <div className="record-list compact-row">
            {openExceptions.slice(0, 5).map((item) => (
              <article className="record-card" key={item.id}>
                <div className="between compact-row">
                  <strong>{item.title}</strong>
                  <StatusChip value={item.severity} />
                </div>
                <p>{item.responsible || "-"} · {item.createdAt}</p>
                <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => mutate("handle-exception", () => api.handleQualityException(item.id, { rootCause: "已复核原因", correctiveAction: "已完成纠正措施" }))}>
                  <CheckCircle2 size={15} />关闭
                </button>
              </article>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
