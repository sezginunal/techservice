import { ApiError } from "./api";
import type { Lang } from "./i18n";

export function canExport(role: string) {
  return role === "owner" || role === "admin";
}

export async function downloadExcel(path: string, lang: Lang) {
  const params = new URLSearchParams(path.includes("?") ? path.slice(path.indexOf("?") + 1) : "");
  params.set("lang", lang);
  const zone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  if (zone) {
    params.set("tz", zone);
  }
  const base = path.includes("?") ? path.slice(0, path.indexOf("?")) : path;
  const response = await fetch(`${base}?${params}`);
  if (response.status === 401) {
    throw new ApiError("unauthorized");
  }
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as { error?: string };
    throw new ApiError(body.error ?? "unexpected");
  }
  const blob = await response.blob();
  const header = response.headers.get("Content-Disposition") ?? "";
  const match = /filename="([^"]+)"/.exec(header);
  const name = match?.[1] ?? "export.xlsx";
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
}
