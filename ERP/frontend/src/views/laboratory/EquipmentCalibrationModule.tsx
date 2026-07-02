import { CheckCircle2, Wrench } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { StatusChip } from "../../components/StatusChip";
import type { LaboratoryCalibration, LaboratoryEquipment } from "../../services/types";
import type { CalibrationForm, EquipmentForm, SubmitHandler } from "./LaboratoryModuleTypes";

type Props = {
  equipment: LaboratoryEquipment[];
  calibrations: LaboratoryCalibration[];
  equipmentForm: EquipmentForm;
  setEquipmentForm: Dispatch<SetStateAction<EquipmentForm>>;
  calibrationForm: CalibrationForm;
  setCalibrationForm: Dispatch<SetStateAction<CalibrationForm>>;
  busy: string;
  onSubmitEquipment: SubmitHandler;
  onSubmitCalibration: SubmitHandler;
};

export function EquipmentCalibrationModule({
  equipment,
  calibrations,
  equipmentForm,
  setEquipmentForm,
  calibrationForm,
  setCalibrationForm,
  busy,
  onSubmitEquipment,
  onSubmitCalibration
}: Props) {
  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-4">
        <h3>仪器登记</h3>
        <form className="lab-form" onSubmit={onSubmitEquipment}>
          <label><span>名称</span><input value={equipmentForm.name} onChange={(event) => setEquipmentForm({ ...equipmentForm, name: event.target.value })} /></label>
          <label><span>型号</span><input value={equipmentForm.model} onChange={(event) => setEquipmentForm({ ...equipmentForm, model: event.target.value })} /></label>
          <label><span>序列号</span><input value={equipmentForm.serialNo} onChange={(event) => setEquipmentForm({ ...equipmentForm, serialNo: event.target.value })} /></label>
          <label><span>下次校准</span><input value={equipmentForm.nextCalibrationAt} onChange={(event) => setEquipmentForm({ ...equipmentForm, nextCalibrationAt: event.target.value })} /></label>
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><Wrench size={16} />登记仪器</button>
        </form>
      </div>

      <div className="panel span-4">
        <h3>校准记录</h3>
        <form className="lab-form" onSubmit={onSubmitCalibration}>
          <label>
            <span>仪器</span>
            <select value={calibrationForm.equipmentId} onChange={(event) => setCalibrationForm({ ...calibrationForm, equipmentId: event.target.value })}>
              {equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </select>
          </label>
          <label><span>证书号</span><input value={calibrationForm.certificateNo} onChange={(event) => setCalibrationForm({ ...calibrationForm, certificateNo: event.target.value })} /></label>
          <label><span>下次到期</span><input value={calibrationForm.nextDueAt} onChange={(event) => setCalibrationForm({ ...calibrationForm, nextDueAt: event.target.value })} /></label>
          <button className="soft-button icon-button-text" type="submit" disabled={busy !== "" || !equipment.length}><CheckCircle2 size={16} />保存校准</button>
        </form>
      </div>

      <div className="panel span-4">
        <div className="between">
          <h3>仪器状态</h3>
          <span className="muted">{equipment.length} 台仪器</span>
        </div>
        <div className="record-list compact-row">
          {equipment.slice(0, 6).map((item) => (
            <article className="record-card" key={item.id}>
              <div className="between compact-row">
                <strong>{item.name}</strong>
                <StatusChip value={item.status} />
              </div>
              <p>{item.equipmentNo} · {item.model || "-"} · 下次 {item.nextCalibrationAt || "-"}</p>
            </article>
          ))}
          {!equipment.length ? <p className="muted">暂无仪器</p> : null}
          {calibrations.slice(0, 3).map((item) => (
            <article className="record-card" key={`cal-${item.id}`}>
              <strong>{item.calibrationNo}</strong>
              <p>{item.certificateNo || "-"} · {item.nextDueAt || "-"} · <StatusChip value={item.result} /></p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
