import type { AuthResponse, Component, Finding, Project, Scan } from "../types";

const TOKEN_KEY = "devsecurity.session";

export class ApiError extends Error {
  constructor(message: string, public status: number, public code?: string) {
    super(message);
  }
}

export function getToken(): string | null {
  return sessionStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (token) {
    sessionStorage.setItem(TOKEN_KEY, token);
  } else {
    sessionStorage.removeItem(TOKEN_KEY);
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = getToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, { ...init, headers });
  if (!response.ok) {
    let message = `请求失败 (${response.status})`;
    let code: string | undefined;
    try {
      const payload = await response.json();
      message = payload.error?.message ?? message;
      code = payload.error?.code;
    } catch {
      // Keep a stable message when a gateway returns non-JSON data.
    }
    if (response.status === 401) {
      setToken(null);
    }
    throw new ApiError(message, response.status, code);
  }
  return response.json() as Promise<T>;
}

export const api = {
  login: (username: string, password: string) =>
    request<AuthResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password })
    }),
  me: () => request<AuthResponse["user"]>("/api/v1/auth/me"),
  projects: () => request<Project[]>("/api/v1/projects"),
  project: (id: string) => request<Project>(`/api/v1/projects/${id}`),
  createProject: (name: string, description: string) =>
    request<Project>("/api/v1/projects", {
      method: "POST",
      body: JSON.stringify({ name, description })
    }),
  scans: (projectId: string) => request<Scan[]>(`/api/v1/projects/${projectId}/scans`),
  scan: (scanId: string) => request<Scan>(`/api/v1/scans/${scanId}`),
  components: (scanId: string) => request<Component[]>(`/api/v1/scans/${scanId}/components`),
  findings: (scanId: string) => request<Finding[]>(`/api/v1/scans/${scanId}/findings`),
  createGitScan: (projectId: string, gitUrl: string, revision: string) =>
    request<Scan>(`/api/v1/projects/${projectId}/scans`, {
      method: "POST",
      body: JSON.stringify({ gitUrl, revision })
    }),
  createUploadScan: (projectId: string, file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<Scan>(`/api/v1/projects/${projectId}/scans/upload`, {
      method: "POST",
      body: form
    });
  }
};
