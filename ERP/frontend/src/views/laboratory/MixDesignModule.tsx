import { Ban, CheckCircle2, Plus, Trash2 } from "lucide-react";
import { useMemo, useState, type Dispatch, type FormEvent, type SetStateAction } from "react";
import { ActionGroup, Button, DataTable, Dialog, DialogForm, Field, FormActions, nameOf, Panel, SectionGrid, SelectInput, StatusChip, TextInput, buildDataTableRowContextMenu } from "../../components";
import { api } from "../../services/api";
import type { LaboratoryOverview, Material, MixDesign, MixDesignPlantProfile, Plant, PlantBufferLocation, Product, Site } from "../../services/types";
import type { MixForm, MutateAction, SubmitHandler } from "./LaboratoryModuleTypes";
import { materialSummary, parseMixMaterials, parseNumber, productName, today, trialsForMix } from "./laboratoryHelpers";

type ProfileMaterialForm = {
  materialId: string;
  dosage: string;
  adjustment: string;
  unit: string;
  bufferId: string;
  remark: string;
};

type ProfileForm = {
  plantId: string;
  code: string;
  version: string;
  scope: string;
  effectiveFrom: string;
  effectiveTo: string;
  remark: string;
  materials: ProfileMaterialForm[];
};

type BaseDialogMode = "create" | "revise";

type ApprovalForm = {
  trialRunId: string;
  effectiveFrom: string;
  effectiveTo: string;
};

type Props = {
  mode: "base" | "plant";
  overview: LaboratoryOverview;
  productOptions: Product[];
  siteOptions: Site[];
  materialOptions: Material[];
  mixForm: MixForm;
  setMixForm: Dispatch<SetStateAction<MixForm>>;
  busy: string;
  mutate: MutateAction;
  onReload: () => Promise<void>;
  onSubmitMix: SubmitHandler;
};

export function MixDesignModule({
  mode,
  overview,
  productOptions,
  siteOptions,
  materialOptions,
  mixForm,
  setMixForm,
  busy,
  mutate,
  onReload,
  onSubmitMix
}: Props) {
  const sortedMixes = useMemo(() => [...overview.mixDesigns].sort((a, b) => b.id - a.id), [overview.mixDesigns]);
  const approvedMixes = useMemo(() => sortedMixes.filter((item) => item.status === "approved"), [sortedMixes]);
  const sortedProfiles = useMemo(() => [...overview.mixDesignPlantProfiles].sort((a, b) => b.id - a.id), [overview.mixDesignPlantProfiles]);
  const plantOptions = useMemo(() => overview.plants.filter((item) => item.status === "running" || item.status === "active" || item.status === ""), [overview.plants]);
  const plantBufferOptions = useMemo(() => overview.plantBufferLocations.filter((item) => item.status === "active" || item.status === "running" || item.status === ""), [overview.plantBufferLocations]);
  const activeMaterialOptions = useMemo(() => materialOptions.filter((item) => item.status === "active" || item.status === ""), [materialOptions]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [baseDialogMode, setBaseDialogMode] = useState<BaseDialogMode>("create");
  const [editingMix, setEditingMix] = useState<MixDesign | null>(null);
  const [approvalMix, setApprovalMix] = useState<MixDesign | null>(null);
  const [approvalDialogOpen, setApprovalDialogOpen] = useState(false);
  const [approvalForm, setApprovalForm] = useState<ApprovalForm>({
    trialRunId: "",
    effectiveFrom: today,
    effectiveTo: ""
  });
  const [retireMix, setRetireMix] = useState<MixDesign | null>(null);
  const [retireDialogOpen, setRetireDialogOpen] = useState(false);
  const [profileDialogOpen, setProfileDialogOpen] = useState(false);
  const [profileBaseMixId, setProfileBaseMixId] = useState<number | null>(null);
  const [profileForm, setProfileForm] = useState<ProfileForm>({
    plantId: "",
    code: "",
    version: "",
    scope: "",
    effectiveFrom: today,
    effectiveTo: "",
    remark: "",
    materials: []
  });
  const selectedMaterialIds = new Set(mixForm.materials.map((item) => item.materialId).filter(Boolean));
  const mixMaterialsValid = mixForm.materials.length > 0
    && mixForm.materials.every((item) => Number(item.materialId) > 0 && Number(item.dosage) > 0)
    && selectedMaterialIds.size === mixForm.materials.filter((item) => item.materialId).length;
  const selectedBaseMix = profileBaseMixId ? sortedMixes.find((item) => item.id === profileBaseMixId) : undefined;
  const profilePlantOptions = selectedBaseMix ? plantOptions.filter((item) => selectedBaseMix.siteId === 0 || item.siteId === selectedBaseMix.siteId) : plantOptions;
  const selectedProfileMaterialIds = new Set(profileForm.materials.map((item) => item.materialId).filter(Boolean));
  const profileMaterialsValid = Boolean(selectedBaseMix && profileForm.plantId)
    && profileForm.materials.length > 0
    && profileForm.materials.every((item) => Number(item.materialId) > 0 && (Number(item.dosage) > 0 || Number(item.adjustment) !== 0 || Number(item.bufferId) > 0))
    && selectedProfileMaterialIds.size === profileForm.materials.filter((item) => item.materialId).length;
  const approvalTrials = useMemo(
    () => approvalMix ? trialsForMix(approvalMix, overview.trialRuns).filter((item) => item.result === "passed") : [],
    [approvalMix, overview.trialRuns]
  );

  function closeDialog() {
    setDialogOpen(false);
    setEditingMix(null);
    setBaseDialogMode("create");
  }

  function openCreateDialog() {
    setBaseDialogMode("create");
    setEditingMix(null);
    setDialogOpen(true);
  }

  function openReviseDialog(mix: MixDesign) {
    setBaseDialogMode("revise");
    setEditingMix(mix);
    setMixForm({
      productId: String(mix.productId),
      siteId: String(mix.siteId),
      code: mix.code,
      version: `${mix.version || "v1"}-rev`,
      strengthGrade: mix.strengthGrade,
      slump: mix.slump,
      scope: mix.scope,
      materials: mix.materials.map((item) => ({
        materialId: String(item.materialId),
        dosage: String(item.dosage || ""),
        unit: item.unit || "kg/t"
      }))
    });
    setDialogOpen(true);
  }

  function openApprovalDialog(mix: MixDesign) {
    const passedTrials = trialsForMix(mix, overview.trialRuns).filter((item) => item.result === "passed");
    setApprovalMix(mix);
    setApprovalForm({
      trialRunId: passedTrials[0] ? String(passedTrials[0].id) : "",
      effectiveFrom: mix.effectiveFrom || today,
      effectiveTo: mix.effectiveTo || ""
    });
    setApprovalDialogOpen(true);
  }

  function closeApprovalDialog() {
    setApprovalDialogOpen(false);
    setApprovalMix(null);
  }

  function openRetireDialog(mix: MixDesign) {
    setRetireMix(mix);
    setRetireDialogOpen(true);
  }

  function closeRetireDialog() {
    setRetireDialogOpen(false);
    setRetireMix(null);
  }

  function closeProfileDialog() {
    setProfileDialogOpen(false);
  }

  function profileMaterialForm(baseMaterial: MixDesign["materials"][number]): ProfileMaterialForm {
    return {
      materialId: String(baseMaterial.materialId),
      dosage: "",
      adjustment: "",
      unit: baseMaterial.unit || "kg/t",
      bufferId: "",
      remark: ""
    };
  }

  function buildProfileForm(mix: MixDesign | undefined): ProfileForm {
    const plant = mix
      ? plantOptions.find((item) => mix.siteId === 0 || item.siteId === mix.siteId) || plantOptions[0]
      : plantOptions[0];
    return {
      plantId: plant ? String(plant.id) : "",
      code: mix && plant ? `${mix.code}-${plant.code}` : "",
      version: mix ? `${mix.version}-line` : "line-v1",
      scope: plant ? `${plant.name}生产线配比` : "生产线配比",
      effectiveFrom: today,
      effectiveTo: "",
      remark: "",
      materials: (mix?.materials || []).slice(0, 2).map(profileMaterialForm)
    };
  }

  function openProfileDialog(mix: MixDesign | undefined = approvedMixes[0]) {
    if (!mix) return;
    setProfileBaseMixId(mix.id);
    setProfileForm(buildProfileForm(mix));
    setProfileDialogOpen(true);
  }

  function updateMixMaterial(index: number, patch: Partial<MixForm["materials"][number]>) {
    setMixForm((value) => ({
      ...value,
      materials: value.materials.map((item, itemIndex) => itemIndex === index ? { ...item, ...patch } : item)
    }));
  }

  function addMixMaterial() {
    const used = new Set(mixForm.materials.map((item) => item.materialId));
    const next = activeMaterialOptions.find((item) => !used.has(String(item.id))) || activeMaterialOptions[0];
    setMixForm((value) => ({
      ...value,
      materials: [...value.materials, { materialId: next ? String(next.id) : "", dosage: "", unit: "kg/t" }]
    }));
  }

  function removeMixMaterial(index: number) {
    setMixForm((value) => ({
      ...value,
      materials: value.materials.filter((_, itemIndex) => itemIndex !== index)
    }));
  }

  function plantLabel(id: number, fallbackCode = "") {
    const plant = plantOptions.find((item) => item.id === id);
    if (!plant) return fallbackCode || "-";
    return `${plant.name} · ${plant.code}`;
  }

  function bufferCanCarryMaterial(buffer: PlantBufferLocation, materialIdValue: string) {
    const materialId = Number(materialIdValue);
    if (!materialId) return true;
    return buffer.materialId === materialId || Boolean(buffer.allowedMaterialIds?.includes(materialId));
  }

  function bufferOptionsFor(materialIdValue: string) {
    const plantId = Number(profileForm.plantId);
    return plantBufferOptions.filter((item) => item.plantId === plantId && bufferCanCarryMaterial(item, materialIdValue));
  }

  function updateProfileMaterial(index: number, patch: Partial<ProfileMaterialForm>) {
    setProfileForm((value) => ({
      ...value,
      materials: value.materials.map((item, itemIndex) => itemIndex === index ? { ...item, ...patch } : item)
    }));
  }

  function addProfileMaterial() {
    if (!selectedBaseMix) return;
    const used = new Set(profileForm.materials.map((item) => item.materialId));
    const next = selectedBaseMix.materials.find((item) => !used.has(String(item.materialId))) || selectedBaseMix.materials[0];
    if (!next) return;
    setProfileForm((value) => ({
      ...value,
      materials: [...value.materials, profileMaterialForm(next)]
    }));
  }

  function removeProfileMaterial(index: number) {
    setProfileForm((value) => ({
      ...value,
      materials: value.materials.filter((_, itemIndex) => itemIndex !== index)
    }));
  }

  function baseMaterial(mix: MixDesign | undefined, materialId: number) {
    return mix?.materials.find((item) => item.materialId === materialId);
  }

  function profileMix(profile: MixDesignPlantProfile) {
    return sortedMixes.find((item) => item.id === profile.mixDesignId);
  }

  function profileMaterialSummary(profile: MixDesignPlantProfile) {
    const mix = profileMix(profile);
    if (!profile.materials.length) return "-";
    return profile.materials.map((item) => {
      const base = baseMaterial(mix, item.materialId);
      const unit = item.unit || base?.unit || "kg/t";
      const dosage = item.dosage > 0 ? `${item.dosage}${unit}` : item.adjustment !== 0 ? `${base?.dosage ?? 0}${unit} ${item.adjustment > 0 ? "+" : ""}${item.adjustment}` : "仅换仓";
      const buffer = item.bufferCode ? ` -> ${item.bufferCode}` : "";
      return `${nameOf(materialOptions, item.materialId)} ${dosage}${buffer}`;
    }).join(" / ");
  }

  function selectedBaseMaterialOptions() {
    return selectedBaseMix?.materials || [];
  }

  async function submitBaseDialog(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (baseDialogMode === "create") {
      await onSubmitMix(event);
      closeDialog();
      return;
    }
    if (!editingMix) return;
    await mutate("revise", () => api.reviseLaboratoryMixDesign(editingMix.id, {
      productId: parseNumber(mixForm.productId),
      siteId: parseNumber(mixForm.siteId),
      code: mixForm.code,
      version: mixForm.version,
      strengthGrade: mixForm.strengthGrade,
      slump: mixForm.slump,
      scope: mixForm.scope,
      materials: parseMixMaterials(mixForm.materials)
    }));
    closeDialog();
  }

  async function submitApproval(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!approvalMix) return;
    await mutate("approve", () => api.approveLaboratoryMixDesign(approvalMix.id, {
      trialRunId: parseNumber(approvalForm.trialRunId),
      effectiveFrom: approvalForm.effectiveFrom || today,
      effectiveTo: approvalForm.effectiveTo
    }));
    closeApprovalDialog();
  }

  async function submitRetire(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!retireMix) return;
    await mutate("retire", () => api.retireLaboratoryMixDesign(retireMix.id));
    closeRetireDialog();
  }

  async function submitProfile() {
    if (!selectedBaseMix) return;
    const payload = {
      plantId: parseNumber(profileForm.plantId),
      code: profileForm.code,
      version: profileForm.version,
      scope: profileForm.scope,
      status: "draft",
      effectiveFrom: profileForm.effectiveFrom || today,
      effectiveTo: profileForm.effectiveTo,
      remark: profileForm.remark,
      materials: profileForm.materials
        .map((item) => {
          const base = baseMaterial(selectedBaseMix, parseNumber(item.materialId));
          return {
            materialId: parseNumber(item.materialId),
            dosage: parseNumber(item.dosage),
            adjustment: parseNumber(item.adjustment),
            unit: item.unit || base?.unit || "kg/t",
            bufferId: parseNumber(item.bufferId),
            bufferCode: "",
            remark: item.remark
          };
        })
        .filter((item) => item.materialId > 0 && (item.dosage > 0 || item.adjustment !== 0 || item.bufferId > 0))
    };
    await mutate("createProfile", () => api.createMixDesignPlantProfile(selectedBaseMix.id, payload));
    closeProfileDialog();
    await onReload();
  }

  return (
    <SectionGrid className="laboratory-module">
      {mode === "base" ? (
      <Panel className="span-12">
        <DataTable
          title="基础配比"
          data={sortedMixes}
          rowKey={(item) => item.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<MixDesign>({
            actions: [
              {
                key: "revise-mix",
                label: "修订该基础配比",
                disabled: (mix) => busy !== "" || !mix.materials.length,
                onSelect: (mix) => openReviseDialog(mix)
              },
              {
                key: "approve-mix",
                label: "打开审批该配比",
                disabled: (mix) => busy !== "" || mix.status === "approved" || mix.status === "retired",
                onSelect: (mix) => openApprovalDialog(mix)
              },
              {
                key: "create-profile",
                label: "生成生产线配比",
                disabled: (mix) => busy !== "" || mix.status !== "approved" || !plantOptions.length,
                onSelect: (mix) => openProfileDialog(mix)
              }
            ],
            copyFields: [
              { key: "mix", label: "配比版本", value: (mix) => `${mix.code} ${mix.version}` },
              { key: "product", label: "产品", value: (mix) => productName(productOptions, mix.productId) },
              { key: "scope", label: "适用范围", value: (mix) => mix.scope },
              { key: "materials", label: "材料摘要", value: (mix) => materialSummary(mix.materials, materialOptions) }
            ]
          })}
          columns={[
            { key: "version", title: "版本", width: "126px", render: (mix) => <strong>{mix.code} {mix.version}</strong> },
            { key: "product", title: "产品", width: "150px", render: (mix) => productName(productOptions, mix.productId) },
            { key: "site", title: "站点", width: "52px", render: (mix) => nameOf(siteOptions, mix.siteId) },
            {
              key: "status",
              title: "当前状态",
              width: "88px",
              render: (mix) => <StatusChip value={mix.isCurrent ? "active" : mix.status} />
            },
            {
              key: "strength",
              title: "强度",
              width: "72px",
              render: (mix) => `${mix.strengthGrade || "-"} / ${mix.slump || "-"}`
            },
            { key: "scope", title: "适用范围", width: "96px", render: (mix) => mix.scope || "通用基础配比" },
            { key: "materials", title: "材料", width: "42%", render: (mix) => <MaterialSummaryList items={mix.materials} materials={materialOptions} /> },
            {
              key: "trials",
              title: "试配",
              width: "72px",
              render: (mix) => {
                const linkedTrials = trialsForMix(mix, overview.trialRuns);
                const latestTrial = linkedTrials[0];
                return latestTrial ? `${linkedTrials.length}次 · ${latestTrial.strength28d}MPa` : "待试配";
              }
            },
            {
              key: "actions",
              title: "操作",
              width: "154px",
              render: (mix) => (
                <ActionGroup as="span">
                  <Button icon={<CheckCircle2 size={15} />} disabled={busy !== "" || mix.status === "approved" || mix.status === "retired"} onClick={() => openApprovalDialog(mix)}>审批</Button>
                  <Button icon={<Plus size={15} />} disabled={busy !== "" || !mix.materials.length} onClick={() => openReviseDialog(mix)}>修订</Button>
                  <Button icon={<Ban size={15} />} disabled={busy !== "" || mix.status === "retired"} onClick={() => openRetireDialog(mix)}>停用</Button>
                </ActionGroup>
              )
            }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<Plus size={16} />} onClick={openCreateDialog}>新增基础配比</Button>
            </ActionGroup>
          }
          emptyText="暂无基础配比"
        />
      </Panel>
      ) : null}

      {mode === "plant" ? (
      <Panel className="span-12">
        <DataTable
          title="生产线配比"
          data={sortedProfiles}
          rowKey={(item) => item.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<MixDesignPlantProfile>({
            actions: [
              {
                key: "focus-plant",
                label: "只看该生产线",
                onSelect: (profile, helpers) => helpers.searchText(plantLabel(profile.plantId, profile.plantCode))
              },
              {
                key: "focus-base-mix",
                label: "只看该基础配比",
                onSelect: (profile, helpers) => helpers.searchText(productionMixName(profileMix(profile)))
              }
            ],
            copyFields: [
              { key: "profile", label: "生产线配比", value: (profile) => `${profile.code} ${profile.version}` },
              { key: "base", label: "基础配比", value: (profile) => productionMixName(profileMix(profile)) },
              { key: "plant", label: "生产线", value: (profile) => plantLabel(profile.plantId, profile.plantCode) },
              { key: "materials", label: "微调摘要", value: (profile) => profileMaterialSummary(profile) }
            ]
          })}
          columns={[
            { key: "version", title: "版本", render: (profile) => <strong>{profile.code} {profile.version}</strong> },
            { key: "base", title: "基础配比", render: (profile) => productionMixName(profileMix(profile)) },
            { key: "plant", title: "生产线", render: (profile) => plantLabel(profile.plantId, profile.plantCode) },
            { key: "status", title: "状态", render: (profile) => <StatusChip value={profile.isCurrent ? "active" : profile.status} /> },
            { key: "scope", title: "范围", render: (profile) => profile.scope || "生产线配比" },
            { key: "materials", title: "微调", render: (profile) => profileMaterialSummary(profile) },
            { key: "effective", title: "生效", render: (profile) => `${profile.effectiveFrom || "-"}${profile.effectiveTo ? ` ~ ${profile.effectiveTo}` : ""}` },
            {
              key: "actions",
              title: "操作",
              render: (profile) => (
                <ActionGroup as="span">
                  <Button icon={<CheckCircle2 size={15} />} disabled={busy !== "" || profile.status === "approved"} onClick={() => void mutate("approveProfile", () => api.approveMixDesignPlantProfile(profile.id, { effectiveFrom: profile.effectiveFrom || today, effectiveTo: profile.effectiveTo }))}>审批</Button>
                  <Button icon={<Ban size={15} />} disabled={busy !== "" || profile.status === "retired"} onClick={() => void mutate("retireProfile", () => api.retireMixDesignPlantProfile(profile.id))}>停用</Button>
                </ActionGroup>
              )
            }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<Plus size={16} />} disabled={!approvedMixes.length || !plantOptions.length} onClick={() => openProfileDialog()}>新增生产线配比</Button>
            </ActionGroup>
          }
          emptyText="暂无生产线配比"
        />
      </Panel>
      ) : null}

      {mode === "base" ? (
      <Dialog open={dialogOpen} title={baseDialogMode === "revise" ? "修订基础配比" : "新增基础配比"} className="master-dialog" closeDisabled={busy !== ""} onClose={closeDialog}>
            <DialogForm
              onSubmit={submitBaseDialog}
            >
              <Field label="产品">
                <SelectInput value={mixForm.productId} onChange={(event) => setMixForm({ ...mixForm, productId: event.target.value })}>
                  {productOptions.map((item) => <option key={item.id} value={item.id}>{item.name} {item.spec}</option>)}
                </SelectInput>
              </Field>
              <Field label="站点">
                <SelectInput value={mixForm.siteId} onChange={(event) => setMixForm({ ...mixForm, siteId: event.target.value })}>
                  {siteOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </SelectInput>
              </Field>
              <Field label="编号"><TextInput value={mixForm.code} onChange={(event) => setMixForm({ ...mixForm, code: event.target.value })} /></Field>
              <Field label="版本"><TextInput value={mixForm.version} onChange={(event) => setMixForm({ ...mixForm, version: event.target.value })} /></Field>
              <Field label="混合料规格"><TextInput value={mixForm.strengthGrade} onChange={(event) => setMixForm({ ...mixForm, strengthGrade: event.target.value })} /></Field>
              <Field label="油石比"><TextInput value={mixForm.slump} onChange={(event) => setMixForm({ ...mixForm, slump: event.target.value })} /></Field>
              <Field label="适用范围" spanAll><TextInput value={mixForm.scope} onChange={(event) => setMixForm({ ...mixForm, scope: event.target.value })} /></Field>
              <Field label="材料用量" spanAll>
                <div className="mix-material-editor">
                  {mixForm.materials.map((material, index) => {
                    const usedByOtherRows = new Set(mixForm.materials.filter((_, itemIndex) => itemIndex !== index).map((item) => item.materialId));
                    return (
                      <div className="mix-material-row" key={`${index}-${material.materialId || "empty"}`}>
                        <SelectInput
                          value={material.materialId}
                          onChange={(event) => {
                            updateMixMaterial(index, { materialId: event.target.value, unit: material.unit || "kg/t" });
                          }}
                        >
                          <option value="">选择物料</option>
                          {activeMaterialOptions.map((item) => (
                            <option key={item.id} value={item.id} disabled={usedByOtherRows.has(String(item.id))}>
                              {item.name} {item.spec || ""} · {item.unit}
                            </option>
                          ))}
                        </SelectInput>
                        <TextInput
                          type="number"
                          min="0"
                          step="0.01"
                          value={material.dosage}
                          onChange={(event) => updateMixMaterial(index, { dosage: event.target.value })}
                          placeholder="用量"
                        />
                        <TextInput
                          value={material.unit || "kg/t"}
                          onChange={(event) => updateMixMaterial(index, { unit: event.target.value })}
                          placeholder="单位"
                        />
                        <Button
                          icon={<Trash2 size={14} />}
                          disabled={busy !== "" || mixForm.materials.length <= 1}
                          onClick={() => removeMixMaterial(index)}
                        >
                          删除
                        </Button>
                      </div>
                    );
                  })}
                  <Button icon={<Plus size={14} />} disabled={busy !== "" || !activeMaterialOptions.length} onClick={addMixMaterial}>添加物料</Button>
                </div>
              </Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<Plus size={16} />} disabled={busy !== "" || !productOptions.length || !siteOptions.length || !activeMaterialOptions.length || !mixMaterialsValid}>
                  {baseDialogMode === "revise" ? "保存修订草稿" : "保存基础配比"}
                </Button>
              </FormActions>
            </DialogForm>
      </Dialog>
      ) : null}

      {mode === "base" ? (
      <Dialog open={approvalDialogOpen} title="审批基础配比" className="master-dialog" closeDisabled={busy !== ""} onClose={closeApprovalDialog}>
        <DialogForm onSubmit={submitApproval}>
          <Field label="基础配比">
            <TextInput value={approvalMix ? `${approvalMix.code} ${approvalMix.version}` : ""} disabled />
          </Field>
          <Field label="合格试配">
            <SelectInput value={approvalForm.trialRunId} onChange={(event) => setApprovalForm({ ...approvalForm, trialRunId: event.target.value })}>
              <option value="">不关联试配记录</option>
              {approvalTrials.map((trial) => (
                <option key={trial.id} value={trial.id}>
                  {trial.trialNo} · 28d {trial.strength28d}MPa
                </option>
              ))}
            </SelectInput>
          </Field>
          <Field label="生效日期">
            <TextInput type="date" value={approvalForm.effectiveFrom} onChange={(event) => setApprovalForm({ ...approvalForm, effectiveFrom: event.target.value })} />
          </Field>
          <Field label="截止日期">
            <TextInput type="date" value={approvalForm.effectiveTo} onChange={(event) => setApprovalForm({ ...approvalForm, effectiveTo: event.target.value })} />
          </Field>
          <FormActions>
            <Button disabled={busy !== ""} onClick={closeApprovalDialog}>取消</Button>
            <Button variant="primary" type="submit" icon={<CheckCircle2 size={16} />} disabled={busy !== "" || !approvalMix}>审批通过</Button>
          </FormActions>
        </DialogForm>
      </Dialog>
      ) : null}

      {mode === "base" ? (
      <Dialog open={retireDialogOpen} title="停用基础配比" className="master-dialog" closeDisabled={busy !== ""} onClose={closeRetireDialog}>
        <DialogForm onSubmit={submitRetire}>
          <Field label="基础配比">
            <TextInput value={retireMix ? `${retireMix.code} ${retireMix.version}` : ""} disabled />
          </Field>
          <Field label="当前状态">
            <TextInput value={retireMix?.isCurrent ? "当前生效" : retireMix?.status || ""} disabled />
          </Field>
          <FormActions>
            <Button disabled={busy !== ""} onClick={closeRetireDialog}>取消</Button>
            <Button variant="danger" type="submit" icon={<Ban size={16} />} disabled={busy !== "" || !retireMix}>确认停用</Button>
          </FormActions>
        </DialogForm>
      </Dialog>
      ) : null}

      {mode === "plant" ? (
      <Dialog open={profileDialogOpen} title="生产线配比" size="wide" className="master-dialog" closeDisabled={busy !== ""} onClose={closeProfileDialog}>
        <DialogForm
          onSubmit={async (event) => {
            event.preventDefault();
            await submitProfile();
          }}
        >
          <Field label="基础配比">
            <SelectInput
              value={profileBaseMixId || ""}
              onChange={(event) => {
                const mix = approvedMixes.find((item) => item.id === Number(event.target.value));
                if (!mix) return;
                setProfileBaseMixId(mix.id);
                setProfileForm(buildProfileForm(mix));
              }}
            >
              {approvedMixes.map((mix) => <option key={mix.id} value={mix.id}>{productionMixName(mix)} · {productName(productOptions, mix.productId)}</option>)}
            </SelectInput>
          </Field>
          <Field label="生产线">
            <SelectInput
              value={profileForm.plantId}
              onChange={(event) => setProfileForm({
                ...profileForm,
                plantId: event.target.value,
                materials: profileForm.materials.map((item) => ({ ...item, bufferId: "" }))
              })}
            >
              <option value="">选择生产线</option>
              {profilePlantOptions.map((plant: Plant) => <option key={plant.id} value={plant.id}>{plant.name} · {plant.code}</option>)}
            </SelectInput>
          </Field>
          <Field label="编号"><TextInput value={profileForm.code} onChange={(event) => setProfileForm({ ...profileForm, code: event.target.value })} /></Field>
          <Field label="版本"><TextInput value={profileForm.version} onChange={(event) => setProfileForm({ ...profileForm, version: event.target.value })} /></Field>
          <Field label="生效日期"><TextInput type="date" value={profileForm.effectiveFrom} onChange={(event) => setProfileForm({ ...profileForm, effectiveFrom: event.target.value })} /></Field>
          <Field label="截止日期"><TextInput type="date" value={profileForm.effectiveTo} onChange={(event) => setProfileForm({ ...profileForm, effectiveTo: event.target.value })} /></Field>
          <Field label="适用范围" spanAll><TextInput value={profileForm.scope} onChange={(event) => setProfileForm({ ...profileForm, scope: event.target.value })} /></Field>
          <Field label="微调物料" spanAll>
            <div className="mix-material-editor profile-material-editor">
              {profileForm.materials.map((material, index) => {
                const usedByOtherRows = new Set(profileForm.materials.filter((_, itemIndex) => itemIndex !== index).map((item) => item.materialId));
                const buffers = bufferOptionsFor(material.materialId);
                return (
                  <div className="mix-material-row profile-material-row" key={`${index}-${material.materialId || "empty"}`}>
                    <SelectInput
                      value={material.materialId}
                      onChange={(event) => updateProfileMaterial(index, { materialId: event.target.value, bufferId: "" })}
                    >
                      <option value="">选择物料</option>
                      {selectedBaseMaterialOptions().map((item) => (
                        <option key={item.materialId} value={item.materialId} disabled={usedByOtherRows.has(String(item.materialId))}>
                          {nameOf(materialOptions, item.materialId)} · 基础 {item.dosage}{item.unit}
                        </option>
                      ))}
                    </SelectInput>
                    <TextInput
                      type="number"
                      min="0"
                      step="0.01"
                      value={material.dosage}
                      onChange={(event) => updateProfileMaterial(index, { dosage: event.target.value })}
                      placeholder="覆盖用量"
                    />
                    <TextInput
                      type="number"
                      step="0.01"
                      value={material.adjustment}
                      onChange={(event) => updateProfileMaterial(index, { adjustment: event.target.value })}
                      placeholder="增减量"
                    />
                    <SelectInput value={material.bufferId} onChange={(event) => updateProfileMaterial(index, { bufferId: event.target.value })}>
                      <option value="">不指定筒仓</option>
                      {buffers.map((buffer) => <option key={buffer.id} value={buffer.id}>{buffer.name} · {buffer.code}</option>)}
                    </SelectInput>
                    <Button
                      icon={<Trash2 size={14} />}
                      disabled={busy !== "" || profileForm.materials.length <= 1}
                      onClick={() => removeProfileMaterial(index)}
                    >
                      删除
                    </Button>
                  </div>
                );
              })}
              <Button icon={<Plus size={14} />} disabled={busy !== "" || !selectedBaseMix || profileForm.materials.length >= (selectedBaseMix?.materials.length || 0)} onClick={addProfileMaterial}>添加微调项</Button>
            </div>
          </Field>
          <Field label="备注" spanAll><TextInput value={profileForm.remark} onChange={(event) => setProfileForm({ ...profileForm, remark: event.target.value })} /></Field>
          <FormActions>
            <Button disabled={busy !== ""} onClick={closeProfileDialog}>取消</Button>
            <Button variant="primary" type="submit" icon={<Plus size={16} />} disabled={busy !== "" || !profileMaterialsValid}>保存为待审批</Button>
          </FormActions>
        </DialogForm>
      </Dialog>
      ) : null}
    </SectionGrid>
  );
}

function productionMixName(mix: MixDesign | undefined) {
  return mix ? `${mix.code} ${mix.version}` : "-";
}

function MaterialSummaryList({ items, materials }: { items: MixDesign["materials"]; materials: Material[] }) {
  if (!items.length) {
    return <span className="muted">-</span>;
  }

  return (
    <div className="mix-material-summary" title={materialSummary(items, materials)}>
      {items.map((item, index) => {
        const material = materials.find((entry) => entry.id === item.materialId);
        const name = material?.name || nameOf(materials, item.materialId);
        const spec = material?.spec || "";

        return (
          <span className="mix-material-token" key={`${item.materialId}-${index}`}>
            <span className="mix-material-token-main">
              <strong>{name}</strong>
              {spec ? <span>{spec}</span> : null}
            </span>
            <span className="mix-material-dosage">{item.dosage}{item.unit}</span>
          </span>
        );
      })}
    </div>
  );
}
