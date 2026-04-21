export interface User {
  id: string;
  email: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Task {
  id: string;
  userId?: string;
  title: string;
  description?: string;
  categoryId?: string | null;
  dueDate?: string | null;
  completed: boolean;
  completedAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface Category {
  id: string;
  name: string;
  createdAt?: string;
}

export interface TaskInput {
  title: string;
  description?: string;
  categoryId?: string | null;
  dueDate?: string | null;
}

export interface CategoryInput {
  name: string;
}

export type StatusFilter = 'all' | 'active' | 'completed';
export type DateRangeFilter = 'all' | 'today' | 'week' | 'month' | 'overdue';
export type SortOrder = 'due-asc' | 'due-desc' | 'created-desc' | 'created-asc';

export interface TaskFilters {
  status: StatusFilter;
  categoryId: string;
  dateRange: DateRangeFilter;
  q: string;
  sort: SortOrder;
}

export class ApiError extends Error {
  status: number;
  body: unknown;
  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}
