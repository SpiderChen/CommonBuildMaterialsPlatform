import { Ban, CheckCircle2, Plus, RotateCcw } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { nameOf } from "../../components/names";
import { StatusChip } from "../../components/StatusChip";
import { api } from "../../services/api";
import type { LaboratoryOverview, Material, Product, Site } from "../../services/types";
import type { MixForm, MutateAction, SubmitHandler } from "./LaboratoryModuleTypes";
import { materialSummary, productName, today } from "./laboratoryHelpers";

type Props = {
  overview: LaboratoryOverview;
  productOptions: Product[];
  siteOptions: Site[];
  materials: Material[];
  mixForm: MixForm;
  setMixForm: Dispatch<SetStateAction<MixForm>>;
  busy: string;
  mutate: MutateAction;
  onReload: () => Promise<void>;
  onSubmitMix: SubmitHandler;
};

export function MixDesignModule({
  overview,
  productOptions,
  siteOptions,
  materials,
  mixForm,
  setMixForm,
  busy,
  mutate,
  onReload,
  onSubmitMix
}: Props) {
  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-8">
        <div className="between">
          <div>
            <h3>配比版本管理</h3>
            <p className="muted">生产任务默认使用同站点当前生效配比，历史批次保留原配比版本</p>
          </div>
          <button className="soft-button icon-button-text" type="button" onClick={() => void onReload()}>
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
                <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => void mutate("approve", () => api.approveLaboratoryMixDesign(mix.id, { effectiveFrom: today }))}>
                  <CheckCircle2 size={15} />审批生效
                </button>
                <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => void mutate("revise", () => api.reviseLaboratoryMixDesign(mix.id, { version: `${mix.version}-rev` }))}>
                  <Plus size={15} />修订
                </button>
                <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => void mutate("retire", () => api.retireLaboratoryMixDesign(mix.id))}>
                  <Ban size={15} />停用
                </button>
              </div>
            </article>
          ))}
          {!overview.mixDesigns.length ? <p className="muted">暂无配比版本</p> : null}
        </div>
      </div>

      <div className="panel span-4">
        <h3>新增配比</h3>
        <form className="lab-form" onSubmit={onSubmitMix}>
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
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== "" || !productOptions.length || !siteOptions.length}><Plus size={16} />保存配比</button>
        </form>
      </div>
    </section>
  );
}
