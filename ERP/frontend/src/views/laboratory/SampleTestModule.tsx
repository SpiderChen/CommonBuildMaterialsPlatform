import { Plus, TestTube2 } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import type { LaboratoryEquipment, LaboratorySample, MixDesign, Product } from "../../services/types";
import type { SampleForm, SubmitHandler, TestForm } from "./LaboratoryModuleTypes";

type Props = {
  productOptions: Product[];
  mixDesigns: MixDesign[];
  samples: LaboratorySample[];
  equipment: LaboratoryEquipment[];
  sampleForm: SampleForm;
  setSampleForm: Dispatch<SetStateAction<SampleForm>>;
  testForm: TestForm;
  setTestForm: Dispatch<SetStateAction<TestForm>>;
  busy: string;
  onSubmitSample: SubmitHandler;
  onSubmitTest: SubmitHandler;
};

export function SampleTestModule({
  productOptions,
  mixDesigns,
  samples,
  equipment,
  sampleForm,
  setSampleForm,
  testForm,
  setTestForm,
  busy,
  onSubmitSample,
  onSubmitTest
}: Props) {
  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-6">
        <h3>样品登记</h3>
        <form className="lab-form" onSubmit={onSubmitSample}>
          <label>
            <span>产品</span>
            <select value={sampleForm.productId} onChange={(event) => setSampleForm({ ...sampleForm, productId: event.target.value })}>
              {productOptions.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
            </select>
          </label>
          <label>
            <span>配比</span>
            <select value={sampleForm.mixDesignId} onChange={(event) => setSampleForm({ ...sampleForm, mixDesignId: event.target.value })}>
              {mixDesigns.map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
            </select>
          </label>
          <label><span>计划试验</span><input value={sampleForm.plannedTestAt} onChange={(event) => setSampleForm({ ...sampleForm, plannedTestAt: event.target.value })} /></label>
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== "" || !productOptions.length || !mixDesigns.length}><Plus size={16} />登记样品</button>
        </form>
      </div>

      <div className="panel span-6">
        <h3>试验复核</h3>
        <form className="lab-form" onSubmit={onSubmitTest}>
          <label>
            <span>样品</span>
            <select value={testForm.sampleId} onChange={(event) => setTestForm({ ...testForm, sampleId: event.target.value })}>
              {samples.map((item) => <option key={item.id} value={item.id}>{item.sampleNo} · {item.sampleType}</option>)}
            </select>
          </label>
          <label>
            <span>仪器</span>
            <select value={testForm.equipmentId} onChange={(event) => setTestForm({ ...testForm, equipmentId: event.target.value })}>
              {equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
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
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== "" || !samples.length || !equipment.length}><TestTube2 size={16} />试验并复核</button>
        </form>
      </div>
    </section>
  );
}
