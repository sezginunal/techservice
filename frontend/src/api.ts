export type Role = "technician" | "owner" | "admin";

export type User = {
  id: number;
  username: string;
  role: Role;
  active: boolean;
};

export type Customer = {
  id: number;
  name: string;
  phone: string;
  email: string;
};

export type Photo = {
  id: string;
  mediaType: string;
  size: number;
};

export type Part = {
  id: string;
  name: string;
  note: string;
  photos: Photo[];
};

export type StatusChange = {
  id: number;
  actorUsername: string;
  from: string;
  to: string;
  note: string;
  hasPhoto: boolean;
  createdAt: string;
};

export type Device = {
  id: string;
  name: string;
  brand: string;
  model: string;
  serial: string;
  note: string;
  status: string;
  createdAt: string;
  customer: Customer;
  parts: Part[];
  photos: Photo[];
  history: StatusChange[];
};

export type DeviceItem = {
  id: string;
  name: string;
  brand: string;
  model: string;
  customerName: string;
  status: string;
  createdAt: string;
};

export type AuditDetail = {
  name?: string;
  previousName?: string;
  uuid?: string;
  customer?: string;
  username?: string;
  role?: Role;
  from?: string;
  to?: string;
};

export type AuditItem = {
  id: number;
  at: string;
  actorUsername: string;
  actorRole: Role;
  event: string;
  detail: AuditDetail;
};

export type Page<T> = {
  items: T[];
  hasMore: boolean;
};

export class ApiError extends Error {
  code: string;
  missing: string[];
  invalid: string[];
  tooLong: string[];
  photo: string;

  constructor(
    code: string,
    extra?: { missing?: string[]; invalid?: string[]; tooLong?: string[]; photo?: string },
  ) {
    super(code);
    this.code = code;
    this.missing = extra?.missing ?? [];
    this.invalid = extra?.invalid ?? [];
    this.tooLong = extra?.tooLong ?? [];
    this.photo = extra?.photo ?? "";
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (init?.body != null && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, { ...init, headers });

  if (response.status === 204) {
    return undefined as T;
  }

  const body = (await response.json().catch(() => ({}))) as {
    error?: string;
    missing?: string[];
    invalid?: string[];
    tooLong?: string[];
    photo?: string;
  };
  if (!response.ok) {
    throw new ApiError(body.error ?? "unexpected", body);
  }
  return body as T;
}

export function photoSrc(deviceId: string, photoId: string) {
  return `/api/devices/${deviceId}/photos/${photoId}`;
}

export function statusPhotoSrc(deviceId: string, changeId: number) {
  return `/api/devices/${deviceId}/status/${changeId}/photo`;
}
