export type Role = "technician" | "owner" | "admin";

export type User = {
  id: number;
  username: string;
  role: Role;
  active: boolean;
};

export class ApiError extends Error {
  code: string;

  constructor(code: string) {
    super(code);
    this.code = code;
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const body = (await response.json().catch(() => ({}))) as { error?: string };
  if (!response.ok) {
    throw new ApiError(body.error ?? "unexpected");
  }
  return body as T;
}
