import { AlertCircle, CheckCircle2, Info, X } from "lucide-react";
import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { cx } from "./utils";

export type MessageTone = "success" | "error" | "warning" | "info";

export type MessageOptions = {
  message: ReactNode;
  tone?: MessageTone;
  durationMs?: number;
};

export type MessageItem = Required<Pick<MessageOptions, "tone" | "durationMs">> & {
  id: number;
  message: ReactNode;
};

type MessageShortcutOptions = Omit<MessageOptions, "message" | "tone">;

export type MessageApi = {
  open: (options: MessageOptions) => number;
  success: (message: ReactNode, options?: MessageShortcutOptions) => number;
  error: (message: ReactNode, options?: MessageShortcutOptions) => number;
  warning: (message: ReactNode, options?: MessageShortcutOptions) => number;
  info: (message: ReactNode, options?: MessageShortcutOptions) => number;
  remove: (id: number) => void;
  clear: () => void;
};

const MessageContext = createContext<MessageApi | null>(null);

function defaultDuration(tone: MessageTone) {
  return tone === "error" ? 4500 : 2600;
}

function iconForTone(tone: MessageTone) {
  if (tone === "success") return <CheckCircle2 size={18} />;
  if (tone === "info") return <Info size={18} />;
  return <AlertCircle size={18} />;
}

export type MessageProps = {
  item: MessageItem;
  onClose: (id: number) => void;
};

export function Message({ item, onClose }: MessageProps) {
  useEffect(() => {
    if (item.durationMs <= 0) return undefined;
    const timer = window.setTimeout(() => onClose(item.id), item.durationMs);
    return () => window.clearTimeout(timer);
  }, [item.durationMs, item.id, onClose]);

  return (
    <section className={cx("ui-message", `ui-message--${item.tone}`)} role={item.tone === "error" ? "alert" : "status"}>
      <span className="ui-message__icon" aria-hidden="true">{iconForTone(item.tone)}</span>
      <div className="ui-message__content">{item.message}</div>
      <button className="ui-message__close" type="button" aria-label="关闭提示" onClick={() => onClose(item.id)}>
        <X size={14} />
      </button>
    </section>
  );
}

export type MessageProviderProps = {
  children: ReactNode;
};

export function MessageProvider({ children }: MessageProviderProps) {
  const [items, setItems] = useState<MessageItem[]>([]);
  const nextId = useRef(0);

  const remove = useCallback((id: number) => {
    setItems((current) => current.filter((item) => item.id !== id));
  }, []);

  const clear = useCallback(() => {
    setItems([]);
  }, []);

  const open = useCallback((options: MessageOptions) => {
    const tone = options.tone || "info";
    const id = nextId.current + 1;
    nextId.current = id;
    setItems((current) => [
      ...current.slice(-3),
      {
        id,
        message: options.message,
        tone,
        durationMs: options.durationMs ?? defaultDuration(tone)
      }
    ]);
    return id;
  }, []);

  const api = useMemo<MessageApi>(() => ({
    open,
    success: (message, options) => open({ ...options, message, tone: "success" }),
    error: (message, options) => open({ ...options, message, tone: "error" }),
    warning: (message, options) => open({ ...options, message, tone: "warning" }),
    info: (message, options) => open({ ...options, message, tone: "info" }),
    remove,
    clear
  }), [clear, open, remove]);

  return (
    <MessageContext.Provider value={api}>
      {children}
      {createPortal(
        <div className="ui-message-stack" aria-live="polite" aria-atomic="false">
          {items.map((item) => (
            <Message item={item} key={item.id} onClose={remove} />
          ))}
        </div>,
        document.body
      )}
    </MessageContext.Provider>
  );
}

export function useMessage() {
  const value = useContext(MessageContext);
  if (!value) {
    throw new Error("useMessage must be used inside MessageProvider.");
  }
  return value;
}
