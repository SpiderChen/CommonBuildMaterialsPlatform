import type { DataDictionary } from "./types";

export type DictionaryOption = {
  code: string;
  label: string;
};

export function activeDictionaryOptions(
  dictionaries: DataDictionary[] | null | undefined,
  type: string,
  fallbackItems: DictionaryOption[] = []
): DictionaryOption[] {
  const items = (Array.isArray(dictionaries) ? dictionaries : [])
    .filter((item) => item.type === type && (item.status === "" || item.status === "active"))
    .sort((a, b) => (a.sort || 0) - (b.sort || 0) || a.code.localeCompare(b.code))
    .map((item) => ({ code: item.code, label: item.label || item.code }));
  return items.length ? items : fallbackItems;
}

export function dictionaryLabel(
  dictionaries: DataDictionary[] | null | undefined,
  type: string,
  code: string | null | undefined,
  fallbackLabel?: string
) {
  const value = code || "";
  if (!value) return fallbackLabel || "-";
  const item = activeDictionaryOptions(dictionaries, type).find((candidate) => candidate.code === value);
  return item?.label || fallbackLabel || value || "-";
}
