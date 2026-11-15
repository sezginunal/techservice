import type { Lang } from "./i18n";

export type NamedState = {
  code: string;
  nameEn: string;
  nameTr: string;
};

export function statusLabel(status: string, states: NamedState[], lang: Lang) {
  const item = states.find((state) => state.code === status);
  if (!item) {
    return status;
  }
  return lang === "tr" ? item.nameTr : item.nameEn;
}

export function choosableStates<T extends NamedState & { active: boolean }>(states: T[]) {
  return states.filter((state) => state.active);
}

export function listedStates<T extends NamedState & { active: boolean; used: boolean }>(states: T[]) {
  return states.filter((state) => state.active || state.used);
}

export function formatWhen(value: string, lang: Lang) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(lang === "tr" ? "tr-TR" : "en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function dayStart(value: string) {
  if (!value) {
    return "";
  }
  const [year, month, day] = value.split("-").map(Number);
  return new Date(year, month - 1, day).toISOString();
}

export function nextDay(value: string) {
  if (!value) {
    return "";
  }
  const [year, month, day] = value.split("-").map(Number);
  return new Date(year, month - 1, day + 1).toISOString();
}
