import { KpiCard } from "../../components/KpiCard";
import type { LaboratoryOverview } from "../../services/types";

export function LaboratoryKpiStrip({ overview }: { overview: LaboratoryOverview }) {
  return (
    <section className="kpi-grid compact">
      <KpiCard label="当前配比" value={overview.kpis.currentMixDesigns} />
      <KpiCard label="待审批配比" value={overview.kpis.pendingMixDesigns} />
      <KpiCard label="样品待试验" value={overview.kpis.pendingSamples} />
      <KpiCard label="试验待复核" value={overview.kpis.pendingReviews} />
      <KpiCard label="仪器临期" value={overview.kpis.calibrationDue + overview.kpis.calibrationOverdue} />
    </section>
  );
}
