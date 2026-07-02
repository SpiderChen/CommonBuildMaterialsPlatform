import { forwardRef, type AnchorHTMLAttributes, type ButtonHTMLAttributes, type ReactNode } from "react";
import { cx } from "./utils";

type ButtonVariant = "primary" | "soft" | "danger" | "ghost";
type ButtonSize = "sm" | "md";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: ReactNode;
};

export function Button({ variant = "soft", size = "md", icon, children, className, type = "button", ...props }: ButtonProps) {
  return (
    <button
      className={cx("ui-button", `ui-button--${variant}`, `ui-button--${size}`, icon && children ? "ui-button--with-icon" : "", className)}
      type={type}
      data-slot="ui-button"
      {...props}
    >
      {icon}
      {children}
    </button>
  );
}

export type BareButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  active?: boolean;
};

export const BareButton = forwardRef<HTMLButtonElement, BareButtonProps>(function BareButton({ active = false, children, className, type = "button", ...props }, ref) {
  return (
    <button ref={ref} className={cx("ui-bare-button", active ? "is-active active" : "", className)} type={type} data-slot="ui-bare-button" {...props}>
      {children}
    </button>
  );
});

export type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon: ReactNode;
  label: string;
  variant?: ButtonVariant;
};

export function IconButton({ icon, label, variant = "ghost", className, type = "button", ...props }: IconButtonProps) {
  return (
    <button
      aria-label={label}
      className={cx("ui-icon-button", `ui-icon-button--${variant}`, className)}
      data-slot="ui-icon-button"
      title={props.title || label}
      type={type}
      {...props}
    >
      {icon}
    </button>
  );
}

export type ChipButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  active?: boolean;
  icon?: ReactNode;
};

export function ChipButton({ active = false, icon, children, className, type = "button", ...props }: ChipButtonProps) {
  return (
    <button
      className={cx("ui-chip-button", active ? "is-active active" : "", icon && children ? "ui-chip-button--with-icon" : "", className)}
      type={type}
      data-slot="ui-chip-button"
      {...props}
    >
      {icon}
      {children}
    </button>
  );
}

export type ButtonLinkProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: ReactNode;
};

export function ButtonLink({ variant = "soft", size = "md", icon, children, className, ...props }: ButtonLinkProps) {
  return (
    <a className={cx("ui-button", `ui-button--${variant}`, `ui-button--${size}`, icon && children ? "ui-button--with-icon" : "", className)} {...props}>
      {icon}
      {children}
    </a>
  );
}
