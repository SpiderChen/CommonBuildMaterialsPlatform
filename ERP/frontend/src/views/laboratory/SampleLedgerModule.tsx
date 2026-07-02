import { useMemo, useState } from "react";
import { ChipButton, DataTable, Panel, SectionGrid, StatusChip, buildDataTableRowContextMenu } from "../../components";
import type { DataDictionary, LaboratorySample, LaboratoryTestRecord, Material, Product } from "../../services/types";
import {
  latestTestForSample,
  sampleSubject,
  sampleTypeLabel,
  shortDate,
  sourceTypeLabel
} from "./laboratoryHelpers";

type Props = {
  dictionaries: DataDictionary[];
  samples: LaboratorySample[];
  tests: LaboratoryTestRecord[];
  productOptions: Product[];
  materials: Material[];
  onReload: () => Promise<void>;
};

export function SampleLedgerModule({ dictionaries, samples, tests, productOptions, materials, onReload }: Props) {
  const [statusFilter, setStatusFilter] = useState("all");
  const filteredSamples = useMemo(() => {
    if (statusFilter === "all") return samples;
    if (statusFilter === "pending") return samples.filter((sample) => sample.status !== "completed");
    if (statusFilter === "failed") return samples.filter((sample) => (sample.result || "") === "failed" || latestTestForSample(sample, tests)?.result === "failed");
    return samples.filter((sample) => sample.status === statusFilter);
  }, [samples, statusFilter, tests]);

  return (
    <SectionGrid className="laboratory-module">
      <Panel className="span-12 sample-ledger-panel">
        <DataTable
          data={filteredSamples}
          rowKey={(sample) => sample.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<LaboratorySample>({
            actions: [
              {
                key: "focus-sample",
                label: "只看该样品",
                onSelect: (sample, helpers) => helpers.searchText(sample.sampleNo)
              },
              {
                key: "focus-source",
                label: "只看同来源",
                onSelect: (sample, helpers) => helpers.searchText(sourceTypeLabel(sample.sourceType, dictionaries))
              }
            ],
            copyFields: [
              { key: "sample", label: "样品编号", value: (sample) => sample.sampleNo },
              { key: "subject", label: "产品/材料", value: (sample) => sampleSubject(sample, productOptions, materials) },
              { key: "source", label: "来源", value: (sample) => sourceTypeLabel(sample.sourceType, dictionaries) },
              { key: "latest", label: "最新试验", value: (sample) => {
                const test = latestTestForSample(sample, tests);
                return test ? `${test.metric} ${test.value}${test.unit}` : "";
              } }
            ]
          })}
          headerAction={
            <>
              <div className="lab-filter-group ledger-filter-group">
                {[
                  ["all", "全部"],
                  ["pending", "待完成"],
                  ["completed", "已完成"],
                  ["failed", "异常"]
                ].map(([value, label]) => (
                  <ChipButton active={statusFilter === value} key={value} onClick={() => setStatusFilter(value)}>
                    {label}
                  </ChipButton>
                ))}
              </div>
              <span className="muted">{filteredSamples.length} 个样品 · {tests.length} 条试验</span>
            </>
          }
          columns={[
            { key: "sample", title: "样品", render: (sample) => sample.sampleNo },
            { key: "source", title: "来源", render: (sample) => sourceTypeLabel(sample.sourceType, dictionaries) },
            { key: "type", title: "类型", render: (sample) => sampleTypeLabel(sample.sampleType, dictionaries) },
            { key: "product", title: "产品/材料", render: (sample) => sampleSubject(sample, productOptions, materials) },
            { key: "plan", title: "计划", render: (sample) => shortDate(sample.plannedTestAt) },
            { key: "latest", title: "最新试验", render: (sample) => {
              const test = latestTestForSample(sample, tests);
              return test ? `${test.metric} ${test.value}${test.unit}` : "-";
            } },
            { key: "status", title: "状态", render: (sample) => <StatusChip value={sample.result || sample.status} /> }
          ]}
          emptyText="暂无样品"
          pageSize={10}
          showPagination={true}
        />
      </Panel>
    </SectionGrid>
  );
}
