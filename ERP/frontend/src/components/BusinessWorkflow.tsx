import { CheckCircle2 } from "lucide-react";
import type { ReactNode } from "react";
import { ActionGroup, Button, Field, StatusChip, TextInput, statusLabel } from "./ui";
import type { WorkflowDelivery, WorkflowEvent, WorkflowInstance, WorkflowLog, WorkflowOutbox, WorkflowOverview, WorkflowTask } from "../services/types";

export type BusinessWorkflowItems = {
  instances: WorkflowInstance[];
  tasks: WorkflowTask[];
  logs: WorkflowLog[];
  events: WorkflowEvent[];
  outbox: WorkflowOutbox[];
  deliveries: WorkflowDelivery[];
};

export function normalizeBusinessWorkflowKey(value?: string | number | null) {
  return String(value || "").trim().toLowerCase().replace(/[\s_-]+/g, "");
}

export function findBusinessWorkflowItems(
  overview: WorkflowOverview | null | undefined,
  resourceNames: string | string[],
  resourceId?: number | string | null,
  resourceNo?: string | null
): BusinessWorkflowItems {
  const names = Array.isArray(resourceNames) ? resourceNames : [resourceNames];
  const resources = new Set(names.map(normalizeBusinessWorkflowKey));
  const targetId = Number(resourceId || 0);
  const targetNo = normalizeBusinessWorkflowKey(resourceNo);
  const resourceMatches = (resource?: string) => resources.has(normalizeBusinessWorkflowKey(resource));

  const instances = (overview?.instances || [])
    .filter((item) => {
      const resourceMatched = resourceMatches(item.resource);
      const idMatched = targetId > 0 && Number(item.resourceId || 0) === targetId;
      const noMatched = targetNo !== "" && normalizeBusinessWorkflowKey(item.resourceNo) === targetNo;
      return resourceMatched && (idMatched || noMatched);
    })
    .sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));

  const instanceIds = new Set(instances.map((item) => item.id));
  const events = (overview?.events || [])
    .filter((item) => {
      const resourceMatched = resourceMatches(item.resource);
      const idMatched = targetId > 0 && Number(item.resourceId || 0) === targetId;
      const noMatched = targetNo !== "" && normalizeBusinessWorkflowKey(item.resourceNo) === targetNo;
      return resourceMatched && (idMatched || noMatched);
    })
    .sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));

  const eventIds = new Set([
    ...events.map((item) => item.id),
    ...instances.map((item) => item.triggerEventId || 0).filter(Boolean)
  ]);

  const tasks = (overview?.tasks || [])
    .filter((item) => instanceIds.has(item.instanceId))
    .sort((left, right) => left.step - right.step || left.id - right.id);

  const logs = (overview?.logs || [])
    .filter((item) => item.instanceId && instanceIds.has(item.instanceId))
    .sort((left, right) => (left.createdAt || "").localeCompare(right.createdAt || ""));

  for (const log of logs) {
    if (log.triggerEventId) eventIds.add(log.triggerEventId);
  }

  const outbox = (overview?.outbox || [])
    .filter((item) => {
      if (item.instanceId && instanceIds.has(item.instanceId)) return true;
      if (item.triggerEventId && eventIds.has(item.triggerEventId)) return true;
      const resourceMatched = resourceMatches(item.resource);
      const idMatched = targetId > 0 && Number(item.resourceId || 0) === targetId;
      return resourceMatched && idMatched;
    })
    .sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));

  const outboxIds = new Set(outbox.map((item) => item.id));
  const outboxNos = new Set(outbox.map((item) => item.outboxNo));
  const deliveries = (overview?.deliveries || [])
    .filter((item) => outboxIds.has(item.outboxId) || outboxNos.has(item.outboxNo))
    .sort((left, right) => (right.createdAt || "").localeCompare(left.createdAt || ""));

  return { instances, tasks, logs, events, outbox, deliveries };
}

export function currentBusinessWorkflowTask(instance: WorkflowInstance | null, tasks: WorkflowTask[]) {
  if (!instance) return null;
  return tasks.find((task) => task.id === instance.currentTaskId) || tasks.find((task) => task.instanceId === instance.id && task.status === "pending") || null;
}

export function businessWorkflowDeliverySummary(items: WorkflowDelivery[]) {
  if (!items.length) return "暂无投递";
  const succeeded = items.filter((item) => item.status === "succeeded").length;
  const pending = items.filter((item) => item.status === "pending" || item.status === "processing").length;
  const failed = items.filter((item) => item.status === "failed" || item.status === "dead").length;
  return `${items.length} 条投递 / 成功 ${succeeded} / 待处理 ${pending} / 失败 ${failed}`;
}

type RoleLabel = (roleCode: string) => ReactNode;

type BusinessWorkflowStatusProps = {
  overview: WorkflowOverview | null | undefined;
  resourceNames: string | string[];
  resourceId?: number | string | null;
  resourceNo?: string | null;
  baseStatus: ReactNode;
  roleLabel?: RoleLabel;
};

export function BusinessWorkflowStatus({
  overview,
  resourceNames,
  resourceId,
  resourceNo,
  baseStatus,
  roleLabel = (roleCode) => roleCode
}: BusinessWorkflowStatusProps) {
  const { instances, tasks } = findBusinessWorkflowItems(overview, resourceNames, resourceId, resourceNo);
  const instance = instances[0] || null;
  if (!instance) return baseStatus;
  const task = currentBusinessWorkflowTask(instance, tasks);
  return (
    <span className="approval-inline-status">
      {baseStatus}
      {instance.status === "pending" && task ? (
        <span className="approval-inline-badge">流程：{roleLabel(task.roleCode)}</span>
      ) : (
        <span className="approval-inline-badge">流程：{statusLabel(instance.status)}</span>
      )}
    </span>
  );
}

type BusinessWorkflowTimelineProps = {
  overview: WorkflowOverview | null | undefined;
  resourceNames: string | string[];
  resourceId?: number | string | null;
  resourceNo?: string | null;
  emptyText?: string;
  roleLabel?: RoleLabel;
  comment?: string;
  commentLabel?: string;
  busy?: boolean;
  canActTask?: (task: WorkflowTask | null) => boolean;
  onCommentChange?: (value: string) => void;
  onActTask?: (task: WorkflowTask, action: "approve" | "reject") => void;
};

export function BusinessWorkflowTimeline({
  overview,
  resourceNames,
  resourceId,
  resourceNo,
  emptyText = "暂无工作流",
  roleLabel = (roleCode) => roleCode,
  comment = "",
  commentLabel = "流程意见",
  busy = false,
  canActTask = () => false,
  onCommentChange,
  onActTask
}: BusinessWorkflowTimelineProps) {
  const { instances, tasks, logs, events, outbox, deliveries } = findBusinessWorkflowItems(overview, resourceNames, resourceId, resourceNo);
  const instance = instances[0] || null;
  if (!instance) {
    const event = events[0] || null;
    return (
      <div className="workflow-business-block">
        <div className="workflow-approval-head">
          <b>工作流</b>
          <StatusChip value={event?.status || "none"} />
        </div>
        {event ? (
          <>
            <span>事件：{event.eventNo} / {event.eventType} / {statusLabel(event.status)}</span>
            {event.reason ? <span>原因：{event.reason}</span> : null}
            {event.error ? <span>异常：{event.error}</span> : null}
          </>
        ) : <span>{emptyText}</span>}
      </div>
    );
  }

  const currentTask = currentBusinessWorkflowTask(instance, tasks);
  const instanceTasks = tasks.filter((task) => task.instanceId === instance.id);
  const instanceLogs = logs.filter((log) => log.instanceId === instance.id);
  const triggerEvent = events.find((event) => event.id === instance.triggerEventId) || events[0] || null;
  const canAct = canActTask(currentTask) && Boolean(currentTask && onActTask);

  return (
    <div className="workflow-business-block">
      <div className="workflow-approval-head">
        <b>{instance.definitionName || instance.definitionCode}</b>
        <StatusChip value={instance.status} />
      </div>
      <span>{instance.instanceNo} / {instance.title || instance.resourceNo}</span>
      {triggerEvent ? <span>触发事件：{triggerEvent.eventNo} / {triggerEvent.eventType} / {statusLabel(triggerEvent.status)}</span> : null}
      <span>
        当前节点：
        {currentTask ? (
          <>{currentTask.stepName || currentTask.stepCode} / {roleLabel(currentTask.roleCode)}</>
        ) : statusLabel(instance.status)}
      </span>
      {instance.reason ? <span>原因：{instance.reason}</span> : null}
      {outbox.length ? <span>出口：{outbox.length} 个事件 / {businessWorkflowDeliverySummary(deliveries)}</span> : null}
      {currentTask && canAct ? (
        <>
          <Field label={commentLabel}>
            <TextInput value={comment} onChange={(event) => onCommentChange?.(event.target.value)} />
          </Field>
          <ActionGroup className="compact-actions">
            <Button type="button" variant="primary" icon={<CheckCircle2 size={13} />} disabled={busy} onClick={() => onActTask?.(currentTask, "approve")}>通过</Button>
            <Button type="button" disabled={busy} onClick={() => onActTask?.(currentTask, "reject")}>驳回</Button>
          </ActionGroup>
        </>
      ) : currentTask ? <span>待 {roleLabel(currentTask.roleCode)} 处理</span> : null}
      <div className="workflow-business-steps">
        {instanceTasks.map((task) => (
          <span className="workflow-business-step" key={task.id}>
            <b>{task.step}</b>{task.stepName || task.stepCode} / {roleLabel(task.roleCode)} / {statusLabel(task.status)}
          </span>
        ))}
      </div>
      <div className="workflow-business-log">
        {instanceLogs.slice(-5).map((log) => (
          <span key={log.id}>{log.createdAt} / {log.action} / {statusLabel(log.status)}{log.actor ? ` / ${log.actor}` : ""}</span>
        ))}
        {!instanceLogs.length ? <span>暂无流程日志</span> : null}
      </div>
    </div>
  );
}
