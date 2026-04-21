import { useCallback, useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { TaskFilters } from '../api/types';

const STORAGE_KEY = 'todo.filters.v1';

const DEFAULT_FILTERS: TaskFilters = {
  status: 'all',
  categoryId: '',
  dateRange: 'all',
  q: '',
  sort: 'due-asc',
};

function readStorage(): Partial<TaskFilters> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as Partial<TaskFilters>) : {};
  } catch {
    return {};
  }
}

function writeStorage(filters: TaskFilters) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(filters));
  } catch {
    /* ignore */
  }
}

function filtersFromParams(params: URLSearchParams): Partial<TaskFilters> {
  const out: Partial<TaskFilters> = {};
  const status = params.get('status');
  if (status === 'all' || status === 'active' || status === 'completed') out.status = status;
  const categoryId = params.get('category');
  if (categoryId != null) out.categoryId = categoryId;
  const dateRange = params.get('dateRange');
  if (
    dateRange === 'all' ||
    dateRange === 'today' ||
    dateRange === 'week' ||
    dateRange === 'month' ||
    dateRange === 'overdue'
  ) {
    out.dateRange = dateRange;
  }
  const q = params.get('q');
  if (q != null) out.q = q;
  const sort = params.get('sort');
  if (sort === 'due-asc' || sort === 'due-desc' || sort === 'created-desc' || sort === 'created-asc') {
    out.sort = sort;
  }
  return out;
}

function paramsFromFilters(filters: TaskFilters): URLSearchParams {
  const p = new URLSearchParams();
  if (filters.status !== 'all') p.set('status', filters.status);
  if (filters.categoryId) p.set('category', filters.categoryId);
  if (filters.dateRange !== 'all') p.set('dateRange', filters.dateRange);
  if (filters.q) p.set('q', filters.q);
  if (filters.sort !== 'due-asc') p.set('sort', filters.sort);
  return p;
}

export function usePersistentFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const initial = useRef<TaskFilters | null>(null);
  if (initial.current === null) {
    initial.current = {
      ...DEFAULT_FILTERS,
      ...readStorage(),
      ...filtersFromParams(searchParams),
    };
  }
  const [filters, setFiltersState] = useState<TaskFilters>(initial.current);

  useEffect(() => {
    writeStorage(filters);
    const next = paramsFromFilters(filters);
    const current = searchParams.toString();
    if (next.toString() !== current) {
      setSearchParams(next, { replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters]);

  const setFilters = useCallback((patch: Partial<TaskFilters>) => {
    setFiltersState((prev) => ({ ...prev, ...patch }));
  }, []);

  const reset = useCallback(() => setFiltersState(DEFAULT_FILTERS), []);

  return { filters, setFilters, reset };
}
