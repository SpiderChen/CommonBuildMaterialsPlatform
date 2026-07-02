import { Panel } from "./Surface";

export type KpiCardProps = {
  label: string;
  value: string | number;
  suffix?: string;
};

export function KpiCard({ label, value, suffix }: KpiCardProps) {
  return (
    <Panel className="kpi-card">
      <span>{label}</span>
      <strong>
        {value}
        {suffix ? <small>{suffix}</small> : null}
      </strong>
    </Panel>
  );
}
