import { X } from "lucide-react";
import { type CSSProperties, type HTMLAttributes, type MouseEvent, type ReactNode, useEffect } from "react";
import { createPortal } from "react-dom";
import { IconButton } from "./Button";
import { cx } from "./utils";

export type DialogSize = "sm" | "md" | "lg" | "xl" | "wide";
export type DialogTone = "success" | "error";

export type DialogProps = {
  open: boolean;
  title: ReactNode;
  description?: ReactNode;
  ariaLabel?: string;
  size?: DialogSize;
  className?: string;
  backdropClassName?: string;
  backdropStyle?: CSSProperties;
  style?: CSSProperties;
  bodyClassName?: string;
  children: ReactNode;
  closeDisabled?: boolean;
  closeLabel?: string;
  feedback?: ReactNode;
  feedbackTone?: DialogTone;
  footer?: ReactNode;
  onClose?: () => void;
};

export function Dialog({
  open,
  title,
  description,
  ariaLabel,
  size = "md",
  className,
  backdropClassName,
  backdropStyle,
  style,
  bodyClassName,
  children,
  closeDisabled = false,
  closeLabel = "关闭",
  feedback,
  feedbackTone = "success",
  footer,
  onClose
}: DialogProps) {
  useEffect(() => {
    if (!open || !onClose || closeDisabled) return undefined;
    function closeByEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose?.();
      }
    }
    window.addEventListener("keydown", closeByEscape);
    return () => window.removeEventListener("keydown", closeByEscape);
  }, [closeDisabled, onClose, open]);

  function closeByBackdrop(event: MouseEvent<HTMLDivElement>) {
    if (event.target === event.currentTarget && onClose && !closeDisabled) {
      onClose();
    }
  }

  if (!open) {
    return null;
  }

  return createPortal(
    <div
      className={cx("ui-dialog-backdrop", backdropClassName)}
      data-slot="ui-dialog-backdrop"
      role="presentation"
      style={backdropStyle}
      onMouseDown={closeByBackdrop}
    >
      <section
        className={cx("ui-dialog", `ui-dialog--${size}`, className)}
        data-size={size}
        data-slot="ui-dialog"
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel || String(title)}
        style={style}
      >
        <div className="ui-dialog__header" data-slot="ui-dialog-header">
          <div className="ui-dialog__title-block" data-slot="ui-dialog-title-block">
            <h3>{title}</h3>
            {description ? <p className="ui-dialog__description">{description}</p> : null}
          </div>
          {onClose ? <IconButton icon={<X size={16} />} label={closeLabel} disabled={closeDisabled} onClick={onClose} /> : null}
        </div>
        {feedback ? (
          <section className={cx("ui-dialog__feedback", `ui-dialog__feedback--${feedbackTone}`)} data-slot="ui-dialog-feedback">
            <span>{feedback}</span>
          </section>
        ) : null}
        <div className={cx("ui-dialog__body", bodyClassName)} data-slot="ui-dialog-body">
          {children}
        </div>
        {footer ? <div className="ui-dialog__footer" data-slot="ui-dialog-footer">{footer}</div> : null}
      </section>
    </div>,
    document.body
  );
}

export type DialogContentProps = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
};

export function DialogContent({ children, className, ...props }: DialogContentProps) {
  return (
    <div className={cx("ui-dialog-content", className)} data-slot="ui-dialog-content" {...props}>
      {children}
    </div>
  );
}
