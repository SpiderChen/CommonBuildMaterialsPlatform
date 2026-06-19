import type { BootstrapData } from "../services/types";

export function nameOf<T extends { id: number; name?: string; plateNo?: string; spec?: string }>(
  items: T[] | undefined,
  id: number,
  fallback = "-"
) {
  const item = items?.find((entry) => entry.id === id);
  if (!item) return fallback;
  if (item.plateNo) return item.plateNo;
  if (item.spec) return `${item.name || ""} ${item.spec}`;
  return item.name || fallback;
}

export function productLabel(data: BootstrapData | null, id: number) {
  const product = data?.products.find((item) => item.id === id);
  if (!product) return "-";
  return `${product.name} ${product.spec}`;
}

