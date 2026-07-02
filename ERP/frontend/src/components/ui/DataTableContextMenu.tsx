import { Copy, MousePointerClick } from "lucide-react";
import type { ReactNode } from "react";
import type { ContextMenuItem } from "./ContextMenu";
import type { DataTableContextMenuHelpers, DataTableProps } from "./DataTable";

export type DataTableCopyField<T> = {
  key: string;
  label: string;
  value: (row: T, helpers: DataTableContextMenuHelpers<T>) => unknown;
  icon?: ReactNode;
  hint?: ReactNode;
  disabled?: (row: T, helpers: DataTableContextMenuHelpers<T>) => boolean;
};

export type DataTableRowAction<T> = {
  key: string;
  label: ReactNode;
  icon?: ReactNode;
  hint?: ReactNode;
  danger?: boolean;
  disabled?: (row: T, helpers: DataTableContextMenuHelpers<T>) => boolean;
  onSelect: (row: T, helpers: DataTableContextMenuHelpers<T>) => void;
};

export type DataTableRowContextMenuOptions<T> = {
  actions?: DataTableRowAction<T>[];
  copyFields?: DataTableCopyField<T>[];
};

function menuValue(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (value instanceof Date) return value.toISOString();
  if (Array.isArray(value)) return value.map(menuValue).filter(Boolean).join(" / ");
  if (typeof value === "object") {
    try {
      return JSON.stringify(value);
    } catch {
      return String(value);
    }
  }
  return String(value);
}

export function buildDataTableRowContextMenu<T>({
  actions = [],
  copyFields = []
}: DataTableRowContextMenuOptions<T>): NonNullable<DataTableProps<T>["rowContextMenu"]> {
  return (row, _index, helpers) => {
    const actionItems: ContextMenuItem[] = actions.map((action) => ({
      key: `custom-action-${action.key}`,
      label: action.label,
      icon: action.icon || <MousePointerClick size={14} />,
      hint: action.hint,
      danger: action.danger,
      disabled: action.disabled?.(row, helpers),
      onSelect: () => action.onSelect(row, helpers)
    }));

    const copyItems: ContextMenuItem[] = copyFields.map((field) => {
      const value = menuValue(field.value(row, helpers));
      return {
        key: `custom-copy-${field.key}`,
        label: `复制${field.label}`,
        icon: field.icon || <Copy size={14} />,
        hint: field.hint,
        disabled: !value || field.disabled?.(row, helpers),
        onSelect: () => helpers.copyText(value, field.label)
      };
    });

    return [
      ...actionItems,
      ...(actionItems.length && copyItems.length ? [{ key: "custom-action-copy-separator", type: "separator" as const }] : []),
      ...copyItems
    ];
  };
}
