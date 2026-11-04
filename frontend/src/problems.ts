import { ApiError } from "./api";
import type { Copy } from "./i18n";
import { messageFor } from "./Login";

const maxPhotoBytes = 8 * 1024 * 1024;
export const maxPhotos = 30;

export function fileProblem(file: File): "type" | "size" | null {
  const allowed =
    file.type === "image/jpeg" ||
    file.type === "image/png" ||
    file.type === "image/webp" ||
    (file.type === "" && /\.(jpe?g|png|webp)$/i.test(file.name));
  if (!allowed) {
    return "type";
  }
  if (file.size > maxPhotoBytes) {
    return "size";
  }
  return null;
}

export function problemText(copy: Copy, err: unknown) {
  if (!(err instanceof ApiError) || err.code !== "validation") {
    if (err instanceof ApiError && err.code === "not_found") {
      return copy.notFound;
    }
    return messageFor(copy, err);
  }
  const lines = [
    ...err.missing.map((field) => fieldLine(copy, field, "missing")),
    ...err.invalid.map((field) => fieldLine(copy, field, "invalid")),
    ...err.tooLong.map((field) => copy.tooLong.replace("{field}", fieldLabel(copy, field))),
  ];
  if (err.photo === "type") {
    lines.push(copy.photoType);
  }
  if (err.photo === "size") {
    lines.push(copy.photoSize);
  }
  if (err.photo === "limit") {
    lines.push(copy.photoLimit);
  }
  return lines.filter(Boolean).join(" ");
}

function fieldLine(copy: Copy, field: string, kind: "missing" | "invalid") {
  const part = /^parts\.(\d+)\.(name|note)$/.exec(field);
  if (part) {
    const sentence = part[2] === "name" ? copy.partNameRequired : copy.partNoteRequired;
    return sentence.replace("{n}", String(Number(part[1]) + 1));
  }
  if (field === "email") {
    return copy.emailInvalid;
  }
  if (kind === "invalid" && field === "phone") {
    return copy.phoneInvalid;
  }
  if (kind === "invalid" && field === "photos") {
    return copy.photosStale;
  }
  if (kind === "invalid" && field === "part") {
    return copy.partsStale;
  }
  const required: Record<string, string> = {
    customer: copy.customerNameRequired,
    customerName: copy.customerNameRequired,
    brand: copy.brandRequired,
    model: copy.modelRequired,
    note: copy.noteRequired,
    status: kind === "invalid" ? copy.statusSame : copy.statusRequired,
    statusPhoto: copy.statusPhotoOne,
  };
  return required[field] ?? copy.unexpected;
}

function fieldLabel(copy: Copy, field: string) {
  const part = /^parts\.(\d+)\.(name|note)$/.exec(field);
  if (part) {
    return part[2] === "name" ? copy.partName : copy.partNote;
  }
  const labels: Record<string, string> = {
    customerName: copy.customerName,
    brand: copy.brand,
    model: copy.model,
    serial: copy.serial,
    note: copy.conditionNote,
    statusNote: copy.statusNote,
    phone: copy.phone,
    email: copy.email,
    address: copy.address,
  };
  return labels[field] ?? field;
}
