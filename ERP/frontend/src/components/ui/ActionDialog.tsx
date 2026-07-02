import { Menu } from "lucide-react";
import { createContext, type ReactNode, useContext } from "react";
import { Button, type ButtonProps } from "./Button";
import { Dialog, type DialogSize, type DialogTone } from "./Dialog";
import { cx } from "./utils";

export type ActionDialogProps = {
  open: boolean;
  title: ReactNode;
  buttonLabel?: ReactNode;
  disabled?: boolean;
  triggerIcon?: ReactNode;
  triggerVariant?: ButtonProps["variant"];
  showTrigger?: boolean;
  size?: DialogSize;
  className?: string;
  bodyClassName?: string;
  closeDisabled?: boolean;
  feedback?: ReactNode;
  feedbackTone?: DialogTone;
  onOpen: () => void;
  onClose: () => void;
  children: ReactNode;
};

function isDetailDialog(title: ReactNode, buttonLabel: ReactNode) {
  const label = typeof buttonLabel === "string" ? buttonLabel : "";
  const titleText = typeof title === "string" ? title : "";
  return label === "明细" || label === "详情" || titleText.includes("明细") || titleText.includes("详情");
}

export function ActionDialog({
  open,
  title,
  buttonLabel = "操作",
  disabled = false,
  triggerIcon = <Menu size={13} />,
  triggerVariant = "soft",
  showTrigger = true,
  size,
  className,
  bodyClassName,
  closeDisabled = false,
  feedback,
  feedbackTone = "success",
  onOpen,
  onClose,
  children
}: ActionDialogProps) {
  const detail = isDetailDialog(title, buttonLabel);
  return (
    <>
      {showTrigger ? <Button icon={triggerIcon} variant={triggerVariant} disabled={disabled} onClick={onOpen}>{buttonLabel}</Button> : null}
      <Dialog
        open={open}
        title={title}
        size={size || (detail ? "xl" : "lg")}
        className={cx("action-dialog", detail ? "detail-dialog" : "", className)}
        bodyClassName={cx("dialog-form", bodyClassName)}
        closeDisabled={closeDisabled}
        feedback={feedback}
        feedbackTone={feedbackTone}
        onClose={onClose}
      >
        {children}
      </Dialog>
    </>
  );
}

type ActionDialogScopeContextValue = {
  activeId: string | null;
  closeDisabled?: boolean;
  feedback?: ReactNode;
  feedbackTone?: DialogTone;
  onActiveIdChange: (id: string | null) => void;
  onBeforeOpen?: () => void;
};

const ActionDialogScopeContext = createContext<ActionDialogScopeContextValue | null>(null);

export type ActionDialogScopeProps = ActionDialogScopeContextValue & {
  children: ReactNode;
};

export function ActionDialogScope({ children, ...value }: ActionDialogScopeProps) {
  return (
    <ActionDialogScopeContext.Provider value={value}>
      {children}
    </ActionDialogScopeContext.Provider>
  );
}

export type ScopedActionDialogProps = Omit<ActionDialogProps, "open" | "onOpen" | "onClose" | "closeDisabled" | "feedback" | "feedbackTone"> & {
  id: string;
  onClose?: () => void;
  onOpen?: () => void;
};

export function ScopedActionDialog({ id, onClose, onOpen, ...props }: ScopedActionDialogProps) {
  const scope = useContext(ActionDialogScopeContext);
  if (!scope) {
    throw new Error("ScopedActionDialog must be rendered inside ActionDialogScope.");
  }

  const open = scope.activeId === id;
  return (
    <ActionDialog
      {...props}
      open={open}
      closeDisabled={scope.closeDisabled}
      feedback={open ? scope.feedback : undefined}
      feedbackTone={scope.feedbackTone}
      onOpen={() => {
        onOpen?.();
        scope.onBeforeOpen?.();
        scope.onActiveIdChange(id);
      }}
      onClose={() => {
        onClose?.();
        scope.onActiveIdChange(scope.activeId === id ? null : scope.activeId);
      }}
    />
  );
}
