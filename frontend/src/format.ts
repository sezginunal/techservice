import type { Lang } from "./i18n";

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

export function statusLabel(status: string, registered: string) {
  return status === "registered" ? registered : status;
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
