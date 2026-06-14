import type {
  Task,
  TaskActivity,
  TaskInput,
  TaskListResult,
  TaskQuery,
  User,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const TOKEN_KEY = "tm_token";

// ── token storage ───────────────────────────────────────
// The JWT is kept in localStorage so the session survives a page refresh.
export const tokenStore = {
  get(): string | null {
    if (typeof window === "undefined") return null;
    return window.localStorage.getItem(TOKEN_KEY);
  },
  set(token: string) {
    window.localStorage.setItem(TOKEN_KEY, token);
  },
  clear() {
    window.localStorage.removeItem(TOKEN_KEY);
  },
};

// ApiError carries the HTTP status plus the parsed error envelope from the API.
export class ApiError extends Error {
  status: number;
  details?: unknown;
  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.status = status;
    this.details = details;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = tokenStore.get();
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type") && options.body) {
    headers.set("Content-Type", "application/json");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);

  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, { ...options, headers });
  } catch {
    throw new ApiError(0, "Could not reach the server. Is the backend running?");
  }

  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);

  if (!res.ok) {
    const message =
      body?.error?.message || `Request failed (${res.status})`;
    throw new ApiError(res.status, message, body?.error?.details);
  }
  return body as T;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export const api = {
  baseUrl: API_URL,

  signup(email: string, password: string) {
    return request<AuthResponse>("/api/auth/signup", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  login(email: string, password: string) {
    return request<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  me() {
    return request<{ user: User }>("/api/auth/me");
  },

  listTasks(query: TaskQuery = {}) {
    const params = new URLSearchParams();
    if (query.status) params.set("status", query.status);
    if (query.search) params.set("search", query.search);
    if (query.sort) params.set("sort", query.sort);
    if (query.order) params.set("order", query.order);
    if (query.page) params.set("page", String(query.page));
    if (query.pageSize) params.set("pageSize", String(query.pageSize));
    if (query.scope) params.set("scope", query.scope);
    const qs = params.toString();
    return request<TaskListResult>(`/api/tasks${qs ? `?${qs}` : ""}`);
  },

  getTask(id: string) {
    return request<{ task: Task }>(`/api/tasks/${id}`);
  },

  createTask(input: TaskInput) {
    return request<{ task: Task }>("/api/tasks", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  updateTask(id: string, input: Partial<TaskInput>) {
    return request<{ task: Task }>(`/api/tasks/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  deleteTask(id: string) {
    return request<void>(`/api/tasks/${id}`, { method: "DELETE" });
  },

  getActivity(id: string) {
    return request<{ activity: TaskActivity[] }>(`/api/tasks/${id}/activity`);
  },
};
