import { type ReactNode, useEffect, useMemo } from "react";
import { createPortal } from "react-dom";
import { BareButton } from "./Button";
import { cx } from "./utils";

export type ContextMenuActionItem = {
  key: string;
  label: ReactNode;
  icon?: ReactNode;
  hint?: ReactNode;
  disabled?: boolean;
  danger?: boolean;
  onSelect?: () => void;
};

export type ContextMenuSeparatorItem = {
  key: string;
  type: "separator";
};

export type ContextMenuItem = ContextMenuActionItem | ContextMenuSeparatorItem;

export type ContextMenuPosition = {
  x: number;
  y: number;
};

export type ContextMenuProps = {
  items: ContextMenuItem[];
  position: ContextMenuPosition;
  onClose: () => void;
  className?: string;
  label?: string;
  width?: number;
};

function isSeparator(item: ContextMenuItem): item is ContextMenuSeparatorItem {
  return "type" in item && item.type === "separator";
}

function normalizeItems(items: ContextMenuItem[]) {
  const result: ContextMenuItem[] = [];
  items.forEach((item) => {
    if (isSeparator(item)) {
      if (!result.length || isSeparator(result[result.length - 1])) return;
      result.push(item);
      return;
    }
    result.push(item);
  });

  while (result.length && isSeparator(result[result.length - 1])) {
    result.pop();
  }
  return result;
}

function clampedPosition(position: ContextMenuPosition, width: number, itemCount: number) {
  const viewportWidth = typeof window === "undefined" ? width + position.x + 8 : window.innerWidth;
  const viewportHeight = typeof window === "undefined" ? 360 + position.y + 8 : window.innerHeight;
  const estimatedHeight = Math.min(420, 10 + itemCount * 34);
  return {
    x: Math.max(8, Math.min(position.x, viewportWidth - width - 8)),
    y: Math.max(8, Math.min(position.y, viewportHeight - estimatedHeight - 8))
  };
}

export function ContextMenu({ items, position, onClose, className, label = "快捷操作", width = 184 }: ContextMenuProps) {
  const normalizedItems = useMemo(() => normalizeItems(items), [items]);
  const menuPosition = clampedPosition(position, width, normalizedItems.length);

  useEffect(() => {
    function closeByKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("click", onClose);
    window.addEventListener("resize", onClose);
    window.addEventListener("blur", onClose);
    window.addEventListener("keydown", closeByKey);
    window.addEventListener("scroll", onClose, true);
    return () => {
      window.removeEventListener("click", onClose);
      window.removeEventListener("resize", onClose);
      window.removeEventListener("blur", onClose);
      window.removeEventListener("keydown", closeByKey);
      window.removeEventListener("scroll", onClose, true);
    };
  }, [onClose]);

  if (!normalizedItems.length) {
    return null;
  }

  return createPortal(
    <div
      className={cx("ui-context-menu", className)}
      style={{ left: menuPosition.x, top: menuPosition.y, width }}
      role="menu"
      aria-label={label}
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => {
        event.preventDefault();
        event.stopPropagation();
      }}
    >
      {normalizedItems.map((item) => {
        if (isSeparator(item)) {
          return <div className="ui-context-menu__separator" role="separator" key={item.key} />;
        }

        return (
          <BareButton
            className={cx("ui-context-menu__item", item.danger ? "is-danger" : "")}
            disabled={item.disabled}
            key={item.key}
            role="menuitem"
            onClick={() => {
              if (item.disabled) return;
              item.onSelect?.();
              onClose();
            }}
          >
            {item.icon ? <span className="ui-context-menu__icon" aria-hidden="true">{item.icon}</span> : null}
            <span className="ui-context-menu__label">{item.label}</span>
            {item.hint ? <span className="ui-context-menu__hint">{item.hint}</span> : null}
          </BareButton>
        );
      })}
    </div>,
    document.body
  );
}
