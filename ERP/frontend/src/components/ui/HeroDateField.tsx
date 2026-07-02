import { type ChangeEvent } from "react";
import { Field, TextInput } from "./Field";

export type HeroDateFieldProps = {
  className?: string;
  defaultValue?: string;
  disabled?: boolean;
  label: string;
  mode?: "date" | "date-time";
  name?: string;
  onChange?: (value: string) => void;
  value?: string;
};

function dateInputValue(value: string) {
  return value ? value.slice(0, 10) : "";
}

function dateTimeInputValue(value: string) {
  if (!value) return "";
  const normalized = value.replace(" ", "T").slice(0, 19);
  if (normalized.length === 10) return `${normalized}T00:00:00`;
  if (normalized.length === 16) return `${normalized}:00`;
  return normalized;
}

function dateTimeOutputValue(value: string) {
  if (!value) return "";
  return value.length === 16 ? `${value}:00` : value;
}

export function HeroDateField({ className, defaultValue, disabled, label, mode = "date", name, onChange, value }: HeroDateFieldProps) {
  const dateTime = mode === "date-time";
  const normalize = dateTime ? dateTimeInputValue : dateInputValue;
  const inputProps = onChange
    ? {
        value: normalize(value || ""),
        onChange: (event: ChangeEvent<HTMLInputElement>) => onChange(dateTime ? dateTimeOutputValue(event.target.value) : event.target.value)
      }
    : {
        defaultValue: normalize(defaultValue || "")
      };

  return (
    <Field className={className} label={label}>
      <TextInput
        {...inputProps}
        aria-label={label}
        className="ui-date-field"
        disabled={disabled}
        name={name}
        step={dateTime ? 1 : undefined}
        type={dateTime ? "datetime-local" : "date"}
      />
    </Field>
  );
}
