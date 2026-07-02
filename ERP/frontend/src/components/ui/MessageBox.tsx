import { AlertCircle, CheckCircle2, Info } from "lucide-react";
import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from "react";
import { Button } from "./Button";
import { Dialog } from "./Dialog";
import { cx } from "./utils";

export type MessageBoxTone = "error" | "success" | "warning" | "info";
type MessageBoxButtonVariant = "primary" | "soft" | "danger" | "ghost";

export type MessageBoxOptions = {
  title?: ReactNode;
  message: ReactNode;
  tone?: MessageBoxTone;
  confirmLabel?: ReactNode;
  cancelLabel?: ReactNode;
  confirmVariant?: MessageBoxButtonVariant;
  onConfirm?: () => void;
  onCancel?: () => void;
  onClose?: () => void;
};

type MessageBoxState = MessageBoxOptions & {
  id: number;
  mode: "message" | "confirm";
  resolve?: (confirmed: boolean) => void;
};

type MessageBoxContextValue = {
  showMessage: (options: MessageBoxOptions) => void;
  confirmMessage: (options: MessageBoxOptions) => Promise<boolean>;
  showError: (error: unknown, fallback?: string, title?: ReactNode) => void;
  closeMessage: () => void;
};

const MessageBoxContext = createContext<MessageBoxContextValue | null>(null);

function messageFromError(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) {
    return error.message;
  }
  if (typeof error === "string" && error.trim()) {
    return error;
  }
  return fallback;
}

function titleForTone(tone: MessageBoxTone) {
  if (tone === "error") return "异常提示";
  if (tone === "success") return "操作完成";
  if (tone === "warning") return "提示";
  return "消息";
}

function iconForTone(tone: MessageBoxTone) {
  if (tone === "success") return <CheckCircle2 size={22} />;
  if (tone === "info") return <Info size={22} />;
  return <AlertCircle size={22} />;
}

export type MessageBoxProps = {
  open: boolean;
  title?: ReactNode;
  message?: ReactNode;
  tone?: MessageBoxTone;
  confirmLabel?: ReactNode;
  cancelLabel?: ReactNode;
  confirmVariant?: MessageBoxButtonVariant;
  onCancel?: () => void;
  onClose: () => void;
};

export function MessageBox({
  open,
  title,
  message,
  tone = "info",
  confirmLabel = "知道了",
  cancelLabel,
  confirmVariant = "primary",
  onCancel,
  onClose
}: MessageBoxProps) {
  const footer = cancelLabel ? (
    <>
      <Button onClick={onCancel || onClose}>{cancelLabel}</Button>
      <Button variant={confirmVariant} onClick={onClose}>{confirmLabel}</Button>
    </>
  ) : (
    <Button variant={confirmVariant} onClick={onClose}>{confirmLabel}</Button>
  );

  return (
    <Dialog
      open={open}
      title={title || titleForTone(tone)}
      size="sm"
      className={cx("ui-message-box", `ui-message-box--${tone}`)}
      bodyClassName="ui-message-box__body"
      footer={footer}
      onClose={onCancel || onClose}
    >
      <div className="ui-message-box__content">
        <span className="ui-message-box__icon" aria-hidden="true">{iconForTone(tone)}</span>
        <p>{message}</p>
      </div>
    </Dialog>
  );
}

export type MessageBoxProviderProps = {
  children: ReactNode;
};

export function MessageBoxProvider({ children }: MessageBoxProviderProps) {
  const [message, setMessage] = useState<MessageBoxState | null>(null);

  const showMessage = useCallback((options: MessageBoxOptions) => {
    setMessage({ ...options, tone: options.tone || "info", mode: "message", id: Date.now() });
  }, []);

  const confirmMessage = useCallback((options: MessageBoxOptions) => {
    return new Promise<boolean>((resolve) => {
      setMessage({
        ...options,
        tone: options.tone || "warning",
        confirmLabel: options.confirmLabel || "确认执行",
        cancelLabel: options.cancelLabel || "取消",
        confirmVariant: options.confirmVariant || "primary",
        mode: "confirm",
        resolve,
        id: Date.now()
      });
    });
  }, []);

  const showError = useCallback((error: unknown, fallback = "操作失败", title: ReactNode = "异常提示") => {
    showMessage({
      title,
      tone: "error",
      message: messageFromError(error, fallback)
    });
  }, [showMessage]);

  const closeMessage = useCallback(() => {
    setMessage((current) => {
      current?.onConfirm?.();
      current?.onClose?.();
      current?.resolve?.(true);
      return null;
    });
  }, []);

  const cancelMessage = useCallback(() => {
    setMessage((current) => {
      current?.onCancel?.();
      current?.onClose?.();
      current?.resolve?.(false);
      return null;
    });
  }, []);

  const value = useMemo<MessageBoxContextValue>(() => ({
    showMessage,
    confirmMessage,
    showError,
    closeMessage
  }), [closeMessage, confirmMessage, showError, showMessage]);

  return (
    <MessageBoxContext.Provider value={value}>
      {children}
      <MessageBox
        open={Boolean(message)}
        title={message?.title}
        message={message?.message}
        tone={message?.tone}
        confirmLabel={message?.confirmLabel}
        cancelLabel={message?.cancelLabel}
        confirmVariant={message?.confirmVariant}
        onCancel={cancelMessage}
        onClose={closeMessage}
      />
    </MessageBoxContext.Provider>
  );
}

export function useMessageBox() {
  const value = useContext(MessageBoxContext);
  if (!value) {
    throw new Error("useMessageBox must be used inside MessageBoxProvider.");
  }
  return value;
}
