import { nameOf } from "../../components/names";
import { StatusChip } from "../../components/StatusChip";
import type { LaboratorySample, LaboratoryTestRecord, Material, Product } from "../../services/types";
import { latestTestForSample, productName } from "./laboratoryHelpers";

type Props = {
  samples: LaboratorySample[];
  tests: LaboratoryTestRecord[];
  productOptions: Product[];
  materials: Material[];
};

export function SampleLedgerModule({ samples, tests, productOptions, materials }: Props) {
  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-12">
        <div className="between">
          <h3>样品台账与试验结果</h3>
          <span className="muted">{samples.length} 个样品 · {tests.length} 条试验</span>
        </div>
        <table>
          <thead>
            <tr><th>样品</th><th>来源</th><th>产品/物料</th><th>计划</th><th>最新试验</th><th>状态</th></tr>
          </thead>
          <tbody>
            {samples.slice(0, 12).map((sample) => {
              const test = latestTestForSample(sample, tests);
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
            {!samples.length ? <tr><td colSpan={6}>暂无样品</td></tr> : null}
          </tbody>
        </table>
      </div>
    </section>
  );
}
