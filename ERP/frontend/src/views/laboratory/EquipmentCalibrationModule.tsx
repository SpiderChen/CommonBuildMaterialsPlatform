import { CheckCircle2, Wrench } from "lucide-react";
import { useState, type Dispatch, type SetStateAction } from "react";
import { ActionGroup, Button, DataTable, Dialog, DialogForm, Field, FormActions, HeroDateField, nameOf, Panel, SectionGrid, SelectInput, StatusChip, TextInput, buildDataTableRowContextMenu } from "../../components";
import type { DataDictionary, LaboratoryCalibration, LaboratoryEquipment, Site } from "../../services/types";
import type { CalibrationForm, EquipmentForm, SubmitHandler } from "./LaboratoryModuleTypes";
import { calibrationState, laboratoryDictionaryOptions, latestCalibrationForEquipment, shortDate } from "./laboratoryHelpers";

type Props = {
  dictionaries: DataDictionary[];
  equipment: LaboratoryEquipment[];
  calibrations: LaboratoryCalibration[];
  siteOptions: Site[];
  equipmentForm: EquipmentForm;
  setEquipmentForm: Dispatch<SetStateAction<EquipmentForm>>;
  calibrationForm: CalibrationForm;
  setCalibrationForm: Dispatch<SetStateAction<CalibrationForm>>;
  busy: string;
  onReload: () => Promise<void>;
  onSubmitEquipment: SubmitHandler;
  onSubmitCalibration: SubmitHandler;
};

export function EquipmentCalibrationModule({
  dictionaries,
  equipment,
  calibrations,
  siteOptions,
  equipmentForm,
  setEquipmentForm,
  calibrationForm,
  setCalibrationForm,
  busy,
  onReload,
  onSubmitEquipment,
  onSubmitCalibration
}: Props) {
  const activeList = [...equipment].sort((a, b) => b.id - a.id);
  const qualityResultOptions = laboratoryDictionaryOptions(dictionaries, "quality_result");
  const [equipmentDialogOpen, setEquipmentDialogOpen] = useState(false);
  const [calibrationDialogOpen, setCalibrationDialogOpen] = useState(false);

  function closeEquipmentDialog() {
    setEquipmentDialogOpen(false);
  }

  function closeCalibrationDialog() {
    setCalibrationDialogOpen(false);
  }

  return (
    <SectionGrid className="laboratory-module">
      <Panel className="span-12">
        <DataTable
          data={activeList}
          rowKey={(item) => item.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<LaboratoryEquipment>({
            actions: [
              {
                key: "calibrate-equipment",
                label: "记录该设备校准",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  setCalibrationForm({
                    ...calibrationForm,
                    equipmentId: String(item.id),
                    nextDueAt: item.nextCalibrationAt || calibrationForm.nextDueAt
                  });
                  setCalibrationDialogOpen(true);
                }
              },
              {
                key: "create-same-site-equipment",
                label: "登记同站点设备",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  setEquipmentForm({
                    ...equipmentForm,
                    siteId: String(item.siteId),
                    model: item.model || equipmentForm.model
                  });
                  setEquipmentDialogOpen(true);
                }
              }
            ],
            copyFields: [
              { key: "equipment", label: "设备编号", value: (item) => item.equipmentNo },
              { key: "name", label: "设备名称", value: (item) => item.name },
              { key: "serial", label: "序列号", value: (item) => item.serialNo },
              { key: "next", label: "下次校准", value: (item) => item.nextCalibrationAt }
            ]
          })}
          columns={[
            { key: "name", title: "设备名称", render: (item) => item.name },
            { key: "model", title: "型号", render: (item) => item.model || "-" },
            { key: "site", title: "站点", render: (item) => nameOf(siteOptions, item.siteId) },
            { key: "status", title: "校准状态", render: (item) => <StatusChip value={calibrationState(item).tone} /> },
            { key: "next", title: "下次校准", render: (item) => shortDate(item.nextCalibrationAt) },
            { key: "certificate", title: "最近证书", render: (item) => latestCalibrationForEquipment(item, calibrations)?.certificateNo || "-" }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button icon={<CheckCircle2 size={16} />} disabled={busy !== "" || !equipment.length} onClick={() => setCalibrationDialogOpen(true)}>记录校准</Button>
              <Button variant="primary" icon={<Wrench size={16} />} disabled={busy !== ""} onClick={() => setEquipmentDialogOpen(true)}>登记设备</Button>
            </ActionGroup>
          }
          emptyText="暂无设备"
        />
      </Panel>

      <Dialog open={equipmentDialogOpen} title="登记设备" className="master-dialog" closeDisabled={busy !== ""} onClose={closeEquipmentDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitEquipment(event);
                closeEquipmentDialog();
              }}
            >
              <Field label="站点">
                <SelectInput value={equipmentForm.siteId} onChange={(event) => setEquipmentForm({ ...equipmentForm, siteId: event.target.value })}>
                  {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="名称"><TextInput value={equipmentForm.name} onChange={(event) => setEquipmentForm({ ...equipmentForm, name: event.target.value })} /></Field>
              <Field label="型号"><TextInput value={equipmentForm.model} onChange={(event) => setEquipmentForm({ ...equipmentForm, model: event.target.value })} /></Field>
              <Field label="序列号"><TextInput value={equipmentForm.serialNo} onChange={(event) => setEquipmentForm({ ...equipmentForm, serialNo: event.target.value })} /></Field>
              <Field label="周期(天)"><TextInput value={equipmentForm.calibrationCycleDays} onChange={(event) => setEquipmentForm({ ...equipmentForm, calibrationCycleDays: event.target.value })} /></Field>
                <HeroDateField label="上次校准" value={equipmentForm.lastCalibrationAt} onChange={(lastCalibrationAt) => setEquipmentForm({ ...equipmentForm, lastCalibrationAt })} />
                <HeroDateField label="下次校准" value={equipmentForm.nextCalibrationAt} onChange={(nextCalibrationAt) => setEquipmentForm({ ...equipmentForm, nextCalibrationAt })} />
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeEquipmentDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<Wrench size={16} />} disabled={busy !== ""}>保存设备</Button>
              </FormActions>
            </DialogForm>
      </Dialog>

      <Dialog open={calibrationDialogOpen} title="记录校准" className="master-dialog" closeDisabled={busy !== ""} onClose={closeCalibrationDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitCalibration(event);
                closeCalibrationDialog();
              }}
            >
              <Field label="设备">
                <SelectInput value={calibrationForm.equipmentId} onChange={(event) => setCalibrationForm({ ...calibrationForm, equipmentId: event.target.value })}>
                  {equipment.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="结果">
                <SelectInput value={calibrationForm.result} onChange={(event) => setCalibrationForm({ ...calibrationForm, result: event.target.value })}>
                  {qualityResultOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
                <HeroDateField label="校准日期" value={calibrationForm.calibratedAt} onChange={(calibratedAt) => setCalibrationForm({ ...calibrationForm, calibratedAt })} />
              <Field label="证书号"><TextInput value={calibrationForm.certificateNo} onChange={(event) => setCalibrationForm({ ...calibrationForm, certificateNo: event.target.value })} /></Field>
              <Field label="机构"><TextInput value={calibrationForm.agency} onChange={(event) => setCalibrationForm({ ...calibrationForm, agency: event.target.value })} /></Field>
                <HeroDateField label="下次到期" value={calibrationForm.nextDueAt} onChange={(nextDueAt) => setCalibrationForm({ ...calibrationForm, nextDueAt })} />
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeCalibrationDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<CheckCircle2 size={16} />} disabled={busy !== "" || !equipment.length}>保存校准</Button>
              </FormActions>
            </DialogForm>
      </Dialog>
    </SectionGrid>
  );
}
