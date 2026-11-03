import type { Copy, Lang } from "./i18n";

export const deviceStatuses = [
  "registered",
  "inspecting",
  "repair_warranty",
  "repair_out_of_warranty",
  "waiting_partner_part",
  "waiting_own_part",
  "final_testing",
  "preparing_delivery",
  "in_storage",
  "left_storage",
  "delivered",
] as const;

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

export function statusLabel(status: string, copy: Copy) {
  if (status in copy.statuses) {
    return copy.statuses[status as keyof typeof copy.statuses];
  }
  return status;
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
