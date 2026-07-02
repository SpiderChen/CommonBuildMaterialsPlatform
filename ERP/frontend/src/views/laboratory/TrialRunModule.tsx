import { FlaskConical } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { StatusChip } from "../../components/StatusChip";
import type { MixDesign, MixDesignTrialRun } from "../../services/types";
import type { SubmitHandler, TrialForm } from "./LaboratoryModuleTypes";

type Props = {
  trialForm: TrialForm;
  setTrialForm: Dispatch<SetStateAction<TrialForm>>;
  draftMixes: MixDesign[];
  currentMixes: MixDesign[];
  trialRuns: MixDesignTrialRun[];
  busy: string;
  onSubmitTrial: SubmitHandler;
};

export function TrialRunModule({
  trialForm,
  setTrialForm,
  draftMixes,
  currentMixes,
  trialRuns,
  busy,
  onSubmitTrial
}: Props) {
  const mixOptions = [...draftMixes, ...currentMixes];

  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-5">
        <h3>试配记录</h3>
        <form className="lab-form" onSubmit={onSubmitTrial}>
          <label>
            <span>配比</span>
            <select value={trialForm.mixDesignId} onChange={(event) => setTrialForm({ ...trialForm, mixDesignId: event.target.value })}>
              {mixOptions.map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
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
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== "" || !mixOptions.length}><FlaskConical size={16} />记录试配</button>
        </form>
      </div>

      <div className="panel span-7">
        <div className="between">
          <h3>试配台账</h3>
          <span className="muted">{trialRuns.length} 条记录</span>
        </div>
        <div className="record-list compact-row">
          {trialRuns.slice(0, 8).map((item) => (
            <div className="record-card" key={item.id}>
              <strong>{item.trialNo}</strong>
              <p>{item.targetStrength} · 7d {item.strength7d}MPa · 28d {item.strength28d}MPa · <StatusChip value={item.result} /></p>
              <p className="muted">{item.tester || "-"} · {item.testedAt || item.createdAt || "-"}</p>
            </div>
          ))}
          {!trialRuns.length ? <p className="muted">暂无试配记录</p> : null}
        </div>
      </div>
    </section>
  );
}
