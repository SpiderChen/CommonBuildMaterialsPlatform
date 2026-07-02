export type SensitiveActionPrompt = {
  title: string;
  message: string;
  confirmLabel: string;
  confirmVariant?: "primary" | "danger";
};

const destructivePattern = /删除|作废|取消|停用|禁用|关闭|结清|重置|回滚|驳回|失败|红字|冲销|清空|注销|吊销|MFA|密码|delete|void|cancel|retire|disable|reject|reset|rollback|fail/i;
const sensitivePattern = /删除|作废|取消|停用|禁用|关闭|结清|重置|回滚|通过|驳回|审批|确认|提交|发布|投递|重放|改派|升级|领取|失败|税务|税控|发票|红字|冲销|付款|收款|催收|短信|签收|补打|下达|完成|派车|重开|入库|调拨|结算|调整|转仓|库存|推进|状态|标记处理|自动化|修订|复核|校准|异常|用户|角色|权限|字典|工作流|订阅|客户|司机|车辆|绑定|MFA|approve|reject|retire|delete|void|cancel|submit|publish|dispatch|reset|rollback|replay|fail|settle|payment|invoice|tax|sign|advance|complete|adjust|transfer|reassign|escalate|status/i;
const externalPattern = /税务|税控|短信|催收|签收|投递|发布|重放|发送|领取|外部|接口/i;

export function sensitiveActionPrompt(label: string, success = label): SensitiveActionPrompt | null {
  const actionText = `${label} ${success}`.trim();
  if (!sensitivePattern.test(actionText)) {
    return null;
  }

  const actionName = success || label || "当前操作";
  const destructive = destructivePattern.test(actionText);
  const external = externalPattern.test(actionText);
  const message = external
    ? `即将执行“${actionName}”。该操作可能触发外部系统、通知或业务状态变更，请确认单据和参数已核对无误。`
    : `即将执行“${actionName}”。该操作会改变业务状态或关键数据，请确认已核对相关单据。`;

  return {
    title: destructive ? "确认危险操作" : "确认敏感操作",
    message,
    confirmLabel: destructive ? "确认执行" : "确认提交",
    confirmVariant: destructive ? "danger" : "primary"
  };
}
