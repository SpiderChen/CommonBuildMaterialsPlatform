import {
  Children,
  type ChangeEvent,
  type CSSProperties,
  forwardRef,
  isValidElement,
  type InputHTMLAttributes,
  type LabelHTMLAttributes,
  type ReactElement,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
  useEffect,
  useImperativeHandle,
  useLayoutEffect,
  useMemo,
  useRef,
  useState
} from "react";
import { createPortal } from "react-dom";
import { cx } from "./utils";

export type FieldProps = LabelHTMLAttributes<HTMLLabelElement> & {
  label: ReactNode;
  children: ReactNode;
  spanAll?: boolean;
};

export function Field({ label, children, className, spanAll = false, ...props }: FieldProps) {
  return (
    <label className={cx("ui-field", spanAll ? "span-all" : "", className)} data-slot="ui-field" {...props}>
      <span>{label}</span>
      {children}
    </label>
  );
}

export type IconFieldProps = LabelHTMLAttributes<HTMLLabelElement> & {
  icon: ReactNode;
  children: ReactNode;
  label?: ReactNode;
};

export function IconField({ icon, label, children, className, ...props }: IconFieldProps) {
  return (
    <label className={cx("ui-icon-field", className)} data-slot="ui-icon-field" {...props}>
      {icon}
      {label ? <span className="sr-only">{label}</span> : null}
      {children}
    </label>
  );
}

export type TextInputProps = InputHTMLAttributes<HTMLInputElement>;

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(function TextInput({ className, ...props }, ref) {
  return <input ref={ref} className={cx("ui-text-input", className)} data-slot="ui-input" {...props} />;
});

export type SelectInputProps = SelectHTMLAttributes<HTMLSelectElement>;

type SelectOptionElementProps = {
  children?: ReactNode;
  disabled?: boolean;
  value?: string | number;
};

type SelectOptionItem = {
  disabled?: boolean;
  label: string;
  value: string;
};

type SelectMenuPlacement = {
  left: number;
  top: number;
  width: number;
  maxHeight: number;
};

function textFromNode(node: ReactNode): string {
  if (node === null || node === undefined || typeof node === "boolean") return "";
  if (typeof node === "string" || typeof node === "number" || typeof node === "bigint") return String(node);
  if (Array.isArray(node)) return node.map(textFromNode).join("");
  if (isValidElement<{ children?: ReactNode }>(node)) return textFromNode(node.props.children);
  return "";
}

function optionsFromChildren(children: ReactNode): SelectOptionItem[] {
  return Children.toArray(children).flatMap((child) => {
    if (!isValidElement<SelectOptionElementProps>(child)) return [];
    const element = child as ReactElement<SelectOptionElementProps>;
    if (element.type === "option") {
      const label = textFromNode(element.props.children);
      return [{
        disabled: element.props.disabled,
        label,
        value: element.props.value === undefined ? label : String(element.props.value)
      }];
    }
    if (element.type === "optgroup") {
      return optionsFromChildren(element.props.children);
    }
    return [];
  });
}

function stringValue(value: SelectInputProps["value"] | SelectInputProps["defaultValue"] | undefined) {
  if (Array.isArray(value)) return String(value[0] ?? "");
  return value === undefined ? "" : String(value);
}

function selectMenuPlacement(anchor: HTMLElement): SelectMenuPlacement {
  const rect = anchor.getBoundingClientRect();
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth;
  const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
  const margin = 8;
  const gap = 4;
  const minHeight = 80;
  const maxHeight = 280;
  const availableBelow = viewportHeight - rect.bottom - margin - gap;
  const availableAbove = rect.top - margin - gap;
  const openBelow = availableBelow >= minHeight || availableBelow >= availableAbove;
  const available = Math.max(minHeight, Math.min(maxHeight, openBelow ? availableBelow : availableAbove));
  const width = Math.max(rect.width, 1);
  const left = Math.min(Math.max(margin, rect.left), Math.max(margin, viewportWidth - width - margin));
  const top = openBelow ? rect.bottom + gap : Math.max(margin, rect.top - available - gap);

  return { left, top, width, maxHeight: available };
}

export const SelectInput = forwardRef<HTMLSelectElement, SelectInputProps>(function SelectInput({
  className,
  children,
  defaultValue,
  disabled,
  multiple,
  onChange,
  value,
  ...props
}, ref) {
  const nativeRef = useRef<HTMLSelectElement>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const options = useMemo(() => optionsFromChildren(children), [children]);
  const controlled = value !== undefined;
  const initialValue = stringValue(value ?? defaultValue) || options.find((option) => !option.disabled)?.value || "";
  const [open, setOpen] = useState(false);
  const [internalValue, setInternalValue] = useState(initialValue);
  const [menuPlacement, setMenuPlacement] = useState<SelectMenuPlacement | null>(null);
  const selectedValue = controlled ? stringValue(value) : internalValue;
  const selectedOption = options.find((option) => option.value === selectedValue);
  const menuIsPortaled = open && Boolean(rootRef.current?.closest("[data-slot='ui-dialog']"));

  useImperativeHandle(ref, () => nativeRef.current as HTMLSelectElement);

  useEffect(() => {
    if (controlled) {
      setInternalValue(stringValue(value));
    }
  }, [controlled, value]);

  useEffect(() => {
    if (!controlled && !selectedValue && options.length) {
      setInternalValue(options.find((option) => !option.disabled)?.value || "");
    }
  }, [controlled, options, selectedValue]);

  useEffect(() => {
    if (!open) return undefined;
    function handlePointerDown(event: MouseEvent) {
      const target = event.target as Node;
      if (!rootRef.current?.contains(target) && !menuRef.current?.contains(target)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  useLayoutEffect(() => {
    if (!open || !menuIsPortaled) {
      setMenuPlacement(null);
      return undefined;
    }

    function updatePlacement() {
      if (rootRef.current) {
        setMenuPlacement(selectMenuPlacement(rootRef.current));
      }
    }

    updatePlacement();
    window.addEventListener("resize", updatePlacement);
    document.addEventListener("scroll", updatePlacement, true);
    return () => {
      window.removeEventListener("resize", updatePlacement);
      document.removeEventListener("scroll", updatePlacement, true);
    };
  }, [menuIsPortaled, open]);

  function commit(nextValue: string) {
    if (disabled) return;
    const option = options.find((item) => item.value === nextValue);
    if (option?.disabled) return;
    if (!controlled) {
      setInternalValue(nextValue);
    }
    setOpen(false);
    if (nativeRef.current) {
      nativeRef.current.value = nextValue;
      onChange?.({ target: nativeRef.current, currentTarget: nativeRef.current } as ChangeEvent<HTMLSelectElement>);
    }
  }

  if (multiple) {
    return (
      <select ref={ref} className={cx("ui-select-input", className)} data-slot="ui-select" disabled={disabled} multiple={multiple} onChange={onChange} value={value} defaultValue={defaultValue} {...props}>
        {children}
      </select>
    );
  }

  const menuStyle: CSSProperties | undefined = menuIsPortaled
    ? {
      left: menuPlacement?.left ?? 0,
      top: menuPlacement?.top ?? 0,
      width: menuPlacement?.width,
      maxHeight: menuPlacement?.maxHeight,
      visibility: menuPlacement ? undefined : "hidden"
    }
    : undefined;
  const menu = open ? (
    <div ref={menuRef} className={cx("ui-select-menu", menuIsPortaled ? "is-portaled" : "")} style={menuStyle} role="listbox">
      {options.map((option) => (
        <button
          type="button"
          key={option.value}
          className={cx("ui-select-option", option.value === selectedValue ? "is-selected" : "")}
          disabled={option.disabled}
          role="option"
          aria-selected={option.value === selectedValue}
          onClick={() => commit(option.value)}
        >
          {option.label}
        </button>
      ))}
    </div>
  ) : null;

  return (
    <div ref={rootRef} className={cx("ui-select-input", disabled ? "is-disabled" : "", className)} data-slot="ui-select">
      <select
        ref={nativeRef}
        className="ui-select-native"
        tabIndex={-1}
        aria-hidden="true"
        disabled={disabled}
        onChange={onChange}
        value={selectedValue}
        {...props}
      >
        {children}
      </select>
      <button
        type="button"
        className="ui-select-trigger"
        disabled={disabled}
        aria-expanded={open}
        aria-haspopup="listbox"
        onClick={() => setOpen((value) => !value)}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            setOpen(false);
          }
          if (event.key === "ArrowDown") {
            event.preventDefault();
            setOpen(true);
          }
        }}
      >
        <span>{selectedOption?.label || selectedValue || "-"}</span>
        <i aria-hidden="true" />
      </button>
      {menuIsPortaled && menu ? createPortal(menu, document.body) : menu}
    </div>
  );
});

export type TextAreaInputProps = TextareaHTMLAttributes<HTMLTextAreaElement>;

export const TextAreaInput = forwardRef<HTMLTextAreaElement, TextAreaInputProps>(function TextAreaInput({ className, ...props }, ref) {
  return <textarea ref={ref} className={cx("ui-textarea-input", className)} data-slot="ui-textarea" {...props} />;
});
