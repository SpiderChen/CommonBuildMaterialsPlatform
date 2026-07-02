import { CheckCircle2, Plus } from "lucide-react";
import { useState, type Dispatch, type SetStateAction } from "react";
import { ActionGroup, BusinessWorkflowStatus, BusinessWorkflowTimeline, Button, DataTable, Dialog, DialogForm, Field, FormActions, Panel, SectionGrid, SelectInput, StatusChip, TextAreaInput, TextInput, buildDataTableRowContextMenu } from "../../components";
import { api } from "../../services/api";
import { hasPermission } from "../../services/permissions";
import type { DataDictionary, QualityException, WorkflowOverview, WorkflowTask } from "../../services/types";
import type { ExceptionForm, MutateAction, SubmitHandler } from "./LaboratoryModuleTypes";
import { laboratoryDictionaryOptions } from "./laboratoryHelpers";

type Props = {
  dictionaries: DataDictionary[];
  exceptions: QualityException[];
  workflowOverview: WorkflowOverview | null;
  currentRoleCode: string;
  currentPermissions: string[];
  exceptionForm: ExceptionForm;
  setExceptionForm: Dispatch<SetStateAction<ExceptionForm>>;
  busy: string;
  mutate: MutateAction;
  onReload: () => Promise<void>;
  onSubmitException: SubmitHandler;
};

const qualityExceptionWorkflowResources = ["quality_exception", "quality_exceptions"];

export function ExceptionClosureModule({
  dictionaries,
  exceptions,
  workflowOverview,
  currentRoleCode,
  currentPermissions,
  exceptionForm,
  setExceptionForm,
  busy,
  mutate,
  onReload,
  onSubmitException
}: Props) {
  const sortedExceptions = [...exceptions].sort((a, b) => b.id - a.id);
  const severityOptions = laboratoryDictionaryOptions(dictionaries, "severity_level");
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [handleDialogOpen, setHandleDialogOpen] = useState(false);
  const [handlingId, setHandlingId] = useState<number | null>(null);
  const [workflowComment, setWorkflowComment] = useState("同意处理");

  const handlingItem = exceptions.find((item) => item.id === handlingId) || null;
  const handlingTitle = handlingItem?.title || "待处理异常";

  function closeCreateDialog() {
    setCreateDialogOpen(false);
  }

  function closeHandleDialog() {
    setHandleDialogOpen(false);
    setHandlingId(null);
  }

  function openHandleDialog(item: QualityException) {
    setExceptionForm((value) => ({
      ...value,
      rootCause: item.rootCause || value.rootCause,
      correctiveAction: item.correctiveAction || value.correctiveAction
    }));
    setHandlingId(item.id);
    setHandleDialogOpen(true);
  }

  function workflowStatusForException(item: QualityException) {
    return (
      <BusinessWorkflowStatus
        overview={workflowOverview}
        resourceNames={qualityExceptionWorkflowResources}
        resourceId={item.id}
        resourceNo={item.exceptionNo}
        baseStatus={<StatusChip value={item.status} />}
      />
    );
  }

  function canActWorkflowTask(task: WorkflowTask | null) {
    return Boolean(task && task.status === "pending" && (task.roleCode === currentRoleCode || hasPermission(currentPermissions, "*")));
  }

  async function actWorkflowTask(task: WorkflowTask, action: "approve" | "reject") {
    await mutate(`workflow-task-${action}-${task.id}`, () => api.actWorkflowTask(task.id, action, workflowComment));
  }

  function workflowTimelineForException(item: QualityException | null) {
    if (!item) return null;
    return (
      <BusinessWorkflowTimeline
        overview={workflowOverview}
        resourceNames={qualityExceptionWorkflowResources}
        resourceId={item.id}
        resourceNo={item.exceptionNo}
        emptyText="当前质量异常暂无工作流实例"
        comment={workflowComment}
        busy={busy !== ""}
        canActTask={canActWorkflowTask}
        onActTask={actWorkflowTask}
        onCommentChange={setWorkflowComment}
      />
    );
  }

  return (
    <SectionGrid className="laboratory-module">
      <Panel className="span-12">
        <DataTable
          data={sortedExceptions}
          rowKey={(item) => item.id}
          onRefresh={onReload}
          rowContextMenu={buildDataTableRowContextMenu<QualityException>({
            actions: [
              {
                key: "handle-exception",
                label: "处理该异常",
                disabled: (item) => busy !== "" || item.status === "closed",
                onSelect: (item) => openHandleDialog(item)
              },
              {
                key: "focus-owner",
                label: "只看该责任人",
                disabled: (item) => !item.responsible,
                onSelect: (item, helpers) => helpers.searchText(item.responsible)
              }
            ],
            copyFields: [
              { key: "exception", label: "异常编号", value: (item) => item.exceptionNo },
              { key: "title", label: "异常主题", value: (item) => item.title },
              { key: "owner", label: "责任人", value: (item) => item.responsible },
              { key: "description", label: "异常描述", value: (item) => item.description }
            ]
          })}
          columns={[
            { key: "exceptionNo", title: "编号", render: (item) => item.exceptionNo },
            { key: "title", title: "主题", render: (item) => item.title },
            { key: "severity", title: "等级", render: (item) => <StatusChip value={item.severity} /> },
            { key: "status", title: "状态", render: (item) => workflowStatusForException(item) },
            { key: "owner", title: "责任人", render: (item) => item.responsible || "-" },
            { key: "time", title: "创建时间", render: (item) => item.createdAt },
            { key: "description", title: "描述", render: (item) => item.description || "未填写描述" },
            {
              key: "action",
              title: "操作",
              render: (item) => item.status === "closed" ? <span className="muted">-</span> : (
                <Button icon={<CheckCircle2 size={15} />} disabled={busy !== ""} onClick={() => openHandleDialog(item)}>关闭</Button>
              )
            }
          ]}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<Plus size={16} />} onClick={() => setCreateDialogOpen(true)} disabled={busy !== ""}>创建异常</Button>
            </ActionGroup>
          }
          emptyText="暂无异常"
        />
      </Panel>

      <Dialog open={createDialogOpen} title="创建异常" className="master-dialog" closeDisabled={busy !== ""} onClose={closeCreateDialog}>
            <DialogForm
              onSubmit={async (event) => {
                await onSubmitException(event);
                closeCreateDialog();
              }}
            >
              <Field label="主题"><TextInput value={exceptionForm.title} onChange={(event) => setExceptionForm({ ...exceptionForm, title: event.target.value })} /></Field>
              <Field label="等级">
                <SelectInput value={exceptionForm.severity} onChange={(event) => setExceptionForm({ ...exceptionForm, severity: event.target.value })}>
                  {severityOptions.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
                </SelectInput>
              </Field>
              <Field label="负责人"><TextInput value={exceptionForm.responsible} onChange={(event) => setExceptionForm({ ...exceptionForm, responsible: event.target.value })} /></Field>
              <Field label="描述" spanAll><TextAreaInput value={exceptionForm.description} onChange={(event) => setExceptionForm({ ...exceptionForm, description: event.target.value })} /></Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeCreateDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<CheckCircle2 size={16} />} disabled={busy !== ""}>保存异常</Button>
              </FormActions>
            </DialogForm>
      </Dialog>

      <Dialog open={handleDialogOpen && handlingId !== null} title={`关闭异常 ${handlingTitle}`} className="master-dialog" closeDisabled={busy !== ""} onClose={closeHandleDialog}>
            <DialogForm
              onSubmit={async (event) => {
                event.preventDefault();
                if (handlingId === null) return;
                await mutate("handle-exception", () => api.handleQualityException(handlingId, {
                  rootCause: exceptionForm.rootCause,
                  correctiveAction: exceptionForm.correctiveAction
                }));
                closeHandleDialog();
              }}
            >
              <div className="form-span-all">
                {workflowTimelineForException(handlingItem)}
              </div>
              <Field label="根因" spanAll><TextAreaInput value={exceptionForm.rootCause} onChange={(event) => setExceptionForm({ ...exceptionForm, rootCause: event.target.value })} /></Field>
              <Field label="措施" spanAll><TextAreaInput value={exceptionForm.correctiveAction} onChange={(event) => setExceptionForm({ ...exceptionForm, correctiveAction: event.target.value })} /></Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={closeHandleDialog}>取消</Button>
                <Button variant="primary" type="submit" icon={<CheckCircle2 size={16} />} disabled={busy !== ""}>确认关闭</Button>
              </FormActions>
            </DialogForm>
      </Dialog>
    </SectionGrid>
  );
}
