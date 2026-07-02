import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "./utils";

export type LayoutElement = "div" | "section" | "aside" | "nav" | "main" | "header";

export type LayoutProps = HTMLAttributes<HTMLElement> & {
  as?: LayoutElement;
  children: ReactNode;
};

export type SectionGridProps = LayoutProps;
export type SectionHeaderProps = LayoutProps;
export type SplitRowProps = LayoutProps;
export type ViewStackProps = LayoutProps;
export type LayoutRegionProps = LayoutProps;

export function LayoutRegion({ as: Component = "div", children, className, ...props }: LayoutRegionProps) {
  return (
    <Component className={cx("ui-layout-region", className)} data-slot="ui-layout-region" {...props}>
      {children}
    </Component>
  );
}

export function ViewStack({ as: Component = "div", children, className, ...props }: ViewStackProps) {
  return (
    <Component className={cx("ui-view-stack", "view-stack", className)} data-slot="ui-view-stack" {...props}>
      {children}
    </Component>
  );
}

export function SectionGrid({ as: Component = "section", children, className, ...props }: LayoutProps) {
  return (
    <Component className={cx("ui-section-grid", "grid-12", className)} data-slot="ui-section-grid" {...props}>
      {children}
    </Component>
  );
}

export function SectionHeader({ as: Component = "div", children, className, ...props }: LayoutProps) {
  return (
    <Component className={cx("ui-section-header", className)} data-slot="ui-section-header" {...props}>
      {children}
    </Component>
  );
}

export function SplitRow({ as: Component = "div", children, className, ...props }: LayoutProps) {
  return (
    <Component className={cx("ui-split-row", "between", className)} data-slot="ui-split-row" {...props}>
      {children}
    </Component>
  );
}

export type MetricListProps = LayoutProps & {
  compact?: boolean;
};

export function MetricList({ as: Component = "div", children, className, compact = false, ...props }: MetricListProps) {
  return (
    <Component className={cx("ui-metric-list", "metric-list", compact ? "compact" : "", className)} data-slot="ui-metric-list" {...props}>
      {children}
    </Component>
  );
}

export type ChipListProps = LayoutProps & {
  compact?: boolean;
};

export function ChipList({ as: Component = "div", children, className, compact = false, ...props }: ChipListProps) {
  return (
    <Component className={cx("ui-chip-list", "permission-chip-list", compact ? "compact" : "", className)} data-slot="ui-chip-list" {...props}>
      {children}
    </Component>
  );
}
