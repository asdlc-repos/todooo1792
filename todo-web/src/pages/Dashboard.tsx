import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { ApiError, Category, Task } from '../api/types';
import { usePersistentFilters } from '../hooks/usePersistentFilters';
import { useToast } from '../context/ToastContext';
import { formatDateTime, formatRelative, isDueToday, isOverdue } from '../utils/dateUtils';

export default function Dashboard() {
  const { filters, setFilters, reset } = usePersistentFilters();
  const toast = useToast();
  const navigate = useNavigate();

  const [tasks, setTasks] = useState<Task[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchInput, setSearchInput] = useState(filters.q);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [t, c] = await Promise.all([api.listTasks(filters), api.listCategories()]);
      setTasks(Array.isArray(t) ? t : []);
      setCategories(Array.isArray(c) ? c : []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load tasks.');
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    const h = window.setTimeout(() => {
      if (searchInput !== filters.q) setFilters({ q: searchInput });
    }, 300);
    return () => window.clearTimeout(h);
  }, [searchInput, filters.q, setFilters]);

  const categoryName = useMemo(() => {
    const map = new Map<string, string>();
    categories.forEach((c) => map.set(c.id, c.name));
    return (id?: string | null) => (id ? map.get(id) || 'Unknown' : 'No category');
  }, [categories]);

  const sortedTasks = useMemo(() => {
    const copy = [...tasks];
    const cmp = (a: Task, b: Task) => {
      switch (filters.sort) {
        case 'due-asc': {
          // Undated last; completed pushed below active.
          if (a.completed !== b.completed) return a.completed ? 1 : -1;
          const av = a.dueDate ? new Date(a.dueDate).getTime() : Number.POSITIVE_INFINITY;
          const bv = b.dueDate ? new Date(b.dueDate).getTime() : Number.POSITIVE_INFINITY;
          return av - bv;
        }
        case 'due-desc': {
          if (a.completed !== b.completed) return a.completed ? 1 : -1;
          const av = a.dueDate ? new Date(a.dueDate).getTime() : Number.NEGATIVE_INFINITY;
          const bv = b.dueDate ? new Date(b.dueDate).getTime() : Number.NEGATIVE_INFINITY;
          return bv - av;
        }
        case 'created-desc': {
          const av = a.createdAt ? new Date(a.createdAt).getTime() : 0;
          const bv = b.createdAt ? new Date(b.createdAt).getTime() : 0;
          return bv - av;
        }
        case 'created-asc': {
          const av = a.createdAt ? new Date(a.createdAt).getTime() : 0;
          const bv = b.createdAt ? new Date(b.createdAt).getTime() : 0;
          return av - bv;
        }
      }
    };
    copy.sort(cmp);
    return copy;
  }, [tasks, filters.sort]);

  const counts = useMemo(() => {
    const c = { total: tasks.length, active: 0, completed: 0, overdue: 0, dueToday: 0 };
    const byCategory = new Map<string, number>();
    for (const t of tasks) {
      if (t.completed) c.completed++;
      else c.active++;
      if (isOverdue(t.dueDate, t.completed)) c.overdue++;
      if (isDueToday(t.dueDate, t.completed)) c.dueToday++;
      const key = t.categoryId || '';
      byCategory.set(key, (byCategory.get(key) || 0) + 1);
    }
    return { ...c, byCategory };
  }, [tasks]);

  const handleToggleComplete = async (task: Task) => {
    const next = !task.completed;
    setTasks((prev) =>
      prev.map((t) =>
        t.id === task.id
          ? { ...t, completed: next, completedAt: next ? new Date().toISOString() : null }
          : t
      )
    );
    try {
      const updated = await api.completeTask(task.id, next);
      setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, ...updated } : t)));
      toast.success(next ? 'Task completed.' : 'Task reopened.');
    } catch (err) {
      setTasks((prev) =>
        prev.map((t) =>
          t.id === task.id
            ? { ...t, completed: task.completed, completedAt: task.completedAt }
            : t
        )
      );
      toast.error(err instanceof ApiError ? err.message : 'Failed to update task.');
    }
  };

  const handleDelete = async (task: Task) => {
    if (!window.confirm(`Delete "${task.title}"? This cannot be undone.`)) return;
    const prev = tasks;
    setTasks((p) => p.filter((t) => t.id !== task.id));
    try {
      await api.deleteTask(task.id);
      toast.success('Task deleted.');
    } catch (err) {
      setTasks(prev);
      toast.error(err instanceof ApiError ? err.message : 'Failed to delete task.');
    }
  };

  return (
    <div className="dashboard">
      <div className="page-header">
        <div>
          <h1>Dashboard</h1>
          <p className="muted">
            {counts.total} task{counts.total === 1 ? '' : 's'} · {counts.active} active ·{' '}
            {counts.completed} done
            {counts.overdue > 0 && <> · <span className="pill pill-danger">{counts.overdue} overdue</span></>}
            {counts.dueToday > 0 && <> · <span className="pill pill-warn">{counts.dueToday} due today</span></>}
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => navigate('/tasks/new')}>
          + New task
        </button>
      </div>

      <section className="filter-bar" aria-label="Filters">
        <input
          type="search"
          placeholder="Search title or description…"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          aria-label="Search"
          className="search-input"
        />
        <select
          value={filters.status}
          onChange={(e) => setFilters({ status: e.target.value as typeof filters.status })}
          aria-label="Status"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="completed">Completed</option>
        </select>
        <select
          value={filters.categoryId}
          onChange={(e) => setFilters({ categoryId: e.target.value })}
          aria-label="Category"
        >
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
        <select
          value={filters.dateRange}
          onChange={(e) => setFilters({ dateRange: e.target.value as typeof filters.dateRange })}
          aria-label="Date range"
        >
          <option value="all">Any date</option>
          <option value="today">Today</option>
          <option value="week">This week</option>
          <option value="month">This month</option>
          <option value="overdue">Overdue</option>
        </select>
        <select
          value={filters.sort}
          onChange={(e) => setFilters({ sort: e.target.value as typeof filters.sort })}
          aria-label="Sort"
        >
          <option value="due-asc">Due: earliest first</option>
          <option value="due-desc">Due: latest first</option>
          <option value="created-desc">Newest</option>
          <option value="created-asc">Oldest</option>
        </select>
        <button className="btn btn-ghost" onClick={() => { reset(); setSearchInput(''); }}>
          Reset
        </button>
      </section>

      {categories.length > 0 && (
        <section className="category-summary" aria-label="By category">
          <span className="summary-label">By category:</span>
          <span className="chip">Uncategorized · {counts.byCategory.get('') || 0}</span>
          {categories.map((c) => (
            <span key={c.id} className="chip">
              {c.name} · {counts.byCategory.get(c.id) || 0}
            </span>
          ))}
        </section>
      )}

      {error && (
        <div className="error-banner" role="alert">
          {error} <button className="btn btn-link" onClick={load}>Retry</button>
        </div>
      )}

      {loading ? (
        <div className="empty">Loading tasks…</div>
      ) : sortedTasks.length === 0 ? (
        <div className="empty">
          <p>No tasks match your filters.</p>
          <Link className="btn btn-primary" to="/tasks/new">Create your first task</Link>
        </div>
      ) : (
        <ul className="task-list">
          {sortedTasks.map((task) => {
            const overdue = isOverdue(task.dueDate, task.completed);
            const today = isDueToday(task.dueDate, task.completed);
            const cls =
              'task-item' +
              (task.completed ? ' completed' : '') +
              (overdue ? ' overdue' : '') +
              (today ? ' due-today' : '');
            return (
              <li key={task.id} className={cls}>
                <label className="task-check" aria-label={task.completed ? 'Mark incomplete' : 'Mark complete'}>
                  <input
                    type="checkbox"
                    checked={task.completed}
                    onChange={() => handleToggleComplete(task)}
                  />
                  <span className="checkmark" aria-hidden />
                </label>
                <div className="task-body">
                  <div className="task-row">
                    <Link to={`/tasks/${task.id}`} className="task-title">
                      {task.title}
                    </Link>
                    <div className="task-tags">
                      {task.categoryId && (
                        <span className="chip chip-cat">{categoryName(task.categoryId)}</span>
                      )}
                      {overdue && <span className="chip chip-danger">Overdue</span>}
                      {today && <span className="chip chip-warn">Due today</span>}
                    </div>
                  </div>
                  {task.description && <p className="task-desc">{task.description}</p>}
                  <div className="task-meta">
                    {task.dueDate ? (
                      <span>Due {formatDateTime(task.dueDate)} ({formatRelative(task.dueDate)})</span>
                    ) : (
                      <span className="muted">No due date</span>
                    )}
                    {task.completed && task.completedAt && (
                      <span> · Completed {formatDateTime(task.completedAt)}</span>
                    )}
                  </div>
                </div>
                <div className="task-actions">
                  <Link className="btn btn-ghost" to={`/tasks/${task.id}`}>Edit</Link>
                  <button className="btn btn-danger-ghost" onClick={() => handleDelete(task)}>
                    Delete
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
