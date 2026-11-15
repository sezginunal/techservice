import type { Lang } from "./i18n";

export type Role = "technician" | "owner" | "admin";

export type Brand = {
  name: string;
  logo: boolean;
};

export type Shop = Brand & {
  lang: Lang;
};

export type CatalogModel = {
  name: string;
  image: boolean;
};

export type CatalogBrand = {
  brand: string;
  models: CatalogModel[];
};

export const logoSrc = "/api/brand/logo";

export function catalogImageSrc(brand: string, model: string) {
  const params = new URLSearchParams({ brand, model });
  return `/api/catalog/image?${params}`;
}

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
  address: string;
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
  deviceId: string;
  brand: string;
  model: string;
  image: boolean;
  serial: string;
  note: string;
  status: string;
  createdAt: string;
  customer: Customer;
  parts: Part[];
  photos: Photo[];
  history: StatusChange[];
};

export type HomeDay = {
  date: string;
  count: number;
};

export type HomeStatus = {
  status: string;
  count: number;
};

export type HomeModel = {
  brand: string;
  model: string;
  count: number;
  image: boolean;
};

export type HomeIdle = {
  id: string;
  brand: string;
  model: string;
  serial: string;
  customerName: string;
  status: string;
  days: number;
  since: string;
};

export type HomeRecent = {
  id: string;
  brand: string;
  model: string;
  serial: string;
  image: boolean;
  status: string;
  createdAt: string;
};

export type HomeQueue = {
  status: string;
  more: boolean;
  items: HomeIdle[];
};

export type HomeSummary = {
  registered7: number;
  registered30: number;
  days: HomeDay[];
  inShop: number;
  delivered: number;
  idle7: number;
  byStatus: HomeStatus[];
  queues: HomeQueue[];
  models: HomeModel[];
  other: number;
  unchanged: HomeIdle[];
  unchangedMore: boolean;
  latest: HomeRecent[];
};

export type ServiceState = {
  code: string;
  nameEn: string;
  nameTr: string;
  position: number;
  active: boolean;
  showOnHome: boolean;
  fixed: boolean;
  locked: boolean;
  used: boolean;
};

export type DeviceItem = {
  id: string;
  brand: string;
  model: string;
  serial: string;
  image: boolean;
  customerName: string;
  status: string;
  createdAt: string;
};

export type Identity = {
  id: string;
  brand: string;
  model: string;
  serial: string;
  image: boolean;
};

export type EarlierService = {
  id: string;
  customerName: string;
  phone: string;
  status: string;
  createdAt: string;
};

export type DeviceService = {
  id: string;
  customerName: string;
  status: string;
  createdAt: string;
};

export type SparePart = {
  id: string;
  brand: string;
  model: string;
  onHand: number;
  existed?: boolean;
  added?: boolean;
};

export type SpareMove = {
  id: number;
  at: string;
  actorUsername: string;
  quantity: number;
  kind: "receipt" | "write_off" | "use" | "return";
  note: string;
  status?: string;
  serviceId?: string;
  partId?: string;
  brand?: string;
  model?: string;
  deviceBrand?: string;
  deviceModel?: string;
  serial?: string;
};

export type SpareReturnable = {
  id: string;
  brand: string;
  model: string;
  returnable: number;
};

export type DeviceMatch = {
  matched: boolean;
  deviceId: string;
  openServiceId: string;
  items: EarlierService[];
  hasMore: boolean;
};

export type AuditDetail = {
  name?: string;
  previousName?: string;
  brand?: string;
  model?: string;
  serial?: string;
  previousBrand?: string;
  previousModel?: string;
  previousSerial?: string;
  previousCustomer?: string;
  uuid?: string;
  customer?: string;
  username?: string;
  role?: Role;
  from?: string;
  to?: string;
  nameEn?: string;
  nameTr?: string;
  previousNameEn?: string;
  previousNameTr?: string;
  active?: string;
  previousActive?: string;
  home?: string;
  previousHome?: string;
  showOnHome?: string;
  quantity?: number;
  deviceBrand?: string;
  deviceModel?: string;
  status?: string;
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
  taken: string[];
  photo: string;
  serviceId: string;
  onHand: number;
  returnable: number;

  constructor(
    code: string,
    extra?: {
      missing?: string[];
      invalid?: string[];
      tooLong?: string[];
      taken?: string[];
      photo?: string;
      serviceId?: string;
      onHand?: number;
      returnable?: number;
    },
  ) {
    super(code);
    this.code = code;
    this.missing = extra?.missing ?? [];
    this.invalid = extra?.invalid ?? [];
    this.tooLong = extra?.tooLong ?? [];
    this.taken = extra?.taken ?? [];
    this.photo = extra?.photo ?? "";
    this.serviceId = extra?.serviceId ?? "";
    this.onHand = extra?.onHand ?? 0;
    this.returnable = extra?.returnable ?? 0;
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
    taken?: string[];
    photo?: string;
    serviceId?: string;
    onHand?: number;
    returnable?: number;
  };
  if (!response.ok) {
    throw new ApiError(body.error ?? "unexpected", body);
  }
  return body as T;
}

export function photoSrc(serviceId: string, photoId: string) {
  return `/api/services/${serviceId}/photos/${photoId}`;
}

export function statusPhotoSrc(serviceId: string, changeId: number) {
  return `/api/services/${serviceId}/status/${changeId}/photo`;
}
