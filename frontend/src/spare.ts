import { ApiError } from "./api";
import type { Copy } from "./i18n";

export const maxStock = 99999;
export const maxSpareNote = 4000;

export function nameFields(brand: string, model: string) {
  const missing: string[] = [];
  const tooLong: string[] = [];
  for (const [field, value] of [
    ["brand", brand],
    ["model", model],
  ] as const) {
    if (!value.trim()) {
      missing.push(field);
    } else if ([...value.trim()].length > 200) {
      tooLong.push(field);
    }
  }
  return { missing, tooLong };
}

export function rangeText(copy: Copy, min: number, max: number) {
  return copy.wholeNumber.replace("{min}", String(min)).replace("{max}", String(max));
}

export function spareProblem(copy: Copy, err: unknown, min: number, max: number) {
  if (!(err instanceof ApiError)) {
    return null;
  }
  if (err.code === "stock_short") {
    return copy.onlyInStock.replace("{n}", String(err.onHand));
  }
  if (err.code === "stock_cap") {
    return copy.stockCap;
  }
  if (err.code === "return_short") {
    return rangeText(copy, 1, err.returnable);
  }
  if (err.code === "validation" && err.missing.includes("quantity")) {
    return copy.quantityRequired;
  }
  if (err.code === "validation" && err.invalid.includes("quantity")) {
    return rangeText(copy, min, max);
  }
  if (err.code === "validation" && err.taken.includes("spare")) {
    return copy.spareTaken;
  }
  return null;
}

export function moveKind(copy: Copy, kind: string) {
  if (kind === "receipt") {
    return copy.moveAdded;
  }
  if (kind === "write_off") {
    return copy.moveWrittenOff;
  }
  if (kind === "use") {
    return copy.moveUsed;
  }
  if (kind === "return") {
    return copy.moveReturned;
  }
  return kind;
}

export function noteTooLong(value: string) {
  return [...value.trim()].length > maxSpareNote;
}
