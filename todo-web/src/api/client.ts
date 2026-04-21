import {
  ApiError,
  AuthResponse,
  Category,
  CategoryInput,
  Task,
  TaskFilters,
  TaskInput,
  User,
} from './types';

declare global {
  interface Window {
    __APP_CONFIG__?: { API_BASE_URL?: string };
  }
}

const TOKEN_KEY = 'todo.jwt';
const USER_KEY = 'todo.user';

function resolveBaseUrl(): string {
  const runtime = typeof window !== 'undefined' ? window.__APP_CONFIG__?.API_BASE_URL : undefined;
  if (runtime && runtime.trim() && runtime !== '__API_BASE_URL__') {
    return runtime.replace(/\/$/, '');
  }
  // Default: same origin under /api (nginx proxies to todo-api). Works without any env config.
  if (typeof window !== 'undefined') {
    return `${window.location.origin}/api`;
  }
  return '/api';
}

export const tokenStore = {
  get: (): string | null => {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set: (token: string) => {
    try {
      localStorage.setItem(TOKEN_KEY, token);
    } catch {
      /* ignore */
    }
  },
  clear: () => {
    try {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    } catch {
      /* ignore */
    }
  },
};

export const userStore = {
  get: (): User | null => {
    try {
      const raw = localStorage.getItem(USER_KEY);
      return raw ? (JSON.parse(raw) as User) : null;
    } catch {
      return null;
    }
  },
  set: (user: User) => {
    try {
      localStorage.setItem(USER_KEY, JSON.stringify(user));
    } catch {
      /* ignore */
    }
  },
};

type RequestOptions = {
  method?: string;
  body?: unknown;
  auth?: boolean;
  query?: Record<string, string | undefined>;
};

let onUnauthorized: (() => void) | null = null;
export function registerUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, auth = true, query } = opts;
  const url = new URL(resolveBaseUrl() + path, window.location.origin);
  if (query) {
    Object.entries(query).forEach(([k, v]) => {
      if (v != null && v !== '') url.searchParams.set(k, v);
    });
  }
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (auth) {
    const token = tokenStore.get();
    if (token) headers['Authorization'] = `Bearer ${token}`;
  }
  let res: Response;
  try {
    res = await fetch(url.toString(), {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (err) {
    throw new ApiError(0, 'Network error — please check your connection.');
  }
  if (res.status === 401 && auth) {
    tokenStore.clear();
    if (onUnauthorized) onUnauthorized();
    throw new ApiError(401, 'Your session has expired. Please log in again.');
  }
  let payload: unknown = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = text;
    }
  }
  if (!res.ok) {
    const message =
      (payload && typeof payload === 'object' && 'message' in (payload as Record<string, unknown>)
        ? String((payload as Record<string, unknown>).message)
        : null) ||
      (payload && typeof payload === 'object' && 'error' in (payload as Record<string, unknown>)
        ? String((payload as Record<string, unknown>).error)
        : null) ||
      defaultMessageFor(res.status);
    throw new ApiError(res.status, message, payload);
  }
  return payload as T;
}

function defaultMessageFor(status: number): string {
  if (status === 400) return 'Invalid request.';
  if (status === 403) return 'You do not have permission to perform this action.';
  if (status === 404) return 'Not found.';
  if (status === 409) return 'Conflict — item already exists.';
  if (status === 422) return 'Validation failed.';
  if (status === 423) return 'Account locked. Try again later.';
  if (status === 429) return 'Too many requests. Please slow down.';
  if (status >= 500) return 'Server error. Please try again.';
  return `Request failed (${status}).`;
}

export const api = {
  health: () => request<unknown>('/health', { auth: false }),

  register: (email: string, password: string) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: { email, password },
      auth: false,
    }),

  login: (email: string, password: string) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: { email, password },
      auth: false,
    }),

  logout: () => request<void>('/auth/logout', { method: 'POST' }).catch(() => undefined),

  passwordReset: (email: string) =>
    request<{ message?: string }>('/auth/password-reset', {
      method: 'POST',
      body: { email },
      auth: false,
    }),

  updateAccount: (payload: { email?: string; password?: string; currentPassword?: string }) =>
    request<User>('/account', { method: 'PUT', body: payload }),

  listTasks: (filters: Partial<TaskFilters>) =>
    request<Task[]>('/tasks', {
      query: {
        category: filters.categoryId || undefined,
        status: filters.status && filters.status !== 'all' ? filters.status : undefined,
        dateRange: filters.dateRange && filters.dateRange !== 'all' ? filters.dateRange : undefined,
        q: filters.q || undefined,
      },
    }),

  getTask: (id: string) => request<Task>(`/tasks/${encodeURIComponent(id)}`),

  createTask: (input: TaskInput) =>
    request<Task>('/tasks', { method: 'POST', body: input }),

  updateTask: (id: string, input: TaskInput) =>
    request<Task>(`/tasks/${encodeURIComponent(id)}`, { method: 'PUT', body: input }),

  deleteTask: (id: string) =>
    request<void>(`/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' }),

  completeTask: (id: string, completed: boolean) =>
    request<Task>(`/tasks/${encodeURIComponent(id)}/complete`, {
      method: 'POST',
      body: { completed },
    }),

  listCategories: () => request<Category[]>('/categories'),

  createCategory: (input: CategoryInput) =>
    request<Category>('/categories', { method: 'POST', body: input }),

  updateCategory: (id: string, input: CategoryInput) =>
    request<Category>(`/categories/${encodeURIComponent(id)}`, { method: 'PUT', body: input }),

  deleteCategory: (id: string) =>
    request<void>(`/categories/${encodeURIComponent(id)}`, { method: 'DELETE' }),
};
