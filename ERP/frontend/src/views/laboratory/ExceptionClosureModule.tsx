import { CheckCircle2, XCircle } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { StatusChip } from "../../components/StatusChip";
import { api } from "../../services/api";
import type { QualityException } from "../../services/types";
import type { ExceptionForm, MutateAction, SubmitHandler } from "./LaboratoryModuleTypes";
import { exceptionSummary } from "./laboratoryHelpers";

type Props = {
  exceptions: QualityException[];
  openExceptions: QualityException[];
  exceptionForm: ExceptionForm;
  setExceptionForm: Dispatch<SetStateAction<ExceptionForm>>;
  busy: string;
  mutate: MutateAction;
  onSubmitException: SubmitHandler;
};

export function ExceptionClosureModule({
  exceptions,
  openExceptions,
  exceptionForm,
  setExceptionForm,
  busy,
  mutate,
  onSubmitException
}: Props) {
  return (
    <section className="grid-12 laboratory-module">
      <div className="panel span-4">
        <div className="between">
          <h3>创建异常</h3>
          <span className="muted">{exceptionSummary(exceptions)}</span>
        </div>
        <form className="lab-form" onSubmit={onSubmitException}>
          <label><span>标题</span><input value={exceptionForm.title} onChange={(event) => setExceptionForm({ ...exceptionForm, title: event.target.value })} /></label>
          <label>
            <span>等级</span>
            <select value={exceptionForm.severity} onChange={(event) => setExceptionForm({ ...exceptionForm, severity: event.target.value })}>
              <option value="low">低</option>
              <option value="medium">中</option>
              <option value="high">高</option>
            </select>
          </label>
          <label><span>责任人</span><input value={exceptionForm.responsible} onChange={(event) => setExceptionForm({ ...exceptionForm, responsible: event.target.value })} /></label>
          <label className="span-all"><span>描述</span><textarea value={exceptionForm.description} onChange={(event) => setExceptionForm({ ...exceptionForm, description: event.target.value })} /></label>
          <button className="primary-button icon-button-text" type="submit" disabled={busy !== ""}><XCircle size={16} />创建异常</button>
        </form>
      </div>

      <div className="panel span-8">
        <div className="between">
          <h3>异常闭环</h3>
          <span className="muted">{openExceptions.length} 个待处理</span>
        </div>
        <div className="record-list compact-row">
          {openExceptions.slice(0, 8).map((item) => (
            <article className="record-card" key={item.id}>
              <div className="between compact-row">
                <strong>{item.title}</strong>
                <StatusChip value={item.severity} />
              </div>
              <p>{item.responsible || "-"} · {item.createdAt}</p>
              <button className="soft-button icon-button-text" type="button" disabled={busy !== ""} onClick={() => void mutate("handle-exception", () => api.handleQualityException(item.id, { rootCause: "已复核原因", correctiveAction: "已完成纠正措施" }))}>
                <CheckCircle2 size={15} />关闭
              </button>
            </article>
          ))}
          {!openExceptions.length ? <p className="muted">暂无待处理异常</p> : null}
        </div>
      </div>
    </section>
  );
}
