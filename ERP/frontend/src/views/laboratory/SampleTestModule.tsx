import { Plus, TestTube2 } from "lucide-react";
import { useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { ActionGroup, Button, DataTable, Dialog, DialogForm, Field, FormActions, HeroDateField, Panel, SectionGrid, SelectInput, StatusChip, TextInput, buildDataTableRowContextMenu } from "../../components";
import type { DataDictionary, LaboratoryEquipment, LaboratorySample, LaboratoryTestRecord, MixDesign, Product, Site } from "../../services/types";
import type { SampleForm, SubmitHandler, TestForm } from "./LaboratoryModuleTypes";
import { laboratoryDictionaryOptions, latestTestForSample, sampleTypeLabel, shortDate, sourceTypeLabel } from "./laboratoryHelpers";

type Props = {
  dictionaries: DataDictionary[];
  productOptions: Product[];
  siteOptions: Site[];
  mixDesigns: MixDesign[];
  samples: LaboratorySample[];
  tests: LaboratoryTestRecord[];
  equipment: LaboratoryEquipment[];
  sampleForm: SampleForm;
  setSampleForm: Dispatch<SetStateAction<SampleForm>>;
  testForm: TestForm;
  setTestForm: Dispatch<SetStateAction<TestForm>>;
  busy: string;
  onReload: () => Promise<void>;
  onSubmitSample: SubmitHandler;
  onSubmitTest: SubmitHandler;
};

export function SampleTestModule({
  dictionaries,
  productOptions,
  siteOptions,
  mixDesigns,
  samples,
  tests,
  equipment,
  sampleForm,
  setSampleForm,
  testForm,
  setTestForm,
  busy,
  onReload,
  onSubmitSample,
  onSubmitTest
}: Props) {
  const recentSamples = useMemo(() => [...samples].sort((a, b) => b.id - a.id), [samples]);
  const sampleTypeOptions = laboratoryDictionaryOptions(dictionaries, "laboratory_test_type");
  const qualityResultOptions = laboratoryDictionaryOptions(dictionaries, "quality_result");

  const [sampleDialogOpen, setSampleDialogOpen] = useState(false);
  const [testDialogOpen, setTestDialogOpen] = useState(false);

  function closeSampleDialog() {
    setSampleDialogOpen(false);
  }

  function closeTestDialog() {
    setTestDialogOpen(false);
  }

  return (
    <SectionGrid className="laboratory-module">
      <Panel className="span-12">
        <DataTable
          data={recentSamples}
          rowKey={(item) => item.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<LaboratorySample>({
            actions: [
              {
                key: "review-sample",
                label: "录入该样品复核",
                disabled: () => busy !== "" || !equipment.length,
                onSelect: (sample) => {
                  setTestForm({ ...testForm, sampleId: String(sample.id) });
                  setTestDialogOpen(true);
                }
              },
              {
                key: "create-same-sample",
                label: "新增同类样品",
                disabled: () => busy !== "",
                onSelect: (sample) => {
                  setSampleForm({
                    ...sampleForm,
                    siteId: String(sample.siteId || sampleForm.siteId),
                    productId: String(sample.productId || sampleForm.productId),
                    mixDesignId: String(sample.mixDesignId || sampleForm.mixDesignId),
                    sampleType: sample.sampleType || sampleForm.sampleType
                  });
                  setSampleDialogOpen(true);
                }
              }
            ],
            copyFields: [
              { key: "sample", label: "样品编号", value: (sample) => sample.sampleNo },
              { key: "type", label: "样品类型", value: (sample) => sampleTypeLabel(sample.sampleType, dictionaries) },
              { key: "source", label: "来源", value: (sample) => sourceTypeLabel(sample.sourceType, dictionaries) },
              { key: "latest", label: "最新结果", value: (sample) => {
                const latest = latestTestForSample(sample, tests);
                return latest ? `${latest.metric} ${latest.value}${latest.unit}` : "待录入";
              } }
            ]
          })}
          columns={[
            { key: "sampleNo", title: "样品编号", render: (sample) => sample.sampleNo },
            { key: "type", title: "类型", render: (sample) => sampleTypeLabel(sample.sampleType, dictionaries) },
            { key: "source", title: "来源", render: (sample) => sourceTypeLabel(sample.sourceType, dictionaries) },
            { key: "plan", title: "计划时间", render: (sample) => shortDate(sample.plannedTestAt) },
            {
              key: "latest",
              title: "最新结果",
              render: (sample) => {
                const latest = latestTestForSample(sample, tests);
                return latest ? `${latest.metric} ${latest.value}${latest.unit}` : "待录入";
              }
            },
            { key: "status", title: "状态", render: (sample) => <StatusChip value={sample.result || sample.status} /> }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<Plus size={16} />} onClick={() => setSampleDialogOpen(true)} disabled={busy !== ""}>新增样品</Button>
              <Button variant="primary" icon={<TestTube2 size={16} />} onClick={() => setTestDialogOpen(true)} disabled={busy !== "" || !samples.length}>试验复核</Button>
            </ActionGroup>
          }
          emptyText="暂无样品"
          pageSize={10}
          showPagination={true}
        />
      </Panel>

      <Dialog open={sampleDialogOpen} title="新增样品" className="master-dialog" closeDisabled={busy !== ""} onClose={closeSampleDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitSample(event);
                closeSampleDialog();
              }}
            >
              <Field label="站点">
                <SelectInput value={sampleForm.siteId} onChange={(event) => setSampleForm({ ...sampleForm, siteId: event.target.value })}>
                  {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="产品">
                <SelectInput value={sampleForm.productId} onChange={(event) => setSampleForm({ ...sampleForm, productId: event.target.value })}>
                  {productOptions.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
                </SelectInput>
              </Field>
              <Field label="配比">
                <SelectInput value={sampleForm.mixDesignId} onChange={(event) => setSampleForm({ ...sampleForm, mixDesignId: event.target.value })}>
                  {mixDesigns.map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
                </SelectInput>
              </Field>
              <Field label="样品类型">
                <SelectInput value={sampleForm.sampleType} onChange={(event) => setSampleForm({ ...sampleForm, sampleType: event.target.value })}>
                  {sampleTypeOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
                <HeroDateField label="计划检测日期" value={sampleForm.plannedTestAt} onChange={(plannedTestAt) => setSampleForm({ ...sampleForm, plannedTestAt })} />
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeSampleDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<Plus size={16} />} disabled={busy !== "" || !productOptions.length || !mixDesigns.length}>保存样品</Button>
              </FormActions>
            </DialogForm>
      </Dialog>

      <Dialog open={testDialogOpen} title="试验复核" className="master-dialog" closeDisabled={busy !== ""} onClose={closeTestDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitTest(event);
                closeTestDialog();
              }}
            >
              <Field label="样品">
                <SelectInput value={testForm.sampleId} onChange={(event) => setTestForm({ ...testForm, sampleId: event.target.value })}>
                  {samples.map((item) => <option key={item.id} value={item.id}>{item.sampleNo} · {item.sampleType}</option>)}
                </SelectInput>
              </Field>
              <Field label="仪器">
                <SelectInput value={testForm.equipmentId} onChange={(event) => setTestForm({ ...testForm, equipmentId: event.target.value })}>
                  {equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="指标"><TextInput value={testForm.metric} onChange={(event) => setTestForm({ ...testForm, metric: event.target.value })} /></Field>
              <Field label="结果值"><TextInput value={testForm.value} onChange={(event) => setTestForm({ ...testForm, value: event.target.value })} /></Field>
              <Field label="单位"><TextInput value={testForm.unit} onChange={(event) => setTestForm({ ...testForm, unit: event.target.value })} /></Field>
              <Field label="判定">
                <SelectInput value={testForm.result} onChange={(event) => setTestForm({ ...testForm, result: event.target.value })}>
                  {qualityResultOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeTestDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<TestTube2 size={16} />} disabled={busy !== "" || !samples.length || !equipment.length}>保存复核</Button>
              </FormActions>
            </DialogForm>
      </Dialog>
    </SectionGrid>
  );
}
