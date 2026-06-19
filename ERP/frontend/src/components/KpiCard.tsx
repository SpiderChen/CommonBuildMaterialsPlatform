export function KpiCard({ label, value, suffix }: { label: string; value: string | number; suffix?: string }) {
  return (
    <section className="kpi-card panel">
      <span>{label}</span>
      <strong>
        {value}
        {suffix ? <small>{suffix}</small> : null}
      </strong>
    </section>
  );
}
