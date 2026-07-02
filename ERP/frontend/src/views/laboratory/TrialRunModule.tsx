import { FlaskConical, Plus } from "lucide-react";
import { useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { ActionGroup, Button, DataTable, Dialog, DialogForm, Field, FormActions, Panel, SectionGrid, SelectInput, StatusChip, TextInput, buildDataTableRowContextMenu } from "../../components";
import type { DataDictionary, MixDesign, MixDesignTrialRun } from "../../services/types";
import type { SubmitHandler, TrialForm } from "./LaboratoryModuleTypes";
import { laboratoryDictionaryOptions, shortDate } from "./laboratoryHelpers";

type Props = {
  dictionaries: DataDictionary[];
  trialForm: TrialForm;
  setTrialForm: Dispatch<SetStateAction<TrialForm>>;
  draftMixes: MixDesign[];
  currentMixes: MixDesign[];
  trialRuns: MixDesignTrialRun[];
  busy: string;
  onReload: () => Promise<void>;
  onSubmitTrial: SubmitHandler;
};

export function TrialRunModule({
  dictionaries,
  trialForm,
  setTrialForm,
  draftMixes,
  currentMixes,
  trialRuns,
  busy,
  onReload,
  onSubmitTrial
}: Props) {
  const mixOptions = [...draftMixes, ...currentMixes];
  const qualityResultOptions = laboratoryDictionaryOptions(dictionaries, "quality_result");
  const sortedRuns = useMemo(() => [...trialRuns].sort((a, b) => b.id - a.id), [trialRuns]);
  const [dialogOpen, setDialogOpen] = useState(false);

  function closeDialog() {
    setDialogOpen(false);
  }

  return (
    <SectionGrid className="laboratory-module">
      <Panel className="span-12">
        <DataTable
          data={sortedRuns}
          rowKey={(item) => item.id}
          pageSize={10}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<MixDesignTrialRun>({
            actions: [
              {
                key: "create-same-mix-trial",
                label: "基于该配比新增试配",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  setTrialForm({
                    ...trialForm,
                    mixDesignId: String(item.mixDesignId),
                    result: item.result || trialForm.result
                  });
                  setDialogOpen(true);
                }
              },
              {
                key: "focus-trial",
                label: "只看该试配",
                onSelect: (item, helpers) => helpers.searchText(item.trialNo)
              }
            ],
            copyFields: [
              { key: "trial", label: "试配编号", value: (item) => item.trialNo },
              { key: "strength", label: "强度结果", value: (item) => `7d ${item.strength7d} MPa / 28d ${item.strength28d} MPa` },
              { key: "params", label: "试验参数", value: (item) => `沥青用量 ${item.water} / 矿料级配 ${item.sandRate}% / 添加剂 ${item.admixtureRate}%` }
            ]
          })}
          columns={[
            { key: "trialNo", title: "试配编号", render: (item) => item.trialNo },
            {
              key: "mix",
              title: "配比",
              render: (item) => {
                const mix = mixOptions.find((candidate) => candidate.id === item.mixDesignId);
                return mix ? `${mix.code} ${mix.version}` : item.targetStrength;
              }
            },
            { key: "strength7d", title: "7d", render: (item) => `${item.strength7d} MPa` },
            { key: "strength28d", title: "28d", render: (item) => `${item.strength28d} MPa` },
            { key: "result", title: "结论", render: (item) => <StatusChip value={item.result} /> },
            { key: "params", title: "试验参数", render: (item) => `沥青用量 ${item.water} · 矿料级配 ${item.sandRate}% · 添加剂 ${item.admixtureRate}%` },
            { key: "time", title: "记录时间", render: (item) => shortDate(item.testedAt || item.createdAt) }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<Plus size={16} />} onClick={() => setDialogOpen(true)} disabled={busy !== ""}>新增试配</Button>
            </ActionGroup>
          }
          emptyText="暂无试配记录"
        />
      </Panel>

      <Dialog open={dialogOpen} title="记录试配" className="master-dialog" closeDisabled={busy !== ""} onClose={closeDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitTrial(event);
                closeDialog();
              }}
            >
              <Field label="配比">
                <SelectInput value={trialForm.mixDesignId} onChange={(event) => setTrialForm({ ...trialForm, mixDesignId: event.target.value })}>
                  {mixOptions.map((item) => <option key={item.id} value={item.id}>{item.code} {item.version}</option>)}
                </SelectInput>
              </Field>
              <Field label="7d 强度"><TextInput value={trialForm.strength7d} onChange={(event) => setTrialForm({ ...trialForm, strength7d: event.target.value })} /></Field>
              <Field label="28d 强度"><TextInput value={trialForm.strength28d} onChange={(event) => setTrialForm({ ...trialForm, strength28d: event.target.value })} /></Field>
              <Field label="用水量"><TextInput value={trialForm.water} onChange={(event) => setTrialForm({ ...trialForm, water: event.target.value })} /></Field>
              <Field label="砂率"><TextInput value={trialForm.sandRate} onChange={(event) => setTrialForm({ ...trialForm, sandRate: event.target.value })} /></Field>
              <Field label="添加剂率"><TextInput value={trialForm.admixtureRate} onChange={(event) => setTrialForm({ ...trialForm, admixtureRate: event.target.value })} /></Field>
              <Field label="结论">
                <SelectInput value={trialForm.result} onChange={(event) => setTrialForm({ ...trialForm, result: event.target.value })}>
                  {qualityResultOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<FlaskConical size={16} />} disabled={busy !== "" || !mixOptions.length}>保存试配</Button>
              </FormActions>
            </DialogForm>
      </Dialog>
    </SectionGrid>
  );
}
