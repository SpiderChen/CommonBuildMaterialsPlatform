import { type ButtonHTMLAttributes, type HTMLAttributes, type ReactNode } from "react";
import { cx } from "./utils";

type SurfaceElement = "section" | "article" | "div" | "aside";

type SurfaceProps = HTMLAttributes<HTMLElement> & {
  as?: SurfaceElement;
  children: ReactNode;
  tone?: "default" | "muted" | "warning" | "danger";
};

export function Panel({ as: Component = "section", children, className, tone = "default", ...props }: SurfaceProps) {
  return (
    <Component className={cx("ui-panel", `ui-panel--${tone}`, className)} data-slot="ui-panel" {...props}>
      {children}
    </Component>
  );
}

export function Card({ as: Component = "article", children, className, tone = "default", ...props }: SurfaceProps) {
  return (
    <Component className={cx("ui-card", `ui-card--${tone}`, className)} data-slot="ui-card" {...props}>
      {children}
    </Component>
  );
}

export type SelectableCardProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  selected?: boolean;
  tone?: "default" | "muted" | "warning" | "danger";
};

export function SelectableCard({ children, className, selected = false, tone = "default", type = "button", ...props }: SelectableCardProps) {
  return (
    <button className={cx("ui-selectable-card", `ui-selectable-card--${tone}`, selected ? "is-selected selected" : "", className)} type={type} data-slot="ui-selectable-card" {...props}>
      {children}
    </button>
  );
}

export type EmptyStateProps = HTMLAttributes<HTMLDivElement> & {
  title?: ReactNode;
  action?: ReactNode;
};

export function EmptyState({ title = "暂无数据", action, children, className, ...props }: EmptyStateProps) {
  return (
    <div className={cx("ui-empty-state", className)} data-slot="ui-empty-state" {...props}>
      <div>
        {title ? <b>{title}</b> : null}
        {children ? <span>{children}</span> : null}
      </div>
      {action}
    </div>
  );
}

export type FeedbackBannerProps = HTMLAttributes<HTMLElement> & {
  children: ReactNode;
  tone?: "success" | "error" | "warning" | "info";
};

export function FeedbackBanner({ children, className, tone = "info", ...props }: FeedbackBannerProps) {
  return (
    <section className={cx("ui-feedback-banner", tone, className)} data-slot="ui-feedback-banner" {...props}>
      {children}
    </section>
  );
}

export type ActionGroupProps = HTMLAttributes<HTMLElement> & {
  as?: "div" | "span";
  children: ReactNode;
};

export function ActionGroup({ as: Component = "div", children, className, ...props }: ActionGroupProps) {
  return (
    <Component className={cx("ui-action-group", className)} data-slot="ui-action-group" {...props}>
      {children}
    </Component>
  );
}
