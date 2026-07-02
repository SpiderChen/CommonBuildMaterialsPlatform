import { type FormHTMLAttributes, type HTMLAttributes, type ReactNode } from "react";
import { cx } from "./utils";

export type FormGridProps = FormHTMLAttributes<HTMLFormElement> & {
  children: ReactNode;
};

export function FormGrid({ children, className, ...props }: FormGridProps) {
  return (
    <form className={cx("ui-form-grid", className)} {...props}>
      {children}
    </form>
  );
}

export type DialogFormProps = FormGridProps;

export function DialogForm({ children, className, ...props }: DialogFormProps) {
  return (
    <FormGrid className={cx("dialog-form", className)} {...props}>
      {children}
    </FormGrid>
  );
}

export type SystemFormProps = FormGridProps;

export function SystemForm({ children, className, ...props }: SystemFormProps) {
  return (
    <FormGrid className={cx("system-form", className)} {...props}>
      {children}
    </FormGrid>
  );
}

export type LoginFormProps = FormGridProps;

export function LoginForm({ children, className, ...props }: LoginFormProps) {
  return (
    <form className={cx("ui-login-form", className)} {...props}>
      {children}
    </form>
  );
}

export type QuickFormProps = FormGridProps;

export function QuickForm({ children, className, ...props }: QuickFormProps) {
  return (
    <form className={cx("ui-quick-form", className)} {...props}>
      {children}
    </form>
  );
}

export type InlineFormProps = FormGridProps;

export function InlineForm({ children, className, ...props }: InlineFormProps) {
  return (
    <form className={cx("ui-inline-form", className)} {...props}>
      {children}
    </form>
  );
}

export type WorkflowFormProps = FormGridProps;

export function WorkflowForm({ children, className, ...props }: WorkflowFormProps) {
  return (
    <form className={cx("ui-workflow-form", className)} {...props}>
      {children}
    </form>
  );
}

export type FormActionsProps = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
  spanAll?: boolean;
};

export function FormActions({ children, className, spanAll = false, ...props }: FormActionsProps) {
  return (
    <div className={cx("ui-form-actions", spanAll ? "ui-form-actions--span-all" : "", className)} {...props}>
      {children}
    </div>
  );
}
